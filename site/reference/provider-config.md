---
title: Provider Configuration
parent: Reference
nav_order: 1
---

# Provider Configuration

## Configuration Reference

| Config Key | Environment Variable | Type | Default | Secret | Description |
|---|---|---|---|---|---|
| `lagoon:apiUrl` | `LAGOON_API_URL` | string | `https://api.lagoon.sh/graphql` | Yes | Lagoon GraphQL API endpoint |
| `lagoon:token` | `LAGOON_TOKEN` | string | — | Yes | Pre-configured JWT authentication token |
| `lagoon:jwtSecret` | `LAGOON_JWT_SECRET` | string | — | Yes | Lagoon core JWT secret; provider generates admin tokens automatically |
| `lagoon:jwtAudience` | `LAGOON_JWT_AUDIENCE` | string | `api.dev` | No | Audience claim used when generating tokens from `jwtSecret` |
| `lagoon:insecure` | `LAGOON_INSECURE` | bool | `false` | No | Disable TLS certificate verification |

{: .warning }
> Always use `--secret` when setting token or jwtSecret via the Pulumi CLI so that the value is encrypted in the stack state file.

## Authentication Priority

The provider resolves credentials in this order. The first value found is used; the rest are ignored.

1. `lagoon:token` — Pulumi config secret
2. `lagoon:jwtSecret` — Pulumi config secret (generates a one-hour admin token; auto-refreshed)
3. `LAGOON_TOKEN` — environment variable
4. `LAGOON_JWT_SECRET` — environment variable

If none of these are set, the provider fails with an authentication error at startup.

### When to use `token` vs `jwtSecret`

Use `lagoon:token` when you have a pre-existing JWT token from the Lagoon UI or CLI (`lagoon login`). Tokens expire on a schedule set by your Lagoon instance (typically 24 hours), so they require periodic rotation.

Use `lagoon:jwtSecret` when you have access to the Lagoon core `JWTSECRET` Kubernetes secret — typically in internal or self-hosted setups. The provider generates a fresh token before each operation and refreshes it automatically, so there is no token expiry concern. This is the recommended approach for fully automated deployments.

## Setting Configuration

### Via Pulumi CLI

```bash
# API endpoint (required for non-default Lagoon instances)
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql

# Authentication — use one or the other
pulumi config set lagoon:token <your-jwt-token> --secret
pulumi config set lagoon:jwtSecret <your-jwt-secret> --secret

# Optional: custom JWT audience
pulumi config set lagoon:jwtAudience api.dev

# Optional: disable TLS verification for self-signed certs
pulumi config set lagoon:insecure true
```

### Via Environment Variables

```bash
export LAGOON_API_URL=https://api.lagoon.example.com/graphql
export LAGOON_TOKEN=<your-jwt-token>
# or
export LAGOON_JWT_SECRET=<your-jwt-secret>
```

Environment variables are useful in CI/CD pipelines where you want to avoid storing credentials in the Pulumi stack state, or when running the same program against different environments by swapping variables without changing config.

## Retrieving the JWT Secret from Kubernetes

For self-hosted Lagoon, the JWT secret is stored in the `lagoon-core` namespace:

```bash
kubectl get secret -n lagoon-core lagoon-core-secrets \
  -o jsonpath='{.data.JWTSECRET}' | base64 -d
```

Store the output as a Pulumi secret:

```bash
pulumi config set lagoon:jwtSecret "$(kubectl get secret -n lagoon-core lagoon-core-secrets \
  -o jsonpath='{.data.JWTSECRET}' | base64 -d)" --secret
```

{: .note }
> The provider trims leading and trailing whitespace from all credential values automatically. Shell pipelines sometimes append a trailing newline, which would silently corrupt an HMAC signing key without this normalization.
