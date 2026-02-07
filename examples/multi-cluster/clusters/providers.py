"""Kubernetes provider setup for multi-cluster Lagoon example.

This module creates Kubernetes providers for each cluster,
enabling resources to be deployed to the correct cluster.
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s
from config import ClusterOutputs


def create_k8s_provider(
    name: str,
    cluster: ClusterOutputs,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> k8s.Provider:
    """Create a Kubernetes provider for a cluster.

    Args:
        name: Pulumi resource name for the provider
        cluster: Cluster outputs containing kubeconfig
        opts: Pulumi resource options

    Returns:
        Kubernetes Provider instance
    """
    return k8s.Provider(
        name,
        kubeconfig=cluster.kubeconfig,
        context=cluster.context_name,
        opts=pulumi.ResourceOptions(
            depends_on=[cluster.cluster_resource],
            parent=opts.parent if opts else None,
        ),
    )


def create_k8s_provider_from_kubeconfig(
    name: str,
    kubeconfig: pulumi.Output[str],
    context: Optional[str] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> k8s.Provider:
    """Create a Kubernetes provider from a kubeconfig string.

    Args:
        name: Pulumi resource name for the provider
        kubeconfig: Kubeconfig YAML string
        context: Optional context name to use
        opts: Pulumi resource options

    Returns:
        Kubernetes Provider instance
    """
    return k8s.Provider(
        name,
        kubeconfig=kubeconfig,
        context=context,
        opts=opts,
    )
