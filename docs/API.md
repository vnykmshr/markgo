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

## Preview API Endpoints

### Overview

The Preview API provides real-time preview capabilities for draft articles with WebSocket-based live reload functionality.

### Authentication

Preview endpoints use session-based authentication with auto-generated tokens.

### Base URL

All preview endpoints are prefixed with `/api/preview`

### Endpoints

#### Create Preview Session
`POST /api/preview/sessions`

Creates a new preview session for a draft article.

**Request Body:**
```json
{
  "article_slug": "example-article-slug"
}
```

**Response:**
```json
{
  "session_id": "session_abc123",
  "url": "http://localhost:3000/preview/session_abc123",
  "auth_token": "auth_token_here",
  "created_at": "2025-09-19T10:30:00Z"
}
```

**Status Codes:**
- `200` - Session created successfully
- `400` - Invalid request body or article slug
- `404` - Draft article not found
- `429` - Maximum sessions exceeded

#### List Active Sessions
`GET /api/preview/sessions`

Returns list of active preview sessions.

**Response:**
```json
{
  "sessions": [
    {
      "session_id": "session_abc123",
      "article_slug": "example-article",
      "url": "http://localhost:3000/preview/session_abc123",
      "created_at": "2025-09-19T10:30:00Z",
      "last_accessed": "2025-09-19T10:35:00Z",
      "client_count": 2
    }
  ],
  "stats": {
    "active_sessions": 1,
    "total_clients": 2,
    "files_watched": 1
  }
}
```

#### Delete Preview Session
`DELETE /api/preview/sessions/{sessionId}`

Terminates a preview session and closes all WebSocket connections.

**Response:**
```json
{
  "message": "Preview session deleted successfully"
}
```

**Status Codes:**
- `200` - Session deleted successfully
- `404` - Session not found

#### WebSocket Connection
`GET /api/preview/ws/{sessionId}`

Establishes WebSocket connection for real-time updates.

**WebSocket Messages:**

**Connection Confirmation:**
```json
{
  "type": "connected",
  "session_id": "session_abc123",
  "timestamp": 1695117000
}
```

**File Change Reload:**
```json
{
  "type": "reload",
  "data": {
    "article_slug": "example-article",
    "file_path": "/path/to/article.md",
    "reason": "file_changed"
  },
  "timestamp": 1695117000
}
```

#### Serve Preview Page
`GET /preview/{sessionId}`

Serves the preview page for the article associated with the session.

**Response:** HTML page with embedded WebSocket client for live reload.

### Error Responses

All API endpoints return errors in a consistent format:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": "Additional error context (optional)"
}
```

### Configuration

Preview service configuration options:

- `PREVIEW_ENABLED` - Enable/disable preview service
- `PREVIEW_PORT` - WebSocket port (default: 8081)
- `PREVIEW_MAX_SESSIONS` - Maximum concurrent sessions (default: 10)
- `PREVIEW_SESSION_TIMEOUT` - Session timeout (default: 30m)

## Performance Targets

| Metric | Target | Purpose |
|--------|--------|---------|
| Throughput | ≥1000 req/s | Competitive advantage vs Ghost |
| 95th Percentile | <50ms | 4x faster than Ghost ~200ms |
| Average Response | <30ms | Excellent user experience |
| Error Rate | <1% | Production reliability |
| Success Rate | >99% | Production readiness |

*This documentation is automatically generated from Go source code comments.*
