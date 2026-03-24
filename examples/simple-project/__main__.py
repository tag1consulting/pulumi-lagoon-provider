"""Simple Lagoon project example.

This example demonstrates how to use the Pulumi Lagoon provider to:
1. Create a Lagoon project
2. Create production and development environments
3. Add environment variables
4. Configure notifications (Slack, Email, RocketChat, Microsoft Teams)
5. Link notifications to projects

Prerequisites:
- Lagoon API access (set LAGOON_API_URL and LAGOON_JWT_SECRET or LAGOON_TOKEN)
- A deploy target (Kubernetes cluster) ID
"""

import pulumi
import pulumi_lagoon
from pulumi_lagoon.lagoon import (
    Environment,
    EnvironmentArgs,
    NotificationEmail,
    NotificationEmailArgs,
    NotificationMicrosoftTeams,
    NotificationMicrosoftTeamsArgs,
    NotificationRocketChat,
    NotificationRocketChatArgs,
    NotificationSlack,
    NotificationSlackArgs,
    Project,
    ProjectArgs,
    ProjectNotification,
    ProjectNotificationArgs,
    Variable,
    VariableArgs,
)

# Get configuration
config = pulumi.Config()
deploytarget_id = config.require_int("deploytargetId")

# Optional: customize project name
project_name = config.get("projectName") or "example-drupal-site"

# Create Lagoon provider instance
# Auth from env vars: LAGOON_API_URL, LAGOON_JWT_SECRET or LAGOON_TOKEN
lagoon_config = pulumi.Config("lagoon")
lagoon_provider = pulumi_lagoon.Provider(
    "lagoon-provider",
    insecure=lagoon_config.get_bool("insecure") or False,
)

lagoon_opts = pulumi.ResourceOptions(provider=lagoon_provider)

# Create a Lagoon project
project = Project(
    "example-project",
    ProjectArgs(
        name=project_name,
        git_url="git@github.com:example/drupal-site.git",
        deploytarget_id=deploytarget_id,
        production_environment="main",
        branches="^(main|develop|stage)$",
        pullrequests="^(PR-.*)",
    ),
    opts=lagoon_opts,
)

# Create production environment
prod_env = Environment(
    "production",
    EnvironmentArgs(
        name="main",
        project_id=project.lagoon_id,
        deploy_type="branch",
        deploy_base_ref="main",
        environment_type="production",
    ),
    opts=lagoon_opts,
)

# Create development environment
# Note: auto_idle is not supported in AddEnvironmentInput - must be set via
# updateEnvironment mutation after creation (not yet implemented in provider)
dev_env = Environment(
    "development",
    EnvironmentArgs(
        name="develop",
        project_id=project.lagoon_id,
        deploy_type="branch",
        deploy_base_ref="develop",
        environment_type="development",
    ),
    opts=lagoon_opts,
)

# Add a project-level variable (applies to all environments)
project_var = Variable(
    "api-url",
    VariableArgs(
        name="API_BASE_URL",
        value="https://api.example.com",
        project_id=project.lagoon_id,
        scope="runtime",
    ),
    opts=lagoon_opts,
)

# Add environment-specific variable for production
prod_db_host = Variable(
    "prod-db-host",
    VariableArgs(
        name="DATABASE_HOST",
        value="mysql-prod.example.com",
        project_id=project.lagoon_id,
        environment_id=prod_env.lagoon_id,
        scope="runtime",
    ),
    opts=lagoon_opts,
)

# Add environment-specific variable for development
dev_db_host = Variable(
    "dev-db-host",
    VariableArgs(
        name="DATABASE_HOST",
        value="mysql-dev.example.com",
        project_id=project.lagoon_id,
        environment_id=dev_env.lagoon_id,
        scope="runtime",
    ),
    opts=lagoon_opts,
)

# =============================================================================
# Notifications
# =============================================================================

# Create a Slack notification for deployment alerts
slack_notification = NotificationSlack(
    "slack-deploys",
    NotificationSlackArgs(
        name=f"{project_name}-slack-deploys",
        webhook="https://example.com/slack-webhook-placeholder",  # Replace with real Slack webhook URL
        channel="#deployments",
    ),
    opts=lagoon_opts,
)

# Create an Email notification for critical alerts
email_notification = NotificationEmail(
    "email-alerts",
    NotificationEmailArgs(
        name=f"{project_name}-email-alerts",
        email_address="ops-team@example.com",
    ),
    opts=lagoon_opts,
)

# Create a RocketChat notification for team chat
rocketchat_notification = NotificationRocketChat(
    "rocketchat-team",
    NotificationRocketChatArgs(
        name=f"{project_name}-rocketchat",
        webhook="https://rocketchat.example.com/hooks/XXXXX/YYYYY",
        channel="#lagoon-builds",
    ),
    opts=lagoon_opts,
)

# Create a Microsoft Teams notification
teams_notification = NotificationMicrosoftTeams(
    "teams-alerts",
    NotificationMicrosoftTeamsArgs(
        name=f"{project_name}-teams",
        webhook="https://outlook.office.com/webhook/XXXXX/IncomingWebhook/YYYYY/ZZZZZ",
    ),
    opts=lagoon_opts,
)

# Link Slack notification to the project
project_slack = ProjectNotification(
    "project-slack",
    ProjectNotificationArgs(
        project_name=project.name,
        notification_type="slack",
        notification_name=slack_notification.name,
    ),
    opts=pulumi.ResourceOptions(
        depends_on=[project, slack_notification],
        provider=lagoon_provider,
    ),
)

# Link Email notification to the project
project_email = ProjectNotification(
    "project-email",
    ProjectNotificationArgs(
        project_name=project.name,
        notification_type="email",
        notification_name=email_notification.name,
    ),
    opts=pulumi.ResourceOptions(
        depends_on=[project, email_notification],
        provider=lagoon_provider,
    ),
)

# Link RocketChat notification to the project
project_rocketchat = ProjectNotification(
    "project-rocketchat",
    ProjectNotificationArgs(
        project_name=project.name,
        notification_type="rocketchat",
        notification_name=rocketchat_notification.name,
    ),
    opts=pulumi.ResourceOptions(
        depends_on=[project, rocketchat_notification],
        provider=lagoon_provider,
    ),
)

# Link Microsoft Teams notification to the project
project_teams = ProjectNotification(
    "project-teams",
    ProjectNotificationArgs(
        project_name=project.name,
        notification_type="microsoftteams",
        notification_name=teams_notification.name,
    ),
    opts=pulumi.ResourceOptions(
        depends_on=[project, teams_notification],
        provider=lagoon_provider,
    ),
)

# Export useful outputs
pulumi.export("project_id", project.id)
pulumi.export("project_name", project.name)
pulumi.export("production_url", prod_env.route)
pulumi.export("development_url", dev_env.route)
pulumi.export("production_environment_id", prod_env.id)
pulumi.export("development_environment_id", dev_env.id)

# Notification outputs
pulumi.export("slack_notification_name", slack_notification.name)
pulumi.export("email_notification_name", email_notification.name)
pulumi.export("rocketchat_notification_name", rocketchat_notification.name)
pulumi.export("teams_notification_name", teams_notification.name)
