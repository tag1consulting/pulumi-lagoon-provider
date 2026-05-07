---
title: Simple Project
parent: Examples
nav_order: 1
---

# Simple Project Example

The `simple-project` example demonstrates the Lagoon provider API against a live Lagoon instance. It covers the most commonly used resource types and is the fastest way to verify that the provider is working correctly with your credentials.

Source: [`examples/simple-project/`](https://github.com/tag1consulting/pulumi-lagoon-provider/tree/main/examples/simple-project)

## What It Demonstrates

- Creating a Lagoon project
- Creating production and development environments
- Setting project-level variables with different scopes
- Configuring all four notification types (Slack, Email, RocketChat, Microsoft Teams)
- Linking notifications to a project via `ProjectNotification`

## Resources Created

| Resource | Count | Notes |
|----------|-------|-------|
| `Project` | 1 | With git URL and deploy target |
| `Environment` | 2 | Production (`main`) and development (`develop`) |
| `Variable` | 3 | Mix of build, runtime, and global scopes |
| `NotificationSlack` | 1 | Slack webhook |
| `NotificationEmail` | 1 | Email address |
| `NotificationRocketChat` | 1 | RocketChat webhook |
| `NotificationMicrosoftTeams` | 1 | Teams webhook |
| `ProjectNotification` | 4 | One per notification, linked to the project |

## Prerequisites

- Python 3.9+
- [Pulumi CLI](https://www.pulumi.com/docs/install/) installed
- Access to a Lagoon instance with API credentials
- A valid Lagoon API token or JWT secret

## Quick Start

```bash
# Clone the repository
git clone https://github.com/tag1consulting/pulumi-lagoon-provider.git
cd pulumi-lagoon-provider/examples/simple-project

# Create a Python virtual environment
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Initialize a Pulumi stack
pulumi stack init dev

# Configure the provider
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql
pulumi config set lagoon:token <your-jwt-token> --secret

# Preview and deploy
pulumi preview
pulumi up
```

## Configuration

| Config Key | Description | Required |
|------------|-------------|----------|
| `lagoon:apiUrl` | GraphQL API endpoint | Yes |
| `lagoon:token` | JWT authentication token | Yes (or `lagoon:jwtSecret`) |
| `lagoon:jwtSecret` | JWT secret for auto-generated tokens | Alternative to `token` |

See the [Configuration reference](../../getting-started/configuration/) for all available options.

## Cleaning Up

```bash
pulumi destroy
pulumi stack rm dev
```

{: .warning }
> `pulumi destroy` removes all resources created by the stack, including Lagoon projects and environments. Lagoon environments may have associated data (databases, file mounts). Verify you have backups before destroying.
