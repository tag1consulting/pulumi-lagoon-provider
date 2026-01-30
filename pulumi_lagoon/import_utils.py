"""Import utilities for Pulumi Lagoon Provider.

This module provides utilities for parsing composite import IDs during
`pulumi import` operations. Import IDs encode the information needed to
fetch resources from the Lagoon API when props are not available.

Import ID Formats:
    - LagoonProject: {numeric_id}
    - LagoonDeployTarget: {numeric_id}
    - LagoonEnvironment: {project_id}:{env_name}
    - LagoonVariable: {project_id}:{env_id}:{var_name} or {project_id}::{var_name}
    - LagoonDeployTargetConfig: {project_id}:{config_id}
"""

from typing import List, Optional, Tuple

from .exceptions import LagoonValidationError


class ImportIdParser:
    """Utility class for parsing import IDs and detecting import scenarios.

    During `pulumi import`, the read() method is called with only the import ID
    and empty/minimal props. This class helps detect that scenario and parse
    the composite import ID to extract the necessary information.

    Example:
        >>> # In a resource provider's read() method:
        >>> if ImportIdParser.is_import_scenario(id, props, ["name", "project_id"]):
        >>>     project_id, env_name = ImportIdParser.parse_environment_id(id)
        >>> else:
        >>>     project_id = int(props["project_id"])
        >>>     env_name = props["name"]
    """

    @staticmethod
    def is_import_scenario(id: str, props: dict, required_props: List[str]) -> bool:
        """Detect if this is an import scenario vs a normal refresh.

        During import, props will be empty or missing required fields.
        During refresh, props will have all the fields from the state.

        Args:
            id: The resource ID passed to read()
            props: The props dict passed to read()
            required_props: List of prop keys that would be present in a refresh

        Returns:
            True if this appears to be an import (missing required props)
        """
        if not props:
            return True

        for prop in required_props:
            if prop not in props or props.get(prop) is None:
                return True

        return False

    @staticmethod
    def parse_environment_id(import_id: str) -> Tuple[int, str]:
        """Parse an environment import ID.

        Format: {project_id}:{env_name}
        Example: "123:main" -> (123, "main")

        Args:
            import_id: The import ID in format "project_id:env_name"

        Returns:
            Tuple of (project_id, env_name)

        Raises:
            LagoonValidationError: If the format is invalid
        """
        parts = import_id.split(":", 1)

        if len(parts) != 2:
            raise LagoonValidationError(
                f"Invalid environment import ID format: '{import_id}'. "
                f"Expected format: 'project_id:env_name' (e.g., '123:main')"
            )

        project_id_str, env_name = parts

        if not project_id_str:
            raise LagoonValidationError(
                f"Invalid environment import ID: '{import_id}'. Project ID cannot be empty."
            )

        if not env_name:
            raise LagoonValidationError(
                f"Invalid environment import ID: '{import_id}'. Environment name cannot be empty."
            )

        try:
            project_id = int(project_id_str)
        except ValueError:
            raise LagoonValidationError(
                f"Invalid environment import ID: '{import_id}'. "
                f"Project ID must be a number, got '{project_id_str}'."
            )

        if project_id <= 0:
            raise LagoonValidationError(
                f"Invalid environment import ID: '{import_id}'. "
                f"Project ID must be positive, got {project_id}."
            )

        return project_id, env_name

    @staticmethod
    def parse_variable_id(import_id: str) -> Tuple[int, Optional[int], str]:
        """Parse a variable import ID.

        Format: {project_id}:{env_id}:{var_name} (environment-level)
        Format: {project_id}::{var_name} (project-level, empty env_id)
        Example: "123:456:DATABASE_HOST" -> (123, 456, "DATABASE_HOST")
        Example: "123::API_KEY" -> (123, None, "API_KEY")

        Args:
            import_id: The import ID

        Returns:
            Tuple of (project_id, environment_id or None, var_name)

        Raises:
            LagoonValidationError: If the format is invalid
        """
        parts = import_id.split(":", 2)

        if len(parts) != 3:
            raise LagoonValidationError(
                f"Invalid variable import ID format: '{import_id}'. "
                f"Expected format: 'project_id:env_id:var_name' or "
                f"'project_id::var_name' for project-level variables. "
                f"(e.g., '123:456:DATABASE_HOST' or '123::API_KEY')"
            )

        project_id_str, env_id_str, var_name = parts

        if not project_id_str:
            raise LagoonValidationError(
                f"Invalid variable import ID: '{import_id}'. Project ID cannot be empty."
            )

        if not var_name:
            raise LagoonValidationError(
                f"Invalid variable import ID: '{import_id}'. Variable name cannot be empty."
            )

        try:
            project_id = int(project_id_str)
        except ValueError:
            raise LagoonValidationError(
                f"Invalid variable import ID: '{import_id}'. "
                f"Project ID must be a number, got '{project_id_str}'."
            )

        if project_id <= 0:
            raise LagoonValidationError(
                f"Invalid variable import ID: '{import_id}'. "
                f"Project ID must be positive, got {project_id}."
            )

        # env_id is optional (empty string means project-level variable)
        environment_id: Optional[int] = None
        if env_id_str:
            try:
                environment_id = int(env_id_str)
            except ValueError:
                raise LagoonValidationError(
                    f"Invalid variable import ID: '{import_id}'. "
                    f"Environment ID must be a number or empty, got '{env_id_str}'."
                )

            if environment_id <= 0:
                raise LagoonValidationError(
                    f"Invalid variable import ID: '{import_id}'. "
                    f"Environment ID must be positive, got {environment_id}."
                )

        return project_id, environment_id, var_name

    @staticmethod
    def parse_deploy_target_config_id(import_id: str) -> Tuple[int, int]:
        """Parse a deploy target config import ID.

        Format: {project_id}:{config_id}
        Example: "123:5" -> (123, 5)

        Args:
            import_id: The import ID in format "project_id:config_id"

        Returns:
            Tuple of (project_id, config_id)

        Raises:
            LagoonValidationError: If the format is invalid
        """
        parts = import_id.split(":", 1)

        if len(parts) != 2:
            raise LagoonValidationError(
                f"Invalid deploy target config import ID format: '{import_id}'. "
                f"Expected format: 'project_id:config_id' (e.g., '123:5')"
            )

        project_id_str, config_id_str = parts

        if not project_id_str:
            raise LagoonValidationError(
                f"Invalid deploy target config import ID: '{import_id}'. "
                f"Project ID cannot be empty."
            )

        if not config_id_str:
            raise LagoonValidationError(
                f"Invalid deploy target config import ID: '{import_id}'. Config ID cannot be empty."
            )

        try:
            project_id = int(project_id_str)
        except ValueError:
            raise LagoonValidationError(
                f"Invalid deploy target config import ID: '{import_id}'. "
                f"Project ID must be a number, got '{project_id_str}'."
            )

        try:
            config_id = int(config_id_str)
        except ValueError:
            raise LagoonValidationError(
                f"Invalid deploy target config import ID: '{import_id}'. "
                f"Config ID must be a number, got '{config_id_str}'."
            )

        if project_id <= 0:
            raise LagoonValidationError(
                f"Invalid deploy target config import ID: '{import_id}'. "
                f"Project ID must be positive, got {project_id}."
            )

        if config_id <= 0:
            raise LagoonValidationError(
                f"Invalid deploy target config import ID: '{import_id}'. "
                f"Config ID must be positive, got {config_id}."
            )

        return project_id, config_id
