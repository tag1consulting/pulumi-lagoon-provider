"""Shared test fixtures for pulumi-lagoon tests."""

import os
from unittest.mock import Mock, patch

import pytest

# Sample data for tests
SAMPLE_PROJECT = {
    "id": 1,
    "name": "test-project",
    "gitUrl": "git@github.com:test/test-repo.git",
    "openshift": 1,
    "productionEnvironment": "main",
    "branches": "^(main|develop)$",
    "pullrequests": ".*",
    "created": "2024-01-01T00:00:00Z",
}

SAMPLE_ENVIRONMENT = {
    "id": 1,
    "name": "main",
    "project": {"id": 1, "name": "test-project"},
    "environmentType": "PRODUCTION",
    "deployType": "BRANCH",
    "deployBaseRef": "main",
    "deployHeadRef": None,
    "deployTitle": None,
    "autoIdle": 1,
    "route": "https://main.test-project.example.com",
    "routes": "https://main.test-project.example.com",
    "created": "2024-01-01T00:00:00Z",
}

SAMPLE_VARIABLE = {
    "id": 1,
    "name": "DATABASE_HOST",
    "value": "mysql.example.com",
    "scope": "RUNTIME",
    "project": {"id": 1, "name": "test-project"},
    "environment": None,
}

SAMPLE_DEPLOY_TARGET = {
    "id": 1,
    "name": "prod-cluster",
    "consoleUrl": "https://kubernetes.example.com:6443",
    "cloudProvider": "aws",
    "cloudRegion": "us-east-1",
    "sshHost": "ssh.lagoon.example.com",
    "sshPort": "22",
    "buildImage": None,
    "disabled": False,
    "routerPattern": None,
    "sharedBastionSecret": None,
    "created": "2024-01-01T00:00:00Z",
}

# Notification sample data
SAMPLE_NOTIFICATION_SLACK = {
    "id": 1,
    "name": "deploy-alerts",
    "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
    "channel": "#deployments",
}

SAMPLE_NOTIFICATION_ROCKETCHAT = {
    "id": 2,
    "name": "team-chat",
    "webhook": "https://rocketchat.example.com/hooks/xxx/yyy",
    "channel": "#alerts",
}

SAMPLE_NOTIFICATION_EMAIL = {
    "id": 3,
    "name": "ops-team",
    "emailAddress": "ops@example.com",
}

SAMPLE_NOTIFICATION_MICROSOFTTEAMS = {
    "id": 4,
    "name": "teams-alerts",
    "webhook": "https://outlook.office.com/webhook/xxx/yyy/zzz",
}


@pytest.fixture
def mock_session():
    """Create a mock requests session."""
    with patch("requests.Session") as mock:
        session_instance = Mock()
        mock.return_value = session_instance
        yield session_instance


@pytest.fixture
def mock_response():
    """Create a factory for mock HTTP responses."""

    def _make_response(data=None, errors=None, status_code=200):
        response = Mock()
        response.status_code = status_code

        json_data = {}
        if data is not None:
            json_data["data"] = data
        if errors is not None:
            json_data["errors"] = errors

        response.json.return_value = json_data
        response.raise_for_status = Mock()

        if status_code >= 400:
            from requests import HTTPError

            response.raise_for_status.side_effect = HTTPError(f"HTTP {status_code}")

        return response

    return _make_response


@pytest.fixture
def lagoon_client(mock_response):
    """Create a LagoonClient instance with mocked session."""
    from pulumi_lagoon.client import LagoonClient

    with patch("requests.Session") as mock_session_class:
        mock_session = Mock()
        mock_session.headers = Mock()
        mock_session.headers.update = Mock()
        mock_session_class.return_value = mock_session

        client = LagoonClient(api_url="https://api.test.lagoon.sh/graphql", token="test-jwt-token")

        # Replace the session with our mock
        client.session = mock_session

        yield client


@pytest.fixture
def sample_project():
    """Return sample project data."""
    return SAMPLE_PROJECT.copy()


@pytest.fixture
def sample_environment():
    """Return sample environment data."""
    return SAMPLE_ENVIRONMENT.copy()


@pytest.fixture
def sample_variable():
    """Return sample variable data."""
    return SAMPLE_VARIABLE.copy()


@pytest.fixture
def sample_deploy_target():
    """Return sample deploy target data."""
    return SAMPLE_DEPLOY_TARGET.copy()


@pytest.fixture
def sample_notification_slack():
    """Return sample Slack notification data."""
    return SAMPLE_NOTIFICATION_SLACK.copy()


@pytest.fixture
def sample_notification_rocketchat():
    """Return sample RocketChat notification data."""
    return SAMPLE_NOTIFICATION_ROCKETCHAT.copy()


@pytest.fixture
def sample_notification_email():
    """Return sample Email notification data."""
    return SAMPLE_NOTIFICATION_EMAIL.copy()


@pytest.fixture
def sample_notification_microsoftteams():
    """Return sample Microsoft Teams notification data."""
    return SAMPLE_NOTIFICATION_MICROSOFTTEAMS.copy()


@pytest.fixture
def env_vars():
    """Set environment variables for testing and clean up after."""
    original_env = os.environ.copy()

    # Set test environment variables
    test_vars = {
        "LAGOON_API_URL": "https://api.test.lagoon.sh/graphql",
        "LAGOON_TOKEN": "test-token-from-env",
    }

    os.environ.update(test_vars)

    yield test_vars

    # Restore original environment
    os.environ.clear()
    os.environ.update(original_env)


@pytest.fixture
def clean_env():
    """Remove Lagoon environment variables for testing defaults."""
    original_env = os.environ.copy()

    # Remove Lagoon env vars
    for key in list(os.environ.keys()):
        if key.startswith("LAGOON_"):
            del os.environ[key]

    yield

    # Restore original environment
    os.environ.clear()
    os.environ.update(original_env)


@pytest.fixture
def mock_pulumi_config():
    """Mock Pulumi config for testing."""
    with patch("pulumi.Config") as mock_config_class:
        mock_config = Mock()
        mock_config_class.return_value = mock_config

        # Default: return None for all config values
        mock_config.get.return_value = None
        mock_config.get_secret.return_value = None

        yield mock_config
