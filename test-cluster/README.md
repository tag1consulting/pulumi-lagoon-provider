# Lagoon Test Cluster

This directory contains a Pulumi program that creates a complete local Lagoon installation for testing the pulumi-lagoon-provider.

> **Quick Start**: For full environment setup (cluster + provider + example), use the unified setup from the repository root:
> ```bash
> # From repository root
> ./scripts/setup-complete.sh
> # Or: make setup-all
> ```

## What Gets Deployed

The Pulumi program creates:

1. **kind Kubernetes cluster** - A local 3-node cluster (1 control plane, 2 workers)
2. **ingress-nginx** - For HTTP routing to services  
3. **cert-manager** - For TLS certificate management
4. **metrics-server** - For resource metrics
5. **Harbor** - Container registry (required by Lagoon)
6. **Lagoon Core** - Main Lagoon control plane
   - API (GraphQL)
   - UI (Web interface)
   - Keycloak (Authentication)
   - Database (PostgreSQL)
   - Message broker (RabbitMQ)
7. **Lagoon Build Deploy** - Build/deploy controller

## Prerequisites

- **Docker** - Running Docker Desktop or Docker daemon
- **kind** - Kubernetes in Docker CLI
- **kubectl** - Kubernetes CLI
- **Pulumi** - Infrastructure as Code CLI
- **Python 3.8+** - For the Pulumi program

## Quick Start

### 1. Set up Python environment

```bash
cd test-cluster
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 2. Initialize Pulumi stack

```bash
pulumi stack init dev
```

### 3. Configure ports (optional)

By default, the ingress controller uses ports 80 (HTTP) and 443 (HTTPS). If these ports are already in use by other services (e.g., another local development environment like DDEV, Lando, or another web server), you can configure alternative ports:

```bash
# Use custom ports (e.g., 8080 for HTTP and 8443 for HTTPS)
pulumi config set ingressHttpPort 8080
pulumi config set ingressHttpsPort 8443
```

**Note:** When using custom ports, you'll need to include the port in URLs:
- `https://ui.lagoon.test:8443` instead of `https://ui.lagoon.test`
- `https://api.lagoon.test:8443/graphql` instead of `https://api.lagoon.test/graphql`

### 4. Deploy the cluster

```bash
pulumi up
```

**Total deployment time: ~4-5 minutes**

### 5. Access the services

After deployment, services are available at (default ports 80/443):

| Service | URL | Credentials |
|---------|-----|-------------|
| Lagoon UI | https://ui.lagoon.test | Via Keycloak |
| Lagoon API | https://api.lagoon.test/graphql | API token required |
| Keycloak | https://keycloak.lagoon.test/auth | Check pod logs |
| Harbor | https://harbor.lagoon.test | admin / Harbor12345 |

> **Note:** If you configured custom ports, add the port to the URLs (e.g., `https://ui.lagoon.test:8443`)

## Lagoon CLI Setup

To configure the lagoon CLI for the test cluster:

```bash
./scripts/setup-lagoon-cli.sh
```

This obtains an OAuth token and configures the CLI automatically. See [LAGOON_CLI_SETUP.md](./LAGOON_CLI_SETUP.md) for detailed instructions, manual configuration, and troubleshooting.

## Getting Credentials

To get the admin API token:

```bash
kubectl --context kind-lagoon-test -n lagoon get secret lagoon-core-api-admin-token -o jsonpath='{.data.token}' | base64 -d
```

To get the Keycloak admin password:

```bash
kubectl --context kind-lagoon-test logs -n lagoon deployment/lagoon-core-keycloak | grep "admin user"
```

## Cleanup

```bash
pulumi destroy
```

## Troubleshooting

### Port conflicts (ports 80/443 in use)

If you see errors about ports 80 or 443 being in use, you likely have another service (like DDEV, Lando, Apache, nginx, or another development environment) running on these ports.

**Option 1: Stop the conflicting service**
```bash
# Check what's using port 80
sudo lsof -i :80
# Check what's using port 443
sudo lsof -i :443
```

**Option 2: Use alternative ports**
```bash
# Configure custom ingress ports
pulumi config set ingressHttpPort 8080
pulumi config set ingressHttpsPort 8443
pulumi up
```

After using custom ports, access services with the port in the URL:
- `https://ui.lagoon.test:8443`
- `https://api.lagoon.test:8443/graphql`
- `https://keycloak.lagoon.test:8443/auth`
- `https://harbor.lagoon.test:8443`

### Services not accessible

Wait 2-3 minutes for pods to be ready:

```bash
kubectl --context kind-lagoon-test get pods -n lagoon
```

### /etc/hosts not updated

Manually add:

```bash
echo "127.0.0.1 api.lagoon.test ui.lagoon.test keycloak.lagoon.test harbor.lagoon.test" | sudo tee -a /etc/hosts
```

## WSL2 + Windows Browser Support

The deployment automatically detects WSL2 and configures both:
- WSL2 `/etc/hosts` pointing to `127.0.0.1`
- Windows `C:\Windows\System32\drivers\etc\hosts` pointing to your WSL2 IP

### Automatic Configuration

During `pulumi up`, the script will:
1. Detect if running in WSL2
2. Get your WSL2 IP address
3. Update WSL2 /etc/hosts
4. Attempt to update Windows hosts file via PowerShell

### Manual Windows Hosts Configuration

If automatic update fails (requires admin rights), manually add to `C:\Windows\System32\drivers\etc\hosts`:

```
<WSL2_IP> api.lagoon.test ui.lagoon.test keycloak.lagoon.test harbor.lagoon.test
```

Get your WSL2 IP with:
```bash
hostname -I | awk '{print $1}'
```

Or check Pulumi exports:
```bash
pulumi stack output wsl2_ip
```

### Windows PowerShell (Run as Administrator)

```powershell
# Add entry to Windows hosts file
$wslIP = "<your-wsl2-ip>"
Add-Content -Path C:\Windows\System32\drivers\etc\hosts -Value "$wslIP api.lagoon.test ui.lagoon.test keycloak.lagoon.test harbor.lagoon.test"
```
