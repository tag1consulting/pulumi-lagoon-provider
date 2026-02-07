#!/bin/bash
# Get Lagoon OAuth token via Keycloak Direct Access Grants flow
#
# This script obtains an OAuth access token from Keycloak that can be used
# to authenticate with the Lagoon API and CLI.
#
# Usage: ./get-lagoon-token.sh [options]
#   -u, --user USERNAME    Keycloak username (default: lagoonadmin)
#   -p, --password PASS    Keycloak password (default: lagoonadmin)
#   -q, --quiet            Only output the token (for scripting)
#   -h, --help             Show this help message

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default configuration
CLUSTER_NAME="${LAGOON_CLUSTER_NAME:-lagoon-test}"
CONTEXT="kind-${CLUSTER_NAME}"
# Use HTTPS ingress endpoint (with self-signed cert)
KEYCLOAK_URL="${LAGOON_KEYCLOAK_URL:-https://keycloak.lagoon.test}"
KEYCLOAK_REALM="lagoon"
KEYCLOAK_CLIENT="lagoon-ui"
USERNAME="${LAGOON_USER:-lagoonadmin}"
# Password can be set via env var, or will be fetched from k8s secret
PASSWORD="${LAGOON_PASSWORD:-}"
QUIET=false
# Use -k for insecure mode (self-signed certs)
CURL_OPTS="-k"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--user)
            USERNAME="$2"
            shift 2
            ;;
        -p|--password)
            PASSWORD="$2"
            shift 2
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -u, --user USERNAME    Keycloak username (default: lagoonadmin)"
            echo "  -p, --password PASS    Keycloak password (default: lagoonadmin)"
            echo "  -q, --quiet            Only output the token (for scripting)"
            echo "  -h, --help             Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  LAGOON_CLUSTER_NAME    Kind cluster name (default: lagoon-test)"
            echo "  LAGOON_KEYCLOAK_URL    Keycloak URL (default: https://keycloak.lagoon.test)"
            echo "  LAGOON_USER            Username (default: lagoonadmin)"
            echo "  LAGOON_PASSWORD        Password (default: fetched from k8s secret)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

log() {
    if [ "$QUIET" = false ]; then
        echo "$@" >&2
    fi
}

error() {
    echo "ERROR: $*" >&2
    exit 1
}

# Check if cluster exists
if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    error "Cluster '${CLUSTER_NAME}' not found. Run 'make deploy' in examples/multi-cluster/ first."
fi

# Fetch password from Kubernetes secret if not provided
if [ -z "$PASSWORD" ]; then
    log "Fetching lagoonadmin password from Kubernetes secret..."
    PASSWORD=$(kubectl --context "${CONTEXT}" get secret lagoon-core-keycloak -n lagoon \
        -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' 2>/dev/null | base64 -d) || {
        error "Failed to fetch password from Kubernetes secret.
Make sure the cluster is deployed and the lagoon namespace exists."
    }
    if [ -z "$PASSWORD" ]; then
        error "Password secret is empty. Lagoon may not be fully deployed."
    fi
fi

# Check if Keycloak is accessible
log "Checking Keycloak connectivity at ${KEYCLOAK_URL}..."
if ! curl -sf ${CURL_OPTS} "${KEYCLOAK_URL}/health/ready" >/dev/null 2>&1 && \
   ! curl -sf ${CURL_OPTS} "${KEYCLOAK_URL}/auth/realms/${KEYCLOAK_REALM}" >/dev/null 2>&1; then
    log "Warning: Could not reach Keycloak health endpoint directly."
    log "Attempting token request anyway..."
fi

# Get OAuth token via Direct Access Grants
log "Requesting OAuth token for user '${USERNAME}'..."

TOKEN_ENDPOINT="${KEYCLOAK_URL}/auth/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token"

RESPONSE=$(curl -sf ${CURL_OPTS} -X POST "${TOKEN_ENDPOINT}" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=password" \
    -d "client_id=${KEYCLOAK_CLIENT}" \
    -d "username=${USERNAME}" \
    -d "password=${PASSWORD}" 2>&1) || {
    error "Failed to get token from Keycloak.

Possible causes:
- Keycloak is not ready (wait a few minutes after cluster deployment)
- Keycloak ingress not accessible (check /etc/hosts has keycloak.lagoon.test)
- Invalid username/password

Try:
  kubectl --context ${CONTEXT} -n lagoon get pods | grep keycloak
  curl -kv ${TOKEN_ENDPOINT}"
}

# Extract access token
ACCESS_TOKEN=$(echo "${RESPONSE}" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
    error "Failed to extract access_token from response: ${RESPONSE}"
fi

# Extract token expiration
EXPIRES_IN=$(echo "${RESPONSE}" | grep -o '"expires_in":[0-9]*' | cut -d':' -f2)
if [ -n "$EXPIRES_IN" ] && [ "$QUIET" = false ]; then
    log "Token obtained successfully (expires in ${EXPIRES_IN} seconds)"
fi

# Output the token
echo "${ACCESS_TOKEN}"
