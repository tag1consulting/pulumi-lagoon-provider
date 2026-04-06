# Managing Lagoon Notifications with Pulumi

This guide covers how to manage Lagoon notifications using the Pulumi Lagoon provider.

## Overview

Lagoon supports multiple notification types to alert teams about deployments, builds, and other events:

- **Slack** - Send notifications to Slack channels via webhooks
- **RocketChat** - Send notifications to RocketChat channels via webhooks
- **Email** - Send notifications via email
- **Microsoft Teams** - Send notifications to Microsoft Teams via webhooks

The notification workflow in Lagoon involves two steps:
1. **Create the notification** - Define the notification configuration (webhook URL, channel, email, etc.)
2. **Link to project** - Associate the notification with one or more projects

## Notification Resources

### NotificationSlack

Creates a Slack notification configuration.

```python
from pulumi_lagoon import NotificationSlack, NotificationSlackArgs

slack_alerts = NotificationSlack("deploy-alerts",
    NotificationSlackArgs(
        name="deploy-alerts",
        webhook="https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXX",
        channel="#deployments",
    )
)
```

**Arguments:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | `str` | Yes | Unique name for the notification |
| `webhook` | `str` | Yes | Slack incoming webhook URL (must be HTTPS) |
| `channel` | `str` | Yes | Slack channel (e.g., `#deployments` or `deployments`) |

**Outputs:**
- `lagoon_id` - The Lagoon internal ID
- `name` - The notification name
- `webhook` - The webhook URL
- `channel` - The channel name

### NotificationRocketChat

Creates a RocketChat notification configuration.

```python
from pulumi_lagoon import NotificationRocketChat, NotificationRocketChatArgs

rocketchat_alerts = NotificationRocketChat("team-chat",
    NotificationRocketChatArgs(
        name="team-chat",
        webhook="https://rocketchat.example.com/hooks/xxxxx/yyyyy",
        channel="#alerts",
    )
)
```

**Arguments:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | `str` | Yes | Unique name for the notification |
| `webhook` | `str` | Yes | RocketChat webhook URL (must be HTTPS) |
| `channel` | `str` | Yes | RocketChat channel |

### NotificationEmail

Creates an Email notification configuration.

```python
from pulumi_lagoon import NotificationEmail, NotificationEmailArgs

email_ops = NotificationEmail("ops-team",
    NotificationEmailArgs(
        name="ops-team",
        email_address="ops@example.com",
    )
)
```

**Arguments:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | `str` | Yes | Unique name for the notification |
| `email_address` | `str` | Yes | Email address to receive notifications |

### NotificationMicrosoftTeams

Creates a Microsoft Teams notification configuration.

```python
from pulumi_lagoon import NotificationMicrosoftTeams, NotificationMicrosoftTeamsArgs

teams_alerts = NotificationMicrosoftTeams("teams-alerts",
    NotificationMicrosoftTeamsArgs(
        name="teams-alerts",
        webhook="https://outlook.office.com/webhook/xxxxx/IncomingWebhook/yyyyy",
    )
)
```

**Arguments:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | `str` | Yes | Unique name for the notification |
| `webhook` | `str` | Yes | Microsoft Teams webhook URL (must be HTTPS) |

### ProjectNotification

Links a notification to a project. This enables the project to receive notifications.

```python
import pulumi
from pulumi_lagoon import (
    NotificationSlack, NotificationSlackArgs,
    Project, ProjectArgs,
    ProjectNotification, ProjectNotificationArgs,
)

# Create a project
project = Project("my-site",
    ProjectArgs(
        name="my-site",
        git_url="git@github.com:example/my-site.git",
        deploytarget_id=1,
    )
)

# Create a Slack notification
slack_alerts = NotificationSlack("deploy-alerts",
    NotificationSlackArgs(
        name="deploy-alerts",
        webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
        channel="#deployments",
    )
)

# Link the notification to the project
project_notification = ProjectNotification("project-slack",
    ProjectNotificationArgs(
        project_name=project.name,
        notification_type="slack",
        notification_name=slack_alerts.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, slack_alerts])
)
```

**Arguments:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `project_name` | `str` | Yes | Name of the project to link |
| `notification_type` | `str` | Yes | Type: `slack`, `rocketchat`, `email`, or `microsoftteams` |
| `notification_name` | `str` | Yes | Name of the notification to link |

**Outputs:**
- `project_name` - The project name
- `notification_type` - The notification type
- `notification_name` - The notification name
- `project_id` - The Lagoon internal project ID

## Complete Example

Here's a complete example showing how to set up multiple notifications for a project:

```python
import pulumi
from pulumi_lagoon import (
    NotificationEmail, NotificationEmailArgs,
    NotificationMicrosoftTeams, NotificationMicrosoftTeamsArgs,
    NotificationSlack, NotificationSlackArgs,
    Project, ProjectArgs,
    ProjectNotification, ProjectNotificationArgs,
)

# Create the project
project = Project("my-drupal-site",
    ProjectArgs(
        name="my-drupal-site",
        git_url="git@github.com:example/my-drupal-site.git",
        deploytarget_id=1,
        production_environment="main",
        branches="^(main|develop|stage)$",
    )
)

# === Create Notifications ===

# Slack for deployment notifications
slack_deploys = NotificationSlack("slack-deploys",
    NotificationSlackArgs(
        name="drupal-deploys",
        webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
        channel="#drupal-deployments",
    )
)

# Email for critical alerts
email_alerts = NotificationEmail("email-alerts",
    NotificationEmailArgs(
        name="critical-alerts",
        email_address="alerts@example.com",
    )
)

# Microsoft Teams for the dev team
teams_dev = NotificationMicrosoftTeams("teams-dev",
    NotificationMicrosoftTeamsArgs(
        name="dev-team-alerts",
        webhook="https://outlook.office.com/webhook/xxx/IncomingWebhook/yyy",
    )
)

# === Link Notifications to Project ===

project_slack = ProjectNotification("project-slack",
    ProjectNotificationArgs(
        project_name=project.name,
        notification_type="slack",
        notification_name=slack_deploys.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, slack_deploys])
)

project_email = ProjectNotification("project-email",
    ProjectNotificationArgs(
        project_name=project.name,
        notification_type="email",
        notification_name=email_alerts.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, email_alerts])
)

project_teams = ProjectNotification("project-teams",
    ProjectNotificationArgs(
        project_name=project.name,
        notification_type="microsoftteams",
        notification_name=teams_dev.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, teams_dev])
)

# Export notification names
pulumi.export("slack_notification", slack_deploys.name)
pulumi.export("email_notification", email_alerts.name)
pulumi.export("teams_notification", teams_dev.name)
```

## Sharing Notifications Across Projects

A single notification can be linked to multiple projects:

```python
import pulumi
from pulumi_lagoon import (
    NotificationSlack, NotificationSlackArgs,
    Project, ProjectArgs,
    ProjectNotification, ProjectNotificationArgs,
)

# Create a shared Slack notification
shared_slack = NotificationSlack("shared-slack",
    NotificationSlackArgs(
        name="team-deploys",
        webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
        channel="#all-deployments",
    )
)

# Create multiple projects
project_a = Project("project-a", ProjectArgs(...))
project_b = Project("project-b", ProjectArgs(...))
project_c = Project("project-c", ProjectArgs(...))

# Link the same notification to all projects
for idx, project in enumerate([project_a, project_b, project_c]):
    ProjectNotification(f"notification-{idx}",
        ProjectNotificationArgs(
            project_name=project.name,
            notification_type="slack",
            notification_name=shared_slack.name,
        ),
        opts=pulumi.ResourceOptions(depends_on=[project, shared_slack])
    )
```

## Importing Existing Notifications

### Import ID Formats

| Resource | Format | Example |
|----------|--------|---------|
| `NotificationSlack` | `{name}` | `deploy-alerts` |
| `NotificationRocketChat` | `{name}` | `team-chat` |
| `NotificationEmail` | `{name}` | `ops-team` |
| `NotificationMicrosoftTeams` | `{name}` | `teams-alerts` |
| `ProjectNotification` | `{project}:{type}:{name}` | `my-project:slack:deploy-alerts` |

### Import Examples

```bash
# Import a Slack notification
pulumi import lagoon:lagoon:NotificationSlack my-slack deploy-alerts

# Import an Email notification
pulumi import lagoon:lagoon:NotificationEmail my-email ops-team

# Import a project notification association
pulumi import lagoon:lagoon:ProjectNotification my-assoc my-project:slack:deploy-alerts
```

After importing, add the corresponding resource definition to your Pulumi code:

```python
import pulumi
from pulumi_lagoon import NotificationSlack, NotificationSlackArgs

# After importing "deploy-alerts"
slack_alerts = NotificationSlack("my-slack",
    NotificationSlackArgs(
        name="deploy-alerts",
        webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
        channel="#deployments",
    ),
    opts=pulumi.ResourceOptions(import_="deploy-alerts")
)
```

## Validation Rules

### Notification Names
- Must start with a letter
- Can contain letters, numbers, hyphens, and underscores
- Maximum 100 characters

### Webhook URLs
- Must use HTTPS (HTTP not allowed for security)
- Must be a valid URL format

### Email Addresses
- Must be a valid email format (user@domain.tld)

### Notification Types
Valid values for `notification_type` in `ProjectNotification`:
- `slack`
- `rocketchat`
- `email`
- `microsoftteams`

## Troubleshooting

### Notification Not Triggering

1. **Verify webhook URL is correct** - Test the webhook independently
2. **Check project association** - Use `pulumi stack output` to verify the notification is linked
3. **Review Lagoon logs** - Check the Lagoon API logs for notification delivery errors

### Import Fails

1. **Verify notification exists** - Use the Lagoon CLI to confirm the notification name
2. **Check spelling** - Notification names are case-sensitive
3. **Verify project association format** - Use `project:type:name` format

### Permission Errors

1. **Verify API token** - Ensure your token has permission to manage notifications
2. **Check project access** - You need project admin access to link notifications

## Related Resources

- [Lagoon Notifications Documentation](https://docs.lagoon.sh/using-lagoon-advanced/notifications/)
- [Slack Incoming Webhooks](https://api.slack.com/messaging/webhooks)
- [Microsoft Teams Webhooks](https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook)
