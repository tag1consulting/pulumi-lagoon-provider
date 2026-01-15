#!/bin/bash
# Setup dnsmasq for wildcard *.lagoon.test domains in WSL

echo "Setting up dnsmasq for *.lagoon.test wildcard domains..."
echo ""

# Install dnsmasq
if ! command -v dnsmasq &> /dev/null; then
    echo "Installing dnsmasq..."
    sudo apt-get update
    sudo apt-get install -y dnsmasq
else
    echo "dnsmasq already installed"
fi

# Create dnsmasq configuration for lagoon.test
echo "Configuring dnsmasq for *.lagoon.test..."
sudo tee /etc/dnsmasq.d/lagoon.conf > /dev/null <<EOF
# Resolve all *.lagoon.test domains to 127.0.0.1
address=/lagoon.test/127.0.0.1

# Listen only on localhost
listen-address=127.0.0.1

# Don't read /etc/hosts
no-hosts

# Use Google DNS as upstream
server=8.8.8.8
server=8.8.4.4
EOF

# Restart dnsmasq
echo "Restarting dnsmasq..."
sudo systemctl restart dnsmasq
sudo systemctl enable dnsmasq

# Configure WSL to use local dnsmasq
echo "Configuring WSL DNS..."
sudo tee /etc/resolv.conf > /dev/null <<EOF
nameserver 127.0.0.1
nameserver 8.8.8.8
EOF

# Make resolv.conf immutable so WSL doesn't overwrite it
sudo chattr +i /etc/resolv.conf

echo ""
echo "=========================================="
echo "dnsmasq setup complete!"
echo "=========================================="
echo ""
echo "All *.lagoon.test domains now resolve to 127.0.0.1"
echo ""
echo "Test with:"
echo "  dig api.lagoon.test"
echo "  dig anything.lagoon.test"
echo "  curl http://api.lagoon.test/graphql"
echo ""
echo "IMPORTANT: Windows still needs hosts file entries for browser access!"
echo "Run: ./update-windows-hosts.ps1 in PowerShell as Administrator"
echo ""
