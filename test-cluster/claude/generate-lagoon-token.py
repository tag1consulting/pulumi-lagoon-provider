#!/usr/bin/env python3
"""
Generate a Lagoon API JWT token for testing

This script creates a JWT token that can be used to authenticate with
the Lagoon GraphQL API.
"""

import jwt
import time
import sys

# JWT secret from lagoon-core-secrets
JWT_SECRET = "UHRnCTAslHmYuSzfPWIGeubzSsVEMZew"

# Token payload - create an admin-level token
payload = {
    "role": "admin",
    "iss": "lagoon-local",
    "sub": "admin",
    "aud": "api.lagoon.test",
    "iat": int(time.time()),
    "exp": int(time.time()) + (365 * 24 * 60 * 60),  # 1 year expiration
}

# Generate token
token = jwt.encode(payload, JWT_SECRET, algorithm="HS256")

print(token)
