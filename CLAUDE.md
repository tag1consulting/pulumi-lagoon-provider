# CLAUDE.md

This file provides guidance to Claude Code when working with the Pulumi Lagoon Provider project.

## Project Overview

**pulumi-lagoon-provider** is a Pulumi dynamic provider for managing Lagoon resources (projects, environments, variables, etc.) as infrastructure-as-code.

This provider allows you to declaratively manage Lagoon hosting platform resources using Pulumi, enabling infrastructure-as-code workflows for Lagoon project management.

## Project Status

**Status**: v0.1.2 Released (Experimental)

The provider is available on PyPI (`pip install pulumi-lagoon`). This is a Python-based dynamic provider with comprehensive resource support. A native Go provider may be built in the future.

## Architecture

### Phase 1: Dynamic Provider (Current)
- Python-based Pulumi dynamic provider
- Direct GraphQL API integration with Lagoon
- Supports resources: Projects, Environments, Variables, Deploy Targets, Deploy Target Configs, Tasks, and Notifications (Slack, RocketChat, Email, Microsoft Teams)

### Phase 2: Native Provider (Future)
- Go-based native provider using Pulumi SDK
- Generated SDKs for Python, TypeScript, Go
- Full production-ready implementation

## Development Environment

### Prerequisites
- Python 3.8+
- Pulumi CLI installed
- Access to a Lagoon instance with API credentials
- GraphQL API endpoint and authentication token

### Setup Commands
```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Install in development mode
pip install -e .
```

## Project Structure

```
pulumi-lagoon-provider/
├── pulumi_lagoon/           # Main provider package
│   ├── __init__.py         # Package exports
│   ├── client.py           # Lagoon GraphQL API client
│   ├── config.py           # Provider configuration
│   ├── project.py          # LagoonProject resource
│   ├── environment.py      # LagoonEnvironment resource
│   ├── variable.py         # LagoonVariable resource
│   ├── deploytarget.py     # LagoonDeployTarget resource
│   ├── deploytarget_config.py  # LagoonDeployTargetConfig resource
│   ├── task.py             # LagoonTask resource
│   ├── notification_slack.py   # LagoonNotificationSlack resource
│   ├── notification_rocketchat.py  # LagoonNotificationRocketChat resource
│   ├── notification_email.py   # LagoonNotificationEmail resource
│   ├── notification_microsoftteams.py  # LagoonNotificationMicrosoftTeams resource
│   └── project_notification.py # LagoonProjectNotification resource
│
├── scripts/                 # SHARED operational scripts
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
│   ├── simple-project/     # Provider usage example (uses pulumi_lagoon)
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
├── tests/                  # Unit and integration tests
├── memory-bank/            # Documentation and planning
├── setup.py               # Python package configuration
├── requirements.txt       # Python dependencies
├── Makefile               # Development workflow automation
└── README.md             # Project documentation
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

1. **Unit Tests**: Test individual resource providers
2. **Integration Tests**: Test against a real Lagoon instance (requires test environment)
3. **Example Validation**: Ensure examples work correctly

## Development Workflow

1. Make changes to provider code
2. Install in development mode: `pip install -e .`
3. Test with example projects in `examples/`
4. Run tests: `pytest tests/`
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

### Long-term (Phase 3) - In Progress
- [x] Notification resources (Slack, RocketChat, Email, Microsoft Teams)
- [x] Project notification associations
- [x] Task resources (advanced task definitions)
- [ ] Additional resources (Groups)
- [ ] Native Go provider
- [ ] Multi-language SDK generation
- [ ] Community adoption

## Contributing

This is currently an early-stage project. Once the core functionality is working, we'll open it up for community contributions.

## References

- [Pulumi Dynamic Providers](https://www.pulumi.com/docs/intro/concepts/resources/dynamic-providers/)
- [Lagoon Documentation](https://docs.lagoon.sh/)
- [Lagoon GraphQL API](https://api.lagoon.sh/graphql)
