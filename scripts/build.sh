#!/bin/bash

# Exit on error
set -e

# Display header
echo "====================================="
echo "  kubectl-kanvas-snapshot Builder"
echo "====================================="

# Set environment variables for the build
export CGO_ENABLED=0

# Go to project root directory
cd "$(dirname "$0")/.."

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi
echo "Go version: $(go version)"

# Build binary
echo "Building kubectl-kanvas-snapshot binary..."
go build -o kubectl-kanvas-snapshot -ldflags="-s -w" .

# Make executable
chmod +x kubectl-kanvas-snapshot

# Show build info
echo "Build completed: $(pwd)/kubectl-kanvas-snapshot"
echo "Plugin size: $(ls -lh kubectl-kanvas-snapshot | awk '{print $5}')"

echo "====================================="
echo "To test the plugin, run:"
echo "./kubectl-kanvas-snapshot --help"
echo "=====================================" 