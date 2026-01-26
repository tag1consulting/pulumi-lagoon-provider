"""Lagoon build-deploy CRDs installation.

This module provides functions for installing the CustomResourceDefinitions
required by the lagoon-build-deploy controller.

The CRDs define:
- LagoonBuild: Represents a build/deploy operation in Lagoon
- LagoonTask: Represents a task (drush command, backup, etc.) in Lagoon

These CRDs must be installed before the lagoon-remote Helm chart with
lagoon-build-deploy enabled, otherwise the controller will fail to start.
"""

from typing import Optional

import pulumi
import pulumi_kubernetes as k8s


def install_lagoon_build_deploy_crds(
    name: str,
    provider: k8s.Provider,
    opts: Optional[pulumi.ResourceOptions] = None,
) -> k8s.yaml.ConfigGroup:
    """Install the Lagoon build-deploy CRDs.

    These CRDs are required by the lagoon-build-deploy controller and must
    be installed before the lagoon-remote Helm release.

    Args:
        name: Pulumi resource name prefix
        provider: Kubernetes provider for the target cluster
        opts: Pulumi resource options

    Returns:
        ConfigGroup containing the CRD resources
    """
    # CRD definitions from lagoon-build-deploy chart v0.39.0
    # Generated with: helm show crds uselagoon/lagoon-build-deploy --version ~0.39.0
    crd_yaml = """
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.16.5
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
  - additionalPrinterColumns:
    - description: Status of the LagoonBuild
      jsonPath: .status.phase
      name: Status
      type: string
    - description: The build step of the LagoonBuild
      jsonPath: .status.conditions[?(@.type == "BuildStep")].reason
      name: BuildStep
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1beta1
    schema:
      openAPIV3Schema:
        description: LagoonBuild is the Schema for the lagoonbuilds API
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            description: LagoonBuildSpec defines the desired state of LagoonBuild
            type: object
          status:
            description: LagoonBuildStatus defines the observed state of LagoonBuild
            properties:
              conditions:
                items:
                  properties:
                    lastTransitionTime:
                      format: date-time
                      type: string
                    message:
                      maxLength: 32768
                      type: string
                    observedGeneration:
                      format: int64
                      minimum: 0
                      type: integer
                    reason:
                      maxLength: 1024
                      minLength: 1
                      pattern: ^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$
                      type: string
                    status:
                      enum:
                      - "True"
                      - "False"
                      - Unknown
                      type: string
                    type:
                      maxLength: 316
                      pattern: ^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$
                      type: string
                  required:
                  - lastTransitionTime
                  - message
                  - reason
                  - status
                  - type
                  type: object
                type: array
              phase:
                type: string
            type: object
        type: object
    served: false
    storage: false
    subresources: {}
  - additionalPrinterColumns:
    - description: Status of the LagoonBuild
      jsonPath: .status.phase
      name: Status
      type: string
    - description: The build step of the LagoonBuild
      jsonPath: .status.conditions[?(@.type == "BuildStep")].reason
      name: BuildStep
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1beta2
    schema:
      openAPIV3Schema:
        description: LagoonBuild is the Schema for the lagoonbuilds API
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            description: LagoonBuildSpec defines the desired state of LagoonBuild
            properties:
              branch:
                properties:
                  name:
                    type: string
                type: object
              build:
                properties:
                  bulkId:
                    type: string
                  ci:
                    type: string
                  image:
                    type: string
                  priority:
                    type: integer
                  type:
                    type: string
                required:
                - type
                type: object
              gitReference:
                type: string
              project:
                properties:
                  deployTarget:
                    type: string
                  environment:
                    type: string
                  environmentId:
                    type: integer
                  environmentIdling:
                    type: integer
                  environmentType:
                    type: string
                  gitUrl:
                    type: string
                  id:
                    type: integer
                  key:
                    format: byte
                    type: string
                  monitoring:
                    properties:
                      contact:
                        type: string
                      statuspageID:
                        type: string
                    type: object
                  name:
                    type: string
                  namespacePattern:
                    type: string
                  organization:
                    properties:
                      id:
                        type: integer
                      name:
                        type: string
                    type: object
                  productionEnvironment:
                    type: string
                  projectIdling:
                    type: integer
                  projectSecret:
                    type: string
                  registry:
                    type: string
                  routerPattern:
                    type: string
                  standbyEnvironment:
                    type: string
                  storageCalculator:
                    type: integer
                  subfolder:
                    type: string
                  uiLink:
                    type: string
                  variables:
                    properties:
                      environment:
                        format: byte
                        type: string
                      project:
                        format: byte
                        type: string
                    type: object
                required:
                - deployTarget
                - environment
                - environmentType
                - gitUrl
                - key
                - monitoring
                - name
                - productionEnvironment
                - projectSecret
                - standbyEnvironment
                - variables
                type: object
              promote:
                properties:
                  sourceEnvironment:
                    type: string
                  sourceProject:
                    type: string
                type: object
              pullrequest:
                properties:
                  baseBranch:
                    type: string
                  baseSha:
                    type: string
                  headBranch:
                    type: string
                  headSha:
                    type: string
                  number:
                    type: string
                  title:
                    type: string
                type: object
            required:
            - build
            - gitReference
            - project
            type: object
          status:
            description: LagoonBuildStatus defines the observed state of LagoonBuild
            properties:
              conditions:
                items:
                  properties:
                    lastTransitionTime:
                      format: date-time
                      type: string
                    message:
                      maxLength: 32768
                      type: string
                    observedGeneration:
                      format: int64
                      minimum: 0
                      type: integer
                    reason:
                      maxLength: 1024
                      minLength: 1
                      pattern: ^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$
                      type: string
                    status:
                      enum:
                      - "True"
                      - "False"
                      - Unknown
                      type: string
                    type:
                      maxLength: 316
                      pattern: ^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$
                      type: string
                  required:
                  - lastTransitionTime
                  - message
                  - reason
                  - status
                  - type
                  type: object
                type: array
              phase:
                type: string
            type: object
        type: object
    served: true
    storage: true
    subresources: {}
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.16.5
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
  - additionalPrinterColumns:
    - description: Status of the LagoonTask
      jsonPath: .status.phase
      name: Status
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1beta1
    schema:
      openAPIV3Schema:
        description: LagoonTask is the Schema for the lagoontasks API
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            description: LagoonTaskSpec defines the desired state of LagoonTask
            type: object
          status:
            description: LagoonTaskStatus defines the observed state of LagoonTask
            properties:
              conditions:
                items:
                  properties:
                    lastTransitionTime:
                      format: date-time
                      type: string
                    message:
                      maxLength: 32768
                      type: string
                    observedGeneration:
                      format: int64
                      minimum: 0
                      type: integer
                    reason:
                      maxLength: 1024
                      minLength: 1
                      pattern: ^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$
                      type: string
                    status:
                      enum:
                      - "True"
                      - "False"
                      - Unknown
                      type: string
                    type:
                      maxLength: 316
                      pattern: ^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$
                      type: string
                  required:
                  - lastTransitionTime
                  - message
                  - reason
                  - status
                  - type
                  type: object
                type: array
              phase:
                type: string
            type: object
        type: object
    served: false
    storage: false
    subresources: {}
  - additionalPrinterColumns:
    - description: Status of the LagoonTask
      jsonPath: .status.phase
      name: Status
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1beta2
    schema:
      openAPIV3Schema:
        description: LagoonTask is the Schema for the lagoontasks API
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            description: LagoonTaskSpec defines the desired state of LagoonTask
            properties:
              advancedTask:
                properties:
                  JSONPayload:
                    type: string
                  deployerToken:
                    type: boolean
                  runnerImage:
                    type: string
                  sshKey:
                    type: boolean
                type: object
              environment:
                properties:
                  environmentType:
                    type: string
                  id:
                    type: integer
                  name:
                    type: string
                  project:
                    type: string
                required:
                - environmentType
                - name
                - project
                type: object
              key:
                type: string
              misc:
                properties:
                  backup:
                    properties:
                      backupId:
                        type: string
                      id:
                        type: string
                      source:
                        type: string
                    required:
                    - backupId
                    - id
                    - source
                    type: object
                  id:
                    type: string
                  miscResource:
                    format: byte
                    type: string
                  name:
                    type: string
                required:
                - id
                type: object
              project:
                properties:
                  id:
                    type: integer
                  name:
                    type: string
                  namespacePattern:
                    type: string
                  organization:
                    properties:
                      id:
                        type: integer
                      name:
                        type: string
                    type: object
                  variables:
                    properties:
                      environment:
                        format: byte
                        type: string
                      project:
                        format: byte
                        type: string
                    type: object
                required:
                - name
                type: object
              task:
                properties:
                  apiHost:
                    type: string
                  command:
                    type: string
                  id:
                    type: string
                  name:
                    type: string
                  service:
                    type: string
                  sshHost:
                    type: string
                  sshPort:
                    type: string
                  taskName:
                    type: string
                required:
                - id
                type: object
            type: object
          status:
            description: LagoonTaskStatus defines the observed state of LagoonTask
            properties:
              conditions:
                items:
                  properties:
                    lastTransitionTime:
                      format: date-time
                      type: string
                    message:
                      maxLength: 32768
                      type: string
                    observedGeneration:
                      format: int64
                      minimum: 0
                      type: integer
                    reason:
                      maxLength: 1024
                      minLength: 1
                      pattern: ^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$
                      type: string
                    status:
                      enum:
                      - "True"
                      - "False"
                      - Unknown
                      type: string
                    type:
                      maxLength: 316
                      pattern: ^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$
                      type: string
                  required:
                  - lastTransitionTime
                  - message
                  - reason
                  - status
                  - type
                  type: object
                type: array
              phase:
                type: string
            type: object
        type: object
    served: true
    storage: true
    subresources: {}
"""

    crds = k8s.yaml.ConfigGroup(
        f"{name}-lagoon-crds",
        yaml=crd_yaml,
        opts=pulumi.ResourceOptions(
            provider=provider,
            parent=opts.parent if opts else None,
            depends_on=opts.depends_on if opts else None,
        ),
    )

    return crds
