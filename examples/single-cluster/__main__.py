"""Single-cluster Lagoon deployment.

IMPORTANT: This is a placeholder. Use test-cluster/ or examples/multi-cluster/ instead.

After the multi-cluster branch merges, this will be refactored to:
1. Import infrastructure modules from multi-cluster (clusters/, infrastructure/, lagoon/, etc.)
2. Configure a single Kind cluster instead of two
3. Provide a simpler alternative to the full multi-cluster deployment

For now:
- Use `test-cluster/` for single-cluster deployments
- Use `examples/multi-cluster/` for production-like multi-cluster setups

The shared scripts in `scripts/` work with both:
  LAGOON_PRESET=single ./scripts/check-cluster-health.sh
  LAGOON_PRESET=multi-prod ./scripts/check-cluster-health.sh
"""

import pulumi

pulumi.log.warn(
    "examples/single-cluster is a placeholder. "
    "Use test-cluster/ or examples/multi-cluster/ instead."
)

pulumi.log.info(
    "After multi-cluster branch merges, this will be refactored to use shared modules."
)

# Placeholder exports
pulumi.export("status", "placeholder")
pulumi.export("message", "Use test-cluster/ for single-cluster deployments")
