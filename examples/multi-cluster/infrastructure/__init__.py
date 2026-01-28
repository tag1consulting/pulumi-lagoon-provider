"""Infrastructure module for multi-cluster Lagoon example.

This module provides functions for installing common infrastructure components:
- ingress-nginx
- cert-manager
- TLS certificates
- CoreDNS configuration for local domain resolution
"""

from .ingress import install_ingress_nginx
from .certmanager import install_cert_manager, create_cluster_issuer
from .certificates import create_wildcard_certificate
from .coredns import setup_coredns_for_lagoon, get_kind_node_internal_ip, patch_coredns_for_lagoon

__all__ = [
    "install_ingress_nginx",
    "install_cert_manager",
    "create_cluster_issuer",
    "create_wildcard_certificate",
    "setup_coredns_for_lagoon",
    "get_kind_node_internal_ip",
    "patch_coredns_for_lagoon",
]
