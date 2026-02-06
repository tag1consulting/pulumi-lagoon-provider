# Next Session Quickstart - Native Go Pulumi Lagoon Provider

**Date Updated**: 2026-02-06
**Status**: Phase 3 COMPLETE - Tests + PR Open

---

## RESUME FROM HERE (Monday)

**Branch**: `native-go-provider`
**PR**: https://github.com/tag1consulting/pulumi-lagoon-provider/pull/37 (Draft, targeting `develop`)
**Commit**: `a0eda7d` - "Add native Go provider with 11 resources and 191 unit tests"

### What's Done
- All 11 resources implemented and compiling
- 191 unit tests passing across 3 packages (client, config, resources)
- GitHub Actions workflow for Go tests (`.github/workflows/test-go.yml`)
- Draft PR #37 open targeting `develop`

### What to do Monday

1. **Check PR status**:
   ```bash
   gh pr view 37
   gh pr checks 37
   ```

2. **Verify tests still pass**:
   ```bash
   cd provider && CGO_ENABLED=0 go test ./... -count=1
   ```

3. **Decide next steps** (see "Remaining Work" below)

---

## Quick Reference

### Build & Test Commands
```bash
# Build the provider binary
cd provider && CGO_ENABLED=0 go build -o bin/pulumi-resource-lagoon ./cmd/pulumi-resource-lagoon

# Run all tests
cd provider && CGO_ENABLED=0 go test ./... -v

# Run tests with coverage
cd provider && CGO_ENABLED=0 go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# Type check only (fast)
cd provider && go vet ./...
```

### CRITICAL: CGO_ENABLED=0
Both `go build` AND `go test` require `CGO_ENABLED=0` on this system due to Linuxbrew ld vs system gcc conflict. Without it, you get linker errors.

---

## Project Structure (Native Go Provider)

```
pulumi-lagoon-provider-native/
├── provider/
│   ├── cmd/pulumi-resource-lagoon/
│   │   └── main.go                     # Binary entry point
│   ├── pkg/
│   │   ├── config/
│   │   │   ├── config.go               # Provider config (auth, JWT, client factory)
│   │   │   └── config_test.go          # 13 tests
│   │   ├── client/
│   │   │   ├── client.go               # Core GraphQL client (retry, token refresh)
│   │   │   ├── errors.go               # Typed errors (LagoonAPIError, etc.)
│   │   │   ├── queries.go              # GraphQL query/mutation constants
│   │   │   ├── project.go              # Project CRUD
│   │   │   ├── environment.go          # Environment CRUD
│   │   │   ├── variable.go             # Variable CRUD (dual API v2.30.0+/legacy)
│   │   │   ├── deploytarget.go         # Deploy target + config CRUD
│   │   │   ├── notification.go         # All 4 notification types + project notification
│   │   │   ├── task.go                 # Task definition CRUD
│   │   │   ├── testutil_test.go        # Mock GraphQL server helper
│   │   │   ├── client_test.go          # 26 tests
│   │   │   ├── errors_test.go          # 11 tests
│   │   │   ├── project_test.go         # 9 tests
│   │   │   ├── environment_test.go     # 7 tests
│   │   │   ├── variable_test.go        # 10 tests
│   │   │   ├── deploytarget_test.go    # 15 tests
│   │   │   ├── notification_test.go    # 21 tests
│   │   │   └── task_test.go            # 12 tests
│   │   ├── resources/
│   │   │   ├── project.go              # LagoonProject resource
│   │   │   ├── environment.go          # LagoonEnvironment
│   │   │   ├── variable.go             # LagoonVariable
│   │   │   ├── deploytarget.go         # LagoonDeployTarget
│   │   │   ├── deploytarget_config.go  # LagoonDeployTargetConfig
│   │   │   ├── notification_slack.go   # LagoonNotificationSlack
│   │   │   ├── notification_rocketchat.go
│   │   │   ├── notification_email.go
│   │   │   ├── notification_microsoftteams.go
│   │   │   ├── project_notification.go # LagoonProjectNotification
│   │   │   ├── task.go                 # LagoonTask
│   │   │   ├── helpers.go              # Shared helpers
│   │   │   ├── helpers_test.go         # 6 tests
│   │   │   └── diff_test.go            # 40+ tests (Diff for all resources)
│   │   └── provider/
│   │       └── provider.go             # Provider assembly (wires config + resources)
│   ├── bin/                            # Build output (gitignored)
│   └── go.mod                          # Go 1.24, pulumi-go-provider v0.25.0
├── .github/workflows/
│   ├── test.yml                        # Python tests (existing)
│   └── test-go.yml                     # Go tests (NEW - 3 jobs: test, vet, build)
└── memory-bank/                        # This documentation
```

---

## Remaining Work (Plan Phases)

### Phase 4: SDK Generation (NOT STARTED)
- Generate Python SDK: `pulumi package gen-sdk ./bin/pulumi-resource-lagoon --language python`
- Generate TypeScript SDK: `pulumi package gen-sdk ./bin/pulumi-resource-lagoon --language nodejs`
- Set up `.goreleaser.yml` for cross-platform builds
- Create `examples/py-native/` and `examples/ts-native/`
- Update Makefile with Go build targets

### Phase 5: CI/CD + Docs + Migration (NOT STARTED)
- `.github/workflows/release.yml` - goreleaser + SDK publish
- Update `CLAUDE.md` and `README.md` with Go provider docs
- Write `docs/migration-guide.md` (dynamic v0.1.x -> native v0.2.0)

### Optional Improvements
- Increase test coverage (currently 191 tests, good but could add edge cases)
- Integration tests against live Lagoon
- `pulumi import` verification for all 11 resource types
- Secret verification (`pulumi stack export` shows encrypted values)

---

## Key Technical Details

### pulumi-go-provider v0.25.0 API
The v0.25.0 API uses **plain function signatures**, NOT request/response structs:
```go
func (r *Resource) Create(ctx context.Context, name string, input TArgs, preview bool) (id string, output TState, err error)
func (r *Resource) Read(ctx context.Context, id string, inputs TArgs, state TState) (canonicalID string, normalizedInputs TArgs, normalizedState TState, err error)
func (r *Resource) Update(ctx context.Context, id string, olds TState, news TArgs, preview bool) (TState, error)
func (r *Resource) Delete(ctx context.Context, id string, props TState) error
func (r *Resource) Diff(ctx context.Context, id string, olds TState, news TArgs) (p.DiffResponse, error)
```

### Provider Config Pattern
```go
// Get config in any resource method:
cfg := infer.GetConfig[config.LagoonConfig](ctx)
c := cfg.NewClient()
```

### Test Pattern
Tests use `net/http/httptest` mock GraphQL server (see `testutil_test.go`).
Important gotcha: JSON numbers decode as `float64` in Go, not `int`.

---

## Python Dynamic Provider (Still Active)

The original Python dynamic provider (v0.1.2) is in the same repo on `main` branch:
- 11 resources, 513 Python unit tests
- Published on PyPI as `pulumi-lagoon`
- Still the production version until native Go provider is ready

---

## Git Branches
- `main` - Python dynamic provider (v0.1.2, production)
- `develop` - Integration branch (PR target)
- `native-go-provider` - Native Go provider work (current, PR #37 open)
