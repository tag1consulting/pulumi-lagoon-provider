---
title: Lagoon CLI Setup
parent: Guides
nav_order: 3
---

# Lagoon CLI Setup

The Lagoon CLI (`lagoon`) is the primary tool for discovering resource IDs, verifying connectivity, and performing operations that the provider does not expose. You will use it frequently when importing existing resources and when troubleshooting.

## Installation

Download the latest binary from the [Lagoon CLI releases page](https://github.com/uselagoon/lagoon-cli/releases).

**Linux (x86\_64)**

```bash
curl -L https://github.com/uselagoon/lagoon-cli/releases/latest/download/lagoon_linux_amd64 \
  -o /usr/local/bin/lagoon
chmod +x /usr/local/bin/lagoon
```

**macOS (Intel)**

```bash
curl -L https://github.com/uselagoon/lagoon-cli/releases/latest/download/lagoon_darwin_amd64 \
  -o /usr/local/bin/lagoon
chmod +x /usr/local/bin/lagoon
```

**macOS (Apple Silicon)**

```bash
curl -L https://github.com/uselagoon/lagoon-cli/releases/latest/download/lagoon_darwin_arm64 \
  -o /usr/local/bin/lagoon
chmod +x /usr/local/bin/lagoon
```

Verify the installation:

```bash
lagoon version
```

## Configuration

Add a named configuration for your Lagoon instance. Replace the values below with your instance's actual endpoints.

```bash
lagoon config add \
  --lagoon my-lagoon \
  --graphql https://api.lagoon.example.com/graphql \
  --hostname ssh.lagoon.example.com \
  --port 22 \
  --ui https://ui.lagoon.example.com
```

Set the new configuration as the default:

```bash
lagoon config default --lagoon my-lagoon
```

{: .note }
> You can maintain multiple named configurations (e.g., `prod`, `nonprod`) and switch between them with `lagoon config default --lagoon <name>` or by passing `--lagoon <name>` on individual commands.

## Token Setup

Generate an API token using your SSH key:

```bash
lagoon login
```

This authenticates via SSH against the configured `--hostname` and stores a short-lived JWT token in your local config. The token is valid for the duration configured on your Lagoon instance (typically 24 hours).

To use a token directly (for example, in CI):

```bash
lagoon login --token <your-jwt-token>
```

## Common Commands

### Projects

```bash
# List all projects (shows IDs)
lagoon list projects

# Get full details for a specific project (including numeric ID)
lagoon get project --project my-site
```

### Environments

```bash
# List all environments for a project
lagoon list environments --project my-site
```

### Variables

```bash
# List project-level variables
lagoon list variables --project my-site

# List environment-level variables
lagoon list variables --project my-site --environment main
```

### Deploy Targets

```bash
# List all registered deploy targets (shows numeric IDs needed for provider config)
lagoon list deploy-targets
```

### Groups

```bash
# List all groups
lagoon list groups

# List members of a specific group
lagoon list group-members --name my-team
```

### Users

```bash
# Get a user by email address (shows internal Lagoon user ID)
lagoon get user --email user@example.com
```

### Notifications

```bash
# List Slack notifications
lagoon list notification slack

# List all notification types for a project
lagoon get project-notifications --project my-site
```

## SSH Access

You can open an interactive shell inside a running Lagoon environment container over SSH:

```bash
ssh -p 32222 -t lagoon@ssh.lagoon.example.com \
  service=cli project=my-site environment=main
```

Replace `32222` with the SSH port configured for your Lagoon instance (check your `lagoon config` output).

## Troubleshooting

**TLS certificate errors**

If your Lagoon instance uses a self-signed or private CA certificate, add `--skip-tls-verify` to `lagoon config add`. For the provider itself, set `lagoon:insecure` to `true` in your Pulumi config.

**SSH key not recognized**

Ensure your SSH public key is registered in the Lagoon UI under your user account. The CLI authenticates using whatever key `ssh-agent` presents, or the default `~/.ssh/id_ed25519` / `~/.ssh/id_rsa`.

**Token expired**

Run `lagoon login` to generate a fresh token. If you are using `jwtSecret` in the provider, tokens are refreshed automatically — no manual intervention needed.

**Wrong Lagoon instance**

Check which configuration is active:

```bash
lagoon config list
```

The active configuration is marked with an asterisk. Switch with `lagoon config default --lagoon <name>`.
