# Next Session Quickstart - Pulumi Lagoon Provider

**Date Updated**: 2026-01-20
**Status**: Phase 2 In Progress - Multi-Cluster Example Debugged

---

## 🚨 CONTINUE FROM PREVIOUS SESSION

**Branch**: `deploytarget-multi-cluster`
**PR**: https://github.com/tag1consulting/pulumi-lagoon-provider/pull/10 (Draft)

### What to do first:

1. Check if Kind clusters still exist:
   ```bash
   kind get clusters
   ```

2. If clusters exist, check pod status:
   ```bash
   kubectl --context kind-lagoon-prod get pods -n lagoon-core
   kubectl --context kind-lagoon-prod get pods -n harbor
   kubectl --context kind-lagoon-prod get pods -n lagoon
   kubectl --context kind-lagoon-nonprod get pods -n lagoon
   ```

3. If clusters don't exist or need fresh start:
   ```bash
   make multi-cluster-down  # Clean up any remnants
   make multi-cluster-up    # Deploy fresh
   ```

### Issues Fixed (2026-01-20)

1. **RabbitMQ CrashLoopBackOff**: Fixed by deleting PVCs to clear corrupted Mnesia data
   - Root cause: Mnesia table sync timeout between broker pods
   - Fix: Scale StatefulSet to 0, delete PVCs, scale back up, delete pods to force recreation

2. **Service Selector Bug**: Fixed in `lagoon/core.py`
   - Root cause: `create_rabbitmq_nodeport_service()` used selector `app.kubernetes.io/component: broker`
   - Fix: Changed to `app.kubernetes.io/component: {release_name}-lagoon-core-broker`

3. **Cross-cluster RabbitMQ IP**: The nonprod remote had wrong IP
   - Root cause: Pulumi state had stale IP from initial deployment
   - Code fix: Added dynamic IP refresh using container ID triggers

4. **Keycloak Direct Access Grants**: OAuth password grant wasn't working
   - Root cause: The `lagoon-ui` Keycloak client doesn't have Direct Access Grants enabled by default
   - Fix: Added Pulumi Job (`lagoon/keycloak.py`) that auto-configures Keycloak after install
   - Also creates `lagoonadmin` user with `platform-owner` role

---

## TL;DR - Quick Commands

```bash
# Multi-cluster example (current work)
make multi-cluster-up       # Deploy prod + nonprod clusters
make multi-cluster-down     # Tear down everything
make multi-cluster-status   # Check outputs

# Simple example (original, still works)
make setup-all              # Complete setup (~5 min)
make example-up             # Deploy example
make clean-all              # Full teardown
```

---

## Current State

### What's Working
- ✅ `LagoonDeployTarget` resource implemented with validators
- ✅ Multi-cluster example code complete
- ✅ Kind clusters created successfully
- ✅ Harbor registry deploys successfully
- ✅ Lagoon core running (RabbitMQ fixed)
- ✅ Cross-cluster RabbitMQ communication working
- ⏳ Browser-based authentication (needs testing)

### Project Structure
```
pulumi-lagoon-provider/
├── Makefile                    # Main automation (includes multi-cluster targets)
├── pulumi_lagoon/              # Provider package
│   ├── deploytarget.py         # NEW: LagoonDeployTarget resource
│   ├── validators.py           # Updated with deploy target validators
│   └── client.py               # Updated with Kubernetes GraphQL ops
├── examples/
│   ├── simple-project/         # Original example (working)
│   └── multi-cluster/          # NEW: Prod/nonprod clusters
│       ├── __main__.py         # Main orchestration
│       ├── clusters/           # Kind cluster creation
│       ├── infrastructure/     # Ingress, cert-manager, CoreDNS
│       ├── registry/           # Harbor installation
│       └── lagoon/             # Lagoon core + remote
├── test-cluster/               # Kind + Lagoon Pulumi program
└── memory-bank/                # Documentation
```

---

## Multi-Cluster Example Details

### Architecture
```
+---------------------------+     +---------------------------+
|    lagoon-prod cluster    |     |  lagoon-nonprod cluster   |
|---------------------------|     |---------------------------|
| lagoon-core namespace:    |     |                           |
|   - API, UI, Keycloak     |     |                           |
|   - RabbitMQ (broker)     |<----+-- lagoon namespace:       |
|   - SSH, webhooks         |     |     - remote-controller   |
|                           |     |       (nonprod builds)    |
| harbor namespace:         |     |                           |
|   - Harbor Registry       |     |                           |
|                           |     |                           |
| lagoon namespace:         |     |                           |
|   - remote-controller     |     |                           |
|     (prod builds)         |     |                           |
+---------------------------+     +---------------------------+
```

### Key Technical Details

| Component | Details |
|-----------|---------|
| Cross-cluster RabbitMQ | NodePort 30672 (custom service, chart doesn't support fixed NodePort) |
| Keycloak internal URL | `http://{release}-lagoon-core-keycloak.{ns}.svc.cluster.local:8080/auth` |
| Service naming | `{release}-lagoon-core-{component}` (e.g., `prod-core-lagoon-core-api`) |
| lagoon-build-deploy | v0.39.0 (required for K8s 1.22+ CRD compatibility) |

### Accessing Services (Port Forwarding)

```bash
# Start port forwards
kubectl --context kind-lagoon-prod port-forward -n lagoon-core svc/prod-core-lagoon-core-ui 3000:3000
kubectl --context kind-lagoon-prod port-forward -n lagoon-core svc/prod-core-lagoon-core-api 4000:80
kubectl --context kind-lagoon-prod port-forward -n lagoon-core svc/prod-core-lagoon-core-keycloak 8080:8080
```

**Important**: For browser auth, add to `/etc/hosts`:
```
127.0.0.1 prod-core-lagoon-core-keycloak.lagoon-core.svc.cluster.local
```

---

## Makefile Targets Reference

### Multi-Cluster Example (NEW)
```bash
make multi-cluster-up       # Create prod + nonprod Kind clusters with Lagoon
make multi-cluster-down     # Destroy multi-cluster environment
make multi-cluster-preview  # Preview changes
make multi-cluster-status   # Show stack outputs
make multi-cluster-clusters # List Kind clusters
```

### Simple Example (Original)
```bash
make setup-all              # Complete setup: venv, provider, Kind, Lagoon
make example-up             # Deploy example resources
make example-down           # Destroy resources
make clean-all              # Full cleanup
```

### Development
```bash
make provider-install       # Reinstall provider after code changes
pytest tests/unit/ -v       # Run unit tests
```

---

## Debugging Commands

```bash
# Check pod status
kubectl --context kind-lagoon-prod get pods -n lagoon-core
kubectl --context kind-lagoon-prod get pods -n harbor
kubectl --context kind-lagoon-prod get pods -n lagoon
kubectl --context kind-lagoon-nonprod get pods -n lagoon

# View logs
kubectl --context kind-lagoon-prod logs -n lagoon-core -l app.kubernetes.io/component=api --tail=50
kubectl --context kind-lagoon-prod logs -n lagoon-core -l app.kubernetes.io/component=keycloak --tail=50
kubectl --context kind-lagoon-prod logs -n lagoon -l app.kubernetes.io/name=lagoon-build-deploy --tail=50

# Check services
kubectl --context kind-lagoon-prod get svc -n lagoon-core
kubectl --context kind-lagoon-prod get svc -n lagoon-core | grep broker

# Check cross-cluster connectivity
kubectl --context kind-lagoon-nonprod exec -it -n lagoon <pod> -- nc -zv <prod-node-ip> 30672
```

---

## Helm Chart Versions

| Chart | Version | Notes |
|-------|---------|-------|
| ingress-nginx | 4.10.1 | Standard Kubernetes ingress |
| cert-manager | v1.14.4 | TLS certificate management |
| harbor | 1.14.2 | Container registry |
| lagoon-core | 1.0.0 | Lagoon core services |
| lagoon-build-deploy | 0.39.0 | Updated for K8s 1.22+ CRDs |

---

## Known Limitations

1. **Browser Auth**: UI redirects to internal K8s URLs - requires hosts file entry for port-forwarding
2. **Self-Signed Certs**: Browsers show security warnings
3. **S3/MinIO**: Disabled (placeholder config)
4. **Elasticsearch**: Disabled (placeholder config)

---

## Documentation Files

- **Session Summary (2026-01-16)**: `memory-bank/session-summary-2026-01-16.md`
- **Multi-Cluster README**: `examples/multi-cluster/README.md`
- **Implementation Status**: `memory-bank/implementation-status.md`
- **Main README**: `README.md`

---

## Next Steps

1. ✅ ~~Debug Lagoon core timeout issue~~ (Fixed: RabbitMQ Mnesia data)
2. ✅ ~~Verify cross-cluster RabbitMQ communication~~ (Working)
3. Test browser-based authentication with port-forwarding
4. Run `pulumi up` to apply the service selector fix
5. Mark PR as ready for review once working
6. Consider adding integration tests

---

**Summary**: Multi-cluster infrastructure is operational! RabbitMQ and cross-cluster communication issues were fixed on 2026-01-20. Code fix in `lagoon/core.py` for service selector needs to be applied via `pulumi up`. Branch is `deploytarget-multi-cluster`, PR #10 is open as draft. Next: test browser auth, then mark PR ready for review.
