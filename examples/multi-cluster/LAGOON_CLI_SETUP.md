# Lagoon CLI Setup for Multi-Cluster Example

This guide explains how to configure the [lagoon CLI](https://github.com/uselagoon/lagoon-cli) to work with the local multi-cluster Lagoon deployment.

## Prerequisites

1. **Clusters running** - Deploy with `make deploy` from this directory or `make multi-cluster-deploy` from the repository root
2. **lagoon CLI installed** - See [Installation](#installation) below

## Quick Setup

From the `examples/multi-cluster` directory, run:

```bash
./scripts/setup-lagoon-cli.sh
```

This script:
1. Obtains an OAuth token from Keycloak (using `lagoonadmin` credentials)
2. Configures the lagoon CLI with the test cluster endpoints
3. Sets the configuration as default
4. Verifies with `lagoon whoami`

### Options

```bash
./scripts/setup-lagoon-cli.sh -h              # Show help
./scripts/setup-lagoon-cli.sh -f              # Force reconfigure
./scripts/setup-lagoon-cli.sh -u user -p pass # Custom credentials
./scripts/setup-lagoon-cli.sh -n my-cluster   # Custom config name
```

## Installation

### Linux (amd64)

```bash
curl -L https://github.com/uselagoon/lagoon-cli/releases/latest/download/lagoon-cli-linux-amd64 -o lagoon
chmod +x lagoon
sudo mv lagoon /usr/local/bin/
```

### macOS (Intel)

```bash
curl -L https://github.com/uselagoon/lagoon-cli/releases/latest/download/lagoon-cli-darwin-amd64 -o lagoon
chmod +x lagoon
sudo mv lagoon /usr/local/bin/
```

### macOS (Apple Silicon)

```bash
curl -L https://github.com/uselagoon/lagoon-cli/releases/latest/download/lagoon-cli-darwin-arm64 -o lagoon
chmod +x lagoon
sudo mv lagoon /usr/local/bin/
```

### Verify Installation

```bash
lagoon version
```

## Manual Configuration

If you prefer to configure the CLI manually:

### 1. Get OAuth Token

```bash
# Get token (outputs to stdout)
./scripts/get-lagoon-token.sh -q

# Or with custom credentials
./scripts/get-lagoon-token.sh -u myuser -p mypassword -q
```

### 2. Configure CLI

```bash
TOKEN=$(./scripts/get-lagoon-token.sh -q)

lagoon config add \
    --lagoon local-test \
    --graphql http://localhost:30030/graphql \
    --ui http://localhost:31311 \
    --token "$TOKEN"

lagoon config default --lagoon local-test
```

### 3. Verify

```bash
lagoon whoami
lagoon list projects
```

## Token Refresh

OAuth tokens from Keycloak expire after approximately **5 minutes** (300 seconds).

When your token expires, you'll see authentication errors. Simply re-run:

```bash
./scripts/setup-lagoon-cli.sh -f
```

The `-f` flag forces reconfiguration, obtaining a fresh token.

## Service Endpoints

The test cluster exposes services via NodePorts to avoid TLS certificate issues:

| Service | NodePort URL | Purpose |
|---------|--------------|---------|
| Lagoon API | `http://localhost:30030/graphql` | GraphQL API endpoint |
| Keycloak | `http://localhost:30370` | OAuth authentication |
| Lagoon UI | `http://localhost:31311` | Web interface |

These HTTP endpoints work without needing to trust the test cluster's self-signed CA certificate.

### Alternative: HTTPS with Ingress

If you need to use the HTTPS ingress endpoints:

1. **Linux** - Trust the CA certificate:
   ```bash
   # Export the CA certificate first
   kubectl --context kind-lagoon-test -n cert-manager get secret lagoon-test-ca-secret -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/lagoon-test-ca.crt

   # Trust it system-wide
   sudo cp /tmp/lagoon-test-ca.crt /usr/local/share/ca-certificates/lagoon-test-ca.crt
   sudo update-ca-certificates
   ```

2. **macOS** - Trust the CA certificate:
   ```bash
   # Export the CA certificate first
   kubectl --context kind-lagoon-test -n cert-manager get secret lagoon-test-ca-secret -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/lagoon-test-ca.crt

   # Trust it system-wide
   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain /tmp/lagoon-test-ca.crt
   ```

Then configure with HTTPS URLs:
```bash
lagoon config add \
    --lagoon local-test-https \
    --graphql https://api.lagoon.test/graphql \
    --ui https://ui.lagoon.test \
    --token "$TOKEN"
```

## Troubleshooting

### "lagoon CLI is not installed"

Install the CLI following the [Installation](#installation) section above.

### "Cluster 'lagoon-test' not found"

The clusters aren't running. Deploy them:

```bash
cd examples/multi-cluster
make deploy

# Or from repository root:
make multi-cluster-deploy
```

### "Failed to get token from Keycloak"

1. **Check if Keycloak is ready:**
   ```bash
   kubectl --context kind-lagoon-test -n lagoon get pods | grep keycloak
   ```
   Wait for the pod to be `Running` and `Ready`.

2. **Check Keycloak connectivity:**
   ```bash
   curl -v http://localhost:30370/auth/realms/lagoon
   ```

3. **Verify NodePort is accessible:**
   ```bash
   kubectl --context kind-lagoon-test -n lagoon get svc lagoon-core-keycloak
   ```

### "Configuration verification failed"

The token may have already expired. Re-run with force:

```bash
./scripts/setup-lagoon-cli.sh -f
```

### "401 Unauthorized" errors

Your token has expired. Refresh it:

```bash
./scripts/setup-lagoon-cli.sh -f
```

### "connection refused" errors

Check that the services are running and NodePorts are accessible:

```bash
# Check all Lagoon pods
kubectl --context kind-lagoon-test -n lagoon get pods

# Check NodePort services
kubectl --context kind-lagoon-test -n lagoon get svc | grep NodePort

# Test API connectivity
curl http://localhost:30030/graphql -X POST \
    -H "Content-Type: application/json" \
    -d '{"query":"{ lagoonVersion }"}'
```

## Default Credentials

| User | Password | Role |
|------|----------|------|
| lagoonadmin | lagoonadmin | Platform Admin |

## Example Commands

After successful setup:

```bash
# Show current user
lagoon whoami

# List all projects
lagoon list projects

# Get help on adding a project
lagoon add project --help

# Get details about an environment
lagoon get environment --project myproject --environment main
```

## See Also

- [Lagoon CLI Documentation](https://docs.lagoon.sh/lagoon-cli/)
- [Multi-Cluster README](./README.md)
