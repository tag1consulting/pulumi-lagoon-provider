---
layout: home
title: Home
nav_exclude: true
hero_title: "Pulumi Lagoon Provider"
hero_tagline: "Manage Lagoon hosting platform resources as infrastructure-as-code. Native SDKs for Python, TypeScript, Go, and C#."
---

<div class="features">
  <div class="feature">
    <h3>18 Resources</h3>
    <p>Manage projects, environments, variables, deploy targets, notifications, routes, tasks, users, and groups declaratively.</p>
  </div>
  <div class="feature">
    <h3>4 Language SDKs</h3>
    <p>Native support for Python, TypeScript/JavaScript, Go, and .NET/C#. Use your preferred language and toolchain.</p>
  </div>
  <div class="feature">
    <h3>Import Existing Infrastructure</h3>
    <p>Bring existing Lagoon resources under Pulumi management with <code>pulumi import</code>.</p>
  </div>
  <div class="feature">
    <h3>Multi-Cluster Ready</h3>
    <p>Route deployments across production and non-production Kubernetes clusters with deploy target configurations.</p>
  </div>
</div>

## Quick Start

Create a Lagoon project with a single resource declaration:

```python
import pulumi
import pulumi_lagoon as lagoon

project = lagoon.Project("my-site",
    lagoon.ProjectArgs(
        name="my-drupal-site",
        git_url="git@github.com:org/repo.git",
        deploytarget_id=1,
        production_environment="main",
    )
)

pulumi.export("deploy_key", project.public_key)
```

Run `pulumi up` to create the project in Lagoon. The exported `deploy_key` is the SSH deploy key Lagoon generates — add it to your Git repository to enable deployments.

## Learn More

- [Getting Started](getting-started/) — Installation, configuration, and your first deployment
- [Resources](resources/) — Full reference for all 18 provider resources
- [Guides](guides/) — Task-oriented walkthroughs for common scenarios
- [Examples](examples/) — Complete working examples including multi-cluster deployments
