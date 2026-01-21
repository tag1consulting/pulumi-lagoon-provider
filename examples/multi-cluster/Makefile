# Multi-cluster Lagoon Example Makefile
#
# This Makefile provides convenience targets for managing the multi-cluster example
#
# Usage:
#   make help          - Show this help
#   make setup         - Initial setup (venv + dependencies)
#   make preview       - Preview changes
#   make up            - Deploy resources (non-interactive)
#   make down          - Destroy all resources
#   make refresh       - Refresh state from clusters
#   make status        - Show stack outputs
#   make clusters      - List Kind clusters
#   make clean         - Delete Kind clusters directly

.PHONY: help setup preview up down refresh status clusters clean \
        port-forwards token check fix-rabbitmq ensure-lagoon-admin \
        deploy verify wait-for-pods test-api

# Default target
help:
	@echo "Multi-cluster Lagoon Example"
	@echo ""
	@echo "Quick Start:"
	@echo "  make deploy             - Full deployment with automatic retry (RECOMMENDED)"
	@echo "  make verify             - Verify deployment health and test API"
	@echo ""
	@echo "Setup:"
	@echo "  make setup              - Create virtual environment and install dependencies"
	@echo "  make check              - Check cluster health and connectivity"
	@echo ""
	@echo "Pulumi Commands (auto-refresh token):"
	@echo "  make preview            - Preview pending changes"
	@echo "  make up                 - Deploy resources (non-interactive, may timeout)"
	@echo "  make down               - Destroy all resources (non-interactive)"
	@echo "  make refresh            - Refresh state from clusters"
	@echo "  make status             - Show stack outputs"
	@echo ""
	@echo "Cluster Management:"
	@echo "  make clusters           - List all Kind clusters"
	@echo "  make clean              - Delete Kind clusters without Pulumi"
	@echo ""
	@echo "Utility targets:"
	@echo "  make port-forwards      - Start kubectl port-forwards"
	@echo "  make token              - Get a fresh OAuth token"
	@echo "  make ensure-lagoon-admin - Create lagoonadmin user in Keycloak"
	@echo "  make fix-rabbitmq       - Fix RabbitMQ password mismatch"
	@echo "  make wait-for-pods      - Wait for all Lagoon pods to be running"
	@echo "  make test-api           - Test Lagoon API access"
	@echo ""
	@echo "Note: 'make deploy' handles Helm timeouts automatically by running"
	@echo "      refresh + up cycles until all resources are deployed."

# Setup virtual environment
setup:
	@if [ ! -d "venv" ]; then \
		echo "Creating Python virtual environment..."; \
		python3 -m venv venv; \
	fi
	@./venv/bin/pip install --upgrade pip -q
	@./venv/bin/pip install -r requirements.txt -q
	@echo "Setup complete."

# Check cluster health
check:
	./scripts/check-cluster-health.sh

# Start port-forwards
port-forwards:
	@echo "Starting port-forwards to prod cluster..."
	@pkill -f "port-forward.*lagoon-core" 2>/dev/null || true
	@kubectl --context kind-lagoon-prod port-forward -n lagoon-core svc/prod-core-lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 &
	@kubectl --context kind-lagoon-prod port-forward -n lagoon-core svc/prod-core-lagoon-core-api 7080:80 >/dev/null 2>&1 &
	@sleep 2
	@echo "Port-forwards started:"
	@echo "  Keycloak: http://localhost:8080/auth"
	@echo "  API:      http://localhost:7080/graphql"

# Get fresh token (must be sourced to export variables)
token:
	@echo "To get a token, run:"
	@echo "  source <(./scripts/run-pulumi.sh --token-only 2>/dev/null)"
	@echo ""
	@echo "Or use the run-pulumi.sh wrapper which handles tokens automatically."

# Preview changes
preview: setup
	./scripts/run-pulumi.sh preview

# Deploy resources
up: setup
	./scripts/run-pulumi.sh up --yes

# Destroy all resources
down:
	./scripts/run-pulumi.sh destroy --yes

# Refresh state
refresh:
	./scripts/run-pulumi.sh refresh --yes

# Show stack outputs
status:
	@./scripts/run-pulumi.sh stack output 2>/dev/null || echo "Stack not initialized. Run 'make up' first."

# List Kind clusters
clusters:
	@echo "Kind clusters:"
	@kind get clusters 2>/dev/null || echo "No clusters found"
	@echo ""
	@echo "Kubernetes contexts:"
	@kubectl config get-contexts 2>/dev/null | grep -E "(CURRENT|kind)" || echo "No Kind contexts found"

# Clean up clusters directly (without Pulumi)
clean:
	@echo "Deleting Kind clusters..."
	-kind delete cluster --name lagoon-prod 2>/dev/null
	-kind delete cluster --name lagoon-nonprod 2>/dev/null
	@echo "Killing port-forwards..."
	@pkill -f "port-forward.*lagoon-core" 2>/dev/null || true
	@echo "Done."

# Ensure lagoonadmin user exists in Keycloak
ensure-lagoon-admin:
	@echo "Ensuring lagoonadmin user exists in Keycloak..."
	@echo "Starting temporary port-forward to Keycloak..."
	@pkill -f "port-forward.*lagoon-core" 2>/dev/null || true
	@kubectl --context kind-lagoon-prod port-forward -n lagoon-core svc/prod-core-lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 & \
		PF_PID=$$!; \
		sleep 2; \
		./scripts/create-lagoon-admin.sh; \
		kill $$PF_PID 2>/dev/null || true

# Fix RabbitMQ password mismatch
fix-rabbitmq:
	@echo "Fixing RabbitMQ password on both clusters..."
	./scripts/fix-rabbitmq-password.sh prod || true
	./scripts/fix-rabbitmq-password.sh nonprod || true

# =============================================================================
# Automated Deployment Targets
# =============================================================================

# Full deployment with automatic retry on Helm timeout
# This handles the common case where Lagoon core Helm release times out
# but pods are actually running successfully
deploy: setup
	@echo "=========================================="
	@echo "Multi-cluster Lagoon Deployment"
	@echo "=========================================="
	@echo ""
	@echo "This will deploy the complete multi-cluster Lagoon environment."
	@echo "The Lagoon core Helm release may timeout (15 min) but pods usually"
	@echo "start successfully. This script handles that automatically."
	@echo ""
	@# Initial deployment attempt
	./scripts/run-pulumi.sh up --yes || true
	@echo ""
	@echo "Checking if deployment completed..."
	@# Wait for pods to be ready
	@$(MAKE) wait-for-pods
	@# Refresh state and complete any remaining resources
	@echo ""
	@echo "Refreshing Pulumi state and completing deployment..."
	./scripts/run-pulumi.sh refresh --yes || true
	./scripts/run-pulumi.sh up --yes || true
	@echo ""
	@echo "=========================================="
	@echo "Deployment complete! Running verification..."
	@echo "=========================================="
	@$(MAKE) verify

# Wait for Lagoon pods to be running
wait-for-pods:
	@echo "Waiting for Lagoon pods to be ready..."
	@# Wait for prod cluster Lagoon core pods
	@echo "Checking prod cluster..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do \
		RUNNING=$$(kubectl --context kind-lagoon-prod get pods -n lagoon-core --no-headers 2>/dev/null | grep -c Running || echo 0); \
		if [ "$$RUNNING" -ge 30 ]; then \
			echo "  Prod cluster: $$RUNNING pods running"; \
			break; \
		fi; \
		echo "  Waiting for pods... ($$RUNNING running, attempt $$i/20)"; \
		sleep 15; \
	done
	@# Check nonprod cluster (may not have pods yet on first run)
	@echo "Checking nonprod cluster..."
	@NONPROD_NS=$$(kubectl --context kind-lagoon-nonprod get ns lagoon --no-headers 2>/dev/null | wc -l || echo 0); \
	if [ "$$NONPROD_NS" -gt 0 ]; then \
		for i in 1 2 3 4 5; do \
			RUNNING=$$(kubectl --context kind-lagoon-nonprod get pods -n lagoon --no-headers 2>/dev/null | grep -c Running || echo 0); \
			if [ "$$RUNNING" -ge 1 ]; then \
				echo "  Nonprod cluster: $$RUNNING pods running"; \
				break; \
			fi; \
			echo "  Waiting for nonprod pods... (attempt $$i/5)"; \
			sleep 10; \
		done; \
	else \
		echo "  Nonprod lagoon namespace not created yet (will be created on next pulumi up)"; \
	fi

# Verify deployment health and test API
verify: check
	@echo ""
	@echo "=========================================="
	@echo "Testing Lagoon API Access"
	@echo "=========================================="
	@$(MAKE) test-api
	@echo ""
	@echo "=========================================="
	@echo "Verification Complete!"
	@echo "=========================================="
	@echo ""
	@echo "Your multi-cluster Lagoon environment is ready."
	@echo ""
	@echo "Next steps:"
	@echo "  1. Access the Lagoon UI: make port-forwards && open http://localhost:8080"
	@echo "  2. Create a project using the Pulumi provider"
	@echo "  3. Check the README.md for more examples"

# Test Lagoon API access
test-api:
	@echo "Starting port-forwards..."
	@pkill -f "port-forward.*lagoon-core" 2>/dev/null || true
	@kubectl --context kind-lagoon-prod port-forward -n lagoon-core svc/prod-core-lagoon-core-keycloak 8080:8080 >/dev/null 2>&1 &
	@kubectl --context kind-lagoon-prod port-forward -n lagoon-core svc/prod-core-lagoon-core-api 7080:80 >/dev/null 2>&1 &
	@sleep 3
	@echo ""
	@echo "Getting OAuth token..."
	@PASSWORD=$$(kubectl --context kind-lagoon-prod -n lagoon-core get secret prod-core-lagoon-core-keycloak -o jsonpath='{.data.KEYCLOAK_LAGOON_ADMIN_PASSWORD}' | base64 -d); \
	TOKEN=$$(curl -sf "http://localhost:8080/auth/realms/lagoon/protocol/openid-connect/token" \
		-d "grant_type=password" \
		-d "client_id=lagoon-ui" \
		-d "username=lagoonadmin" \
		-d "password=$$PASSWORD" 2>/dev/null | jq -r '.access_token // empty'); \
	if [ -n "$$TOKEN" ]; then \
		echo "  Token acquired successfully"; \
		echo ""; \
		echo "Testing API..."; \
		RESULT=$$(curl -sf "http://localhost:7080/graphql" \
			-H "Authorization: Bearer $$TOKEN" \
			-H "Content-Type: application/json" \
			-d '{"query": "{ allProjects { id } }"}' 2>/dev/null); \
		if echo "$$RESULT" | jq -e '.data' >/dev/null 2>&1; then \
			echo "  API responding correctly"; \
		else \
			echo "  WARNING: API returned unexpected response"; \
			echo "  $$RESULT"; \
		fi; \
	else \
		echo "  WARNING: Could not get OAuth token"; \
		echo "  This might be normal if Keycloak is still initializing."; \
		echo "  Try running: make ensure-lagoon-admin"; \
	fi
