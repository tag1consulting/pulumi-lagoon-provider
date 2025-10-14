"""Tests for provider configuration."""

import pytest
from unittest.mock import patch
# from pulumi_lagoon.config import LagoonConfig


# TODO: Implement tests
# Example test structure:

# def test_config_from_environment():
#     """Test configuration from environment variables."""
#     with patch.dict('os.environ', {
#         'LAGOON_API_URL': 'https://api.test.com/graphql',
#         'LAGOON_TOKEN': 'test-token'
#     }):
#         config = LagoonConfig()
#         assert config.api_url == 'https://api.test.com/graphql'
#         assert config.token == 'test-token'

# def test_config_missing_token():
#     """Test that missing token raises error."""
#     with patch.dict('os.environ', {}, clear=True):
#         with pytest.raises(ValueError, match="token must be provided"):
#             LagoonConfig()
