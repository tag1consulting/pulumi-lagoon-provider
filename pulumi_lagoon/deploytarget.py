"""Lagoon DeployTarget resource - Dynamic provider for managing Lagoon deploy targets."""

from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi.dynamic as dynamic

from .config import LagoonConfig
from .validators import (
    validate_cloud_provider,
    validate_console_url,
    validate_deploy_target_name,
    validate_ssh_host,
    validate_ssh_port,
)


@dataclass
class LagoonDeployTargetArgs:
    """Arguments for creating a Lagoon deploy target (Kubernetes cluster)."""

    name: pulumi.Input[str]
    """Deploy target name (unique identifier)."""

    console_url: pulumi.Input[str]
    """Kubernetes API URL (e.g., https://kubernetes.default.svc)."""

    cloud_provider: Optional[pulumi.Input[str]] = None
    """Cloud provider (e.g., 'aws', 'gcp', 'azure', 'kind'). Default: 'kind'."""

    cloud_region: Optional[pulumi.Input[str]] = None
    """Cloud region (e.g., 'us-east-1'). Default: 'local'."""

    ssh_host: Optional[pulumi.Input[str]] = None
    """SSH host for builds."""

    ssh_port: Optional[pulumi.Input[str]] = None
    """SSH port for builds. Default: '22'."""

    build_image: Optional[pulumi.Input[str]] = None
    """Custom build image for this deploy target."""

    disabled: Optional[pulumi.Input[bool]] = None
    """Whether this deploy target is disabled. Default: False."""

    router_pattern: Optional[pulumi.Input[str]] = None
    """Route pattern for environments on this deploy target."""

    shared_bastion_secret: Optional[pulumi.Input[str]] = None
    """Shared bastion secret name for this deploy target."""

    # API configuration - allows passing token/secret dynamically
    api_url: Optional[pulumi.Input[str]] = None
    """Lagoon API URL. If not provided, uses LAGOON_API_URL env var or config."""

    api_token: Optional[pulumi.Input[str]] = None
    """Lagoon API token. If not provided, uses LAGOON_TOKEN env var or config."""

    jwt_secret: Optional[pulumi.Input[str]] = None
    """JWT secret for generating admin tokens. Alternative to api_token."""


class LagoonDeployTargetProvider(dynamic.ResourceProvider):
    """Dynamic provider implementation for Lagoon deploy targets (Kubernetes clusters)."""

    def _get_client(self, inputs=None):
        """Get configured Lagoon API client.

        If api_url, api_token, or jwt_secret are provided in inputs, use those
        instead of the default configuration. This allows passing tokens dynamically
        as Pulumi Outputs that are resolved at runtime.
        """
        from .client import LagoonClient

        # Check if inputs provide API configuration
        if inputs:
            api_url = inputs.get("api_url")
            api_token = inputs.get("api_token")
            jwt_secret = inputs.get("jwt_secret")

            # If we have explicit configuration, use it
            if api_url and (api_token or jwt_secret):
                if api_token:
                    return LagoonClient(api_url, api_token)
                elif jwt_secret:
                    # Generate admin token from JWT secret
                    token = self._generate_admin_token(jwt_secret)
                    return LagoonClient(api_url, token)

        # Fall back to default configuration
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
            "exp": now + 3600,  # 1 hour validity
        }
        return pyjwt.encode(payload, jwt_secret, algorithm="HS256")

    def create(self, inputs):
        """Create a new Lagoon deploy target."""
        # Input validation (fail fast)
        validate_deploy_target_name(inputs.get("name"))
        validate_console_url(inputs.get("console_url"))

        if inputs.get("cloud_provider"):
            validate_cloud_provider(inputs["cloud_provider"])
        if inputs.get("ssh_host"):
            validate_ssh_host(inputs["ssh_host"])
        if inputs.get("ssh_port"):
            validate_ssh_port(inputs["ssh_port"])

        client = self._get_client(inputs)

        # Prepare input data for API call
        # Use defaults from the API if not provided
        create_args = {
            "name": inputs["name"],
            "console_url": inputs["console_url"],
        }

        # Add optional fields with defaults
        if inputs.get("cloud_provider"):
            create_args["cloud_provider"] = inputs["cloud_provider"]
        else:
            create_args["cloud_provider"] = "kind"

        if inputs.get("cloud_region"):
            create_args["cloud_region"] = inputs["cloud_region"]
        else:
            create_args["cloud_region"] = "local"

        # Add other optional fields (camelCase for GraphQL API)
        if inputs.get("ssh_host"):
            create_args["sshHost"] = inputs["ssh_host"]
        if inputs.get("ssh_port"):
            create_args["sshPort"] = inputs["ssh_port"]
        if inputs.get("build_image"):
            create_args["buildImage"] = inputs["build_image"]
        if inputs.get("disabled") is not None:
            create_args["disabled"] = inputs["disabled"]
        if inputs.get("router_pattern"):
            create_args["routerPattern"] = inputs["router_pattern"]
        if inputs.get("shared_bastion_secret"):
            create_args["sharedBastionSecret"] = inputs["shared_bastion_secret"]

        # Create deploy target via API
        result = client.add_kubernetes(**create_args)

        # Return outputs
        outs = {
            "id": result["id"],
            "name": result["name"],
            "console_url": result["consoleUrl"],
            "cloud_provider": result.get("cloudProvider"),
            "cloud_region": result.get("cloudRegion"),
            "ssh_host": result.get("sshHost"),
            "ssh_port": result.get("sshPort"),
            "build_image": result.get("buildImage"),
            "disabled": result.get("disabled", False),
            "router_pattern": result.get("routerPattern"),
            "shared_bastion_secret": result.get("sharedBastionSecret"),
            "created": result.get("created"),
        }

        return dynamic.CreateResult(id_=str(result["id"]), outs=outs)

    def update(self, id, old_inputs, new_inputs):
        """Update an existing Lagoon deploy target."""
        # Validate changed inputs (fail fast)
        if new_inputs.get("console_url") != old_inputs.get("console_url"):
            validate_console_url(new_inputs.get("console_url"))
        if new_inputs.get("cloud_provider") != old_inputs.get("cloud_provider"):
            if new_inputs.get("cloud_provider"):
                validate_cloud_provider(new_inputs["cloud_provider"])
        if new_inputs.get("ssh_host") != old_inputs.get("ssh_host"):
            if new_inputs.get("ssh_host"):
                validate_ssh_host(new_inputs["ssh_host"])
        if new_inputs.get("ssh_port") != old_inputs.get("ssh_port"):
            if new_inputs.get("ssh_port"):
                validate_ssh_port(new_inputs["ssh_port"])

        client = self._get_client(new_inputs)

        # Prepare update data
        update_args = {
            "k8s_id": int(id),
        }

        # Check which fields have changed and include them
        if new_inputs.get("console_url") != old_inputs.get("console_url"):
            update_args["consoleUrl"] = new_inputs["console_url"]
        if new_inputs.get("cloud_provider") != old_inputs.get("cloud_provider"):
            update_args["cloudProvider"] = new_inputs.get("cloud_provider")
        if new_inputs.get("cloud_region") != old_inputs.get("cloud_region"):
            update_args["cloudRegion"] = new_inputs.get("cloud_region")
        if new_inputs.get("ssh_host") != old_inputs.get("ssh_host"):
            update_args["sshHost"] = new_inputs.get("ssh_host")
        if new_inputs.get("ssh_port") != old_inputs.get("ssh_port"):
            update_args["sshPort"] = new_inputs.get("ssh_port")
        if new_inputs.get("build_image") != old_inputs.get("build_image"):
            update_args["buildImage"] = new_inputs.get("build_image")
        if new_inputs.get("disabled") != old_inputs.get("disabled"):
            update_args["disabled"] = new_inputs.get("disabled")
        if new_inputs.get("router_pattern") != old_inputs.get("router_pattern"):
            update_args["routerPattern"] = new_inputs.get("router_pattern")
        if new_inputs.get("shared_bastion_secret") != old_inputs.get("shared_bastion_secret"):
            update_args["sharedBastionSecret"] = new_inputs.get("shared_bastion_secret")

        # Only update if there are changes beyond the ID
        if len(update_args) > 1:
            result = client.update_kubernetes(**update_args)

            # Return updated outputs
            outs = {
                "id": result["id"],
                "name": result["name"],
                "console_url": result["consoleUrl"],
                "cloud_provider": result.get("cloudProvider"),
                "cloud_region": result.get("cloudRegion"),
                "ssh_host": result.get("sshHost"),
                "ssh_port": result.get("sshPort"),
                "build_image": result.get("buildImage"),
                "disabled": result.get("disabled", False),
                "router_pattern": result.get("routerPattern"),
                "shared_bastion_secret": result.get("sharedBastionSecret"),
                "created": old_inputs.get("created"),  # Created timestamp doesn't change
            }
        else:
            # No changes, return old inputs
            outs = new_inputs

        return dynamic.UpdateResult(outs=outs)

    def delete(self, id, props):
        """Delete a Lagoon deploy target."""
        client = self._get_client(props)

        # Delete via API (uses name, not ID)
        target_name = props["name"]
        client.delete_kubernetes(target_name)

    def read(self, id, props):
        """Read/refresh a Lagoon deploy target from API."""
        client = self._get_client(props)

        # Query current state
        result = client.get_kubernetes_by_id(int(id))

        if not result:
            # Deploy target no longer exists
            return None

        # Return current state
        outs = {
            "id": result["id"],
            "name": result["name"],
            "console_url": result["consoleUrl"],
            "cloud_provider": result.get("cloudProvider"),
            "cloud_region": result.get("cloudRegion"),
            "ssh_host": result.get("sshHost"),
            "ssh_port": result.get("sshPort"),
            "build_image": result.get("buildImage"),
            "disabled": result.get("disabled", False),
            "router_pattern": result.get("routerPattern"),
            "shared_bastion_secret": result.get("sharedBastionSecret"),
            "created": result.get("created"),
        }

        return dynamic.ReadResult(id_=str(result["id"]), outs=outs)


class LagoonDeployTarget(dynamic.Resource):
    """
    A Lagoon deploy target (Kubernetes cluster) resource.

    Manages a deploy target in Lagoon, which represents a Kubernetes cluster
    where projects can be deployed.

    ## Example Usage

    ```python
    import pulumi
    import pulumi_lagoon as lagoon

    # Create a production deploy target
    prod_cluster = lagoon.LagoonDeployTarget("production",
        name="production-cluster",
        console_url="https://kubernetes.prod.example.com:6443",
        cloud_provider="aws",
        cloud_region="us-east-1",
        ssh_host="ssh.lagoon.prod.example.com",
    )

    # Create a development deploy target using Kind
    dev_cluster = lagoon.LagoonDeployTarget("development",
        name="dev-cluster",
        console_url="https://kubernetes.default.svc",
        cloud_provider="kind",
        cloud_region="local",
    )

    pulumi.export("prod_target_id", prod_cluster.id)
    pulumi.export("dev_target_id", dev_cluster.id)
    ```
    """

    # Output properties
    # Note: id is inherited from base class as Output[str]

    name: pulumi.Output[str]
    """The deploy target name."""

    console_url: pulumi.Output[str]
    """The Kubernetes API URL."""

    cloud_provider: pulumi.Output[Optional[str]]
    """The cloud provider."""

    cloud_region: pulumi.Output[Optional[str]]
    """The cloud region."""

    ssh_host: pulumi.Output[Optional[str]]
    """SSH host for builds."""

    ssh_port: pulumi.Output[Optional[str]]
    """SSH port for builds."""

    build_image: pulumi.Output[Optional[str]]
    """Custom build image."""

    disabled: pulumi.Output[bool]
    """Whether the deploy target is disabled."""

    router_pattern: pulumi.Output[Optional[str]]
    """Route pattern for environments."""

    shared_bastion_secret: pulumi.Output[Optional[str]]
    """Shared bastion secret name."""

    created: pulumi.Output[Optional[str]]
    """Creation timestamp."""

    def __init__(
        self,
        resource_name: str,
        args: LagoonDeployTargetArgs,
        opts: Optional[pulumi.ResourceOptions] = None,
    ):
        """
        Create a LagoonDeployTarget resource.

        Args:
            resource_name: The Pulumi resource name
            args: The deploy target configuration arguments
            opts: Optional resource options
        """
        # Prepare inputs
        inputs = {
            "name": args.name,
            "console_url": args.console_url,
            "cloud_provider": args.cloud_provider,
            "cloud_region": args.cloud_region,
            "ssh_host": args.ssh_host,
            "ssh_port": args.ssh_port,
            "build_image": args.build_image,
            "disabled": args.disabled,
            "router_pattern": args.router_pattern,
            "shared_bastion_secret": args.shared_bastion_secret,
            # API configuration (allows dynamic token passing)
            "api_url": args.api_url,
            "api_token": args.api_token,
            "jwt_secret": args.jwt_secret,
            # Outputs (set by provider)
            "id": None,
            "created": None,
        }

        super().__init__(LagoonDeployTargetProvider(), resource_name, inputs, opts)
