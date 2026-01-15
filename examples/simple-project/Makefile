# Makefile for Lagoon Pulumi Provider Example
#
# This Makefile provides convenient shortcuts for common operations.
# All targets that run Pulumi commands use the run-pulumi.sh wrapper
# which automatically handles token refresh and venv activation.
#
# Usage:
#   make help          - Show this help
#   make setup         - Initial setup (quickstart)
#   make preview       - Preview changes
#   make up            - Deploy resources (interactive)
#   make destroy       - Destroy all resources
#   make refresh       - Refresh state from Lagoon
#   make output        - Show stack outputs
#   make clean         - Kill port-forwards and clean up
#   make install       - Install provider in development mode

.PHONY: help setup preview up destroy refresh output clean port-forwards token check install

# Default target
help:
	@echo "Lagoon Pulumi Provider - Example Project"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Setup targets:"
	@echo "  setup         - Run quickstart script (first-time setup)"
	@echo "  check         - Check cluster health and connectivity"
	@echo "  port-forwards - Start kubectl port-forwards"
	@echo "  token         - Get a fresh OAuth token"
	@echo ""
	@echo "Pulumi targets (auto-refresh token):"
	@echo "  preview       - Preview pending changes"
	@echo "  up            - Deploy resources (interactive)"
	@echo "  destroy       - Destroy all resources (interactive)"
	@echo "  refresh       - Refresh state from Lagoon API"
	@echo "  output        - Show stack outputs"
	@echo ""
	@echo "Utility targets:"
	@echo "  install       - Install provider in development mode"
	@echo "  clean         - Kill port-forwards and clean temp files"
	@echo ""
	@echo "Note: Most targets use the run-pulumi.sh wrapper which"
	@echo "      automatically handles token refresh and venv activation."

# Initial setup using quickstart script
setup:
	@echo "Running quickstart setup..."
	@echo "Note: This script should be sourced to set environment variables:"
	@echo "  source ./scripts/quickstart.sh"
	@echo ""
	./scripts/quickstart.sh

# Check cluster health
check:
	./scripts/check-cluster-health.sh

# Start port-forwards
port-forwards:
	./scripts/setup-port-forwards.sh

# Get fresh token (must be sourced to export variables)
token:
	@echo "To get a token, run:"
	@echo "  source ./scripts/get-lagoon-token.sh"
	@echo ""
	@echo "Or use the run-pulumi.sh wrapper which handles tokens automatically."

# Preview changes
preview:
	./scripts/run-pulumi.sh preview

# Deploy resources
up:
	./scripts/run-pulumi.sh up

# Destroy all resources
destroy:
	./scripts/run-pulumi.sh destroy

# Refresh state from API
refresh:
	./scripts/run-pulumi.sh refresh

# Show stack outputs
output:
	./scripts/run-pulumi.sh stack output

# Clean up
clean:
	@echo "Killing port-forwards..."
	@pkill -f "port-forward.*lagoon-core" 2>/dev/null || true
	@echo "Done."

# Install provider in development mode
install:
	@echo "Installing provider in development mode..."
	@if [ -f ../../venv/bin/activate ]; then \
		. ../../venv/bin/activate && pip install -e ../..; \
	else \
		echo "Virtual environment not found at ../../venv"; \
		echo "Create it with: python3 -m venv ../../venv && source ../../venv/bin/activate && pip install -e ../.."; \
	fi
