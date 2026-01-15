#!/bin/bash
# Fix RabbitMQ password mismatch for lagoon-build-deploy
#
# Usage:
#   ./scripts/fix-rabbitmq-password.sh
#
# This fixes a common issue where the lagoon-build-deploy secret has
# a placeholder password instead of the actual RabbitMQ broker password.
#
# Symptoms:
#   - lagoon-build-deploy pod in CrashLoopBackOff
#   - Log error: "username or password not allowed" (403)
#   - "Failed to initialize message queue manager"

set -e

CONTEXT="${KUBE_CONTEXT:-kind-lagoon-test}"
NAMESPACE="${LAGOON_NAMESPACE:-lagoon}"

echo "Checking RabbitMQ password configuration..."
echo "Context: $CONTEXT"
echo ""

# Get current passwords
BUILD_DEPLOY_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret lagoon-build-deploy -o jsonpath='{.data.RABBITMQ_PASSWORD}' 2>/dev/null | base64 -d)
BROKER_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret lagoon-core-broker -o jsonpath='{.data.RABBITMQ_PASSWORD}' 2>/dev/null | base64 -d)

echo "lagoon-build-deploy password: ${BUILD_DEPLOY_PASSWORD:0:10}..."
echo "lagoon-core-broker password:  ${BROKER_PASSWORD:0:10}..."
echo ""

if [ "$BUILD_DEPLOY_PASSWORD" = "$BROKER_PASSWORD" ]; then
    echo "Passwords match. No fix needed."
    exit 0
fi

echo "Password mismatch detected. Updating lagoon-build-deploy secret..."

# Get the correct password (base64 encoded)
CORRECT_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret lagoon-core-broker -o jsonpath='{.data.RABBITMQ_PASSWORD}')

# Patch the secret
kubectl --context "$CONTEXT" -n "$NAMESPACE" patch secret lagoon-build-deploy \
    -p "{\"data\":{\"RABBITMQ_PASSWORD\":\"$CORRECT_PASSWORD\"}}"

echo "Secret updated."
echo ""

# Find and restart the build-deploy pod
POD_NAME=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get pods -l app.kubernetes.io/name=lagoon-build-deploy -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ -n "$POD_NAME" ]; then
    echo "Restarting pod: $POD_NAME"
    kubectl --context "$CONTEXT" -n "$NAMESPACE" delete pod "$POD_NAME"

    echo "Waiting for new pod to start..."
    sleep 5

    # Check new pod status
    NEW_POD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get pods -l app.kubernetes.io/name=lagoon-build-deploy -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    STATUS=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get pod "$NEW_POD" -o jsonpath='{.status.phase}' 2>/dev/null)

    echo "New pod: $NEW_POD"
    echo "Status: $STATUS"

    if [ "$STATUS" = "Running" ]; then
        echo ""
        echo "Fix successful! Checking logs..."
        kubectl --context "$CONTEXT" -n "$NAMESPACE" logs "$NEW_POD" --tail=5
    fi
else
    echo "No lagoon-build-deploy pod found."
fi
