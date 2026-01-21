"""Centralized configuration for the multi-cluster Lagoon example.

This module provides:
- Domain names and version constants
- Namespace names and default configurations
- Pulumi config accessors
- Dataclasses for type-safe configuration passing between modules
"""

from dataclasses import dataclass, field
from typing import Optional

import pulumi
from pulumi import Config

# =============================================================================
# Version Constants
# =============================================================================

VERSIONS = {
    "ingress_nginx": "4.10.1",
    "cert_manager": "v1.14.4",
    "harbor": "1.14.2",
    "lagoon_core": "1.0.0",
    "lagoon_build_deploy": "0.39.0",  # Updated for k8s 1.22+ CRD compatibility
}

# Kind node image - use a version compatible with WSL2/cgroup v2
KIND_NODE_IMAGE = "kindest/node:v1.31.0"

# =============================================================================
# Domain Configuration
# =============================================================================

DEFAULT_BASE_DOMAIN = "lagoon.local"


@dataclass
class DomainConfig:
    """Domain configuration for all services."""

    base: str = DEFAULT_BASE_DOMAIN

    @property
    def lagoon_api(self) -> str:
        return f"api.{self.base}"

    @property
    def lagoon_ui(self) -> str:
        return f"ui.{self.base}"

    @property
    def lagoon_keycloak(self) -> str:
        return f"keycloak.{self.base}"

    @property
    def lagoon_webhook(self) -> str:
        return f"webhook.{self.base}"

    @property
    def lagoon_ssh(self) -> str:
        return f"ssh.{self.base}"

    @property
    def harbor(self) -> str:
        return f"harbor.{self.base}"

    @property
    def harbor_notary(self) -> str:
        return f"notary.{self.base}"


# =============================================================================
# Namespace Configuration
# =============================================================================

@dataclass
class NamespaceConfig:
    """Kubernetes namespace configuration."""

    ingress: str = "ingress-nginx"
    cert_manager: str = "cert-manager"
    harbor: str = "harbor"
    lagoon_core: str = "lagoon-core"
    lagoon_remote: str = "lagoon"


# =============================================================================
# Cluster Configuration
# =============================================================================

@dataclass
class ClusterConfig:
    """Configuration for a Kind cluster."""

    name: str
    http_port: int
    https_port: int
    is_production: bool = False
    node_labels: dict = field(default_factory=dict)

    @property
    def context_name(self) -> str:
        return f"kind-{self.name}"


DEFAULT_CLUSTERS = {
    "prod": ClusterConfig(
        name="lagoon-prod",
        http_port=8080,
        https_port=8443,
        is_production=True,
        node_labels={
            "lagoon.sh/environment-type": "production",
            "ingress-ready": "true",  # Required for ingress-nginx DaemonSet
        },
    ),
    "nonprod": ClusterConfig(
        name="lagoon-nonprod",
        http_port=9080,
        https_port=9443,
        is_production=False,
        node_labels={
            "lagoon.sh/environment-type": "development",
            "ingress-ready": "true",  # Required for ingress-nginx DaemonSet
        },
    ),
}


# =============================================================================
# Component Output Dataclasses
# =============================================================================

@dataclass
class ClusterOutputs:
    """Outputs from cluster creation."""

    name: str
    kubeconfig: pulumi.Output[str]
    context_name: str
    cluster_resource: pulumi.Resource


@dataclass
class IngressOutputs:
    """Outputs from ingress-nginx installation."""

    service: pulumi.Resource
    namespace: str
    class_name: str = "nginx"


@dataclass
class CertManagerOutputs:
    """Outputs from cert-manager installation."""

    issuer: pulumi.Resource
    namespace: str
    issuer_name: str


@dataclass
class HarborOutputs:
    """Outputs from Harbor installation."""

    release: pulumi.Resource
    url: str
    admin_password: pulumi.Output[str]
    namespace: str


@dataclass
class LagoonSecretsOutputs:
    """Outputs from Lagoon secrets generation."""

    ssh_private_key: pulumi.Output[str]
    ssh_public_key: pulumi.Output[str]
    rabbitmq_password: pulumi.Output[str]
    keycloak_admin_password: pulumi.Output[str]
    api_db_password: pulumi.Output[str]


@dataclass
class LagoonCoreOutputs:
    """Outputs from Lagoon core installation."""

    release: pulumi.Resource
    api_url: str
    ui_url: str
    keycloak_url: str
    rabbitmq_host: str  # Internal Kubernetes service name
    rabbitmq_password: pulumi.Output[str]
    namespace: str
    ssh_host: str  # Internal Kubernetes service name for SSH
    # External connection info for cross-cluster communication
    rabbitmq_nodeport: int = 30672  # NodePort for RabbitMQ


@dataclass
class LagoonRemoteOutputs:
    """Outputs from Lagoon remote/build-deploy installation."""

    release: pulumi.Resource
    namespace: str


# =============================================================================
# Pulumi Config Accessors
# =============================================================================

class MultiClusterConfig:
    """Pulumi configuration accessor for multi-cluster example."""

    def __init__(self):
        self._config = Config()
        self._lagoon_config = Config("lagoon")

    @property
    def create_clusters(self) -> bool:
        """Whether to create Kind clusters (default: True)."""
        val = self._config.get_bool("createClusters")
        return val if val is not None else True

    @property
    def base_domain(self) -> str:
        """Base domain for services (default: lagoon.local)."""
        return self._config.get("baseDomain") or DEFAULT_BASE_DOMAIN

    @property
    def harbor_admin_password(self) -> Optional[pulumi.Output[str]]:
        """Harbor admin password (optional, generated if not provided)."""
        return self._config.get_secret("harborAdminPassword")

    @property
    def lagoon_api_url(self) -> Optional[str]:
        """Lagoon API URL for provider configuration."""
        return self._lagoon_config.get("apiUrl")

    @property
    def lagoon_token(self) -> Optional[pulumi.Output[str]]:
        """Lagoon API token for provider configuration."""
        return self._lagoon_config.get_secret("token")

    @property
    def ssh_host(self) -> Optional[str]:
        """SSH host for Lagoon builds."""
        return self._config.get("sshHost")

    @property
    def install_harbor(self) -> bool:
        """Whether to install Harbor registry (default: True)."""
        val = self._config.get_bool("installHarbor")
        return val if val is not None else True

    @property
    def install_lagoon(self) -> bool:
        """Whether to install Lagoon core (default: True)."""
        val = self._config.get_bool("installLagoon")
        return val if val is not None else True

    @property
    def helm_timeout(self) -> int:
        """Helm release timeout in seconds (default: 1800 = 30 minutes).

        Lagoon core takes a long time to initialize. If you experience timeouts,
        you can increase this value:
            pulumi config set helmTimeout 3600  # 1 hour
        """
        val = self._config.get_int("helmTimeout")
        return val if val is not None else 1800  # Default 30 minutes

    @property
    def create_example_project(self) -> bool:
        """Whether to create the example Drupal project (default: True).

        Set to false to skip example project creation:
            pulumi config set createExampleProject false
        """
        val = self._config.get_bool("createExampleProject")
        return val if val is not None else True

    @property
    def example_project_name(self) -> str:
        """Name for the example Drupal project (default: drupal-example).

        Set a custom name:
            pulumi config set exampleProjectName my-drupal-site
        """
        return self._config.get("exampleProjectName") or "drupal-example"

    @property
    def example_project_git_url(self) -> str:
        """Git URL for the example project (default: Lagoon's Drupal base).

        Set a custom Git URL:
            pulumi config set exampleProjectGitUrl https://github.com/myorg/myrepo.git
        """
        return (
            self._config.get("exampleProjectGitUrl")
            or "https://github.com/lagoon-examples/drupal-base.git"
        )

    def get_domain_config(self) -> DomainConfig:
        """Get domain configuration based on Pulumi config."""
        return DomainConfig(base=self.base_domain)

    def get_namespace_config(self) -> NamespaceConfig:
        """Get namespace configuration."""
        return NamespaceConfig()

    def get_cluster_config(self, cluster_key: str) -> ClusterConfig:
        """Get cluster configuration by key (prod/nonprod)."""
        return DEFAULT_CLUSTERS.get(cluster_key, DEFAULT_CLUSTERS["prod"])


# Global config instance
config = MultiClusterConfig()
