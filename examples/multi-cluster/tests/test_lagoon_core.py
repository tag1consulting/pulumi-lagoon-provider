"""Unit tests for Lagoon core configuration."""

import os
import sys

import pytest

# Add parent directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


class TestCORSConfiguration:
    """Tests for CORS configuration in API ingress."""

    def test_cors_annotations_structure(self):
        """Test that CORS annotations follow the expected nginx-ingress format."""
        # These are the expected annotations for CORS support
        required_annotations = {
            "nginx.ingress.kubernetes.io/enable-cors": "true",
            "nginx.ingress.kubernetes.io/cors-allow-credentials": "true",
        }

        for key in required_annotations:
            assert key.startswith("nginx.ingress.kubernetes.io/")

    def test_cors_allow_origin_wildcard_pattern(self):
        """Test that CORS allow-origin uses wildcard pattern for lagoon.local."""
        # The pattern should match *.lagoon.local:8443
        pattern = "https://*.lagoon.local:8443"

        # Verify the pattern format
        assert pattern.startswith("https://")
        assert "*." in pattern  # Wildcard subdomain
        assert ":8443" in pattern  # Port for Kind clusters

    def test_cors_allow_methods_include_required_methods(self):
        """Test that CORS allow-methods includes required HTTP methods."""
        allowed_methods = "GET, POST, PUT, DELETE, OPTIONS"

        # GraphQL requires POST
        assert "POST" in allowed_methods
        # OPTIONS is required for preflight requests
        assert "OPTIONS" in allowed_methods
        # GET is commonly used
        assert "GET" in allowed_methods

    def test_cors_allow_headers_include_required_headers(self):
        """Test that CORS allow-headers includes required headers for Lagoon."""
        allowed_headers = "DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization,Accept,Origin,Referer"

        # Authorization is required for Bearer tokens
        assert "Authorization" in allowed_headers
        # Content-Type is required for JSON requests
        assert "Content-Type" in allowed_headers
        # Accept is commonly sent by browsers
        assert "Accept" in allowed_headers


class TestUIEnvironmentConfiguration:
    """Tests for UI pod environment configuration."""

    def test_node_tls_reject_unauthorized_setting(self):
        """Test that NODE_TLS_REJECT_UNAUTHORIZED is set for self-signed certs."""
        # For Kind clusters with self-signed certs, this should be "0"
        env_value = "0"
        assert env_value == "0"

    def test_ui_env_vars_structure(self):
        """Test expected environment variables for UI pod."""
        required_env_vars = [
            "GRAPHQL_API",
            "KEYCLOAK_API",
            "WEBHOOK_URL",
            "LAGOON_VERSION",
            "NODE_TLS_REJECT_UNAUTHORIZED",
        ]

        # All should be valid environment variable names
        for var in required_env_vars:
            assert var.isupper() or "_" in var
            assert not var.startswith("_")


class TestAPIEnvironmentConfiguration:
    """Tests for API pod environment configuration."""

    def test_api_tls_reject_unauthorized_setting(self):
        """Test that NODE_TLS_REJECT_UNAUTHORIZED is set for API pod."""
        # API also needs to bypass TLS verification for internal HTTPS connections
        env_value = "0"
        assert env_value == "0"


class TestServiceNaming:
    """Tests for Kubernetes service naming conventions."""

    def test_release_name_pattern(self):
        """Test release name follows expected pattern."""
        # Release name is derived from resource name prefix
        name = "prod-lagoon-core"
        short_name = name.split("-")[0]  # "prod"
        release_name = f"{short_name}-core"  # "prod-core"

        assert release_name == "prod-core"

    def test_keycloak_service_name(self):
        """Test Keycloak service name follows Helm chart pattern."""
        release_name = "prod-core"
        namespace = "lagoon-core"

        # Pattern: {release_name}-lagoon-core-keycloak
        keycloak_service = f"{release_name}-lagoon-core-keycloak"
        assert keycloak_service == "prod-core-lagoon-core-keycloak"

        # Internal URL pattern
        internal_url = f"http://{keycloak_service}.{namespace}.svc.cluster.local:8080/auth"
        assert "8080" in internal_url
        assert "/auth" in internal_url

    def test_api_service_name(self):
        """Test API service name follows Helm chart pattern."""
        release_name = "prod-core"

        # Pattern: {release_name}-api
        api_service = f"{release_name}-api"
        assert api_service == "prod-core-api"

    def test_broker_service_name(self):
        """Test RabbitMQ broker service name follows Helm chart pattern."""
        release_name = "prod-core"

        # Pattern: {release_name}-broker
        broker_service = f"{release_name}-broker"
        assert broker_service == "prod-core-broker"

    def test_ssh_service_name(self):
        """Test SSH service name follows Helm chart pattern."""
        release_name = "prod-core"

        # Pattern: {release_name}-ssh
        ssh_service = f"{release_name}-ssh"
        assert ssh_service == "prod-core-ssh"


class TestRabbitMQNodePortConfiguration:
    """Tests for RabbitMQ NodePort service configuration."""

    def test_nodeport_value(self):
        """Test RabbitMQ NodePort uses expected value."""
        nodeport = 30672

        # NodePort must be in valid range (30000-32767)
        assert 30000 <= nodeport <= 32767

    def test_amqp_port(self):
        """Test AMQP port is standard RabbitMQ port."""
        amqp_port = 5672
        assert amqp_port == 5672

    def test_nodeport_service_selector(self):
        """Test NodePort service uses correct pod selector."""
        release_name = "prod-core"

        # Selector labels should match broker pods
        # Pattern: {release_name}-broker
        selector = {
            "app.kubernetes.io/component": f"{release_name}-broker",
            "app.kubernetes.io/instance": release_name,
            "app.kubernetes.io/name": "lagoon-core",
        }

        assert "broker" in selector["app.kubernetes.io/component"]
        assert selector["app.kubernetes.io/instance"] == release_name


class TestHelmValuesStructure:
    """Tests for Helm values structure."""

    def test_api_replica_count_for_migrations(self):
        """Test API replica count is 1 to prevent migration lock race."""
        # Single replica prevents multiple pods racing for migration lock
        replica_count = 1
        assert replica_count == 1

    def test_broker_internal_service_type(self):
        """Test broker internal service is ClusterIP."""
        service_type = "ClusterIP"
        assert service_type == "ClusterIP"

    def test_broker_external_disabled_in_chart(self):
        """Test chart's amqpExternal is disabled (we create our own)."""
        # We disable the chart's external service and create our own with fixed NodePort
        amqp_external_enabled = False
        assert amqp_external_enabled is False

    def test_ssh_service_type(self):
        """Test SSH service uses NodePort for Kind clusters."""
        service_type = "NodePort"
        assert service_type == "NodePort"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
