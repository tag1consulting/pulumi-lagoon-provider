# Domain-Based Access for Lagoon Test Cluster

## Quick Start

### 1. Deploy Cluster
```bash
cd /home/gchaix/repos/pulumi-lagoon-provider/test-cluster
pulumi up
```

### 2. Configure DNS (Both WSL and Windows)

**Run this in WSL:**
```bash
./scripts/update-all-hosts.sh
```

**Then run this in Windows PowerShell as Administrator:**
```powershell
# Copy the command shown by the script above, it will look like:
powershell.exe -ExecutionPolicy Bypass -File "\\wsl$\Ubuntu\home\gchaix\repos\pulumi-lagoon-provider\test-cluster\scripts\update-windows-hosts.ps1"
```

### 3. Access Services

| Service | URL | Purpose |
|---------|-----|---------|
| Lagoon API | http://api.lagoon.test/graphql | GraphQL API endpoint |
| Keycloak | http://keycloak.lagoon.test/auth | Authentication |
| Lagoon UI | http://ui.lagoon.test | Web interface |
| Harbor | http://harbor.lagoon.test | Container registry |

## Why Two Hosts Files?

```
┌─────────────────────────────────────────────────────────────┐
│                      Windows Host OS                         │
│                                                              │
│  C:\Windows\System32\drivers\etc\hosts                      │
│  127.0.0.1 api.lagoon.test                                  │
│  127.0.0.1 keycloak.lagoon.test                             │
│  127.0.0.1 ui.lagoon.test                                   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                   WSL (Ubuntu)                        │  │
│  │                                                       │  │
│  │  /etc/hosts                                          │  │
│  │  127.0.0.1 api.lagoon.test                           │  │
│  │  127.0.0.1 keycloak.lagoon.test                      │  │
│  │  127.0.0.1 ui.lagoon.test                            │  │
│  │                                                       │  │
│  │  ┌────────────────────────────────────────────┐     │  │
│  │  │  kind cluster (lagoon-test)                 │     │  │
│  │  │                                             │     │  │
│  │  │  Port 80 → ingress-nginx                   │     │  │
│  │  │             ↓                               │     │  │
│  │  │  api.lagoon.test → lagoon-core-api         │     │  │
│  │  │  keycloak.lagoon.test → keycloak           │     │  │
│  │  │  ui.lagoon.test → lagoon-core-ui           │     │  │
│  │  └────────────────────────────────────────────┘     │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘

Windows Browser Request:
1. Browser looks up ui.lagoon.test
2. Windows hosts file resolves to 127.0.0.1
3. Request goes to localhost:80
4. WSL forwards to kind (extraPortMappings)
5. ingress-nginx routes to lagoon-core-ui
6. Response back through same path

WSL curl Request:
1. curl looks up api.lagoon.test
2. WSL /etc/hosts resolves to 127.0.0.1
3. Request goes to localhost:80 (WSL localhost = kind)
4. ingress-nginx routes to lagoon-core-api
5. Response returned
```

## Testing

### From WSL:
```bash
curl http://api.lagoon.test/graphql
curl http://keycloak.lagoon.test/auth
curl http://ui.lagoon.test
```

### From Windows Browser:
```
http://ui.lagoon.test
http://keycloak.lagoon.test/auth
http://harbor.lagoon.test
```

## Troubleshooting

### "Server not found" in Windows Browser
**Problem**: Didn't update Windows hosts file
**Fix**: Run the PowerShell script as Administrator

### "Server not found" in WSL curl
**Problem**: Didn't update WSL /etc/hosts
**Fix**: Run `./scripts/update-hosts.sh`

### "Connection refused"
**Problem**: Cluster not running or ports not mapped
**Fix**:
```bash
# Check cluster
kubectl --context kind-lagoon-test get pods -A

# Check port mapping
docker ps --filter "name=lagoon-test-control-plane" --format "{{.Ports}}"
# Should show 0.0.0.0:80->80/tcp
```

### Ingress not routing correctly
**Fix**:
```bash
# Check ingress resources
kubectl --context kind-lagoon-test get ingress -n lagoon

# Check ingress controller
kubectl --context kind-lagoon-test get pods -n ingress-nginx
```

## Manual Hosts File Updates

If scripts don't work, manually add these entries:

### WSL `/etc/hosts`:
```bash
sudo nano /etc/hosts
# Add this line:
127.0.0.1 api.lagoon.test keycloak.lagoon.test ui.lagoon.test harbor.lagoon.test
```

### Windows `C:\Windows\System32\drivers\etc\hosts`:
```
1. Open Notepad as Administrator
2. Open C:\Windows\System32\drivers\etc\hosts
3. Add this line:
127.0.0.1 api.lagoon.test keycloak.lagoon.test ui.lagoon.test harbor.lagoon.test
4. Save
```

## Removing Domains

### WSL:
```bash
sudo sed -i '/# Lagoon test cluster/d' /etc/hosts
sudo sed -i '/lagoon.test/d' /etc/hosts
```

### Windows PowerShell (as Administrator):
```powershell
$hostsFile = "C:\Windows\System32\drivers\etc\hosts"
$content = Get-Content $hostsFile | Where-Object { $_ -notmatch "lagoon.test" -and $_ -notmatch "Lagoon test cluster" }
$content | Set-Content $hostsFile
```

## Credentials

See pulumi outputs:
```bash
pulumi stack output --show-secrets
```

Or check the main README for default credentials.

## More Information

See `/home/gchaix/repos/pulumi-lagoon-provider/memory-bank/domain-based-access.md` for complete documentation.
