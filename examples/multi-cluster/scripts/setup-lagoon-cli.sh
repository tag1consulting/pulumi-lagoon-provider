#!/bin/bash
# Configure the lagoon CLI to work with the local test cluster
#
# This script:
# 1. Obtains an OAuth token from Keycloak using get-lagoon-token.sh
# 2. Configures the lagoon CLI with the test cluster settings
# 3. Verifies the configuration with 'lagoon whoami'
#
# Usage: ./setup-lagoon-cli.sh [options]
#   -n, --name NAME        Configuration name (default: local-test)
#   -u, --user USERNAME    Keycloak username (default: lagoonadmin)
#   -p, --password PASS    Keycloak password (default: lagoonadmin)
#   -f, --force            Overwrite existing configuration
#   -h, --help             Show this help message

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default configuration
CONFIG_NAME="${LAGOON_CONFIG_NAME:-local-test}"
USERNAME="${LAGOON_USER:-lagoonadmin}"
PASSWORD="${LAGOON_PASSWORD:-lagoonadmin}"
FORCE=false

# Lagoon API endpoints (using HTTPS ingress with self-signed certs)
API_URL="https://api.lagoon.test/graphql"
UI_URL="https://ui.lagoon.test"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            CONFIG_NAME="$2"
            shift 2
            ;;
        -u|--user)
            USERNAME="$2"
            shift 2
            ;;
        -p|--password)
            PASSWORD="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Configure the lagoon CLI to work with the local test cluster."
            echo ""
            echo "Options:"
            echo "  -n, --name NAME        Configuration name (default: local-test)"
            echo "  -u, --user USERNAME    Keycloak username (default: lagoonadmin)"
            echo "  -p, --password PASS    Keycloak password (default: lagoonadmin)"
            echo "  -f, --force            Overwrite existing configuration"
            echo "  -h, --help             Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  LAGOON_CONFIG_NAME     Configuration name (default: local-test)"
            echo "  LAGOON_USER            Username (default: lagoonadmin)"
            echo "  LAGOON_PASSWORD        Password (default: lagoonadmin)"
            echo ""
            echo "Example:"
            echo "  $0                      # Use defaults"
            echo "  $0 -f                   # Force reconfigure"
            echo "  $0 -u myuser -p mypass  # Custom credentials"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "=== Lagoon CLI Configuration ==="
echo ""

# Check if lagoon CLI is installed
if ! command -v lagoon &> /dev/null; then
    echo "ERROR: lagoon CLI is not installed."
    echo ""
    echo "Install it from: https://github.com/uselagoon/lagoon-cli"
    echo ""
    echo "Quick install (Linux/macOS):"
    echo "  curl -L https://github.com/uselagoon/lagoon-cli/releases/latest/download/lagoon-cli-\$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 -o lagoon"
    echo "  chmod +x lagoon"
    echo "  sudo mv lagoon /usr/local/bin/"
    exit 1
fi
echo "Found lagoon CLI: $(command -v lagoon)"

# Check if configuration already exists
EXISTING_CONFIG=$(lagoon config list 2>/dev/null | grep -w "${CONFIG_NAME}" || true)
if [ -n "$EXISTING_CONFIG" ] && [ "$FORCE" = false ]; then
    echo ""
    echo "Configuration '${CONFIG_NAME}' already exists."
    echo "Use -f/--force to overwrite, or use a different name with -n/--name."
    echo ""
    echo "Current configuration:"
    lagoon config list
    exit 0
fi

# Remove existing configuration if forcing
if [ -n "$EXISTING_CONFIG" ] && [ "$FORCE" = true ]; then
    echo "Removing existing configuration '${CONFIG_NAME}'..."
    lagoon config delete --lagoon "${CONFIG_NAME}" --force 2>/dev/null || true
fi

# Get OAuth token
echo ""
echo "Obtaining OAuth token from Keycloak..."
TOKEN=$("${SCRIPT_DIR}/get-lagoon-token.sh" -u "${USERNAME}" -p "${PASSWORD}" -q) || {
    echo "ERROR: Failed to obtain token. See errors above."
    exit 1
}
echo "Token obtained successfully"

# Configure lagoon CLI
echo ""
echo "Configuring lagoon CLI..."
echo "  Name: ${CONFIG_NAME}"
echo "  API:  ${API_URL}"
echo "  UI:   ${UI_URL}"

lagoon config add \
    --lagoon "${CONFIG_NAME}" \
    --graphql "${API_URL}" \
    --ui "${UI_URL}" \
    --token "${TOKEN}"

# Set as default
echo ""
echo "Setting '${CONFIG_NAME}' as default..."
lagoon config default --lagoon "${CONFIG_NAME}"

# Verify configuration
echo ""
echo "Verifying configuration..."
echo ""
if lagoon whoami; then
    echo ""
    echo "=== Configuration Successful ==="
    echo ""
    echo "You can now use the lagoon CLI. Examples:"
    echo "  lagoon list projects"
    echo "  lagoon add project --help"
    echo "  lagoon get environment --help"
    echo ""
    echo "Note: Tokens expire after ~5 minutes. Re-run this script to refresh."
else
    echo ""
    echo "ERROR: Configuration verification failed."
    echo "The token may have expired or the API is not accessible."
    exit 1
fi
