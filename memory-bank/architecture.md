# Pulumi Lagoon Provider - Architecture

## System Architecture

### High-Level Overview

```
┌─────────────────┐
│  Pulumi Program │  (User's __main__.py)
│   (Python/TS)  │
└────────┬────────┘
         │
         │ Uses
         ▼
┌─────────────────────────┐
│  pulumi_lagoon Package  │
│  ┌──────────────────┐   │
│  │ LagoonProject    │   │
│  │ LagoonEnvironment│   │
│  │ LagoonVariable   │   │
│  └──────────────────┘   │
└────────┬────────────────┘
         │
         │ Calls
         ▼
┌─────────────────────────┐
│   LagoonClient          │
│   (GraphQL Client)      │
└────────┬────────────────┘
         │
         │ HTTP/GraphQL
         ▼
┌─────────────────────────┐
│   Lagoon API Server     │
│   (GraphQL Endpoint)    │
└────────┬────────────────┘
         │
         │ Manages
         ▼
┌─────────────────────────┐
│  Lagoon Core Services   │
│  ┌──────────────────┐   │
│  │ Projects         │   │
│  │ Environments     │   │
│  │ Deployments      │   │
│  └──────────────────┘   │
└─────────────────────────┘
```

## Component Architecture

### 1. Resource Layer

Each resource type (Project, Environment, Variable) implements the Pulumi Dynamic Provider interface:

```python
# Resource Definition
class LagoonProject(pulumi.dynamic.Resource):
    """User-facing resource class."""

    # Inputs (what user provides)
    name: str
    git_url: str
    deploytarget_id: int

    # Outputs (what provider returns)
    id: pulumi.Output[int]
    created: pulumi.Output[str]

    def __init__(self, name, args, opts=None):
        # Initialize with provider
        super().__init__(
            LagoonProjectProvider(),
            name,
            {...},
            opts
        )

# Provider Implementation
class LagoonProjectProvider(pulumi.dynamic.ResourceProvider):
    """CRUD operations implementation."""

    def create(self, inputs) -> CreateResult:
        # Create via API

    def update(self, id, old, new) -> UpdateResult:
        # Update via API

    def delete(self, id, props):
        # Delete via API

    def read(self, id, props) -> ReadResult:
        # Refresh from API
```

### 2. Client Layer

The GraphQL client handles all API communication:

```python
class LagoonClient:
    """GraphQL API client for Lagoon."""

    def __init__(self, api_url: str, token: str):
        self.api_url = api_url
        self.token = token
        self.session = requests.Session()

    def _execute(self, query: str, variables: dict) -> dict:
        """Execute GraphQL query with error handling."""

    # High-level operations
    def create_project(self, **kwargs):
        """Create project via addProject mutation."""

    def get_project_by_name(self, name: str):
        """Get project details via query."""
```

### 3. Configuration Layer

Configuration management for provider settings:

```python
class LagoonConfig:
    """Provider configuration."""

    def __init__(self):
        config = pulumi.Config("lagoon")

        # API endpoint
        self.api_url = (
            config.get("apiUrl") or
            os.environ.get("LAGOON_API_URL") or
            "https://api.lagoon.sh/graphql"
        )

        # Authentication
        self.token = (
            config.get_secret("token") or
            os.environ.get("LAGOON_TOKEN")
        )

    def get_client(self) -> LagoonClient:
        """Create configured client instance."""
        return LagoonClient(self.api_url, self.token)
```

## Data Flow

### Resource Creation Flow

```
1. User Code
   lagoon.LagoonProject("mysite", args)
   │
   ▼
2. Resource Constructor
   Validates inputs, initializes provider
   │
   ▼
3. Provider.create()
   Prepares API call
   │
   ▼
4. LagoonClient.create_project()
   Executes GraphQL mutation
   │
   ▼
5. Lagoon API
   Creates project, returns ID
   │
   ▼
6. Provider.create() returns
   CreateResult with outputs
   │
   ▼
7. Pulumi State
   Stores resource state
```

### State Refresh Flow

```
1. Pulumi Refresh Command
   │
   ▼
2. Provider.read()
   │
   ▼
3. LagoonClient.get_project()
   Query current state
   │
   ▼
4. Lagoon API
   Returns current state
   │
   ▼
5. Provider.read() returns
   ReadResult with current state
   │
   ▼
6. Pulumi
   Compares with stored state
   Detects drift if different
```

## GraphQL API Integration

### Authentication

```python
# JWT Token in HTTP Headers
headers = {
    "Authorization": f"Bearer {token}",
    "Content-Type": "application/json"
}
```

### Query Structure

```graphql
# Create Project
mutation AddProject($input: AddProjectInput!) {
  addProject(input: $input) {
    id
    name
    gitUrl
    productionEnvironment
    openshift
    created
  }
}

# Variables
{
  "input": {
    "name": "my-project",
    "gitUrl": "git@github.com:org/repo.git",
    "openshift": 1,
    "productionEnvironment": "main"
  }
}
```

### Error Handling

```python
try:
    response = self.session.post(url, json=payload, headers=headers)
    response.raise_for_status()

    data = response.json()

    if "errors" in data:
        # GraphQL errors
        raise LagoonAPIError(data["errors"])

    return data["data"]

except requests.HTTPError as e:
    # HTTP errors
    raise LagoonConnectionError(str(e))
```

## Resource Relationships

### Dependency Graph

```
LagoonProject
    ├── LagoonEnvironment (many)
    │   ├── LagoonVariable (many)
    │   └── LagoonRoute (many)
    └── LagoonVariable (many, project-scoped)

LagoonDeployTarget
    └── LagoonProject (many)

LagoonGroup
    └── LagoonProject (many-to-many)
```

### Pulumi Dependencies

```python
# Explicit dependency
project = lagoon.LagoonProject("site", ...)

env = lagoon.LagoonEnvironment("prod",
    project_id=project.id,  # Implicit dependency
    ...
)

var = lagoon.LagoonVariable("db",
    project_id=project.id,
    environment_id=env.id,  # Depends on both
    ...
)
```

## State Management

### State Storage

Pulumi stores resource state in its backend (local file, S3, Pulumi Cloud, etc.):

```json
{
  "resources": [
    {
      "type": "lagoon:index:Project",
      "urn": "urn:pulumi:dev::my-stack::lagoon:index:Project::mysite",
      "id": "42",
      "inputs": {
        "name": "my-site",
        "gitUrl": "git@github.com:org/repo.git",
        "deploytargetId": 1
      },
      "outputs": {
        "id": 42,
        "created": "2025-10-14T12:00:00Z"
      }
    }
  ]
}
```

### Drift Detection

Provider must implement `read()` to enable drift detection:

```python
def read(self, id, props):
    """Refresh resource state from API."""
    client = get_client()
    current = client.get_project_by_id(int(id))

    if not current:
        # Resource deleted outside Pulumi
        return None

    return pulumi.dynamic.ReadResult(
        id=id,
        outs={
            "name": current["name"],
            "gitUrl": current["gitUrl"],
            # ... other properties
        }
    )
```

## Error Handling Strategy

### Error Categories

1. **Configuration Errors**
   - Missing API URL
   - Invalid credentials
   - Action: Fail fast with clear message

2. **API Errors**
   - Network failures
   - Authentication failures
   - Rate limiting
   - Action: Retry with backoff, then fail

3. **Resource Errors**
   - Resource already exists
   - Resource not found
   - Invalid input
   - Action: Surface GraphQL errors to user

4. **State Errors**
   - Drift detected
   - Concurrent modification
   - Action: Inform user, suggest refresh

### Error Classes

```python
class LagoonError(Exception):
    """Base exception for all Lagoon provider errors."""

class LagoonConfigError(LagoonError):
    """Configuration error."""

class LagoonAPIError(LagoonError):
    """API request failed."""

class LagoonResourceError(LagoonError):
    """Resource operation failed."""
```

## Testing Strategy

### Unit Tests
Test individual components in isolation:

```python
def test_lagoon_client_create_project():
    """Test GraphQL client project creation."""
    mock_response = {"data": {"addProject": {"id": 42}}}

    with patch('requests.Session.post') as mock_post:
        mock_post.return_value.json.return_value = mock_response

        client = LagoonClient("http://test", "token")
        result = client.create_project(
            name="test",
            git_url="git@github.com:test/test.git",
            openshift=1
        )

        assert result["id"] == 42
```

### Integration Tests
Test against real/test Lagoon instance:

```python
@pytest.mark.integration
def test_full_project_lifecycle():
    """Test creating, updating, and deleting a project."""
    # Requires LAGOON_TEST_API_URL and LAGOON_TEST_TOKEN
    # Creates real resources, then cleans up
```

### Example Tests
Validate example programs work:

```bash
cd examples/simple-project
pulumi preview  # Should succeed
```

## Security Considerations

### Secrets Management
- API tokens stored as Pulumi secrets
- Never log sensitive data
- Secure credential rotation support

### Network Security
- HTTPS only for API communication
- Certificate validation
- Optional proxy support

### Audit Trail
- Log all API operations (without secrets)
- Track resource changes
- Enable Lagoon audit integration

## Performance Considerations

### API Rate Limiting
- Implement backoff/retry
- Batch operations where possible
- Cache read-only data

### Parallel Operations
- Pulumi handles parallelism
- Provider must be thread-safe
- Use connection pooling

## Future Enhancements

### Phase 2 Features
- Import existing resources
- Bulk operations support
- Advanced validation
- Custom resource transformations

### Phase 3 Features (Native Provider)
- Multi-language SDKs
- Improved performance
- Better IDE integration
- Component resources

## References

- [Pulumi Dynamic Providers](https://www.pulumi.com/docs/intro/concepts/resources/dynamic-providers/)
- [Pulumi Resource Model](https://www.pulumi.com/docs/intro/concepts/resources/)
- [Lagoon GraphQL API](https://api.lagoon.sh/graphql)
