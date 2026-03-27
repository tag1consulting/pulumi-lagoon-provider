# Release Process — Pulumi Lagoon Provider

**Last Updated**: 2026-03-27

## Claude Code Automation

Two Claude Code tools assist with releases:

### `/release` skill

Invoke with `/release` in Claude Code. Walks through the complete release interactively:
- Prompts for the target version
- Runs the `release-validator` agent as a pre-flight check
- Creates a worktree, runs `make release-prep`, updates the changelog
- Guides through PR, both git tags, and GitHub release creation
- Provides a post-release verification checklist

### `release-validator` agent

Pre-flight checklist runner. Used automatically by `/release`, or invoke manually:
> "Run the release-validator for v0.X.Y"

Checks: version consistency across all SDK manifests, test passage, go vet, changelog entry, clean working tree, and CI status on main. Reports PASS/FAIL per check with remediation steps.

---

## Pre-Release: `make release-prep`

One command handles version bumps, SDK regeneration, and tests:

```bash
make release-prep VERSION=0.X.Y
```

This target (in order):
1. Bumps version strings in three source-of-truth files **first**, so the
   provider binary and SDKs carry the new version from the start:
   - `provider/cmd/pulumi-resource-lagoon/main.go` (`var Version`)
   - `Makefile` (`PROVIDER_VERSION`)
   - `provider/schema.json` (`version` field)
2. Rebuilds the provider binary with the new version embedded via ldflags
3. Regenerates all SDKs (Python, Node.js, Go) — uses a temp directory so
   hand-maintained files (README.pypi.md, package-lock.json, go.mod/go.sum,
   pyproject.toml) are preserved; rsync uses `--delete` on generated subdirs
   so stale files are removed
4. Updates package manifest version strings:
   - `sdk/python/pyproject.toml` (`version`)
   - `sdk/nodejs/package.json` (first `version` field)
   - Note: `sdk/python/pulumi_lagoon/pulumi-plugin.json` is now updated
     automatically by the SDK regeneration (no separate sed step)
5. Runs the full test suite

After `release-prep` completes, still do manually:
- Update `RELEASE_NOTES.md`
- Live testing against kind-lagoon cluster (pulumi up / refresh / destroy)
- Commit, push, and open PR to main

## Release Steps

### 1. Tag the release on main

```bash
git tag v0.X.Y
git push origin v0.X.Y
```

### 2. Create the Go module subdirectory tag

**IMPORTANT**: Push this tag **before** creating the GitHub release so the
`warm-go-proxy` CI job finds it when the release event fires.

The Go module proxy requires a subdirectory-prefixed tag for nested modules.
Without this, `go get` cannot resolve the SDK at a tagged version and falls
back to pseudo-versions.

```bash
git tag "sdk/go/lagoon/v0.X.Y" "v0.X.Y^{}"
git push origin "sdk/go/lagoon/v0.X.Y"
```

Note: Use `v0.X.Y^{}` (not `v0.X.Y`) to dereference the annotated tag and point
to the underlying commit. Otherwise Git creates a nested tag-of-a-tag.

### 3. Create the GitHub Release

```bash
gh release create v0.X.Y --title "v0.X.Y" --notes-file RELEASE_NOTES_EXCERPT.md
```

This triggers the `publish.yml` workflow which:
- Builds the Python SDK wheel; tests on Python 3.9, 3.11, 3.12; publishes to PyPI via OIDC
- Builds the TypeScript SDK; tests on Node.js 22, 24; publishes to npm via OIDC
- Warms the Go module proxy (`warm-go-proxy` job)

### 4. Post-Release Verification

- [ ] PyPI: `pip install pulumi-lagoon==0.X.Y` works
- [ ] npm: `npm view @tag1consulting/pulumi-lagoon version` shows `0.X.Y`
- [ ] Go proxy: `make go-proxy-warmup VERSION=0.X.Y` returns 200 OK for all endpoints (automated via CI, run manually as fallback)
- [ ] pkg.go.dev: `https://pkg.go.dev/github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon@v0.X.Y` shows documentation (may take up to an hour after proxy warm-up)
- [ ] GitHub release page shows correct notes

## Known Gotchas

### Go module tags (discovered in #64)
The Go SDK lives at `sdk/go/lagoon/` with its own `go.mod`. The Go module proxy
matches tags by stripping the module's subdirectory prefix, so it looks for
`sdk/go/lagoon/v0.X.Y`, NOT `v0.X.Y`. If you only push the repo-level tag,
the proxy returns 404 and users get pseudo-versions.

### SDK regeneration and hand-maintained files
`pulumi package gen-sdk` replaces the entire SDK output directory. The Makefile
targets (`go-sdk-python`, `go-sdk-nodejs`, `go-sdk-go`) generate into a temp
directory (`.sdk-gen-tmp/`) and rsync over, so hand-maintained files are preserved.
If you run `pulumi package gen-sdk` directly, you'll lose `go.mod`, `go.sum`,
`README.pypi.md`, `package-lock.json`, and the pyproject.toml license fix.
Always use the Makefile targets.

### PyPI README
The publish workflow copies `sdk/python/README.pypi.md` (Python-specific) to
`sdk/python/README.md` at build time. Keep `README.pypi.md` updated when adding
new resources or changing configuration.

### `release-prep` self-modifies the Makefile (discovered in #68)
The `sed -i 's/PROVIDER_VERSION ?= .*/PROVIDER_VERSION ?= $(VERSION)/' Makefile` command
modifies the Makefile in place — including the sed command line itself, because
`PROVIDER_VERSION ?= .*` matches the sed command line. The `^` anchor fixes this:
`sed -i 's/^PROVIDER_VERSION ?= .*/PROVIDER_VERSION ?= $(VERSION)/' Makefile`
The tab-indented recipe line starts with `\t`, so the anchored pattern never matches it.

### Node.js SDK has two version fields
`sdk/nodejs/package.json` has both `.version` (top-level) and `.pulumi.version` (nested).
Both must be updated in `release-prep`. The `jq` expression uses:
`.version = $v | .pulumi.version = $v`

### Go SDK LICENSE file required for pkg.go.dev (discovered in #69)
pkg.go.dev uses `licensecheck` to scan the module zip. The Go module proxy creates zip
archives scoped to the module subdirectory (`sdk/go/lagoon/`). The repo root `LICENSE`
is **outside that boundary** and invisible to the checker — even though Apache 2.0 is
on the approved list, pkg.go.dev shows "Documentation not displayed due to license
restrictions" without it.

Fix: `sdk/go/lagoon/LICENSE` must exist. The `go-sdk-go` Makefile target automates this:
```makefile
rsync -a --delete --exclude='go.mod' --exclude='go.sum' --exclude='LICENSE' ...
cp LICENSE sdk/go/lagoon/LICENSE
```
The `--exclude='LICENSE'` prevents rsync from deleting the file during regeneration.

### npm publishing requires GitHub environment setup (OIDC)
The `publish-npm` job in `publish.yml` uses OIDC trusted publishing (no `NPM_TOKEN`
secret). It requires:
1. A GitHub environment named `npm` in the repository settings
2. npm trusted publishing configured at npmjs.com for `@tag1consulting/pulumi-lagoon`
   (Granular Access Token → OIDC, linked to this repository and environment)
3. The first publish uses `--access public` (required for scoped packages)

### Go proxy warm-up timing
The `warm-go-proxy` CI job fires when the GitHub release is published. It assumes
the `sdk/go/lagoon/v0.X.Y` tag already exists at that point. **Always push both
tags before creating the GitHub release** — if the Go tag is missing, the job
fails with a clear error message. Run `make go-proxy-warmup VERSION=0.X.Y`
manually after pushing the tag to recover.

### Token expiry during live testing
Keycloak OAuth tokens expire in 5 minutes. For live testing, use JWT admin tokens
generated from the `JWTSECRET` (1-hour expiry). The `run-pulumi.sh` script handles
this automatically. When setting tokens in Pulumi stack config, use `lagoon:token`
with a fresh JWT — avoid `lagoon:jwtSecret` in config as it creates provider
replacement churn.
