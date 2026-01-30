"""Lagoon database migrations management.

Lagoon v2.30.0 has a bug where the init container only runs migrations from
dist/migrations/lagoon/migrations/ (1 file) instead of the Knex migrations
in database/migrations/ (44 files that create base tables like openshift).

This module provides functions to check and run Knex migrations.
"""

import os
from typing import Optional

import pulumi
from pulumi_command import local as command


def ensure_knex_migrations(
    name: str,
    context: str,
    namespace: str,
    core_secrets_name: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> command.Command:
    """Ensure Knex migrations have been run on the Lagoon API database.

    This creates a Command resource that checks if the base schema tables
    exist and runs Knex migrations if they are missing.

    Args:
        name: Pulumi resource name prefix
        context: Kubernetes context (e.g., "kind-lagoon-prod")
        namespace: Lagoon core namespace (e.g., "lagoon-core")
        core_secrets_name: Name of the lagoon-core-secrets secret
        opts: Pulumi resource options

    Returns:
        Command resource that ensures migrations are run
    """
    # Get the script directory (relative to repository root)
    # The script is at scripts/ensure-knex-migrations.sh
    script_dir = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))
    script_path = os.path.join(script_dir, "scripts", "ensure-knex-migrations.sh")

    # Build the command to run the migration check
    # Use environment variables to configure the script
    migration_command = f"""
export KUBE_CONTEXT="{context}"
export LAGOON_NAMESPACE="{namespace}"
export CORE_SECRETS="{core_secrets_name}"

# Wait a bit for the API to be fully ready after Helm install
sleep 10

# Run the migration check script
"{script_path}"
"""

    return command.Command(
        f"{name}-ensure-knex-migrations",
        create=migration_command,
        # Also check on updates in case something changed
        update=migration_command,
        opts=pulumi.ResourceOptions(
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )


def check_knex_migrations_inline(
    name: str,
    context: str,
    namespace: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> command.Command:
    """Check if Knex migrations have been run (inline version without external script).

    This is a simpler version that directly executes kubectl commands to check
    and run migrations without requiring the external script.

    Args:
        name: Pulumi resource name prefix
        context: Kubernetes context (e.g., "kind-lagoon-prod")
        namespace: Lagoon core namespace (e.g., "lagoon-core")
        opts: Pulumi resource options

    Returns:
        Command resource that ensures migrations are run
    """
    # Inline script that checks and runs migrations
    migration_script = f'''
#!/bin/bash
set -e

CONTEXT="{context}"
NAMESPACE="{namespace}"

echo "Checking Knex migrations status..."

# Find the API pod
API_POD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get pods --no-headers 2>/dev/null | grep -E '\\-api\\-' | head -1 | awk '{{print $1}}')

if [ -z "$API_POD" ]; then
    echo "ERROR: Could not find API pod"
    exit 1
fi

echo "Found API pod: $API_POD"

# Wait for pod to be ready
echo "Waiting for API pod to be ready..."
kubectl --context "$CONTEXT" -n "$NAMESPACE" wait --for=condition=ready pod "$API_POD" --timeout=300s

# Check if openshift table exists (indicates Knex migrations have run)
echo "Checking if base schema exists..."
RESULT=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" exec "$API_POD" -c api -- \\
    sh -c 'node -e "
const knex = require(\\"knex\\")(\{{
    client: \\"mysql2\\",
    connection: \{{
        host: process.env.API_DB_HOST || \\"mariadb\\",
        port: process.env.API_DB_PORT || 3306,
        user: process.env.API_DB_USER || \\"api\\",
        password: process.env.API_DB_PASSWORD || \\"api\\",
        database: process.env.API_DB_DATABASE || \\"infrastructure\\"
    \}}
\}});

knex.schema.hasTable(\\"openshift\\").then(exists => \{{
    console.log(exists ? \\"EXISTS\\" : \\"MISSING\\");
    process.exit(0);
\}}).catch(err => \{{
    console.error(err.message);
    process.exit(1);
\}}).finally(() => knex.destroy());
"' 2>&1)

if echo "$RESULT" | grep -q "EXISTS"; then
    echo "Knex migrations already applied (openshift table exists)"
    exit 0
elif echo "$RESULT" | grep -q "MISSING"; then
    echo "Running Knex migrations..."
    kubectl --context "$CONTEXT" -n "$NAMESPACE" exec "$API_POD" -c api -- \\
        sh -c 'cd /app/services/api && npx knex migrate:latest --knexfile database/knexfile.js'

    echo "Verifying migrations..."
    VERIFY=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" exec "$API_POD" -c api -- \\
        sh -c 'node -e "
const knex = require(\\"knex\\")(\{{
    client: \\"mysql2\\",
    connection: \{{
        host: process.env.API_DB_HOST || \\"mariadb\\",
        port: process.env.API_DB_PORT || 3306,
        user: process.env.API_DB_USER || \\"api\\",
        password: process.env.API_DB_PASSWORD || \\"api\\",
        database: process.env.API_DB_DATABASE || \\"infrastructure\\"
    \}}
\}});

knex.schema.hasTable(\\"openshift\\").then(exists => \{{
    console.log(exists ? \\"EXISTS\\" : \\"MISSING\\");
    process.exit(0);
\}}).catch(err => \{{
    console.error(err.message);
    process.exit(1);
\}}).finally(() => knex.destroy());
"' 2>&1)

    if echo "$VERIFY" | grep -q "EXISTS"; then
        echo "Knex migrations completed successfully"
        exit 0
    else
        echo "ERROR: Migrations failed - openshift table still missing"
        exit 1
    fi
else
    echo "ERROR: Failed to check migration status: $RESULT"
    exit 1
fi
'''

    return command.Command(
        f"{name}-ensure-knex-migrations",
        create=migration_script,
        update=migration_script,
        opts=pulumi.ResourceOptions(
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )
