"""Lagoon Task resource - Dynamic provider for managing Lagoon advanced task definitions."""

from dataclasses import dataclass
from typing import Any, Dict, List, Optional

import pulumi
import pulumi.dynamic as dynamic

from .client import LagoonAPIError, LagoonConnectionError
from .config import LagoonConfig
from .validators import (
    validate_positive_int,
    validate_required,
    validate_task_command_or_image,
    validate_task_permission,
    validate_task_scope,
    validate_task_type,
)


@dataclass
class LagoonTaskArgs:
    """Arguments for creating a Lagoon advanced task definition."""

    name: pulumi.Input[str]
    """Task definition name."""

    type: pulumi.Input[str]
    """Task type: 'command' or 'image'."""

    service: pulumi.Input[str]
    """Service container name to run the task in."""

    command: Optional[pulumi.Input[str]] = None
    """Command to execute (required if type='command')."""

    image: Optional[pulumi.Input[str]] = None
    """Container image to run (required if type='image')."""

    project_id: Optional[pulumi.Input[int]] = None
    """Project ID (for project-scoped tasks)."""

    environment_id: Optional[pulumi.Input[int]] = None
    """Environment ID (for environment-scoped tasks)."""

    group_name: Optional[pulumi.Input[str]] = None
    """Group name (for group-scoped tasks)."""

    system_wide: Optional[pulumi.Input[bool]] = None
    """If True, task is available system-wide (platform admin only)."""

    description: Optional[pulumi.Input[str]] = None
    """Task description."""

    permission: Optional[pulumi.Input[str]] = None
    """Permission level: 'guest', 'developer', or 'maintainer'."""

    confirmation_text: Optional[pulumi.Input[str]] = None
    """Text to display for user confirmation before running."""

    arguments: Optional[pulumi.Input[List[Dict[str, Any]]]] = None
    """List of argument definitions [{name, display_name, type}]."""


class LagoonTaskProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon advanced task definitions."""

    def _get_client(self):
        """Get configured Lagoon API client."""
        config = LagoonConfig()
        return config.get_client()

    def create(self, inputs):
        """Create a new Lagoon advanced task definition."""
        # Input validation (fail fast)
        validate_required(inputs.get("name"), "name")
        task_type = validate_task_type(inputs.get("type"))
        validate_required(inputs.get("service"), "service")

        # Validate command/image based on type
        command = inputs.get("command")
        image = inputs.get("image")
        validate_task_command_or_image(task_type, command, image)

        # Validate scope (exactly one required)
        project_id = inputs.get("project_id")
        environment_id = inputs.get("environment_id")
        group_name = inputs.get("group_name")
        system_wide = inputs.get("system_wide")

        if project_id is not None:
            project_id = validate_positive_int(project_id, "project_id")
        if environment_id is not None:
            environment_id = validate_positive_int(environment_id, "environment_id")

        validate_task_scope(project_id, environment_id, group_name, system_wide)

        # Validate optional permission
        permission = validate_task_permission(inputs.get("permission"))

        # Validate arguments if provided
        arguments = inputs.get("arguments")
        if arguments:
            from .validators import validate_task_argument_type

            for arg in arguments:
                validate_required(arg.get("name"), "argument name")
                if arg.get("type"):
                    validate_task_argument_type(arg.get("type"))

        client = self._get_client()

        # Create task via API
        result = client.add_advanced_task_definition(
            name=inputs["name"],
            task_type=task_type,
            service=inputs["service"],
            command=command,
            image=image,
            project_id=project_id,
            environment_id=environment_id,
            group_name=group_name,
            system_wide=system_wide,
            description=inputs.get("description"),
            permission=permission,
            confirmation_text=inputs.get("confirmation_text"),
            arguments=arguments,
        )

        task_id = str(result.get("id"))

        # Return outputs
        outs = {
            "id": result.get("id"),
            "name": result.get("name"),
            "type": result.get("type", "").lower(),
            "service": result.get("service"),
            "command": result.get("command"),
            "image": result.get("image"),
            "description": result.get("description"),
            "permission": result.get("permission", "").lower()
            if result.get("permission")
            else None,
            "confirmation_text": result.get("confirmationText"),
            "project_id": result.get("projectId") or project_id,
            "environment_id": result.get("environmentId") or environment_id,
            "group_name": result.get("groupName") or group_name,
            "system_wide": system_wide,
            "arguments": self._normalize_arguments(result.get("advancedTaskDefinitionArguments")),
            "created": result.get("created"),
        }

        return dynamic.CreateResult(id_=task_id, outs=outs)

    def _normalize_arguments(self, api_arguments: Optional[list]) -> Optional[list]:
        """Normalize API arguments to snake_case format."""
        if not api_arguments:
            return None
        return [
            {
                "id": arg.get("id"),
                "name": arg.get("name"),
                "display_name": arg.get("displayName"),
                "type": arg.get("type", "").lower() if arg.get("type") else None,
            }
            for arg in api_arguments
        ]

    def update(self, id, old_inputs, new_inputs):
        """Update an existing Lagoon task definition.

        Lagoon task definitions are updated via delete-recreate pattern.
        """
        # Input validation for new values (fail fast)
        validate_required(new_inputs.get("name"), "name")
        new_task_type = validate_task_type(new_inputs.get("type"))
        validate_required(new_inputs.get("service"), "service")

        # Validate command/image based on type
        new_command = new_inputs.get("command")
        new_image = new_inputs.get("image")
        validate_task_command_or_image(new_task_type, new_command, new_image)

        # Validate scope
        new_project_id = new_inputs.get("project_id")
        new_environment_id = new_inputs.get("environment_id")
        new_group_name = new_inputs.get("group_name")
        new_system_wide = new_inputs.get("system_wide")

        if new_project_id is not None:
            new_project_id = validate_positive_int(new_project_id, "project_id")
        if new_environment_id is not None:
            new_environment_id = validate_positive_int(new_environment_id, "environment_id")

        validate_task_scope(new_project_id, new_environment_id, new_group_name, new_system_wide)

        # Validate optional permission
        new_permission = validate_task_permission(new_inputs.get("permission"))

        # Validate arguments if provided
        new_arguments = new_inputs.get("arguments")
        if new_arguments:
            from .validators import validate_task_argument_type

            for arg in new_arguments:
                validate_required(arg.get("name"), "argument name")
                if arg.get("type"):
                    validate_task_argument_type(arg.get("type"))

        client = self._get_client()

        # Delete old task definition
        old_task_id = int(id)
        try:
            client.delete_advanced_task_definition(old_task_id)
        except LagoonAPIError:
            # Task might not exist or might be already deleted
            # This is acceptable during update - continue with recreation
            pass
        except LagoonConnectionError:
            # Connection errors should propagate - can't safely continue
            raise

        # Create new task definition with updated values
        result = client.add_advanced_task_definition(
            name=new_inputs["name"],
            task_type=new_task_type,
            service=new_inputs["service"],
            command=new_command,
            image=new_image,
            project_id=new_project_id,
            environment_id=new_environment_id,
            group_name=new_group_name,
            system_wide=new_system_wide,
            description=new_inputs.get("description"),
            permission=new_permission,
            confirmation_text=new_inputs.get("confirmation_text"),
            arguments=new_arguments,
        )

        # Return updated outputs
        outs = {
            "id": result.get("id"),
            "name": result.get("name"),
            "type": result.get("type", "").lower(),
            "service": result.get("service"),
            "command": result.get("command"),
            "image": result.get("image"),
            "description": result.get("description"),
            "permission": result.get("permission", "").lower()
            if result.get("permission")
            else None,
            "confirmation_text": result.get("confirmationText"),
            "project_id": result.get("projectId") or new_project_id,
            "environment_id": result.get("environmentId") or new_environment_id,
            "group_name": result.get("groupName") or new_group_name,
            "system_wide": new_system_wide,
            "arguments": self._normalize_arguments(result.get("advancedTaskDefinitionArguments")),
            "created": result.get("created"),
        }

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a Lagoon task definition."""
        client = self._get_client()

        # Ensure ID is an integer
        task_id = int(id)

        client.delete_advanced_task_definition(task_id)

    def read(self, id, props):
        """Read/refresh a Lagoon task definition from API.

        Supports both refresh (props available) and import (props empty).
        For import, the ID format is a numeric task ID.
        """
        from .import_utils import ImportIdParser

        client = self._get_client()

        # Detect import vs refresh scenario
        if ImportIdParser.is_import_scenario(id, props, ["name", "type"]):
            # Import: parse numeric ID
            task_id = ImportIdParser.parse_task_id(id)
        else:
            # Refresh: use ID directly (already numeric string)
            task_id = int(id)

        # Query current state
        result = client.get_advanced_task_definition_by_id(task_id)

        if not result:
            # Task no longer exists
            return None

        # Return current state
        outs = {
            "id": result.get("id"),
            "name": result.get("name"),
            "type": result.get("type", "").lower(),
            "service": result.get("service"),
            "command": result.get("command"),
            "image": result.get("image"),
            "description": result.get("description"),
            "permission": result.get("permission", "").lower()
            if result.get("permission")
            else None,
            "confirmation_text": result.get("confirmationText"),
            "project_id": result.get("projectId"),
            "environment_id": result.get("environmentId"),
            "group_name": result.get("groupName"),
            "system_wide": result.get("groupName") is None
            and result.get("projectId") is None
            and result.get("environmentId") is None,
            "arguments": self._normalize_arguments(result.get("advancedTaskDefinitionArguments")),
            "created": result.get("created"),
        }

        return dynamic.ReadResult(id_=str(task_id), outs=outs)


class LagoonTask(dynamic.Resource, module="lagoon", name="Task"):
    """
    A Lagoon advanced task definition resource.

    Manages an advanced task definition in Lagoon that can be executed on-demand
    within environments. Tasks can execute commands in existing service containers
    or run special container images.

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Create a command-type task (runs in existing service container)
    yarn_audit = lagoon.LagoonTask("yarn-audit",
        name="run-yarn-audit",
        type="command",
        service="node",
        command="yarn audit",
        project_id=project.id,
        permission="developer",
        description="Run yarn security audit",
    )

    # Create an image-type task (runs a special container)
    backup_task = lagoon.LagoonTask("db-backup",
        name="database-backup",
        type="image",
        service="cli",
        image="amazeeio/database-tools:latest",
        project_id=project.id,
        permission="maintainer",
        description="Create database backup",
        confirmation_text="This will create a full database backup. Continue?",
    )

    # Create a task with arguments
    deploy_task = lagoon.LagoonTask("deploy-branch",
        name="deploy-to-branch",
        type="command",
        service="cli",
        command="drush deploy",
        project_id=project.id,
        permission="developer",
        arguments=[
            {"name": "target_env", "display_name": "Target Environment", "type": "environment_source_name"},
        ],
    )
    ```

    ## Import

    Tasks can be imported using their numeric ID:
    ```bash
    pulumi import lagoon:index:Task my-task 123
    ```
    """

    # Output properties
    id: pulumi.Output[int]
    """The task definition ID."""

    name: pulumi.Output[str]
    """The task definition name."""

    type: pulumi.Output[str]
    """The task type ('command' or 'image')."""

    service: pulumi.Output[str]
    """The service container name."""

    command: pulumi.Output[Optional[str]]
    """The command to execute (for command-type tasks)."""

    image: pulumi.Output[Optional[str]]
    """The container image (for image-type tasks)."""

    description: pulumi.Output[Optional[str]]
    """The task description."""

    permission: pulumi.Output[Optional[str]]
    """The permission level ('guest', 'developer', 'maintainer')."""

    confirmation_text: pulumi.Output[Optional[str]]
    """The confirmation text displayed before running."""

    project_id: pulumi.Output[Optional[int]]
    """The project ID (for project-scoped tasks)."""

    environment_id: pulumi.Output[Optional[int]]
    """The environment ID (for environment-scoped tasks)."""

    group_name: pulumi.Output[Optional[str]]
    """The group name (for group-scoped tasks)."""

    system_wide: pulumi.Output[Optional[bool]]
    """Whether this is a system-wide task."""

    arguments: pulumi.Output[Optional[List[Dict[str, Any]]]]
    """The task argument definitions."""

    created: pulumi.Output[Optional[str]]
    """The creation timestamp."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonTaskArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonTask resource.

        Args:
            resource_name: The Pulumi resource name
            args: The task configuration arguments
            opts: Optional resource options
        """
        # Prepare inputs
        inputs = {
            "name": args.name,
            "type": args.type,
            "service": args.service,
            "command": args.command,
            "image": args.image,
            "project_id": args.project_id,
            "environment_id": args.environment_id,
            "group_name": args.group_name,
            "system_wide": args.system_wide,
            "description": args.description,
            "permission": args.permission,
            "confirmation_text": args.confirmation_text,
            "arguments": args.arguments,
            # Outputs (set by provider)
            "id": None,
            "created": None,
        }

        super().__init__(LagoonTaskProvider(), resource_name, inputs, opts)
