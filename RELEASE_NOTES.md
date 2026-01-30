# Release v0.1.0

The initial release of the Pulumi Lagoon Provider - a Python dynamic provider for managing [Lagoon](https://www.lagoon.sh/) resources as infrastructure-as-code.

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
# Clone and install in development mode
git clone https://github.com/tag1consulting/pulumi-lagoon-provider.git
cd pulumi-lagoon-provider
python3 -m venv venv
source venv/bin/activate
pip install -e .
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
- Not yet published to PyPI (install from source)

## What's Next

- PyPI package publishing
- Additional resources (Groups, Notifications)
- Native Go provider for multi-language SDK support

## License

Apache License 2.0

## Acknowledgments

Built for the Lagoon community by Tag1 Consulting.
