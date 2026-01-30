"""Lagoon module for multi-cluster Lagoon example.

This module provides functions for installing and configuring
Lagoon core and remote (build-deploy) components.
"""

from .secrets import generate_lagoon_secrets
from .core import install_lagoon_core, create_rabbitmq_nodeport_service
from .remote import install_lagoon_remote
from .crds import install_lagoon_build_deploy_crds
from .keycloak import configure_keycloak_for_cli_auth
from .project import (
    create_deploy_targets,
    create_example_drupal_project,
    DeployTargetPair,
    ExampleProjectOutputs,
)
from .migrations import (
    ensure_knex_migrations,
    check_knex_migrations_inline,
)

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
