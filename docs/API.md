# MarkGo API Documentation
Generated on: Thu Sep 11 07:35:12 +0545 2025

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

#### [Init Tool](./cmd-init-package.md)
Quick blog initialization tool with interactive setup wizard.

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
