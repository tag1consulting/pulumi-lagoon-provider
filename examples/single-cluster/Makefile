# Single-cluster Lagoon Example Makefile

.PHONY: help setup deploy preview up down clean status

help:
	@echo "Single-cluster Lagoon Example"
	@echo ""
	@echo "Quick Start:"
	@echo "  make deploy             - Full deployment"
	@echo "  make preview            - Preview changes"
	@echo ""
	@echo "Commands:"
	@echo "  make setup              - Initialize Pulumi stack"
	@echo "  make up                 - Deploy resources"
	@echo "  make down               - Destroy resources"
	@echo "  make status             - Show stack outputs"
	@echo "  make clean              - Delete Kind cluster"
	@echo ""
	@echo "Utilities:"
	@echo "  make pods               - List all pods"
	@echo "  make logs-core          - View Lagoon core logs"
	@echo "  make logs-remote        - View Lagoon remote logs"

# Use shared venv from repository root
VENV := ../../venv
PYTHON := $(VENV)/bin/python
PULUMI := pulumi

# Default cluster name
CLUSTER_NAME ?= lagoon

setup:
	@if [ ! -d "$(VENV)" ]; then \
		echo "Creating virtual environment..."; \
		python3 -m venv $(VENV); \
		$(VENV)/bin/pip install -r ../multi-cluster/requirements.txt; \
	fi
	@$(PULUMI) stack init dev 2>/dev/null || echo "Stack already exists"

deploy: setup
	$(PULUMI) up --yes

preview:
	$(PULUMI) preview

up:
	$(PULUMI) up

down:
	$(PULUMI) destroy --yes

status:
	$(PULUMI) stack output

clean:
	@echo "Deleting Kind cluster..."
	kind delete cluster --name $(CLUSTER_NAME) 2>/dev/null || true

# Utility targets
pods:
	kubectl get pods -A --context kind-$(CLUSTER_NAME)

logs-core:
	kubectl logs -n lagoon-core -l app.kubernetes.io/name=lagoon-core --context kind-$(CLUSTER_NAME) -f --tail=100

logs-remote:
	kubectl logs -n lagoon -l app.kubernetes.io/name=lagoon-remote --context kind-$(CLUSTER_NAME) -f --tail=100

builds:
	kubectl get lagoonbuilds -A --context kind-$(CLUSTER_NAME)
