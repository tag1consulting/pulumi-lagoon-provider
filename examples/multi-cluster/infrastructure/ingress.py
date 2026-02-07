"""Ingress-nginx installation for multi-cluster Lagoon example.

This module provides functions for installing and configuring
ingress-nginx on Kind clusters.
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s
from config import VERSIONS, IngressOutputs, NamespaceConfig


def install_ingress_nginx(
    name: str,
    provider: k8s.Provider,
    namespace_config: Optional[NamespaceConfig] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> IngressOutputs:
    """Install ingress-nginx using Helm.

    Configures ingress-nginx for Kind clusters with hostPort mode
    to allow access via the Kind node ports.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        namespace_config: Optional namespace configuration
        opts: Pulumi resource options

    Returns:
        IngressOutputs with service reference and namespace
    """
    ns_config = namespace_config or NamespaceConfig()
    namespace = ns_config.ingress

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

    # Generate a short release name to avoid Kubernetes 63-char limit
    # Extract a short prefix from the name (e.g., "prod" from "prod-ingress")
    short_name = name.split("-")[0] if "-" in name else name
    release_name = f"{short_name}-nginx"

    # Install ingress-nginx using Helm
    # Configuration optimized for Kind clusters
    release = k8s.helm.v3.Release(
        f"{name}-release",
        name=release_name,  # Explicit short name to avoid 63-char limit
        chart="ingress-nginx",
        version=VERSIONS["ingress_nginx"],
        namespace=namespace,
        repository_opts=k8s.helm.v3.RepositoryOptsArgs(
            repo="https://kubernetes.github.io/ingress-nginx",
        ),
        values={
            "controller": {
                "hostPort": {
                    "enabled": True,
                },
                "service": {
                    "type": "NodePort",
                },
                "watchIngressWithoutClass": True,
                "ingressClassResource": {
                    "name": "nginx",
                    "enabled": True,
                    "default": True,
                },
                "config": {
                    "proxy-buffer-size": "16k",
                    "proxy-buffers": "4 16k",
                    "use-forwarded-headers": "true",
                },
                "resources": {
                    "requests": {
                        "cpu": "100m",
                        "memory": "90Mi",
                    },
                },
                # Kind-specific: use hostNetwork for better compatibility
                "hostNetwork": True,
                "kind": "DaemonSet",
                "nodeSelector": {
                    "ingress-ready": "true",
                    "kubernetes.io/os": "linux",
                },
                "tolerations": [
                    {
                        "key": "node-role.kubernetes.io/control-plane",
                        "operator": "Equal",
                        "effect": "NoSchedule",
                    },
                    {
                        "key": "node-role.kubernetes.io/master",
                        "operator": "Equal",
                        "effect": "NoSchedule",
                    },
                ],
                # Disable admission webhook for local development (avoids timing issues)
                "admissionWebhooks": {
                    "enabled": False,
                },
            },
            "defaultBackend": {
                "enabled": False,
            },
        },
        # Wait for the release to be fully deployed before returning
        wait_for_jobs=True,
        timeout=300,
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[ns] + (opts.depends_on if opts and opts.depends_on else []),
            parent=opts.parent if opts else None,
        ),
    )

    return IngressOutputs(
        service=release,
        namespace=namespace,
        class_name="nginx",
    )
