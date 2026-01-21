#!/bin/bash
# Create the lagoonadmin user in Keycloak
#
# The Lagoon Helm chart does NOT automatically create the lagoonadmin user.
# This script creates the user with the password from the Helm secret
# and assigns the platform-owner role.
#
# Usage:
#   ./scripts/create-lagoon-admin.sh
#
# Prerequisites:
#   - Port-forward to Keycloak on localhost:8080
#   - Or set KEYCLOAK_URL environment variable

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

CONTEXT="${KUBE_CONTEXT:-kind-lagoon-prod}"
NAMESPACE="${LAGOON_NAMESPACE:-lagoon}"
KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"

echo "Creating lagoonadmin user in Keycloak..."

# Get admin credentials
KEYCLOAK_ADMIN_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret lagoon-core-keycloak -o jsonpath='{.data.KEYCLOAK_ADMIN_PASSWORD}' | base64 -d)
LAGOON_ADMIN_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret lagoon-core-keycloak -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d)

# Get admin token from master realm
echo "  Getting Keycloak admin token..."
ADMIN_TOKEN=$(curl -s "${KEYCLOAK_URL}/auth/realms/master/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" \
    -d "username=admin" \
    -d "password=${KEYCLOAK_ADMIN_PASSWORD}" | jq -r '.access_token')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo "Error: Failed to get Keycloak admin token"
    exit 1
fi

# Check if lagoonadmin user already exists
echo "  Checking if lagoonadmin user exists..."
EXISTING_USER=$(curl -s "${KEYCLOAK_URL}/auth/admin/realms/lagoon/users?username=lagoonadmin" \
    -H "Authorization: Bearer $ADMIN_TOKEN")

USER_COUNT=$(echo "$EXISTING_USER" | jq 'length')

if [ "$USER_COUNT" -gt 0 ]; then
    echo "  User lagoonadmin already exists."
    USER_ID=$(echo "$EXISTING_USER" | jq -r '.[0].id')
else
    # Create the user
    echo "  Creating lagoonadmin user..."
    CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${KEYCLOAK_URL}/auth/admin/realms/lagoon/users" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "lagoonadmin",
            "email": "lagoonadmin@lagoon.local",
            "enabled": true,
            "emailVerified": true,
            "firstName": "Lagoon",
            "lastName": "Admin"
        }')

    HTTP_CODE=$(echo "$CREATE_RESPONSE" | tail -n 1)

    if [ "$HTTP_CODE" != "201" ] && [ "$HTTP_CODE" != "409" ]; then
        echo "Error: Failed to create user (HTTP $HTTP_CODE)"
        echo "$CREATE_RESPONSE"
        exit 1
    fi

    # Get the user ID
    USER_ID=$(curl -s "${KEYCLOAK_URL}/auth/admin/realms/lagoon/users?username=lagoonadmin" \
        -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')

    echo "  User created with ID: $USER_ID"
fi

# Set the password
echo "  Setting password..."
curl -s -X PUT "${KEYCLOAK_URL}/auth/admin/realms/lagoon/users/${USER_ID}/reset-password" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"type\": \"password\",
        \"value\": \"${LAGOON_ADMIN_PASSWORD}\",
        \"temporary\": false
    }"

# Get the platform-owner role
echo "  Assigning platform-owner role..."
ROLE_INFO=$(curl -s "${KEYCLOAK_URL}/auth/admin/realms/lagoon/roles/platform-owner" \
    -H "Authorization: Bearer $ADMIN_TOKEN")

ROLE_ID=$(echo "$ROLE_INFO" | jq -r '.id')
ROLE_NAME=$(echo "$ROLE_INFO" | jq -r '.name')

if [ -z "$ROLE_ID" ] || [ "$ROLE_ID" = "null" ]; then
    echo "Warning: platform-owner role not found. User may have limited permissions."
else
    # Assign the role
    curl -s -X POST "${KEYCLOAK_URL}/auth/admin/realms/lagoon/users/${USER_ID}/role-mappings/realm" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d "[{\"id\": \"${ROLE_ID}\", \"name\": \"${ROLE_NAME}\"}]"
    echo "  Role assigned."
fi

echo ""
echo "lagoonadmin user is ready!"
echo "  Username: lagoonadmin"
echo "  Password: (from KEYCLOAK_LAGOON_ADMIN_PASSWORD secret)"
