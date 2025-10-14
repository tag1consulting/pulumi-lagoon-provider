"""GraphQL client for Lagoon API."""

from typing import Dict, Any, Optional, List
import requests
import json


class LagoonAPIError(Exception):
    """Exception raised for Lagoon API errors."""
    pass


class LagoonConnectionError(Exception):
    """Exception raised for connection errors."""
    pass


class LagoonClient:
    """GraphQL API client for Lagoon."""

    def __init__(self, api_url: str, token: str):
        """
        Initialize Lagoon API client.

        Args:
            api_url: Lagoon GraphQL API endpoint URL
            token: Authentication token (JWT)
        """
        self.api_url = api_url
        self.token = token
        self.session = requests.Session()
        self.session.headers.update({
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json",
        })

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
        payload = {
            "query": query,
        }

        if variables:
            payload["variables"] = variables

        try:
            response = self.session.post(
                self.api_url,
                json=payload,
                timeout=30
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
    def create_project(
        self,
        name: str,
        git_url: str,
        openshift: int,
        **kwargs
    ) -> Dict[str, Any]:
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
                openshift
                productionEnvironment
                branches
                pullrequests
                created
            }
        }
        """

        input_data = {
            "name": name,
            "gitUrl": git_url,
            "openshift": openshift,
            **kwargs
        }

        result = self._execute(mutation, {"input": input_data})
        return result.get("addProject", {})

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
                openshift
                productionEnvironment
                branches
                pullrequests
                created
            }
        }
        """

        result = self._execute(query, {"name": name})
        return result.get("projectByName")

    def update_project(
        self,
        project_id: int,
        **kwargs
    ) -> Dict[str, Any]:
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
                openshift
                productionEnvironment
                branches
                pullrequests
            }
        }
        """

        input_data = {
            "id": project_id,
            **kwargs
        }

        result = self._execute(mutation, {"input": input_data})
        return result.get("updateProject", {})

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

    # Environment operations (stubs - to be implemented)
    def add_or_update_environment(self, **kwargs) -> Dict[str, Any]:
        """Add or update an environment."""
        # TODO: Implement
        raise NotImplementedError("Environment operations not yet implemented")

    def delete_environment(self, name: str, project: str) -> str:
        """Delete an environment."""
        # TODO: Implement
        raise NotImplementedError("Environment operations not yet implemented")

    # Variable operations (stubs - to be implemented)
    def add_env_variable(self, **kwargs) -> Dict[str, Any]:
        """Add an environment variable."""
        # TODO: Implement
        raise NotImplementedError("Variable operations not yet implemented")

    def delete_env_variable(self, name: str, project: str, environment: Optional[str] = None) -> str:
        """Delete an environment variable."""
        # TODO: Implement
        raise NotImplementedError("Variable operations not yet implemented")
