"""Lagoon Project resource - Dynamic provider for managing Lagoon projects."""

from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi.dynamic as dynamic

from .config import LagoonConfig
from .validators import (
    validate_git_url,
    validate_positive_int,
    validate_project_name,
    validate_regex_pattern,
)


@dataclass
class LagoonProjectArgs:
    """Arguments for creating a Lagoon project."""

    name: pulumi.Input[str]
    """Project name (unique identifier)."""

    git_url: pulumi.Input[str]
    """Git repository URL (e.g., git@github.com:org/repo.git)."""

    deploytarget_id: pulumi.Input[int]
    """Deploy target (Kubernetes cluster) ID."""

    production_environment: Optional[pulumi.Input[str]] = None
    """Name of the production branch (default: 'main')."""

    branches: Optional[pulumi.Input[str]] = None
    """Regex pattern for branches to deploy (e.g., '^(main|develop)$')."""

    pullrequests: Optional[pulumi.Input[str]] = None
    """Regex pattern for pull requests to deploy (e.g., '^(PR-.*)$')."""

    openshift_project_pattern: Optional[pulumi.Input[str]] = None
    """Pattern for Kubernetes namespace naming."""

    auto_idle: Optional[pulumi.Input[int]] = None
    """Auto-idle configuration (minutes of inactivity before auto-idle)."""

    storage_calc: Optional[pulumi.Input[int]] = None
    """Storage calculation setting."""

    # API configuration - allows passing token/secret dynamically
    api_url: Optional[pulumi.Input[str]] = None
    """Lagoon API URL. If not provided, uses LAGOON_API_URL env var or config."""

    api_token: Optional[pulumi.Input[str]] = None
    """Lagoon API token. If not provided, uses LAGOON_TOKEN env var or config."""

    jwt_secret: Optional[pulumi.Input[str]] = None
    """JWT secret for generating admin tokens. Alternative to api_token."""


class LagoonProjectProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon projects."""

    def _get_client(self, inputs=None):
        """Get configured Lagoon API client.

        If api_url, api_token, or jwt_secret are provided in inputs, use those
        instead of the default configuration.
        """
        from .client import LagoonClient

        if inputs:
            api_url = inputs.get("api_url")
            api_token = inputs.get("api_token")
            jwt_secret = inputs.get("jwt_secret")

            if api_url and (api_token or jwt_secret):
                if api_token:
                    return LagoonClient(api_url, api_token)
                elif jwt_secret:
                    token = self._generate_admin_token(jwt_secret)
                    return LagoonClient(api_url, token)

        config = LagoonConfig()
        return config.get_client()

    def _generate_admin_token(self, jwt_secret: str) -> str:
        """Generate an admin JWT token from the JWT secret."""
        import time

        import jwt as pyjwt

        now = int(time.time())
        payload = {
            "role": "admin",
            "iss": "lagoon-api",
            "sub": "lagoonadmin",
            "aud": "api.dev",
            "iat": now,
            "exp": now + 3600,
        }
        return pyjwt.encode(payload, jwt_secret, algorithm="HS256")

    def create(self, inputs):
        """Create a new Lagoon project."""
        # Input validation (fail fast)
        validate_project_name(inputs.get("name"))
        validate_git_url(inputs.get("git_url"))
        validate_positive_int(inputs.get("deploytarget_id"), "deploytarget_id")

        # Validate optional regex patterns
        if inputs.get("branches"):
            validate_regex_pattern(inputs["branches"], "branches")
        if inputs.get("pullrequests"):
            validate_regex_pattern(inputs["pullrequests"], "pullrequests")

        client = self._get_client(inputs)

        # Prepare input data - use snake_case for client method arguments
        create_args = {
            "name": inputs["name"],
            "git_url": inputs["git_url"],
            "openshift": inputs["deploytarget_id"],
        }

        # Add optional fields (these go into **kwargs and need camelCase for GraphQL)
        if inputs.get("production_environment"):
            create_args["productionEnvironment"] = inputs["production_environment"]
        if inputs.get("branches"):
            create_args["branches"] = inputs["branches"]
        if inputs.get("pullrequests"):
            create_args["pullrequests"] = inputs["pullrequests"]
        if inputs.get("openshift_project_pattern"):
            create_args["openshiftProjectPattern"] = inputs["openshift_project_pattern"]
        if inputs.get("auto_idle") is not None:
            create_args["autoIdle"] = inputs["auto_idle"]
        if inputs.get("storage_calc") is not None:
            create_args["storageCalc"] = inputs["storage_calc"]

        # Create project via API
        result = client.create_project(**create_args)

        # Return outputs
        outs = {
            "id": result["id"],
            "name": result["name"],
            "git_url": result["gitUrl"],
            "deploytarget_id": result["openshift"],
            "production_environment": result.get("productionEnvironment"),
            "branches": result.get("branches"),
            "pullrequests": result.get("pullrequests"),
            "created": result.get("created"),
        }

        return dynamic.CreateResult(id_=str(result["id"]), outs=outs)

    def update(self, id, old_inputs, new_inputs):
        """Update an existing Lagoon project."""
        # Validate changed inputs (fail fast)
        # Note: name cannot be changed after creation, but validate if provided
        if new_inputs.get("git_url") != old_inputs.get("git_url"):
            validate_git_url(new_inputs.get("git_url"))
        if new_inputs.get("deploytarget_id") != old_inputs.get("deploytarget_id"):
            validate_positive_int(new_inputs.get("deploytarget_id"), "deploytarget_id")
        if new_inputs.get("branches") != old_inputs.get("branches"):
            if new_inputs.get("branches"):
                validate_regex_pattern(new_inputs["branches"], "branches")
        if new_inputs.get("pullrequests") != old_inputs.get("pullrequests"):
            if new_inputs.get("pullrequests"):
                validate_regex_pattern(new_inputs["pullrequests"], "pullrequests")

        client = self._get_client(new_inputs)

        # Store project ID for the API call (passed as positional argument)
        project_id = int(id)

        # Prepare update data (without id - passed separately to client method)
        update_args = {}

        # Check which fields have changed and include them
        if new_inputs.get("git_url") != old_inputs.get("git_url"):
            update_args["gitUrl"] = new_inputs["git_url"]
        if new_inputs.get("deploytarget_id") != old_inputs.get("deploytarget_id"):
            update_args["openshift"] = new_inputs["deploytarget_id"]
        if new_inputs.get("production_environment") != old_inputs.get("production_environment"):
            update_args["productionEnvironment"] = new_inputs.get("production_environment")
        if new_inputs.get("branches") != old_inputs.get("branches"):
            update_args["branches"] = new_inputs.get("branches")
        if new_inputs.get("pullrequests") != old_inputs.get("pullrequests"):
            update_args["pullrequests"] = new_inputs.get("pullrequests")
        if new_inputs.get("openshift_project_pattern") != old_inputs.get(
            "openshift_project_pattern"
        ):
            update_args["openshiftProjectPattern"] = new_inputs.get("openshift_project_pattern")
        if new_inputs.get("auto_idle") != old_inputs.get("auto_idle"):
            update_args["autoIdle"] = new_inputs.get("auto_idle")
        if new_inputs.get("storage_calc") != old_inputs.get("storage_calc"):
            update_args["storageCalc"] = new_inputs.get("storage_calc")

        # Only update if there are changes
        if len(update_args) > 0:
            result = client.update_project(project_id, **update_args)

            # Return updated outputs
            outs = {
                "id": result["id"],
                "name": result["name"],
                "git_url": result["gitUrl"],
                "deploytarget_id": result["openshift"],
                "production_environment": result.get("productionEnvironment"),
                "branches": result.get("branches"),
                "pullrequests": result.get("pullrequests"),
                "created": old_inputs.get("created"),  # Created timestamp doesn't change
            }
        else:
            # No changes, return old inputs
            outs = new_inputs

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a Lagoon project."""
        client = self._get_client(props)

        # Delete via API (uses name, not ID)
        project_name = props["name"]
        client.delete_project(project_name)

    def read(self, id, props):
        """Read/refresh a Lagoon project from API."""
        client = self._get_client(props)

        # Query current state
        result = client.get_project_by_id(int(id))

        if not result:
            # Project no longer exists
            return None

        # Return current state
        outs = {
            "id": result["id"],
            "name": result["name"],
            "git_url": result["gitUrl"],
            "deploytarget_id": result["openshift"],
            "production_environment": result.get("productionEnvironment"),
            "branches": result.get("branches"),
            "pullrequests": result.get("pullrequests"),
            "created": result.get("created"),
        }

        return dynamic.ReadResult(id_=str(result["id"]), outs=outs)


class LagoonProject(dynamic.Resource, module="lagoon", name="Project"):
    """
    A Lagoon project resource.

    Manages a project in Lagoon, which represents an application or website
    that will be deployed to a Kubernetes cluster.

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    project = lagoon.LagoonProject("my-site",
        name="my-drupal-site",
        git_url="git@github.com:org/repo.git",
        deploytarget_id=1,
        production_environment="main",
        branches="^(main|develop|stage)$",
        pullrequests="^(PR-.*)",
    )

    pulumi.export("project_id", project.id)
    ```
    """

    # Output properties
    # Note: id is inherited from base class as Output[str]

    name: pulumi.Output[str]
    """The project name."""

    git_url: pulumi.Output[str]
    """The Git repository URL."""

    deploytarget_id: pulumi.Output[int]
    """The deploy target (Kubernetes cluster) ID."""

    production_environment: pulumi.Output[Optional[str]]
    """The production branch name."""

    branches: pulumi.Output[Optional[str]]
    """Branch deployment regex pattern."""

    pullrequests: pulumi.Output[Optional[str]]
    """Pull request deployment regex pattern."""

    created: pulumi.Output[Optional[str]]
    """Creation timestamp."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonProjectArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonProject resource.

        Args:
            resource_name: The Pulumi resource name
            args: The project configuration arguments
            opts: Optional resource options
        """
        # Prepare inputs
        inputs = {
            "name": args.name,
            "git_url": args.git_url,
            "deploytarget_id": args.deploytarget_id,
            "production_environment": args.production_environment,
            "branches": args.branches,
            "pullrequests": args.pullrequests,
            "openshift_project_pattern": args.openshift_project_pattern,
            "auto_idle": args.auto_idle,
            "storage_calc": args.storage_calc,
            # API configuration (allows dynamic token passing)
            "api_url": args.api_url,
            "api_token": args.api_token,
            "jwt_secret": args.jwt_secret,
            # Outputs (set by provider)
            "id": None,
            "created": None,
        }

        super().__init__(LagoonProjectProvider(), resource_name, inputs, opts)
