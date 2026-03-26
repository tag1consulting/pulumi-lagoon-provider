# Release v0.2.6 (2026-03-26)

Maintenance release switching npm publishing to OIDC trusted publishing, upgrading the Node.js toolchain, and fixing the PyPI badge in the Python SDK README.

## Improvements

- **npm OIDC publishing**: Replaced token-based npm publishing (`NPM_TOKEN` secret) with npm's OIDC trusted publishing (GA since July 2025), matching the existing PyPI pattern. No secrets required — GitHub Actions exchanges an OIDC token directly with npmjs.com. Adds `--provenance` for SLSA attestation and a verified publisher badge on npmjs.com.
- **Node.js toolchain upgrade**: Dropped Node.js 18 (EOL April 2025) and 20 (maintenance EOL); build and publish now use Node.js 24 (Active LTS); test matrix is `['22', '24']`.
- **`npm ci` in CI**: Switched from `npm install` to `npm ci` in all CI jobs for reproducible, lock-file-pinned installs.
- **PyPI badge**: Fixed stale `badge.fury.io` badge in `sdk/python/README.pypi.md`; now uses `shields.io` dynamic badge matching the root README.

## Installation

```bash
pip install pulumi-lagoon==0.2.6
npm install @tag1consulting/pulumi-lagoon@0.2.6
go get github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon@v0.2.6
```

---

# Release v0.2.5 (2026-03-26)

Patch release fixing the Node.js SDK so `@tag1consulting/pulumi-lagoon` can be installed and required correctly.

## Bug Fixes

- **Node.js SDK `require` path**: `bin/utilities.js` called `require('./package.json')` which resolves to `bin/package.json` — a file that doesn't exist after `npm install`. Fixed to `require('../package.json')` so it correctly finds the package root. This caused the `test-install-nodejs` smoke test to fail in v0.2.4, blocking npm publishing entirely.
- **Makefile `go-sdk-nodejs`**: Added `sed` fixup so future SDK regenerations automatically patch the generated `utilities.ts` path, preventing regression.

## Installation

```bash
pip install pulumi-lagoon==0.2.5
npm install @tag1consulting/pulumi-lagoon@0.2.5
go get github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon@v0.2.5
```

---

# Release v0.2.4 (2026-03-26)

Maintenance release fixing pkg.go.dev license display, PyPI badge accuracy, README version text, and adding automated npm publishing for the TypeScript SDK.

## Bug Fixes

- **Go SDK LICENSE**: Added `LICENSE` file to `sdk/go/lagoon/` directory so pkg.go.dev can detect the Apache 2.0 license and display documentation. The Go module proxy creates zip archives scoped to the module subdirectory; the repo root `LICENSE` was outside that boundary and invisible to the license checker.
- **PyPI badge**: Switched from `badge.fury.io` to `shields.io` for the PyPI version badge to avoid CloudFront CDN caching that showed stale version `0.1.2`
- **README version text**: Updated status line from `v0.2.2` to `v0.2.4`
- **Makefile go-sdk-go**: Added `cp LICENSE sdk/go/lagoon/LICENSE` to automate the LICENSE copy on every SDK regeneration; added `--exclude='LICENSE'` to rsync so the file is not deleted when the SDK is regenerated

## New Features

- **npm publishing**: Added `build-nodejs` and `publish-npm` jobs to `.github/workflows/publish.yml`. The TypeScript SDK (`@tag1consulting/pulumi-lagoon`) is now automatically published to npm on each GitHub Release. Requires a `NPM_TOKEN` secret in a GitHub `npm` environment.

## Installation

```bash
pip install pulumi-lagoon==0.2.4
npm install @tag1consulting/pulumi-lagoon@0.2.4
go get github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon@v0.2.4
```

---

# Release v0.2.3 (2026-03-26)

Maintenance release with CI improvements, tooling fixes, and developer experience improvements.

## Bug Fixes

- **Makefile SDK sync**: Added `--delete` to rsync commands for generated SDK subdirectories so stale files are removed when resources are removed from the schema
- **`release-prep` version propagation**: Restructured to bump version strings first, then build — provider binary and all SDKs now carry the new version from the start
- **`release-prep` portability**: Replaced GNU-specific `sed '0,/pattern/s//'` with `jq` for updating `sdk/nodejs/package.json` version; works on both Linux and macOS
- **`release-prep` self-modification bug**: Added `^` anchor to the `PROVIDER_VERSION` sed pattern so the command no longer matches and corrupts itself when the Makefile is updated in place
- **`release-prep` Node.js SDK version drift**: The `jq` version update now sets both `.version` and `.pulumi.version` in `sdk/nodejs/package.json` to keep both fields in sync

## Improvements

- **CodeRabbit configuration**: Added `.coderabbit.yaml` enabling sequence diagrams, high-level summaries, related issues/PRs, code review effort estimation, and architectural tone in reviews
- **Ruff configuration**: Added `ruff.toml` with per-file-ignores for generated SDK Python files to eliminate false-positive lint noise on regeneration
- **CodeQL configuration**: Added `.github/codeql/codeql-config.yml` to exclude generated SDK directories from CodeQL analysis
- **CODEOWNERS**: Simplified to `@gchaix` only; CodeRabbit reviews via its own configuration rather than as a required code owner

## Installation

```bash
pip install pulumi-lagoon==0.2.3
```

---

# Release v0.2.2 (2026-03-25)

Feature release adding the Group resource, SDK path fixes, and significantly expanded test coverage.

## New Features

### Group Resource

New `Group` resource for managing Lagoon groups as infrastructure-as-code:

```python
from pulumi_lagoon.lagoon import Group

group = Group("my-team",
    name="my-team",
)
```

- `pulumi import lagoon:lagoon:Group my-group my-group-name`

## Bug Fixes

- **Group ID type**: Lagoon returns UUIDs (strings) for group IDs, not integers. Fixed `Group.ID` and `GroupState.LagoonID` types and updated the schema accordingly.
- **Group delete idempotency**: `DeleteGroup` now correctly handles "Group not found" errors so that deleting an already-removed group succeeds instead of erroring.
- **Group parentGroup queries reverted**: The `parentGroup` field is not exposed on Lagoon's `GroupInterface` type. Reverted query additions that caused `Cannot query field "parentGroup"` errors. The `Read` handler preserves `parentGroupName` from Pulumi state instead.
- **Task fallback predicate tightened**: `isFieldNotFoundOrLegacyError` now only triggers legacy API fallback for `advancedTasksForEnvironment` field errors, preventing false-positive fallbacks on unrelated API errors.
- **SDK path fix**: Corrected doubled SDK directory paths (`sdk/python/python/` → `sdk/python/`, `sdk/nodejs/nodejs/` → `sdk/nodejs/`) caused by `pulumi package gen-sdk`
- **PyPI README**: Fixed missing project description on PyPI by copying the root README into the Python SDK before building
- **npm badge**: Fixed broken npm version badge that used the non-existent `badge.npmjs.com` service

## Improvements

- **Test coverage**: Expanded from ~40% to 87.4% (512 tests, up from ~200)
- **CI**: Added `test-build.yml` workflow to validate the Python SDK builds correctly on every PR

## Installation

```bash
pip install pulumi-lagoon==0.2.2
```

---

# Release v0.2.1 (2026-03-25)

Maintenance release with security updates, CI improvements, and Lagoon CLI setup tooling.

## Security

- **Dependency updates**: Bumped `github.com/go-git/go-git/v5` from 5.13.1 to 5.16.5 and `github.com/cloudflare/circl` from 1.6.1 to 1.6.3 across all Go modules (provider, SDK, test)

## New Features

### Lagoon CLI Setup Scripts

New automated setup scripts and documentation for configuring the Lagoon CLI against local and multi-cluster deployments:

- `scripts/setup-lagoon-cli.sh` — Configures the Lagoon CLI with SSH and API endpoints, creates an admin user, and verifies connectivity
- `examples/multi-cluster/scripts/setup-lagoon-cli.sh` — Multi-cluster variant with prod/nonprod support
- `examples/multi-cluster/scripts/get-lagoon-token.sh` — Retrieves a Lagoon API token via SSH key authentication
- `docs/lagoon-cli-setup.md` — Full Lagoon CLI setup guide
- `examples/multi-cluster/LAGOON_CLI_SETUP.md` — Multi-cluster-specific CLI setup guide

## Improvements

- **Notification helpers**: Refactored notification CRUD operations to use shared helper functions, reducing duplication across all notification resource types
- **CI**: Updated `upload-artifact` and `download-artifact` GitHub Actions to Node.js 24 compatible versions

## Cleanup

- Removed legacy Python v0.1.x build artifacts: `setup.py`, `requirements.txt`, and the old `test.yml` workflow
- Updated README to reflect v0.2.x native Go provider structure

## Installation

```bash
pip install pulumi-lagoon==0.2.1
```

---

# Release v0.2.0 (2026-03-24)

This is a major release that replaces the Python dynamic provider with a native Go provider and generated multi-language SDKs.

## Breaking Changes

### Python SDK: Class names and import paths changed

The `pulumi_lagoon` package on PyPI now ships the **native Go provider SDK** instead of the Python dynamic provider. All class names, module paths, and resource behaviors have changed.

**Old (v0.1.x dynamic provider):**
```python
from pulumi_lagoon import LagoonProject, LagoonEnvironment, LagoonVariable
from pulumi_lagoon import LagoonNotificationSlack, LagoonProjectNotification
```

**New (v0.2.x native SDK):**
```python
from pulumi_lagoon.lagoon import Project, Environment, Variable
from pulumi_lagoon.lagoon import NotificationSlack, ProjectNotification
```

### Python 3.8 no longer supported

The native SDK requires Python 3.9 or later.

### Resource ID semantics changed

Notification resources now use numeric Lagoon IDs as the Pulumi resource ID (instead of the resource name). This may require `pulumi import` to re-adopt existing resources.

## Migration Guide

1. Upgrade: `pip install --upgrade pulumi-lagoon`
2. Update your Pulumi programs to use the new class names and import paths (see above)
3. If you have existing state with v0.1.x resources, you may need to run `pulumi refresh` or re-import resources using `lagoon:lagoon:*` type tokens (e.g., `pulumi import lagoon:lagoon:NotificationSlack my-slack my-notification-name`)

## Highlights

- **Native Go provider**: Built with `pulumi-go-provider` for full type safety and performance
- **Multi-language SDKs**: Generated Python, TypeScript, and Go SDKs from a single schema
- **Improved correctness**: Proper Read/Diff/Update lifecycle for all resources; idempotent creates; graceful not-found handling
- **Comprehensive tests**: 198 unit tests covering all resource types
- **TypeScript SDK**: `@tag1consulting/pulumi-lagoon` on npm
- **Go SDK**: `github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon`

## New Features

- All resources support create-or-update semantics (idempotent)
- `DeployTargetConfig` resource for managing deploy target configurations
- Improved Diff logic prevents spurious updates for optional fields with API defaults
- `ErrNotFound` handling in Read methods marks deleted resources for re-creation

## Resources

All resources from v0.1.x are available under the new `pulumi_lagoon.lagoon` module:
- `Project` (was `LagoonProject`)
- `Environment` (was `LagoonEnvironment`)
- `Variable` (was `LagoonVariable`)
- `DeployTarget` (was `LagoonDeployTarget`)
- `DeployTargetConfig` (was `LagoonDeployTargetConfig`)
- `NotificationSlack` (was `LagoonNotificationSlack`)
- `NotificationRocketChat` (was `LagoonNotificationRocketChat`)
- `NotificationEmail` (was `LagoonNotificationEmail`)
- `NotificationMicrosoftTeams` (was `LagoonNotificationMicrosoftTeams`)
- `ProjectNotification` (was `LagoonProjectNotification`)
- `Task` (was `LagoonTask`)

---

# Release v0.1.2 (2026-02-06)

This release adds multi-version Lagoon API compatibility, supporting Lagoon versions v2.24.1 through v2.30.0.

## Highlights

- **Multi-Version API Compatibility**: Automatic fallback to older API queries/mutations for Lagoon versions prior to v2.30.0
- **CRD Version Detection**: Automatic selection of correct CRD storage version based on Lagoon chart version
- **SSL Verification Option**: Support for self-signed certificates in development environments
- **Expanded Test Coverage**: 507 unit tests with 92% code coverage

## New Features

### Multi-Version Lagoon API Support

The provider now automatically detects and adapts to different Lagoon API versions:

| Lagoon Version | Chart Version | Status |
|----------------|---------------|--------|
| v2.30.0 | 1.59.0 | Primary API |
| v2.28.0 | 1.56.0 | Fallback support |
| v2.24.1 | 1.52.0 | Fallback support |

**API Changes Handled:**
- `get_project_by_id`: Uses `allProjects` + filter (projectById doesn't exist in older versions)
- `get_kubernetes_by_id`: Uses `allKubernetes` + filter
- `add_env_variable`: Falls back to `addEnvVariable` mutation
- `delete_env_variable`: Falls back to `deleteEnvVariable` mutation
- `get_env_variable_by_name`: Falls back to `envVariablesByProjectEnvironment` query
- `get_advanced_tasks_by_environment`: Falls back to `advancedTasksByEnvironment` query

### CRD Version Detection

For Kubernetes deployments, the provider automatically selects the correct CRD storage version:
- **Chart < 1.58.0** (Lagoon ≤ v2.28.0): v1beta1 as storage version
- **Chart ≥ 1.58.0** (Lagoon ≥ v2.29.0): v1beta2 as storage version
- Both v1beta1 and v1beta2 are always served (required by lagoon-remote controller)

### SSL Verification Option

`LagoonDeployTarget` now supports a `verify_ssl` parameter for environments with self-signed certificates:

```python
deploy_target = lagoon.LagoonDeployTarget("local-cluster",
    name="local-kind",
    console_url="https://kubernetes.default.svc",
    api_url="https://api.lagoon.local/graphql",
    jwt_secret=jwt_secret,
    verify_ssl=False,  # For self-signed certs
)
```

## Bug Fixes

- Fixed `get_project_by_id` to work with Lagoon versions that don't support `projectById` query
- Fixed environment variable operations for older Lagoon API versions
- Fixed advanced task queries to handle different response formats across versions

## Documentation

- Updated README with multi-version compatibility information
- Added version compatibility table to documentation

## Requirements

- Python 3.8+
- Pulumi CLI 3.x
- Lagoon v2.24.1 or later

## Installation

```bash
pip install pulumi-lagoon==0.1.2
```

## Full Changelog

See the [commit history](https://github.com/tag1consulting/pulumi-lagoon-provider/compare/v0.1.1...v0.1.2) for all changes.

---

# Release v0.1.1 (2026-02-02)

This release adds notification and task management resources to the Pulumi Lagoon Provider.

## Highlights

- **Notification Resources**: Full CRUD support for all Lagoon notification types (Slack, RocketChat, Email, Microsoft Teams) plus project notification linking
- **Task Resources**: Manage advanced task definitions for on-demand commands and container-based tasks
- **Expanded Test Coverage**: 467+ unit tests

## New Features

### Notification Resources
- **LagoonNotificationSlack** - Create and manage Slack webhook notifications
- **LagoonNotificationRocketChat** - Create and manage RocketChat webhook notifications
- **LagoonNotificationEmail** - Create and manage email notifications
- **LagoonNotificationMicrosoftTeams** - Create and manage Microsoft Teams webhook notifications
- **LagoonProjectNotification** - Link notifications to projects for deployment alerts

### Task Resources
- **LagoonTask** - Create and manage advanced task definitions with support for:
  - Command-type tasks (execute commands in existing service containers)
  - Image-type tasks (run custom container images)
  - Multiple scope options: project, environment, group, or system-wide
  - Permission levels: guest, developer, maintainer
  - Task arguments with configurable types
  - Confirmation text before execution

### Import Support
All new resources support `pulumi import`:
- `pulumi import lagoon:index:NotificationSlack my-slack deploy-alerts`
- `pulumi import lagoon:index:NotificationRocketChat my-rc team-chat`
- `pulumi import lagoon:index:NotificationEmail my-email ops-team`
- `pulumi import lagoon:index:NotificationMicrosoftTeams my-teams teams-alerts`
- `pulumi import lagoon:index:ProjectNotification my-assoc my-project:slack:deploy-alerts`
- `pulumi import lagoon:index:Task my-task 123`

## Example Usage

### Notifications
```python
import pulumi_lagoon as lagoon

# Create a Slack notification
slack_notify = lagoon.LagoonNotificationSlack("deploy-alerts",
    name="deploy-alerts",
    webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
    channel="#deployments",
)

# Link notification to a project
project_notify = lagoon.LagoonProjectNotification("project-slack",
    project_name="my-project",
    notification_type="slack",
    notification_name="deploy-alerts",
)
```

### Tasks
```python
import pulumi_lagoon as lagoon

# Create a command-type task
yarn_audit = lagoon.LagoonTask("yarn-audit",
    name="run-yarn-audit",
    type="command",
    service="node",
    command="yarn audit",
    project_id=project.id,
    permission="developer",
    description="Run yarn security audit",
)

# Create an image-type task
backup_task = lagoon.LagoonTask("db-backup",
    name="database-backup",
    type="image",
    service="cli",
    image="amazeeio/database-tools:latest",
    project_id=project.id,
    permission="maintainer",
    confirmation_text="This will create a full database backup. Continue?",
)
```

## Documentation

- See [docs/notifications.md](docs/notifications.md) for detailed notification resource documentation
- Task resource documentation and examples are included in the main [README.md](README.md#supported-resources)

## Requirements

- Python 3.8+
- Pulumi CLI 3.x
- Access to a Lagoon instance with API credentials

## Installation

```bash
pip install pulumi-lagoon
```

## Full Changelog

See the [commit history](https://github.com/tag1consulting/pulumi-lagoon-provider/compare/v0.1.0...v0.1.1) for all changes.

---

# Release v0.1.0 (2026-01-30)

The initial release of the Pulumi Lagoon Provider, providing a Python dynamic provider for managing [Lagoon](https://www.lagoon.sh/) resources as infrastructure-as-code.

## Highlights

This release provides a complete, working dynamic provider that enables declarative management of Lagoon hosting platform resources using Pulumi.

## Features

### Core Resources
- **LagoonProject** - Create and manage Lagoon projects with full CRUD support
- **LagoonEnvironment** - Manage environments (branch/PR deployments)
- **LagoonVariable** - Manage project and environment-level variables with build/runtime/global scopes

### Deploy Target Resources
- **LagoonDeployTarget** - Manage Kubernetes cluster deploy targets
- **LagoonDeployTargetConfig** - Configure deployment routing to specific clusters based on branch patterns

### Infrastructure
- GraphQL client with comprehensive error handling
- CORS support and TLS bypass for local development
- Token refresh handling for Lagoon's 5-minute token expiration
- Resource import support for adopting existing Lagoon infrastructure

### Examples & Automation
- **simple-project/** - Basic provider usage example
- **single-cluster/** - Complete Lagoon stack on a single Kind cluster
- **multi-cluster/** - Production-like deployment with separate prod/nonprod clusters
- Makefile automation for common operations
- Port-forward management and health checks

### Testing
- 300+ unit tests with 95% code coverage
- Integration test framework
- CI/CD pipeline via GitHub Actions

## Requirements

- Python 3.8+
- Pulumi CLI 3.x
- Access to a Lagoon instance with API credentials

## Installation

```bash
pip install pulumi-lagoon
```

## Quick Start

```python
import pulumi_lagoon as lagoon

# Create a Lagoon project
project = lagoon.LagoonProject("my-site",
    name="my-site",
    git_url="git@github.com:org/repo.git",
    deploytarget_id=1,
    production_environment="main",
    branches="^(main|develop)$",
)

# Create an environment
env = lagoon.LagoonEnvironment("production",
    project_id=project.id,
    name="main",
    environment_type="production",
    deploy_type="branch",
)

# Add a variable
var = lagoon.LagoonVariable("api-key",
    project_id=project.id,
    name="API_KEY",
    value="secret-value",
    scope="runtime",
)
```

## Known Limitations

- This is a **dynamic provider** (Python-based), not a native provider
- Dynamic providers run in a subprocess and cannot access Pulumi config secrets directly - use environment variables for `LAGOON_TOKEN`

## License

Apache License 2.0

## Acknowledgments

Built for the Lagoon community by Tag1 Consulting.
