"""Unit tests for LagoonNotificationRocketChat provider."""

from unittest.mock import Mock, patch

import pytest

from pulumi_lagoon.exceptions import LagoonValidationError


class TestLagoonNotificationRocketChatProviderCreate:
    """Tests for LagoonNotificationRocketChatProvider.create method."""

    @patch("pulumi_lagoon.notification_rocketchat.LagoonConfig")
    def test_create_rocketchat_notification(
        self, mock_config_class, sample_notification_rocketchat
    ):
        """Test creating a RocketChat notification."""
        from pulumi_lagoon.notification_rocketchat import LagoonNotificationRocketChatProvider

        mock_client = Mock()
        mock_client.add_notification_rocketchat.return_value = sample_notification_rocketchat
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationRocketChatProvider()

        inputs = {
            "name": "team-chat",
            "webhook": "https://rocketchat.example.com/hooks/xxx/yyy",
            "channel": "#alerts",
        }

        result = provider.create(inputs)

        assert result.id == "team-chat"
        assert result.outs["name"] == "team-chat"
        assert result.outs["webhook"] == "https://rocketchat.example.com/hooks/xxx/yyy"
        assert result.outs["channel"] == "#alerts"
        assert result.outs["lagoon_id"] == 2

        mock_client.add_notification_rocketchat.assert_called_once_with(
            name="team-chat",
            webhook="https://rocketchat.example.com/hooks/xxx/yyy",
            channel="#alerts",
        )


class TestLagoonNotificationRocketChatProviderUpdate:
    """Tests for LagoonNotificationRocketChatProvider.update method."""

    @patch("pulumi_lagoon.notification_rocketchat.LagoonConfig")
    def test_update_rocketchat_notification_webhook(
        self, mock_config_class, sample_notification_rocketchat
    ):
        """Test updating a RocketChat notification's webhook."""
        from pulumi_lagoon.notification_rocketchat import LagoonNotificationRocketChatProvider

        updated_notification = sample_notification_rocketchat.copy()
        updated_notification["webhook"] = "https://rocketchat.example.com/hooks/new/webhook"

        mock_client = Mock()
        mock_client.update_notification_rocketchat.return_value = updated_notification
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationRocketChatProvider()

        old_inputs = {
            "name": "team-chat",
            "webhook": "https://rocketchat.example.com/hooks/xxx/yyy",
            "channel": "#alerts",
        }

        new_inputs = {
            "name": "team-chat",
            "webhook": "https://rocketchat.example.com/hooks/new/webhook",
            "channel": "#alerts",
        }

        result = provider.update("team-chat", old_inputs, new_inputs)

        assert result.outs["webhook"] == "https://rocketchat.example.com/hooks/new/webhook"
        mock_client.update_notification_rocketchat.assert_called_once()


class TestLagoonNotificationRocketChatProviderDelete:
    """Tests for LagoonNotificationRocketChatProvider.delete method."""

    @patch("pulumi_lagoon.notification_rocketchat.LagoonConfig")
    def test_delete_rocketchat_notification(self, mock_config_class):
        """Test deleting a RocketChat notification."""
        from pulumi_lagoon.notification_rocketchat import LagoonNotificationRocketChatProvider

        mock_client = Mock()
        mock_client.delete_notification_rocketchat.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationRocketChatProvider()

        props = {
            "name": "team-chat",
            "webhook": "https://rocketchat.example.com/hooks/xxx/yyy",
            "channel": "#alerts",
        }

        provider.delete("team-chat", props)

        mock_client.delete_notification_rocketchat.assert_called_once_with(name="team-chat")


class TestLagoonNotificationRocketChatProviderRead:
    """Tests for LagoonNotificationRocketChatProvider.read method."""

    @patch("pulumi_lagoon.notification_rocketchat.LagoonConfig")
    def test_read_rocketchat_notification_exists(
        self, mock_config_class, sample_notification_rocketchat
    ):
        """Test reading an existing RocketChat notification."""
        from pulumi_lagoon.notification_rocketchat import LagoonNotificationRocketChatProvider

        mock_client = Mock()
        mock_client.get_notification_rocketchat_by_name.return_value = (
            sample_notification_rocketchat
        )
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationRocketChatProvider()

        props = {"name": "team-chat"}

        result = provider.read("team-chat", props)

        assert result.outs["name"] == "team-chat"
        assert result.outs["lagoon_id"] == 2
        mock_client.get_notification_rocketchat_by_name.assert_called_once_with("team-chat")

    @patch("pulumi_lagoon.notification_rocketchat.LagoonConfig")
    def test_read_rocketchat_notification_not_found(self, mock_config_class):
        """Test reading a RocketChat notification that doesn't exist."""
        from pulumi_lagoon.notification_rocketchat import LagoonNotificationRocketChatProvider

        mock_client = Mock()
        mock_client.get_notification_rocketchat_by_name.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationRocketChatProvider()

        props = {"name": "deleted-notification"}

        result = provider.read("deleted-notification", props)

        assert result is None

    @patch("pulumi_lagoon.notification_rocketchat.LagoonConfig")
    def test_read_import_scenario(self, mock_config_class, sample_notification_rocketchat):
        """Test read() during import scenario."""
        from pulumi_lagoon.notification_rocketchat import LagoonNotificationRocketChatProvider

        mock_client = Mock()
        mock_client.get_notification_rocketchat_by_name.return_value = (
            sample_notification_rocketchat
        )
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationRocketChatProvider()

        # Import scenario: ID is the name, props is empty
        result = provider.read("team-chat", {})

        assert result is not None
        assert result.outs["name"] == "team-chat"


class TestLagoonNotificationRocketChatProviderValidation:
    """Tests for input validation in LagoonNotificationRocketChatProvider."""

    @patch("pulumi_lagoon.notification_rocketchat.LagoonConfig")
    def test_create_invalid_name(self, mock_config_class):
        """Test that invalid notification names are rejected."""
        from pulumi_lagoon.notification_rocketchat import LagoonNotificationRocketChatProvider

        provider = LagoonNotificationRocketChatProvider()

        inputs = {
            "name": "123-invalid",  # Starts with number
            "webhook": "https://rocketchat.example.com/hooks/xxx/yyy",
            "channel": "#alerts",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.notification_rocketchat.LagoonConfig")
    def test_create_invalid_webhook(self, mock_config_class):
        """Test that invalid webhooks are rejected."""
        from pulumi_lagoon.notification_rocketchat import LagoonNotificationRocketChatProvider

        provider = LagoonNotificationRocketChatProvider()

        inputs = {
            "name": "team-chat",
            "webhook": "not-a-valid-url",
            "channel": "#alerts",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "webhook" in str(exc.value).lower()


class TestLagoonNotificationRocketChatArgs:
    """Tests for LagoonNotificationRocketChatArgs dataclass."""

    def test_args_creation(self):
        """Test creating args for RocketChat notification."""
        from pulumi_lagoon.notification_rocketchat import LagoonNotificationRocketChatArgs

        args = LagoonNotificationRocketChatArgs(
            name="team-chat",
            webhook="https://rocketchat.example.com/hooks/xxx/yyy",
            channel="#alerts",
        )

        assert args.name == "team-chat"
        assert args.webhook == "https://rocketchat.example.com/hooks/xxx/yyy"
        assert args.channel == "#alerts"
