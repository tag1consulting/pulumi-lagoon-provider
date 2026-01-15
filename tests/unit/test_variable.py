"""Unit tests for LagoonVariable provider."""

from unittest.mock import Mock, patch


class TestLagoonVariableProviderCreate:
    """Tests for LagoonVariableProvider.create method."""

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_create_project_variable(self, mock_config_class, sample_variable):
        """Test creating a project-level variable."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        mock_client = Mock()
        mock_client.add_env_variable.return_value = sample_variable
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        inputs = {
            "name": "DATABASE_HOST",
            "value": "mysql.example.com",
            "project_id": 1,
            "scope": "runtime",
        }

        result = provider.create(inputs)

        assert result.outs["name"] == "DATABASE_HOST"
        assert result.outs["value"] == "mysql.example.com"

        # Verify project-level variable (no environment)
        call_kwargs = mock_client.add_env_variable.call_args[1]
        assert call_kwargs["project"] == 1
        assert "environment" not in call_kwargs

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_create_environment_variable(self, mock_config_class, sample_variable):
        """Test creating an environment-level variable."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        env_variable = sample_variable.copy()
        env_variable["environment"] = {"id": 1, "name": "main"}

        mock_client = Mock()
        mock_client.add_env_variable.return_value = env_variable
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        inputs = {
            "name": "DATABASE_HOST",
            "value": "mysql.example.com",
            "project_id": 1,
            "environment_id": 1,
            "scope": "runtime",
        }

        provider.create(inputs)

        # Verify environment-level variable
        call_kwargs = mock_client.add_env_variable.call_args[1]
        assert call_kwargs["environment"] == 1

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_create_variable_handles_string_ids(
        self, mock_config_class, sample_variable
    ):
        """Test that string IDs are converted to int."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        mock_client = Mock()
        mock_client.add_env_variable.return_value = sample_variable
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        inputs = {
            "name": "TEST",
            "value": "value",
            "project_id": "1",  # String
            "environment_id": "2",  # String
            "scope": "runtime",
        }

        provider.create(inputs)

        call_kwargs = mock_client.add_env_variable.call_args[1]
        assert call_kwargs["project"] == 1
        assert call_kwargs["environment"] == 2

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_create_variable_id_generation(self, mock_config_class, sample_variable):
        """Test that variable IDs are generated correctly."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        mock_client = Mock()
        mock_client.add_env_variable.return_value = sample_variable
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        # Project-level variable
        inputs = {
            "name": "VAR_NAME",
            "value": "value",
            "project_id": 1,
            "scope": "runtime",
        }

        result = provider.create(inputs)
        assert result.id == "p1_VAR_NAME"

        # Environment-level variable
        inputs["environment_id"] = 2
        result = provider.create(inputs)
        assert result.id == "p1e2_VAR_NAME"


class TestLagoonVariableProviderUpdate:
    """Tests for LagoonVariableProvider.update method."""

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_update_variable_value(self, mock_config_class, sample_variable):
        """Test updating a variable's value."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        updated_variable = sample_variable.copy()
        updated_variable["value"] = "new-value"

        mock_client = Mock()
        mock_client.delete_env_variable.return_value = "success"
        mock_client.add_env_variable.return_value = updated_variable
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        old_inputs = {
            "name": "DATABASE_HOST",
            "value": "mysql.example.com",
            "project_id": 1,
            "scope": "runtime",
        }

        new_inputs = {
            "name": "DATABASE_HOST",
            "value": "new-mysql.example.com",
            "project_id": 1,
            "scope": "runtime",
        }

        provider.update("p1_DATABASE_HOST", old_inputs, new_inputs)

        # Should delete old and create new
        mock_client.delete_env_variable.assert_called_once()
        mock_client.add_env_variable.assert_called_once()

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_update_handles_delete_failure(self, mock_config_class, sample_variable):
        """Test update continues if delete fails."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        mock_client = Mock()
        mock_client.delete_env_variable.side_effect = Exception("Variable not found")
        mock_client.add_env_variable.return_value = sample_variable
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        old_inputs = {
            "name": "VAR",
            "value": "old",
            "project_id": 1,
            "scope": "runtime",
        }

        new_inputs = {
            "name": "VAR",
            "value": "new",
            "project_id": 1,
            "scope": "runtime",
        }

        # Should not raise, should continue with create
        provider.update("p1_VAR", old_inputs, new_inputs)

        mock_client.add_env_variable.assert_called_once()


class TestLagoonVariableProviderDelete:
    """Tests for LagoonVariableProvider.delete method."""

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_delete_project_variable(self, mock_config_class):
        """Test deleting a project-level variable."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        mock_client = Mock()
        mock_client.delete_env_variable.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        props = {
            "name": "DATABASE_HOST",
            "project_id": 1,
        }

        provider.delete("p1_DATABASE_HOST", props)

        call_kwargs = mock_client.delete_env_variable.call_args[1]
        assert call_kwargs["name"] == "DATABASE_HOST"
        assert call_kwargs["project"] == 1
        assert "environment" not in call_kwargs

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_delete_environment_variable(self, mock_config_class):
        """Test deleting an environment-level variable."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        mock_client = Mock()
        mock_client.delete_env_variable.return_value = "success"
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        props = {
            "name": "DATABASE_HOST",
            "project_id": 1,
            "environment_id": 2,
        }

        provider.delete("p1e2_DATABASE_HOST", props)

        call_kwargs = mock_client.delete_env_variable.call_args[1]
        assert call_kwargs["environment"] == 2


class TestLagoonVariableProviderRead:
    """Tests for LagoonVariableProvider.read method."""

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_read_variable_exists(self, mock_config_class, sample_variable):
        """Test reading an existing variable."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        mock_client = Mock()
        mock_client.get_env_variable_by_name.return_value = sample_variable
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        props = {
            "name": "DATABASE_HOST",
            "project_id": 1,
        }

        result = provider.read("p1_DATABASE_HOST", props)

        assert result.outs["name"] == "DATABASE_HOST"
        mock_client.get_env_variable_by_name.assert_called_once_with(
            name="DATABASE_HOST", project=1, environment=None
        )

    @patch("pulumi_lagoon.variable.LagoonConfig")
    def test_read_variable_not_found(self, mock_config_class):
        """Test reading a variable that doesn't exist."""
        from pulumi_lagoon.variable import LagoonVariableProvider

        mock_client = Mock()
        mock_client.get_env_variable_by_name.return_value = None
        mock_config = Mock()
        mock_config.get_client.return_value = mock_client
        mock_config_class.return_value = mock_config

        provider = LagoonVariableProvider()

        props = {
            "name": "DELETED_VAR",
            "project_id": 1,
        }

        result = provider.read("p1_DELETED_VAR", props)

        assert result is None


class TestLagoonVariableArgs:
    """Tests for LagoonVariableArgs dataclass."""

    def test_args_project_level(self):
        """Test creating args for project-level variable."""
        from pulumi_lagoon.variable import LagoonVariableArgs

        args = LagoonVariableArgs(
            name="DATABASE_HOST",
            value="mysql.example.com",
            project_id=1,
            scope="runtime",
        )

        assert args.name == "DATABASE_HOST"
        assert args.value == "mysql.example.com"
        assert args.project_id == 1
        assert args.scope == "runtime"
        assert args.environment_id is None

    def test_args_environment_level(self):
        """Test creating args for environment-level variable."""
        from pulumi_lagoon.variable import LagoonVariableArgs

        args = LagoonVariableArgs(
            name="DATABASE_HOST",
            value="mysql.example.com",
            project_id=1,
            scope="runtime",
            environment_id=2,
        )

        assert args.environment_id == 2

    def test_args_different_scopes(self):
        """Test different variable scopes."""
        from pulumi_lagoon.variable import LagoonVariableArgs

        for scope in ["build", "runtime", "global", "container_registry"]:
            args = LagoonVariableArgs(
                name="TEST",
                value="value",
                project_id=1,
                scope=scope,
            )
            assert args.scope == scope
