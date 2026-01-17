"""Unit tests for LagoonDeployTarget provider."""

import pytest
from unittest.mock import Mock, patch

from pulumi_lagoon.exceptions import LagoonValidationError


class TestLagoonDeployTargetProviderCreate:
    """Tests for LagoonDeployTargetProvider.create method."""

    @patch("pulumi_lagoon.deploytarget.LagoonConfig")
    def test_create_deploy_target_minimal(
        self, mock_config_class, sample_deploy_target
    ):
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
    def test_create_deploy_target_defaults(
        self, mock_config_class, sample_deploy_target
    ):
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
    def test_update_deploy_target_changes(
        self, mock_config_class, sample_deploy_target
    ):
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
