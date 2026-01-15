#!/usr/bin/env python3
"""
Example: Testing pulumi-lagoon-provider against local test cluster

This example demonstrates how to use the pulumi-lagoon-provider with a local
Lagoon instance running in kind.

Prerequisites:
1. Test cluster is deployed: cd ../test-cluster && pulumi up
2. Credentials obtained: ./scripts/get-credentials.sh
3. /etc/hosts configured: 127.0.0.1 api.lagoon.test
4. Provider installed: cd .. && pip install -e .

Usage:
    # Set environment variables (or use pulumi config)
    export LAGOON_API_URL='http://api.lagoon.test/graphql'
    export LAGOON_TOKEN='<token-from-get-credentials.sh>'

    # Initialize Pulumi stack
    pulumi stack init test

    # Preview/deploy
    pulumi preview
    pulumi up

    # Clean up
    pulumi destroy
"""

import os
import sys
import pulumi

# Add parent directory to path to import pulumi_lagoon
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", ".."))

import pulumi_lagoon as lagoon

# Configuration - will use LAGOON_API_URL and LAGOON_TOKEN env vars
# or can be set via:
#   pulumi config set lagoon:apiUrl http://api.lagoon.test/graphql
#   pulumi config set lagoon:token <token> --secret

config = pulumi.Config()

# Get deploytarget ID - you may need to query this from Lagoon first
# For local test cluster, typically the first/only deploy target will be ID 1
deploytarget_id = config.get_int("deploytargetId") or 1

# Create a test project
test_project = lagoon.LagoonProject(
    "test-drupal-site",
    lagoon.LagoonProjectArgs(
        name="test-drupal-site",
        git_url="https://github.com/amazeeio/drupal-example.git",
        deploytarget_id=deploytarget_id,
        production_environment="main",
        branches="^(main|develop)$",
        pullrequests="true",
        openshift_project_pattern="${project}-${environment}",
        development_environments_limit=2,
        storage_calc=1,
        environments_limit=3,
    ),
)

# Create a production environment
prod_environment = lagoon.LagoonEnvironment(
    "test-prod-env",
    lagoon.LagoonEnvironmentArgs(
        name="main",
        project_id=test_project.id,
        deploy_type="branch",
        environment_type="production",
    ),
)

# Create a development environment
dev_environment = lagoon.LagoonEnvironment(
    "test-dev-env",
    lagoon.LagoonEnvironmentArgs(
        name="develop",
        project_id=test_project.id,
        deploy_type="branch",
        environment_type="development",
    ),
)

# Create project-level variable (available to all environments)
project_var = lagoon.LagoonVariable(
    "project-database-type",
    lagoon.LagoonVariableArgs(
        name="DATABASE_TYPE",
        value="mariadb",
        project_id=test_project.id,
        scope="build",
    ),
)

# Create production environment variable
prod_var_db_host = lagoon.LagoonVariable(
    "prod-db-host",
    lagoon.LagoonVariableArgs(
        name="DATABASE_HOST",
        value="mariadb.prod.example.com",
        project_id=test_project.id,
        environment_id=prod_environment.id,
        scope="runtime",
    ),
)

# Create production API key (secret)
prod_var_api_key = lagoon.LagoonVariable(
    "prod-api-key",
    lagoon.LagoonVariableArgs(
        name="API_KEY",
        value="prod-secret-key-12345",
        project_id=test_project.id,
        environment_id=prod_environment.id,
        scope="runtime",
    ),
)

# Create development environment variable
dev_var_db_host = lagoon.LagoonVariable(
    "dev-db-host",
    lagoon.LagoonVariableArgs(
        name="DATABASE_HOST",
        value="mariadb.dev.example.com",
        project_id=test_project.id,
        environment_id=dev_environment.id,
        scope="runtime",
    ),
)

# Export useful information
pulumi.export("project_id", test_project.id)
pulumi.export("project_name", test_project.name)
pulumi.export("prod_environment_id", prod_environment.id)
pulumi.export("dev_environment_id", dev_environment.id)
pulumi.export(
    "lagoon_ui_url",
    test_project.id.apply(
        lambda pid: f"http://ui.lagoon.test/projects/{pid}"
    ),
)

# Success message
pulumi.log.info("✓ Test resources created successfully!")
pulumi.log.info("View in Lagoon UI: http://ui.lagoon.test")
