"""Lagoon RocketChat Notification resource - Dynamic provider for managing Lagoon RocketChat notifications."""

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
class LagoonNotificationRocketChatArgs:
    """Arguments for creating a Lagoon RocketChat notification."""

    name: pulumi.Input[str]
    """Notification name."""

    webhook: pulumi.Input[str]
    """RocketChat webhook URL."""

    channel: pulumi.Input[str]
    """RocketChat channel (e.g., '#alerts')."""


class LagoonNotificationRocketChatProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon RocketChat notifications."""

    def _get_client(self):
        """Get configured Lagoon API client."""
        config = LagoonConfig()
        return config.get_client()

    def create(self, inputs):
        """Create a new Lagoon RocketChat notification."""
        # Input validation (fail fast)
        validate_notification_name(inputs.get("name"))
        validate_webhook_url(inputs.get("webhook"))
        validate_channel_name(inputs.get("channel"))

        client = self._get_client()

        # Create notification via API
        result = client.add_notification_rocketchat(
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
        """Update an existing Lagoon RocketChat notification."""
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
        result = client.update_notification_rocketchat(
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
        """Delete a Lagoon RocketChat notification."""
        client = self._get_client()

        # Delete via API
        client.delete_notification_rocketchat(name=props["name"])

    def read(self, id, props):
        """Read/refresh a Lagoon RocketChat notification from API.

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
        result = client.get_notification_rocketchat_by_name(name)

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


class LagoonNotificationRocketChat(
    dynamic.Resource, module="lagoon", name="NotificationRocketChat"
):
    """
    A Lagoon RocketChat notification resource.

    Manages a RocketChat notification in Lagoon that can be linked to projects
    to receive deployment and other notifications.

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Create a RocketChat notification
    rocketchat_alerts = lagoon.LagoonNotificationRocketChat("team-chat",
        lagoon.LagoonNotificationRocketChatArgs(
            name="team-chat",
            webhook="https://rocketchat.example.com/hooks/xxx/yyy",
            channel="#alerts",
        )
    )

    # Export the notification name
    pulumi.export("rocketchat_notification_name", rocketchat_alerts.name)
    ```

    ## Import

    RocketChat notifications can be imported using the notification name:

    ```bash
    pulumi import lagoon:index:NotificationRocketChat team-chat team-chat
    ```
    """

    # Output properties
    lagoon_id: pulumi.Output[int]
    """The Lagoon internal ID for this notification."""

    name: pulumi.Output[str]
    """The notification name."""

    webhook: pulumi.Output[str]
    """The RocketChat webhook URL."""

    channel: pulumi.Output[str]
    """The RocketChat channel."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonNotificationRocketChatArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonNotificationRocketChat resource.

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

        super().__init__(LagoonNotificationRocketChatProvider(), resource_name, inputs, opts)
