#!/bin/bash
# Fix RabbitMQ password mismatch for lagoon-remote
#
# Usage:
#   ./scripts/fix-rabbitmq-password.sh
#   LAGOON_PRESET=multi-prod ./scripts/fix-rabbitmq-password.sh
#
# This fixes a common issue where the lagoon-remote secret has
# a placeholder password instead of the actual RabbitMQ broker password.
#
# Symptoms:
#   - lagoon-remote-kubernetes-build-deploy pod in CrashLoopBackOff
#   - Log error: "username or password not allowed" (403)
#   - "Failed to initialize message queue manager"
#
# See scripts/common.sh for configuration options.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

echo "Checking RabbitMQ password configuration..."
echo "  Context:   $KUBE_CONTEXT"
echo "  Namespace: $LAGOON_NAMESPACE"
echo ""

# Verify required secrets are configured
if [ -z "$REMOTE_SECRET" ] || [ -z "$BROKER_SECRET" ]; then
    log_error "REMOTE_SECRET and BROKER_SECRET must be configured"
    print_config
    exit 1
fi

# Get current passwords
REMOTE_PASSWORD=$(get_secret_value "$REMOTE_SECRET" "RABBITMQ_PASSWORD")
BROKER_PASSWORD=$(get_secret_value "$BROKER_SECRET" "RABBITMQ_PASSWORD")

if [ -z "$BROKER_PASSWORD" ]; then
    log_error "Could not get broker password from $BROKER_SECRET"
    exit 1
fi

echo "Remote secret ($REMOTE_SECRET): ${REMOTE_PASSWORD:0:10}..."
echo "Broker secret ($BROKER_SECRET): ${BROKER_PASSWORD:0:10}..."
echo ""

if [ "$REMOTE_PASSWORD" = "$BROKER_PASSWORD" ]; then
    log_info "Passwords match. No fix needed."
    exit 0
fi

log_warn "Password mismatch detected. Updating $REMOTE_SECRET..."

# Get the correct password (base64 encoded)
CORRECT_PASSWORD=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get secret "$BROKER_SECRET" \
    -o jsonpath='{.data.RABBITMQ_PASSWORD}')

# Patch the secret
kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" patch secret "$REMOTE_SECRET" \
    -p "{\"data\":{\"RABBITMQ_PASSWORD\":\"$CORRECT_PASSWORD\"}}"

log_info "Secret updated."
echo ""

# Find and restart the build-deploy pod
if [ -n "$REMOTE_DEPLOYMENT" ]; then
    POD_NAME=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get pods \
        -l "app.kubernetes.io/name=$REMOTE_DEPLOYMENT" \
        -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)

    # Try alternative label if first one fails
    if [ -z "$POD_NAME" ]; then
        POD_NAME=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get pods \
            -l "app.kubernetes.io/instance=${REMOTE_DEPLOYMENT%-kubernetes-build-deploy}" \
            -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
    fi

    if [ -n "$POD_NAME" ]; then
        log_info "Restarting pod: $POD_NAME"
        kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" delete pod "$POD_NAME"

        echo "Waiting for new pod to start..."
        sleep 5

        # Check new pod status
        NEW_POD=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get pods \
            -l "app.kubernetes.io/name=$REMOTE_DEPLOYMENT" \
            -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

        if [ -n "$NEW_POD" ]; then
            STATUS=$(kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" get pod "$NEW_POD" \
                -o jsonpath='{.status.phase}' 2>/dev/null || echo "Unknown")
            echo "New pod: $NEW_POD"
            echo "Status: $STATUS"

            if [ "$STATUS" = "Running" ]; then
                echo ""
                log_info "Fix successful! Checking logs..."
                kubectl --context "$KUBE_CONTEXT" -n "$LAGOON_NAMESPACE" logs "$NEW_POD" --tail=5 2>/dev/null || true
            fi
        fi
    else
        log_warn "No $REMOTE_DEPLOYMENT pod found."
    fi
fi
