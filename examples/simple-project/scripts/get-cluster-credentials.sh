#!/bin/bash
# Extract all Lagoon credentials from the test cluster
#
# Usage:
#   ./scripts/get-cluster-credentials.sh
#   # or source it to set environment variables
#   source ./scripts/get-cluster-credentials.sh
#
# This retrieves credentials from Kubernetes secrets and optionally
# exports them as environment variables.

set -e

CONTEXT="${KUBE_CONTEXT:-kind-lagoon-test}"
NAMESPACE="${LAGOON_NAMESPACE:-lagoon}"

echo "Extracting credentials from cluster: $CONTEXT"
echo "Namespace: $NAMESPACE"
echo ""

# Helper function to decode base64
decode_secret() {
    kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret "$1" -o jsonpath="{.data.$2}" 2>/dev/null | base64 -d
}

# Get Keycloak credentials
KEYCLOAK_ADMIN_PASSWORD=$(decode_secret lagoon-core-keycloak KEYCLOAK_ADMIN_PASSWORD)
LAGOON_ADMIN_PASSWORD=$(decode_secret lagoon-core-keycloak KEYCLOAK_LAGOON_ADMIN_PASSWORD)
LAGOON_UI_CLIENT_SECRET=$(decode_secret lagoon-core-keycloak KEYCLOAK_LAGOON_UI_OIDC_CLIENT_SECRET)

# Get RabbitMQ credentials
RABBITMQ_PASSWORD=$(decode_secret lagoon-core-broker RABBITMQ_PASSWORD)

echo "=== Keycloak Credentials ==="
echo "KEYCLOAK_ADMIN_PASSWORD: $KEYCLOAK_ADMIN_PASSWORD"
echo ""
echo "=== Lagoon Admin Credentials ==="
echo "Username: lagoonadmin"
echo "LAGOON_PASSWORD: $LAGOON_ADMIN_PASSWORD"
echo ""
echo "=== Lagoon UI Client ==="
echo "CLIENT_ID: lagoon-ui"
echo "CLIENT_SECRET: $LAGOON_UI_CLIENT_SECRET"
echo ""
echo "=== RabbitMQ Credentials ==="
echo "Username: lagoon"
echo "RABBITMQ_PASSWORD: $RABBITMQ_PASSWORD"
echo ""

# If sourced, export the variables
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
    export KEYCLOAK_ADMIN_PASSWORD
    export LAGOON_PASSWORD="$LAGOON_ADMIN_PASSWORD"
    export CLIENT_SECRET="$LAGOON_UI_CLIENT_SECRET"
    export RABBITMQ_PASSWORD
    echo "Environment variables exported: KEYCLOAK_ADMIN_PASSWORD, LAGOON_PASSWORD, CLIENT_SECRET, RABBITMQ_PASSWORD"
else
    echo "To export these as environment variables, source this script:"
    echo "  source ./scripts/get-cluster-credentials.sh"
fi
