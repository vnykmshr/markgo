# Security Guide

This document outlines security considerations and best practices for deploying and operating MarkGo.

## 🔒 Environment Variables & Secrets Management

### Critical: Never Commit Secrets

**IMPORTANT**: The `.env` file is automatically excluded from version control via `.gitignore`. However, you must ensure:

1. **Never commit `.env` to Git**
   ```bash
   # Verify .env is ignored
   git check-ignore .env
   # Should output: .env
   ```

2. **Use `.env.example` as a template**
   - Copy `.env.example` to `.env` for local development
   - `.env.example` contains no secrets, only placeholders
   - Safe to commit `.env.example` to show required configuration

3. **Rotate secrets if accidentally committed**
   - If you accidentally commit secrets to Git:
     - Immediately rotate all exposed credentials (SMTP passwords, API keys, etc.)
     - Remove the commit from Git history (use `git filter-branch` or BFG Repo-Cleaner)
     - Never assume deletion removes the secret - anyone who pulled the repo has it

### Required Environment Variables with Security Impact

```bash
# SMTP Credentials (for contact form)
EMAIL_SMTP_PASSWORD="your-smtp-password-here"  # Keep secret!

# Admin Credentials (for /admin and /debug endpoints)
ADMIN_USERNAME="admin"           # Change from default
ADMIN_PASSWORD="secure-password" # Use strong password (16+ chars, mixed case, numbers, symbols)
```

### Deployment Security Checklist

- [ ] `.env` file is NOT in version control
- [ ] SMTP password is stored securely (use secrets manager in production)
- [ ] Admin password is strong and unique
- [ ] Admin credentials are different from default values
- [ ] Environment-specific `.env` files are used (dev, staging, prod)

---

## 🔐 Authentication & Authorization

### Session-Based Authentication

MarkGo uses **session-based authentication** for admin, compose, and draft endpoints:
- Login form at `/login` with session cookie
- `HttpOnly`, `SameSite=Strict` cookie attributes
- Sessions expire after inactivity

**Recommendations**:

1. **Always use HTTPS in production**
   ```nginx
   # Example nginx config
   server {
       listen 443 ssl;
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;

       location / {
           proxy_pass http://localhost:3000;
       }
   }
   ```

2. **Restrict admin access by IP** (recommended for production)
   ```nginx
   location /admin {
       allow 192.168.1.0/24;  # Your office network
       allow 10.0.0.1;         # Your home IP
       deny all;
       proxy_pass http://localhost:3000;
   }

   location /debug {
       allow 127.0.0.1;  # Localhost only
       deny all;
       proxy_pass http://localhost:3000;
   }
   ```

3. **Disable debug endpoints in production**
   - Debug endpoints (`/debug/*`) are **only enabled** when `ENVIRONMENT=development`
   - Production deployments automatically disable these endpoints
   - If you need debugging in production, use production-grade APM tools instead

### Endpoints Requiring Authentication

| Endpoint Pattern | Auth Required | Environment | Purpose |
|-----------------|---------------|-------------|---------|
| `/admin/*` | ✅ Session | All | Admin dashboard and management |
| `/compose/*` | ✅ Session + CSRF | All | Content creation and editing |
| `/debug/*` | ✅ Session | Development only | Runtime profiling and debugging |
| `/api/*` | ❌ None | All | Public API endpoints |
| `/writing/*` | ❌ None | All | Public content |

---

## 🛡️ Security Hardening Applied

### CORS Protection

**What we fixed**: Prevented localhost bypass vulnerability where `localhost.evil.com` could bypass CORS restrictions.

**Implementation**: Exact origin matching using map lookup:
```go
// Secure: Only explicitly allowed origins
allowedMap := make(map[string]bool)
for _, origin := range allowedOrigins {
    allowedMap[origin] = true
}

// Exact match required
if origin != "" && allowedMap[origin] {
    c.Header("Access-Control-Allow-Origin", origin)
}
```

**Configuration**:
```bash
# In .env - specify exact allowed origins
CORS_ALLOWED_ORIGINS="https://yourdomain.com,https://api.yourdomain.com"
```

### Rate Limiting

**Protection against**:
- Brute force attacks on contact form
- API abuse and spam
- Memory exhaustion attacks

**Implementation**:
- IP-based rate limiting with 10,000 IP cap (prevents unbounded memory growth)
- Uses `RemoteAddr` (not `X-Forwarded-For`) to prevent header spoofing
- Background cleanup every 1 minute removes stale entries
- Separate limits for general requests and contact form

**Configuration**:
```bash
# General rate limit
RATE_LIMIT_GENERAL_REQUESTS=100     # requests per window
RATE_LIMIT_GENERAL_WINDOW=1m        # 1 minute window

# Contact form rate limit (stricter)
RATE_LIMIT_CONTACT_REQUESTS=3       # requests per window
RATE_LIMIT_CONTACT_WINDOW=15m       # 15 minute window
```

**Note**: Math-based CAPTCHA is implemented on contact form for additional protection.

### Security Headers

Automatically applied to all responses:
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

### Input Validation

- Email addresses validated using RFC-compliant regex
- Markdown content sanitized with bluemonday
- File paths validated to prevent directory traversal
- All user input validated before processing

---

## Security Features (v3.1+)

### CSRF Protection

Double-submit cookie pattern on all compose routes:
- `SameSite=Strict`, `HttpOnly` cookie + hidden form field
- Constant-time comparison prevents timing attacks
- Token generation failure aborts with 500 (prevents empty-token bypass)
- `Secure` flag conditional on environment (HTTPS in production, HTTP in development)

### Session-Based Authentication

Cookie-based sessions replaced Basic Auth for admin, compose, and draft routes:
- Login form at `/login` with POST action
- Session cookie with `HttpOnly`, `SameSite=Strict`
- Middleware-enforced on all authenticated routes

### Slug Validation

Regex `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$` with length limits:
- Prevents CRLF injection in redirects
- Prevents filesystem traversal via URL params

### Image Upload Validation

- Content type detection via `http.DetectContentType` (not file extension)
- Atomic file writes (temp + rename) prevent partial content on disk

## Known Limitations

### 1. No Account Lockout

**Risk Level**: Low

**Mitigation**: Rate limiting on login endpoint prevents brute force. Deploy behind reverse proxy with IP whitelisting for additional protection.

### 2. Single Admin Account

**Risk Level**: Low

**Why**: Target audience is single-author blogs. Credentials in `.env`.

**Mitigation**: Use strong passwords. Consider oauth2-proxy for multi-user setups.

---

## 📋 Security Checklist for Production Deployment

### Before Deploying

- [ ] Change default admin credentials
- [ ] Use strong passwords (16+ characters, mixed case, numbers, symbols)
- [ ] Review and set `CORS_ALLOWED_ORIGINS` to exact allowed domains
- [ ] Configure rate limits appropriate for your traffic
- [ ] Ensure `.env` is in `.gitignore` and not committed
- [ ] Set `ENVIRONMENT=production` to disable debug endpoints

### Deployment Configuration

- [ ] Deploy behind reverse proxy (nginx, Caddy, etc.)
- [ ] Enable HTTPS/TLS with valid certificates
- [ ] Configure IP whitelisting for `/admin` endpoints
- [ ] Set up firewall rules to restrict admin access
- [ ] Use secrets manager for sensitive config (AWS Secrets Manager, HashiCorp Vault, etc.)

### Post-Deployment

- [ ] Test that `/debug` endpoints are disabled (should return 404)
- [ ] Verify CORS headers are correctly applied
- [ ] Test rate limiting with multiple requests
- [ ] Confirm admin login works with strong password
- [ ] Monitor logs for suspicious activity
- [ ] Set up automated security scanning (Dependabot, Snyk, etc.)

### Monitoring

- [ ] Enable structured logging in production
- [ ] Monitor failed admin login attempts
- [ ] Track rate limit violations
- [ ] Alert on unusual traffic patterns
- [ ] Regular dependency updates for security patches

---

## 🔍 Security Auditing

### Dependency Scanning

```bash
# Check for known vulnerabilities
go list -json -deps ./... | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy
```

### Code Scanning

```bash
# Run gosec security scanner
gosec ./...

# Run staticcheck
staticcheck ./...
```

### Testing

```bash
# Run all tests including security-related tests
go test ./... -v

# Run specific security tests
go test ./internal/middleware -v  # CORS and rate limiting
go test ./internal/handlers -v    # Contact form validation
```

---

## 📞 Reporting Security Issues

If you discover a security vulnerability in MarkGo:

1. **DO NOT** open a public GitHub issue
2. Email security concerns to the maintainer privately
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact assessment
   - Suggested fix (if available)

We take security seriously and will respond to legitimate reports promptly.

---

## 📚 Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Policy](https://golang.org/security)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Let's Encrypt (Free SSL/TLS)](https://letsencrypt.org/)

---

**Last Updated**: 2026-02-12
**Version**: 3.1.0
