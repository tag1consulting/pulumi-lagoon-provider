"""Unit tests for Lagoon GraphQL client."""

import pytest
from unittest.mock import Mock, patch, MagicMock
import requests

from pulumi_lagoon.client import LagoonClient, LagoonAPIError, LagoonConnectionError


class TestLagoonClientInit:
    """Tests for LagoonClient initialization."""

    def test_client_initialization(self):
        """Test client initializes with correct credentials."""
        with patch("requests.Session"):
            client = LagoonClient(
                api_url="https://api.test.com/graphql", token="test-token"
            )
            assert client.api_url == "https://api.test.com/graphql"
            assert client.token == "test-token"
            assert client.verify_ssl is True

    def test_client_ssl_verification_disabled_by_env(self):
        """Test SSL verification can be disabled via environment variable."""
        with patch("requests.Session"):
            with patch.dict("os.environ", {"LAGOON_INSECURE": "true"}):
                client = LagoonClient(
                    api_url="https://api.test.com/graphql", token="test-token"
                )
                assert client.verify_ssl is False

    def test_client_ssl_verification_explicit(self):
        """Test SSL verification can be set explicitly."""
        with patch("requests.Session"):
            client = LagoonClient(
                api_url="https://api.test.com/graphql",
                token="test-token",
                verify_ssl=False,
            )
            assert client.verify_ssl is False

    def test_client_sets_auth_headers(self):
        """Test client sets proper authorization headers."""
        with patch("requests.Session") as mock_session_class:
            mock_session = Mock()
            mock_headers = MagicMock()
            mock_session.headers = mock_headers
            mock_session_class.return_value = mock_session

            LagoonClient(api_url="https://api.test.com/graphql", token="test-token")

            mock_headers.update.assert_called_once()
            call_args = mock_headers.update.call_args[0][0]
            assert call_args["Authorization"] == "Bearer test-token"
            assert call_args["Content-Type"] == "application/json"


class TestLagoonClientExecute:
    """Tests for the _execute method."""

    def test_execute_success(self, lagoon_client, mock_response):
        """Test successful GraphQL execution."""
        response = mock_response(data={"test": "value"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client._execute("query { test }")

        assert result == {"test": "value"}
        lagoon_client.session.post.assert_called_once()

    def test_execute_with_variables(self, lagoon_client, mock_response):
        """Test GraphQL execution with variables."""
        response = mock_response(data={"test": "value"})
        lagoon_client.session.post.return_value = response

        lagoon_client._execute(
            "query Test($id: Int!) { test(id: $id) }", variables={"id": 42}
        )

        call_kwargs = lagoon_client.session.post.call_args[1]
        assert call_kwargs["json"]["variables"] == {"id": 42}

    def test_execute_graphql_error(self, lagoon_client, mock_response):
        """Test GraphQL error handling."""
        response = mock_response(errors=[{"message": "Project not found"}])
        lagoon_client.session.post.return_value = response

        with pytest.raises(LagoonAPIError, match="Project not found"):
            lagoon_client._execute("query { test }")

    def test_execute_http_error(self, lagoon_client):
        """Test HTTP error handling."""
        lagoon_client.session.post.side_effect = requests.HTTPError("401 Unauthorized")

        with pytest.raises(LagoonConnectionError, match="HTTP error"):
            lagoon_client._execute("query { test }")

    def test_execute_connection_error(self, lagoon_client):
        """Test connection error handling."""
        lagoon_client.session.post.side_effect = requests.ConnectionError(
            "Connection refused"
        )

        with pytest.raises(LagoonConnectionError, match="Connection error"):
            lagoon_client._execute("query { test }")

    def test_execute_invalid_json(self, lagoon_client):
        """Test invalid JSON response handling."""
        import json

        response = Mock()
        response.raise_for_status = Mock()
        response.json.side_effect = json.JSONDecodeError("Invalid JSON", "", 0)
        lagoon_client.session.post.return_value = response

        with pytest.raises(LagoonAPIError, match="Invalid JSON response"):
            lagoon_client._execute("query { test }")


class TestProjectOperations:
    """Tests for project CRUD operations."""

    def test_create_project_success(self, lagoon_client, mock_response, sample_project):
        """Test successful project creation."""
        response = mock_response(data={"addProject": sample_project})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.create_project(
            name="test-project",
            git_url="git@github.com:test/test-repo.git",
            openshift=1,
            productionEnvironment="main",
            branches="^(main|develop)$",
        )

        assert result["id"] == 1
        assert result["name"] == "test-project"

    def test_create_project_normalizes_openshift(self, lagoon_client, mock_response):
        """Test that openshift dict is normalized to ID."""
        project_data = {
            "id": 1,
            "name": "test-project",
            "gitUrl": "git@github.com:test/repo.git",
            "openshift": {"id": 5, "name": "k8s-cluster"},
            "productionEnvironment": "main",
        }
        response = mock_response(data={"addProject": project_data})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.create_project(
            name="test-project", git_url="git@github.com:test/repo.git", openshift=5
        )

        assert result["openshift"] == 5

    def test_get_project_by_name(self, lagoon_client, mock_response, sample_project):
        """Test getting project by name."""
        response = mock_response(data={"projectByName": sample_project})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_project_by_name("test-project")

        assert result["name"] == "test-project"
        assert result["id"] == 1

    def test_get_project_by_name_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent project returns None."""
        response = mock_response(data={"projectByName": None})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_project_by_name("nonexistent")

        assert result is None

    def test_get_project_by_id(self, lagoon_client, mock_response, sample_project):
        """Test getting project by ID."""
        response = mock_response(data={"projectById": sample_project})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_project_by_id(1)

        assert result["id"] == 1

    def test_update_project(self, lagoon_client, mock_response, sample_project):
        """Test updating a project."""
        updated = sample_project.copy()
        updated["branches"] = "^(main|develop|staging)$"
        response = mock_response(data={"updateProject": updated})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.update_project(
            project_id=1, branches="^(main|develop|staging)$"
        )

        assert result["branches"] == "^(main|develop|staging)$"

    def test_delete_project(self, lagoon_client, mock_response):
        """Test deleting a project."""
        response = mock_response(data={"deleteProject": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_project("test-project")

        assert result == "success"


class TestEnvironmentOperations:
    """Tests for environment CRUD operations."""

    def test_add_environment(self, lagoon_client, mock_response, sample_environment):
        """Test adding an environment."""
        response = mock_response(data={"addOrUpdateEnvironment": sample_environment})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_or_update_environment(
            name="main", project=1, deploy_type="branch", environment_type="production"
        )

        assert result["id"] == 1
        assert result["name"] == "main"

    def test_add_environment_uppercase_enums(
        self, lagoon_client, mock_response, sample_environment
    ):
        """Test that deploy_type and environment_type are uppercased."""
        response = mock_response(data={"addOrUpdateEnvironment": sample_environment})
        lagoon_client.session.post.return_value = response

        lagoon_client.add_or_update_environment(
            name="main", project=1, deploy_type="branch", environment_type="production"
        )

        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["deployType"] == "BRANCH"
        assert input_data["environmentType"] == "PRODUCTION"

    def test_get_environment_by_name(
        self, lagoon_client, mock_response, sample_environment
    ):
        """Test getting environment by name."""
        response = mock_response(data={"environmentByName": sample_environment})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_environment_by_name("main", project_id=1)

        assert result["name"] == "main"

    def test_get_environment_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent environment returns None."""
        response = mock_response(data={"environmentByName": None})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_environment_by_name("nonexistent", project_id=1)

        assert result is None

    def test_delete_environment(self, lagoon_client, mock_response):
        """Test deleting an environment."""
        response = mock_response(data={"deleteEnvironment": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_environment(
            name="develop", project=1, execute=True
        )

        assert result == "success"


class TestVariableOperations:
    """Tests for variable CRUD operations."""

    def test_add_project_variable(self, lagoon_client, mock_response, sample_variable):
        """Test adding a project-level variable."""
        response = mock_response(data={"addEnvVariable": sample_variable})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_env_variable(
            name="DATABASE_HOST", value="mysql.example.com", project=1, scope="runtime"
        )

        assert result["name"] == "DATABASE_HOST"

        # Verify PROJECT type was used
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["type"] == "PROJECT"
        assert input_data["typeId"] == 1

    def test_add_environment_variable(
        self, lagoon_client, mock_response, sample_variable
    ):
        """Test adding an environment-level variable."""
        env_var = sample_variable.copy()
        env_var["environment"] = {"id": 1, "name": "main"}
        response = mock_response(data={"addEnvVariable": env_var})
        lagoon_client.session.post.return_value = response

        lagoon_client.add_env_variable(
            name="DATABASE_HOST",
            value="mysql.example.com",
            project=1,
            scope="runtime",
            environment=1,
        )

        # Verify ENVIRONMENT type was used
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["type"] == "ENVIRONMENT"
        assert input_data["typeId"] == 1

    def test_add_variable_uppercase_scope(
        self, lagoon_client, mock_response, sample_variable
    ):
        """Test that scope is uppercased."""
        response = mock_response(data={"addEnvVariable": sample_variable})
        lagoon_client.session.post.return_value = response

        lagoon_client.add_env_variable(
            name="TEST", value="value", project=1, scope="build"
        )

        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["scope"] == "BUILD"

    def test_get_variable_by_name(self, lagoon_client, mock_response, sample_variable):
        """Test getting variable by name."""
        response = mock_response(
            data={"envVariablesByProjectEnvironment": [sample_variable]}
        )
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_env_variable_by_name(name="DATABASE_HOST", project=1)

        assert result["name"] == "DATABASE_HOST"

    def test_get_variable_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent variable returns None."""
        response = mock_response(data={"envVariablesByProjectEnvironment": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_env_variable_by_name(name="NONEXISTENT", project=1)

        assert result is None

    def test_get_variable_with_environment(
        self, lagoon_client, mock_response, sample_variable
    ):
        """Test getting environment-scoped variable."""
        response = mock_response(
            data={"envVariablesByProjectEnvironment": [sample_variable]}
        )
        lagoon_client.session.post.return_value = response

        lagoon_client.get_env_variable_by_name(
            name="DATABASE_HOST", project=1, environment=1
        )

        call_kwargs = lagoon_client.session.post.call_args[1]
        variables = call_kwargs["json"]["variables"]
        assert variables["environment"] == 1

    def test_delete_variable(self, lagoon_client, mock_response):
        """Test deleting a variable."""
        response = mock_response(data={"deleteEnvVariable": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_env_variable(name="DATABASE_HOST", project=1)

        assert result == "success"

    def test_delete_environment_variable(self, lagoon_client, mock_response):
        """Test deleting an environment-scoped variable."""
        response = mock_response(data={"deleteEnvVariable": "success"})
        lagoon_client.session.post.return_value = response

        lagoon_client.delete_env_variable(
            name="DATABASE_HOST", project=1, environment=1
        )

        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["environment"] == 1
