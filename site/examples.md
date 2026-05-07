---
title: Examples
nav_order: 4
has_children: true
---

# Examples

Three example projects demonstrate different deployment scenarios, from a minimal provider usage test to a full production-like multi-cluster setup.

Full source code for all examples is in the [`examples/`](https://github.com/tag1consulting/pulumi-lagoon-provider/tree/main/examples) directory of the repository.

| Example | What it covers |
|---------|---------------|
| [Simple Project](simple-project/) | Provider API: projects, environments, variables, all 4 notification types |
| [Single Cluster](single-cluster/) | Complete Lagoon stack on one Kind cluster; good for local development and testing |
| [Multi-Cluster](multi-cluster/) | Production-like setup with separate prod and nonprod clusters |
