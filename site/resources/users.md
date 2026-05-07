---
title: Users
parent: Resources
nav_order: 6
---

# Users

User resources manage Lagoon users and their access assignments. The three resources cover user lifecycle, group membership, and platform-level roles.

---

## User

A `User` manages a Lagoon user via the GraphQL API. The user's email address is the primary identifier and is force-new — changing it recreates the resource.

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `email` | string | Yes | User's email address (force-new) |
| `firstName` | string | No | User's first name |
| `lastName` | string | No | User's last name |
| `comment` | string | No | Optional comment about the user |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| `lagoonId` | string | Lagoon internal user ID |

### Import

Import using the user's email address:

```bash
pulumi import lagoon:lagoon:User alice alice@example.com
```

---

## UserGroupAssignment

A `UserGroupAssignment` assigns a user to a group with a specified role. Role changes are applied in-place using Lagoon's `addUserToGroup` upsert mutation. Changing `userEmail` or `groupName` triggers a replace.

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `userEmail` | string | Yes | User's email address (force-new) |
| `groupName` | string | Yes | Group name (force-new) |
| `role` | string | Yes | Role within the group: `GUEST`, `REPORTER`, `DEVELOPER`, `MAINTAINER`, or `OWNER` |

### Import

Import using the email address and group name, separated by a colon:

```bash
pulumi import lagoon:lagoon:UserGroupAssignment alice-dev-team alice@example.com:dev-team
```

{: .note }
> Role changes are applied in-place. Changing `userEmail` or `groupName` triggers a replace, which removes and re-adds the assignment.

---

## UserPlatformRole

A `UserPlatformRole` assigns a platform-level role to a Lagoon user. Both fields are force-new — changing either triggers a replace, which removes the old role and creates a new assignment.

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `userEmail` | string | Yes | User's email address (force-new) |
| `role` | string | Yes | Platform role: `OWNER` or `VIEWER` (force-new) |

### Import

Import using the email address and role, separated by a colon:

```bash
pulumi import lagoon:lagoon:UserPlatformRole alice-platform alice@example.com:VIEWER
```

---

## Examples

### Complete user setup: create, assign to group, assign platform role

<div class="code-tabs" markdown="0">
  <input type="radio" id="user-example-python" name="user-example" checked>
  <label for="user-example-python">Python</label>
  <input type="radio" id="user-example-ts" name="user-example">
  <label for="user-example-ts">TypeScript</label>
  <input type="radio" id="user-example-go" name="user-example">
  <label for="user-example-go">Go</label>
  <input type="radio" id="user-example-csharp" name="user-example">
  <label for="user-example-csharp">C#</label>
  <div class="tab-content" markdown="1">

```python
import pulumi
import pulumi_lagoon as lagoon

# Create the user
alice = lagoon.User("alice",
    lagoon.UserArgs(
        email="alice@example.com",
        first_name="Alice",
        last_name="Smith",
        comment="Lead developer",
    )
)

# Assign the user to a group as a developer
alice_dev = lagoon.UserGroupAssignment("alice-dev-team",
    lagoon.UserGroupAssignmentArgs(
        user_email="alice@example.com",
        group_name="dev-team",
        role="DEVELOPER",
    ),
    opts=pulumi.ResourceOptions(depends_on=[alice])
)

# Grant platform viewer access
alice_platform = lagoon.UserPlatformRole("alice-platform-viewer",
    lagoon.UserPlatformRoleArgs(
        user_email="alice@example.com",
        role="VIEWER",
    ),
    opts=pulumi.ResourceOptions(depends_on=[alice])
)

pulumi.export("alice_lagoon_id", alice.lagoon_id)
```

  </div>
  <div class="tab-content" markdown="1">

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as lagoon from "@tag1consulting/pulumi-lagoon";

const alice = new lagoon.User("alice", {
    email: "alice@example.com",
    firstName: "Alice",
    lastName: "Smith",
    comment: "Lead developer",
});

const aliceDev = new lagoon.UserGroupAssignment("alice-dev-team", {
    userEmail: "alice@example.com",
    groupName: "dev-team",
    role: "DEVELOPER",
}, { dependsOn: [alice] });

const alicePlatform = new lagoon.UserPlatformRole("alice-platform-viewer", {
    userEmail: "alice@example.com",
    role: "VIEWER",
}, { dependsOn: [alice] });

export const aliceLagoonId = alice.lagoonId;
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
        alice, err := lagoon.NewUser(ctx, "alice", &lagoon.UserArgs{
            Email:     pulumi.String("alice@example.com"),
            FirstName: pulumi.String("Alice"),
            LastName:  pulumi.String("Smith"),
            Comment:   pulumi.String("Lead developer"),
        })
        if err != nil {
            return err
        }

        _, err = lagoon.NewUserGroupAssignment(ctx, "alice-dev-team", &lagoon.UserGroupAssignmentArgs{
            UserEmail: pulumi.String("alice@example.com"),
            GroupName: pulumi.String("dev-team"),
            Role:      pulumi.String("DEVELOPER"),
        }, pulumi.DependsOn([]pulumi.Resource{alice}))
        if err != nil {
            return err
        }

        _, err = lagoon.NewUserPlatformRole(ctx, "alice-platform-viewer", &lagoon.UserPlatformRoleArgs{
            UserEmail: pulumi.String("alice@example.com"),
            Role:      pulumi.String("VIEWER"),
        }, pulumi.DependsOn([]pulumi.Resource{alice}))
        if err != nil {
            return err
        }

        ctx.Export("aliceLagoonId", alice.LagoonId)
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
    var alice = new User("alice", new UserArgs
    {
        Email = "alice@example.com",
        FirstName = "Alice",
        LastName = "Smith",
        Comment = "Lead developer",
    });

    var aliceDev = new UserGroupAssignment("alice-dev-team", new UserGroupAssignmentArgs
    {
        UserEmail = "alice@example.com",
        GroupName = "dev-team",
        Role = "DEVELOPER",
    }, new CustomResourceOptions { DependsOn = { alice } });

    var alicePlatform = new UserPlatformRole("alice-platform-viewer", new UserPlatformRoleArgs
    {
        UserEmail = "alice@example.com",
        Role = "VIEWER",
    }, new CustomResourceOptions { DependsOn = { alice } });

    return new Dictionary<string, object?>
    {
        ["aliceLagoonId"] = alice.LagoonId,
    };
});
```

  </div>
</div>
