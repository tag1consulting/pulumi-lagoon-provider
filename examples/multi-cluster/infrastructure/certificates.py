"""TLS certificate generation for multi-cluster Lagoon example.

This module provides functions for creating TLS certificates
using cert-manager.
"""

from typing import List, Optional

import pulumi
import pulumi_kubernetes as k8s
from config import CertManagerOutputs


def create_wildcard_certificate(
    name: str,
    domain: str,
    namespace: str,
    issuer: CertManagerOutputs,
    provider: k8s.Provider,
    additional_dns_names: Optional[List[str]] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> k8s.apiextensions.CustomResource:
    """Create a wildcard TLS certificate for a domain.

    Args:
        name: Pulumi resource name prefix
        domain: Base domain for the wildcard certificate
        namespace: Namespace to create the certificate secret in
        issuer: ClusterIssuer outputs to use for signing
        provider: Kubernetes provider for the target cluster
        additional_dns_names: Additional DNS names to include
        opts: Pulumi resource options

    Returns:
        Certificate custom resource
    """
    dns_names = [
        domain,
        f"*.{domain}",
    ]
    if additional_dns_names:
        dns_names.extend(additional_dns_names)

    certificate = k8s.apiextensions.CustomResource(
        f"{name}-certificate",
        api_version="cert-manager.io/v1",
        kind="Certificate",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=f"{name}-tls",
            namespace=namespace,
        ),
        spec={
            "secretName": f"{name}-tls",
            "duration": "8760h",  # 1 year
            "renewBefore": "720h",  # 30 days
            "privateKey": {
                "algorithm": "RSA",
                "size": 2048,
            },
            "dnsNames": dns_names,
            "issuerRef": {
                "name": issuer.issuer_name,
                "kind": "ClusterIssuer",
                "group": "cert-manager.io",
            },
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[issuer.issuer],
            parent=opts.parent if opts else None,
        ),
    )

    return certificate


def create_certificate(
    name: str,
    dns_names: List[str],
    namespace: str,
    issuer: CertManagerOutputs,
    provider: k8s.Provider,
    secret_name: Optional[str] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> k8s.apiextensions.CustomResource:
    """Create a TLS certificate for specific DNS names.

    Args:
        name: Pulumi resource name prefix
        dns_names: List of DNS names for the certificate
        namespace: Namespace to create the certificate secret in
        issuer: ClusterIssuer outputs to use for signing
        provider: Kubernetes provider for the target cluster
        secret_name: Optional name for the TLS secret (defaults to {name}-tls)
        opts: Pulumi resource options

    Returns:
        Certificate custom resource
    """
    cert_secret_name = secret_name or f"{name}-tls"

    certificate = k8s.apiextensions.CustomResource(
        f"{name}-certificate",
        api_version="cert-manager.io/v1",
        kind="Certificate",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=cert_secret_name,
            namespace=namespace,
        ),
        spec={
            "secretName": cert_secret_name,
            "duration": "8760h",  # 1 year
            "renewBefore": "720h",  # 30 days
            "privateKey": {
                "algorithm": "RSA",
                "size": 2048,
            },
            "dnsNames": dns_names,
            "issuerRef": {
                "name": issuer.issuer_name,
                "kind": "ClusterIssuer",
                "group": "cert-manager.io",
            },
        },
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[issuer.issuer],
            parent=opts.parent if opts else None,
        ),
    )

    return certificate


def get_certificate_secret_name(cert_name: str) -> str:
    """Get the TLS secret name for a certificate.

    Args:
        cert_name: Name used when creating the certificate

    Returns:
        Name of the TLS secret created by cert-manager
    """
    return f"{cert_name}-tls"
