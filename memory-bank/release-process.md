# Release Process — Pulumi Lagoon Provider

**Last Updated**: 2026-03-26

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
git tag -a v0.X.Y -m "Release v0.X.Y"
git push origin v0.X.Y
```

### 2. Create the Go module subdirectory tag

**IMPORTANT**: The Go module proxy requires a subdirectory-prefixed tag for nested
modules. Without this, `go get` cannot resolve the SDK at a tagged version and
falls back to pseudo-versions.

```bash
git tag -a "sdk/go/lagoon/v0.X.Y" "v0.X.Y^{}" -m "Go SDK v0.X.Y"
git push origin "sdk/go/lagoon/v0.X.Y"
```

Note: Use `v0.X.Y^{}` (not `v0.X.Y`) to dereference the annotated tag and point
to the underlying commit. Otherwise Git creates a nested tag-of-a-tag.

Verify it resolves on the Go proxy:

```bash
GOPROXY=https://proxy.golang.org go list -m \
  github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon@v0.X.Y
```

### 3. Create the GitHub Release

```bash
gh release create v0.X.Y --title "v0.X.Y" --notes-file RELEASE_NOTES_EXCERPT.md
```

This triggers the `publish.yml` workflow which:
- Builds the Python SDK wheel
- Tests installation on Python 3.9, 3.11, 3.12
- Publishes to PyPI (`pulumi-lagoon`)

### 4. Post-Release Verification

- [ ] PyPI: `pip install pulumi-lagoon==0.X.Y` works
- [ ] npm: `npm install @tag1consulting/pulumi-lagoon@0.X.Y` works (if npm publish is set up)
- [ ] Go: `go get github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon@v0.X.Y` resolves
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

### Token expiry during live testing
Keycloak OAuth tokens expire in 5 minutes. For live testing, use JWT admin tokens
generated from the `JWTSECRET` (1-hour expiry). The `run-pulumi.sh` script handles
this automatically. When setting tokens in Pulumi stack config, use `lagoon:token`
with a fresh JWT — avoid `lagoon:jwtSecret` in config as it creates provider
replacement churn.
