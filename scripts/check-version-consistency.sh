#!/usr/bin/env bash
# Check that version strings match across all provider and SDK files.
#
# This script is the safety net for releases. Every published artifact
# that carries a version string should be listed here so a stale value
# fails CI rather than shipping silently.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$REPO_ROOT"

# Authoritative source: provider main.go.
VERSION_MAIN=$(grep -E '^var Version = ' provider/cmd/pulumi-resource-lagoon/main.go | sed 's/.*"\(.*\)".*/\1/')

# Each entry below is "Label|extracted value". Extraction is per-file because
# formats differ (TOML, JSON top-level, JSON nested, XML, plaintext).
declare -a checks=()

add_check() {
    checks+=("$1|$2")
}

add_check "Makefile"                    "$(grep -E '^PROVIDER_VERSION \?= ' Makefile | awk '{print $3}')"
add_check "schema.json"                 "$(jq -r '.version' provider/schema.json)"
add_check "pyproject.toml"              "$(grep -E '^\s*version = ' sdk/python/pyproject.toml | sed 's/.*"\(.*\)".*/\1/')"
add_check "package.json"                "$(jq -r '.version' sdk/nodejs/package.json)"
add_check "package-lock.json (root)"    "$(jq -r '.version' sdk/nodejs/package-lock.json)"
add_check "package-lock.json (pkg)"     "$(jq -r '.packages."".version' sdk/nodejs/package-lock.json)"
add_check "go pulumi-plugin.json"       "$(jq -r '.version' sdk/go/lagoon/pulumi-plugin.json)"
add_check "python pulumi-plugin.json"   "$(jq -r '.version' sdk/python/pulumi_lagoon/pulumi-plugin.json)"
add_check "dotnet pulumi-plugin.json"   "$(jq -r '.version' sdk/dotnet/pulumi-plugin.json)"
add_check "dotnet version.txt"          "$(tr -d '[:space:]' < sdk/dotnet/version.txt)"
add_check "dotnet csproj <Version>"     "$(grep -oE '<Version>[^<]+</Version>' sdk/dotnet/Tag1Consulting.Lagoon.csproj | sed -E 's#</?Version>##g')"

echo "Version check (expected: $VERSION_MAIN)"
echo "  main.go: $VERSION_MAIN"

mismatched=()
for entry in "${checks[@]}"; do
    label="${entry%%|*}"
    value="${entry#*|}"
    printf '  %-30s %s\n' "$label:" "$value"
    if [ "$value" != "$VERSION_MAIN" ]; then
        mismatched+=("$label ($value)")
    fi
done
echo

if [ ${#mismatched[@]} -eq 0 ]; then
    echo "✓ All versions match: $VERSION_MAIN"
    exit 0
fi

echo "✗ Version mismatch detected:"
for m in "${mismatched[@]}"; do
    echo "  main.go ($VERSION_MAIN) != $m"
done
exit 1
