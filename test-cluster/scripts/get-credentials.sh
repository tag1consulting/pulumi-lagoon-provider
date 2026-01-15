#!/bin/bash
# Extract Lagoon API credentials from the test cluster

set -e

CLUSTER_NAME="${1:-lagoon-test}"
CONTEXT="kind-${CLUSTER_NAME}"

echo "=== Lagoon Test Cluster Credentials ==="
echo

# Check if cluster exists
if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "ERROR: Cluster '${CLUSTER_NAME}' not found"
    echo "Available clusters:"
    kind get clusters
    exit 1
fi

# Check if lagoon namespace exists
if ! kubectl --context "${CONTEXT}" get namespace lagoon &>/dev/null; then
    echo "ERROR: Lagoon namespace not found in cluster"
    echo "Has the cluster been fully deployed? Run: pulumi up"
    exit 1
fi

echo "Cluster: ${CLUSTER_NAME}"
echo "Context: ${CONTEXT}"
echo

# Get API endpoint
echo "=== Lagoon API ==="
API_URL="http://api.lagoon.test/graphql"
echo "URL: ${API_URL}"
echo

# Try to get admin token
echo "=== Admin Token ==="
if kubectl --context "${CONTEXT}" -n lagoon get secret lagoon-core-api-admin-token &>/dev/null; then
    TOKEN=$(kubectl --context "${CONTEXT}" -n lagoon get secret lagoon-core-api-admin-token -o jsonpath='{.data.token}' 2>/dev/null | base64 -d 2>/dev/null || echo "")

    if [ -n "$TOKEN" ]; then
        echo "Token: ${TOKEN}"
    else
        echo "Token secret exists but could not extract value"
        echo "Try: kubectl --context ${CONTEXT} -n lagoon get secret lagoon-core-api-admin-token -o yaml"
    fi
else
    echo "Admin token secret not found yet"
    echo "Lagoon may still be initializing. Wait a few minutes and try again."
    echo
    echo "Check Lagoon pods status:"
    kubectl --context "${CONTEXT}" -n lagoon get pods
fi

echo
echo "=== Harbor Registry ==="
echo "URL: http://harbor.lagoon.test"
echo "Username: admin"
echo "Password: Harbor12345"
echo

echo "=== /etc/hosts Configuration ==="
echo "Add this line to your /etc/hosts file:"
echo "127.0.0.1 api.lagoon.test ui.lagoon.test harbor.lagoon.test"
echo

echo "=== Quick Test ==="
if [ -n "$TOKEN" ]; then
    echo "Test the GraphQL API:"
    echo "curl -X POST ${API_URL} \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -H 'Authorization: Bearer ${TOKEN}' \\"
    echo "  -d '{\"query\":\"{ allProjects { id name } }\"}'"
    echo

    echo "Or configure pulumi-lagoon-provider:"
    echo "export LAGOON_API_URL='${API_URL}'"
    echo "export LAGOON_TOKEN='${TOKEN}'"
    echo
    echo "# Or in your Pulumi program:"
    echo "pulumi config set lagoon:apiUrl '${API_URL}'"
    echo "pulumi config set lagoon:token '${TOKEN}' --secret"
fi

echo
echo "=== Useful Commands ==="
echo "View all Lagoon pods:"
echo "  kubectl --context ${CONTEXT} -n lagoon get pods"
echo
echo "View Lagoon API logs:"
echo "  kubectl --context ${CONTEXT} -n lagoon logs -l app.kubernetes.io/name=api --tail=100 -f"
echo
echo "Access Lagoon API pod:"
echo "  kubectl --context ${CONTEXT} -n lagoon exec -it deploy/lagoon-core-api -- /bin/sh"
echo
echo "Port-forward to API (alternative to ingress):"
echo "  kubectl --context ${CONTEXT} -n lagoon port-forward svc/lagoon-core-api 3000:3000"
echo
