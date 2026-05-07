---
title: Single Cluster
parent: Examples
nav_order: 2
---

# Single Cluster Example

The `single-cluster` example deploys a complete Lagoon stack to a single Kind (Kubernetes in Docker) cluster on your local machine. It is designed for local development, provider testing, and learning how Lagoon components fit together.

Source: [`examples/single-cluster/`](https://github.com/tag1consulting/pulumi-lagoon-provider/tree/main/examples/single-cluster)

## What It Deploys

| Component | Details |
|-----------|---------|
| Kind cluster | Single-node local Kubernetes cluster |
| ingress-nginx | Ingress controller with NodePort for local access |
| cert-manager | TLS certificate management |
| Harbor registry | Container image registry (optional) |
| Lagoon Core | API, UI, broker, keycloak, database |
| Lagoon Remote | Build controller and SSH gateway |

## Prerequisites

- Docker (with sufficient resources — 8 GB RAM recommended)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- Python 3.9+
- Helm 3
- jq

## Quick Start

The example includes a `Makefile` with targets that handle cluster creation, stack deployment, and health checking.

```bash
cd examples/single-cluster

# Create the Kind cluster and deploy the full Lagoon stack
make cluster-up

# Wait for Lagoon to finish initializing (takes 5–10 minutes)
make wait-for-lagoon
```

Once complete, access the Lagoon UI at `http://lagoon.172.18.0.2.nip.io` (the exact IP depends on your Docker network; `make cluster-up` prints the URL).

## Configuration

The example is configured via Pulumi config. Most values have sensible defaults for local development.

| Config Key | Default | Description |
|------------|---------|-------------|
| `createCluster` | `true` | Whether to create the Kind cluster (set to `false` to use an existing cluster) |
| `clusterName` | `lagoon` | Kind cluster name |
| `baseDomain` | auto-detected | Base domain using nip.io with the cluster IP |
| `httpPort` | `80` | HTTP NodePort |
| `httpsPort` | `443` | HTTPS NodePort |
| `installHarbor` | `true` | Install Harbor container registry |
| `installLagoon` | `true` | Install Lagoon Core and Remote |

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make cluster-up` | Create cluster and deploy full Lagoon stack |
| `make cluster-down` | Destroy the Kind cluster |
| `make cluster-status` | Show cluster and pod status |
| `make check-health` | Run health checks against all Lagoon components |
| `make port-forwards` | Start kubectl port-forwards for direct service access |
| `make wait-for-lagoon` | Poll until Lagoon Core reports healthy |

## Cleaning Up

```bash
make cluster-down
```

This removes the Kind cluster entirely. No Pulumi state cleanup is needed for local development setups unless you are using a remote backend.

{: .note }
> Kind clusters use Docker bridge networking. The IP address used for nip.io-based routing is determined at cluster creation time. If your Docker network changes, recreate the cluster with `make cluster-down && make cluster-up`.
