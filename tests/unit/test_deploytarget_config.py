"""Unit tests for LagoonDeployTargetConfig provider."""

from unittest.mock import Mock, patch

import pytest

# Sample deploy target config data for tests
SAMPLE_DEPLOY_TARGET_CONFIG = {
    "id": 5,
    "projectId": 123,
    "deployTargetId": 1,
    "branches": "main",
    "pullrequests": "false",
    "weight": 10,
    "deployTargetProjectPattern": None,
}


@pytest.fixture
def sample_deploy_target_config():
    """Return sample deploy target config data."""
    return SAMPLE_DEPLOY_TARGET_CONFIG.copy()


class TestLagoonDeployTargetConfigProviderCreate:
    """Tests for LagoonDeployTargetConfigProvider.create method."""

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_create_minimal(self, mock_config_class, sample_deploy_target_config):
        """Test creating a deploy target config with minimal arguments."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        mock_client = Mock()
        mock_client.add_deploy_target_config.return_value = sample_deploy_target_config
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        inputs = {
            "project_id": 123,
            "deploy_target_id": 1,
        }

        result = provider.create(inputs)

        assert result.id == "5"
        assert result.outs["project_id"] == 123
        mock_client.add_deploy_target_config.assert_called_once()

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_create_with_all_options(self, mock_config_class, sample_deploy_target_config):
        """Test creating a deploy target config with all options."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        mock_client = Mock()
        mock_client.add_deploy_target_config.return_value = sample_deploy_target_config
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        inputs = {
            "project_id": 123,
            "deploy_target_id": 1,
            "branches": "main",
            "pullrequests": "false",
            "weight": 10,
            "deploy_target_project_pattern": "${project}-${environment}",
        }

        provider.create(inputs)

        call_kwargs = mock_client.add_deploy_target_config.call_args[1]
        assert call_kwargs["project"] == 123
        assert call_kwargs["deploy_target"] == 1
        assert call_kwargs["branches"] == "main"


class TestLagoonDeployTargetConfigProviderUpdate:
    """Tests for LagoonDeployTargetConfigProvider.update method."""

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_update_branches(self, mock_config_class, sample_deploy_target_config):
        """Test updating a deploy target config."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        updated_config = sample_deploy_target_config.copy()
        updated_config["branches"] = "^(main|develop)$"

        mock_client = Mock()
        mock_client.update_deploy_target_config.return_value = updated_config
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        old_inputs = {
            "project_id": 123,
            "deploy_target_id": 1,
            "branches": "main",
        }

        new_inputs = {
            "project_id": 123,
            "deploy_target_id": 1,
            "branches": "^(main|develop)$",
        }

        provider.update("5", old_inputs, new_inputs)

        mock_client.update_deploy_target_config.assert_called_once()


class TestLagoonDeployTargetConfigProviderDelete:
    """Tests for LagoonDeployTargetConfigProvider.delete method."""

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_delete(self, mock_config_class):
        """Test deleting a deploy target config."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        mock_client = Mock()
        mock_client.delete_deploy_target_config.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        props = {
            "project_id": 123,
            "deploy_target_id": 1,
        }

        provider.delete("5", props)

        mock_client.delete_deploy_target_config.assert_called_once_with(5, 123)


class TestLagoonDeployTargetConfigProviderRead:
    """Tests for LagoonDeployTargetConfigProvider.read method."""

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_read_exists(self, mock_config_class, sample_deploy_target_config):
        """Test reading an existing deploy target config."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        mock_client = Mock()
        mock_client.get_deploy_target_config_by_id.return_value = sample_deploy_target_config
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        props = {
            "project_id": 123,
            "deploy_target_id": 1,
        }

        result = provider.read("5", props)

        assert result is not None
        assert result.outs["branches"] == "main"
        mock_client.get_deploy_target_config_by_id.assert_called_once_with(5, 123)

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_read_not_found(self, mock_config_class):
        """Test reading a deploy target config that doesn't exist."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        mock_client = Mock()
        mock_client.get_deploy_target_config_by_id.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        props = {
            "project_id": 123,
        }

        result = provider.read("999", props)

        assert result is None


class TestLagoonDeployTargetConfigProviderImport:
    """Tests for import functionality in LagoonDeployTargetConfigProvider."""

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_read_import_scenario_empty_props(self, mock_config_class, sample_deploy_target_config):
        """Test read() during import with empty props."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        mock_client = Mock()
        mock_client.get_deploy_target_config_by_id.return_value = sample_deploy_target_config
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        # Import scenario: composite ID, empty props
        result = provider.read("123:5", {})

        assert result is not None
        assert result.outs["id"] == 5
        assert result.outs["project_id"] == 123
        mock_client.get_deploy_target_config_by_id.assert_called_once_with(5, 123)

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_read_refresh_scenario_uses_props(self, mock_config_class, sample_deploy_target_config):
        """Test read() during refresh uses props, not ID parsing."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        mock_client = Mock()
        mock_client.get_deploy_target_config_by_id.return_value = sample_deploy_target_config
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        # Refresh scenario: numeric ID with full props
        props = {"project_id": 123, "deploy_target_id": 1}
        result = provider.read("5", props)

        assert result is not None
        # In refresh, the ID is the config_id and project_id comes from props
        mock_client.get_deploy_target_config_by_id.assert_called_once_with(5, 123)

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_read_import_not_found(self, mock_config_class):
        """Test read() during import when config doesn't exist."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        mock_client = Mock()
        mock_client.get_deploy_target_config_by_id.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        result = provider.read("999:888", {})

        assert result is None

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_read_import_invalid_format(self, mock_config_class):
        """Test read() during import with invalid ID format."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider
        from pulumi_lagoon.exceptions import LagoonValidationError

        provider = LagoonDeployTargetConfigProvider()

        with pytest.raises(LagoonValidationError) as exc:
            provider.read("invalid-no-colon", {})
        assert "project_id:config_id" in str(exc.value)

    @patch("pulumi_lagoon.deploytarget_config.LagoonConfig")
    def test_read_import_large_ids(self, mock_config_class, sample_deploy_target_config):
        """Test read() during import with large ID values."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigProvider

        large_config = sample_deploy_target_config.copy()
        large_config["id"] = 999999
        large_config["projectId"] = 888888

        mock_client = Mock()
        mock_client.get_deploy_target_config_by_id.return_value = large_config
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonDeployTargetConfigProvider()

        result = provider.read("888888:999999", {})

        assert result is not None
        mock_client.get_deploy_target_config_by_id.assert_called_once_with(999999, 888888)


class TestLagoonDeployTargetConfigArgs:
    """Tests for LagoonDeployTargetConfigArgs dataclass."""

    def test_args_minimal(self):
        """Test creating args with minimal required fields."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigArgs

        args = LagoonDeployTargetConfigArgs(
            project_id=123,
            deploy_target_id=1,
        )

        assert args.project_id == 123
        assert args.deploy_target_id == 1
        assert args.branches is None
        assert args.pullrequests is None
        assert args.weight is None

    def test_args_full(self):
        """Test creating args with all fields."""
        from pulumi_lagoon.deploytarget_config import LagoonDeployTargetConfigArgs

        args = LagoonDeployTargetConfigArgs(
            project_id=123,
            deploy_target_id=1,
            branches="^main$",
            pullrequests="true",
            weight=10,
            deploy_target_project_pattern="${project}-${env}",
        )

        assert args.branches == "^main$"
        assert args.pullrequests == "true"
        assert args.weight == 10
        assert args.deploy_target_project_pattern == "${project}-${env}"
