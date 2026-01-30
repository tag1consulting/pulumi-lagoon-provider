# Pulumi Lagoon Provider

[![Tests](https://github.com/tag1consulting/pulumi-lagoon-provider/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/tag1consulting/pulumi-lagoon-provider/actions/workflows/test.yml)
[![Python 3.8+](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/downloads/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A Pulumi dynamic provider for managing [Lagoon](https://www.lagoon.sh/) resources as infrastructure-as-code.

## Overview

This provider enables you to manage Lagoon hosting platform resources (projects, environments, variables, etc.) using Pulumi, bringing infrastructure-as-code practices to your Lagoon workflows.

**Status**: 🚧 Early Development / Proof of Concept

## Features

- **Declarative Configuration**: Manage Lagoon resources alongside your AWS/Kubernetes infrastructure
- **State Management**: Pulumi tracks resource state and detects drift
- **Type Safety**: Python type hints for better IDE support
- **GitOps Ready**: Version control your Lagoon configurations

## Supported Resources

### Core Resources (Complete)
- `LagoonProject` - Manage Lagoon projects
- `LagoonEnvironment` - Manage environments (branches/PRs)
- `LagoonVariable` - Manage project and environment variables

### Deploy Target Resources (Complete)
- `LagoonDeployTarget` - Manage Kubernetes cluster deploy targets
- `LagoonDeployTargetConfig` - Configure project deployment routing to specific clusters based on branch patterns

### Planned
- `LagoonGroup` - Manage user groups and permissions
- `LagoonNotification` - Manage notification integrations
- `LagoonTask` - Manage tasks and backups

## Quick Start - Complete Test Environment

Set up a complete local development environment with Kind cluster, Lagoon, and the provider:

```bash
# Clone the repository
git clone https://github.com/tag1consulting/pulumi-lagoon-provider.git
cd pulumi-lagoon-provider

# Option 1: Use the setup script (recommended)
./scripts/setup-complete.sh

# Option 2: Use Make
make setup-all

# After setup, deploy the example project
make example-up
```

**What gets created:**
- Kind Kubernetes cluster (`lagoon`)
- Lagoon Core (API, UI, Keycloak, RabbitMQ, databases)
- Harbor container registry
- Ingress controller with TLS
- Python virtual environment with provider installed
- Example project ready to deploy

**Total setup time: ~15-20 minutes**

## Installation (Manual)

```bash
# Clone the repository
git clone https://github.com/tag1consulting/pulumi-lagoon-provider.git
cd pulumi-lagoon-provider

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install in development mode
pip install -e .
```

## Configuration

**For the local test cluster** (after running setup):

```bash
# The example project handles this automatically via the run-pulumi.sh wrapper
cd examples/simple-project
./scripts/run-pulumi.sh up
```

**For external Lagoon instances**, use environment variables (recommended for dynamic providers):

```bash
export LAGOON_API_URL=https://api.lagoon.example.com/graphql
export LAGOON_TOKEN=YOUR_TOKEN
```

**Note**: Dynamic providers run in a subprocess and cannot access Pulumi config secrets.
Always use environment variables for `LAGOON_TOKEN`.

## Usage

```python
import pulumi
import pulumi_lagoon as lagoon

# Create a Lagoon project
project = lagoon.LagoonProject("my-drupal-site",
    name="my-drupal-site",
    git_url="git@github.com:org/repo.git",
    deploytarget_id=1,
    production_environment="main",
    branches="^(main|develop|stage)$",
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
db_config = lagoon.LagoonVariable("database-host",
    project_id=project.id,
    environment_id=prod_env.id,
    name="DATABASE_HOST",
    value="mysql.production.example.com",
    scope="runtime",
)

# Export project details
pulumi.export("project_id", project.id)
pulumi.export("production_url", prod_env.route)
```

## Examples

See the `examples/` directory for complete examples:

- `simple-project/` - Use the Lagoon provider to create projects/environments/variables via API
- `single-cluster/` - Deploy complete Lagoon stack to a single Kind cluster
- `multi-cluster/` - Production-like deployment with separate prod/nonprod Kind clusters

### Multi-Cluster Example

The multi-cluster example demonstrates a production-like Lagoon deployment with:

- **Production cluster** (`lagoon-prod`): Lagoon core services, Harbor registry, and production workloads
- **Non-production cluster** (`lagoon-nonprod`): Development/staging workloads that connect to prod core
- **Deploy Target Configs**: Route deployments to the appropriate cluster based on branch patterns
  - `main` branch → production cluster
  - `develop`, `feature/*` branches → non-production cluster
  - Pull requests → non-production cluster

```bash
# Deploy the multi-cluster environment
make multi-cluster-deploy

# Verify deployment
make multi-cluster-verify

# Access information (URLs, credentials)
cd examples/multi-cluster && make show-access-info
```

The example creates a complete Drupal project with multi-cluster routing configured automatically.

## Importing Existing Resources

You can import existing Lagoon resources into Pulumi state using `pulumi import`. This is useful when adopting infrastructure-as-code for existing Lagoon projects.

### Import ID Formats

| Resource | Import ID Format | Example |
|----------|-----------------|---------|
| `LagoonProject` | `{numeric_id}` | `pulumi import lagoon:index:Project my-project 123` |
| `LagoonDeployTarget` | `{numeric_id}` | `pulumi import lagoon:index:DeployTarget my-target 1` |
| `LagoonEnvironment` | `{project_id}:{env_name}` | `pulumi import lagoon:index:Environment my-env 123:main` |
| `LagoonVariable` | `{project_id}:{env_id}:{var_name}` | `pulumi import lagoon:index:Variable my-var 123:456:DATABASE_HOST` |
| `LagoonVariable` (project-level) | `{project_id}::{var_name}` | `pulumi import lagoon:index:Variable my-var 123::API_KEY` |
| `LagoonDeployTargetConfig` | `{project_id}:{config_id}` | `pulumi import lagoon:index:DeployTargetConfig my-config 123:5` |

### Finding Resource IDs

Use the Lagoon CLI to find resource IDs:

```bash
# List projects and their IDs
lagoon list projects

# Get project details including environment IDs
lagoon get project --project my-project

# List variables for a project
lagoon list variables --project my-project
```

### Import Examples

```bash
# Import an existing project (ID 123)
pulumi import lagoon:index:Project my-site 123

# Import an environment named "main" from project 123
pulumi import lagoon:index:Environment prod-env 123:main

# Import a project-level variable
pulumi import lagoon:index:Variable api-key 123::API_KEY

# Import an environment-level variable (project 123, environment 456)
pulumi import lagoon:index:Variable db-host 123:456:DATABASE_HOST

# Import a deploy target config
pulumi import lagoon:index:DeployTargetConfig routing-config 123:5
```

After importing, you'll need to add the corresponding resource definition to your Pulumi code.

## Development

### Prerequisites
- Python 3.8 or later
- Pulumi CLI
- Docker (for local test cluster)
- kind CLI (for local test cluster)
- kubectl

### Quick Setup

```bash
# Complete automated setup (creates Kind cluster, installs Lagoon, provider)
make setup-all

# Or just install the provider for use with existing Lagoon
make venv provider-install
```

### Make Targets

```bash
# Setup
make setup-all          # Complete setup: venv, provider, Kind cluster, Lagoon, user, deploy target
make venv               # Create Python virtual environment
make provider-install   # Install provider in development mode
make cluster-up         # Create Kind cluster + deploy Lagoon via Pulumi
make cluster-down       # Destroy Kind cluster and Lagoon resources

# Lagoon Setup (called by setup-all)
make ensure-lagoon-admin   # Create lagoonadmin user in Keycloak
make ensure-deploy-target  # Create deploy target in Lagoon + set Pulumi config

# Example Project (simple-project)
make example-setup      # Initialize example Pulumi stack
make example-preview    # Preview changes (auto token refresh)
make example-up         # Deploy example resources (auto token refresh)
make example-down       # Destroy example resources
make example-output     # Show stack outputs

# Multi-cluster Example
make multi-cluster-up       # Create prod + nonprod Kind clusters with full Lagoon stack:
                            #   - prod cluster: lagoon-core + lagoon-remote + Harbor
                            #   - nonprod cluster: lagoon-remote only (connects to prod core)
make multi-cluster-down     # Destroy multi-cluster environment
make multi-cluster-preview  # Preview multi-cluster changes
make multi-cluster-status   # Show multi-cluster stack outputs
make multi-cluster-clusters # List all Kind clusters

# Cleanup
make clean              # Kill port-forwards, remove temp files
make clean-all          # Full cleanup: clean + destroy cluster + remove venvs

# Status
make cluster-status     # Show cluster and Lagoon status
make help               # Show all available targets
```

**Note**: All targets that interact with Lagoon automatically handle:
- Port-forward setup (temporary, cleaned up after)
- Token refresh (5-minute token expiration handled)
- Direct Access Grants enablement in Keycloak
- User and deploy target creation (if needed)

### Manual Setup

```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Install in development mode
pip install -e .

# Run tests
pytest tests/
```

### Project Structure

```
pulumi-lagoon-provider/
├── pulumi_lagoon/           # Main provider package
│   ├── __init__.py         # Package exports
│   ├── client.py           # Lagoon GraphQL API client
│   ├── config.py           # Provider configuration
│   ├── exceptions.py       # Custom exceptions
│   ├── validators.py       # Input validation
│   ├── project.py          # LagoonProject resource
│   ├── environment.py      # LagoonEnvironment resource
│   ├── variable.py         # LagoonVariable resource
│   ├── deploytarget.py     # LagoonDeployTarget resource
│   └── deploytarget_config.py  # LagoonDeployTargetConfig resource
├── examples/
│   ├── simple-project/     # API usage example (assumes Lagoon exists)
│   │   ├── __main__.py     # Creates projects/environments/variables
│   │   └── scripts/        # Helper scripts
│   ├── single-cluster/     # Single Kind cluster deployment
│   │   ├── __main__.py     # Deploys full Lagoon stack
│   │   └── (symlinks)      # Reuses multi-cluster modules
│   └── multi-cluster/      # Production-like multi-cluster deployment
│       ├── __main__.py     # Two-cluster deployment
│       ├── clusters/       # Kind cluster management
│       ├── infrastructure/ # Ingress, cert-manager, CoreDNS
│       ├── lagoon/         # Lagoon core and remote
│       ├── registry/       # Harbor installation
│       └── tests/          # Unit tests for multi-cluster config (39 tests)
├── scripts/                # Shared operational scripts
├── tests/                  # Unit and integration tests
├── memory-bank/            # Planning and architecture docs
├── Makefile                # Top-level automation
└── README.md              # This file
```

## Architecture

This is a **Pulumi dynamic provider** written in Python. It communicates with the Lagoon GraphQL API to perform CRUD operations on resources.

Key components:
- **Resource Definitions**: User-facing classes (e.g., `LagoonProject`)
- **Provider Implementations**: CRUD logic for each resource type
- **GraphQL Client**: Handles API communication
- **Configuration**: Manages API credentials and settings

For detailed architecture information, see `memory-bank/architecture.md`.

## Roadmap

### Phase 1: MVP (Complete)
- [x] Project setup and structure
- [x] GraphQL client implementation
- [x] Core resources (Project, Environment, Variable)
- [x] Basic examples with automation scripts
- [x] Test cluster setup via Pulumi (Kind + Lagoon)
- [x] Documentation

### Phase 2: Deploy Targets & Multi-Cluster (Complete)
- [x] Comprehensive error handling
- [x] Unit tests (300+ tests, 95% coverage)
- [x] Integration test framework
- [x] CI/CD pipeline (GitHub Actions)
- [x] DeployTarget resource for Kubernetes cluster management
- [x] DeployTargetConfig resource for branch-based deployment routing
- [x] Multi-cluster example (prod/nonprod separation)
- [x] Two-phase deployment pattern (infrastructure first, then API-dependent resources)
- [x] CORS and TLS configuration for browser access with self-signed certificates
- [x] Comprehensive documentation for browser access setup

### Phase 3: Production Ready (Current)
- [ ] Additional resources (Group, Notification)
- [ ] Advanced examples
- [ ] PyPI package publishing
- [ ] Community feedback integration

### Phase 4: Native Provider (Future)
- [ ] Go implementation
- [ ] Multi-language SDK generation (Python, TypeScript, Go)
- [ ] Enhanced performance

## Contributing

This project is in early development. Contributions, feedback, and bug reports are welcome!

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## Resources

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Lagoon Documentation](https://docs.lagoon.sh/)
- [Lagoon GraphQL API](https://api.lagoon.sh/graphql)

## Support

For issues and questions:
- GitHub Issues: [Create an issue](https://github.com/tag1consulting/pulumi-lagoon-provider/issues)
- Lagoon Community: [Slack](https://amazeeio.rocket.chat/)

## Acknowledgments

Built for the Lagoon community by infrastructure engineers who believe in infrastructure-as-code.
