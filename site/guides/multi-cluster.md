---
title: Multi-Cluster Deployments
parent: Guides
nav_order: 1
---

# Multi-Cluster Deployments

Deploying Lagoon across multiple Kubernetes clusters gives you hard isolation between production and non-production workloads, independent scaling, and clearer cost attribution. This guide covers the architecture and shows how to register deploy targets and route branches to the right cluster using the Lagoon provider.

## Architecture Overview

A typical multi-cluster setup has two clusters:

- **Production cluster** — runs Lagoon Core (API, UI, message broker), Lagoon Remote, and Harbor registry. Hosts the `main` branch and any other production environments.
- **Nonprod cluster** — runs Lagoon Remote only, connected to the production core. Hosts feature branches, pull request environments, and the `develop` branch.

This separation means a misbehaving nonprod deployment cannot affect the production API or database, and you can apply stricter resource quotas to nonprod without impacting production uptime.

{: .note }
> Lagoon Core should run on only one cluster. The nonprod cluster runs Lagoon Remote and registers itself as a deploy target against the production core API.

## Register Deploy Targets

Each cluster must be registered as a `DeployTarget` in Lagoon. The deploy target record tells Lagoon where to send build jobs and how to reach the cluster's API.

```python
import pulumi
import pulumi_lagoon as lagoon

# Production cluster — also hosts Lagoon Core
prod_target = lagoon.DeployTarget("prod-cluster",
    lagoon.DeployTargetArgs(
        name="prod",
        console_url="https://kubernetes.prod.example.com",
        token="<prod-cluster-token>",
        router_pattern="${environment}.${project}.prod.example.com",
        ssh_host="ssh.prod.example.com",
        ssh_port="22",
        build_image="uselagoon/build-deploy-tool:latest",
    )
)

# Nonprod cluster — remote only
nonprod_target = lagoon.DeployTarget("nonprod-cluster",
    lagoon.DeployTargetArgs(
        name="nonprod",
        console_url="https://kubernetes.nonprod.example.com",
        token="<nonprod-cluster-token>",
        router_pattern="${environment}.${project}.nonprod.example.com",
        ssh_host="ssh.nonprod.example.com",
        ssh_port="22",
        build_image="uselagoon/build-deploy-tool:latest",
    )
)

pulumi.export("prod_target_id", prod_target.lagoon_id)
pulumi.export("nonprod_target_id", nonprod_target.lagoon_id)
```

## Configure Branch Routing

`DeployTargetConfig` attaches routing rules to a project. Rules are evaluated in order; the first match wins. Use branch regex patterns to send production branches to the prod cluster and everything else to nonprod.

```python
import pulumi
import pulumi_lagoon as lagoon

project = lagoon.Project("my-site",
    lagoon.ProjectArgs(
        name="my-site",
        git_url="git@github.com:myorg/my-site.git",
        deploytarget_id=prod_target.lagoon_id,   # default target
        production_environment="main",
        branches=".*",        # allow all branches
        pullrequests=".*",    # allow all PRs
    )
)

# Route 'main' branch to production cluster
prod_routing = lagoon.DeployTargetConfig("prod-routing",
    lagoon.DeployTargetConfigArgs(
        project_id=project.lagoon_id,
        deploytarget=prod_target.lagoon_id,
        branches="^main$",
        pullrequests="false",
        weight=100,
    )
)

# Route everything else to nonprod cluster
nonprod_routing = lagoon.DeployTargetConfig("nonprod-routing",
    lagoon.DeployTargetConfigArgs(
        project_id=project.lagoon_id,
        deploytarget=nonprod_target.lagoon_id,
        branches="^(?!main$).*",   # any branch that is not 'main'
        pullrequests=".*",
        weight=1,
    )
)
```

{: .tip }
> Use the `weight` field to control evaluation order when multiple rules could match. Higher weight rules are evaluated first.

## Best Practices

**Security isolation**

Apply restrictive Kubernetes `NetworkPolicy` on the nonprod cluster so that nonprod workloads cannot reach production databases or internal services. The clusters share only the Lagoon Core API (over HTTPS); they should not share a VPC subnet or internal DNS namespace.

**Cost management**

Enable auto-idle (`auto_idle=1`) on all nonprod environments. Lagoon will scale idle deployments to zero after a configurable period, substantially reducing compute costs for infrequently accessed feature branches.

**Network connectivity**

The nonprod cluster's Lagoon Remote must be able to reach the production core's RabbitMQ broker and API endpoint. Ensure firewall rules allow outbound TCP from the nonprod cluster to:
- Production RabbitMQ (typically port 5672 or 5671 with TLS)
- Production Lagoon API (typically port 443)

The production cluster does not need inbound access to the nonprod cluster.

**Deploy target tokens**

Store the Kubernetes API tokens for each cluster as Pulumi secrets:

```bash
pulumi config set --secret lagoon:prodClusterToken "$(kubectl --context prod-ctx get secret ...)"
pulumi config set --secret lagoon:nonprodClusterToken "$(kubectl --context nonprod-ctx get secret ...)"
```

Reference them in your program via `pulumi.Config` rather than hardcoding them.

## Full Working Example

The `examples/multi-cluster/` directory in the repository contains a complete working deployment that spins up two Kind clusters (Kubernetes in Docker) and installs the full Lagoon stack. See the [Multi-Cluster example page](../examples/multi-cluster/) for setup instructions.
