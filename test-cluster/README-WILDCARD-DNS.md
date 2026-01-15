# Wildcard DNS for *.lagoon.test

## Problem

The `/etc/hosts` file does NOT support wildcard entries:
```
# This DOES NOT WORK:
127.0.0.1 *.lagoon.test
```

You must list each domain explicitly:
```
# This works but requires manual updates:
127.0.0.1 api.lagoon.test keycloak.lagoon.test ui.lagoon.test harbor.lagoon.test
```

## Solutions

### Option 1: dnsmasq (Recommended for WSL)

A lightweight DNS server that runs locally and supports wildcards.

**Pros:**
- Wildcard support: `*.lagoon.test` → `127.0.0.1`
- Automatic for all subdomains
- Fast and lightweight
- Easy to set up in WSL

**Cons:**
- Only works for WSL (Windows still needs hosts file)
- Requires running a service
- Modifies system DNS resolution

**Setup:**
```bash
cd /home/gchaix/repos/pulumi-lagoon-provider/test-cluster
./scripts/setup-dnsmasq.sh
```

**Test:**
```bash
dig api.lagoon.test
dig anything.lagoon.test  # Any subdomain works!
curl http://api.lagoon.test/graphql
```

**For Windows Browsers:** Still need to update Windows hosts file manually

### Option 2: Acrylic DNS Proxy (For Windows)

Similar to dnsmasq but for Windows.

**Download:** http://mayakron.altervista.org/wikibase/show.php?id=AcrylicHome

**Configuration:**
1. Install Acrylic DNS Proxy
2. Edit `AcrylicHosts.txt`:
   ```
   127.0.0.1 *.lagoon.test
   ```
3. Set Windows DNS to `127.0.0.1`
4. Restart Acrylic service

**Pros:**
- Works for Windows browsers
- Wildcard support
- Can also handle WSL (WSL uses Windows DNS)

**Cons:**
- Windows-only
- More complex setup
- Changes system DNS

### Option 3: Manual Hosts File (Current Approach)

List each domain explicitly.

**Pros:**
- No additional software
- Simple and reliable
- Works everywhere

**Cons:**
- No wildcard support
- Must update for each new subdomain
- Two files to maintain (WSL + Windows)

**Current Scripts:**
```bash
# WSL
./scripts/update-hosts.sh

# Windows (PowerShell as Admin)
./scripts/update-windows-hosts.ps1
```

## Comparison

| Solution | WSL | Windows | Wildcard | Complexity |
|----------|-----|---------|----------|------------|
| hosts file | ✓ | ✓ | ✗ | Low |
| dnsmasq | ✓ | ✗ | ✓ | Medium |
| Acrylic | ✓ | ✓ | ✓ | Medium |
| Both dnsmasq + Acrylic | ✓ | ✓ | ✓ | High |

## Recommended Approach

### For Current Needs (Few Domains)
**Use hosts files** - Simple, reliable, no additional software

Domains needed:
- api.lagoon.test
- keycloak.lagoon.test
- ui.lagoon.test
- harbor.lagoon.test

### For Future (Many Subdomains)
**Use dnsmasq in WSL** for development

When you need many Lagoon projects with their own subdomains:
- project1-main.lagoon.test
- project1-dev.lagoon.test
- project2-prod.lagoon.test
- etc.

## How DDEV Does It

DDEV uses similar approaches:

1. **ddev-router** container runs dnsmasq
2. Listens on 127.0.0.1:53
3. Resolves `*.ddev.site` → 127.0.0.1
4. On macOS: Automatically configures `/etc/resolver/ddev.site`
5. On Windows: Prompts to install certificate and configure DNS
6. On Linux: Uses systemd-resolved or dnsmasq

For Lagoon test cluster, we're essentially replicating this but simpler since we only need `*.lagoon.test`.

## Implementation Status

✅ **Implemented**: Manual hosts files (Option 3)
- Scripts: `update-hosts.sh`, `update-windows-hosts.ps1`, `update-all-hosts.sh`
- Works for current needs (4 domains)

✅ **Implemented**: dnsmasq setup script (Option 1)
- Script: `setup-dnsmasq.sh`
- Ready to use if wildcards needed

❌ **Not Implemented**: Acrylic setup (Option 2)
- Manual setup required
- Only needed if Windows wildcard support required

## Decision

**For now**: Stick with hosts files approach
- Simple, reliable, works everywhere
- Only 4 domains to manage
- No extra services to run

**Future**: Switch to dnsmasq if:
- Deploying multiple Lagoon projects
- Each project has multiple environments
- Need project-specific subdomains (pr-123.project.lagoon.test)

## Quick Commands

### Current Approach (Hosts Files):
```bash
cd /home/gchaix/repos/pulumi-lagoon-provider/test-cluster
./scripts/update-all-hosts.sh
# Then run PowerShell script on Windows as shown
```

### Wildcard Approach (dnsmasq):
```bash
cd /home/gchaix/repos/pulumi-lagoon-provider/test-cluster
./scripts/setup-dnsmasq.sh
# Windows still needs hosts file for browser access
```

### Check Current DNS Resolution:
```bash
# Test specific domain
dig api.lagoon.test

# Test if wildcard works (only works with dnsmasq)
dig random-subdomain.lagoon.test

# Check current nameserver
cat /etc/resolv.conf
```
