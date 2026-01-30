"""Lagoon Microsoft Teams Notification resource - Dynamic provider for managing Lagoon MS Teams notifications."""

from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi.dynamic as dynamic

from .config import LagoonConfig
from .import_utils import ImportIdParser
from .validators import (
    validate_notification_name,
    validate_webhook_url,
)


@dataclass
class LagoonNotificationMicrosoftTeamsArgs:
    """Arguments for creating a Lagoon Microsoft Teams notification."""

    name: pulumi.Input[str]
    """Notification name."""

    webhook: pulumi.Input[str]
    """Microsoft Teams webhook URL."""


class LagoonNotificationMicrosoftTeamsProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon Microsoft Teams notifications."""

    def _get_client(self):
        """Get configured Lagoon API client."""
        config = LagoonConfig()
        return config.get_client()

    def create(self, inputs):
        """Create a new Lagoon Microsoft Teams notification."""
        # Input validation (fail fast)
        validate_notification_name(inputs.get("name"))
        validate_webhook_url(inputs.get("webhook"))

        client = self._get_client()

        # Create notification via API
        result = client.add_notification_microsoftteams(
            name=inputs["name"],
            webhook=inputs["webhook"],
        )

        # Use notification name as ID (names are unique in Lagoon)
        notification_id = inputs["name"]

        outs = {
            "lagoon_id": result.get("id"),
            "name": result["name"],
            "webhook": result["webhook"],
        }

        return dynamic.CreateResult(id_=notification_id, outs=outs)

    def update(self, id, old_inputs, new_inputs):
        """Update an existing Lagoon Microsoft Teams notification."""
        # Input validation for new values (fail fast)
        validate_notification_name(new_inputs.get("name"))
        validate_webhook_url(new_inputs.get("webhook"))

        client = self._get_client()

        # Build update kwargs
        update_kwargs = {}
        if new_inputs.get("webhook") != old_inputs.get("webhook"):
            update_kwargs["webhook"] = new_inputs["webhook"]

        # Update notification via API (use old name to identify)
        result = client.update_notification_microsoftteams(
            name=old_inputs["name"],
            **update_kwargs,
        )

        outs = {
            "lagoon_id": result.get("id"),
            "name": result["name"],
            "webhook": result["webhook"],
        }

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a Lagoon Microsoft Teams notification."""
        client = self._get_client()

        # Delete via API
        client.delete_notification_microsoftteams(name=props["name"])

    def read(self, id, props):
        """Read/refresh a Lagoon Microsoft Teams notification from API.

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
        result = client.get_notification_microsoftteams_by_name(name)

        if not result:
            # Notification no longer exists
            return None

        outs = {
            "lagoon_id": result.get("id"),
            "name": result["name"],
            "webhook": result["webhook"],
        }

        return dynamic.ReadResult(id_=id, outs=outs)


class LagoonNotificationMicrosoftTeams(
    dynamic.Resource, module="lagoon", name="NotificationMicrosoftTeams"
):
    """
    A Lagoon Microsoft Teams notification resource.

    Manages a Microsoft Teams notification in Lagoon that can be linked to projects
    to receive deployment and other notifications.

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Create a Microsoft Teams notification
    teams_alerts = lagoon.LagoonNotificationMicrosoftTeams("teams-alerts",
        lagoon.LagoonNotificationMicrosoftTeamsArgs(
            name="teams-alerts",
            webhook="https://outlook.office.com/webhook/xxx/yyy/zzz",
        )
    )

    # Export the notification name
    pulumi.export("teams_notification_name", teams_alerts.name)
    ```

    ## Import

    Microsoft Teams notifications can be imported using the notification name:

    ```bash
    pulumi import lagoon:index:NotificationMicrosoftTeams teams-alerts teams-alerts
    ```
    """

    # Output properties
    lagoon_id: pulumi.Output[int]
    """The Lagoon internal ID for this notification."""

    name: pulumi.Output[str]
    """The notification name."""

    webhook: pulumi.Output[str]
    """The Microsoft Teams webhook URL."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonNotificationMicrosoftTeamsArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonNotificationMicrosoftTeams resource.

        Args:
            resource_name: The Pulumi resource name
            args: The notification configuration arguments
            opts: Optional resource options
        """
        inputs = {
            "name": args.name,
            "webhook": args.webhook,
            # Outputs (set by provider)
            "lagoon_id": None,
        }

        super().__init__(LagoonNotificationMicrosoftTeamsProvider(), resource_name, inputs, opts)
