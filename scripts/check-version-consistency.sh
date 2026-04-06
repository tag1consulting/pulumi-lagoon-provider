#!/usr/bin/env bash
# Check that version strings match across all provider files

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$REPO_ROOT"

# Extract versions from each file
VERSION_MAIN=$(grep -E '^var Version = ' provider/cmd/pulumi-resource-lagoon/main.go | sed 's/.*"\(.*\)".*/\1/')
VERSION_MAKEFILE=$(grep -E '^PROVIDER_VERSION \?= ' Makefile | awk '{print $3}')
VERSION_SCHEMA=$(grep -E '^\s*"version":' provider/schema.json | head -1 | sed 's/.*"\(.*\)".*/\1/')
VERSION_PYTHON=$(grep -E '^\s*version = ' sdk/python/pyproject.toml | sed 's/.*"\(.*\)".*/\1/')
VERSION_NODEJS=$(grep -E '^\s*"version":' sdk/nodejs/package.json | head -1 | sed 's/.*"\(.*\)".*/\1/')

echo "Version check:"
echo "  main.go:         $VERSION_MAIN"
echo "  Makefile:        $VERSION_MAKEFILE"
echo "  schema.json:     $VERSION_SCHEMA"
echo "  pyproject.toml:  $VERSION_PYTHON"
echo "  package.json:    $VERSION_NODEJS"
echo

# Check all match
if [ "$VERSION_MAIN" = "$VERSION_MAKEFILE" ] && \
   [ "$VERSION_MAIN" = "$VERSION_SCHEMA" ] && \
   [ "$VERSION_MAIN" = "$VERSION_PYTHON" ] && \
   [ "$VERSION_MAIN" = "$VERSION_NODEJS" ]; then
  echo "✓ All versions match: $VERSION_MAIN"
  exit 0
else
  echo "✗ Version mismatch detected:"
  [ "$VERSION_MAIN" != "$VERSION_MAKEFILE" ] && echo "  main.go ($VERSION_MAIN) != Makefile ($VERSION_MAKEFILE)"
  [ "$VERSION_MAIN" != "$VERSION_SCHEMA" ] && echo "  main.go ($VERSION_MAIN) != schema.json ($VERSION_SCHEMA)"
  [ "$VERSION_MAIN" != "$VERSION_PYTHON" ] && echo "  main.go ($VERSION_MAIN) != pyproject.toml ($VERSION_PYTHON)"
  [ "$VERSION_MAIN" != "$VERSION_NODEJS" ] && echo "  main.go ($VERSION_MAIN) != package.json ($VERSION_NODEJS)"
  exit 1
fi
