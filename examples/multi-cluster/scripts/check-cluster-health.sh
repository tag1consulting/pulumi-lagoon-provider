#!/bin/bash
# Check Multi-cluster Lagoon health
#
# Usage:
#   ./scripts/check-cluster-health.sh
#
# Verifies:
#   - Both Kind clusters are running
#   - All Lagoon pods are healthy on both clusters
#   - Services are accessible

set -e

NAMESPACE="${LAGOON_NAMESPACE:-lagoon}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

check_cluster() {
    local cluster_name="$1"
    local context="kind-lagoon-$cluster_name"

    echo -e "\n${BLUE}========== $cluster_name Cluster ==========${NC}"

    # Check if kind cluster exists
    echo "=== Kind Cluster ==="
    if docker ps --filter "name=lagoon-$cluster_name-control-plane" --format "{{.Names}}" | grep -q "lagoon-$cluster_name"; then
        echo -e "${GREEN}OK${NC} - Kind cluster 'lagoon-$cluster_name' is running"
    else
        echo -e "${RED}FAILED${NC} - Kind cluster 'lagoon-$cluster_name' not found"
        return 1
    fi

    # Check kubectl connectivity
    echo ""
    echo "=== Kubernetes Connectivity ==="
    if kubectl --context "$context" cluster-info >/dev/null 2>&1; then
        echo -e "${GREEN}OK${NC} - kubectl can connect to cluster"
    else
        echo -e "${RED}FAILED${NC} - Cannot connect to cluster"
        return 1
    fi

    # Check Lagoon pods
    echo ""
    echo "=== Lagoon Pods ==="
    UNHEALTHY=$(kubectl --context "$context" get pods -n "$NAMESPACE" --no-headers 2>/dev/null | grep -v "Running\|Completed" | wc -l)
    TOTAL=$(kubectl --context "$context" get pods -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
    RUNNING=$(kubectl --context "$context" get pods -n "$NAMESPACE" --no-headers 2>/dev/null | grep Running | wc -l)

    echo "Total pods: $TOTAL"
    echo "Running: $RUNNING"
    echo "Unhealthy: $UNHEALTHY"

    if [ "$UNHEALTHY" -gt 0 ]; then
        echo ""
        echo -e "${YELLOW}Unhealthy pods:${NC}"
        kubectl --context "$context" get pods -n "$NAMESPACE" --no-headers | grep -v "Running\|Completed"
    fi

    # Check specific critical pods based on cluster type
    echo ""
    echo "=== Critical Services ==="

    if [ "$cluster_name" = "prod" ]; then
        CRITICAL_PODS=("lagoon-core-api" "lagoon-core-keycloak" "lagoon-core-broker" "lagoon-build-deploy")
    else
        CRITICAL_PODS=("lagoon-build-deploy")
    fi

    for POD_PREFIX in "${CRITICAL_PODS[@]}"; do
        POD_STATUS=$(kubectl --context "$context" get pods -n "$NAMESPACE" -l "app.kubernetes.io/name=$POD_PREFIX" -o jsonpath='{.items[0].status.phase}' 2>/dev/null || echo "NotFound")
        if [ "$POD_STATUS" = "Running" ]; then
            echo -e "${GREEN}OK${NC} - $POD_PREFIX: Running"
        else
            echo -e "${RED}FAILED${NC} - $POD_PREFIX: $POD_STATUS"
        fi
    done

    # Check network access
    echo ""
    echo "=== Network Access ==="
    CLUSTER_IP=$(docker inspect "lagoon-$cluster_name-control-plane" --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>/dev/null)
    echo "Cluster IP: $CLUSTER_IP"

    return 0
}

echo "Checking Multi-cluster Lagoon health..."
echo ""

# Check both clusters
PROD_OK=0
NONPROD_OK=0

check_cluster "prod" && PROD_OK=1 || true
check_cluster "nonprod" && NONPROD_OK=1 || true

# Check port-forwards
echo ""
echo -e "${BLUE}========== Port Forwards ==========${NC}"
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

# Service accessibility (prod only since it hosts core services)
echo ""
echo "=== Service Accessibility ==="
PROD_IP=$(docker inspect "lagoon-prod-control-plane" --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>/dev/null)

if curl -s --connect-timeout 2 "http://localhost:8080/auth" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - Keycloak accessible at localhost:8080"
elif [ -n "$PROD_IP" ] && curl -s -k --connect-timeout 2 "https://keycloak.$PROD_IP.nip.io" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - Keycloak accessible via ingress"
else
    echo -e "${YELLOW}WARNING${NC} - Keycloak not accessible"
    echo "  Run: make port-forwards"
fi

if curl -s --connect-timeout 2 "http://localhost:7080/" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - API accessible at localhost:7080"
elif [ -n "$PROD_IP" ] && curl -s -k --connect-timeout 2 "https://api.$PROD_IP.nip.io" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - API accessible via ingress"
else
    echo -e "${YELLOW}WARNING${NC} - API not accessible"
    echo "  Run: make port-forwards"
fi

# Summary
echo ""
echo -e "${BLUE}========== Summary ==========${NC}"
if [ "$PROD_OK" -eq 1 ] && [ "$NONPROD_OK" -eq 1 ]; then
    echo -e "${GREEN}Both clusters are healthy and ready for use${NC}"
    echo ""
    echo "Quick start:"
    echo "  1. make port-forwards"
    echo "  2. make preview (or make up)"
else
    echo -e "${RED}Some clusters have issues that need attention${NC}"
    echo ""
    echo "Common fixes:"
    echo "  - RabbitMQ auth: make fix-rabbitmq"
    echo "  - Check logs: kubectl --context kind-lagoon-prod logs -n lagoon <pod-name>"
fi
