# Pulumi Lagoon Provider - Complete Development Environment
#
# This Makefile provides a unified interface for setting up and managing
# the complete development/test environment including:
# - Kind cluster(s)
# - Lagoon installation via Helm
# - Python provider installation
# - Example project deployment
#
# Usage:
#   make help                - Show this help
#   make setup-all           - Complete setup from scratch
#   make cluster-up          - Create Kind cluster and install Lagoon
#   make cluster-down        - Destroy Kind cluster
#   make provider-install    - Install provider in development mode
#   make example-up          - Deploy example project
#   make example-down        - Destroy example project resources

.PHONY: help setup-all cluster-up cluster-down cluster-status \
        provider-install provider-test \
        example-up example-down example-preview example-output \
        ensure-lagoon-admin ensure-deploy-target ensure-migrations \
        port-forwards check-health \
        multi-cluster-up multi-cluster-down multi-cluster-preview multi-cluster-status multi-cluster-clusters \
        multi-cluster-deploy multi-cluster-verify multi-cluster-port-forwards multi-cluster-test-api \
        clean clean-all venv

# Variables
VENV_DIR := venv
PYTHON := python3
CLUSTER_NAME := lagoon-test
TEST_CLUSTER_DIR := test-cluster
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
	@echo "  make venv            - Create Python virtual environment"
	@echo "  make provider-install - Install provider in development mode"
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
	@echo "  make multi-cluster-port-forwards - Start kubectl port-forwards"
	@echo "  make multi-cluster-test-api - Test Lagoon API access"
	@echo ""
	@echo "Testing:"
	@echo "  make provider-test   - Run provider tests"
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
	@echo "  LAGOON_PRESET=single      - Single-cluster (default, test-cluster)"
	@echo "  LAGOON_PRESET=multi-prod  - Multi-cluster production"
	@echo "  LAGOON_PRESET=multi-nonprod - Multi-cluster non-production"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - Docker installed and running"
	@echo "  - kind CLI installed (https://kind.sigs.k8s.io/)"
	@echo "  - kubectl installed"
	@echo "  - pulumi CLI installed"
	@echo "  - Python 3.8+"

#==============================================================================
# Complete Setup
#==============================================================================

setup-all: check-prerequisites venv provider-install cluster-up wait-for-lagoon ensure-lagoon-admin example-setup
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
# Python Virtual Environment
#==============================================================================

venv:
	@if [ ! -d "$(VENV_DIR)" ]; then \
		echo "Creating Python virtual environment..."; \
		$(PYTHON) -m venv $(VENV_DIR); \
		. $(VENV_DIR)/bin/activate && pip install --upgrade pip; \
		echo "Virtual environment created at $(VENV_DIR)"; \
	else \
		echo "Virtual environment already exists at $(VENV_DIR)"; \
	fi

#==============================================================================
# Provider Installation
#==============================================================================

provider-install: venv
	@echo "Installing provider in development mode..."
	@. $(VENV_DIR)/bin/activate && pip install -e .
	@echo "Provider installed successfully."

provider-test: venv
	@echo "Running provider tests..."
	@. $(VENV_DIR)/bin/activate && pytest tests/ -v

#==============================================================================
# Kind Cluster + Lagoon (Single-cluster via test-cluster)
#==============================================================================

cluster-up: venv
	@echo "Setting up Kind cluster and Lagoon..."
	@echo "This will take approximately 10-15 minutes."
	@echo ""
	@cd $(TEST_CLUSTER_DIR) && \
		if [ ! -d "venv" ]; then \
			$(PYTHON) -m venv venv; \
		fi && \
		. venv/bin/activate && \
		pip install -q -r requirements.txt && \
		pulumi stack select dev 2>/dev/null || pulumi stack init dev && \
		pulumi up --yes
	@echo ""
	@echo "Kind cluster and Lagoon deployed successfully!"

cluster-down:
	@echo "Destroying Kind cluster..."
	@cd $(TEST_CLUSTER_DIR) && \
		. venv/bin/activate 2>/dev/null && \
		pulumi destroy --yes 2>/dev/null || true
	@kind delete cluster --name $(CLUSTER_NAME) 2>/dev/null || true
	@echo "Cluster destroyed."

cluster-status:
	@echo "Cluster Status:"
	@echo "==============="
	@kind get clusters 2>/dev/null | grep -q $(CLUSTER_NAME) && echo "Kind cluster '$(CLUSTER_NAME)': Running" || echo "Kind cluster '$(CLUSTER_NAME)': Not found"
	@echo ""
	@echo "Lagoon Pods:"
	@kubectl --context kind-$(CLUSTER_NAME) get pods -n lagoon 2>/dev/null || echo "Could not get pod status"

wait-for-lagoon:
	@echo "Waiting for Lagoon to be ready..."
	@echo "Waiting for database migration job to complete..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12; do \
		JOB_STATUS=$$(kubectl --context kind-$(CLUSTER_NAME) get job lagoon-core-api-migratedb -n lagoon -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' 2>/dev/null); \
		JOB_FAILED=$$(kubectl --context kind-$(CLUSTER_NAME) get job lagoon-core-api-migratedb -n lagoon -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}' 2>/dev/null); \
		if [ "$$JOB_STATUS" = "True" ]; then \
			echo "Migration job completed successfully."; \
			break; \
		elif [ "$$JOB_FAILED" = "True" ]; then \
			echo "Migration job failed, deleting for retry..."; \
			kubectl --context kind-$(CLUSTER_NAME) delete job lagoon-core-api-migratedb -n lagoon 2>/dev/null || true; \
			sleep 10; \
		else \
			echo "Waiting for migration job... (attempt $$i/12)"; \
			sleep 10; \
		fi; \
	done
	@echo "Waiting for Lagoon core pods..."
	@kubectl --context kind-$(CLUSTER_NAME) wait --for=condition=ready pod \
		-l app.kubernetes.io/name=lagoon-core --field-selector=status.phase=Running -n lagoon --timeout=300s 2>/dev/null || true
	@echo "Waiting for Broker pods..."
	@kubectl --context kind-$(CLUSTER_NAME) wait --for=condition=ready pod \
		-l app.kubernetes.io/name=broker -n lagoon --timeout=300s 2>/dev/null || true
	@echo "Checking pod status..."
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
	@kubectl --context kind-$(CLUSTER_NAME) port-forward -n lagoon svc/lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 & \
		PF_PID=$$!; \
		sleep 2; \
		LAGOON_PRESET=single $(SCRIPTS_DIR)/create-lagoon-admin.sh; \
		kill $$PF_PID 2>/dev/null || true

ensure-deploy-target:
	@echo "Ensuring deploy target exists in Lagoon..."
	@echo "Starting temporary port-forwards..."
	@kubectl --context kind-$(CLUSTER_NAME) port-forward -n lagoon svc/lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 & \
		KC_PID=$$!; \
		kubectl --context kind-$(CLUSTER_NAME) port-forward -n lagoon svc/lagoon-core-api 7080:80 >/dev/null 2>&1 & \
		API_PID=$$!; \
		sleep 3; \
		DEPLOY_TARGET_ID=$$(LAGOON_PRESET=single $(SCRIPTS_DIR)/ensure-deploy-target.sh) || { kill $$KC_PID $$API_PID 2>/dev/null; exit 1; }; \
		kill $$KC_PID $$API_PID 2>/dev/null || true; \
		echo "Deploy target ID: $$DEPLOY_TARGET_ID"; \
		cd $(EXAMPLE_DIR) && pulumi config set deploytargetId $$DEPLOY_TARGET_ID 2>/dev/null || true

#==============================================================================
# Example Project (simple-project - provider usage demo)
#==============================================================================

example-setup: venv provider-install
	@echo "Setting up example project..."
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

multi-cluster-up: venv provider-install
	@echo "Creating multi-cluster environment (prod + nonprod)..."
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) up

multi-cluster-deploy: venv provider-install
	@echo "Deploying multi-cluster environment with automatic retry..."
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) deploy

multi-cluster-verify:
	@echo "Verifying multi-cluster deployment..."
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) verify

multi-cluster-down:
	@echo "Destroying multi-cluster environment..."
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) down

multi-cluster-preview: venv provider-install
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

multi-cluster-test-api:
	@cd $(MULTI_CLUSTER_DIR) && $(MAKE) test-api

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

clean-all: clean cluster-down
	@rm -rf $(VENV_DIR) 2>/dev/null || true
	@rm -rf $(TEST_CLUSTER_DIR)/venv 2>/dev/null || true
	@echo "Full cleanup complete."
