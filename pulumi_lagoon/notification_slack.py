"""Lagoon Slack Notification resource - Dynamic provider for managing Lagoon Slack notifications."""

from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi.dynamic as dynamic

from .config import LagoonConfig
from .import_utils import ImportIdParser
from .validators import (
    validate_channel_name,
    validate_notification_name,
    validate_webhook_url,
)


@dataclass
class LagoonNotificationSlackArgs:
    """Arguments for creating a Lagoon Slack notification."""

    name: pulumi.Input[str]
    """Notification name."""

    webhook: pulumi.Input[str]
    """Slack webhook URL."""

    channel: pulumi.Input[str]
    """Slack channel (e.g., '#alerts')."""


class LagoonNotificationSlackProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon Slack notifications."""

    def _get_client(self):
        """Get configured Lagoon API client."""
        config = LagoonConfig()
        return config.get_client()

    def create(self, inputs):
        """Create a new Lagoon Slack notification."""
        # Input validation (fail fast)
        validate_notification_name(inputs.get("name"))
        validate_webhook_url(inputs.get("webhook"))
        validate_channel_name(inputs.get("channel"))

        client = self._get_client()

        # Create notification via API
        result = client.add_notification_slack(
            name=inputs["name"],
            webhook=inputs["webhook"],
            channel=inputs["channel"],
        )

        # Use notification name as ID (names are unique in Lagoon)
        notification_id = inputs["name"]

        outs = {
            "lagoon_id": result.get("id"),
            "name": result["name"],
            "webhook": result["webhook"],
            "channel": result["channel"],
        }

        return dynamic.CreateResult(id_=notification_id, outs=outs)

    def update(self, id, old_inputs, new_inputs):
        """Update an existing Lagoon Slack notification."""
        # Input validation for new values (fail fast)
        validate_notification_name(new_inputs.get("name"))
        validate_webhook_url(new_inputs.get("webhook"))
        validate_channel_name(new_inputs.get("channel"))

        client = self._get_client()

        # Build update kwargs
        update_kwargs = {}
        if new_inputs.get("webhook") != old_inputs.get("webhook"):
            update_kwargs["webhook"] = new_inputs["webhook"]
        if new_inputs.get("channel") != old_inputs.get("channel"):
            update_kwargs["channel"] = new_inputs["channel"]

        # Update notification via API (use old name to identify)
        result = client.update_notification_slack(
            name=old_inputs["name"],
            **update_kwargs,
        )

        outs = {
            "lagoon_id": result.get("id"),
            "name": result["name"],
            "webhook": result["webhook"],
            "channel": result["channel"],
        }

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a Lagoon Slack notification."""
        client = self._get_client()

        # Delete via API
        client.delete_notification_slack(name=props["name"])

    def read(self, id, props):
        """Read/refresh a Lagoon Slack notification from API.

        Supports both refresh (props available) and import (props empty).
        For import, the ID format is: {name}
        """
        client = self._get_client()

        # Detect import vs refresh scenario
        if ImportIdParser.is_import_scenario(id, props, ["name"]):
            # Import: use ID as name
            name = id
        else:
            # Refresh: use props from state
            name = props["name"]

        # Query current state
        result = client.get_notification_slack_by_name(name)

        if not result:
            # Notification no longer exists
            return None

        outs = {
            "lagoon_id": result.get("id"),
            "name": result["name"],
            "webhook": result["webhook"],
            "channel": result["channel"],
        }

        return dynamic.ReadResult(id_=id, outs=outs)


class LagoonNotificationSlack(dynamic.Resource, module="lagoon", name="NotificationSlack"):
    """
    A Lagoon Slack notification resource.

    Manages a Slack notification in Lagoon that can be linked to projects
    to receive deployment and other notifications.

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Create a Slack notification
    slack_alerts = lagoon.LagoonNotificationSlack("deploy-alerts",
        lagoon.LagoonNotificationSlackArgs(
            name="deploy-alerts",
            webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
            channel="#deployments",
        )
    )

    # Export the notification name
    pulumi.export("slack_notification_name", slack_alerts.name)
    ```

    ## Import

    Slack notifications can be imported using the notification name:

    ```bash
    pulumi import lagoon:index:NotificationSlack deploy-alerts deploy-alerts
    ```
    """

    # Output properties
    lagoon_id: pulumi.Output[int]
    """The Lagoon internal ID for this notification."""

    name: pulumi.Output[str]
    """The notification name."""

    webhook: pulumi.Output[str]
    """The Slack webhook URL."""

    channel: pulumi.Output[str]
    """The Slack channel."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonNotificationSlackArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonNotificationSlack resource.

        Args:
            resource_name: The Pulumi resource name
            args: The notification configuration arguments
            opts: Optional resource options
        """
        inputs = {
            "name": args.name,
            "webhook": args.webhook,
            "channel": args.channel,
            # Outputs (set by provider)
            "lagoon_id": None,
        }

        super().__init__(LagoonNotificationSlackProvider(), resource_name, inputs, opts)
