---
title: Quick Start
parent: Getting Started
nav_order: 3
---

# Quick Start

This walkthrough gets you from zero to a running Lagoon project in about five minutes. By the end you will have a Lagoon project, an environment, and a runtime variable managed by Pulumi.

## Step 1: Install the SDK

#### Python

```bash
pip install pulumi-lagoon
```

See [Installation](installation/) for TypeScript, Go, and .NET instructions.

## Step 2: Create a New Pulumi Project

```bash
mkdir my-lagoon-infra && cd my-lagoon-infra
pulumi new python
```

For other runtimes, replace `python` with `typescript`, `go`, or `csharp`.

## Step 3: Configure the Provider

```bash
pulumi config set lagoon:apiUrl https://api.lagoon.example.com/graphql
pulumi config set lagoon:token <your-jwt-token> --secret
```

## Step 4: Write Your Infrastructure Code

Replace the generated entry point with the following. The example creates a project, a production environment, and a database host variable.

#### Python

```python
import pulumi
import pulumi_lagoon as lagoon

project = lagoon.Project("my-site",
    lagoon.ProjectArgs(
        name="my-drupal-site",
        git_url="git@github.com:org/repo.git",
        deploytarget_id=1,
        production_environment="main",
        branches="^(main|develop|stage)$",
    )
)

pulumi.export("deploy_key", project.public_key)

prod_env = lagoon.Environment("production",
    lagoon.EnvironmentArgs(
        name="main",
        project_id=project.lagoon_id,
        deploy_type="branch",
        deploy_base_ref="main",
        environment_type="production",
    )
)

db_host = lagoon.Variable("db-host",
    lagoon.VariableArgs(
        name="DATABASE_HOST",
        value="mysql.production.example.com",
        project_id=project.lagoon_id,
        environment_id=prod_env.lagoon_id,
        scope="runtime",
    )
)

pulumi.export("project_id", project.lagoon_id)
pulumi.export("production_url", prod_env.route)
```

#### TypeScript

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as lagoon from "@tag1consulting/pulumi-lagoon";

const project = new lagoon.Project("my-site", {
    name: "my-drupal-site",
    gitUrl: "git@github.com:org/repo.git",
    deploytargetId: 1,
    productionEnvironment: "main",
    branches: "^(main|develop|stage)$",
});

export const deployKey = project.publicKey;

const prodEnv = new lagoon.Environment("production", {
    name: "main",
    projectId: project.lagoonId,
    deployType: "branch",
    deployBaseRef: "main",
    environmentType: "production",
});

const dbHost = new lagoon.Variable("db-host", {
    name: "DATABASE_HOST",
    value: "mysql.production.example.com",
    projectId: project.lagoonId,
    environmentId: prodEnv.lagoonId,
    scope: "runtime",
});

export const projectId = project.lagoonId;
```

#### Go

```go
package main

import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    lagoon "github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon/lagoon"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        project, err := lagoon.NewProject(ctx, "my-site", &lagoon.ProjectArgs{
            Name:                  pulumi.String("my-drupal-site"),
            GitUrl:                pulumi.String("git@github.com:org/repo.git"),
            DeploytargetId:        pulumi.Int(1),
            ProductionEnvironment: pulumi.String("main"),
            Branches:              pulumi.StringPtrInput(pulumi.String("^(main|develop|stage)$")),
        })
        if err != nil {
            return err
        }

        ctx.Export("deployKey", project.PublicKey)

        prodEnv, err := lagoon.NewEnvironment(ctx, "production", &lagoon.EnvironmentArgs{
            Name:            pulumi.String("main"),
            ProjectId:       project.LagoonId,
            DeployType:      pulumi.String("branch"),
            DeployBaseRef:   pulumi.StringPtrInput(pulumi.String("main")),
            EnvironmentType: pulumi.String("production"),
        })
        if err != nil {
            return err
        }

        ctx.Export("projectId", project.LagoonId)
        _ = prodEnv
        return nil
    })
}
```

#### C#

```csharp
using System.Collections.Generic;
using Pulumi;
using Tag1Consulting.Lagoon.Lagoon;

return await Deployment.RunAsync(() =>
{
    var project = new Project("my-site", new ProjectArgs
    {
        Name = "my-drupal-site",
        GitUrl = "git@github.com:org/repo.git",
        DeploytargetId = 1,
        ProductionEnvironment = "main",
        Branches = "^(main|develop|stage)$",
    });

    var prodEnv = new Environment("production", new EnvironmentArgs
    {
        Name = "main",
        ProjectId = project.LagoonId,
        DeployType = "branch",
        DeployBaseRef = "main",
        EnvironmentType = "production",
    });

    return new Dictionary<string, object?>
    {
        ["projectId"] = project.LagoonId,
        ["deployKey"] = project.PublicKey,
    };
});
```

## Step 5: Deploy

```bash
pulumi up
```

Pulumi shows a preview of the resources it will create and prompts for confirmation. After you confirm, it creates the project, environment, and variable in Lagoon.

The `deploy_key` output is the SSH deploy key Lagoon generated for the project. Copy it and add it to your Git repository so that Lagoon can clone the code during deployments.

## Next Steps

- [Resources](../resources/) — Full reference for all 18 provider resources
- [Importing Resources](importing-resources/) — Bring existing Lagoon resources under Pulumi management
- [Guides](../guides/) — Task-oriented walkthroughs for common scenarios
