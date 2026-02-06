"""Unit tests for Lagoon GraphQL client."""

from unittest.mock import MagicMock, Mock, patch

import pytest
import requests

from pulumi_lagoon.client import LagoonAPIError, LagoonClient, LagoonConnectionError


class TestLagoonClientInit:
    """Tests for LagoonClient initialization."""

    def test_client_initialization(self):
        """Test client initializes with correct credentials."""
        with patch("requests.Session"):
            client = LagoonClient(api_url="https://api.test.com/graphql", token="test-token")
            assert client.api_url == "https://api.test.com/graphql"
            assert client.token == "test-token"
            assert client.verify_ssl is True

    def test_client_ssl_verification_disabled_by_env(self):
        """Test SSL verification can be disabled via environment variable."""
        with patch("requests.Session"):
            with patch.dict("os.environ", {"LAGOON_INSECURE": "true"}):
                client = LagoonClient(api_url="https://api.test.com/graphql", token="test-token")
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

        lagoon_client._execute("query Test($id: Int!) { test(id: $id) }", variables={"id": 42})

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
        lagoon_client.session.post.side_effect = requests.ConnectionError("Connection refused")

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
        """Test getting project by ID.

        Note: The implementation now uses allProjects and filters by ID
        for compatibility with Lagoon v2.30.0+ which doesn't have projectById.
        """
        response = mock_response(data={"allProjects": [sample_project]})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_project_by_id(1)

        assert result["id"] == 1

    def test_update_project(self, lagoon_client, mock_response, sample_project):
        """Test updating a project."""
        updated = sample_project.copy()
        updated["branches"] = "^(main|develop|staging)$"
        response = mock_response(data={"updateProject": updated})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.update_project(project_id=1, branches="^(main|develop|staging)$")

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

    def test_get_environment_by_name(self, lagoon_client, mock_response, sample_environment):
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

        result = lagoon_client.delete_environment(name="develop", project=1, execute=True)

        assert result == "success"


class TestVariableOperations:
    """Tests for variable CRUD operations.

    Note: Variable methods now call get_project_by_id internally to convert
    project IDs to names for Lagoon v2.30.0+ compatibility. Tests need to
    mock both the allProjects query and the actual mutation.
    """

    def test_add_project_variable(
        self, lagoon_client, mock_response, sample_variable, sample_project
    ):
        """Test adding a project-level variable."""
        # Mock responses for: 1) allProjects (to get project name), 2) mutation
        project_response = mock_response(data={"allProjects": [sample_project]})
        mutation_response = mock_response(data={"addOrUpdateEnvVariableByName": sample_variable})
        lagoon_client.session.post.side_effect = [project_response, mutation_response]

        result = lagoon_client.add_env_variable(
            name="DATABASE_HOST", value="mysql.example.com", project=1, scope="runtime"
        )

        assert result["name"] == "DATABASE_HOST"

    def test_add_environment_variable(
        self, lagoon_client, mock_response, sample_variable, sample_project, sample_environment
    ):
        """Test adding an environment-level variable."""
        env_var = sample_variable.copy()
        env_var["environment"] = {"id": 1, "name": "main"}
        # Mock responses for: 1) allProjects, 2) environmentById, 3) mutation
        project_response = mock_response(data={"allProjects": [sample_project]})
        env_response = mock_response(data={"environmentById": sample_environment})
        mutation_response = mock_response(data={"addOrUpdateEnvVariableByName": env_var})
        lagoon_client.session.post.side_effect = [project_response, env_response, mutation_response]

        lagoon_client.add_env_variable(
            name="DATABASE_HOST",
            value="mysql.example.com",
            project=1,
            scope="runtime",
            environment=1,
        )

        # Verify mutation was called (3rd call)
        assert lagoon_client.session.post.call_count == 3

    def test_add_variable_uppercase_scope(
        self, lagoon_client, mock_response, sample_variable, sample_project
    ):
        """Test that scope is uppercased."""
        project_response = mock_response(data={"allProjects": [sample_project]})
        mutation_response = mock_response(data={"addOrUpdateEnvVariableByName": sample_variable})
        lagoon_client.session.post.side_effect = [project_response, mutation_response]

        lagoon_client.add_env_variable(name="TEST", value="value", project=1, scope="build")

        # Check the mutation call (2nd call)
        call_kwargs = lagoon_client.session.post.call_args_list[1][1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["scope"] == "BUILD"

    def test_get_variable_by_name(
        self, lagoon_client, mock_response, sample_variable, sample_project
    ):
        """Test getting variable by name."""
        # Mock responses for: 1) allProjects (to get project name), 2) variable query
        project_response = mock_response(data={"allProjects": [sample_project]})
        var_response = mock_response(
            data={"getEnvVariablesByProjectEnvironmentName": [sample_variable]}
        )
        lagoon_client.session.post.side_effect = [project_response, var_response]

        result = lagoon_client.get_env_variable_by_name(name="DATABASE_HOST", project=1)

        assert result["name"] == "DATABASE_HOST"

    def test_get_variable_not_found(self, lagoon_client, mock_response, sample_project):
        """Test getting nonexistent variable returns None."""
        project_response = mock_response(data={"allProjects": [sample_project]})
        var_response = mock_response(data={"getEnvVariablesByProjectEnvironmentName": []})
        lagoon_client.session.post.side_effect = [project_response, var_response]

        result = lagoon_client.get_env_variable_by_name(name="NONEXISTENT", project=1)

        assert result is None

    def test_get_variable_with_environment(
        self, lagoon_client, mock_response, sample_variable, sample_project, sample_environment
    ):
        """Test getting environment-scoped variable."""
        # Mock responses for: 1) allProjects, 2) environmentById, 3) variable query
        project_response = mock_response(data={"allProjects": [sample_project]})
        env_response = mock_response(data={"environmentById": sample_environment})
        var_response = mock_response(
            data={"getEnvVariablesByProjectEnvironmentName": [sample_variable]}
        )
        lagoon_client.session.post.side_effect = [project_response, env_response, var_response]

        lagoon_client.get_env_variable_by_name(name="DATABASE_HOST", project=1, environment=1)

        # Verify variable query includes environment name (from 3rd call)
        call_kwargs = lagoon_client.session.post.call_args_list[2][1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["environment"] == "main"  # Environment name, not ID

    def test_delete_variable(self, lagoon_client, mock_response, sample_project):
        """Test deleting a variable."""
        project_response = mock_response(data={"allProjects": [sample_project]})
        delete_response = mock_response(data={"deleteEnvVariableByName": "success"})
        lagoon_client.session.post.side_effect = [project_response, delete_response]

        result = lagoon_client.delete_env_variable(name="DATABASE_HOST", project=1)

        assert result == "success"

    def test_delete_environment_variable(
        self, lagoon_client, mock_response, sample_project, sample_environment
    ):
        """Test deleting an environment-scoped variable."""
        project_response = mock_response(data={"allProjects": [sample_project]})
        env_response = mock_response(data={"environmentById": sample_environment})
        delete_response = mock_response(data={"deleteEnvVariableByName": "success"})
        lagoon_client.session.post.side_effect = [project_response, env_response, delete_response]

        lagoon_client.delete_env_variable(name="DATABASE_HOST", project=1, environment=1)

        # Verify delete was called with environment name
        call_kwargs = lagoon_client.session.post.call_args_list[2][1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["environment"] == "main"


class TestNotificationSlackOperations:
    """Tests for Slack notification CRUD operations."""

    def test_add_notification_slack(self, lagoon_client, mock_response, sample_notification_slack):
        """Test adding a Slack notification."""
        response = mock_response(data={"addNotificationSlack": sample_notification_slack})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_notification_slack(
            name="deploy-alerts",
            webhook="https://hooks.slack.com/services/xxx/yyy/zzz",
            channel="#deployments",
        )

        assert result["id"] == 1
        assert result["name"] == "deploy-alerts"
        assert result["webhook"] == "https://hooks.slack.com/services/xxx/yyy/zzz"
        assert result["channel"] == "#deployments"

        # Verify the GraphQL mutation was called with correct input
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["name"] == "deploy-alerts"
        assert input_data["webhook"] == "https://hooks.slack.com/services/xxx/yyy/zzz"
        assert input_data["channel"] == "#deployments"

    def test_get_all_notification_slack(
        self, lagoon_client, mock_response, sample_notification_slack
    ):
        """Test getting all Slack notifications."""
        all_notifications = [
            {**sample_notification_slack, "__typename": "NotificationSlack"},
            {
                "__typename": "NotificationEmail",
                "id": 2,
                "name": "email-alert",
                "emailAddress": "test@example.com",
            },
        ]
        response = mock_response(data={"allNotifications": all_notifications})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_all_notification_slack()

        assert len(result) == 1
        assert result[0]["name"] == "deploy-alerts"
        assert result[0]["__typename"] == "NotificationSlack"

    def test_get_notification_slack_by_name(
        self, lagoon_client, mock_response, sample_notification_slack
    ):
        """Test getting Slack notification by name."""
        all_notifications = [
            {**sample_notification_slack, "__typename": "NotificationSlack"},
        ]
        response = mock_response(data={"allNotifications": all_notifications})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_notification_slack_by_name("deploy-alerts")

        assert result is not None
        assert result["name"] == "deploy-alerts"

    def test_get_notification_slack_by_name_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent Slack notification returns None."""
        response = mock_response(data={"allNotifications": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_notification_slack_by_name("nonexistent")

        assert result is None

    def test_update_notification_slack(
        self, lagoon_client, mock_response, sample_notification_slack
    ):
        """Test updating a Slack notification."""
        updated = sample_notification_slack.copy()
        updated["webhook"] = "https://hooks.slack.com/services/new/webhook/url"
        response = mock_response(data={"updateNotificationSlack": updated})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.update_notification_slack(
            name="deploy-alerts",
            webhook="https://hooks.slack.com/services/new/webhook/url",
        )

        assert result["webhook"] == "https://hooks.slack.com/services/new/webhook/url"

        # Verify patch format is used
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["name"] == "deploy-alerts"
        assert input_data["patch"]["webhook"] == "https://hooks.slack.com/services/new/webhook/url"

    def test_delete_notification_slack(self, lagoon_client, mock_response):
        """Test deleting a Slack notification."""
        response = mock_response(data={"deleteNotificationSlack": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_notification_slack(name="deploy-alerts")

        assert result == "success"

        # Verify correct input
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["name"] == "deploy-alerts"


class TestNotificationRocketChatOperations:
    """Tests for RocketChat notification CRUD operations."""

    def test_add_notification_rocketchat(
        self, lagoon_client, mock_response, sample_notification_rocketchat
    ):
        """Test adding a RocketChat notification."""
        response = mock_response(data={"addNotificationRocketChat": sample_notification_rocketchat})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_notification_rocketchat(
            name="team-chat",
            webhook="https://rocketchat.example.com/hooks/xxx/yyy",
            channel="#alerts",
        )

        assert result["id"] == 2
        assert result["name"] == "team-chat"
        assert result["webhook"] == "https://rocketchat.example.com/hooks/xxx/yyy"
        assert result["channel"] == "#alerts"

    def test_get_all_notification_rocketchat(
        self, lagoon_client, mock_response, sample_notification_rocketchat
    ):
        """Test getting all RocketChat notifications."""
        all_notifications = [
            {**sample_notification_rocketchat, "__typename": "NotificationRocketChat"},
            {
                "__typename": "NotificationSlack",
                "id": 1,
                "name": "slack-alert",
                "webhook": "https://slack.com",
                "channel": "#test",
            },
        ]
        response = mock_response(data={"allNotifications": all_notifications})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_all_notification_rocketchat()

        assert len(result) == 1
        assert result[0]["name"] == "team-chat"
        assert result[0]["__typename"] == "NotificationRocketChat"

    def test_get_notification_rocketchat_by_name(
        self, lagoon_client, mock_response, sample_notification_rocketchat
    ):
        """Test getting RocketChat notification by name."""
        all_notifications = [
            {**sample_notification_rocketchat, "__typename": "NotificationRocketChat"},
        ]
        response = mock_response(data={"allNotifications": all_notifications})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_notification_rocketchat_by_name("team-chat")

        assert result is not None
        assert result["name"] == "team-chat"

    def test_get_notification_rocketchat_by_name_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent RocketChat notification returns None."""
        response = mock_response(data={"allNotifications": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_notification_rocketchat_by_name("nonexistent")

        assert result is None

    def test_update_notification_rocketchat(
        self, lagoon_client, mock_response, sample_notification_rocketchat
    ):
        """Test updating a RocketChat notification."""
        updated = sample_notification_rocketchat.copy()
        updated["channel"] = "#new-channel"
        response = mock_response(data={"updateNotificationRocketChat": updated})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.update_notification_rocketchat(
            name="team-chat",
            channel="#new-channel",
        )

        assert result["channel"] == "#new-channel"

        # Verify patch format is used
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["name"] == "team-chat"
        assert input_data["patch"]["channel"] == "#new-channel"

    def test_delete_notification_rocketchat(self, lagoon_client, mock_response):
        """Test deleting a RocketChat notification."""
        response = mock_response(data={"deleteNotificationRocketChat": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_notification_rocketchat(name="team-chat")

        assert result == "success"


class TestNotificationEmailOperations:
    """Tests for Email notification CRUD operations."""

    def test_add_notification_email(self, lagoon_client, mock_response, sample_notification_email):
        """Test adding an Email notification."""
        response = mock_response(data={"addNotificationEmail": sample_notification_email})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_notification_email(
            name="ops-team",
            email_address="ops@example.com",
        )

        assert result["id"] == 3
        assert result["name"] == "ops-team"
        assert result["emailAddress"] == "ops@example.com"

        # Verify the GraphQL mutation uses emailAddress (camelCase)
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["name"] == "ops-team"
        assert input_data["emailAddress"] == "ops@example.com"

    def test_get_all_notification_email(
        self, lagoon_client, mock_response, sample_notification_email
    ):
        """Test getting all Email notifications."""
        all_notifications = [
            {**sample_notification_email, "__typename": "NotificationEmail"},
            {
                "__typename": "NotificationSlack",
                "id": 1,
                "name": "slack-alert",
                "webhook": "https://slack.com",
                "channel": "#test",
            },
        ]
        response = mock_response(data={"allNotifications": all_notifications})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_all_notification_email()

        assert len(result) == 1
        assert result[0]["name"] == "ops-team"
        assert result[0]["__typename"] == "NotificationEmail"

    def test_get_notification_email_by_name(
        self, lagoon_client, mock_response, sample_notification_email
    ):
        """Test getting Email notification by name."""
        all_notifications = [
            {**sample_notification_email, "__typename": "NotificationEmail"},
        ]
        response = mock_response(data={"allNotifications": all_notifications})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_notification_email_by_name("ops-team")

        assert result is not None
        assert result["name"] == "ops-team"

    def test_get_notification_email_by_name_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent Email notification returns None."""
        response = mock_response(data={"allNotifications": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_notification_email_by_name("nonexistent")

        assert result is None

    def test_update_notification_email(
        self, lagoon_client, mock_response, sample_notification_email
    ):
        """Test updating an Email notification."""
        updated = sample_notification_email.copy()
        updated["emailAddress"] = "new-ops@example.com"
        response = mock_response(data={"updateNotificationEmail": updated})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.update_notification_email(
            name="ops-team",
            emailAddress="new-ops@example.com",
        )

        assert result["emailAddress"] == "new-ops@example.com"

        # Verify patch format is used
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["name"] == "ops-team"
        assert input_data["patch"]["emailAddress"] == "new-ops@example.com"

    def test_delete_notification_email(self, lagoon_client, mock_response):
        """Test deleting an Email notification."""
        response = mock_response(data={"deleteNotificationEmail": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_notification_email(name="ops-team")

        assert result == "success"


class TestNotificationMicrosoftTeamsOperations:
    """Tests for Microsoft Teams notification CRUD operations."""

    def test_add_notification_microsoftteams(
        self, lagoon_client, mock_response, sample_notification_microsoftteams
    ):
        """Test adding a Microsoft Teams notification."""
        response = mock_response(
            data={"addNotificationMicrosoftTeams": sample_notification_microsoftteams}
        )
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_notification_microsoftteams(
            name="teams-alerts",
            webhook="https://outlook.office.com/webhook/xxx/yyy/zzz",
        )

        assert result["id"] == 4
        assert result["name"] == "teams-alerts"
        assert result["webhook"] == "https://outlook.office.com/webhook/xxx/yyy/zzz"

    def test_get_all_notification_microsoftteams(
        self, lagoon_client, mock_response, sample_notification_microsoftteams
    ):
        """Test getting all Microsoft Teams notifications."""
        all_notifications = [
            {**sample_notification_microsoftteams, "__typename": "NotificationMicrosoftTeams"},
            {
                "__typename": "NotificationSlack",
                "id": 1,
                "name": "slack-alert",
                "webhook": "https://slack.com",
                "channel": "#test",
            },
        ]
        response = mock_response(data={"allNotifications": all_notifications})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_all_notification_microsoftteams()

        assert len(result) == 1
        assert result[0]["name"] == "teams-alerts"
        assert result[0]["__typename"] == "NotificationMicrosoftTeams"

    def test_get_notification_microsoftteams_by_name(
        self, lagoon_client, mock_response, sample_notification_microsoftteams
    ):
        """Test getting Microsoft Teams notification by name."""
        all_notifications = [
            {**sample_notification_microsoftteams, "__typename": "NotificationMicrosoftTeams"},
        ]
        response = mock_response(data={"allNotifications": all_notifications})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_notification_microsoftteams_by_name("teams-alerts")

        assert result is not None
        assert result["name"] == "teams-alerts"

    def test_get_notification_microsoftteams_by_name_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent Microsoft Teams notification returns None."""
        response = mock_response(data={"allNotifications": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_notification_microsoftteams_by_name("nonexistent")

        assert result is None

    def test_update_notification_microsoftteams(
        self, lagoon_client, mock_response, sample_notification_microsoftteams
    ):
        """Test updating a Microsoft Teams notification."""
        updated = sample_notification_microsoftteams.copy()
        updated["webhook"] = "https://outlook.office.com/webhook/new/url/here"
        response = mock_response(data={"updateNotificationMicrosoftTeams": updated})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.update_notification_microsoftteams(
            name="teams-alerts",
            webhook="https://outlook.office.com/webhook/new/url/here",
        )

        assert result["webhook"] == "https://outlook.office.com/webhook/new/url/here"

        # Verify patch format is used
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["name"] == "teams-alerts"
        assert input_data["patch"]["webhook"] == "https://outlook.office.com/webhook/new/url/here"

    def test_delete_notification_microsoftteams(self, lagoon_client, mock_response):
        """Test deleting a Microsoft Teams notification."""
        response = mock_response(data={"deleteNotificationMicrosoftTeams": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_notification_microsoftteams(name="teams-alerts")

        assert result == "success"


class TestProjectNotificationOperations:
    """Tests for project notification association operations."""

    def test_add_notification_to_project(self, lagoon_client, mock_response, sample_project):
        """Test adding a notification to a project."""
        response = mock_response(
            data={"addNotificationToProject": {"id": 1, "name": "test-project"}}
        )
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_notification_to_project(
            project="test-project",
            notification_type="slack",
            notification_name="deploy-alerts",
        )

        assert result["id"] == 1
        assert result["name"] == "test-project"

        # Verify notification type is uppercased
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["project"] == "test-project"
        assert input_data["notificationType"] == "SLACK"
        assert input_data["notificationName"] == "deploy-alerts"

    def test_add_notification_to_project_all_types(self, lagoon_client, mock_response):
        """Test adding different notification types to a project."""
        response = mock_response(
            data={"addNotificationToProject": {"id": 1, "name": "test-project"}}
        )
        lagoon_client.session.post.return_value = response

        for notification_type in ["slack", "rocketchat", "email", "microsoftteams"]:
            lagoon_client.add_notification_to_project(
                project="test-project",
                notification_type=notification_type,
                notification_name="test-notification",
            )

            call_kwargs = lagoon_client.session.post.call_args[1]
            input_data = call_kwargs["json"]["variables"]["input"]
            assert input_data["notificationType"] == notification_type.upper()

    def test_remove_notification_from_project(self, lagoon_client, mock_response):
        """Test removing a notification from a project."""
        response = mock_response(
            data={"removeNotificationFromProject": {"id": 1, "name": "test-project"}}
        )
        lagoon_client.session.post.return_value = response

        result = lagoon_client.remove_notification_from_project(
            project="test-project",
            notification_type="slack",
            notification_name="deploy-alerts",
        )

        assert result["id"] == 1
        assert result["name"] == "test-project"

        # Verify notification type is uppercased
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["notificationType"] == "SLACK"

    def test_get_project_notifications(
        self, lagoon_client, mock_response, sample_notification_slack, sample_notification_email
    ):
        """Test getting all notifications linked to a project."""
        project_data = {
            "id": 1,
            "name": "test-project",
            "notifications": [
                {**sample_notification_slack, "__typename": "NotificationSlack"},
                {**sample_notification_email, "__typename": "NotificationEmail"},
            ],
        }
        response = mock_response(data={"projectByName": project_data})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_project_notifications("test-project")

        assert len(result["slack"]) == 1
        assert result["slack"][0]["name"] == "deploy-alerts"
        assert len(result["email"]) == 1
        assert result["email"][0]["name"] == "ops-team"
        assert len(result["rocketchat"]) == 0
        assert len(result["microsoftteams"]) == 0

    def test_get_project_notifications_empty(self, lagoon_client, mock_response):
        """Test getting notifications for project with no notifications."""
        project_data = {
            "id": 1,
            "name": "test-project",
            "notifications": [],
        }
        response = mock_response(data={"projectByName": project_data})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_project_notifications("test-project")

        assert result["slack"] == []
        assert result["email"] == []
        assert result["rocketchat"] == []
        assert result["microsoftteams"] == []

    def test_get_project_notifications_project_not_found(self, lagoon_client, mock_response):
        """Test getting notifications for nonexistent project returns empty dict."""
        response = mock_response(data={"projectByName": None})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_project_notifications("nonexistent")

        assert result == {}

    def test_check_project_notification_exists_true(
        self, lagoon_client, mock_response, sample_notification_slack
    ):
        """Test checking that a notification exists on a project."""
        project_data = {
            "id": 1,
            "name": "test-project",
            "notifications": [
                {**sample_notification_slack, "__typename": "NotificationSlack"},
            ],
        }
        response = mock_response(data={"projectByName": project_data})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.check_project_notification_exists(
            project_name="test-project",
            notification_type="slack",
            notification_name="deploy-alerts",
        )

        assert result is True

    def test_check_project_notification_exists_false(
        self, lagoon_client, mock_response, sample_notification_slack
    ):
        """Test checking that a notification does not exist on a project."""
        project_data = {
            "id": 1,
            "name": "test-project",
            "notifications": [
                {**sample_notification_slack, "__typename": "NotificationSlack"},
            ],
        }
        response = mock_response(data={"projectByName": project_data})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.check_project_notification_exists(
            project_name="test-project",
            notification_type="slack",
            notification_name="nonexistent",
        )

        assert result is False

    def test_check_project_notification_exists_wrong_type(
        self, lagoon_client, mock_response, sample_notification_slack
    ):
        """Test checking notification with wrong type returns false."""
        project_data = {
            "id": 1,
            "name": "test-project",
            "notifications": [
                {**sample_notification_slack, "__typename": "NotificationSlack"},
            ],
        }
        response = mock_response(data={"projectByName": project_data})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.check_project_notification_exists(
            project_name="test-project",
            notification_type="email",
            notification_name="deploy-alerts",
        )

        assert result is False


class TestGetAllNotifications:
    """Tests for the _get_all_notifications internal method."""

    def test_get_all_notifications_mixed_types(
        self,
        lagoon_client,
        mock_response,
        sample_notification_slack,
        sample_notification_rocketchat,
        sample_notification_email,
        sample_notification_microsoftteams,
    ):
        """Test getting all notifications with mixed types."""
        all_notifications = [
            {**sample_notification_slack, "__typename": "NotificationSlack"},
            {**sample_notification_rocketchat, "__typename": "NotificationRocketChat"},
            {**sample_notification_email, "__typename": "NotificationEmail"},
            {**sample_notification_microsoftteams, "__typename": "NotificationMicrosoftTeams"},
        ]
        response = mock_response(data={"allNotifications": all_notifications})
        lagoon_client.session.post.return_value = response

        result = lagoon_client._get_all_notifications()

        assert len(result) == 4
        types = [n["__typename"] for n in result]
        assert "NotificationSlack" in types
        assert "NotificationRocketChat" in types
        assert "NotificationEmail" in types
        assert "NotificationMicrosoftTeams" in types

    def test_get_all_notifications_empty(self, lagoon_client, mock_response):
        """Test getting all notifications when none exist."""
        response = mock_response(data={"allNotifications": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client._get_all_notifications()

        assert result == []

    def test_get_all_notifications_none_response(self, lagoon_client, mock_response):
        """Test getting all notifications when response is None returns empty list."""
        response = mock_response(data={"allNotifications": None})
        lagoon_client.session.post.return_value = response

        result = lagoon_client._get_all_notifications()

        # Client should handle None response gracefully by returning empty list
        assert result == []


class TestKubernetesOperations:
    """Tests for Kubernetes deploy target CRUD operations."""

    def test_add_kubernetes(self, lagoon_client, mock_response, sample_deploy_target):
        """Test adding a Kubernetes deploy target."""
        response = mock_response(data={"addKubernetes": sample_deploy_target})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_kubernetes(
            name="prod-cluster",
            console_url="https://kubernetes.example.com:6443",
            cloud_provider="aws",
            cloud_region="us-east-1",
        )

        assert result["id"] == 1
        assert result["name"] == "prod-cluster"
        assert result["consoleUrl"] == "https://kubernetes.example.com:6443"

        # Verify the GraphQL mutation was called with correct input
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["name"] == "prod-cluster"
        assert input_data["consoleUrl"] == "https://kubernetes.example.com:6443"
        assert input_data["cloudProvider"] == "aws"
        assert input_data["cloudRegion"] == "us-east-1"

    def test_add_kubernetes_with_optional_params(
        self, lagoon_client, mock_response, sample_deploy_target
    ):
        """Test adding a Kubernetes deploy target with optional parameters."""
        response = mock_response(data={"addKubernetes": sample_deploy_target})
        lagoon_client.session.post.return_value = response

        lagoon_client.add_kubernetes(
            name="prod-cluster",
            console_url="https://kubernetes.example.com:6443",
            cloud_provider="aws",
            cloud_region="us-east-1",
            sshHost="ssh.lagoon.example.com",
            sshPort="22",
        )

        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["sshHost"] == "ssh.lagoon.example.com"
        assert input_data["sshPort"] == "22"

    def test_get_all_kubernetes(self, lagoon_client, mock_response, sample_deploy_target):
        """Test getting all Kubernetes deploy targets."""
        response = mock_response(data={"allKubernetes": [sample_deploy_target]})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_all_kubernetes()

        assert len(result) == 1
        assert result[0]["name"] == "prod-cluster"

    def test_get_all_kubernetes_empty(self, lagoon_client, mock_response):
        """Test getting all Kubernetes deploy targets when none exist."""
        response = mock_response(data={"allKubernetes": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_all_kubernetes()

        assert result == []

    def test_get_kubernetes_by_id(self, lagoon_client, mock_response, sample_deploy_target):
        """Test getting Kubernetes deploy target by ID.

        Note: The implementation now uses allKubernetes and filters by ID
        for compatibility with Lagoon v2.30.0+ which doesn't have kubernetes(id:).
        """
        response = mock_response(data={"allKubernetes": [sample_deploy_target]})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_kubernetes_by_id(1)

        assert result["id"] == 1
        assert result["name"] == "prod-cluster"

    def test_get_kubernetes_by_id_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent Kubernetes deploy target returns None."""
        response = mock_response(data={"allKubernetes": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_kubernetes_by_id(999)

        assert result is None

    def test_get_kubernetes_by_name(self, lagoon_client, mock_response, sample_deploy_target):
        """Test getting Kubernetes deploy target by name."""
        response = mock_response(data={"allKubernetes": [sample_deploy_target]})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_kubernetes_by_name("prod-cluster")

        assert result is not None
        assert result["name"] == "prod-cluster"

    def test_get_kubernetes_by_name_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent Kubernetes deploy target by name returns None."""
        response = mock_response(data={"allKubernetes": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_kubernetes_by_name("nonexistent")

        assert result is None

    def test_update_kubernetes(self, lagoon_client, mock_response, sample_deploy_target):
        """Test updating a Kubernetes deploy target."""
        updated = sample_deploy_target.copy()
        updated["cloudRegion"] = "us-west-2"
        response = mock_response(data={"updateKubernetes": updated})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.update_kubernetes(k8s_id=1, cloudRegion="us-west-2")

        assert result["cloudRegion"] == "us-west-2"

        # Verify the GraphQL mutation was called with correct input
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["id"] == 1
        assert input_data["cloudRegion"] == "us-west-2"

    def test_delete_kubernetes(self, lagoon_client, mock_response):
        """Test deleting a Kubernetes deploy target."""
        response = mock_response(data={"deleteKubernetes": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_kubernetes(name="prod-cluster")

        assert result == "success"

        # Verify correct input
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["name"] == "prod-cluster"


class TestDeployTargetConfigOperations:
    """Tests for deploy target configuration CRUD operations."""

    def test_add_deploy_target_config(self, lagoon_client, mock_response):
        """Test adding a deploy target configuration."""
        config_data = {
            "id": 1,
            "weight": 1,
            "branches": "^main$",
            "pullrequests": "false",
            "deployTargetProjectPattern": None,
            "deployTarget": {"id": 1, "name": "prod-cluster"},
            "project": {"id": 1, "name": "test-project"},
        }
        response = mock_response(data={"addDeployTargetConfig": config_data})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_deploy_target_config(
            project=1,
            deploy_target=1,
            branches="^main$",
            pullrequests="false",
            weight=1,
        )

        assert result["id"] == 1
        assert result["branches"] == "^main$"
        assert result["deployTargetId"] == 1
        assert result["projectId"] == 1

    def test_add_deploy_target_config_with_pattern(self, lagoon_client, mock_response):
        """Test adding a deploy target configuration with namespace pattern."""
        config_data = {
            "id": 1,
            "weight": 1,
            "branches": "^main$",
            "pullrequests": "false",
            "deployTargetProjectPattern": "${project}-${environment}",
            "deployTarget": {"id": 1, "name": "prod-cluster"},
            "project": {"id": 1, "name": "test-project"},
        }
        response = mock_response(data={"addDeployTargetConfig": config_data})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.add_deploy_target_config(
            project=1,
            deploy_target=1,
            branches="^main$",
            deploy_target_project_pattern="${project}-${environment}",
        )

        assert result["deployTargetProjectPattern"] == "${project}-${environment}"

        # Verify pattern was included in input
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["deployTargetProjectPattern"] == "${project}-${environment}"

    def test_get_deploy_target_configs_by_project(self, lagoon_client, mock_response):
        """Test getting deploy target configs for a project."""
        configs = [
            {
                "id": 1,
                "weight": 1,
                "branches": "^main$",
                "pullrequests": "false",
                "deployTargetProjectPattern": None,
                "deployTarget": {"id": 1, "name": "prod-cluster"},
                "project": {"id": 1, "name": "test-project"},
            },
            {
                "id": 2,
                "weight": 2,
                "branches": "^develop$",
                "pullrequests": "true",
                "deployTargetProjectPattern": None,
                "deployTarget": {"id": 2, "name": "dev-cluster"},
                "project": {"id": 1, "name": "test-project"},
            },
        ]
        response = mock_response(data={"deployTargetConfigsByProjectId": configs})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_deploy_target_configs_by_project(project=1)

        assert len(result) == 2
        assert result[0]["deployTargetId"] == 1
        assert result[1]["deployTargetId"] == 2

    def test_get_deploy_target_configs_empty(self, lagoon_client, mock_response):
        """Test getting deploy target configs when none exist."""
        response = mock_response(data={"deployTargetConfigsByProjectId": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_deploy_target_configs_by_project(project=1)

        assert result == []

    def test_get_deploy_target_config_by_id(self, lagoon_client, mock_response):
        """Test getting a specific deploy target config by ID."""
        configs = [
            {
                "id": 1,
                "weight": 1,
                "branches": "^main$",
                "pullrequests": "false",
                "deployTargetProjectPattern": None,
                "deployTarget": {"id": 1, "name": "prod-cluster"},
                "project": {"id": 1, "name": "test-project"},
            },
        ]
        response = mock_response(data={"deployTargetConfigsByProjectId": configs})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_deploy_target_config_by_id(config_id=1, project=1)

        assert result is not None
        assert result["id"] == 1

    def test_get_deploy_target_config_by_id_not_found(self, lagoon_client, mock_response):
        """Test getting nonexistent deploy target config returns None."""
        response = mock_response(data={"deployTargetConfigsByProjectId": []})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.get_deploy_target_config_by_id(config_id=999, project=1)

        assert result is None

    def test_update_deploy_target_config(self, lagoon_client, mock_response):
        """Test updating a deploy target configuration."""
        updated_config = {
            "id": 1,
            "weight": 5,
            "branches": "^(main|develop)$",
            "pullrequests": "true",
            "deployTargetProjectPattern": None,
            "deployTarget": {"id": 1, "name": "prod-cluster"},
            "project": {"id": 1, "name": "test-project"},
        }
        response = mock_response(data={"updateDeployTargetConfig": updated_config})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.update_deploy_target_config(
            config_id=1,
            branches="^(main|develop)$",
            weight=5,
        )

        assert result["branches"] == "^(main|develop)$"
        assert result["weight"] == 5
        assert result["deployTargetId"] == 1

    def test_delete_deploy_target_config(self, lagoon_client, mock_response):
        """Test deleting a deploy target configuration."""
        response = mock_response(data={"deleteDeployTargetConfig": "success"})
        lagoon_client.session.post.return_value = response

        result = lagoon_client.delete_deploy_target_config(config_id=1, project=1)

        assert result == "success"

        # Verify correct input
        call_kwargs = lagoon_client.session.post.call_args[1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["id"] == 1
        assert input_data["project"] == 1


class TestAPIFallbackCode:
    """Tests for backward compatibility fallback code for older Lagoon API versions."""

    def test_add_env_variable_fallback_on_cannot_query_field(self, lagoon_client, mock_response):
        """Test that add_env_variable falls back to old API on 'Cannot query field' error."""
        # Call sequence:
        # 1. get_project_by_id (via allProjects fallback since projectById may fail)
        # 2. addOrUpdateEnvVariableByName (fails with "Cannot query field")
        # 3. addEnvVariable (old API, succeeds)
        project_response = mock_response(
            data={"allProjects": [{"id": 1, "name": "test-project"}]}
        )
        error_response = mock_response(
            errors=[{"message": "Cannot query field 'addOrUpdateEnvVariableByName'"}]
        )
        success_response = mock_response(
            data={"addEnvVariable": {"id": 1, "name": "TEST_VAR", "value": "test", "scope": "RUNTIME"}}
        )
        lagoon_client.session.post.side_effect = [project_response, error_response, success_response]

        result = lagoon_client.add_env_variable(
            name="TEST_VAR", value="test", scope="runtime", project=1
        )

        assert result["id"] == 1
        assert result["name"] == "TEST_VAR"

    def test_add_env_variable_fallback_with_environment(self, lagoon_client, mock_response):
        """Test that add_env_variable fallback works for environment variables."""
        # Call sequence:
        # 1. get_project_by_id (success)
        # 2. environmentById (may fail but we handle it)
        # 3. addOrUpdateEnvVariableByName (fails)
        # 4. addEnvVariable (old API, succeeds)
        project_response = mock_response(
            data={"allProjects": [{"id": 1, "name": "test-project"}]}
        )
        env_response = mock_response(
            data={"environmentById": {"id": 5, "name": "main"}}
        )
        error_response = mock_response(errors=[{"message": "Cannot query field"}])
        success_response = mock_response(
            data={"addEnvVariable": {"id": 2, "name": "DB_HOST", "value": "localhost", "scope": "RUNTIME"}}
        )
        lagoon_client.session.post.side_effect = [project_response, env_response, error_response, success_response]

        result = lagoon_client.add_env_variable(
            name="DB_HOST", value="localhost", scope="runtime", project=1, environment=5
        )

        assert result["id"] == 2
        # Verify the old API format was used with ENVIRONMENT type
        call_kwargs = lagoon_client.session.post.call_args_list[3][1]
        input_data = call_kwargs["json"]["variables"]["input"]
        assert input_data["type"] == "ENVIRONMENT"
        assert input_data["typeId"] == 5

    def test_add_env_variable_reraises_unrelated_error(self, lagoon_client, mock_response):
        """Test that add_env_variable re-raises errors that are not API version related."""
        # Call sequence:
        # 1. get_project_by_id (success)
        # 2. addOrUpdateEnvVariableByName (fails with unrelated error)
        project_response = mock_response(
            data={"allProjects": [{"id": 1, "name": "test-project"}]}
        )
        error_response = mock_response(errors=[{"message": "Permission denied"}])
        lagoon_client.session.post.side_effect = [project_response, error_response]

        with pytest.raises(LagoonAPIError, match="Permission denied"):
            lagoon_client.add_env_variable(
                name="TEST_VAR", value="test", scope="runtime", project=1
            )

    def test_get_env_variable_fallback(self, lagoon_client, mock_response):
        """Test that get_env_variable_by_name falls back to envVariablesByProjectEnvironment."""
        # Call sequence:
        # 1. get_project_by_id (success)
        # 2. getEnvVariablesByProjectEnvironmentName (fails)
        # 3. envVariablesByProjectEnvironment (old API, success)
        project_response = mock_response(
            data={"allProjects": [{"id": 1, "name": "test-project"}]}
        )
        error_response = mock_response(errors=[{"message": "Cannot query field"}])
        old_api_response = mock_response(
            data={
                "envVariablesByProjectEnvironment": [
                    {"id": 1, "name": "TEST_VAR", "value": "test_value", "scope": "RUNTIME"}
                ]
            }
        )
        lagoon_client.session.post.side_effect = [project_response, error_response, old_api_response]

        result = lagoon_client.get_env_variable_by_name("TEST_VAR", project=1)

        assert result is not None
        assert result["name"] == "TEST_VAR"
        assert result["value"] == "test_value"

    def test_delete_env_variable_fallback(self, lagoon_client, mock_response):
        """Test that delete_env_variable falls back to old API format."""
        # Call sequence:
        # 1. get_project_by_id (success)
        # 2. deleteEnvVariableByName (fails)
        # 3. deleteEnvVariable old API (success)
        project_response = mock_response(
            data={"allProjects": [{"id": 1, "name": "test-project"}]}
        )
        error_response = mock_response(errors=[{"message": "Cannot query field"}])
        success_response = mock_response(data={"deleteEnvVariable": "success"})
        lagoon_client.session.post.side_effect = [project_response, error_response, success_response]

        result = lagoon_client.delete_env_variable(name="TEST_VAR", project=1)

        assert result == "success"

    def test_get_project_by_id_fallback(self, lagoon_client, mock_response, sample_project):
        """Test that get_project_by_id falls back to allProjects query."""
        # This method now uses allProjects directly, so just test it works
        all_projects_response = mock_response(data={"allProjects": [sample_project]})
        lagoon_client.session.post.return_value = all_projects_response

        result = lagoon_client.get_project_by_id(1)

        assert result is not None
        assert result["id"] == 1
        assert result["name"] == "test-project"

    def test_get_project_by_id_not_found(self, lagoon_client, mock_response):
        """Test that get_project_by_id returns None when project not found."""
        all_projects_response = mock_response(data={"allProjects": []})
        lagoon_client.session.post.return_value = all_projects_response

        result = lagoon_client.get_project_by_id(999)

        assert result is None

    def test_get_kubernetes_by_id_uses_get_all(self, lagoon_client, mock_response, sample_deploy_target):
        """Test that get_kubernetes_by_id filters from get_all_kubernetes."""
        all_k8s_response = mock_response(data={"allKubernetes": [sample_deploy_target]})
        lagoon_client.session.post.return_value = all_k8s_response

        result = lagoon_client.get_kubernetes_by_id(1)

        assert result is not None
        assert result["id"] == 1
        assert result["name"] == "prod-cluster"

    def test_get_kubernetes_by_id_not_found(self, lagoon_client, mock_response):
        """Test that get_kubernetes_by_id returns None when not found."""
        all_k8s_response = mock_response(data={"allKubernetes": []})
        lagoon_client.session.post.return_value = all_k8s_response

        result = lagoon_client.get_kubernetes_by_id(999)

        assert result is None


class TestAdvancedTaskFallback:
    """Tests for advanced task API fallback code."""

    def test_get_advanced_tasks_by_environment_fallback(self, lagoon_client, mock_response):
        """Test that get_advanced_tasks_by_environment falls back to older API."""
        # New API fails
        error_response = mock_response(errors=[{"message": "Cannot query field"}])
        # Old API succeeds with different format (project/environment as objects)
        old_api_response = mock_response(
            data={
                "advancedTasksByEnvironment": [
                    {
                        "id": 1,
                        "name": "run-tests",
                        "type": "COMMAND",
                        "service": "cli",
                        "command": "yarn test",
                        "permission": "DEVELOPER",
                        "project": {"id": 1, "name": "test-project"},
                        "environment": {"id": 5, "name": "main"},
                        "groupName": None,
                    }
                ]
            }
        )
        lagoon_client.session.post.side_effect = [error_response, old_api_response]

        result = lagoon_client.get_advanced_tasks_by_environment(5)

        assert len(result) == 1
        assert result[0]["name"] == "run-tests"
        # Verify normalization added projectId and environmentId
        assert result[0]["projectId"] == 1
        assert result[0]["environmentId"] == 5

    def test_get_advanced_tasks_normalization_new_api(self, lagoon_client, mock_response):
        """Test that new API format (int IDs) gets normalized correctly."""
        new_api_response = mock_response(
            data={
                "advancedTasksForEnvironment": [
                    {
                        "id": 1,
                        "name": "run-tests",
                        "type": "COMMAND",
                        "service": "cli",
                        "command": "yarn test",
                        "permission": "DEVELOPER",
                        "project": 1,  # New API returns int
                        "environment": 5,  # New API returns int
                        "groupName": None,
                    }
                ]
            }
        )
        lagoon_client.session.post.return_value = new_api_response

        result = lagoon_client.get_advanced_tasks_by_environment(5)

        assert len(result) == 1
        # Verify normalization works for int values too
        assert result[0]["projectId"] == 1
        assert result[0]["environmentId"] == 5

    def test_get_advanced_tasks_reraises_unrelated_error(self, lagoon_client, mock_response):
        """Test that unrelated errors are re-raised."""
        error_response = mock_response(errors=[{"message": "Access denied"}])
        lagoon_client.session.post.return_value = error_response

        with pytest.raises(LagoonAPIError, match="Access denied"):
            lagoon_client.get_advanced_tasks_by_environment(5)
