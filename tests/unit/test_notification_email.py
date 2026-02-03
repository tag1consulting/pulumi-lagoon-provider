"""Unit tests for LagoonNotificationEmail provider."""

from unittest.mock import Mock, patch

import pytest

from pulumi_lagoon.exceptions import LagoonValidationError


class TestLagoonNotificationEmailProviderCreate:
    """Tests for LagoonNotificationEmailProvider.create method."""

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_create_email_notification(self, mock_config_class, sample_notification_email):
        """Test creating an Email notification."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        mock_client = Mock()
        mock_client.add_notification_email.return_value = sample_notification_email
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationEmailProvider()

        inputs = {
            "name": "ops-team",
            "email_address": "ops@example.com",
        }

        result = provider.create(inputs)

        assert result.id == "ops-team"
        assert result.outs["name"] == "ops-team"
        assert result.outs["email_address"] == "ops@example.com"
        assert result.outs["lagoon_id"] == 3

        mock_client.add_notification_email.assert_called_once_with(
            name="ops-team",
            email_address="ops@example.com",
        )

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_create_email_notification_complex_address(
        self, mock_config_class, sample_notification_email
    ):
        """Test creating an Email notification with complex email address."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        sample_notification_email["emailAddress"] = "ops+alerts@example.co.uk"

        mock_client = Mock()
        mock_client.add_notification_email.return_value = sample_notification_email
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationEmailProvider()

        inputs = {
            "name": "ops-team",
            "email_address": "ops+alerts@example.co.uk",
        }

        result = provider.create(inputs)
        assert result.outs["email_address"] == "ops+alerts@example.co.uk"


class TestLagoonNotificationEmailProviderUpdate:
    """Tests for LagoonNotificationEmailProvider.update method."""

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_update_email_notification_address(self, mock_config_class, sample_notification_email):
        """Test updating an Email notification's address."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        updated_notification = sample_notification_email.copy()
        updated_notification["emailAddress"] = "new-ops@example.com"

        mock_client = Mock()
        mock_client.update_notification_email.return_value = updated_notification
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationEmailProvider()

        old_inputs = {
            "name": "ops-team",
            "email_address": "ops@example.com",
        }

        new_inputs = {
            "name": "ops-team",
            "email_address": "new-ops@example.com",
        }

        result = provider.update("ops-team", old_inputs, new_inputs)

        assert result.outs["email_address"] == "new-ops@example.com"
        mock_client.update_notification_email.assert_called_once()


class TestLagoonNotificationEmailProviderDelete:
    """Tests for LagoonNotificationEmailProvider.delete method."""

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_delete_email_notification(self, mock_config_class):
        """Test deleting an Email notification."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        mock_client = Mock()
        mock_client.delete_notification_email.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationEmailProvider()

        props = {
            "name": "ops-team",
            "email_address": "ops@example.com",
        }

        provider.delete("ops-team", props)

        mock_client.delete_notification_email.assert_called_once_with(name="ops-team")


class TestLagoonNotificationEmailProviderRead:
    """Tests for LagoonNotificationEmailProvider.read method."""

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_read_email_notification_exists(self, mock_config_class, sample_notification_email):
        """Test reading an existing Email notification."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        mock_client = Mock()
        mock_client.get_notification_email_by_name.return_value = sample_notification_email
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationEmailProvider()

        props = {"name": "ops-team"}

        result = provider.read("ops-team", props)

        assert result.outs["name"] == "ops-team"
        assert result.outs["email_address"] == "ops@example.com"
        assert result.outs["lagoon_id"] == 3
        mock_client.get_notification_email_by_name.assert_called_once_with("ops-team")

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_read_email_notification_not_found(self, mock_config_class):
        """Test reading an Email notification that doesn't exist."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        mock_client = Mock()
        mock_client.get_notification_email_by_name.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationEmailProvider()

        props = {"name": "deleted-notification"}

        result = provider.read("deleted-notification", props)

        assert result is None

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_read_import_scenario(self, mock_config_class, sample_notification_email):
        """Test read() during import scenario."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        mock_client = Mock()
        mock_client.get_notification_email_by_name.return_value = sample_notification_email
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonNotificationEmailProvider()

        # Import scenario: ID is the name, props is empty
        result = provider.read("ops-team", {})

        assert result is not None
        assert result.outs["name"] == "ops-team"


class TestLagoonNotificationEmailProviderValidation:
    """Tests for input validation in LagoonNotificationEmailProvider."""

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_create_invalid_name(self, mock_config_class):
        """Test that invalid notification names are rejected."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        provider = LagoonNotificationEmailProvider()

        inputs = {
            "name": "123-invalid",  # Starts with number
            "email_address": "ops@example.com",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_create_invalid_email_no_at(self, mock_config_class):
        """Test that email without @ is rejected."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        provider = LagoonNotificationEmailProvider()

        inputs = {
            "name": "ops-team",
            "email_address": "not-an-email",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "email" in str(exc.value).lower()

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_create_invalid_email_no_domain(self, mock_config_class):
        """Test that email without domain is rejected."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        provider = LagoonNotificationEmailProvider()

        inputs = {
            "name": "ops-team",
            "email_address": "ops@",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "email" in str(exc.value).lower()

    @patch("pulumi_lagoon.notification_email.LagoonConfig")
    def test_create_empty_email(self, mock_config_class):
        """Test that empty email is rejected."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailProvider

        provider = LagoonNotificationEmailProvider()

        inputs = {
            "name": "ops-team",
            "email_address": "",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "email_address" in str(exc.value).lower()


class TestLagoonNotificationEmailArgs:
    """Tests for LagoonNotificationEmailArgs dataclass."""

    def test_args_creation(self):
        """Test creating args for Email notification."""
        from pulumi_lagoon.notification_email import LagoonNotificationEmailArgs

        args = LagoonNotificationEmailArgs(
            name="ops-team",
            email_address="ops@example.com",
        )

        assert args.name == "ops-team"
        assert args.email_address == "ops@example.com"
