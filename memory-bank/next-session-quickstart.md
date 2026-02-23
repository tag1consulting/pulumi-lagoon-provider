# Next Session Quickstart - Native Go Pulumi Lagoon Provider

**Date Updated**: 2026-02-23
**Status**: Phase 5 IN PROGRESS - Multi-cluster integration tested, TypeScript/Go SDKs remain

---

## RESUME FROM HERE

**Branch**: `native-go-provider`
**PR**: https://github.com/tag1consulting/pulumi-lagoon-provider/pull/37 (Draft, targeting `develop`)
**Latest Commit**: `912b350` - "Fix multi-cluster Lagoon service names and missing dependencies"

### What's Done
- All 11 resources implemented and compiling
- 191 unit tests passing across 3 packages (client, config, resources)
- SDK generation complete (Python, TypeScript, Go)
- GitHub Actions workflow for Go tests (`.github/workflows/test-go.yml`)
- Draft PR #37 open targeting `develop`
- Integration testing:
  - single-cluster: TESTED (Create, Read)
  - simple-project: TESTED (full CRUD)
  - multi-cluster: TESTED (Create + Read for DeployTarget, Project, DeployTargetConfig)

### Multi-Cluster Environment Status
- Two Kind clusters running: `kind-lagoon-prod` + `kind-lagoon-nonprod`
- Full Lagoon stack deployed (55 resources)
- Build-deploy pods connected to RabbitMQ broker on both clusters
- Port-forwards active (Keycloak :8080, API :7080)

### What To Do Next

1. **If multi-cluster is still running, verify**:
   ```bash
   cd examples/multi-cluster
   kind get clusters
   make check
   ```

2. **Test TypeScript SDK** (not yet done):
   - Create a TypeScript example using the generated SDK in `sdk/nodejs/`

3. **Test Go SDK** (not yet done):
   - Create a Go example using the generated SDK in `sdk/go/`

4. **Known issues to address**:
   - Provider token rotation causes cascading replacements (JWT changes each run)
   - Lagoon overrides project `branches`/`pullrequests` fields when DeployTargetConfig is set

5. **When done testing, tear down**:
   ```bash
   cd examples/multi-cluster && make clean-all
   ```

---

## Quick Reference

### Build & Test Commands
```bash
# Build the provider binary
cd provider && CGO_ENABLED=0 go build -o bin/pulumi-resource-lagoon ./cmd/pulumi-resource-lagoon

# Run all tests
cd provider && CGO_ENABLED=0 go test ./... -v

# Type check only (fast)
cd provider && go vet ./...

# SDK generation
make go-schema && make go-sdk-all
```

### Multi-Cluster Commands
```bash
cd examples/multi-cluster

# Full deployment (~7 min)
make deploy

# Health check
make check

# API test
make test-api

# Port forwards
make port-forwards

# Pulumi operations (auto token refresh)
./scripts/run-pulumi.sh preview
./scripts/run-pulumi.sh refresh --yes

# Teardown
make clean-all
```

### CRITICAL: CGO_ENABLED=0
Both `go build` AND `go test` require `CGO_ENABLED=0` on this system due to Linuxbrew ld vs system gcc conflict.

---

## Key Technical Details

### pulumi-go-provider v1.3.0 API
The v1.3.0 API uses **request/response structs**:
```go
func (r *Resource) Create(ctx context.Context, req infer.CreateRequest[TArgs]) (infer.CreateResponse[TState], error)
func (r *Resource) Read(ctx context.Context, req infer.ReadRequest[TArgs, TState]) (infer.ReadResponse[TArgs, TState], error)
func (r *Resource) Update(ctx context.Context, req infer.UpdateRequest[TArgs, TState]) (infer.UpdateResponse[TState], error)
func (r *Resource) Delete(ctx context.Context, req infer.DeleteRequest[TState]) (infer.DeleteResponse, error)
func (r *Resource) Diff(ctx context.Context, req infer.DiffRequest[TArgs, TState]) (infer.DiffResponse, error)
```

### Multi-Cluster Helm Service Naming
The lagoon-core Helm chart creates services as `{release_name}-lagoon-core-{component}`:
- Broker: `prod-core-lagoon-core-broker`
- SSH: `prod-core-lagoon-core-ssh`
- Keycloak: `prod-core-lagoon-core-keycloak`
- Pod labels: `app.kubernetes.io/component: prod-core-lagoon-core-broker`

### Provider Config Pattern
```go
cfg := infer.GetConfig[config.LagoonConfig](ctx)
c := cfg.NewClient()
```

---

## Git Branches
- `main` - Python dynamic provider (v0.1.2, production)
- `develop` - Integration branch (PR target)
- `native-go-provider` - Native Go provider work (current, PR #37 open)
