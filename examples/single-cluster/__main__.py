"""Single-cluster Lagoon deployment example.

This example deploys a complete Lagoon stack to a single Kind cluster:
- Kind cluster with ingress support
- ingress-nginx for routing
- cert-manager with self-signed certificates
- Harbor container registry
- Lagoon core (API, UI, Keycloak, RabbitMQ, etc.)
- Lagoon remote with build-deploy controller

This is a simplified version of the multi-cluster example, suitable for:
- Local development
- Testing Lagoon features
- Learning Lagoon architecture

For production-like deployments with separate prod/nonprod clusters,
see the multi-cluster example.
"""

import pulumi
import pulumi_kubernetes as k8s

from config import (
    config,
    VERSIONS,
    KIND_NODE_IMAGE,
    DomainConfig,
    NamespaceConfig,
)

# Import cluster management
from clusters import create_kind_cluster, create_k8s_provider

# Import infrastructure components
from infrastructure import (
    install_ingress_nginx,
    install_cert_manager,
    create_wildcard_certificate,
)
from infrastructure.coredns import patch_coredns_for_lagoon, restart_coredns

# Import registry
from registry import install_harbor

# Import Lagoon components
from lagoon import (
    generate_lagoon_secrets,
    install_lagoon_core,
    install_lagoon_remote,
    install_lagoon_build_deploy_crds,
    configure_keycloak_for_cli_auth,
)

# =============================================================================
# Configuration
# =============================================================================

domain_config = config.get_domain_config()
namespace_config = config.get_namespace_config()
cluster_config = config.get_cluster_config()

pulumi.log.info(f"Deploying Lagoon to single cluster: {cluster_config.name}")
pulumi.log.info(f"Base domain: {domain_config.base}")
pulumi.log.info(f"Ports: HTTP={cluster_config.http_port}, HTTPS={cluster_config.https_port}")

# =============================================================================
# Phase 1: Create Kind Cluster
# =============================================================================

cluster = None
provider = None

if config.create_cluster:
    pulumi.log.info("Creating Kind cluster...")

    cluster = create_kind_cluster(
        cluster_config.name,
        cluster_config,
        node_image=KIND_NODE_IMAGE,
    )

    provider = create_k8s_provider(
        f"{cluster_config.name}-provider",
        cluster.kubeconfig,
        cluster.context_name,
    )

    pulumi.export("cluster_name", cluster.name)
    pulumi.export("cluster_context", cluster.context_name)
else:
    pulumi.log.info("Skipping cluster creation (createCluster=false)")
    # Use default provider for existing cluster
    provider = k8s.Provider(
        "existing-cluster-provider",
        context=cluster_config.context_name,
    )

# =============================================================================
# Phase 2: Install Infrastructure Components
# =============================================================================

ingress = None
cert_manager = None
wildcard_cert = None

if provider is not None:
    pulumi.log.info("Installing infrastructure components...")

    # Install ingress-nginx
    ingress = install_ingress_nginx(
        "ingress",
        provider,
        namespace_config,
        opts=pulumi.ResourceOptions(
            depends_on=[cluster.cluster_resource] if cluster else None,
        ),
    )

    pulumi.export("ingress_class", ingress.class_name)

    # Install cert-manager
    cert_manager = install_cert_manager(
        "cert-manager",
        provider,
        namespace_config,
        opts=pulumi.ResourceOptions(
            depends_on=[ingress.service],
        ),
    )

    # Create wildcard certificate
    wildcard_cert = create_wildcard_certificate(
        "wildcard-cert",
        provider,
        domain_config,
        cert_manager,
        namespace_config,
    )

# =============================================================================
# Phase 3: Configure CoreDNS for Local Domain Resolution
# =============================================================================

if provider is not None and cluster is not None:
    pulumi.log.info("Configuring CoreDNS for local domain resolution...")

    from clusters import get_kind_node_internal_ip

    # Get the cluster's node IP
    node_ip = get_kind_node_internal_ip(
        "node-ip",
        cluster.name,
        opts=pulumi.ResourceOptions(depends_on=[cluster.cluster_resource]),
    )

    # Patch CoreDNS to resolve *.lagoon.local to the node IP
    coredns_patch = patch_coredns_for_lagoon(
        "coredns",
        provider,
        domain_config,
        node_ip,
        cluster.name,
        opts=pulumi.ResourceOptions(
            depends_on=[cluster.cluster_resource, ingress.service],
        ),
    )

    # Restart CoreDNS to pick up the new configuration
    coredns_restart = restart_coredns(
        "coredns",
        cluster.name,
        coredns_patch,
    )

    pulumi.export("node_ip", node_ip)

# =============================================================================
# Phase 4: Install Harbor Registry
# =============================================================================

harbor = None

if provider is not None and config.install_harbor:
    pulumi.log.info("Installing Harbor registry...")

    harbor = install_harbor(
        "harbor",
        provider,
        domain_config,
        namespace_config,
        admin_password=config.harbor_admin_password,
        opts=pulumi.ResourceOptions(
            depends_on=[wildcard_cert] if wildcard_cert else [ingress.service],
        ),
    )

    pulumi.export("harbor_url", harbor.url)
    pulumi.export("harbor_admin_password", harbor.admin_password)

# =============================================================================
# Phase 5: Install Lagoon Core
# =============================================================================

lagoon_secrets = None
lagoon_core = None

if provider is not None and config.install_lagoon:
    pulumi.log.info("Generating Lagoon secrets...")

    lagoon_secrets = generate_lagoon_secrets("lagoon-secrets")

    pulumi.export("lagoon_ssh_public_key", lagoon_secrets.ssh_public_key)

    pulumi.log.info("Installing Lagoon core...")

    core_depends = [ingress.service]
    if harbor:
        core_depends.append(harbor.release)
    if wildcard_cert:
        core_depends.append(wildcard_cert)

    lagoon_core = install_lagoon_core(
        "lagoon-core",
        provider,
        domain_config,
        lagoon_secrets,
        namespace_config=namespace_config,
        harbor_outputs=harbor,
        helm_timeout=config.helm_timeout,
        opts=pulumi.ResourceOptions(depends_on=core_depends),
    )

    pulumi.export("lagoon_api_url", lagoon_core.api_url)
    pulumi.export("lagoon_ui_url", lagoon_core.ui_url)
    pulumi.export("lagoon_keycloak_url", lagoon_core.keycloak_url)

    # Configure Keycloak for CLI authentication
    pulumi.log.info("Configuring Keycloak for CLI authentication...")

    keycloak_config = configure_keycloak_for_cli_auth(
        "keycloak-config",
        provider,
        lagoon_core,
        namespace_config,
    )

# =============================================================================
# Phase 6: Install Lagoon Remote (Build-Deploy Controller)
# =============================================================================

lagoon_remote = None

if lagoon_core is not None and lagoon_secrets is not None:
    pulumi.log.info("Installing Lagoon remote (build-deploy controller)...")

    # Install CRDs first
    lagoon_crds = install_lagoon_build_deploy_crds(
        "lagoon-crds",
        provider,
        opts=pulumi.ResourceOptions(
            depends_on=[lagoon_core.release],
        ),
    )

    # Install lagoon-remote with build-deploy enabled
    lagoon_remote = install_lagoon_remote(
        "lagoon-remote",
        provider,
        lagoon_core,
        lagoon_secrets,
        target_name=config.deploy_target_name,
        is_production=False,
        namespace_config=namespace_config,
        opts=pulumi.ResourceOptions(
            depends_on=[lagoon_core.release, lagoon_crds],
        ),
    )

    pulumi.export("lagoon_remote_namespace", lagoon_remote.namespace)
    pulumi.export("deploy_target_name", config.deploy_target_name)

# =============================================================================
# Summary Outputs
# =============================================================================

pulumi.export("domain_config", {
    "base": domain_config.base,
    "api": domain_config.lagoon_api,
    "ui": domain_config.lagoon_ui,
    "keycloak": domain_config.lagoon_keycloak,
    "harbor": domain_config.harbor,
})

pulumi.export("installation_summary", {
    "cluster_created": cluster is not None,
    "harbor_installed": harbor is not None,
    "lagoon_installed": lagoon_core is not None,
    "build_deploy_installed": lagoon_remote is not None,
})
