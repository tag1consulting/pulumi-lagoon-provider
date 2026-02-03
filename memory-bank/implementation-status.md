# Pulumi Lagoon Provider - Implementation Status

**Last Updated**: 2026-02-03
**Status**: v0.1.1 Released on PyPI - Phase 3 In Progress

---

## Current Work (2026-01-21)

### LagoonDeployTarget Resource
**Status**: ✅ COMPLETE (code) | ⏳ TESTING

New resource for managing Kubernetes deploy targets:
- `pulumi_lagoon/deploytarget.py` - Full CRUD implementation
- `pulumi_lagoon/validators.py` - Deploy target validators added
- `pulumi_lagoon/client.py` - Kubernetes GraphQL operations added
- `tests/unit/test_deploytarget.py` - Unit tests
- `tests/unit/test_validators.py` - Validator tests

GraphQL Operations:
- `add_kubernetes()` - Create deploy target
- `get_all_kubernetes()` - List all deploy targets
- `get_kubernetes_by_id()` - Query by ID
- `get_kubernetes_by_name()` - Query by name
- `update_kubernetes()` - Update deploy target
- `delete_kubernetes()` - Delete deploy target

### Multi-Cluster Example
**Status**: ✅ COMPLETE (2026-01-28)

Location: `examples/multi-cluster/`

Architecture:
- Production cluster (`lagoon-prod`): Lagoon core, Harbor, prod remote controller
- Non-production cluster (`lagoon-nonprod`): Nonprod remote controller

Components:
- `clusters/` - Kind cluster creation
- `infrastructure/` - ingress-nginx, cert-manager, CoreDNS
- `registry/` - Harbor container registry
- `lagoon/` - Lagoon core and remote (build-deploy)

Issues Fixed (2026-01-28):
- Keycloak config job secret name: Changed `prod-core-keycloak` to `prod-core-lagoon-core-keycloak`

Issues Fixed (2026-01-20):
- RabbitMQ CrashLoopBackOff: Cleared corrupted Mnesia data by deleting PVCs
- Service selector bug in `lagoon/core.py`: Changed selector to match actual pod labels
- Cross-cluster RabbitMQ IP: Added dynamic IP refresh using container ID triggers
- Keycloak Direct Access Grants: Added `lagoon/keycloak.py` for automatic configuration

New Features:
- `lagoon/keycloak.py`: Kubernetes Job that configures Keycloak for CLI auth
  - Enables Direct Access Grants for lagoon-ui client (OAuth password grant)
  - Creates lagoonadmin user with platform-owner role
- Dynamic IP refresh: Cluster IPs now automatically refresh when Kind clusters change
- Port-forwarding make targets: `make port-forwards-all`, `make test-ui`

Completed:
- All code fixes applied via `pulumi up`
- Port-forwarding access to Lagoon UI tested and working
- CLI authentication via OAuth password grant tested and working
- Browser authentication setup documented (requires hosts file entry)

**Branch**: `deploytarget-multi-cluster`
**PR**: https://github.com/tag1consulting/pulumi-lagoon-provider/pull/10 (Draft)

---

## Completed Work

### 1. GraphQL Client (pulumi_lagoon/client.py)
**Status**: ✅ COMPLETE

All core operations implemented:
- **Project Operations**:
  - `create_project()` - Create new project
  - `get_project_by_name()` - Query by name
  - `get_project_by_id()` - Query by ID
  - `update_project()` - Update existing project
  - `delete_project()` - Delete project

- **Environment Operations**:
  - `add_or_update_environment()` - Create/update environment
  - `get_environment_by_name()` - Query environment
  - `delete_environment()` - Delete environment

- **Variable Operations**:
  - `add_env_variable()` - Create variable
  - `get_env_variable_by_name()` - Query variable
  - `delete_env_variable()` - Delete variable

- **Error Handling**:
  - `LagoonAPIError` - API/GraphQL errors
  - `LagoonConnectionError` - Network errors
  - Proper timeout and retry logic

### 2. Configuration (pulumi_lagoon/config.py)
**Status**: ✅ COMPLETE

Features:
- Pulumi config integration (`lagoon:apiUrl`, `lagoon:token`)
- Environment variable fallback (`LAGOON_API_URL`, `LAGOON_TOKEN`)
- Secret management for tokens
- Client factory method

### 3. Resource Providers
**Status**: ✅ ALL COMPLETE

#### LagoonProject (pulumi_lagoon/project.py)
- Full CRUD implementation
- Dynamic provider with proper state management
- Drift detection via `read()` method
- Properties: name, git_url, deploytarget_id, production_environment, branches, pullrequests, etc.
- 268 lines, fully documented

#### LagoonEnvironment (pulumi_lagoon/environment.py)
- Full CRUD implementation
- Depends on project_id
- Properties: name, deploy_type, environment_type, route, routes, etc.
- 294 lines, fully documented

#### LagoonVariable (pulumi_lagoon/variable.py)
- Full CRUD implementation
- Supports project-level and environment-level variables
- Properties: name, value, scope, project_id, environment_id
- 277 lines, fully documented

### 4. Package Structure (pulumi_lagoon/__init__.py)
**Status**: ✅ COMPLETE

Clean exports:
- All three resource classes
- All argument dataclasses
- Client classes for advanced usage
- Configuration class

### 5. Example Code (examples/simple-project/)
**Status**: ✅ COMPLETE

Features:
- Complete working example using all three resources
- Demonstrates resource dependencies
- Shows project → environments → variables workflow
- Comprehensive README with setup instructions
- Includes troubleshooting guide

### 6. Automation Scripts (examples/simple-project/scripts/)
**Status**: ✅ COMPLETE

Scripts:
- `run-pulumi.sh` - Wrapper that handles token refresh before each command
- `quickstart.sh` - Full automated first-time setup
- `create-lagoon-admin.sh` - Creates lagoonadmin user in Keycloak
- `ensure-deploy-target.sh` - Creates deploy target if none exist
- `get-lagoon-token.sh` - Manual token acquisition
- `enable-direct-access-grants.sh` - Enable password auth in Keycloak
- `fix-rabbitmq-password.sh` - Fix RabbitMQ auth issues
- Various utility scripts for debugging

### 7. Makefile Automation
**Status**: ✅ COMPLETE

Targets:
- `setup-all` - Complete setup from scratch (~5 minutes)
- `ensure-lagoon-admin` - Create Keycloak user with fresh token
- `ensure-deploy-target` - Create deploy target with fresh token
- `example-up/preview/down` - Example project management
- `clean-all` - Full teardown (cluster, venvs, temp files)

Token Handling:
- Each target starts temporary port-forwards
- Each script fetches fresh tokens (handles 5-min expiration)
- Port-forwards cleaned up after each operation

## Remaining Work

### Code Cleanup
**Status**: ⏸️ LOW PRIORITY

1. **Clarify lagoon imports in single-cluster example** (2026-01-26)
   - `examples/single-cluster/__main__.py` has confusing imports:
     - `from lagoon import ...` - local infrastructure deployment functions
     - `import pulumi_lagoon as lagoon` - Lagoon API provider
   - Both end up using `lagoon.` prefix which is confusing
   - Suggested fix: Rename `import pulumi_lagoon as lagoon` to just `import pulumi_lagoon`
   - Then use `pulumi_lagoon.LagoonDeployTarget(...)` instead of `lagoon.LagoonDeployTarget(...)`
   - **Note**: This is a minor code style issue that doesn't affect functionality. Deferred to future cleanup.

### Provider Features
**Status**: ✅ PARTIAL

1. **Import functionality** (Phase 2) - ✅ COMPLETE
   - All resources support `pulumi import` with composite ID formats
   - LagoonProject: `{numeric_id}`
   - LagoonDeployTarget: `{numeric_id}`
   - LagoonEnvironment: `{project_id}:{env_name}`
   - LagoonVariable: `{project_id}:{env_id}:{var_name}` or `{project_id}::{var_name}`
   - LagoonDeployTargetConfig: `{project_id}:{config_id}`
   - Import utilities in `pulumi_lagoon/import_utils.py`
   - Full unit test coverage in `tests/unit/test_import_utils.py`

2. **Add SSH key authentication** (Phase 2+)
   - Currently only JWT token authentication is supported
   - SSH key auth would enable service account workflows
   - Would require changes to `pulumi_lagoon/config.py` and `client.py`

### Testing
**Status**: ✅ COMPLETE (2026-01-26)

All unit tests implemented and passing (240 tests):
- `tests/unit/test_client.py` - GraphQL client tests (386 lines)
- `tests/unit/test_config.py` - Configuration tests (146 lines)
- `tests/unit/test_project.py` - Project resource tests (366 lines)
- `tests/unit/test_environment.py` - Environment resource tests (389 lines)
- `tests/unit/test_variable.py` - Variable resource tests (521 lines)
- `tests/unit/test_deploytarget.py` - Deploy target tests (411 lines)
- `tests/unit/test_validators.py` - Validator tests (788 lines)

Run tests with: `pytest tests/unit/ -v`

### Documentation Updates
**Status**: ✅ COMPLETE (2026-01-26)

README.md includes:
- Complete usage examples (LagoonProject, LagoonEnvironment, LagoonVariable)
- Testing instructions (`pytest tests/`)
- Make targets documentation
- Project structure overview
- Examples directory descriptions

No significant API quirks discovered during testing.

## How to Use Right Now

### Installation
```bash
cd /home/gchaix/repos/pulumi-lagoon-provider
pip install -e .
```

### Basic Usage
```python
import pulumi
import pulumi_lagoon as lagoon

# Create a project
project = lagoon.LagoonProject("my-site",
    lagoon.LagoonProjectArgs(
        name="my-drupal-site",
        git_url="git@github.com:org/repo.git",
        deploytarget_id=1,
        production_environment="main",
    )
)

# Create an environment
env = lagoon.LagoonEnvironment("prod",
    lagoon.LagoonEnvironmentArgs(
        name="main",
        project_id=project.id,
        deploy_type="branch",
        environment_type="production",
    )
)

# Create a variable
var = lagoon.LagoonVariable("db-host",
    lagoon.LagoonVariableArgs(
        name="DATABASE_HOST",
        value="mysql.prod.example.com",
        project_id=project.id,
        environment_id=env.id,
        scope="runtime",
    )
)
```

### Configuration
```bash
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql
pulumi config set lagoon:token YOUR_TOKEN --secret
```

## Next Steps

### Immediate (Phase 1 Completion) - ✅ DONE
1. ~~**Test against real Lagoon instance**~~ ✅ Tested and working
2. ~~**Implement automation scripts**~~ ✅ Complete
3. ~~**Update README**~~ ✅ Updated with Makefile targets

### Remaining (Optional)
1. **Implement unit tests**
   - Mock GraphQL responses
   - Test error handling
   - Test resource lifecycle

2. **Additional documentation**
   - API quirks discovered during testing
   - Advanced usage examples

### Phase 2 (Complete)
- ✅ LagoonDeployTarget resource implemented
- ✅ LagoonDeployTargetConfig resource implemented
- ✅ Multi-cluster example complete and working
- ✅ CI/CD pipeline (GitHub Actions)
- ✅ PyPI publication (v0.1.0 released 2026-01-30)
- ⏳ Additional resources (Group, Notification) - deferred to Phase 3

### Phase 3 (Long-term)
- Native Go provider
- Multi-language SDK generation

## File Locations

### Core Implementation
- `pulumi_lagoon/__init__.py` - Package exports
- `pulumi_lagoon/config.py` - Configuration
- `pulumi_lagoon/client.py` - GraphQL client
- `pulumi_lagoon/project.py` - Project resource
- `pulumi_lagoon/environment.py` - Environment resource
- `pulumi_lagoon/variable.py` - Variable resource

### Examples
- `examples/simple-project/__main__.py` - Working example
- `examples/simple-project/README.md` - Example documentation

### Tests (Templates Only)
- `tests/test_client.py`
- `tests/test_config.py`

### Documentation
- `README.md` - Main project documentation
- `CLAUDE.md` - Project-specific Claude instructions
- `memory-bank/planning.md` - Original planning document
- `memory-bank/architecture.md` - Architecture documentation
- `memory-bank/implementation-status.md` - This file

## Git Status

**Current Branch**: main

**Changes Made** (not yet committed):
- Updated `pulumi_lagoon/client.py` - Added environment and variable operations
- Implemented `pulumi_lagoon/project.py` - Complete resource
- Implemented `pulumi_lagoon/environment.py` - Complete resource
- Implemented `pulumi_lagoon/variable.py` - Complete resource
- Updated `pulumi_lagoon/__init__.py` - Added exports
- Updated `examples/simple-project/__main__.py` - Working example
- Updated `examples/simple-project/README.md` - Comprehensive guide

**Suggested Commit Message**:
```
Implement Phase 1: Core resource providers

- Add complete GraphQL client with project, environment, and variable operations
- Implement LagoonProject dynamic resource provider with full CRUD support
- Implement LagoonEnvironment dynamic resource provider
- Implement LagoonVariable dynamic resource provider
- Update package exports to include all resources
- Add comprehensive working example with documentation
- All core Phase 1 functionality complete and ready for testing
```

## Success Criteria Met

✅ Can create/update/delete Lagoon projects via Pulumi
✅ Can manage environments and variables
✅ Working example demonstrating all resources
✅ Documentation sufficient for early adopters
✅ Proper dependency handling (project → environment → variable)
✅ Drift detection support via read() methods
✅ Clean API with dataclass arguments
✅ Automated setup via `make setup-all` (~5 minutes)
✅ Token expiration handled automatically
✅ Full end-to-end testing passed (2026-01-13)

## Recent Fixes

### Issue #6: Keycloak Migration Job Wait Fix (2026-01-20)
**Problem**: The `lagoon-core-api-migratedb` job pod would fail on initial deploy because the database wasn't ready yet. The `kubectl wait --for=condition=ready` commands would then block for the full timeout (300s/5 minutes) because Job pods don't become "ready" - they either complete or fail.

**Solution**: Updated the wait logic in both `Makefile` and `scripts/setup-complete.sh` to:
1. First wait for the migration job to complete by name (`lagoon-core-api-migratedb`), checking for Complete or Failed conditions
2. If the job fails, delete it to allow Helm to retry on next deployment
3. Then wait for deployment pods using `--field-selector=status.phase=Running` to exclude completed/failed job pods from the wait

**Files Modified**:
- `Makefile` - `wait-for-lagoon` target
- `scripts/setup-complete.sh` - `wait_for_lagoon()` function

## Known Limitations

1. ~~**No unit tests yet**~~ ✅ 240+ tests passing (2026-01-26)
2. ~~**Not tested against real Lagoon**~~ ✅ Tested and working
3. ~~**No import functionality**~~ ✅ Import support complete (2026-01-26)
4. ~~**Limited validation**~~ ✅ Comprehensive validation in validators.py (470 lines)
5. **Token-based auth only** - SSH key authentication not supported

## Questions Answered

1. ~~Does the GraphQL schema match real Lagoon API?~~ ✅ Yes, tested and working
2. ~~Are there rate limits that need handling?~~ Not observed in testing
3. ~~Should we implement retry logic?~~ Not needed for current use cases
4. ~~How should we handle concurrent modifications?~~ Pulumi state handles this
5. ~~Do we need import functionality for Phase 1?~~ No, deferred to Phase 2
