# Single-Cluster Lagoon Example

This example deploys a complete Lagoon stack to a single Kind cluster. It's a simplified version of the multi-cluster example, suitable for local development and testing.

## What Gets Deployed

- **Kind cluster** with ingress support
- **ingress-nginx** for HTTP/HTTPS routing
- **cert-manager** with self-signed certificates
- **Harbor** container registry
- **Lagoon Core** (API, UI, Keycloak, RabbitMQ, MariaDB, etc.)
- **Lagoon Remote** with build-deploy controller

## Prerequisites

- Docker
- Kind (`brew install kind` or see [kind.sigs.k8s.io](https://kind.sigs.k8s.io/))
- kubectl
- Helm
- Python 3.8+
- Pulumi CLI

## Quick Start

```bash
# From repository root
cd examples/single-cluster

# Initialize Pulumi stack
pulumi stack init dev

# Deploy
pulumi up
```

## Configuration

| Config Key | Default | Description |
|------------|---------|-------------|
| `createCluster` | `true` | Create a Kind cluster |
| `clusterName` | `lagoon` | Kind cluster name |
| `baseDomain` | `lagoon.local` | Base domain for services |
| `httpPort` | `8080` | HTTP port for ingress |
| `httpsPort` | `8443` | HTTPS port for ingress |
| `installHarbor` | `true` | Install Harbor registry |
| `installLagoon` | `true` | Install Lagoon |
| `deployTargetName` | `local-kind` | Name for the Lagoon deploy target |
| `helmTimeout` | `1800` | Helm timeout in seconds |

Example:
```bash
pulumi config set baseDomain lagoon.test
pulumi config set httpPort 80
pulumi config set httpsPort 443
```

## Accessing Services

After deployment, add these entries to `/etc/hosts`:

```
127.0.0.1 api.lagoon.local ui.lagoon.local keycloak.lagoon.local harbor.lagoon.local
```

Then access:
- **Lagoon UI**: https://ui.lagoon.local:8443
- **Lagoon API**: https://api.lagoon.local:8443/graphql
- **Keycloak**: https://keycloak.lagoon.local:8443
- **Harbor**: https://harbor.lagoon.local:8443

## Architecture

This example reuses modules from the multi-cluster example via symlinks:
- `clusters/` - Kind cluster management
- `infrastructure/` - Ingress, cert-manager
- `lagoon/` - Lagoon core and remote installation
- `registry/` - Harbor installation

The main difference from multi-cluster is that everything runs on a single cluster, simplifying the deployment and eliminating cross-cluster communication setup.

## Clean Up

```bash
# Destroy Pulumi resources
pulumi destroy

# Delete Kind cluster (if Pulumi doesn't clean it up)
kind delete cluster --name lagoon
```

## Troubleshooting

### Helm Timeouts

Lagoon core can take a while to initialize. If you see timeout errors:

```bash
pulumi config set helmTimeout 3600  # Increase to 1 hour
pulumi up
```

### Check Pod Status

```bash
kubectl get pods -A --context kind-lagoon
```

### View Lagoon Core Logs

```bash
kubectl logs -n lagoon-core -l app.kubernetes.io/name=lagoon-core --context kind-lagoon
```

## Comparison with Multi-Cluster

| Feature | Single-Cluster | Multi-Cluster |
|---------|---------------|---------------|
| Clusters | 1 | 2 (prod + nonprod) |
| Cross-cluster RabbitMQ | No | Yes (NodePort) |
| Deploy targets | 1 | 2 |
| Complexity | Simple | Production-like |
| Use case | Development | Production testing |

For production-like deployments with separate environments, see the [multi-cluster example](../multi-cluster/).
