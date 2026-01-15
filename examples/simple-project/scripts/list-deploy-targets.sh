#!/bin/bash
# List all Kubernetes deploy targets in Lagoon
#
# Usage:
#   ./scripts/list-deploy-targets.sh
#
# Prerequisites:
#   - LAGOON_TOKEN and LAGOON_API_URL environment variables set
#   - curl and jq installed

set -e

# Required variables
if [ -z "$LAGOON_TOKEN" ]; then
    echo "Error: LAGOON_TOKEN environment variable is required" >&2
    echo "Run: source ./scripts/get-lagoon-token.sh" >&2
    exit 1
fi

LAGOON_API_URL="${LAGOON_API_URL:-http://localhost:7080/graphql}"

# Query deploy targets
RESPONSE=$(curl -s -k -H "Authorization: Bearer $LAGOON_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"query": "{ allKubernetes { id name cloudProvider cloudRegion consoleUrl disabled } }"}' \
    "$LAGOON_API_URL")

# Check for errors
if echo "$RESPONSE" | jq -e '.errors' > /dev/null 2>&1; then
    ERROR=$(echo "$RESPONSE" | jq -r '.errors[0].message')
    echo "Error: $ERROR" >&2
    exit 1
fi

# Display results
TARGETS=$(echo "$RESPONSE" | jq '.data.allKubernetes')

if [ "$TARGETS" = "[]" ] || [ "$TARGETS" = "null" ]; then
    echo "No deploy targets found."
    echo ""
    echo "Create one with:"
    echo "  ./scripts/add-deploy-target.sh local-kind https://kubernetes.default.svc"
else
    echo "Deploy Targets:"
    echo "$TARGETS" | jq -r '.[] | "  ID: \(.id) | Name: \(.name) | Provider: \(.cloudProvider // "n/a") | Region: \(.cloudRegion // "n/a")"'
fi
