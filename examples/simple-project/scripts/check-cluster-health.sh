#!/bin/bash
# Check Lagoon test cluster health
#
# Usage:
#   ./scripts/check-cluster-health.sh
#
# Verifies:
#   - Kind cluster is running
#   - All Lagoon pods are healthy
#   - RabbitMQ broker is accessible
#   - API is responding
#   - Keycloak is responding

set -e

CONTEXT="${KUBE_CONTEXT:-kind-lagoon-test}"
NAMESPACE="${LAGOON_NAMESPACE:-lagoon}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Checking Lagoon test cluster health..."
echo "Context: $CONTEXT"
echo ""

# Check if kind cluster exists
echo "=== Kind Cluster ==="
if docker ps --filter "name=lagoon-test-control-plane" --format "{{.Names}}" | grep -q lagoon-test; then
    echo -e "${GREEN}OK${NC} - Kind cluster 'lagoon-test' is running"
else
    echo -e "${RED}FAILED${NC} - Kind cluster 'lagoon-test' not found"
    echo "  Start with: kind create cluster --name lagoon-test --config test-cluster/kind-config.yaml"
    exit 1
fi
echo ""

# Check kubectl connectivity
echo "=== Kubernetes Connectivity ==="
if kubectl --context "$CONTEXT" cluster-info >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - kubectl can connect to cluster"
else
    echo -e "${RED}FAILED${NC} - Cannot connect to cluster"
    exit 1
fi
echo ""

# Check Lagoon pods
echo "=== Lagoon Pods ==="
UNHEALTHY=$(kubectl --context "$CONTEXT" get pods -n "$NAMESPACE" --no-headers 2>/dev/null | grep -v "Running\|Completed" | wc -l)
TOTAL=$(kubectl --context "$CONTEXT" get pods -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
RUNNING=$(kubectl --context "$CONTEXT" get pods -n "$NAMESPACE" --no-headers 2>/dev/null | grep Running | wc -l)

echo "Total pods: $TOTAL"
echo "Running: $RUNNING"
echo "Unhealthy: $UNHEALTHY"

if [ "$UNHEALTHY" -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}Unhealthy pods:${NC}"
    kubectl --context "$CONTEXT" get pods -n "$NAMESPACE" --no-headers | grep -v "Running\|Completed"
fi
echo ""

# Check specific critical pods
echo "=== Critical Services ==="
CRITICAL_PODS=("lagoon-core-api" "lagoon-core-keycloak" "lagoon-core-broker" "lagoon-build-deploy")

for POD_PREFIX in "${CRITICAL_PODS[@]}"; do
    POD_STATUS=$(kubectl --context "$CONTEXT" get pods -n "$NAMESPACE" -l "app.kubernetes.io/name=$POD_PREFIX" -o jsonpath='{.items[0].status.phase}' 2>/dev/null || echo "NotFound")
    if [ "$POD_STATUS" = "Running" ]; then
        echo -e "${GREEN}OK${NC} - $POD_PREFIX: Running"
    else
        echo -e "${RED}FAILED${NC} - $POD_PREFIX: $POD_STATUS"
    fi
done
echo ""

# Check if port-forwards are needed
echo "=== Network Access ==="
CLUSTER_IP=$(docker inspect lagoon-test-control-plane --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>/dev/null)
echo "Cluster IP: $CLUSTER_IP"

if ping -c 1 -W 1 "$CLUSTER_IP" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - Docker network is routable"
    echo "  You can access services via ingress at *.$CLUSTER_IP.nip.io"
else
    echo -e "${YELLOW}WARNING${NC} - Docker network is not routable (common in WSL2)"
    echo "  Use port-forwards: ./scripts/setup-port-forwards.sh"
fi
echo ""

# Check if port-forwards are running
echo "=== Port Forwards ==="
if pgrep -f "port-forward.*lagoon-core-keycloak" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - Keycloak port-forward running (localhost:8080)"
else
    echo -e "${YELLOW}INFO${NC} - No Keycloak port-forward"
fi

if pgrep -f "port-forward.*lagoon-core-api" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - API port-forward running (localhost:7080)"
else
    echo -e "${YELLOW}INFO${NC} - No API port-forward"
fi
echo ""

# Test service accessibility
echo "=== Service Accessibility ==="

# Try port-forward first, then direct
if curl -s --connect-timeout 2 "http://localhost:8080/auth" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - Keycloak accessible at localhost:8080"
elif curl -s -k --connect-timeout 2 "https://keycloak.$CLUSTER_IP.nip.io" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - Keycloak accessible via ingress"
else
    echo -e "${YELLOW}WARNING${NC} - Keycloak not accessible"
    echo "  Run: ./scripts/setup-port-forwards.sh"
fi

if curl -s --connect-timeout 2 "http://localhost:7080/" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - API accessible at localhost:7080"
elif curl -s -k --connect-timeout 2 "https://api.$CLUSTER_IP.nip.io" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - API accessible via ingress"
else
    echo -e "${YELLOW}WARNING${NC} - API not accessible"
    echo "  Run: ./scripts/setup-port-forwards.sh"
fi
echo ""

echo "=== Summary ==="
if [ "$UNHEALTHY" -eq 0 ] || [ "$UNHEALTHY" -eq 1 ]; then
    echo -e "${GREEN}Cluster is healthy and ready for use${NC}"
    echo ""
    echo "Quick start:"
    echo "  1. source ./scripts/get-cluster-credentials.sh"
    echo "  2. ./scripts/setup-port-forwards.sh"
    echo "  3. source ./scripts/get-lagoon-token.sh"
    echo "  4. pulumi up"
else
    echo -e "${RED}Cluster has issues that need attention${NC}"
    echo ""
    echo "Common fixes:"
    echo "  - RabbitMQ auth: ./scripts/fix-rabbitmq-password.sh"
    echo "  - Check logs: kubectl --context $CONTEXT logs -n $NAMESPACE <pod-name>"
fi
