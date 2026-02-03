"""Unit tests for LagoonTask provider."""

from unittest.mock import Mock, patch

import pytest

from pulumi_lagoon.client import LagoonAPIError, LagoonConnectionError
from pulumi_lagoon.exceptions import LagoonValidationError


class TestLagoonTaskProviderCreate:
    """Tests for LagoonTaskProvider.create method."""

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_command_task(self, mock_config_class, sample_task):
        """Test creating a command-type task."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.add_advanced_task_definition.return_value = sample_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        inputs = {
            "name": "run-yarn-audit",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "project_id": 1,
        }

        result = provider.create(inputs)

        assert result.outs["name"] == "run-yarn-audit"
        assert result.outs["type"] == "command"
        assert result.outs["service"] == "node"
        assert result.id == "1"

        # Verify API call
        call_kwargs = mock_client.add_advanced_task_definition.call_args[1]
        assert call_kwargs["name"] == "run-yarn-audit"
        assert call_kwargs["task_type"] == "command"
        assert call_kwargs["command"] == "yarn audit"
        assert call_kwargs["project_id"] == 1

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_image_task(self, mock_config_class):
        """Test creating an image-type task."""
        from pulumi_lagoon.task import LagoonTaskProvider

        image_task = {
            "id": 2,
            "name": "db-backup",
            "type": "IMAGE",
            "service": "cli",
            "image": "amazeeio/database-tools:latest",
            "command": None,
            "projectId": 1,
            "created": "2024-01-01T00:00:00Z",
        }

        mock_client = Mock()
        mock_client.add_advanced_task_definition.return_value = image_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        inputs = {
            "name": "db-backup",
            "type": "image",
            "service": "cli",
            "image": "amazeeio/database-tools:latest",
            "project_id": 1,
        }

        result = provider.create(inputs)

        assert result.outs["name"] == "db-backup"
        assert result.outs["type"] == "image"
        assert result.outs["image"] == "amazeeio/database-tools:latest"

        # Verify API call
        call_kwargs = mock_client.add_advanced_task_definition.call_args[1]
        assert call_kwargs["task_type"] == "image"
        assert call_kwargs["image"] == "amazeeio/database-tools:latest"

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_environment_scoped_task(self, mock_config_class, sample_task):
        """Test creating an environment-scoped task."""
        from pulumi_lagoon.task import LagoonTaskProvider

        env_task = sample_task.copy()
        env_task["environmentId"] = 5
        env_task["projectId"] = None

        mock_client = Mock()
        mock_client.add_advanced_task_definition.return_value = env_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        inputs = {
            "name": "run-yarn-audit",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "environment_id": 5,
        }

        provider.create(inputs)

        call_kwargs = mock_client.add_advanced_task_definition.call_args[1]
        assert call_kwargs["environment_id"] == 5
        assert call_kwargs["project_id"] is None

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_group_scoped_task(self, mock_config_class, sample_task):
        """Test creating a group-scoped task."""
        from pulumi_lagoon.task import LagoonTaskProvider

        group_task = sample_task.copy()
        group_task["groupName"] = "developers"
        group_task["projectId"] = None

        mock_client = Mock()
        mock_client.add_advanced_task_definition.return_value = group_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        inputs = {
            "name": "run-yarn-audit",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "group_name": "developers",
        }

        provider.create(inputs)

        call_kwargs = mock_client.add_advanced_task_definition.call_args[1]
        assert call_kwargs["group_name"] == "developers"

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_task_with_permission(self, mock_config_class, sample_task):
        """Test creating a task with permission level."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.add_advanced_task_definition.return_value = sample_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        inputs = {
            "name": "run-yarn-audit",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "project_id": 1,
            "permission": "maintainer",
        }

        provider.create(inputs)

        call_kwargs = mock_client.add_advanced_task_definition.call_args[1]
        assert call_kwargs["permission"] == "maintainer"

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_task_with_arguments(self, mock_config_class, sample_task):
        """Test creating a task with argument definitions."""
        from pulumi_lagoon.task import LagoonTaskProvider

        task_with_args = sample_task.copy()
        task_with_args["advancedTaskDefinitionArguments"] = [
            {
                "id": 1,
                "name": "target_env",
                "displayName": "Target Environment",
                "type": "ENVIRONMENT_SOURCE_NAME",
            }
        ]

        mock_client = Mock()
        mock_client.add_advanced_task_definition.return_value = task_with_args
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        inputs = {
            "name": "deploy-task",
            "type": "command",
            "service": "cli",
            "command": "drush deploy",
            "project_id": 1,
            "arguments": [
                {
                    "name": "target_env",
                    "display_name": "Target Environment",
                    "type": "environment_source_name",
                }
            ],
        }

        result = provider.create(inputs)

        assert result.outs["arguments"] is not None
        assert len(result.outs["arguments"]) == 1
        assert result.outs["arguments"][0]["name"] == "target_env"

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_task_with_confirmation_text(self, mock_config_class, sample_task):
        """Test creating a task with confirmation text."""
        from pulumi_lagoon.task import LagoonTaskProvider

        task_with_confirm = sample_task.copy()
        task_with_confirm["confirmationText"] = "Are you sure?"

        mock_client = Mock()
        mock_client.add_advanced_task_definition.return_value = task_with_confirm
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        inputs = {
            "name": "dangerous-task",
            "type": "command",
            "service": "cli",
            "command": "rm -rf /tmp/*",
            "project_id": 1,
            "confirmation_text": "Are you sure?",
        }

        result = provider.create(inputs)

        assert result.outs["confirmation_text"] == "Are you sure?"


class TestLagoonTaskProviderUpdate:
    """Tests for LagoonTaskProvider.update method."""

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_update_task_command(self, mock_config_class, sample_task):
        """Test updating a task's command."""
        from pulumi_lagoon.task import LagoonTaskProvider

        updated_task = sample_task.copy()
        updated_task["command"] = "yarn audit --level high"
        updated_task["id"] = 2  # New ID after recreate

        mock_client = Mock()
        mock_client.delete_advanced_task_definition.return_value = "success"
        mock_client.add_advanced_task_definition.return_value = updated_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        old_inputs = {
            "name": "run-yarn-audit",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "project_id": 1,
        }

        new_inputs = {
            "name": "run-yarn-audit",
            "type": "command",
            "service": "node",
            "command": "yarn audit --level high",
            "project_id": 1,
        }

        provider.update("1", old_inputs, new_inputs)

        # Should delete old and create new
        mock_client.delete_advanced_task_definition.assert_called_once_with(1)
        mock_client.add_advanced_task_definition.assert_called_once()

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_update_handles_delete_failure(self, mock_config_class, sample_task):
        """Test update continues if delete fails with API error."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.delete_advanced_task_definition.side_effect = LagoonAPIError("Task not found")
        mock_client.add_advanced_task_definition.return_value = sample_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        old_inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "command": "old",
            "project_id": 1,
        }

        new_inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "command": "new",
            "project_id": 1,
        }

        # Should not raise - LagoonAPIError is caught and ignored
        provider.update("1", old_inputs, new_inputs)

        mock_client.add_advanced_task_definition.assert_called_once()

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_update_connection_error_propagates(self, mock_config_class):
        """Test that LagoonConnectionError during delete propagates."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.delete_advanced_task_definition.side_effect = LagoonConnectionError(
            "Network error"
        )
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        old_inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "command": "old",
            "project_id": 1,
        }

        new_inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "command": "new",
            "project_id": 1,
        }

        # Should raise - connection errors should propagate
        with pytest.raises(LagoonConnectionError):
            provider.update("1", old_inputs, new_inputs)


class TestLagoonTaskProviderDelete:
    """Tests for LagoonTaskProvider.delete method."""

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_delete_task(self, mock_config_class):
        """Test deleting a task."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.delete_advanced_task_definition.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        props = {
            "name": "run-yarn-audit",
            "type": "command",
            "project_id": 1,
        }

        provider.delete("123", props)

        mock_client.delete_advanced_task_definition.assert_called_once_with(123)

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_delete_handles_string_id(self, mock_config_class):
        """Test that delete converts string ID to int."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.delete_advanced_task_definition.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        provider.delete("456", {})

        mock_client.delete_advanced_task_definition.assert_called_once_with(456)


class TestLagoonTaskProviderRead:
    """Tests for LagoonTaskProvider.read method."""

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_read_task_exists(self, mock_config_class, sample_task):
        """Test reading an existing task."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.get_advanced_task_definition_by_id.return_value = sample_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        props = {
            "name": "run-yarn-audit",
            "type": "command",
            "project_id": 1,
        }

        result = provider.read("1", props)

        assert result.outs["name"] == "run-yarn-audit"
        mock_client.get_advanced_task_definition_by_id.assert_called_once_with(1)

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_read_task_not_found(self, mock_config_class):
        """Test reading a task that doesn't exist."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.get_advanced_task_definition_by_id.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        props = {
            "name": "deleted-task",
            "type": "command",
        }

        result = provider.read("999", props)

        assert result is None

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_read_import_scenario(self, mock_config_class, sample_task):
        """Test read() during import (empty props)."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.get_advanced_task_definition_by_id.return_value = sample_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        # Import scenario: numeric ID, empty props
        result = provider.read("123", {})

        assert result is not None
        mock_client.get_advanced_task_definition_by_id.assert_called_once_with(123)


class TestLagoonTaskArgs:
    """Tests for LagoonTaskArgs dataclass."""

    def test_args_command_task(self):
        """Test creating args for command-type task."""
        from pulumi_lagoon.task import LagoonTaskArgs

        args = LagoonTaskArgs(
            name="run-yarn-audit",
            type="command",
            service="node",
            command="yarn audit",
            project_id=1,
        )

        assert args.name == "run-yarn-audit"
        assert args.type == "command"
        assert args.service == "node"
        assert args.command == "yarn audit"
        assert args.project_id == 1
        assert args.image is None

    def test_args_image_task(self):
        """Test creating args for image-type task."""
        from pulumi_lagoon.task import LagoonTaskArgs

        args = LagoonTaskArgs(
            name="db-backup",
            type="image",
            service="cli",
            image="amazeeio/database-tools:latest",
            project_id=1,
        )

        assert args.type == "image"
        assert args.image == "amazeeio/database-tools:latest"
        assert args.command is None

    def test_args_with_arguments(self):
        """Test creating args with task arguments."""
        from pulumi_lagoon.task import LagoonTaskArgs

        args = LagoonTaskArgs(
            name="deploy-task",
            type="command",
            service="cli",
            command="drush deploy",
            project_id=1,
            arguments=[
                {
                    "name": "target_env",
                    "display_name": "Target Environment",
                    "type": "environment_source_name",
                }
            ],
        )

        assert args.arguments is not None
        assert len(args.arguments) == 1


class TestLagoonTaskProviderValidation:
    """Tests for input validation in LagoonTaskProvider."""

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_missing_name(self, mock_config_class):
        """Test that missing name is rejected."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        inputs = {
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "project_id": 1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_invalid_type(self, mock_config_class):
        """Test that invalid type is rejected."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        inputs = {
            "name": "task",
            "type": "invalid_type",
            "service": "node",
            "command": "yarn audit",
            "project_id": 1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "type" in str(exc.value)

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_command_without_command(self, mock_config_class):
        """Test that command-type task without command is rejected."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "project_id": 1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "command" in str(exc.value).lower()

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_image_without_image(self, mock_config_class):
        """Test that image-type task without image is rejected."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        inputs = {
            "name": "task",
            "type": "image",
            "service": "cli",
            "project_id": 1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "image" in str(exc.value).lower()

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_command_with_image(self, mock_config_class):
        """Test that command-type task with image is rejected."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "image": "some-image",
            "project_id": 1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "image" in str(exc.value).lower()

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_no_scope(self, mock_config_class):
        """Test that missing scope is rejected."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "scope" in str(exc.value).lower()

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_multiple_scopes(self, mock_config_class):
        """Test that multiple scopes are rejected."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "project_id": 1,
            "environment_id": 2,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "scope" in str(exc.value).lower() or "multiple" in str(exc.value).lower()

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_invalid_permission(self, mock_config_class):
        """Test that invalid permission is rejected."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "project_id": 1,
            "permission": "invalid_permission",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "permission" in str(exc.value)

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_create_invalid_project_id(self, mock_config_class):
        """Test that invalid project_id is rejected."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        inputs = {
            "name": "task",
            "type": "command",
            "service": "node",
            "command": "yarn audit",
            "project_id": -1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "project_id" in str(exc.value)


class TestLagoonTaskProviderImport:
    """Tests for import functionality in LagoonTaskProvider."""

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_read_import_by_id(self, mock_config_class, sample_task):
        """Test read() during import with numeric ID."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.get_advanced_task_definition_by_id.return_value = sample_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        # Import scenario: numeric ID, empty props
        result = provider.read("456", {})

        assert result is not None
        assert result.outs["name"] == "run-yarn-audit"
        mock_client.get_advanced_task_definition_by_id.assert_called_once_with(456)

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_read_import_invalid_format(self, mock_config_class):
        """Test read() during import with invalid ID format."""
        from pulumi_lagoon.task import LagoonTaskProvider

        provider = LagoonTaskProvider()

        with pytest.raises(LagoonValidationError) as exc:
            provider.read("not-a-number", {})
        assert "must be a number" in str(exc.value)

    @patch("pulumi_lagoon.task.LagoonConfig")
    def test_read_refresh_uses_props(self, mock_config_class, sample_task):
        """Test read() during refresh uses props, not ID parsing."""
        from pulumi_lagoon.task import LagoonTaskProvider

        mock_client = Mock()
        mock_client.get_advanced_task_definition_by_id.return_value = sample_task
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonTaskProvider()

        # Refresh scenario: ID with full props
        props = {"name": "run-yarn-audit", "type": "command"}
        result = provider.read("1", props)

        assert result is not None
        mock_client.get_advanced_task_definition_by_id.assert_called_once_with(1)


class TestLagoonTaskClientMethods:
    """Tests for task-related client methods."""

    def test_add_advanced_task_definition(self, lagoon_client, mock_response):
        """Test creating a task definition via client."""
        response = mock_response(
            data={
                "addAdvancedTaskDefinition": {
                    "id": 1,
                    "name": "test-task",
                    "type": "COMMAND",
                    "service": "node",
                    "command": "yarn audit",
                    "project": {"id": 1, "name": "test"},
                }
            }
        )
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_advanced_task_definition(
            name="test-task",
            task_type="COMMAND",
            service="node",
            command="yarn audit",
            project_id=1,
        )

        assert result["name"] == "test-task"
        assert result["projectId"] == 1

    def test_get_advanced_task_definition_by_id(self, lagoon_client, mock_response):
        """Test getting a task definition by ID."""
        response = mock_response(
            data={
                "advancedTaskDefinitionById": {
                    "id": 1,
                    "name": "test-task",
                    "type": "COMMAND",
                    "service": "node",
                    "command": "yarn audit",
                }
            }
        )
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_advanced_task_definition_by_id(1)

        assert result["name"] == "test-task"

    def test_delete_advanced_task_definition(self, lagoon_client, mock_response):
        """Test deleting a task definition."""
        response = mock_response(data={"deleteAdvancedTaskDefinition": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_advanced_task_definition(1)

        assert result == "success"

    def test_get_advanced_tasks_by_environment(self, lagoon_client, mock_response):
        """Test getting tasks for an environment."""
        response = mock_response(
            data={
                "advancedTasksByEnvironment": [
                    {
                        "id": 1,
                        "name": "task1",
                        "type": "COMMAND",
                        "service": "node",
                        "project": {"id": 1, "name": "test"},
                    },
                    {
                        "id": 2,
                        "name": "task2",
                        "type": "IMAGE",
                        "service": "cli",
                        "project": {"id": 1, "name": "test"},
                    },
                ]
            }
        )
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_advanced_tasks_by_environment(5)

        assert len(result) == 2
        assert result[0]["name"] == "task1"
        assert result[1]["name"] == "task2"
