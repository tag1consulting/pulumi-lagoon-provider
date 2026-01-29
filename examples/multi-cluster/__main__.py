"""Multi-cluster Lagoon example - Main orchestration.

This example demonstrates deploying a complete Lagoon infrastructure:
1. Creating Kind clusters (production and non-production)
2. Installing infrastructure components (ingress-nginx, cert-manager)
3. Installing Harbor container registry
4. Installing Lagoon core components
5. Installing Lagoon remote/build-deploy on each cluster

Prerequisites:
- Docker installed and running (for Kind)
- Pulumi CLI installed
- Kind CLI installed

Usage:
    pulumi up           # Deploy everything
    pulumi destroy      # Tear down everything

Configuration:
    pulumi config set createClusters true/false  # Whether to create Kind clusters
    pulumi config set baseDomain lagoon.local    # Base domain for services
    pulumi config set installHarbor true/false   # Whether to install Harbor
    pulumi config set installLagoon true/false   # Whether to install Lagoon
"""

import pulumi
import pulumi_kubernetes as k8s
from pulumi_command import local as command

from config import (
    config,
    DEFAULT_CLUSTERS,
    DomainConfig,
    NamespaceConfig,
)
from clusters import create_kind_cluster, create_k8s_provider
from infrastructure import (
    install_ingress_nginx,
    install_cert_manager,
    create_cluster_issuer,
    create_wildcard_certificate,
    setup_coredns_for_lagoon,
    get_kind_node_internal_ip,
    patch_coredns_for_lagoon,
)
from registry import install_harbor
from lagoon import (
    generate_lagoon_secrets,
    install_lagoon_core,
    install_lagoon_remote,
    install_lagoon_build_deploy_crds,
    create_rabbitmq_nodeport_service,
    configure_keycloak_for_cli_auth,
    create_deploy_targets,
    create_example_drupal_project,
    ensure_knex_migrations,
)


# =============================================================================
# Configuration
# =============================================================================

domain_config = config.get_domain_config()
namespace_config = config.get_namespace_config()
create_clusters = config.create_clusters
install_harbor_registry = config.install_harbor
install_lagoon_components = config.install_lagoon


# =============================================================================
# Phase 1: Create Kind Clusters
# =============================================================================

prod_cluster = None
nonprod_cluster = None
prod_provider = None
nonprod_provider = None

if create_clusters:
    pulumi.log.info("Creating Kind clusters...")

    # Create production cluster
    prod_cluster_config = config.get_cluster_config("prod")
    prod_cluster = create_kind_cluster(
        "prod-cluster",
        prod_cluster_config,
    )

    # Create non-production cluster
    nonprod_cluster_config = config.get_cluster_config("nonprod")
    nonprod_cluster = create_kind_cluster(
        "nonprod-cluster",
        nonprod_cluster_config,
    )

    # Create Kubernetes providers for each cluster
    prod_provider = create_k8s_provider("prod-k8s", prod_cluster)
    nonprod_provider = create_k8s_provider("nonprod-k8s", nonprod_cluster)

    # Export cluster information
    pulumi.export("prod_cluster_name", prod_cluster.name)
    pulumi.export("prod_cluster_context", prod_cluster.context_name)
    pulumi.export("nonprod_cluster_name", nonprod_cluster.name)
    pulumi.export("nonprod_cluster_context", nonprod_cluster.context_name)
else:
    pulumi.log.info("Skipping Kind cluster creation (createClusters=false)")
    pulumi.log.warn("You must have existing Kind clusters named 'lagoon-prod' and 'lagoon-nonprod'")


# =============================================================================
# Phase 2: Install Infrastructure Components (Production Cluster)
# =============================================================================

if prod_provider is not None:
    pulumi.log.info("Installing infrastructure on production cluster...")

    # Install ingress-nginx
    prod_ingress = install_ingress_nginx(
        "prod-ingress",
        prod_provider,
        namespace_config,
        opts=pulumi.ResourceOptions(depends_on=[prod_cluster.cluster_resource]),
    )

    # Configure CoreDNS to resolve *.lagoon.local to the ingress controller
    # This allows pods inside the cluster to reach services via their external domain names
    pulumi.log.info("Configuring CoreDNS for local domain resolution...")
    prod_coredns = setup_coredns_for_lagoon(
        "prod-coredns",
        prod_provider,
        domain_config,
        prod_cluster.name,  # Kind cluster name (e.g., "lagoon-prod")
        opts=pulumi.ResourceOptions(
            depends_on=[prod_cluster.cluster_resource, prod_ingress.service],
        ),
    )

    # Install cert-manager
    prod_cert_manager = install_cert_manager(
        "prod-certmanager",
        prod_provider,
        namespace_config,
        opts=pulumi.ResourceOptions(depends_on=[prod_cluster.cluster_resource]),
    )

    # Create ClusterIssuer
    prod_issuer = create_cluster_issuer(
        "prod-issuer",
        prod_provider,
        prod_cert_manager,
    )

    # Create namespaces for Lagoon and Harbor BEFORE certificates
    # (certificates need the namespace to exist)
    lagoon_core_ns = k8s.core.v1.Namespace(
        "prod-lagoon-core-ns",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=namespace_config.lagoon_core,
        ),
        opts=pulumi.ResourceOptions(
            provider=prod_provider,
            depends_on=[prod_cluster.cluster_resource],
        ),
    )

    harbor_ns = k8s.core.v1.Namespace(
        "prod-harbor-ns",
        metadata=k8s.meta.v1.ObjectMetaArgs(
            name=namespace_config.harbor,
        ),
        opts=pulumi.ResourceOptions(
            provider=prod_provider,
            depends_on=[prod_cluster.cluster_resource],
        ),
    )

    # Create wildcard certificate for Lagoon domain
    prod_cert = create_wildcard_certificate(
        "prod-lagoon",
        domain_config.base,
        namespace_config.lagoon_core,
        prod_issuer,
        prod_provider,
        opts=pulumi.ResourceOptions(depends_on=[lagoon_core_ns]),
    )

    pulumi.export("prod_ingress_class", prod_ingress.class_name)


# =============================================================================
# Phase 3: Install Harbor Registry (Production Cluster)
# =============================================================================

prod_harbor = None
if prod_provider is not None and install_harbor_registry:
    pulumi.log.info("Installing Harbor registry on production cluster...")

    # Create certificate for Harbor (namespace already created above)
    harbor_cert = create_wildcard_certificate(
        "prod-harbor",
        domain_config.base,
        namespace_config.harbor,
        prod_issuer,
        prod_provider,
        opts=pulumi.ResourceOptions(depends_on=[harbor_ns]),
    )

    prod_harbor = install_harbor(
        "prod-harbor",
        prod_provider,
        domain_config,
        tls_secret_name="prod-harbor-tls",
        namespace_config=namespace_config,
        opts=pulumi.ResourceOptions(
            depends_on=[prod_ingress.service, harbor_cert, harbor_ns],
        ),
    )

    pulumi.export("harbor_url", prod_harbor.url)
    pulumi.export("harbor_admin_password", pulumi.Output.secret(prod_harbor.admin_password))


# =============================================================================
# Phase 4: Generate Lagoon Secrets
# =============================================================================

lagoon_secrets = None
if install_lagoon_components:
    pulumi.log.info("Generating Lagoon secrets...")
    lagoon_secrets = generate_lagoon_secrets("lagoon")

    # Export SSH public key (useful for configuring Git webhooks)
    pulumi.export("lagoon_ssh_public_key", lagoon_secrets.ssh_public_key)


# =============================================================================
# Phase 5: Install Lagoon Core (Production Cluster)
# =============================================================================

lagoon_core = None
if prod_provider is not None and install_lagoon_components and lagoon_secrets is not None:
    pulumi.log.info("Installing Lagoon core on production cluster...")

    # Build dependency list - include Harbor and CoreDNS if available
    lagoon_core_deps = [prod_ingress.service, prod_cert, lagoon_core_ns, prod_coredns]
    if prod_harbor is not None:
        lagoon_core_deps.append(prod_harbor.release)

    lagoon_core = install_lagoon_core(
        "prod-lagoon-core",
        prod_provider,
        domain_config,
        lagoon_secrets,
        tls_secret_name="prod-lagoon-tls",
        harbor=prod_harbor,
        namespace_config=namespace_config,
        helm_timeout=config.helm_timeout,
        opts=pulumi.ResourceOptions(
            depends_on=lagoon_core_deps,
        ),
    )

    pulumi.export("lagoon_api_url", lagoon_core.api_url)
    pulumi.export("lagoon_ui_url", lagoon_core.ui_url)
    pulumi.export("lagoon_keycloak_url", lagoon_core.keycloak_url)

    # Create a NodePort service for cross-cluster RabbitMQ access
    # The Helm chart doesn't support fixed NodePorts, so we create our own
    rabbitmq_nodeport_svc = create_rabbitmq_nodeport_service(
        "prod-lagoon",
        "prod-core",  # Helm release name
        lagoon_core.namespace,
        prod_provider,
        nodeport=30672,  # Fixed NodePort for cross-cluster communication
        opts=pulumi.ResourceOptions(
            depends_on=[lagoon_core.release],
        ),
    )

    # Configure Keycloak for CLI/programmatic authentication
    # This enables Direct Access Grants (OAuth password grant) for the lagoon-ui
    # client and creates the lagoonadmin user if it doesn't exist.
    # Without this, CLI tools and scripts cannot authenticate to the Lagoon API.
    pulumi.log.info("Configuring Keycloak for CLI authentication...")
    keycloak_config_job = configure_keycloak_for_cli_auth(
        "prod-lagoon",
        prod_provider,
        namespace=lagoon_core.namespace,
        keycloak_service="prod-core-lagoon-core-keycloak",
        keycloak_admin_secret="prod-core-lagoon-core-keycloak",
        opts=pulumi.ResourceOptions(
            depends_on=[lagoon_core.release],
        ),
    )


# =============================================================================
# Phase 6: Install Lagoon Remote (Both Clusters)
# =============================================================================

if lagoon_core is not None and lagoon_secrets is not None:
    # Install Lagoon remote on production cluster
    if prod_provider is not None:
        pulumi.log.info("Installing Lagoon remote on production cluster...")

        # Install Lagoon build-deploy CRDs first (required before Helm release)
        prod_lagoon_crds = install_lagoon_build_deploy_crds(
            "prod-lagoon-crds",
            prod_provider,
            context="kind-lagoon-prod",
            opts=pulumi.ResourceOptions(
                depends_on=[lagoon_core.release],
            ),
        )

        prod_remote = install_lagoon_remote(
            "prod-lagoon-remote",
            prod_provider,
            lagoon_core,
            lagoon_secrets,
            target_name="lagoon-prod",
            is_production=True,
            namespace_config=namespace_config,
            opts=pulumi.ResourceOptions(
                depends_on=[lagoon_core.release, prod_lagoon_crds],
            ),
        )

        pulumi.export("prod_remote_namespace", prod_remote.namespace)

    # Install Lagoon remote on non-production cluster
    if nonprod_provider is not None:
        pulumi.log.info("Installing Lagoon remote on non-production cluster...")

        # Install ingress-nginx on nonprod cluster
        nonprod_ingress = install_ingress_nginx(
            "nonprod-ingress",
            nonprod_provider,
            namespace_config,
            opts=pulumi.ResourceOptions(depends_on=[nonprod_cluster.cluster_resource]),
        )

        # Get the prod cluster's node IP for cross-cluster communication
        # Both Kind clusters share the Docker network, so this IP is reachable
        prod_node_ip = get_kind_node_internal_ip(
            "prod-node-ip",
            prod_cluster.name,  # Kind cluster name (e.g., "lagoon-prod")
            opts=pulumi.ResourceOptions(depends_on=[prod_cluster.cluster_resource]),
        )

        # Configure CoreDNS on nonprod to resolve *.lagoon.local to prod cluster
        # This enables cross-cluster communication via the same domain names
        pulumi.log.info("Configuring CoreDNS on nonprod for cross-cluster resolution...")
        nonprod_coredns_patch = patch_coredns_for_lagoon(
            "nonprod-coredns",
            nonprod_provider,
            domain_config,
            prod_node_ip,  # Use prod node IP so nonprod can reach prod services
            nonprod_cluster.name,  # Use nonprod cluster name for kubectl context
            opts=pulumi.ResourceOptions(
                depends_on=[nonprod_cluster.cluster_resource, nonprod_ingress.service],
            ),
        )

        # Import restart_coredns from infrastructure module
        from infrastructure.coredns import restart_coredns as restart_coredns_fn

        # Restart CoreDNS on nonprod to pick up the new configuration
        nonprod_coredns = restart_coredns_fn(
            "nonprod-coredns",
            nonprod_cluster.name,
            nonprod_coredns_patch,
        )

        # Build external RabbitMQ host for cross-cluster connection
        # Format: <prod_node_ip>:<nodeport>
        external_rabbitmq_host = prod_node_ip.apply(
            lambda ip: f"{ip}:{lagoon_core.rabbitmq_nodeport}"
        )

        # Install Lagoon build-deploy CRDs on nonprod cluster
        nonprod_lagoon_crds = install_lagoon_build_deploy_crds(
            "nonprod-lagoon-crds",
            nonprod_provider,
            context="kind-lagoon-nonprod",
            opts=pulumi.ResourceOptions(
                depends_on=[nonprod_cluster.cluster_resource],
            ),
        )

        nonprod_remote = install_lagoon_remote(
            "nonprod-lagoon-remote",
            nonprod_provider,
            lagoon_core,
            lagoon_secrets,
            target_name="lagoon-nonprod",
            is_production=False,
            namespace_config=namespace_config,
            external_rabbitmq_host=external_rabbitmq_host,
            opts=pulumi.ResourceOptions(
                depends_on=[lagoon_core.release, nonprod_ingress.service, nonprod_coredns, nonprod_lagoon_crds],
            ),
        )

        pulumi.export("nonprod_remote_namespace", nonprod_remote.namespace)
        pulumi.export("prod_node_ip", prod_node_ip)


# =============================================================================
# Phase 7: Ensure Database Migrations
# =============================================================================

# Lagoon v2.30.0 has a bug where Knex migrations aren't run by the init container.
# This check ensures the base schema tables exist before we try to use the API
# for deploy target management.
knex_migrations = None
if lagoon_core is not None:
    pulumi.log.info("Ensuring Lagoon database migrations are applied...")

    knex_migrations = ensure_knex_migrations(
        "prod-lagoon",
        context="kind-lagoon-prod",
        namespace=namespace_config.lagoon_core,
        core_secrets_name="prod-core-lagoon-core-secrets",  # Not used by script but required by function
        opts=pulumi.ResourceOptions(
            depends_on=[lagoon_core.release, keycloak_config_job],
        ),
    )


# =============================================================================
# Phase 8: Create Example Drupal Project (Optional)
# =============================================================================

example_project = None
if lagoon_core is not None and config.create_example_project:
    pulumi.log.info("Creating example Drupal project with multi-cluster routing...")

    # Create deploy targets and example project via a Command resource
    # This approach uses kubectl port-forward and curl to create resources via the
    # Lagoon GraphQL API, avoiding the need for the provider to have a pre-configured
    # token on first run.
    #
    # The Command:
    # 1. Starts port-forwards to the API
    # 2. Generates an admin JWT token using the JWTSECRET
    # 3. Creates deploy targets for prod and nonprod clusters
    # 4. Creates the example Drupal project with deploy target configurations
    example_setup_deps = [lagoon_core.release, keycloak_config_job]
    if knex_migrations is not None:
        example_setup_deps.append(knex_migrations)

    example_setup = command.Command(
        "create-deploy-targets-and-project",
        create=f'''
#!/bin/bash
set -e

CONTEXT="kind-lagoon-prod"
NAMESPACE="lagoon-core"
API_SVC="prod-core-lagoon-core-api"
CORE_SECRETS="prod-core-lagoon-core-secrets"
SSH_HOST="{lagoon_core.ssh_host}"
ROUTER_PATTERN="${{environment}}.${{project}}.{domain_config.base}"
PROJECT_NAME="{config.example_project_name}"
GIT_URL="{config.example_project_git_url}"

echo "Setting up Lagoon deploy targets and example project..."

# Start port-forward to API (run in background, will be cleaned up when script exits)
kubectl --context "$CONTEXT" port-forward -n "$NAMESPACE" "svc/$API_SVC" 7070:80 &
PF_PID=$!
trap "kill $PF_PID 2>/dev/null || true" EXIT

# Wait for port-forward to be ready
for i in $(seq 1 15); do
    if curl -sf "http://localhost:7070/" >/dev/null 2>&1; then
        echo "Port-forward ready"
        break
    fi
    sleep 1
done

# Get JWTSECRET and generate admin token
JWT_SECRET=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret "$CORE_SECRETS" -o jsonpath='{{.data.JWTSECRET}}' | base64 -d)
TOKEN=$(python3 -c "
import jwt, time
payload = {{'role': 'admin', 'iss': 'lagoon-api', 'sub': 'lagoonadmin', 'aud': 'api.dev', 'iat': int(time.time()), 'exp': int(time.time()) + 3600}}
print(jwt.encode(payload, '$JWT_SECRET', algorithm='HS256'))
")

API_URL="http://localhost:7070/graphql"

# Function to call GraphQL API
graphql() {{
    curl -sf "$API_URL" \\
        -H "Authorization: Bearer $TOKEN" \\
        -H "Content-Type: application/json" \\
        -d "$1"
}}

# Create nonprod deploy target
echo "Creating nonprod deploy target..."
NONPROD_RESULT=$(graphql '{{"query": "mutation {{ addKubernetes(input: {{ name: \\"lagoon-nonprod\\", consoleUrl: \\"https://kubernetes.default.svc\\", cloudProvider: \\"kind\\", cloudRegion: \\"local\\", sshHost: \\"'"$SSH_HOST"'\\", sshPort: \\"22\\", routerPattern: \\"'"$ROUTER_PATTERN"'\\" }}) {{ id name }} }}"}}')
NONPROD_ID=$(echo "$NONPROD_RESULT" | jq -r '.data.addKubernetes.id // empty')
if [ -z "$NONPROD_ID" ]; then
    # Check if it already exists
    NONPROD_ID=$(graphql '{{"query": "{{ kubernetes(name: \\"lagoon-nonprod\\") {{ id }} }}"}}' | jq -r '.data.kubernetes.id // empty')
fi
echo "Nonprod deploy target ID: $NONPROD_ID"

# Create prod deploy target
echo "Creating prod deploy target..."
PROD_RESULT=$(graphql '{{"query": "mutation {{ addKubernetes(input: {{ name: \\"lagoon-prod\\", consoleUrl: \\"https://kubernetes.default.svc\\", cloudProvider: \\"kind\\", cloudRegion: \\"local\\", sshHost: \\"'"$SSH_HOST"'\\", sshPort: \\"22\\", routerPattern: \\"'"$ROUTER_PATTERN"'\\" }}) {{ id name }} }}"}}')
PROD_ID=$(echo "$PROD_RESULT" | jq -r '.data.addKubernetes.id // empty')
if [ -z "$PROD_ID" ]; then
    # Check if it already exists
    PROD_ID=$(graphql '{{"query": "{{ kubernetes(name: \\"lagoon-prod\\") {{ id }} }}"}}' | jq -r '.data.kubernetes.id // empty')
fi
echo "Prod deploy target ID: $PROD_ID"

# Create example project
echo "Creating example project: $PROJECT_NAME..."
PROJECT_RESULT=$(graphql '{{"query": "mutation {{ addProject(input: {{ name: \\"'"$PROJECT_NAME"'\\", gitUrl: \\"'"$GIT_URL"'\\", kubernetes: '"$PROD_ID"', branches: \\"^(main|develop|feature/.*)$\\", pullrequests: \\".*\\", productionEnvironment: \\"main\\" }}) {{ id name }} }}"}}')
PROJECT_ID=$(echo "$PROJECT_RESULT" | jq -r '.data.addProject.id // empty')
if [ -z "$PROJECT_ID" ]; then
    # Check if it already exists
    PROJECT_ID=$(graphql '{{"query": "{{ projectByName(name: \\"'"$PROJECT_NAME"'\\") {{ id }} }}"}}' | jq -r '.data.projectByName.id // empty')
fi
echo "Project ID: $PROJECT_ID"

# Add deploy target config for prod (main branch only)
echo "Configuring production routing..."
graphql '{{"query": "mutation {{ addDeployTargetConfig(input: {{ project: '"$PROJECT_ID"', deployTarget: '"$PROD_ID"', branches: \\"main\\", pullrequests: \\"false\\", weight: 10 }}) {{ id }} }}"}}'

# Add deploy target config for nonprod (all other branches + PRs)
echo "Configuring non-production routing..."
graphql '{{"query": "mutation {{ addDeployTargetConfig(input: {{ project: '"$PROJECT_ID"', deployTarget: '"$NONPROD_ID"', branches: \\".*\\", pullrequests: \\"true\\", weight: 1 }}) {{ id }} }}"}}'

echo ""
echo "Setup complete!"
echo "  Nonprod deploy target: $NONPROD_ID"
echo "  Prod deploy target: $PROD_ID"
echo "  Project: $PROJECT_NAME (ID: $PROJECT_ID)"

# Output JSON for Pulumi to parse
echo ""
echo "PULUMI_OUTPUT_JSON:{{\\"nonprod_id\\": \\"$NONPROD_ID\\", \\"prod_id\\": \\"$PROD_ID\\", \\"project_id\\": \\"$PROJECT_ID\\"}}"
''',
        opts=pulumi.ResourceOptions(
            depends_on=example_setup_deps,
        ),
    )

    # Export the IDs (they'll be available after the command runs)
    pulumi.export("prod_deploy_target_id", example_setup.stdout.apply(
        lambda s: next((l.split("PULUMI_OUTPUT_JSON:")[1] for l in s.split("\n") if "PULUMI_OUTPUT_JSON:" in l), "{}")).apply(
        lambda j: __import__("json").loads(j).get("prod_id", "unknown")
    ))
    pulumi.export("nonprod_deploy_target_id", example_setup.stdout.apply(
        lambda s: next((l.split("PULUMI_OUTPUT_JSON:")[1] for l in s.split("\n") if "PULUMI_OUTPUT_JSON:" in l), "{}")).apply(
        lambda j: __import__("json").loads(j).get("nonprod_id", "unknown")
    ))
    pulumi.export("example_project_id", example_setup.stdout.apply(
        lambda s: next((l.split("PULUMI_OUTPUT_JSON:")[1] for l in s.split("\n") if "PULUMI_OUTPUT_JSON:" in l), "{}")).apply(
        lambda j: __import__("json").loads(j).get("project_id", "unknown")
    ))
    pulumi.export("example_project_name", config.example_project_name)


# =============================================================================
# Summary Exports
# =============================================================================

pulumi.export("domain_config", {
    "base": domain_config.base,
    "api": domain_config.lagoon_api,
    "ui": domain_config.lagoon_ui,
    "keycloak": domain_config.lagoon_keycloak,
    "harbor": domain_config.harbor,
})

pulumi.export("installation_summary", {
    "clusters_created": create_clusters,
    "harbor_installed": install_harbor_registry and prod_harbor is not None,
    "lagoon_installed": install_lagoon_components and lagoon_core is not None,
    "example_project_created": config.create_example_project and example_project is not None,
})
