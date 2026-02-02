"""Simple Lagoon project example.

This example demonstrates how to use the Pulumi Lagoon provider to:
1. Create a Lagoon project
2. Create production and development environments
3. Add environment variables
4. Configure notifications (Slack, Email, RocketChat, Microsoft Teams)
5. Link notifications to projects

Prerequisites:
- Lagoon API access (set LAGOON_API_URL and LAGOON_TOKEN)
- A deploy target (Kubernetes cluster) ID
"""

import pulumi

import pulumi_lagoon as lagoon

# Get configuration
config = pulumi.Config()
deploytarget_id = config.require_int("deploytargetId")

# Optional: customize project name
project_name = config.get("projectName") or "example-drupal-site"

# Create a Lagoon project
project = lagoon.LagoonProject(
    "example-project",
    lagoon.LagoonProjectArgs(
        name=project_name,
        git_url="git@github.com:example/drupal-site.git",
        deploytarget_id=deploytarget_id,
        production_environment="main",
        branches="^(main|develop|stage)$",
        pullrequests="^(PR-.*)",
    ),
)

# Create production environment
prod_env = lagoon.LagoonEnvironment(
    "production",
    lagoon.LagoonEnvironmentArgs(
        name="main",
        project_id=project.id,
        deploy_type="branch",
        environment_type="production",
    ),
)

# Create development environment
# Note: auto_idle is not supported in AddEnvironmentInput - must be set via
# updateEnvironment mutation after creation (not yet implemented in provider)
dev_env = lagoon.LagoonEnvironment(
    "development",
    lagoon.LagoonEnvironmentArgs(
        name="develop",
        project_id=project.id,
        deploy_type="branch",
        environment_type="development",
    ),
)

# Add a project-level variable (applies to all environments)
project_var = lagoon.LagoonVariable(
    "api-url",
    lagoon.LagoonVariableArgs(
        name="API_BASE_URL",
        value="https://api.example.com",
        project_id=project.id,
        scope="runtime",
    ),
)

# Add environment-specific variable for production
prod_db_host = lagoon.LagoonVariable(
    "prod-db-host",
    lagoon.LagoonVariableArgs(
        name="DATABASE_HOST",
        value="mysql-prod.example.com",
        project_id=project.id,
        environment_id=prod_env.id,
        scope="runtime",
    ),
)

# Add environment-specific variable for development
dev_db_host = lagoon.LagoonVariable(
    "dev-db-host",
    lagoon.LagoonVariableArgs(
        name="DATABASE_HOST",
        value="mysql-dev.example.com",
        project_id=project.id,
        environment_id=dev_env.id,
        scope="runtime",
    ),
)

# =============================================================================
# Notifications
# =============================================================================

# Create a Slack notification for deployment alerts
slack_notification = lagoon.LagoonNotificationSlack(
    "slack-deploys",
    lagoon.LagoonNotificationSlackArgs(
        name=f"{project_name}-slack-deploys",
        webhook="https://example.com/slack-webhook-placeholder",
        channel="#deployments",
    ),
)

# Create an Email notification for critical alerts
email_notification = lagoon.LagoonNotificationEmail(
    "email-alerts",
    lagoon.LagoonNotificationEmailArgs(
        name=f"{project_name}-email-alerts",
        email_address="ops-team@example.com",
    ),
)

# Create a RocketChat notification for team chat
rocketchat_notification = lagoon.LagoonNotificationRocketChat(
    "rocketchat-team",
    lagoon.LagoonNotificationRocketChatArgs(
        name=f"{project_name}-rocketchat",
        webhook="https://rocketchat.example.com/hooks/XXXXX/YYYYY",
        channel="#lagoon-builds",
    ),
)

# Create a Microsoft Teams notification
teams_notification = lagoon.LagoonNotificationMicrosoftTeams(
    "teams-alerts",
    lagoon.LagoonNotificationMicrosoftTeamsArgs(
        name=f"{project_name}-teams",
        webhook="https://outlook.office.com/webhook/XXXXX/IncomingWebhook/YYYYY/ZZZZZ",
    ),
)

# Link Slack notification to the project
project_slack = lagoon.LagoonProjectNotification(
    "project-slack",
    lagoon.LagoonProjectNotificationArgs(
        project_name=project.name,
        notification_type="slack",
        notification_name=slack_notification.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, slack_notification]),
)

# Link Email notification to the project
project_email = lagoon.LagoonProjectNotification(
    "project-email",
    lagoon.LagoonProjectNotificationArgs(
        project_name=project.name,
        notification_type="email",
        notification_name=email_notification.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, email_notification]),
)

# Link RocketChat notification to the project
project_rocketchat = lagoon.LagoonProjectNotification(
    "project-rocketchat",
    lagoon.LagoonProjectNotificationArgs(
        project_name=project.name,
        notification_type="rocketchat",
        notification_name=rocketchat_notification.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, rocketchat_notification]),
)

# Link Microsoft Teams notification to the project
project_teams = lagoon.LagoonProjectNotification(
    "project-teams",
    lagoon.LagoonProjectNotificationArgs(
        project_name=project.name,
        notification_type="microsoftteams",
        notification_name=teams_notification.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, teams_notification]),
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
