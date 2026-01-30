"""Input validation utilities for Lagoon resources."""

import re
from typing import Any, Optional, Set

from .exceptions import LagoonValidationError

# Valid enum values for Lagoon API
VALID_DEPLOY_TYPES: Set[str] = {"branch", "pullrequest"}
VALID_ENVIRONMENT_TYPES: Set[str] = {"production", "development", "standby"}
VALID_SCOPES: Set[str] = {
    "build",
    "runtime",
    "global",
    "container_registry",
    "internal_container_registry",
}


def validate_required(value: Any, field_name: str) -> None:
    """Validate that a required value is not None or empty.

    Args:
        value: The value to validate
        field_name: Name of the field (for error messages)

    Raises:
        LagoonValidationError: If value is None or empty string
    """
    if value is None:
        raise LagoonValidationError(
            f"Required field '{field_name}' is missing",
            field=field_name,
            value=value,
            suggestion=f"Provide a value for '{field_name}'",
        )
    if isinstance(value, str) and not value.strip():
        raise LagoonValidationError(
            f"Required field '{field_name}' cannot be empty",
            field=field_name,
            value=value,
            suggestion=f"Provide a non-empty value for '{field_name}'",
        )


def validate_project_name(name: str) -> None:
    """Validate Lagoon project name.

    Rules:
    - Must start with a lowercase letter
    - Can only contain lowercase letters, numbers, and hyphens
    - Cannot end with a hyphen
    - Maximum 58 characters (Lagoon/K8s limitation)

    Args:
        name: Project name to validate

    Raises:
        LagoonValidationError: If name is invalid
    """
    validate_required(name, "name")

    # Pattern: starts with lowercase letter, contains only lowercase letters,
    # numbers, and hyphens, ends with lowercase letter or number
    pattern = r"^[a-z][a-z0-9-]*[a-z0-9]$|^[a-z]$"

    if len(name) > 58:
        raise LagoonValidationError(
            "Project name exceeds maximum length of 58 characters",
            field="name",
            value=name,
            suggestion="Use a shorter project name (max 58 characters)",
        )

    if not re.match(pattern, name):
        raise LagoonValidationError(
            "Invalid project name format",
            field="name",
            value=name,
            suggestion="Project name must start with a lowercase letter, contain only "
            "lowercase letters, numbers, and hyphens, and not end with a hyphen",
        )


def validate_git_url(git_url: str) -> None:
    """Validate Git URL format.

    Accepts:
    - SSH format: git@github.com:org/repo.git
    - HTTPS format: https://github.com/org/repo.git

    Args:
        git_url: Git repository URL to validate

    Raises:
        LagoonValidationError: If URL format is invalid
    """
    validate_required(git_url, "git_url")

    # SSH format: git@host:path.git or git@host:path
    ssh_pattern = r"^git@[\w.-]+:[\w./-]+(?:\.git)?$"
    # HTTPS format: https://host/path.git or https://host/path
    https_pattern = r"^https?://[\w.-]+/[\w./-]+(?:\.git)?$"

    if not (re.match(ssh_pattern, git_url) or re.match(https_pattern, git_url)):
        raise LagoonValidationError(
            "Invalid Git URL format",
            field="git_url",
            value=git_url,
            suggestion="Use SSH format (git@github.com:org/repo.git) or "
            "HTTPS format (https://github.com/org/repo.git)",
        )


def validate_positive_int(value: Any, field_name: str, allow_zero: bool = False) -> int:
    """Validate that a value is a positive integer.

    Args:
        value: The value to validate (can be int or string)
        field_name: Name of the field (for error messages)
        allow_zero: If True, zero is allowed

    Returns:
        The validated integer value

    Raises:
        LagoonValidationError: If value is not a valid positive integer
    """
    try:
        int_value = int(value)
    except (ValueError, TypeError):
        raise LagoonValidationError(
            f"Field '{field_name}' must be an integer",
            field=field_name,
            value=value,
            suggestion=f"Provide a valid integer for '{field_name}'",
        )

    if allow_zero:
        if int_value < 0:
            raise LagoonValidationError(
                f"Field '{field_name}' must be non-negative",
                field=field_name,
                value=value,
                suggestion=f"Provide a non-negative integer for '{field_name}'",
            )
    else:
        if int_value <= 0:
            raise LagoonValidationError(
                f"Field '{field_name}' must be a positive integer",
                field=field_name,
                value=value,
                suggestion=f"Provide a positive integer for '{field_name}'",
            )

    return int_value


def validate_enum(value: str, field_name: str, valid_values: Set[str]) -> str:
    """Validate that a value is in a set of valid values.

    Args:
        value: The value to validate
        field_name: Name of the field (for error messages)
        valid_values: Set of valid values

    Returns:
        The normalized (lowercase) value

    Raises:
        LagoonValidationError: If value is not in valid_values
    """
    validate_required(value, field_name)

    normalized = value.lower().strip()

    if normalized not in valid_values:
        raise LagoonValidationError(
            f"Invalid value for '{field_name}'",
            field=field_name,
            value=value,
            suggestion=f"Valid values are: {', '.join(sorted(valid_values))}",
        )

    return normalized


def validate_deploy_type(deploy_type: str) -> str:
    """Validate deploy type enum.

    Args:
        deploy_type: The deploy type to validate

    Returns:
        The normalized deploy type

    Raises:
        LagoonValidationError: If deploy_type is invalid
    """
    return validate_enum(deploy_type, "deploy_type", VALID_DEPLOY_TYPES)


def validate_environment_type(env_type: str) -> str:
    """Validate environment type enum.

    Args:
        env_type: The environment type to validate

    Returns:
        The normalized environment type

    Raises:
        LagoonValidationError: If env_type is invalid
    """
    return validate_enum(env_type, "environment_type", VALID_ENVIRONMENT_TYPES)


def validate_scope(scope: str) -> str:
    """Validate variable scope enum.

    Args:
        scope: The variable scope to validate

    Returns:
        The normalized scope

    Raises:
        LagoonValidationError: If scope is invalid
    """
    return validate_enum(scope, "scope", VALID_SCOPES)


def validate_regex_pattern(pattern: Optional[str], field_name: str) -> None:
    """Validate that a string is a valid regex pattern.

    Args:
        pattern: The regex pattern to validate (None is allowed for optional fields)
        field_name: Name of the field (for error messages)

    Raises:
        LagoonValidationError: If pattern is not a valid regex
    """
    if pattern is None:
        return  # Optional field

    try:
        re.compile(pattern)
    except re.error as e:
        raise LagoonValidationError(
            f"Invalid regex pattern in '{field_name}'",
            field=field_name,
            value=pattern,
            suggestion=f"Fix the regex syntax error: {str(e)}",
        )


def validate_variable_name(name: str) -> None:
    """Validate environment variable name.

    Rules:
    - Must start with a letter or underscore
    - Can only contain letters, numbers, and underscores

    Args:
        name: Variable name to validate

    Raises:
        LagoonValidationError: If name is invalid
    """
    validate_required(name, "name")

    pattern = r"^[a-zA-Z_][a-zA-Z0-9_]*$"

    if not re.match(pattern, name):
        raise LagoonValidationError(
            "Invalid variable name format",
            field="name",
            value=name,
            suggestion="Variable name must start with a letter or underscore, "
            "and contain only letters, numbers, and underscores",
        )


def validate_environment_name(name: str) -> None:
    """Validate environment name.

    Rules similar to project name but more lenient for branch names.

    Args:
        name: Environment name to validate

    Raises:
        LagoonValidationError: If name is invalid
    """
    validate_required(name, "name")

    # Allow most common branch naming patterns
    pattern = r"^[a-zA-Z0-9][a-zA-Z0-9._/-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$"

    if len(name) > 63:
        raise LagoonValidationError(
            "Environment name exceeds maximum length of 63 characters",
            field="name",
            value=name,
            suggestion="Use a shorter environment name (max 63 characters)",
        )

    if not re.match(pattern, name):
        raise LagoonValidationError(
            "Invalid environment name format",
            field="name",
            value=name,
            suggestion="Environment name must start and end with alphanumeric characters",
        )


# Valid cloud providers for deploy targets
VALID_CLOUD_PROVIDERS: Set[str] = {
    "kind",
    "aws",
    "gcp",
    "azure",
    "openstack",
    "digitalocean",
    "linode",
    "vsphere",
    "bare_metal",
    "other",
}


def validate_deploy_target_name(name: str) -> None:
    """Validate deploy target name.

    Rules:
    - Must start with a lowercase letter or number
    - Can only contain lowercase letters, numbers, and hyphens
    - Cannot end with a hyphen
    - Maximum 63 characters (Kubernetes limitation)

    Args:
        name: Deploy target name to validate

    Raises:
        LagoonValidationError: If name is invalid
    """
    validate_required(name, "name")

    # Pattern: starts with lowercase letter or number, contains only lowercase letters,
    # numbers, and hyphens, ends with lowercase letter or number
    pattern = r"^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$"

    if len(name) > 63:
        raise LagoonValidationError(
            "Deploy target name exceeds maximum length of 63 characters",
            field="name",
            value=name,
            suggestion="Use a shorter deploy target name (max 63 characters)",
        )

    if not re.match(pattern, name):
        raise LagoonValidationError(
            "Invalid deploy target name format",
            field="name",
            value=name,
            suggestion="Deploy target name must start with a lowercase letter or number, "
            "contain only lowercase letters, numbers, and hyphens, and not end with a hyphen",
        )


def validate_console_url(url: str) -> None:
    """Validate Kubernetes console/API URL.

    Accepts:
    - HTTPS URLs: https://kubernetes.example.com
    - HTTP URLs (for development): http://localhost:8080

    Args:
        url: Kubernetes API URL to validate

    Raises:
        LagoonValidationError: If URL format is invalid
    """
    validate_required(url, "console_url")

    # HTTPS/HTTP URL pattern
    pattern = r"^https?://[\w.-]+(?::\d+)?(?:/[\w./-]*)?$"

    if not re.match(pattern, url):
        raise LagoonValidationError(
            "Invalid Kubernetes console URL format",
            field="console_url",
            value=url,
            suggestion="Use HTTPS format (https://kubernetes.example.com) or "
            "HTTP format for development (http://localhost:8080)",
        )


def validate_cloud_provider(provider: str) -> str:
    """Validate cloud provider value.

    Args:
        provider: The cloud provider to validate

    Returns:
        The normalized (lowercase) cloud provider

    Raises:
        LagoonValidationError: If provider is invalid
    """
    return validate_enum(provider, "cloud_provider", VALID_CLOUD_PROVIDERS)


def validate_ssh_port(port: Any) -> int:
    """Validate SSH port number.

    Args:
        port: The port number to validate (can be int or string)

    Returns:
        The validated port number

    Raises:
        LagoonValidationError: If port is invalid
    """
    try:
        port_int = int(port)
    except (ValueError, TypeError):
        raise LagoonValidationError(
            "SSH port must be an integer",
            field="ssh_port",
            value=port,
            suggestion="Provide a valid port number (e.g., 22)",
        )

    if port_int < 1 or port_int > 65535:
        raise LagoonValidationError(
            "SSH port must be between 1 and 65535",
            field="ssh_port",
            value=port,
            suggestion="Provide a valid port number (1-65535, typically 22)",
        )

    return port_int


def validate_ssh_host(host: Optional[str]) -> None:
    """Validate SSH host for builds.

    Args:
        host: SSH hostname to validate (None is allowed for optional field)

    Raises:
        LagoonValidationError: If host format is invalid
    """
    if host is None:
        return  # Optional field

    # Hostname pattern: alphanumeric with hyphens and dots, or IP address
    hostname_pattern = r"^[\w.-]+$"
    ip_pattern = r"^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$"

    if not (re.match(hostname_pattern, host) or re.match(ip_pattern, host)):
        raise LagoonValidationError(
            "Invalid SSH host format",
            field="ssh_host",
            value=host,
            suggestion="Provide a valid hostname (e.g., ssh.lagoon.example.com) or IP address",
        )
