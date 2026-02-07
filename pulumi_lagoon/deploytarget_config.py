"""Lagoon DeployTargetConfig resource - Dynamic provider for deploy target configurations.

Deploy target configurations route specific branches or pull requests to specific
Kubernetes clusters, enabling multi-cluster deployment strategies.
"""

from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi.dynamic as dynamic

from .config import LagoonConfig


@dataclass
class LagoonDeployTargetConfigArgs:
    """Arguments for creating a Lagoon deploy target configuration."""

    project_id: pulumi.Input[int]
    """Project ID to configure."""

    deploy_target_id: pulumi.Input[int]
    """Deploy target (Kubernetes cluster) ID to route matching deployments to."""

    branches: Optional[pulumi.Input[str]] = None
    """Branch regex pattern (e.g., 'main', '^feature/.*$'). Empty string means no branch matching."""

    pullrequests: Optional[pulumi.Input[str]] = None
    """Whether to handle pull requests ('true' or 'false'). Default: 'false'."""

    weight: Optional[pulumi.Input[int]] = None
    """Priority weight (higher = more priority). Default: 1."""

    deploy_target_project_pattern: Optional[pulumi.Input[str]] = None
    """Optional namespace pattern for deployments."""

    # API configuration - allows passing token/secret dynamically
    api_url: Optional[pulumi.Input[str]] = None
    """Lagoon API URL. If not provided, uses LAGOON_API_URL env var or config."""

    api_token: Optional[pulumi.Input[str]] = None
    """Lagoon API token. If not provided, uses LAGOON_TOKEN env var or config."""

    jwt_secret: Optional[pulumi.Input[str]] = None
    """JWT secret for generating admin tokens. Alternative to api_token."""


class LagoonDeployTargetConfigProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon deploy target configurations."""

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
        """Create a new deploy target configuration."""
        client = self._get_client(inputs)

        # Prepare input data for API call
        create_args = {
            "project": inputs["project_id"],
            "deploy_target": inputs["deploy_target_id"],
            "branches": inputs.get("branches", ""),
            "pullrequests": inputs.get("pullrequests", "false"),
            "weight": inputs.get("weight", 1),
        }

        if inputs.get("deploy_target_project_pattern"):
            create_args["deploy_target_project_pattern"] = inputs["deploy_target_project_pattern"]

        # Create deploy target config via API
        result = client.add_deploy_target_config(**create_args)

        # Return outputs
        outs = {
            "id": result["id"],
            "project_id": result.get("projectId") or inputs["project_id"],
            "deploy_target_id": result.get("deployTargetId") or inputs["deploy_target_id"],
            "branches": result.get("branches", ""),
            "pullrequests": result.get("pullrequests", "false"),
            "weight": result.get("weight", 1),
            "deploy_target_project_pattern": result.get("deployTargetProjectPattern"),
        }

        return dynamic.CreateResult(id_=str(result["id"]), outs=outs)

    def update(self, id, old_inputs, new_inputs):
        """Update an existing deploy target configuration."""
        client = self._get_client(new_inputs)

        # Prepare update data
        update_args = {}

        # Check which fields have changed
        if new_inputs.get("branches") != old_inputs.get("branches"):
            update_args["branches"] = new_inputs.get("branches", "")
        if new_inputs.get("pullrequests") != old_inputs.get("pullrequests"):
            update_args["pullrequests"] = new_inputs.get("pullrequests", "false")
        if new_inputs.get("weight") != old_inputs.get("weight"):
            update_args["weight"] = new_inputs.get("weight", 1)
        if new_inputs.get("deploy_target_project_pattern") != old_inputs.get(
            "deploy_target_project_pattern"
        ):
            update_args["deployTargetProjectPattern"] = new_inputs.get(
                "deploy_target_project_pattern"
            )

        # Only update if there are changes
        if update_args:
            result = client.update_deploy_target_config(int(id), **update_args)

            outs = {
                "id": result["id"],
                "project_id": result.get("projectId") or new_inputs["project_id"],
                "deploy_target_id": result.get("deployTargetId") or new_inputs["deploy_target_id"],
                "branches": result.get("branches", ""),
                "pullrequests": result.get("pullrequests", "false"),
                "weight": result.get("weight", 1),
                "deploy_target_project_pattern": result.get("deployTargetProjectPattern"),
            }
        else:
            # No changes, return new inputs
            outs = new_inputs

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a deploy target configuration."""
        client = self._get_client(props)

        # Delete via API
        client.delete_deploy_target_config(int(id), props["project_id"])

    def read(self, id, props):
        """Read/refresh a deploy target configuration from API.

        Supports both refresh (props available) and import (props empty).
        For import, the ID format is: project_id:config_id
        """
        from .import_utils import ImportIdParser

        client = self._get_client(props)

        # Detect import vs refresh scenario
        if ImportIdParser.is_import_scenario(id, props, ["project_id"]):
            # Import: parse composite ID
            project_id, config_id = ImportIdParser.parse_deploy_target_config_id(id)
        else:
            # Refresh: use props from state, ID is the config_id
            config_id = int(id)
            project_id = int(props["project_id"])

        # Query current state
        result = client.get_deploy_target_config_by_id(config_id, project_id)

        if not result:
            # Config no longer exists
            return None

        # Return current state
        outs = {
            "id": result["id"],
            "project_id": result.get("projectId") or props["project_id"],
            "deploy_target_id": result.get("deployTargetId") or props["deploy_target_id"],
            "branches": result.get("branches", ""),
            "pullrequests": result.get("pullrequests", "false"),
            "weight": result.get("weight", 1),
            "deploy_target_project_pattern": result.get("deployTargetProjectPattern"),
        }

        return dynamic.ReadResult(id_=str(result["id"]), outs=outs)


class LagoonDeployTargetConfig(dynamic.Resource):
    """
    A Lagoon deploy target configuration resource.

    Routes specific branches or pull requests to a specific Kubernetes cluster.
    This enables multi-cluster deployment strategies where production branches
    deploy to one cluster and development branches deploy to another.

    Higher weight configurations take priority when multiple configurations match.

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Assume we have deploy targets for prod and nonprod clusters
    prod_target = lagoon.LagoonDeployTarget("prod-target", ...)
    nonprod_target = lagoon.LagoonDeployTarget("nonprod-target", ...)

    # Create a project
    project = lagoon.LagoonProject("my-project", ...)

    # Route 'main' branch to production cluster with higher priority
    prod_config = lagoon.LagoonDeployTargetConfig("prod-routing",
        args=lagoon.LagoonDeployTargetConfigArgs(
            project_id=project.id,
            deploy_target_id=prod_target.id,
            branches="main",
            pullrequests="false",
            weight=10,  # Higher weight = higher priority
        ),
    )

    # Route all other branches and PRs to nonprod cluster
    nonprod_config = lagoon.LagoonDeployTargetConfig("nonprod-routing",
        args=lagoon.LagoonDeployTargetConfigArgs(
            project_id=project.id,
            deploy_target_id=nonprod_target.id,
            branches=".*",  # Match all branches
            pullrequests="true",  # Handle PRs
            weight=1,  # Lower weight = lower priority (fallback)
        ),
    )
    ```
    """

    # Output properties
    id: pulumi.Output[str]
    """The deploy target config ID."""

    project_id: pulumi.Output[int]
    """The project ID."""

    deploy_target_id: pulumi.Output[int]
    """The deploy target ID."""

    branches: pulumi.Output[str]
    """Branch regex pattern."""

    pullrequests: pulumi.Output[str]
    """Whether pull requests are handled."""

    weight: pulumi.Output[int]
    """Priority weight."""

    deploy_target_project_pattern: pulumi.Output[Optional[str]]
    """Namespace pattern for deployments."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonDeployTargetConfigArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonDeployTargetConfig resource.

        Args:
            resource_name: The Pulumi resource name
            args: The deploy target config arguments
            opts: Optional resource options
        """
        # Prepare inputs
        inputs = {
            "project_id": args.project_id,
            "deploy_target_id": args.deploy_target_id,
            "branches": args.branches,
            "pullrequests": args.pullrequests,
            "weight": args.weight,
            "deploy_target_project_pattern": args.deploy_target_project_pattern,
            # API configuration (allows dynamic token passing)
            "api_url": args.api_url,
            "api_token": args.api_token,
            "jwt_secret": args.jwt_secret,
            # Output (set by provider)
            "id": None,
        }

        super().__init__(LagoonDeployTargetConfigProvider(), resource_name, inputs, opts)
