"""Unit tests for LagoonNotificationSlack provider."""

from unittest.mock import Mock, patch

import pytest

from pulumi_lagoon.exceptions import LagoonValidationError


class TestLagoonNotificationSlackProviderCreate:
    """Tests for LagoonNotificationSlackProvider.create method."""

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_create_slack_notification(self, mock_config_class, sample_notification_slack):
        """Test creating a Slack notification."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        mock_client = Mock()
        mock_client.add_notification_slack.return_value = sample_notification_slack
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationSlackProvider()

        inputs = {
            "name": "deploy-alerts",
            "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
            "channel": "#deployments",
        }

        result = provider.create(inputs)

        assert result.id == "deploy-alerts"
        assert result.outs["name"] == "deploy-alerts"
        assert result.outs["webhook"] == "https://hooks.slack.com/services/xxx/yyy/zzz"
        assert result.outs["channel"] == "#deployments"
        assert result.outs["lagoon_id"] == 1

        mock_client.add_notification_slack.assert_called_once_with(
            name="deploy-alerts",
            webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
            channel="#deployments",
        )

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_create_slack_notification_with_channel_no_hash(
        self, mock_config_class, sample_notification_slack
    ):
        """Test creating a Slack notification with channel name without #."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        sample_notification_slack["channel"] = "deployments"

        mock_client = Mock()
        mock_client.add_notification_slack.return_value = sample_notification_slack
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationSlackProvider()

        inputs = {
            "name": "deploy-alerts",
            "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
            "channel": "deployments",
        }

        result = provider.create(inputs)
        assert result.outs["channel"] == "deployments"


class TestLagoonNotificationSlackProviderUpdate:
    """Tests for LagoonNotificationSlackProvider.update method."""

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_update_slack_notification_webhook(self, mock_config_class, sample_notification_slack):
        """Test updating a Slack notification's webhook."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        updated_notification = sample_notification_slack.copy()
        updated_notification["webhook"] = "https://hooks.slack.com/services/new/webhook/url"

        mock_client = Mock()
        mock_client.update_notification_slack.return_value = updated_notification
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationSlackProvider()

        old_inputs = {
            "name": "deploy-alerts",
            "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
            "channel": "#deployments",
        }

        new_inputs = {
            "name": "deploy-alerts",
            "webhook": "https://hooks.slack.com/services/new/webhook/url",
            "channel": "#deployments",
        }

        result = provider.update("deploy-alerts", old_inputs, new_inputs)

        assert result.outs["webhook"] == "https://hooks.slack.com/services/new/webhook/url"
        mock_client.update_notification_slack.assert_called_once()

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_update_slack_notification_channel(self, mock_config_class, sample_notification_slack):
        """Test updating a Slack notification's channel."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        updated_notification = sample_notification_slack.copy()
        updated_notification["channel"] = "#new-channel"

        mock_client = Mock()
        mock_client.update_notification_slack.return_value = updated_notification
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationSlackProvider()

        old_inputs = {
            "name": "deploy-alerts",
            "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
            "channel": "#deployments",
        }

        new_inputs = {
            "name": "deploy-alerts",
            "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
            "channel": "#new-channel",
        }

        result = provider.update("deploy-alerts", old_inputs, new_inputs)

        assert result.outs["channel"] == "#new-channel"


class TestLagoonNotificationSlackProviderDelete:
    """Tests for LagoonNotificationSlackProvider.delete method."""

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_delete_slack_notification(self, mock_config_class):
        """Test deleting a Slack notification."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        mock_client = Mock()
        mock_client.delete_notification_slack.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationSlackProvider()

        props = {
            "name": "deploy-alerts",
            "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
            "channel": "#deployments",
        }

        provider.delete("deploy-alerts", props)

        mock_client.delete_notification_slack.assert_called_once_with(name="deploy-alerts")


class TestLagoonNotificationSlackProviderRead:
    """Tests for LagoonNotificationSlackProvider.read method."""

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_read_slack_notification_exists(self, mock_config_class, sample_notification_slack):
        """Test reading an existing Slack notification."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        mock_client = Mock()
        mock_client.get_notification_slack_by_name.return_value = sample_notification_slack
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationSlackProvider()

        props = {"name": "deploy-alerts"}

        result = provider.read("deploy-alerts", props)

        assert result.outs["name"] == "deploy-alerts"
        assert result.outs["lagoon_id"] == 1
        mock_client.get_notification_slack_by_name.assert_called_once_with("deploy-alerts")

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_read_slack_notification_not_found(self, mock_config_class):
        """Test reading a Slack notification that doesn't exist."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        mock_client = Mock()
        mock_client.get_notification_slack_by_name.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationSlackProvider()

        props = {"name": "deleted-notification"}

        result = provider.read("deleted-notification", props)

        assert result is None

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_read_import_scenario(self, mock_config_class, sample_notification_slack):
        """Test read() during import scenario."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        mock_client = Mock()
        mock_client.get_notification_slack_by_name.return_value = sample_notification_slack
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationSlackProvider()

        # Import scenario: ID is the name, props is empty
        result = provider.read("deploy-alerts", {})

        assert result is not None
        assert result.outs["name"] == "deploy-alerts"
        mock_client.get_notification_slack_by_name.assert_called_once_with("deploy-alerts")


class TestLagoonNotificationSlackProviderValidation:
    """Tests for input validation in LagoonNotificationSlackProvider."""

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_create_invalid_name(self, mock_config_class):
        """Test that invalid notification names are rejected."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        provider = LagoonNotificationSlackProvider()

        inputs = {
            "name": "123-invalid",  # Starts with number
            "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
            "channel": "#deployments",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_create_invalid_webhook_http(self, mock_config_class):
        """Test that HTTP webhooks are rejected (must be HTTPS)."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        provider = LagoonNotificationSlackProvider()

        inputs = {
            "name": "deploy-alerts",
            "webhook": "http://hooks.slack.com/services/xxx/yyy/zzz",  # HTTP not allowed
            "channel": "#deployments",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "webhook" in str(exc.value).lower()

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_create_empty_channel(self, mock_config_class):
        """Test that empty channel is rejected."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        provider = LagoonNotificationSlackProvider()

        inputs = {
            "name": "deploy-alerts",
            "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
            "channel": "",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "channel" in str(exc.value).lower()

    @patch("pulumi_lagoon.notification_slack.LagoonConfig")
    def test_create_name_too_long(self, mock_config_class):
        """Test that names over 100 characters are rejected."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackProvider

        provider = LagoonNotificationSlackProvider()

        inputs = {
            "name": "a" * 101,  # 101 characters
            "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
            "channel": "#deployments",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "100" in str(exc.value)


class TestLagoonNotificationSlackArgs:
    """Tests for LagoonNotificationSlackArgs dataclass."""

    def test_args_creation(self):
        """Test creating args for Slack notification."""
        from pulumi_lagoon.notification_slack import LagoonNotificationSlackArgs

        args = LagoonNotificationSlackArgs(
            name="deploy-alerts",
            webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
            channel="#deployments",
        )

        assert args.name == "deploy-alerts"
        assert args.webhook == "https://hooks.slack.com/services/xxx/yyy/zzz"
        assert args.channel == "#deployments"
