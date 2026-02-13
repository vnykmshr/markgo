# Changelog

All notable changes to MarkGo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [3.3.0] - 2026-02-13

Article asset uploads, reactive auth, brand identity, and security hardening across the board.

### Added

**Article Asset Uploads**
- Slug-scoped file uploads: images, PDFs, and other assets stored per-article
- Upload UI in compose page with save-first guard, multi-type support, and progress feedback
- Extension-based blocklist (19 executable/script/HTML types) with case-insensitive matching
- Filepath containment check prevents directory traversal attacks
- Random 4-byte hex suffix in filenames prevents collision
- Upload directory writability probe at startup with warning on failure
- Dynamic `maxFileSize` from server config (no hardcoded client-side limit)
- `Content-Disposition: inline` for images, `attachment` for all other uploaded types

**Reactive Authentication**
- Login without page reload — DOM swap preserves SPA state
- Draft recovery toast after login (unsaved work detected)
- Popover module with AbortController lifecycle cleanup
- Search + subscribe header popovers

**Admin & Drafts**
- Admin dashboard with clickable stat tiles and Edit action buttons
- `/admin/writing` page with direct edit links to published articles
- One-click publish from drafts list with fade animation
- Edit link on published article pages

**Compose Enhancements**
- Draft-first compose: two buttons (Save Draft / Publish) replace checkbox
- Full compose auto-save with localStorage failure warning
- Quick compose 401 auto-login flow

**Brand Identity**
- Bean Brace brand: logo, favicon, Open Graph images

### Changed

- `fallbackErrorHTML` extracted to `internal/errors.FallbackErrorHTML` (shared between handlers and recovery middleware)
- 3 missing templates added to `requiredTemplates` startup validation (admin_home, category, tag)
- Upload route registers even when directory creation fails (Gin static serves existing files)
- Frontend CSS deduplication and layout consolidation
- Active link management consolidated in navigation module (removed duplicate from router)
- Button hover polish: prevent blue-on-blue text

### Fixed

- CSRF sync after reactive auth DOM swap (popover destroy + fresh token fetch)
- Bare `catch {}` blocks replaced with proper error logging (drafts, compose-sheet)
- Double-invocation guard on draft card removal animation
- `og:image` default wired correctly with deterministic meta tag ordering
- Nav active states, header consistency, tag ordering, article footer

### Security

- Upload blocklist expanded to 19 types (scripts, HTML-renderable, Java archives)
- Upload filepath containment via `filepath.Abs` + prefix check
- Upload config validation: empty path rejected, max size enforced
- `X-Content-Type-Options: nosniff` on all uploaded files
- Popover AbortController cleanup prevents event listener leaks

### Testing

- 28 upload security tests: no-extension, case-insensitive blocking, all 19 blocked extensions
- HandleSubmit/HandleEdit: valid, missing-slug, empty-title, frontmatter-error, write-error paths
- refreshCSRFToken: success, cookie-missing, cookie-empty, invalid-format paths
- injectAuthState: authenticated/unauthenticated paths
- PublishDraft reload-failure path test
- AdminHome, Writing, formatDuration, HumansTxt handler tests
- UploadConfig boundary tests (exactly-at-limit, over-limit, empty-path)
- Coverage: 52% → 58.7%

### Removed

- 6 unused `logo-*.png` theme assets (~78KB)
- Dead Twitter meta comment from base.html
- Stale draft-publishing comment from article service
- Duplicate `updateActiveLinks` from SPA router

---

## [3.2.0] - 2026-02-13

Semantic HTML, SEO, admin redesign, and security hardening. Web standards compliance across the board.

### Added

**SEO & Structured Data**
- Canonical URLs on all public pages with conditional rendering (empty on 404/admin/compose)
- Visible breadcrumb navigation with JSON-LD BreadcrumbList schema
- JSON-LD structured data: BlogPosting (articles), CollectionPage (listings), WebSite (home)
- Open Graph and meta description on all page types

**Admin Dashboard**
- Redesigned admin dashboard consistent with blog UX (card-based layout)
- Admin popover in header with Dashboard, Drafts, Sign out
- Auth-aware UI across all pages (login/admin popover based on session state)
- Sample articles section with one-click creation

**Public Compose**
- Compose page accessible without authentication (login deferred to publish)
- Compose link in header nav when admin is configured
- FAB (floating action button) visible for all users when admin configured
- Dynamic CTA: "Publish" / "Save Draft" / "Update" based on draft checkbox and edit state
- Fetch-based form submit with 401/403 handling (toast + login popover trigger)

**PWA**
- Share target in web manifest for receiving shared content

### Changed

- Semantic HTML: single `<h1>` per page, proper heading hierarchy, `<section>`/`<article>` elements
- 404/offline page CSS class renamed from `error-page` to `error-content` (fixes layout collision with body class)
- SessionAware middleware generates CSRF tokens for unauthenticated visitors (enables login popover on all pages)
- SPA router syncs CSRF meta tag and hidden inputs after content swap (prevents SPA desync)
- Compose error responses shown via toast instead of DOM swap (preserves event listeners)

### Security

- CSRF cookie reuse validates token format (64-char hex) before accepting
- CSRF cookie max-age refreshed on reuse to prevent silent expiry
- `isValidCSRFToken()` rejects corrupted, truncated, or injected cookie values

### Removed

- Editorial and bold theme stubs (unused CSS)
- Stale CDN references in design docs

---

## [3.1.0] - 2026-02-12

MarkGo reimagined as a blogging companion app. SPA navigation, installable PWA, mobile-native UX, quick capture, offline compose. Single binary with embedded web assets — no filesystem setup required.

### Added

**SPA & Navigation**
- App shell router with instant content swaps (Turbo Drive pattern — fetch HTML, DOMParser, swap `<main>`)
- Prefetch on hover (65ms delay, 5-entry cache, 30s TTL)
- CSS-only progress bar with prefers-reduced-motion support
- History management with redirect detection and same-page guard

**PWA & Offline**
- Service Worker with 3-tier caching: precache (offline.html), stale-while-revalidate (static), network-first (HTML)
- Installable PWA with dynamic web manifest generated from config
- Offline compose queue via IndexedDB with auto-sync on reconnect
- Install prompt with visit count threshold and 30-day dismiss

**Mobile-Native UX**
- Bottom navigation (5-item tab bar, hidden on desktop)
- Full-screen search overlay (mobile) / centered modal (desktop, Cmd/Ctrl+K)
- Visual viewport handling for iOS keyboards
- Dynamic theme-color meta tag from computed background

**Quick Capture**
- Floating action button (FAB) with compose sheet overlay
- Quick publish API (POST /compose/quick → JSON response)
- Auto-save drafts to localStorage with 2s debounce and recovery
- Optimistic feed update (prepend card on publish)
- Keyboard shortcut: Cmd/Ctrl+N

**Compose & Editing**
- Article editing with CSRF protection and atomic file writes (temp+rename)
- Server-side markdown preview with live toggle
- Image upload with drag-and-drop and content type detection
- Draft management page with edit links
- Description and categories in compose form

**Accessibility**
- Skip-to-content link
- SPA route announcer (`aria-live="polite"`)
- Screen reader labels on all form inputs
- Color theme picker with `role="radiogroup"` semantics
- Focus management after content swaps

**Frontend Architecture**
- ES modules: app.js entry + 9 shell modules + 4 page modules (replaces 590-line IIFE)
- Self-hosted fonts and highlight.js (zero CDN dependencies)
- Design token system in CSS custom properties (all colors, spacing, typography)
- Theme popover: Light/Dark/Auto mode + 5 color presets
- Toast notification system

**About Page**
- Config-driven unified about page (avatar, tagline, bio, location, 5 social platforms)
- Bio from `articles/about.md` or `ABOUT_BIO` config
- Contact: mailto link (email only) or SMTP form
- `/contact` → 301 redirect to `/about#contact`

**Backend**
- Go embed: templates and static assets compiled into binary
- Filesystem-first fallback: override embedded assets with `STATIC_PATH`/`TEMPLATES_PATH`
- Session-based authentication replacing BasicAuth
- Typed error system replacing string comparison
- Feed service extracted from handlers
- 11 focused handler types (from 2 monolithic handlers)

**CLI**
- `markgo new --type thought` and `markgo new --type link` quick-post commands

### Changed

- `/articles` renamed to `/writing` throughout
- Content types (article/thought/link) inferred automatically from frontmatter
- Feed page replaces homepage with type-specific card templates
- All CSS converted to mobile-first (320px base → 481px → 769px → 1025px)
- Admin CSS converted from max-width to min-width breakpoints
- All `innerHTML` replaced with DOM API (DOMParser, createElementNS)
- Pagination centralized with page clamping
- SEO URLs: `/article/` → `/articles/` in sitemap and schema

### Removed

- **Export command** — `markgo export` and static site generation removed (MarkGo requires a backend)
- **Deploy workflow** — `.github/workflows/deploy.yml` deleted
- **Giscus comments** — Third-party comment system removed; replaced with "Reply by email" link
- **Handler-level cache** — obcache removed from handlers (article service cache remains)
- **CDN dependencies** — All fonts and highlight.js self-hosted
- **Analytics config** — Dead `ANALYTICS_*` fields removed
- **Composed facade** — Replaced with direct handler routing
- **~4,000 lines of legacy CSS** — Rebuilt on design tokens
- **~1,000 lines of test bloat** — Mocks now return canned data

### Fixed

- Content negotiation: `Accept: */*` correctly returns HTML
- CSRF token generation failure aborts 500 (prevents empty-token bypass)
- Slug validation with length limits on URL params
- Template parse errors include actual Go error detail
- `NewPagination` guards against division by zero
- Silent error swallowing replaced with structured logging
- Race condition in cache counter increments (atomic ops)
- Cache goroutine leak (cleanup via `stopCh` channel)

### Security

- CSRF double-submit cookie on all compose routes (SameSite=Strict, constant-time compare)
- Slug regex prevents CRLF injection in redirects
- Image upload validates content type via `http.DetectContentType`
- Atomic file writes prevent partial content on disk
- Rate limiting excludes static assets (prevents false positives)

**98 commits since v2.3.1**

---

## [2.3.1] - 2026-02-04

### Fixed

- **78 golangci-lint issues resolved**: httpNoBody (25), importShadow (14), hugeParam (8), paramTypeCombine (3), unnecessaryBlock (3), errcheck (3), gosec G602 (1), gocyclo (1), emptyStringTest (1), misspell (1), revive naming (1), sprintfQuotedString nolint (3)
- **Contact page title**: Now includes blog title suffix consistent with all other pages
- **Page header layout inconsistency**: Standardized all pages to full-width hero pattern (categories, tags, contact, about)
- **Page header spacing**: Added breathing space between hero sections and body content across all breakpoints
- **Navbar tagline overflow**: Removed fallback to long description; empty tagline renders nothing
- **Placeholder GitHub links**: Replaced `@yourusername` in contact and about templates with actual project URL

### Changed

- **Serve command refactored**: Extracted `setupServer` and `configureGinMode` to reduce `Run()` complexity (gocyclo fix)
- **Export command**: `ExportConfig` renamed to `Config` to avoid package stutter (`export.Config`)

---

## [2.3.0] - 2026-02-04

### Added

- **BLOG_TAGLINE config**: Concise navbar branding separate from full blog description
- **Mobile-first CSS architecture**: Complete design system with CSS custom properties, breakpoints at 481px/769px/1025px
- **Theme system**: Independent light/dark mode toggle and color presets (default, ocean, forest, sunset, berry)
- **FOUC prevention**: Inline script in `<head>` reads localStorage before render
- **Article repository tests** (`repository_test.go`): LoadAll, GetBySlug, GetByTag, GetByCategory, slug generation, reading time
- **Search service tests** (`search_test.go`): Basic search, scoring, filters, suggestions, stop words
- **Cache coordinator tests** (`cache_test.go`): CRUD, concurrent access with race detector, shutdown cleanup
- **Content processor tests** (`content_test.go`): Markdown processing, excerpts, duplicate titles, image/link extraction
- **CI lint step**: golangci-lint in CI with `only-new-issues: true`
- **CI coverage threshold**: Build fails if coverage drops below 45%

### Changed

- **Unified color palette**: Replaced amber accent and rainbow gradients with restrained monochrome system — accent is now a lighter variant of each theme's primary hue
- **About page sidebar**: 6 hardcoded gradient cards reduced to 1 accent (profile) + 5 neutral cards with borders
- **Social share buttons**: Platform-specific brand colors replaced with `var(--color-primary)` for consistency
- **Tech icons**: Individual brand colors standardized to `var(--color-primary-dark)`
- **Tag cloud**: Filled blue pills changed to outlined style matching article card tags
- **JS showMessage**: Hardcoded hex colors replaced with CSS custom property reads (with fallback values)
- **golangci-lint v1 to v2 migration**: Config schema updated to v2 format, action upgraded to v7 with golangci-lint v2.8.0
- **quic-go updated** to v0.57.0 (resolves CVE)
- **GitHub Actions pinned**: `softprops/action-gh-release` pinned to SHA
- **Documentation accuracy**: Fixed broken links, stale coverage numbers, nonexistent Makefile targets, wrong PORT defaults, and obsolete binary references

### Fixed

- **YAML injection in article creation**: `SanitizeForYAML()` escapes backslashes and newlines
- **Cache race condition**: Counter increments now use `atomic.AddInt64` instead of mutating under `RLock`
- **Cache goroutine leak**: Cleanup goroutine stops cleanly via `stopCh` channel
- **Navbar horizontal overflow**: Prevented on mobile viewports
- **Color theme and dark mode separation**: Independent `data-theme` and `data-color-theme` attributes
- **Small-phone spacing**: Restored via `min-width: 481px` breakpoint
- **CLI hardening**: `serve --help`/`--port`, hardened export flags, wired build info
- **CI ldflags**: Aligned with constants package

---

## [2.2.0] - 2025-10-24

### Major Changes

**Stress Testing Tool Graduation:**
- **WebStress** - Independent stress testing tool (graduated from examples/stress-test/)
  - Migrated to standalone repository: https://github.com/vnykmshr/webstress
  - Enhanced with clean architecture and comprehensive documentation
  - Generalized to work with any web application (not just MarkGo)
  - Added migration guide: examples/STRESS_TESTING.md

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

- **Documentation**: Comprehensive cleanup and simplification
  - Updated README.md binary size (38MB → ~27MB)
  - Moved status reports to historical documentation
  - Pruned outdated documentation
  - Simplified structure for better maintainability

### Removed

- **examples/stress-test/**: Graduated to independent WebStress project
  - 6 files removed (~50KB of code)
  - Replaced with migration guide (examples/STRESS_TESTING.md)
  - Full functionality preserved in https://github.com/vnykmshr/webstress

### Fixed

- **Dockerfile build path**: Updated from `cmd/server` to `cmd/markgo` (critical regression)
- **Discovered**: Article not found returns 200 (error page), not 404 (documented in tests)

### Maintenance

- Cleaned 7.5MB of local artifacts (temp_articles/, dist/)
- All tests passing with updated dependencies
- CI/CD pipelines verified and passing
- Documentation hygiene improvements

**Hygiene Score Improvement: 78/100 → 91/100** (+13 points)

**Commits in this release:** 8
- Week 1 critical hygiene fixes
- 33 dependency updates with comprehensive testing
- Operational documentation (runbook, checklist)
- Handler test coverage improvements
- Documentation cleanup and simplification
- WebStress graduation

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
  - Aliases: `server`, `start` → `serve`; `new-article` → `new`

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
- Admin interface with metrics
- Rate limiting and security middleware

---

For detailed commit history, see: https://github.com/vnykmshr/markgo/commits/main
