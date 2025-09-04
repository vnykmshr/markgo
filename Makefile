# Makefile for MarkGo Engine

# Variables
BINARY_NAME=markgo
MAIN_PATH=./cmd/server

BUILD_DIR=./build
DOCKER_IMAGE=markgo
DOCKER_TAG=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(shell git describe --tags --always --dirty)"
BUILD_FLAGS=-trimpath

.PHONY: help build clean test coverage lint fmt vet deps tidy run dev docker docker-build docker-run install new-article

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

new-article: ## Build the new-article CLI tool
	@echo "Building new-article..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/new-article ./cmd/new-article
	@echo "Build complete: $(BUILD_DIR)/new-article"

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-darwin: ## Build for macOS
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-all: build-linux build-windows build-darwin ## Build for all platforms



# Development targets
run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

dev: ## Run with hot reload (requires air)
	@echo "Starting development server with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Error: 'air' not found. Install with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi

install: ## Install the application
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) $(MAIN_PATH)



# Testing targets
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

test-short: ## Run short tests
	@echo "Running short tests..."
	$(GOTEST) -short -v ./...

coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Code quality targets
fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	$(GOCMD) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

lint: ## Run linter
	@echo "Running linter..."
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run; \
	else \
		echo "Error: '$(GOLINT)' not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

# Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOGET) -d ./...

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	$(GOMOD) vendor

update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	DOCKER_BUILDKIT=1 docker build -f deployments/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@if [ -f .env ]; then \
		docker run -p 3000:3000 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG); \
	else \
		docker run -p 3000:3000 $(DOCKER_IMAGE):$(DOCKER_TAG); \
	fi

docker: docker-build docker-run ## Build and run Docker container

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	@if [ -f deployments/docker-compose.yml ]; then \
		docker-compose -f deployments/docker-compose.yml up -d; \
	else \
		echo "Error: docker-compose.yml not found in deployments/"; \
		exit 1; \
	fi

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	@if [ -f deployments/docker-compose.yml ]; then \
		docker-compose -f deployments/docker-compose.yml down; \
	else \
		echo "Error: docker-compose.yml not found in deployments/"; \
		exit 1; \
	fi

docker-logs: ## Show Docker container logs
	@echo "Showing Docker logs..."
	docker logs $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker images and containers..."
	docker container prune -f
	docker image prune -f
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true

docker-shell: ## Open shell in running container
	@echo "Opening shell in $(DOCKER_IMAGE) container..."
	docker exec -it $(shell docker ps -q --filter ancestor=$(DOCKER_IMAGE):$(DOCKER_TAG)) /bin/sh

docker-inspect: ## Inspect Docker image
	@echo "Inspecting Docker image..."
	docker inspect $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image to registry..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-build-multiarch: ## Build multi-architecture Docker image
	@echo "Building multi-architecture Docker image..."
	docker buildx build --platform linux/amd64,linux/arm64 -f deployments/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-build-push: ## Build and push Docker image
	@echo "Building and pushing Docker image..."
	DOCKER_BUILDKIT=1 docker build -f deployments/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) . && \
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-dev: ## Run Docker container in development mode with volume mounts
	@echo "Running Docker container in development mode..."
	docker run -p 3000:3000 \
		-v $(PWD)/articles:/app/articles \
		-v $(PWD)/web:/app/web \
		-v $(PWD)/.env:/app/.env \
		--name markgo-dev \
		--rm \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Deployment targets
release: clean build-all ## Create release builds
	@echo "Creating release..."
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BUILD_DIR)/$(BINARY_NAME)-* $(BUILD_DIR)/release/
	@echo "Release builds created in $(BUILD_DIR)/release/"

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

migrate-articles: ## Copy articles from Node.js version
	@echo "Migrating articles from Node.js version..."
	@if [ -d "../old-blog/articles" ]; then \
		cp -r ../old-blog/articles/* ./articles/; \
		echo "Articles migrated successfully"; \
	else \
		echo "Source articles directory not found: ../old-blog/articles"; \
	fi

setup: ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Please edit .env file with your configuration"
	@$(GOMOD) download
	@mkdir -p articles web/static web/templates
	@echo "Setup complete. Edit .env file and run 'make run' to start"

# Statistics
stats: ## Show project statistics
	@echo "Project Statistics:"
	@echo "==================="
	@echo -n "Total Go files: "
	@find . -name "*.go" -not -path "./vendor/*" | wc -l
	@echo -n "Lines of code: "
	@find . -name "*.go" -not -path "./vendor/*" -exec cat {} \; | wc -l
	@echo -n "Test files: "
	@find . -name "*_test.go" -not -path "./vendor/*" | wc -l
	@echo "Dependencies:"
	@$(GOMOD) graph | wc -l

# Development tools installation
install-dev-tools: ## Install development tools
	@echo "Installing development tools..."
	@$(GOCMD) install github.com/air-verse/air@latest
	@$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"



# Production targets
prod-build: ## Build for production
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) \
		-trimpath \
		-ldflags "-s -w -X main.version=$(shell git describe --tags --always --dirty) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

systemd-install: prod-build ## Install as systemd service
	@echo "Installing systemd service..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo cp deployments/markgo.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@sudo systemctl enable markgo
	@echo "Systemd service installed. Start with: sudo systemctl start markgo"

# Help is default
.DEFAULT_GOAL := help
