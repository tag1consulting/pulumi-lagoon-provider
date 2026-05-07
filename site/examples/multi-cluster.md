---
title: Multi-Cluster
parent: Examples
nav_order: 3
---

# Multi-Cluster Example

The `multi-cluster` example deploys a production-like Lagoon setup across two Kind (Kubernetes in Docker) clusters: one for production workloads and one for non-production. This mirrors a real multi-cluster architecture with isolated environments.

Source: [`examples/multi-cluster/`](https://github.com/tag1consulting/pulumi-lagoon-provider/tree/main/examples/multi-cluster)

## Architecture

| Cluster | Components | Purpose |
|---------|-----------|---------|
| Prod | Lagoon Core, Lagoon Remote, Harbor | Hosts `main` branch and production environments |
| Nonprod | Lagoon Remote only | Hosts feature branches and PR environments |

The nonprod cluster's Lagoon Remote connects to the production core API and message broker. Both clusters are created as Kind clusters on the same Docker host, so this setup requires a machine with sufficient resources.

{: .note }
> 16 GB of RAM is recommended for running both clusters simultaneously. The example can be run with 12 GB if Harbor is disabled on the nonprod cluster.

## Prerequisites

- Docker (16 GB+ RAM recommended)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- Python 3.9+
- Helm 3
- jq

## Quick Start

```bash
cd examples/multi-cluster

# Create both clusters and deploy the full stack
make multi-cluster-up

# Check status of both clusters
make multi-cluster-status
```

Startup takes 10–20 minutes as both clusters initialize and Lagoon components come up.

## Configuration

| Config Key | Default | Description |
|------------|---------|-------------|
| `baseDomain` | auto-detected | Base domain for nip.io routing |
| `installHarbor` | `true` | Install Harbor on the prod cluster |
| `installLagoon` | `true` | Install Lagoon Core and Remote |
| `helmTimeout` | `15m` | Timeout for Helm install operations |
| `createExampleProject` | `false` | Automatically create a sample Lagoon project after deployment |

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make multi-cluster-up` | Create both clusters and deploy all components |
| `make multi-cluster-down` | Destroy both clusters |
| `make multi-cluster-status` | Show status of all pods on both clusters |
| `make multi-cluster-verify` | Run connectivity and health checks across both clusters |

## Operational Scripts

The `scripts/` directory contains parameterized operational scripts. Most accept a `LAGOON_PRESET` environment variable to target either cluster:

```bash
# Check health of the prod cluster
LAGOON_PRESET=multi-prod ./scripts/check-cluster-health.sh

# Check health of the nonprod cluster
LAGOON_PRESET=multi-nonprod ./scripts/check-cluster-health.sh
```

Available presets: `single` (default), `multi-prod`, `multi-nonprod`.

## Cleaning Up

```bash
make multi-cluster-down
```

This removes both Kind clusters. Docker network resources created by Kind are also cleaned up.

{: .tip }
> If you want to reprovision only one cluster without destroying the other, use `kind delete cluster --name <cluster-name>` directly, then re-run the relevant Pulumi stack. The clusters are managed by separate Pulumi stacks and can be destroyed independently.
