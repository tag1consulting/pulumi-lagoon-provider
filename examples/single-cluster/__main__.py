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

# Import cluster management
from clusters import create_k8s_provider, create_kind_cluster
from config import (
    config,
)

# Import infrastructure components
from infrastructure import (
    create_cluster_issuer,
    create_wildcard_certificate,
    install_cert_manager,
    install_ingress_nginx,
)
from infrastructure.coredns import patch_coredns_for_lagoon, restart_coredns

# Import Lagoon components
from lagoon import (
    configure_keycloak_for_cli_auth,
    ensure_knex_migrations,
    generate_lagoon_secrets,
    install_lagoon_build_deploy_crds,
    install_lagoon_core,
    install_lagoon_remote,
)

# Import registry
from registry import install_harbor

# Import pulumi_lagoon for deploy targets and projects
import pulumi_lagoon as lagoon

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
    )

    provider = create_k8s_provider(
        f"{cluster_config.name}-provider",
        cluster,
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
cluster_issuer = None
lagoon_core_ns = None
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

    # Create ClusterIssuer
    cluster_issuer = create_cluster_issuer(
        "cluster-issuer",
        provider,
        cert_manager,
    )

    # Create namespace for Lagoon core BEFORE certificates
    # (certificates need the namespace to exist)
    lagoon_core_ns = k8s.core.v1.Namespace(
        "lagoon-core-ns",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=namespace_config.lagoon_core,
        ),
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[cluster.cluster_resource] if cluster else None,
        ),
    )

    # Create wildcard certificate
    wildcard_cert = create_wildcard_certificate(
        "wildcard-cert",
        domain_config.base,
        namespace_config.lagoon_core,
        cluster_issuer,
        provider,
        opts=pulumi.ResourceOptions(depends_on=[lagoon_core_ns]),
    )

# =============================================================================
# Phase 3: Configure CoreDNS for Local Domain Resolution
# =============================================================================

coredns_setup = None

if provider is not None and cluster is not None:
    pulumi.log.info("Configuring CoreDNS for local domain resolution...")

    from infrastructure import get_kind_node_internal_ip

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
    coredns_setup = restart_coredns(
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

    # Create Harbor namespace
    harbor_ns = k8s.core.v1.Namespace(
        "harbor-namespace",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=namespace_config.harbor,
        ),
        opts=pulumi.ResourceOptions(
            provider=provider,
            depends_on=[cluster.cluster_resource] if cluster else None,
        ),
    )

    # Create certificate for Harbor
    harbor_cert = create_wildcard_certificate(
        "harbor",
        domain_config.base,
        namespace_config.harbor,
        cluster_issuer,
        provider,
        opts=pulumi.ResourceOptions(depends_on=[harbor_ns]),
    )

    harbor = install_harbor(
        "harbor",
        provider,
        domain_config,
        tls_secret_name="harbor-tls",
        admin_password=config.harbor_admin_password,
        namespace_config=namespace_config,
        opts=pulumi.ResourceOptions(
            depends_on=[ingress.service, harbor_cert, harbor_ns],
        ),
    )

    pulumi.export("harbor_url", harbor.url)
    pulumi.export("harbor_admin_password", pulumi.Output.secret(harbor.admin_password))

# =============================================================================
# Phase 5: Install Lagoon Core
# =============================================================================

lagoon_secrets = None
lagoon_core = None
knex_migrations = None
keycloak_config = None

if provider is not None and config.install_lagoon:
    pulumi.log.info("Generating Lagoon secrets...")

    lagoon_secrets = generate_lagoon_secrets("lagoon-secrets")

    pulumi.export("lagoon_ssh_public_key", lagoon_secrets.ssh_public_key)

    pulumi.log.info("Installing Lagoon core...")

    # Build dependency list - include namespace, cert, CoreDNS, and optionally Harbor
    core_depends = [ingress.service, wildcard_cert, lagoon_core_ns]
    if coredns_setup is not None:
        core_depends.append(coredns_setup)
    if harbor is not None:
        core_depends.append(harbor.release)

    lagoon_core = install_lagoon_core(
        "lagoon-core",
        provider,
        domain_config,
        lagoon_secrets,
        tls_secret_name="wildcard-cert-tls",
        harbor=harbor,
        namespace_config=namespace_config,
        helm_timeout=config.helm_timeout,
        opts=pulumi.ResourceOptions(depends_on=core_depends),
    )

    pulumi.export("lagoon_api_url", lagoon_core.api_url)
    pulumi.export("lagoon_ui_url", lagoon_core.ui_url)
    pulumi.export("lagoon_keycloak_url", lagoon_core.keycloak_url)

    # Configure Keycloak for CLI authentication
    pulumi.log.info("Configuring Keycloak for CLI authentication...")

    # The keycloak service name follows pattern: {release_name}-keycloak
    # For name="lagoon-core", release_name="lagoon-core", so service="lagoon-core-keycloak"
    keycloak_config = configure_keycloak_for_cli_auth(
        "keycloak-config",
        provider,
        namespace=lagoon_core.namespace,
        keycloak_service="lagoon-core-keycloak",
        keycloak_admin_secret="lagoon-core-keycloak",
        opts=pulumi.ResourceOptions(
            depends_on=[lagoon_core.release],
        ),
    )

    # Ensure database migrations are applied
    # Lagoon v2.30.0 has a bug where Knex migrations aren't run by the init container.
    # This check ensures the base schema tables exist before we try to use the API.
    pulumi.log.info("Ensuring Lagoon database migrations are applied...")

    knex_migrations = ensure_knex_migrations(
        "lagoon",
        context=cluster_config.context_name,
        namespace=namespace_config.lagoon_core,
        core_secrets_name="lagoon-core-lagoon-core-secrets",
        opts=pulumi.ResourceOptions(
            depends_on=[lagoon_core.release, keycloak_config],
        ),
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
        context=cluster_config.context_name,
        opts=pulumi.ResourceOptions(
            depends_on=[lagoon_core.release],
        ),
    )

    # Build dependencies for remote - include migrations if available
    remote_depends = [lagoon_core.release, lagoon_crds]
    if knex_migrations is not None:
        remote_depends.append(knex_migrations)

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
            depends_on=remote_depends,
        ),
    )

    pulumi.export("lagoon_remote_namespace", lagoon_remote.namespace)
    pulumi.export("deploy_target_name", config.deploy_target_name)

# =============================================================================
# Phase 7: Create Deploy Target and Example Project (Optional)
# =============================================================================

deploy_target = None
example_project = None

if lagoon_core is not None and config.create_example_project:
    pulumi.log.info("Creating deploy target and example project...")

    # Determine dependencies for deploy target
    deploy_target_deps = [lagoon_core.release, keycloak_config]
    if knex_migrations is not None:
        deploy_target_deps.append(knex_migrations)

    # Create a deploy target for the single cluster
    # This registers the Kind cluster as a deploy target in Lagoon
    deploy_target = lagoon.LagoonDeployTarget(
        "single-cluster-target",
        args=lagoon.LagoonDeployTargetArgs(
            name=config.deploy_target_name,
            console_url="https://kubernetes.default.svc",  # Internal K8s API
            cloud_provider="kind",
            cloud_region="local",
            ssh_host=lagoon_core.ssh_host,
            ssh_port="22",
            # Router pattern determines how routes are generated
            # Format: ${environment}.${project}.${cluster-domain}
            router_pattern=f"${{environment}}.${{project}}.{domain_config.base}",
        ),
        opts=pulumi.ResourceOptions(
            depends_on=deploy_target_deps,
        ),
    )

    pulumi.export("deploy_target_id", deploy_target.id)

    # Create the example Drupal project
    example_project = lagoon.LagoonProject(
        "example-project",
        args=lagoon.LagoonProjectArgs(
            name=config.example_project_name,
            git_url=config.example_project_git_url,
            deploytarget_id=deploy_target.id.apply(lambda x: int(x)),
            production_environment="main",
            # Branch pattern - which branches can be deployed
            branches="^(main|develop|feature/.*)$",
            # PR pattern - which PRs can be deployed
            pullrequests=".*",
        ),
        opts=pulumi.ResourceOptions(
            depends_on=[deploy_target],
        ),
    )

    pulumi.export("example_project_id", example_project.id)
    pulumi.export("example_project_name", config.example_project_name)

# =============================================================================
# Summary Outputs
# =============================================================================

pulumi.export(
    "domain_config",
    {
        "base": domain_config.base,
        "api": domain_config.lagoon_api,
        "ui": domain_config.lagoon_ui,
        "keycloak": domain_config.lagoon_keycloak,
        "harbor": domain_config.harbor,
    },
)

pulumi.export(
    "installation_summary",
    {
        "cluster_created": cluster is not None,
        "harbor_installed": harbor is not None,
        "lagoon_installed": lagoon_core is not None,
        "build_deploy_installed": lagoon_remote is not None,
        "example_project_created": config.create_example_project and example_project is not None,
    },
)
