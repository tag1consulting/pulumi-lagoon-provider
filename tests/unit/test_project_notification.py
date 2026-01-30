"""Unit tests for LagoonProjectNotification provider."""

from unittest.mock import Mock, patch

import pytest

from pulumi_lagoon.exceptions import LagoonValidationError


class TestLagoonProjectNotificationProviderCreate:
    """Tests for LagoonProjectNotificationProvider.create method."""

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_create_project_notification_slack(self, mock_config_class, sample_project):
        """Test creating a Slack project notification association."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        mock_client = Mock()
        mock_client.add_notification_to_project.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectNotificationProvider()

        inputs = {
            "project_name": "test-project",
            "notification_type": "slack",
            "notification_name": "deploy-alerts",
        }

        result = provider.create(inputs)

        assert result.id == "test-project:slack:deploy-alerts"
        assert result.outs["project_name"] == "test-project"
        assert result.outs["notification_type"] == "slack"
        assert result.outs["notification_name"] == "deploy-alerts"
        assert result.outs["project_id"] == 1

        mock_client.add_notification_to_project.assert_called_once_with(
            project="test-project",
            notification_type="slack",
            notification_name="deploy-alerts",
        )

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_create_project_notification_email(self, mock_config_class, sample_project):
        """Test creating an Email project notification association."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        mock_client = Mock()
        mock_client.add_notification_to_project.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectNotificationProvider()

        inputs = {
            "project_name": "test-project",
            "notification_type": "email",
            "notification_name": "ops-team",
        }

        result = provider.create(inputs)

        assert result.id == "test-project:email:ops-team"
        assert result.outs["notification_type"] == "email"

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_create_project_notification_rocketchat(self, mock_config_class, sample_project):
        """Test creating a RocketChat project notification association."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        mock_client = Mock()
        mock_client.add_notification_to_project.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectNotificationProvider()

        inputs = {
            "project_name": "test-project",
            "notification_type": "rocketchat",
            "notification_name": "team-chat",
        }

        result = provider.create(inputs)

        assert result.id == "test-project:rocketchat:team-chat"
        assert result.outs["notification_type"] == "rocketchat"

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_create_project_notification_microsoftteams(self, mock_config_class, sample_project):
        """Test creating a Microsoft Teams project notification association."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        mock_client = Mock()
        mock_client.add_notification_to_project.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectNotificationProvider()

        inputs = {
            "project_name": "test-project",
            "notification_type": "microsoftteams",
            "notification_name": "teams-alerts",
        }

        result = provider.create(inputs)

        assert result.id == "test-project:microsoftteams:teams-alerts"
        assert result.outs["notification_type"] == "microsoftteams"


class TestLagoonProjectNotificationProviderUpdate:
    """Tests for LagoonProjectNotificationProvider.update method."""

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_update_project_notification_different_notification(
        self, mock_config_class, sample_project
    ):
        """Test updating to a different notification."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        mock_client = Mock()
        mock_client.remove_notification_from_project.return_value = "success"
        mock_client.add_notification_to_project.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectNotificationProvider()

        old_inputs = {
            "project_name": "test-project",
            "notification_type": "slack",
            "notification_name": "old-alerts",
        }

        new_inputs = {
            "project_name": "test-project",
            "notification_type": "slack",
            "notification_name": "new-alerts",
        }

        result = provider.update("test-project:slack:old-alerts", old_inputs, new_inputs)

        assert result.outs["notification_name"] == "new-alerts"
        mock_client.remove_notification_from_project.assert_called_once()
        mock_client.add_notification_to_project.assert_called_once()


class TestLagoonProjectNotificationProviderDelete:
    """Tests for LagoonProjectNotificationProvider.delete method."""

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_delete_project_notification(self, mock_config_class):
        """Test deleting a project notification association."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        mock_client = Mock()
        mock_client.remove_notification_from_project.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectNotificationProvider()

        props = {
            "project_name": "test-project",
            "notification_type": "slack",
            "notification_name": "deploy-alerts",
        }

        provider.delete("test-project:slack:deploy-alerts", props)

        mock_client.remove_notification_from_project.assert_called_once_with(
            project="test-project",
            notification_type="slack",
            notification_name="deploy-alerts",
        )


class TestLagoonProjectNotificationProviderRead:
    """Tests for LagoonProjectNotificationProvider.read method."""

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_read_project_notification_exists(self, mock_config_class, sample_project):
        """Test reading an existing project notification association."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        mock_client = Mock()
        mock_client.check_project_notification_exists.return_value = True
        mock_client.get_project_by_name.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectNotificationProvider()

        props = {
            "project_name": "test-project",
            "notification_type": "slack",
            "notification_name": "deploy-alerts",
        }

        result = provider.read("test-project:slack:deploy-alerts", props)

        assert result.outs["project_name"] == "test-project"
        assert result.outs["notification_type"] == "slack"
        assert result.outs["notification_name"] == "deploy-alerts"
        assert result.outs["project_id"] == 1

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_read_project_notification_not_found(self, mock_config_class):
        """Test reading a project notification that doesn't exist."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        mock_client = Mock()
        mock_client.check_project_notification_exists.return_value = False
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectNotificationProvider()

        props = {
            "project_name": "test-project",
            "notification_type": "slack",
            "notification_name": "deleted-alerts",
        }

        result = provider.read("test-project:slack:deleted-alerts", props)

        assert result is None

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_read_import_scenario(self, mock_config_class, sample_project):
        """Test read() during import scenario."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        mock_client = Mock()
        mock_client.check_project_notification_exists.return_value = True
        mock_client.get_project_by_name.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectNotificationProvider()

        # Import scenario: composite ID, empty props
        result = provider.read("test-project:slack:deploy-alerts", {})

        assert result is not None
        assert result.outs["project_name"] == "test-project"
        assert result.outs["notification_type"] == "slack"
        assert result.outs["notification_name"] == "deploy-alerts"


class TestLagoonProjectNotificationProviderValidation:
    """Tests for input validation in LagoonProjectNotificationProvider."""

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_create_invalid_project_name(self, mock_config_class):
        """Test that invalid project names are rejected."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        provider = LagoonProjectNotificationProvider()

        inputs = {
            "project_name": "Invalid_Project",  # Underscores not allowed
            "notification_type": "slack",
            "notification_name": "deploy-alerts",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_create_invalid_notification_type(self, mock_config_class):
        """Test that invalid notification types are rejected."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        provider = LagoonProjectNotificationProvider()

        inputs = {
            "project_name": "test-project",
            "notification_type": "invalid",  # Not a valid type
            "notification_name": "deploy-alerts",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "notification_type" in str(exc.value).lower()

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_create_invalid_notification_name(self, mock_config_class):
        """Test that invalid notification names are rejected."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        provider = LagoonProjectNotificationProvider()

        inputs = {
            "project_name": "test-project",
            "notification_type": "slack",
            "notification_name": "123-invalid",  # Starts with number
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.project_notification.LagoonConfig")
    def test_create_empty_notification_type(self, mock_config_class):
        """Test that empty notification type is rejected."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationProvider

        provider = LagoonProjectNotificationProvider()

        inputs = {
            "project_name": "test-project",
            "notification_type": "",
            "notification_name": "deploy-alerts",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "notification_type" in str(exc.value).lower()


class TestLagoonProjectNotificationArgs:
    """Tests for LagoonProjectNotificationArgs dataclass."""

    def test_args_creation(self):
        """Test creating args for project notification."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationArgs

        args = LagoonProjectNotificationArgs(
            project_name="test-project",
            notification_type="slack",
            notification_name="deploy-alerts",
        )

        assert args.project_name == "test-project"
        assert args.notification_type == "slack"
        assert args.notification_name == "deploy-alerts"

    def test_args_all_notification_types(self):
        """Test args with different notification types."""
        from pulumi_lagoon.project_notification import LagoonProjectNotificationArgs

        for notification_type in ["slack", "rocketchat", "email", "microsoftteams"]:
            args = LagoonProjectNotificationArgs(
                project_name="test-project",
                notification_type=notification_type,
                notification_name="alerts",
            )
            assert args.notification_type == notification_type


class TestLagoonProjectNotificationImportIdParser:
    """Tests for import ID parsing for project notifications."""

    def test_parse_valid_import_id(self):
        """Test parsing a valid import ID."""
        from pulumi_lagoon.import_utils import ImportIdParser

        project_name, notification_type, notification_name = (
            ImportIdParser.parse_project_notification_id("my-project:slack:deploy-alerts")
        )

        assert project_name == "my-project"
        assert notification_type == "slack"
        assert notification_name == "deploy-alerts"

    def test_parse_valid_import_id_all_types(self):
        """Test parsing import IDs for all notification types."""
        from pulumi_lagoon.import_utils import ImportIdParser

        for notification_type in ["slack", "rocketchat", "email", "microsoftteams"]:
            project_name, parsed_type, notification_name = (
                ImportIdParser.parse_project_notification_id(
                    f"my-project:{notification_type}:alerts"
                )
            )
            assert parsed_type == notification_type

    def test_parse_import_id_invalid_format(self):
        """Test parsing an invalid import ID format."""
        from pulumi_lagoon.import_utils import ImportIdParser

        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_project_notification_id("invalid-format")
        assert "project_name:notification_type:notification_name" in str(exc.value)

    def test_parse_import_id_empty_project(self):
        """Test parsing import ID with empty project name."""
        from pulumi_lagoon.import_utils import ImportIdParser

        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_project_notification_id(":slack:alerts")
        assert "Project name cannot be empty" in str(exc.value)

    def test_parse_import_id_empty_type(self):
        """Test parsing import ID with empty notification type."""
        from pulumi_lagoon.import_utils import ImportIdParser

        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_project_notification_id("my-project::alerts")
        assert "Notification type cannot be empty" in str(exc.value)

    def test_parse_import_id_empty_name(self):
        """Test parsing import ID with empty notification name."""
        from pulumi_lagoon.import_utils import ImportIdParser

        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_project_notification_id("my-project:slack:")
        assert "Notification name cannot be empty" in str(exc.value)

    def test_parse_import_id_invalid_type(self):
        """Test parsing import ID with invalid notification type."""
        from pulumi_lagoon.import_utils import ImportIdParser

        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_project_notification_id("my-project:invalid:alerts")
        assert "must be one of" in str(exc.value)
