# Makefile for MarkGo

# Variables
BINARY_NAME=markgo
MAIN_PATH=./cmd/server
BUILD_DIR=./build
DOCKER_IMAGE=markgo
DOCKER_TAG=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(shell git describe --tags --always --dirty)"
BUILD_FLAGS=-trimpath

.PHONY: help build build-all build-release clean test test-race coverage lint fmt run dev docker install tidy

# Default target
help: ## Show this help message
	@echo "MarkGo - Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build main server binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build all CLI tools (server, init, new, export)
	@echo "Building all tools..."
	@mkdir -p $(BUILD_DIR)
	@echo "  Building server..."
	@CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "  Building init..."
	@CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/init ./cmd/init
	@echo "  Building new-article..."
	@CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/new-article ./cmd/new-article
	@echo "  Building export..."
	@CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/export ./cmd/export
	@echo "✓ All builds complete"

build-release: ## Build for all platforms (Linux, macOS, Windows)
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)
	@echo "  Linux amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "  macOS amd64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "  macOS arm64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "  Windows amd64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✓ Release builds complete"

# Development targets
run: ## Run the server
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

dev: ## Run in development mode (restart manually for changes)
	@echo "Starting development server..."
	@echo "Note: Restart server manually to reload templates/config"
	ENVIRONMENT=development $(GOCMD) run $(MAIN_PATH)

install: ## Install server binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) $(MAIN_PATH)

# Testing targets
test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"
	@$(GOCMD) tool cover -func=coverage.out | grep total

# Code quality targets
fmt: ## Format code with gofmt
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	$(GOCMD) fmt ./...
	@echo "✓ Code formatted"

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run; \
		echo "✓ Lint complete"; \
	else \
		echo "Error: '$(GOLINT)' not found."; \
		echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Dependency management
tidy: ## Tidy and verify dependencies
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "✓ Dependencies tidied"

# Docker targets
docker: ## Build and run Docker container
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Running Docker container..."
	docker run -p 3000:3000 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

# Note: For benchmarks, profiling, and other advanced tasks, use go test directly:
#   Benchmarks:  go test -bench=. -benchmem ./...
#   CPU Profile: go test -cpuprofile=cpu.prof ./...
#   Mem Profile: go test -memprofile=mem.prof ./...
#   Profiling:   go tool pprof cpu.prof
