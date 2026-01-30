"""CoreDNS configuration for local domain resolution.

This module configures CoreDNS in Kind clusters to resolve *.lagoon.local
domains to the Kind node's internal IP, enabling internal pod-to-pod
communication using the same domain names used externally.

Since ingress-nginx uses hostNetwork mode in Kind, traffic needs to be
routed to the node's IP address where the ingress controller is listening.
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s
from pulumi_command import local as command

from config import DomainConfig


def get_kind_container_id(
    name: str,
    cluster_name: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> command.Command:
    """Get the Docker container ID for a Kind cluster node.

    This is used as a trigger for IP lookups - when the container ID changes
    (cluster recreated), the IP lookup will be refreshed.

    Args:
        name: Pulumi resource name prefix
        cluster_name: Kind cluster name (e.g., "lagoon-prod")
        opts: Pulumi resource options

    Returns:
        Command resource with container ID in stdout
    """
    container_name = f"{cluster_name}-control-plane"

    # Get container ID - this changes when cluster is recreated
    return command.Command(
        f"{name}-container-id",
        create=f"docker inspect -f '{{{{.Id}}}}' {container_name} 2>/dev/null || echo 'not-found'",
        # Also run on update to detect changes
        update=f"docker inspect -f '{{{{.Id}}}}' {container_name} 2>/dev/null || echo 'not-found'",
        opts=pulumi.ResourceOptions(
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )


def get_kind_node_internal_ip(
    name: str,
    cluster_name: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> pulumi.Output[str]:
    """Get the internal IP of the Kind control-plane node.

    In Kind clusters with hostNetwork ingress, pods need to connect
    to the node's internal IP to reach the ingress controller.

    Uses docker inspect to get the node IP directly from the container,
    which is more reliable than kubectl as it doesn't require kubeconfig.

    This function automatically refreshes the IP when the cluster container
    changes (e.g., after cluster recreation), using the container ID as a trigger.

    Args:
        name: Pulumi resource name prefix
        cluster_name: Kind cluster name (e.g., "lagoon-prod")
        opts: Pulumi resource options

    Returns:
        Output containing the node's internal IP address
    """
    # First get the container ID - this will be used as a trigger
    # When the container ID changes (cluster recreated), the IP lookup refreshes
    container_id_cmd = get_kind_container_id(
        f"{name}-trigger",
        cluster_name,
        opts,
    )

    # Use docker inspect to get the Kind node's IP address
    # Kind nodes are Docker containers named <cluster-name>-control-plane
    container_name = f"{cluster_name}-control-plane"

    # The IP lookup uses the container ID as a trigger
    # If the container ID changes, this command will be recreated
    get_node_ip = command.Command(
        f"{name}-get-node-ip",
        create=f"docker inspect -f '{{{{.NetworkSettings.Networks.kind.IPAddress}}}}' {container_name}",
        # Also run on update to get fresh IP
        update=f"docker inspect -f '{{{{.NetworkSettings.Networks.kind.IPAddress}}}}' {container_name}",
        # Trigger refresh when container ID changes
        triggers=[container_id_cmd.stdout],
        opts=pulumi.ResourceOptions(
            parent=opts.parent if opts else None,
            depends_on=[container_id_cmd] + (opts.depends_on if opts and opts.depends_on else []),
        ),
    )

    # Return the stdout which contains the IP
    return get_node_ip.stdout.apply(lambda s: s.strip() if s else "127.0.0.1")


def patch_coredns_for_lagoon(
    name: str,
    provider: k8s.Provider,
    domain_config: DomainConfig,
    node_ip: pulumi.Output[str],
    cluster_name: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> command.Command:
    """Patch the CoreDNS ConfigMap to add local domain resolution.

    This modifies the CoreDNS Corefile to include hosts entries
    for resolving *.lagoon.local to the Kind node's internal IP
    where the ingress-nginx controller is running.

    Uses kubectl patch to modify the existing ConfigMap created by kubeadm,
    avoiding Server-Side Apply conflicts.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster (unused, kept for API compat)
        domain_config: Domain configuration with base domain
        node_ip: The internal IP of the Kind node
        cluster_name: Kind cluster name for kubectl context
        opts: Pulumi resource options

    Returns:
        Command resource that patches CoreDNS
    """
    base_domain = domain_config.base

    # Build the hosts block for the Corefile
    def build_corefile(ip: str) -> str:
        domains = [
            domain_config.lagoon_api,
            domain_config.lagoon_ui,
            domain_config.lagoon_keycloak,
            domain_config.lagoon_webhook,
            domain_config.harbor,
            f"git.{base_domain}",
            f"broker.{base_domain}",
            f"ssh.{base_domain}",
            base_domain,
        ]
        hosts_entries = "\n        ".join([f"{ip} {domain}" for domain in domains])

        # Standard Kind CoreDNS Corefile with added hosts block
        # The hosts block must come BEFORE kubernetes plugin for it to work
        # Note: Escaped for shell string
        return f""".:53 {{
    errors
    health {{
       lameduck 5s
    }}
    ready
    hosts {{
        {hosts_entries}
        fallthrough
    }}
    kubernetes cluster.local in-addr.arpa ip6.arpa {{
       pods insecure
       fallthrough in-addr.arpa ip6.arpa
       ttl 30
    }}
    prometheus :9153
    forward . /etc/resolv.conf {{
       max_concurrent 1000
    }}
    cache 30
    loop
    reload
    loadbalance
}}
"""

    # Build kubectl patch command
    def build_patch_command(ip: str) -> str:
        corefile = build_corefile(ip)
        # Escape the corefile for JSON
        corefile_escaped = corefile.replace("\\", "\\\\").replace('"', '\\"').replace("\n", "\\n")
        context = f"kind-{cluster_name}"
        return f'''kubectl --context {context} patch configmap coredns -n kube-system --type merge -p '{{"data":{{"Corefile":"{corefile_escaped}"}}}}'  '''

    patch_cmd = node_ip.apply(build_patch_command)

    # Use kubectl to patch the existing CoreDNS ConfigMap
    # The node_ip trigger ensures this re-runs when the IP changes
    coredns_patch = command.Command(
        f"{name}-coredns-patch",
        create=patch_cmd,
        # Also run on update when IP changes
        update=patch_cmd,
        # Trigger re-patch when the node IP changes
        triggers=[node_ip],
        opts=pulumi.ResourceOptions(
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )

    return coredns_patch


def restart_coredns(
    name: str,
    cluster_name: str,
    coredns_patch: command.Command,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> command.Command:
    """Restart CoreDNS to pick up the new configuration.

    Uses kubectl rollout restart to trigger a new deployment.

    Args:
        name: Pulumi resource name prefix
        cluster_name: Kind cluster name for kubectl context
        coredns_patch: The CoreDNS patch command (for dependency)
        opts: Pulumi resource options

    Returns:
        Command resource that restarts CoreDNS
    """
    context = f"kind-{cluster_name}"

    # Use kubectl to restart CoreDNS
    restart_cmd = command.Command(
        f"{name}-coredns-restart",
        create=f"kubectl --context {context} rollout restart deployment coredns -n kube-system",
        opts=pulumi.ResourceOptions(
            depends_on=[coredns_patch] + (opts.depends_on if opts and opts.depends_on else []),
            parent=opts.parent if opts else None,
        ),
    )

    return restart_cmd


def setup_coredns_for_lagoon(
    name: str,
    provider: k8s.Provider,
    domain_config: DomainConfig,
    cluster_name: str,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> command.Command:
    """Set up CoreDNS to resolve Lagoon domains to the ingress controller.

    This is a convenience function that:
    1. Gets the Kind node's internal IP
    2. Patches the CoreDNS ConfigMap with hosts entries
    3. Triggers a CoreDNS restart

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster (unused, kept for API compat)
        domain_config: Domain configuration with base domain
        cluster_name: Kind cluster name (e.g., "lagoon-prod")
        opts: Pulumi resource options

    Returns:
        The CoreDNS restart Command (final resource in the chain)
    """
    # Get the node IP using docker inspect
    node_ip = get_kind_node_internal_ip(
        f"{name}-node-ip",
        cluster_name,
        opts,
    )

    # Patch CoreDNS using kubectl
    coredns_patch = patch_coredns_for_lagoon(
        name,
        provider,
        domain_config,
        node_ip,
        cluster_name,
        opts,
    )

    # Restart CoreDNS using kubectl
    restart_cmd = restart_coredns(
        name,
        cluster_name,
        coredns_patch,
        opts,
    )

    return restart_cmd
