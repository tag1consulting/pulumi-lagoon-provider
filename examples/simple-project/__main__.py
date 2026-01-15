"""Simple Lagoon project example.

This example demonstrates how to use the Pulumi Lagoon provider to:
1. Create a Lagoon project
2. Create production and development environments
3. Add environment variables

Prerequisites:
- Lagoon API access (set LAGOON_API_URL and LAGOON_TOKEN)
- A deploy target (Kubernetes cluster) ID
"""

import pulumi
import pulumi_lagoon as lagoon

# Get configuration
config = pulumi.Config()
deploytarget_id = config.require_int("deploytargetId")

# Optional: customize project name
project_name = config.get("projectName") or "example-drupal-site"

# Create a Lagoon project
project = lagoon.LagoonProject("example-project",
    lagoon.LagoonProjectArgs(
        name=project_name,
        git_url="git@github.com:example/drupal-site.git",
        deploytarget_id=deploytarget_id,
        production_environment="main",
        branches="^(main|develop|stage)$",
        pullrequests="^(PR-.*)",
    )
)

# Create production environment
prod_env = lagoon.LagoonEnvironment("production",
    lagoon.LagoonEnvironmentArgs(
        name="main",
        project_id=project.id,
        deploy_type="branch",
        environment_type="production",
    )
)

# Create development environment
# Note: auto_idle is not supported in AddEnvironmentInput - must be set via
# updateEnvironment mutation after creation (not yet implemented in provider)
dev_env = lagoon.LagoonEnvironment("development",
    lagoon.LagoonEnvironmentArgs(
        name="develop",
        project_id=project.id,
        deploy_type="branch",
        environment_type="development",
    )
)

# Add a project-level variable (applies to all environments)
project_var = lagoon.LagoonVariable("api-url",
    lagoon.LagoonVariableArgs(
        name="API_BASE_URL",
        value="https://api.example.com",
        project_id=project.id,
        scope="runtime",
    )
)

# Add environment-specific variable for production
prod_db_host = lagoon.LagoonVariable("prod-db-host",
    lagoon.LagoonVariableArgs(
        name="DATABASE_HOST",
        value="mysql-prod.example.com",
        project_id=project.id,
        environment_id=prod_env.id,
        scope="runtime",
    )
)

# Add environment-specific variable for development
dev_db_host = lagoon.LagoonVariable("dev-db-host",
    lagoon.LagoonVariableArgs(
        name="DATABASE_HOST",
        value="mysql-dev.example.com",
        project_id=project.id,
        environment_id=dev_env.id,
        scope="runtime",
    )
)

# Export useful outputs
pulumi.export("project_id", project.id)
pulumi.export("project_name", project.name)
pulumi.export("production_url", prod_env.route)
pulumi.export("development_url", dev_env.route)
pulumi.export("production_environment_id", prod_env.id)
pulumi.export("development_environment_id", dev_env.id)
