#!/bin/bash
# Enable Direct Access Grants on the lagoon-ui Keycloak client
#
# This is required to obtain tokens via password grant (username/password).
# By default, the lagoon-ui client only supports browser-based OAuth flow.
#
# Usage:
#   ./scripts/enable-direct-access-grants.sh
#
# Prerequisites:
#   - curl and jq installed
#   - Keycloak accessible (via port-forward or direct)
#   - Keycloak admin credentials
#
# Configuration via environment variables:
#   KEYCLOAK_URL           - Keycloak base URL (default: http://localhost:8080)
#   KEYCLOAK_ADMIN_USER    - Keycloak admin username (default: admin)
#   KEYCLOAK_ADMIN_PASSWORD - Keycloak admin password (auto-fetched from cluster if not set)
#   KUBE_CONTEXT           - Kubernetes context (default: kind-lagoon-test)

set -e

# Configuration with defaults (prefer port-forward URL)
KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"
KEYCLOAK_ADMIN_USER="${KEYCLOAK_ADMIN_USER:-admin}"
KUBE_CONTEXT="${KUBE_CONTEXT:-kind-lagoon-test}"

# Try to get admin password from cluster if not provided
if [ -z "$KEYCLOAK_ADMIN_PASSWORD" ]; then
    if command -v kubectl &> /dev/null; then
        echo "Fetching Keycloak admin password from cluster..."
        KEYCLOAK_ADMIN_PASSWORD=$(kubectl --context "$KUBE_CONTEXT" -n lagoon get secret lagoon-core-keycloak \
            -o jsonpath='{.data.KEYCLOAK_ADMIN_PASSWORD}' 2>/dev/null | base64 -d 2>/dev/null)
    fi
fi

if [ -z "$KEYCLOAK_ADMIN_PASSWORD" ]; then
    echo "Error: KEYCLOAK_ADMIN_PASSWORD environment variable is required" >&2
    echo "Set it manually or ensure kubectl can access the cluster:" >&2
    echo "  kubectl --context $KUBE_CONTEXT -n lagoon get secret lagoon-core-keycloak \\" >&2
    echo "    -o jsonpath='{.data.KEYCLOAK_ADMIN_PASSWORD}' | base64 -d" >&2
    exit 1
fi

echo "Getting Keycloak admin token from $KEYCLOAK_URL..."

# Get admin token from master realm
ADMIN_TOKEN=$(curl -s -k -X POST "${KEYCLOAK_URL}/auth/realms/master/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" \
    -d "username=${KEYCLOAK_ADMIN_USER}" \
    -d "password=${KEYCLOAK_ADMIN_PASSWORD}" | jq -r '.access_token')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo "Error: Failed to get admin token" >&2
    echo "Make sure Keycloak is accessible at $KEYCLOAK_URL" >&2
    echo "If using port-forward, run: ./scripts/setup-port-forwards.sh" >&2
    exit 1
fi

echo "Finding lagoon-ui client..."

# Get lagoon-ui client ID
CLIENT_UUID=$(curl -s -k -H "Authorization: Bearer $ADMIN_TOKEN" \
    "${KEYCLOAK_URL}/auth/admin/realms/lagoon/clients?clientId=lagoon-ui" | jq -r '.[0].id')

if [ -z "$CLIENT_UUID" ] || [ "$CLIENT_UUID" = "null" ]; then
    echo "Error: lagoon-ui client not found" >&2
    exit 1
fi

echo "Enabling Direct Access Grants on lagoon-ui client..."

# Get current client config
CLIENT_CONFIG=$(curl -s -k -H "Authorization: Bearer $ADMIN_TOKEN" \
    "${KEYCLOAK_URL}/auth/admin/realms/lagoon/clients/${CLIENT_UUID}")

# Check if already enabled
ALREADY_ENABLED=$(echo "$CLIENT_CONFIG" | jq -r '.directAccessGrantsEnabled')
if [ "$ALREADY_ENABLED" = "true" ]; then
    echo "Direct Access Grants is already enabled."
    exit 0
fi

# Update with direct access grants enabled
UPDATED_CONFIG=$(echo "$CLIENT_CONFIG" | jq '.directAccessGrantsEnabled = true')

# Apply the update
HTTP_CODE=$(curl -s -k -o /dev/null -w "%{http_code}" \
    -X PUT "${KEYCLOAK_URL}/auth/admin/realms/lagoon/clients/${CLIENT_UUID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$UPDATED_CONFIG")

if [ "$HTTP_CODE" = "204" ]; then
    echo "Success! Direct Access Grants enabled on lagoon-ui client."
    echo ""
    echo "You can now obtain tokens using:"
    echo "  source ./scripts/get-lagoon-token.sh"
else
    echo "Error: Failed to update client (HTTP $HTTP_CODE)" >&2
    exit 1
fi
