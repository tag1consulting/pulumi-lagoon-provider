# Next Session Quickstart - Native Go Pulumi Lagoon Provider

**Date Updated**: 2026-03-24
**Status**: v0.2.0 RELEASE IN PROGRESS — PR #39 under review

---

## CURRENT STATE

**Development Branch**: `native-go-provider` (merged to `develop` via PR #37)
**Release PR**: https://github.com/tag1consulting/pulumi-lagoon-provider/pull/39 (`develop` → `main`)

### What's Done
- All 11 resources implemented and compiling
- 198 unit tests passing across 3 packages (client: 118, config: 13, resources: 67)
- SDK generation complete (Python, TypeScript, Go)
- GitHub Actions workflow for Go tests (`.github/workflows/test-go.yml`)
- PR #37 merged to `develop`; release PR #39 (`develop` → `main`) under review
- Integration testing:
  - single-cluster: TESTED (Create, Read)
  - simple-project: TESTED (full CRUD)
  - multi-cluster: TESTED (Create, Read, token rotation replacement cycle all verified)
- Token rotation resilience: create-or-update, idempotent deletes, graceful read-not-found
- Lagoon update mutation input format fixed (`{id, patch: {...}}` structure)

### Multi-Cluster Environment Status
- Two Kind clusters may still be running: `kind-lagoon-prod` + `kind-lagoon-nonprod`
- Check with `kind get clusters`

### What To Do Next

1. **If multi-cluster is still running, tear down**:
   ```bash
   cd examples/multi-cluster && make clean-all
   ```

2. **Regenerate SDKs** (provider code changed significantly):
   ```bash
   make go-schema && make go-sdk-all
   ```

3. **TypeScript SDK** — TESTED (2026-03-24, full CRUD verified)
4. **Go SDK** — TESTED (2026-03-24, full CRUD verified)
5. **Release PR #39** (`develop` → `main`) — under review

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

### Plugin Install (must rm first!)
```bash
pulumi plugin rm resource lagoon 0.2.0-dev --yes
pulumi plugin install resource lagoon 0.2.0-dev --file provider/bin/pulumi-resource-lagoon
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

### Lagoon Update Mutation Input Format
All update mutations require `{id, patch: {...fields...}}` — NOT flat fields:
```go
c.Execute(ctx, mutation, map[string]any{
    "input": map[string]any{
        "id":    id,
        "patch": fieldsMap,
    },
})
```

### Create-or-Update Pattern
DeployTarget, Project, DeployTargetConfig all implement:
1. Attempt create → if `IsDuplicateEntry(err)` → look up existing → update
2. Strip immutable fields (e.g., `name`) before calling update
3. Read returns empty response on not-found (signals Pulumi to remove from state)
4. Delete ignores not-found/API errors (idempotent)

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
