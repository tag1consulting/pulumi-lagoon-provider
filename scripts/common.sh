#!/bin/bash
# Common functions and configuration for Lagoon scripts
#
# This file should be sourced by other scripts:
#   source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
#
# Configuration via environment variables:
#
# Cluster Configuration:
#   KUBE_CONTEXT         - Kubernetes context (default: kind-lagoon-test)
#   LAGOON_NAMESPACE     - Namespace for Lagoon core (default: lagoon)
#   KIND_CLUSTER_NAME    - Kind cluster name for docker inspect (default: lagoon-test)
#
# Service Names (vary between single-cluster and multi-cluster):
#   KEYCLOAK_SVC         - Keycloak service name (default: lagoon-core-keycloak)
#   API_SVC              - API service name (default: lagoon-core-api)
#   KEYCLOAK_SECRET      - Keycloak secret name (default: lagoon-core-keycloak)
#   BROKER_SECRET        - RabbitMQ broker secret (default: lagoon-core-broker)
#   REMOTE_SECRET        - Remote controller secret (default: lagoon-remote)
#   REMOTE_DEPLOYMENT    - Remote controller deployment (default: lagoon-remote-kubernetes-build-deploy)
#
# Presets:
#   LAGOON_PRESET        - Use a preset configuration: "single" or "multi-prod" or "multi-nonprod"
#
# Examples:
#   # Single-cluster (test-cluster style):
#   export LAGOON_PRESET=single
#   ./scripts/check-cluster-health.sh
#
#   # Multi-cluster production:
#   export LAGOON_PRESET=multi-prod
#   ./scripts/check-cluster-health.sh
#
#   # Custom configuration:
#   export KUBE_CONTEXT=my-context
#   export LAGOON_NAMESPACE=my-lagoon
#   ./scripts/check-cluster-health.sh

# Apply presets if specified
case "${LAGOON_PRESET:-}" in
    single|"")
        # Single-cluster defaults (test-cluster style)
        KUBE_CONTEXT="${KUBE_CONTEXT:-kind-lagoon-test}"
        LAGOON_NAMESPACE="${LAGOON_NAMESPACE:-lagoon}"
        KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-lagoon-test}"
        KEYCLOAK_SVC="${KEYCLOAK_SVC:-lagoon-core-keycloak}"
        API_SVC="${API_SVC:-lagoon-core-api}"
        KEYCLOAK_SECRET="${KEYCLOAK_SECRET:-lagoon-core-keycloak}"
        BROKER_SECRET="${BROKER_SECRET:-lagoon-core-broker}"
        REMOTE_SECRET="${REMOTE_SECRET:-lagoon-remote}"
        REMOTE_DEPLOYMENT="${REMOTE_DEPLOYMENT:-lagoon-remote-kubernetes-build-deploy}"
        ;;
    multi-prod)
        # Multi-cluster production defaults
        KUBE_CONTEXT="${KUBE_CONTEXT:-kind-lagoon-prod}"
        LAGOON_NAMESPACE="${LAGOON_NAMESPACE:-lagoon-core}"
        KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-lagoon-prod}"
        KEYCLOAK_SVC="${KEYCLOAK_SVC:-prod-core-lagoon-core-keycloak}"
        API_SVC="${API_SVC:-prod-core-lagoon-core-api}"
        KEYCLOAK_SECRET="${KEYCLOAK_SECRET:-prod-core-lagoon-core-keycloak}"
        BROKER_SECRET="${BROKER_SECRET:-prod-core-lagoon-core-broker}"
        REMOTE_SECRET="${REMOTE_SECRET:-prod-lagoon-remote-lagoon-remote}"
        REMOTE_DEPLOYMENT="${REMOTE_DEPLOYMENT:-prod-lagoon-remote-lagoon-remote-kubernetes-build-deploy}"
        ;;
    multi-nonprod)
        # Multi-cluster non-production defaults
        KUBE_CONTEXT="${KUBE_CONTEXT:-kind-lagoon-nonprod}"
        LAGOON_NAMESPACE="${LAGOON_NAMESPACE:-lagoon-remote}"
        KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-lagoon-nonprod}"
        # Note: nonprod doesn't have core services, only remote
        KEYCLOAK_SVC="${KEYCLOAK_SVC:-}"
        API_SVC="${API_SVC:-}"
        KEYCLOAK_SECRET="${KEYCLOAK_SECRET:-}"
        BROKER_SECRET="${BROKER_SECRET:-}"
        REMOTE_SECRET="${REMOTE_SECRET:-nonprod-lagoon-remote-lagoon-remote}"
        REMOTE_DEPLOYMENT="${REMOTE_DEPLOYMENT:-nonprod-lagoon-remote-lagoon-remote-kubernetes-build-deploy}"
        ;;
    *)
        echo "Unknown LAGOON_PRESET: $LAGOON_PRESET" >&2
        echo "Valid presets: single, multi-prod, multi-nonprod" >&2
        exit 1
        ;;
esac

# Export all variables
export KUBE_CONTEXT LAGOON_NAMESPACE KIND_CLUSTER_NAME
export KEYCLOAK_SVC API_SVC KEYCLOAK_SECRET BROKER_SECRET
export REMOTE_SECRET REMOTE_DEPLOYMENT

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_debug() {
    if [ "${DEBUG:-}" = "1" ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Check if kubectl can connect to cluster
check_cluster_connectivity() {
    if ! kubectl --context "$KUBE_CONTEXT" cluster-info >/dev/null 2>&1; then
        log_error "Cannot connect to cluster '$KUBE_CONTEXT'"
        return 1
    fi
    return 0
}

# Check if port-forwards are running
check_port_forwards() {
    if curl -s --connect-timeout 2 "http://localhost:8080/auth/" >/dev/null 2>&1 && \
       curl -s --connect-timeout 2 "http://localhost:7080/" >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

# Get a value from a Kubernetes secret
get_secret_value() {
    local secret_name="$1"
    local key="$2"
    kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get secret "$secret_name" \
        -o jsonpath="{.data.$key}" 2>/dev/null | base64 -d 2>/dev/null
}

# Print current configuration (for debugging)
print_config() {
    echo "Current configuration:"
    echo "  KUBE_CONTEXT:      $KUBE_CONTEXT"
    echo "  LAGOON_NAMESPACE:  $LAGOON_NAMESPACE"
    echo "  KIND_CLUSTER_NAME: $KIND_CLUSTER_NAME"
    echo "  KEYCLOAK_SVC:      $KEYCLOAK_SVC"
    echo "  API_SVC:           $API_SVC"
    echo "  KEYCLOAK_SECRET:   $KEYCLOAK_SECRET"
    echo "  BROKER_SECRET:     $BROKER_SECRET"
    echo "  REMOTE_SECRET:     $REMOTE_SECRET"
    echo "  REMOTE_DEPLOYMENT: $REMOTE_DEPLOYMENT"
}
