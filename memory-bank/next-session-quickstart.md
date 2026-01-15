# Next Session Quickstart - Pulumi Lagoon Provider

**Date Updated**: 2026-01-13
**Status**: Phase 1 Complete - Full End-to-End Testing Passed

---

## TL;DR - Quick Commands

```bash
# Full setup from scratch (~5 minutes)
make setup-all

# Deploy example project
make example-up

# Check status
make cluster-status

# Full teardown
make clean-all
```

---

## Current State

### What's Working
- ✅ `make setup-all` - Complete automated setup
- ✅ `make example-up` - Deploys 7 Lagoon resources
- ✅ `make clean-all` - Full teardown
- ✅ Token expiration handled automatically
- ✅ Keycloak user created automatically
- ✅ Deploy target created automatically
- ✅ Direct Access Grants enabled automatically

### Project Structure
```
pulumi-lagoon-provider/
├── Makefile                    # Main automation
├── pulumi_lagoon/              # Provider package
├── examples/simple-project/    # Example with helper scripts
│   ├── scripts/
│   │   ├── run-pulumi.sh       # Token-refreshing wrapper
│   │   ├── create-lagoon-admin.sh
│   │   └── ensure-deploy-target.sh
│   └── Makefile
├── test-cluster/               # Kind + Lagoon Pulumi program
└── memory-bank/                # Documentation
```

---

## Common Tasks

### Start Fresh Development Environment
```bash
make setup-all      # Creates everything (~5 min)
make example-up     # Deploys example
```

### Destroy and Recreate
```bash
make clean-all      # Destroys everything
make setup-all      # Recreates
```

### Just Work on Provider Code
```bash
# If cluster already running:
make provider-install   # Reinstall after code changes
make example-preview    # Test changes
```

### Debug Issues
```bash
make cluster-status     # Check pod status
kubectl --context kind-lagoon-test logs -n lagoon <pod>
```

---

## Makefile Targets Reference

### Setup
```bash
make setup-all          # Complete setup: venv, provider, Kind, Lagoon, user, deploy target
make venv               # Create Python virtual environment
make provider-install   # Install provider in development mode
make cluster-up         # Create Kind cluster + deploy Lagoon
make cluster-down       # Destroy Kind cluster
```

### Lagoon Setup (called automatically by setup-all)
```bash
make ensure-lagoon-admin   # Create lagoonadmin user in Keycloak
make ensure-deploy-target  # Create deploy target + set Pulumi config
```

### Example Project
```bash
make example-setup      # Initialize Pulumi stack
make example-preview    # Preview changes (auto token refresh)
make example-up         # Deploy resources (auto token refresh)
make example-down       # Destroy resources
make example-output     # Show stack outputs
```

### Cleanup
```bash
make clean              # Kill port-forwards, remove temp files
make clean-all          # Full cleanup: clean + destroy cluster + remove venvs
```

---

## Key Files Changed This Session (2026-01-13)

### New Scripts
- `examples/simple-project/scripts/create-lagoon-admin.sh` - Creates Keycloak user
- `examples/simple-project/scripts/ensure-deploy-target.sh` - Creates deploy target

### Modified Files
- `Makefile` - Added `ensure-lagoon-admin`, `ensure-deploy-target`, fixed `clean`
- `examples/simple-project/scripts/quickstart.sh` - Auto user creation
- `examples/simple-project/scripts/run-pulumi.sh` - Auto user creation
- `README.md` - Updated Makefile documentation

---

## Known Issues & Fixes

| Issue | Cause | Fix |
|-------|-------|-----|
| lagoonadmin user missing | Helm chart doesn't create it | `create-lagoon-admin.sh` (auto-called) |
| "Client not allowed for direct access grants" | Keycloak config | Auto-enabled by scripts |
| Token expired | 5-min expiration | Use `run-pulumi.sh` wrapper |
| `make clean` kills itself | pkill pattern too broad | Fixed with `[k]ubectl` pattern |

---

## Token Handling

All Makefile targets get fresh tokens automatically:

| Target | Token Type | Handled By |
|--------|-----------|------------|
| `ensure-lagoon-admin` | Keycloak admin | `create-lagoon-admin.sh` |
| `ensure-deploy-target` | Lagoon OAuth | `ensure-deploy-target.sh` |
| `example-preview/up` | Lagoon OAuth | `run-pulumi.sh` |

---

## Resources Created by Example

| Resource | ID | Type |
|----------|-----|------|
| Project | 1 | example-drupal-site |
| Environment | 1 | develop (development) |
| Environment | 2 | main (production) |
| Variable | 1 | API_BASE_URL (project) |
| Variable | 2 | DATABASE_HOST (production) |
| Variable | 3 | DATABASE_HOST (development) |

---

## Manual Access (if needed)

### Get Token Manually
```bash
cd examples/simple-project
source ./scripts/get-lagoon-token.sh
echo "Token: ${LAGOON_TOKEN:0:20}..."
```

### Credentials
```bash
# Lagoon admin password
kubectl --context kind-lagoon-test -n lagoon get secret lagoon-core-keycloak \
  -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d

# Keycloak admin password
kubectl --context kind-lagoon-test -n lagoon get secret lagoon-core-keycloak \
  -o jsonpath='{.data.KEYCLOAK_ADMIN_PASSWORD}' | base64 -d
```

---

## Documentation Files

- **Session Summary**: `memory-bank/session-summary-2026-01-13.md`
- **Implementation Status**: `memory-bank/implementation-status.md`
- **Example README**: `examples/simple-project/README.md`
- **Main README**: `README.md`

---

**Summary**: Phase 1 complete. Use `make setup-all` for fresh setup, `make example-up` to deploy. All automation working.
