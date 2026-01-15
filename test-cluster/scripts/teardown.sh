#!/bin/bash
# Teardown script for Lagoon test cluster

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
CLUSTER_NAME="${1:-lagoon-test}"

echo "=== Lagoon Test Cluster Teardown ==="
echo

cd "$PROJECT_DIR"

# Check if virtual environment exists and activate
if [ -d "venv" ]; then
    source venv/bin/activate
fi

# Option 1: Use Pulumi to destroy (preserves state)
if [ "$1" == "--pulumi" ]; then
    echo "Destroying cluster using Pulumi..."
    pulumi destroy --yes
    echo "Cluster destroyed via Pulumi"
else
    # Option 2: Direct kind deletion (faster)
    echo "Destroying kind cluster directly..."
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        kind delete cluster --name "${CLUSTER_NAME}"
        echo "Cluster '${CLUSTER_NAME}' deleted"
    else
        echo "Cluster '${CLUSTER_NAME}' not found"
    fi

    # Clean up Pulumi state to match
    echo
    echo "Note: Cluster was deleted directly. To clean Pulumi state, run:"
    echo "  pulumi refresh"
    echo "  or"
    echo "  pulumi stack rm dev --yes"
fi

echo
echo "=== Teardown complete! ==="
echo
echo "To start fresh, run: ./scripts/setup.sh"
echo
