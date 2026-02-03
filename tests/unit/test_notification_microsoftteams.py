"""Unit tests for LagoonNotificationMicrosoftTeams provider."""

from unittest.mock import Mock, patch

import pytest

from pulumi_lagoon.exceptions import LagoonValidationError


class TestLagoonNotificationMicrosoftTeamsProviderCreate:
    """Tests for LagoonNotificationMicrosoftTeamsProvider.create method."""

    @patch("pulumi_lagoon.notification_microsoftteams.LagoonConfig")
    def test_create_teams_notification(self, mock_config_class, sample_notification_microsoftteams):
        """Test creating a Microsoft Teams notification."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsProvider,
        )

        mock_client = Mock()
        mock_client.add_notification_microsoftteams.return_value = (
            sample_notification_microsoftteams
        )
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationMicrosoftTeamsProvider()

        inputs = {
            "name": "teams-alerts",
            "webhook": "https://outlook.office.com/webhook/xxx/yyy/zzz",
        }

        result = provider.create(inputs)

        assert result.id == "teams-alerts"
        assert result.outs["name"] == "teams-alerts"
        assert result.outs["webhook"] == "https://outlook.office.com/webhook/xxx/yyy/zzz"
        assert result.outs["lagoon_id"] == 4

        mock_client.add_notification_microsoftteams.assert_called_once_with(
            name="teams-alerts",
            webhook="https://outlook.office.com/webhook/xxx/yyy/zzz",
        )


class TestLagoonNotificationMicrosoftTeamsProviderUpdate:
    """Tests for LagoonNotificationMicrosoftTeamsProvider.update method."""

    @patch("pulumi_lagoon.notification_microsoftteams.LagoonConfig")
    def test_update_teams_notification_webhook(
        self, mock_config_class, sample_notification_microsoftteams
    ):
        """Test updating a Microsoft Teams notification's webhook."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsProvider,
        )

        updated_notification = sample_notification_microsoftteams.copy()
        updated_notification["webhook"] = "https://outlook.office.com/webhook/new/url/here"

        mock_client = Mock()
        mock_client.update_notification_microsoftteams.return_value = updated_notification
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationMicrosoftTeamsProvider()

        old_inputs = {
            "name": "teams-alerts",
            "webhook": "https://outlook.office.com/webhook/xxx/yyy/zzz",
        }

        new_inputs = {
            "name": "teams-alerts",
            "webhook": "https://outlook.office.com/webhook/new/url/here",
        }

        result = provider.update("teams-alerts", old_inputs, new_inputs)

        assert result.outs["webhook"] == "https://outlook.office.com/webhook/new/url/here"
        mock_client.update_notification_microsoftteams.assert_called_once()


class TestLagoonNotificationMicrosoftTeamsProviderDelete:
    """Tests for LagoonNotificationMicrosoftTeamsProvider.delete method."""

    @patch("pulumi_lagoon.notification_microsoftteams.LagoonConfig")
    def test_delete_teams_notification(self, mock_config_class):
        """Test deleting a Microsoft Teams notification."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsProvider,
        )

        mock_client = Mock()
        mock_client.delete_notification_microsoftteams.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationMicrosoftTeamsProvider()

        props = {
            "name": "teams-alerts",
            "webhook": "https://outlook.office.com/webhook/xxx/yyy/zzz",
        }

        provider.delete("teams-alerts", props)

        mock_client.delete_notification_microsoftteams.assert_called_once_with(name="teams-alerts")


class TestLagoonNotificationMicrosoftTeamsProviderRead:
    """Tests for LagoonNotificationMicrosoftTeamsProvider.read method."""

    @patch("pulumi_lagoon.notification_microsoftteams.LagoonConfig")
    def test_read_teams_notification_exists(
        self, mock_config_class, sample_notification_microsoftteams
    ):
        """Test reading an existing Microsoft Teams notification."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsProvider,
        )

        mock_client = Mock()
        mock_client.get_notification_microsoftteams_by_name.return_value = (
            sample_notification_microsoftteams
        )
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationMicrosoftTeamsProvider()

        props = {"name": "teams-alerts"}

        result = provider.read("teams-alerts", props)

        assert result.outs["name"] == "teams-alerts"
        assert result.outs["lagoon_id"] == 4
        mock_client.get_notification_microsoftteams_by_name.assert_called_once_with("teams-alerts")

    @patch("pulumi_lagoon.notification_microsoftteams.LagoonConfig")
    def test_read_teams_notification_not_found(self, mock_config_class):
        """Test reading a Microsoft Teams notification that doesn't exist."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsProvider,
        )

        mock_client = Mock()
        mock_client.get_notification_microsoftteams_by_name.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationMicrosoftTeamsProvider()

        props = {"name": "deleted-notification"}

        result = provider.read("deleted-notification", props)

        assert result is None

    @patch("pulumi_lagoon.notification_microsoftteams.LagoonConfig")
    def test_read_import_scenario(self, mock_config_class, sample_notification_microsoftteams):
        """Test read() during import scenario."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsProvider,
        )

        mock_client = Mock()
        mock_client.get_notification_microsoftteams_by_name.return_value = (
            sample_notification_microsoftteams
        )
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationMicrosoftTeamsProvider()

        # Import scenario: ID is the name, props is empty
        result = provider.read("teams-alerts", {})

        assert result is not None
        assert result.outs["name"] == "teams-alerts"


class TestLagoonNotificationMicrosoftTeamsProviderValidation:
    """Tests for input validation in LagoonNotificationMicrosoftTeamsProvider."""

    @patch("pulumi_lagoon.notification_microsoftteams.LagoonConfig")
    def test_create_invalid_name(self, mock_config_class):
        """Test that invalid notification names are rejected."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsProvider,
        )

        provider = LagoonNotificationMicrosoftTeamsProvider()

        inputs = {
            "name": "123-invalid",  # Starts with number
            "webhook": "https://outlook.office.com/webhook/xxx/yyy/zzz",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.notification_microsoftteams.LagoonConfig")
    def test_create_invalid_webhook_http(self, mock_config_class):
        """Test that HTTP webhooks are rejected (must be HTTPS)."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsProvider,
        )

        provider = LagoonNotificationMicrosoftTeamsProvider()

        inputs = {
            "name": "teams-alerts",
            "webhook": "http://outlook.office.com/webhook/xxx/yyy/zzz",  # HTTP not allowed
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "webhook" in str(exc.value).lower()

    @patch("pulumi_lagoon.notification_microsoftteams.LagoonConfig")
    def test_create_empty_webhook(self, mock_config_class):
        """Test that empty webhook is rejected."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsProvider,
        )

        provider = LagoonNotificationMicrosoftTeamsProvider()

        inputs = {
            "name": "teams-alerts",
            "webhook": "",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "webhook" in str(exc.value).lower()


class TestLagoonNotificationMicrosoftTeamsArgs:
    """Tests for LagoonNotificationMicrosoftTeamsArgs dataclass."""

    def test_args_creation(self):
        """Test creating args for Microsoft Teams notification."""
        from pulumi_lagoon.notification_microsoftteams import (
            LagoonNotificationMicrosoftTeamsArgs,
        )

        args = LagoonNotificationMicrosoftTeamsArgs(
            name="teams-alerts",
            webhook="https://outlook.office.com/webhook/xxx/yyy/zzz",
        )

        assert args.name == "teams-alerts"
        assert args.webhook == "https://outlook.office.com/webhook/xxx/yyy/zzz"
