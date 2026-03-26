# Pulumi Lagoon Provider — Python SDK

[![PyPI version](https://badge.fury.io/py/pulumi-lagoon.svg)](https://pypi.org/project/pulumi-lagoon/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/tag1consulting/pulumi-lagoon-provider/blob/main/LICENSE)

A Pulumi provider for managing [Lagoon](https://www.lagoon.sh/) resources as infrastructure-as-code.

## Installation

```bash
pip install pulumi-lagoon
```

## Configuration

```bash
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql
pulumi config set --secret lagoon:token YOUR_TOKEN
```

Or via environment variables:

```bash
export LAGOON_API_URL=https://api.lagoon.example.com/graphql
export LAGOON_TOKEN=YOUR_TOKEN
```

## Supported Resources

| Resource | Description |
|----------|-------------|
| `Project` | Lagoon projects (applications/sites) |
| `Environment` | Environments (branch/PR deployments) |
| `Variable` | Project and environment variables |
| `DeployTarget` | Kubernetes cluster deploy targets |
| `DeployTargetConfig` | Branch-pattern routing to deploy targets |
| `NotificationSlack` | Slack deployment notifications |
| `NotificationRocketChat` | RocketChat deployment notifications |
| `NotificationEmail` | Email deployment notifications |
| `NotificationMicrosoftTeams` | Microsoft Teams deployment notifications |
| `ProjectNotification` | Link notifications to projects |
| `Task` | Advanced task definitions (command and image types) |
| `Group` | Groups for organizing projects and users |

## Usage

```python
import pulumi
from pulumi_lagoon.lagoon import Project, Environment, Variable, Group

project = Project("my-site",
    name="my-drupal-site",
    git_url="git@github.com:org/repo.git",
    deploytarget_id=1,
    production_environment="main",
    branches="^(main|develop|stage)$",
)

prod_env = Environment("production",
    name="main",
    project_id=project.lagoon_id,
    deploy_type="branch",
    deploy_base_ref="main",
    environment_type="production",
)

db_config = Variable("db-host",
    name="DATABASE_HOST",
    value="mysql.production.example.com",
    project_id=project.lagoon_id,
    environment_id=prod_env.lagoon_id,
    scope="runtime",
)

team = Group("my-team",
    name="my-team",
)

pulumi.export("project_id", project.lagoon_id)
```

## Importing Existing Resources

```bash
pulumi import lagoon:lagoon:Project my-site 123
pulumi import lagoon:lagoon:Environment prod-env 123:main
pulumi import lagoon:lagoon:Variable api-key 123::API_KEY
pulumi import lagoon:lagoon:Group my-team my-team
```

## Multi-Language Support

This provider also has SDKs for [TypeScript/JavaScript](https://www.npmjs.com/package/@tag1consulting/pulumi-lagoon) and [Go](https://pkg.go.dev/github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon).

## Documentation

- [GitHub Repository](https://github.com/tag1consulting/pulumi-lagoon-provider)
- [Lagoon Documentation](https://docs.lagoon.sh/)
- [Pulumi Documentation](https://www.pulumi.com/docs/)

## License

Apache License 2.0
