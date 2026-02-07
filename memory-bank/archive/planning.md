# Pulumi Lagoon Provider - Planning Document

**Created**: 2025-10-14
**Status**: Initial Planning Phase

## Project Genesis

This project emerged from an analysis of infrastructure repositories showing extensive use of Lagoon (open source hosting platform) alongside Pulumi for infrastructure-as-code. The goal is to create a Pulumi provider that allows declarative management of Lagoon resources.

## Problem Statement

Currently, Lagoon project/environment/variable management is done through:
- Manual CLI operations (`lagoon` CLI)
- Web UI interactions
- Ad-hoc scripts

This creates challenges:
- No infrastructure-as-code for Lagoon resources
- Manual, error-prone processes
- No version control for Lagoon configurations
- Difficult to replicate environments
- No drift detection

## Solution: Pulumi Lagoon Provider

A Pulumi provider that treats Lagoon resources as first-class infrastructure components, manageable through Pulumi programs alongside AWS/EKS/Kubernetes resources.

## Architecture Approaches Considered

### Option 1: Dynamic Provider (Python) ✅ SELECTED FOR PHASE 1
**Pros:**
- Quick to build and iterate
- Python-based (matches existing Pulumi codebases)
- Good for proof of concept
- Direct GraphQL API integration

**Cons:**
- Limited to Python runtime
- Not distributable as standard package
- Less performant than native

**Decision**: Start here to validate concept

### Option 2: Native Provider (Go)
**Pros:**
- Production-ready
- Generates SDKs for multiple languages
- Better performance
- Standard distribution model

**Cons:**
- More complex implementation
- Requires Go expertise
- Longer development time

**Decision**: Phase 2 after validating with dynamic provider

### Option 3: Bridged Terraform Provider
**Status**: No Terraform Lagoon provider exists
**Decision**: Not viable currently

## Resource Model

### Core Resources (Phase 1)

#### LagoonProject
Represents a Lagoon project (application/site).

**GraphQL Mapping**: `addProject` mutation, `project` query

**Key Properties**:
- `name`: Project identifier
- `git_url`: Source repository
- `deploytarget_id`: Target Kubernetes cluster
- `production_environment`: Production branch name
- `branches`: Branch deployment regex
- `pullrequests`: PR deployment regex (optional)
- `openshift_project_pattern`: Namespace pattern
- `auto_idle`: Auto-idle configuration
- `storage_calc`: Storage calculation settings

**Outputs**:
- `id`: Lagoon project ID
- `created`: Creation timestamp
- `environments`: List of environments

#### LagoonEnvironment
Represents a deployed environment (branch or PR).

**GraphQL Mapping**: `addOrUpdateEnvironment` mutation

**Key Properties**:
- `project_id`: Parent project
- `name`: Environment name (typically branch name)
- `environment_type`: production, development, etc.
- `deploy_type`: branch or pullrequest
- `openshift_project_name`: Kubernetes namespace

**Outputs**:
- `id`: Environment ID
- `route`: Primary route URL
- `routes`: All routes
- `created`: Creation timestamp

#### LagoonVariable
Represents environment or project-scoped variables.

**GraphQL Mapping**: `addEnvVariable` mutation

**Key Properties**:
- `project_id`: Parent project
- `environment_id`: Environment (optional for project-level vars)
- `name`: Variable name
- `value`: Variable value
- `scope`: build, runtime, or global

**Outputs**:
- `id`: Variable ID

### Extended Resources (Phase 2)

- **LagoonDeployTarget**: Manage Kubernetes cluster targets
- **LagoonGroup**: User groups and permissions
- **LagoonNotification**: Slack, webhook, etc.
- **LagoonTask**: One-off and recurring tasks
- **LagoonBackup**: Backup schedules and retention

## Technical Implementation

### GraphQL Client Architecture

```python
class LagoonClient:
    def __init__(self, api_url: str, token: str):
        """Initialize client with API endpoint and auth token."""

    def query(self, query: str, variables: dict) -> dict:
        """Execute GraphQL query/mutation."""

    # Project operations
    def create_project(self, **kwargs) -> dict:
    def get_project(self, name: str) -> dict:
    def update_project(self, id: int, **kwargs) -> dict:
    def delete_project(self, name: str) -> dict:

    # Environment operations
    def add_or_update_environment(self, **kwargs) -> dict:
    def delete_environment(self, name: str, project: str) -> dict:

    # Variable operations
    def add_variable(self, **kwargs) -> dict:
    def delete_variable(self, name: str, project: str, env: str = None) -> dict:
```

### Dynamic Provider Pattern

```python
class LagoonProject(pulumi.dynamic.Resource):
    """Dynamic provider for Lagoon projects."""

    def __init__(self, name, args, opts=None):
        super().__init__(
            LagoonProjectProvider(),
            name,
            {
                'name': args.name,
                'git_url': args.git_url,
                'deploytarget_id': args.deploytarget_id,
                # ... other properties
                'id': None,  # Set by provider
            },
            opts
        )

class LagoonProjectProvider(pulumi.dynamic.ResourceProvider):
    """Provider implementation with CRUD operations."""

    def create(self, inputs):
        """Create new Lagoon project via API."""
        client = self._get_client()
        result = client.create_project(**inputs)
        return pulumi.dynamic.CreateResult(
            id=str(result['id']),
            outs={**inputs, 'id': result['id']}
        )

    def update(self, id, old_inputs, new_inputs):
        """Update existing project."""

    def delete(self, id, props):
        """Delete project."""

    def read(self, id, props):
        """Refresh state from API."""
```

## Configuration & Authentication

### Provider Configuration

Via Pulumi config:
```bash
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql
pulumi config set lagoon:token <token> --secret
```

Via environment variables:
```bash
export LAGOON_API_URL=https://api.lagoon.example.com/graphql
export LAGOON_TOKEN=<token>
```

### Authentication Methods
1. **JWT Token** (primary): Long-lived API token
2. **SSH Key** (future): SSH-based authentication

## Example Usage

```python
import pulumi
import pulumi_lagoon as lagoon

# Configure provider
config = pulumi.Config("lagoon")

# Create a Lagoon project
project = lagoon.LagoonProject("my-drupal-site",
    name="my-drupal-site",
    git_url="git@github.com:org/repo.git",
    deploytarget_id=1,
    production_environment="main",
    branches="^(main|develop)$",
    pullrequests="^(PR-.*)",
)

# Create production environment
prod_env = lagoon.LagoonEnvironment("production",
    project_id=project.id,
    name="main",
    environment_type="production",
    deploy_type="branch",
)

# Add environment variable
db_host = lagoon.LagoonVariable("db-host",
    project_id=project.id,
    environment_id=prod_env.id,
    name="DATABASE_HOST",
    value="mysql.example.com",
    scope="runtime",
)

# Export outputs
pulumi.export("project_id", project.id)
pulumi.export("production_url", prod_env.route)
```

## Integration with Existing Infrastructure

Can be used alongside existing Pulumi EKS/Kubernetes code:

```python
import pulumi
import pulumi_eks as eks
import pulumi_lagoon as lagoon

# Existing EKS cluster
cluster = eks.Cluster("lagoon-cluster", ...)

# Deploy Lagoon to cluster (via Helm)
lagoon_core = helm3.Chart("lagoon-core", ...)

# Register cluster as deploy target
deploytarget = lagoon.LagoonDeployTarget("dev-cluster",
    name="dev-eks",
    kubernetes_endpoint=cluster.endpoint,
    router_pattern="${project}-${environment}.dev.example.com",
)

# Now create projects targeting this cluster
project = lagoon.LagoonProject("site",
    deploytarget_id=deploytarget.id,
    # ...
)
```

## Development Roadmap

### Phase 1: MVP (Weeks 1-4)
- [x] Project setup and structure
- [ ] GraphQL client implementation
- [ ] LagoonProject resource
- [ ] LagoonEnvironment resource
- [ ] LagoonVariable resource
- [ ] Basic example
- [ ] README documentation

### Phase 2: Polish (Weeks 5-8)
- [ ] Error handling improvements
- [ ] Unit tests
- [ ] Integration tests
- [ ] Multiple examples
- [ ] Advanced documentation

### Phase 3: Extended Resources (Weeks 9-12)
- [ ] DeployTarget resource
- [ ] Group resource
- [ ] Notification resource
- [ ] Task resource

### Phase 4: Native Provider (Month 4+)
- [ ] Go implementation
- [ ] Schema definition
- [ ] SDK generation
- [ ] Migration guide

## Success Criteria

### Phase 1 Success
- Can create/update/delete Lagoon projects via Pulumi
- Can manage environments and variables
- Working example that deploys a real project
- Documentation sufficient for early adopters

### Production Ready
- Comprehensive resource coverage
- Robust error handling
- Full test coverage
- Production deployments
- Community adoption

## Risks & Mitigations

### Risk: Lagoon API Changes
**Impact**: High
**Mitigation**: Version detection, API compatibility layer

### Risk: Authentication Complexity
**Impact**: Medium
**Mitigation**: Support multiple auth methods, clear documentation

### Risk: State Synchronization
**Impact**: Medium
**Mitigation**: Implement proper refresh operations, handle drift

### Risk: Adoption
**Impact**: Medium
**Mitigation**: Excellent documentation, examples, community engagement

## Open Questions

1. **API Versioning**: How to handle multiple Lagoon versions?
2. **Multi-tenancy**: How to handle multiple Lagoon instances?
3. **Import**: How to import existing Lagoon resources?
4. **Testing**: How to test without affecting production?

## Next Steps

1. Implement GraphQL client
2. Build LagoonProject resource
3. Create working example
4. Validate against real Lagoon instance
5. Iterate based on feedback
