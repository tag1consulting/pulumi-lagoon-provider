# Pulumi Lagoon Provider - Implementation Status

**Last Updated**: 2026-03-24
**Status**: Native Go Provider v0.2.0 READY FOR RELEASE — PR #39 in review

---

## Native Go Provider (v0.2.0)

### Overview
Migration from Python dynamic provider (v0.1.2) to native Go provider using `pulumi-go-provider` v1.3.0 with `infer` package and builder pattern.

**Development Branch**: `native-go-provider` (merged to `develop` via PR #37)
**Release PR**: https://github.com/tag1consulting/pulumi-lagoon-provider/pull/39 (`develop` → `main`)

### Phase 1: Scaffolding + GraphQL Client - COMPLETE
- Go module initialized (`go.mod` with Go 1.24, pulumi-go-provider v1.3.0)
- Provider binary entry point (`provider/cmd/pulumi-resource-lagoon/main.go`)
- Centralized config with JWT generation, secret tags (`provider/pkg/config/config.go`)
- Full GraphQL client with retry, token refresh, API version detection (`provider/pkg/client/`)
- All GraphQL queries/mutations ported from Python (`provider/pkg/client/queries.go`, 644 lines)

### Phase 2: Core Resources - COMPLETE
- LagoonProject - full CRUD + Diff + Check + Read(import)
- LagoonEnvironment - depends on Project
- LagoonVariable - dual API support (v2.30.0+ "new" / legacy), secret value

### Phase 3: Remaining Resources + Tests - COMPLETE
- LagoonDeployTarget + LagoonDeployTargetConfig
- All 4 notification resources (Slack, RocketChat, Email, Microsoft Teams)
- LagoonProjectNotification
- LagoonTask (dual API, scope validation)
- 198 unit tests across 3 packages
- GitHub Actions workflow (`test-go.yml`)

### Phase 4: SDK Generation - COMPLETE
- Python SDK: `sdk/python/python/pulumi_lagoon/`
- TypeScript SDK: `sdk/nodejs/nodejs/`
- Go SDK: `sdk/go/go/lagoon/`
- GoReleaser config (`.goreleaser.yml`) for cross-platform builds
- Makefile targets: `make go-build`, `make go-test`, `make go-schema`, `make go-sdk-all`

### Phase 5: Integration Testing - COMPLETE
- **single-cluster**: TESTED (Create, Read verified via `pulumi refresh`)
- **simple-project**: TESTED (full CRUD — Create, Read, Update, Delete all verified)
- **multi-cluster**: TESTED (2026-02-23, re-tested with all fixes)
  - 55 resources deployed (2 Kind clusters + full Lagoon stack)
  - Native provider created: 2 DeployTargets, 1 Project, 2 DeployTargetConfigs
  - Read verified via `pulumi refresh` (all resources read correctly from API)
  - API queries confirmed correct data (deploy target routing, project config)
  - Token rotation tested: `pulumi up` succeeds despite JWT change triggering replacements
  - Fixed: Helm service name mismatch (`{release}-lagoon-core-{component}` not `{release}-{component}`)
  - Fixed: NodePort selector, broker/SSH/Keycloak internal hostnames
  - Fixed: Missing `jwt_secret` in `LagoonSecretsOutputs` dataclass
  - Fixed: Missing `PyJWT` dependency in `requirements.txt`
  - Fixed: Update mutations need `{id, patch: {...}}` input structure (not flat)
  - Fixed: Create-or-update semantics for DeployTarget, Project, DeployTargetConfig
  - Fixed: Idempotent Delete and graceful Read-not-found for state cleanup
- **TypeScript SDK**: TESTED (2026-03-24, full CRUD verified against Kind cluster)
- **Go SDK**: TESTED (2026-03-24, full CRUD verified against Kind cluster)

### Phase 6: CI/CD + Docs + Migration - COMPLETE
- Release workflow (goreleaser + SDK publish) — GoReleaser config, PyPI publish workflow
- Documentation updates (CLAUDE.md, README.md, RELEASE_NOTES.md)
- Migration guide (dynamic v0.1.x -> native v0.2.0) — in RELEASE_NOTES.md

---

## Test Coverage Summary

### Go Tests (Native Provider) - 198 tests

| Package | Tests | Time | Description |
|---------|-------|------|-------------|
| `pkg/client` | 118 | ~5s | GraphQL client, all resource CRUD operations, error helpers |
| `pkg/config` | 13 | <1s | JWT generation, Configure() validation, client factory |
| `pkg/resources` | 67 | <1s | Helper functions, Diff() for all 11 resources |

### Integration Testing (Live Lagoon on Kind)

#### simple-project (16 Lagoon resources)

| Operation | Status | Details |
|-----------|--------|---------|
| **Create** | PASS | All 16 resources created successfully |
| **Read** | PASS | `pulumi refresh` — 16 unchanged (idempotent) |
| **Update** | PASS | Variable value change applied and verified |
| **Delete** | PASS | All 16 resources destroyed cleanly |
| **Re-create** | PASS | Full roundtrip: empty → create → refresh → destroy |

#### multi-cluster (5 native provider resources)

| Operation | Status | Details |
|-----------|--------|---------|
| **Create** | PASS | 2 DeployTargets + 1 Project + 2 DeployTargetConfigs |
| **Read** | PASS | `pulumi refresh` read all resources from API |
| **API verify** | PASS | GraphQL queries confirm correct data |

### Python Tests (Dynamic Provider) - 513 tests
(Still on `main` branch, separate test suite)

---

## Key Issues Encountered & Resolved

### 1. pulumi-go-provider API Migration (v0.25.0 → v1.3.0)
**Problem**: Initial implementation used v0.25.0 plain function signatures; upgraded to v1.3.0 request/response structs.
**Solution**: All 11 resources rewritten to use `infer.CreateRequest[I]` / `infer.CreateResponse[O]` pattern.

### 2. CGO_ENABLED=0 Required
**Problem**: `go build` and `go test` fail with Linuxbrew ld vs system gcc linker errors.
**Solution**: Always use `CGO_ENABLED=0` prefix. `go vet` works without it.

### 3. Environment `deployBaseRef` Required by Lagoon API
**Problem**: Lagoon's `AddOrUpdateEnvironmentInput.deployBaseRef` is `String!` but was treated as optional.
**Solution**: Default to environment name when not provided in Create/Update.

### 4. Environment Delete Needs Project Name (Not ID)
**Problem**: Lagoon's `DeleteEnvironmentInput.project` is `String!` expecting project name.
**Solution**: Look up project name via `GetProjectByID` before calling `deleteEnvironment`.

### 5. Helm Service Name Mismatch in Multi-Cluster
**Problem**: Code assumed service pattern `{release}-{component}` but lagoon-core chart uses `{release}-lagoon-core-{component}`.
**Solution**: Fixed broker, SSH, Keycloak hostnames and NodePort selector in `lagoon/core.py`.

### 6. Provider Token Rotation Causes Replacements
**Problem**: JWT token changes each `pulumi up` run, causing provider replacement which cascades to all native resources.
**Solution**: Implemented create-or-update semantics (adopt existing on duplicate), idempotent deletes (ignore not-found), and graceful read-not-found (signal state cleanup). Token rotation now completes successfully.

### 7. Lagoon Update Mutations Require {id, patch: {...}} Input
**Problem**: `UpdateKubernetesInput`, `UpdateProjectInput`, `UpdateDeployTargetConfigInput` all require `{id: Int!, patch: {...fields...}}` structure, not flat fields.
**Solution**: Fixed `UpdateDeployTarget`, `UpdateProject`, `UpdateDeployTargetConfig` client methods to wrap fields in `patch` object.

### 8. Lagoon "Already Exists" Error Variants
**Problem**: Lagoon returns different error messages for duplicates — MySQL "Duplicate entry" for deploy targets, "Project already exists" for projects.
**Solution**: `IsDuplicateEntry()` now does case-insensitive matching for both "duplicate entry" and "already exists".

---

## Resource Specifications (All 11 Implemented)

### LagoonProject (`lagoon:lagoon:Project`)
- ForceNew: `name`
- Computed: `lagoonId`, `created`

### LagoonEnvironment (`lagoon:lagoon:Environment`)
- ForceNew: `name`, `projectId`
- Computed: `lagoonId`, `route`, `routes`, `created`

### LagoonVariable (`lagoon:lagoon:Variable`)
- ForceNew: `name`, `projectId`, `environmentId`
- Secret: `value`
- Computed: `lagoonId`

### LagoonDeployTarget (`lagoon:lagoon:DeployTarget`)
- ForceNew: `name`
- Computed: `lagoonId`, `created`

### LagoonDeployTargetConfig (`lagoon:lagoon:DeployTargetConfig`)
- ForceNew: `projectId`, `deployTargetId`
- Computed: `lagoonId`

### LagoonTask (`lagoon:lagoon:Task`)
- ForceNew: `type`, `projectId`, `environmentId`, `groupName`, `systemWide`
- Computed: `lagoonId`, `created`

### Notification Resources
- **NotificationSlack**: ForceNew `name`, Secret `webhook`
- **NotificationRocketChat**: ForceNew `name`, Secret `webhook`
- **NotificationEmail**: ForceNew `name`
- **NotificationMicrosoftTeams**: ForceNew `name`, Secret `webhook`

### LagoonProjectNotification (`lagoon:lagoon:ProjectNotification`)
- ForceNew: ALL fields (Lagoon API doesn't support updating associations)

---

## File Locations

### Native Go Provider (on `native-go-provider` branch)
- Binary entry point: `provider/cmd/pulumi-resource-lagoon/main.go`
- Provider assembly: `provider/pkg/provider/provider.go`
- Config: `provider/pkg/config/config.go`
- Client: `provider/pkg/client/` (9 source files)
- Resources: `provider/pkg/resources/` (12 source files)
- Tests: 11 test files across client/, config/, resources/
- CI: `.github/workflows/test-go.yml`
- Go module: `provider/go.mod` (Go 1.24)
- SDKs: `sdk/python/`, `sdk/nodejs/`, `sdk/go/`

### Multi-Cluster Example
- Makefile: `examples/multi-cluster/Makefile`
- Orchestration: `examples/multi-cluster/__main__.py` (8-phase deployment)
- Native resources: `examples/multi-cluster/lagoon/project.py`
- Lagoon core: `examples/multi-cluster/lagoon/core.py`
- Config classes: `examples/multi-cluster/config.py`

## Python Dynamic Provider Status (v0.1.2)

The original Python provider remains on `main` branch and is the production version:
- 11 resources, 513 unit tests, published on PyPI
- Still works, still supported
- Will be superseded by native Go provider at v0.2.0
