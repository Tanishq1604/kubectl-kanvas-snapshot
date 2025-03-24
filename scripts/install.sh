#!/bin/bash

# Exit on error
set -e

# Display header
echo "====================================="
echo "kubectl-kanvas-snapshot Installer"
echo "====================================="

# Go to project root directory
cd "$(dirname "$0")/.."

# Check if the binary exists, build it if not
if [ ! -f "./kubectl-kanvas-snapshot" ]; then
    echo "Binary not found. Building it first..."
    ./scripts/build.sh
fi

# Install to path
INSTALL_DIR="/usr/local/bin"
echo "Installing to $INSTALL_DIR..."

# Check if we have permission to write to the directory
if [ -w "$INSTALL_DIR" ]; then
    cp kubectl-kanvas-snapshot "$INSTALL_DIR"
else
    # Use sudo if we don't have permission
    echo "Need sudo permission to install to $INSTALL_DIR"
    sudo cp kubectl-kanvas-snapshot "$INSTALL_DIR"
fi

echo "====================================="
echo "Installation complete!"
echo "You can now use the plugin with:"
echo "kubectl kanvas-snapshot"
echo "=====================================" 