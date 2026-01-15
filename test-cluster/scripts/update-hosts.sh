#!/bin/bash
# Script to update /etc/hosts with Lagoon test domains

HOSTS_ENTRIES="127.0.0.1 api.lagoon.test keycloak.lagoon.test ui.lagoon.test harbor.lagoon.test"
HOSTS_FILE="/etc/hosts"
MARKER="# Lagoon test cluster"

# Check if entries already exist
if grep -q "$MARKER" "$HOSTS_FILE" 2>/dev/null; then
    echo "Lagoon hosts entries already exist in $HOSTS_FILE"
    exit 0
fi

echo "Adding Lagoon hosts entries to $HOSTS_FILE"
echo "" | sudo tee -a "$HOSTS_FILE" > /dev/null
echo "$MARKER" | sudo tee -a "$HOSTS_FILE" > /dev/null
echo "$HOSTS_ENTRIES" | sudo tee -a "$HOSTS_FILE" > /dev/null

echo "Done! Added:"
echo "  $HOSTS_ENTRIES"
