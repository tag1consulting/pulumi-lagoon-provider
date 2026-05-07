---
title: Installation
parent: Getting Started
nav_order: 1
---

# Installation

Install the SDK for your language of choice. The provider binary is installed automatically the first time you run `pulumi up` or `pulumi preview`.

## SDK Installation

#### Python

```bash
pip install pulumi-lagoon
```

[View on PyPI](https://pypi.org/project/pulumi-lagoon/)

#### TypeScript / JavaScript

```bash
npm install @tag1consulting/pulumi-lagoon
# or
yarn add @tag1consulting/pulumi-lagoon
```

[View on npm](https://www.npmjs.com/package/@tag1consulting/pulumi-lagoon)

#### Go

```bash
go get github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon
```

[View on pkg.go.dev](https://pkg.go.dev/github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon)

#### .NET / C#

```bash
dotnet add package Tag1Consulting.Lagoon
```

[View on NuGet](https://www.nuget.org/packages/Tag1Consulting.Lagoon)

## Provider Binary

The provider binary (`pulumi-resource-lagoon`) is installed automatically when Pulumi resolves the plugin for your stack. No manual action is required in most cases.

If you need to install it manually — for example, in a CI environment without internet access or to pin a specific version — run:

```bash
pulumi plugin install resource lagoon <version> \
  --server github://api.github.com/tag1consulting/pulumi-lagoon-provider
```

Replace `<version>` with the version you want, e.g., `0.5.0`. Versions are listed on the [GitHub releases page](https://github.com/tag1consulting/pulumi-lagoon-provider/releases).
