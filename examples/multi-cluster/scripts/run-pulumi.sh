#!/bin/bash
# Wrapper script for running Pulumi commands with Lagoon token refresh
# Usage: ./scripts/run-pulumi.sh <pulumi-command> [args...]
#
# This script:
# 1. Refreshes the Lagoon API token using lagoon-cli
# 2. Runs the specified Pulumi command
# 3. Works with the multi-cluster example

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to refresh Lagoon token
refresh_lagoon_token() {
    log_info "Refreshing Lagoon API token..."

    if ! command -v lagoon &> /dev/null; then
        log_warn "lagoon-cli not found. Using existing LAGOON_TOKEN if set."
        return 0
    fi

    # Get token from lagoon-cli
    TOKEN=$(lagoon get token 2>/dev/null || echo "")

    if [ -n "$TOKEN" ]; then
        export LAGOON_TOKEN="$TOKEN"
        log_info "Lagoon token refreshed successfully"
    else
        log_warn "Could not refresh token. Using existing LAGOON_TOKEN if set."
    fi
}

# Function to check prerequisites
check_prerequisites() {
    local missing=()

    if ! command -v pulumi &> /dev/null; then
        missing+=("pulumi")
    fi

    if ! command -v kind &> /dev/null; then
        missing+=("kind")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing[*]}"
        exit 1
    fi
}

# Main
check_prerequisites

# Refresh token before running Pulumi
refresh_lagoon_token

# Change to project directory
cd "$PROJECT_DIR"

# Activate virtual environment if it exists
if [ -f "venv/bin/activate" ]; then
    source venv/bin/activate
fi

# Run Pulumi with all arguments
log_info "Running: pulumi $*"
pulumi "$@"
