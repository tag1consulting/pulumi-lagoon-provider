"""Keycloak configuration for Lagoon authentication.

This module provides functions for configuring Keycloak after Lagoon core
installation, including enabling Direct Access Grants for OAuth password
authentication and creating admin users.

Background:
The Lagoon UI uses Keycloak for authentication. By default, the lagoon-ui
client in Keycloak does not have Direct Access Grants enabled, which means
OAuth password grant (Resource Owner Password Credentials) doesn't work.
This is needed for:
- CLI tools that need to authenticate programmatically
- Scripts that automate Lagoon operations
- Testing authentication without a browser

This module creates a Kubernetes Job that runs after Lagoon core is installed
to configure Keycloak via its Admin REST API.
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s


def configure_keycloak_for_cli_auth(
    name: str,
    provider: k8s.Provider,
    namespace: str,
    keycloak_service: str,
    keycloak_admin_secret: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> k8s.batch.v1.Job:
    """Configure Keycloak to enable CLI/programmatic authentication.

    This creates a Kubernetes Job that:
    1. Waits for Keycloak to be ready
    2. Gets an admin token from Keycloak
    3. Enables Direct Access Grants for the lagoon-ui client
    4. Optionally creates the lagoonadmin user if it doesn't exist

    Direct Access Grants enables the OAuth "password" grant type, which allows
    CLI tools and scripts to authenticate using username/password without
    requiring a browser-based OAuth flow.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        namespace: Namespace where Keycloak is running
        keycloak_service: Name of the Keycloak service (e.g., "prod-core-lagoon-core-keycloak")
        keycloak_admin_secret: Name of the secret containing Keycloak admin credentials
        opts: Pulumi resource options

    Returns:
        Kubernetes Job resource
    """
    # The configuration script that runs inside the Job
    # This script:
    # 1. Waits for Keycloak to be ready
    # 2. Gets admin token
    # 3. Enables Direct Access Grants for lagoon-ui client
    # 4. Creates lagoonadmin user if it doesn't exist
    config_script = """#!/bin/sh
set -e

KEYCLOAK_URL="http://${KEYCLOAK_SERVICE}:8080/auth"
REALM="lagoon"
CLIENT_ID="lagoon-ui"

echo "Waiting for Keycloak to be ready..."
max_attempts=60
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if curl -sf "${KEYCLOAK_URL}/realms/${REALM}/.well-known/openid-configuration" > /dev/null 2>&1; then
        echo "Keycloak is ready"
        break
    fi
    attempt=$((attempt + 1))
    echo "Waiting for Keycloak... (attempt $attempt/$max_attempts)"
    sleep 5
done

if [ $attempt -eq $max_attempts ]; then
    echo "ERROR: Keycloak did not become ready in time"
    exit 1
fi

echo "Getting admin token..."
ADMIN_TOKEN=$(curl -sf -X POST "${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "client_id=admin-cli" \
    -d "grant_type=password" \
    -d "username=admin" \
    -d "password=${KEYCLOAK_ADMIN_PASSWORD}" | jq -r '.access_token')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo "ERROR: Failed to get admin token"
    exit 1
fi
echo "Got admin token"

echo "Getting lagoon-ui client configuration..."
CLIENT_CONFIG=$(curl -sf "${KEYCLOAK_URL}/admin/realms/${REALM}/clients?clientId=${CLIENT_ID}" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}")

CLIENT_UUID=$(echo "$CLIENT_CONFIG" | jq -r '.[0].id')
DIRECT_ACCESS_ENABLED=$(echo "$CLIENT_CONFIG" | jq -r '.[0].directAccessGrantsEnabled')

if [ -z "$CLIENT_UUID" ] || [ "$CLIENT_UUID" = "null" ]; then
    echo "ERROR: lagoon-ui client not found"
    exit 1
fi

echo "Client UUID: $CLIENT_UUID"
echo "Current directAccessGrantsEnabled: $DIRECT_ACCESS_ENABLED"

if [ "$DIRECT_ACCESS_ENABLED" = "true" ]; then
    echo "Direct Access Grants already enabled"
else
    echo "Enabling Direct Access Grants for lagoon-ui client..."

    # Get full client config and update it
    FULL_CONFIG=$(curl -sf "${KEYCLOAK_URL}/admin/realms/${REALM}/clients/${CLIENT_UUID}" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}")

    UPDATED_CONFIG=$(echo "$FULL_CONFIG" | jq '.directAccessGrantsEnabled = true')

    curl -sf -X PUT "${KEYCLOAK_URL}/admin/realms/${REALM}/clients/${CLIENT_UUID}" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "$UPDATED_CONFIG"

    echo "Direct Access Grants enabled successfully"
fi

# Create lagoonadmin user if it doesn't exist
echo "Checking if lagoonadmin user exists..."
USER_EXISTS=$(curl -sf "${KEYCLOAK_URL}/admin/realms/${REALM}/users?username=lagoonadmin" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq 'length')

if [ "$USER_EXISTS" = "0" ]; then
    echo "Creating lagoonadmin user..."

    # Create the user
    curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/users" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "lagoonadmin",
            "enabled": true,
            "emailVerified": true,
            "credentials": [{
                "type": "password",
                "value": "'"${KEYCLOAK_LAGOON_ADMIN_PASSWORD}"'",
                "temporary": false
            }]
        }'

    # Get the new user's ID
    USER_ID=$(curl -sf "${KEYCLOAK_URL}/admin/realms/${REALM}/users?username=lagoonadmin" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq -r '.[0].id')

    if [ -n "$USER_ID" ] && [ "$USER_ID" != "null" ]; then
        echo "User created with ID: $USER_ID"

        # Get platform-owner role ID
        ROLE_ID=$(curl -sf "${KEYCLOAK_URL}/admin/realms/${REALM}/roles" \
            -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq -r '.[] | select(.name=="platform-owner") | .id')

        if [ -n "$ROLE_ID" ] && [ "$ROLE_ID" != "null" ]; then
            echo "Assigning platform-owner role..."
            curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/users/${USER_ID}/role-mappings/realm" \
                -H "Authorization: Bearer ${ADMIN_TOKEN}" \
                -H "Content-Type: application/json" \
                -d '[{"id": "'"${ROLE_ID}"'", "name": "platform-owner"}]'
            echo "Role assigned successfully"
        else
            echo "WARNING: platform-owner role not found"
        fi
    else
        echo "WARNING: Failed to get user ID after creation"
    fi
else
    echo "lagoonadmin user already exists"
fi

echo ""
echo "=== Keycloak Configuration Complete ==="
echo "Direct Access Grants: enabled for lagoon-ui client"
echo "lagoonadmin user: configured"
"""

    # Create ConfigMap with the script
    config_map = k8s.core.v1.ConfigMap(
        f"{name}-keycloak-config-script",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=f"{name}-keycloak-config",
            namespace=namespace,
        ),
        data={
            "configure-keycloak.sh": config_script,
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )

    # Create the Job that runs the configuration script
    job = k8s.batch.v1.Job(
        f"{name}-keycloak-config-job",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=f"{name}-keycloak-config",
            namespace=namespace,
        ),
        spec=k8s.batch.v1.JobSpecArgs(
            ttl_seconds_after_finished=300,  # Clean up 5 minutes after completion
            backoff_limit=3,
            template=k8s.core.v1.PodTemplateSpecArgs(
                spec=k8s.core.v1.PodSpecArgs(
                    restart_policy="Never",
                    containers=[
                        k8s.core.v1.ContainerArgs(
                            name="configure-keycloak",
                            image="curlimages/curl:8.5.0",
                            command=["/bin/sh", "/scripts/configure-keycloak.sh"],
                            env=[
                                k8s.core.v1.EnvVarArgs(
                                    name="KEYCLOAK_SERVICE",
                                    value=keycloak_service,
                                ),
                                k8s.core.v1.EnvVarArgs(
                                    name="KEYCLOAK_ADMIN_PASSWORD",
                                    value_from=k8s.core.v1.EnvVarSourceArgs(
                                        secret_key_ref=k8s.core.v1.SecretKeySelectorArgs(
                                            name=keycloak_admin_secret,
                                            key="KEYCLOAK_ADMIN_PASSWORD",
                                        ),
                                    ),
                                ),
                                k8s.core.v1.EnvVarArgs(
                                    name="KEYCLOAK_LAGOON_ADMIN_PASSWORD",
                                    value_from=k8s.core.v1.EnvVarSourceArgs(
                                        secret_key_ref=k8s.core.v1.SecretKeySelectorArgs(
                                            name=keycloak_admin_secret,
                                            key="KEYCLOAK_LAGOON_ADMIN_PASSWORD",
                                        ),
                                    ),
                                ),
                            ],
                            volume_mounts=[
                                k8s.core.v1.VolumeMountArgs(
                                    name="scripts",
                                    mount_path="/scripts",
                                ),
                            ],
                        ),
                    ],
                    volumes=[
                        k8s.core.v1.VolumeArgs(
                            name="scripts",
                            config_map=k8s.core.v1.ConfigMapVolumeSourceArgs(
                                name=config_map.metadata.name,
                                default_mode=0o755,
                            ),
                        ),
                    ],
                ),
            ),
        ),
        opts=pulumi.ResourceOptions(
            provider=provider,
            parent=opts.parent if opts else None,
            depends_on=[config_map] + (opts.depends_on if opts and opts.depends_on else []),
        ),
    )

    return job
