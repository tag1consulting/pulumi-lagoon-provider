"""Unit tests for LagoonEnvironment provider."""

import pytest
from unittest.mock import Mock, patch

from pulumi_lagoon.exceptions import LagoonValidationError


class TestLagoonEnvironmentProviderCreate:
    """Tests for LagoonEnvironmentProvider.create method."""

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_create_environment_minimal(self, mock_config_class, sample_environment):
        """Test creating an environment with minimal arguments."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        # Setup mock
        mock_client = Mock()
        mock_client.add_or_update_environment.return_value = sample_environment
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonEnvironmentProvider()

        inputs = {
            "name": "main",
            "project_id": 1,
            "deploy_type": "branch",
            "environment_type": "production",
        }

        result = provider.create(inputs)

        assert result.id == "1"
        assert result.outs["name"] == "main"
        assert result.outs["id"] == 1

        mock_client.add_or_update_environment.assert_called_once()

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_create_environment_sets_deploy_base_ref(
        self, mock_config_class, sample_environment
    ):
        """Test that deployBaseRef defaults to environment name."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        # Setup mock
        mock_client = Mock()
        mock_client.add_or_update_environment.return_value = sample_environment
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonEnvironmentProvider()

        inputs = {
            "name": "main",
            "project_id": 1,
            "deploy_type": "branch",
            "environment_type": "production",
        }

        provider.create(inputs)

        call_kwargs = mock_client.add_or_update_environment.call_args[1]
        assert call_kwargs["deployBaseRef"] == "main"

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_create_environment_with_pullrequest(
        self, mock_config_class, sample_environment
    ):
        """Test creating a pull request environment."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        pr_environment = sample_environment.copy()
        pr_environment["name"] = "pr-123"
        pr_environment["deployType"] = "PULLREQUEST"
        pr_environment["environmentType"] = "DEVELOPMENT"

        mock_client = Mock()
        mock_client.add_or_update_environment.return_value = pr_environment
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonEnvironmentProvider()

        inputs = {
            "name": "pr-123",
            "project_id": 1,
            "deploy_type": "pullrequest",
            "environment_type": "development",
            "deploy_base_ref": "main",
            "deploy_head_ref": "feature-branch",
            "deploy_title": "Add new feature",
        }

        provider.create(inputs)

        call_kwargs = mock_client.add_or_update_environment.call_args[1]
        assert call_kwargs["deployBaseRef"] == "main"
        assert call_kwargs["deployHeadRef"] == "feature-branch"
        assert call_kwargs["deployTitle"] == "Add new feature"

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_create_handles_string_project_id(
        self, mock_config_class, sample_environment
    ):
        """Test that string project_id is converted to int."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        mock_client = Mock()
        mock_client.add_or_update_environment.return_value = sample_environment
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonEnvironmentProvider()

        inputs = {
            "name": "main",
            "project_id": "1",  # String instead of int
            "deploy_type": "branch",
            "environment_type": "production",
        }

        provider.create(inputs)

        call_kwargs = mock_client.add_or_update_environment.call_args[1]
        assert call_kwargs["project"] == 1


class TestLagoonEnvironmentProviderUpdate:
    """Tests for LagoonEnvironmentProvider.update method."""

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_update_environment(self, mock_config_class, sample_environment):
        """Test updating an environment."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        mock_client = Mock()
        mock_client.add_or_update_environment.return_value = sample_environment
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonEnvironmentProvider()

        old_inputs = {
            "name": "main",
            "project_id": 1,
            "deploy_type": "branch",
            "environment_type": "production",
        }

        new_inputs = {
            "name": "main",
            "project_id": 1,
            "deploy_type": "branch",
            "environment_type": "production",
            "auto_idle": 60,
        }

        provider.update("1", old_inputs, new_inputs)

        mock_client.add_or_update_environment.assert_called_once()


class TestLagoonEnvironmentProviderDelete:
    """Tests for LagoonEnvironmentProvider.delete method."""

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_delete_environment(self, mock_config_class):
        """Test deleting an environment."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        mock_client = Mock()
        mock_client.delete_environment.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonEnvironmentProvider()

        props = {
            "name": "develop",
            "project_id": 1,
        }

        provider.delete("1", props)

        mock_client.delete_environment.assert_called_once_with(
            name="develop", project=1, execute=True
        )


class TestLagoonEnvironmentProviderRead:
    """Tests for LagoonEnvironmentProvider.read method."""

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_read_environment_exists(self, mock_config_class, sample_environment):
        """Test reading an existing environment."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        mock_client = Mock()
        mock_client.get_environment_by_name.return_value = sample_environment
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonEnvironmentProvider()

        props = {
            "name": "main",
            "project_id": 1,
        }

        result = provider.read("1", props)

        assert result.id == "1"
        assert result.outs["name"] == "main"
        mock_client.get_environment_by_name.assert_called_once_with(
            name="main", project_id=1
        )

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_read_environment_not_found(self, mock_config_class):
        """Test reading an environment that doesn't exist."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        mock_client = Mock()
        mock_client.get_environment_by_name.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonEnvironmentProvider()

        props = {
            "name": "deleted-env",
            "project_id": 1,
        }

        result = provider.read("999", props)

        assert result is None


class TestLagoonEnvironmentArgs:
    """Tests for LagoonEnvironmentArgs dataclass."""

    def test_args_minimal(self):
        """Test creating args with minimal required fields."""
        from pulumi_lagoon.environment import LagoonEnvironmentArgs

        args = LagoonEnvironmentArgs(
            name="main",
            project_id=1,
            deploy_type="branch",
            environment_type="production",
        )

        assert args.name == "main"
        assert args.project_id == 1
        assert args.deploy_type == "branch"
        assert args.environment_type == "production"
        assert args.deploy_base_ref is None

    def test_args_pullrequest(self):
        """Test creating args for a pull request environment."""
        from pulumi_lagoon.environment import LagoonEnvironmentArgs

        args = LagoonEnvironmentArgs(
            name="pr-123",
            project_id=1,
            deploy_type="pullrequest",
            environment_type="development",
            deploy_base_ref="main",
            deploy_head_ref="feature",
            deploy_title="New feature",
        )

        assert args.deploy_type == "pullrequest"
        assert args.deploy_base_ref == "main"
        assert args.deploy_head_ref == "feature"
        assert args.deploy_title == "New feature"


class TestLagoonEnvironmentProviderValidation:
    """Tests for input validation in LagoonEnvironmentProvider."""

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_create_invalid_environment_name(self, mock_config_class):
        """Test that invalid environment names are rejected."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        provider = LagoonEnvironmentProvider()

        inputs = {
            "name": "-invalid-name",
            "project_id": 1,
            "deploy_type": "branch",
            "environment_type": "production",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_create_invalid_project_id(self, mock_config_class):
        """Test that invalid project_id is rejected."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        provider = LagoonEnvironmentProvider()

        inputs = {
            "name": "main",
            "project_id": -1,
            "deploy_type": "branch",
            "environment_type": "production",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "project_id" in str(exc.value)

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_create_invalid_deploy_type(self, mock_config_class):
        """Test that invalid deploy_type is rejected."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        provider = LagoonEnvironmentProvider()

        inputs = {
            "name": "main",
            "project_id": 1,
            "deploy_type": "invalid_type",
            "environment_type": "production",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "deploy_type" in str(exc.value)
        assert "branch" in str(exc.value)

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_create_invalid_environment_type(self, mock_config_class):
        """Test that invalid environment_type is rejected."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        provider = LagoonEnvironmentProvider()

        inputs = {
            "name": "main",
            "project_id": 1,
            "deploy_type": "branch",
            "environment_type": "staging",  # Invalid - should be production/development/standby
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "environment_type" in str(exc.value)
        assert "production" in str(exc.value)

    @patch("pulumi_lagoon.environment.LagoonConfig")
    def test_update_invalid_deploy_type(self, mock_config_class):
        """Test that invalid deploy_type in update is rejected."""
        from pulumi_lagoon.environment import LagoonEnvironmentProvider

        provider = LagoonEnvironmentProvider()

        old_inputs = {
            "name": "main",
            "project_id": 1,
            "deploy_type": "branch",
            "environment_type": "production",
        }
        new_inputs = {
            "name": "main",
            "project_id": 1,
            "deploy_type": "invalid",
            "environment_type": "production",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.update("1", old_inputs, new_inputs)
        assert "deploy_type" in str(exc.value)
