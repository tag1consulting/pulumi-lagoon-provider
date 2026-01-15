#!/bin/bash
# Quickstart script for the Lagoon Pulumi provider example
#
# Usage:
#   ./scripts/quickstart.sh
#
# This script:
#   1. Verifies prerequisites
#   2. Sets up port-forwards
#   3. Gets credentials from secrets
#   4. Enables Direct Access Grants (if needed)
#   5. Gets an OAuth token
#   6. Runs pulumi preview
#
# After running this, you can run 'pulumi up' to deploy.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

CONTEXT="${KUBE_CONTEXT:-kind-lagoon-test}"
NAMESPACE="${LAGOON_NAMESPACE:-lagoon}"

echo "======================================"
echo "Lagoon Pulumi Provider - Quick Start"
echo "======================================"
echo ""

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl not found"
    exit 1
fi

if ! command -v pulumi &> /dev/null; then
    echo "Error: pulumi not found"
    exit 1
fi

if ! command -v curl &> /dev/null; then
    echo "Error: curl not found"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq not found"
    exit 1
fi

echo "  All prerequisites found."
echo ""

# Check cluster
echo "Checking cluster connectivity..."
if ! kubectl --context "$CONTEXT" cluster-info >/dev/null 2>&1; then
    echo "Error: Cannot connect to cluster '$CONTEXT'"
    echo "Make sure the Kind cluster is running."
    exit 1
fi
echo "  Connected to cluster."
echo ""

# Fix RabbitMQ password if needed
echo "Checking RabbitMQ configuration..."
BUILD_DEPLOY_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret lagoon-build-deploy -o jsonpath='{.data.RABBITMQ_PASSWORD}' 2>/dev/null | base64 -d)
BROKER_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret lagoon-core-broker -o jsonpath='{.data.RABBITMQ_PASSWORD}' 2>/dev/null | base64 -d)

if [ "$BUILD_DEPLOY_PASSWORD" != "$BROKER_PASSWORD" ]; then
    echo "  Fixing RabbitMQ password mismatch..."
    ./scripts/fix-rabbitmq-password.sh >/dev/null 2>&1
    echo "  Fixed."
else
    echo "  RabbitMQ configuration OK."
fi
echo ""

# Set up port-forwards
echo "Setting up port-forwards..."
pkill -f "port-forward.*lagoon-core" 2>/dev/null || true
sleep 1

kubectl --context "$CONTEXT" port-forward -n "$NAMESPACE" svc/lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 &
kubectl --context "$CONTEXT" port-forward -n "$NAMESPACE" svc/lagoon-core-api 7080:80 >/dev/null 2>&1 &
sleep 2

# Verify port-forwards
if ! curl -s --connect-timeout 2 "http://localhost:8080/" >/dev/null 2>&1; then
    echo "Error: Keycloak port-forward failed"
    exit 1
fi

if ! curl -s --connect-timeout 2 "http://localhost:7080/" >/dev/null 2>&1; then
    echo "Error: API port-forward failed"
    exit 1
fi

echo "  Port-forwards active."
echo ""

# Get credentials
echo "Getting credentials from cluster..."
LAGOON_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret lagoon-core-keycloak -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d)
export LAGOON_PASSWORD

echo "  Credentials retrieved."
echo ""

# Get token using simplified approach (no client secret needed for password grant)
echo "Getting OAuth token..."

get_token() {
    curl -s "http://localhost:8080/auth/realms/lagoon/protocol/openid-connect/token" \
        -d "grant_type=password" \
        -d "client_id=lagoon-ui" \
        -d "username=lagoonadmin" \
        -d "password=$LAGOON_PASSWORD" 2>&1
}

# First, try without client secret (simpler setup)
TOKEN_RESPONSE=$(get_token)

if echo "$TOKEN_RESPONSE" | jq -e '.access_token' > /dev/null 2>&1; then
    LAGOON_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
    export LAGOON_TOKEN
    export LAGOON_API_URL="http://localhost:7080/graphql"
    echo "  Token acquired."
else
    # Check if Direct Access Grants is the issue
    if echo "$TOKEN_RESPONSE" | grep -q "Client not allowed for direct access grants"; then
        echo "  Enabling Direct Access Grants..."

        KEYCLOAK_ADMIN_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret lagoon-core-keycloak -o jsonpath='{.data.KEYCLOAK_ADMIN_PASSWORD}' | base64 -d)

        # Get admin token
        ADMIN_TOKEN=$(curl -s "http://localhost:8080/auth/realms/master/protocol/openid-connect/token" \
            -d "grant_type=password" \
            -d "client_id=admin-cli" \
            -d "username=admin" \
            -d "password=$KEYCLOAK_ADMIN_PASSWORD" | jq -r '.access_token')

        # Enable Direct Access Grants
        curl -s -X PUT "http://localhost:8080/auth/admin/realms/lagoon/clients/$(curl -s "http://localhost:8080/auth/admin/realms/lagoon/clients?clientId=lagoon-ui" -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -H "Content-Type: application/json" \
            -d '{"directAccessGrantsEnabled": true}' >/dev/null

        # Retry token
        TOKEN_RESPONSE=$(get_token)
    fi

    # Check if user doesn't exist (invalid_grant with "Invalid user credentials")
    if echo "$TOKEN_RESPONSE" | grep -q "Invalid user credentials"; then
        echo "  lagoonadmin user not found. Creating..."
        ./scripts/create-lagoon-admin.sh
        echo "  Retrying token acquisition..."
        TOKEN_RESPONSE=$(get_token)
    fi

    # Final check
    if echo "$TOKEN_RESPONSE" | jq -e '.access_token' > /dev/null 2>&1; then
        LAGOON_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
        export LAGOON_TOKEN
        export LAGOON_API_URL="http://localhost:7080/graphql"
        echo "  Token acquired."
    else
        echo "Error: Failed to get token"
        echo "$TOKEN_RESPONSE"
        exit 1
    fi
fi
echo ""

# Test API
echo "Testing API connection..."
API_RESPONSE=$(curl -s -X POST "$LAGOON_API_URL" \
    -H "Authorization: Bearer $LAGOON_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"query": "{ allProjects { id name } }"}')

if echo "$API_RESPONSE" | jq -e '.data' > /dev/null 2>&1; then
    PROJECT_COUNT=$(echo "$API_RESPONSE" | jq '.data.allProjects | length')
    echo "  API working. Found $PROJECT_COUNT existing project(s)."
else
    echo "Error: API test failed"
    echo "$API_RESPONSE"
    exit 1
fi
echo ""

# Check for deploy target
echo "Checking deploy targets..."
DEPLOY_TARGETS=$(curl -s -X POST "$LAGOON_API_URL" \
    -H "Authorization: Bearer $LAGOON_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"query": "{ allKubernetes { id name } }"}')

TARGET_COUNT=$(echo "$DEPLOY_TARGETS" | jq '.data.allKubernetes | length')

if [ "$TARGET_COUNT" -eq 0 ]; then
    echo "  No deploy targets found. Creating 'local-kind'..."
    ./scripts/add-deploy-target.sh local-kind https://kubernetes.default.svc >/dev/null 2>&1
    echo "  Deploy target created."
else
    echo "  Found $TARGET_COUNT deploy target(s)."
fi
echo ""

# Initialize Pulumi stack if needed
echo "Checking Pulumi stack..."
if ! pulumi stack ls 2>/dev/null | grep -q "test"; then
    echo "  Initializing 'test' stack..."
    pulumi stack init test 2>/dev/null || true
fi
pulumi stack select test 2>/dev/null || true

# Set deploytarget ID
DEPLOY_TARGET_ID=$(echo "$DEPLOY_TARGETS" | jq -r '.data.allKubernetes[0].id // 1')
pulumi config set deploytargetId "$DEPLOY_TARGET_ID" 2>/dev/null || true
echo "  Stack configured with deploytargetId=$DEPLOY_TARGET_ID"
echo ""

# Run preview
echo "======================================"
echo "Running pulumi preview..."
echo "======================================"
echo ""

pulumi preview

echo ""
echo "======================================"
echo "Setup complete!"
echo "======================================"
echo ""
echo "To deploy the resources, run:"
echo "  pulumi up"
echo ""
echo "Or use the wrapper script (automatically refreshes token):"
echo "  ./scripts/run-pulumi.sh up"
echo ""
echo "Environment variables are set for this session:"
echo "  LAGOON_API_URL=$LAGOON_API_URL"
echo "  LAGOON_TOKEN=<set>"
echo ""
echo "IMPORTANT: Tokens expire in 5 minutes!"
echo "If your token expires, either:"
echo "  1. Run: source ./scripts/quickstart.sh"
echo "  2. Use: ./scripts/run-pulumi.sh <command> (recommended)"
