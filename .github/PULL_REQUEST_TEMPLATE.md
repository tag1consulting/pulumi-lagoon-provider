## Summary

<!-- Briefly describe what this PR does and link related issues (e.g. "Closes #123") -->

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring
- [ ] Other (please describe)

## Checklist

- [ ] Tests pass (`make go-test`)
- [ ] **Schema change?** The following all count as schema changes and require regeneration:
  - adding, removing, or renaming a field on a `*Args` or `*State` struct
  - changing or adding a `pulumi:"..."` struct tag
  - changing or adding an `Annotate` description
  - adding a new resource
  - changing any input/output type signature

  If any of the above applies, run `make go-schema && make go-sdk-all` and commit the
  regenerated `provider/schema.json` and all `sdk/{python,nodejs,go,dotnet}/` changes
  in this PR. CI's `verify-sdks` workflow will fail otherwise.

  Use the Pulumi CLI version pinned in [`.pulumiversion`](../.pulumiversion) — see
  [CONTRIBUTING.md](../CONTRIBUTING.md#prerequisites). If your local CLI version does
  not match, install the pinned version or note the mismatch in the PR description
  and ask a maintainer to regenerate.

## Test Plan

<!-- How has this been tested? -->
