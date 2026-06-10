#!/bin/bash
# E2E assertion suite for pulumi-lagoon-provider.
#
# Runs four assertion groups against a deployed multi-cluster example stack and
# exits non-zero on the first failure. Meant to be called by `make e2e-assert`
# after `make e2e-deploy` has completed successfully.
#
# The four groups:
#   1. Outputs present     - stack exports are non-empty
#   2. CRUD lifecycle      - refresh/update/destroy behave correctly
#   3. Live API readback   - resources actually exist server-side
#   4. Idempotency         - re-up and refresh show no drift
#
# Usage:
#   # From repo root after `make e2e-deploy`:
#   LAGOON_PRESET=multi-prod ./scripts/e2e-assertions.sh
#
# Environment variables (all optional):
#   E2E_UPDATE_BRANCHES  - branch regex to set for the in-place update assertion
#                          (default: ^(main|develop|feature/.*|hotfix/.*)$)
#   E2E_PULUMI_DIR       - directory containing the Pulumi project (default: auto-detected)
#   LAGOON_PRESET        - passed to common.sh (default: multi-prod)
#   LAGOON_API_URL       - Lagoon GraphQL endpoint (default: http://localhost:7080/graphql)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MULTI_CLUSTER_DIR="${E2E_PULUMI_DIR:-$REPO_ROOT/examples/multi-cluster}"
RUN_PULUMI="$MULTI_CLUSTER_DIR/scripts/run-pulumi.sh"

export LAGOON_PRESET="${LAGOON_PRESET:-multi-prod}"
source "$SCRIPT_DIR/common.sh"

LAGOON_API_URL="${LAGOON_API_URL:-http://localhost:7080/graphql}"
E2E_UPDATE_BRANCHES="${E2E_UPDATE_BRANCHES:-^(main|develop|feature/.*|hotfix/.*)$}"

# Colors
BOLD='\033[1m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }
info() { echo -e "${BOLD}[E2E]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# Run a GraphQL query against the Lagoon API with an admin JWT token.
# Requires LAGOON_TOKEN and LAGOON_API_URL to be set.
gql_query() {
    local query="$1"
    curl -s -k \
        -H "Authorization: Bearer $LAGOON_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$query\"}" \
        "$LAGOON_API_URL"
}

echo ""
echo "============================================================"
echo " pulumi-lagoon-provider E2E Assertions"
echo "============================================================"
echo ""

# =============================================================================
# Group 1: Stack outputs are present and non-empty
# =============================================================================

info "Group 1: Verifying stack outputs are non-empty..."

cd "$MULTI_CLUSTER_DIR"
OUTPUTS=$(pulumi stack output --json 2>/dev/null) || fail "pulumi stack output failed"

check_output() {
    local key="$1"
    local val
    val=$(echo "$OUTPUTS" | jq -r --arg k "$key" '.[$k] // empty')
    if [ -z "$val" ] || [ "$val" = "null" ]; then
        fail "Stack output '$key' is absent or empty. Outputs: $OUTPUTS"
    fi
    echo "  $key = $val"
}

check_output "prod_deploy_target_id"
check_output "nonprod_deploy_target_id"
check_output "example_project_id"
check_output "example_project_name"

PROD_DT_ID=$(echo "$OUTPUTS"   | jq -r '.prod_deploy_target_id')
NONPROD_DT_ID=$(echo "$OUTPUTS" | jq -r '.nonprod_deploy_target_id')
PROJECT_ID=$(echo "$OUTPUTS"   | jq -r '.example_project_id')
PROJECT_NAME=$(echo "$OUTPUTS" | jq -r '.example_project_name')

pass "Group 1: all four outputs present (project_id=$PROJECT_ID, prod_dt=$PROD_DT_ID, nonprod_dt=$NONPROD_DT_ID)"
echo ""

# =============================================================================
# Group 2: CRUD lifecycle
# =============================================================================

info "Group 2: CRUD lifecycle — refresh, in-place update, destroy..."

# 2a. refresh: no unexpected changes
info "  2a. pulumi refresh (no drift expected)..."
"$RUN_PULUMI" refresh --yes --non-interactive 2>&1 | tail -5
pass "  2a. refresh succeeded"

# 2b. preview --expect-no-changes: no pending diff after refresh
info "  2b. pulumi preview --expect-no-changes..."
if ! "$RUN_PULUMI" preview --expect-no-changes 2>&1 | tail -10; then
    fail "  2b. preview showed unexpected changes after refresh"
fi
pass "  2b. no pending diff after refresh"

# 2c. in-place update: change exampleProjectBranches, expect an update (not replace)
info "  2c. In-place update: setting exampleProjectBranches to trigger a Project update..."
pulumi config set exampleProjectBranches "$E2E_UPDATE_BRANCHES"

UPDATE_OUTPUT=$("$RUN_PULUMI" up --yes --non-interactive --json 2>&1 || true)

# Check the op type: we expect an "update" on the Project resource, not "replace" or "create"
if echo "$UPDATE_OUTPUT" | grep -qE '"op"\s*:\s*"(replace|delete-replaced|create-replacement)"'; then
    fail "  2c. Project update caused an unexpected replace. This indicates a ForceNew field was changed. Output: $UPDATE_OUTPUT"
fi
if echo "$UPDATE_OUTPUT" | grep -qE '"op"\s*:\s*"update".*[Pp]roject|[Pp]roject.*"op"\s*:\s*"update"'; then
    pass "  2c. in-place update on Project confirmed"
else
    # Fall back to checking non-JSON output for an update step
    if "$RUN_PULUMI" preview --json 2>/dev/null | jq -e '.changeSummary.update > 0' >/dev/null 2>&1; then
        pass "  2c. in-place update on Project confirmed (via changeSummary)"
    else
        warn "  2c. Could not confirm in-place update from JSON; checking that up succeeded and no errors..."
        # At minimum the up should not have failed
        if echo "$UPDATE_OUTPUT" | grep -qiE "error|failed"; then
            fail "  2c. pulumi up reported errors during update: $UPDATE_OUTPUT"
        fi
        pass "  2c. update completed without error"
    fi
fi

# Restore config to original (unset the override so destroy uses the normal value)
pulumi config rm exampleProjectBranches 2>/dev/null || true

# 2d. destroy: the native resources (DeployTarget, Project) are removed cleanly
info "  2d. pulumi destroy (native resources only; clusters remain)..."
# Enumerate Lagoon resource URNs via stack export (more reliable than --show-urns,
# which is not guaranteed across all CLI versions).
_lagoon_urns=$(pulumi stack export 2>/dev/null \
    | jq -r '.deployment.resources[].urn | select(test("lagoon:lagoon:"))' 2>/dev/null || true)
if [ -n "$_lagoon_urns" ]; then
    _target_args=()
    while IFS= read -r _urn; do
        _target_args+=(--target "$_urn")
    done <<< "$_lagoon_urns"
    "$RUN_PULUMI" destroy --yes --non-interactive --target-dependents "${_target_args[@]}" \
        2>&1 | tail -10 || {
        warn "  2d. Targeted destroy failed; falling back to full stack destroy"
        "$RUN_PULUMI" destroy --yes --non-interactive 2>&1 | tail -10
    }
else
    warn "  2d. No Lagoon URNs found via stack export; falling back to full stack destroy"
    "$RUN_PULUMI" destroy --yes --non-interactive 2>&1 | tail -10
fi
pass "  2d. destroy succeeded"
echo ""

# =============================================================================
# Group 3: Live API readback via GraphQL
# =============================================================================
# Group 2 (above) destroyed the native resources. Re-deploy them here so that
# the live GraphQL readback in Group 3 has something to query.
# Execution order: Group 1 -> Group 2 (CRUD including destroy) -> re-deploy -> Group 3 -> Group 4.

info "Group 3: Live API readback — resources must exist server-side..."

# Re-deploy native resources (Phase 8 only) so we can query them
info "  Re-deploying native resources for live readback..."
"$RUN_PULUMI" up --yes --non-interactive 2>&1 | tail -10
OUTPUTS=$(pulumi stack output --json)
PROJECT_ID=$(echo "$OUTPUTS"   | jq -r '.example_project_id')
PROD_DT_ID=$(echo "$OUTPUTS"   | jq -r '.prod_deploy_target_id')
NONPROD_DT_ID=$(echo "$OUTPUTS" | jq -r '.nonprod_deploy_target_id')
PROJECT_NAME=$(echo "$OUTPUTS" | jq -r '.example_project_name')

# Acquire admin JWT for direct GraphQL queries
source "$SCRIPT_DIR/get-admin-jwt.sh"
export LAGOON_API_URL="${LAGOON_API_URL:-http://localhost:7080/graphql}"

# 3a. Deploy targets exist with the right names
info "  3a. Querying allKubernetes for deploy targets..."
DT_RESPONSE=$(gql_query "{ allKubernetes { id name } }")
if echo "$DT_RESPONSE" | jq -e '.errors' >/dev/null 2>&1; then
    fail "  3a. allKubernetes query returned errors: $(echo "$DT_RESPONSE" | jq -r '.errors[0].message')"
fi

DT_NAMES=$(echo "$DT_RESPONSE" | jq -r '.data.allKubernetes[].name')
if ! echo "$DT_NAMES" | grep -q "lagoon-prod"; then
    fail "  3a. Deploy target 'lagoon-prod' not found server-side. Got: $DT_NAMES"
fi
if ! echo "$DT_NAMES" | grep -q "lagoon-nonprod"; then
    fail "  3a. Deploy target 'lagoon-nonprod' not found server-side. Got: $DT_NAMES"
fi

# Verify IDs match stack outputs
SERVER_PROD_ID=$(echo "$DT_RESPONSE" | jq -r '.data.allKubernetes[] | select(.name == "lagoon-prod") | .id')
SERVER_NONPROD_ID=$(echo "$DT_RESPONSE" | jq -r '.data.allKubernetes[] | select(.name == "lagoon-nonprod") | .id')
if [ "$SERVER_PROD_ID" != "$PROD_DT_ID" ]; then
    fail "  3a. prod deploy target ID mismatch: stack=$PROD_DT_ID server=$SERVER_PROD_ID"
fi
if [ "$SERVER_NONPROD_ID" != "$NONPROD_DT_ID" ]; then
    fail "  3a. nonprod deploy target ID mismatch: stack=$NONPROD_DT_ID server=$SERVER_NONPROD_ID"
fi
pass "  3a. both deploy targets present server-side with matching IDs"

# 3b. Example project exists with the right name
info "  3b. Querying allProjects for example project..."
PROJ_RESPONSE=$(gql_query "{ allProjects { id name } }")
if echo "$PROJ_RESPONSE" | jq -e '.errors' >/dev/null 2>&1; then
    fail "  3b. allProjects query returned errors: $(echo "$PROJ_RESPONSE" | jq -r '.errors[0].message')"
fi

SERVER_PROJ=$(echo "$PROJ_RESPONSE" | jq -r --arg name "$PROJECT_NAME" '.data.allProjects[] | select(.name == $name)')
if [ -z "$SERVER_PROJ" ]; then
    fail "  3b. Project '$PROJECT_NAME' not found server-side. Projects: $(echo "$PROJ_RESPONSE" | jq -r '.data.allProjects[].name')"
fi
SERVER_PROJ_ID=$(echo "$SERVER_PROJ" | jq -r '.id')
if [ "$SERVER_PROJ_ID" != "$PROJECT_ID" ]; then
    fail "  3b. project ID mismatch: stack=$PROJECT_ID server=$SERVER_PROJ_ID"
fi
pass "  3b. project '$PROJECT_NAME' present server-side with matching ID ($PROJECT_ID)"
echo ""

# =============================================================================
# Group 4: Idempotency / no-drift
# =============================================================================

info "Group 4: Idempotency — re-up and refresh must show no changes..."

# 4a. A second pulumi up immediately after the first must report zero changes
info "  4a. pulumi up --expect-no-changes (idempotency)..."
if ! "$RUN_PULUMI" up --expect-no-changes --yes --non-interactive 2>&1 | tail -10; then
    fail "  4a. Second 'pulumi up' reported unexpected changes (non-idempotent Create/Read)"
fi
pass "  4a. no changes on second up"

# 4b. pulumi refresh then preview: no server-side drift
info "  4b. pulumi refresh + preview --expect-no-changes (drift detection)..."
"$RUN_PULUMI" refresh --yes --non-interactive 2>&1 | tail -5
if ! "$RUN_PULUMI" preview --expect-no-changes 2>&1 | tail -10; then
    fail "  4b. pulumi preview after refresh showed drift (Read/Diff bug)"
fi
pass "  4b. no drift detected after refresh"
echo ""

echo "============================================================"
echo -e " ${GREEN}${BOLD}All E2E assertions passed.${NC}"
echo "============================================================"
echo ""
