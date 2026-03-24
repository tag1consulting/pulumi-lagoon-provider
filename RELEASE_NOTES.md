# Release v0.2.0 (2026-03-24)

This is a major release that replaces the Python dynamic provider with a native Go provider and generated multi-language SDKs.

## Breaking Changes

### Python SDK: Class names and import paths changed

The `pulumi_lagoon` package on PyPI now ships the **native Go provider SDK** instead of the Python dynamic provider. All class names, module paths, and resource behaviors have changed.

**Old (v0.1.x dynamic provider):**
```python
from pulumi_lagoon import LagoonProject, LagoonEnvironment, LagoonVariable
from pulumi_lagoon import LagoonNotificationSlack, LagoonProjectNotification
```

**New (v0.2.x native SDK):**
```python
from pulumi_lagoon.lagoon import Project, Environment, Variable
from pulumi_lagoon.lagoon import NotificationSlack, ProjectNotification
```

### Python 3.8 no longer supported

The native SDK requires Python 3.9 or later.

### Resource ID semantics changed

Notification resources now use numeric Lagoon IDs as the Pulumi resource ID (instead of the resource name). This may require `pulumi import` to re-adopt existing resources.

## Migration Guide

1. Upgrade: `pip install --upgrade pulumi-lagoon`
2. Update your Pulumi programs to use the new class names and import paths (see above)
3. If you have existing state with v0.1.x resources, you may need to run `pulumi refresh` or re-import resources

## Highlights

- **Native Go provider**: Built with `pulumi-go-provider` for full type safety and performance
- **Multi-language SDKs**: Generated Python, TypeScript, and Go SDKs from a single schema
- **Improved correctness**: Proper Read/Diff/Update lifecycle for all resources; idempotent creates; graceful not-found handling
- **Comprehensive tests**: 200+ unit tests covering all resource types
- **TypeScript SDK**: `@tag1consulting/pulumi-lagoon` on npm
- **Go SDK**: `github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon`

## New Features

- All resources support create-or-update semantics (idempotent)
- `DeployTargetConfig` resource for managing deploy target configurations
- Improved Diff logic prevents spurious updates for optional fields with API defaults
- `ErrNotFound` handling in Read methods marks deleted resources for re-creation

## Resources

All resources from v0.1.x are available under the new `pulumi_lagoon.lagoon` module:
- `Project` (was `LagoonProject`)
- `Environment` (was `LagoonEnvironment`)
- `Variable` (was `LagoonVariable`)
- `DeployTarget` (was `LagoonDeployTarget`)
- `DeployTargetConfig` (was `LagoonDeployTargetConfig`)
- `NotificationSlack` (was `LagoonNotificationSlack`)
- `NotificationRocketChat` (was `LagoonNotificationRocketChat`)
- `NotificationEmail` (was `LagoonNotificationEmail`)
- `NotificationMicrosoftTeams` (was `LagoonNotificationMicrosoftTeams`)
- `ProjectNotification` (was `LagoonProjectNotification`)

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
