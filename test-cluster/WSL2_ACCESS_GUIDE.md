# Accessing Lagoon from Windows (WSL2)

**Environment**: WSL2 on Windows 11
**Status**: kind cluster running in WSL2 with port forwarding to Windows

## Quick Answer

Since you're running kind in WSL2, the cluster is accessible from Windows via **localhost**!

## Port Mappings

Your kind cluster has these ports exposed to Windows:

| Service | Port | Windows Access | WSL2 Access |
|---------|------|----------------|-------------|
| HTTP | 80 | `http://localhost` | `http://localhost` |
| HTTPS | 443 | `https://localhost` | `https://localhost` |
| API | 3000 | `http://localhost:3000` | `http://localhost:3000` |

## Option 1: Direct Access via localhost (Recommended)

### From Windows

Access Lagoon services directly:

**Lagoon API (Primary)**:
```
http://localhost:3000
```

**Harbor Registry**:
```
http://localhost
```

### Testing from Windows PowerShell

```powershell
# Test if Lagoon API is accessible
Invoke-WebRequest -Uri "http://localhost:3000" -Method GET

# Or use curl (if installed)
curl http://localhost:3000
```

### For pulumi-lagoon-provider

Set these environment variables (Windows PowerShell):

```powershell
$env:LAGOON_API_URL = "http://localhost:3000/graphql"
$env:LAGOON_TOKEN = "your-token-here"
```

Or in Windows CMD:
```cmd
set LAGOON_API_URL=http://localhost:3000/graphql
set LAGOON_TOKEN=your-token-here
```

## Option 2: Use Domain Names (Optional)

If you prefer using domain names like `api.lagoon.test`:

### Step 1: Edit Windows hosts file

**Location**: `C:\Windows\System32\drivers\etc\hosts`

**How to Edit**:
1. Open Notepad **as Administrator** (Right-click → Run as administrator)
2. File → Open → Navigate to `C:\Windows\System32\drivers\etc\`
3. Change file filter from "Text Documents (*.txt)" to **"All Files (*.*)"**
4. Select `hosts` file
5. Add these lines at the end:

```
127.0.0.1 api.lagoon.test
127.0.0.1 ui.lagoon.test
127.0.0.1 harbor.lagoon.test
```

6. Save the file (requires Administrator)

### Step 2: Test Access

From Windows browser or PowerShell:
```
http://api.lagoon.test:3000
http://harbor.lagoon.test
```

## Option 3: kubectl Port Forwarding (For Internal Services)

Some services are only accessible inside the cluster. Use port forwarding:

### From WSL2 Terminal

Forward Lagoon API:
```bash
kubectl --context kind-lagoon-test port-forward -n lagoon svc/lagoon-core-api 8080:80
```

Then access from Windows:
```
http://localhost:8080
```

Forward Harbor:
```bash
kubectl --context kind-lagoon-test port-forward -n harbor svc/harbor-core 8081:80
```

Then access from Windows:
```
http://localhost:8081
```

## Getting Lagoon API Token

### From WSL2

Run this in your WSL2 terminal:

```bash
cd /home/gchaix/repos/pulumi-lagoon-provider/test-cluster
./scripts/get-credentials.sh
```

This will output:
```
API URL: http://api.lagoon.test/graphql
Token: <your-token>
```

Copy the token for use in your provider configuration.

## Testing Access

### From WSL2

```bash
# Test Lagoon API
curl http://localhost:3000

# Test with actual GraphQL query
curl -X POST http://localhost:3000/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"query": "{ allProjects { name } }"}'
```

### From Windows (PowerShell)

```powershell
# Simple connectivity test
Test-NetConnection -ComputerName localhost -Port 3000

# HTTP request test
Invoke-WebRequest -Uri "http://localhost:3000" -Method GET
```

### From Windows (Browser)

Open your browser and navigate to:
```
http://localhost:3000
```

You should see a Lagoon response or GraphQL playground.

## Troubleshooting

### Port 3000 not accessible from Windows

1. Check if kind cluster is running:
   ```bash
   # In WSL2
   docker ps | grep lagoon-test
   ```

2. Verify ports are exposed:
   ```bash
   # In WSL2
   docker port lagoon-test-control-plane
   ```

3. Check Windows Firewall:
   - Windows might be blocking localhost ports
   - Try disabling Windows Firewall temporarily to test

### Hosts file changes not taking effect

1. Flush DNS cache (Windows PowerShell as Administrator):
   ```powershell
   ipconfig /flushdns
   ```

2. Restart browser

3. Verify hosts file:
   ```powershell
   Get-Content C:\Windows\System32\drivers\etc\hosts
   ```

### WSL2 IP Address Changes

If localhost doesn't work, WSL2 might have a different IP. Get it with:

```bash
# In WSL2
ip addr show eth0 | grep inet | awk '{print $2}' | cut -d/ -f1
```

Then use that IP instead of localhost from Windows.

## Recommended Setup for Development

### For Python Development (Windows)

If you're developing in Windows but testing in WSL2:

1. **Install Python provider locally** (Windows):
   ```powershell
   pip install -e .
   ```

2. **Set environment variables** (Windows):
   ```powershell
   $env:LAGOON_API_URL = "http://localhost:3000/graphql"
   $env:LAGOON_TOKEN = "your-token-here"
   ```

3. **Run Pulumi** (Windows):
   ```powershell
   pulumi up
   ```

### For Python Development (WSL2)

If you're developing in WSL2:

1. **Use localhost or 127.0.0.1**:
   ```bash
   export LAGOON_API_URL="http://localhost:3000/graphql"
   export LAGOON_TOKEN="your-token-here"
   ```

2. **Run Pulumi** (WSL2):
   ```bash
   pulumi up
   ```

## Summary

**Simplest approach for WSL2**:
- ✅ Use `http://localhost:3000` for Lagoon API
- ✅ Use `http://localhost` for Harbor
- ✅ No hosts file editing required!
- ✅ Works from both Windows and WSL2

**If you want domain names**:
- Edit `C:\Windows\System32\drivers\etc\hosts` (Windows, requires Administrator)
- Add `127.0.0.1 api.lagoon.test ui.lagoon.test harbor.lagoon.test`
- Flush DNS cache
- Use `http://api.lagoon.test:3000`

**Recommended**: Start with localhost (Option 1), it's simpler and works immediately!

## Next Steps

1. Get your API token:
   ```bash
   ./scripts/get-credentials.sh
   ```

2. Test connectivity:
   ```bash
   curl http://localhost:3000
   ```

3. Configure your provider and test!

---

**Note**: The `/etc/hosts` file in WSL2 is separate from the Windows hosts file. If you edit WSL2's `/etc/hosts`, it only affects WSL2, not Windows. To make domain names work in both, you need to edit the Windows hosts file.
