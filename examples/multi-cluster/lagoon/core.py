"""Lagoon core installation for multi-cluster Lagoon example.

This module provides functions for installing the lagoon-core Helm chart
which includes Keycloak, API, UI, RabbitMQ, and other core components.
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s

from config import (
    VERSIONS,
    DomainConfig,
    NamespaceConfig,
    LagoonSecretsOutputs,
    LagoonCoreOutputs,
    HarborOutputs,
)


def install_lagoon_core(
    name: str,
    provider: k8s.Provider,
    domain_config: DomainConfig,
    secrets: LagoonSecretsOutputs,
    tls_secret_name: str,
    harbor: Optional[HarborOutputs] = None,
    ingress_class: str = "nginx",
    namespace_config: Optional[NamespaceConfig] = None,
    helm_timeout: int = 1800,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> LagoonCoreOutputs:
    """Install Lagoon core using Helm.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        domain_config: Domain configuration for Lagoon URLs
        secrets: Generated secrets for Lagoon
        tls_secret_name: Name of TLS secret for HTTPS
        harbor: Optional Harbor outputs for registry integration
        ingress_class: Ingress class to use (default: nginx)
        namespace_config: Optional namespace configuration
        helm_timeout: Helm release timeout in seconds (default: 1800 = 30 minutes)
        opts: Pulumi resource options

    Returns:
        LagoonCoreOutputs with Lagoon core information
    """
    ns_config = namespace_config or NamespaceConfig()
    namespace = ns_config.lagoon_core

    # Generate release name early (needed for internal service URLs)
    short_name = name.split("-")[0] if "-" in name else name
    release_name = f"{short_name}-core"

    # Note: Namespace should be created externally before calling this function
    # to ensure certificates can be created in the namespace first

    # Construct Lagoon URLs
    api_url = f"https://{domain_config.lagoon_api}/graphql"
    ui_url = f"https://{domain_config.lagoon_ui}"
    keycloak_url = f"https://{domain_config.lagoon_keycloak}"
    webhook_url = f"https://{domain_config.lagoon_webhook}"

    # Harbor URL and registry - required by lagoon-core chart
    # Use the provided Harbor or a default
    harbor_url = harbor.url if harbor else f"https://{domain_config.harbor}"
    # Registry is the Harbor hostname without protocol
    registry = harbor_url.replace("https://", "").replace("http://", "")
    # Harbor admin password - required for Lagoon to push images
    harbor_admin_password = harbor.admin_password if harbor else "Harbor2024!"

    # Internal Keycloak URL for API communication (must include /auth path)
    # Service name follows pattern: {release_name}-keycloak
    keycloak_internal_url = f"http://{release_name}-keycloak.{namespace}.svc.cluster.local:8080/auth"

    # Build Helm values
    # Note: lagoon-core requires many configuration values for S3/storage
    helm_values = {
        # Top-level Keycloak URL configuration - used by API and other services
        "keycloakAPIURL": keycloak_internal_url,
        "harborURL": harbor_url,
        "registry": registry,
        "harborAdminPassword": harbor_admin_password,
        "rabbitMQPassword": secrets.rabbitmq_password,
        "keycloakAdminPassword": secrets.keycloak_admin_password,
        # S3/MinIO configuration - required by lagoon-core for file storage
        # Using dummy values for local development (features disabled)
        "s3FilesAccessKeyID": "AKIAIOSFODNN7EXAMPLE",
        "s3FilesSecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
        "s3FilesBucket": "lagoon-files",
        "s3FilesHost": "s3.lagoon.local",
        # S3 for backups - also required
        "s3BAASAccessKeyID": "AKIAIOSFODNN7EXAMPLE",
        "s3BAASSecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
        # Elasticsearch/OpenSearch - required for logging
        "elasticsearchURL": "http://elasticsearch.lagoon.local:9200",
        "kibanaURL": "http://kibana.lagoon.local:5601",
        # Lagoon routes/domains configuration
        "lagoonRoutes": f".{domain_config.base}",
        "routeSuffix": f".{domain_config.base}",
        # Git server configuration
        "gitHost": f"git.{domain_config.base}",
        # API JWT secret
        "apiJWTSecret": "changeme-jwt-secret-for-local-dev-only",
        "api": {
            "enabled": True,
            "serviceMonitor": {
                "enabled": False,
            },
            "ingress": {
                "enabled": True,
                "ingressClassName": ingress_class,
                "annotations": {
                    "nginx.ingress.kubernetes.io/ssl-redirect": "false",
                },
                "hosts": [
                    {
                        "host": domain_config.lagoon_api,
                        "paths": ["/"],
                    }
                ],
                "tls": [
                    {
                        "secretName": tls_secret_name,
                        "hosts": [domain_config.lagoon_api],
                    }
                ],
            },
            "additionalEnvs": {
                "LAGOON_UI_URL": ui_url,
                # Disable TLS verification for any remaining HTTPS connections
                "NODE_TLS_REJECT_UNAUTHORIZED": "0",
            },
            "resources": {
                "requests": {
                    "cpu": "100m",
                    "memory": "256Mi",
                },
            },
        },
        "apiDB": {
            "enabled": True,
            "serviceMonitor": {
                "enabled": False,
            },
            "password": secrets.api_db_password,
            "resources": {
                "requests": {
                    "cpu": "100m",
                    "memory": "256Mi",
                },
            },
        },
        "ui": {
            "enabled": True,
            "ingress": {
                "enabled": True,
                "ingressClassName": ingress_class,
                "annotations": {
                    "nginx.ingress.kubernetes.io/ssl-redirect": "true",
                },
                "hosts": [
                    {
                        "host": domain_config.lagoon_ui,
                        "paths": ["/"],
                    }
                ],
                "tls": [
                    {
                        "secretName": tls_secret_name,
                        "hosts": [domain_config.lagoon_ui],
                    }
                ],
            },
            "resources": {
                "requests": {
                    "cpu": "50m",
                    "memory": "128Mi",
                },
            },
        },
        "keycloak": {
            "enabled": True,
            "serviceMonitor": {
                "enabled": False,
            },
            "ingress": {
                "enabled": True,
                "ingressClassName": ingress_class,
                "annotations": {
                    "nginx.ingress.kubernetes.io/ssl-redirect": "true",
                    "nginx.ingress.kubernetes.io/proxy-buffer-size": "64k",
                },
                "hosts": [
                    {
                        "host": domain_config.lagoon_keycloak,
                        "paths": ["/"],
                    }
                ],
                "tls": [
                    {
                        "secretName": tls_secret_name,
                        "hosts": [domain_config.lagoon_keycloak],
                    }
                ],
            },
            "resources": {
                "requests": {
                    "cpu": "100m",
                    "memory": "512Mi",
                },
            },
        },
        "broker": {
            "enabled": True,
            "serviceMonitor": {
                "enabled": False,
            },
            "service": {
                # Keep internal service as ClusterIP
                "type": "ClusterIP",
                # Disable chart's amqpExternal - we'll create our own with fixed NodePort
                "amqpExternal": {
                    "enabled": False,
                },
            },
            "resources": {
                "requests": {
                    "cpu": "100m",
                    "memory": "256Mi",
                },
            },
        },
        "webhookHandler": {
            "enabled": True,
            "ingress": {
                "enabled": True,
                "ingressClassName": ingress_class,
                "annotations": {
                    "nginx.ingress.kubernetes.io/ssl-redirect": "true",
                },
                "hosts": [
                    {
                        "host": domain_config.lagoon_webhook,
                        "paths": ["/"],
                    }
                ],
                "tls": [
                    {
                        "secretName": tls_secret_name,
                        "hosts": [domain_config.lagoon_webhook],
                    }
                ],
            },
            "resources": {
                "requests": {
                    "cpu": "50m",
                    "memory": "128Mi",
                },
            },
        },
        "webhooks2tasks": {
            "enabled": True,
            "resources": {
                "requests": {
                    "cpu": "50m",
                    "memory": "128Mi",
                },
            },
        },
        "actionsHandler": {
            "enabled": True,
            "resources": {
                "requests": {
                    "cpu": "50m",
                    "memory": "128Mi",
                },
            },
        },
        "autoIdler": {
            "enabled": False,
        },
        "storageCalculator": {
            "enabled": False,
        },
        "logs2notifications": {
            "enabled": False,
        },
        "drushAlias": {
            "enabled": False,
        },
        "ssh": {
            "enabled": True,
            "privateHostKey": secrets.ssh_private_key,
            "service": {
                # Use NodePort for Kind clusters (LoadBalancer stays pending without MetalLB)
                "type": "NodePort",
                "port": 22,
            },
            "resources": {
                "requests": {
                    "cpu": "50m",
                    "memory": "128Mi",
                },
            },
        },
        "sshPortal": {
            "enabled": False,
        },
        "controllerHandler": {
            "enabled": True,
            "resources": {
                "requests": {
                    "cpu": "50m",
                    "memory": "128Mi",
                },
            },
        },
        # Disable optional components for minimal setup
        "backupHandler": {
            "enabled": False,
        },
        "insightsHandler": {
            "enabled": False,
        },
    }

    # Install lagoon-core using Helm
    release = k8s.helm.v3.Release(
        f"{name}-release",
        name=release_name,
        chart="lagoon-core",
        version=VERSIONS["lagoon_core"],
        namespace=namespace,
        repository_opts=k8s.helm.v3.RepositoryOptsArgs(
            repo="https://uselagoon.github.io/lagoon-charts/",
        ),
        values=helm_values,
        # Lagoon core takes a long time to initialize - use configurable timeout
        timeout=helm_timeout,
        wait_for_jobs=True,
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=opts.depends_on if opts else None,
            parent=opts.parent if opts else None,
        ),
    )

    # RabbitMQ is exposed internally via service and externally via NodePort
    # Service name follows pattern: {release_name}-lagoon-core-broker
    rabbitmq_host = f"{release_name}-lagoon-core-broker.{namespace}.svc.cluster.local"
    rabbitmq_nodeport = 30672  # Matches amqpNodePort in helm values

    # SSH service for build connections
    # Service name follows pattern: {release_name}-lagoon-core-ssh
    ssh_host = f"{release_name}-lagoon-core-ssh.{namespace}.svc.cluster.local"

    return LagoonCoreOutputs(
        release=release,
        api_url=api_url,
        ui_url=ui_url,
        keycloak_url=keycloak_url,
        rabbitmq_host=rabbitmq_host,
        rabbitmq_password=secrets.rabbitmq_password,
        namespace=namespace,
        ssh_host=ssh_host,
        rabbitmq_nodeport=rabbitmq_nodeport,
    )


def create_rabbitmq_nodeport_service(
    name: str,
    release_name: str,
    namespace: str,
    provider: k8s.Provider,
    nodeport: int = 30672,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> k8s.core.v1.Service:
    """Create a NodePort service for RabbitMQ cross-cluster access.

    The lagoon-core Helm chart doesn't support setting a fixed NodePort,
    so we create our own service that points to the broker pods.

    Args:
        name: Pulumi resource name prefix
        release_name: Helm release name (e.g., "prod-core")
        namespace: Namespace where broker is deployed
        provider: Kubernetes provider
        nodeport: Fixed NodePort to use (default: 30672)
        opts: Pulumi resource options

    Returns:
        Kubernetes Service resource
    """
    return k8s.core.v1.Service(
        f"{name}-broker-nodeport",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=f"{release_name}-broker-external",
            namespace=namespace,
            labels={
                "app.kubernetes.io/component": "broker",
                "app.kubernetes.io/name": "lagoon-core",
            },
        ),
        spec=k8s.core.v1.ServiceSpecArgs(
            type="NodePort",
            ports=[
                k8s.core.v1.ServicePortArgs(
                    name="amqp",
                    port=5672,
                    target_port=5672,
                    node_port=nodeport,
                    protocol="TCP",
                ),
            ],
            # Select the broker pods using the same labels as the Helm chart
            # Note: The Helm chart sets component to "{release_name}-broker"
            selector={
                "app.kubernetes.io/component": f"{release_name}-broker",
                "app.kubernetes.io/instance": release_name,
                "app.kubernetes.io/name": "lagoon-core",
            },
        ),
        opts=pulumi.ResourceOptions(
            provider=provider,
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )


def create_lagoon_ssh_host_secret(
    name: str,
    namespace: str,
    ssh_private_key: pulumi.Output[str],
    provider: k8s.Provider,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> k8s.core.v1.Secret:
    """Create the SSH host key secret for Lagoon.

    Args:
        name: Pulumi resource name prefix
        namespace: Namespace to create the secret in
        ssh_private_key: SSH private key content
        provider: Kubernetes provider
        opts: Pulumi resource options

    Returns:
        Kubernetes Secret resource
    """
    return k8s.core.v1.Secret(
        f"{name}-ssh-host-key",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name="lagoon-sshportal-host-key",
            namespace=namespace,
        ),
        type="Opaque",
        string_data={
            "ssh_host_rsa_key": ssh_private_key,
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )
