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
#   - PyJWT installed into examples/multi-cluster/venv (via `make e2e-setup` or
#     `pip install -r examples/multi-cluster/requirements.txt`); this script
#     resolves that venv's python3 directly rather than relying on $PATH
#
# Environment variables (all optional, sensible defaults for multi-cluster prod):
#   KUBE_CONTEXT    - kubectl context (default: kind-lagoon-prod)
#   LAGOON_NAMESPACE - namespace (default: lagoon-core)
#   CORE_SECRETS    - secret name holding JWTSECRET (default: prod-core-lagoon-core-secrets)
#   JWT_AUDIENCE    - JWT aud claim (default: api.dev)
#   JWT_VALIDITY    - token validity in seconds (default: 3600)
#   LAGOON_API_URL  - if set, skip overriding; if unset, set to http://localhost:7080/graphql

set -e

_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
_REPO_ROOT="$(cd "$_SCRIPT_DIR/.." && pwd)"

# Resolve the venv's Python directly rather than relying on `python3` from
# $PATH: e2e-assert's caller cd's into examples/multi-cluster before sourcing
# this script, but a bare `python3` still resolves through the invoking
# shell's PATH, not the venv, so PyJWT (installed only into the venv by
# `make e2e-setup`) was silently unavailable. Check the same candidate
# locations run-pulumi.sh already searches for the venv.
_PYTHON="python3"
for _venv in "./venv" "$_REPO_ROOT/examples/multi-cluster/venv" "$_REPO_ROOT/venv"; do
    if [ -x "$_venv/bin/python3" ]; then
        _PYTHON="$_venv/bin/python3"
        break
    fi
done

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

# Generate the HS256 token via Python; secret is passed via an env var to avoid
# shell-escaping issues, disk writes, and the SC2259 pipe-vs-heredoc conflict.
# (A heredoc overrides piped stdin, so env var is the correct approach here.)
# PyJWT must be available in the interpreter resolved above.
_token=$(_JWTSECRET="$_jwt_secret" "$_PYTHON" - <<EOF
import jwt, time, os
secret = os.environ['_JWTSECRET']
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
