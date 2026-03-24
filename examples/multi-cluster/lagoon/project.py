"""Example Lagoon project setup for multi-cluster deployment.

This module demonstrates how to:
1. Register Kubernetes clusters as deploy targets in Lagoon
2. Create a project with deploy target configurations
3. Route production branches to prod cluster and dev branches to nonprod cluster
"""

import re
from dataclasses import dataclass
from typing import Optional

import pulumi
import pulumi_lagoon
from config import DomainConfig
from pulumi_lagoon.lagoon import (
    DeployTarget,
    DeployTargetArgs,
    DeployTargetConfig,
    DeployTargetConfigArgs,
    Project,
    ProjectArgs,
)


@dataclass
class DeployTargetPair:
    """Outputs from creating deploy target pair for prod/nonprod clusters."""

    prod_target: DeployTarget
    nonprod_target: DeployTarget


@dataclass
class ExampleProjectOutputs:
    """Outputs from example project creation."""

    project: Project
    prod_config: DeployTargetConfig
    nonprod_config: DeployTargetConfig
    project_id: pulumi.Output[int]
    project_name: str


def create_deploy_targets(
    name: str,
    prod_cluster_name: str,
    nonprod_cluster_name: str,
    domain_config: DomainConfig,
    ssh_host: str,
    ssh_port: str = "22",
    lagoon_provider: Optional[pulumi_lagoon.Provider] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> DeployTargetPair:
    """
    Register both Kind clusters as deploy targets in Lagoon.

    This creates entries in Lagoon's Kubernetes registry that allow
    projects to be deployed to these clusters.

    Args:
        name: Pulumi resource name prefix
        prod_cluster_name: Name for the production deploy target
        nonprod_cluster_name: Name for the non-production deploy target
        domain_config: Domain configuration for router patterns
        ssh_host: SSH host for Lagoon builds (the SSH service)
        ssh_port: SSH port (default: "22")
        lagoon_provider: Native Lagoon provider instance
        opts: Pulumi resource options

    Returns:
        DeployTargetPair with both deploy targets
    """
    # Merge provider into opts
    provider_opts = pulumi.ResourceOptions(provider=lagoon_provider) if lagoon_provider else None
    merged_opts = pulumi.ResourceOptions.merge(opts, provider_opts) if provider_opts else opts

    # Production deploy target
    # In a real setup, console_url would be the actual Kubernetes API endpoint
    # For Kind clusters, we use the internal cluster URL
    prod_target = DeployTarget(
        f"{name}-prod-target",
        args=DeployTargetArgs(
            name=prod_cluster_name,
            console_url="https://kubernetes.default.svc",  # Internal K8s API
            cloud_provider="kind",
            cloud_region="local",
            ssh_host=ssh_host,
            ssh_port=ssh_port,
            # Router pattern determines how routes are generated
            # Format: ${environment}.${project}.${cluster-domain}
            router_pattern=f"${{environment}}.${{project}}.{domain_config.base}",
        ),
        opts=merged_opts,
    )

    # Non-production deploy target
    nonprod_target = DeployTarget(
        f"{name}-nonprod-target",
        args=DeployTargetArgs(
            name=nonprod_cluster_name,
            console_url="https://kubernetes.default.svc",  # Internal K8s API
            cloud_provider="kind",
            cloud_region="local",
            ssh_host=ssh_host,
            ssh_port=ssh_port,
            router_pattern=f"${{environment}}.${{project}}.{domain_config.base}",
        ),
        opts=merged_opts,
    )

    return DeployTargetPair(
        prod_target=prod_target,
        nonprod_target=nonprod_target,
    )


def create_example_drupal_project(
    name: str,
    deploy_targets: DeployTargetPair,
    git_url: str = "https://github.com/lagoon-examples/drupal-base.git",
    production_environment: str = "main",
    lagoon_provider: Optional[pulumi_lagoon.Provider] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> ExampleProjectOutputs:
    """
    Create an example Drupal project with multi-cluster routing.

    This creates a Lagoon project configured to:
    - Deploy 'main' branch to the production cluster
    - Deploy all other branches and PRs to the nonprod cluster

    Args:
        name: Pulumi resource name prefix and project name
        deploy_targets: The deploy target pair (prod/nonprod)
        git_url: Git repository URL (default: Lagoon's Drupal example)
        production_environment: Name of the production branch (default: "main")
        lagoon_provider: Native Lagoon provider instance
        opts: Pulumi resource options

    Returns:
        ExampleProjectOutputs with project and configuration resources
    """
    # Merge provider into opts
    provider_opts = pulumi.ResourceOptions(provider=lagoon_provider) if lagoon_provider else None

    # Create the project
    # The deploytarget_id here is the "default" target, but deploy target
    # configurations will override this for specific branches/PRs
    project = Project(
        f"{name}-project",
        args=ProjectArgs(
            name=name,
            git_url=git_url,
            # Use prod target as default (required by Lagoon)
            deploytarget_id=deploy_targets.prod_target.lagoon_id,
            production_environment=production_environment,
            # Branch pattern - which branches can be deployed
            branches=f"^({re.escape(production_environment)}|develop|feature/.*)$",
            # PR pattern - which PRs can be deployed
            pullrequests=".*",
        ),
        opts=pulumi.ResourceOptions.merge(
            pulumi.ResourceOptions.merge(
                opts,
                pulumi.ResourceOptions(
                    depends_on=[deploy_targets.prod_target, deploy_targets.nonprod_target],
                ),
            ),
            provider_opts,
        ),
    )

    # Deploy target configuration for production (main branch)
    # Higher weight = higher priority
    prod_config = DeployTargetConfig(
        f"{name}-prod-routing",
        args=DeployTargetConfigArgs(
            project_id=project.lagoon_id,
            deploy_target_id=deploy_targets.prod_target.lagoon_id,
            branches=production_environment,  # Only match 'main' (or configured prod branch)
            pullrequests="false",  # Production doesn't accept PRs
            weight=10,  # Higher priority - matches first for 'main'
        ),
        opts=pulumi.ResourceOptions.merge(
            pulumi.ResourceOptions.merge(
                opts,
                pulumi.ResourceOptions(depends_on=[project]),
            ),
            provider_opts,
        ),
    )

    # Deploy target configuration for non-production (all other branches + PRs)
    # Lower weight = fallback when prod config doesn't match
    nonprod_config = DeployTargetConfig(
        f"{name}-nonprod-routing",
        args=DeployTargetConfigArgs(
            project_id=project.lagoon_id,
            deploy_target_id=deploy_targets.nonprod_target.lagoon_id,
            branches=".*",  # Match all branches (fallback)
            pullrequests="true",  # Accept all PRs
            weight=1,  # Lower priority - only used when prod config doesn't match
        ),
        opts=pulumi.ResourceOptions.merge(
            pulumi.ResourceOptions.merge(
                opts,
                pulumi.ResourceOptions(depends_on=[project]),
            ),
            provider_opts,
        ),
    )

    return ExampleProjectOutputs(
        project=project,
        prod_config=prod_config,
        nonprod_config=nonprod_config,
        project_id=project.lagoon_id,
        project_name=name,
    )
