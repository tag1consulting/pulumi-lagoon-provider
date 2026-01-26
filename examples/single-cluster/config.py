"""Centralized configuration for the single-cluster Lagoon example.

This module provides configuration for a simple single-cluster Lagoon deployment.
It reuses components from the multi-cluster example but configures them for
a single Kind cluster.
"""

from dataclasses import dataclass, field
from typing import Optional

import pulumi
from pulumi import Config

# =============================================================================
# Version Constants (same as multi-cluster)
# =============================================================================

VERSIONS = {
    "ingress_nginx": "4.10.1",
    "cert_manager": "v1.14.4",
    "harbor": "1.14.2",
    "lagoon_core": "1.59.0",
    "lagoon_remote": "0.103.0",
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


# Single cluster configuration
DEFAULT_CLUSTER = ClusterConfig(
    name="lagoon",
    http_port=8080,
    https_port=8443,
    is_production=False,
    node_labels={
        "lagoon.sh/environment-type": "development",
        "ingress-ready": "true",
    },
)


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
    rabbitmq_host: str
    rabbitmq_password: pulumi.Output[str]
    namespace: str
    ssh_host: str
    rabbitmq_nodeport: int = 30672


@dataclass
class LagoonRemoteOutputs:
    """Outputs from Lagoon remote/build-deploy installation."""

    release: pulumi.Resource
    namespace: str


# =============================================================================
# Pulumi Config Accessors
# =============================================================================

class SingleClusterConfig:
    """Pulumi configuration accessor for single-cluster example."""

    def __init__(self):
        self._config = Config()

    @property
    def create_cluster(self) -> bool:
        """Whether to create the Kind cluster (default: True)."""
        val = self._config.get_bool("createCluster")
        return val if val is not None else True

    @property
    def base_domain(self) -> str:
        """Base domain for services (default: lagoon.local)."""
        return self._config.get("baseDomain") or DEFAULT_BASE_DOMAIN

    @property
    def cluster_name(self) -> str:
        """Kind cluster name (default: lagoon)."""
        return self._config.get("clusterName") or "lagoon"

    @property
    def http_port(self) -> int:
        """HTTP port for ingress (default: 8080)."""
        val = self._config.get_int("httpPort")
        return val if val is not None else 8080

    @property
    def https_port(self) -> int:
        """HTTPS port for ingress (default: 8443)."""
        val = self._config.get_int("httpsPort")
        return val if val is not None else 8443

    @property
    def harbor_admin_password(self) -> Optional[pulumi.Output[str]]:
        """Harbor admin password (optional, generated if not provided)."""
        return self._config.get_secret("harborAdminPassword")

    @property
    def install_harbor(self) -> bool:
        """Whether to install Harbor registry (default: True)."""
        val = self._config.get_bool("installHarbor")
        return val if val is not None else True

    @property
    def install_lagoon(self) -> bool:
        """Whether to install Lagoon (default: True)."""
        val = self._config.get_bool("installLagoon")
        return val if val is not None else True

    @property
    def helm_timeout(self) -> int:
        """Helm release timeout in seconds (default: 1800 = 30 minutes)."""
        val = self._config.get_int("helmTimeout")
        return val if val is not None else 1800

    @property
    def deploy_target_name(self) -> str:
        """Name for the deploy target (default: local-kind)."""
        return self._config.get("deployTargetName") or "local-kind"

    @property
    def create_example_project(self) -> bool:
        """Whether to create the example Drupal project (default: False)."""
        val = self._config.get_bool("createExampleProject")
        return val if val is not None else False

    @property
    def example_project_name(self) -> str:
        """Name for the example project (default: lagoon-drupal-example)."""
        return self._config.get("exampleProjectName") or "lagoon-drupal-example"

    @property
    def example_project_git_url(self) -> str:
        """Git URL for the example project."""
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

    def get_cluster_config(self) -> ClusterConfig:
        """Get cluster configuration."""
        return ClusterConfig(
            name=self.cluster_name,
            http_port=self.http_port,
            https_port=self.https_port,
            is_production=False,
            node_labels={
                "lagoon.sh/environment-type": "development",
                "ingress-ready": "true",
            },
        )


# Global config instance
config = SingleClusterConfig()
