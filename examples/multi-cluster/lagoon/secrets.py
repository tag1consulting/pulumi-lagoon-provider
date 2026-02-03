"""Lagoon secrets generation for multi-cluster Lagoon example.

This module provides functions for generating SSH keys, passwords,
and other secrets required by Lagoon components.
"""

from typing import Optional

import pulumi
from config import LagoonSecretsOutputs
from pulumi_random import RandomPassword
from pulumi_tls import PrivateKey


def generate_lagoon_secrets(
    name: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> LagoonSecretsOutputs:
    """Generate all secrets needed for Lagoon deployment.

    This includes:
    - SSH key pair for build-deploy communication
    - RabbitMQ password
    - Keycloak admin password
    - API database password

    Args:
        name: Pulumi resource name prefix
        opts: Pulumi resource options

    Returns:
        LagoonSecretsOutputs with all generated secrets
    """
    parent = opts.parent if opts else None

    # Generate SSH key pair for Lagoon builds
    ssh_key = PrivateKey(
        f"{name}-ssh-key",
        algorithm="RSA",
        rsa_bits=4096,
        opts=pulumi.ResourceOptions(parent=parent),
    )

    # Generate RabbitMQ password
    rabbitmq_password = RandomPassword(
        f"{name}-rabbitmq-password",
        length=32,
        special=False,
        opts=pulumi.ResourceOptions(parent=parent),
    )

    # Generate Keycloak admin password
    keycloak_password = RandomPassword(
        f"{name}-keycloak-password",
        length=24,
        special=False,
        opts=pulumi.ResourceOptions(parent=parent),
    )

    # Generate API database password
    api_db_password = RandomPassword(
        f"{name}-api-db-password",
        length=32,
        special=False,
        opts=pulumi.ResourceOptions(parent=parent),
    )

    return LagoonSecretsOutputs(
        ssh_private_key=ssh_key.private_key_openssh,
        ssh_public_key=ssh_key.public_key_openssh,
        rabbitmq_password=rabbitmq_password.result,
        keycloak_admin_password=keycloak_password.result,
        api_db_password=api_db_password.result,
    )


def create_ssh_key_secret(
    name: str,
    namespace: str,
    private_key: pulumi.Output[str],
    public_key: pulumi.Output[str],
    provider,
    opts: Optional[pulumi.ResourceOptions] = None,
):
    """Create a Kubernetes secret containing SSH keys.

    Args:
        name: Pulumi resource name prefix
        namespace: Namespace to create the secret in
        private_key: SSH private key content
        public_key: SSH public key content
        provider: Kubernetes provider
        opts: Pulumi resource options

    Returns:
        Kubernetes Secret resource
    """

    import pulumi_kubernetes as k8s

    # Create secret with SSH keys
    return k8s.core.v1.Secret(
        f"{name}-ssh-secret",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=f"{name}-ssh-keys",
            namespace=namespace,
        ),
        type="Opaque",
        string_data={
            "id_rsa": private_key,
            "id_rsa.pub": public_key,
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )


def generate_jwt_secret(
    name: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> pulumi.Output[str]:
    """Generate a JWT secret for Lagoon API.

    Args:
        name: Pulumi resource name prefix
        opts: Pulumi resource options

    Returns:
        JWT secret string
    """
    secret = RandomPassword(
        f"{name}-jwt-secret",
        length=64,
        special=False,
        opts=pulumi.ResourceOptions(
            parent=opts.parent if opts else None,
        ),
    )
    return secret.result


def generate_keycloak_client_secret(
    name: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> pulumi.Output[str]:
    """Generate a Keycloak client secret.

    Args:
        name: Pulumi resource name prefix
        opts: Pulumi resource options

    Returns:
        Client secret string
    """
    secret = RandomPassword(
        f"{name}-keycloak-client-secret",
        length=32,
        special=False,
        opts=pulumi.ResourceOptions(
            parent=opts.parent if opts else None,
        ),
    )
    return secret.result
