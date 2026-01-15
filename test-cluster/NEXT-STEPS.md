# Next Steps After Deployment

Once `pulumi up` completes successfully, follow these steps:

## 1. Configure DNS Resolution

### WSL (Required for curl/CLI):
```bash
cd /home/gchaix/repos/pulumi-lagoon-provider/test-cluster
./scripts/update-hosts.sh
```

### Windows (Required for Browser Access):
In **Windows PowerShell as Administrator**:
```powershell
# Get the exact path from the update-all-hosts.sh output, or use:
cd C:\Users\[YourUsername]  # Navigate to somewhere you can access WSL files
wsl.exe /home/gchaix/repos/pulumi-lagoon-provider/test-cluster/scripts/update-windows-hosts.ps1
```

Or manually add to `C:\Windows\System32\drivers\etc\hosts`:
```
127.0.0.1 api.lagoon.test keycloak.lagoon.test ui.lagoon.test harbor.lagoon.test
```

## 2. Verify Services

### From WSL:
```bash
# Test API
curl http://api.lagoon.test/graphql

# Test Keycloak
curl http://keycloak.lagoon.test/auth

# Test UI
curl http://ui.lagoon.test

# Test Harbor
curl http://harbor.lagoon.test
```

### From Windows Browser:
- Lagoon UI: http://ui.lagoon.test
- Keycloak: http://keycloak.lagoon.test/auth
- Harbor: http://harbor.lagoon.test
- API Playground: http://api.lagoon.test/graphql

## 3. Get Credentials

```bash
# Keycloak admin password
kubectl --context kind-lagoon-test -n lagoon logs deployment/lagoon-core-keycloak | grep 'Admin password'

# RabbitMQ password (for reference)
kubectl --context kind-lagoon-test get secret lagoon-core-broker -n lagoon -o jsonpath='{.data.RABBITMQ_PASSWORD}' | base64 -d

# Harbor (already known)
# Username: admin
# Password: Harbor12345
```

## 4. Check Cluster Health

```bash
# All pods should be Running (except completed jobs)
kubectl --context kind-lagoon-test get pods -A

# Check Lagoon namespace specifically
kubectl --context kind-lagoon-test get pods -n lagoon

# Check ingress resources
kubectl --context kind-lagoon-test get ingress -A
```

## 5. Test Lagoon API

```bash
# GraphQL introspection query
curl -X POST http://api.lagoon.test/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __schema { types { name } } }"}'
```

## 6. Access Lagoon UI

1. Open http://ui.lagoon.test in browser
2. Click "Login with Keycloak"
3. You'll be redirected to http://keycloak.lagoon.test/auth
4. Login with Keycloak admin credentials

## Troubleshooting

### "Server not found" errors:
- **WSL**: Run `./scripts/update-hosts.sh`
- **Windows**: Run PowerShell script as Administrator

### "Connection refused" errors:
```bash
# Check if cluster is running
kubectl --context kind-lagoon-test get pods -A

# Check if ports are mapped
docker ps --filter "name=lagoon-test-control-plane" --format "{{.Ports}}"
# Should show: 0.0.0.0:80->80/tcp
```

### Ingress not working:
```bash
# Check ingress controller
kubectl --context kind-lagoon-test get pods -n ingress-nginx

# Check ingress resources
kubectl --context kind-lagoon-test describe ingress -n lagoon

# Check service endpoints
kubectl --context kind-lagoon-test get endpoints -n lagoon
```

### Pods in CrashLoopBackOff:
```bash
# Check pod logs
kubectl --context kind-lagoon-test logs -n lagoon <pod-name>

# Common issues:
# - Database migrations: Check lagoon-core-api-migratedb job logs
# - RabbitMQ: Check if fix-rabbitmq-password command ran
```

## Files Reference

- **Main Pulumi Code**: `/home/gchaix/repos/pulumi-lagoon-provider/test-cluster/__main__.py`
- **Lagoon Config**: `/home/gchaix/repos/pulumi-lagoon-provider/test-cluster/config/lagoon-values.yaml`
- **kind Config**: `/home/gchaix/repos/pulumi-lagoon-provider/test-cluster/config/kind-config.yaml`
- **Hosts Scripts**: `/home/gchaix/repos/pulumi-lagoon-provider/test-cluster/scripts/`

## Documentation

- **Domain Setup**: `README-DOMAINS.md`
- **Wildcard DNS**: `README-WILDCARD-DNS.md`
- **Detailed Guide**: `/home/gchaix/repos/pulumi-lagoon-provider/memory-bank/domain-based-access.md`

## Quick Commands

```bash
# Show all pulumi outputs
pulumi stack output

# Get specific output
pulumi stack output lagoon_api_url

# Destroy and recreate (if needed)
pulumi destroy -y
pulumi up -y
```
