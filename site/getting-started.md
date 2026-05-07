---
title: Getting Started
nav_order: 1
has_children: true
---

# Getting Started

The Pulumi Lagoon Provider lets you manage [Lagoon](https://lagoon.sh/) hosting platform resources — projects, environments, variables, notifications, routes, users, and more — as declarative infrastructure-as-code using your preferred programming language.

## Prerequisites

Before you begin, make sure you have:

- **Pulumi CLI** installed ([pulumi.com/docs/install](https://www.pulumi.com/docs/install/))
- **Access to a Lagoon instance** with API credentials:
  - GraphQL API endpoint (e.g., `https://api.lagoon.example.com/graphql`)
  - A JWT authentication token **or** the Lagoon core `JWTSECRET`
- **One language runtime:**
  - Python 3.9+
  - Node.js 18+
  - Go 1.22+
  - .NET 6.0+

## In This Section

[**Installation**](installation/) — Install the SDK for your language and set up the provider binary.

[**Configuration**](configuration/) — Configure the provider with your Lagoon API endpoint and authentication credentials.

[**Quick Start**](quick-start/) — A five-minute walkthrough that creates a Lagoon project, environment, and variable using all four supported languages.

[**Importing Resources**](importing-resources/) — Bring existing Lagoon resources under Pulumi management using `pulumi import`.
