# Simple Project Example

This example demonstrates basic usage of the Pulumi Lagoon provider to create a project.

## Prerequisites

- Pulumi CLI installed
- Access to a Lagoon instance
- Lagoon API credentials

## Configuration

```bash
# Set your Lagoon API endpoint
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql

# Set your authentication token (stored encrypted)
pulumi config set lagoon:token YOUR_TOKEN --secret

# Optional: Set your deploy target ID
pulumi config set deploytargetId 1
```

## Usage

```bash
# Preview changes
pulumi preview

# Deploy
pulumi up

# Destroy
pulumi destroy
```

## What This Creates

- A Lagoon project configured with:
  - Git repository
  - Production environment (main branch)
  - Branch and PR deployment patterns
  - Deploy target assignment

## Next Steps

Once the basic project resource is implemented, this example will be expanded to include:
- Environment creation
- Variable management
- Multiple environments (dev, staging, production)
