# Configuration Reference Files

These YAML files are **reference templates only** and are not used by the Pulumi program.

The actual Kind cluster configuration is generated dynamically by `clusters/kind.py`, and Lagoon Helm values are constructed inline in `lagoon/core.py`.

- `kind-prod.yaml` — Reference Kind config for the production cluster
- `kind-nonprod.yaml` — Reference Kind config for the non-production cluster
- `lagoon-values.yaml` — Reference Lagoon Helm values (uses placeholder domain `lagoon.test`)
