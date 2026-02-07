#!/bin/bash
# Get an OAuth token from Keycloak for Lagoon API access
#
# Usage:
#   source scripts/get-token.sh
#   # or
#   export LAGOON_TOKEN=$(./scripts/get-token.sh)
#
# See scripts/common.sh for configuration options.
#
# Additional configuration:
#   KEYCLOAK_URL      - Keycloak base URL (default: http://localhost:8080)
#   LAGOON_USERNAME   - Lagoon admin username (default: lagoonadmin)
#   LAGOON_PASSWORD   - Lagoon admin password (fetched from secret if not set)
#   CLIENT_ID         - Keycloak client ID (default: lagoon-ui)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

# Configuration with defaults
KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"
KEYCLOAK_REALM="${KEYCLOAK_REALM:-lagoon}"
LAGOON_USERNAME="${LAGOON_USERNAME:-lagoonadmin}"
CLIENT_ID="${CLIENT_ID:-lagoon-ui}"

# Try to get password from cluster if not provided
if [ -z "$LAGOON_PASSWORD" ]; then
    if [ -n "$KEYCLOAK_SECRET" ]; then
        LAGOON_PASSWORD=$(get_secret_value "$KEYCLOAK_SECRET" "KEYCLOAK_LAGOON_ADMIN_PASSWORD")
    fi
fi

if [ -z "$LAGOON_PASSWORD" ]; then
    echo "Error: LAGOON_PASSWORD environment variable is required" >&2
    echo "Set it manually or ensure kubectl can access the cluster." >&2
    exit 1
fi

# Build token request
TOKEN_DATA="grant_type=password&client_id=${CLIENT_ID}&username=${LAGOON_USERNAME}&password=${LAGOON_PASSWORD}"

# Add client secret if provided
if [ -n "$CLIENT_SECRET" ]; then
    TOKEN_DATA="${TOKEN_DATA}&client_secret=${CLIENT_SECRET}"
fi

# Get token
TOKEN_RESPONSE=$(curl -s -k -X POST "${KEYCLOAK_URL}/auth/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "$TOKEN_DATA" 2>&1)

# Check for errors
if echo "$TOKEN_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    ERROR=$(echo "$TOKEN_RESPONSE" | jq -r '.error_description // .error')
    echo "Error getting token: $ERROR" >&2

    if echo "$ERROR" | grep -q "Client not allowed for direct access grants"; then
        echo "" >&2
        echo "The lagoon-ui client needs 'Direct Access Grants' enabled." >&2
        echo "Run: ./scripts/enable-direct-access-grants.sh" >&2
    fi
    exit 1
fi

# Extract token
TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "Error: Failed to extract token from response" >&2
    echo "$TOKEN_RESPONSE" >&2
    exit 1
fi

# Determine API URL based on KEYCLOAK_URL pattern
if [[ "$KEYCLOAK_URL" == *"localhost"* ]]; then
    DEFAULT_API_URL="http://localhost:7080/graphql"
else
    DEFAULT_API_URL=$(echo "$KEYCLOAK_URL" | sed 's/keycloak/api/' | sed 's/:8080/:443/')/graphql
fi

# If sourced, export the variables; if executed, just print the token
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
    export LAGOON_TOKEN="$TOKEN"
    export LAGOON_API_URL="${LAGOON_API_URL:-$DEFAULT_API_URL}"
    echo "LAGOON_TOKEN exported (expires in 5 minutes)"
    echo "LAGOON_API_URL=$LAGOON_API_URL"
else
    echo "$TOKEN"
fi
