#!/bin/bash
# Set up kubectl port-forwards for accessing Lagoon services
#
# Usage:
#   ./scripts/setup-port-forwards.sh
#   LAGOON_PRESET=multi-prod ./scripts/setup-port-forwards.sh
#
# This script sets up port-forwards for:
#   - Keycloak: localhost:8080 -> keycloak-service:8080
#   - API: localhost:7080 -> api-service:80
#
# See scripts/common.sh for configuration options.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

echo "Setting up port-forwards for Lagoon services..."
echo "  Context:   $KUBE_CONTEXT"
echo "  Namespace: $LAGOON_NAMESPACE"
echo ""

# Check for required services
if [ -z "$KEYCLOAK_SVC" ] || [ -z "$API_SVC" ]; then
    log_error "Keycloak and API service names are required"
    log_error "This preset may not have core services (e.g., multi-nonprod)"
    exit 1
fi

# Kill any existing port-forwards
pkill -f "port-forward.*$KEYCLOAK_SVC" 2>/dev/null || true
pkill -f "port-forward.*$API_SVC" 2>/dev/null || true
sleep 1

# Start Keycloak port-forward
echo "Starting Keycloak port-forward (localhost:8080)..."
kubectl --context "$KUBE_CONTEXT" port-forward -n "$LAGOON_NAMESPACE" "svc/$KEYCLOAK_SVC" 8080:8080 >/dev/null 2>&1 &
KEYCLOAK_PID=$!

# Start API port-forward
echo "Starting API port-forward (localhost:7080)..."
kubectl --context "$KUBE_CONTEXT" port-forward -n "$LAGOON_NAMESPACE" "svc/$API_SVC" 7080:80 >/dev/null 2>&1 &
API_PID=$!

sleep 2

# Verify port-forwards
if curl -s --connect-timeout 2 "http://localhost:8080/" >/dev/null 2>&1; then
    echo -e "  Keycloak: ${GREEN}OK${NC} (PID: $KEYCLOAK_PID)"
else
    echo -e "  Keycloak: ${RED}FAILED${NC}"
fi

if curl -s --connect-timeout 2 "http://localhost:7080/" >/dev/null 2>&1; then
    echo -e "  API: ${GREEN}OK${NC} (PID: $API_PID)"
else
    echo -e "  API: ${RED}FAILED${NC}"
fi

echo ""
echo "Port-forwards are running in the background."
echo "To stop them: pkill -f 'port-forward.*lagoon'"
echo ""
echo "Service URLs:"
echo "  Keycloak: http://localhost:8080/auth"
echo "  API:      http://localhost:7080/graphql"
echo ""
echo "Set these environment variables for the provider:"
echo "  export LAGOON_API_URL=http://localhost:7080/graphql"
echo "  export KEYCLOAK_URL=http://localhost:8080"
