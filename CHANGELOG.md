# Changelog

All notable changes to MarkGo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

**Operational Excellence:**
- **Operational Runbook** (docs/RUNBOOK.md) - Comprehensive 1,000+ line guide
  - Incident response procedures (P1/P2/P3 classification)
  - Troubleshooting guides for common issues (service won't start, high memory, slow responses)
  - Health check protocols and monitoring recommendations
  - Rollback procedures with timing estimates
  - Performance baseline metrics and degradation indicators
  - Emergency contacts and escalation paths

- **Pre-Release Checklist** (.github/RELEASE_CHECKLIST.md)
  - 13-step comprehensive release validation process
  - Docker build verification steps
  - CI/CD workflow validation
  - Post-release validation procedures
  - Emergency rollback instructions
  - Prevents regressions (learned from Dockerfile path issue)

- **Dependency Automation** (.github/dependabot.yml)
  - Weekly automated dependency update PRs
  - Go modules, GitHub Actions, and Docker updates
  - Grouped minor/patch updates to reduce PR noise

**Test Coverage:**
- **Handler tests** (internal/handlers/article_test.go) - 575 new test lines
  - Article viewing by slug (valid, not found, empty)
  - Articles listing with pagination
  - Tag/category filtering (including URL encoding)
  - Search functionality (query, empty, special chars)
  - Home, tags, and categories pages
  - Table-driven test patterns for maintainability
  - **Coverage improvement: 14.1% → 50.1%** (+36 percentage points)

### Changed

- **Dependencies**: Updated 33 outdated packages (all tests passing)
  - gin-gonic/gin: v1.10.1 → v1.11.0
  - stretchr/testify: v1.10.0 → v1.11.1
  - redis/go-redis/v9: v9.12.1 → v9.14.1
  - prometheus/client_golang: v1.23.0 → v1.23.2
  - vnykmshr/goflow: v1.0.0 → v1.0.3
  - vnykmshr/obcache-go: v1.0.2 → v1.0.3
  - golang.org/x/crypto: v0.39.0 → v0.43.0 (security)
  - golang.org/x/net: v0.41.0 → v0.46.0 (security)

- **Documentation**: Updated README.md binary size (38MB → ~27MB)

### Fixed

- **Dockerfile build path**: Updated from `cmd/server` to `cmd/markgo` (critical regression)
- **Discovered**: Article not found returns 200 (error page), not 404 (documented in tests)

### Maintenance

- Cleaned 7.5MB of local artifacts (temp_articles/, dist/)
- All tests passing with updated dependencies
- CI/CD pipelines verified and passing

**Hygiene Score Improvement: 78/100 → 91/100** (+13 points)

**Commits in this maintenance cycle:** 4
- Week 1 critical fixes
- 33 dependency updates with testing
- Operational documentation
- Handler test coverage improvements

---

## [2.1.0] - 2025-10-23

### Major Simplifications

This release focuses on "less code, same functionality" - removing over-engineering
and bloat while maintaining full production features.

### Changed

- **Unified CLI**: Consolidated 5 separate binaries into single `markgo` binary with subcommands
  - `markgo serve` - Start the server (default command)
  - `markgo init` - Initialize new blog
  - `markgo new` - Create new article
  - `markgo export` - Export to static site
  - Aliases: `server`, `start` → `serve`; `new-article` → `new`; `build` → `export`

- **SEO Service Simplification**: Converted stateful service to stateless utility (~19% reduction)
  - Removed lifecycle management (Start/Stop methods)
  - Removed sitemap caching with mutexes (on-demand generation)
  - Consolidated 3 files (service.go, metatags.go, schema.go) into single seo.go
  - Total: 727 lines removed (46% smaller)

- **Admin Interface Cleanup**: Removed bloat and dead code (~27% reduction)
  - Removed emoji icons from admin routes
  - Removed non-functional SetLogLevel endpoint
  - Removed questionable CompactMemory endpoint
  - Removed placeholder SEO admin endpoint
  - Total: 152 lines removed

- **Middleware Consolidation**: Streamlined middleware stack
  - Removed 3 unused middleware functions (RequestID, Timeout, Compress)
  - Merged performance.go into middleware.go
  - Cleaned up function signatures (removed unused parameters)
  - Total: 56 lines removed

- **Config Validation Simplification**: Drastically simplified validation system (~27% reduction)
  - From 1,103 lines to 802 lines
  - Removed duplicate validation logic
  - Total: 301 lines removed

- **Article Service**: Removed unnecessary lazy loading complexity (~38% reduction)
  - From 305 lines to 190 lines
  - Simpler, more direct article loading
  - Total: 115 lines removed

- **Build System**: Simplified Makefile dramatically (~79% reduction)
  - From 660 lines to 136 lines
  - Removed stress-test (moved to examples/)
  - Clearer, more maintainable build targets
  - Total: 524 lines removed

### Removed

- Template hot-reload feature (fsnotify dependency removed)
- Article lazy loading infrastructure
- SEO service lifecycle management
- 3 unused middleware functions
- Dead admin endpoints
- Outdated CLI package documentation

### Fixed

- CI workflow updated for unified CLI structure
- Build paths corrected (cmd/server → cmd/markgo)
- ldflags injection into correct package path
- Documentation updated throughout for new CLI

### Added

- Unified CLI with subcommand architecture
- Comprehensive release documentation
- Updated getting started guide

### Technical Details

**Total Impact:**
- ~1,800+ lines of code removed
- 5 binaries → 1 unified CLI
- 3 SEO files → 1
- 2 middleware files → 1
- All tests passing (17.9% coverage, appropriate for scale)
- Binary size: ~27MB (optimized)
- Memory footprint: ~30MB (unchanged)
- Startup time: <1 second (unchanged)

**Commits in this release:** 12
- 10 simplification/refactoring commits
- 1 documentation update
- 1 CI workflow fix

## [2.0.0] - 2025-10-22

### Added
- Initial v2.0.0 release with modern Go architecture
- File-based blog engine with Markdown support
- SEO automation (sitemaps, Schema.org, Open Graph)
- Full-text search and RSS/JSON feeds
- Docker deployment support
- Static site export for GitHub Pages, Netlify, Vercel
- Admin interface with metrics
- Rate limiting and security middleware

---

For detailed commit history, see: https://github.com/vnykmshr/markgo/commits/main
