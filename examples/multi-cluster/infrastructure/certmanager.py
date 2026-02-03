"""Cert-manager installation for multi-cluster Lagoon example.

This module provides functions for installing cert-manager
and creating ClusterIssuers for TLS certificate management.
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s
from config import VERSIONS, CertManagerOutputs, NamespaceConfig


def install_cert_manager(
    name: str,
    provider: k8s.Provider,
    namespace_config: Optional[NamespaceConfig] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> k8s.helm.v3.Release:
    """Install cert-manager using Helm.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        namespace_config: Optional namespace configuration
        opts: Pulumi resource options

    Returns:
        Helm release resource
    """
    ns_config = namespace_config or NamespaceConfig()
    namespace = ns_config.cert_manager

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

    # Generate a short release name
    short_name = name.split("-")[0] if "-" in name else name
    release_name = f"{short_name}-cm"

    # Install cert-manager using Helm
    release = k8s.helm.v3.Release(
        f"{name}-release",
        name=release_name,
        chart="cert-manager",
        version=VERSIONS["cert_manager"],
        namespace=namespace,
        repository_opts=k8s.helm.v3.RepositoryOptsArgs(
            repo="https://charts.jetstack.io",
        ),
        values={
            "installCRDs": True,
            "resources": {
                "requests": {
                    "cpu": "10m",
                    "memory": "32Mi",
                },
            },
            "webhook": {
                "resources": {
                    "requests": {
                        "cpu": "10m",
                        "memory": "32Mi",
                    },
                },
            },
            "cainjector": {
                "resources": {
                    "requests": {
                        "cpu": "10m",
                        "memory": "32Mi",
                    },
                },
            },
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[ns],
            parent=opts.parent if opts else None,
        ),
    )

    return release


def create_cluster_issuer(
    name: str,
    provider: k8s.Provider,
    cert_manager_release: k8s.helm.v3.Release,
    issuer_name: str = "selfsigned-issuer",
    opts: Optional[pulumi.ResourceOptions] = None,
) -> CertManagerOutputs:
    """Create a self-signed ClusterIssuer.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        cert_manager_release: Cert-manager Helm release to depend on
        issuer_name: Name for the ClusterIssuer
        opts: Pulumi resource options

    Returns:
        CertManagerOutputs with issuer reference
    """
    # Create self-signed ClusterIssuer
    # Using apiextensions to create the custom resource
    issuer = k8s.apiextensions.CustomResource(
        f"{name}-issuer",
        api_version="cert-manager.io/v1",
        kind="ClusterIssuer",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=issuer_name,
        ),
        spec={
            "selfSigned": {},
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[cert_manager_release],
            parent=opts.parent if opts else None,
        ),
    )

    return CertManagerOutputs(
        issuer=issuer,
        namespace="cert-manager",
        issuer_name=issuer_name,
    )


def create_ca_issuer(
    name: str,
    provider: k8s.Provider,
    cert_manager_release: k8s.helm.v3.Release,
    namespace_config: Optional[NamespaceConfig] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> CertManagerOutputs:
    """Create a CA-based ClusterIssuer with self-signed root CA.

    This creates a two-tier PKI:
    1. Self-signed ClusterIssuer for root CA
    2. Root CA certificate
    3. CA ClusterIssuer using the root CA

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        cert_manager_release: Cert-manager Helm release to depend on
        namespace_config: Optional namespace configuration
        opts: Pulumi resource options

    Returns:
        CertManagerOutputs with the CA issuer
    """
    ns_config = namespace_config or NamespaceConfig()

    # Create self-signed issuer for root CA
    selfsigned_issuer = k8s.apiextensions.CustomResource(
        f"{name}-selfsigned-issuer",
        api_version="cert-manager.io/v1",
        kind="ClusterIssuer",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=f"{name}-selfsigned",
        ),
        spec={
            "selfSigned": {},
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[cert_manager_release],
            parent=opts.parent if opts else None,
        ),
    )

    # Create root CA certificate
    root_ca = k8s.apiextensions.CustomResource(
        f"{name}-root-ca",
        api_version="cert-manager.io/v1",
        kind="Certificate",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=f"{name}-root-ca",
            namespace=ns_config.cert_manager,
        ),
        spec={
            "isCA": True,
            "commonName": f"{name}-root-ca",
            "secretName": f"{name}-root-ca-secret",
            "duration": "87600h",  # 10 years
            "renewBefore": "8760h",  # 1 year
            "privateKey": {
                "algorithm": "ECDSA",
                "size": 256,
            },
            "issuerRef": {
                "name": f"{name}-selfsigned",
                "kind": "ClusterIssuer",
                "group": "cert-manager.io",
            },
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[selfsigned_issuer],
            parent=opts.parent if opts else None,
        ),
    )

    # Create CA ClusterIssuer
    ca_issuer = k8s.apiextensions.CustomResource(
        f"{name}-ca-issuer",
        api_version="cert-manager.io/v1",
        kind="ClusterIssuer",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=f"{name}-ca",
        ),
        spec={
            "ca": {
                "secretName": f"{name}-root-ca-secret",
            },
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[root_ca],
            parent=opts.parent if opts else None,
        ),
    )

    return CertManagerOutputs(
        issuer=ca_issuer,
        namespace=ns_config.cert_manager,
        issuer_name=f"{name}-ca",
    )
