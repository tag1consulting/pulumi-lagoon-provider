"""Lagoon Project Notification resource - Dynamic provider for managing Lagoon project notification associations."""

from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi.dynamic as dynamic

from .config import LagoonConfig
from .import_utils import ImportIdParser
from .validators import (
    validate_notification_name,
    validate_notification_type,
    validate_project_name,
)


@dataclass
class LagoonProjectNotificationArgs:
    """Arguments for linking a notification to a Lagoon project."""

    project_name: pulumi.Input[str]
    """Project name to link the notification to."""

    notification_type: pulumi.Input[str]
    """Type of notification: 'slack', 'rocketchat', 'email', or 'microsoftteams'."""

    notification_name: pulumi.Input[str]
    """Name of the notification to link."""


class LagoonProjectNotificationProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon project notification associations."""

    def _get_client(self):
        """Get configured Lagoon API client."""
        config = LagoonConfig()
        return config.get_client()

    def _create_composite_id(
        self, project_name: str, notification_type: str, notification_name: str
    ) -> str:
        """Create a unique composite ID for the association."""
        return f"{project_name}:{notification_type}:{notification_name}"

    def create(self, inputs):
        """Create a new project notification association."""
        # Input validation (fail fast)
        validate_project_name(inputs.get("project_name"))
        notification_type = validate_notification_type(inputs.get("notification_type"))
        validate_notification_name(inputs.get("notification_name"))

        client = self._get_client()

        # Create association via API
        result = client.add_notification_to_project(
            project=inputs["project_name"],
            notification_type=notification_type,
            notification_name=inputs["notification_name"],
        )

        # Create composite ID
        association_id = self._create_composite_id(
            inputs["project_name"],
            notification_type,
            inputs["notification_name"],
        )

        outs = {
            "project_name": inputs["project_name"],
            "notification_type": notification_type,
            "notification_name": inputs["notification_name"],
            "project_id": result.get("id"),
        }

        return dynamic.CreateResult(id_=association_id, outs=outs)

    def update(self, id, old_inputs, new_inputs):
        """Update an existing project notification association.

        Note: Lagoon doesn't support updating associations directly.
        If any field changes, we delete the old association and create a new one.
        """
        # Input validation for new values (fail fast)
        validate_project_name(new_inputs.get("project_name"))
        new_notification_type = validate_notification_type(new_inputs.get("notification_type"))
        validate_notification_name(new_inputs.get("notification_name"))

        client = self._get_client()

        # Parse old values
        old_notification_type = old_inputs.get("notification_type", "").lower()

        # Remove old association
        client.remove_notification_from_project(
            project=old_inputs["project_name"],
            notification_type=old_notification_type,
            notification_name=old_inputs["notification_name"],
        )

        # Create new association
        result = client.add_notification_to_project(
            project=new_inputs["project_name"],
            notification_type=new_notification_type,
            notification_name=new_inputs["notification_name"],
        )

        outs = {
            "project_name": new_inputs["project_name"],
            "notification_type": new_notification_type,
            "notification_name": new_inputs["notification_name"],
            "project_id": result.get("id"),
        }

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a project notification association."""
        client = self._get_client()

        notification_type = props.get("notification_type", "").lower()

        # Remove association via API
        client.remove_notification_from_project(
            project=props["project_name"],
            notification_type=notification_type,
            notification_name=props["notification_name"],
        )

    def read(self, id, props):
        """Read/refresh a project notification association from API.

        Supports both refresh (props available) and import (props empty).
        For import, the ID format is: {project_name}:{notification_type}:{notification_name}
        """
        client = self._get_client()

        # Detect import vs refresh scenario
        if ImportIdParser.is_import_scenario(
            id, props, ["project_name", "notification_type", "notification_name"]
        ):
            # Import: parse composite ID
            project_name, notification_type, notification_name = (
                ImportIdParser.parse_project_notification_id(id)
            )
        else:
            # Refresh: use props from state
            project_name = props["project_name"]
            notification_type = props.get("notification_type", "").lower()
            notification_name = props["notification_name"]

        # Check if association still exists
        exists = client.check_project_notification_exists(
            project_name=project_name,
            notification_type=notification_type,
            notification_name=notification_name,
        )

        if not exists:
            # Association no longer exists
            return None

        # Get project info to get project_id
        project = client.get_project_by_name(project_name)

        outs = {
            "project_name": project_name,
            "notification_type": notification_type,
            "notification_name": notification_name,
            "project_id": project.get("id") if project else None,
        }

        return dynamic.ReadResult(id_=id, outs=outs)


class LagoonProjectNotification(dynamic.Resource, module="lagoon", name="ProjectNotification"):
    """
    A Lagoon project notification association resource.

    Links a notification (Slack, RocketChat, Email, or Microsoft Teams) to a project.
    This allows the project to receive notifications for deployments, builds, and other events.

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Create a project first
    project = lagoon.LagoonProject("my-site",
        lagoon.LagoonProjectArgs(
            name="my-site",
            git_url="git@github.com:example/my-site.git",
            deploytarget_id=1,
        )
    )

    # Create a Slack notification
    slack_alerts = lagoon.LagoonNotificationSlack("deploy-alerts",
        lagoon.LagoonNotificationSlackArgs(
            name="deploy-alerts",
            webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
            channel="#deployments",
        )
    )

    # Link the notification to the project
    project_notification = lagoon.LagoonProjectNotification("project-slack",
        lagoon.LagoonProjectNotificationArgs(
            project_name=project.name,
            notification_type="slack",
            notification_name=slack_alerts.name,
        ),
        opts=pulumi.ResourceOptions(depends_on=[project, slack_alerts])
    )

    # Export the association details
    pulumi.export("project_notification_type", project_notification.notification_type)
    ```

    ## Import

    Project notification associations can be imported using the format:
    `{project_name}:{notification_type}:{notification_name}`

    ```bash
    pulumi import lagoon:index:ProjectNotification project-slack my-project:slack:deploy-alerts
    ```
    """

    # Output properties
    project_name: pulumi.Output[str]
    """The project name."""

    notification_type: pulumi.Output[str]
    """The notification type (slack, rocketchat, email, microsoftteams)."""

    notification_name: pulumi.Output[str]
    """The notification name."""

    project_id: pulumi.Output[Optional[int]]
    """The Lagoon internal project ID."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonProjectNotificationArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonProjectNotification resource.

        Args:
            resource_name: The Pulumi resource name
            args: The association configuration arguments
            opts: Optional resource options
        """
        inputs = {
            "project_name": args.project_name,
            "notification_type": args.notification_type,
            "notification_name": args.notification_name,
            # Outputs (set by provider)
            "project_id": None,
        }

        super().__init__(LagoonProjectNotificationProvider(), resource_name, inputs, opts)
