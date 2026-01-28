# Single-cluster Lagoon Example Makefile

.PHONY: help setup deploy preview up down clean clean-all status fix-migration-lock

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
	@echo "  make clean-all          - Delete cluster and remove Pulumi state"
	@echo ""
	@echo "Utilities:"
	@echo "  make pods               - List all pods"
	@echo "  make logs-core          - View Lagoon core logs"
	@echo "  make logs-remote        - View Lagoon remote logs"
	@echo "  make fix-migration-lock - Fix API migration lock (CrashLoopBackOff)"

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

# Full cleanup including Pulumi state
clean-all: clean
	@echo "Removing Pulumi state..."
	@rm -rf .pulumi 2>/dev/null || true
	@rm -f Pulumi.dev.yaml 2>/dev/null || true
	@echo "Full cleanup complete."

# Utility targets
pods:
	kubectl get pods -A --context kind-$(CLUSTER_NAME)

logs-core:
	kubectl logs -n lagoon-core -l app.kubernetes.io/name=lagoon-core --context kind-$(CLUSTER_NAME) -f --tail=100

logs-remote:
	kubectl logs -n lagoon -l app.kubernetes.io/name=lagoon-remote --context kind-$(CLUSTER_NAME) -f --tail=100

builds:
	kubectl get lagoonbuilds -A --context kind-$(CLUSTER_NAME)

# Fix migration lock if API pods are stuck in CrashLoopBackOff
# This clears the database migration lock and restarts the API pods
fix-migration-lock:
	@echo "Fixing Lagoon API migration lock..."
	@echo ""
	@echo "Step 1: Scale down API deployment to stop all pods..."
	@kubectl --context kind-$(CLUSTER_NAME) scale deployment -n lagoon-core lagoon-core-lagoon-core-api --replicas=0 || true
	@sleep 5
	@echo ""
	@echo "Step 2: Clear migration lock in database..."
	@kubectl --context kind-$(CLUSTER_NAME) exec -n lagoon-core lagoon-core-lagoon-core-api-db-0 -- \
		mysql -u root -pLag00n infrastructure -e \
		"DELETE FROM lagoon_migrations_lock WHERE \`index\` > 1; UPDATE lagoon_migrations_lock SET is_locked = 0; SELECT * FROM lagoon_migrations_lock;" \
		2>/dev/null || echo "  Could not connect to database (may not be running yet)"
	@echo ""
	@echo "Step 3: Scale API deployment back to 1 replica..."
	@kubectl --context kind-$(CLUSTER_NAME) scale deployment -n lagoon-core lagoon-core-lagoon-core-api --replicas=1 || true
	@echo ""
	@echo "Waiting for API pod to start..."
	@sleep 20
	@kubectl --context kind-$(CLUSTER_NAME) get pods -n lagoon-core | grep api
	@echo ""
	@echo "If the pod is Running (2/2), the migration lock has been fixed."
	@echo "Run 'make up' or 'make deploy' to continue deployment."
