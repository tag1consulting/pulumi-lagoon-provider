"""Pulumi Lagoon Provider - Manage Lagoon resources as infrastructure-as-code."""

__version__ = "0.1.0"

# Configuration
from .config import LagoonConfig

# Resources
from .project import LagoonProject, LagoonProjectArgs
from .environment import LagoonEnvironment, LagoonEnvironmentArgs
from .variable import LagoonVariable, LagoonVariableArgs
from .deploytarget import LagoonDeployTarget, LagoonDeployTargetArgs
from .deploytarget_config import LagoonDeployTargetConfig, LagoonDeployTargetConfigArgs

# Client (for advanced use cases)
from .client import LagoonClient

# Exceptions (centralized)
from .exceptions import (
    LagoonAPIError,
    LagoonConnectionError,
    LagoonProviderError,
    LagoonValidationError,
    LagoonResourceNotFoundError,
)

__all__ = [
    # Configuration
    "LagoonConfig",
    # Resources
    "LagoonProject",
    "LagoonProjectArgs",
    "LagoonEnvironment",
    "LagoonEnvironmentArgs",
    "LagoonVariable",
    "LagoonVariableArgs",
    "LagoonDeployTarget",
    "LagoonDeployTargetArgs",
    "LagoonDeployTargetConfig",
    "LagoonDeployTargetConfigArgs",
    # Client
    "LagoonClient",
    # Exceptions
    "LagoonAPIError",
    "LagoonConnectionError",
    "LagoonProviderError",
    "LagoonValidationError",
    "LagoonResourceNotFoundError",
]
