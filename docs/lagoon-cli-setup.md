# Lagoon CLI Setup

This guide explains how to install and configure the [lagoon CLI](https://github.com/uselagoon/lagoon-cli) to work with a local test cluster deployed via the examples in this repository.

## Prerequisites

- A running Lagoon stack — see `examples/single-cluster/` or `examples/multi-cluster/`
- `kubectl` with access to the cluster
- `curl` and `jq` (for token retrieval)

## Quick Setup

From the repository root, run the automated setup script:

```bash
# Single-cluster (default)
./scripts/setup-lagoon-cli.sh

# Multi-cluster (production cluster)
LAGOON_PRESET=multi-prod ./scripts/setup-lagoon-cli.sh

# Force reconfigure if already set up
./scripts/setup-lagoon-cli.sh -f
```

This script:
1. Checks that the lagoon CLI is installed
2. Obtains an OAuth token from Keycloak
3. Runs `lagoon config add` with the correct endpoints
4. Verifies with `lagoon whoami`

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

### Verify

```bash
lagoon version
```

## Manual Configuration

If you prefer to configure the CLI manually or need to understand what the script does:

### Step 1: Start port-forwards

The Lagoon API and Keycloak are accessible via NodePort or port-forward:

```bash
# Start port-forwards for Keycloak (8080) and API (7080)
./scripts/setup-port-forwards.sh

# Or with multi-cluster preset
LAGOON_PRESET=multi-prod ./scripts/setup-port-forwards.sh
```

### Step 2: Get an OAuth token

```bash
# Get token (outputs to stdout)
export LAGOON_TOKEN=$(./scripts/get-token.sh)

# Or with explicit credentials
LAGOON_TOKEN=$(LAGOON_USERNAME=lagoonadmin LAGOON_PASSWORD=lagoonadmin ./scripts/get-token.sh)
```

### Step 3: Configure the CLI

```bash
lagoon config add \
    --lagoon local-test \
    --graphql http://localhost:7080/graphql \
    --ui http://localhost:8080 \
    --token "$LAGOON_TOKEN"

lagoon config default --lagoon local-test
```

### Step 4: Verify

```bash
lagoon whoami
lagoon list projects
```

## Token Refresh

Keycloak tokens expire after approximately **5 minutes**. When you see authentication errors, refresh your token:

```bash
./scripts/setup-lagoon-cli.sh -f
```

## TLS Certificate Handling

The test clusters use self-signed TLS certificates. The `setup-port-forwards.sh` script forwards to HTTP endpoints, avoiding TLS entirely. If you need to use HTTPS ingress endpoints:

### Linux — Trust the CA

```bash
# Export the CA certificate from the cluster
KUBE_CONTEXT=kind-lagoon  # or kind-lagoon-prod for multi-cluster
kubectl --context $KUBE_CONTEXT -n cert-manager \
  get secret lagoon-ca-secret -o jsonpath='{.data.ca\.crt}' | base64 -d \
  > /tmp/lagoon-ca.crt

# Add to system trust store
sudo cp /tmp/lagoon-ca.crt /usr/local/share/ca-certificates/lagoon-ca.crt
sudo update-ca-certificates
```

### macOS — Trust the CA

```bash
# Export the CA certificate from the cluster
KUBE_CONTEXT=kind-lagoon  # or kind-lagoon-prod for multi-cluster
kubectl --context $KUBE_CONTEXT -n cert-manager \
  get secret lagoon-ca-secret -o jsonpath='{.data.ca\.crt}' | base64 -d \
  > /tmp/lagoon-ca.crt

# Add to system Keychain
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain /tmp/lagoon-ca.crt
```

## Adding an SSH Key

To use `lagoon deploy`, `lagoon ssh`, or git push operations, add your SSH public key to your Lagoon user:

### Via Lagoon UI

1. Open the Lagoon UI in your browser (check `pulumi stack output lagoon_ui_url`)
2. Log in as `lagoonadmin` (password from the cluster secret)
3. Go to **Settings** → **SSH Keys**
4. Click **Add SSH Key**, paste your public key, and save

### Via Lagoon CLI

```bash
# Add your default SSH public key
lagoon add ssh-key \
  --key-value "$(cat ~/.ssh/id_ed25519.pub)" \
  --key-name "my-dev-key"

# Verify it was added
lagoon list ssh-keys
```

### Find Your SSH Public Key

```bash
# Ed25519 key (recommended)
cat ~/.ssh/id_ed25519.pub

# RSA key (alternative)
cat ~/.ssh/id_rsa.pub
```

## SSH Access to Environments

After adding your SSH key, connect to Lagoon environments via SSH:

```bash
# Find the SSH host and port
pulumi stack output lagoon_api_url  # API URL (for reference)

# SSH into an environment (format: <environment>.<project>@<ssh-host> -p <ssh-port>)
# The SSH service is exposed as a NodePort — find the port.
# Use the KUBE_CONTEXT and LAGOON_NAMESPACE that match your preset:
#   single-cluster:  KUBE_CONTEXT=kind-lagoon    LAGOON_NAMESPACE=lagoon-core
#   multi-cluster:   KUBE_CONTEXT=kind-lagoon-prod  LAGOON_NAMESPACE=lagoon-core
kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get svc | grep -i ssh

# Connect (example — actual port will differ)
ssh -p <nodeport> main.my-project@localhost
```

## Example Commands

After successful setup:

```bash
# Show current user
lagoon whoami

# List all projects
lagoon list projects

# Get details about a project
lagoon get project --project my-project

# List environments for a project
lagoon list environments --project my-project

# List variables for a project
lagoon list variables --project my-project

# Add a new project
lagoon add project --help

# Deploy an environment
lagoon deploy branch --project my-project --branch main
```

## Troubleshooting

### "lagoon CLI is not installed"

Install the CLI following the [Installation](#installation) section above.

### "Failed to get token from Keycloak"

1. Ensure the `lagoonadmin` user exists — it's created automatically by the Pulumi stack, but may need the Keycloak configuration job to complete first.

2. Check Keycloak is ready:
   ```bash
   kubectl --context kind-lagoon -n lagoon-core get pods | grep keycloak
   ```

3. Ensure port-forwards are active:
   ```bash
   curl -s http://localhost:8080/auth/realms/lagoon | jq '.realm'
   # Should return "lagoon"
   ```

4. Ensure Direct Access Grants are enabled on the `lagoon-ui` Keycloak client:
   ```bash
   ./scripts/enable-direct-access-grants.sh
   ```

### "401 Unauthorized" or "Invalid Token"

Token has expired (5-minute TTL). Refresh it:

```bash
./scripts/setup-lagoon-cli.sh -f
```

### "connection refused" on port 7080 or 8080

Port-forwards are not running. Start them:

```bash
./scripts/setup-port-forwards.sh
```

### "Config already exists"

Use `-f` to force reconfigure:

```bash
./scripts/setup-lagoon-cli.sh -f
```

## Default Credentials

| User | Password | Role |
|------|----------|------|
| `lagoonadmin` | Retrieved from cluster secret (see below) | Platform Admin |

To retrieve the `lagoonadmin` password from the cluster:

```bash
# Single-cluster
kubectl --context kind-lagoon -n lagoon-core \
  get secret lagoon-core-keycloak -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d

# Multi-cluster
kubectl --context kind-lagoon-prod -n lagoon-core \
  get secret prod-core-keycloak -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d
```

## See Also

- [Lagoon CLI Documentation](https://docs.lagoon.sh/lagoon-cli/)
- [Single-Cluster Example](../examples/single-cluster/README.md)
- [Multi-Cluster Example](../examples/multi-cluster/README.md)
- [Lagoon Documentation](https://docs.lagoon.sh/)
