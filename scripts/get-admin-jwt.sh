#!/bin/bash
# Mint an admin JWT token signed with the cluster JWTSECRET.
#
# Usage (standalone):
#   export LAGOON_TOKEN=$(./scripts/get-admin-jwt.sh)
#   export LAGOON_API_URL=http://localhost:7080/graphql
#
# Usage (sourced — also sets LAGOON_API_URL):
#   source ./scripts/get-admin-jwt.sh
#
# Admin JWT tokens are required for deploy-target (Kubernetes/OpenShift) operations.
# Keycloak OAuth tokens lack the necessary permissions for those GraphQL mutations.
#
# Prerequisites:
#   - kubectl configured with the prod context (KUBE_CONTEXT, defaults to kind-lagoon-prod)
#   - The lagoon-core namespace (LAGOON_NAMESPACE, defaults to lagoon-core)
#   - CORE_SECRETS secret (defaults to prod-core-lagoon-core-secrets)
#   - Python 3 with PyJWT available (provided by the example venv)
#
# Environment variables (all optional, sensible defaults for multi-cluster prod):
#   KUBE_CONTEXT    - kubectl context (default: kind-lagoon-prod)
#   LAGOON_NAMESPACE - namespace (default: lagoon-core)
#   CORE_SECRETS    - secret name holding JWTSECRET (default: prod-core-lagoon-core-secrets)
#   JWT_AUDIENCE    - JWT aud claim (default: api.dev)
#   JWT_VALIDITY    - token validity in seconds (default: 3600)
#   LAGOON_API_URL  - if set, skip overriding; if unset, set to http://localhost:7080/graphql

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

_CONTEXT="${KUBE_CONTEXT:-kind-lagoon-prod}"
_NAMESPACE="${LAGOON_NAMESPACE:-lagoon-core}"
_CORE_SECRETS="${CORE_SECRETS:-prod-core-lagoon-core-secrets}"
_AUDIENCE="${JWT_AUDIENCE:-api.dev}"
_VALIDITY="${JWT_VALIDITY:-3600}"

# Resolve color helpers without re-sourcing common.sh if already sourced
_log_info()  { echo -e "\033[0;32m[INFO]\033[0m $1" >&2; }
_log_error() { echo -e "\033[0;31m[ERROR]\033[0m $1" >&2; }

# Get JWTSECRET from the core secrets
_jwt_secret=$(kubectl --context "$_CONTEXT" -n "$_NAMESPACE" get secret "$_CORE_SECRETS" \
    -o jsonpath='{.data.JWTSECRET}' 2>/dev/null | base64 -d 2>/dev/null)

if [ -z "$_jwt_secret" ]; then
    _log_error "Could not read JWTSECRET from $_CONTEXT/$_NAMESPACE/$_CORE_SECRETS"
    exit 1
fi

# Write secret to a temp file to avoid shell-escaping issues with special characters
_secret_file=$(mktemp)
echo "$_jwt_secret" > "$_secret_file"

# Generate the HS256 token via Python; PyJWT must be available in the active venv
_token=$(python3 - <<EOF
import jwt, time, sys

with open('$_secret_file', 'r') as f:
    secret = f.read().strip()

now = int(time.time())
payload = {
    'role': 'admin',
    'iss': 'lagoon-api',
    'sub': 'lagoonadmin',
    'aud': '$_AUDIENCE',
    'iat': now,
    'exp': now + int('$_VALIDITY'),
}
print(jwt.encode(payload, secret, algorithm='HS256'))
EOF
)

rm -f "$_secret_file"

if [ -z "$_token" ] || echo "$_token" | grep -qE "Traceback|Error:|ModuleNotFoundError"; then
    _log_error "Failed to generate admin JWT token"
    echo "$_token" >&2
    exit 1
fi

# When sourced, export into the caller's environment.
# When executed, print the token to stdout (for command substitution).
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
    export LAGOON_TOKEN="$_token"
    export LAGOON_API_URL="${LAGOON_API_URL:-http://localhost:7080/graphql}"
    _log_info "Admin JWT exported (valid for ${_VALIDITY}s)"
else
    printf '%s\n' "$_token"
fi
