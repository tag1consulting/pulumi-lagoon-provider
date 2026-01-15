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

### 3. Deploy the cluster

```bash
pulumi up
```

**Total deployment time: ~4-5 minutes**

### 4. Access the services

After deployment, services are available at:

| Service | URL | Credentials |
|---------|-----|-------------|
| Lagoon UI | http://ui.lagoon.test | Via Keycloak |
| Lagoon API | http://api.lagoon.test/graphql | API token required |
| Keycloak | http://keycloak.lagoon.test/auth | Check pod logs |
| Harbor | http://harbor.lagoon.test | admin / Harbor12345 |

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
