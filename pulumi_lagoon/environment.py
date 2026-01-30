"""Lagoon Environment resource - Dynamic provider for managing Lagoon environments."""

from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi.dynamic as dynamic

from .config import LagoonConfig
from .validators import (
    validate_deploy_type,
    validate_environment_name,
    validate_environment_type,
    validate_positive_int,
)


@dataclass
class LagoonEnvironmentArgs:
    """Arguments for creating a Lagoon environment."""

    name: pulumi.Input[str]
    """Environment name (typically the branch name)."""

    project_id: pulumi.Input[int]
    """Parent project ID."""

    deploy_type: pulumi.Input[str]
    """Deployment type: 'branch' or 'pullrequest'."""

    environment_type: pulumi.Input[str]
    """Environment type: 'production', 'development', etc."""

    deploy_base_ref: Optional[pulumi.Input[str]] = None
    """Base reference for deployment (for pull requests)."""

    deploy_head_ref: Optional[pulumi.Input[str]] = None
    """Head reference for deployment (for pull requests)."""

    deploy_title: Optional[pulumi.Input[str]] = None
    """Deployment title (for pull requests)."""

    openshift_project_name: Optional[pulumi.Input[str]] = None
    """Kubernetes namespace name (auto-generated if not provided)."""

    auto_idle: Optional[pulumi.Input[int]] = None
    """Auto-idle setting in minutes."""


class LagoonEnvironmentProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon environments."""

    def _get_client(self):
        """Get configured Lagoon API client."""
        config = LagoonConfig()
        return config.get_client()

    def create(self, inputs):
        """Create a new Lagoon environment."""
        # Input validation (fail fast)
        validate_environment_name(inputs.get("name"))
        project_id = validate_positive_int(inputs.get("project_id"), "project_id")
        validate_deploy_type(inputs.get("deploy_type"))
        validate_environment_type(inputs.get("environment_type"))

        client = self._get_client()

        # Prepare input data
        create_args = {
            "name": inputs["name"],
            "project": project_id,
            "deploy_type": inputs["deploy_type"],
            "environment_type": inputs["environment_type"],
            # deployBaseRef is required - default to environment name (branch name) if not provided
            "deployBaseRef": inputs.get("deploy_base_ref") or inputs["name"],
        }

        # Add optional fields
        # (deployBaseRef is now handled above)
        if inputs.get("deploy_head_ref"):
            create_args["deployHeadRef"] = inputs["deploy_head_ref"]
        if inputs.get("deploy_title"):
            create_args["deployTitle"] = inputs["deploy_title"]
        if inputs.get("openshift_project_name"):
            create_args["openshiftProjectName"] = inputs["openshift_project_name"]
        # Note: autoIdle is not part of AddEnvironmentInput in newer Lagoon versions
        # It needs to be set via updateEnvironment after creation

        # Create environment via API
        result = client.add_or_update_environment(**create_args)

        # Return outputs
        outs = {
            "id": result["id"],
            "name": result["name"],
            "project_id": result["project"]["id"],
            "deploy_type": result.get("deployType"),
            "environment_type": result.get("environmentType"),
            "deploy_base_ref": result.get("deployBaseRef"),
            "deploy_head_ref": result.get("deployHeadRef"),
            "deploy_title": result.get("deployTitle"),
            "route": result.get("route"),
            "routes": result.get("routes"),
            "created": result.get("created"),
        }

        return dynamic.CreateResult(id_=str(result["id"]), outs=outs)

    def update(self, id, old_inputs, new_inputs):
        """Update an existing Lagoon environment."""
        # Input validation (fail fast)
        validate_environment_name(new_inputs.get("name"))
        project_id = validate_positive_int(new_inputs.get("project_id"), "project_id")
        validate_deploy_type(new_inputs.get("deploy_type"))
        validate_environment_type(new_inputs.get("environment_type"))

        client = self._get_client()

        # Lagoon uses addOrUpdateEnvironment for both create and update
        update_args = {
            "name": new_inputs["name"],
            "project": project_id,
            "deploy_type": new_inputs["deploy_type"],
            "environment_type": new_inputs["environment_type"],
        }

        # Add optional fields
        if new_inputs.get("deploy_base_ref"):
            update_args["deployBaseRef"] = new_inputs["deploy_base_ref"]
        if new_inputs.get("deploy_head_ref"):
            update_args["deployHeadRef"] = new_inputs["deploy_head_ref"]
        if new_inputs.get("deploy_title"):
            update_args["deployTitle"] = new_inputs["deploy_title"]
        if new_inputs.get("openshift_project_name"):
            update_args["openshiftProjectName"] = new_inputs["openshift_project_name"]
        if new_inputs.get("auto_idle") is not None:
            update_args["autoIdle"] = new_inputs["auto_idle"]

        # Update via API
        result = client.add_or_update_environment(**update_args)

        # Return updated outputs
        outs = {
            "id": result["id"],
            "name": result["name"],
            "project_id": result["project"]["id"],
            "deploy_type": result.get("deployType"),
            "environment_type": result.get("environmentType"),
            "deploy_base_ref": result.get("deployBaseRef"),
            "deploy_head_ref": result.get("deployHeadRef"),
            "deploy_title": result.get("deployTitle"),
            "route": result.get("route"),
            "routes": result.get("routes"),
            "created": old_inputs.get("created"),  # Preserve original created time
        }

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a Lagoon environment."""
        client = self._get_client()

        # Ensure project_id is an integer
        project_id = int(props["project_id"])

        # Delete via API
        # Note: execute=True is required to actually delete (safety feature)
        client.delete_environment(name=props["name"], project=project_id, execute=True)

    def read(self, id, props):
        """Read/refresh a Lagoon environment from API.

        Supports both refresh (props available) and import (props empty).
        For import, the ID format is: project_id:env_name
        """
        from .import_utils import ImportIdParser

        client = self._get_client()

        # Detect import vs refresh scenario
        if ImportIdParser.is_import_scenario(id, props, ["name", "project_id"]):
            # Import: parse composite ID
            project_id, env_name = ImportIdParser.parse_environment_id(id)
        else:
            # Refresh: use props from state
            project_id = int(props["project_id"])
            env_name = props["name"]

        # Query current state
        result = client.get_environment_by_name(name=env_name, project_id=project_id)

        if not result:
            # Environment no longer exists
            return None

        # Return current state
        outs = {
            "id": result["id"],
            "name": result["name"],
            "project_id": result["project"]["id"],
            "deploy_type": result.get("deployType"),
            "environment_type": result.get("environmentType"),
            "route": result.get("route"),
            "routes": result.get("routes"),
            "created": result.get("created"),
        }

        return dynamic.ReadResult(id_=str(result["id"]), outs=outs)


class LagoonEnvironment(dynamic.Resource, module="lagoon", name="Environment"):
    """
    A Lagoon environment resource.

    Manages an environment in Lagoon, which represents a deployed version of
    a project (typically corresponding to a Git branch or pull request).

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Create project first
    project = lagoon.LagoonProject("my-site", ...)

    # Create production environment
    prod_env = lagoon.LagoonEnvironment("production",
        name="main",
        project_id=project.id,
        deploy_type="branch",
        environment_type="production",
    )

    # Create development environment
    dev_env = lagoon.LagoonEnvironment("development",
        name="develop",
        project_id=project.id,
        deploy_type="branch",
        environment_type="development",
    )

    pulumi.export("prod_url", prod_env.route)
    ```
    """

    # Output properties
    # Note: id is inherited from base class as Output[str]

    name: pulumi.Output[str]
    """The environment name."""

    project_id: pulumi.Output[int]
    """The parent project ID."""

    deploy_type: pulumi.Output[str]
    """The deployment type (branch or pullrequest)."""

    environment_type: pulumi.Output[str]
    """The environment type (production, development, etc.)."""

    deploy_base_ref: pulumi.Output[Optional[str]]
    """Base reference for deployment."""

    deploy_head_ref: pulumi.Output[Optional[str]]
    """Head reference for deployment."""

    deploy_title: pulumi.Output[Optional[str]]
    """Deployment title."""

    route: pulumi.Output[Optional[str]]
    """Primary route URL for the environment."""

    routes: pulumi.Output[Optional[str]]
    """All routes for the environment."""

    created: pulumi.Output[Optional[str]]
    """Creation timestamp."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonEnvironmentArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonEnvironment resource.

        Args:
            resource_name: The Pulumi resource name
            args: The environment configuration arguments
            opts: Optional resource options
        """
        # Prepare inputs
        inputs = {
            "name": args.name,
            "project_id": args.project_id,
            "deploy_type": args.deploy_type,
            "environment_type": args.environment_type,
            "deploy_base_ref": args.deploy_base_ref,
            "deploy_head_ref": args.deploy_head_ref,
            "deploy_title": args.deploy_title,
            "openshift_project_name": args.openshift_project_name,
            "auto_idle": args.auto_idle,
            # Outputs (set by provider)
            "id": None,
            "route": None,
            "routes": None,
            "created": None,
        }

        super().__init__(LagoonEnvironmentProvider(), resource_name, inputs, opts)
