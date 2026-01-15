# Simple Project Example

This example demonstrates comprehensive usage of the Pulumi Lagoon provider to create and manage Lagoon resources.

> **Tested with**: Lagoon Core v2.x, Keycloak 22.x, Kubernetes 1.27+

## What's Automated vs. Manual

### Fully Automated (via scripts)

| Task | Script | Notes |
|------|--------|-------|
| Port-forward setup | `run-pulumi.sh` | Auto-starts if needed |
| Token acquisition | `run-pulumi.sh` | Auto-refreshes before each command |
| Direct Access Grants | `run-pulumi.sh` | Auto-enables in Keycloak if needed |
| Deploy target creation | `quickstart.sh` | Creates `local-kind` if none exist |
| RabbitMQ password fix | `quickstart.sh` | Detects and fixes mismatch |
| Pulumi stack init | `quickstart.sh` | Creates `test` stack |

### Manual One-Time Setup (before first use)

| Task | How to Do It | Why Manual |
|------|--------------|------------|
| Create Kind cluster | `kind create cluster --name lagoon-test` | Cluster creation is outside provider scope |
| Install Lagoon via Helm | See test-cluster docs | Complex Helm installation with many values |
| Install Python provider | `pip install -e ../..` | Requires virtual environment setup |

### Runtime Considerations

| Issue | Cause | Solution |
|-------|-------|----------|
| Token expiration | OAuth tokens expire in 5 minutes | Use `./scripts/run-pulumi.sh` wrapper |
| Port-forwards dying | kubectl port-forwards can timeout | `run-pulumi.sh` auto-restarts them |
| Config via env vars only | Dynamic providers can't read Pulumi config secrets | Set `LAGOON_TOKEN` env var (handled by wrapper) |

## Prerequisites

- Pulumi CLI installed
- Python 3.8+ with virtual environment
- kubectl configured
- Access to a Lagoon instance (or the local Kind test cluster)
- `curl` and `jq` installed (for helper scripts)

## Quick Start (Automated)

For the fastest setup with the local Kind test cluster:

```bash
# 1. Set up Python environment (from repository root)
cd /path/to/pulumi-lagoon-provider  # Replace with actual path
python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip  # Recommended: pip 21.0+
pip install -e .

# 2. Go to example directory
cd examples/simple-project

# 3. Run the quickstart script (does everything automatically)
source ./scripts/quickstart.sh

# 4. Deploy using the wrapper script (handles token refresh)
./scripts/run-pulumi.sh up
```

Or use Make:

```bash
make setup    # First-time setup
make preview  # Preview changes
make up       # Deploy resources
```

### Why Use run-pulumi.sh?

The `run-pulumi.sh` wrapper is **strongly recommended** because:
1. **Token refresh**: OAuth tokens expire every 5 minutes; the wrapper gets a fresh token before each command
2. **Port-forward management**: Automatically starts/restarts port-forwards if they're not running
3. **Direct Access Grants**: Auto-enables in Keycloak if needed
4. **Environment setup**: Sets `LAGOON_TOKEN` and `LAGOON_API_URL` correctly

If you run `pulumi up` directly, you must ensure `LAGOON_TOKEN` is set and not expired.

The quickstart script handles:
- Port-forwards for service access
- Credential extraction from secrets
- Direct Access Grants configuration
- Token acquisition
- Deploy target setup
- Pulumi stack initialization

## Manual Setup

### Step 1: Check Cluster Health

```bash
./scripts/check-cluster-health.sh
```

This verifies:
- Kind cluster is running
- All Lagoon pods are healthy
- Services are accessible

### Step 2: Fix Common Issues

If the `lagoon-build-deploy` pod is in CrashLoopBackOff with "username or password not allowed" errors:

```bash
./scripts/fix-rabbitmq-password.sh
```

### Step 3: Set Up Port Forwards

In WSL2 or Docker Desktop environments, the Docker network may not be directly routable. Set up port-forwards:

```bash
./scripts/setup-port-forwards.sh
```

This creates:
- `localhost:8080` → Keycloak
- `localhost:7080` → Lagoon API

### Step 4: Get Credentials

Extract all credentials from Kubernetes secrets:

```bash
source ./scripts/get-cluster-credentials.sh
```

This exports:
- `LAGOON_PASSWORD` - Admin user password
- `CLIENT_SECRET` - Keycloak client secret
- `KEYCLOAK_ADMIN_PASSWORD` - Keycloak admin password

### Step 5: Enable Direct Access Grants (First Time Only)

The Keycloak client needs to allow password-based authentication:

```bash
./scripts/enable-direct-access-grants.sh
```

### Step 6: Get OAuth Token

```bash
export KEYCLOAK_URL=http://localhost:8080
export LAGOON_API_URL=http://localhost:7080/graphql
source ./scripts/get-lagoon-token.sh
```

### Step 7: Configure Deploy Target

Ensure a deploy target exists:

```bash
./scripts/list-deploy-targets.sh

# If empty, add one:
./scripts/add-deploy-target.sh local-kind https://kubernetes.default.svc
```

### Step 8: Configure Pulumi

```bash
pulumi stack init test
pulumi config set deploytargetId 1  # Use ID from list-deploy-targets.sh
```

### Step 9: Deploy

```bash
pulumi preview  # Review changes
pulumi up       # Apply changes
```

## Authentication

Lagoon uses Keycloak for authentication. You need an OAuth token to access the API.

### Token Details

- Tokens expire after **5 minutes** (300 seconds)
- For long-running operations, refresh the token with `source ./scripts/get-lagoon-token.sh`
- The provider uses `LAGOON_TOKEN` and `LAGOON_API_URL` environment variables

### Getting Credentials from Test Cluster

```bash
# All credentials at once
source ./scripts/get-cluster-credentials.sh

# Or individually:

# Lagoon admin password
kubectl --context kind-lagoon-test -n lagoon get secret lagoon-core-keycloak \
  -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d && echo

# Keycloak admin password (for enabling direct access grants)
kubectl --context kind-lagoon-test -n lagoon get secret lagoon-core-keycloak \
  -o jsonpath='{.data.KEYCLOAK_ADMIN_PASSWORD}' | base64 -d && echo

# Client secret for lagoon-ui
kubectl --context kind-lagoon-test -n lagoon get secret lagoon-core-keycloak \
  -o jsonpath='{.data.KEYCLOAK_LAGOON_UI_OIDC_CLIENT_SECRET}' | base64 -d && echo
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LAGOON_API_URL` | GraphQL API endpoint | `https://api.lagoon.test/graphql` |
| `LAGOON_TOKEN` | OAuth bearer token | (required) |
| `LAGOON_INSECURE` | Skip SSL verification | `false` |
| `KEYCLOAK_URL` | Keycloak base URL | `https://keycloak.lagoon.test` |
| `KUBE_CONTEXT` | Kubernetes context | `kind-lagoon-test` |

### For Port-Forward Access (WSL2/Docker Desktop)

```bash
export LAGOON_API_URL=http://localhost:7080/graphql
export KEYCLOAK_URL=http://localhost:8080
```

### Using Pulumi Config

```bash
# These non-secret values work with Pulumi config
pulumi config set deploytargetId 1
pulumi config set projectName my-custom-project
```

> **Note**: Due to a limitation in Pulumi dynamic providers, `lagoon:apiUrl` and
> `lagoon:token` **cannot** be read from Pulumi config (even with `--secret`).
> You **must** use environment variables `LAGOON_API_URL` and `LAGOON_TOKEN` instead.
> The `run-pulumi.sh` wrapper handles this automatically.

## What This Creates

1. **Lagoon Project** (default: `example-drupal-site`, configurable via `projectName`)
   - Git repository connection
   - Production environment: main branch
   - Branch pattern: `^(main|develop|stage)$`
   - PR pattern: `^(PR-.*)`

2. **Environments**
   - **production** (main): Production branch deployment
   - **development** (develop): Development environment

3. **Variables**
   - **API_BASE_URL**: Project-level runtime variable
   - **DATABASE_HOST**: Environment-specific variable for each env

### Known Limitations

- **auto_idle**: The `auto_idle` parameter is not supported in Lagoon's `AddEnvironmentInput`.
  It must be set via `updateEnvironment` mutation after creation (not yet implemented in provider).
- **Pulumi config secrets**: Dynamic providers run in a subprocess and cannot read Pulumi config
  secrets. Use environment variables (`LAGOON_TOKEN`, `LAGOON_API_URL`) instead.

## Helper Scripts

| Script | Description |
|--------|-------------|
| `scripts/run-pulumi.sh` | **Recommended** - Wrapper that refreshes token before running pulumi |
| `scripts/quickstart.sh` | Automated first-time setup - does everything |
| `scripts/check-cluster-health.sh` | Verify cluster status |
| `scripts/setup-port-forwards.sh` | Create kubectl port-forwards |
| `scripts/get-cluster-credentials.sh` | Extract credentials from secrets |
| `scripts/get-lagoon-token.sh` | Get OAuth token from Keycloak |
| `scripts/enable-direct-access-grants.sh` | Enable password auth in Keycloak |
| `scripts/fix-rabbitmq-password.sh` | Fix RabbitMQ auth for build-deploy |
| `scripts/add-deploy-target.sh` | Register a Kubernetes cluster |
| `scripts/list-deploy-targets.sh` | List registered deploy targets |

## Makefile Targets

For convenience, a Makefile is provided:

```bash
make help      # Show all available targets
make setup     # Run quickstart script
make preview   # Preview changes (with auto token refresh)
make up        # Deploy resources (with auto token refresh)
make destroy   # Destroy resources
make output    # Show stack outputs
make clean     # Kill port-forwards
```

## Troubleshooting

### lagoon-build-deploy CrashLoopBackOff

**Symptom**: Pod logs show "username or password not allowed" (403)

**Cause**: The `lagoon-build-deploy` secret has a placeholder password

**Fix**:
```bash
./scripts/fix-rabbitmq-password.sh
```

### Cannot Connect to Services (WSL2/Docker Desktop)

**Symptom**: Connection timeout when accessing `https://api.lagoon.test`

**Cause**: Docker network not routable from host

**Fix**:
```bash
./scripts/setup-port-forwards.sh
export LAGOON_API_URL=http://localhost:7080/graphql
export KEYCLOAK_URL=http://localhost:8080
```

### "Client not allowed for direct access grants"

**Cause**: Keycloak client needs to allow password-based authentication

**Fix**:
```bash
./scripts/enable-direct-access-grants.sh
```

### "Unauthorized - Bearer token required" or "Lagoon API token must be provided"

**Cause**: Token missing or expired (tokens last 5 minutes)

**Fix** (recommended - use the wrapper script):
```bash
./scripts/run-pulumi.sh up
```

The wrapper automatically gets a fresh token before each command.

**Alternative** (manual token refresh):
```bash
source ./scripts/get-lagoon-token.sh
pulumi up
```

### SSL Certificate Errors

**Cause**: Self-signed certificates in test cluster

**Fix**:
```bash
export LAGOON_INSECURE=true
```

Or use port-forwards which avoid SSL entirely:
```bash
./scripts/setup-port-forwards.sh
export LAGOON_API_URL=http://localhost:7080/graphql
```

### "Deploy target not found"

**Fix**:
```bash
./scripts/list-deploy-targets.sh
./scripts/add-deploy-target.sh local-kind https://kubernetes.default.svc
```

### Import Errors

```bash
# From repository root
source venv/bin/activate
pip install -e .
```

## Test Cluster Reference

If using the local Kind test cluster (`kind-lagoon-test`):

| Service | Direct URL | Port-Forward URL |
|---------|-----------|------------------|
| Lagoon API | `https://api.<cluster-ip>.nip.io/graphql` | `http://localhost:7080/graphql` |
| Keycloak | `https://keycloak.<cluster-ip>.nip.io` | `http://localhost:8080` |
| Lagoon UI | `https://ui.<cluster-ip>.nip.io` | (not typically needed) |
| Harbor | `https://harbor.<cluster-ip>.nip.io` | (not typically needed) |

> **Note**: Keycloak endpoints use `/auth/` as part of the path (e.g., `http://localhost:8080/auth/realms/lagoon/...`).
> The helper scripts handle this automatically.

Get cluster IP:
```bash
docker inspect lagoon-test-control-plane --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'
```

## Outputs

After deployment (example output with default project name):

```bash
$ pulumi stack output
project_id                  42
project_name                example-drupal-site
production_url              https://main.example-drupal-site.lagoon.test
development_url             https://develop.example-drupal-site.lagoon.test
production_environment_id   123
development_environment_id  124
```

## Customization

### Using Different Git Repository

Edit `__main__.py`:
```python
git_url="git@github.com:your-org/your-repo.git",
```

### Adding More Environments

```python
stage_env = lagoon.LagoonEnvironment("staging",
    lagoon.LagoonEnvironmentArgs(
        name="stage",
        project_id=project.id,
        deploy_type="branch",
        environment_type="development",
        # Note: auto_idle is not yet supported in the provider
    )
)
```

### Adding More Variables

```python
build_mode = lagoon.LagoonVariable("build-mode",
    lagoon.LagoonVariableArgs(
        name="NODE_ENV",
        value="production",
        project_id=project.id,
        scope="build",
    )
)
```
