---
title: Lagoon Installation & Configuration
meta_desc: Information on how to install and configure the Lagoon provider for Pulumi.
layout: package
---

## Installation

The Lagoon provider is available for the following languages:

* **TypeScript/JavaScript**: [`@tag1consulting/pulumi-lagoon`](https://www.npmjs.com/package/@tag1consulting/pulumi-lagoon) on npm
* **Python**: [`pulumi_lagoon`](https://pypi.org/project/pulumi-lagoon/) on PyPI
* **Go**: [`github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon`](https://pkg.go.dev/github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon) on pkg.go.dev

### Provider Binary

The provider binary is distributed via GitHub Releases and is installed automatically when you run `pulumi up`. To install it manually:

```bash
pulumi plugin install resource lagoon <version> --server github://api.github.com/tag1consulting/pulumi-lagoon-provider
```

Replace `<version>` with the desired release (e.g., `v0.2.8`).

## Configuration

The provider can be configured via Pulumi config or environment variables.

### Required Configuration

At least one authentication method must be provided:

| Config key | Environment variable | Description |
|---|---|---|
| `lagoon:token` | `LAGOON_TOKEN` | A pre-configured JWT authentication token |
| `lagoon:jwtSecret` | `LAGOON_JWT_SECRET` | The Lagoon core `JWTSECRET`; used to generate admin tokens on the fly |

### Optional Configuration

| Config key | Environment variable | Default | Description |
|---|---|---|---|
| `lagoon:apiUrl` | `LAGOON_API_URL` | `https://api.lagoon.sh/graphql` | The Lagoon GraphQL API endpoint |
| `lagoon:jwtAudience` | — | `api.dev` | Audience claim for generated JWT tokens |
| `lagoon:insecure` | — | `false` | Disable SSL certificate verification |

### Setting Configuration

```bash
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql
pulumi config set lagoon:token your-token --secret
```

Or via environment variables before running Pulumi:

```bash
export LAGOON_API_URL=https://api.lagoon.example.com/graphql
export LAGOON_TOKEN=your-token
```
