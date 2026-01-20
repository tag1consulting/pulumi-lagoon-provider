"""Lagoon module for multi-cluster Lagoon example.

This module provides functions for installing and configuring
Lagoon core and remote (build-deploy) components.
"""

from .secrets import generate_lagoon_secrets
from .core import install_lagoon_core, create_rabbitmq_nodeport_service
from .remote import install_lagoon_remote
from .keycloak import configure_keycloak_for_cli_auth

__all__ = [
    "generate_lagoon_secrets",
    "install_lagoon_core",
    "install_lagoon_remote",
    "create_rabbitmq_nodeport_service",
    "configure_keycloak_for_cli_auth",
]
