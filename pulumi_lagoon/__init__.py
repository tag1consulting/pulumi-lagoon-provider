"""Pulumi Lagoon Provider - Manage Lagoon resources as infrastructure-as-code."""

__version__ = "0.1.0"

# Configuration
# Client (for advanced use cases)
from .client import LagoonClient
from .config import LagoonConfig
from .deploytarget import LagoonDeployTarget, LagoonDeployTargetArgs
from .deploytarget_config import LagoonDeployTargetConfig, LagoonDeployTargetConfigArgs
from .environment import LagoonEnvironment, LagoonEnvironmentArgs

# Exceptions (centralized)
from .exceptions import (
    LagoonAPIError,
    LagoonConnectionError,
    LagoonProviderError,
    LagoonResourceNotFoundError,
    LagoonValidationError,
)

# Import utilities
from .import_utils import ImportIdParser

# Notification resources
from .notification_email import LagoonNotificationEmail, LagoonNotificationEmailArgs
from .notification_microsoftteams import (
    LagoonNotificationMicrosoftTeams,
    LagoonNotificationMicrosoftTeamsArgs,
)
from .notification_rocketchat import (
    LagoonNotificationRocketChat,
    LagoonNotificationRocketChatArgs,
)
from .notification_slack import LagoonNotificationSlack, LagoonNotificationSlackArgs

# Resources
from .project import LagoonProject, LagoonProjectArgs
from .project_notification import LagoonProjectNotification, LagoonProjectNotificationArgs
from .variable import LagoonVariable, LagoonVariableArgs

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
    # Notification resources
    "LagoonNotificationSlack",
    "LagoonNotificationSlackArgs",
    "LagoonNotificationRocketChat",
    "LagoonNotificationRocketChatArgs",
    "LagoonNotificationEmail",
    "LagoonNotificationEmailArgs",
    "LagoonNotificationMicrosoftTeams",
    "LagoonNotificationMicrosoftTeamsArgs",
    "LagoonProjectNotification",
    "LagoonProjectNotificationArgs",
    # Client
    "LagoonClient",
    # Import utilities
    "ImportIdParser",
    # Exceptions
    "LagoonAPIError",
    "LagoonConnectionError",
    "LagoonProviderError",
    "LagoonValidationError",
    "LagoonResourceNotFoundError",
]
