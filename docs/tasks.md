# Managing Lagoon Tasks with Pulumi

This guide covers how to manage Lagoon advanced task definitions using the Pulumi Lagoon provider.

## Overview

Lagoon supports advanced task definitions that allow you to define reusable commands and container-based jobs. Tasks can be triggered on-demand via the Lagoon UI, API, or CLI.

**Task Types:**
- **Command tasks** - Execute a command in an existing service container
- **Image tasks** - Run a custom container image as a Kubernetes Job

**Task Scope:**
Tasks can be scoped to different levels:
- **Project-level** - Available to all environments in a project
- **Environment-level** - Available only to a specific environment
- **Group-level** - Available to all projects in a group
- **System-wide** - Available to all projects (platform admin only)

## Task Resource

The `Task` resource defines an advanced task definition that can be triggered on-demand.

### Command Task Example

```python
from pulumi_lagoon import Task, TaskArgs

drush_cache_rebuild = Task("drush-cr",
    TaskArgs(
        name="Drush Cache Rebuild",
        type="command",
        service="cli",
        command="drush cache-rebuild",
        project_id=123,
        permission="developer",
        description="Clear all Drupal caches using drush cr",
        confirmation_text="This will clear all caches. Continue?",
    )
)
```

### Image Task Example

```python
from pulumi_lagoon import Task, TaskArgs

database_backup = Task("db-backup",
    TaskArgs(
        name="Database Backup",
        type="image",
        image="uselagoon/db-backup:latest",
        service="cli",
        environment_id=456,
        permission="maintainer",
        description="Create a database backup and upload to S3",
        confirmation_text="Create a new database backup?",
        arguments=[
            {
                "name": "BACKUP_TYPE",
                "display_name": "Backup Type",
                "type": "STRING",
            },
            {
                "name": "SOURCE_ENV",
                "display_name": "Source Environment",
                "type": "ENVIRONMENT_SOURCE_NAME",
            },
        ],
    )
)
```

### Properties

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | `str` | Yes | Task definition name (displayed in UI) |
| `type` | `str` | Yes | Task type: `command` or `image` |
| `service` | `str` | Yes | Service container to run the task in |
| `command` | `str` | Conditional | Command to execute (required if `type='command'`) |
| `image` | `str` | Conditional | Container image (required if `type='image'`) |
| `project_id` | `int` | No | Project ID (for project-scoped tasks) |
| `environment_id` | `int` | No | Environment ID (for environment-scoped tasks) |
| `group_name` | `str` | No | Group name (for group-scoped tasks) |
| `system_wide` | `bool` | No | Make task available system-wide (admin only) |
| `permission` | `str` | No | Permission level: `guest`, `developer`, or `maintainer` (default: `developer`) |
| `description` | `str` | No | Task description (shown in UI and CLI) |
| `confirmation_text` | `str` | No | Confirmation prompt text (shown before execution) |
| `arguments` | `list` | No | List of argument definitions (see below) |

### Task Scoping Rules

Exactly one of the following should be set to define the task scope:
- `project_id` - Task available to all environments in the project
- `environment_id` - Task available only to the specified environment
- `group_name` - Task available to all projects in the group
- `system_wide=true` - Task available to all projects (admin only)

> **Note:** The scope constraint is enforced by the Lagoon API, not by the provider. If you set multiple scopes (or none), the API will return an error. The exact validation may vary by Lagoon version.

### Task Arguments

Tasks can define arguments that are prompted for when the task is triggered. Each argument is a dictionary with:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `str` | Yes | Argument name (environment variable name) |
| `display_name` | `str` | No | Display name in UI (defaults to `name`) |
| `type` | `str` | Yes | Argument type (see below) |

**Argument Types:**
- `STRING` - Free-form text input
- `ENVIRONMENT_SOURCE_NAME` - Dropdown of environment names in the project

### Outputs

- `lagoon_id` - The Lagoon internal task ID
- `name` - Task name
- `type` - Task type
- `service` - Service name
- `command` - Command (if command task)
- `image` - Image (if image task)
- All other input properties

### Resource ID

Numeric Lagoon task ID (e.g., `456`)

## Complete Examples

### Drush Commands for Drupal

```python
import pulumi
from pulumi_lagoon import Task, TaskArgs

# Drush cache rebuild
drush_cr = Task("drush-cr",
    TaskArgs(
        name="Drush Cache Rebuild",
        type="command",
        service="cli",
        command="drush cache-rebuild",
        project_id=123,
        permission="developer",
        description="Clear all Drupal caches",
        confirmation_text="Clear all caches?",
    )
)

# Drush SQL dump
drush_sql_dump = Task("drush-sql-dump",
    TaskArgs(
        name="Drush SQL Dump",
        type="command",
        service="cli",
        command="drush sql-dump --gzip > /tmp/db-backup-$(date +%Y%m%d-%H%M%S).sql.gz",
        project_id=123,
        permission="maintainer",
        description="Create a gzipped database backup",
        confirmation_text="Create database backup?",
    )
)

# Drush update database
drush_updb = Task("drush-updb",
    TaskArgs(
        name="Drush Update Database",
        type="command",
        service="cli",
        command="drush updatedb -y",
        project_id=123,
        permission="maintainer",
        description="Run Drupal database updates",
        confirmation_text="WARNING: This will apply pending database updates. Continue?",
    )
)

# Export task IDs
pulumi.export("drush_cr_id", drush_cr.lagoon_id)
pulumi.export("drush_sql_dump_id", drush_sql_dump.lagoon_id)
pulumi.export("drush_updb_id", drush_updb.lagoon_id)
```

### Database Sync with Arguments

```python
from pulumi_lagoon import Task, TaskArgs

db_sync = Task("db-sync",
    TaskArgs(
        name="Sync Database from Environment",
        type="command",
        service="cli",
        command="lagoon-sync sync mariadb -p ${SOURCE_PROJECT} -e ${SOURCE_ENV}",
        environment_id=456,
        permission="maintainer",
        description="Sync database from another environment",
        confirmation_text="This will OVERWRITE the current database. Continue?",
        arguments=[
            {
                "name": "SOURCE_PROJECT",
                "display_name": "Source Project",
                "type": "STRING",
            },
            {
                "name": "SOURCE_ENV",
                "display_name": "Source Environment",
                "type": "ENVIRONMENT_SOURCE_NAME",
            },
        ],
    )
)
```

### Image-Based Task with Custom Container

```python
from pulumi_lagoon import Task, TaskArgs

security_scan = Task("security-scan",
    TaskArgs(
        name="Security Vulnerability Scan",
        type="image",
        image="aquasec/trivy:latest",
        service="cli",
        project_id=123,
        permission="maintainer",
        description="Run Trivy security scan on the codebase",
        confirmation_text="Run security scan?",
    )
)
```

### System-Wide Admin Task

```python
from pulumi_lagoon import Task, TaskArgs

platform_health_check = Task("health-check",
    TaskArgs(
        name="Platform Health Check",
        type="command",
        service="cli",
        command="/usr/local/bin/health-check.sh",
        system_wide=True,
        permission="maintainer",
        description="Run platform-wide health checks (admin only)",
    )
)
```

## Permission Levels

| Permission | Who Can Trigger | Description |
|------------|----------------|-------------|
| `guest` | All project viewers | Read-only access users |
| `developer` | Developers and above | Standard development team members |
| `maintainer` | Maintainers and admins | Senior developers and project leads |

**Default:** `developer`

## Triggering Tasks

Once defined, tasks can be triggered via:

### Lagoon CLI

```bash
# List available tasks
lagoon list tasks --project my-project --environment main

# Run a task
lagoon run task --project my-project --environment main --task "Drush Cache Rebuild"

# Run task with arguments
lagoon run task --project my-project --environment main \
  --task "Sync Database from Environment" \
  --args SOURCE_PROJECT=my-project,SOURCE_ENV=production
```

### Lagoon UI

1. Navigate to the environment in the Lagoon UI
2. Click on "Tasks"
3. Select the task from the dropdown
4. Fill in any required arguments
5. Click "Run Task"

### GraphQL API

```graphql
mutation RunTask {
  taskDrushCacheClear(
    environment: 456
  ) {
    id
    name
    status
  }
}
```

## GraphQL API Reference

The provider uses these Lagoon GraphQL operations to manage task definitions.

**Create a task definition:**

```graphql
mutation AddAdvancedTaskDefinition($input: AddAdvancedTaskDefinitionInput!) {
    addAdvancedTaskDefinition(input: $input) {
        id
        name
        description
        type
        service
        command
        image
        permission
        confirmationText
        advancedTaskDefinitionArguments { id name displayName type }
        project { id name }
        environment { id name }
        groupName
    }
}
```

**Query a task definition by ID:**

```graphql
query AdvancedTaskDefinitionById($id: Int!) {
    advancedTaskDefinitionById(id: $id) {
        id
        name
        description
        type
        service
        command
        image
        permission
        confirmationText
        advancedTaskDefinitionArguments { id name displayName type }
        project { id name }
        environment { id name }
        groupName
    }
}
```

## Importing Existing Tasks

### Import ID Format

| Resource | Format | Example |
|----------|--------|---------|
| `Task` | `{numeric_id}` | `456` |

### Import Example

```bash
# Import an existing task
pulumi import lagoon:lagoon:Task drush-cr 456
```

After importing, add the corresponding resource definition to your Pulumi code:

```python
from pulumi_lagoon import Task, TaskArgs

# After importing task ID 456
drush_cr = Task("drush-cr",
    TaskArgs(
        name="Drush Cache Rebuild",
        type="command",
        service="cli",
        command="drush cache-rebuild",
        project_id=123,
        permission="developer",
    ),
    opts=pulumi.ResourceOptions(import_="456")
)
```

To find the task ID, use the Lagoon CLI:

```bash
lagoon list tasks --project my-project --environment main
```

## Validation Rules

### Task Names
- Must be unique within the task's scope
- Maximum 100 characters
- No special validation rules (can contain spaces, punctuation, etc.)

### Task Types
Valid values for `type`:
- `command` - Execute a command in a service container
- `image` - Run a custom container image

### Permission Levels
Valid values for `permission`:
- `guest`
- `developer`
- `maintainer`

### Service Names
- Must be a valid service name from the project's .lagoon.yml
- Common values: `cli`, `nginx`, `php`, `mariadb`

### Command Tasks
- Require `command` to be set
- Cannot have `image` set
- Command is executed in the specified service container's default shell

### Image Tasks
- Require `image` to be set
- Cannot have `command` set
- Image must be a valid Docker image reference

## Troubleshooting

### Task Not Appearing in UI

1. **Verify scope** - Check that the task is scoped correctly (project, environment, group, or system)
2. **Check permissions** - Ensure your user has the required permission level
3. **Refresh** - Task definitions are cached; try logging out and back in

### Task Execution Fails

1. **Service not found** - Verify the service name matches a service in .lagoon.yml
2. **Command not found** - Check that the command exists in the service container
3. **Image pull fails** - For image tasks, verify the image is accessible and the tag exists
4. **Permission denied** - Ensure the user triggering the task has sufficient permissions

### Import Fails

1. **Task ID not found** - Verify the task exists: `lagoon list tasks`
2. **Wrong format** - Use the numeric task ID (not the task name)

### Scope Conflicts

- **Scope conflict error from Lagoon API** - Set exactly one of: `project_id`, `environment_id`, `group_name`, or `system_wide=true`
- **System-wide requires admin** - Only platform admins can create system-wide tasks

## Best Practices

### Task Naming
- Use descriptive names that indicate what the task does
- Include tool names for clarity (e.g., "Drush Cache Rebuild" not just "Cache Rebuild")

### Confirmation Text
- Always add confirmation text for destructive operations
- Be explicit about what will happen (e.g., "This will OVERWRITE the database")

### Permissions
- Use `maintainer` for destructive operations (database syncs, updates)
- Use `developer` for safe operations (cache clears, log viewing)
- Rarely use `guest` (read-only operations only)

### Environment Variables
- Use arguments to make tasks reusable with different parameters
- Use `ENVIRONMENT_SOURCE_NAME` type for environment selection (provides dropdown)
- Document expected argument values in the description

### Image Tasks
- Pin image versions (`:v1.2.3`) rather than using `:latest`
- Test images locally before using in tasks
- Consider creating custom images with pre-installed tools

## Related Resources

- [Lagoon Advanced Tasks Documentation](https://docs.lagoon.sh/using-lagoon-advanced/advanced-tasks/)
- [Lagoon CLI](https://docs.lagoon.sh/using-lagoon-the-basics/lagoon-cli/)
- [Lagoon GraphQL API](https://api.lagoon.sh/graphql)
