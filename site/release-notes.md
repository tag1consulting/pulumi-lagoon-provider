---
title: Release Notes
nav_order: 7
---

# Release Notes

---

## v0.5.3 (2026-07-21)

Bug fix release closing a nil-pointer panic that could crash the provider under concurrent resource operations, plus a local e2e release-gate harness repair. No provider API or schema changes — existing programs require no updates.

### Bug Fixes

- **Nil-pointer panic under concurrent resource creation** ([#265](https://github.com/tag1consulting/pulumi-lagoon-provider/issues/265)): `LagoonConfig` cached its Lagoon API client via a `*sync.Once` field plus a separate `*client.Client` field. `infer.GetConfig` returns independent value copies of the provider config per resource operation; the `*sync.Once` pointer was correctly shared across those copies, but the `*client.Client` field was not, so only the copy whose goroutine won the `sync.Once` race ever had a populated client — every other concurrent copy's `NewClient()` returned nil, and the next call into that nil client panicked. This reliably crashed the provider when creating two or more `DeployTarget`s concurrently, as in the `examples/multi-cluster` example. Fixed by routing the cached client through a single shared holder struct instead of two independently-copied fields.

### Build and CI

- Fixed five bugs in the local `make e2e` release-gate harness (ambient-plugin resolution, a missing local Pulumi backend directory, passphrase handling, a stack-selection mismatch, and an incorrect relative path in `run-pulumi.sh`) discovered while validating this and the prior release.

---

## v0.5.2 (2026-07-21)

Security and maintenance release. Closes the `js-yaml` quadratic-CPU-consumption advisory (GHSA-52cp-r559-cp3m) left open by an incomplete automated dependency update, plus routine transitive dependency and CI tooling refreshes. No provider API or resource behavior changes — existing programs require no updates.

### Security

- **`js-yaml` quadratic CPU consumption** (GHSA-52cp-r559-cp3m): Upgraded `js-yaml` from v4.2.0 to v4.3.0 in `sdk/nodejs` and `claude/ts-test`. An automated dependency update batch had already fixed a critical `tar` decompression DoS and a high-severity `brace-expansion` DoS in the same npm dependency tree, but an internal error while bumping `@opentelemetry/core` caused `js-yaml` to silently drop out of that batch. `npm audit` now reports zero vulnerabilities in both directories.
- **`@opentelemetry/core` transitive bump**: `claude/ts-test` picks up the `@opentelemetry/core` v1.30.1 → v2.9.0+ bump previously applied only to `sdk/nodejs`, resolving 13 related moderate-severity advisories across the `@opentelemetry/*` package family.

### Dependency Updates

- `golang.org/x/net` v0.53.0/v0.54.0 → v0.55.0 (and transitive `golang.org/x/{crypto,sys,term,text}` peers) across the provider, Go SDK, and internal Go test harness modules
- GitHub Actions tooling: `actions/setup-go` v7, `actions/setup-dotnet` v6, `actions/setup-node` v7, `actions/setup-python` v7, `pypa/gh-action-pypi-publish` v1.14.1
- `ruby` v4.0.6 for the documentation site build
- Modernized the AI PR review slash-command workflow to use the upstream reusable workflow in place of a hand-rolled parser

---

## v0.5.1 (2026-06-01)

Security and maintenance release. Resolves seven reachable `golang.org/x/crypto` SSH vulnerabilities in the provider binary, adds this documentation site, and refreshes transitive dependencies and tooling. No provider API or resource behavior changes — existing programs require no updates.

### Security

- **`golang.org/x/crypto` SSH vulnerabilities** (GO-2026-5013, GO-2026-5015, GO-2026-5017, GO-2026-5018, GO-2026-5019, GO-2026-5020, GO-2026-5021): Upgraded `golang.org/x/crypto` from v0.50.0 to v0.52.0 in both the provider and Go SDK modules. `govulncheck` confirmed all seven were reachable through the provider's SSH key handling and reports zero reachable vulnerabilities after the upgrade.

### Documentation

- **GitHub Pages documentation site**: Added this Jekyll-based documentation site covering installation, quick-start, the complete resource reference, guides, examples, and troubleshooting.
- **README and reference updates**: Documented the `User`, `UserGroupAssignment`, and `UserPlatformRole` resources, corrected resource and test counts, fixed import-ID formats, and refreshed the Go prerequisite version.

### Build and CI

- **Pulumi CLI pin updated to 3.244.0** so local SDK generation matches the CI toolchain.
- **GitHub Pages deployment workflow** added to build and publish this site.
- **AI PR review inputs**: Wired additional engine and feature inputs into the AI review workflow.
- Removed an unused code-review tooling configuration file.

### Dependency Updates

- `golang.org/x/crypto` v0.50.0 → v0.52.0 (security; see above), with companion `golang.org/x/{net,sys,term,text}` upgrades
- Transitive Go module bumps: `go-git/go-git` v5.18.0 → v5.19.1, `cyphar/filepath-securejoin` v0.4.1 → v0.6.1, `pjbgf/sha1cd` v0.3.2 → v0.6.0, and the `golang.org/x/{mod,sync,tools,exp}` suite
- Transitive npm dependency updates in the Node.js SDK
- Ruby 4.0.5 for the documentation site build

---

## v0.5.0 (2026-05-06)

Feature release exposing the project deploy key as a Pulumi output, plus Makefile portability improvements and CI enhancements.

### New Features

**Expose project deploy key as `publicKey` output**

`Project` now exposes a `publicKey` output containing the SSH deploy key generated by Lagoon during project creation. This allows you to programmatically add the deploy key to your Git repository without needing the Lagoon CLI or UI — essential for fully automated infrastructure-as-code workflows.

```python
project = lagoon.Project("my-project",
    lagoon.ProjectArgs(
        name="my-project",
        git_url="git@github.com:org/repo.git",
        deploytarget_id=1,
        production_environment="main",
        branches="main",
    )
)

# Use the deploy key output to configure the Git repository
pulumi.export("deploy_key", project.public_key)
```

### Build and CI

- **Makefile portability**: Fixed 14 `sed` commands for BSD/macOS compatibility
- **Pulumi CLI version pinning**: SDK generation targets enforce the version pinned in `.pulumiversion` to prevent codegen drift
- **AI review for fork PRs**: Switched to `pull_request_target` trigger so the review workflow runs on external contributor PRs
- **Schema change documentation**: Added "Making Schema Changes" guidance to CONTRIBUTING.md and the PR template

### Dependency Updates

- golangci-lint v2.12.2
- Pulumi GitHub Actions v7
- npm dependency updates

---

## v0.4.1 (2026-05-01)

Patch release fixing the provider replace cascade that caused every resource to be replaced on every `pulumi up` when config inputs were re-evaluated.

### Bug Fixes

**Fix provider replace cascade on config changes**

The provider now implements `DiffConfig` to prevent unnecessary provider replacements when configuration values change. Previously, any change to `jwtSecret`, `token`, `apiUrl`, or other config fields triggered a provider replace, which cascaded into replacing every resource associated with the provider.

No provider config change requires a replace — changing credentials or the API URL only affects how the provider authenticates, not which resources it manages. The diff also normalizes whitespace so trailing newlines in secrets are not detected as changes. Empty `jwtAudience` and `"api.dev"` are treated as equivalent, matching the runtime default.

---

## v0.4.0 (2026-05-01)

Feature release adding user resource management: full CRUD for Lagoon users, group role assignments, and platform role assignments.

### New Resources

**`lagoon:lagoon:User`** — Full CRUD for Lagoon users via the GraphQL API. Uses email as the primary identifier. Supports optional `firstName`, `lastName`, and `comment` fields with in-place updates.

**`lagoon:lagoon:UserGroupAssignment`** — Assigns a user to a Lagoon group with a specific role (`GUEST`, `REPORTER`, `DEVELOPER`, `MAINTAINER`, or `OWNER`). Role changes are applied in-place via Lagoon's upsert semantics.

**`lagoon:lagoon:UserPlatformRole`** — Assigns a platform-level role (`OWNER` or `VIEWER`) to a user. Both fields are force-new; changing either triggers a replace.

```python
import pulumi
import pulumi_lagoon as lagoon

admin_user = lagoon.User("lagoonadmin",
    lagoon.UserArgs(
        email="admin@lagoon.example.com",
        first_name="Lagoon",
        last_name="Admin",
    )
)

lagoon.UserPlatformRole("lagoonadmin-platform-owner",
    lagoon.UserPlatformRoleArgs(
        user_email=admin_user.email,
        role="OWNER",
    )
)

team_group = lagoon.Group("mysite-team",
    lagoon.GroupArgs(name="project-mysite")
)

lagoon.UserGroupAssignment("lagoonadmin-team",
    lagoon.UserGroupAssignmentArgs(
        user_email=admin_user.email,
        group_name=team_group.name,
        role="MAINTAINER",
    )
)
```

**Limitations in this release:**
- SSH key management is not included. Use the Lagoon UI/CLI or a `pulumi-command` resource.
- Direct user-to-project role assignment is not supported. Grant project access through a group using `UserGroupAssignment`.

### Bug Fixes

**Fix JWT secret whitespace causing "invalid signature"** — The provider now trims leading and trailing whitespace from `jwtSecret`, `token`, and `jwtAudience` values before use. Trailing newlines from shell pipelines would silently corrupt the HMAC signing key.

**Debug logging for token generation** — The provider logs debug messages during `Configure` showing whether a token was generated from a JWT secret or loaded from an environment variable. Visible with `pulumi --logtostderr -v=9`.

---

For the complete release history including v0.3.0 (.NET/C# SDK), v0.2.x (native Go provider, route resources, group resources), and v0.1.x (original Python dynamic provider), see [RELEASE_NOTES.md](https://github.com/tag1consulting/pulumi-lagoon-provider/blob/main/RELEASE_NOTES.md) on GitHub.
