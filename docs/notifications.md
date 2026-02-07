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

### LagoonNotificationSlack

Creates a Slack notification configuration.

```python
import pulumi_lagoon as lagoon

slack_alerts = lagoon.LagoonNotificationSlack("deploy-alerts",
    lagoon.LagoonNotificationSlackArgs(
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

### LagoonNotificationRocketChat

Creates a RocketChat notification configuration.

```python
import pulumi_lagoon as lagoon

rocketchat_alerts = lagoon.LagoonNotificationRocketChat("team-chat",
    lagoon.LagoonNotificationRocketChatArgs(
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

### LagoonNotificationEmail

Creates an Email notification configuration.

```python
import pulumi_lagoon as lagoon

email_ops = lagoon.LagoonNotificationEmail("ops-team",
    lagoon.LagoonNotificationEmailArgs(
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

### LagoonNotificationMicrosoftTeams

Creates a Microsoft Teams notification configuration.

```python
import pulumi_lagoon as lagoon

teams_alerts = lagoon.LagoonNotificationMicrosoftTeams("teams-alerts",
    lagoon.LagoonNotificationMicrosoftTeamsArgs(
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

### LagoonProjectNotification

Links a notification to a project. This enables the project to receive notifications.

```python
import pulumi
import pulumi_lagoon as lagoon

# Create a project
project = lagoon.LagoonProject("my-site",
    lagoon.LagoonProjectArgs(
        name="my-site",
        git_url="git@github.com:example/my-site.git",
        deploytarget_id=1,
    )
)

# Create a Slack notification
slack_alerts = lagoon.LagoonNotificationSlack("deploy-alerts",
    lagoon.LagoonNotificationSlackArgs(
        name="deploy-alerts",
        webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
        channel="#deployments",
    )
)

# Link the notification to the project
project_notification = lagoon.LagoonProjectNotification("project-slack",
    lagoon.LagoonProjectNotificationArgs(
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
import pulumi_lagoon as lagoon

# Create the project
project = lagoon.LagoonProject("my-drupal-site",
    lagoon.LagoonProjectArgs(
        name="my-drupal-site",
        git_url="git@github.com:example/my-drupal-site.git",
        deploytarget_id=1,
        production_environment="main",
        branches="^(main|develop|stage)$",
    )
)

# === Create Notifications ===

# Slack for deployment notifications
slack_deploys = lagoon.LagoonNotificationSlack("slack-deploys",
    lagoon.LagoonNotificationSlackArgs(
        name="drupal-deploys",
        webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
        channel="#drupal-deployments",
    )
)

# Email for critical alerts
email_alerts = lagoon.LagoonNotificationEmail("email-alerts",
    lagoon.LagoonNotificationEmailArgs(
        name="critical-alerts",
        email_address="alerts@example.com",
    )
)

# Microsoft Teams for the dev team
teams_dev = lagoon.LagoonNotificationMicrosoftTeams("teams-dev",
    lagoon.LagoonNotificationMicrosoftTeamsArgs(
        name="dev-team-alerts",
        webhook="https://outlook.office.com/webhook/xxx/IncomingWebhook/yyy",
    )
)

# === Link Notifications to Project ===

project_slack = lagoon.LagoonProjectNotification("project-slack",
    lagoon.LagoonProjectNotificationArgs(
        project_name=project.name,
        notification_type="slack",
        notification_name=slack_deploys.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, slack_deploys])
)

project_email = lagoon.LagoonProjectNotification("project-email",
    lagoon.LagoonProjectNotificationArgs(
        project_name=project.name,
        notification_type="email",
        notification_name=email_alerts.name,
    ),
    opts=pulumi.ResourceOptions(depends_on=[project, email_alerts])
)

project_teams = lagoon.LagoonProjectNotification("project-teams",
    lagoon.LagoonProjectNotificationArgs(
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
import pulumi_lagoon as lagoon

# Create a shared Slack notification
shared_slack = lagoon.LagoonNotificationSlack("shared-slack",
    lagoon.LagoonNotificationSlackArgs(
        name="team-deploys",
        webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
        channel="#all-deployments",
    )
)

# Create multiple projects
project_a = lagoon.LagoonProject("project-a", ...)
project_b = lagoon.LagoonProject("project-b", ...)
project_c = lagoon.LagoonProject("project-c", ...)

# Link the same notification to all projects
for idx, project in enumerate([project_a, project_b, project_c]):
    lagoon.LagoonProjectNotification(f"notification-{idx}",
        lagoon.LagoonProjectNotificationArgs(
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
| `LagoonNotificationSlack` | `{name}` | `deploy-alerts` |
| `LagoonNotificationRocketChat` | `{name}` | `team-chat` |
| `LagoonNotificationEmail` | `{name}` | `ops-team` |
| `LagoonNotificationMicrosoftTeams` | `{name}` | `teams-alerts` |
| `LagoonProjectNotification` | `{project}:{type}:{name}` | `my-project:slack:deploy-alerts` |

### Import Examples

```bash
# Import a Slack notification
pulumi import lagoon:index:NotificationSlack my-slack deploy-alerts

# Import an Email notification
pulumi import lagoon:index:NotificationEmail my-email ops-team

# Import a project notification association
pulumi import lagoon:index:ProjectNotification my-assoc my-project:slack:deploy-alerts
```

After importing, add the corresponding resource definition to your Pulumi code:

```python
# After importing "deploy-alerts"
slack_alerts = lagoon.LagoonNotificationSlack("my-slack",
    lagoon.LagoonNotificationSlackArgs(
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
Valid values for `notification_type` in `LagoonProjectNotification`:
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
