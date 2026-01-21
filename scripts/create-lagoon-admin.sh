#!/bin/bash
# Create the lagoonadmin user in Keycloak
#
# The Lagoon Helm chart does NOT automatically create the lagoonadmin user.
# This script creates the user with the password from the Helm secret
# and assigns the platform-owner role.
#
# Usage:
#   ./scripts/create-lagoon-admin.sh
#   LAGOON_PRESET=multi-prod ./scripts/create-lagoon-admin.sh
#
# See scripts/common.sh for configuration options.
#
# Additional configuration:
#   KEYCLOAK_URL - Keycloak base URL (default: http://localhost:8080)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"

echo "Creating lagoonadmin user in Keycloak..."
echo "  Context:   $KUBE_CONTEXT"
echo "  Namespace: $LAGOON_NAMESPACE"
echo "  Keycloak:  $KEYCLOAK_URL"

# Check for required secret
if [ -z "$KEYCLOAK_SECRET" ]; then
    log_error "KEYCLOAK_SECRET is required"
    exit 1
fi

# Get admin credentials
KEYCLOAK_ADMIN_PASSWORD=$(get_secret_value "$KEYCLOAK_SECRET" "KEYCLOAK_ADMIN_PASSWORD")
LAGOON_ADMIN_PASSWORD=$(get_secret_value "$KEYCLOAK_SECRET" "KEYCLOAK_LAGOON_ADMIN_PASSWORD")

if [ -z "$KEYCLOAK_ADMIN_PASSWORD" ]; then
    log_error "Could not get KEYCLOAK_ADMIN_PASSWORD from secret $KEYCLOAK_SECRET"
    exit 1
fi

if [ -z "$LAGOON_ADMIN_PASSWORD" ]; then
    log_error "Could not get KEYCLOAK_LAGOON_ADMIN_PASSWORD from secret $KEYCLOAK_SECRET"
    exit 1
fi

# Get admin token from master realm
log_info "Getting Keycloak admin token..."
ADMIN_TOKEN=$(curl -s "${KEYCLOAK_URL}/auth/realms/master/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" \
    -d "username=admin" \
    -d "password=${KEYCLOAK_ADMIN_PASSWORD}" | jq -r '.access_token')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    log_error "Failed to get Keycloak admin token"
    exit 1
fi

# Check if lagoonadmin user already exists
log_info "Checking if lagoonadmin user exists..."
EXISTING_USER=$(curl -s "${KEYCLOAK_URL}/auth/admin/realms/lagoon/users?username=lagoonadmin" \
    -H "Authorization: Bearer $ADMIN_TOKEN")

USER_COUNT=$(echo "$EXISTING_USER" | jq 'length')

if [ "$USER_COUNT" -gt 0 ]; then
    log_info "User lagoonadmin already exists."
    USER_ID=$(echo "$EXISTING_USER" | jq -r '.[0].id')
else
    # Create the user
    log_info "Creating lagoonadmin user..."
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
        log_error "Failed to create user (HTTP $HTTP_CODE)"
        echo "$CREATE_RESPONSE"
        exit 1
    fi

    # Get the user ID
    USER_ID=$(curl -s "${KEYCLOAK_URL}/auth/admin/realms/lagoon/users?username=lagoonadmin" \
        -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')

    log_info "User created with ID: $USER_ID"
fi

# Set the password
log_info "Setting password..."
curl -s -X PUT "${KEYCLOAK_URL}/auth/admin/realms/lagoon/users/${USER_ID}/reset-password" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"type\": \"password\",
        \"value\": \"${LAGOON_ADMIN_PASSWORD}\",
        \"temporary\": false
    }"

# Get the platform-owner role
log_info "Assigning platform-owner role..."
ROLE_INFO=$(curl -s "${KEYCLOAK_URL}/auth/admin/realms/lagoon/roles/platform-owner" \
    -H "Authorization: Bearer $ADMIN_TOKEN")

ROLE_ID=$(echo "$ROLE_INFO" | jq -r '.id')
ROLE_NAME=$(echo "$ROLE_INFO" | jq -r '.name')

if [ -z "$ROLE_ID" ] || [ "$ROLE_ID" = "null" ]; then
    log_warn "platform-owner role not found. User may have limited permissions."
else
    # Assign the role
    curl -s -X POST "${KEYCLOAK_URL}/auth/admin/realms/lagoon/users/${USER_ID}/role-mappings/realm" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d "[{\"id\": \"${ROLE_ID}\", \"name\": \"${ROLE_NAME}\"}]"
    log_info "Role assigned."
fi

echo ""
echo "lagoonadmin user is ready!"
echo "  Username: lagoonadmin"
echo "  Password: (from KEYCLOAK_LAGOON_ADMIN_PASSWORD secret)"
