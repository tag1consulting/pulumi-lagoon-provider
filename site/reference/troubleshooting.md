---
title: Troubleshooting
parent: Reference
nav_order: 4
---

# Troubleshooting

## Authentication Errors

### "lagoon authentication required"

The provider could not find any credentials. Set either `lagoon:token` or `lagoon:jwtSecret` in your Pulumi config:

```bash
pulumi config set lagoon:token <your-jwt-token> --secret
# or
pulumi config set lagoon:jwtSecret <your-jwt-secret> --secret
```

Alternatively, export `LAGOON_TOKEN` or `LAGOON_JWT_SECRET` as environment variables.

### "Legacy token invalid: invalid signature" or "invalid signature"

The JWT secret contains extra whitespace. This is common when extracting secrets from Kubernetes via shell pipelines (`base64 -d` on Linux adds a trailing newline). The provider trims whitespace automatically as of v0.4.0. If you are on an older version, upgrade:

```bash
pip install --upgrade pulumi-lagoon
```

### Token expired

If you are using `lagoon:token`, generate a fresh one:

```bash
lagoon login
lagoon config whoami   # confirm the token works
pulumi config set lagoon:token "$(cat ~/.lagoon/config.yaml | grep token | ...)" --secret
```

Prefer `lagoon:jwtSecret` for automated workflows — the provider refreshes tokens automatically and there is no expiry concern.

### Self-signed certificate errors

For development instances with private CA certificates, disable TLS verification:

```bash
pulumi config set lagoon:insecure true
```

{: .warning }
> Do not set `lagoon:insecure` to `true` in production. Use a properly signed certificate or add your CA to the system trust store instead.

---

## Import Issues

### "not found" when running pulumi import

The resource does not exist at the ID you specified, or the ID format is wrong. Verify:

1. The resource exists in Lagoon (use `lagoon list projects`, `lagoon list environments`, etc.)
2. You are using the correct ID format — see the [Import ID Reference](import-ids/)
3. IDs are case-sensitive: `My-Project` and `my-project` are different values

### Unexpected changes after import

Run `pulumi preview` after importing. If the preview shows changes, your Pulumi code does not match the values returned by the Lagoon API. Common mismatches:

- **Optional fields with API defaults**: Fields like `branches` and `pullrequests` on a project may have API-assigned defaults. Either set them explicitly in your code to match, or accept the update to let the provider write the desired values.
- **Fields the provider does not manage**: Some Lagoon fields (e.g., certain internal metadata) are read-only and will not appear in diffs.
- **Whitespace or encoding differences**: Variable values, webhook URLs, etc. must match exactly.

---

## Diff and State Issues

### Provider replaced on every `pulumi up`

Symptom: `pulumi preview` shows the provider resource itself being replaced, which then cascades to replacing all associated resources.

This was a bug fixed in v0.4.1. Upgrade the provider:

```bash
pip install --upgrade pulumi-lagoon
```

If you are already on v0.4.1+, check whether `jwtAudience` is set. The values `""` and `"api.dev"` are treated as equivalent; if your stack config has one and the provider default is the other, the diff engine may see them as different on older versions.

### Check the installed provider version

```bash
pulumi plugin ls | grep lagoon
```

If the plugin version does not match your SDK version, reinstall:

```bash
pulumi plugin install resource lagoon <version>
```

### State drifted from actual Lagoon resources

If resources were deleted directly in Lagoon (via UI or CLI) without going through Pulumi, the Pulumi state will reference resources that no longer exist. Options:

1. **Re-create**: run `pulumi up` — the provider's `Read` method detects missing resources and marks them for re-creation.
2. **Remove from state**: use `pulumi state delete <urn>` to remove the stale record without affecting Lagoon.
3. **Refresh**: run `pulumi refresh` to sync state from the live API (use with caution — this can mark resources as deleted if they temporarily return empty results from the Lagoon API).

---

## Build and Development Issues

### SDK regeneration produces unexpected diffs

The Pulumi CLI version used for SDK generation must match the version in `.pulumiversion`. Run:

```bash
make check-pulumi-version
```

If there is a mismatch, install the pinned version:

```bash
curl -fsSL https://get.pulumi.com | sh -s -- --version $(cat .pulumiversion)
```

### Go build fails with CGO errors

The provider binary must be built with `CGO_ENABLED=0`:

```bash
CGO_ENABLED=0 make go-build
```

The Makefile sets this automatically via the `go-build` target. If you are running `go build` directly, add the flag.

### `make go-test` fails with "connection refused"

Unit tests use a mock GraphQL server and do not require a live Lagoon instance. If you see connection errors, check that nothing is binding to the random port the mock server selects. Run `make go-test` in isolation to rule out port conflicts.
