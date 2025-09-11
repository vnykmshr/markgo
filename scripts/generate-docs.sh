#!/bin/bash

# MarkGo Documentation Generator
# Generates comprehensive API documentation from Go source code

set -e

echo "🚀 Generating comprehensive MarkGo documentation..."

# Core packages
PACKAGES=(
    "internal/models:models-package.md"
    "internal/config:config-package.md" 
    "internal/services:services-package.md"
    "internal/handlers:handlers-package.md"
    "internal/middleware:middleware-package.md"
    "internal/errors:errors-package.md"
    "cmd/server:cmd-server-package.md"
    "cmd/new-article:cmd-new-article-package.md"
    "cmd/stress-test:stress-test-package.md"
)

# Generate package documentation
for package_info in "${PACKAGES[@]}"; do
    IFS=":" read -r package file <<< "$package_info"
    echo "  📦 Generating $package documentation..."
    go doc -all ./$package > docs/$file
done

# Generate API documentation
cat > docs/API.md << 'EOF'
# MarkGo API Documentation
Generated on: $(date)

## Overview

MarkGo is a high-performance, file-based blog engine built with Go. This documentation provides a comprehensive overview of the internal APIs, data structures, and interfaces.

## Architecture

MarkGo follows a clean architecture pattern with the following layers:

- **Models**: Core data structures and domain objects
- **Services**: Business logic and data processing
- **Handlers**: HTTP request handling and routing
- **Middleware**: Request processing pipeline
- **Config**: Configuration management

## Package Documentation

### Core Packages

#### [Models Package](./models-package.md)
Core data structures including Article, ContactMessage, SearchResult, Feed, and Pagination.

#### [Config Package](./config-package.md)
Configuration management including environment-specific settings and validation.

#### [Services Package](./services-package.md)
Business logic services including article management, search, email, and templates.

#### [Handlers Package](./handlers-package.md)
HTTP request handlers for web interface, admin functionality, and API endpoints.

#### [Middleware Package](./middleware-package.md)
HTTP middleware components including security, rate limiting, and performance monitoring.

#### [Errors Package](./errors-package.md)
Domain-specific error handling including structured error types and utilities.

### Command Line Tools

#### [MarkGo Server](./cmd-server-package.md)
Main web server application with HTTP configuration and graceful shutdown.

#### [New Article Tool](./cmd-new-article-package.md)
Interactive article creation CLI tool with templates and validation.

#### [Stress Test Tool](./stress-test-package.md)
Performance testing tool with automatic URL discovery and validation.

## Performance Targets

| Metric | Target | Purpose |
|--------|--------|---------|
| Throughput | ≥1000 req/s | Competitive advantage vs Ghost |
| 95th Percentile | <50ms | 4x faster than Ghost ~200ms |
| Average Response | <30ms | Excellent user experience |
| Error Rate | <1% | Production reliability |
| Success Rate | >99% | Production readiness |

*This documentation is automatically generated from Go source code comments.*
EOF

# Generate documentation index
cat > docs/README.md << 'EOF'
# MarkGo Documentation Index

Complete documentation for the MarkGo high-performance blog engine.

## Package Documentation (Generated from Source Code)

| Package | Description | Documentation |
|---------|-------------|---------------|
| **Models** | Core data structures with memory optimization | [models-package.md](./models-package.md) |
| **Config** | Configuration management and validation | [config-package.md](./config-package.md) |
| **Services** | Business logic and service layer | [services-package.md](./services-package.md) |
| **Handlers** | HTTP request handlers and routing | [handlers-package.md](./handlers-package.md) |
| **Middleware** | HTTP middleware pipeline components | [middleware-package.md](./middleware-package.md) |
| **Errors** | Domain-specific error handling | [errors-package.md](./errors-package.md) |

## Command Line Tools

| Tool | Purpose | Documentation |
|------|---------|---------------|
| **Server** | Main MarkGo web server | [cmd-server-package.md](./cmd-server-package.md) |
| **New Article** | Interactive article creation tool | [cmd-new-article-package.md](./cmd-new-article-package.md) |
| **Stress Test** | Performance testing and validation | [stress-test-package.md](./stress-test-package.md) |

## Documentation Generation

```bash
# Regenerate all documentation
make docs

# Serve documentation locally
make docs-serve
```

*Generated on: $(date)*
EOF

# Replace date placeholders
sed -i '' "s/\$(date)/$(date)/g" docs/API.md docs/README.md

echo "✅ Documentation generation complete!"
echo "   📁 Generated $(ls docs/*.md | wc -l | tr -d ' ') documentation files"
echo "   📊 Total size: $(du -sh docs/ | cut -f1)"
echo "   🔗 View: docs/README.md or docs/API.md"