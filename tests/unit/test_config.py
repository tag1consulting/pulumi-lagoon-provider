"""Unit tests for Lagoon provider configuration."""

import os
from unittest.mock import Mock, patch

import pytest


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

            # Simulate Pulumi config values (token is now retrieved via get(), not get_secret())
            mock_config.get.side_effect = lambda key: {
                "apiUrl": "https://api.pulumi-config.lagoon.sh/graphql",
                "token": "pulumi-secret-token",
                "sshKeyPath": "/pulumi/key/path",
            }.get(key)

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

                # Pulumi config returns values (token via get(), not get_secret())
                mock_config.get.side_effect = lambda key: {
                    "apiUrl": "https://api.pulumi.lagoon.sh/graphql",
                    "token": "pulumi-token",
                }.get(key)

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


class TestLagoonConfigJwtSecret:
    """Tests for JWT token generation from secret."""

    def test_config_generates_token_from_jwt_secret(self, clean_env):
        """Test token is generated from JWT secret when no direct token provided."""
        with patch("pulumi.Config") as mock_config_class:
            mock_config = Mock()
            mock_config_class.return_value = mock_config

            # Only jwtSecret is provided, not token
            mock_config.get.side_effect = lambda key: {
                "jwtSecret": "test-jwt-secret-key",
            }.get(key)

            # Mock the jwt module
            with patch("pulumi_lagoon.config.LagoonConfig._generate_admin_token") as mock_gen:
                mock_gen.return_value = "generated-jwt-token"

                from pulumi_lagoon.config import LagoonConfig

                config = LagoonConfig()

                mock_gen.assert_called_once_with("test-jwt-secret-key")
                assert config.token == "generated-jwt-token"

    def test_generate_admin_token_success(self, clean_env, mock_pulumi_config):
        """Test _generate_admin_token creates valid JWT."""
        # First set up with a token so we can create the config
        with patch.dict(os.environ, {"LAGOON_TOKEN": "temp-token"}):
            from pulumi_lagoon.config import LagoonConfig

            config = LagoonConfig()

        # Now test the _generate_admin_token method directly
        with patch("jwt.encode") as mock_encode:
            mock_encode.return_value = "test-generated-token"

            result = config._generate_admin_token("test-secret")

            mock_encode.assert_called_once()
            call_args = mock_encode.call_args
            payload = call_args[0][0]

            assert payload["role"] == "admin"
            assert payload["iss"] == "lagoon-api"
            assert payload["sub"] == "lagoonadmin"
            assert "iat" in payload
            assert "exp" in payload
            assert result == "test-generated-token"

    def test_generate_admin_token_pyjwt_not_installed(self, clean_env, mock_pulumi_config):
        """Test _generate_admin_token raises error when PyJWT not installed."""
        with patch.dict(os.environ, {"LAGOON_TOKEN": "temp-token"}):
            from pulumi_lagoon.config import LagoonConfig

            config = LagoonConfig()

        # Mock jwt module import to raise ImportError
        import builtins

        original_import = builtins.__import__

        def mock_import(name, *args, **kwargs):
            if name == "jwt":
                raise ImportError("No module named 'jwt'")
            return original_import(name, *args, **kwargs)

        with patch.object(builtins, "__import__", mock_import):
            with pytest.raises(ValueError, match="PyJWT is required"):
                config._generate_admin_token("test-secret")

    def test_generate_admin_token_encoding_error(self, clean_env, mock_pulumi_config):
        """Test _generate_admin_token handles encoding errors."""
        with patch.dict(os.environ, {"LAGOON_TOKEN": "temp-token"}):
            from pulumi_lagoon.config import LagoonConfig

            config = LagoonConfig()

        with patch("jwt.encode") as mock_encode:
            mock_encode.side_effect = Exception("Encoding failed")

            with pytest.raises(ValueError, match="Failed to generate admin token"):
                config._generate_admin_token("bad-secret")

    def test_config_from_jwt_secret_env_var(self, clean_env):
        """Test token is generated from LAGOON_JWT_SECRET env var."""
        with patch.dict(os.environ, {"LAGOON_JWT_SECRET": "env-jwt-secret"}):
            with patch("pulumi.Config") as mock_config_class:
                mock_config = Mock()
                mock_config_class.return_value = mock_config
                mock_config.get.return_value = None

                with patch("pulumi_lagoon.config.LagoonConfig._generate_admin_token") as mock_gen:
                    mock_gen.return_value = "generated-token-from-env"

                    from pulumi_lagoon.config import LagoonConfig

                    config = LagoonConfig()

                    mock_gen.assert_called_once_with("env-jwt-secret")
                    assert config.token == "generated-token-from-env"


class TestLagoonConfigRequiredValues:
    """Tests for required configuration value handling."""

    def test_get_config_value_raises_for_missing_required(self, clean_env):
        """Test _get_config_value raises error for missing required config."""
        with patch.dict(os.environ, {"LAGOON_TOKEN": "temp-token"}):
            with patch("pulumi.Config") as mock_config_class:
                mock_config = Mock()
                mock_config_class.return_value = mock_config
                mock_config.get.return_value = None

                from pulumi_lagoon.config import LagoonConfig

                config = LagoonConfig()

                # Test calling _get_config_value with required=True and no default
                with pytest.raises(ValueError, match="must be provided via"):
                    config._get_config_value(
                        mock_config,
                        "customKey",
                        "CUSTOM_ENV_VAR",
                        default=None,
                        required=True,
                    )
