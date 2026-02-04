# MarkGo Configuration Guide

This guide covers all configuration options available in MarkGo. Configuration is managed through environment variables, typically set in a `.env` file.

## Quick Start

1. Copy the example configuration:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your settings:
   ```bash
   # Minimum required changes for local development
   BLOG_TITLE="My Awesome Blog"
   BLOG_AUTHOR="Your Name"
   BLOG_AUTHOR_EMAIL="your@email.com"
   ```

3. For production deployment, see the [Production Configuration](#production-configuration) section.

## Configuration Sections

### Core Configuration

#### `ENVIRONMENT`
**Default:** `development`  
**Values:** `development`, `production`, `test`

Sets the runtime environment. This affects default values for other settings and enables/disables certain features.

**Environment Effects:**
- **Production:** Enables optimizations, sets `GIN_MODE=release`, disables debug endpoints
- **Development:** Enables debug endpoints, template hot-reload, detailed logging, sets `GIN_MODE=debug`
- **Test:** Minimal output, sets `GIN_MODE=test`, optimized for testing

#### `PORT`
**Default:** `3000`
**Range:** `1-65535`

The port number the server will listen on.

#### `BASE_URL`
**Default:** `http://localhost:3000`
**Format:** Full URL with protocol

The base URL of your blog. Used for:
- RSS feed URLs
- Sitemap generation
- Social media metadata (OpenGraph, Twitter Cards)
- Canonical URLs

**Production Example:** `https://yourdomain.com`

### Paths & Directories

#### `ARTICLES_PATH`
**Default:** `./articles`

Path to the directory containing your blog articles (Markdown files).

#### `STATIC_PATH`
**Default:** `./web/static`

Path to static assets (CSS, JavaScript, images).

#### `TEMPLATES_PATH`
**Default:** `./web/templates`

Path to HTML templates directory.

### Blog Information

#### `BLOG_TITLE`
**Default:** `Your Blog Title`

The title of your blog, displayed in the header and used in metadata.

#### `BLOG_TAGLINE`
**Default:** *(empty — falls back to `BLOG_DESCRIPTION`)*

A short tagline displayed below the blog title in the navbar. Keep it concise (3-5 words). If not set, the full description is used instead.

#### `BLOG_DESCRIPTION`
**Default:** `Your blog description goes here`

A brief description of your blog, used for SEO metadata, RSS feeds, footer, and hero sections.

#### `BLOG_AUTHOR`
**Default:** `Your Name`

The primary author name for your blog.

#### `BLOG_AUTHOR_EMAIL`
**Default:** `your.email@example.com`

Author's email address, used in RSS feeds and metadata.

#### `BLOG_LANGUAGE`
**Default:** `en`  
**Format:** ISO 639-1 language code

The primary language of your blog content. Examples: `en`, `es`, `fr`, `en-US`.

#### `BLOG_THEME`
**Default:** `default`  
**Values:** `default`, `dark`, `custom`

Theme selection for your blog's appearance.

#### `BLOG_POSTS_PER_PAGE`
**Default:** `10`  
**Range:** `1-100`

Number of articles to display per page in article listings.

### Cache Configuration

MarkGo uses an intelligent caching system to improve performance.

#### `CACHE_TTL`
**Default:** `3600` (1 hour)  
**Unit:** Seconds

How long cached items remain valid. Recommended values:
- Development: `300` (5 minutes)
- Production: `3600-86400` (1-24 hours)

#### `CACHE_MAX_SIZE`
**Default:** `1000`

Maximum number of items to keep in cache. Higher values use more memory but improve cache hit rates.

#### `CACHE_CLEANUP_INTERVAL`
**Default:** `600` (10 minutes)  
**Unit:** Seconds

How often to clean expired items from cache.

### Email Configuration

Used for the contact form functionality.

#### `EMAIL_HOST`
**Default:** (empty)

SMTP server hostname. Leave empty to disable email functionality.

**Common values:**
- Gmail: `smtp.gmail.com`
- Outlook: `smtp-mail.outlook.com`
- Custom SMTP server: `mail.yourdomain.com`

#### `EMAIL_PORT`
**Default:** `587`

SMTP server port.

**Common values:**
- `25` - Unencrypted (not recommended)
- `587` - STARTTLS (recommended)
- `465` - SSL/TLS

#### `EMAIL_USERNAME`
**Default:** (empty)

SMTP authentication username.

#### `EMAIL_PASSWORD`
**Default:** (empty)

SMTP authentication password.

#### `EMAIL_FROM`
**Default:** `noreply@yourdomain.com`

The "From" email address for outgoing messages.

#### `EMAIL_TO`
**Default:** `your.email@example.com`

Where contact form submissions will be sent.

#### `EMAIL_USE_SSL`
**Default:** `true`

Enable SSL/TLS encryption for email connections.

### Rate Limiting & Security

#### General Rate Limiting

**`RATE_LIMIT_GENERAL_REQUESTS`**  
**Default:** `100`

Number of requests allowed per time window for general API endpoints.

**`RATE_LIMIT_GENERAL_WINDOW`**  
**Default:** `900s` (15 minutes)

Time window for general rate limiting.

#### Contact Form Rate Limiting

**`RATE_LIMIT_CONTACT_REQUESTS`**  
**Default:** `5`

Number of contact form submissions allowed per time window.

**`RATE_LIMIT_CONTACT_WINDOW`**  
**Default:** `3600s` (1 hour)

Time window for contact form rate limiting.

#### CORS Configuration

**`CORS_ALLOWED_ORIGINS`**  
**Default:** `http://localhost:3000,http://localhost:8080`

Comma-separated list of allowed origins for CORS requests.

**Production example:** `https://yourdomain.com,https://www.yourdomain.com`

**`CORS_ALLOWED_METHODS`**  
**Default:** `GET,POST,PUT,DELETE,OPTIONS`

**`CORS_ALLOWED_HEADERS`**  
**Default:** `Origin,Content-Type,Accept,Authorization`

### Admin Interface

#### `ADMIN_USERNAME`
**Default:** (empty)

Admin panel username. Leave empty to disable admin interface.

#### `ADMIN_PASSWORD`
**Default:** (empty)

Admin panel password.

**Security Notes:**
- Use a strong, unique password for production
- Consider disabling admin interface if not needed
- Avoid common usernames like `admin`, `root`, `administrator`

### Server Timeouts

#### `SERVER_READ_TIMEOUT`
**Default:** `15s`

Maximum time to read the entire request, including body.

#### `SERVER_WRITE_TIMEOUT`
**Default:** `15s`

Maximum time before timing out writes of the response.

#### `SERVER_IDLE_TIMEOUT`
**Default:** `60s`

Maximum time to wait for the next request when keep-alives are enabled.

### Logging Configuration

#### `LOG_LEVEL`
**Default:** `info`  
**Values:** `debug`, `info`, `warn`, `error`

Minimum log level to output. Lower levels include higher levels.

#### `LOG_FORMAT`
**Default:** `json`  
**Values:** `json`, `text`

Log output format. JSON is better for production log analysis.

#### `LOG_OUTPUT`
**Default:** `stdout`  
**Values:** `stdout`, `stderr`, `file`

Where to send log output.

#### `LOG_FILE`
**Default:** (empty)

Log file path. Required when `LOG_OUTPUT=file`.

#### Log Rotation Settings

**`LOG_MAX_SIZE`** - Maximum log file size in MB before rotation (default: `100`)  
**`LOG_MAX_BACKUPS`** - Number of rotated log files to keep (default: `3`)  
**`LOG_MAX_AGE`** - Maximum age in days to keep old log files (default: `28`)  
**`LOG_COMPRESS`** - Compress rotated log files (default: `true`)

#### Advanced Logging

**`LOG_ADD_SOURCE`** - Add source file and line numbers (default: `false`)  
**`LOG_TIME_FORMAT`** - Custom time format for text logs (default: ISO 8601)

### Comments System

#### `COMMENTS_ENABLED`
**Default:** `false`

Enable/disable comments functionality.

#### `COMMENTS_PROVIDER`
**Default:** `giscus`  
**Values:** `giscus`, `disqus`, `utterances`

Comments system provider.

#### Giscus Configuration

**`GISCUS_REPO`** - GitHub repository in format `owner/repo`  
**`GISCUS_REPO_ID`** - GitHub repository ID  
**`GISCUS_CATEGORY`** - Discussion category (default: `General`)  
**`GISCUS_CATEGORY_ID`** - Discussion category ID  
**`GISCUS_THEME`** - Theme preference (default: `preferred_color_scheme`)  
**`GISCUS_LANGUAGE`** - Language code (default: `en`)  
**`GISCUS_REACTIONS_ENABLED`** - Enable emoji reactions (default: `true`)

### Analytics Integration

#### `ANALYTICS_ENABLED`
**Default:** `false`

Enable analytics tracking.

#### `ANALYTICS_PROVIDER`
**Default:** (empty)  
**Values:** `google`, `plausible`, `custom`

Analytics service provider.

#### Provider-Specific Settings

**Google Analytics:**
- `ANALYTICS_TRACKING_ID` - Google Analytics tracking ID (e.g., `GA_MEASUREMENT_ID`)

**Plausible Analytics:**
- `ANALYTICS_DOMAIN` - Your domain for Plausible

**Custom Analytics:**
- `ANALYTICS_DATA_API` - Custom API endpoint
- `ANALYTICS_CUSTOM_CODE` - Custom tracking code

### Preview Service Configuration

#### `PREVIEW_ENABLED`
**Default:** `false`

Enable/disable the live preview service for drafts. When enabled, allows real-time preview of articles during editing.

#### `PREVIEW_PORT`
**Default:** `8081`
**Range:** `1-65535`

Port for the preview service WebSocket connections.

#### `PREVIEW_BASE_URL`
**Default:** (auto-generated from BASE_URL)
**Format:** Full URL with protocol

Base URL for preview sessions. If not specified, automatically derived from BASE_URL.

#### `PREVIEW_MAX_SESSIONS`
**Default:** `10`
**Range:** `1-100`

Maximum number of concurrent preview sessions allowed.

#### `PREVIEW_SESSION_TIMEOUT`
**Default:** `30m`
**Format:** Duration (e.g., `30m`, `1h`, `2h30m`)

How long preview sessions remain active without activity before cleanup.

### Development Settings

#### `TEMPLATE_HOT_RELOAD`
**Default:** `true` (development only)

Automatically reload templates when they change during development.

#### `DEBUG`
**Default:** `false`

Add debug information to responses and enable additional logging.

## Production Configuration

### Essential Production Settings

```bash
# Core settings
ENVIRONMENT=production
BASE_URL=https://yourdomain.com

# Security
ADMIN_USERNAME=secure_admin_name
ADMIN_PASSWORD=very_secure_random_password

# Performance
CACHE_TTL=86400
LOG_LEVEL=warn
LOG_OUTPUT=file
LOG_FILE=/var/log/markgo/app.log

# CORS (be specific about origins)
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com

# Email (if using contact form)
EMAIL_HOST=smtp.yourdomain.com
EMAIL_USERNAME=noreply@yourdomain.com
EMAIL_PASSWORD=smtp_password
EMAIL_FROM=noreply@yourdomain.com
EMAIL_TO=contact@yourdomain.com
```

### Security Checklist

- [ ] Use HTTPS in `BASE_URL`
- [ ] Set strong `ADMIN_PASSWORD` or disable admin interface
- [ ] Configure specific `CORS_ALLOWED_ORIGINS` (not wildcard)
- [ ] Set appropriate rate limits
- [ ] Use `warn` or `error` log level
- [ ] Configure log rotation for file output
- [ ] Validate all placeholder values are updated

### Performance Optimization

- Increase `CACHE_TTL` to 24 hours (`86400`)
- Set `CACHE_MAX_SIZE` based on your content volume
- Use file-based logging with rotation
- Monitor and adjust rate limits based on traffic

### Monitoring

Enable logging and consider:
- Log aggregation service (ELK stack, Fluentd)
- Application monitoring (New Relic, DataDog)
- Analytics integration for user behavior tracking

## Environment-Specific Examples

### Development
```bash
ENVIRONMENT=development
PORT=3000
BASE_URL=http://localhost:3000
CACHE_TTL=300
LOG_LEVEL=debug
TEMPLATE_HOT_RELOAD=true
DEBUG=true
```

### Staging
```bash
ENVIRONMENT=production
PORT=3000
BASE_URL=https://staging.yourdomain.com
CACHE_TTL=3600
LOG_LEVEL=info
LOG_OUTPUT=file
LOG_FILE=/var/log/markgo-staging/app.log
```

### Production
```bash
ENVIRONMENT=production
PORT=3000
BASE_URL=https://yourdomain.com
CACHE_TTL=86400
LOG_LEVEL=warn
LOG_OUTPUT=file
LOG_FILE=/var/log/markgo/app.log
ANALYTICS_ENABLED=true
ANALYTICS_PROVIDER=plausible
ANALYTICS_DOMAIN=yourdomain.com
```

## Validation

MarkGo includes comprehensive configuration validation that checks for:

- **Errors:** Invalid values that prevent startup
- **Warnings:** Suboptimal settings that may cause issues
- **Recommendations:** Suggestions for better security/performance

Run with `-validate` flag to see validation results:
```bash
./markgo -validate
```

## Troubleshooting

### Common Issues

**Server won't start:**
- Check port availability: `lsof -i :3000`
- Verify file permissions on paths
- Check configuration validation output

**Performance issues:**
- Increase `CACHE_TTL` and `CACHE_MAX_SIZE`
- Monitor log output for errors
- Check server timeout values

**Email not working:**
- Verify SMTP credentials and settings
- Check firewall rules for SMTP ports
- Test with a simple SMTP client

**Admin panel inaccessible:**
- Verify `ADMIN_USERNAME` and `ADMIN_PASSWORD` are set
- Check if admin routes are properly configured
- Review access logs for blocked requests

### Getting Help

- Check server logs for error messages
- Use debug mode in development: `DEBUG=true`
- Run configuration validation
- Review this documentation for correct syntax

## Migration Guide

### From v0.x to v1.0

Configuration changes in v1.0:
- Renamed `ARTICLES_PER_PAGE` to `BLOG_POSTS_PER_PAGE`
- Added prefix `BLOG_` to blog-related settings
- Replaced simple rate limiting with separate general/contact limits
- Added comprehensive logging configuration
- Introduced analytics and comments configuration

Update your `.env` file according to the new structure in `.env.example`.