#!/bin/bash
# Configure the lagoon CLI to work with a local test cluster
#
# Usage:
#   ./scripts/setup-lagoon-cli.sh [options]
#   LAGOON_PRESET=multi-prod ./scripts/setup-lagoon-cli.sh
#
# Options:
#   -n, --name NAME        Configuration name (default: local-test)
#   -f, --force            Overwrite existing configuration
#   -h, --help             Show this help message
#
# See scripts/common.sh and docs/lagoon-cli-setup.md for full documentation.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

# multi-nonprod has no core services (no Keycloak, no API) — fail fast with a clear message
if [ "${LAGOON_PRESET:-}" = "multi-nonprod" ]; then
    echo "ERROR: LAGOON_PRESET=multi-nonprod is not supported for CLI setup."
    echo "The non-production cluster has no Keycloak or API services."
    echo "Use the production cluster instead:"
    echo "  LAGOON_PRESET=multi-prod ./scripts/setup-lagoon-cli.sh"
    exit 1
fi

# Defaults
CONFIG_NAME="${LAGOON_CONFIG_NAME:-local-test}"
FORCE=false

# Service endpoints (using port-forwarded HTTP to avoid TLS)
API_URL="${LAGOON_API_URL:-http://localhost:7080/graphql}"
UI_URL="${LAGOON_UI_URL:-http://localhost:8080}"

while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            CONFIG_NAME="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Configure the lagoon CLI to work with a local test cluster."
            echo ""
            echo "Options:"
            echo "  -n, --name NAME    Configuration name (default: local-test)"
            echo "  -f, --force        Overwrite existing configuration"
            echo "  -h, --help         Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  LAGOON_PRESET      Cluster preset: single (default), multi-prod, multi-nonprod"
            echo "  LAGOON_CONFIG_NAME Configuration name (default: local-test)"
            echo "  LAGOON_API_URL     API endpoint (default: http://localhost:7080/graphql)"
            echo "  LAGOON_UI_URL      UI endpoint (default: http://localhost:8080)"
            echo "  LAGOON_USERNAME    Keycloak username (default: lagoonadmin)"
            echo "  LAGOON_PASSWORD    Keycloak password (fetched from cluster if not set)"
            echo ""
            echo "Examples:"
            echo "  $0                              # Single-cluster with defaults"
            echo "  $0 -f                           # Force reconfigure"
            echo "  LAGOON_PRESET=multi-prod $0     # Multi-cluster production"
            echo ""
            echo "See docs/lagoon-cli-setup.md for full documentation."
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "=== Lagoon CLI Configuration ==="
echo "  Preset:    ${LAGOON_PRESET:-single}"
echo "  Context:   $KUBE_CONTEXT"
echo "  Config:    $CONFIG_NAME"
echo ""

# Check if lagoon CLI is installed
if ! command -v lagoon &> /dev/null; then
    echo "ERROR: lagoon CLI is not installed."
    echo ""
    echo "Install it from: https://github.com/uselagoon/lagoon-cli/releases/latest"
    echo ""
    echo "Quick install:"
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac
    echo "  curl -L https://github.com/uselagoon/lagoon-cli/releases/latest/download/lagoon-cli-${OS}-${ARCH} -o lagoon"
    echo "  chmod +x lagoon && sudo mv lagoon /usr/local/bin/"
    exit 1
fi
echo "Found lagoon CLI: $(command -v lagoon) ($(lagoon version 2>/dev/null || echo 'version unknown'))"

# Check if configuration already exists
EXISTING_CONFIG=$(lagoon config list 2>/dev/null | grep -w "${CONFIG_NAME}" || true)
if [ -n "$EXISTING_CONFIG" ] && [ "$FORCE" = false ]; then
    echo ""
    echo "Configuration '${CONFIG_NAME}' already exists. Use -f to overwrite."
    echo ""
    echo "Current configs:"
    lagoon config list
    exit 0
fi

# Remove existing config if forcing
if [ -n "$EXISTING_CONFIG" ] && [ "$FORCE" = true ]; then
    echo "Removing existing configuration '${CONFIG_NAME}'..."
    lagoon config delete --lagoon "${CONFIG_NAME}" --force 2>/dev/null || true
fi

# Ensure port-forwards are running before getting token
if ! curl -s --connect-timeout 2 "http://localhost:8080/" >/dev/null 2>&1; then
    echo ""
    echo "Keycloak is not reachable at http://localhost:8080."
    echo "Starting port-forwards..."
    "$SCRIPT_DIR/setup-port-forwards.sh"
    # Poll until Keycloak is reachable (up to 30s)
    TIMEOUT=30
    ELAPSED=0
    until curl -s --connect-timeout 2 "http://localhost:8080/" >/dev/null 2>&1; do
        if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
            echo "ERROR: Keycloak did not become reachable within ${TIMEOUT}s."
            echo "Check port-forward status: kubectl --context \"$KUBE_CONTEXT\" get pods -n \"$LAGOON_NAMESPACE\" | grep keycloak"
            exit 1
        fi
        sleep 2
        ELAPSED=$((ELAPSED + 2))
    done
fi

# Get OAuth token
echo ""
echo "Obtaining OAuth token from Keycloak..."
TOKEN=$("$SCRIPT_DIR/get-token.sh") || {
    echo "ERROR: Failed to obtain token. Check that Keycloak is running and accessible."
    echo "Run: ./scripts/setup-port-forwards.sh"
    exit 1
}
echo "Token obtained successfully."

# Configure lagoon CLI
echo ""
echo "Running lagoon config add..."
echo "  Name:    ${CONFIG_NAME}"
echo "  GraphQL: ${API_URL}"
echo "  UI:      ${UI_URL}"

lagoon config add \
    --lagoon "${CONFIG_NAME}" \
    --graphql "${API_URL}" \
    --ui "${UI_URL}" \
    --token "${TOKEN}"

# Set as default
echo ""
echo "Setting '${CONFIG_NAME}' as default configuration..."
lagoon config default --lagoon "${CONFIG_NAME}"

# Verify
echo ""
echo "Verifying configuration..."
echo ""
if lagoon whoami; then
    echo ""
    echo "=== Configuration Successful ==="
    echo ""
    echo "Example commands:"
    echo "  lagoon list projects"
    echo "  lagoon get project --project <name>"
    echo "  lagoon list variables --project <name>"
    echo ""
    echo "Note: Tokens expire after ~5 minutes. Re-run with -f to refresh."
    echo "See docs/lagoon-cli-setup.md for full documentation."
else
    echo ""
    echo "ERROR: Verification failed — token may have expired or API is unreachable."
    echo "Try: ./scripts/setup-port-forwards.sh && ./scripts/setup-lagoon-cli.sh -f"
    exit 1
fi
