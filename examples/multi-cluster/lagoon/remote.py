"""Lagoon remote (build-deploy) installation for multi-cluster Lagoon example.

This module provides functions for installing the lagoon-build-deploy Helm chart
which handles building and deploying Lagoon projects on remote clusters.
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s

from config import (
    VERSIONS,
    NamespaceConfig,
    LagoonSecretsOutputs,
    LagoonCoreOutputs,
    LagoonRemoteOutputs,
)


def install_lagoon_remote(
    name: str,
    provider: k8s.Provider,
    core_outputs: LagoonCoreOutputs,
    secrets: LagoonSecretsOutputs,
    target_name: str,
    is_production: bool = False,
    namespace_config: Optional[NamespaceConfig] = None,
    external_rabbitmq_host: Optional[str] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> LagoonRemoteOutputs:
    """Install Lagoon build-deploy (remote) using Helm.

    This installs the build controller and related components that
    handle building and deploying projects on a remote cluster.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        core_outputs: Outputs from the Lagoon core installation
        secrets: Generated secrets for Lagoon
        target_name: Name for this deploy target
        is_production: Whether this is a production cluster
        namespace_config: Optional namespace configuration
        external_rabbitmq_host: Optional external hostname for RabbitMQ
            (for cross-cluster communication). If provided, uses this
            instead of the internal Kubernetes service name.
        opts: Pulumi resource options

    Returns:
        LagoonRemoteOutputs with remote installation information
    """
    ns_config = namespace_config or NamespaceConfig()
    namespace = ns_config.lagoon_remote

    # Create namespace
    ns = k8s.core.v1.Namespace(
        f"{name}-namespace",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=namespace,
            labels={
                "lagoon.sh/environment-type": "production" if is_production else "development",
            },
        ),
        opts=pulumi.ResourceOptions(
            provider=provider,
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )

    # Create SSH key secret for build-deploy
    ssh_secret = k8s.core.v1.Secret(
        f"{name}-ssh-secret",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name="lagoon-build-deploy-ssh",
            namespace=namespace,
        ),
        type="Opaque",
        string_data={
            "ssh-privatekey": secrets.ssh_private_key,
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[ns],
            parent=opts.parent if opts else None,
        ),
    )

    # Determine RabbitMQ hostname - use external if provided (cross-cluster),
    # otherwise use internal Kubernetes service name (same-cluster)
    rabbitmq_hostname = external_rabbitmq_host or core_outputs.rabbitmq_host

    # Build Helm values for lagoon-build-deploy 0.39.0+
    helm_values = {
        "lagoonTargetName": target_name,
        "rabbitMQHostname": rabbitmq_hostname,
        "rabbitMQPassword": core_outputs.rabbitmq_password,
        "rabbitMQUsername": "lagoon",
        # Token service (SSH) - uses the core SSH service
        "lagoonTokenHost": core_outputs.ssh_host,
        "lagoonTokenPort": "22",
        # Lagoon API endpoint (without /graphql)
        "lagoonAPIHost": core_outputs.api_url.replace("/graphql", ""),
        # Harbor integration
        "harbor": {
            "enabled": False,  # Disabled for simple setup
        },
    }

    # Add production-specific settings
    if is_production:
        helm_values["lagoonFeatureFlagForceRootlessWorkload"] = "enabled"

    # Generate a short release name
    short_name = name.split("-")[0] if "-" in name else name
    release_name = f"{short_name}-remote"

    # Install lagoon-build-deploy using Helm
    release = k8s.helm.v3.Release(
        f"{name}-release",
        name=release_name,
        chart="lagoon-build-deploy",
        version=VERSIONS["lagoon_build_deploy"],
        namespace=namespace,
        repository_opts=k8s.helm.v3.RepositoryOptsArgs(
            repo="https://uselagoon.github.io/lagoon-charts/",
        ),
        values=helm_values,
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[ns, ssh_secret],
            parent=opts.parent if opts else None,
        ),
    )

    return LagoonRemoteOutputs(
        release=release,
        namespace=namespace,
    )


def install_lagoon_remote_minimal(
    name: str,
    provider: k8s.Provider,
    rabbitmq_host: str,
    rabbitmq_password: pulumi.Output[str],
    lagoon_token_host: str,
    lagoon_api_host: str,
    target_name: str,
    namespace_config: Optional[NamespaceConfig] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> LagoonRemoteOutputs:
    """Install Lagoon build-deploy with minimal configuration.

    This is a simpler version that takes explicit connection parameters
    instead of requiring a LagoonCoreOutputs object.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        rabbitmq_host: RabbitMQ broker hostname
        rabbitmq_password: RabbitMQ password
        lagoon_token_host: Token service (SSH) hostname
        lagoon_api_host: Lagoon API hostname (without /graphql)
        target_name: Name for this deploy target
        namespace_config: Optional namespace configuration
        opts: Pulumi resource options

    Returns:
        LagoonRemoteOutputs with remote installation information
    """
    ns_config = namespace_config or NamespaceConfig()
    namespace = ns_config.lagoon_remote

    # Create namespace
    ns = k8s.core.v1.Namespace(
        f"{name}-namespace",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=namespace,
        ),
        opts=pulumi.ResourceOptions(
            provider=provider,
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )

    # Build Helm values for lagoon-build-deploy 0.39.0+
    helm_values = {
        "lagoonTargetName": target_name,
        "rabbitMQHostname": rabbitmq_host,
        "rabbitMQPassword": rabbitmq_password,
        "rabbitMQUsername": "lagoon",
        "lagoonTokenHost": lagoon_token_host,
        "lagoonTokenPort": "22",
        "lagoonAPIHost": lagoon_api_host,
        "harbor": {
            "enabled": False,
        },
    }

    # Generate a short release name
    short_name = name.split("-")[0] if "-" in name else name
    release_name = f"{short_name}-remote"

    # Install lagoon-build-deploy using Helm
    release = k8s.helm.v3.Release(
        f"{name}-release",
        name=release_name,
        chart="lagoon-build-deploy",
        version=VERSIONS["lagoon_build_deploy"],
        namespace=namespace,
        repository_opts=k8s.helm.v3.RepositoryOptsArgs(
            repo="https://uselagoon.github.io/lagoon-charts/",
        ),
        values=helm_values,
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[ns],
            parent=opts.parent if opts else None,
        ),
    )

    return LagoonRemoteOutputs(
        release=release,
        namespace=namespace,
    )
