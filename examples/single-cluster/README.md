# Single-Cluster Lagoon Deployment

This example deploys a complete Lagoon environment on a single Kind cluster, suitable for development and testing.

## Overview

This deployment creates:
- A Kind Kubernetes cluster
- Ingress-nginx controller
- Cert-manager for TLS certificates
- Harbor container registry
- Lagoon Core (API, UI, Keycloak, RabbitMQ)
- Lagoon Remote (build controller)

## Quick Start

```bash
# From repository root
cd examples/single-cluster

# Create virtual environment and install dependencies
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Initialize Pulumi stack
pulumi stack init dev

# Deploy
pulumi up
```

## Configuration

```bash
# Optional: customize cluster name (default: lagoon-test)
pulumi config set clusterName my-lagoon

# Optional: customize base domain (default: lagoon.test)
pulumi config set baseDomain lagoon.local
```

## Using the Shared Scripts

This example uses the shared scripts from the repository root:

```bash
# Check cluster health
LAGOON_PRESET=single ../../scripts/check-cluster-health.sh

# Set up port forwards
LAGOON_PRESET=single ../../scripts/setup-port-forwards.sh

# Get an OAuth token
LAGOON_PRESET=single source ../../scripts/get-token.sh

# Fix RabbitMQ password issues
LAGOON_PRESET=single ../../scripts/fix-rabbitmq-password.sh
```

Or use the convenience symlinks in the `scripts/` directory.

## Comparison to Multi-Cluster

| Feature | Single-Cluster | Multi-Cluster |
|---------|---------------|---------------|
| Kind clusters | 1 | 2 (prod + nonprod) |
| Use case | Development, testing | Production-like, multi-env |
| Complexity | Simple | More complex |
| Deploy targets | 1 | 2 |

For production-like environments with separate clusters for production and non-production workloads, see `examples/multi-cluster/`.

## Architecture

```
Kind Cluster (lagoon-test)
├── ingress-nginx (namespace: ingress-nginx)
├── cert-manager (namespace: cert-manager)
├── harbor (namespace: harbor)
└── lagoon
    ├── lagoon-core (API, UI, Keycloak, Broker)
    └── lagoon-remote (build controller)
```

## Cleanup

```bash
pulumi destroy
```

This will remove all Kubernetes resources and delete the Kind cluster.
