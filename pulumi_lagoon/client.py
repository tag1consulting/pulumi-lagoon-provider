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

        Note: Lagoon v2.30.0+ does not have a projectById query, so we query
        allProjects and filter by ID.

        Args:
            project_id: Project ID

        Returns:
            Project data or None if not found
        """
        query = """
        query AllProjects {
            allProjects {
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

        result = self._execute(query)
        all_projects = result.get("allProjects", [])

        # Filter for the specific project ID
        for project in all_projects:
            if project.get("id") == project_id:
                # Normalize openshift to just the ID for consistency
                if project.get("openshift") and isinstance(project["openshift"], dict):
                    project["openshift"] = project["openshift"].get("id")
                return project

        return None

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

        Note:
            Uses addOrUpdateEnvVariableByName for Lagoon v2.30.0+
            (which takes project/environment names as Strings)
            with fallback to addEnvVariable for older versions
            (which takes project/environment IDs as Ints).
        """
        # Try newer API first (Lagoon v2.30.0+)
        # The new API uses project NAME (String) not ID (Int)
        mutation_new = """
        mutation AddOrUpdateEnvVariableByName($input: EnvVariableByNameInput!) {
            addOrUpdateEnvVariableByName(input: $input) {
                id
                name
                value
                scope
            }
        }
        """

        # Get project name from ID for the new API
        project_data = self.get_project_by_id(project)
        if project_data is None:
            raise LagoonAPIError(f"Project with ID {project} not found")
        project_name = project_data.get("name")

        # Build input for new API using names
        input_data_new: Dict[str, Any] = {
            "name": name,
            "value": value,
            "scope": scope.upper(),
            "project": project_name,
        }

        # If environment is specified, get environment name
        if environment is not None:
            env_query = """
            query EnvironmentById($id: Int!) {
                environmentById(id: $id) {
                    id
                    name
                }
            }
            """
            try:
                env_result = self._execute(env_query, {"id": environment})
                env_data = env_result.get("environmentById")
                if env_data:
                    input_data_new["environment"] = env_data.get("name")
            except (LagoonAPIError, LagoonConnectionError):
                # If environmentById fails, we'll handle it in the fallback
                pass

        try:
            result = self._execute(mutation_new, {"input": input_data_new})
            return result.get("addOrUpdateEnvVariableByName", {})
        except (LagoonAPIError, LagoonConnectionError) as e:
            # Fallback to older API for Lagoon versions < 2.30.0
            if "Cannot query field" in str(e) or "400" in str(e) or "Unknown argument" in str(e):
                mutation_old = """
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
                    input_data_old = {
                        "name": name,
                        "value": value,
                        "type": "ENVIRONMENT",
                        "typeId": environment,
                        "scope": scope.upper(),
                        **kwargs,
                    }
                else:
                    input_data_old = {
                        "name": name,
                        "value": value,
                        "type": "PROJECT",
                        "typeId": project,
                        "scope": scope.upper(),
                        **kwargs,
                    }
                result = self._execute(mutation_old, {"input": input_data_old})
                return result.get("addEnvVariable", {})
            else:
                raise

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

        Note:
            Uses getEnvVariablesByProjectEnvironmentName for Lagoon v2.30.0+
            (which takes project/environment names as Strings)
            with fallback to envVariablesByProjectEnvironment for older versions
            (which takes project/environment IDs as Ints).
        """
        # Try newer API first (Lagoon v2.30.0+)
        # The new API uses project NAME (String) not ID (Int)
        # Note: The new API returns EnvKeyValue type which doesn't have project/environment fields
        query_new = """
        query GetEnvVariablesByProjectEnvironmentName($input: EnvVariableByProjectEnvironmentNameInput!) {
            getEnvVariablesByProjectEnvironmentName(input: $input) {
                id
                name
                value
                scope
            }
        }
        """

        # Get project name from ID for the new API
        project_data = self.get_project_by_id(project)
        if project_data is None:
            return None
        project_name = project_data.get("name")

        # Build input for new API using names
        input_data: Dict[str, Any] = {"project": project_name}

        # If environment is specified, we need to get environment name too
        if environment is not None:
            # Query the environment to get its name
            env_query = """
            query EnvironmentById($id: Int!) {
                environmentById(id: $id) {
                    id
                    name
                }
            }
            """
            try:
                env_result = self._execute(env_query, {"id": environment})
                env_data = env_result.get("environmentById")
                if env_data:
                    input_data["environment"] = env_data.get("name")
            except (LagoonAPIError, LagoonConnectionError):
                # If environmentById fails, we'll handle it in the fallback
                pass

        try:
            result = self._execute(query_new, {"input": input_data})
            all_vars = result.get("getEnvVariablesByProjectEnvironmentName", [])
        except (LagoonAPIError, LagoonConnectionError) as e:
            # Fallback to older API for Lagoon versions < 2.30.0
            if "Cannot query field" in str(e) or "400" in str(e):
                query_old = """
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
                result = self._execute(query_old, variables)
                all_vars = result.get("envVariablesByProjectEnvironment", [])
            else:
                raise

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

        Note:
            Uses deleteEnvVariableByName for Lagoon v2.30.0+
            (which takes project/environment names as Strings)
            with fallback to deleteEnvVariable for older versions
            (which takes project/environment IDs as Ints).
        """
        # Try newer API first (Lagoon v2.30.0+)
        # The new API uses project NAME (String) not ID (Int)
        mutation_new = """
        mutation DeleteEnvVariableByName($input: DeleteEnvVariableByNameInput!) {
            deleteEnvVariableByName(input: $input)
        }
        """

        # Get project name from ID for the new API
        project_data = self.get_project_by_id(project)
        if project_data is None:
            raise LagoonAPIError(f"Project with ID {project} not found")
        project_name = project_data.get("name")

        # Build input for new API using names
        input_data_new: Dict[str, Any] = {"name": name, "project": project_name}

        # If environment is specified, get environment name
        if environment is not None:
            env_query = """
            query EnvironmentById($id: Int!) {
                environmentById(id: $id) {
                    id
                    name
                }
            }
            """
            try:
                env_result = self._execute(env_query, {"id": environment})
                env_data = env_result.get("environmentById")
                if env_data:
                    input_data_new["environment"] = env_data.get("name")
            except (LagoonAPIError, LagoonConnectionError):
                # If environmentById fails, we'll handle it in the fallback
                pass

        try:
            result = self._execute(mutation_new, {"input": input_data_new})
            return result.get("deleteEnvVariableByName", "")
        except (LagoonAPIError, LagoonConnectionError) as e:
            # Fallback to older API for Lagoon versions < 2.30.0
            if "Cannot query field" in str(e) or "400" in str(e):
                mutation_old = """
                mutation DeleteEnvVariable($input: DeleteEnvVariableInput!) {
                    deleteEnvVariable(input: $input)
                }
                """
                input_data_old: Dict[str, Any] = {"name": name, "project": project}
                if environment is not None:
                    input_data_old["environment"] = environment
                result = self._execute(mutation_old, {"input": input_data_old})
                return result.get("deleteEnvVariable", "")
            else:
                raise

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

        Note: Lagoon v2.30.0+ does not have a kubernetes(id:) query, so we query
        allKubernetes and filter by ID.

        Args:
            k8s_id: Kubernetes/deploy target ID

        Returns:
            Deploy target data or None if not found
        """
        # Query all and filter - compatible with all Lagoon versions
        all_k8s = self.get_all_kubernetes()

        for k8s in all_k8s:
            if k8s.get("id") == k8s_id:
                return k8s

        return None

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

    # Notification operations - Slack

    def add_notification_slack(self, name: str, webhook: str, channel: str) -> Dict[str, Any]:
        """
        Add a Slack notification.

        Args:
            name: Notification name
            webhook: Slack webhook URL
            channel: Slack channel (e.g., "#alerts")

        Returns:
            Notification data
        """
        mutation = """
        mutation AddNotificationSlack($input: AddNotificationSlackInput!) {
            addNotificationSlack(input: $input) {
                id
                name
                webhook
                channel
            }
        }
        """

        input_data = {"name": name, "webhook": webhook, "channel": channel}

        result = self._execute(mutation, {"input": input_data})
        return result.get("addNotificationSlack", {})

    def _get_all_notifications(self) -> list:
        """
        Get all notifications of all types.

        Returns:
            List of notification data with __typename
        """
        query = """
        query AllNotifications {
            allNotifications {
                __typename
                ... on NotificationSlack {
                    id
                    name
                    webhook
                    channel
                }
                ... on NotificationRocketChat {
                    id
                    name
                    webhook
                    channel
                }
                ... on NotificationEmail {
                    id
                    name
                    emailAddress
                }
                ... on NotificationMicrosoftTeams {
                    id
                    name
                    webhook
                }
            }
        }
        """

        result = self._execute(query)
        return result.get("allNotifications") or []

    def get_all_notification_slack(self) -> list:
        """
        Get all Slack notifications.

        Returns:
            List of Slack notification data
        """
        all_notifications = self._get_all_notifications()
        return [n for n in all_notifications if n.get("__typename") == "NotificationSlack"]

    def get_notification_slack_by_name(self, name: str) -> Optional[Dict[str, Any]]:
        """
        Get Slack notification by name.

        Args:
            name: Notification name

        Returns:
            Notification data or None if not found
        """
        all_notifications = self.get_all_notification_slack()

        for notification in all_notifications:
            if notification.get("name") == name:
                return notification

        return None

    def update_notification_slack(self, name: str, **kwargs) -> Dict[str, Any]:
        """
        Update a Slack notification.

        Args:
            name: Notification name (used to identify the notification)
            **kwargs: Properties to update (webhook, channel)

        Returns:
            Updated notification data
        """
        mutation = """
        mutation UpdateNotificationSlack($input: UpdateNotificationSlackInput!) {
            updateNotificationSlack(input: $input) {
                id
                name
                webhook
                channel
            }
        }
        """

        # Lagoon API uses patch format: {name: "...", patch: {field: value}}
        input_data = {"name": name, "patch": kwargs}

        result = self._execute(mutation, {"input": input_data})
        return result.get("updateNotificationSlack", {})

    def delete_notification_slack(self, name: str) -> str:
        """
        Delete a Slack notification.

        Args:
            name: Notification name

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteNotificationSlack($input: DeleteNotificationSlackInput!) {
            deleteNotificationSlack(input: $input)
        }
        """

        input_data = {"name": name}

        result = self._execute(mutation, {"input": input_data})
        return result.get("deleteNotificationSlack", "")

    # Notification operations - RocketChat

    def add_notification_rocketchat(self, name: str, webhook: str, channel: str) -> Dict[str, Any]:
        """
        Add a RocketChat notification.

        Args:
            name: Notification name
            webhook: RocketChat webhook URL
            channel: RocketChat channel (e.g., "#alerts")

        Returns:
            Notification data
        """
        mutation = """
        mutation AddNotificationRocketChat($input: AddNotificationRocketChatInput!) {
            addNotificationRocketChat(input: $input) {
                id
                name
                webhook
                channel
            }
        }
        """

        input_data = {"name": name, "webhook": webhook, "channel": channel}

        result = self._execute(mutation, {"input": input_data})
        return result.get("addNotificationRocketChat", {})

    def get_all_notification_rocketchat(self) -> list:
        """
        Get all RocketChat notifications.

        Returns:
            List of RocketChat notification data
        """
        all_notifications = self._get_all_notifications()
        return [n for n in all_notifications if n.get("__typename") == "NotificationRocketChat"]

    def get_notification_rocketchat_by_name(self, name: str) -> Optional[Dict[str, Any]]:
        """
        Get RocketChat notification by name.

        Args:
            name: Notification name

        Returns:
            Notification data or None if not found
        """
        all_notifications = self.get_all_notification_rocketchat()

        for notification in all_notifications:
            if notification.get("name") == name:
                return notification

        return None

    def update_notification_rocketchat(self, name: str, **kwargs) -> Dict[str, Any]:
        """
        Update a RocketChat notification.

        Args:
            name: Notification name (used to identify the notification)
            **kwargs: Properties to update (webhook, channel)

        Returns:
            Updated notification data
        """
        mutation = """
        mutation UpdateNotificationRocketChat($input: UpdateNotificationRocketChatInput!) {
            updateNotificationRocketChat(input: $input) {
                id
                name
                webhook
                channel
            }
        }
        """

        # Lagoon API uses patch format: {name: "...", patch: {field: value}}
        input_data = {"name": name, "patch": kwargs}

        result = self._execute(mutation, {"input": input_data})
        return result.get("updateNotificationRocketChat", {})

    def delete_notification_rocketchat(self, name: str) -> str:
        """
        Delete a RocketChat notification.

        Args:
            name: Notification name

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteNotificationRocketChat($input: DeleteNotificationRocketChatInput!) {
            deleteNotificationRocketChat(input: $input)
        }
        """

        input_data = {"name": name}

        result = self._execute(mutation, {"input": input_data})
        return result.get("deleteNotificationRocketChat", "")

    # Notification operations - Email

    def add_notification_email(self, name: str, email_address: str) -> Dict[str, Any]:
        """
        Add an Email notification.

        Args:
            name: Notification name
            email_address: Email address to send notifications to

        Returns:
            Notification data
        """
        mutation = """
        mutation AddNotificationEmail($input: AddNotificationEmailInput!) {
            addNotificationEmail(input: $input) {
                id
                name
                emailAddress
            }
        }
        """

        input_data = {"name": name, "emailAddress": email_address}

        result = self._execute(mutation, {"input": input_data})
        return result.get("addNotificationEmail", {})

    def get_all_notification_email(self) -> list:
        """
        Get all Email notifications.

        Returns:
            List of Email notification data
        """
        all_notifications = self._get_all_notifications()
        return [n for n in all_notifications if n.get("__typename") == "NotificationEmail"]

    def get_notification_email_by_name(self, name: str) -> Optional[Dict[str, Any]]:
        """
        Get Email notification by name.

        Args:
            name: Notification name

        Returns:
            Notification data or None if not found
        """
        all_notifications = self.get_all_notification_email()

        for notification in all_notifications:
            if notification.get("name") == name:
                return notification

        return None

    def update_notification_email(self, name: str, **kwargs) -> Dict[str, Any]:
        """
        Update an Email notification.

        Args:
            name: Notification name (used to identify the notification)
            **kwargs: Properties to update (emailAddress)

        Returns:
            Updated notification data
        """
        mutation = """
        mutation UpdateNotificationEmail($input: UpdateNotificationEmailInput!) {
            updateNotificationEmail(input: $input) {
                id
                name
                emailAddress
            }
        }
        """

        # Lagoon API uses patch format: {name: "...", patch: {field: value}}
        input_data = {"name": name, "patch": kwargs}

        result = self._execute(mutation, {"input": input_data})
        return result.get("updateNotificationEmail", {})

    def delete_notification_email(self, name: str) -> str:
        """
        Delete an Email notification.

        Args:
            name: Notification name

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteNotificationEmail($input: DeleteNotificationEmailInput!) {
            deleteNotificationEmail(input: $input)
        }
        """

        input_data = {"name": name}

        result = self._execute(mutation, {"input": input_data})
        return result.get("deleteNotificationEmail", "")

    # Notification operations - Microsoft Teams

    def add_notification_microsoftteams(self, name: str, webhook: str) -> Dict[str, Any]:
        """
        Add a Microsoft Teams notification.

        Args:
            name: Notification name
            webhook: Microsoft Teams webhook URL

        Returns:
            Notification data
        """
        mutation = """
        mutation AddNotificationMicrosoftTeams($input: AddNotificationMicrosoftTeamsInput!) {
            addNotificationMicrosoftTeams(input: $input) {
                id
                name
                webhook
            }
        }
        """

        input_data = {"name": name, "webhook": webhook}

        result = self._execute(mutation, {"input": input_data})
        return result.get("addNotificationMicrosoftTeams", {})

    def get_all_notification_microsoftteams(self) -> list:
        """
        Get all Microsoft Teams notifications.

        Returns:
            List of Microsoft Teams notification data
        """
        all_notifications = self._get_all_notifications()
        return [n for n in all_notifications if n.get("__typename") == "NotificationMicrosoftTeams"]

    def get_notification_microsoftteams_by_name(self, name: str) -> Optional[Dict[str, Any]]:
        """
        Get Microsoft Teams notification by name.

        Args:
            name: Notification name

        Returns:
            Notification data or None if not found
        """
        all_notifications = self.get_all_notification_microsoftteams()

        for notification in all_notifications:
            if notification.get("name") == name:
                return notification

        return None

    def update_notification_microsoftteams(self, name: str, **kwargs) -> Dict[str, Any]:
        """
        Update a Microsoft Teams notification.

        Args:
            name: Notification name (used to identify the notification)
            **kwargs: Properties to update (webhook)

        Returns:
            Updated notification data
        """
        mutation = """
        mutation UpdateNotificationMicrosoftTeams($input: UpdateNotificationMicrosoftTeamsInput!) {
            updateNotificationMicrosoftTeams(input: $input) {
                id
                name
                webhook
            }
        }
        """

        # Lagoon API uses patch format: {name: "...", patch: {field: value}}
        input_data = {"name": name, "patch": kwargs}

        result = self._execute(mutation, {"input": input_data})
        return result.get("updateNotificationMicrosoftTeams", {})

    def delete_notification_microsoftteams(self, name: str) -> str:
        """
        Delete a Microsoft Teams notification.

        Args:
            name: Notification name

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteNotificationMicrosoftTeams($input: DeleteNotificationMicrosoftTeamsInput!) {
            deleteNotificationMicrosoftTeams(input: $input)
        }
        """

        input_data = {"name": name}

        result = self._execute(mutation, {"input": input_data})
        return result.get("deleteNotificationMicrosoftTeams", "")

    # Project notification association operations

    def add_notification_to_project(
        self, project: str, notification_type: str, notification_name: str
    ) -> Dict[str, Any]:
        """
        Add a notification to a project.

        Args:
            project: Project name
            notification_type: Type of notification (slack, rocketchat, email, microsoftteams)
            notification_name: Name of the notification to add

        Returns:
            Project data with notifications
        """
        mutation = """
        mutation AddNotificationToProject($input: AddNotificationToProjectInput!) {
            addNotificationToProject(input: $input) {
                id
                name
            }
        }
        """

        input_data = {
            "project": project,
            "notificationType": notification_type.upper(),
            "notificationName": notification_name,
        }

        result = self._execute(mutation, {"input": input_data})
        return result.get("addNotificationToProject", {})

    def remove_notification_from_project(
        self, project: str, notification_type: str, notification_name: str
    ) -> Dict[str, Any]:
        """
        Remove a notification from a project.

        Args:
            project: Project name
            notification_type: Type of notification (slack, rocketchat, email, microsoftteams)
            notification_name: Name of the notification to remove

        Returns:
            Project data
        """
        mutation = """
        mutation RemoveNotificationFromProject($input: RemoveNotificationFromProjectInput!) {
            removeNotificationFromProject(input: $input) {
                id
                name
            }
        }
        """

        input_data = {
            "project": project,
            "notificationType": notification_type.upper(),
            "notificationName": notification_name,
        }

        result = self._execute(mutation, {"input": input_data})
        return result.get("removeNotificationFromProject", {})

    def get_project_notifications(self, project_name: str) -> Dict[str, Any]:
        """
        Get all notifications linked to a project.

        Args:
            project_name: Project name

        Returns:
            Dict with notification lists by type
        """
        query = """
        query ProjectByName($name: String!) {
            projectByName(name: $name) {
                id
                name
                notifications {
                    ... on NotificationSlack {
                        __typename
                        id
                        name
                        webhook
                        channel
                    }
                    ... on NotificationRocketChat {
                        __typename
                        id
                        name
                        webhook
                        channel
                    }
                    ... on NotificationEmail {
                        __typename
                        id
                        name
                        emailAddress
                    }
                    ... on NotificationMicrosoftTeams {
                        __typename
                        id
                        name
                        webhook
                    }
                }
            }
        }
        """

        result = self._execute(query, {"name": project_name})
        project = result.get("projectByName")

        if not project:
            return {}

        # Organize notifications by type
        notifications = {
            "slack": [],
            "rocketchat": [],
            "email": [],
            "microsoftteams": [],
        }

        for notification in project.get("notifications", []):
            typename = notification.get("__typename", "")
            if typename == "NotificationSlack":
                notifications["slack"].append(notification)
            elif typename == "NotificationRocketChat":
                notifications["rocketchat"].append(notification)
            elif typename == "NotificationEmail":
                notifications["email"].append(notification)
            elif typename == "NotificationMicrosoftTeams":
                notifications["microsoftteams"].append(notification)

        return notifications

    def check_project_notification_exists(
        self, project_name: str, notification_type: str, notification_name: str
    ) -> bool:
        """
        Check if a specific notification is linked to a project.

        Args:
            project_name: Project name
            notification_type: Type of notification (slack, rocketchat, email, microsoftteams)
            notification_name: Name of the notification

        Returns:
            True if the notification is linked to the project
        """
        notifications = self.get_project_notifications(project_name)
        type_notifications = notifications.get(notification_type.lower(), [])

        for notification in type_notifications:
            if notification.get("name") == notification_name:
                return True

        return False

    # Advanced Task Definition operations
    def add_advanced_task_definition(
        self,
        name: str,
        task_type: str,
        service: str,
        command: Optional[str] = None,
        image: Optional[str] = None,
        project_id: Optional[int] = None,
        environment_id: Optional[int] = None,
        group_name: Optional[str] = None,
        system_wide: Optional[bool] = None,
        description: Optional[str] = None,
        permission: Optional[str] = None,
        confirmation_text: Optional[str] = None,
        arguments: Optional[list] = None,
    ) -> Dict[str, Any]:
        """
        Add an advanced task definition.

        Args:
            name: Task definition name
            task_type: Task type ("COMMAND" or "IMAGE")
            service: Service container name to run in
            command: Command to execute (required for COMMAND type)
            image: Container image (required for IMAGE type)
            project_id: Project ID (for project-scoped tasks)
            environment_id: Environment ID (for environment-scoped tasks)
            group_name: Group name (for group-scoped tasks)
            system_wide: If True, task is available system-wide (platform admin only)
            description: Task description
            permission: Permission level ("GUEST", "DEVELOPER", "MAINTAINER")
            confirmation_text: Text to display for user confirmation
            arguments: List of argument definitions [{name, displayName, type}]

        Returns:
            Task definition data
        """
        # Use inline fragments for union type response
        mutation = """
        mutation AddAdvancedTaskDefinition($input: AddAdvancedTaskDefinitionInput!) {
            addAdvancedTaskDefinition(input: $input) {
                ... on AdvancedTaskDefinitionCommand {
                    id
                    name
                    description
                    type
                    service
                    command
                    permission
                    confirmationText
                    advancedTaskDefinitionArguments {
                        id
                        name
                        displayName
                        type
                    }
                    project {
                        id
                        name
                    }
                    environment {
                        id
                        name
                    }
                    groupName
                    created
                }
                ... on AdvancedTaskDefinitionImage {
                    id
                    name
                    description
                    type
                    service
                    image
                    permission
                    confirmationText
                    advancedTaskDefinitionArguments {
                        id
                        name
                        displayName
                        type
                    }
                    project {
                        id
                        name
                    }
                    environment {
                        id
                        name
                    }
                    groupName
                    created
                }
            }
        }
        """

        input_data: Dict[str, Any] = {
            "name": name,
            "type": task_type.upper(),
            "service": service,
        }

        # Add command or image based on type
        if command:
            input_data["command"] = command
        if image:
            input_data["image"] = image

        # Add scope (exactly one should be set)
        if project_id is not None:
            input_data["project"] = project_id
        if environment_id is not None:
            input_data["environment"] = environment_id
        if group_name is not None:
            input_data["groupName"] = group_name
        if system_wide is True:
            input_data["systemWide"] = True

        # Add optional fields
        if description:
            input_data["description"] = description
        if permission:
            input_data["permission"] = permission.upper()
        if confirmation_text:
            input_data["confirmationText"] = confirmation_text
        if arguments:
            # Convert to API format
            input_data["advancedTaskDefinitionArguments"] = [
                {
                    "name": arg.get("name"),
                    "displayName": arg.get("display_name") or arg.get("displayName"),
                    "type": arg.get("type", "STRING").upper(),
                }
                for arg in arguments
            ]

        result = self._execute(mutation, {"input": input_data})
        task = result.get("addAdvancedTaskDefinition", {})

        # Normalize nested objects
        if task.get("project") and isinstance(task["project"], dict):
            task["projectId"] = task["project"].get("id")
        if task.get("environment") and isinstance(task["environment"], dict):
            task["environmentId"] = task["environment"].get("id")

        return task

    def get_advanced_task_definition_by_id(self, task_id: int) -> Optional[Dict[str, Any]]:
        """
        Get an advanced task definition by ID.

        Args:
            task_id: Task definition ID

        Returns:
            Task definition data or None if not found
        """
        query = """
        query AdvancedTaskDefinitionById($id: Int!) {
            advancedTaskDefinitionById(id: $id) {
                ... on AdvancedTaskDefinitionCommand {
                    id
                    name
                    description
                    type
                    service
                    command
                    permission
                    confirmationText
                    advancedTaskDefinitionArguments {
                        id
                        name
                        displayName
                        type
                    }
                    project {
                        id
                        name
                    }
                    environment {
                        id
                        name
                    }
                    groupName
                    created
                }
                ... on AdvancedTaskDefinitionImage {
                    id
                    name
                    description
                    type
                    service
                    image
                    permission
                    confirmationText
                    advancedTaskDefinitionArguments {
                        id
                        name
                        displayName
                        type
                    }
                    project {
                        id
                        name
                    }
                    environment {
                        id
                        name
                    }
                    groupName
                    created
                }
            }
        }
        """

        result = self._execute(query, {"id": task_id})
        task = result.get("advancedTaskDefinitionById")

        if task:
            # Normalize nested objects
            if task.get("project") and isinstance(task["project"], dict):
                task["projectId"] = task["project"].get("id")
            if task.get("environment") and isinstance(task["environment"], dict):
                task["environmentId"] = task["environment"].get("id")

        return task

    def delete_advanced_task_definition(self, task_id: int) -> str:
        """
        Delete an advanced task definition.

        Args:
            task_id: Task definition ID

        Returns:
            Success message
        """
        mutation = """
        mutation DeleteAdvancedTaskDefinition($id: Int!) {
            deleteAdvancedTaskDefinition(id: $id)
        }
        """

        result = self._execute(mutation, {"id": task_id})
        return result.get("deleteAdvancedTaskDefinition", "")

    def get_advanced_tasks_by_environment(self, environment_id: int) -> list:
        """
        Get all advanced task definitions available for an environment.

        Args:
            environment_id: Environment ID

        Returns:
            List of task definition data

        Note:
            Uses advancedTasksForEnvironment for Lagoon v2.30.0+
            with fallback to advancedTasksByEnvironment for older versions.
        """
        # Try newer API first (Lagoon v2.30.0+)
        # Note: In v2.30.0+, project and environment are Int IDs, not objects
        query_new = """
        query AdvancedTasksForEnvironment($environment: Int!) {
            advancedTasksForEnvironment(environment: $environment) {
                ... on AdvancedTaskDefinitionCommand {
                    id
                    name
                    description
                    type
                    service
                    command
                    permission
                    project
                    environment
                    groupName
                }
                ... on AdvancedTaskDefinitionImage {
                    id
                    name
                    description
                    type
                    service
                    image
                    permission
                    project
                    environment
                    groupName
                }
            }
        }
        """

        try:
            result = self._execute(query_new, {"environment": environment_id})
            tasks = result.get("advancedTasksForEnvironment", [])
        except (LagoonAPIError, LagoonConnectionError) as e:
            # Fallback to older API for Lagoon versions < 2.30.0
            if "Cannot query field" in str(e) or "400" in str(e):
                query_old = """
                query AdvancedTasksByEnvironment($environment: Int!) {
                    advancedTasksByEnvironment(environment: $environment) {
                        ... on AdvancedTaskDefinitionCommand {
                            id
                            name
                            description
                            type
                            service
                            command
                            permission
                            project {
                                id
                                name
                            }
                            environment {
                                id
                                name
                            }
                            groupName
                        }
                        ... on AdvancedTaskDefinitionImage {
                            id
                            name
                            description
                            type
                            service
                            image
                            permission
                            project {
                                id
                                name
                            }
                            environment {
                                id
                                name
                            }
                            groupName
                        }
                    }
                }
                """
                result = self._execute(query_old, {"environment": environment_id})
                tasks = result.get("advancedTasksByEnvironment", [])
            else:
                raise

        # Normalize project/environment fields for consistency
        # Old API returns objects: {"id": 1, "name": "..."}
        # New API returns Int IDs directly: 1
        for task in tasks:
            project_val = task.get("project")
            env_val = task.get("environment")

            if isinstance(project_val, dict):
                # Old API format
                task["projectId"] = project_val.get("id")
            elif isinstance(project_val, int):
                # New API format
                task["projectId"] = project_val

            if isinstance(env_val, dict):
                # Old API format
                task["environmentId"] = env_val.get("id")
            elif isinstance(env_val, int):
                # New API format
                task["environmentId"] = env_val

        return tasks
