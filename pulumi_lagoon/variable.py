"""Lagoon Variable resource - Dynamic provider for managing Lagoon variables."""

from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi.dynamic as dynamic

from .client import LagoonAPIError, LagoonConnectionError
from .config import LagoonConfig
from .validators import (
    validate_positive_int,
    validate_required,
    validate_scope,
    validate_variable_name,
)


@dataclass
class LagoonVariableArgs:
    """Arguments for creating a Lagoon variable."""

    name: pulumi.Input[str]
    """Variable name."""

    value: pulumi.Input[str]
    """Variable value."""

    project_id: pulumi.Input[int]
    """Parent project ID."""

    scope: pulumi.Input[str]
    """Variable scope: 'build', 'runtime', 'global', 'container_registry', or 'internal_container_registry'."""

    environment_id: Optional[pulumi.Input[int]] = None
    """Environment ID (if omitted, variable is project-scoped)."""


class LagoonVariableProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon variables."""

    def _get_client(self):
        """Get configured Lagoon API client."""
        config = LagoonConfig()
        return config.get_client()

    def _create_variable_id(self, project_id: int, env_id: Optional[int], name: str) -> str:
        """Create a unique ID for the variable."""
        if env_id:
            return f"p{project_id}e{env_id}_{name}"
        return f"p{project_id}_{name}"

    def create(self, inputs):
        """Create a new Lagoon variable."""
        # Input validation (fail fast)
        validate_variable_name(inputs.get("name"))
        validate_required(inputs.get("value"), "value")
        project_id = validate_positive_int(inputs.get("project_id"), "project_id")
        validate_scope(inputs.get("scope"))

        environment_id = None
        if inputs.get("environment_id"):
            environment_id = validate_positive_int(inputs["environment_id"], "environment_id")

        client = self._get_client()

        # Prepare input data
        create_args = {
            "name": inputs["name"],
            "value": inputs["value"],
            "project": project_id,
            "scope": inputs["scope"],
        }

        # Add environment if specified
        if environment_id:
            create_args["environment"] = environment_id

        # Create variable via API
        result = client.add_env_variable(**create_args)

        # Generate a unique ID for the variable
        var_id = self._create_variable_id(project_id, environment_id, inputs["name"])

        # Return outputs - note: new API doesn't return project/environment objects
        outs = {
            "id": result.get("id", var_id),  # Use API ID if available, otherwise use generated ID
            "name": result["name"],
            "value": result["value"],
            "project_id": inputs["project_id"],  # Store from inputs
            "environment_id": inputs.get("environment_id"),  # Store from inputs
            "scope": result["scope"],
        }

        return dynamic.CreateResult(id_=var_id, outs=outs)

    def update(self, id, old_inputs, new_inputs):
        """Update an existing Lagoon variable."""
        # Input validation for new values (fail fast)
        validate_variable_name(new_inputs.get("name"))
        validate_required(new_inputs.get("value"), "value")
        new_project_id = validate_positive_int(new_inputs.get("project_id"), "project_id")
        validate_scope(new_inputs.get("scope"))

        new_environment_id = None
        if new_inputs.get("environment_id"):
            new_environment_id = validate_positive_int(
                new_inputs["environment_id"], "environment_id"
            )

        client = self._get_client()

        # Parse old IDs (these should be valid since they were previously created)
        old_project_id = int(old_inputs["project_id"])
        old_environment_id = (
            int(old_inputs["environment_id"]) if old_inputs.get("environment_id") else None
        )

        # In Lagoon, updating a variable is done by deleting and re-creating
        # First, delete the old variable
        delete_args = {
            "name": old_inputs["name"],
            "project": old_project_id,
        }
        if old_environment_id:
            delete_args["environment"] = old_environment_id

        try:
            client.delete_env_variable(**delete_args)
        except LagoonAPIError:
            # Variable might not exist or might be already deleted
            # This is acceptable during update - continue with recreation
            pass
        except LagoonConnectionError:
            # Connection errors should propagate - can't safely continue
            raise

        # Create new variable with updated values
        create_args = {
            "name": new_inputs["name"],
            "value": new_inputs["value"],
            "project": new_project_id,
            "scope": new_inputs["scope"],
        }

        if new_environment_id:
            create_args["environment"] = new_environment_id

        result = client.add_env_variable(**create_args)

        # Generate new ID
        var_id = self._create_variable_id(new_project_id, new_environment_id, new_inputs["name"])

        # Return updated outputs - note: new API doesn't return project/environment objects
        outs = {
            "id": result.get("id", var_id),
            "name": result["name"],
            "value": result["value"],
            "project_id": new_project_id,  # Store from inputs
            "environment_id": new_environment_id,  # Store from inputs
            "scope": result["scope"],
        }

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a Lagoon variable."""
        client = self._get_client()

        # Ensure IDs are integers (Pulumi may pass them as strings)
        project_id = int(props["project_id"])
        environment_id = int(props["environment_id"]) if props.get("environment_id") else None

        # Delete via API
        delete_args = {
            "name": props["name"],
            "project": project_id,
        }

        if environment_id:
            delete_args["environment"] = environment_id

        client.delete_env_variable(**delete_args)

    def read(self, id, props):
        """Read/refresh a Lagoon variable from API.

        Supports both refresh (props available) and import (props empty).
        For import, the ID format is:
          - Environment-level: project_id:env_id:var_name
          - Project-level: project_id::var_name (empty env_id)
        """
        from .import_utils import ImportIdParser

        client = self._get_client()

        # Detect import vs refresh scenario
        if ImportIdParser.is_import_scenario(id, props, ["name", "project_id"]):
            # Import: parse composite ID
            project_id, environment_id, var_name = ImportIdParser.parse_variable_id(id)
        else:
            # Refresh: use props from state
            project_id = int(props["project_id"])
            environment_id = int(props["environment_id"]) if props.get("environment_id") else None
            var_name = props["name"]

        # Query current state
        result = client.get_env_variable_by_name(
            name=var_name, project=project_id, environment=environment_id
        )

        if not result:
            # Variable no longer exists
            return None

        # Return current state
        outs = {
            "id": result.get("id", id),
            "name": result["name"],
            "value": result["value"],
            "project_id": result["project"]["id"],
            "environment_id": result.get("environment", {}).get("id")
            if result.get("environment")
            else None,
            "scope": result["scope"],
        }

        return dynamic.ReadResult(id_=id, outs=outs)


class LagoonVariable(dynamic.Resource, module="lagoon", name="Variable"):
    """
    A Lagoon variable resource.

    Manages an environment variable in Lagoon, either at the project level
    or environment level. Variables can have different scopes (build, runtime, global).

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Create project and environment first
    project = lagoon.LagoonProject("my-site", ...)
    prod_env = lagoon.LagoonEnvironment("production", ...)

    # Project-level variable (applies to all environments)
    project_var = lagoon.LagoonVariable("api-key",
        name="API_KEY",
        value="secret-key-123",
        project_id=project.id,
        scope="runtime",
    )

    # Environment-level variable (specific to one environment)
    env_var = lagoon.LagoonVariable("db-host",
        name="DATABASE_HOST",
        value="mysql.production.example.com",
        project_id=project.id,
        environment_id=prod_env.id,
        scope="runtime",
    )

    # Build-time variable
    build_var = lagoon.LagoonVariable("build-mode",
        name="BUILD_MODE",
        value="production",
        project_id=project.id,
        scope="build",
    )
    ```
    """

    # Output properties
    id: pulumi.Output[str]
    """The variable ID (composite key)."""

    name: pulumi.Output[str]
    """The variable name."""

    value: pulumi.Output[str]
    """The variable value."""

    project_id: pulumi.Output[int]
    """The parent project ID."""

    environment_id: pulumi.Output[Optional[int]]
    """The environment ID (None for project-level variables)."""

    scope: pulumi.Output[str]
    """The variable scope (build, runtime, global, etc.)."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonVariableArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonVariable resource.

        Args:
            resource_name: The Pulumi resource name
            args: The variable configuration arguments
            opts: Optional resource options
        """
        # Prepare inputs
        inputs = {
            "name": args.name,
            "value": args.value,
            "project_id": args.project_id,
            "scope": args.scope,
            "environment_id": args.environment_id,
            # Outputs (set by provider)
            "id": None,
        }

        super().__init__(LagoonVariableProvider(), resource_name, inputs, opts)
