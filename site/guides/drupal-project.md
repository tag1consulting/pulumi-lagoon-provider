---
title: Managing a Drupal Project
parent: Guides
nav_order: 2
---

# Managing a Drupal Project

This guide walks through the complete lifecycle of a Drupal site managed with the Lagoon provider: creating the project, wiring up the deploy key, configuring environments, adding variables, and setting up deployment notifications.

For a minimal working reference, see `examples/simple-project/` in the repository.

## 1. Create the Project

Start by declaring the project with its Git URL and the ID of the deploy target cluster where environments will run. Use `lagoon list deploy-targets` with the Lagoon CLI to find your deploy target ID.

```python
import pulumi
import pulumi_lagoon as lagoon

project = lagoon.Project("my-drupal-site",
    lagoon.ProjectArgs(
        name="my-drupal-site",
        git_url="git@github.com:myorg/my-drupal-site.git",
        deploytarget_id=1,
        production_environment="main",
        branches="^(main|develop)$",
        auto_idle=1,
        storage_calc=1,
    )
)

pulumi.export("project_id", project.lagoon_id)
pulumi.export("deploy_key", project.public_key)
```

Run `pulumi up` to create the project and retrieve the deploy key.

## 2. Add the Deploy Key to Your Git Host

Lagoon generates an SSH deploy key when a project is created. The `public_key` output contains this key. Add it as a read-only deploy key on your Git repository so Lagoon can clone the code during builds.

```bash
# Retrieve the deploy key after pulumi up
pulumi stack output deploy_key
```

Copy the output and add it to your repository's deploy keys:
- **GitHub**: Settings > Deploy keys > Add deploy key (read-only)
- **GitLab**: Settings > Repository > Deploy keys
- **Bitbucket**: Repository settings > Access keys

{: .important }
> Deployments will fail with a clone error until the deploy key is registered with your Git host. This step cannot be automated through the Lagoon provider — it requires a separate API call to your Git hosting platform.

## 3. Create Environments

Declare your production and development environments. The environment `name` must match the Git branch name.

```python
prod_env = lagoon.Environment("prod-env",
    lagoon.EnvironmentArgs(
        name="main",
        project_id=project.lagoon_id,
        deploy_type="branch",
        environment_type="production",
    )
)

dev_env = lagoon.Environment("dev-env",
    lagoon.EnvironmentArgs(
        name="develop",
        project_id=project.lagoon_id,
        deploy_type="branch",
        environment_type="development",
        auto_idle=1,
    )
)

pulumi.export("prod_url", prod_env.route)
pulumi.export("dev_url", dev_env.route)
```

{: .tip }
> Set `auto_idle=1` on development environments to let Lagoon scale idle pods to zero, reducing cluster resource usage between deployments.

## 4. Add Variables

Use `Variable` resources to set environment-specific configuration. Variables can be scoped to a single environment or to the entire project. Project-level variables apply to all environments unless overridden at the environment level.

```python
# Project-level variables (all environments)
db_host = lagoon.Variable("db-host",
    lagoon.VariableArgs(
        name="DB_HOST",
        value="db.internal.example.com",
        project_id=project.lagoon_id,
        scope="runtime",
    )
)

composer_auth = lagoon.Variable("composer-auth",
    lagoon.VariableArgs(
        name="COMPOSER_AUTH",
        value='{"http-basic":{"repo.example.com":{"username":"user","password":"<secret-redacted>"}}}',
        project_id=project.lagoon_id,
        scope="build",
    )
)

# Environment-level override: production database host
prod_db_host = lagoon.Variable("prod-db-host",
    lagoon.VariableArgs(
        name="DB_HOST",
        value="db-prod.internal.example.com",
        project_id=project.lagoon_id,
        environment_id=prod_env.lagoon_id,
        scope="runtime",
    )
)

# API key available at both build and runtime
api_key = lagoon.Variable("drupal-api-key",
    lagoon.VariableArgs(
        name="EXTERNAL_API_KEY",
        value="<secret-redacted>",
        project_id=project.lagoon_id,
        scope="global",
    )
)
```

| Scope | When available |
|-------|---------------|
| `build` | During the container build phase only |
| `runtime` | During container execution only |
| `global` | During both build and runtime |

## 5. Set Up Slack Notifications

Create a Slack notification configuration and link it to the project. Lagoon will send deployment status messages to the configured channel.

```python
slack_notify = lagoon.NotificationSlack("deploy-alerts",
    lagoon.NotificationSlackArgs(
        name="my-drupal-site-deployments",
        webhook="https://hooks.example.com/services/YOUR/SLACK/WEBHOOK",
        channel="#deployments",
    )
)
```

## 6. Link Notifications to the Project

A `ProjectNotification` attaches a notification to a project. You need one `ProjectNotification` per notification channel.

```python
project_slack = lagoon.ProjectNotification("project-slack",
    lagoon.ProjectNotificationArgs(
        project_name=project.name,
        notification_type="slack",
        notification_name=slack_notify.name,
    )
)
```

## Complete Example

Putting it all together:

```python
import pulumi
import pulumi_lagoon as lagoon

# Project
project = lagoon.Project("my-drupal-site",
    lagoon.ProjectArgs(
        name="my-drupal-site",
        git_url="git@github.com:myorg/my-drupal-site.git",
        deploytarget_id=1,
        production_environment="main",
        branches="^(main|develop)$",
        auto_idle=1,
    )
)

# Environments
prod_env = lagoon.Environment("prod-env",
    lagoon.EnvironmentArgs(
        name="main",
        project_id=project.lagoon_id,
        deploy_type="branch",
        environment_type="production",
    )
)

dev_env = lagoon.Environment("dev-env",
    lagoon.EnvironmentArgs(
        name="develop",
        project_id=project.lagoon_id,
        deploy_type="branch",
        environment_type="development",
        auto_idle=1,
    )
)

# Variables
lagoon.Variable("db-host",
    lagoon.VariableArgs(
        name="DB_HOST",
        value="db.internal.example.com",
        project_id=project.lagoon_id,
        scope="runtime",
    )
)

lagoon.Variable("composer-auth",
    lagoon.VariableArgs(
        name="COMPOSER_AUTH",
        value='{"http-basic":{"repo.example.com":{"username":"user","password":"<secret-redacted>"}}}',
        project_id=project.lagoon_id,
        scope="build",
    )
)

# Notifications
slack_notify = lagoon.NotificationSlack("deploy-alerts",
    lagoon.NotificationSlackArgs(
        name="my-drupal-site-deployments",
        webhook="https://hooks.example.com/services/YOUR/SLACK/WEBHOOK",
        channel="#deployments",
    )
)

lagoon.ProjectNotification("project-slack",
    lagoon.ProjectNotificationArgs(
        project_name=project.name,
        notification_type="slack",
        notification_name=slack_notify.name,
    )
)

# Outputs
pulumi.export("project_id", project.lagoon_id)
pulumi.export("deploy_key", project.public_key)
pulumi.export("prod_url", prod_env.route)
pulumi.export("dev_url", dev_env.route)
```

After `pulumi up`:

1. Copy the `deploy_key` output and add it to your GitHub/GitLab/Bitbucket repository.
2. Push a commit to the `main` or `develop` branch to trigger the first deployment.
3. Watch for deployment notifications in your Slack channel.
