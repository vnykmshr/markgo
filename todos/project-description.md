# Project: MarkGo Engine
A modern, high-performance file-based blog engine built with Go. MarkGo combines the simplicity of file-based content management with the power of a dynamic web server, delivering blazing-fast performance and developer-friendly workflows.

## Features
- **File-based Content Management**: Uses Markdown files with YAML front matter stored in `/articles/`
- **Full-text Search**: Search functionality with relevance scoring across blog content  
- **SEO Optimization**: Structured data, sitemaps, RSS/JSON feeds, and proper meta tags
- **Responsive Web Interface**: Modern HTML templates with CSS for clean, responsive design
- **CLI Article Creator**: Command-line tool for creating new blog posts
- **Admin Interface**: Protected admin endpoints for cache management and statistics
- **Contact Form**: Email integration for visitor contact functionality
- **High Performance**: Built-in caching system and optimized for fast response times

## Tech Stack
- **Go 1.24.4** - Primary backend language
- **Gin v1.10.1** - HTTP web framework
- **Goldmark v1.7.12** - Markdown processing
- **testify v1.10.0** - Testing framework
- **Docker** - Containerization with multi-stage builds
- **Nginx** - Reverse proxy and static file serving
- **Air** - Hot-reload development server

## Structure
- `/cmd/server/` - Main web server application entry point
- `/cmd/new-article/` - CLI tool for creating articles
- `/internal/` - Private application code (services, handlers, models, config)
- `/web/` - Frontend assets and HTML templates
- `/articles/` - Blog content as Markdown files
- `/deployments/` - Docker and deployment configurations

## Architecture
Clean architecture with service-oriented design:
- **Service Layer**: ArticleService, CacheService, EmailService, SearchService, TemplateService
- **HTTP Layer**: Gin handlers with middleware for security, logging, CORS, rate limiting
- **Models**: Article, SearchResult, ContactMessage, Feed structures
- **Configuration**: Environment-based config with comprehensive settings

## Commands
- Build: `make build` or `go build ./cmd/server`
- Test: `make test` or `go test ./...`
- Lint: `make check` (includes fmt, vet, lint, test)
- Dev/Run: `make dev` (hot-reload) or `make run`

## Testing
- Uses `testify` framework with comprehensive test coverage (80%+)
- Table-driven tests and mock-based testing patterns
- HTTP handler testing with `httptest`
- Test files organized alongside source code (`*_test.go`)
- Coverage reporting and benchmark support
- Run tests with `go test ./...` or use Makefile targets

## Key Performance Characteristics
- **Response Time**: Sub-100ms average, <50ms 95th percentile
- **Memory Usage**: ~30MB base memory footprint
- **Binary Size**: ~29MB single executable
- **Concurrency**: Goroutine-based request handling
- **Caching**: Intelligent in-memory caching with TTL

## Production Features
- **Security**: Rate limiting, CORS, security headers, input validation
- **Monitoring**: Health checks, metrics endpoints, structured logging
- **Deployment**: Docker with multi-stage builds, systemd service files
- **Observability**: Built-in analytics, request monitoring, error tracking
- **SEO**: Automatic sitemap generation, structured data, RSS/JSON feeds

## Competitive Advantages
- **vs Hugo**: Dynamic features (search, contact forms, real-time updates)
- **vs Ghost**: 10x better performance, single binary deployment
- **vs Jekyll**: Native Go speed, no build step required
- **vs WordPress**: Developer-friendly, Git workflow, superior performance

## Release Status
- **Version**: 1.0.0 (ready for initial release)
- **License**: MIT (maximum adoption potential)
- **Tests**: 328+ tests passing, 80%+ coverage
- **Documentation**: Complete with README, guides, API docs
- **Personal Info**: Fully sanitized for public release

## Future Development Areas
- Plugin system for extensibility
- Multi-author workflows  
- Advanced analytics integration
- Theme marketplace
- Migration tools from other platforms
- API for headless CMS usage
- Performance optimizations
- Community-contributed themes and plugins

## Open Source Community Positioning
- **Target Audience**: Go developers, technical bloggers, performance-focused users
- **Value Proposition**: "Hugo's dynamic cousin" - same performance + server features
- **Differentiation**: File-based + dynamic features + developer workflow
- **Community Strategy**: Documentation-first, contribution-friendly, maintenance-focused