#!/bin/bash
# List all Kubernetes deploy targets in Lagoon
#
# Usage:
#   ./scripts/list-deploy-targets.sh
#   LAGOON_PRESET=multi-prod ./scripts/list-deploy-targets.sh
#
# Prerequisites:
#   - LAGOON_TOKEN environment variable set, OR
#   - Port-forwards running (script will try to get token)
#   - curl and jq installed
#
# See scripts/common.sh for configuration options.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

LAGOON_API_URL="${LAGOON_API_URL:-http://localhost:7080/graphql}"

# Get token if not set
if [ -z "$LAGOON_TOKEN" ]; then
    if [ -n "$KEYCLOAK_SECRET" ]; then
        log_info "Getting OAuth token..."
        LAGOON_TOKEN=$("$SCRIPT_DIR/get-token.sh" 2>/dev/null) || true
    fi
fi

if [ -z "$LAGOON_TOKEN" ]; then
    log_error "LAGOON_TOKEN environment variable is required"
    echo "Run: source ./scripts/get-token.sh" >&2
    exit 1
fi

# Query deploy targets
RESPONSE=$(curl -s -k -H "Authorization: Bearer $LAGOON_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"query": "{ allKubernetes { id name cloudProvider cloudRegion consoleUrl disabled } }"}' \
    "$LAGOON_API_URL")

# Check for errors
if echo "$RESPONSE" | jq -e '.errors' > /dev/null 2>&1; then
    ERROR=$(echo "$RESPONSE" | jq -r '.errors[0].message')
    log_error "$ERROR"
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
