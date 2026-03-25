# Pulumi Lagoon Provider

[![PyPI version](https://badge.fury.io/py/pulumi-lagoon.svg)](https://pypi.org/project/pulumi-lagoon/)
[![npm version](https://badge.npmjs.com/v/@tag1consulting/pulumi-lagoon.svg)](https://www.npmjs.com/package/@tag1consulting/pulumi-lagoon)
[![Go Reference](https://pkg.go.dev/badge/github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon.svg)](https://pkg.go.dev/github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon)
[![Go Tests](https://github.com/tag1consulting/pulumi-lagoon-provider/actions/workflows/test-go.yml/badge.svg?branch=main)](https://github.com/tag1consulting/pulumi-lagoon-provider/actions/workflows/test-go.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A Pulumi provider for managing [Lagoon](https://www.lagoon.sh/) resources as infrastructure-as-code.

## Overview

This provider enables you to manage Lagoon hosting platform resources (projects, environments, variables, deploy targets, notifications, tasks, etc.) using Pulumi, with native SDKs for Python, TypeScript/JavaScript, and Go.

**Status**: v0.2.0 — Native Go Provider

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

## Installation

### Python

```bash
pip install pulumi-lagoon
```

### TypeScript / JavaScript

```bash
npm install @tag1consulting/pulumi-lagoon
# or
yarn add @tag1consulting/pulumi-lagoon
```

### Go

```bash
go get github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon
```

## Configuration

```bash
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql
pulumi config set --secret lagoon:token YOUR_TOKEN
# or use a JWT secret for admin token generation:
pulumi config set --secret lagoon:jwtSecret YOUR_JWT_SECRET
```

Or via environment variables:

```bash
export LAGOON_API_URL=https://api.lagoon.example.com/graphql
export LAGOON_TOKEN=YOUR_TOKEN
```

## Usage

### Python

```python
import pulumi
import pulumi_lagoon as lagoon
from pulumi_lagoon.lagoon import Project, ProjectArgs, Environment, EnvironmentArgs, Variable, VariableArgs

project = Project("my-site",
    ProjectArgs(
        name="my-drupal-site",
        git_url="git@github.com:org/repo.git",
        deploytarget_id=1,
        production_environment="main",
        branches="^(main|develop|stage)$",
    )
)

prod_env = Environment("production",
    EnvironmentArgs(
        name="main",
        project_id=project.lagoon_id,
        deploy_type="branch",
        deploy_base_ref="main",
        environment_type="production",
    )
)

db_config = Variable("db-host",
    VariableArgs(
        name="DATABASE_HOST",
        value="mysql.production.example.com",
        project_id=project.lagoon_id,
        environment_id=prod_env.lagoon_id,
        scope="runtime",
    )
)

pulumi.export("project_id", project.lagoon_id)
pulumi.export("production_url", prod_env.route)
```

### TypeScript

```typescript
import * as lagoon from "@tag1consulting/pulumi-lagoon";

const project = new lagoon.lagoon.Project("my-site", {
    name: "my-drupal-site",
    gitUrl: "git@github.com:org/repo.git",
    deploytargetId: 1,
    productionEnvironment: "main",
    branches: "^(main|develop|stage)$",
});

export const projectId = project.lagoonId;
```

### Go

```go
import (
    lagoon "github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon/lagoon"
)

project, err := lagoon.NewProject(ctx, "my-site", &lagoon.ProjectArgs{
    Name:                  pulumi.String("my-drupal-site"),
    GitUrl:                pulumi.String("git@github.com:org/repo.git"),
    DeploytargetId:        pulumi.Int(1),
    ProductionEnvironment: pulumi.String("main"),
})
```

## Examples

See the `examples/` directory for complete examples:

- `simple-project/` — Create Lagoon projects, environments, variables, and notifications via the API
- `single-cluster/` — Deploy a complete Lagoon stack to a single Kind cluster
- `multi-cluster/` — Production-like deployment with separate prod/nonprod Kind clusters

### Multi-Cluster Example

```bash
# Deploy prod + nonprod Kind clusters with full Lagoon stack
make multi-cluster-up

# Verify deployment
make multi-cluster-status

# Tear down
make multi-cluster-down
```

## Importing Existing Resources

Use `pulumi import` to bring existing Lagoon resources under Pulumi management:

| Resource | Import ID Format | Example |
|----------|-----------------|---------|
| `lagoon:lagoon:Project` | `{numeric_id}` | `123` |
| `lagoon:lagoon:DeployTarget` | `{numeric_id}` | `1` |
| `lagoon:lagoon:Environment` | `{project_id}:{env_name}` | `123:main` |
| `lagoon:lagoon:Variable` | `{project_id}:{env_id}:{var_name}` | `123:456:DATABASE_HOST` |
| `lagoon:lagoon:Variable` (project-level) | `{project_id}::{var_name}` | `123::API_KEY` |
| `lagoon:lagoon:DeployTargetConfig` | `{project_id}:{config_id}` | `123:5` |
| `lagoon:lagoon:NotificationSlack` | `{name}` | `deploy-alerts` |
| `lagoon:lagoon:ProjectNotification` | `{project}:{type}:{name}` | `my-project:slack:deploy-alerts` |
| `lagoon:lagoon:Task` | `{numeric_id}` | `456` |

```bash
# Import an existing project (ID 123)
pulumi import lagoon:lagoon:Project my-site 123

# Import an environment
pulumi import lagoon:lagoon:Environment prod-env 123:main

# Import a project-level variable
pulumi import lagoon:lagoon:Variable api-key 123::API_KEY
```

Use the [Lagoon CLI](docs/lagoon-cli-setup.md) to find resource IDs:

```bash
lagoon list projects
lagoon get project --project my-project
```

## Development

### Prerequisites

- Go 1.22+
- Pulumi CLI
- Docker, Kind, kubectl (for local test clusters)

### Build

```bash
cd provider
CGO_ENABLED=0 go build ./cmd/pulumi-resource-lagoon/
```

### Test

```bash
cd provider
CGO_ENABLED=0 go test ./... -count=1
```

### Makefile Targets

```bash
make go-build       # Build the provider binary
make go-test        # Run all Go tests (198 tests)
make go-schema      # Regenerate provider schema
make go-sdk-all     # Regenerate all language SDKs
make go-sdk-python  # Regenerate Python SDK
make go-sdk-nodejs  # Regenerate TypeScript SDK
make go-sdk-go      # Regenerate Go SDK
```

### Provider Structure

```
pulumi-lagoon-provider/
├── provider/                    # Native Go provider
│   ├── cmd/pulumi-resource-lagoon/  # Provider binary entrypoint
│   ├── pkg/client/              # Lagoon GraphQL client
│   ├── pkg/config/              # Provider configuration
│   ├── pkg/resources/           # 11 resource implementations
│   └── schema.json              # Pulumi schema
├── sdk/                         # Generated multi-language SDKs
│   ├── python/python/           # Python SDK (PyPI: pulumi-lagoon)
│   ├── nodejs/nodejs/           # TypeScript SDK (npm: @tag1consulting/pulumi-lagoon)
│   └── go/go/lagoon/            # Go SDK
├── examples/
│   ├── simple-project/          # Provider usage example
│   ├── single-cluster/          # Single Kind cluster deployment
│   └── multi-cluster/           # Production-like multi-cluster deployment
├── scripts/                     # Shared operational scripts
├── docs/                        # Additional documentation
└── memory-bank/                 # Architecture and planning docs
```

## Architecture

A **native Go provider** using [`pulumi-go-provider`](https://github.com/pulumi/pulumi-go-provider) v1.3.0 with the `infer` package. The provider communicates with the Lagoon GraphQL API and generates multi-language SDKs from a single schema.

Key properties:
- All sensitive fields (`token`, `jwtSecret`, `webhook`, `value`) are marked secret and encrypted in state
- All 11 resources implement `Diff` with per-field `DetailedDiff` (Update vs UpdateReplace)
- All resources support `pulumi import`
- JWT token generation is centralized and configurable (`jwtAudience` config field)

## Contributing

Contributions, feedback, and bug reports are welcome!

1. Fork the repository
2. Create a feature branch off `main`
3. Make your changes with tests
4. Submit a pull request

## License

Apache License 2.0 — See [LICENSE](LICENSE) for details.

## Resources

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Lagoon Documentation](https://docs.lagoon.sh/)
- [Lagoon GraphQL API](https://api.lagoon.sh/graphql)

## Support

- GitHub Issues: [Create an issue](https://github.com/tag1consulting/pulumi-lagoon-provider/issues)
- Lagoon Community: [RocketChat](https://amazeeio.rocket.chat/)
