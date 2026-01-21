#!/bin/bash
# Ensure a deploy target exists in Lagoon
#
# This script:
# 1. Gets a fresh OAuth token
# 2. Checks if any deploy targets exist
# 3. Creates one if none exist
# 4. Outputs the deploy target ID
#
# Usage:
#   ./scripts/ensure-deploy-target.sh
#   DEPLOY_TARGET_ID=$(./scripts/ensure-deploy-target.sh)
#   LAGOON_PRESET=multi-prod ./scripts/ensure-deploy-target.sh
#
# See scripts/common.sh for configuration options.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"
LAGOON_API_URL="${LAGOON_API_URL:-http://localhost:7080/graphql}"

# Enable Direct Access Grants if needed
enable_direct_access_grants() {
    local admin_password
    admin_password=$(get_secret_value "$KEYCLOAK_SECRET" "KEYCLOAK_ADMIN_PASSWORD")

    local admin_token
    admin_token=$(curl -s "${KEYCLOAK_URL}/auth/realms/master/protocol/openid-connect/token" \
        -d "grant_type=password" \
        -d "client_id=admin-cli" \
        -d "username=admin" \
        -d "password=$admin_password" | jq -r '.access_token')

    local client_id
    client_id=$(curl -s "${KEYCLOAK_URL}/auth/admin/realms/lagoon/clients?clientId=lagoon-ui" \
        -H "Authorization: Bearer $admin_token" | jq -r '.[0].id')

    curl -s -X PUT "${KEYCLOAK_URL}/auth/admin/realms/lagoon/clients/$client_id" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d '{"directAccessGrantsEnabled": true}' >/dev/null
}

# Get fresh OAuth token
get_fresh_token() {
    local password
    password=$(get_secret_value "$KEYCLOAK_SECRET" "KEYCLOAK_LAGOON_ADMIN_PASSWORD")

    local response
    response=$(curl -s "${KEYCLOAK_URL}/auth/realms/lagoon/protocol/openid-connect/token" \
        -d "grant_type=password" \
        -d "client_id=lagoon-ui" \
        -d "username=lagoonadmin" \
        -d "password=$password" 2>&1)

    local token
    token=$(echo "$response" | jq -r '.access_token // empty')

    # Check if Direct Access Grants need to be enabled
    if [ -z "$token" ] && echo "$response" | grep -q "Client not allowed for direct access grants"; then
        echo "Enabling Direct Access Grants..." >&2
        enable_direct_access_grants

        # Retry
        response=$(curl -s "${KEYCLOAK_URL}/auth/realms/lagoon/protocol/openid-connect/token" \
            -d "grant_type=password" \
            -d "client_id=lagoon-ui" \
            -d "username=lagoonadmin" \
            -d "password=$password" 2>&1)
        token=$(echo "$response" | jq -r '.access_token // empty')
    fi

    if [ -z "$token" ]; then
        echo "Error: Failed to get OAuth token" >&2
        echo "$response" >&2
        exit 1
    fi

    echo "$token"
}

echo "Checking deploy targets..." >&2

# Get fresh token
TOKEN=$(get_fresh_token)

# Check existing deploy targets
DEPLOY_TARGETS=$(curl -s -X POST "$LAGOON_API_URL" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"query": "{ allKubernetes { id name } }"}')

TARGET_COUNT=$(echo "$DEPLOY_TARGETS" | jq '.data.allKubernetes | length')

if [ "$TARGET_COUNT" -gt 0 ]; then
    # Return first deploy target ID
    DEPLOY_TARGET_ID=$(echo "$DEPLOY_TARGETS" | jq -r '.data.allKubernetes[0].id')
    echo "Found existing deploy target (ID: $DEPLOY_TARGET_ID)" >&2
    echo "$DEPLOY_TARGET_ID"
    exit 0
fi

echo "No deploy targets found. Creating 'local-kind'..." >&2

# Token may have been used, get a fresh one for the mutation
TOKEN=$(get_fresh_token)

# Create deploy target
CREATE_RESPONSE=$(curl -s -X POST "$LAGOON_API_URL" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "mutation AddKubernetes($input: AddKubernetesInput!) { addKubernetes(input: $input) { id name } }",
        "variables": {
            "input": {
                "name": "local-kind",
                "consoleUrl": "https://kubernetes.default.svc",
                "token": "test-token",
                "routerPattern": "${environment}.${project}.lagoon.local",
                "cloudProvider": "kind",
                "cloudRegion": "local"
            }
        }
    }')

DEPLOY_TARGET_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.addKubernetes.id // empty')

if [ -z "$DEPLOY_TARGET_ID" ]; then
    echo "Error: Failed to create deploy target" >&2
    echo "$CREATE_RESPONSE" >&2
    exit 1
fi

echo "Created deploy target 'local-kind' (ID: $DEPLOY_TARGET_ID)" >&2
echo "$DEPLOY_TARGET_ID"
