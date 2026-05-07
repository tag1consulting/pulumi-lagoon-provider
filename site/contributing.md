---
title: Contributing
nav_order: 6
---

# Contributing

Contributions, bug reports, and feedback are welcome. Please review the [Code of Conduct](https://github.com/tag1consulting/pulumi-lagoon-provider/blob/main/CODE_OF_CONDUCT.md) and [Security Policy](https://github.com/tag1consulting/pulumi-lagoon-provider/blob/main/SECURITY.md) before opening a pull request.

## Prerequisites

- Go 1.26+
- [Pulumi CLI](https://www.pulumi.com/docs/install/) — version must match `.pulumiversion`. SDK generation output varies between CLI versions, so mismatches cause spurious diffs in the generated files. Install the pinned version with:
  ```bash
  curl -fsSL https://get.pulumi.com | sh -s -- --version $(cat .pulumiversion)
  ```
- Docker (for running integration tests)
- Kind (for cluster-based examples and integration tests)
- kubectl
- jq

## Development Setup

```bash
git clone https://github.com/tag1consulting/pulumi-lagoon-provider.git
cd pulumi-lagoon-provider

# Build the provider binary
make go-build

# Run unit tests (no Lagoon instance required)
make go-test

# Run Go static analysis
make go-vet
```

## Project Structure

```text
provider/
  pkg/client/        # GraphQL API client — one file per resource domain
  pkg/resources/     # Pulumi resource CRUD implementations
    client_iface.go  # LagoonClient interface (update when adding resources)
  schema.json        # Pulumi schema — regenerated, do not hand-edit

sdk/                 # Generated SDKs — do not hand-edit
  python/
  nodejs/
  go/
  dotnet/

examples/            # Usage examples
  simple-project/
  single-cluster/
  multi-cluster/
```

{: .note }
> The `sdk/` directory and `provider/schema.json` are generated files. Never edit them directly — they will be overwritten on the next regeneration.

## Adding a New Resource

Follow these steps to add a new resource type:

1. **Add GraphQL queries** to `provider/pkg/client/queries.go`.

2. **Create the client file** `provider/pkg/client/<resource>.go` with CRUD methods implementing the Lagoon API calls. Return `errors.ErrNotFound` from `provider/pkg/client/errors.go` for missing resources. Add a `<resource>_test.go` alongside it.

3. **Add method signatures** for the new methods to the `LagoonClient` interface in `provider/pkg/resources/client_iface.go`.

4. **Create the resource file** `provider/pkg/resources/<resource>.go` following the pattern in `project.go`:
   - `type <Resource> struct{}` — empty struct as the receiver
   - `type <Resource>Args struct` — input fields with `pulumi:"fieldName"` tags; use pointer types for optional fields
   - `type <Resource>State struct` — embeds Args, adds computed outputs (e.g., `LagoonID int`)
   - Implement `Annotate` on all three types
   - Implement five lifecycle methods: `Create`, `Update`, `Delete`, `Read`, `Diff`
   - Add a `<resource>_crud_test.go` using the mock client in `mock_client_test.go`

5. **Register the resource** in the provider constructor in `provider/pkg/provider/`.

6. **Regenerate the schema and SDKs** (see [Making Schema Changes](#making-schema-changes)).

7. **Verify tests pass**: `make go-test`

## Making Schema Changes

Any change to the Pulumi-facing surface of a resource requires regenerating `provider/schema.json` and all four SDKs. This includes:

- Adding, removing, or renaming a field on a `*Args` or `*State` struct
- Changing or adding a `pulumi:"..."` struct tag
- Changing or adding an `Annotate` description
- Adding a new resource
- Changing any input or output type signature

Regenerate with the pinned Pulumi CLI version:

```bash
make check-pulumi-version   # Verify your CLI matches .pulumiversion
make go-schema              # Regenerate provider/schema.json
make go-sdk-all             # Regenerate sdk/python, sdk/nodejs, sdk/go, sdk/dotnet
```

Commit `provider/schema.json` and all `sdk/` changes in the same PR as the provider source change. CI's `verify-sdks` workflow fails any PR whose committed artifacts drift from a fresh regeneration with the pinned CLI version.

## Testing

Unit tests use a mock GraphQL server and require no live Lagoon instance:

```bash
make go-test
```

There are 490+ unit tests covering all resource types. New resources must include:
- `provider/pkg/client/<resource>_test.go`
- `provider/pkg/resources/<resource>_crud_test.go`

Integration tests require a live Lagoon instance. See `examples/simple-project/` for setup instructions.

## Code Style

- Format code with `gofmt` before committing. CI will reject improperly formatted Go files.
- Run `make go-vet` and fix all reported issues.
- Follow the patterns established in existing resource files. Introduce new abstractions only after discussion in a GitHub issue.

## Pull Request Process

1. Fork the repository and create a branch off `main` (e.g., `feature/my-resource` or `fix/123-description`).
2. Make your changes, including tests.
3. If you changed the schema, regenerate SDKs with `make go-sdk-all` using the pinned Pulumi CLI version, and include the generated changes in your PR.
4. Open a pull request describing what you changed and why, linking any related issues.
5. Check the schema change checkbox in the PR template if applicable.
6. CI will run tests and verify that committed SDKs match a fresh regeneration.

## Reporting Bugs and Requesting Features

Open a [GitHub issue](https://github.com/tag1consulting/pulumi-lagoon-provider/issues).

**Bug reports** should include: Pulumi version (`pulumi version`), provider version (`pulumi plugin ls | grep lagoon`), Lagoon version, the full resource configuration (with secrets redacted), and the complete error output.

**Feature requests** should describe the Lagoon API capability you want to expose and your use case.
