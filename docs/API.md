# MarkGo HTTP API Reference

**Version:** 2.3.0

## Overview

MarkGo provides HTTP endpoints for browsing articles, search, RSS feeds, and administration. All endpoints return HTML unless otherwise specified.

## Base Configuration

- **Default Port:** 3000 (configurable via `PORT` env variable)
- **Base URL:** Set via `BASE_URL` env variable

## Public Endpoints

### Home & Articles

#### `GET /`
Homepage displaying recent articles.

**Response:** HTML page

---

#### `GET /articles`
List all published articles with pagination.

**Query Parameters:**
- `page` (optional): Page number (default: 1)

**Response:** HTML page with article list

---

#### `GET /articles/:slug`
View individual article by slug.

**Parameters:**
- `slug`: Article slug (from filename or frontmatter)

**Response:**
- `200 OK`: HTML article page
- `200 OK`: Error page (if article not found)

---

### Filtering

#### `GET /tags`
List all tags with article counts.

**Response:** HTML page with tag cloud

---

#### `GET /tags/:tag`
List articles with specific tag.

**Parameters:**
- `tag`: Tag name

**Response:** HTML page with filtered articles

---

#### `GET /categories`
List all categories with article counts.

**Response:** HTML page with categories

---

#### `GET /categories/:category`
List articles in specific category.

**Parameters:**
- `category`: Category name (URL-encoded for spaces)

**Response:** HTML page with filtered articles

---

### Search

#### `GET /search`
Full-text search articles.

**Query Parameters:**
- `q`: Search query string

**Response:** HTML page with search results

---

### Special Pages

#### `GET /about`
About page (if `about.md` exists in articles directory).

**Response:** HTML page

---

#### `POST /contact`
Contact form submission.

**Content-Type:** `application/x-www-form-urlencoded`

**Form Fields:**
- `name`: Sender name (required)
- `email`: Sender email (required)
- `message`: Message content (required)

**Response:**
- `200 OK`: Success message
- `400 Bad Request`: Validation errors
- `429 Too Many Requests`: Rate limit exceeded

**Rate Limit:** Configurable, default 1 request per minute per IP

---

## Feeds

#### `GET /feed.xml`
RSS 2.0 feed of recent articles.

**Response:** XML (application/rss+xml)

---

#### `GET /feed.json`
JSON feed of recent articles.

**Response:** JSON (application/json)

**Example Response:**
```json
{
  "version": "https://jsonfeed.org/version/1",
  "title": "Blog Title",
  "home_page_url": "https://example.com",
  "feed_url": "https://example.com/feed.json",
  "items": [
    {
      "id": "article-slug",
      "url": "https://example.com/articles/article-slug",
      "title": "Article Title",
      "content_html": "<p>Content...</p>",
      "date_published": "2025-10-23T00:00:00Z",
      "tags": ["tag1", "tag2"]
    }
  ]
}
```

---

## SEO & Crawlers

#### `GET /sitemap.xml`
XML sitemap for search engines.

**Response:** XML (application/xml)

---

#### `GET /robots.txt`
Robots.txt for web crawlers.

**Response:** Plain text

---

## Health & Monitoring

#### `GET /health`
Health check endpoint.

**Response:** JSON

**Example Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-23T10:00:00Z",
  "uptime": 3600,
  "version": "v2.1.0"
}
```

**Status Codes:**
- `200 OK`: Service healthy
- `503 Service Unavailable`: Service degraded

---

#### `GET /metrics`
Prometheus metrics endpoint.

**Response:** Plain text (Prometheus format)

**Metrics Include:**
- HTTP request counters
- Request duration histograms
- Go runtime metrics

---

## Admin Endpoints

Admin endpoints require Basic Authentication (configured via `ADMIN_USERNAME` and `ADMIN_PASSWORD`).

#### `GET /admin`
Admin dashboard.

**Authentication:** Basic Auth required

**Response:** HTML admin page with statistics

---

#### `GET /admin/stats`
Retrieve site statistics.

**Authentication:** Basic Auth required

**Response:** JSON

**Example Response:**
```json
{
  "articles": {
    "total": 42,
    "published": 40,
    "drafts": 2
  },
  "tags": 15,
  "categories": 8,
  "cache": {
    "enabled": true,
    "hits": 1000,
    "misses": 50
  }
}
```

---

#### `POST /admin/cache/clear`
Clear application cache.

**Authentication:** Basic Auth required

**Response:** JSON

**Example Response:**
```json
{
  "message": "Cache cleared successfully"
}
```

---

#### `POST /admin/articles/reload`
Reload articles from disk without restarting.

**Authentication:** Basic Auth required

**Response:** JSON

**Example Response:**
```json
{
  "message": "Articles reloaded successfully",
  "count": 42
}
```

---

## Debug Endpoints (Development Only)

Debug endpoints are only available when `ENVIRONMENT=development`. They require Basic Auth if admin credentials are configured.

**Available Endpoints:**
- `GET /debug/pprof/*`: Go pprof profiling endpoints
- See Go pprof documentation for details

---

## Static Assets

#### `GET /static/*`
Serve static files (CSS, JavaScript, images).

**Response:** File content with appropriate Content-Type

**Caching:** Long-term cache headers in production

---

## Error Responses

MarkGo typically returns HTML error pages for user-facing endpoints. Status codes:

- `200 OK`: Success (even for "article not found" - returns error page)
- `400 Bad Request`: Invalid request parameters
- `404 Not Found`: Route not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service unavailable

---

## Rate Limiting

Rate limiting is applied per IP address:

- **General requests:** Configurable (default: 10 req/s)
- **Contact form:** Configurable (default: 1 req/min)

Rate limit headers:
- `X-RateLimit-Limit`: Request limit
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Reset timestamp

---

## Security Headers

All responses include security headers:

- `Strict-Transport-Security`: HSTS for HTTPS
- `X-Frame-Options`: SAMEORIGIN
- `X-Content-Type-Options`: nosniff
- `X-XSS-Protection`: 1; mode=block
- `Content-Security-Policy`: Configured CSP

---

## CORS

CORS is enabled with configurable allowed origins (via `CORS_ALLOWED_ORIGINS`).

---

## Further Reading

- [Configuration Guide](configuration.md) - All configuration options
- [Getting Started](GETTING-STARTED.md) - Setup and usage
- [Architecture](architecture.md) - Technical architecture

---

**Last Updated:** February 2026 (v2.3.0)
