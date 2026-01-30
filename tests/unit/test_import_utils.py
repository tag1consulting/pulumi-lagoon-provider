"""Unit tests for import utilities."""

import pytest

from pulumi_lagoon.exceptions import LagoonValidationError
from pulumi_lagoon.import_utils import ImportIdParser


class TestIsImportScenario:
    """Tests for ImportIdParser.is_import_scenario method."""

    def test_empty_props_is_import(self):
        """Test that empty props indicates import scenario."""
        assert ImportIdParser.is_import_scenario("123:main", {}, ["name", "project_id"])

    def test_none_props_is_import(self):
        """Test that None props indicates import scenario."""
        assert ImportIdParser.is_import_scenario("123:main", None, ["name", "project_id"])

    def test_missing_required_prop_is_import(self):
        """Test that missing required props indicates import scenario."""
        props = {"name": "main"}  # Missing project_id
        assert ImportIdParser.is_import_scenario("123:main", props, ["name", "project_id"])

    def test_none_required_prop_is_import(self):
        """Test that None value for required prop indicates import scenario."""
        props = {"name": "main", "project_id": None}
        assert ImportIdParser.is_import_scenario("123:main", props, ["name", "project_id"])

    def test_full_props_is_refresh(self):
        """Test that full props indicates refresh scenario."""
        props = {"name": "main", "project_id": 123}
        assert not ImportIdParser.is_import_scenario("123:main", props, ["name", "project_id"])

    def test_extra_props_is_refresh(self):
        """Test that extra props beyond required still indicates refresh."""
        props = {"name": "main", "project_id": 123, "extra": "value"}
        assert not ImportIdParser.is_import_scenario("123:main", props, ["name", "project_id"])

    def test_empty_required_list(self):
        """Test with empty required props list."""
        # If no props are required, any non-empty props dict is considered refresh
        assert ImportIdParser.is_import_scenario("123", {}, [])
        assert not ImportIdParser.is_import_scenario("123", {"anything": "value"}, [])


class TestParseEnvironmentId:
    """Tests for ImportIdParser.parse_environment_id method."""

    def test_valid_format(self):
        """Test parsing valid environment import ID."""
        project_id, env_name = ImportIdParser.parse_environment_id("123:main")
        assert project_id == 123
        assert env_name == "main"

    def test_valid_format_with_branch_name(self):
        """Test parsing environment ID with typical branch name."""
        project_id, env_name = ImportIdParser.parse_environment_id("456:develop")
        assert project_id == 456
        assert env_name == "develop"

    def test_env_name_with_special_chars(self):
        """Test that environment name can contain special characters."""
        project_id, env_name = ImportIdParser.parse_environment_id("123:feature/JIRA-123")
        assert project_id == 123
        assert env_name == "feature/JIRA-123"

    def test_env_name_with_colon(self):
        """Test that only first colon is used as separator."""
        project_id, env_name = ImportIdParser.parse_environment_id("123:name:with:colons")
        assert project_id == 123
        assert env_name == "name:with:colons"

    def test_missing_separator(self):
        """Test error when separator is missing."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_environment_id("123main")
        assert "project_id:env_name" in str(exc.value)

    def test_empty_project_id(self):
        """Test error when project ID is empty."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_environment_id(":main")
        assert "Project ID cannot be empty" in str(exc.value)

    def test_empty_env_name(self):
        """Test error when environment name is empty."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_environment_id("123:")
        assert "Environment name cannot be empty" in str(exc.value)

    def test_non_numeric_project_id(self):
        """Test error when project ID is not a number."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_environment_id("abc:main")
        assert "must be a number" in str(exc.value)

    def test_negative_project_id(self):
        """Test error when project ID is negative."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_environment_id("-1:main")
        assert "must be positive" in str(exc.value)

    def test_zero_project_id(self):
        """Test error when project ID is zero."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_environment_id("0:main")
        assert "must be positive" in str(exc.value)


class TestParseVariableId:
    """Tests for ImportIdParser.parse_variable_id method."""

    def test_environment_level_variable(self):
        """Test parsing environment-level variable import ID."""
        project_id, env_id, var_name = ImportIdParser.parse_variable_id("123:456:DATABASE_HOST")
        assert project_id == 123
        assert env_id == 456
        assert var_name == "DATABASE_HOST"

    def test_project_level_variable(self):
        """Test parsing project-level variable import ID (empty env_id)."""
        project_id, env_id, var_name = ImportIdParser.parse_variable_id("123::API_KEY")
        assert project_id == 123
        assert env_id is None
        assert var_name == "API_KEY"

    def test_var_name_with_underscores(self):
        """Test variable name with underscores."""
        project_id, env_id, var_name = ImportIdParser.parse_variable_id("1:2:MY_VAR_NAME")
        assert var_name == "MY_VAR_NAME"

    def test_var_name_with_colon(self):
        """Test that only first two colons are used as separators."""
        project_id, env_id, var_name = ImportIdParser.parse_variable_id("1:2:name:with:colons")
        assert project_id == 1
        assert env_id == 2
        assert var_name == "name:with:colons"

    def test_missing_separators(self):
        """Test error when separators are missing."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id("123DATABASE_HOST")
        assert "project_id:env_id:var_name" in str(exc.value)

    def test_only_one_separator(self):
        """Test error when only one separator is present."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id("123:DATABASE_HOST")
        assert "project_id:env_id:var_name" in str(exc.value)

    def test_empty_project_id(self):
        """Test error when project ID is empty."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id(":456:VAR")
        assert "Project ID cannot be empty" in str(exc.value)

    def test_empty_var_name(self):
        """Test error when variable name is empty."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id("123:456:")
        assert "Variable name cannot be empty" in str(exc.value)

    def test_non_numeric_project_id(self):
        """Test error when project ID is not a number."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id("abc:456:VAR")
        assert "Project ID must be a number" in str(exc.value)

    def test_non_numeric_env_id(self):
        """Test error when environment ID is not a number (and not empty)."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id("123:abc:VAR")
        assert "Environment ID must be a number or empty" in str(exc.value)

    def test_negative_project_id(self):
        """Test error when project ID is negative."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id("-1:456:VAR")
        assert "Project ID must be positive" in str(exc.value)

    def test_negative_env_id(self):
        """Test error when environment ID is negative."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id("123:-1:VAR")
        assert "Environment ID must be positive" in str(exc.value)

    def test_zero_project_id(self):
        """Test error when project ID is zero."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id("0:456:VAR")
        assert "Project ID must be positive" in str(exc.value)

    def test_zero_env_id(self):
        """Test error when environment ID is zero."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_variable_id("123:0:VAR")
        assert "Environment ID must be positive" in str(exc.value)


class TestParseDeployTargetConfigId:
    """Tests for ImportIdParser.parse_deploy_target_config_id method."""

    def test_valid_format(self):
        """Test parsing valid deploy target config import ID."""
        project_id, config_id = ImportIdParser.parse_deploy_target_config_id("123:5")
        assert project_id == 123
        assert config_id == 5

    def test_large_ids(self):
        """Test parsing with large ID values."""
        project_id, config_id = ImportIdParser.parse_deploy_target_config_id("999999:888888")
        assert project_id == 999999
        assert config_id == 888888

    def test_missing_separator(self):
        """Test error when separator is missing."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id("1235")
        assert "project_id:config_id" in str(exc.value)

    def test_empty_project_id(self):
        """Test error when project ID is empty."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id(":5")
        assert "Project ID cannot be empty" in str(exc.value)

    def test_empty_config_id(self):
        """Test error when config ID is empty."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id("123:")
        assert "Config ID cannot be empty" in str(exc.value)

    def test_non_numeric_project_id(self):
        """Test error when project ID is not a number."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id("abc:5")
        assert "Project ID must be a number" in str(exc.value)

    def test_non_numeric_config_id(self):
        """Test error when config ID is not a number."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id("123:abc")
        assert "Config ID must be a number" in str(exc.value)

    def test_negative_project_id(self):
        """Test error when project ID is negative."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id("-1:5")
        assert "Project ID must be positive" in str(exc.value)

    def test_negative_config_id(self):
        """Test error when config ID is negative."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id("123:-5")
        assert "Config ID must be positive" in str(exc.value)

    def test_zero_project_id(self):
        """Test error when project ID is zero."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id("0:5")
        assert "Project ID must be positive" in str(exc.value)

    def test_zero_config_id(self):
        """Test error when config ID is zero."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id("123:0")
        assert "Config ID must be positive" in str(exc.value)

    def test_extra_colons_ignored(self):
        """Test that extra colons are included in config_id (which will fail)."""
        with pytest.raises(LagoonValidationError) as exc:
            ImportIdParser.parse_deploy_target_config_id("123:5:extra")
        assert "Config ID must be a number" in str(exc.value)
