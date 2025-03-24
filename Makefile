# Makefile for kubectl-kanvas-snapshot

PLUGIN_NAME=kubectl-kanvas-snapshot
PLUGIN_VERSION=0.1.0
BINARY_NAME=kubectl-kanvas-snapshot
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_LINT=golangci-lint
GO_FMT=$(GO_CMD) fmt
GOPATH=$(shell go env GOPATH)
BINARY_DIR=$(GOPATH)/bin

# Build settings
LDFLAGS=-ldflags="-s -w"
BUILD_DIR=./build

.PHONY: all build clean install test lint fmt help

all: build

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@$(GO_BUILD) $(LDFLAGS) -o $(BINARY_NAME) .

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -rf dist/

install: ## Install the plugin to the kubectl plugins directory
	@echo "Installing $(BINARY_NAME)..."
	@./scripts/install.sh

test: ## Run the tests
	@echo "Running tests..."
	@$(GO_TEST) -v ./...

lint: ## Run the linter
	@echo "Running linter..."
	@$(GO_LINT) run ./...

fmt: ## Format the code
	@echo "Formatting code..."
	@$(GO_FMT) ./...

build-all: ## Build for all platforms
	@echo "Building for all supported platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 .
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 .
	GOOS=windows GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe .
	@echo "Builds available in $(BUILD_DIR)/"

run-example: build ## Run the plugin with a sample manifest
	@echo "Running the plugin with a sample manifest..."
	./$(BINARY_NAME) -f samples/deployment.yaml --name "sample-deployment"

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help