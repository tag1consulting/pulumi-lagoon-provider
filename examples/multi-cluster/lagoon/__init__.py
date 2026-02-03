"""Lagoon module for multi-cluster Lagoon example.

This module provides functions for installing and configuring
Lagoon core and remote (build-deploy) components.
"""

from .core import create_rabbitmq_nodeport_service, install_lagoon_core
from .crds import install_lagoon_build_deploy_crds
from .keycloak import configure_keycloak_for_cli_auth
from .migrations import (
    check_knex_migrations_inline,
    ensure_knex_migrations,
)
from .project import (
    DeployTargetPair,
    ExampleProjectOutputs,
    create_deploy_targets,
    create_example_drupal_project,
)
from .remote import install_lagoon_remote
from .secrets import generate_lagoon_secrets

__all__ = [
    "generate_lagoon_secrets",
    "install_lagoon_core",
    "install_lagoon_remote",
    "install_lagoon_build_deploy_crds",
    "create_rabbitmq_nodeport_service",
    "configure_keycloak_for_cli_auth",
    "create_deploy_targets",
    "create_example_drupal_project",
    "DeployTargetPair",
    "ExampleProjectOutputs",
    "ensure_knex_migrations",
    "check_knex_migrations_inline",
]
