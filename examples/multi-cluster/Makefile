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
        port-forwards token check fix-rabbitmq ensure-lagoon-admin

# Default target
help:
	@echo "Multi-cluster Lagoon Example"
	@echo ""
	@echo "Setup:"
	@echo "  make setup              - Create virtual environment and install dependencies"
	@echo "  make check              - Check cluster health and connectivity"
	@echo ""
	@echo "Pulumi Commands (auto-refresh token):"
	@echo "  make preview            - Preview pending changes"
	@echo "  make up                 - Deploy resources (non-interactive)"
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
	@echo ""
	@echo "Note: Most targets use the run-pulumi.sh wrapper which"
	@echo "      automatically handles token refresh and port-forwards."

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
