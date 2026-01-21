#!/bin/bash
# Wrapper script for running Pulumi commands with automatic token refresh
#
# Usage:
#   ./scripts/run-pulumi.sh preview
#   ./scripts/run-pulumi.sh up
#   ./scripts/run-pulumi.sh up --yes
#   ./scripts/run-pulumi.sh destroy
#   ./scripts/run-pulumi.sh stack output
#
# This script automatically:
#   1. Checks/starts port-forwards to the prod cluster
#   2. Gets a fresh OAuth token from Keycloak
#   3. Runs the specified Pulumi command
#
# Prerequisites:
#   - kubectl configured with kind-lagoon-prod context
#   - Multi-cluster setup running
#   - Python virtual environment with provider installed

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Activate virtual environment if it exists
if [ -d "venv" ] && [ -f "venv/bin/activate" ]; then
    # shellcheck source=/dev/null
    source "venv/bin/activate"
fi

CONTEXT="${KUBE_CONTEXT:-kind-lagoon-prod}"
NAMESPACE="${LAGOON_NAMESPACE:-lagoon-core}"
# Service names in multi-cluster setup include the Helm release prefix
KEYCLOAK_SVC="${KEYCLOAK_SVC:-prod-core-lagoon-core-keycloak}"
API_SVC="${API_SVC:-prod-core-lagoon-core-api}"
KEYCLOAK_SECRET="${KEYCLOAK_SECRET:-prod-core-lagoon-core-keycloak}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if port-forwards are running
check_port_forwards() {
    if curl -s --connect-timeout 2 "http://localhost:8080/auth/" >/dev/null 2>&1 && \
       curl -s --connect-timeout 2 "http://localhost:7080/" >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

# Start port-forwards
start_port_forwards() {
    log_info "Starting port-forwards to $CONTEXT..."

    # Kill existing port-forwards
    pkill -f "port-forward.*lagoon-core" 2>/dev/null || true
    sleep 1

    # Start new port-forwards
    kubectl --context "$CONTEXT" port-forward -n "$NAMESPACE" "svc/$KEYCLOAK_SVC" 8080:8080 >/dev/null 2>&1 &
    kubectl --context "$CONTEXT" port-forward -n "$NAMESPACE" "svc/$API_SVC" 7080:80 >/dev/null 2>&1 &

    # Wait for them to be ready
    local retries=10
    while ! check_port_forwards && [ $retries -gt 0 ]; do
        sleep 1
        retries=$((retries - 1))
    done

    if ! check_port_forwards; then
        log_error "Failed to start port-forwards"
        exit 1
    fi

    log_info "Port-forwards ready"
}

# Get OAuth token
get_token() {
    log_info "Getting OAuth token..."

    # Get password from secret
    local password
    password=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret "$KEYCLOAK_SECRET" \
        -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d)

    # Get token
    local response
    response=$(curl -s "http://localhost:8080/auth/realms/lagoon/protocol/openid-connect/token" \
        -d "grant_type=password" \
        -d "client_id=lagoon-ui" \
        -d "username=lagoonadmin" \
        -d "password=$password" 2>&1)

    local token
    token=$(echo "$response" | jq -r '.access_token // empty')

    if [ -z "$token" ]; then
        # Check if we need to enable Direct Access Grants
        if echo "$response" | grep -q "Client not allowed for direct access grants"; then
            log_warn "Enabling Direct Access Grants in Keycloak..."

            local admin_password
            admin_password=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret "$KEYCLOAK_SECRET" \
                -o jsonpath='{.data.KEYCLOAK_ADMIN_PASSWORD}' | base64 -d)

            local admin_token
            admin_token=$(curl -s "http://localhost:8080/auth/realms/master/protocol/openid-connect/token" \
                -d "grant_type=password" \
                -d "client_id=admin-cli" \
                -d "username=admin" \
                -d "password=$admin_password" | jq -r '.access_token')

            local client_id
            client_id=$(curl -s "http://localhost:8080/auth/admin/realms/lagoon/clients?clientId=lagoon-ui" \
                -H "Authorization: Bearer $admin_token" | jq -r '.[0].id')

            curl -s -X PUT "http://localhost:8080/auth/admin/realms/lagoon/clients/$client_id" \
                -H "Authorization: Bearer $admin_token" \
                -H "Content-Type: application/json" \
                -d '{"directAccessGrantsEnabled": true}' >/dev/null

            # Retry getting token
            response=$(curl -s "http://localhost:8080/auth/realms/lagoon/protocol/openid-connect/token" \
                -d "grant_type=password" \
                -d "client_id=lagoon-ui" \
                -d "username=lagoonadmin" \
                -d "password=$password" 2>&1)

            token=$(echo "$response" | jq -r '.access_token // empty')
        fi
    fi

    # Check if user doesn't exist (invalid_grant with "Invalid user credentials")
    if [ -z "$token" ] && echo "$response" | grep -q "Invalid user credentials"; then
        log_warn "lagoonadmin user not found. Creating..."
        ./scripts/create-lagoon-admin.sh

        # Retry getting token
        response=$(curl -s "http://localhost:8080/auth/realms/lagoon/protocol/openid-connect/token" \
            -d "grant_type=password" \
            -d "client_id=lagoon-ui" \
            -d "username=lagoonadmin" \
            -d "password=$password" 2>&1)

        token=$(echo "$response" | jq -r '.access_token // empty')
    fi

    if [ -z "$token" ]; then
        log_error "Failed to get OAuth token"
        echo "$response" >&2
        exit 1
    fi

    export LAGOON_TOKEN="$token"
    export LAGOON_API_URL="http://localhost:7080/graphql"

    log_info "Token acquired (valid for ~5 minutes)"
}

# Main execution
main() {
    if [ $# -eq 0 ]; then
        echo "Usage: $0 <pulumi-command> [args...]"
        echo ""
        echo "Examples:"
        echo "  $0 preview"
        echo "  $0 up"
        echo "  $0 up --yes"
        echo "  $0 destroy"
        echo "  $0 stack output"
        exit 1
    fi

    # Check cluster connectivity - if cluster doesn't exist yet, skip token refresh
    # (Pulumi will create the cluster on first run)
    if ! kubectl --context "$CONTEXT" cluster-info >/dev/null 2>&1; then
        log_warn "Cluster '$CONTEXT' not found - running Pulumi without token refresh"
        log_warn "This is expected on first run (Pulumi will create the clusters)"
        log_info "Running: pulumi $*"
        echo ""
        pulumi "$@"
        return
    fi

    # Cluster exists - do full token refresh flow
    # Ensure port-forwards are running
    if ! check_port_forwards; then
        start_port_forwards
    else
        log_info "Port-forwards already running"
    fi

    # Get fresh token
    get_token

    # Run pulumi command
    log_info "Running: pulumi $*"
    echo ""

    pulumi "$@"
}

main "$@"
