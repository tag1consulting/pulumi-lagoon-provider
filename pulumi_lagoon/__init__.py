"""Pulumi Lagoon Provider - Manage Lagoon resources as infrastructure-as-code."""

__version__ = "0.1.0"

# Configuration
from .config import LagoonConfig

# Resources
from .project import LagoonProject, LagoonProjectArgs
from .environment import LagoonEnvironment, LagoonEnvironmentArgs
from .variable import LagoonVariable, LagoonVariableArgs

# Client (for advanced use cases)
from .client import LagoonClient, LagoonAPIError, LagoonConnectionError

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
    # Client
    "LagoonClient",
    "LagoonAPIError",
    "LagoonConnectionError",
]
