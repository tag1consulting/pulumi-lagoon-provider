#!/bin/bash
# Master script to update hosts files on both WSL and Windows

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================"
echo "Lagoon Test Cluster - Hosts File Setup"
echo "========================================"
echo ""

# Update WSL /etc/hosts
echo "1. Updating WSL /etc/hosts..."
"$SCRIPT_DIR/update-hosts.sh"
echo ""

# Update Windows hosts file
echo "2. Updating Windows hosts file..."
echo "   This requires Administrator privileges on Windows."
echo ""

# Check if we're in WSL
if grep -qi microsoft /proc/version 2>/dev/null; then
    echo "   Detected WSL environment."
    echo ""
    echo "   To update Windows hosts file, run this command in Windows PowerShell AS ADMINISTRATOR:"
    echo ""

    # Convert WSL path to Windows path
    WINDOWS_PATH=$(wslpath -w "$SCRIPT_DIR/update-windows-hosts.ps1")

    echo "   powershell.exe -ExecutionPolicy Bypass -File \"$WINDOWS_PATH\""
    echo ""
    echo "   Or manually add these entries to C:\\Windows\\System32\\drivers\\etc\\hosts:"
    echo "   127.0.0.1 api.lagoon.test keycloak.lagoon.test ui.lagoon.test harbor.lagoon.test"
    echo ""
else
    echo "   Not running in WSL, skipping Windows hosts file update."
    echo ""
fi

echo "========================================"
echo "Setup Complete!"
echo "========================================"
echo ""
echo "WSL DNS Resolution: ✓"
echo "Windows DNS Resolution: Manual step required (see above)"
echo ""
echo "After updating Windows hosts file, you can access:"
echo "  - From WSL: curl http://api.lagoon.test/graphql"
echo "  - From Windows Browser: http://ui.lagoon.test"
echo ""
