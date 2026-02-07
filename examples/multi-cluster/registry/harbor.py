"""Harbor registry installation for multi-cluster Lagoon example.

This module provides functions for installing and configuring
Harbor container registry.
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s
from config import VERSIONS, DomainConfig, HarborOutputs, NamespaceConfig
from pulumi_random import RandomPassword


def install_harbor(
    name: str,
    provider: k8s.Provider,
    domain_config: DomainConfig,
    tls_secret_name: str,
    ingress_class: str = "nginx",
    admin_password: Optional[pulumi.Output[str]] = None,
    namespace_config: Optional[NamespaceConfig] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> HarborOutputs:
    """Install Harbor registry using Helm.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        domain_config: Domain configuration for Harbor URLs
        tls_secret_name: Name of TLS secret for HTTPS
        ingress_class: Ingress class to use (default: nginx)
        admin_password: Optional admin password (generated if not provided)
        namespace_config: Optional namespace configuration
        opts: Pulumi resource options

    Returns:
        HarborOutputs with registry information
    """
    ns_config = namespace_config or NamespaceConfig()
    namespace = ns_config.harbor

    # Note: Namespace should be created externally before calling this function
    # to ensure certificates can be created in the namespace first

    # Generate admin password if not provided
    if admin_password is None:
        password = RandomPassword(
            f"{name}-admin-password",
            length=24,
            special=False,
            opts=pulumi.ResourceOptions(
                parent=opts.parent if opts else None,
            ),
        )
        admin_pwd = password.result
    else:
        admin_pwd = admin_password

    # Harbor URLs
    harbor_url = f"https://{domain_config.harbor}"

    # Generate a short release name
    short_name = name.split("-")[0] if "-" in name else name
    release_name = f"{short_name}-harbor"

    # Install Harbor using Helm
    release = k8s.helm.v3.Release(
        f"{name}-release",
        name=release_name,
        chart="harbor",
        version=VERSIONS["harbor"],
        namespace=namespace,
        repository_opts=k8s.helm.v3.RepositoryOptsArgs(
            repo="https://helm.goharbor.io",
        ),
        values={
            "expose": {
                "type": "ingress",
                "tls": {
                    "enabled": True,
                    "certSource": "secret",
                    "secret": {
                        "secretName": tls_secret_name,
                    },
                },
                "ingress": {
                    "hosts": {
                        "core": domain_config.harbor,
                    },
                    "className": ingress_class,
                    "annotations": {
                        "nginx.ingress.kubernetes.io/ssl-redirect": "true",
                        "nginx.ingress.kubernetes.io/proxy-body-size": "0",
                    },
                },
            },
            "externalURL": harbor_url,
            "harborAdminPassword": admin_pwd,
            "persistence": {
                "enabled": True,
                "persistentVolumeClaim": {
                    "registry": {
                        "size": "5Gi",
                    },
                    "chartmuseum": {
                        "size": "1Gi",
                    },
                    "jobservice": {
                        "jobLog": {
                            "size": "1Gi",
                        },
                    },
                    "database": {
                        "size": "1Gi",
                    },
                    "redis": {
                        "size": "1Gi",
                    },
                    "trivy": {
                        "size": "1Gi",
                    },
                },
            },
            # Disable components not needed for dev
            "chartmuseum": {
                "enabled": False,
            },
            "trivy": {
                "enabled": False,
            },
            "notary": {
                "enabled": False,
            },
            # Resource limits for local development
            "core": {
                "resources": {
                    "requests": {
                        "cpu": "100m",
                        "memory": "256Mi",
                    },
                },
            },
            "jobservice": {
                "resources": {
                    "requests": {
                        "cpu": "100m",
                        "memory": "256Mi",
                    },
                },
            },
            "registry": {
                "resources": {
                    "requests": {
                        "cpu": "100m",
                        "memory": "256Mi",
                    },
                },
            },
            "portal": {
                "resources": {
                    "requests": {
                        "cpu": "100m",
                        "memory": "128Mi",
                    },
                },
            },
            "database": {
                "internal": {
                    "resources": {
                        "requests": {
                            "cpu": "100m",
                            "memory": "256Mi",
                        },
                    },
                },
            },
            "redis": {
                "internal": {
                    "resources": {
                        "requests": {
                            "cpu": "100m",
                            "memory": "128Mi",
                        },
                    },
                },
            },
        },
        # Wait for Harbor to be fully deployed
        wait_for_jobs=True,
        timeout=600,
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=opts.depends_on if opts else None,
            parent=opts.parent if opts else None,
        ),
    )

    return HarborOutputs(
        release=release,
        url=harbor_url,
        admin_password=admin_pwd,
        namespace=namespace,
    )
