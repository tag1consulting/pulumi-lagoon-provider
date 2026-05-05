# Contributing to pulumi-lagoon-provider

Contributions, bug reports, and feedback are welcome. Please follow our [Code of Conduct](CODE_OF_CONDUCT.md) in all interactions.

## Prerequisites

- Go 1.26+
- [Pulumi CLI](https://www.pulumi.com/docs/install/) — version must match `.pulumiversion` (currently 3.234.0). SDK generation output varies between CLI versions, so mismatches cause spurious diffs. Install a specific version with: `curl -fsSL https://get.pulumi.com | sh -s -- --version $(cat .pulumiversion)`
- A running Lagoon instance with API credentials (required for integration tests only; unit tests are self-contained)

## Development Setup

```bash
git clone https://github.com/tag1consulting/pulumi-lagoon-provider.git
cd pulumi-lagoon-provider
make go-build   # Build the provider binary
make go-test    # Run unit tests (no Lagoon instance needed)
make go-vet     # Run Go static analysis
```

## Project Structure

```
provider/
  pkg/client/        # GraphQL API client — one file per resource domain
  pkg/resources/     # Pulumi resource CRUD implementations
    client_iface.go  # LagoonClient interface (update when adding resources)
  schema.json        # Pulumi schema — regenerated, do not hand-edit
sdk/                 # Generated SDKs — do not hand-edit
examples/            # Usage examples
```

## Adding a New Resource

1. **Add GraphQL queries** to `provider/pkg/client/queries.go`.

2. **Create the client file** `provider/pkg/client/<resource>.go` with CRUD methods implementing the Lagoon API calls. Return `errors.ErrNotFound` (from `provider/pkg/client/errors.go`) for missing resources. Add a `<resource>_test.go` alongside it.

3. **Add method signatures** for the new methods to the `LagoonClient` interface in `provider/pkg/resources/client_iface.go`.

4. **Create the resource file** `provider/pkg/resources/<resource>.go` following the pattern in `project.go`:
   - `type <Resource> struct{}` — empty struct as receiver
   - `type <Resource>Args struct` — input fields with `pulumi:"fieldName"` tags; use pointer types for optional fields
   - `type <Resource>State struct` — embeds Args, adds computed outputs (e.g., `LagoonID int`)
   - Implement `Annotate` on all three types
   - Implement five lifecycle methods: `Create`, `Update`, `Delete`, `Read`, `Diff`
     - `Create`: handle "already exists" by adopting the existing resource
     - `Delete`: treat "not found" as success (idempotent)
     - `Diff`: return `DetailedDiff` distinguishing `Update` vs `UpdateReplace`
   - Add a `<resource>_crud_test.go` using the mock client in `mock_client_test.go`

5. **Register the resource** in the provider constructor (see `provider/pkg/provider/`).

6. **Regenerate the schema and SDKs:**
   ```bash
   make go-schema    # Regenerates provider/schema.json
   make go-sdk-all   # Regenerates sdk/python, sdk/nodejs, sdk/go
   ```

7. **Verify tests pass:**
   ```bash
   make go-test
   ```

## Testing

Unit tests use a mock GraphQL server and require no live Lagoon instance. New resources must include:
- `provider/pkg/client/<resource>_test.go`
- `provider/pkg/resources/<resource>_crud_test.go`

Integration tests require a live Lagoon instance. See `examples/simple-project/` for setup.

## Code Style

- Format code with `gofmt` before committing.
- Run `make go-vet` and fix any reported issues.
- Follow the patterns established in existing resource files. Introduce new abstractions only after discussion.

## Pull Request Process

1. Fork the repository and create a branch off `main` (e.g., `feature/my-resource` or `fix/123-description`).
2. Make your changes with tests.
3. If you changed the schema, regenerate SDKs with `make go-sdk-all` (using the Pulumi CLI version from `.pulumiversion`) and include the generated changes in your PR. CI will verify that committed SDKs match what the pinned CLI produces.
4. Open a pull request describing what you changed and why, linking any related issues.

## Reporting Bugs and Requesting Features

Open a [GitHub issue](https://github.com/tag1consulting/pulumi-lagoon-provider/issues).

**Bug reports** should include: Pulumi version, provider version, Lagoon version, the resource configuration, and the full error output.

**Feature requests** should describe the Lagoon API capability and your use case.
