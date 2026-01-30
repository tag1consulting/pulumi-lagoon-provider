"""GraphQL client for Lagoon API."""

import json
import os
from typing import Any, Dict, Optional

import requests


class LagoonAPIError(Exception):
    """Exception raised for Lagoon API errors."""

    pass


class LagoonConnectionError(Exception):
    """Exception raised for connection errors."""

    pass


class LagoonClient:
    """GraphQL API client for Lagoon."""

    def __init__(self, api_url: str, token: str, verify_ssl: Optional[bool] = None):
        """
        Initialize Lagoon API client.

        Args:
            api_url: Lagoon GraphQL API endpoint URL
            token: Authentication token (JWT)
            verify_ssl: Whether to verify SSL certificates (default: True,
                        can be overridden with LAGOON_INSECURE=true env var)
        """
        self.api_url = api_url
        self.token = token

        # Determine SSL verification setting
        if verify_ssl is None:
            # Check environment variable
            insecure = os.environ.get("LAGOON_INSECURE", "").lower()
            self.verify_ssl = insecure not in ("true", "1", "yes")
        else:
            self.verify_ssl = verify_ssl

        self.session = requests.Session()
        self.session.headers.update(
            {
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json",
            }
        )

        # Disable SSL verification warnings if insecure mode
        if not self.verify_ssl:
            import urllib3

            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    def _execute(self, query: str, variables: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Execute a GraphQL query or mutation.

        Args:
            query: GraphQL query or mutation string
            variables: Optional variables for the query

        Returns:
            The data portion of the GraphQL response

        Raises:
            LagoonAPIError: If the API returns an error
            LagoonConnectionError: If there's a connection issue
        """
        payload: Dict[str, Any] = {
            "query": query,
        }

        if variables:
            payload["variables"] = variables

        try:
            response = self.session.post(
                self.api_url, json=payload, timeout=30, verify=self.verify_ssl
            )
            response.raise_for_status()

            data = response.json()

            # Check for GraphQL errors
            if "errors" in data:
                error_messages = [error.get("message", str(error)) for error in data["errors"]]
                raise LagoonAPIError(f"GraphQL errors: {'; '.join(error_messages)}")

            # Return the data portion
            return data.get("data", {})

        except requests.HTTPError as e:
            raise LagoonConnectionError(f"HTTP error: {e}")
        except requests.RequestException as e:
            raise LagoonConnectionError(f"Connection error: {e}")
        except json.JSONDecodeError as e:
            raise LagoonAPIError(f"Invalid JSON response: {e}")

    # Project operations
    def create_project(self, name: str, git_url: str, openshift: int, **kwargs) -> Dict[str, Any]:
        """
        Create a new Lagoon project.

        Args:
            name: Project name
            git_url: Git repository URL
            openshift: Deploy target (OpenShift/Kubernetes) ID
            **kwargs: Additional project properties

        Returns:
            Created project data
        """
        mutation = """
        mutation AddProject($input: AddProjectInput!) {
            addProject(input: $input) {
                id
                name
                gitUrl
                openshift {
                    id
                    name
                }
                productionEnvironment
                branches
                pullrequests
                created
            }
        }
        """

        input_data = {"name": name, "gitUrl": git_url, "openshift": openshift, **kwargs}

        result = self._execute(mutation, {"input": input_data})
        project = result.get("addProject", {})

        # Normalize openshift to just the ID for consistency
        if project.get("openshift") and isinstance(project["openshift"], dict):
            project["openshift"] = project["openshift"].get("id")

        return project

    def get_project_by_name(self, name: str) -> Optional[Dict[str, Any]]:
        """
        Get project details by name.

        Args:
            name: Project name

        Returns:
            Project data or None if not found
        """
        query = """
        query ProjectByName($name: String!) {
            projectByName(name: $name) {
                id
                name
                gitUrl
                openshift {
                    id
                    name
                }
                productionEnvironment
                branches
                pullrequests
                created
            }
        }
        """

        result = self._execute(query, {"name": name})
        project = result.get("projectByName")

        # Normalize openshift to just the ID for consistency
        if project and project.get("openshift") and isinstance(project["openshift"], dict):
            project["openshift"] = project["openshift"].get("id")

        return project

    def get_project_by_id(self, project_id: int) -> Optional[Dict[str, Any]]:
        """
        Get project details by ID.

        Args:
            project_id: Project ID

        Returns:
            Project data or None if not found
        """
        query = """
        query ProjectById($id: Int!) {
            projectById(id: $id) {
                id
                name
                gitUrl
                openshift {
                    id
                    name
                }
                productionEnvironment
                branches
                pullrequests
                created
            }
        }
        """

        result = self._execute(query, {"id": project_id})
        project = result.get("projectById")

        # Normalize openshift to just the ID for consistency
        if project and project.get("openshift") and isinstance(project["openshift"], dict):
            project["openshift"] = project["openshift"].get("id")

        return project

    def update_project(self, project_id: int, **kwargs) -> Dict[str, Any]:
        """
        Update an existing project.

        Args:
            project_id: Project ID
            **kwargs: Project properties to update

        Returns:
            Updated project data
        """
        mutation = """
        mutation UpdateProject($input: UpdateProjectInput!) {
            updateProject(input: $input) {
                id
                name
                gitUrl
                openshift {
                    id
                    name
                }
                productionEnvironment
                branches
                pullrequests
            }
        }
        """

        input_data = {"id": project_id, **kwargs}

        result = self._execute(mutation, {"input": input_data})
        project = result.get("updateProject", {})

        # Normalize openshift to just the ID for consistency
        if project.get("openshift") and isinstance(project["openshift"], dict):
            project["openshift"] = project["openshift"].get("id")

        return project

    def delete_project(self, name: str) -> str:
        """
        Delete a project.

        Args:
            name: Project name

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteProject($input: DeleteProjectInput!) {
            deleteProject(input: $input)
        }
        """

        result = self._execute(mutation, {"input": {"project": name}})
        return result.get("deleteProject", "")

    # Environment operations
    def add_or_update_environment(
        self, name: str, project: int, deploy_type: str, environment_type: str, **kwargs
    ) -> Dict[str, Any]:
        """
        Add or update an environment.

        Args:
            name: Environment name (typically branch name)
            project: Project ID
            deploy_type: "branch" or "pullrequest"
            environment_type: "production", "development", etc.
            **kwargs: Additional environment properties

        Returns:
            Environment data
        """
        mutation = """
        mutation AddOrUpdateEnvironment($input: AddEnvironmentInput!) {
            addOrUpdateEnvironment(input: $input) {
                id
                name
                project {
                    id
                    name
                }
                environmentType
                deployType
                deployBaseRef
                deployHeadRef
                deployTitle
                autoIdle
                route
                routes
                created
            }
        }
        """

        input_data = {
            "name": name,
            "project": project,
            "deployType": deploy_type.upper(),  # Lagoon expects uppercase enum
            "environmentType": environment_type.upper(),  # Lagoon expects uppercase enum
            **kwargs,
        }

        result = self._execute(mutation, {"input": input_data})
        return result.get("addOrUpdateEnvironment", {})

    def get_environment_by_name(self, name: str, project_id: int) -> Optional[Dict[str, Any]]:
        """
        Get environment by name and project.

        Args:
            name: Environment name
            project_id: Project ID

        Returns:
            Environment data or None if not found
        """
        query = """
        query EnvironmentByName($name: String!, $project: Int!) {
            environmentByName(name: $name, project: $project) {
                id
                name
                project {
                    id
                    name
                }
                environmentType
                deployType
                route
                routes
                created
            }
        }
        """

        result = self._execute(query, {"name": name, "project": project_id})
        return result.get("environmentByName")

    def delete_environment(self, name: str, project: int, execute: bool = False) -> str:
        """
        Delete an environment.

        Args:
            name: Environment name
            project: Project ID
            execute: Whether to actually execute the deletion (Lagoon safety feature)

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteEnvironment($input: DeleteEnvironmentInput!) {
            deleteEnvironment(input: $input)
        }
        """

        input_data = {"name": name, "project": project, "execute": execute}

        result = self._execute(mutation, {"input": input_data})
        return result.get("deleteEnvironment", "")

    # Variable operations
    def add_env_variable(
        self,
        name: str,
        value: str,
        project: int,
        scope: str,
        environment: Optional[int] = None,
        **kwargs,
    ) -> Dict[str, Any]:
        """
        Add an environment variable.

        Args:
            name: Variable name
            value: Variable value
            project: Project ID
            scope: Variable scope ("build", "runtime", "global", "container_registry", "internal_container_registry")
            environment: Environment ID (optional, for environment-scoped variables)
            **kwargs: Additional variable properties

        Returns:
            Variable data
        """
        mutation = """
        mutation AddEnvVariable($input: EnvVariableInput!) {
            addEnvVariable(input: $input) {
                id
                name
                value
                scope
            }
        }
        """

        # Lagoon uses type/typeId to specify whether this is a project or environment variable
        if environment is not None:
            input_data = {
                "name": name,
                "value": value,
                "type": "ENVIRONMENT",
                "typeId": environment,
                "scope": scope.upper(),  # Lagoon expects uppercase
                **kwargs,
            }
        else:
            input_data = {
                "name": name,
                "value": value,
                "type": "PROJECT",
                "typeId": project,
                "scope": scope.upper(),  # Lagoon expects uppercase
                **kwargs,
            }

        result = self._execute(mutation, {"input": input_data})
        return result.get("addEnvVariable", {})

    def get_env_variable_by_name(
        self, name: str, project: int, environment: Optional[int] = None
    ) -> Optional[Dict[str, Any]]:
        """
        Get an environment variable by name.

        Args:
            name: Variable name
            project: Project ID
            environment: Environment ID (optional)

        Returns:
            Variable data or None if not found
        """
        # Note: Lagoon doesn't have a direct query for single variable
        # We need to get all variables and filter
        query = """
        query EnvVariablesByProjectEnvironment($project: Int!, $environment: Int) {
            envVariablesByProjectEnvironment(input: {project: $project, environment: $environment}) {
                id
                name
                value
                scope
                project {
                    id
                    name
                }
                environment {
                    id
                    name
                }
            }
        }
        """

        variables = {"project": project}
        if environment is not None:
            variables["environment"] = environment

        result = self._execute(query, variables)
        all_vars = result.get("envVariablesByProjectEnvironment", [])

        # Filter for the specific variable name
        for var in all_vars:
            if var.get("name") == name:
                return var

        return None

    def delete_env_variable(
        self, name: str, project: int, environment: Optional[int] = None
    ) -> str:
        """
        Delete an environment variable.

        Args:
            name: Variable name
            project: Project ID
            environment: Environment ID (optional, for environment-scoped variables)

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteEnvVariable($input: DeleteEnvVariableInput!) {
            deleteEnvVariable(input: $input)
        }
        """

        input_data = {"name": name, "project": project}

        if environment is not None:
            input_data["environment"] = environment

        result = self._execute(mutation, {"input": input_data})
        return result.get("deleteEnvVariable", "")

    # Deploy target (Kubernetes) operations
    def add_kubernetes(
        self,
        name: str,
        console_url: str,
        cloud_provider: str = "kind",
        cloud_region: str = "local",
        **kwargs,
    ) -> Dict[str, Any]:
        """
        Add a Kubernetes deploy target.

        Args:
            name: Deploy target name
            console_url: Kubernetes API URL
            cloud_provider: Cloud provider (e.g., "aws", "gcp", "kind")
            cloud_region: Cloud region (e.g., "us-east-1", "local")
            **kwargs: Additional properties (sshHost, sshPort, buildImage, disabled, etc.)

        Returns:
            Deploy target data
        """
        mutation = """
        mutation AddKubernetes($input: AddKubernetesInput!) {
            addKubernetes(input: $input) {
                id
                name
                consoleUrl
                cloudProvider
                cloudRegion
                sshHost
                sshPort
                buildImage
                disabled
                routerPattern
                created
            }
        }
        """

        input_data = {
            "name": name,
            "consoleUrl": console_url,
            "cloudProvider": cloud_provider,
            "cloudRegion": cloud_region,
            **kwargs,
        }

        result = self._execute(mutation, {"input": input_data})
        return result.get("addKubernetes", {})

    def get_all_kubernetes(self) -> list:
        """
        Get all Kubernetes deploy targets.

        Returns:
            List of deploy target data
        """
        query = """
        query AllKubernetes {
            allKubernetes {
                id
                name
                consoleUrl
                cloudProvider
                cloudRegion
                sshHost
                sshPort
                buildImage
                disabled
                routerPattern
                created
            }
        }
        """

        result = self._execute(query)
        return result.get("allKubernetes", [])

    def get_kubernetes_by_id(self, k8s_id: int) -> Optional[Dict[str, Any]]:
        """
        Get Kubernetes deploy target by ID.

        Args:
            k8s_id: Kubernetes/deploy target ID

        Returns:
            Deploy target data or None if not found
        """
        query = """
        query KubernetesById($id: Int!) {
            kubernetes(id: $id) {
                id
                name
                consoleUrl
                cloudProvider
                cloudRegion
                sshHost
                sshPort
                buildImage
                disabled
                routerPattern
                created
            }
        }
        """

        result = self._execute(query, {"id": k8s_id})
        return result.get("kubernetes")

    def get_kubernetes_by_name(self, name: str) -> Optional[Dict[str, Any]]:
        """
        Get Kubernetes deploy target by name.

        Args:
            name: Deploy target name

        Returns:
            Deploy target data or None if not found
        """
        # Lagoon doesn't have a direct query for Kubernetes by name,
        # so we get all and filter
        all_k8s = self.get_all_kubernetes()

        for k8s in all_k8s:
            if k8s.get("name") == name:
                return k8s

        return None

    def update_kubernetes(self, k8s_id: int, **kwargs) -> Dict[str, Any]:
        """
        Update a Kubernetes deploy target.

        Args:
            k8s_id: Kubernetes/deploy target ID
            **kwargs: Properties to update

        Returns:
            Updated deploy target data
        """
        mutation = """
        mutation UpdateKubernetes($input: UpdateKubernetesInput!) {
            updateKubernetes(input: $input) {
                id
                name
                consoleUrl
                cloudProvider
                cloudRegion
                sshHost
                sshPort
                buildImage
                disabled
                routerPattern
            }
        }
        """

        input_data = {"id": k8s_id, **kwargs}

        result = self._execute(mutation, {"input": input_data})
        return result.get("updateKubernetes", {})

    def delete_kubernetes(self, name: str) -> str:
        """
        Delete a Kubernetes deploy target.

        Args:
            name: Deploy target name

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteKubernetes($input: DeleteKubernetesInput!) {
            deleteKubernetes(input: $input)
        }
        """

        input_data = {"name": name}

        result = self._execute(mutation, {"input": input_data})
        return result.get("deleteKubernetes", "")

    # Deploy target configuration operations
    def add_deploy_target_config(
        self,
        project: int,
        deploy_target: int,
        branches: str = "",
        pullrequests: str = "false",
        weight: int = 1,
        deploy_target_project_pattern: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Add a deploy target configuration to route specific branches/PRs to a cluster.

        Deploy target configurations allow routing different branches or pull requests
        to different Kubernetes clusters. Higher weight configurations take priority.

        Args:
            project: Project ID
            deploy_target: Deploy target (Kubernetes) ID
            branches: Regex pattern for branches to match (e.g., "main", "^feature/.*$")
            pullrequests: Whether to handle PRs ("true" or "false")
            weight: Priority weight (higher = more priority)
            deploy_target_project_pattern: Optional namespace pattern

        Returns:
            Deploy target config data
        """
        mutation = """
        mutation AddDeployTargetConfig($input: AddDeployTargetConfigInput!) {
            addDeployTargetConfig(input: $input) {
                id
                weight
                branches
                pullrequests
                deployTargetProjectPattern
                deployTarget {
                    id
                    name
                }
                project {
                    id
                    name
                }
            }
        }
        """

        input_data = {
            "project": project,
            "deployTarget": deploy_target,
            "branches": branches,
            "pullrequests": pullrequests,
            "weight": weight,
        }

        if deploy_target_project_pattern:
            input_data["deployTargetProjectPattern"] = deploy_target_project_pattern

        result = self._execute(mutation, {"input": input_data})
        config = result.get("addDeployTargetConfig", {})

        # Normalize nested objects to IDs for consistency
        if config.get("deployTarget") and isinstance(config["deployTarget"], dict):
            config["deployTargetId"] = config["deployTarget"].get("id")
        if config.get("project") and isinstance(config["project"], dict):
            config["projectId"] = config["project"].get("id")

        return config

    def get_deploy_target_configs_by_project(self, project: int) -> list:
        """
        Get all deploy target configurations for a project.

        Args:
            project: Project ID

        Returns:
            List of deploy target config data
        """
        query = """
        query DeployTargetConfigsByProjectId($project: Int!) {
            deployTargetConfigsByProjectId(project: $project) {
                id
                weight
                branches
                pullrequests
                deployTargetProjectPattern
                deployTarget {
                    id
                    name
                }
                project {
                    id
                    name
                }
            }
        }
        """

        result = self._execute(query, {"project": project})
        configs = result.get("deployTargetConfigsByProjectId", [])

        # Normalize nested objects
        for config in configs:
            if config.get("deployTarget") and isinstance(config["deployTarget"], dict):
                config["deployTargetId"] = config["deployTarget"].get("id")
            if config.get("project") and isinstance(config["project"], dict):
                config["projectId"] = config["project"].get("id")

        return configs

    def get_deploy_target_config_by_id(
        self, config_id: int, project: int
    ) -> Optional[Dict[str, Any]]:
        """
        Get a deploy target configuration by ID.

        Args:
            config_id: Deploy target config ID
            project: Project ID (needed to query configs)

        Returns:
            Deploy target config data or None if not found
        """
        configs = self.get_deploy_target_configs_by_project(project)

        for config in configs:
            if config.get("id") == config_id:
                return config

        return None

    def update_deploy_target_config(self, config_id: int, **kwargs) -> Dict[str, Any]:
        """
        Update a deploy target configuration.

        Args:
            config_id: Deploy target config ID
            **kwargs: Properties to update (branches, pullrequests, weight, etc.)

        Returns:
            Updated deploy target config data
        """
        mutation = """
        mutation UpdateDeployTargetConfig($input: UpdateDeployTargetConfigInput!) {
            updateDeployTargetConfig(input: $input) {
                id
                weight
                branches
                pullrequests
                deployTargetProjectPattern
                deployTarget {
                    id
                    name
                }
                project {
                    id
                    name
                }
            }
        }
        """

        input_data = {"id": config_id, **kwargs}

        result = self._execute(mutation, {"input": input_data})
        config = result.get("updateDeployTargetConfig", {})

        # Normalize nested objects
        if config.get("deployTarget") and isinstance(config["deployTarget"], dict):
            config["deployTargetId"] = config["deployTarget"].get("id")
        if config.get("project") and isinstance(config["project"], dict):
            config["projectId"] = config["project"].get("id")

        return config

    def delete_deploy_target_config(self, config_id: int, project: int) -> str:
        """
        Delete a deploy target configuration.

        Args:
            config_id: Deploy target config ID
            project: Project ID

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteDeployTargetConfig($input: DeleteDeployTargetConfigInput!) {
            deleteDeployTargetConfig(input: $input)
        }
        """

        input_data = {"id": config_id, "project": project}

        result = self._execute(mutation, {"input": input_data})
        return result.get("deleteDeployTargetConfig", "")
