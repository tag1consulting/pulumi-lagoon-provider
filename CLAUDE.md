# CLAUDE.md

This file provides guidance to Claude Code when working with the Pulumi Lagoon Provider project.

## Project Overview

**pulumi-lagoon-provider** is a Pulumi native provider for managing Lagoon resources (projects, environments, variables, etc.) as infrastructure-as-code.

This provider allows you to declaratively manage Lagoon hosting platform resources using Pulumi, enabling infrastructure-as-code workflows for Lagoon project management.

## Project Status

**Status**: v0.2.6 Released (Native Go Provider)

The provider is available on PyPI (`pip install pulumi-lagoon`), npm (`@tag1consulting/pulumi-lagoon`), and Go (`github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon`). v0.2.x ships the native Go provider with generated multi-language SDKs, replacing the Python dynamic provider.

## Architecture

### Native Go Provider (Current)
- Go-based native provider using `pulumi-go-provider` v1.3.0 (`infer` package)
- Generated SDKs for Python, TypeScript, and Go from a single schema
- Published to PyPI (`pulumi-lagoon`), npm (`@tag1consulting/pulumi-lagoon`), and Go module
- 198+ unit tests in `provider/`; comprehensive resource CRUD lifecycle

> **Historical note**: v0.1.x was a Python dynamic provider, fully superseded in v0.2.0. The legacy code was removed in v0.2.6 (issue #77).

## Development Environment

### Prerequisites
- Go 1.21+
- Pulumi CLI installed
- Access to a Lagoon instance with API credentials
- GraphQL API endpoint and authentication token

### Setup Commands
```bash
# Build the provider binary
make go-build

# Run Go unit tests
make go-test

# Regenerate Python SDK after schema changes
make go-sdk-python
```

## Project Structure

```
pulumi-lagoon-provider/
├── provider/                # Native Go provider (authoritative source)
│   ├── cmd/pulumi-resource-lagoon/main.go  # Provider entry point
│   ├── pkg/client/         # Lagoon GraphQL API client (Go)
│   ├── pkg/config/         # Provider configuration
│   ├── pkg/provider/       # Pulumi provider implementation
│   ├── pkg/resources/      # Resource CRUD implementations
│   ├── schema.json         # Pulumi schema definition
│   └── go.mod              # Go module definition
│
├── sdk/                    # Generated multi-language SDKs
│   ├── python/             # Generated Python SDK (published to PyPI)
│   ├── nodejs/             # Generated TypeScript/Node.js SDK (published to npm)
│   └── go/                 # Generated Go SDK
│
├── scripts/                # SHARED operational scripts
│   ├── common.sh           # Common functions and configuration
│   ├── run-pulumi.sh       # Wrapper with auto token refresh
│   ├── get-token.sh        # Get OAuth token
│   ├── setup-port-forwards.sh  # Start kubectl port-forwards
│   ├── check-cluster-health.sh # Verify cluster health
│   ├── create-lagoon-admin.sh  # Create Keycloak admin user
│   ├── fix-rabbitmq-password.sh # Fix RabbitMQ auth issues
│   └── ...                 # Other operational scripts
│
├── examples/
│   ├── simple-project/     # Provider usage example (uses sdk/python/)
│   │   ├── __main__.py     # Creates Lagoon projects/envs/vars via API
│   │   └── scripts/        # Helper scripts
│   │
│   ├── single-cluster/     # Single Kind cluster with full Lagoon stack
│   │   ├── __main__.py     # Deploys complete Lagoon to one cluster
│   │   ├── config.py       # Single-cluster configuration
│   │   └── (symlinks)      # Reuses modules from multi-cluster
│   │
│   └── multi-cluster/      # Production-like multi-cluster deployment
│       ├── __main__.py     # Deploys 2 Kind clusters + full Lagoon
│       ├── config.py       # Multi-cluster configuration
│       ├── clusters/       # Kind cluster management
│       ├── infrastructure/ # Ingress, cert-manager, CoreDNS
│       ├── lagoon/         # Core and remote installation
│       └── registry/       # Harbor installation
│
├── docs/                   # Additional documentation
│   └── notifications.md    # Notification resource documentation
├── memory-bank/            # Documentation and planning
├── RELEASE_NOTES.md        # Version changelog
├── Makefile                # Development workflow automation
└── README.md               # Project documentation
```

## Shared Scripts

All operational scripts are in `scripts/` and are parameterized via environment variables:

```bash
# Single-cluster (test-cluster style - default)
LAGOON_PRESET=single ./scripts/check-cluster-health.sh

# Multi-cluster production
LAGOON_PRESET=multi-prod ./scripts/check-cluster-health.sh

# Multi-cluster non-production
LAGOON_PRESET=multi-nonprod ./scripts/check-cluster-health.sh
```

The `LAGOON_PRESET` variable configures:
- Kubernetes context
- Namespace
- Service names
- Secret names

See `scripts/common.sh` for all configuration options.

## Key Resources

### LagoonProject
Manages a Lagoon project (application/site).

**Properties:**
- `name`: Project name
- `git_url`: Git repository URL
- `deploytarget_id`: Target Kubernetes cluster ID
- `production_environment`: Name of production branch
- `branches`: Branch regex pattern
- `pullrequests`: PR regex pattern (optional)

### LagoonEnvironment
Manages a Lagoon environment (branch/PR deployment).

**Properties:**
- `project_id`: Parent project ID
- `name`: Environment name (branch name)
- `environment_type`: production, development, etc.
- `deploy_type`: branch or pullrequest

### LagoonVariable
Manages environment or project-level variables.

**Properties:**
- `project_id`: Parent project ID
- `environment_id`: Environment ID (optional, project-level if omitted)
- `name`: Variable name
- `value`: Variable value
- `scope`: build, runtime, or global

### Notification Resources

#### LagoonNotificationSlack
Manages Slack notification configurations.

**Properties:**
- `name`: Notification name
- `webhook`: Slack webhook URL
- `channel`: Slack channel (e.g., '#deployments')

#### LagoonNotificationRocketChat
Manages RocketChat notification configurations.

**Properties:**
- `name`: Notification name
- `webhook`: RocketChat webhook URL
- `channel`: RocketChat channel

#### LagoonNotificationEmail
Manages Email notification configurations.

**Properties:**
- `name`: Notification name
- `email_address`: Email address to send notifications to

#### LagoonNotificationMicrosoftTeams
Manages Microsoft Teams notification configurations.

**Properties:**
- `name`: Notification name
- `webhook`: Microsoft Teams webhook URL

#### LagoonProjectNotification
Links a notification to a project.

**Properties:**
- `project_name`: Project name
- `notification_type`: Type of notification (slack, rocketchat, email, microsoftteams)
- `notification_name`: Name of the notification to link

### LagoonTask
Manages advanced task definitions (on-demand commands and container-based tasks).

**Properties:**
- `name`: Task definition name
- `type`: Task type (`command` or `image`)
- `service`: Service container name to run the task in
- `command`: Command to execute (required if type='command')
- `image`: Container image to run (required if type='image')
- `project_id`: Project ID (for project-scoped tasks)
- `environment_id`: Environment ID (for environment-scoped tasks)
- `group_name`: Group name (for group-scoped tasks)
- `system_wide`: If true, task is available system-wide (platform admin only)
- `permission`: Permission level (`guest`, `developer`, `maintainer`)
- `description`: Task description (optional)
- `confirmation_text`: Text to display for user confirmation (optional)
- `arguments`: List of argument definitions (optional)

## Lagoon API Integration

### GraphQL API
The provider interacts with Lagoon's GraphQL API:
- Authentication: JWT token or SSH key
- Endpoint: `https://<lagoon-instance>/graphql`
- Key operations: mutations for create/update/delete, queries for read/list

### Example Queries
```graphql
# Create project
mutation AddProject($input: AddProjectInput!) {
  addProject(input: $input) {
    id
    name
    gitUrl
  }
}

# List projects
query AllProjects {
  allProjects {
    id
    name
    gitUrl
    productionEnvironment
  }
}
```

## Testing Strategy

1. **Go Unit Tests**: `make go-test` — tests in `provider/pkg/` with a mock GraphQL server
2. **Integration Tests**: Require a live Lagoon instance; run manually against the example stack
3. **SDK Build Test**: CI builds all SDKs and verifies `import pulumi_lagoon` succeeds

## Development Workflow

1. Make changes to Go provider code in `provider/`
2. Build and test: `make go-build && make go-test`
3. Regenerate SDKs if schema changed: `make go-sdk-all`
4. Test with example projects in `examples/`
5. Update documentation as needed

## Configuration

Provider configuration via Pulumi config:
```bash
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql
pulumi config set lagoon:token <your-token> --secret
```

Or via environment variables:
```bash
export LAGOON_API_URL=https://api.lagoon.example.com/graphql
export LAGOON_TOKEN=<your-token>
```

## Goals

### Short-term (Phase 1) - Complete
- [x] Implement core resources (Project, Environment, Variable)
- [x] GraphQL client with proper error handling
- [x] Working examples
- [x] Basic documentation
- [x] Unit tests

### Medium-term (Phase 2) - Complete
- [x] Deploy target resources (DeployTarget, DeployTargetConfig)
- [x] Multi-cluster example
- [x] Integration tests
- [x] Comprehensive documentation
- [x] Published to PyPI

### Long-term (Phase 3) - Complete
- [x] Notification resources (Slack, RocketChat, Email, Microsoft Teams)
- [x] Project notification associations
- [x] Task resources (advanced task definitions)
- [x] Native Go provider
- [x] Multi-language SDK generation (Python, TypeScript, Go)
- [x] Additional resources (Groups)
- [x] Remove legacy Python dynamic provider code
- [ ] Community adoption

## Contributing

This is currently an early-stage project. Once the core functionality is working, we'll open it up for community contributions.

## References

- [Pulumi Native Providers](https://www.pulumi.com/docs/iac/concepts/resources/providers/)
- [Lagoon Documentation](https://docs.lagoon.sh/)
- [Lagoon GraphQL API](https://api.lagoon.sh/graphql)
