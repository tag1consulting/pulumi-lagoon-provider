#!/bin/bash
# Ensure Knex migrations have been run on the Lagoon API database
#
# Lagoon v2.30.0 has a bug where the init container only runs migrations from
# dist/migrations/lagoon/migrations/ (1 file) instead of the Knex migrations
# in database/migrations/ (44 files that create base tables like openshift).
#
# This script checks if the base schema tables exist and runs Knex migrations
# if they are missing.
#
# Usage:
#   ./scripts/ensure-knex-migrations.sh
#
# Configuration via environment variables or LAGOON_PRESET:
#   LAGOON_PRESET=single       - Single-cluster (default)
#   LAGOON_PRESET=multi-prod   - Multi-cluster production
#
# See scripts/common.sh for all configuration options.

set -e

# Determine script location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source common configuration
source "$SCRIPT_DIR/common.sh"

# API pod label selector varies by deployment type
get_api_pod() {
    local label_selector
    local pod

    # Try different label selectors based on deployment type
    if [ "${LAGOON_PRESET:-single}" = "single" ]; then
        label_selector="app.kubernetes.io/name=lagoon-core,app.kubernetes.io/component=api"
    else
        # Multi-cluster uses Helm release name prefix
        label_selector="app.kubernetes.io/instance=prod-core,app.kubernetes.io/component=api"
    fi

    pod=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get pods \
        -l "$label_selector" \
        -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

    # Fallback: grep for the actual api pod (not actions-handler or other api-like pods)
    # The API pod name ends with -api- followed by a deployment hash
    if [ -z "$pod" ]; then
        pod=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get pods \
            --no-headers 2>/dev/null | grep -E 'lagoon-core-api-[a-z0-9]+-[a-z0-9]+' | head -1 | awk '{print $1}')
    fi

    # Another fallback: more general pattern
    if [ -z "$pod" ]; then
        pod=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get pods \
            --no-headers 2>/dev/null | grep -E '\-api\-[a-z0-9]+-[a-z0-9]+$' | grep -v actions-handler | head -1 | awk '{print $1}')
    fi

    echo "$pod"
}

# Wait for API pod to be ready
# Note: log messages go to stderr so they don't interfere with the pod name output
wait_for_api_pod() {
    log_info "Waiting for API pod to be ready..." >&2

    local retries=30
    local pod=""

    while [ $retries -gt 0 ]; do
        pod=$(get_api_pod)

        if [ -n "$pod" ]; then
            local status
            status=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get pod "$pod" \
                -o jsonpath='{.status.phase}' 2>/dev/null)

            if [ "$status" = "Running" ]; then
                # Check if the api container is ready
                local ready
                ready=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get pod "$pod" \
                    -o jsonpath='{.status.containerStatuses[?(@.name=="api")].ready}' 2>/dev/null)

                if [ "$ready" = "true" ]; then
                    log_info "API pod '$pod' is ready" >&2
                    echo "$pod"
                    return 0
                fi
            fi
        fi

        log_debug "Waiting for API pod... (attempt $((31 - retries))/30)" >&2
        sleep 5
        retries=$((retries - 1))
    done

    log_error "Timeout waiting for API pod to be ready" >&2
    return 1
}

# Check if Knex migrations have been run by checking for the openshift table
check_knex_migrations() {
    local pod="$1"

    log_info "Checking if Knex migrations have been run..."

    # Check if the openshift table exists (created by Knex migrations)
    local result
    result=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" exec "$pod" -c api -- \
        sh -c 'node -e "
const knex = require(\"knex\")({
    client: \"mysql2\",
    connection: {
        host: process.env.API_DB_HOST || \"mariadb\",
        port: process.env.API_DB_PORT || 3306,
        user: process.env.API_DB_USER || \"api\",
        password: process.env.API_DB_PASSWORD || \"api\",
        database: process.env.API_DB_DATABASE || \"infrastructure\"
    }
});

knex.schema.hasTable(\"openshift\").then(exists => {
    console.log(exists ? \"EXISTS\" : \"MISSING\");
    process.exit(0);
}).catch(err => {
    console.error(err.message);
    process.exit(1);
}).finally(() => knex.destroy());
"' 2>&1)

    if echo "$result" | grep -q "EXISTS"; then
        log_info "Knex migrations have already been run (openshift table exists)"
        return 0
    elif echo "$result" | grep -q "MISSING"; then
        log_warn "Knex migrations have NOT been run (openshift table missing)"
        return 1
    else
        log_error "Failed to check migration status: $result"
        return 2
    fi
}

# Run Knex migrations
run_knex_migrations() {
    local pod="$1"

    log_info "Running Knex migrations..."

    # Run the migrations
    local output
    output=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" exec "$pod" -c api -- \
        sh -c 'cd /app/services/api && npx knex migrate:latest --knexfile database/knexfile.js' 2>&1)

    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        log_info "Knex migrations completed successfully"
        echo "$output" | grep -E "^Batch|^Migration|migrations" || true
        return 0
    else
        log_error "Knex migrations failed"
        echo "$output" >&2
        return 1
    fi
}

# Get migration status details
get_migration_status() {
    local pod="$1"

    log_info "Getting migration status details..."

    kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" exec "$pod" -c api -- \
        sh -c 'cd /app/services/api && npx knex migrate:status --knexfile database/knexfile.js' 2>&1 || true
}

# Main execution
main() {
    log_info "Ensuring Knex migrations are run..."
    log_debug "Using context: $KUBE_CONTEXT, namespace: $LAGOON_NAMESPACE"

    # Check cluster connectivity
    if ! check_cluster_connectivity; then
        log_error "Cannot connect to cluster"
        exit 1
    fi

    # Wait for API pod
    local api_pod
    api_pod=$(wait_for_api_pod)

    if [ -z "$api_pod" ]; then
        log_error "Could not find API pod"
        exit 1
    fi

    # Check if migrations have been run
    if check_knex_migrations "$api_pod"; then
        log_info "Database schema is up to date"

        if [ "${SHOW_STATUS:-}" = "1" ]; then
            get_migration_status "$api_pod"
        fi

        exit 0
    fi

    # Migrations need to be run
    log_warn "Database schema is missing base tables"

    if [ "${DRY_RUN:-}" = "1" ]; then
        log_info "DRY_RUN=1, not running migrations"
        get_migration_status "$api_pod"
        exit 0
    fi

    # Run migrations
    if run_knex_migrations "$api_pod"; then
        log_info "Database schema has been initialized"

        # Verify the migrations worked
        if check_knex_migrations "$api_pod"; then
            log_info "Migration verification passed"
            exit 0
        else
            log_error "Migration verification failed - tables still missing"
            exit 1
        fi
    else
        log_error "Failed to run migrations"
        exit 1
    fi
}

main "$@"
