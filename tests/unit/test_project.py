"""Unit tests for LagoonProject provider."""

from unittest.mock import Mock, patch


class TestLagoonProjectProviderCreate:
    """Tests for LagoonProjectProvider.create method."""

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_create_project_minimal(self, mock_config_class, sample_project):
        """Test creating a project with minimal arguments."""
        from pulumi_lagoon.project import LagoonProjectProvider

        # Setup mock
        mock_client = Mock()
        mock_client.create_project.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectProvider()

        inputs = {
            "name": "test-project",
            "git_url": "git@github.com:test/test-repo.git",
            "deploytarget_id": 1,
        }

        result = provider.create(inputs)

        assert result.id == "1"
        assert result.outs["name"] == "test-project"
        assert result.outs["id"] == 1

        mock_client.create_project.assert_called_once()

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_create_project_full(self, mock_config_class, sample_project):
        """Test creating a project with all arguments."""
        from pulumi_lagoon.project import LagoonProjectProvider

        # Setup mock
        mock_client = Mock()
        mock_client.create_project.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectProvider()

        inputs = {
            "name": "test-project",
            "git_url": "git@github.com:test/test-repo.git",
            "deploytarget_id": 1,
            "production_environment": "main",
            "branches": "^(main|develop)$",
            "pullrequests": ".*",
            "openshift_project_pattern": "${project}-${environment}",
            "auto_idle": 4,
            "storage_calc": 1,
        }

        provider.create(inputs)

        # Verify the API was called with correct arguments
        call_kwargs = mock_client.create_project.call_args[1]
        assert call_kwargs["name"] == "test-project"
        assert call_kwargs["git_url"] == "git@github.com:test/test-repo.git"
        assert call_kwargs["openshift"] == 1
        assert call_kwargs["productionEnvironment"] == "main"
        assert call_kwargs["branches"] == "^(main|develop)$"


class TestLagoonProjectProviderUpdate:
    """Tests for LagoonProjectProvider.update method."""

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_update_project_changes(self, mock_config_class, sample_project):
        """Test updating a project with changed values."""
        from pulumi_lagoon.project import LagoonProjectProvider

        # Setup mock
        updated_project = sample_project.copy()
        updated_project["branches"] = "^(main|develop|staging)$"

        mock_client = Mock()
        mock_client.update_project.return_value = updated_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectProvider()

        old_inputs = {
            "name": "test-project",
            "git_url": "git@github.com:test/test-repo.git",
            "deploytarget_id": 1,
            "branches": "^(main|develop)$",
        }

        new_inputs = {
            "name": "test-project",
            "git_url": "git@github.com:test/test-repo.git",
            "deploytarget_id": 1,
            "branches": "^(main|develop|staging)$",
        }

        result = provider.update("1", old_inputs, new_inputs)

        mock_client.update_project.assert_called_once()
        assert result.outs["branches"] == "^(main|develop|staging)$"

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_update_project_no_changes(self, mock_config_class):
        """Test update with no actual changes returns inputs."""
        from pulumi_lagoon.project import LagoonProjectProvider

        # Setup mock
        mock_client = Mock()
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectProvider()

        inputs = {
            "name": "test-project",
            "git_url": "git@github.com:test/test-repo.git",
            "deploytarget_id": 1,
        }

        result = provider.update("1", inputs, inputs)

        # Should not call update_project when nothing changed
        mock_client.update_project.assert_not_called()
        assert result.outs == inputs


class TestLagoonProjectProviderDelete:
    """Tests for LagoonProjectProvider.delete method."""

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_delete_project(self, mock_config_class):
        """Test deleting a project."""
        from pulumi_lagoon.project import LagoonProjectProvider

        # Setup mock
        mock_client = Mock()
        mock_client.delete_project.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectProvider()

        props = {
            "name": "test-project",
            "git_url": "git@github.com:test/test-repo.git",
            "deploytarget_id": 1,
        }

        provider.delete("1", props)

        mock_client.delete_project.assert_called_once_with("test-project")


class TestLagoonProjectProviderRead:
    """Tests for LagoonProjectProvider.read method."""

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_read_project_exists(self, mock_config_class, sample_project):
        """Test reading an existing project."""
        from pulumi_lagoon.project import LagoonProjectProvider

        # Setup mock
        mock_client = Mock()
        mock_client.get_project_by_id.return_value = sample_project
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectProvider()

        props = {"name": "test-project"}

        result = provider.read("1", props)

        assert result.id == "1"
        assert result.outs["name"] == "test-project"
        mock_client.get_project_by_id.assert_called_once_with(1)

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_read_project_not_found(self, mock_config_class):
        """Test reading a project that doesn't exist."""
        from pulumi_lagoon.project import LagoonProjectProvider

        # Setup mock - project not found
        mock_client = Mock()
        mock_client.get_project_by_id.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonProjectProvider()

        props = {"name": "deleted-project"}

        result = provider.read("999", props)

        assert result is None


class TestLagoonProjectArgs:
    """Tests for LagoonProjectArgs dataclass."""

    def test_args_minimal(self):
        """Test creating args with minimal required fields."""
        from pulumi_lagoon.project import LagoonProjectArgs

        args = LagoonProjectArgs(
            name="test-project",
            git_url="git@github.com:test/repo.git",
            deploytarget_id=1,
        )

        assert args.name == "test-project"
        assert args.git_url == "git@github.com:test/repo.git"
        assert args.deploytarget_id == 1
        assert args.production_environment is None
        assert args.branches is None

    def test_args_full(self):
        """Test creating args with all fields."""
        from pulumi_lagoon.project import LagoonProjectArgs

        args = LagoonProjectArgs(
            name="test-project",
            git_url="git@github.com:test/repo.git",
            deploytarget_id=1,
            production_environment="main",
            branches="^(main|develop)$",
            pullrequests=".*",
            openshift_project_pattern="${project}-${env}",
            auto_idle=4,
            storage_calc=1,
        )

        assert args.production_environment == "main"
        assert args.branches == "^(main|develop)$"
        assert args.pullrequests == ".*"
