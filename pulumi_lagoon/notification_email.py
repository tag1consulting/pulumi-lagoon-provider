"""Lagoon Email Notification resource - Dynamic provider for managing Lagoon Email notifications."""

from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi.dynamic as dynamic

from .config import LagoonConfig
from .import_utils import ImportIdParser
from .validators import (
    validate_email_address,
    validate_notification_name,
)


@dataclass
class LagoonNotificationEmailArgs:
    """Arguments for creating a Lagoon Email notification."""

    name: pulumi.Input[str]
    """Notification name."""

    email_address: pulumi.Input[str]
    """Email address to send notifications to."""


class LagoonNotificationEmailProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon Email notifications."""

    def _get_client(self):
        """Get configured Lagoon API client."""
        config = LagoonConfig()
        return config.get_client()

    def create(self, inputs):
        """Create a new Lagoon Email notification."""
        # Input validation (fail fast)
        validate_notification_name(inputs.get("name"))
        validate_email_address(inputs.get("email_address"))

        client = self._get_client()

        # Create notification via API
        result = client.add_notification_email(
            name=inputs["name"],
            email_address=inputs["email_address"],
        )

        # Use notification name as ID (names are unique in Lagoon)
        notification_id = inputs["name"]

        outs = {
            "lagoon_id": result.get("id"),
            "name": result["name"],
            "email_address": result["emailAddress"],
        }

        return dynamic.CreateResult(id_=notification_id, outs=outs)

    def update(self, id, old_inputs, new_inputs):
        """Update an existing Lagoon Email notification."""
        # Input validation for new values (fail fast)
        validate_notification_name(new_inputs.get("name"))
        validate_email_address(new_inputs.get("email_address"))

        client = self._get_client()

        # Build update kwargs
        update_kwargs = {}
        if new_inputs.get("email_address") != old_inputs.get("email_address"):
            update_kwargs["emailAddress"] = new_inputs["email_address"]

        # Update notification via API (use old name to identify)
        result = client.update_notification_email(
            name=old_inputs["name"],
            **update_kwargs,
        )

        outs = {
            "lagoon_id": result.get("id"),
            "name": result["name"],
            "email_address": result["emailAddress"],
        }

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a Lagoon Email notification."""
        client = self._get_client()

        # Delete via API
        client.delete_notification_email(name=props["name"])

    def read(self, id, props):
        """Read/refresh a Lagoon Email notification from API.

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
        result = client.get_notification_email_by_name(name)

        if not result:
            # Notification no longer exists
            return None

        outs = {
            "lagoon_id": result.get("id"),
            "name": result["name"],
            "email_address": result["emailAddress"],
        }

        return dynamic.ReadResult(id_=id, outs=outs)


class LagoonNotificationEmail(dynamic.Resource, module="lagoon", name="NotificationEmail"):
    """
    A Lagoon Email notification resource.

    Manages an Email notification in Lagoon that can be linked to projects
    to receive deployment and other notifications.

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Create an Email notification
    email_ops = lagoon.LagoonNotificationEmail("ops-team",
        lagoon.LagoonNotificationEmailArgs(
            name="ops-team",
            email_address="ops@example.com",
        )
    )

    # Export the notification name
    pulumi.export("email_notification_name", email_ops.name)
    ```

    ## Import

    Email notifications can be imported using the notification name:

    ```bash
    pulumi import lagoon:index:NotificationEmail ops-team ops-team
    ```
    """

    # Output properties
    lagoon_id: pulumi.Output[int]
    """The Lagoon internal ID for this notification."""

    name: pulumi.Output[str]
    """The notification name."""

    email_address: pulumi.Output[str]
    """The email address."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonNotificationEmailArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonNotificationEmail resource.

        Args:
            resource_name: The Pulumi resource name
            args: The notification configuration arguments
            opts: Optional resource options
        """
        inputs = {
            "name": args.name,
            "email_address": args.email_address,
            # Outputs (set by provider)
            "lagoon_id": None,
        }

        super().__init__(LagoonNotificationEmailProvider(), resource_name, inputs, opts)
