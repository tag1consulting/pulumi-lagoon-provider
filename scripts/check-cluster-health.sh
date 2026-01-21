#!/bin/bash
# Check Lagoon cluster health
#
# Usage:
#   ./scripts/check-cluster-health.sh
#   LAGOON_PRESET=multi-prod ./scripts/check-cluster-health.sh
#
# See scripts/common.sh for configuration options.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

echo "Checking Lagoon cluster health..."
echo "  Context:   $KUBE_CONTEXT"
echo "  Namespace: $LAGOON_NAMESPACE"
echo ""

# Check if kind cluster exists
echo "=== Kind Cluster ==="
CONTAINER_NAME="${KIND_CLUSTER_NAME}-control-plane"
if docker ps --filter "name=$CONTAINER_NAME" --format "{{.Names}}" | grep -q "$KIND_CLUSTER_NAME"; then
    echo -e "${GREEN}OK${NC} - Kind cluster '$KIND_CLUSTER_NAME' is running"
else
    echo -e "${RED}FAILED${NC} - Kind cluster '$KIND_CLUSTER_NAME' not found"
    exit 1
fi
echo ""

# Check kubectl connectivity
echo "=== Kubernetes Connectivity ==="
if check_cluster_connectivity; then
    echo -e "${GREEN}OK${NC} - kubectl can connect to cluster"
else
    exit 1
fi
echo ""

# Check Lagoon pods
echo "=== Lagoon Pods ==="
UNHEALTHY=$(kubectl --context "$KUBE_CONTEXT" get pods -n "$LAGOON_NAMESPACE" --no-headers 2>/dev/null | grep -v "Running\|Completed" | wc -l)
TOTAL=$(kubectl --context "$KUBE_CONTEXT" get pods -n "$LAGOON_NAMESPACE" --no-headers 2>/dev/null | wc -l)
RUNNING=$(kubectl --context "$KUBE_CONTEXT" get pods -n "$LAGOON_NAMESPACE" --no-headers 2>/dev/null | grep Running | wc -l)

echo "Total pods: $TOTAL"
echo "Running: $RUNNING"
echo "Unhealthy: $UNHEALTHY"

if [ "$UNHEALTHY" -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}Unhealthy pods:${NC}"
    kubectl --context "$KUBE_CONTEXT" get pods -n "$LAGOON_NAMESPACE" --no-headers | grep -v "Running\|Completed"
fi
echo ""

# Check specific critical pods (only if we have core services)
if [ -n "$KEYCLOAK_SVC" ]; then
    echo "=== Critical Services ==="

    # Build list of services to check based on what's configured
    SERVICES_TO_CHECK=()
    [ -n "$API_SVC" ] && SERVICES_TO_CHECK+=("$API_SVC")
    [ -n "$KEYCLOAK_SVC" ] && SERVICES_TO_CHECK+=("$KEYCLOAK_SVC")

    for SVC in "${SERVICES_TO_CHECK[@]}"; do
        if kubectl --context "$KUBE_CONTEXT" get svc -n "$LAGOON_NAMESPACE" "$SVC" >/dev/null 2>&1; then
            echo -e "${GREEN}OK${NC} - $SVC: exists"
        else
            echo -e "${RED}FAILED${NC} - $SVC: not found"
        fi
    done
    echo ""
fi

# Check network access
echo "=== Network Access ==="
CLUSTER_IP=$(docker inspect "$CONTAINER_NAME" --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>/dev/null || echo "unknown")
echo "Cluster IP: $CLUSTER_IP"

if [ "$CLUSTER_IP" != "unknown" ] && ping -c 1 -W 1 "$CLUSTER_IP" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - Docker network is routable"
else
    echo -e "${YELLOW}WARNING${NC} - Docker network may not be routable (common in WSL2)"
    echo "  Use port-forwards: ./scripts/setup-port-forwards.sh"
fi
echo ""

# Check if port-forwards are running
echo "=== Port Forwards ==="
if [ -n "$KEYCLOAK_SVC" ] && pgrep -f "port-forward.*$KEYCLOAK_SVC" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - Keycloak port-forward running (localhost:8080)"
else
    echo -e "${YELLOW}INFO${NC} - No Keycloak port-forward"
fi

if [ -n "$API_SVC" ] && pgrep -f "port-forward.*$API_SVC" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC} - API port-forward running (localhost:7080)"
else
    echo -e "${YELLOW}INFO${NC} - No API port-forward"
fi
echo ""

# Test service accessibility (only if we have core services)
if [ -n "$KEYCLOAK_SVC" ]; then
    echo "=== Service Accessibility ==="

    if curl -s --connect-timeout 2 "http://localhost:8080/auth" >/dev/null 2>&1; then
        echo -e "${GREEN}OK${NC} - Keycloak accessible at localhost:8080"
    elif [ "$CLUSTER_IP" != "unknown" ] && curl -s -k --connect-timeout 2 "https://keycloak.$CLUSTER_IP.nip.io" >/dev/null 2>&1; then
        echo -e "${GREEN}OK${NC} - Keycloak accessible via ingress"
    else
        echo -e "${YELLOW}WARNING${NC} - Keycloak not accessible"
        echo "  Run: ./scripts/setup-port-forwards.sh"
    fi

    if curl -s --connect-timeout 2 "http://localhost:7080/" >/dev/null 2>&1; then
        echo -e "${GREEN}OK${NC} - API accessible at localhost:7080"
    elif [ "$CLUSTER_IP" != "unknown" ] && curl -s -k --connect-timeout 2 "https://api.$CLUSTER_IP.nip.io" >/dev/null 2>&1; then
        echo -e "${GREEN}OK${NC} - API accessible via ingress"
    else
        echo -e "${YELLOW}WARNING${NC} - API not accessible"
        echo "  Run: ./scripts/setup-port-forwards.sh"
    fi
    echo ""
fi

echo "=== Summary ==="
if [ "$UNHEALTHY" -le 1 ]; then
    echo -e "${GREEN}Cluster is healthy and ready for use${NC}"
    echo ""
    echo "Quick start:"
    echo "  1. ./scripts/setup-port-forwards.sh"
    echo "  2. source ./scripts/get-token.sh"
    echo "  3. pulumi up"
else
    echo -e "${RED}Cluster has issues that need attention${NC}"
    echo ""
    echo "Common fixes:"
    echo "  - RabbitMQ auth: ./scripts/fix-rabbitmq-password.sh"
    echo "  - Check logs: kubectl --context $KUBE_CONTEXT logs -n $LAGOON_NAMESPACE <pod-name>"
fi
