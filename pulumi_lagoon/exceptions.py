"""Custom exceptions for Pulumi Lagoon Provider."""

from typing import Any, Optional


class LagoonProviderError(Exception):
    """Base exception for all Lagoon provider errors.

    Provides structured error information including the field that caused
    the error, the invalid value, and a suggestion for how to fix it.
    """

    def __init__(
        self,
        message: str,
        field: Optional[str] = None,
        value: Any = None,
        suggestion: Optional[str] = None,
    ):
        self.field = field
        self.value = value
        self.suggestion = suggestion

        # Build detailed message
        parts = [message]
        if field and value is not None:
            parts.append(f"Field: {field}, Value: {repr(value)}")
        if suggestion:
            parts.append(f"Suggestion: {suggestion}")

        super().__init__("\n".join(parts))


class LagoonValidationError(LagoonProviderError):
    """Raised when input validation fails.

    This exception is raised before any API calls are made, allowing
    users to catch validation errors early in the resource lifecycle.
    """

    pass


class LagoonResourceNotFoundError(LagoonProviderError):
    """Raised when a referenced resource does not exist.

    This exception is raised when an operation references a resource
    (like a project or environment) that cannot be found in Lagoon.
    """

    pass


# Re-export existing exceptions from client.py for backward compatibility
# These are imported here so users can import all exceptions from one place
from .client import LagoonAPIError, LagoonConnectionError  # noqa: E402

__all__ = [
    "LagoonProviderError",
    "LagoonValidationError",
    "LagoonResourceNotFoundError",
    "LagoonAPIError",
    "LagoonConnectionError",
]
