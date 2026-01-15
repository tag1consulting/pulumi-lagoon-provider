"""Unit tests for Lagoon provider configuration."""

import pytest
from unittest.mock import Mock, patch
import os


class TestLagoonConfigEnvVars:
    """Tests for LagoonConfig with environment variables."""

    def test_config_from_env_vars(self, env_vars, mock_pulumi_config):
        """Test configuration loads from environment variables."""
        from pulumi_lagoon.config import LagoonConfig

        config = LagoonConfig()

        assert config.api_url == "https://api.test.lagoon.sh/graphql"
        assert config.token == "test-token-from-env"

    def test_config_missing_token_raises_error(self, clean_env, mock_pulumi_config):
        """Test that missing token raises ValueError."""
        from pulumi_lagoon.config import LagoonConfig

        with pytest.raises(ValueError, match="Lagoon API token must be provided"):
            LagoonConfig()

    def test_config_default_api_url(self, mock_pulumi_config):
        """Test default API URL when not specified."""
        with patch.dict(os.environ, {"LAGOON_TOKEN": "test-token"}, clear=False):
            # Remove LAGOON_API_URL if present
            env = os.environ.copy()
            if "LAGOON_API_URL" in env:
                del os.environ["LAGOON_API_URL"]

            from pulumi_lagoon.config import LagoonConfig

            config = LagoonConfig()

            assert config.api_url == "https://api.lagoon.sh/graphql"

    def test_config_ssh_key_path_optional(self, env_vars, mock_pulumi_config):
        """Test SSH key path is optional."""
        from pulumi_lagoon.config import LagoonConfig

        config = LagoonConfig()

        assert config.ssh_key_path is None

    def test_config_ssh_key_path_from_env(self, mock_pulumi_config):
        """Test SSH key path from environment variable."""
        with patch.dict(
            os.environ,
            {"LAGOON_TOKEN": "test-token", "LAGOON_SSH_KEY_PATH": "/path/to/key"},
        ):
            from pulumi_lagoon.config import LagoonConfig

            config = LagoonConfig()

            assert config.ssh_key_path == "/path/to/key"


class TestLagoonConfigPulumiConfig:
    """Tests for LagoonConfig with Pulumi config values."""

    def test_config_from_pulumi_config(self, clean_env):
        """Test configuration loads from Pulumi config."""
        with patch("pulumi.Config") as mock_config_class:
            mock_config = Mock()
            mock_config_class.return_value = mock_config

            # Simulate Pulumi config values
            mock_config.get.side_effect = lambda key: {
                "apiUrl": "https://api.pulumi-config.lagoon.sh/graphql",
                "sshKeyPath": "/pulumi/key/path",
            }.get(key)
            mock_config.get_secret.return_value = "pulumi-secret-token"

            from pulumi_lagoon.config import LagoonConfig

            config = LagoonConfig()

            assert config.api_url == "https://api.pulumi-config.lagoon.sh/graphql"
            assert config.token == "pulumi-secret-token"
            assert config.ssh_key_path == "/pulumi/key/path"

    def test_config_pulumi_takes_precedence(self):
        """Test Pulumi config takes precedence over environment variables."""
        with patch.dict(
            os.environ,
            {
                "LAGOON_API_URL": "https://api.env.lagoon.sh/graphql",
                "LAGOON_TOKEN": "env-token",
            },
        ):
            with patch("pulumi.Config") as mock_config_class:
                mock_config = Mock()
                mock_config_class.return_value = mock_config

                # Pulumi config returns values
                mock_config.get.side_effect = lambda key: {
                    "apiUrl": "https://api.pulumi.lagoon.sh/graphql",
                }.get(key)
                mock_config.get_secret.return_value = "pulumi-token"

                from pulumi_lagoon.config import LagoonConfig

                config = LagoonConfig()

                # Pulumi values should take precedence
                assert config.api_url == "https://api.pulumi.lagoon.sh/graphql"
                assert config.token == "pulumi-token"


class TestLagoonConfigGetClient:
    """Tests for the get_client factory method."""

    def test_get_client_returns_lagoon_client(self, env_vars, mock_pulumi_config):
        """Test get_client returns a configured LagoonClient."""
        from pulumi_lagoon.config import LagoonConfig

        with patch("pulumi_lagoon.client.LagoonClient") as mock_client_class:
            mock_client = Mock()
            mock_client_class.return_value = mock_client

            config = LagoonConfig()
            client = config.get_client()

            mock_client_class.assert_called_once_with(
                "https://api.test.lagoon.sh/graphql", "test-token-from-env"
            )
            assert client == mock_client


class TestLagoonConfigRepr:
    """Tests for string representation."""

    def test_repr_hides_token(self, env_vars, mock_pulumi_config):
        """Test __repr__ does not expose the token."""
        from pulumi_lagoon.config import LagoonConfig

        config = LagoonConfig()
        repr_str = repr(config)

        assert "test-token-from-env" not in repr_str
        assert "***" in repr_str
        assert "api.test.lagoon.sh" in repr_str
