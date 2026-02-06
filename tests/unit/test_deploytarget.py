"""Unit tests for LagoonDeployTarget provider."""

from unittest.mock import Mock, patch

import pytest

from pulumi_lagoon.exceptions import LagoonValidationError


class TestLagoonDeployTargetProviderCreate:
    """Tests for LagoonDeployTargetProvider.create method."""

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_create_deploy_target_minimal(self, mock_config_class, sample_deploy_target):
        """Test creating a deploy target with minimal arguments."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        # Setup mock
        mock_client = Mock()
        mock_client.add_kubernetes.return_value = sample_deploy_target
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
        }

        result = provider.create(inputs)

        assert result.id == "1"
        assert result.outs["name"] == "prod-cluster"
        assert result.outs["id"] == 1

        mock_client.add_kubernetes.assert_called_once()

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_create_deploy_target_full(self, mock_config_class, sample_deploy_target):
        """Test creating a deploy target with all arguments."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        # Setup mock
        mock_client = Mock()
        mock_client.add_kubernetes.return_value = sample_deploy_target
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
            "cloud_provider": "aws",
            "cloud_region": "us-east-1",
            "ssh_host": "ssh.lagoon.example.com",
            "ssh_port": "22",
            "build_image": "custom-build:latest",
            "disabled": False,
            "router_pattern": "${environment}.${project}.example.com",
        }

        provider.create(inputs)

        # Verify the API was called with correct arguments
        call_kwargs = mock_client.add_kubernetes.call_args[1]
        assert call_kwargs["name"] == "prod-cluster"
        assert call_kwargs["console_url"] == "https://kubernetes.example.com:6443"
        assert call_kwargs["cloud_provider"] == "aws"
        assert call_kwargs["cloud_region"] == "us-east-1"
        assert call_kwargs["sshHost"] == "ssh.lagoon.example.com"

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_create_deploy_target_defaults(self, mock_config_class, sample_deploy_target):
        """Test that defaults are applied for cloud_provider and cloud_region."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        # Modify sample to have kind/local defaults
        default_target = sample_deploy_target.copy()
        default_target["cloudProvider"] = "kind"
        default_target["cloudRegion"] = "local"

        mock_client = Mock()
        mock_client.add_kubernetes.return_value = default_target
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        inputs = {
            "name": "dev-cluster",
            "console_url": "https://kubernetes.default.svc",
        }

        provider.create(inputs)

        call_kwargs = mock_client.add_kubernetes.call_args[1]
        assert call_kwargs["cloud_provider"] == "kind"
        assert call_kwargs["cloud_region"] == "local"


class TestLagoonDeployTargetProviderUpdate:
    """Tests for LagoonDeployTargetProvider.update method."""

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_update_deploy_target_changes(self, mock_config_class, sample_deploy_target):
        """Test updating a deploy target with changed values."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        # Setup mock
        updated_target = sample_deploy_target.copy()
        updated_target["cloudRegion"] = "us-west-2"

        mock_client = Mock()
        mock_client.update_kubernetes.return_value = updated_target
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        old_inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
            "cloud_provider": "aws",
            "cloud_region": "us-east-1",
        }

        new_inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
            "cloud_provider": "aws",
            "cloud_region": "us-west-2",
        }

        result = provider.update("1", old_inputs, new_inputs)

        mock_client.update_kubernetes.assert_called_once()
        assert result.outs["cloud_region"] == "us-west-2"

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_update_deploy_target_no_changes(self, mock_config_class):
        """Test update with no actual changes returns inputs."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        # Setup mock
        mock_client = Mock()
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
            "cloud_provider": "aws",
            "cloud_region": "us-east-1",
        }

        result = provider.update("1", inputs, inputs)

        # Should not call update_kubernetes when nothing changed
        mock_client.update_kubernetes.assert_not_called()
        assert result.outs == inputs


class TestLagoonDeployTargetProviderDelete:
    """Tests for LagoonDeployTargetProvider.delete method."""

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_delete_deploy_target(self, mock_config_class):
        """Test deleting a deploy target."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        # Setup mock
        mock_client = Mock()
        mock_client.delete_kubernetes.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        props = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
        }

        provider.delete("1", props)

        mock_client.delete_kubernetes.assert_called_once_with("prod-cluster")


class TestLagoonDeployTargetProviderRead:
    """Tests for LagoonDeployTargetProvider.read method."""

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_read_deploy_target_exists(self, mock_config_class, sample_deploy_target):
        """Test reading an existing deploy target."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        # Setup mock
        mock_client = Mock()
        mock_client.get_kubernetes_by_id.return_value = sample_deploy_target
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        props = {"name": "prod-cluster"}

        result = provider.read("1", props)

        assert result.id == "1"
        assert result.outs["name"] == "prod-cluster"
        mock_client.get_kubernetes_by_id.assert_called_once_with(1)

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_read_deploy_target_not_found(self, mock_config_class):
        """Test reading a deploy target that doesn't exist."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        # Setup mock - deploy target not found
        mock_client = Mock()
        mock_client.get_kubernetes_by_id.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        props = {"name": "deleted-cluster"}

        result = provider.read("999", props)

        assert result is None


class TestLagoonDeployTargetArgs:
    """Tests for LagoonDeployTargetArgs dataclass."""

    def test_args_minimal(self):
        """Test creating args with minimal required fields."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetArgs

        args = LagoonDeployTargetArgs(
            name="prod-cluster",
            console_url="https://kubernetes.example.com:6443",
        )

        assert args.name == "prod-cluster"
        assert args.console_url == "https://kubernetes.example.com:6443"
        assert args.cloud_provider is None
        assert args.cloud_region is None
        assert args.ssh_host is None

    def test_args_full(self):
        """Test creating args with all fields."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetArgs

        args = LagoonDeployTargetArgs(
            name="prod-cluster",
            console_url="https://kubernetes.example.com:6443",
            cloud_provider="aws",
            cloud_region="us-east-1",
            ssh_host="ssh.lagoon.example.com",
            ssh_port="22",
            build_image="custom-build:latest",
            disabled=False,
            router_pattern="${environment}.${project}.example.com",
            shared_bastion_secret="bastion-secret",
        )

        assert args.cloud_provider == "aws"
        assert args.cloud_region == "us-east-1"
        assert args.ssh_host == "ssh.lagoon.example.com"
        assert args.ssh_port == "22"
        assert args.build_image == "custom-build:latest"
        assert args.disabled is False
        assert args.router_pattern == "${environment}.${project}.example.com"
        assert args.shared_bastion_secret == "bastion-secret"


class TestLagoonDeployTargetProviderValidation:
    """Tests for input validation in LagoonDeployTargetProvider."""

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_create_invalid_deploy_target_name(self, mock_config_class):
        """Test that invalid deploy target names are rejected."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        provider = LagoonDeployTargetProvider()

        inputs = {
            "name": "Invalid-Name",  # Uppercase not allowed
            "console_url": "https://kubernetes.example.com:6443",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_create_invalid_console_url(self, mock_config_class):
        """Test that invalid console URLs are rejected."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        provider = LagoonDeployTargetProvider()

        inputs = {
            "name": "prod-cluster",
            "console_url": "not-a-valid-url",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "console_url" in str(exc.value)

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_create_invalid_cloud_provider(self, mock_config_class):
        """Test that invalid cloud providers are rejected."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        provider = LagoonDeployTargetProvider()

        inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
            "cloud_provider": "invalid_provider",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "cloud_provider" in str(exc.value)

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_create_invalid_ssh_port(self, mock_config_class):
        """Test that invalid SSH ports are rejected."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        provider = LagoonDeployTargetProvider()

        inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
            "ssh_port": "99999",  # Invalid port (out of range)
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "ssh_port" in str(exc.value)

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_update_invalid_console_url(self, mock_config_class):
        """Test that invalid console URL in update is rejected."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        provider = LagoonDeployTargetProvider()

        old_inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
        }
        new_inputs = {
            "name": "prod-cluster",
            "console_url": "invalid-url",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.update("1", old_inputs, new_inputs)
        assert "console_url" in str(exc.value)

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_create_valid_kind_cluster(self, mock_config_class, sample_deploy_target):
        """Test creating a valid Kind cluster deploy target."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        kind_target = sample_deploy_target.copy()
        kind_target["cloudProvider"] = "kind"
        kind_target["cloudRegion"] = "local"

        mock_client = Mock()
        mock_client.add_kubernetes.return_value = kind_target
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        inputs = {
            "name": "dev-cluster",
            "console_url": "https://kubernetes.default.svc",
            "cloud_provider": "kind",
            "cloud_region": "local",
        }

        result = provider.create(inputs)

        assert result.outs["cloud_provider"] == "kind"
        assert result.outs["cloud_region"] == "local"


class TestLagoonDeployTargetProviderClientConfig:
    """Tests for LagoonDeployTargetProvider client configuration."""

    def test_get_client_with_api_token(self, sample_deploy_target):
        """Test creating client with explicit api_url and api_token."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        # Patch the client import inside the module
        with patch("pulumi_lagoon.client.LagoonClient") as mock_client_class:
            mock_client = Mock()
            mock_client.add_kubernetes.return_value = sample_deploy_target
            mock_client_class.return_value = mock_client

            provider = LagoonDeployTargetProvider()

            inputs = {
                "name": "test-cluster",
                "console_url": "https://kubernetes.default.svc",
                "api_url": "https://api.lagoon.example.com/graphql",
                "api_token": "test-bearer-token",
            }

            provider.create(inputs)

            # Verify LagoonClient was created with correct args
            mock_client_class.assert_called_once_with(
                "https://api.lagoon.example.com/graphql",
                "test-bearer-token",
                verify_ssl=True,
            )

    def test_get_client_with_verify_ssl_false(self, sample_deploy_target):
        """Test creating client with verify_ssl=False for self-signed certs."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        with patch("pulumi_lagoon.client.LagoonClient") as mock_client_class:
            mock_client = Mock()
            mock_client.add_kubernetes.return_value = sample_deploy_target
            mock_client_class.return_value = mock_client

            provider = LagoonDeployTargetProvider()

            inputs = {
                "name": "test-cluster",
                "console_url": "https://kubernetes.default.svc",
                "api_url": "https://api.lagoon.local/graphql",
                "api_token": "test-bearer-token",
                "verify_ssl": False,
            }

            provider.create(inputs)

            # Verify LagoonClient was created with verify_ssl=False
            mock_client_class.assert_called_once_with(
                "https://api.lagoon.local/graphql",
                "test-bearer-token",
                verify_ssl=False,
            )

    def test_get_client_with_jwt_secret(self, sample_deploy_target):
        """Test creating client with jwt_secret for token generation."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        with patch("pulumi_lagoon.client.LagoonClient") as mock_client_class:
            mock_client = Mock()
            mock_client.add_kubernetes.return_value = sample_deploy_target
            mock_client_class.return_value = mock_client

            provider = LagoonDeployTargetProvider()

            inputs = {
                "name": "test-cluster",
                "console_url": "https://kubernetes.default.svc",
                "api_url": "https://api.lagoon.example.com/graphql",
                "jwt_secret": "test-jwt-secret-key",
            }

            provider.create(inputs)

            # Verify LagoonClient was called (we can't easily verify the token)
            mock_client_class.assert_called_once()
            call_args = mock_client_class.call_args
            assert call_args[0][0] == "https://api.lagoon.example.com/graphql"
            # Token should be a JWT string
            assert isinstance(call_args[0][1], str)
            assert call_args[1]["verify_ssl"] is True

    def test_get_client_with_jwt_secret_and_verify_ssl_false(self, sample_deploy_target):
        """Test creating client with jwt_secret and verify_ssl=False."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        with patch("pulumi_lagoon.client.LagoonClient") as mock_client_class:
            mock_client = Mock()
            mock_client.add_kubernetes.return_value = sample_deploy_target
            mock_client_class.return_value = mock_client

            provider = LagoonDeployTargetProvider()

            inputs = {
                "name": "test-cluster",
                "console_url": "https://kubernetes.default.svc",
                "api_url": "https://api.lagoon.local/graphql",
                "jwt_secret": "test-jwt-secret-key",
                "verify_ssl": False,
            }

            provider.create(inputs)

            # Verify verify_ssl=False was passed
            call_args = mock_client_class.call_args
            assert call_args[1]["verify_ssl"] is False


class TestLagoonDeployTargetProviderJWTGeneration:
    """Tests for JWT token generation in LagoonDeployTargetProvider."""

    def test_generate_admin_token_valid(self):
        """Test that _generate_admin_token produces a valid JWT."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        import jwt as pyjwt

        provider = LagoonDeployTargetProvider()
        secret = "test-secret-key"

        token = provider._generate_admin_token(secret)

        # Verify it's a valid JWT that can be decoded (without audience verification)
        decoded = pyjwt.decode(token, secret, algorithms=["HS256"], options={"verify_aud": False})

        assert decoded["role"] == "admin"
        assert decoded["iss"] == "lagoon-api"
        assert decoded["sub"] == "lagoonadmin"
        assert decoded["aud"] == "api.dev"
        assert "iat" in decoded
        assert "exp" in decoded
        # Token should be valid for about an hour
        assert decoded["exp"] - decoded["iat"] == 3600

    def test_generate_admin_token_different_secrets(self):
        """Test that different secrets produce different tokens."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        provider = LagoonDeployTargetProvider()

        token1 = provider._generate_admin_token("secret-one")
        token2 = provider._generate_admin_token("secret-two")

        assert token1 != token2


class TestLagoonDeployTargetResourceInit:
    """Tests for LagoonDeployTarget resource initialization."""

    def test_resource_init_constructs_inputs_correctly(self):
        """Test that LagoonDeployTarget.__init__ constructs inputs dict from args."""
        from pulumi_lagoon.deploytarget import LagoonDeployTarget, LagoonDeployTargetArgs

        # Mock the parent class __init__ to capture inputs
        with patch("pulumi.dynamic.Resource.__init__") as mock_init:
            mock_init.return_value = None

            args = LagoonDeployTargetArgs(
                name="test-cluster",
                console_url="https://kubernetes.example.com:6443",
                cloud_provider="aws",
                cloud_region="us-east-1",
                ssh_host="ssh.example.com",
                ssh_port="22",
                build_image="custom-build:v1",
                disabled=False,
                router_pattern="${env}.${project}.example.com",
                shared_bastion_secret="bastion-secret",
                api_url="https://api.lagoon.example.com/graphql",
                api_token="test-token",
                jwt_secret=None,
                verify_ssl=False,
            )

            LagoonDeployTarget("test-resource", args)

            # Verify parent __init__ was called
            mock_init.assert_called_once()
            call_args = mock_init.call_args

            # Check the inputs dict (second positional arg after provider)
            inputs = call_args[0][2]  # provider, name, inputs, opts

            assert inputs["name"] == "test-cluster"
            assert inputs["console_url"] == "https://kubernetes.example.com:6443"
            assert inputs["cloud_provider"] == "aws"
            assert inputs["cloud_region"] == "us-east-1"
            assert inputs["ssh_host"] == "ssh.example.com"
            assert inputs["ssh_port"] == "22"
            assert inputs["build_image"] == "custom-build:v1"
            assert inputs["disabled"] is False
            assert inputs["router_pattern"] == "${env}.${project}.example.com"
            assert inputs["shared_bastion_secret"] == "bastion-secret"
            assert inputs["api_url"] == "https://api.lagoon.example.com/graphql"
            assert inputs["api_token"] == "test-token"
            assert inputs["jwt_secret"] is None
            assert inputs["verify_ssl"] is False
            assert inputs["id"] is None  # Output placeholder
            assert inputs["created"] is None  # Output placeholder

    def test_resource_init_with_minimal_args(self):
        """Test that LagoonDeployTarget.__init__ works with minimal required args."""
        from pulumi_lagoon.deploytarget import LagoonDeployTarget, LagoonDeployTargetArgs

        with patch("pulumi.dynamic.Resource.__init__") as mock_init:
            mock_init.return_value = None

            args = LagoonDeployTargetArgs(
                name="minimal-cluster",
                console_url="https://kubernetes.default.svc",
            )

            LagoonDeployTarget("minimal-resource", args)

            mock_init.assert_called_once()
            call_args = mock_init.call_args
            inputs = call_args[0][2]

            assert inputs["name"] == "minimal-cluster"
            assert inputs["console_url"] == "https://kubernetes.default.svc"
            # Optional fields should be None
            assert inputs["cloud_provider"] is None
            assert inputs["cloud_region"] is None
            assert inputs["ssh_host"] is None

    def test_resource_init_passes_opts(self):
        """Test that LagoonDeployTarget.__init__ passes ResourceOptions correctly."""
        from pulumi_lagoon.deploytarget import LagoonDeployTarget, LagoonDeployTargetArgs

        import pulumi

        with patch("pulumi.dynamic.Resource.__init__") as mock_init:
            mock_init.return_value = None

            args = LagoonDeployTargetArgs(
                name="test-cluster",
                console_url="https://kubernetes.default.svc",
            )

            opts = pulumi.ResourceOptions(protect=True)
            LagoonDeployTarget("test-resource", args, opts=opts)

            mock_init.assert_called_once()
            call_args = mock_init.call_args

            # opts is the 4th positional argument
            passed_opts = call_args[0][3]
            assert passed_opts is opts


class TestLagoonDeployTargetProviderUpdateValidation:
    """Tests for update validation in LagoonDeployTargetProvider."""

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_update_validates_cloud_provider_when_changed(self, mock_config_class, sample_deploy_target):
        """Test that cloud_provider is validated only when it changes."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        provider = LagoonDeployTargetProvider()

        old_inputs = {
            "name": "test-cluster",
            "console_url": "https://kubernetes.default.svc",
            "cloud_provider": "kind",
        }
        new_inputs = {
            "name": "test-cluster",
            "console_url": "https://kubernetes.default.svc",
            "cloud_provider": "invalid_provider",  # Invalid
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.update("1", old_inputs, new_inputs)
        assert "cloud_provider" in str(exc.value)

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_update_validates_ssh_host_when_changed(self, mock_config_class, sample_deploy_target):
        """Test that ssh_host is validated only when it changes."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        provider = LagoonDeployTargetProvider()

        old_inputs = {
            "name": "test-cluster",
            "console_url": "https://kubernetes.default.svc",
            "ssh_host": "ssh.example.com",
        }
        new_inputs = {
            "name": "test-cluster",
            "console_url": "https://kubernetes.default.svc",
            "ssh_host": "invalid host with spaces",  # Invalid
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.update("1", old_inputs, new_inputs)
        assert "ssh_host" in str(exc.value)

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_update_validates_ssh_port_when_changed(self, mock_config_class, sample_deploy_target):
        """Test that ssh_port is validated only when it changes."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        provider = LagoonDeployTargetProvider()

        old_inputs = {
            "name": "test-cluster",
            "console_url": "https://kubernetes.default.svc",
            "ssh_port": "22",
        }
        new_inputs = {
            "name": "test-cluster",
            "console_url": "https://kubernetes.default.svc",
            "ssh_port": "99999",  # Invalid - out of range
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.update("1", old_inputs, new_inputs)
        assert "ssh_port" in str(exc.value)

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_update_with_multiple_field_changes(self, mock_config_class, sample_deploy_target):
        """Test update with multiple fields changing."""
        from pulumi_lagoon.deploytarget import LagoonDeployTargetProvider

        updated_target = sample_deploy_target.copy()
        updated_target["sshHost"] = "new-ssh.example.com"
        updated_target["sshPort"] = "2222"
        updated_target["buildImage"] = "new-build:v2"

        mock_client = Mock()
        mock_client.update_kubernetes.return_value = updated_target
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetProvider()

        old_inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
            "ssh_host": "old-ssh.example.com",
            "ssh_port": "22",
            "build_image": "old-build:v1",
        }
        new_inputs = {
            "name": "prod-cluster",
            "console_url": "https://kubernetes.example.com:6443",
            "ssh_host": "new-ssh.example.com",
            "ssh_port": "2222",
            "build_image": "new-build:v2",
        }

        result = provider.update("1", old_inputs, new_inputs)

        # Verify all changed fields were included
        call_kwargs = mock_client.update_kubernetes.call_args[1]
        assert call_kwargs["sshHost"] == "new-ssh.example.com"
        assert call_kwargs["sshPort"] == "2222"
        assert call_kwargs["buildImage"] == "new-build:v2"
