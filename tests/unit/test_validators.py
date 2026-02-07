"""Unit tests for input validation utilities."""

import pytest

from pulumi_lagoon.exceptions import LagoonValidationError
from pulumi_lagoon.validators import (
    VALID_CLOUD_PROVIDERS,
    VALID_DEPLOY_TYPES,
    VALID_ENVIRONMENT_TYPES,
    VALID_SCOPES,
    validate_cloud_provider,
    validate_console_url,
    validate_deploy_target_name,
    validate_deploy_type,
    validate_enum,
    validate_environment_name,
    validate_environment_type,
    validate_git_url,
    validate_positive_int,
    validate_project_name,
    validate_regex_pattern,
    validate_required,
    validate_scope,
    validate_ssh_host,
    validate_ssh_port,
    validate_variable_name,
)


class TestValidateRequired:
    """Tests for validate_required."""

    def test_valid_string(self):
        """Test that non-empty strings pass validation."""
        # Should not raise
        validate_required("test", "field")

    def test_valid_int(self):
        """Test that integers pass validation."""
        validate_required(123, "field")

    def test_valid_zero(self):
        """Test that zero passes validation."""
        validate_required(0, "field")

    def test_none_raises_error(self):
        """Test that None raises LagoonValidationError."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_required(None, "my_field")
        assert "my_field" in str(exc.value)
        assert "missing" in str(exc.value).lower()

    def test_empty_string_raises_error(self):
        """Test that empty string raises LagoonValidationError."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_required("", "my_field")
        assert "my_field" in str(exc.value)
        assert "empty" in str(exc.value).lower()

    def test_whitespace_only_raises_error(self):
        """Test that whitespace-only string raises LagoonValidationError."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_required("   ", "my_field")
        assert "my_field" in str(exc.value)


class TestValidateProjectName:
    """Tests for validate_project_name."""

    def test_valid_simple_name(self):
        """Test valid simple project name."""
        validate_project_name("my-project")

    def test_valid_single_char(self):
        """Test valid single character name."""
        validate_project_name("a")

    def test_valid_with_numbers(self):
        """Test valid name with numbers."""
        validate_project_name("project-123")

    def test_valid_starts_with_letter(self):
        """Test valid name starting with letter."""
        validate_project_name("z-test-project")

    def test_invalid_starts_with_number(self):
        """Test that name starting with number is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_project_name("123-project")
        assert "name" in str(exc.value)

    def test_invalid_starts_with_hyphen(self):
        """Test that name starting with hyphen is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_project_name("-project")

    def test_invalid_ends_with_hyphen(self):
        """Test that name ending with hyphen is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_project_name("project-")

    def test_invalid_uppercase(self):
        """Test that uppercase letters are rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_project_name("My-Project")
        assert "lowercase" in str(exc.value)

    def test_invalid_underscore(self):
        """Test that underscores are rejected."""
        with pytest.raises(LagoonValidationError):
            validate_project_name("my_project")

    def test_invalid_too_long(self):
        """Test that names over 58 chars are rejected."""
        long_name = "a" * 59
        with pytest.raises(LagoonValidationError) as exc:
            validate_project_name(long_name)
        assert "58" in str(exc.value)

    def test_valid_max_length(self):
        """Test that 58 char name is valid."""
        name = "a" * 58
        validate_project_name(name)

    def test_none_raises_error(self):
        """Test that None raises validation error."""
        with pytest.raises(LagoonValidationError):
            validate_project_name(None)


class TestValidateGitUrl:
    """Tests for validate_git_url."""

    def test_valid_ssh_github(self):
        """Test valid GitHub SSH URL."""
        validate_git_url("git@github.com:org/repo.git")

    def test_valid_ssh_gitlab(self):
        """Test valid GitLab SSH URL."""
        validate_git_url("git@gitlab.com:org/repo.git")

    def test_valid_ssh_without_git_extension(self):
        """Test valid SSH URL without .git extension."""
        validate_git_url("git@github.com:org/repo")

    def test_valid_https_github(self):
        """Test valid GitHub HTTPS URL."""
        validate_git_url("https://github.com/org/repo.git")

    def test_valid_https_without_git_extension(self):
        """Test valid HTTPS URL without .git extension."""
        validate_git_url("https://github.com/org/repo")

    def test_valid_http(self):
        """Test valid HTTP URL (allowed but not recommended)."""
        validate_git_url("http://github.com/org/repo.git")

    def test_invalid_no_protocol(self):
        """Test that URL without protocol is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_git_url("github.com/org/repo.git")
        assert "git_url" in str(exc.value)

    def test_invalid_ftp_protocol(self):
        """Test that FTP protocol is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_git_url("ftp://github.com/org/repo.git")

    def test_invalid_just_host(self):
        """Test that just hostname is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_git_url("https://github.com")

    def test_none_raises_error(self):
        """Test that None raises validation error."""
        with pytest.raises(LagoonValidationError):
            validate_git_url(None)


class TestValidatePositiveInt:
    """Tests for validate_positive_int."""

    def test_valid_positive_int(self):
        """Test valid positive integer."""
        result = validate_positive_int(5, "field")
        assert result == 5

    def test_valid_string_int(self):
        """Test valid integer as string."""
        result = validate_positive_int("123", "field")
        assert result == 123

    def test_zero_rejected_by_default(self):
        """Test that zero is rejected by default."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_positive_int(0, "my_field")
        assert "positive" in str(exc.value).lower()

    def test_zero_allowed_with_flag(self):
        """Test that zero is allowed when allow_zero=True."""
        result = validate_positive_int(0, "field", allow_zero=True)
        assert result == 0

    def test_negative_rejected(self):
        """Test that negative numbers are rejected."""
        with pytest.raises(LagoonValidationError):
            validate_positive_int(-5, "field")

    def test_negative_rejected_with_allow_zero(self):
        """Test that negative numbers are rejected even with allow_zero."""
        with pytest.raises(LagoonValidationError):
            validate_positive_int(-1, "field", allow_zero=True)

    def test_invalid_string_rejected(self):
        """Test that non-numeric strings are rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_positive_int("abc", "field")
        assert "integer" in str(exc.value).lower()

    def test_none_rejected(self):
        """Test that None is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_positive_int(None, "field")

    def test_float_converted(self):
        """Test that floats are converted to int."""
        result = validate_positive_int(5.7, "field")
        assert result == 5


class TestValidateEnum:
    """Tests for validate_enum."""

    def test_valid_value(self):
        """Test valid enum value."""
        result = validate_enum("branch", "field", {"branch", "pullrequest"})
        assert result == "branch"

    def test_case_insensitive(self):
        """Test enum validation is case-insensitive."""
        result = validate_enum("BRANCH", "field", {"branch", "pullrequest"})
        assert result == "branch"

    def test_strips_whitespace(self):
        """Test enum validation strips whitespace."""
        result = validate_enum("  branch  ", "field", {"branch", "pullrequest"})
        assert result == "branch"

    def test_invalid_value_rejected(self):
        """Test invalid enum value is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_enum("invalid", "my_field", {"branch", "pullrequest"})
        assert "my_field" in str(exc.value)
        assert "branch" in str(exc.value)
        assert "pullrequest" in str(exc.value)

    def test_none_rejected(self):
        """Test that None is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_enum(None, "field", {"a", "b"})


class TestValidateDeployType:
    """Tests for validate_deploy_type."""

    def test_valid_branch(self):
        """Test valid 'branch' deploy type."""
        result = validate_deploy_type("branch")
        assert result == "branch"

    def test_valid_pullrequest(self):
        """Test valid 'pullrequest' deploy type."""
        result = validate_deploy_type("pullrequest")
        assert result == "pullrequest"

    def test_case_insensitive(self):
        """Test case-insensitive validation."""
        result = validate_deploy_type("BRANCH")
        assert result == "branch"

    def test_invalid_rejected(self):
        """Test invalid deploy type is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_deploy_type("invalid")
        assert "branch" in str(exc.value)
        assert "pullrequest" in str(exc.value)


class TestValidateEnvironmentType:
    """Tests for validate_environment_type."""

    def test_valid_production(self):
        """Test valid 'production' environment type."""
        result = validate_environment_type("production")
        assert result == "production"

    def test_valid_development(self):
        """Test valid 'development' environment type."""
        result = validate_environment_type("development")
        assert result == "development"

    def test_valid_standby(self):
        """Test valid 'standby' environment type."""
        result = validate_environment_type("standby")
        assert result == "standby"

    def test_invalid_rejected(self):
        """Test invalid environment type is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_environment_type("staging")
        assert "production" in str(exc.value)
        assert "development" in str(exc.value)


class TestValidateScope:
    """Tests for validate_scope."""

    def test_valid_build(self):
        """Test valid 'build' scope."""
        result = validate_scope("build")
        assert result == "build"

    def test_valid_runtime(self):
        """Test valid 'runtime' scope."""
        result = validate_scope("runtime")
        assert result == "runtime"

    def test_valid_global(self):
        """Test valid 'global' scope."""
        result = validate_scope("global")
        assert result == "global"

    def test_valid_container_registry(self):
        """Test valid 'container_registry' scope."""
        result = validate_scope("container_registry")
        assert result == "container_registry"

    def test_valid_internal_container_registry(self):
        """Test valid 'internal_container_registry' scope."""
        result = validate_scope("internal_container_registry")
        assert result == "internal_container_registry"

    def test_invalid_rejected(self):
        """Test invalid scope is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_scope("invalid_scope")


class TestValidateRegexPattern:
    """Tests for validate_regex_pattern."""

    def test_valid_simple_pattern(self):
        """Test valid simple regex pattern."""
        validate_regex_pattern("^main$", "field")

    def test_valid_complex_pattern(self):
        """Test valid complex regex pattern."""
        validate_regex_pattern("^(main|develop|feature/.*)$", "field")

    def test_none_allowed(self):
        """Test that None is allowed (optional field)."""
        validate_regex_pattern(None, "field")

    def test_invalid_pattern_rejected(self):
        """Test invalid regex pattern is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_regex_pattern("[invalid", "branches")
        assert "branches" in str(exc.value)
        assert "regex" in str(exc.value).lower()

    def test_unbalanced_parentheses_rejected(self):
        """Test unbalanced parentheses are rejected."""
        with pytest.raises(LagoonValidationError):
            validate_regex_pattern("^(main$", "field")


class TestValidateVariableName:
    """Tests for validate_variable_name."""

    def test_valid_uppercase(self):
        """Test valid uppercase variable name."""
        validate_variable_name("MY_VAR")

    def test_valid_lowercase(self):
        """Test valid lowercase variable name."""
        validate_variable_name("my_var")

    def test_valid_mixed_case(self):
        """Test valid mixed case variable name."""
        validate_variable_name("My_Var_123")

    def test_valid_starts_with_underscore(self):
        """Test valid name starting with underscore."""
        validate_variable_name("_private")

    def test_valid_single_char(self):
        """Test valid single character name."""
        validate_variable_name("x")

    def test_invalid_starts_with_number(self):
        """Test that name starting with number is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_variable_name("123_VAR")
        assert "name" in str(exc.value)

    def test_invalid_hyphen(self):
        """Test that hyphens are rejected."""
        with pytest.raises(LagoonValidationError):
            validate_variable_name("MY-VAR")

    def test_invalid_space(self):
        """Test that spaces are rejected."""
        with pytest.raises(LagoonValidationError):
            validate_variable_name("MY VAR")

    def test_none_rejected(self):
        """Test that None is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_variable_name(None)


class TestValidateEnvironmentName:
    """Tests for validate_environment_name."""

    def test_valid_simple_name(self):
        """Test valid simple environment name."""
        validate_environment_name("main")

    def test_valid_with_numbers(self):
        """Test valid name with numbers."""
        validate_environment_name("feature123")

    def test_valid_branch_style(self):
        """Test valid branch-style name."""
        validate_environment_name("feature/my-feature")

    def test_valid_with_dots(self):
        """Test valid name with dots."""
        validate_environment_name("v1.2.3")

    def test_valid_pr_style(self):
        """Test valid PR-style name."""
        validate_environment_name("PR-123")

    def test_valid_single_char(self):
        """Test valid single character name."""
        validate_environment_name("a")

    def test_invalid_starts_with_hyphen(self):
        """Test that name starting with hyphen is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_environment_name("-main")

    def test_invalid_ends_with_slash(self):
        """Test that name ending with slash is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_environment_name("feature/")

    def test_invalid_too_long(self):
        """Test that names over 63 chars are rejected."""
        long_name = "a" * 64
        with pytest.raises(LagoonValidationError) as exc:
            validate_environment_name(long_name)
        assert "63" in str(exc.value)

    def test_valid_max_length(self):
        """Test that 63 char name is valid."""
        name = "a" * 63
        validate_environment_name(name)

    def test_none_rejected(self):
        """Test that None is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_environment_name(None)


class TestValidEnumSets:
    """Tests to ensure enum sets are properly defined."""

    def test_valid_deploy_types(self):
        """Test VALID_DEPLOY_TYPES contains expected values."""
        assert "branch" in VALID_DEPLOY_TYPES
        assert "pullrequest" in VALID_DEPLOY_TYPES
        assert len(VALID_DEPLOY_TYPES) == 2

    def test_valid_environment_types(self):
        """Test VALID_ENVIRONMENT_TYPES contains expected values."""
        assert "production" in VALID_ENVIRONMENT_TYPES
        assert "development" in VALID_ENVIRONMENT_TYPES
        assert "standby" in VALID_ENVIRONMENT_TYPES
        assert len(VALID_ENVIRONMENT_TYPES) == 3

    def test_valid_scopes(self):
        """Test VALID_SCOPES contains expected values."""
        assert "build" in VALID_SCOPES
        assert "runtime" in VALID_SCOPES
        assert "global" in VALID_SCOPES
        assert "container_registry" in VALID_SCOPES
        assert "internal_container_registry" in VALID_SCOPES
        assert len(VALID_SCOPES) == 5


class TestExceptionAttributes:
    """Tests for exception attributes and error message quality."""

    def test_validation_error_has_field(self):
        """Test that LagoonValidationError has field attribute."""
        try:
            validate_project_name("Invalid-Name")
        except LagoonValidationError as e:
            assert e.field == "name"

    def test_validation_error_has_value(self):
        """Test that LagoonValidationError has value attribute."""
        try:
            validate_project_name("Invalid-Name")
        except LagoonValidationError as e:
            assert e.value == "Invalid-Name"

    def test_validation_error_has_suggestion(self):
        """Test that LagoonValidationError has suggestion attribute."""
        try:
            validate_project_name("Invalid-Name")
        except LagoonValidationError as e:
            assert e.suggestion is not None
            assert len(e.suggestion) > 0


class TestValidateDeployTargetName:
    """Tests for validate_deploy_target_name."""

    def test_valid_simple_name(self):
        """Test valid simple deploy target name."""
        validate_deploy_target_name("production")

    def test_valid_with_hyphens(self):
        """Test valid name with hyphens."""
        validate_deploy_target_name("prod-cluster-1")

    def test_valid_starts_with_number(self):
        """Test valid name starting with number."""
        validate_deploy_target_name("1-prod-cluster")

    def test_valid_single_char(self):
        """Test valid single character name."""
        validate_deploy_target_name("a")

    def test_valid_single_number(self):
        """Test valid single number name."""
        validate_deploy_target_name("1")

    def test_invalid_uppercase(self):
        """Test that uppercase letters are rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_deploy_target_name("Prod-Cluster")
        assert "name" in str(exc.value)

    def test_invalid_starts_with_hyphen(self):
        """Test that name starting with hyphen is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_deploy_target_name("-production")

    def test_invalid_ends_with_hyphen(self):
        """Test that name ending with hyphen is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_deploy_target_name("production-")

    def test_invalid_underscore(self):
        """Test that underscores are rejected."""
        with pytest.raises(LagoonValidationError):
            validate_deploy_target_name("prod_cluster")

    def test_invalid_too_long(self):
        """Test that names over 63 chars are rejected."""
        long_name = "a" * 64
        with pytest.raises(LagoonValidationError) as exc:
            validate_deploy_target_name(long_name)
        assert "63" in str(exc.value)

    def test_valid_max_length(self):
        """Test that 63 char name is valid."""
        name = "a" * 63
        validate_deploy_target_name(name)

    def test_none_raises_error(self):
        """Test that None raises validation error."""
        with pytest.raises(LagoonValidationError):
            validate_deploy_target_name(None)


class TestValidateConsoleUrl:
    """Tests for validate_console_url."""

    def test_valid_https_url(self):
        """Test valid HTTPS URL."""
        validate_console_url("https://kubernetes.example.com")

    def test_valid_https_with_port(self):
        """Test valid HTTPS URL with port."""
        validate_console_url("https://kubernetes.example.com:6443")

    def test_valid_https_with_path(self):
        """Test valid HTTPS URL with path."""
        validate_console_url("https://kubernetes.example.com/api/v1")

    def test_valid_http_localhost(self):
        """Test valid HTTP URL for development."""
        validate_console_url("http://localhost:8080")

    def test_valid_http_with_ip(self):
        """Test valid HTTP URL with IP address."""
        validate_console_url("http://192.168.1.100:6443")

    def test_invalid_no_protocol(self):
        """Test that URL without protocol is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_console_url("kubernetes.example.com")
        assert "console_url" in str(exc.value)

    def test_invalid_ftp_protocol(self):
        """Test that FTP protocol is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_console_url("ftp://kubernetes.example.com")

    def test_invalid_just_protocol(self):
        """Test that just protocol is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_console_url("https://")

    def test_none_raises_error(self):
        """Test that None raises validation error."""
        with pytest.raises(LagoonValidationError):
            validate_console_url(None)

    def test_empty_raises_error(self):
        """Test that empty string raises validation error."""
        with pytest.raises(LagoonValidationError):
            validate_console_url("")


class TestValidateCloudProvider:
    """Tests for validate_cloud_provider."""

    def test_valid_kind(self):
        """Test valid 'kind' cloud provider."""
        result = validate_cloud_provider("kind")
        assert result == "kind"

    def test_valid_aws(self):
        """Test valid 'aws' cloud provider."""
        result = validate_cloud_provider("aws")
        assert result == "aws"

    def test_valid_gcp(self):
        """Test valid 'gcp' cloud provider."""
        result = validate_cloud_provider("gcp")
        assert result == "gcp"

    def test_valid_azure(self):
        """Test valid 'azure' cloud provider."""
        result = validate_cloud_provider("azure")
        assert result == "azure"

    def test_case_insensitive(self):
        """Test case-insensitive validation."""
        result = validate_cloud_provider("AWS")
        assert result == "aws"

    def test_invalid_rejected(self):
        """Test invalid cloud provider is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_cloud_provider("invalid_provider")
        assert "cloud_provider" in str(exc.value)

    def test_none_rejected(self):
        """Test that None is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_cloud_provider(None)


class TestValidateSshPort:
    """Tests for validate_ssh_port."""

    def test_valid_standard_port(self):
        """Test valid standard SSH port."""
        result = validate_ssh_port(22)
        assert result == 22

    def test_valid_custom_port(self):
        """Test valid custom SSH port."""
        result = validate_ssh_port(2222)
        assert result == 2222

    def test_valid_string_port(self):
        """Test valid port as string."""
        result = validate_ssh_port("22")
        assert result == 22

    def test_valid_min_port(self):
        """Test minimum valid port."""
        result = validate_ssh_port(1)
        assert result == 1

    def test_valid_max_port(self):
        """Test maximum valid port."""
        result = validate_ssh_port(65535)
        assert result == 65535

    def test_invalid_zero(self):
        """Test that zero port is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_ssh_port(0)
        assert "ssh_port" in str(exc.value)

    def test_invalid_negative(self):
        """Test that negative port is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_ssh_port(-1)

    def test_invalid_too_high(self):
        """Test that port over 65535 is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_ssh_port(65536)
        assert "65535" in str(exc.value)

    def test_invalid_string_rejected(self):
        """Test that non-numeric strings are rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_ssh_port("abc")
        assert "integer" in str(exc.value).lower()

    def test_none_rejected(self):
        """Test that None is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_ssh_port(None)


class TestValidateSshHost:
    """Tests for validate_ssh_host."""

    def test_valid_hostname(self):
        """Test valid hostname."""
        validate_ssh_host("ssh.lagoon.example.com")

    def test_valid_simple_hostname(self):
        """Test valid simple hostname."""
        validate_ssh_host("localhost")

    def test_valid_ip_address(self):
        """Test valid IP address."""
        validate_ssh_host("192.168.1.100")

    def test_valid_hostname_with_hyphen(self):
        """Test valid hostname with hyphen."""
        validate_ssh_host("ssh-server.example.com")

    def test_none_allowed(self):
        """Test that None is allowed (optional field)."""
        validate_ssh_host(None)

    def test_invalid_with_space(self):
        """Test that hostname with space is rejected."""
        with pytest.raises(LagoonValidationError) as exc:
            validate_ssh_host("ssh server.example.com")
        assert "ssh_host" in str(exc.value)

    def test_invalid_with_protocol(self):
        """Test that hostname with protocol is rejected."""
        with pytest.raises(LagoonValidationError):
            validate_ssh_host("ssh://server.example.com")


class TestValidCloudProviders:
    """Tests to ensure VALID_CLOUD_PROVIDERS is properly defined."""

    def test_contains_expected_values(self):
        """Test VALID_CLOUD_PROVIDERS contains expected values."""
        assert "kind" in VALID_CLOUD_PROVIDERS
        assert "aws" in VALID_CLOUD_PROVIDERS
        assert "gcp" in VALID_CLOUD_PROVIDERS
        assert "azure" in VALID_CLOUD_PROVIDERS
        assert "openstack" in VALID_CLOUD_PROVIDERS
        assert "digitalocean" in VALID_CLOUD_PROVIDERS
        assert "other" in VALID_CLOUD_PROVIDERS

    def test_minimum_providers(self):
        """Test that at least common providers are present."""
        assert len(VALID_CLOUD_PROVIDERS) >= 5
