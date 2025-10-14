"""Pulumi Lagoon Provider - Manage Lagoon resources as infrastructure-as-code."""

__version__ = "0.1.0"

# Configuration
from .config import LagoonConfig

# Resources - will be implemented in next phase
# from .project import LagoonProject
# from .environment import LagoonEnvironment
# from .variable import LagoonVariable

__all__ = [
    "LagoonConfig",
    # "LagoonProject",
    # "LagoonEnvironment",
    # "LagoonVariable",
]
