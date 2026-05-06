# Pulumi Lagoon Provider - Complete Development Environment
#
# This Makefile provides a unified interface for setting up and managing
# the complete development/test environment including:
# - Kind cluster(s)
# - Lagoon installation via Helm
# - Example project deployment
#
# Usage:
#   make help                - Show this help
#   make setup-all           - Complete setup from scratch
#   make cluster-up          - Create Kind cluster and install Lagoon
#   make cluster-down        - Destroy Kind cluster
#   make example-up          - Deploy example project
#   make example-down        - Destroy example project resources

.PHONY: help setup-all cluster-up cluster-down cluster-status \
        example-up example-down example-preview example-output \
        ensure-lagoon-admin ensure-deploy-target ensure-migrations \
        port-forwards check-health \
        multi-cluster-up multi-cluster-down multi-cluster-preview multi-cluster-status multi-cluster-clusters \
        multi-cluster-deploy multi-cluster-verify multi-cluster-port-forwards multi-cluster-port-forwards-all \
        multi-cluster-test-api multi-cluster-test-ui multi-cluster-info \
        clean clean-all \
        go-build go-test go-vet go-schema go-sdk-clean go-sdk-python go-sdk-nodejs go-sdk-go go-sdk-dotnet go-sdk-all go-install check-release-version release-prep go-proxy-warmup check-versions

# Variables
PYTHON := python3
CLUSTER_NAME := lagoon
SINGLE_CLUSTER_DIR := examples/single-cluster
EXAMPLE_DIR := examples/simple-project
SCRIPTS_DIR := scripts

#==============================================================================
# Help
#==============================================================================

help:
	@echo "Pulumi Lagoon Provider - Development Environment"
	@echo ""
	@echo "Complete Setup (recommended for first-time users):"
	@echo "  make setup-all       - Complete setup from scratch"
	@echo ""
	@echo "Individual Steps:"
	@echo "  make cluster-up      - Create Kind cluster and install Lagoon"
	@echo "  make cluster-down    - Destroy Kind cluster"
	@echo "  make cluster-status  - Check cluster and pod status"
	@echo ""
	@echo "Cluster Operations (using shared scripts):"
	@echo "  make check-health      - Check cluster health"
	@echo "  make port-forwards     - Set up kubectl port-forwards"
	@echo "  make ensure-migrations - Ensure Lagoon Knex migrations are run"
	@echo ""
	@echo "Example Project (simple-project):"
	@echo "  make example-preview - Preview example project changes"
	@echo "  make example-up      - Deploy example project"
	@echo "  make example-down    - Destroy example project resources"
	@echo "  make example-output  - Show example project outputs"
	@echo ""
	@echo "Multi-cluster Example:"
	@echo "  make multi-cluster-deploy  - Deploy with automatic retry (RECOMMENDED)"
	@echo "  make multi-cluster-up      - Create prod + nonprod clusters with full Lagoon stack:"
	@echo "                                 prod: lagoon-core + lagoon-remote + Harbor"
	@echo "                                 nonprod: lagoon-remote only (connects to prod core)"
	@echo "  make multi-cluster-down    - Destroy multi-cluster environment"
	@echo "  make multi-cluster-preview - Preview multi-cluster changes"
	@echo "  make multi-cluster-verify  - Verify deployment and test API"
	@echo "  make multi-cluster-status  - Show multi-cluster outputs"
	@echo "  make multi-cluster-clusters - List all Kind clusters"
	@echo "  make multi-cluster-port-forwards     - Start port-forwards for API/Keycloak"
	@echo "  make multi-cluster-port-forwards-all - Start port-forwards for all services (API, Keycloak, UI)"
	@echo "  make multi-cluster-test-api - Test Lagoon API access"
	@echo "  make multi-cluster-test-ui  - Test all services via port-forward"
	@echo ""
	@echo "Go Provider:"
	@echo "  make go-build        - Build the Go provider binary"
	@echo "  make go-test         - Run Go provider unit tests"
	@echo "  make go-sdk-python   - Regenerate Python SDK"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean           - Kill port-forwards, clean temp files"
	@echo "  make clean-all       - Complete cleanup including Kind cluster"
	@echo ""
	@echo "Shared Scripts (in scripts/ directory):"
	@echo "  ./scripts/check-cluster-health.sh   - Check cluster health"
	@echo "  ./scripts/setup-port-forwards.sh    - Set up port-forwards"
	@echo "  ./scripts/get-token.sh              - Get OAuth token"
	@echo "  ./scripts/fix-rabbitmq-password.sh  - Fix RabbitMQ auth issues"
	@echo "  ./scripts/run-pulumi.sh             - Wrapper with auto token refresh"
	@echo "  ./scripts/ensure-knex-migrations.sh - Check/run Knex migrations"
	@echo ""
	@echo "Script Configuration:"
	@echo "  LAGOON_PRESET=single      - Single-cluster (default)"
	@echo "  LAGOON_PRESET=multi-prod  - Multi-cluster production"
	@echo "  LAGOON_PRESET=multi-nonprod - Multi-cluster non-production"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - Docker installed and running"
	@echo "  - kind CLI installed (https://kind.sigs.k8s.io/)"
	@echo "  - kubectl installed"
	@echo "  - pulumi CLI installed"
	@echo "  - Python 3.9+"

#==============================================================================
# Complete Setup
#==============================================================================

setup-all: check-prerequisites cluster-up wait-for-lagoon ensure-lagoon-admin example-setup
	@echo ""
	@echo "=============================================="
	@echo "Setup Complete!"
	@echo "=============================================="
	@echo ""
	@echo "Lagoon is running on Kind cluster '$(CLUSTER_NAME)'"
	@echo ""
	@echo "Access via port-forwards (recommended for WSL2):"
	@echo "  Lagoon API: http://localhost:7080/graphql"
	@echo "  Keycloak:   http://localhost:8080/auth"
	@echo ""
	@echo "To deploy the example project:"
	@echo "  cd $(EXAMPLE_DIR)"
	@echo "  ./scripts/run-pulumi.sh up"
	@echo ""
	@echo "Or simply run: make example-up"

check-prerequisites:
	@echo "Checking prerequisites..."
	@command -v docker >/dev/null 2>&1 || { echo "ERROR: docker not found"; exit 1; }
	@command -v kind >/dev/null 2>&1 || { echo "ERROR: kind not found"; exit 1; }
	@command -v kubectl >/dev/null 2>&1 || { echo "ERROR: kubectl not found"; exit 1; }
	@command -v pulumi >/dev/null 2>&1 || { echo "ERROR: pulumi not found"; exit 1; }
	@command -v $(PYTHON) >/dev/null 2>&1 || { echo "ERROR: $(PYTHON) not found"; exit 1; }
	@docker info >/dev/null 2>&1 || { echo "ERROR: Docker daemon not running"; exit 1; }
	@echo "All prerequisites satisfied."

#==============================================================================
# Kind Cluster + Lagoon (Single-cluster)
#==============================================================================

cluster-up:
	@echo "Setting up Kind cluster and Lagoon..."
	@echo "This will take approximately 10-15 minutes."
	@echo ""
	@cd $(SINGLE_CLUSTER_DIR) && \
		if [ ! -d "venv" ]; then \
			$(PYTHON) -m venv venv; \
		fi && \
		. venv/bin/activate && \
		pip install -q -r requirements.txt && \
		pulumi stack select dev 2>/dev/null || pulumi stack init dev && \
		pulumi up --yes --skip-preview
	@echo ""
	@echo "Kind cluster and Lagoon deployed successfully!"

cluster-down:
	@echo "Destroying Kind cluster..."
	@kind delete cluster --name $(CLUSTER_NAME) 2>/dev/null || true
	@echo "Cluster destroyed."

cluster-status:
	@echo "Cluster Status:"
	@echo "==============="
	@# Check for single-cluster
	@if kind get clusters 2>/dev/null | grep -q "^lagoon$$"; then \
		echo ""; \
		echo "Single-cluster (lagoon): Running"; \
		CORE_PODS=$$(kubectl --context kind-lagoon get pods -n lagoon-core --no-headers 2>/dev/null | grep -c Running || echo 0); \
		REMOTE_PODS=$$(kubectl --context kind-lagoon get pods -n lagoon --no-headers 2>/dev/null | grep -c Running || echo 0); \
		echo "  lagoon-core: $$CORE_PODS pods running"; \
		echo "  lagoon (remote): $$REMOTE_PODS pods running"; \
	else \
		echo ""; \
		echo "Single-cluster (lagoon): Not found"; \
	fi
	@# Check for multi-cluster prod
	@if kind get clusters 2>/dev/null | grep -q "^lagoon-prod$$"; then \
		echo ""; \
		echo "Multi-cluster prod (lagoon-prod): Running"; \
		CORE_PODS=$$(kubectl --context kind-lagoon-prod get pods -n lagoon-core --no-headers 2>/dev/null | grep -c Running || echo 0); \
		REMOTE_PODS=$$(kubectl --context kind-lagoon-prod get pods -n lagoon --no-headers 2>/dev/null | grep -c Running || echo 0); \
		HARBOR_PODS=$$(kubectl --context kind-lagoon-prod get pods -n harbor --no-headers 2>/dev/null | grep -c Running || echo 0); \
		echo "  lagoon-core: $$CORE_PODS pods running"; \
		echo "  lagoon (remote): $$REMOTE_PODS pods running"; \
		echo "  harbor: $$HARBOR_PODS pods running"; \
	else \
		echo ""; \
		echo "Multi-cluster prod (lagoon-prod): Not found"; \
	fi
	@# Check for multi-cluster nonprod
	@if kind get clusters 2>/dev/null | grep -q "^lagoon-nonprod$$"; then \
		echo ""; \
		echo "Multi-cluster nonprod (lagoon-nonprod): Running"; \
		REMOTE_PODS=$$(kubectl --context kind-lagoon-nonprod get pods -n lagoon --no-headers 2>/dev/null | grep -c Running || echo 0); \
		echo "  lagoon (remote): $$REMOTE_PODS pods running"; \
	else \
		echo ""; \
		echo "Multi-cluster nonprod (lagoon-nonprod): Not found"; \
	fi
	@echo ""

wait-for-lagoon:
	@echo "Waiting for Lagoon to be ready..."
	@echo "Note: Pulumi handles migrations via ensure_knex_migrations"
	@echo "Waiting for Lagoon core pods..."
	@kubectl --context kind-$(CLUSTER_NAME) wait --for=condition=ready pod \
		-l app.kubernetes.io/name=lagoon-core -n lagoon-core --timeout=300s 2>/dev/null || true
	@echo "Waiting for Lagoon remote pods..."
	@kubectl --context kind-$(CLUSTER_NAME) wait --for=condition=ready pod \
		-l app.kubernetes.io/name=lagoon-remote -n lagoon --timeout=300s 2>/dev/null || true
	@echo "Checking pod status..."
	@kubectl --context kind-$(CLUSTER_NAME) get pods -n lagoon-core | head -20
	@kubectl --context kind-$(CLUSTER_NAME) get pods -n lagoon

#==============================================================================
# Shared Script Operations
#==============================================================================

check-health:
	@LAGOON_PRESET=single $(SCRIPTS_DIR)/check-cluster-health.sh

port-forwards:
	@LAGOON_PRESET=single $(SCRIPTS_DIR)/setup-port-forwards.sh

ensure-migrations:
	@LAGOON_PRESET=single $(SCRIPTS_DIR)/ensure-knex-migrations.sh

ensure-lagoon-admin:
	@echo "Ensuring lagoonadmin user exists in Keycloak..."
	@echo "Starting temporary port-forward to Keycloak..."
	@kubectl --context kind-$(CLUSTER_NAME) port-forward -n lagoon-core svc/lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 & \
		PF_PID=$$!; \
		sleep 3; \
		LAGOON_PRESET=single $(SCRIPTS_DIR)/create-lagoon-admin.sh; \
		kill $$PF_PID 2>/dev/null || true

ensure-deploy-target:
	@echo "Ensuring deploy target exists in Lagoon..."
	@echo "Starting temporary port-forwards..."
	@kubectl --context kind-$(CLUSTER_NAME) port-forward -n lagoon-core svc/lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 & \
		KC_PID=$$!; \
		kubectl --context kind-$(CLUSTER_NAME) port-forward -n lagoon-core svc/lagoon-core-api 7080:80 >/dev/null 2>&1 & \
		API_PID=$$!; \
		sleep 3; \
		DEPLOY_TARGET_ID=$$(LAGOON_PRESET=single $(SCRIPTS_DIR)/ensure-deploy-target.sh) || { kill $$KC_PID $$API_PID 2>/dev/null; exit 1; }; \
		kill $$KC_PID $$API_PID 2>/dev/null || true; \
		echo "Deploy target ID: $$DEPLOY_TARGET_ID"; \
		cd $(EXAMPLE_DIR) && pulumi config set deploytargetId $$DEPLOY_TARGET_ID 2>/dev/null || true

#==============================================================================
# Example Project (simple-project - provider usage demo)
#==============================================================================

example-setup:
	@echo "Setting up example project..."
	@$(MAKE) -C $(EXAMPLE_DIR) install
	@cd $(EXAMPLE_DIR) && \
		pulumi stack select test 2>/dev/null || pulumi stack init test
	@$(MAKE) ensure-deploy-target
	@echo "Example project ready."
	@echo "Run 'make example-up' to deploy."

example-preview:
	@cd $(EXAMPLE_DIR) && LAGOON_PRESET=single ./scripts/run-pulumi.sh preview

example-up:
	@cd $(EXAMPLE_DIR) && LAGOON_PRESET=single ./scripts/run-pulumi.sh up --yes

example-down:
	@cd $(EXAMPLE_DIR) && LAGOON_PRESET=single ./scripts/run-pulumi.sh destroy

example-output:
	@cd $(EXAMPLE_DIR) && LAGOON_PRESET=single ./scripts/run-pulumi.sh stack output

#==============================================================================
# Multi-cluster Example
#==============================================================================

MULTI_CLUSTER_DIR := examples/multi-cluster

multi-cluster-up:
	@echo "Creating multi-cluster environment (prod + nonprod)..."
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) deploy

multi-cluster-deploy:
	@echo "Deploying multi-cluster environment with automatic retry..."
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) deploy

multi-cluster-verify:
	@echo "Verifying multi-cluster deployment..."
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) verify

multi-cluster-down:
	@echo "Destroying multi-cluster environment..."
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) down

multi-cluster-preview:
	@echo "Previewing multi-cluster changes..."
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) preview

multi-cluster-status:
	@echo "Multi-cluster stack outputs:"
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) status

multi-cluster-clusters:
	@echo "Kind clusters:"
	@kind get clusters

multi-cluster-port-forwards:
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) port-forwards

multi-cluster-port-forwards-all:
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) port-forwards-all

multi-cluster-test-api:
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) test-api

multi-cluster-test-ui:
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) test-ui

multi-cluster-info:
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) show-access-info

#==============================================================================
# Cleanup
#==============================================================================

clean:
	@echo "Cleaning up..."
	@ps aux | grep '[k]ubectl.*port-forward' | awk '{print $$2}' | xargs -r kill 2>/dev/null || true
	@rm -rf /tmp/lagoon-certs 2>/dev/null || true
	@find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	@find . -type f -name "*.pyc" -delete 2>/dev/null || true
	@echo "Cleanup complete."

clean-all: clean
	@echo "Destroying Kind cluster(s)..."
	@kind delete cluster --name $(CLUSTER_NAME) 2>/dev/null || true
	@kind delete cluster --name lagoon-prod 2>/dev/null || true
	@kind delete cluster --name lagoon-nonprod 2>/dev/null || true
	@echo "Cluster(s) destroyed."
	@echo ""
	@echo "Removing Pulumi state..."
	@# Remove single-cluster Pulumi state
	@if [ -d "$(SINGLE_CLUSTER_DIR)/.pulumi" ]; then \
		echo "  Removing single-cluster Pulumi state..."; \
		rm -rf $(SINGLE_CLUSTER_DIR)/.pulumi; \
	fi
	@if [ -d "$(SINGLE_CLUSTER_DIR)/Pulumi.dev.yaml" ]; then \
		rm -f $(SINGLE_CLUSTER_DIR)/Pulumi.dev.yaml; \
	fi
	@# Remove multi-cluster Pulumi state
	@if [ -d "$(MULTI_CLUSTER_DIR)/.pulumi" ]; then \
		echo "  Removing multi-cluster Pulumi state..."; \
		rm -rf $(MULTI_CLUSTER_DIR)/.pulumi; \
	fi
	@if [ -f "$(MULTI_CLUSTER_DIR)/Pulumi.dev.yaml" ]; then \
		rm -f $(MULTI_CLUSTER_DIR)/Pulumi.dev.yaml; \
	fi
	@# Remove example project Pulumi state
	@if [ -d "$(EXAMPLE_DIR)/.pulumi" ]; then \
		echo "  Removing example project Pulumi state..."; \
		rm -rf $(EXAMPLE_DIR)/.pulumi; \
	fi
	@echo "Pulumi state removed."
	@echo ""
	@rm -rf $(SINGLE_CLUSTER_DIR)/venv 2>/dev/null || true
	@echo "Full cleanup complete."

#==============================================================================
# Go Provider (Native)
#==============================================================================

PROVIDER_VERSION ?= 0.4.1
PROVIDER_BIN     := provider/bin/pulumi-resource-lagoon
GO_BIN           ?= $(if $(GOPATH),$(GOPATH)/bin,$(HOME)/go/bin)

go-build:
	cd provider && mkdir -p bin && CGO_ENABLED=0 go build -ldflags "-X main.Version=$(PROVIDER_VERSION)" \
		-o bin/pulumi-resource-lagoon ./cmd/pulumi-resource-lagoon

go-test:
	cd provider && CGO_ENABLED=0 go test ./... -v -count=1

go-vet:
	cd provider && go vet ./...

go-lint:
	cd provider && golangci-lint run ./...

go-schema: go-build
	pulumi package get-schema ./$(PROVIDER_BIN) > provider/schema.json

go-sdk-clean:
	rm -rf sdk/python sdk/nodejs sdk/go sdk/dotnet

# SDK generation uses a temp directory to avoid deleting hand-maintained files
# (README.pypi.md, pyproject.toml license fix, package-lock.json, go.mod/go.sum).
# Generated files are rsynced over, then the temp directory is removed.
SDK_TMP := .sdk-gen-tmp

go-sdk-python: go-build
	rm -rf $(SDK_TMP)
	pulumi package gen-sdk ./$(PROVIDER_BIN) --language python -o $(SDK_TMP)
	rsync -a --ignore-existing $(SDK_TMP)/python/ sdk/python/
	rsync -a --delete $(SDK_TMP)/python/pulumi_lagoon/ sdk/python/pulumi_lagoon/
	cp LICENSE sdk/python/LICENSE
	rm -rf $(SDK_TMP)
	# Append __version__ so it survives SDK regeneration (generator does not emit it).
	printf '\n# Expose the package version as the conventional __version__ attribute.\n# The code generator does not emit this; it is appended by the go-sdk-python\n# Makefile target after generation so it survives SDK regenerations.\n__version__: str = _utilities.get_version()\n' >> sdk/python/pulumi_lagoon/__init__.py
	# Re-export resource classes at the top level so users can write
	# "from pulumi_lagoon import Project" instead of "from pulumi_lagoon.lagoon import Project".
	printf '\n# Re-export resource classes at the top level for convenience.\n# This is appended by the go-sdk-python Makefile target after generation.\nfrom .lagoon import *\n' >> sdk/python/pulumi_lagoon/__init__.py

go-sdk-nodejs: go-build
	rm -rf $(SDK_TMP)
	pulumi package gen-sdk ./$(PROVIDER_BIN) --language nodejs -o $(SDK_TMP)
	rsync -a --ignore-existing $(SDK_TMP)/nodejs/ sdk/nodejs/
	rsync -a --delete $(SDK_TMP)/nodejs/lagoon/ sdk/nodejs/lagoon/
	rsync -a --delete $(SDK_TMP)/nodejs/types/ sdk/nodejs/types/
	rsync -a --delete $(SDK_TMP)/nodejs/config/ sdk/nodejs/config/
	@for f in index.ts provider.ts utilities.ts tsconfig.json; do \
		if [ -f "$(SDK_TMP)/nodejs/$$f" ]; then cp "$(SDK_TMP)/nodejs/$$f" "sdk/nodejs/$$f"; fi; \
	done
	# Fix generated path: utilities.ts uses require('./package.json') but compiles to bin/
	# where package.json is one level up, so patch it to require('../package.json').
	sed -i.bak "s|require('./package.json')|require('../package.json')|g" sdk/nodejs/utilities.ts && rm -f sdk/nodejs/utilities.ts.bak
	# Re-export resource classes at the top level so users can write
	# import { Project } from "@tag1consulting/pulumi-lagoon".
	printf '\n// Re-export resource classes at the top level for convenience.\n// This is appended by the go-sdk-nodejs Makefile target after generation.\nexport * from "./lagoon";\n' >> sdk/nodejs/index.ts
	# Inject description field into package.json (codegen does not emit it).
	# Guard: only inject if not already present (idempotent across re-runs).
	grep -q '"description"' sdk/nodejs/package.json || \
		{ awk '{print} /"name": "@tag1consulting\/pulumi-lagoon"/{print "    \"description\": \"Manage Lagoon hosting platform resources as infrastructure-as-code.\","}' sdk/nodejs/package.json > sdk/nodejs/package.json.tmp && mv sdk/nodejs/package.json.tmp sdk/nodejs/package.json; }
	grep -q '"description"' sdk/nodejs/package.json || \
		(echo "ERROR: description injection failed in package.json" >&2 && exit 1)
	cp LICENSE sdk/nodejs/LICENSE
	rm -rf $(SDK_TMP)

go-sdk-go: go-build
	rm -rf $(SDK_TMP)
	pulumi package gen-sdk ./$(PROVIDER_BIN) --language go -o $(SDK_TMP)
	rsync -a --delete --exclude='go.mod' --exclude='go.sum' --exclude='LICENSE' $(SDK_TMP)/go/ sdk/go/
	cp LICENSE sdk/go/lagoon/LICENSE
	rm -rf $(SDK_TMP)

# go-sdk-dotnet regenerates the .NET SDK. The rsync uses --delete to refresh all
# generated files but preserves hand-maintained files (logo.png, README.md), then
# post-processes the result:
#   - patches TargetFramework to net8.0 (codegen emits net6.0 which is EOL)
#   - restores logo.png from docs/ (codegen copies the SVG under a .png name;
#     we overwrite it with the real PNG so NuGet package icon renders correctly)
go-sdk-dotnet: go-build
	rm -rf $(SDK_TMP)
	pulumi package gen-sdk ./$(PROVIDER_BIN) --language dotnet -o $(SDK_TMP)
	mkdir -p sdk/dotnet
	rsync -a --delete --exclude='logo.png' --exclude='README.md' $(SDK_TMP)/dotnet/ sdk/dotnet/
	cp LICENSE sdk/dotnet/LICENSE
	# Patch generated TargetFramework to net8.0 (codegen emits net6.0, which is EOL).
	# Match any TFM so this remains effective if codegen ever emits net7.0/net9.0.
	sed -i.bak 's|<TargetFramework>[^<]*</TargetFramework>|<TargetFramework>net8.0</TargetFramework>|' sdk/dotnet/Tag1Consulting.Lagoon.csproj && rm -f sdk/dotnet/Tag1Consulting.Lagoon.csproj.bak
	grep -q '<TargetFramework>net8.0</TargetFramework>' sdk/dotnet/Tag1Consulting.Lagoon.csproj || \
		(echo "ERROR: TargetFramework patch failed in Tag1Consulting.Lagoon.csproj" >&2 && exit 1)
	# Inject PackageReadmeFile property (codegen does not emit it).
	# Guard: only inject if not already present (idempotent across re-runs).
	grep -q '<PackageReadmeFile>' sdk/dotnet/Tag1Consulting.Lagoon.csproj || \
		{ awk '{print} /<PackageIcon>logo.png<\/PackageIcon>/{print "    <PackageReadmeFile>README.md</PackageReadmeFile>"}' sdk/dotnet/Tag1Consulting.Lagoon.csproj > sdk/dotnet/Tag1Consulting.Lagoon.csproj.tmp && mv sdk/dotnet/Tag1Consulting.Lagoon.csproj.tmp sdk/dotnet/Tag1Consulting.Lagoon.csproj; }
	grep -q '<PackageReadmeFile>README.md</PackageReadmeFile>' sdk/dotnet/Tag1Consulting.Lagoon.csproj || \
		(echo "ERROR: PackageReadmeFile patch failed in Tag1Consulting.Lagoon.csproj" >&2 && exit 1)
	# Inject PackageTags property (codegen does not emit it).
	# Guard: only inject if not already present (idempotent across re-runs).
	grep -q '<PackageTags>' sdk/dotnet/Tag1Consulting.Lagoon.csproj || \
		{ awk '{print} /<PackageReadmeFile>README.md<\/PackageReadmeFile>/{print "    <PackageTags>lagoon;pulumi;infrastructure-as-code;kubernetes;hosting</PackageTags>"}' sdk/dotnet/Tag1Consulting.Lagoon.csproj > sdk/dotnet/Tag1Consulting.Lagoon.csproj.tmp && mv sdk/dotnet/Tag1Consulting.Lagoon.csproj.tmp sdk/dotnet/Tag1Consulting.Lagoon.csproj; }
	grep -q '<PackageTags>' sdk/dotnet/Tag1Consulting.Lagoon.csproj || \
		(echo "ERROR: PackageTags injection failed in Tag1Consulting.Lagoon.csproj" >&2 && exit 1)
	# Pack README.md into the nupkg root (codegen does not include it).
	# Guard: only inject if not already present (idempotent across re-runs).
	grep -q '<None Include="README.md"' sdk/dotnet/Tag1Consulting.Lagoon.csproj || \
		{ awk '/<None Include="logo.png">/{print "    <None Include=\"README.md\" Pack=\"true\" PackagePath=\"\" />"} {print}' sdk/dotnet/Tag1Consulting.Lagoon.csproj > sdk/dotnet/Tag1Consulting.Lagoon.csproj.tmp && mv sdk/dotnet/Tag1Consulting.Lagoon.csproj.tmp sdk/dotnet/Tag1Consulting.Lagoon.csproj; }
	# Replace the SVG-disguised-as-PNG that codegen copies with a real PNG.
	cp docs/logo.png sdk/dotnet/logo.png
	# Ensure version.txt has a trailing newline (codegen omits it).
	awk 1 sdk/dotnet/version.txt > sdk/dotnet/version.txt.tmp && mv sdk/dotnet/version.txt.tmp sdk/dotnet/version.txt
	rm -rf $(SDK_TMP)

# go-sdk-all regenerates all SDKs without a clean; use go-sdk-clean first for a
# full reset (note: go-sdk-clean deletes hand-maintained files like README.pypi.md,
# pyproject.toml license fix, package-lock.json, go.mod/go.sum, and
# Tag1Consulting.Lagoon.csproj; only use it when those can be restored from git).
go-sdk-all: go-sdk-python go-sdk-nodejs go-sdk-go go-sdk-dotnet

go-install: go-build
	mkdir -p $(GO_BIN)
	cp $(PROVIDER_BIN) $(GO_BIN)/

# Guard target: ensures VERSION is a bare semver (x.y.z) before expensive work begins.
# Intentionally stricter than CI (which allows pre-release/build suffixes like x.y.z-rc1):
# release-prep only cuts stable releases, and the publish pipeline does not support
# pre-release NuGet/PyPI/npm packages. Use a release tag directly to publish pre-releases.
check-release-version:
ifndef VERSION
	$(error VERSION is required. Usage: make release-prep VERSION=0.3.0)
endif
	@echo "$(VERSION)" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$$' || \
		(echo "ERROR: VERSION must be a bare semver (e.g. 0.3.0), got '$(VERSION)'" >&2 && exit 1)

# Release prep: bump versions first, then rebuild provider and regenerate SDKs,
# then run tests.  Versions are updated before the build so the provider binary
# and generated SDKs carry the new version string from the start.
# Usage: make release-prep VERSION=0.3.0
release-prep: check-release-version
	@echo "=== Setting version $(VERSION) ==="
	sed -i.bak 's/var Version = .*/var Version = "$(VERSION)"/' provider/cmd/pulumi-resource-lagoon/main.go && rm -f provider/cmd/pulumi-resource-lagoon/main.go.bak
	sed -i.bak 's/^PROVIDER_VERSION ?= .*/PROVIDER_VERSION ?= $(VERSION)/' Makefile && rm -f Makefile.bak
	sed -i.bak 's/"version": ".*"/"version": "$(VERSION)"/' provider/schema.json && rm -f provider/schema.json.bak
	sed -i.bak 's/^\*\*Status\*\*: v[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*/\*\*Status\*\*: v$(VERSION)/' README.md && rm -f README.md.bak
	sed -i.bak 's/^\*\*Status\*\*: v[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*/\*\*Status\*\*: v$(VERSION)/' CLAUDE.md && rm -f CLAUDE.md.bak
	printf '# Release v$(VERSION) (%s)\n\nTODO: Write release notes before committing.\n\n---\n\n' "$$(date +%Y-%m-%d)" > /tmp/rn_header.md && cat RELEASE_NOTES.md >> /tmp/rn_header.md && mv /tmp/rn_header.md RELEASE_NOTES.md
	$(MAKE) PROVIDER_VERSION=$(VERSION) go-build go-sdk-python go-sdk-nodejs go-sdk-go go-sdk-dotnet
	sed -i.bak 's/^  version = .*/  version = "$(VERSION)"/' sdk/python/pyproject.toml && rm -f sdk/python/pyproject.toml.bak
	jq --indent 4 --arg v "$(VERSION)" '.version = $$v | .pulumi.version = $$v' sdk/nodejs/package.json > sdk/nodejs/package.json.tmp && mv sdk/nodejs/package.json.tmp sdk/nodejs/package.json
	sed -i.bak 's|<Version>.*</Version>|<Version>$(VERSION)</Version>|' sdk/dotnet/Tag1Consulting.Lagoon.csproj && rm -f sdk/dotnet/Tag1Consulting.Lagoon.csproj.bak
	echo "$(VERSION)" > sdk/dotnet/version.txt
	@echo "=== Running tests ==="
	cd provider && CGO_ENABLED=0 go test ./... -count=1
	@echo ""
	@echo "=== Release prep complete for v$(VERSION) ==="
	@echo "Remaining steps:"
	@echo "  1. Fill in RELEASE_NOTES.md (scaffold inserted at top of file)"
	@echo "  2. Commit, push, and open PR to main"
	@echo "  3. After merge: git tag v$(VERSION) && git push origin v$(VERSION)"
	@echo "  4. Tag Go module: git tag sdk/go/lagoon/v$(VERSION) v$(VERSION)^{} && git push origin sdk/go/lagoon/v$(VERSION)"
	@echo "  5. Create GitHub release (triggers publish.yml: PyPI, npm, NuGet, Go proxy warm-up)"
	@echo "  6. Verify: make go-proxy-warmup VERSION=$(VERSION)  (or check CI)"

# Warm the Go module proxy so the SDK is immediately available via go get and
# appears on pkg.go.dev.  The sdk/go/lagoon/vVERSION tag must already exist
# on the remote before running this.
# Usage: make go-proxy-warmup VERSION=0.3.0
GO_MODULE := github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon
GO_PROXY  := https://proxy.golang.org

go-proxy-warmup: check-release-version
	@echo "=== Warming Go module proxy for $(GO_MODULE)@v$(VERSION) ==="
	@for endpoint in .info .mod .zip; do \
		url="$(GO_PROXY)/$(GO_MODULE)/@v/v$(VERSION)$$endpoint"; \
		echo -n "  GET $$url ... "; \
		status=0; \
		for attempt in 1 2 3; do \
			code=$$(curl -s -o /dev/null -w '%{http_code}' "$$url"); \
			if [ "$$code" = "200" ]; then \
				echo "200 OK"; \
				status=1; \
				break; \
			fi; \
			echo -n "($$code, retry $$attempt/3) "; \
			sleep 10; \
		done; \
		if [ "$$status" = "0" ]; then \
			echo "FAILED"; \
			echo "ERROR: proxy returned non-200 for $$url" >&2; \
			echo "Ensure the tag sdk/go/lagoon/v$(VERSION) has been pushed." >&2; \
			exit 1; \
		fi; \
	done
	@echo "=== Go module proxy warm-up complete ==="

check-versions:
	bash scripts/check-version-consistency.sh
