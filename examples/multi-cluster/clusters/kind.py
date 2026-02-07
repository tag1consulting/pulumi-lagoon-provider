"""Kind cluster creation and management.

This module provides functions for creating and managing Kind clusters
for the multi-cluster Lagoon example.
"""

from typing import Optional

import pulumi
import yaml
from config import KIND_NODE_IMAGE, ClusterConfig, ClusterOutputs
from pulumi_command import local


def generate_kind_config(cluster_config: ClusterConfig) -> str:
    """Generate Kind cluster configuration YAML.

    Args:
        cluster_config: Cluster configuration dataclass

    Returns:
        YAML string for Kind cluster configuration
    """
    config = {
        "kind": "Cluster",
        "apiVersion": "kind.x-k8s.io/v1alpha4",
        "name": cluster_config.name,
        "nodes": [
            {
                "role": "control-plane",
                "kubeadmConfigPatches": [
                    """
kind: InitConfiguration
nodeRegistration:
  kubeletExtraArgs:
    node-labels: "{labels}"
""".format(
                        labels=",".join(f"{k}={v}" for k, v in cluster_config.node_labels.items())
                        if cluster_config.node_labels
                        else ""
                    )
                ],
                "extraPortMappings": [
                    {
                        "containerPort": 80,
                        "hostPort": cluster_config.http_port,
                        "listenAddress": "0.0.0.0",
                        "protocol": "TCP",
                    },
                    {
                        "containerPort": 443,
                        "hostPort": cluster_config.https_port,
                        "listenAddress": "0.0.0.0",
                        "protocol": "TCP",
                    },
                ],
            }
        ],
    }
    return yaml.dump(config, default_flow_style=False)


def create_kind_cluster(
    name: str,
    cluster_config: ClusterConfig,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> ClusterOutputs:
    """Create a Kind cluster.

    Args:
        name: Pulumi resource name prefix
        cluster_config: Configuration for the cluster
        opts: Pulumi resource options

    Returns:
        ClusterOutputs dataclass with cluster information
    """
    # Generate Kind config
    kind_config_yaml = generate_kind_config(cluster_config)

    # Write config to temp file and create cluster
    # Using heredoc to avoid needing a separate config file
    # Check if cluster exists first to make this idempotent
    create_command = f"""
set -e
if kind get clusters 2>/dev/null | grep -q "^{cluster_config.name}$"; then
    echo "Cluster {cluster_config.name} already exists"
    exit 0
fi
cat > /tmp/kind-{cluster_config.name}.yaml << 'KINDCONFIG'
{kind_config_yaml}
KINDCONFIG
kind create cluster --config /tmp/kind-{cluster_config.name}.yaml --image {KIND_NODE_IMAGE} --wait 120s
rm -f /tmp/kind-{cluster_config.name}.yaml
"""

    delete_command = f"kind delete cluster --name {cluster_config.name} || true"

    # Create the cluster
    cluster = local.Command(
        f"{name}-create",
        create=create_command,
        delete=delete_command,
        opts=pulumi.ResourceOptions(
            custom_timeouts=pulumi.CustomTimeouts(create="10m", delete="5m"),
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )

    # Extract kubeconfig
    kubeconfig_cmd = local.Command(
        f"{name}-kubeconfig",
        create=f"kind get kubeconfig --name {cluster_config.name}",
        opts=pulumi.ResourceOptions(
            depends_on=[cluster],
            parent=opts.parent if opts else None,
        ),
    )

    # Get cluster info for verification
    local.Command(
        f"{name}-info",
        create=f"kubectl cluster-info --context {cluster_config.context_name} 2>&1 | head -3 || echo 'Cluster info pending'",
        opts=pulumi.ResourceOptions(
            depends_on=[cluster],
            parent=opts.parent if opts else None,
        ),
    )

    return ClusterOutputs(
        name=cluster_config.name,
        kubeconfig=kubeconfig_cmd.stdout,
        context_name=cluster_config.context_name,
        cluster_resource=cluster,
    )


def get_existing_cluster_kubeconfig(
    name: str,
    cluster_name: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> pulumi.Output[str]:
    """Get kubeconfig for an existing Kind cluster.

    Args:
        name: Pulumi resource name
        cluster_name: Name of the Kind cluster
        opts: Pulumi resource options

    Returns:
        Pulumi Output containing the kubeconfig
    """
    kubeconfig_cmd = local.Command(
        f"{name}-kubeconfig",
        create=f"kind get kubeconfig --name {cluster_name}",
        opts=opts,
    )
    return kubeconfig_cmd.stdout
