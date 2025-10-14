# CLAUDE.md

This file provides guidance to Claude Code when working with the Pulumi Lagoon Provider project.

## Project Overview

**pulumi-lagoon-provider** is a Pulumi dynamic provider for managing Lagoon resources (projects, environments, variables, etc.) as infrastructure-as-code.

This provider allows you to declaratively manage Lagoon hosting platform resources using Pulumi, enabling infrastructure-as-code workflows for Lagoon project management.

## Project Status

**Status**: Initial Development / Proof of Concept

This project is in early development. We're starting with a Python-based dynamic provider to validate the concept before potentially building a full native provider in Go.

## Architecture

### Phase 1: Dynamic Provider (Current)
- Python-based Pulumi dynamic provider
- Direct GraphQL API integration with Lagoon
- Supports core resources: Projects, Environments, Variables, Deploy Targets

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
├── pulumi_lagoon/           # Main package
│   ├── __init__.py         # Package exports
│   ├── client.py           # Lagoon GraphQL API client
│   ├── config.py           # Provider configuration
│   ├── project.py          # LagoonProject resource
│   ├── environment.py      # LagoonEnvironment resource
│   ├── variable.py         # LagoonVariable resource
│   └── deploytarget.py     # LagoonDeployTarget resource
├── examples/               # Example Pulumi programs
│   └── simple-project/     # Basic project example
├── tests/                  # Unit and integration tests
├── memory-bank/            # Documentation and planning
│   ├── planning.md         # Initial planning and architecture decisions
│   └── architecture.md     # Technical architecture notes
├── setup.py               # Python package configuration
├── requirements.txt       # Python dependencies
└── README.md             # Project documentation
```

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

### Short-term (Phase 1)
- [ ] Implement core resources (Project, Environment, Variable)
- [ ] GraphQL client with proper error handling
- [ ] Working examples
- [ ] Basic documentation
- [ ] Unit tests

### Medium-term (Phase 2)
- [ ] Additional resources (Groups, Notifications, Tasks)
- [ ] Integration tests
- [ ] Comprehensive documentation
- [ ] Published to PyPI

### Long-term (Phase 3)
- [ ] Native Go provider
- [ ] Multi-language SDK generation
- [ ] Production-ready release
- [ ] Community adoption

## Contributing

This is currently an early-stage project. Once the core functionality is working, we'll open it up for community contributions.

## References

- [Pulumi Dynamic Providers](https://www.pulumi.com/docs/intro/concepts/resources/dynamic-providers/)
- [Lagoon Documentation](https://docs.lagoon.sh/)
- [Lagoon GraphQL API](https://api.lagoon.sh/graphql)
