# Test Cluster - Testing Notes

**Date**: 2025-10-22
**Tester**: Claude Code
**Status**: Partial deployment successful, timing issues identified

## Summary

Tested the setup instructions in `README.md` step-by-step. The basic infrastructure (kind cluster creation) worked perfectly, but there are dependency/timing issues in the Pulumi program that need to be resolved.

## What Worked ✅

1. **Prerequisites check** - All tools detected correctly:
   - Docker 28.5.1 ✅
   - kind 0.30.0 ✅
   - kubectl v1.34.1 ✅
   - Pulumi v3.203.0 ✅
   - Python 3.14.0 ✅

2. **Setup script** (`./scripts/setup.sh`):
   - Created venv successfully ✅
   - Installed Python dependencies ✅
   - Initialized Pulumi stack `dev` ✅

3. **/etc/hosts configuration**:
   - Successfully added entries ✅
   - Verified entries present ✅

4. **kind cluster creation**:
   - 3-node cluster created successfully ✅
   - Control plane: `lagoon-test-control-plane` ✅
   - Worker 1: `lagoon-test-worker` ✅
   - Worker 2: `lagoon-test-worker2` ✅
   - All nodes showing `Ready` status ✅
   - Took ~53 seconds to create ✅

## Issues Found 🐛

### Issue #1: Dependency Race Conditions

**Problem**: Commands that depend on cluster resources run before those resources are ready.

**Examples**:
1. `wait-for-ingress-nginx` runs before `ingress-nginx` chart is installed
   - Error: `error: no matching resources found`
   - Cause: Command depends on `ingress_nginx` but ingress isn't installed yet

2. `wait-for-lagoon-api` ran before cluster context exists
   - Error: `error: context "kind-lagoon-test" does not exist`
   - Cause: Command runs in parallel with `create-kind-cluster`

3. `selfsigned-issuer` ClusterIssuer created before cert-manager CRDs installed
   - Error: `no matches for kind "ClusterIssuer" in version "cert-manager.io/v1"`
   - Cause: CustomResource depends on `cert_manager` but CRDs take time to register

**Root Cause**:
- Pulumi's dependency system doesn't wait for Helm charts to be fully deployed
- Command resources (`command.local.Command`) run immediately upon dependency completion
- Kubernetes resources take time to become available after Helm chart "completes"

**Attempted Fix**:
Added `get_kubeconfig` to dependencies of wait commands, but this isn't sufficient. The issue is that Helm charts report "complete" before pods are actually running.

### Issue #2: Wait Commands Should Be Optional

**Problem**: Wait commands fail the entire deployment if resources aren't ready immediately.

**Impact**:
- Deployment stops even though resources will eventually become ready
- User must manually run `pulumi up` multiple times
- Poor user experience

**Recommendation**:
- Make wait commands use `|| true` to not fail deployment
- Or remove wait commands entirely and let Pulumi's built-in retry logic handle it
- Or add explicit `sleep` delays (not ideal but works)

## Workaround

The cluster infrastructure is created successfully. User can:

1. Wait a few minutes for resources to initialize
2. Run `pulumi up --yes` again
3. Pulumi will continue from where it left off
4. May need to run 2-3 times for full deployment

## Recommendations

### Short-term Fixes

1. **Remove problematic wait commands**:
   ```python
   # Remove or make optional:
   # - wait-for-ingress-nginx
   # - wait-for-lagoon-api
   ```

2. **Add retries to CustomResources**:
   ```python
   opts=pulumi.ResourceOptions(
       provider=k8s_provider,
       depends_on=[cert_manager],
       ignore_changes=["*"],  # Allow manual fixes
       retry_on_failure=True,
   )
   ```

3. **Add explicit delays** (not elegant but works):
   ```python
   wait_30s = command.local.Command(
       "wait-30s-for-cert-manager",
       create="sleep 30",
       opts=pulumi.ResourceOptions(depends_on=[cert_manager]),
   )
   ```

### Long-term Solutions

1. **Use Pulumi's `kubernetes.yaml.ConfigFile`** instead of raw CustomResources
   - Better handling of CRD timing

2. **Implement proper readiness checks**:
   - Check for CRD existence before creating CustomResources
   - Check for deployment rollout status before dependent resources

3. **Split into multiple Pulumi stacks**:
   - Stack 1: Infrastructure (kind, ingress, cert-manager, metrics-server)
   - Stack 2: Applications (Harbor, Lagoon)
   - Use stack references for dependencies

4. **Add health checks between stages**:
   ```python
   # Wait for cert-manager webhook to be ready
   cert_manager_ready = command.local.Command(
       "wait-cert-manager-webhook",
       create="kubectl wait --for=condition=available --timeout=300s deployment/cert-manager-webhook -n cert-manager",
       opts=pulumi.ResourceOptions(depends_on=[cert_manager]),
   )
   ```

## Test Cluster Status

As of test completion:
- ✅ kind cluster: Running (3 nodes)
- ✅ Kubernetes system pods: Running
- ❌ ingress-nginx: Not verified
- ❌ cert-manager: Partially deployed (CRDs may be pending)
- ❌ metrics-server: Not deployed yet (just added)
- ❌ Harbor: Not deployed
- ❌ Lagoon: Not deployed

## Manual Verification Commands

```bash
# Check cluster
kind get clusters
kubectl --context kind-lagoon-test get nodes

# Check what's deployed
kubectl --context kind-lagoon-test get pods -A

# Check Helm releases
helm list -A --kube-context kind-lagoon-test

# Continue deployment
cd test-cluster
source venv/bin/activate
pulumi up --yes
```

## Updated Components

1. **Fixed dependency**: Added `get_kubeconfig` to wait commands
2. **Added metrics-server**: Now included in deployment
3. **Three-node cluster**: Successfully validates multi-node configuration

## Next Steps

1. Fix dependency/timing issues in `__main__.py`
2. Test full deployment end-to-end
3. Verify all services are accessible
4. Test provider against deployed Lagoon
5. Document any additional issues
6. Update README with clearer expectations about timing

## Conclusion

The **infrastructure setup works correctly** - kind cluster creation, Python environment, prerequisites, etc. are all solid.

The **Pulumi program has timing issues** that prevent one-shot deployment. This is fixable but requires refactoring the dependency chain.

**Recommendation**: Document the current behavior (may need 2-3 `pulumi up` runs) or implement one of the fixes above before considering this production-ready.

---

**Files modified during testing**:
- `test-cluster/__main__.py` - Added `get_kubeconfig` dependencies, added metrics-server
- `test-cluster/README.md` - Updated to mention metrics-server
- `/etc/hosts` - Added lagoon.test entries

**Commands run**:
```bash
cd test-cluster
./scripts/setup.sh          # ✅ Worked
# Added to /etc/hosts         # ✅ Worked
source venv/bin/activate
pulumi preview              # ✅ Worked
pulumi up --yes            # ⚠️  Partial (timing issues)
```
