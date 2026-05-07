---
title: Tasks
parent: Resources
nav_order: 5
---

# Tasks

The `Task` resource manages advanced task definitions in Lagoon. Tasks are on-demand commands or container executions that users can trigger from the Lagoon UI or API. They can be scoped to a project, an environment, a group, or made system-wide (platform admin only).

---

## Task

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Task definition name |
| `type` | string | Yes | Task type: `command` or `image` |
| `service` | string | Yes | Service container name to run the task in |
| `command` | string | No | Command to execute — required when `type` is `command` |
| `image` | string | No | Container image to run — required when `type` is `image` |
| `projectId` | int | No | Scope the task to a specific project |
| `environmentId` | int | No | Scope the task to a specific environment |
| `groupName` | string | No | Scope the task to a specific group |
| `systemWide` | bool | No | Make the task available system-wide (platform admin only) |
| `permission` | string | No | Minimum permission required to run: `guest`, `developer`, or `maintainer` |
| `description` | string | No | Human-readable description shown in the Lagoon UI |
| `confirmationText` | string | No | Text displayed to the user before running the task (requires confirmation) |
| `arguments` | list | No | Input argument definitions — each with `name`, `displayName`, and `type` |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| `lagoonId` | int | Lagoon internal task definition ID |

### Import

Import using the numeric task definition ID:

```bash
pulumi import lagoon:lagoon:Task drush-cache-clear 42
```

### Examples

<div class="code-tabs" markdown="0">
  <input type="radio" id="task-example-python" name="task-example" checked>
  <label for="task-example-python">Python</label>
  <input type="radio" id="task-example-ts" name="task-example">
  <label for="task-example-ts">TypeScript</label>
  <input type="radio" id="task-example-go" name="task-example">
  <label for="task-example-go">Go</label>
  <input type="radio" id="task-example-csharp" name="task-example">
  <label for="task-example-csharp">C#</label>
  <div class="tab-content" markdown="1">

**Example 1: Drush cache-clear command task**

```python
import pulumi
import pulumi_lagoon as lagoon

cache_clear = lagoon.Task("drush-cache-clear",
    lagoon.TaskArgs(
        name="Drush Cache Clear",
        type="command",
        service="cli",
        command="drush cr",
        project_id=123,
        permission="developer",
        description="Clear all Drupal caches using Drush.",
    )
)

pulumi.export("cache_clear_task_id", cache_clear.lagoon_id)
```

**Example 2: Image-based database operations task**

```python
import pulumi
import pulumi_lagoon as lagoon

db_snapshot = lagoon.Task("db-snapshot",
    lagoon.TaskArgs(
        name="Database Snapshot",
        type="image",
        service="cli",
        image="uselagoon/db-tools:v1.0.0",
        project_id=123,
        permission="maintainer",
        description="Create a snapshot of the production database.",
        confirmation_text="This will create a database snapshot. Are you sure?",
    )
)
```

  </div>
  <div class="tab-content" markdown="1">

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as lagoon from "@tag1consulting/pulumi-lagoon";

const cacheClear = new lagoon.Task("drush-cache-clear", {
    name: "Drush Cache Clear",
    type: "command",
    service: "cli",
    command: "drush cr",
    projectId: 123,
    permission: "developer",
    description: "Clear all Drupal caches using Drush.",
});

export const cacheClearTaskId = cacheClear.lagoonId;
```

  </div>
  <div class="tab-content" markdown="1">

```go
package main

import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    lagoon "github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon/lagoon"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        cacheClear, err := lagoon.NewTask(ctx, "drush-cache-clear", &lagoon.TaskArgs{
            Name:        pulumi.String("Drush Cache Clear"),
            Type:        pulumi.String("command"),
            Service:     pulumi.String("cli"),
            Command:     pulumi.String("drush cr"),
            ProjectId:   pulumi.Int(123),
            Permission:  pulumi.String("developer"),
            Description: pulumi.String("Clear all Drupal caches using Drush."),
        })
        if err != nil {
            return err
        }

        ctx.Export("cacheClearTaskId", cacheClear.LagoonId)
        return nil
    })
}
```

  </div>
  <div class="tab-content" markdown="1">

```csharp
using Pulumi;
using Tag1Consulting.Lagoon.Lagoon;

return await Deployment.RunAsync(() =>
{
    var cacheClear = new Task("drush-cache-clear", new TaskArgs
    {
        Name = "Drush Cache Clear",
        Type = "command",
        Service = "cli",
        Command = "drush cr",
        ProjectId = 123,
        Permission = "developer",
        Description = "Clear all Drupal caches using Drush.",
    });

    return new Dictionary<string, object?>
    {
        ["cacheClearTaskId"] = cacheClear.LagoonId,
    };
});
```

  </div>
</div>
