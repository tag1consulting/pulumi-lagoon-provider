# Pulumi Lagoon Provider - Architecture

**Last Updated**: 2026-03-24

## Two Provider Implementations

### 1. Python Dynamic Provider (v0.1.2, production)
- Branch: `main`
- Published on PyPI as `pulumi-lagoon`
- Uses Pulumi dynamic provider interface
- 11 resources, 513 unit tests

### 2. Native Go Provider (v0.2.0 — released)
- Merged via PR #37 → `develop` → `main`
- Uses `pulumi-go-provider` v1.3.0 with `infer` package (request/response builder pattern)
- 11 resources, 191+ unit tests
- Resolves all HIGH/MEDIUM findings from provider-analysis.md

---

## Native Go Provider Architecture

### High-Level Overview

```text
┌─────────────────┐
│  Pulumi Program │  (Python/TS/Go via generated SDK)
└────────┬────────┘
         │ gRPC
         ▼
┌──────────────────────────────────┐
│  pulumi-resource-lagoon binary   │
│  (provider/cmd/.../main.go)      │
│                                  │
│  ┌────────────────────────────┐  │
│  │ provider.go                │  │  infer.Provider() wires everything
│  │   Config: LagoonConfig     │  │
│  │   Resources: [11 types]    │  │
│  └────────────────────────────┘  │
│                                  │
│  ┌────────────────────────────┐  │
│  │ config.go                  │  │  Auth, JWT, client factory
│  │   Configure() validation   │  │  provider:"secret" tags
│  │   NewClient() factory      │  │  Env var fallback
│  └────────────────────────────┘  │
│                                  │
│  ┌────────────────────────────┐  │
│  │ resources/*.go             │  │  11 resources, each with:
│  │   Create/Read/Update/Delete│  │  - TArgs (inputs)
│  │   Diff/Check               │  │  - TState (outputs)
│  └────────────────────────────┘  │
└────────┬─────────────────────────┘
         │
         ▼
┌─────────────────────────┐
│   client/*.go            │  GraphQL client layer
│   - Retry (3x exp backoff)│
│   - Token refresh         │
│   - API version detection │
│   - Typed errors          │
└────────┬────────────────┘
         │ HTTP/GraphQL
         ▼
┌─────────────────────────┐
│   Lagoon API Server     │
│   (GraphQL Endpoint)    │
└─────────────────────────┘
```

### Package Dependency Graph

```text
main.go
  └── provider/provider.go
        ├── config/config.go
        │     └── client/client.go
        └── resources/*.go (11 resource files)
              ├── config/config.go (via infer.GetConfig)
              └── client/*.go (via config.NewClient())
```

### Resource Layer (`provider/pkg/resources/`)

Each resource implements the `infer` interfaces using request/response structs (v1.3.0 API):

```go
type Project struct{}

// Input struct (what user provides)
type ProjectArgs struct {
    Name                 string  `pulumi:"name"`
    GitURL               string  `pulumi:"gitUrl"`
    DeploytargetID       int     `pulumi:"deploytargetId"`
    ProductionEnvironment *string `pulumi:"productionEnvironment,optional"`
    // ...
}

// Output struct (what provider returns)
type ProjectState struct {
    ProjectArgs              // Embeds all inputs
    LagoonID int    `pulumi:"lagoonId"`
    Created  string `pulumi:"created"`
}

// CRUD methods use request/response structs:
func (r *Project) Create(ctx context.Context, req infer.CreateRequest[ProjectArgs]) (infer.CreateResponse[ProjectState], error)
func (r *Project) Read(ctx context.Context, req infer.ReadRequest[ProjectArgs, ProjectState]) (infer.ReadResponse[ProjectArgs, ProjectState], error)
func (r *Project) Update(ctx context.Context, req infer.UpdateRequest[ProjectArgs, ProjectState]) (infer.UpdateResponse[ProjectState], error)
func (r *Project) Delete(ctx context.Context, req infer.DeleteRequest[ProjectState]) (infer.DeleteResponse, error)
func (r *Project) Diff(ctx context.Context, req infer.DiffRequest[ProjectArgs, ProjectState]) (infer.DiffResponse, error)
```

### Client Layer (`provider/pkg/client/`)

Core client with retry and token management:

```go
type Client struct {
    endpoint  string
    token     string
    http      *http.Client
    isNewAPI  bool           // v2.30.0+ detection result
    tokenFunc func() string  // For JWT refresh
}

func (c *Client) Execute(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error)
// - Adds Bearer token header
// - Checks/refreshes token before request
// - Retries 3x with exponential backoff (1s, 2s, 4s) on 5xx/network errors
// - Does NOT retry on 4xx or GraphQL errors
```

Resource-specific files (project.go, environment.go, etc.) wrap Execute() with type-safe methods:

```go
func (c *Client) CreateProject(ctx context.Context, input map[string]any) (*Project, error)
func (c *Client) GetProjectByName(ctx context.Context, name string) (*Project, error)
// etc.
```

### Config Layer (`provider/pkg/config/`)

```go
type LagoonConfig struct {
    APIUrl      string `pulumi:"apiUrl" provider:"secret"`
    Token       string `pulumi:"token,optional" provider:"secret"`
    JWTSecret   string `pulumi:"jwtSecret,optional" provider:"secret"`
    JWTAudience string `pulumi:"jwtAudience,optional"`
    Insecure    bool   `pulumi:"insecure,optional"`
}

func (c *LagoonConfig) Configure(ctx context.Context) error
// - Token takes precedence over JWTSecret
// - Falls back to LAGOON_TOKEN / LAGOON_JWT_SECRET env vars
// - Generates JWT from secret if no token provided

func (c *LagoonConfig) NewClient() *client.Client
// - Creates client with token or tokenFunc (for auto-refresh)
```

---

## Resource Relationships

```
LagoonProject
    ├── LagoonEnvironment (many)
    │   ├── LagoonVariable (many, environment-scoped)
    │   └── LagoonTask (many, environment-scoped)
    ├── LagoonVariable (many, project-scoped)
    ├── LagoonDeployTargetConfig (many)
    ├── LagoonProjectNotification (many)
    │   └── references: NotificationSlack / RocketChat / Email / MicrosoftTeams
    └── LagoonTask (many, project-scoped)

LagoonDeployTarget
    └── LagoonProject (many, via deploytargetId)
        └── LagoonDeployTargetConfig (many)
```

## Error Handling

```go
// Typed errors in client/errors.go
var ErrNotFound = errors.New("resource not found")
var ErrValidation = errors.New("validation error")
var ErrAPI = errors.New("API error")
var ErrConnection = errors.New("connection error")

// Concrete types wrap these sentinels
type LagoonAPIError struct { Message string; Errors []GraphQLError }
type LagoonConnectionError struct { Err error }
type LagoonNotFoundError struct { Resource string; ID string }
type LagoonValidationError struct { Field string; Message string }

// Usage: errors.Is(err, ErrNotFound) works with all concrete types
```

## GraphQL API Integration

### Authentication
```go
// Bearer token in HTTP headers
req.Header.Set("Authorization", "Bearer "+c.token)
req.Header.Set("Content-Type", "application/json")
```

### API Version Detection
```go
// DetectAPIVersion() probes for v2.30.0+ features
// Used by Variable resource for new vs legacy mutation format
func (c *Client) DetectAPIVersion(ctx context.Context) error
func (c *Client) IsNewAPI() bool
```

### Dual API Support (Variables)
```go
// New API (v2.30.0+): addOrUpdateEnvVariableByName mutation
// Legacy API: addEnvVariable mutation
// Client auto-detects and falls back if new API returns "field not found" error
```

## Security

### Secrets in State
Fields tagged with `provider:"secret"` are encrypted in Pulumi state:
- `LagoonConfig.Token`
- `LagoonConfig.JWTSecret`
- `LagoonVariableState.Value`
- `NotificationSlack/RocketChat/MicrosoftTeams.Webhook`

### ForceNew (Replace) Fields
Resources implement `Diff()` to mark immutable fields with `p.UpdateReplace`:
- Names (project, environment, notifications)
- Parent IDs (projectId, environmentId, deploytargetId)
- Types (task type)
- All fields on ProjectNotification (API doesn't support updates)

## Testing Strategy

### Unit Tests (191)
- Mock GraphQL server using `net/http/httptest`
- Shared helper in `testutil_test.go`
- Tests cover: CRUD operations, error handling, normalization, Diff behavior

### CI/CD
- `.github/workflows/test-go.yml` - Three jobs:
  1. `test` - Run tests with coverage
  2. `vet` - Static analysis
  3. `build` - Binary compilation
- Triggers on push to main/develop and PRs
