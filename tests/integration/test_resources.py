"""Integration tests for Lagoon resources.

These tests require a running Lagoon instance and are skipped by default.
To run them, use: pytest -m integration

Environment variables required:
- LAGOON_API_URL: Lagoon GraphQL API endpoint
- LAGOON_TOKEN: Valid authentication token
- LAGOON_DEPLOY_TARGET_ID: ID of a deploy target to use for tests
"""

import pytest
import os

from pulumi_lagoon.client import LagoonClient, LagoonAPIError


# Mark all tests in this module as integration tests
pytestmark = pytest.mark.integration


@pytest.fixture(scope="module")
def lagoon_config():
    """Get Lagoon configuration from environment."""
    api_url = os.environ.get("LAGOON_API_URL")
    token = os.environ.get("LAGOON_TOKEN")
    deploy_target_id = os.environ.get("LAGOON_DEPLOY_TARGET_ID")

    if not all([api_url, token, deploy_target_id]):
        pytest.skip(
            "Integration tests require LAGOON_API_URL, LAGOON_TOKEN, "
            "and LAGOON_DEPLOY_TARGET_ID environment variables"
        )

    return {
        "api_url": api_url,
        "token": token,
        "deploy_target_id": int(deploy_target_id),
    }


@pytest.fixture(scope="module")
def client(lagoon_config):
    """Create a real Lagoon client."""
    return LagoonClient(
        api_url=lagoon_config["api_url"],
        token=lagoon_config["token"]
    )


@pytest.fixture
def test_project_name():
    """Generate a unique test project name."""
    import time
    return f"test-integration-{int(time.time())}"


class TestProjectLifecycle:
    """Integration tests for project CRUD operations."""

    def test_create_get_delete_project(self, client, lagoon_config, test_project_name):
        """Test full project lifecycle: create, read, delete."""
        # Create
        project = client.create_project(
            name=test_project_name,
            git_url="git@github.com:test/integration-test.git",
            openshift=lagoon_config["deploy_target_id"],
            productionEnvironment="main",
            branches="^(main|develop)$",
        )

        assert project["name"] == test_project_name
        assert project["id"] is not None
        project_id = project["id"]

        try:
            # Read by name
            fetched = client.get_project_by_name(test_project_name)
            assert fetched is not None
            assert fetched["id"] == project_id

            # Read by ID
            fetched_by_id = client.get_project_by_id(project_id)
            assert fetched_by_id is not None
            assert fetched_by_id["name"] == test_project_name

        finally:
            # Cleanup
            client.delete_project(test_project_name)

            # Verify deletion
            deleted = client.get_project_by_name(test_project_name)
            assert deleted is None

    def test_update_project(self, client, lagoon_config, test_project_name):
        """Test project update operation."""
        # Create
        project = client.create_project(
            name=test_project_name,
            git_url="git@github.com:test/integration-test.git",
            openshift=lagoon_config["deploy_target_id"],
            branches="^main$",
        )

        try:
            # Update
            updated = client.update_project(
                project_id=project["id"],
                branches="^(main|develop|staging)$",
            )

            assert updated["branches"] == "^(main|develop|staging)$"

            # Verify update persisted
            fetched = client.get_project_by_id(project["id"])
            assert fetched["branches"] == "^(main|develop|staging)$"

        finally:
            client.delete_project(test_project_name)


class TestEnvironmentLifecycle:
    """Integration tests for environment CRUD operations."""

    @pytest.fixture
    def project_with_cleanup(self, client, lagoon_config, test_project_name):
        """Create a project and clean it up after the test."""
        project = client.create_project(
            name=test_project_name,
            git_url="git@github.com:test/integration-test.git",
            openshift=lagoon_config["deploy_target_id"],
            productionEnvironment="main",
        )

        yield project

        # Cleanup
        try:
            client.delete_project(test_project_name)
        except Exception:
            pass

    def test_create_get_delete_environment(self, client, project_with_cleanup):
        """Test full environment lifecycle."""
        project_id = project_with_cleanup["id"]

        # Create
        env = client.add_or_update_environment(
            name="develop",
            project=project_id,
            deploy_type="branch",
            environment_type="development",
            deployBaseRef="develop",
        )

        assert env["name"] == "develop"
        env_id = env["id"]

        try:
            # Read
            fetched = client.get_environment_by_name("develop", project_id)
            assert fetched is not None
            assert fetched["id"] == env_id

        finally:
            # Delete
            client.delete_environment(
                name="develop",
                project=project_id,
                execute=True
            )

            # Verify deletion
            deleted = client.get_environment_by_name("develop", project_id)
            assert deleted is None


class TestVariableLifecycle:
    """Integration tests for variable CRUD operations."""

    @pytest.fixture
    def project_with_cleanup(self, client, lagoon_config, test_project_name):
        """Create a project and clean it up after the test."""
        project = client.create_project(
            name=test_project_name,
            git_url="git@github.com:test/integration-test.git",
            openshift=lagoon_config["deploy_target_id"],
        )

        yield project

        try:
            client.delete_project(test_project_name)
        except Exception:
            pass

    def test_project_level_variable(self, client, project_with_cleanup):
        """Test project-level variable lifecycle."""
        project_id = project_with_cleanup["id"]

        # Create
        var = client.add_env_variable(
            name="TEST_VAR",
            value="test-value",
            project=project_id,
            scope="runtime",
        )

        assert var["name"] == "TEST_VAR"
        assert var["value"] == "test-value"

        try:
            # Read
            fetched = client.get_env_variable_by_name(
                name="TEST_VAR",
                project=project_id
            )
            assert fetched is not None
            assert fetched["value"] == "test-value"

        finally:
            # Delete
            client.delete_env_variable(
                name="TEST_VAR",
                project=project_id
            )

            # Verify deletion
            deleted = client.get_env_variable_by_name(
                name="TEST_VAR",
                project=project_id
            )
            assert deleted is None

    def test_variable_scopes(self, client, project_with_cleanup):
        """Test different variable scopes."""
        project_id = project_with_cleanup["id"]

        scopes = ["BUILD", "RUNTIME", "GLOBAL"]

        for scope in scopes:
            var_name = f"TEST_{scope}_VAR"

            # Create with specific scope
            var = client.add_env_variable(
                name=var_name,
                value=f"{scope.lower()}-value",
                project=project_id,
                scope=scope.lower(),
            )

            assert var["scope"] == scope

            # Cleanup
            client.delete_env_variable(
                name=var_name,
                project=project_id
            )
