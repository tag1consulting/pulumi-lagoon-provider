#!/bin/bash
# Add a Kubernetes deploy target to Lagoon
#
# Deploy targets (also called "Kubernetes" or "OpenShift" in the API) are
# the clusters where Lagoon deploys projects. You need at least one deploy
# target before you can create projects.
#
# Usage:
#   ./scripts/add-deploy-target.sh [name] [console-url]
#
# Examples:
#   # Add local kind cluster
#   ./scripts/add-deploy-target.sh local-kind https://kubernetes.default.svc
#
#   # Add production cluster
#   ./scripts/add-deploy-target.sh prod-cluster https://api.prod.example.com:6443
#
# Prerequisites:
#   - LAGOON_TOKEN and LAGOON_API_URL environment variables set
#   - curl and jq installed
#
# Optional environment variables:
#   CLOUD_PROVIDER - Cloud provider name (default: kind)
#   CLOUD_REGION   - Cloud region (default: local)

set -e

# Arguments
NAME="${1:-local-kind}"
CONSOLE_URL="${2:-https://kubernetes.default.svc}"

# Configuration
CLOUD_PROVIDER="${CLOUD_PROVIDER:-kind}"
CLOUD_REGION="${CLOUD_REGION:-local}"

# Required variables
if [ -z "$LAGOON_TOKEN" ]; then
    echo "Error: LAGOON_TOKEN environment variable is required" >&2
    echo "Run: source ./scripts/get-lagoon-token.sh" >&2
    exit 1
fi

LAGOON_API_URL="${LAGOON_API_URL:-http://localhost:7080/graphql}"

echo "Adding deploy target '$NAME'..."

# Create the deploy target
RESPONSE=$(curl -s -k -H "Authorization: Bearer $LAGOON_TOKEN" \
    -H "Content-Type: application/json" \
    --data-binary @- \
    "$LAGOON_API_URL" <<EOF
{
  "query": "mutation { addKubernetes(input: { name: \"${NAME}\", consoleUrl: \"${CONSOLE_URL}\", cloudProvider: \"${CLOUD_PROVIDER}\", cloudRegion: \"${CLOUD_REGION}\" }) { id name cloudProvider cloudRegion } }"
}
EOF
)

# Check for errors
if echo "$RESPONSE" | jq -e '.errors' > /dev/null 2>&1; then
    ERROR=$(echo "$RESPONSE" | jq -r '.errors[0].message')
    echo "Error: $ERROR" >&2
    exit 1
fi

# Extract and display result
DEPLOY_TARGET=$(echo "$RESPONSE" | jq '.data.addKubernetes')
ID=$(echo "$DEPLOY_TARGET" | jq -r '.id')

echo ""
echo "Deploy target created successfully!"
echo "$DEPLOY_TARGET" | jq '.'
echo ""
echo "Use this ID in your Pulumi configuration:"
echo "  pulumi config set deploytargetId $ID"
