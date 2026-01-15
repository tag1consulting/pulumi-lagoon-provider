"""Configuration management for Pulumi Lagoon provider."""

import os
from typing import Optional
import pulumi


class LagoonConfig:
    """Provider configuration for Lagoon API access."""

    def __init__(self):
        """Initialize configuration from Pulumi config or environment variables."""
        config = pulumi.Config("lagoon")

        # API endpoint configuration
        self.api_url = self._get_config_value(
            config, "apiUrl", "LAGOON_API_URL", default="https://api.lagoon.sh/graphql"
        )

        # Authentication token
        self.token = self._get_secret_value(config, "token", "LAGOON_TOKEN")

        if not self.token:
            raise ValueError(
                "Lagoon API token must be provided via:\n"
                "  - Pulumi config: pulumi config set lagoon:token <token> --secret\n"
                "  - Environment variable: LAGOON_TOKEN"
            )

        # Optional SSH key path for alternative authentication
        self.ssh_key_path = self._get_config_value(
            config, "sshKeyPath", "LAGOON_SSH_KEY_PATH", default=None, required=False
        )

    def _get_config_value(
        self,
        config: pulumi.Config,
        config_key: str,
        env_var: str,
        default: Optional[str] = None,
        required: bool = True,
    ) -> Optional[str]:
        """Get configuration value from Pulumi config or environment variable."""
        # Try Pulumi config first
        value = config.get(config_key)
        if value:
            return value

        # Try environment variable
        value = os.environ.get(env_var)
        if value:
            return value

        # Return default if provided or not required
        if not required or default is not None:
            return default

        raise ValueError(
            f"Configuration value '{config_key}' must be provided via:\n"
            f"  - Pulumi config: pulumi config set lagoon:{config_key} <value>\n"
            f"  - Environment variable: {env_var}"
        )

    def _get_secret_value(
        self, config: pulumi.Config, config_key: str, env_var: str
    ) -> Optional[str]:
        """Get secret configuration value from Pulumi config or environment variable."""
        # Try Pulumi config first (use get() since get_secret() returns Output[str])
        # The secret nature is preserved by how Pulumi stores the config value
        value = config.get(config_key)
        if value:
            return value

        # Try environment variable
        value = os.environ.get(env_var)
        if value:
            return value

        return None

    def get_client(self):
        """Create a configured Lagoon client instance."""
        from .client import LagoonClient

        return LagoonClient(self.api_url, self.token)

    def __repr__(self) -> str:
        """String representation (without exposing token)."""
        return f"LagoonConfig(api_url='{self.api_url}', token='***')"
