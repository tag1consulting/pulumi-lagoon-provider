"""Lagoon remote (build-deploy) installation for multi-cluster Lagoon example.

This module provides functions for installing the lagoon-remote Helm chart
which includes the lagoon-build-deploy subchart. The build-deploy controller
handles building and deploying Lagoon projects on remote clusters via RabbitMQ
message queue communication with lagoon-core.

Installation follows the official Lagoon documentation:
https://docs.lagoon.sh/installing-lagoon/install-lagoon-remote/

The lagoon-remote chart includes:
- lagoon-build-deploy: Build controller that processes build/deploy messages
- docker-host: Docker-in-Docker service for building container images
- dbaas-operator: (optional) Database-as-a-Service operator
- ssh-portal: (optional) SSH access to running environments
"""

from typing import Optional, Union

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
    external_rabbitmq_host: Optional[Union[str, pulumi.Output[str]]] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> LagoonRemoteOutputs:
    """Install Lagoon remote using Helm.

    This installs the lagoon-remote chart which includes:
    - lagoon-build-deploy: The build controller that processes messages from RabbitMQ
    - docker-host: Docker-in-Docker for building container images
    - Supporting infrastructure (priority classes, network policies, etc.)

    The build-deploy controller connects to RabbitMQ in lagoon-core and processes
    build/deploy requests, executing them on the local cluster.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        core_outputs: Outputs from the Lagoon core installation
        secrets: Generated secrets for Lagoon
        target_name: Name for this deploy target (must match Lagoon deploy target)
        is_production: Whether this is a production cluster
        namespace_config: Optional namespace configuration
        external_rabbitmq_host: Optional external hostname:port for RabbitMQ
            (for cross-cluster communication). If provided, uses this
            instead of the internal Kubernetes service name.
        opts: Pulumi resource options

    Returns:
        LagoonRemoteOutputs with remote installation information
    """
    ns_config = namespace_config or NamespaceConfig()
    namespace = ns_config.lagoon_remote

    # Create namespace with appropriate labels
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

    # Create SSH key secret for build-deploy controller
    # This key is used by the controller to authenticate to lagoon-core SSH service
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

    # Build Helm values for lagoon-remote chart
    # The chart structure is:
    # - Top-level values for lagoon-remote (docker-host, etc.)
    # - lagoon-build-deploy.* for the build controller subchart
    # - global.* for values shared across subcharts
    helm_values = {
        # Global values shared across chart and subcharts
        "global": {
            "rabbitMQUsername": "lagoon",
            "rabbitMQPassword": core_outputs.rabbitmq_password,
            "rabbitMQHostname": rabbitmq_hostname,
        },
        # Enable the lagoon-build-deploy subchart
        # This is the critical setting that was missing!
        "lagoon-build-deploy": {
            "enabled": True,
            # Deploy target name - must match the name registered in Lagoon
            "lagoonTargetName": target_name,
            # RabbitMQ connection (also set via global, but explicit here)
            "rabbitMQHostname": rabbitmq_hostname,
            "rabbitMQPassword": core_outputs.rabbitmq_password,
            "rabbitMQUsername": "lagoon",
            # Token service (SSH) - for authenticating build pods
            "lagoonTokenHost": core_outputs.ssh_host,
            "lagoonTokenPort": "22",
            # Lagoon API endpoint (without /graphql suffix)
            "lagoonAPIHost": core_outputs.api_url.replace("/graphql", ""),
            # Feature flags for production environments
            **({"lagoonFeatureFlagForceRootlessWorkload": "enabled"} if is_production else {}),
        },
        # Docker host configuration
        # Note: storage.size cannot be changed after initial creation (StatefulSet limitation)
        "dockerHost": {},
        # Disable optional components for simpler setup
        "dbaas-operator": {
            "enabled": False,
        },
        "sshPortal": {
            "enabled": False,
        },
        # Enable ssh-core service account for lagoon-core connectivity
        "sshCore": {
            "enabled": True,
        },
    }

    # Generate a short release name to avoid Kubernetes name length limits
    short_name = name.split("-")[0] if "-" in name else name
    release_name = f"{short_name}-remote"

    # Install lagoon-remote using Helm
    release = k8s.helm.v3.Release(
        f"{name}-release",
        name=release_name,
        chart="lagoon-remote",
        version=VERSIONS["lagoon_remote"],
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
    rabbitmq_host: Union[str, pulumi.Output[str]],
    rabbitmq_password: pulumi.Output[str],
    lagoon_token_host: str,
    lagoon_api_host: str,
    target_name: str,
    is_production: bool = False,
    namespace_config: Optional[NamespaceConfig] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> LagoonRemoteOutputs:
    """Install Lagoon remote with minimal configuration.

    This is a simpler version that takes explicit connection parameters
    instead of requiring a LagoonCoreOutputs object. Useful when connecting
    to an existing Lagoon installation.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        rabbitmq_host: RabbitMQ broker hostname:port (e.g., "broker.example.com:5672")
        rabbitmq_password: RabbitMQ password
        lagoon_token_host: Token service (SSH) hostname
        lagoon_api_host: Lagoon API hostname (without /graphql suffix)
        target_name: Name for this deploy target
        is_production: Whether this is a production cluster
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

    # Build Helm values for lagoon-remote chart with lagoon-build-deploy subchart
    helm_values = {
        # Global values shared across chart and subcharts
        "global": {
            "rabbitMQUsername": "lagoon",
            "rabbitMQPassword": rabbitmq_password,
            "rabbitMQHostname": rabbitmq_host,
        },
        # Enable the lagoon-build-deploy subchart
        "lagoon-build-deploy": {
            "enabled": True,
            "lagoonTargetName": target_name,
            "rabbitMQHostname": rabbitmq_host,
            "rabbitMQPassword": rabbitmq_password,
            "rabbitMQUsername": "lagoon",
            "lagoonTokenHost": lagoon_token_host,
            "lagoonTokenPort": "22",
            "lagoonAPIHost": lagoon_api_host,
            **({"lagoonFeatureFlagForceRootlessWorkload": "enabled"} if is_production else {}),
        },
        # Docker host configuration
        # Note: storage.size cannot be changed after initial creation (StatefulSet limitation)
        "dockerHost": {},
        # Disable optional components
        "dbaas-operator": {
            "enabled": False,
        },
        "sshPortal": {
            "enabled": False,
        },
        "sshCore": {
            "enabled": True,
        },
    }

    # Generate a short release name
    short_name = name.split("-")[0] if "-" in name else name
    release_name = f"{short_name}-remote"

    # Install lagoon-remote using Helm
    release = k8s.helm.v3.Release(
        f"{name}-release",
        name=release_name,
        chart="lagoon-remote",
        version=VERSIONS["lagoon_remote"],
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
