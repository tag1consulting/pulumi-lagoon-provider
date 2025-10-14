"""Simple Lagoon project example."""

import pulumi

# TODO: Uncomment when LagoonProject is implemented
# import pulumi_lagoon as lagoon

# config = pulumi.Config()
# deploytarget_id = config.require_int("deploytargetId")

# # Create a Lagoon project
# project = lagoon.LagoonProject("example-project",
#     name="example-drupal-site",
#     git_url="git@github.com:example/drupal-site.git",
#     deploytarget_id=deploytarget_id,
#     production_environment="main",
#     branches="^(main|develop|stage)$",
#     pullrequests="^(PR-.*)",
# )

# # Export project ID
# pulumi.export("project_id", project.id)
# pulumi.export("project_name", project.name)

# Placeholder for now
pulumi.export("status", "Lagoon provider resources not yet implemented - coming soon!")
