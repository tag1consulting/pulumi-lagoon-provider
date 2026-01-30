#!/bin/bash
#
# Complete Pulumi Lagoon Provider Development Environment Setup
#
# This script automates the complete setup process:
# 1. Creates Python virtual environment
# 2. Installs the provider in development mode
# 3. Creates a Kind Kubernetes cluster
# 4. Installs Lagoon via Helm (using Pulumi)
# 5. Sets up the example project
#
# Usage:
#   ./scripts/setup-complete.sh [options]
#
# Options:
#   --skip-cluster    Skip Kind cluster creation (use existing)
#   --skip-provider   Skip provider installation
#   --skip-example    Skip example project setup
#   --help            Show this help message
#
# Prerequisites:
#   - Docker installed and running
#   - kind CLI installed (https://kind.sigs.k8s.io/)
#   - kubectl installed
#   - pulumi CLI installed
#   - Python 3.8+
#
# Total time: ~15-20 minutes for full setup

set -e

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
CLUSTER_NAME="${CLUSTER_NAME:-lagoon-test}"
VENV_DIR="${VENV_DIR:-$ROOT_DIR/venv}"
TEST_CLUSTER_DIR="$ROOT_DIR/test-cluster"
EXAMPLE_DIR="$ROOT_DIR/examples/simple-project"

# Flags
SKIP_CLUSTER=false
SKIP_PROVIDER=false
SKIP_EXAMPLE=false

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

#==============================================================================
# Helper Functions
#==============================================================================

log_header() {
    echo ""
    echo -e "${BLUE}================================================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}================================================================${NC}"
    echo ""
}

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${GREEN}[STEP $1]${NC} $2"
}

show_help() {
    cat << EOF
Pulumi Lagoon Provider - Complete Setup Script

Usage: $0 [options]

Options:
  --skip-cluster    Skip Kind cluster creation (use existing cluster)
  --skip-provider   Skip provider installation
  --skip-example    Skip example project setup
  --help            Show this help message

Environment Variables:
  CLUSTER_NAME      Kind cluster name (default: lagoon-test)
  VENV_DIR          Python virtual environment path (default: ./venv)

Prerequisites:
  - Docker installed and running
  - kind CLI installed (https://kind.sigs.k8s.io/)
  - kubectl installed
  - pulumi CLI installed
  - Python 3.8+

Examples:
  # Full setup from scratch
  $0

  # Skip cluster if already running
  $0 --skip-cluster

  # Only set up the provider
  $0 --skip-cluster --skip-example
EOF
}

check_prerequisites() {
    log_step "0" "Checking prerequisites..."

    local missing=()

    if ! command -v docker &> /dev/null; then
        missing+=("docker")
    fi

    if ! command -v kind &> /dev/null; then
        missing+=("kind (https://kind.sigs.k8s.io/)")
    fi

    if ! command -v kubectl &> /dev/null; then
        missing+=("kubectl")
    fi

    if ! command -v pulumi &> /dev/null; then
        missing+=("pulumi")
    fi

    if ! command -v python3 &> /dev/null; then
        missing+=("python3")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing prerequisites:"
        for item in "${missing[@]}"; do
            echo "  - $item"
        done
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi

    log_info "All prerequisites satisfied"
}

#==============================================================================
# Setup Steps
#==============================================================================

setup_venv() {
    log_step "1" "Setting up Python virtual environment..."

    if [ -d "$VENV_DIR" ]; then
        log_info "Virtual environment already exists at $VENV_DIR"
    else
        python3 -m venv "$VENV_DIR"
        log_info "Created virtual environment at $VENV_DIR"
    fi

    # shellcheck source=/dev/null
    source "$VENV_DIR/bin/activate"
    pip install --upgrade pip -q
    log_info "Virtual environment activated"
}

install_provider() {
    log_step "2" "Installing Pulumi Lagoon provider..."

    if [ "$SKIP_PROVIDER" = true ]; then
        log_info "Skipping provider installation (--skip-provider)"
        return
    fi

    cd "$ROOT_DIR"
    # shellcheck source=/dev/null
    source "$VENV_DIR/bin/activate"
    pip install -e . -q
    log_info "Provider installed in development mode"
}

setup_cluster() {
    log_step "3" "Setting up Kind cluster and Lagoon..."

    if [ "$SKIP_CLUSTER" = true ]; then
        log_info "Skipping cluster creation (--skip-cluster)"
        if kind get clusters 2>/dev/null | grep -q "$CLUSTER_NAME"; then
            log_info "Using existing cluster: $CLUSTER_NAME"
        else
            log_warn "No existing cluster found with name: $CLUSTER_NAME"
        fi
        return
    fi

    # Check if cluster already exists
    if kind get clusters 2>/dev/null | grep -q "$CLUSTER_NAME"; then
        log_info "Kind cluster '$CLUSTER_NAME' already exists"
        log_info "Checking Lagoon installation..."
        if kubectl --context "kind-$CLUSTER_NAME" get ns lagoon &>/dev/null; then
            log_info "Lagoon is already installed"
            return
        fi
    fi

    cd "$TEST_CLUSTER_DIR"

    # Set up test-cluster venv
    if [ ! -d "venv" ]; then
        python3 -m venv venv
    fi
    # shellcheck source=/dev/null
    source venv/bin/activate
    pip install -r requirements.txt -q

    # Initialize stack if needed
    pulumi stack select dev 2>/dev/null || pulumi stack init dev

    log_info "Deploying Kind cluster and Lagoon (this takes 10-15 minutes)..."
    pulumi up --yes

    log_info "Cluster and Lagoon deployed successfully"
}

wait_for_lagoon() {
    log_step "4" "Waiting for Lagoon to be ready..."

    if [ "$SKIP_CLUSTER" = true ]; then
        log_info "Skipping wait (--skip-cluster)"
        return
    fi

    # First, wait for the api-migratedb job to complete (success or failure)
    # This prevents blocking on a failed job pod (issue #6)
    log_info "Waiting for database migration job to complete..."
    for i in $(seq 1 12); do
        JOB_STATUS=$(kubectl --context "kind-$CLUSTER_NAME" get job lagoon-core-api-migratedb -n lagoon -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' 2>/dev/null)
        JOB_FAILED=$(kubectl --context "kind-$CLUSTER_NAME" get job lagoon-core-api-migratedb -n lagoon -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}' 2>/dev/null)

        if [ "$JOB_STATUS" = "True" ]; then
            log_info "Migration job completed successfully."
            break
        elif [ "$JOB_FAILED" = "True" ]; then
            log_warn "Migration job failed, deleting for retry..."
            kubectl --context "kind-$CLUSTER_NAME" delete job lagoon-core-api-migratedb -n lagoon 2>/dev/null || true
            sleep 10
        else
            log_info "Waiting for migration job... (attempt $i/12)"
            sleep 10
        fi
    done

    # Now wait for the key deployment pods (excluding completed/failed job pods)
    log_info "Waiting for Lagoon core pods..."
    kubectl --context "kind-$CLUSTER_NAME" wait --for=condition=ready pod \
        -l app.kubernetes.io/name=lagoon-core --field-selector=status.phase=Running -n lagoon --timeout=300s 2>/dev/null || true

    log_info "Waiting for Broker pods..."
    kubectl --context "kind-$CLUSTER_NAME" wait --for=condition=ready pod \
        -l app.kubernetes.io/name=broker -n lagoon --timeout=300s 2>/dev/null || true

    # Show pod status
    log_info "Current pod status:"
    kubectl --context "kind-$CLUSTER_NAME" get pods -n lagoon
}

setup_example() {
    log_step "5" "Setting up example project..."

    if [ "$SKIP_EXAMPLE" = true ]; then
        log_info "Skipping example setup (--skip-example)"
        return
    fi

    cd "$EXAMPLE_DIR"

    # shellcheck source=/dev/null
    source "$VENV_DIR/bin/activate"

    # Initialize stack if needed
    pulumi stack select test 2>/dev/null || pulumi stack init test

    log_info "Example project initialized"
}

show_summary() {
    log_header "Setup Complete!"

    echo "What was set up:"
    echo ""
    if [ "$SKIP_PROVIDER" = false ]; then
        echo "  [x] Python virtual environment at: $VENV_DIR"
        echo "  [x] Pulumi Lagoon provider installed"
    fi
    if [ "$SKIP_CLUSTER" = false ]; then
        echo "  [x] Kind cluster: $CLUSTER_NAME"
        echo "  [x] Lagoon Core (API, UI, Keycloak, RabbitMQ)"
        echo "  [x] Lagoon Build Deploy controller"
        echo "  [x] Harbor registry"
        echo "  [x] Ingress controller"
    fi
    if [ "$SKIP_EXAMPLE" = false ]; then
        echo "  [x] Example project initialized"
    fi

    echo ""
    echo "Next Steps:"
    echo ""
    echo "  1. Activate the virtual environment:"
    echo "     source $VENV_DIR/bin/activate"
    echo ""
    echo "  2. Deploy the example project:"
    echo "     cd $EXAMPLE_DIR"
    echo "     ./scripts/run-pulumi.sh up"
    echo ""
    echo "  Or use the Makefile:"
    echo "     make example-up"
    echo ""
    echo "Access Lagoon (via port-forwards):"
    echo "  - API:      http://localhost:7080/graphql"
    echo "  - Keycloak: http://localhost:8080/auth"
    echo ""
    echo "Get credentials:"
    echo "  kubectl --context kind-$CLUSTER_NAME get secret lagoon-core-keycloak \\"
    echo "    -n lagoon -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d"
    echo ""
}

#==============================================================================
# Main
#==============================================================================

main() {
    # Parse arguments
    while [ $# -gt 0 ]; do
        case "$1" in
            --skip-cluster)
                SKIP_CLUSTER=true
                shift
                ;;
            --skip-provider)
                SKIP_PROVIDER=true
                shift
                ;;
            --skip-example)
                SKIP_EXAMPLE=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    log_header "Pulumi Lagoon Provider - Complete Setup"

    check_prerequisites
    setup_venv
    install_provider
    setup_cluster
    wait_for_lagoon
    setup_example
    show_summary
}

main "$@"
