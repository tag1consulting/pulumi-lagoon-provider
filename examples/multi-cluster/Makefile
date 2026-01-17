# Multi-cluster Lagoon Example Makefile
# This Makefile provides convenience targets for managing the multi-cluster example

.PHONY: help setup up down preview refresh status clusters clean

# Default target
help:
	@echo "Multi-cluster Lagoon Example"
	@echo ""
	@echo "Setup:"
	@echo "  make setup        - Create virtual environment and install dependencies"
	@echo ""
	@echo "Pulumi Commands:"
	@echo "  make up           - Deploy the multi-cluster stack"
	@echo "  make down         - Destroy the multi-cluster stack"
	@echo "  make preview      - Preview changes without deploying"
	@echo "  make refresh      - Refresh stack state from providers"
	@echo "  make status       - Show current stack outputs"
	@echo ""
	@echo "Cluster Management:"
	@echo "  make clusters     - List all Kind clusters"
	@echo "  make clean        - Delete Kind clusters without Pulumi"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - Lagoon instance running (LAGOON_API_URL, LAGOON_TOKEN set)"
	@echo "  - Kind and kubectl installed"
	@echo "  - Pulumi CLI installed"

# Setup virtual environment
setup:
	@if [ ! -d "venv" ]; then \
		echo "Creating Python virtual environment..."; \
		python3 -m venv venv; \
	fi
	@./venv/bin/pip install --upgrade pip -q
	@./venv/bin/pip install -r requirements.txt -q
	@echo "Setup complete."

# Deploy the stack
up: setup
	./scripts/run-pulumi.sh up --yes

# Destroy the stack
down:
	./scripts/run-pulumi.sh destroy --yes

# Preview changes
preview: setup
	./scripts/run-pulumi.sh preview

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
	@echo "Done."
