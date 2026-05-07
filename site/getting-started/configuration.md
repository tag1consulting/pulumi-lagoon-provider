---
title: Configuration
parent: Getting Started
nav_order: 2
---

# Configuration

The provider requires at least one authentication method and, optionally, a custom API endpoint.

## Authentication

You must supply either `lagoon:token` or `lagoon:jwtSecret`. The provider will not start without at least one of these.

| Config Key | Environment Variable | Description |
|---|---|---|
| `lagoon:token` | `LAGOON_TOKEN` | A pre-configured JWT authentication token |
| `lagoon:jwtSecret` | `LAGOON_JWT_SECRET` | The Lagoon core `JWTSECRET`; used to generate admin tokens on the fly |

{: .note }
> JWT tokens generated from `jwtSecret` expire after one hour. The provider refreshes them automatically during long-running operations.

## Optional Configuration

| Config Key | Environment Variable | Default | Description |
|---|---|---|---|
| `lagoon:apiUrl` | `LAGOON_API_URL` | `https://api.lagoon.sh/graphql` | The Lagoon GraphQL API endpoint |
| `lagoon:jwtAudience` | `LAGOON_JWT_AUDIENCE` | `api.dev` | Audience claim used when generating JWT tokens from `jwtSecret` |
| `lagoon:insecure` | `LAGOON_INSECURE` | `false` | Disable SSL certificate verification |

{: .tip }
> For development instances with self-signed TLS certificates, set `lagoon:insecure` to `true`.

## Authentication Priority

When the provider starts, it resolves credentials in this order:

1. `lagoon:token` (Pulumi config)
2. `lagoon:jwtSecret` (Pulumi config)
3. `LAGOON_TOKEN` (environment variable)
4. `LAGOON_JWT_SECRET` (environment variable)

The first value found is used.

## Setting Configuration via Pulumi CLI

```bash
# Optional: override the default API endpoint (default: https://api.lagoon.sh/graphql)
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql

# Authentication — use one of:
pulumi config set lagoon:token <your-jwt-token> --secret
pulumi config set lagoon:jwtSecret <your-jwt-secret> --secret

# Optional
pulumi config set lagoon:insecure true
```

Always use `--secret` for credentials so Pulumi encrypts them in the stack state.

## Setting Configuration via Environment Variables

```bash
export LAGOON_API_URL=https://api.lagoon.example.com/graphql
export LAGOON_TOKEN=<your-jwt-token>
```

Environment variables are useful for CI/CD pipelines where you want to avoid storing credentials in the Pulumi stack configuration.
