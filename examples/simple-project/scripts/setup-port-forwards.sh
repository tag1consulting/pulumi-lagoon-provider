#!/bin/bash
# Set up kubectl port-forwards for accessing Lagoon services
#
# Usage:
#   ./scripts/setup-port-forwards.sh
#
# This script sets up port-forwards for:
#   - Keycloak: localhost:8080 -> lagoon-core-keycloak:8080
#   - API: localhost:7080 -> lagoon-core-api:80
#
# Note: This is necessary when Docker networks aren't directly routable
# (common in WSL2/Docker Desktop environments)

set -e

CONTEXT="${KUBE_CONTEXT:-kind-lagoon-test}"
NAMESPACE="${LAGOON_NAMESPACE:-lagoon}"

echo "Setting up port-forwards for Lagoon services..."
echo "Using context: $CONTEXT"
echo ""

# Kill any existing port-forwards
pkill -f "port-forward.*lagoon-core-keycloak" 2>/dev/null || true
pkill -f "port-forward.*lagoon-core-api" 2>/dev/null || true
sleep 1

# Start Keycloak port-forward
echo "Starting Keycloak port-forward (localhost:8080)..."
kubectl --context "$CONTEXT" port-forward -n "$NAMESPACE" svc/lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 &
KEYCLOAK_PID=$!

# Start API port-forward
echo "Starting API port-forward (localhost:7080)..."
kubectl --context "$CONTEXT" port-forward -n "$NAMESPACE" svc/lagoon-core-api 7080:80 >/dev/null 2>&1 &
API_PID=$!

sleep 2

# Verify port-forwards
if curl -s "http://localhost:8080/" >/dev/null 2>&1; then
    echo "  Keycloak: OK (PID: $KEYCLOAK_PID)"
else
    echo "  Keycloak: FAILED"
fi

if curl -s "http://localhost:7080/" >/dev/null 2>&1; then
    echo "  API: OK (PID: $API_PID)"
else
    echo "  API: FAILED"
fi

echo ""
echo "Port-forwards are running in the background."
echo "To stop them: pkill -f 'port-forward.*lagoon-core'"
echo ""
echo "Service URLs:"
echo "  Keycloak: http://localhost:8080/auth"
echo "  API:      http://localhost:7080/graphql"
echo ""
echo "Set these environment variables for the provider:"
echo "  export LAGOON_API_URL=http://localhost:7080/graphql"
echo "  export KEYCLOAK_URL=http://localhost:8080"
