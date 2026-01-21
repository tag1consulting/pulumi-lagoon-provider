# Pulumi Lagoon Provider - Complete Development Environment
#
# This Makefile provides a unified interface for setting up and managing
# the complete development/test environment including:
# - Kind cluster
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
        ensure-lagoon-admin ensure-deploy-target \
        clean clean-all venv

# Variables
VENV_DIR := venv
PYTHON := python3
CLUSTER_NAME := lagoon-test
TEST_CLUSTER_DIR := test-cluster
EXAMPLE_DIR := examples/simple-project

#==============================================================================
# Help
#==============================================================================

help:
	@echo "Pulumi Lagoon Provider - Development Environment"
	@echo ""
	@echo "Complete Setup (recommended for first-time users):"
	@echo "  make setup-all       - Complete setup from scratch (15-20 min)"
	@echo ""
	@echo "Individual Steps:"
	@echo "  make venv            - Create Python virtual environment"
	@echo "  make provider-install - Install provider in development mode"
	@echo "  make cluster-up      - Create Kind cluster and install Lagoon"
	@echo "  make cluster-down    - Destroy Kind cluster"
	@echo "  make cluster-status  - Check cluster and pod status"
	@echo ""
	@echo "Example Project:"
	@echo "  make example-preview - Preview example project changes"
	@echo "  make example-up      - Deploy example project"
	@echo "  make example-down    - Destroy example project resources"
	@echo "  make example-output  - Show example project outputs"
	@echo ""
	@echo "Testing:"
	@echo "  make provider-test   - Run provider tests"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean           - Kill port-forwards, clean temp files"
	@echo "  make clean-all       - Complete cleanup including Kind cluster"
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
# Kind Cluster + Lagoon
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
	@# First, wait for the api-migratedb job to complete (success or failure)
	@# This prevents blocking on a failed job pod
	@echo "Waiting for database migration job to complete..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12; do \
		JOB_STATUS=$$(kubectl --context kind-$(CLUSTER_NAME) get job -n lagoon -l app.kubernetes.io/component=api-migratedb -o jsonpath='{.items[0].status.conditions[?(@.type=="Complete")].status}' 2>/dev/null); \
		JOB_FAILED=$$(kubectl --context kind-$(CLUSTER_NAME) get job -n lagoon -l app.kubernetes.io/component=api-migratedb -o jsonpath='{.items[0].status.conditions[?(@.type=="Failed")].status}' 2>/dev/null); \
		if [ "$$JOB_STATUS" = "True" ]; then \
			echo "Migration job completed successfully."; \
			break; \
		elif [ "$$JOB_FAILED" = "True" ]; then \
			echo "Migration job failed, deleting for retry..."; \
			kubectl --context kind-$(CLUSTER_NAME) delete job -n lagoon -l app.kubernetes.io/component=api-migratedb 2>/dev/null || true; \
			sleep 10; \
		else \
			echo "Waiting for migration job... (attempt $$i/12)"; \
			sleep 10; \
		fi; \
	done
	@# Now wait for the key deployment pods (excluding job pods)
	@echo "Waiting for Lagoon API pods..."
	@kubectl --context kind-$(CLUSTER_NAME) wait --for=condition=ready pod \
		-l app.kubernetes.io/name=lagoon-core,app.kubernetes.io/component=api -n lagoon --timeout=300s 2>/dev/null || true
	@echo "Waiting for Keycloak pods..."
	@kubectl --context kind-$(CLUSTER_NAME) wait --for=condition=ready pod \
		-l app.kubernetes.io/name=lagoon-core,app.kubernetes.io/component=lagoon-core-keycloak -n lagoon --timeout=300s 2>/dev/null || true
	@echo "Waiting for Broker pods..."
	@kubectl --context kind-$(CLUSTER_NAME) wait --for=condition=ready pod \
		-l app.kubernetes.io/name=broker -n lagoon --timeout=300s 2>/dev/null || true
	@echo "Checking pod status..."
	@kubectl --context kind-$(CLUSTER_NAME) get pods -n lagoon

ensure-lagoon-admin:
	@echo "Ensuring lagoonadmin user exists in Keycloak..."
	@echo "Starting temporary port-forward to Keycloak..."
	@kubectl --context kind-$(CLUSTER_NAME) port-forward -n lagoon svc/lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 & \
		PF_PID=$$!; \
		sleep 2; \
		cd $(EXAMPLE_DIR) && ./scripts/create-lagoon-admin.sh; \
		kill $$PF_PID 2>/dev/null || true

ensure-deploy-target:
	@echo "Ensuring deploy target exists in Lagoon..."
	@echo "Starting temporary port-forwards..."
	@kubectl --context kind-$(CLUSTER_NAME) port-forward -n lagoon svc/lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 & \
		KC_PID=$$!; \
		kubectl --context kind-$(CLUSTER_NAME) port-forward -n lagoon svc/lagoon-core-api 7080:80 >/dev/null 2>&1 & \
		API_PID=$$!; \
		sleep 3; \
		DEPLOY_TARGET_ID=$$(cd $(EXAMPLE_DIR) && ./scripts/ensure-deploy-target.sh) || { kill $$KC_PID $$API_PID 2>/dev/null; exit 1; }; \
		kill $$KC_PID $$API_PID 2>/dev/null || true; \
		echo "Deploy target ID: $$DEPLOY_TARGET_ID"; \
		cd $(EXAMPLE_DIR) && pulumi config set deploytargetId $$DEPLOY_TARGET_ID 2>/dev/null || true

#==============================================================================
# Example Project
#==============================================================================

example-setup: venv provider-install
	@echo "Setting up example project..."
	@cd $(EXAMPLE_DIR) && \
		pulumi stack select test 2>/dev/null || pulumi stack init test
	@$(MAKE) ensure-deploy-target
	@echo "Example project ready."
	@echo "Run 'make example-up' to deploy."

example-preview:
	@cd $(EXAMPLE_DIR) && ./scripts/run-pulumi.sh preview

example-up:
	@cd $(EXAMPLE_DIR) && ./scripts/run-pulumi.sh up

example-down:
	@cd $(EXAMPLE_DIR) && ./scripts/run-pulumi.sh destroy

example-output:
	@cd $(EXAMPLE_DIR) && ./scripts/run-pulumi.sh stack output

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
