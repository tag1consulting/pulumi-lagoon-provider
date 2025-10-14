# Pulumi Lagoon Provider

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

### Phase 1 (Current Development)
- `LagoonProject` - Manage Lagoon projects
- `LagoonEnvironment` - Manage environments (branches/PRs)
- `LagoonVariable` - Manage project and environment variables

### Planned
- `LagoonDeployTarget` - Manage Kubernetes cluster targets
- `LagoonGroup` - Manage user groups and permissions
- `LagoonNotification` - Manage notification integrations
- `LagoonTask` - Manage tasks and backups

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/pulumi-lagoon-provider.git
cd pulumi-lagoon-provider

# Install in development mode
pip install -e .
```

## Configuration

Configure the provider with your Lagoon API credentials:

```bash
# Set Lagoon API URL
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql

# Set authentication token (stored encrypted)
pulumi config set lagoon:token YOUR_TOKEN --secret
```

Or use environment variables:

```bash
export LAGOON_API_URL=https://api.lagoon.example.com/graphql
export LAGOON_TOKEN=YOUR_TOKEN
```

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

- `simple-project/` - Basic project setup
- `multi-environment/` - Project with multiple environments (coming soon)
- `with-eks/` - Integration with Pulumi EKS (coming soon)

## Development

### Prerequisites
- Python 3.8 or later
- Pulumi CLI
- Access to a Lagoon instance

### Setup

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
├── pulumi_lagoon/           # Main package
│   ├── __init__.py         # Package exports
│   ├── client.py           # Lagoon GraphQL API client
│   ├── config.py           # Provider configuration
│   ├── project.py          # LagoonProject resource
│   ├── environment.py      # LagoonEnvironment resource
│   └── variable.py         # LagoonVariable resource
├── examples/               # Example Pulumi programs
├── tests/                  # Tests
├── memory-bank/            # Planning and architecture docs
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

### Phase 1: MVP (Current)
- [x] Project setup and structure
- [ ] GraphQL client implementation
- [ ] Core resources (Project, Environment, Variable)
- [ ] Basic examples
- [ ] Documentation

### Phase 2: Polish
- [ ] Comprehensive error handling
- [ ] Unit and integration tests
- [ ] Additional resources (DeployTarget, Group, Notification)
- [ ] Advanced examples

### Phase 3: Production Ready
- [ ] PyPI package publishing
- [ ] CI/CD pipeline
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

[License TBD]

## Resources

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Lagoon Documentation](https://docs.lagoon.sh/)
- [Lagoon GraphQL API](https://api.lagoon.sh/graphql)

## Support

For issues and questions:
- GitHub Issues: [Create an issue](https://github.com/yourusername/pulumi-lagoon-provider/issues)
- Lagoon Community: [Slack](https://amazeeio.rocket.chat/)

## Acknowledgments

Built for the Lagoon community by infrastructure engineers who believe in infrastructure-as-code.
