"""Registry module for multi-cluster Lagoon example.

This module provides functions for installing Harbor container registry.
"""

from .harbor import install_harbor

__all__ = [
    "install_harbor",
]
