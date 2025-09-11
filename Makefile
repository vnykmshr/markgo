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

.PHONY: help build buildall clean test coverage lint fmt vet deps tidy run dev docker docker-build docker-run install new-article stress-test \
	generate-article validate-articles preview-articles count-articles backup-content \
	docs docs-serve changelog security-scan audit-deps code-complexity dead-code \
	pre-commit commit-check git-hooks health-check metrics monitor \
	profile-cpu profile-memory profile-trace quick-setup dev-reset full-test \
	stress-test-run stress-test-quick stress-test-full

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

stress-test: ## Build the stress-test CLI tool
	@echo "Building stress-test..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/stress-test ./cmd/stress-test
	@echo "Build complete: $(BUILD_DIR)/stress-test"

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

build-all: build new-article stress-test ## Build all cmd tools

build-dist: build-linux build-windows build-darwin ## Build for all platforms

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

stress-test-run: ## Run the stress test tool against running server
	@echo "Running stress test against http://localhost:3000..."
	@if [ ! -f $(BUILD_DIR)/stress-test ]; then $(MAKE) stress-test; fi
	@$(BUILD_DIR)/stress-test -url http://localhost:3000 -duration 1m -concurrency 10

stress-test-quick: ## Run quick stress test (30s, 5 concurrent users)
	@echo "Running quick stress test..."
	@if [ ! -f $(BUILD_DIR)/stress-test ]; then $(MAKE) stress-test; fi
	@$(BUILD_DIR)/stress-test -url http://localhost:3000 -duration 30s -concurrency 5 -output $(BUILD_DIR)/stress-test-results.json

stress-test-full: ## Run comprehensive stress test with report generation
	@echo "Running comprehensive stress test..."
	@if [ ! -f $(BUILD_DIR)/stress-test ]; then $(MAKE) stress-test; fi
	@$(BUILD_DIR)/stress-test -url http://localhost:3000 -duration 2m -concurrency 20 -output $(BUILD_DIR)/stress-test-results.json -verbose

benchmark-integration: ## Run integration benchmarks
	@echo "Running integration benchmarks..."
	$(GOTEST) -run=^$$ -bench=BenchmarkFullRequestFlow -benchmem ./internal/handlers
	$(GOTEST) -run=^$$ -bench=BenchmarkConcurrentArticleAccess -benchmem ./internal/handlers
	$(GOTEST) -run=^$$ -bench=BenchmarkSearchWithLargeDataset -benchmem ./internal/handlers

benchmark-memory: ## Run memory profiling benchmarks
	@echo "Running memory benchmarks..."
	$(GOTEST) -run=^$$ -bench=BenchmarkMemory -benchmem ./internal/handlers
	$(GOTEST) -run=^$$ -bench=BenchmarkBaseline -benchmem ./internal/handlers

benchmark-regression: ## Run performance regression tests
	@echo "Running performance regression tests..."
	@mkdir -p benchmarks/baseline
	$(GOTEST) -run=^$$ -bench=. -benchmem -count=5 ./... | tee benchmarks/current-benchmark.txt
	@echo "Benchmark results saved to benchmarks/current-benchmark.txt"

performance-report: ## Generate comprehensive performance report
	@echo "Generating performance report..."
	@./scripts/competitor-benchmark.sh || echo "Benchmark completed with warnings"

benchmark-all: benchmark benchmark-integration benchmark-memory ## Run all benchmarks

benchmark-ci: ## Run benchmarks for CI/CD pipeline
	@echo "Running CI benchmarks..."
	$(GOTEST) -run=^$$ -bench=BenchmarkFullRequestFlow -benchmem -count=3 ./internal/handlers
	$(GOTEST) -run=^$$ -bench=BenchmarkBaselineResourceUsage -benchmem ./internal/handlers

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
release: clean build-release-dist ## Create comprehensive release packages
	@echo "Creating release packages with archives and checksums..."
	@mkdir -p $(BUILD_DIR)/release
	
	# Create release archives for each platform
	@cd $(BUILD_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [ -f "$$binary" ]; then \
			platform=$$(echo $$binary | sed 's/$(BINARY_NAME)-//'); \
			archive_name="markgo-v$$(git describe --tags --always --dirty)-$$platform"; \
			if [[ "$$platform" == *"windows"* ]]; then \
				zip -j "release/$$archive_name.zip" "$$binary" ../README.md ../CHANGELOG.md ../LICENSE 2>/dev/null || \
				zip -j "release/$$archive_name.zip" "$$binary" 2>/dev/null; \
			else \
				tar -czf "release/$$archive_name.tar.gz" -C . "$$binary" -C .. README.md CHANGELOG.md LICENSE 2>/dev/null || \
				tar -czf "release/$$archive_name.tar.gz" -C . "$$binary" 2>/dev/null; \
			fi; \
			echo "Created archive: $$archive_name"; \
		fi; \
	done
	
	# Create checksums
	@cd $(BUILD_DIR)/release && \
	if command -v sha256sum > /dev/null; then \
		sha256sum * > checksums.txt; \
	elif command -v shasum > /dev/null; then \
		shasum -a 256 * > checksums.txt; \
	else \
		echo "Warning: No checksum utility found"; \
	fi
	
	@echo "✅ Release packages created in $(BUILD_DIR)/release/"
	@echo "📦 Contents:"
	@ls -la $(BUILD_DIR)/release/

# Enhanced build targets for release
build-release-dist: build-release-linux build-release-windows build-release-darwin ## Build release binaries for all platforms

build-release-linux: ## Build complete Linux release binaries
	@echo "Building Linux release binaries..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

build-release-windows: ## Build complete Windows release binaries  
	@echo "Building Windows release binaries..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe $(MAIN_PATH)

build-release-darwin: ## Build complete macOS release binaries
	@echo "Building macOS release binaries..."
	@mkdir -p $(BUILD_DIR)  
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

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

# Forking and customization targets
update-imports: ## Update import paths for forked repository (usage: make update-imports USERNAME=yourusername)
	@if [ -z "$(USERNAME)" ]; then \
		echo "Usage: make update-imports USERNAME=yourusername"; \
		echo "   Or: make update-imports USERNAME=github.com/yourusername/markgo"; \
		exit 1; \
	fi
	@./scripts/update-imports.sh $(USERNAME)



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

# Content Management targets
generate-article: ## Generate a new article using the new-article tool
	@echo "Generating new article..."
	@if [ ! -f $(BUILD_DIR)/new-article ]; then $(MAKE) new-article; fi
	@$(BUILD_DIR)/new-article --interactive

validate-articles: ## Validate all articles for proper frontmatter and structure
	@echo "Validating articles..."
	@./scripts/validate-articles.sh || echo "Article validation completed with warnings"

preview-articles: ## Preview articles locally (requires build)
	@echo "Starting article preview server..."
	@if [ ! -f $(BUILD_DIR)/$(BINARY_NAME) ]; then $(MAKE) build; fi
	@echo "Starting server on http://localhost:3000"
	@$(BUILD_DIR)/$(BINARY_NAME)

count-articles: ## Count articles and show statistics
	@echo "Article Statistics:"
	@echo "=================="
	@echo -n "Total articles: "
	@find articles -name "*.md" -o -name "*.markdown" -o -name "*.mdown" -o -name "*.mkd" | wc -l
	@echo -n "Draft articles: "
	@find articles -name "*.md" -o -name "*.markdown" -o -name "*.mdown" -o -name "*.mkd" -exec grep -l "draft: true" {} \; 2>/dev/null | wc -l || echo "0"
	@echo -n "Featured articles: "
	@find articles -name "*.md" -o -name "*.markdown" -o -name "*.mdown" -o -name "*.mkd" -exec grep -l "featured: true" {} \; 2>/dev/null | wc -l || echo "0"

backup-content: ## Backup articles and static content
	@echo "Backing up content..."
	@mkdir -p backups
	@tar -czf backups/content-backup-$(shell date +%Y%m%d-%H%M%S).tar.gz articles/ web/static/ || true
	@echo "Content backed up to backups/ directory"

# Documentation targets
docs: ## Generate comprehensive project documentation from Go source code
	@mkdir -p docs
	@./scripts/generate-docs.sh

docs-serve: ## Serve documentation locally (requires godoc)
	@echo "Serving documentation on http://localhost:6060"
	@if command -v godoc > /dev/null; then \
		godoc -http=:6060; \
	else \
		echo "Error: 'godoc' not found. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
		exit 1; \
	fi

changelog: ## Generate changelog from git commits
	@echo "Generating changelog..."
	@echo "# Changelog" > CHANGELOG.md
	@echo "" >> CHANGELOG.md
	@git log --pretty=format:"- %s (%h)" --reverse >> CHANGELOG.md
	@echo "Changelog generated: CHANGELOG.md"

# Security and Quality targets
security-scan: ## Run security vulnerability scan
	@echo "Running security scan..."
	@if command -v govulncheck > /dev/null; then \
		govulncheck ./...; \
	else \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	fi

audit-deps: ## Audit dependencies for known vulnerabilities
	@echo "Auditing dependencies..."
	@$(GOMOD) download
	@if command -v nancy > /dev/null; then \
		go list -json -m all | nancy sleuth; \
	else \
		echo "Note: Install nancy for enhanced dependency auditing: go install github.com/sonatypecommunity/nancy@latest"; \
		govulncheck ./...; \
	fi

code-complexity: ## Analyze code complexity
	@echo "Analyzing code complexity..."
	@if command -v gocyclo > /dev/null; then \
		gocyclo -over 10 .; \
	else \
		echo "Installing gocyclo..."; \
		go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; \
		gocyclo -over 10 .; \
	fi

dead-code: ## Find dead/unused code
	@echo "Finding dead code..."
	@if command -v deadcode > /dev/null; then \
		deadcode ./...; \
	else \
		echo "Installing deadcode..."; \
		go install golang.org/x/tools/cmd/deadcode@latest; \
		deadcode ./...; \
	fi

audit: security-scan audit-deps code-complexity dead-code ## Run all security checks

# Development Workflow targets
pre-commit: ## Run pre-commit checks (format, lint, test)
	@echo "Running pre-commit checks..."
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) lint
	@$(MAKE) test-short

commit-check: ## Validate commit message format
	@echo "Checking last commit message format..."
	@git log -1 --pretty=format:"%s" | grep -E "^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?: .{1,50}" || \
		(echo "Commit message should follow format: type(scope): description"; exit 1)

git-hooks: ## Install Git hooks for development
	@echo "Installing Git hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make pre-commit' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed"

# Monitoring and Health targets
health-check: ## Check application health endpoints
	@echo "Checking application health..."
	@curl -f http://localhost:3000/health || echo "Health check failed - is the server running?"

metrics: ## Show application metrics
	@echo "Fetching application metrics..."
	@curl -s http://localhost:3000/metrics || echo "Metrics unavailable - is the server running?"

monitor: ## Monitor application with basic stats
	@echo "Monitoring application (Ctrl+C to stop)..."
	@while true; do \
		echo "=== $(shell date) ==="; \
		curl -s http://localhost:3000/health | head -1; \
		echo "Memory usage: $(shell ps aux | grep '[m]arkgo' | awk '{print $$4}')%"; \
		sleep 5; \
	done

# Advanced Development targets
profile-cpu: ## Profile CPU usage (requires running server)
	@echo "Profiling CPU for 30 seconds..."
	@mkdir -p profiles
	@go tool pprof -http=:8080 http://localhost:3000/debug/pprof/profile?seconds=30

profile-memory: ## Profile memory usage (requires running server)
	@echo "Profiling memory usage..."
	@mkdir -p profiles
	@go tool pprof -http=:8080 http://localhost:3000/debug/pprof/heap

profile-trace: ## Generate execution trace (requires running server)
	@echo "Generating execution trace for 10 seconds..."
	@mkdir -p profiles
	@curl http://localhost:3000/debug/pprof/trace?seconds=10 > profiles/trace.out
	@go tool trace profiles/trace.out

quick-setup: ## Quick development setup (deps + tools + git hooks)
	@echo "Quick development setup..."
	@$(MAKE) deps
	@$(MAKE) install-dev-tools
	@$(MAKE) git-hooks
	@$(MAKE) setup
	@echo "Quick setup complete! Run 'make dev' to start development server."

# All-in-one targets
dev-reset: ## Reset development environment (clean + setup)
	@echo "Resetting development environment..."
	@$(MAKE) clean
	@$(MAKE) setup
	@$(MAKE) deps
	@echo "Development environment reset complete"

full-test: ## Run comprehensive test suite
	@echo "Running comprehensive test suite..."
	@$(MAKE) pre-commit
	@$(MAKE) test-race
	@$(MAKE) benchmark-ci
	@$(MAKE) security-scan
	@echo "Full test suite completed"

# Help is default
.DEFAULT_GOAL := help
