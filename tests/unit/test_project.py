"""Unit tests for LagoonProject provider."""

from unittest.mock import Mock, patch

import pytest

from pulumi_lagoon.exceptions import LagoonValidationError


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

        # Verify correct calling convention: project_id as positional arg, not in kwargs
        call_args, call_kwargs = mock_client.update_project.call_args
        assert call_args == (1,), "Project ID should be first positional argument"
        assert "id" not in call_kwargs, "ID should not be in kwargs"

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


class TestLagoonProjectProviderValidation:
    """Tests for input validation in LagoonProjectProvider."""

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_create_invalid_project_name_uppercase(self, mock_config_class):
        """Test that uppercase project names are rejected."""
        from pulumi_lagoon.project import LagoonProjectProvider

        provider = LagoonProjectProvider()

        inputs = {
            "name": "Invalid-Name",
            "git_url": "git@github.com:test/repo.git",
            "deploytarget_id": 1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "name" in str(exc.value).lower()

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_create_invalid_git_url(self, mock_config_class):
        """Test that invalid git URLs are rejected."""
        from pulumi_lagoon.project import LagoonProjectProvider

        provider = LagoonProjectProvider()

        inputs = {
            "name": "valid-name",
            "git_url": "not-a-valid-url",
            "deploytarget_id": 1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "git_url" in str(exc.value)

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_create_invalid_deploytarget_id(self, mock_config_class):
        """Test that invalid deploytarget_id is rejected."""
        from pulumi_lagoon.project import LagoonProjectProvider

        provider = LagoonProjectProvider()

        inputs = {
            "name": "valid-name",
            "git_url": "git@github.com:test/repo.git",
            "deploytarget_id": -1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "deploytarget_id" in str(exc.value)

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_create_invalid_branches_regex(self, mock_config_class):
        """Test that invalid branch regex patterns are rejected."""
        from pulumi_lagoon.project import LagoonProjectProvider

        provider = LagoonProjectProvider()

        inputs = {
            "name": "valid-name",
            "git_url": "git@github.com:test/repo.git",
            "deploytarget_id": 1,
            "branches": "[invalid-regex",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "branches" in str(exc.value)

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_create_invalid_pullrequests_regex(self, mock_config_class):
        """Test that invalid pullrequests regex patterns are rejected."""
        from pulumi_lagoon.project import LagoonProjectProvider

        provider = LagoonProjectProvider()

        inputs = {
            "name": "valid-name",
            "git_url": "git@github.com:test/repo.git",
            "deploytarget_id": 1,
            "pullrequests": "(unbalanced",
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.create(inputs)
        assert "pullrequests" in str(exc.value)

    @patch("pulumi_lagoon.project.LagoonConfig")
    def test_update_invalid_git_url(self, mock_config_class):
        """Test that invalid git URL in update is rejected."""
        from pulumi_lagoon.project import LagoonProjectProvider

        provider = LagoonProjectProvider()

        old_inputs = {
            "name": "valid-name",
            "git_url": "git@github.com:test/repo.git",
            "deploytarget_id": 1,
        }
        new_inputs = {
            "name": "valid-name",
            "git_url": "invalid-url",
            "deploytarget_id": 1,
        }

        with pytest.raises(LagoonValidationError) as exc:
            provider.update("1", old_inputs, new_inputs)
        assert "git_url" in str(exc.value)


class TestLagoonProjectResourceInit:
    """Tests for LagoonProject resource initialization."""

    def test_resource_init_constructs_inputs_correctly(self):
        """Test that LagoonProject.__init__ constructs inputs dict from args."""
        from pulumi_lagoon.project import LagoonProject, LagoonProjectArgs

        # Mock the parent class __init__ to capture inputs
        with patch("pulumi.dynamic.Resource.__init__") as mock_init:
            mock_init.return_value = None

            args = LagoonProjectArgs(
                name="test-project",
                git_url="git@github.com:test/repo.git",
                deploytarget_id=1,
                production_environment="main",
                branches="^(main|develop)$",
                pullrequests=".*",
                openshift_project_pattern="${project}-${env}",
                auto_idle=1,
                storage_calc=1,
                api_url="https://api.lagoon.example.com/graphql",
                api_token="test-token",
                jwt_secret=None,
            )

            LagoonProject("test-resource", args)

            # Verify parent __init__ was called
            mock_init.assert_called_once()
            call_args = mock_init.call_args

            # Check the inputs dict (second positional arg after provider)
            inputs = call_args[0][2]  # provider, name, inputs, opts

            assert inputs["name"] == "test-project"
            assert inputs["git_url"] == "git@github.com:test/repo.git"
            assert inputs["deploytarget_id"] == 1
            assert inputs["production_environment"] == "main"
            assert inputs["branches"] == "^(main|develop)$"
            assert inputs["pullrequests"] == ".*"
            assert inputs["openshift_project_pattern"] == "${project}-${env}"
            assert inputs["auto_idle"] == 1
            assert inputs["storage_calc"] == 1
            assert inputs["api_url"] == "https://api.lagoon.example.com/graphql"
            assert inputs["api_token"] == "test-token"
            assert inputs["jwt_secret"] is None
            assert inputs["id"] is None  # Output placeholder
            assert inputs["created"] is None  # Output placeholder

    def test_resource_init_with_minimal_args(self):
        """Test that LagoonProject.__init__ works with minimal required args."""
        from pulumi_lagoon.project import LagoonProject, LagoonProjectArgs

        with patch("pulumi.dynamic.Resource.__init__") as mock_init:
            mock_init.return_value = None

            args = LagoonProjectArgs(
                name="minimal-project",
                git_url="git@github.com:test/minimal.git",
                deploytarget_id=1,
            )

            LagoonProject("minimal-resource", args)

            mock_init.assert_called_once()
            call_args = mock_init.call_args
            inputs = call_args[0][2]

            assert inputs["name"] == "minimal-project"
            assert inputs["git_url"] == "git@github.com:test/minimal.git"
            assert inputs["deploytarget_id"] == 1
            # Optional fields should be None
            assert inputs["production_environment"] is None
            assert inputs["branches"] is None
            assert inputs["pullrequests"] is None

    def test_resource_init_passes_opts(self):
        """Test that LagoonProject.__init__ passes ResourceOptions correctly."""
        import pulumi

        from pulumi_lagoon.project import LagoonProject, LagoonProjectArgs

        with patch("pulumi.dynamic.Resource.__init__") as mock_init:
            mock_init.return_value = None

            args = LagoonProjectArgs(
                name="test-project",
                git_url="git@github.com:test/repo.git",
                deploytarget_id=1,
            )

            opts = pulumi.ResourceOptions(protect=True, retain_on_delete=True)
            LagoonProject("test-resource", args, opts=opts)

            mock_init.assert_called_once()
            call_args = mock_init.call_args

            # opts is the 4th positional argument
            passed_opts = call_args[0][3]
            assert passed_opts is opts


class TestLagoonProjectProviderClientConfig:
    """Tests for LagoonProjectProvider client configuration."""

    def test_get_client_with_api_token(self, sample_project):
        """Test creating client with explicit api_url and api_token."""
        from pulumi_lagoon.project import LagoonProjectProvider

        with patch("pulumi_lagoon.client.LagoonClient") as mock_client_class:
            mock_client = Mock()
            mock_client.create_project.return_value = sample_project
            mock_client_class.return_value = mock_client

            provider = LagoonProjectProvider()

            inputs = {
                "name": "test-project",
                "git_url": "git@github.com:test/repo.git",
                "deploytarget_id": 1,
                "api_url": "https://api.lagoon.example.com/graphql",
                "api_token": "test-bearer-token",
            }

            provider.create(inputs)

            # Verify LagoonClient was created with correct args
            mock_client_class.assert_called_once_with(
                "https://api.lagoon.example.com/graphql",
                "test-bearer-token",
            )

    def test_get_client_with_jwt_secret(self, sample_project):
        """Test creating client with jwt_secret for token generation."""
        from pulumi_lagoon.project import LagoonProjectProvider

        with patch("pulumi_lagoon.client.LagoonClient") as mock_client_class:
            mock_client = Mock()
            mock_client.create_project.return_value = sample_project
            mock_client_class.return_value = mock_client

            provider = LagoonProjectProvider()

            inputs = {
                "name": "test-project",
                "git_url": "git@github.com:test/repo.git",
                "deploytarget_id": 1,
                "api_url": "https://api.lagoon.example.com/graphql",
                "jwt_secret": "test-jwt-secret-key",
            }

            provider.create(inputs)

            # Verify LagoonClient was called with generated token
            mock_client_class.assert_called_once()
            call_args = mock_client_class.call_args
            assert call_args[0][0] == "https://api.lagoon.example.com/graphql"
            # Token should be a JWT string
            assert isinstance(call_args[0][1], str)


class TestLagoonProjectProviderJWTGeneration:
    """Tests for JWT token generation in LagoonProjectProvider."""

    def test_generate_admin_token_valid(self):
        """Test that _generate_admin_token produces a valid JWT."""
        import jwt as pyjwt

        from pulumi_lagoon.project import LagoonProjectProvider

        provider = LagoonProjectProvider()
        secret = "test-secret-key"

        token = provider._generate_admin_token(secret)

        # Verify it's a valid JWT that can be decoded
        decoded = pyjwt.decode(token, secret, algorithms=["HS256"], options={"verify_aud": False})

        assert decoded["role"] == "admin"
        assert decoded["iss"] == "lagoon-api"
        assert decoded["sub"] == "lagoonadmin"
