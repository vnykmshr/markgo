# MarkGo Technical Architecture

**Version:** 2.3.0
**Last Updated:** February 2026

## Overview

MarkGo follows a **layered architecture** with clean separation of concerns. The design prioritizes simplicity, performance, and maintainability over complex patterns.

## Architecture Layers

```
┌─────────────────────────────────────────┐
│         HTTP Layer (Gin)                │
│   Routes, Middleware, Static Files      │
├─────────────────────────────────────────┤
│       Application Layer (Handlers)      │
│   Request/Response, Template Rendering  │
├─────────────────────────────────────────┤
│      Business Logic (Services)          │
│   Articles, Search, Templates, Email    │
├─────────────────────────────────────────┤
│      Infrastructure (Config, Cache)     │
│   File System, Configuration, Logging   │
└─────────────────────────────────────────┘
```

## Core Components

### HTTP Layer

**Router:** Gin web framework

**Middleware Pipeline:**
- Recovery (panic handling)
- Logging (request/response)
- Security headers (HSTS, CSP, XSS protection)
- CORS (cross-origin requests)
- Rate limiting (request throttling)

**Routes:**
- Static files: `/static/*`, `/favicon.ico`
- Public: `/`, `/articles/*`, `/tags`, `/categories`, `/search`, `/rss`
- Admin: `/admin/*` (metrics, reload, cache management)
- Health: `/health`, `/metrics`

### Application Layer (Handlers)

Handlers bridge HTTP requests and business logic.

**Key Handlers:**
- `ArticleHandler`: Article viewing, listing, filtering
- `SearchHandler`: Full-text search queries
- `AdminHandler`: Administrative functions
- `HealthHandler`: Health checks and metrics

All handlers use dependency injection for testability.

### Business Logic Layer (Services)

**ArticleService:**
- Loads Markdown files with YAML frontmatter
- Parses and caches parsed articles
- Provides article retrieval, filtering, pagination
- Manages tags and categories

**SearchService:**
- Full-text search with simple ranking
- Query processing and result filtering
- Integrates with article index

**TemplateService:**
- Loads and renders HTML templates
- Provides 30+ custom template functions
- Handles template caching

**EmailService:**
- Contact form processing
- SMTP email delivery
- Rate limiting and spam protection

### Infrastructure Layer

**Configuration:**
- Environment variable based
- Validation on startup
- Hierarchical config structure

**File System:**
- Markdown article storage
- Template and static file serving
- Content scanning and monitoring

**Caching:**
- In-memory caching for parsed articles
- Template render caching
- Configurable TTL and size limits

**Logging:**
- Structured logging with slog
- Request/response logging
- Error tracking

## Request Flow

```
HTTP Request
    ↓
Middleware Chain
    ↓
Handler
    ↓
Service Layer
    ├─→ Check Cache → Return if hit
    └─→ Process Request → Update Cache
    ↓
Response
```

### Article Processing Pipeline

```
Markdown File
    ↓
Read File → Parse YAML Frontmatter
    ↓
Parse Markdown → Render HTML
    ↓
Create Article Object
    ↓
Cache & Index
```

## Design Patterns

### Dependency Injection

Services are injected into handlers via constructors:

```go
func NewArticleHandler(
    config *config.Config,
    logger *slog.Logger,
    templates *services.TemplateService,
    articles services.ArticleServiceInterface,
    // ... other dependencies
) *ArticleHandler
```

### Interface-Based Design

Services implement interfaces for testing and flexibility:

```go
type ArticleServiceInterface interface {
    GetAllArticles() []*models.Article
    GetArticleBySlug(slug string) (*models.Article, error)
    // ... other methods
}
```

### Repository Pattern

Article storage is abstracted behind a repository interface, currently implemented with file system storage.

## Configuration

Configuration is environment-driven with sensible defaults:

```go
// Server
PORT=3000
BASE_URL=http://localhost:3000

// Blog
BLOG_TITLE=My Blog
BLOG_AUTHOR=Author Name

// Features
CACHE_ENABLED=true
SEARCH_ENABLED=true
RSS_ENABLED=true

// Performance
CACHE_TTL=3600
RATE_LIMIT_ENABLED=true
```

See [configuration.md](configuration.md) for complete options.

## Testing Strategy

MarkGo follows a pragmatic testing approach:

**Unit Tests:** Core business logic, utilities, parsers
**Integration Tests:** Handlers with mock services
**Coverage Target:** Focus on critical paths over percentages

Current coverage: ~46% overall, with focused coverage on handlers, article services, and content processing.

Test files follow the `*_test.go` convention and use:
- Standard library `testing` package
- `testify` for assertions
- `httptest` for HTTP testing
- Mock interfaces for service isolation

## Performance Characteristics

**Startup Time:** < 1 second
**Memory Usage:** ~30MB typical workload
**Binary Size:** ~27MB
**Response Time:**
- Cached: < 5ms
- Uncached: < 50ms (first request)

**Scalability:**
- Single binary handles 1000+ req/s
- Horizontal scaling via load balancer
- Stateless design (except in-memory cache)

## Security

**Input Validation:**
- All user input sanitized
- XSS protection via HTML escaping
- CSRF token support for forms

**Security Headers:**
- Strict-Transport-Security
- X-Frame-Options: SAMEORIGIN
- X-Content-Type-Options: nosniff
- Content-Security-Policy

**Rate Limiting:**
- Configurable per endpoint
- Protection against abuse
- Contact form throttling

## Deployment Architecture

**Binary Deployment:**
```
markgo binary
├── Config (.env)
├── Content (articles/)
├── Templates (web/templates/)
└── Static files (web/static/)
```

**Docker Deployment:**
- Multi-stage build
- Minimal base image (scratch)
- Volume mounts for content

**Reverse Proxy (Production):**
```
Client → Nginx/Caddy → MarkGo
         (TLS, caching, compression)
```

See [deployment.md](deployment.md) for detailed deployment strategies.

## CLI Structure

Unified CLI with subcommands (since v2.1.0):

```
markgo/
└── cmd/markgo/          # Main entry point
    └── main.go

internal/commands/       # Command implementations
├── serve/              # Server command (default)
├── init/               # Initialize new blog
├── new/                # Create new article
└── export/             # Static site export
```

## Project Structure

```
markgo/
├── cmd/markgo/              # CLI entry point
├── internal/                # Private application code
│   ├── commands/            # CLI commands
│   ├── config/              # Configuration
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # HTTP middleware
│   ├── models/              # Data structures
│   └── services/            # Business logic
├── web/                     # Frontend assets
│   ├── static/              # CSS, JS, images
│   └── templates/           # HTML templates
├── articles/                # Markdown content
├── deployments/             # Docker, systemd configs
└── docs/                    # Documentation
```

## Evolution: v1.0 → v2.1.0

**Major Simplifications:**

v2.0.0:
- Removed 6,000+ lines of over-engineering
- Eliminated unnecessary object pools
- Simplified caching layers

v2.1.0:
- Consolidated 5 binaries into unified CLI
- Simplified SEO service (stateless)
- Removed template hot-reload (fsnotify)
- Consolidated middleware
- Reduced admin bloat

**Philosophy:** "Less code, same functionality"

## Further Reading

- [Getting Started](GETTING-STARTED.md) - Setup and usage
- [Configuration](configuration.md) - All config options
- [API Documentation](API.md) - HTTP endpoints
- [Deployment](deployment.md) - Production deployment
- [Runbook](RUNBOOK.md) - Operations guide

---

**Last Updated:** February 2026 (v2.3.0)
**Questions?** See [GitHub Issues](https://github.com/vnykmshr/markgo/issues)
