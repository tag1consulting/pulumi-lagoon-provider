#!/bin/bash
# Setup script for Lagoon test cluster

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== Lagoon Test Cluster Setup ==="
echo

# Check prerequisites
echo "Checking prerequisites..."

check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo "ERROR: $1 is not installed. Please install it first."
        exit 1
    fi
    echo "✓ $1 found"
}

check_command docker
check_command kind
check_command kubectl
check_command pulumi

# Check Docker is running
if ! docker info &> /dev/null; then
    echo "ERROR: Docker is not running. Please start Docker first."
    exit 1
fi
echo "✓ Docker is running"

echo
echo "All prerequisites met!"
echo

# Navigate to project directory
cd "$PROJECT_DIR"

# Create Python virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
echo "Activating virtual environment..."
source venv/bin/activate

# Install Python dependencies
echo "Installing Python dependencies..."
pip install -q --upgrade pip
pip install -q -r requirements.txt

# Initialize Pulumi stack if needed
if ! pulumi stack ls 2>/dev/null | grep -q "dev"; then
    echo "Initializing Pulumi stack 'dev'..."
    pulumi stack init dev
else
    echo "Using existing Pulumi stack 'dev'"
    pulumi stack select dev
fi

echo
echo "=== Setup complete! ==="
echo
echo "Next steps:"
echo "1. Review configuration in config/ directory"
echo "2. Run: pulumi preview"
echo "3. Run: pulumi up"
echo "4. Run: ./scripts/get-credentials.sh"
echo
