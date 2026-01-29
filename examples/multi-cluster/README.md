# Multi-Cluster Lagoon Example

This example deploys a complete Lagoon infrastructure across two Kind clusters:
- **Production cluster** (`lagoon-prod`): Hosts Lagoon core services, Harbor registry, and a production remote controller
- **Non-production cluster** (`lagoon-nonprod`): Hosts a development remote controller that connects to the production core

## Prerequisites

- Docker
- Kind (Kubernetes in Docker)
- kubectl
- Pulumi CLI
- Python 3.8+

## Quick Start

```bash
# From the repository root (RECOMMENDED - handles timeouts automatically)
make multi-cluster-deploy
```

Or from this directory:
```bash
make deploy
```

**Note**: The Lagoon core Helm release may timeout (~15-30 minutes) but pods usually
start successfully. The `deploy` target handles this automatically by running
refresh + up cycles until deployment completes.

To verify the deployment:
```bash
make multi-cluster-verify
# or from this directory:
make verify
```

To tear down:
```bash
make multi-cluster-down
```

### Alternative Commands

If you prefer manual control:
```bash
# Single deployment attempt (may timeout on first run)
make multi-cluster-up

# If it times out, run:
make multi-cluster-preview  # Check what's pending
# Then retry until complete
```

## Accessing Services

### Option 1: Port Forwarding (Recommended)

#### Quick Start

```bash
# Start port-forwards for all services (API, Keycloak, UI)
make port-forwards-all

# Test that everything is accessible
make test-ui
```

#### Available Make Targets

| Target | Description |
|--------|-------------|
| `make port-forwards` | Start API and Keycloak port-forwards only |
| `make port-forwards-all` | Start all port-forwards (API, Keycloak, UI) |
| `make port-forwards-stop` | Stop all port-forwards |
| `make test-ui` | Test all services via port-forward |
| `make test-api` | Test API with authentication |

#### Service URLs

| Service | URL | Notes |
|---------|-----|-------|
| Lagoon UI | http://localhost:3000 | Web interface |
| Lagoon API | http://localhost:7080/graphql | GraphQL endpoint |
| Keycloak | http://localhost:8080/auth | Authentication server |
| Harbor | http://localhost:8081 | Container registry (manual port-forward) |

#### Browser Authentication Setup (Required for UI Login)

The Lagoon UI redirects browsers to an internal Kubernetes service URL for Keycloak authentication. For browser-based login to work, you must add a hosts file entry:

**Step 1: Add hosts file entry**

```bash
# Linux/Mac: Edit /etc/hosts (requires sudo)
sudo sh -c 'echo "127.0.0.1 prod-core-lagoon-core-keycloak.lagoon-core.svc.cluster.local" >> /etc/hosts'

# Windows: Edit C:\Windows\System32\drivers\etc\hosts (as Administrator)
# Add this line:
# 127.0.0.1 prod-core-lagoon-core-keycloak.lagoon-core.svc.cluster.local
```

**Step 2: Start port-forwards**

```bash
make port-forwards-all
```

**Step 3: Access the UI**

Open http://localhost:3000 in your browser and log in as `lagoonadmin`.

To get the password:
```bash
kubectl --context kind-lagoon-prod -n lagoon-core get secret prod-core-lagoon-core-keycloak \
  -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d && echo
```

#### Verifying Port-Forward Setup

Run the test target to verify everything is working:

```bash
make test-ui
```

Expected output:
```
Testing UI accessibility...
  UI accessible: http://localhost:3000 -> HTTP 200

Testing API accessibility...
  API accessible: http://localhost:7080/graphql -> HTTP 401 (401=auth required, 200=OK)

Testing Keycloak accessibility...
  Keycloak accessible: http://localhost:8080/auth -> HTTP 200

Testing OAuth token acquisition...
  OAuth token: OK (length=1255 chars)
```

#### Manual Port-Forwards

For services not included in `make port-forwards-all`:

```bash
# Harbor (http://localhost:8081)
kubectl --context kind-lagoon-prod port-forward -n harbor svc/prod-harbor-portal 8081:80 &
```

### Option 2: Host File Entries (Direct Browser Access)

This option allows you to access the Lagoon services directly in your browser using the
configured domain names, without port-forwarding.

#### Step 1: Add hosts file entries

Add the following to your hosts file:

**Linux/Mac:** `/etc/hosts`
**Windows:** `C:\Windows\System32\drivers\etc\hosts`

```
127.0.0.1 ui.lagoon.local api.lagoon.local keycloak.lagoon.local harbor.lagoon.local webhook.lagoon.local
```

> **Note for WSL2 users:** Edit the Windows hosts file, not the Linux one, since your
> browser runs on Windows.

#### Step 2: Accept self-signed certificates for EACH domain

**IMPORTANT:** Before using the UI, you must visit each URL below in your browser and
accept the certificate warning. The browser needs to trust the self-signed certificate
for each domain separately, or cross-origin requests (CORS) will silently fail.

Visit these URLs in order and click "Advanced" → "Accept the Risk and Continue" (Firefox)
or "Advanced" → "Proceed to..." (Chrome):

1. https://keycloak.lagoon.local:8443 - You should see the Keycloak welcome page
2. https://api.lagoon.local:8443/graphql - You should see `{"errors":[{"message":"Unauthorized - Bearer token required"}]}`
3. https://ui.lagoon.local:8443 - Now you can log in

If you skip this step, the UI will load but fail to communicate with the API or Keycloak,
resulting in errors like "TypeError: can't access property 'allProjects'" or CORS failures.

#### Step 3: Log in to the Lagoon UI

Open https://ui.lagoon.local:8443 and log in:
- **Username:** `lagoonadmin`
- **Password:** Run `make show-access-info` to get the password

#### Service URLs

| Service | URL |
|---------|-----|
| Lagoon UI | https://ui.lagoon.local:8443 |
| Lagoon API | https://api.lagoon.local:8443/graphql |
| Keycloak | https://keycloak.lagoon.local:8443 |
| Harbor | https://harbor.lagoon.local:8443 |

## Architecture

```
+---------------------------+     +---------------------------+
|    lagoon-prod cluster    |     |  lagoon-nonprod cluster   |
|---------------------------|     |---------------------------|
| lagoon-core namespace:    |     |                           |
|   - API                   |     |                           |
|   - UI                    |     |                           |
|   - Keycloak              |     |                           |
|   - RabbitMQ (broker)     |<----+-- lagoon namespace:       |
|   - SSH                   |     |     - remote-controller   |
|   - webhooks              |     |       (nonprod builds)    |
|                           |     |                           |
| harbor namespace:         |     |                           |
|   - Harbor Registry       |     |                           |
|                           |     |                           |
| lagoon namespace:         |     |                           |
|   - remote-controller     |     |                           |
|     (prod builds)         |     |                           |
+---------------------------+     +---------------------------+
```

## Configuration

Configuration is managed via Pulumi config. Key settings:

```bash
# Set base domain (default: lagoon.local)
pulumi config set baseDomain lagoon.local

# Disable Harbor installation
pulumi config set installHarbor false

# Disable Lagoon installation
pulumi config set installLagoon false

# Increase Helm timeout for slow environments (default: 1800 seconds = 30 min)
pulumi config set helmTimeout 3600  # 1 hour

# Disable example project creation
pulumi config set createExampleProject false

# Customize example project name (default: drupal-example)
pulumi config set exampleProjectName my-drupal-site

# Use a different Git repository for the example project
pulumi config set exampleProjectGitUrl https://github.com/myorg/myrepo.git
```

### Example Drupal Project

By default, this example creates a Drupal project that demonstrates multi-cluster
deployment routing:

- **Production branch (`main`)** → Deploys to the production cluster (`lagoon-prod`)
- **Development branches (`develop`, `feature/*`)** → Deploy to the non-production cluster (`lagoon-nonprod`)
- **Pull requests** → Deploy to the non-production cluster

This is implemented using Lagoon's Deploy Target Configurations, which route deployments
to different Kubernetes clusters based on branch patterns and priority weights.

To disable the example project:
```bash
pulumi config set createExampleProject false
```

### Helm Timeout

Lagoon core takes a long time to initialize because it runs database migrations and
waits for all pods to be ready. If you experience timeouts:

1. Use `make deploy` instead of `make up` (handles retries automatically)
2. Or increase the timeout: `pulumi config set helmTimeout 3600`
3. Or run refresh + up after a timeout: `make refresh && make up`

## Default Credentials

| Service | Username | Password |
|---------|----------|----------|
| Keycloak Admin | admin | (check Pulumi output or generated secret) |
| Lagoon Admin | lagoonadmin | (check secret below) |
| Harbor Admin | admin | (check `harbor_admin_password` Pulumi output) |

### Getting Passwords from Secrets

```bash
# Keycloak admin password
kubectl --context kind-lagoon-prod -n lagoon-core get secret prod-core-lagoon-core-keycloak \
  -o jsonpath='{.data.KEYCLOAK_ADMIN_PASSWORD}' | base64 -d && echo

# Lagoon admin password
kubectl --context kind-lagoon-prod -n lagoon-core get secret prod-core-lagoon-core-keycloak \
  -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d && echo
```

## Authentication

Lagoon uses Keycloak for authentication. This example automatically configures:

1. **Direct Access Grants** - Enables OAuth password grant for CLI tools
2. **lagoonadmin user** - Creates a platform-owner user for API access

### CLI Authentication (Programmatic)

The easiest way to test API access is with the make target:

```bash
# From repository root
make multi-cluster-test-api

# Or from this directory
make test-api
```

For manual authentication (requires `make port-forwards` running):

```bash
# Get the Lagoon admin password
LAGOON_PASSWORD=$(kubectl --context kind-lagoon-prod -n lagoon-core get secret \
  prod-core-lagoon-core-keycloak -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d)

# Get an OAuth token
TOKEN=$(curl -s -X POST "http://localhost:8080/auth/realms/lagoon/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=lagoon-ui" \
  -d "grant_type=password" \
  -d "username=lagoonadmin" \
  -d "password=$LAGOON_PASSWORD" | jq -r '.access_token')

# Use the token with the API
curl -s http://localhost:7080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ lagoonVersion }"}'
```

### Browser Authentication

Browser-based authentication requires additional setup because the Lagoon UI redirects
to an internal Kubernetes service URL for Keycloak.

**Step 1:** Add hosts file entry:
```bash
# Add to /etc/hosts
127.0.0.1 prod-core-lagoon-core-keycloak.lagoon-core.svc.cluster.local
```

**Step 2:** Start port forwards:
```bash
# Start API and Keycloak port-forwards
make port-forwards

# Manually add UI port-forward
kubectl --context kind-lagoon-prod port-forward -n lagoon-core svc/prod-core-lagoon-core-ui 3000:3000 &
```

**Step 3:** Open http://localhost:3000 and log in as `lagoonadmin`

### How Keycloak Configuration Works

The multi-cluster example includes automatic Keycloak configuration via a Kubernetes Job
that runs after Lagoon core is installed. The job (`prod-lagoon-keycloak-config`):

1. Waits for Keycloak to be ready
2. Gets an admin token
3. Enables Direct Access Grants for the `lagoon-ui` client
4. Creates the `lagoonadmin` user with `platform-owner` role

Without Direct Access Grants, the OAuth password grant flow doesn't work, which means
CLI tools cannot authenticate using username/password. The Lagoon Helm chart does not
enable this by default.

To check the job status:
```bash
kubectl --context kind-lagoon-prod get jobs -n lagoon-core | grep keycloak-config
kubectl --context kind-lagoon-prod logs -n lagoon-core job/prod-lagoon-keycloak-config
```

## Troubleshooting

### Check pod status
```bash
# Production cluster
kubectl --context kind-lagoon-prod get pods -A | grep -E "(lagoon|harbor)"

# Non-production cluster
kubectl --context kind-lagoon-nonprod get pods -n lagoon
```

### View logs
```bash
# Lagoon API logs
kubectl --context kind-lagoon-prod logs -n lagoon-core -l app.kubernetes.io/component=api --tail=50

# Remote controller logs (prod)
kubectl --context kind-lagoon-prod logs -n lagoon -l app.kubernetes.io/name=lagoon-build-deploy --tail=50

# Remote controller logs (nonprod)
kubectl --context kind-lagoon-nonprod logs -n lagoon -l app.kubernetes.io/name=lagoon-build-deploy --tail=50
```

### Cross-cluster connectivity
The nonprod remote controller connects to the prod RabbitMQ via NodePort 30672. Verify connectivity:
```bash
kubectl --context kind-lagoon-prod get svc -n lagoon-core | grep broker
```

## File Structure

```
examples/multi-cluster/
├── __main__.py              # Main Pulumi program
├── config.py                # Configuration and constants
├── clusters/                # Kind cluster management
├── infrastructure/          # Ingress, cert-manager, CoreDNS
├── registry/                # Harbor installation
├── lagoon/                  # Lagoon core and remote installation
└── scripts/                 # Helper scripts
```

## Technical Details

### Helm Chart Versions

| Chart | Version | Notes |
|-------|---------|-------|
| ingress-nginx | 4.10.1 | Standard Kubernetes ingress |
| cert-manager | v1.14.4 | TLS certificate management |
| harbor | 1.14.2 | Container registry |
| lagoon-core | 1.59.0 | Lagoon core services |
| lagoon-remote | 0.103.0 | Remote controller |

### Service Naming Convention

Lagoon core services follow the pattern: `{release-name}-lagoon-core-{component}`

| Component | Service Name | Port |
|-----------|-------------|------|
| API | prod-core-lagoon-core-api | 80 |
| UI | prod-core-lagoon-core-ui | 3000 |
| Keycloak | prod-core-lagoon-core-keycloak | 8080 |
| Broker (RabbitMQ) | prod-core-lagoon-core-broker | 5672 |
| SSH | prod-core-lagoon-core-ssh | 22 |

### Cross-Cluster Communication

The nonprod cluster connects to the prod cluster's RabbitMQ broker via a NodePort service:

- **NodePort Service**: `prod-core-broker-external` on port 30672
- **Connection**: nonprod remote-controller → prod node IP:30672 → RabbitMQ broker

The Helm chart doesn't support setting a fixed NodePort, so a custom service is created by Pulumi.

### CoreDNS Configuration

Both clusters have CoreDNS configured to resolve `*.lagoon.local`:
- **Prod cluster**: Resolves to the ingress controller's ClusterIP
- **Nonprod cluster**: Resolves to the prod cluster's node IP (for cross-cluster access)

### Known Limitations

1. **Browser Authentication**: The Lagoon UI redirects browsers to internal Kubernetes service URLs for Keycloak authentication. When using port-forwarding, you must add a hosts file entry (see "Accessing Services" above).

2. **Self-Signed Certificates**: This example uses self-signed TLS certificates. Browsers will show security warnings. **Important:** You must accept the certificate for each domain (UI, API, Keycloak) separately before the UI will work. See "Option 2: Host File Entries" for detailed instructions.

3. **WSL2 Users**: If running on WSL2, edit the Windows hosts file (`C:\Windows\System32\drivers\etc\hosts`), not the Linux one, since browsers run on Windows.

4. **S3/MinIO**: File storage (S3) is configured with dummy values. Features requiring file storage (backups, file uploads) are disabled.

5. **Elasticsearch/Kibana**: Logging integration is configured with placeholder URLs. Log aggregation is disabled.
