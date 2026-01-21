# Simple Project Example

This example demonstrates how to use the Pulumi Lagoon provider to create and manage Lagoon resources.

## What This Example Does

Uses the `pulumi_lagoon` provider to create:
- A Lagoon **Project** connected to a Git repository
- **Production** and **Development** environments
- **Variables** at project and environment levels

## Prerequisites

- A running Lagoon instance (see `test-cluster/` or `examples/multi-cluster/`)
- Python 3.8+ with the provider installed: `pip install -e ../..`
- `curl` and `jq` for the helper scripts

## Quick Start

```bash
# 1. Ensure Lagoon is running (use test-cluster or multi-cluster)
cd ../../test-cluster && pulumi up

# 2. Come back to this example
cd ../examples/simple-project

# 3. Initialize Pulumi stack
pulumi stack init test

# 4. Configure deploy target ID (get it from Lagoon)
./scripts/list-deploy-targets.sh
pulumi config set deploytargetId <ID>

# 5. Deploy using the wrapper (handles token refresh)
./scripts/run-pulumi.sh up
```

## Configuration

| Config Key | Description | Required |
|------------|-------------|----------|
| `deploytargetId` | Lagoon deploy target ID | Yes |
| `projectName` | Custom project name | No (default: example-drupal-site) |

## What Gets Created

```
Lagoon Project: example-drupal-site
├── Environment: main (production)
│   └── Variable: DATABASE_HOST
├── Environment: develop (development)
│   └── Variable: DATABASE_HOST
└── Variable: API_BASE_URL (project-level)
```

## Helper Scripts

All scripts are symlinks to the shared `scripts/` directory:

| Script | Description |
|--------|-------------|
| `run-pulumi.sh` | Wrapper that auto-refreshes OAuth token |
| `get-lagoon-token.sh` | Get OAuth token manually |
| `list-deploy-targets.sh` | List available deploy targets |
| `ensure-deploy-target.sh` | Create deploy target if none exist |
| `check-cluster-health.sh` | Verify cluster status |
| `setup-port-forwards.sh` | Start kubectl port-forwards |

## Customization

Edit `__main__.py` to:
- Change the Git repository URL
- Add more environments
- Add more variables
- Modify branch patterns

## Troubleshooting

**Token expired**: Use `./scripts/run-pulumi.sh` which auto-refreshes tokens.

**No deploy targets**: Run `./scripts/ensure-deploy-target.sh`

**Connection issues**: Run `./scripts/setup-port-forwards.sh`

For more details, see the main [repository documentation](../../README.md).
