---
title: Notifications
parent: Resources
nav_order: 3
---

# Notifications

Notification resources manage the delivery of Lagoon deployment events to external services. The workflow is two-step: first create a notification configuration, then link it to one or more projects using `ProjectNotification`.

---

## NotificationSlack

A `NotificationSlack` configures a Slack incoming webhook as a notification channel.

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Unique notification name |
| `webhook` | string | Yes | Slack incoming webhook URL (stored as a secret) |
| `channel` | string | Yes | Slack channel to post to (e.g., `#deployments`) |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| `lagoonId` | int | Lagoon internal notification ID |

### Import

```bash
pulumi import lagoon:lagoon:NotificationSlack my-slack-notif my-slack-notif
```

---

## NotificationRocketChat

A `NotificationRocketChat` configures a RocketChat webhook as a notification channel.

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Unique notification name |
| `webhook` | string | Yes | RocketChat webhook URL (stored as a secret) |
| `channel` | string | Yes | RocketChat channel to post to |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| `lagoonId` | int | Lagoon internal notification ID |

### Import

```bash
pulumi import lagoon:lagoon:NotificationRocketChat my-rocketchat-notif my-rocketchat-notif
```

---

## NotificationEmail

A `NotificationEmail` configures an email address to receive deployment notifications.

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Unique notification name |
| `emailAddress` | string | Yes | Email address to send notifications to |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| `lagoonId` | int | Lagoon internal notification ID |

### Import

```bash
pulumi import lagoon:lagoon:NotificationEmail my-email-notif my-email-notif
```

---

## NotificationMicrosoftTeams

A `NotificationMicrosoftTeams` configures a Microsoft Teams incoming webhook as a notification channel.

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Unique notification name |
| `webhook` | string | Yes | Microsoft Teams webhook URL (stored as a secret) |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| `lagoonId` | int | Lagoon internal notification ID |

### Import

```bash
pulumi import lagoon:lagoon:NotificationMicrosoftTeams my-teams-notif my-teams-notif
```

---

## ProjectNotification

A `ProjectNotification` links an existing notification configuration to a project. One notification can be linked to multiple projects, and one project can have multiple notifications.

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `projectName` | string | Yes | Project name to attach the notification to |
| `notificationType` | string | Yes | Notification type: `slack`, `rocketchat`, `email`, or `microsoftteams` |
| `notificationName` | string | Yes | Name of the notification configuration to link |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| `projectId` | int | Lagoon internal ID of the linked project |

### Import

Import using the project name, notification type, and notification name:

```bash
pulumi import lagoon:lagoon:ProjectNotification my-project-slack my-site:slack:my-slack-notif
```

---

## Examples

### Creating a Slack notification and linking it to a project

<div class="code-tabs" markdown="0">
  <input type="radio" id="slack-notif-example-python" name="slack-notif-example" checked>
  <label for="slack-notif-example-python">Python</label>
  <input type="radio" id="slack-notif-example-ts" name="slack-notif-example">
  <label for="slack-notif-example-ts">TypeScript</label>
  <div class="tab-content" markdown="1">

```python
import pulumi
import pulumi_lagoon as lagoon

# Create a Slack notification channel
slack_notif = lagoon.NotificationSlack("deployments-slack",
    lagoon.NotificationSlackArgs(
        name="deployments-slack",
        webhook="https://hooks.example.com/services/YOUR/SLACK/WEBHOOK",
        channel="#deployments",
    )
)

# Create an email notification channel
email_notif = lagoon.NotificationEmail("ops-email",
    lagoon.NotificationEmailArgs(
        name="ops-email",
        email_address="ops@example.com",
    )
)

# Link the Slack notification to the project
lagoon.ProjectNotification("my-site-slack",
    lagoon.ProjectNotificationArgs(
        project_name="my-site",
        notification_type="slack",
        notification_name="deployments-slack",
    ),
    opts=pulumi.ResourceOptions(depends_on=[slack_notif])
)

# Link the email notification to the same project
lagoon.ProjectNotification("my-site-email",
    lagoon.ProjectNotificationArgs(
        project_name="my-site",
        notification_type="email",
        notification_name="ops-email",
    ),
    opts=pulumi.ResourceOptions(depends_on=[email_notif])
)
```

  </div>
  <div class="tab-content" markdown="1">

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as lagoon from "@tag1consulting/pulumi-lagoon";

const slackNotif = new lagoon.NotificationSlack("deployments-slack", {
    name: "deployments-slack",
    webhook: "https://hooks.example.com/services/YOUR/SLACK/WEBHOOK",
    channel: "#deployments",
});

const projectSlack = new lagoon.ProjectNotification("my-site-slack", {
    projectName: "my-site",
    notificationType: "slack",
    notificationName: "deployments-slack",
}, { dependsOn: [slackNotif] });
```

  </div>
</div>
