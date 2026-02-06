# Pulumi Lagoon Provider - Implementation Status

**Last Updated**: 2026-02-06
**Status**: Native Go Provider - Phases 1-3 COMPLETE, PR #37 Open

---

## Native Go Provider (Current Work)

### Overview
Migration from Python dynamic provider (v0.1.2) to native Go provider using `pulumi-go-provider` v0.25.0 with `infer` package.

**Branch**: `native-go-provider`
**PR**: https://github.com/tag1consulting/pulumi-lagoon-provider/pull/37 (Draft -> `develop`)
**Commit**: `a0eda7d`

### Phase 1: Scaffolding + GraphQL Client - COMPLETE
- Go module initialized (`go.mod` with Go 1.24, pulumi-go-provider v0.25.0)
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
- 191 unit tests across 3 packages
- GitHub Actions workflow (`test-go.yml`)

### Phase 4: SDK Generation - NOT STARTED
- Generate Python SDK via `pulumi package gen-sdk`
- Generate TypeScript SDK
- `.goreleaser.yml` for cross-platform builds
- Example programs using generated SDKs

### Phase 5: CI/CD + Docs + Migration - NOT STARTED
- Release workflow (goreleaser + SDK publish)
- Documentation updates (CLAUDE.md, README.md)
- Migration guide (dynamic v0.1.x -> native v0.2.0)

---

## Test Coverage Summary

### Go Tests (Native Provider) - 191 tests
| Package | Tests | Time | Description |
|---------|-------|------|-------------|
| `pkg/client` | 111 | ~5s | GraphQL client, all resource CRUD operations |
| `pkg/config` | 13 | <1s | JWT generation, Configure() validation, client factory |
| `pkg/resources` | 67 | <1s | Helper functions, Diff() for all 11 resources |

Test files:
- `client/testutil_test.go` - Mock GraphQL server helper
- `client/errors_test.go` - 11 tests (error types, Is(), Unwrap())
- `client/client_test.go` - 26 tests (Execute, retry, token refresh, API detection)
- `client/project_test.go` - 9 tests
- `client/environment_test.go` - 7 tests
- `client/variable_test.go` - 10 tests (including dual API)
- `client/deploytarget_test.go` - 15 tests (deploy targets + configs)
- `client/notification_test.go` - 21 tests (all 4 types + project notification)
- `client/task_test.go` - 12 tests (polymorphic project/environment fields)
- `config/config_test.go` - 13 tests
- `resources/helpers_test.go` - 6 tests
- `resources/diff_test.go` - 40+ subtests (Diff for all 11 resources)

### Python Tests (Dynamic Provider) - 513 tests
(Still on `main` branch, separate test suite)

---

## Key Issues Encountered & Resolved

### 1. pulumi-go-provider v0.25.0 API Mismatch
**Problem**: Initial implementation used CreateRequest/CreateResponse structs that don't exist in v0.25.0.
**Solution**: v0.25.0 uses plain function signatures. All 11 resources rewritten.
**Reference**: `memory/api-signatures.md` (auto-memory)

### 2. CGO_ENABLED=0 Required
**Problem**: `go build` and `go test` fail with Linuxbrew ld vs system gcc linker errors.
**Solution**: Always use `CGO_ENABLED=0` prefix. `go vet` works without it.

### 3. JSON float64 in Tests
**Problem**: `TestUpdateProject` failed because JSON numbers decode as `float64`, not `int`.
**Solution**: Type-assert to float64 before comparing: `if id, ok := input["id"].(float64); !ok || int(id) != 42`

### 4. json.RawMessage for Polymorphic Fields
**Problem**: Lagoon API returns nested objects (e.g., `openshift: {id: 1, name: "cluster"}`) that need to be normalized to flat IDs.
**Solution**: Use `json.RawMessage` in raw structs, then `normalizeX()` functions to extract IDs.

---

## Provider Analysis Resolution Status

| Finding | Severity | Status | How Resolved |
|---------|----------|--------|-------------|
| H1 - Delete-then-recreate | HIGH | RESOLVED | `Update()` calls mutation directly; `Diff()` forces replace only on immutable fields |
| H2 - JWT duplication | HIGH | RESOLVED | Single `config.go` with centralized JWT, configurable audience |
| H3 - Plaintext credentials | HIGH | RESOLVED | `provider:"secret"` struct tags encrypt token/jwtSecret in state |
| H4 - O(n) fetch-all queries | HIGH | PARTIAL | Client-level caching planned but not yet implemented |
| H5 - No diff() | HIGH | RESOLVED | All 11 resources implement `Diff()` with forceNew fields |
| H6 - Dual build config | HIGH | RESOLVED | Go module replaces Python packaging |
| M2 - Inconsistent credentials | MEDIUM | RESOLVED | All resources use `infer.GetConfig[LagoonConfig](ctx)` |
| M3/M4 - read() drops fields | MEDIUM | RESOLVED | Schema-driven: all fields in State struct always returned |
| M5 - Brittle API fallback | MEDIUM | RESOLVED | `DetectAPIVersion()` probes once, stores result |
| M7 - Exception hierarchy | MEDIUM | RESOLVED | Clean Go error types with `errors.Is()`/`errors.As()` |
| M9 - No retry | MEDIUM | RESOLVED | HTTP client with exponential backoff (3 retries, 1s/2s/4s) |
| M10 - Secret bypass | MEDIUM | RESOLVED | Native secret support via struct tags |
| M11 - No token refresh | MEDIUM | RESOLVED | Client checks expiry before each request, auto-refreshes |

---

## Resource Specifications (All 11 Implemented)

### LagoonProject (`lagoon:index:Project`)
- ForceNew: `name`
- Computed: `lagoonId`, `created`

### LagoonEnvironment (`lagoon:index:Environment`)
- ForceNew: `name`, `projectId`
- Computed: `lagoonId`, `route`, `routes`, `created`

### LagoonVariable (`lagoon:index:Variable`)
- ForceNew: `name`, `projectId`, `environmentId`
- Secret: `value`
- Computed: `lagoonId`

### LagoonDeployTarget (`lagoon:index:DeployTarget`)
- ForceNew: `name`
- Computed: `lagoonId`, `created`

### LagoonDeployTargetConfig (`lagoon:index:DeployTargetConfig`)
- ForceNew: `projectId`, `deployTargetId`
- Computed: `lagoonId`

### LagoonTask (`lagoon:index:Task`)
- ForceNew: `type`, `projectId`, `environmentId`, `groupName`, `systemWide`
- Computed: `lagoonId`, `created`

### Notification Resources
- **NotificationSlack**: ForceNew `name`, Secret `webhook`
- **NotificationRocketChat**: ForceNew `name`, Secret `webhook`
- **NotificationEmail**: ForceNew `name`
- **NotificationMicrosoftTeams**: ForceNew `name`, Secret `webhook`

### LagoonProjectNotification (`lagoon:index:ProjectNotification`)
- ForceNew: ALL fields (Lagoon API doesn't support updating associations)

---

## Python Dynamic Provider Status (v0.1.2)

The original Python provider remains on `main` branch and is the production version:
- 11 resources, 513 unit tests, published on PyPI
- Still works, still supported
- Will be superseded by native Go provider at v0.2.0

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
