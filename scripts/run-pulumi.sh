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
#   1. Checks/starts port-forwards
#   2. Gets a fresh OAuth token
#   3. Runs the specified Pulumi command
#
# See scripts/common.sh for configuration options.
#
# The script will look for a Python virtual environment in these locations:
#   1. ./venv (local to current directory)
#   2. ../../venv (for examples/*/scripts/run-pulumi.sh symlinks)
#   3. $REPO_ROOT/venv (repository root)

set -e

# Determine script location and repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# If running from a symlink, resolve the actual script location
if [ -L "${BASH_SOURCE[0]}" ]; then
    REAL_SCRIPT="$(readlink -f "${BASH_SOURCE[0]}")"
    SCRIPT_DIR="$(dirname "$REAL_SCRIPT")"
fi

REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Source common configuration
source "$SCRIPT_DIR/common.sh"

# Find and activate virtual environment
activate_venv() {
    local venv_locations=(
        "./venv"
        "../../venv"
        "$REPO_ROOT/venv"
    )

    for venv in "${venv_locations[@]}"; do
        if [ -d "$venv" ] && [ -f "$venv/bin/activate" ]; then
            log_debug "Activating venv: $venv"
            # shellcheck source=/dev/null
            source "$venv/bin/activate"
            return 0
        fi
    done

    log_warn "No virtual environment found. Provider may not be installed."
    return 1
}

# Start port-forwards if needed
ensure_port_forwards() {
    if check_port_forwards; then
        log_info "Port-forwards already running"
        return 0
    fi

    log_info "Starting port-forwards..."

    # Kill existing port-forwards
    pkill -f "port-forward.*$KEYCLOAK_SVC" 2>/dev/null || true
    pkill -f "port-forward.*$API_SVC" 2>/dev/null || true
    sleep 1

    # Start new port-forwards
    kubectl --context "$KUBE_CONTEXT" port-forward -n "$LAGOON_NAMESPACE" "svc/$KEYCLOAK_SVC" 8080:8080 >/dev/null 2>&1 &
    kubectl --context "$KUBE_CONTEXT" port-forward -n "$LAGOON_NAMESPACE" "svc/$API_SVC" 7080:80 >/dev/null 2>&1 &

    # Wait for them to be ready
    local retries=10
    while ! check_port_forwards && [ $retries -gt 0 ]; do
        sleep 1
        retries=$((retries - 1))
    done

    if ! check_port_forwards; then
        log_error "Failed to start port-forwards"
        return 1
    fi

    log_info "Port-forwards ready"
}

# Get admin JWT token (signed with JWTSECRET for full API access)
get_admin_jwt_token() {
    log_info "Getting admin JWT token..."

    # Get JWTSECRET from core secrets
    local jwt_secret
    jwt_secret=$(get_secret_value "$CORE_SECRETS" "JWTSECRET")

    if [ -z "$jwt_secret" ]; then
        log_error "Could not get JWTSECRET from $CORE_SECRETS"
        return 1
    fi

    # Write secret to temp file to avoid shell escaping issues
    local secret_file
    secret_file=$(mktemp)
    echo "$jwt_secret" > "$secret_file"

    # Generate admin JWT token using Python
    local token
    token=$(python3 << EOF
import jwt
import time

with open('$secret_file', 'r') as f:
    secret = f.read().strip()

now = int(time.time())
payload = {
    'role': 'admin',
    'iss': 'lagoon-api',
    'sub': 'lagoonadmin',
    'aud': 'api.dev',
    'iat': now,
    'exp': now + 3600  # 1 hour validity
}
print(jwt.encode(payload, secret, algorithm='HS256'))
EOF
)

    # Clean up temp file
    rm -f "$secret_file"

    if [ -z "$token" ] || echo "$token" | grep -q "Traceback\|Error\|ModuleNotFoundError"; then
        log_error "Failed to generate admin JWT token"
        echo "$token" >&2
        return 1
    fi

    export LAGOON_TOKEN="$token"
    export LAGOON_API_URL="http://localhost:7080/graphql"

    log_info "Admin token acquired (valid for 1 hour)"
}

# Legacy OAuth token function (kept for reference, but OAuth doesn't have deploy target permissions)
get_oauth_token() {
    log_info "Getting OAuth token..."

    # Get password from secret
    local password
    password=$(get_secret_value "$KEYCLOAK_SECRET" "KEYCLOAK_LAGOON_ADMIN_PASSWORD")

    if [ -z "$password" ]; then
        log_error "Could not get password from $KEYCLOAK_SECRET"
        return 1
    fi

    # Get token
    local response
    response=$(curl -s "http://localhost:8080/auth/realms/lagoon/protocol/openid-connect/token" \
        -d "grant_type=password" \
        -d "client_id=lagoon-ui" \
        -d "username=lagoonadmin" \
        -d "password=$password" 2>&1)

    local token
    token=$(echo "$response" | jq -r '.access_token // empty')

    # Handle Direct Access Grants not enabled
    if [ -z "$token" ] && echo "$response" | grep -q "Client not allowed for direct access grants"; then
        log_warn "Enabling Direct Access Grants in Keycloak..."

        local admin_password
        admin_password=$(get_secret_value "$KEYCLOAK_SECRET" "KEYCLOAK_ADMIN_PASSWORD")

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

    # Handle user doesn't exist
    if [ -z "$token" ] && echo "$response" | grep -q "Invalid user credentials"; then
        log_warn "lagoonadmin user not found. Creating..."
        "$SCRIPT_DIR/create-lagoon-admin.sh"

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
        return 1
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
        echo ""
        echo "Configuration:"
        echo "  Set LAGOON_PRESET=single (default), multi-prod, or multi-nonprod"
        echo "  Or set individual variables (see scripts/common.sh)"
        exit 1
    fi

    # Activate virtual environment
    activate_venv || true

    # Check cluster connectivity
    if ! check_cluster_connectivity; then
        exit 1
    fi

    # Ensure port-forwards are running (only for clusters with core services)
    if [ -n "$KEYCLOAK_SVC" ] && [ -n "$API_SVC" ]; then
        ensure_port_forwards || exit 1
        # Use admin JWT token for full API access (including deploy targets)
        get_admin_jwt_token || exit 1
    else
        log_warn "No core services configured - skipping port-forwards and token"
    fi

    # Run pulumi command
    log_info "Running: pulumi $*"
    echo ""

    pulumi "$@"
}

main "$@"
