"""Lagoon build-deploy CRDs installation.

This module provides functions for installing the CustomResourceDefinitions
required by the lagoon-build-deploy controller.

The CRDs define:
- LagoonBuild: Represents a build/deploy operation in Lagoon
- LagoonTask: Represents a task (drush command, backup, etc.) in Lagoon

These CRDs must be installed before the lagoon-remote Helm chart with
lagoon-build-deploy enabled, otherwise the controller will fail to start.

Version compatibility:
- lagoon-remote requires BOTH v1beta1 and v1beta2 API versions to be served
- Lagoon <= v2.28.0 (chart < 1.58.0): Uses v1beta1 as storage version
- Lagoon >= v2.29.0 (chart >= 1.58.0): Uses v1beta2 as storage version
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s
from pulumi_command import local as command


def _parse_version(version: str) -> tuple:
    """Parse a semver version string into a tuple for comparison."""
    # Handle versions like "1.56.0" or "v2.28.0"
    version = version.lstrip("v")
    parts = version.split(".")
    return tuple(int(p) for p in parts[:3])


def _get_storage_version(lagoon_core_version: str) -> str:
    """Determine which CRD API version should be the storage version.

    Args:
        lagoon_core_version: The lagoon-core Helm chart version (e.g., "1.56.0")

    Returns:
        "v1beta1" for older versions, "v1beta2" for newer versions
    """
    try:
        version_tuple = _parse_version(lagoon_core_version)
        # Chart version 1.58.0 and later use v1beta2 as storage
        # Chart versions before 1.58.0 use v1beta1 as storage
        if version_tuple >= (1, 58, 0):
            return "v1beta2"
        else:
            return "v1beta1"
    except (ValueError, IndexError):
        # Default to v1beta2 for unparseable versions (assume latest)
        return "v1beta2"


def _generate_crd_yaml(storage_version: str) -> str:
    """Generate CRD YAML with both v1beta1 and v1beta2 served.

    lagoon-remote requires both API versions to be available. The storage_version
    parameter determines which version is used for storing objects.

    Args:
        storage_version: Either "v1beta1" or "v1beta2"

    Returns:
        YAML string with both CRD definitions
    """
    v1beta1_storage = "true" if storage_version == "v1beta1" else "false"
    v1beta2_storage = "true" if storage_version == "v1beta2" else "false"

    return f"""
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: lagoonbuilds.crd.lagoon.sh
spec:
  group: crd.lagoon.sh
  names:
    kind: LagoonBuild
    listKind: LagoonBuildList
    plural: lagoonbuilds
    singular: lagoonbuild
  scope: Namespaced
  versions:
  - name: v1beta1
    served: true
    storage: {v1beta1_storage}
    additionalPrinterColumns:
    - description: Status of the LagoonBuild
      jsonPath: .status.phase
      name: Status
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    schema:
      openAPIV3Schema:
        description: LagoonBuild is the Schema for the lagoonbuilds API
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
            x-kubernetes-preserve-unknown-fields: true
          status:
            type: object
            x-kubernetes-preserve-unknown-fields: true
    subresources: {{}}
  - name: v1beta2
    served: true
    storage: {v1beta2_storage}
    additionalPrinterColumns:
    - description: Status of the LagoonBuild
      jsonPath: .status.phase
      name: Status
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    schema:
      openAPIV3Schema:
        description: LagoonBuild is the Schema for the lagoonbuilds API
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
            x-kubernetes-preserve-unknown-fields: true
          status:
            type: object
            x-kubernetes-preserve-unknown-fields: true
    subresources: {{}}
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: lagoontasks.crd.lagoon.sh
spec:
  group: crd.lagoon.sh
  names:
    kind: LagoonTask
    listKind: LagoonTaskList
    plural: lagoontasks
    singular: lagoontask
  scope: Namespaced
  versions:
  - name: v1beta1
    served: true
    storage: {v1beta1_storage}
    additionalPrinterColumns:
    - description: Status of the LagoonTask
      jsonPath: .status.phase
      name: Status
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    schema:
      openAPIV3Schema:
        description: LagoonTask is the Schema for the lagoontasks API
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
            x-kubernetes-preserve-unknown-fields: true
          status:
            type: object
            x-kubernetes-preserve-unknown-fields: true
    subresources: {{}}
  - name: v1beta2
    served: true
    storage: {v1beta2_storage}
    additionalPrinterColumns:
    - description: Status of the LagoonTask
      jsonPath: .status.phase
      name: Status
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    schema:
      openAPIV3Schema:
        description: LagoonTask is the Schema for the lagoontasks API
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
            x-kubernetes-preserve-unknown-fields: true
          status:
            type: object
            x-kubernetes-preserve-unknown-fields: true
    subresources: {{}}
"""


def install_lagoon_build_deploy_crds(
    name: str,
    provider: k8s.Provider,
    context: Optional[str] = None,
    lagoon_core_version: Optional[str] = None,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> command.Command:
    """Install the Lagoon build-deploy CRDs.

    These CRDs are required by the lagoon-build-deploy controller and must
    be installed before the lagoon-remote Helm release.

    Uses kubectl apply to install CRDs, which is more reliable than
    pulumi-kubernetes YAML parsing for complex CRDs.

    Both v1beta1 and v1beta2 API versions are always served (required by
    lagoon-remote). The storage version is selected based on the Lagoon version:
    - Lagoon <= v2.28.0 (chart < 1.58.0): v1beta1 is storage version
    - Lagoon >= v2.29.0 (chart >= 1.58.0): v1beta2 is storage version

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster (kept for API compat)
        context: Kubernetes context name (e.g., "kind-lagoon")
        lagoon_core_version: The lagoon-core Helm chart version (e.g., "1.56.0")
            Used to determine which CRD API version is the storage version.
            If not provided, defaults to v1beta2 (latest).
        opts: Pulumi resource options

    Returns:
        Command resource that applies the CRDs
    """
    # Determine which version should be storage
    if lagoon_core_version:
        storage_version = _get_storage_version(lagoon_core_version)
    else:
        storage_version = "v1beta2"  # Default to latest

    pulumi.log.info(
        f"Installing Lagoon CRDs with {storage_version} as storage version "
        f"(for Lagoon chart {lagoon_core_version or 'default'})"
    )

    # Generate CRD YAML with appropriate storage version
    crd_yaml = _generate_crd_yaml(storage_version)

    # Write CRDs to a temp file and apply with kubectl
    # This avoids pulumi-kubernetes YAML parsing issues with complex CRDs
    crd_file = "/tmp/lagoon-crds.yaml"

    # Build the kubectl command with optional context
    context_flag = f"--context {context}" if context else ""
    kubectl_cmd = f"cat > {crd_file} << 'EOFCRDS'\n{crd_yaml}\nEOFCRDS\nkubectl apply -f {crd_file} {context_flag}"

    crds = command.Command(
        f"{name}-lagoon-crds",
        create=kubectl_cmd,
        # Include version in triggers to force re-apply when version changes
        triggers=[storage_version, lagoon_core_version or "default"],
        opts=pulumi.ResourceOptions(
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )

    return crds
