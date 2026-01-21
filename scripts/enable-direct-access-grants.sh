#!/bin/bash
# Enable Direct Access Grants on the lagoon-ui Keycloak client
#
# This is required for password-based OAuth authentication (resource owner password grant).
# Without this, CLI tools and scripts cannot get tokens using username/password.
#
# Usage:
#   ./scripts/enable-direct-access-grants.sh
#   LAGOON_PRESET=multi-prod ./scripts/enable-direct-access-grants.sh
#
# See scripts/common.sh for configuration options.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"

echo "Enabling Direct Access Grants on lagoon-ui client..."
echo "  Context:   $KUBE_CONTEXT"
echo "  Namespace: $LAGOON_NAMESPACE"
echo "  Keycloak:  $KEYCLOAK_URL"

# Check for required secret
if [ -z "$KEYCLOAK_SECRET" ]; then
    log_error "KEYCLOAK_SECRET is required"
    exit 1
fi

# Get admin password
ADMIN_PASSWORD=$(get_secret_value "$KEYCLOAK_SECRET" "KEYCLOAK_ADMIN_PASSWORD")

if [ -z "$ADMIN_PASSWORD" ]; then
    log_error "Could not get KEYCLOAK_ADMIN_PASSWORD from secret $KEYCLOAK_SECRET"
    exit 1
fi

# Get admin token
log_info "Getting Keycloak admin token..."
ADMIN_TOKEN=$(curl -s "${KEYCLOAK_URL}/auth/realms/master/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" \
    -d "username=admin" \
    -d "password=${ADMIN_PASSWORD}" | jq -r '.access_token')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    log_error "Failed to get admin token"
    exit 1
fi

# Get the lagoon-ui client ID
log_info "Finding lagoon-ui client..."
CLIENT_ID=$(curl -s "${KEYCLOAK_URL}/auth/admin/realms/lagoon/clients?clientId=lagoon-ui" \
    -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')

if [ -z "$CLIENT_ID" ] || [ "$CLIENT_ID" = "null" ]; then
    log_error "Could not find lagoon-ui client"
    exit 1
fi

log_info "Client ID: $CLIENT_ID"

# Enable Direct Access Grants
log_info "Enabling Direct Access Grants..."
curl -s -X PUT "${KEYCLOAK_URL}/auth/admin/realms/lagoon/clients/$CLIENT_ID" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"directAccessGrantsEnabled": true}'

log_info "Direct Access Grants enabled for lagoon-ui client"
echo ""
echo "You can now use password-based authentication:"
echo "  source ./scripts/get-token.sh"
