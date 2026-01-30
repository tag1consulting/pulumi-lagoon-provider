"""Unit tests for multi-cluster configuration."""

import pytest
import sys
import os

# Add parent directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from config import DomainConfig, NamespaceConfig, ClusterConfig, DEFAULT_CLUSTERS


class TestDomainConfig:
    """Tests for DomainConfig dataclass."""

    def test_default_values(self):
        """Test default configuration values."""
        config = DomainConfig()
        assert config.base == "lagoon.local"
        assert config.https_port == 8443

    def test_custom_base_domain(self):
        """Test custom base domain configuration."""
        config = DomainConfig(base="mylagoon.example.com")
        assert config.base == "mylagoon.example.com"
        assert config.lagoon_api == "api.mylagoon.example.com"

    def test_port_suffix_for_kind_cluster(self):
        """Test port suffix is added for non-standard ports (Kind uses 8443)."""
        config = DomainConfig(https_port=8443)
        assert config._port_suffix == ":8443"

    def test_port_suffix_for_standard_port(self):
        """Test port suffix is empty for standard HTTPS port 443."""
        config = DomainConfig(https_port=443)
        assert config._port_suffix == ""

    def test_lagoon_api_url_with_port(self):
        """Test API URL includes port suffix for Kind clusters."""
        config = DomainConfig(https_port=8443)
        assert config.lagoon_api_url == "https://api.lagoon.local:8443"

    def test_lagoon_api_url_without_port(self):
        """Test API URL has no port suffix for standard HTTPS."""
        config = DomainConfig(https_port=443)
        assert config.lagoon_api_url == "https://api.lagoon.local"

    def test_lagoon_ui_url_with_port(self):
        """Test UI URL includes port suffix for Kind clusters."""
        config = DomainConfig(https_port=8443)
        assert config.lagoon_ui_url == "https://ui.lagoon.local:8443"

    def test_lagoon_ui_url_without_port(self):
        """Test UI URL has no port suffix for standard HTTPS."""
        config = DomainConfig(https_port=443)
        assert config.lagoon_ui_url == "https://ui.lagoon.local"

    def test_lagoon_keycloak_url_with_port(self):
        """Test Keycloak URL includes port suffix for Kind clusters."""
        config = DomainConfig(https_port=8443)
        assert config.lagoon_keycloak_url == "https://keycloak.lagoon.local:8443"

    def test_lagoon_keycloak_url_without_port(self):
        """Test Keycloak URL has no port suffix for standard HTTPS."""
        config = DomainConfig(https_port=443)
        assert config.lagoon_keycloak_url == "https://keycloak.lagoon.local"

    def test_all_domain_properties(self):
        """Test all domain properties return expected values."""
        config = DomainConfig(base="test.local")
        assert config.lagoon_api == "api.test.local"
        assert config.lagoon_ui == "ui.test.local"
        assert config.lagoon_keycloak == "keycloak.test.local"
        assert config.lagoon_webhook == "webhook.test.local"
        assert config.lagoon_ssh == "ssh.test.local"
        assert config.harbor == "harbor.test.local"
        assert config.harbor_notary == "notary.test.local"

    def test_custom_port_in_urls(self):
        """Test custom port (e.g., 9443) is included in URLs."""
        config = DomainConfig(https_port=9443)
        assert config._port_suffix == ":9443"
        assert config.lagoon_api_url == "https://api.lagoon.local:9443"
        assert config.lagoon_ui_url == "https://ui.lagoon.local:9443"
        assert config.lagoon_keycloak_url == "https://keycloak.lagoon.local:9443"


class TestNamespaceConfig:
    """Tests for NamespaceConfig dataclass."""

    def test_default_namespaces(self):
        """Test default namespace values."""
        config = NamespaceConfig()
        assert config.ingress == "ingress-nginx"
        assert config.cert_manager == "cert-manager"
        assert config.harbor == "harbor"
        assert config.lagoon_core == "lagoon-core"
        assert config.lagoon_remote == "lagoon"


class TestClusterConfig:
    """Tests for ClusterConfig dataclass."""

    def test_context_name_generation(self):
        """Test Kubernetes context name is generated correctly."""
        config = ClusterConfig(name="my-cluster", http_port=8080, https_port=8443)
        assert config.context_name == "kind-my-cluster"

    def test_production_cluster_config(self):
        """Test production cluster configuration."""
        config = DEFAULT_CLUSTERS["prod"]
        assert config.name == "lagoon-prod"
        assert config.http_port == 8080
        assert config.https_port == 8443
        assert config.is_production is True
        assert "lagoon.sh/environment-type" in config.node_labels
        assert config.node_labels["lagoon.sh/environment-type"] == "production"

    def test_nonprod_cluster_config(self):
        """Test non-production cluster configuration."""
        config = DEFAULT_CLUSTERS["nonprod"]
        assert config.name == "lagoon-nonprod"
        assert config.http_port == 9080
        assert config.https_port == 9443
        assert config.is_production is False
        assert config.node_labels["lagoon.sh/environment-type"] == "development"

    def test_ingress_ready_label(self):
        """Test both clusters have ingress-ready label for nginx DaemonSet."""
        for key in ["prod", "nonprod"]:
            config = DEFAULT_CLUSTERS[key]
            assert "ingress-ready" in config.node_labels
            assert config.node_labels["ingress-ready"] == "true"


class TestURLConsistency:
    """Tests for URL consistency across different configurations."""

    def test_url_protocol_consistency(self):
        """Test all URLs use HTTPS protocol."""
        config = DomainConfig()
        assert config.lagoon_api_url.startswith("https://")
        assert config.lagoon_ui_url.startswith("https://")
        assert config.lagoon_keycloak_url.startswith("https://")

    def test_url_domain_consistency(self):
        """Test URLs use correct domain from base."""
        config = DomainConfig(base="custom.domain")
        assert "custom.domain" in config.lagoon_api_url
        assert "custom.domain" in config.lagoon_ui_url
        assert "custom.domain" in config.lagoon_keycloak_url

    def test_api_url_for_graphql_endpoint(self):
        """Test API URL can be used to construct GraphQL endpoint."""
        config = DomainConfig(https_port=8443)
        graphql_url = f"{config.lagoon_api_url}/graphql"
        assert graphql_url == "https://api.lagoon.local:8443/graphql"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
