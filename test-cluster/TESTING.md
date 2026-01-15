# Testing the Pulumi Lagoon Provider

This guide explains how to test the `pulumi-lagoon-provider` using the local test cluster.

## Quick Test Workflow

### 1. Deploy Test Cluster

```bash
cd test-cluster
./scripts/setup.sh
source venv/bin/activate
pulumi up
```

Wait for deployment to complete (~10-20 minutes).

### 2. Configure /etc/hosts

```bash
echo "127.0.0.1 api.lagoon.test ui.lagoon.test harbor.lagoon.test" | sudo tee -a /etc/hosts
```

### 3. Get Lagoon Credentials

```bash
./scripts/get-credentials.sh
```

Save the token displayed - you'll need it for testing.

### 4. Install Provider

```bash
cd ..
pip install -e .
```

### 5. Test Provider with Example

```bash
cd test-cluster/examples

# Set credentials
export LAGOON_API_URL='http://api.lagoon.test/graphql'
export LAGOON_TOKEN='<token-from-step-3>'

# Run test
python3 test-provider.py
```

### 6. Verify in Lagoon UI

Open http://ui.lagoon.test in your browser to see the created resources.

### 7. Cleanup

```bash
cd ../test-cluster
pulumi destroy
```

## Detailed Testing Scenarios

### Scenario 1: Basic CRUD Operations

Test creating, reading, updating, and deleting resources:

```python
import pulumi_lagoon as lagoon

# Create
project = lagoon.LagoonProject("test", lagoon.LagoonProjectArgs(
    name="test-project",
    git_url="https://github.com/example/repo.git",
    deploytarget_id=1,
))

# Read (automatic via Pulumi)
pulumi.export("project_id", project.id)

# Update (change git_url and run `pulumi up` again)
# Delete (run `pulumi destroy`)
```

### Scenario 2: Resource Dependencies

Test that dependencies work correctly:

```python
# Project must exist before environment
env = lagoon.LagoonEnvironment("prod", lagoon.LagoonEnvironmentArgs(
    project_id=project.id,  # Dependency
    name="main",
))

# Environment must exist before environment variable
var = lagoon.LagoonVariable("db-host", lagoon.LagoonVariableArgs(
    project_id=project.id,
    environment_id=env.id,  # Dependency
    name="DB_HOST",
    value="localhost",
))
```

### Scenario 3: Drift Detection

Test that the provider detects changes made outside Pulumi:

```bash
# 1. Create resources via Pulumi
pulumi up

# 2. Manually modify in Lagoon (via GraphQL or UI)
# For example, change a variable value

# 3. Run refresh to detect drift
pulumi refresh

# Should show differences and offer to sync
```

### Scenario 4: Error Handling

Test error scenarios:

```python
# Test 1: Invalid credentials
# Set wrong token and verify error message

# Test 2: Non-existent resource
# Try to create environment with invalid project_id

# Test 3: Duplicate names
# Try to create two projects with same name

# Test 4: Invalid parameters
# Try to create project with invalid git_url format
```

### Scenario 5: Concurrent Operations

Test multiple resources at once:

```python
# Create multiple environments in parallel
envs = [
    lagoon.LagoonEnvironment(f"env-{i}", lagoon.LagoonEnvironmentArgs(
        name=f"branch-{i}",
        project_id=project.id,
        deploy_type="branch",
    ))
    for i in range(5)
]
```

## Manual GraphQL Testing

Test the Lagoon API directly to verify provider behavior:

```bash
# Set variables
export API_URL="http://api.lagoon.test/graphql"
export TOKEN="<your-token>"

# List all projects
curl -X POST "${API_URL}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"query":"{ allProjects { id name gitUrl } }"}'

# Create a project
curl -X POST "${API_URL}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "query": "mutation($input: AddProjectInput!) { addProject(input: $input) { id name } }",
    "variables": {
      "input": {
        "name": "manual-test",
        "gitUrl": "https://github.com/example/repo.git",
        "deploytarget": 1
      }
    }
  }'

# Delete a project
curl -X POST "${API_URL}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "query": "mutation { deleteProject(input: { project: \"manual-test\" }) }"
  }'
```

## Debugging Tips

### Enable Verbose Logging

```bash
# Pulumi verbose mode
pulumi up -v=9

# Python logging
export PULUMI_DEBUG_COMMANDS=true
export PULUMI_DEBUG_PROMISE_LEAKS=true
```

### Inspect Provider Requests

Add logging to the provider code:

```python
# In pulumi_lagoon/client.py
import logging
logging.basicConfig(level=logging.DEBUG)

def execute_graphql(self, query, variables=None):
    logging.debug(f"GraphQL Query: {query}")
    logging.debug(f"Variables: {variables}")
    # ... rest of method
```

### Check Lagoon API Logs

```bash
kubectl --context kind-lagoon-test -n lagoon logs -l app.kubernetes.io/name=api -f
```

### Use Pulumi Console

Export state for inspection:

```bash
pulumi stack export > stack-state.json
cat stack-state.json | jq '.deployment.resources'
```

## Common Issues and Solutions

### Issue: "Connection refused"

**Cause**: Lagoon API not accessible
**Solution**:
1. Check /etc/hosts entry
2. Verify pods are running: `kubectl -n lagoon get pods`
3. Try port-forward: `kubectl -n lagoon port-forward svc/lagoon-core-api 3000:3000`

### Issue: "Unauthorized" or "Invalid token"

**Cause**: Token expired or incorrect
**Solution**:
1. Re-run `./scripts/get-credentials.sh`
2. Check token format (should be JWT)
3. Verify API pod is healthy

### Issue: "Resource already exists"

**Cause**: Pulumi state out of sync
**Solution**:
1. Run `pulumi refresh`
2. Or delete manually and import: `pulumi import`
3. Or start fresh: `pulumi stack rm dev && pulumi stack init dev`

### Issue: Slow performance

**Cause**: Docker resource constraints
**Solution**:
1. Increase Docker resources (8 CPU, 12GB RAM minimum)
2. Close other applications
3. Use `--target` for specific resources

## Automated Testing

### Unit Tests (Coming Soon)

```bash
cd ..
pytest tests/ -v
```

### Integration Tests

Create `test-cluster/examples/test_integration.py`:

```python
import unittest
import pulumi_lagoon as lagoon

class TestLagoonProvider(unittest.TestCase):
    def test_create_project(self):
        # Test project creation
        pass

    def test_create_environment(self):
        # Test environment creation
        pass

    # ... more tests

if __name__ == "__main__":
    unittest.main()
```

## Performance Testing

Test provider performance with many resources:

```python
# Create 10 projects
projects = [
    lagoon.LagoonProject(f"project-{i}", lagoon.LagoonProjectArgs(
        name=f"project-{i}",
        git_url=f"https://github.com/example/repo-{i}.git",
        deploytarget_id=1,
    ))
    for i in range(10)
]

# Create 5 environments per project
envs = [
    lagoon.LagoonEnvironment(f"env-{i}-{j}", lagoon.LagoonEnvironmentArgs(
        name=f"env-{j}",
        project_id=projects[i].id,
        deploy_type="branch",
    ))
    for i in range(10)
    for j in range(5)
]

# Measure time
# pulumi up --watch
```

## CI/CD Testing

Example GitHub Actions workflow:

```yaml
name: Test Provider

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup kind
        uses: helm/kind-action@v1

      - name: Setup Pulumi
        uses: pulumi/setup-pulumi@v2

      - name: Deploy test cluster
        run: |
          cd test-cluster
          ./scripts/setup.sh
          pulumi up --yes

      - name: Test provider
        run: |
          pip install -e .
          cd test-cluster/examples
          export LAGOON_API_URL='http://api.lagoon.test/graphql'
          export LAGOON_TOKEN=$(./scripts/get-credentials.sh | grep Token | cut -d' ' -f2)
          python3 test-provider.py

      - name: Cleanup
        if: always()
        run: |
          cd test-cluster
          pulumi destroy --yes
```

## Reporting Bugs

When reporting provider bugs, include:

1. **Environment**:
   - OS and version
   - Docker version
   - Pulumi version
   - Python version

2. **Reproduction**:
   - Minimal Pulumi program
   - Steps to reproduce
   - Expected vs actual behavior

3. **Logs**:
   - Pulumi output: `pulumi up -v=9`
   - Lagoon API logs
   - Provider error messages

4. **State**:
   - `pulumi stack export` (sanitize secrets!)
   - Resource status

## Next Steps

After successful testing:

1. Document any issues found
2. Update provider code as needed
3. Add unit tests for bug fixes
4. Update documentation
5. Prepare for Phase 2 features
