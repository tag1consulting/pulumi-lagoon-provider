# Release v0.1.2 (2026-02-06)

This release adds multi-version Lagoon API compatibility, supporting Lagoon versions v2.24.1 through v2.30.0.

## Highlights

- **Multi-Version API Compatibility**: Automatic fallback to older API queries/mutations for Lagoon versions prior to v2.30.0
- **CRD Version Detection**: Automatic selection of correct CRD storage version based on Lagoon chart version
- **SSL Verification Option**: Support for self-signed certificates in development environments
- **Expanded Test Coverage**: 507 unit tests with 92% code coverage

## New Features

### Multi-Version Lagoon API Support

The provider now automatically detects and adapts to different Lagoon API versions:

| Lagoon Version | Chart Version | Status |
|----------------|---------------|--------|
| v2.30.0 | 1.59.0 | Primary API |
| v2.28.0 | 1.56.0 | Fallback support |
| v2.24.1 | 1.52.0 | Fallback support |

**API Changes Handled:**
- `get_project_by_id`: Uses `allProjects` + filter (projectById doesn't exist in older versions)
- `get_kubernetes_by_id`: Uses `allKubernetes` + filter
- `add_env_variable`: Falls back to `addEnvVariable` mutation
- `delete_env_variable`: Falls back to `deleteEnvVariable` mutation
- `get_env_variable_by_name`: Falls back to `envVariablesByProjectEnvironment` query
- `get_advanced_tasks_by_environment`: Falls back to `advancedTasksByEnvironment` query

### CRD Version Detection

For Kubernetes deployments, the provider automatically selects the correct CRD storage version:
- **Chart < 1.58.0** (Lagoon ≤ v2.28.0): v1beta1 as storage version
- **Chart ≥ 1.58.0** (Lagoon ≥ v2.29.0): v1beta2 as storage version
- Both v1beta1 and v1beta2 are always served (required by lagoon-remote controller)

### SSL Verification Option

`LagoonDeployTarget` now supports a `verify_ssl` parameter for environments with self-signed certificates:

```python
deploy_target = lagoon.LagoonDeployTarget("local-cluster",
    name="local-kind",
    console_url="https://kubernetes.default.svc",
    api_url="https://api.lagoon.local/graphql",
    jwt_secret=jwt_secret,
    verify_ssl=False,  # For self-signed certs
)
```

## Bug Fixes

- Fixed `get_project_by_id` to work with Lagoon versions that don't support `projectById` query
- Fixed environment variable operations for older Lagoon API versions
- Fixed advanced task queries to handle different response formats across versions

## Documentation

- Updated README with multi-version compatibility information
- Added version compatibility table to documentation

## Requirements

- Python 3.8+
- Pulumi CLI 3.x
- Lagoon v2.24.1 or later

## Installation

```bash
pip install pulumi-lagoon==0.1.2
```

## Full Changelog

See the [commit history](https://github.com/tag1consulting/pulumi-lagoon-provider/compare/v0.1.1...v0.1.2) for all changes.

---

# Release v0.1.1 (2026-02-02)

This release adds notification and task management resources to the Pulumi Lagoon Provider.

## Highlights

- **Notification Resources**: Full CRUD support for all Lagoon notification types (Slack, RocketChat, Email, Microsoft Teams) plus project notification linking
- **Task Resources**: Manage advanced task definitions for on-demand commands and container-based tasks
- **Expanded Test Coverage**: 467+ unit tests

## New Features

### Notification Resources
- **LagoonNotificationSlack** - Create and manage Slack webhook notifications
- **LagoonNotificationRocketChat** - Create and manage RocketChat webhook notifications
- **LagoonNotificationEmail** - Create and manage email notifications
- **LagoonNotificationMicrosoftTeams** - Create and manage Microsoft Teams webhook notifications
- **LagoonProjectNotification** - Link notifications to projects for deployment alerts

### Task Resources
- **LagoonTask** - Create and manage advanced task definitions with support for:
  - Command-type tasks (execute commands in existing service containers)
  - Image-type tasks (run custom container images)
  - Multiple scope options: project, environment, group, or system-wide
  - Permission levels: guest, developer, maintainer
  - Task arguments with configurable types
  - Confirmation text before execution

### Import Support
All new resources support `pulumi import`:
- `pulumi import lagoon:index:NotificationSlack my-slack deploy-alerts`
- `pulumi import lagoon:index:NotificationRocketChat my-rc team-chat`
- `pulumi import lagoon:index:NotificationEmail my-email ops-team`
- `pulumi import lagoon:index:NotificationMicrosoftTeams my-teams teams-alerts`
- `pulumi import lagoon:index:ProjectNotification my-assoc my-project:slack:deploy-alerts`
- `pulumi import lagoon:index:Task my-task 123`

## Example Usage

### Notifications
```python
import pulumi_lagoon as lagoon

# Create a Slack notification
slack_notify = lagoon.LagoonNotificationSlack("deploy-alerts",
    name="deploy-alerts",
    webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
    channel="#deployments",
)

# Link notification to a project
project_notify = lagoon.LagoonProjectNotification("project-slack",
    project_name="my-project",
    notification_type="slack",
    notification_name="deploy-alerts",
)
```

### Tasks
```python
import pulumi_lagoon as lagoon

# Create a command-type task
yarn_audit = lagoon.LagoonTask("yarn-audit",
    name="run-yarn-audit",
    type="command",
    service="node",
    command="yarn audit",
    project_id=project.id,
    permission="developer",
    description="Run yarn security audit",
)

# Create an image-type task
backup_task = lagoon.LagoonTask("db-backup",
    name="database-backup",
    type="image",
    service="cli",
    image="amazeeio/database-tools:latest",
    project_id=project.id,
    permission="maintainer",
    confirmation_text="This will create a full database backup. Continue?",
)
```

## Documentation

- See [docs/notifications.md](docs/notifications.md) for detailed notification resource documentation
- Task resource documentation and examples are included in the main [README.md](README.md#supported-resources)

## Requirements

- Python 3.8+
- Pulumi CLI 3.x
- Access to a Lagoon instance with API credentials

## Installation

```bash
pip install pulumi-lagoon
```

## Full Changelog

See the [commit history](https://github.com/tag1consulting/pulumi-lagoon-provider/compare/v0.1.0...v0.1.1) for all changes.

---

# Release v0.1.0 (2026-01-30)

The initial release of the Pulumi Lagoon Provider, providing a Python dynamic provider for managing [Lagoon](https://www.lagoon.sh/) resources as infrastructure-as-code.

## Highlights

This release provides a complete, working dynamic provider that enables declarative management of Lagoon hosting platform resources using Pulumi.

## Features

### Core Resources
- **LagoonProject** - Create and manage Lagoon projects with full CRUD support
- **LagoonEnvironment** - Manage environments (branch/PR deployments)
- **LagoonVariable** - Manage project and environment-level variables with build/runtime/global scopes

### Deploy Target Resources
- **LagoonDeployTarget** - Manage Kubernetes cluster deploy targets
- **LagoonDeployTargetConfig** - Configure deployment routing to specific clusters based on branch patterns

### Infrastructure
- GraphQL client with comprehensive error handling
- CORS support and TLS bypass for local development
- Token refresh handling for Lagoon's 5-minute token expiration
- Resource import support for adopting existing Lagoon infrastructure

### Examples & Automation
- **simple-project/** - Basic provider usage example
- **single-cluster/** - Complete Lagoon stack on a single Kind cluster
- **multi-cluster/** - Production-like deployment with separate prod/nonprod clusters
- Makefile automation for common operations
- Port-forward management and health checks

### Testing
- 300+ unit tests with 95% code coverage
- Integration test framework
- CI/CD pipeline via GitHub Actions

## Requirements

- Python 3.8+
- Pulumi CLI 3.x
- Access to a Lagoon instance with API credentials

## Installation

```bash
pip install pulumi-lagoon
```

## Quick Start

```python
import pulumi_lagoon as lagoon

# Create a Lagoon project
project = lagoon.LagoonProject("my-site",
    name="my-site",
    git_url="git@github.com:org/repo.git",
    deploytarget_id=1,
    production_environment="main",
    branches="^(main|develop)$",
)

# Create an environment
env = lagoon.LagoonEnvironment("production",
    project_id=project.id,
    name="main",
    environment_type="production",
    deploy_type="branch",
)

# Add a variable
var = lagoon.LagoonVariable("api-key",
    project_id=project.id,
    name="API_KEY",
    value="secret-value",
    scope="runtime",
)
```

## Known Limitations

- This is a **dynamic provider** (Python-based), not a native provider
- Dynamic providers run in a subprocess and cannot access Pulumi config secrets directly - use environment variables for `LAGOON_TOKEN`

## License

Apache License 2.0

## Acknowledgments

Built for the Lagoon community by Tag1 Consulting.
