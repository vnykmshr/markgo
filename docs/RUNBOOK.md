# MarkGo Operational Runbook

**Version:** 2.1.0
**Last Updated:** October 23, 2025
**Target Audience:** Operations, DevOps, SREs running MarkGo in production

This runbook provides procedures for common operational tasks, troubleshooting, and incident response for MarkGo deployments.

---

## Table of Contents

- [Quick Reference](#quick-reference)
- [Health Checks](#health-checks)
- [Common Operations](#common-operations)
- [Troubleshooting](#troubleshooting)
- [Performance Issues](#performance-issues)
- [Incident Response](#incident-response)
- [Rollback Procedures](#rollback-procedures)
- [Monitoring & Alerts](#monitoring--alerts)
- [Emergency Contacts](#emergency-contacts)

---

## Quick Reference

### Service Status

```bash
# Check if service is running
ps aux | grep markgo

# Check service logs (systemd)
journalctl -u markgo -f

# Check service logs (Docker)
docker logs -f markgo-app

# Health check endpoint
curl http://localhost:3000/health
```

### Critical Files

| File | Purpose | Location |
|------|---------|----------|
| `.env` | Configuration | Project root |
| `articles/` | Content | Project root or `/articles` in Docker |
| `logs/*.log` | Application logs | `./logs/` or `/var/log/markgo/` |
| `web/templates/` | HTML templates | `./web/templates/` |
| `web/static/` | Assets | `./web/static/` |

### Quick Actions

```bash
# Restart service (systemd)
sudo systemctl restart markgo

# Restart service (Docker)
docker-compose restart markgo

# Clear cache (requires restart)
# No action needed - cache is in-memory

# Reload articles (graceful)
curl -X POST http://localhost:3000/admin/reload

# Check version
./markgo --version
```

---

## Health Checks

### Application Health

**Endpoint:** `GET /health`

```bash
curl -i http://localhost:3000/health
```

**Expected Response:**
```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "healthy",
  "timestamp": "2025-10-23T01:00:00Z",
  "uptime": 3600,
  "version": "v2.1.0"
}
```

**Status Codes:**
- `200`: Service healthy
- `503`: Service degraded or unavailable

### Docker Health Check

Docker includes automatic healthchecks:

```bash
docker inspect --format='{{.State.Health.Status}}' markgo-app
```

**Expected:** `healthy`

### Process Health

```bash
# Check process is running
pgrep -f markgo

# Check port is listening
netstat -tlnp | grep :3000
# or
ss -tlnp | grep :3000
```

---

## Common Operations

### Deployment

#### Standard Deployment (Systemd)

```bash
# 1. Stop service
sudo systemctl stop markgo

# 2. Backup current binary
cp /usr/local/bin/markgo /usr/local/bin/markgo.backup

# 3. Deploy new binary
scp markgo-linux server:/usr/local/bin/markgo
chmod +x /usr/local/bin/markgo

# 4. Verify version
/usr/local/bin/markgo --version

# 5. Start service
sudo systemctl start markgo

# 6. Check logs
journalctl -u markgo -f

# 7. Health check
curl http://localhost:3000/health
```

#### Docker Deployment

```bash
# 1. Pull new image or rebuild
docker-compose pull  # if using registry
# or
docker-compose build  # if building locally

# 2. Stop and remove old container
docker-compose down

# 3. Start new container
docker-compose up -d

# 4. Check logs
docker-compose logs -f markgo

# 5. Health check
curl http://localhost:8080/health
```

### Configuration Changes

```bash
# 1. Edit .env file
vim .env

# 2. Validate configuration (optional)
./markgo serve --validate-config

# 3. Restart service
sudo systemctl restart markgo
# or
docker-compose restart markgo

# 4. Verify changes took effect
curl http://localhost:3000/admin/config  # if admin enabled
```

### Adding/Updating Content

```bash
# 1. Add new article
vim articles/2025-10-23-new-post.md

# 2. Reload articles (no restart needed)
curl -X POST http://localhost:3000/admin/reload

# 3. Verify article appears
curl http://localhost:3000/api/articles | jq '.articles[] | select(.slug=="new-post")'
```

### Certificate Renewal (HTTPS)

If using reverse proxy (nginx/caddy):

```bash
# Let's Encrypt with Certbot
certbot renew

# Reload nginx
nginx -s reload

# No MarkGo restart needed
```

---

## Troubleshooting

### Issue: Service Won't Start

**Symptoms:** `systemctl start markgo` fails or exits immediately

**Diagnosis:**

```bash
# Check logs
journalctl -u markgo -n 50

# Try manual start to see errors
/usr/local/bin/markgo serve

# Check port availability
netstat -tlnp | grep :3000
```

**Common Causes:**

1. **Port already in use**
   ```bash
   # Find process using port
   lsof -i :3000

   # Kill conflicting process
   kill -9 <PID>

   # Or change port in .env
   echo "PORT=3001" >> .env
   ```

2. **Invalid configuration**
   ```bash
   # Validate .env
   ./markgo serve --validate-config

   # Check for typos, missing required vars
   cat .env | grep -E "BLOG_TITLE|PORT|BASE_URL"
   ```

3. **Missing dependencies (static files)**
   ```bash
   # Verify directories exist
   ls -la web/templates/
   ls -la web/static/

   # Restore from backup if missing
   ```

4. **Permissions issues**
   ```bash
   # Check file ownership
   ls -l /usr/local/bin/markgo

   # Fix ownership
   sudo chown root:root /usr/local/bin/markgo

   # Check execute permission
   chmod +x /usr/local/bin/markgo
   ```

### Issue: High Memory Usage

**Symptoms:** Memory usage > 100MB, OOM kills

**Diagnosis:**

```bash
# Check memory usage
ps aux | grep markgo | awk '{print $6}'  # RSS in KB

# Docker memory stats
docker stats markgo-app
```

**Actions:**

1. **Check article count**
   ```bash
   ls articles/*.md | wc -l
   ```

   Expected: MarkGo handles 100-500 articles in ~30MB

   If > 1000 articles, consider pagination or splitting

2. **Check for memory leaks**
   ```bash
   # Enable pprof (if not in production)
   curl http://localhost:3000/debug/pprof/heap > heap.prof

   # Analyze with go tool
   go tool pprof heap.prof
   ```

3. **Restart service (temporary fix)**
   ```bash
   sudo systemctl restart markgo
   ```

4. **Increase container limits (Docker)**
   ```yaml
   # docker-compose.yml
   services:
     markgo:
       mem_limit: 128m
       memswap_limit: 128m
   ```

### Issue: Slow Response Times

**Symptoms:** Response time > 100ms, timeouts

**Diagnosis:**

```bash
# Check response time
time curl -o /dev/null -s -w '%{time_total}\n' http://localhost:3000/

# Check server logs for slow requests
grep "Slow request" logs/markgo.log

# Check system load
uptime
top
```

**Common Causes:**

1. **Cache disabled**
   ```bash
   # Check .env
   grep CACHE_ENABLED .env

   # Should be: CACHE_ENABLED=true
   ```

2. **Disk I/O (first request after restart)**
   ```bash
   # Normal: first request loads articles from disk
   # Subsequent requests use cache

   # Warm up cache after restart
   curl http://localhost:3000/api/articles >/dev/null
   ```

3. **High CPU usage**
   ```bash
   top -p $(pgrep markgo)

   # If CPU > 80%, check for:
   # - Too many concurrent requests (add rate limiting)
   # - Search queries on large content
   ```

4. **Network issues**
   ```bash
   # Test from server directly
   curl http://localhost:3000/health

   # vs from external
   curl http://yourdomain.com/health

   # Check reverse proxy (nginx/caddy) if external is slow
   ```

### Issue: Articles Not Appearing

**Symptoms:** New articles don't show up on site

**Diagnosis:**

```bash
# Check file exists
ls -la articles/your-article.md

# Check frontmatter is valid
head -20 articles/your-article.md

# Check logs for parsing errors
grep -i "error\|failed" logs/markgo.log | grep articles
```

**Actions:**

1. **Validate frontmatter**
   ```yaml
   # Required fields:
   ---
   title: "Article Title"
   date: 2025-10-23T10:00:00Z
   published: true
   ---
   ```

2. **Check draft status**
   ```yaml
   # published: false means article won't appear
   published: true  # Must be true
   ```

3. **Reload articles**
   ```bash
   curl -X POST http://localhost:3000/admin/reload
   ```

4. **Check file permissions**
   ```bash
   # MarkGo process must be able to read files
   chmod 644 articles/*.md
   ```

5. **Restart if reload doesn't work**
   ```bash
   sudo systemctl restart markgo
   ```

### Issue: Search Not Working

**Symptoms:** Search returns no results or errors

**Diagnosis:**

```bash
# Check if search is enabled
grep SEARCH_ENABLED .env

# Test search endpoint
curl "http://localhost:3000/api/search?q=test"
```

**Actions:**

1. **Enable search**
   ```bash
   echo "SEARCH_ENABLED=true" >> .env
   sudo systemctl restart markgo
   ```

2. **Check article content**
   ```bash
   # Search requires article content, not just titles
   # Verify articles have body content
   ```

3. **Try different queries**
   ```bash
   # Search is case-insensitive
   curl "http://localhost:3000/api/search?q=golang"
   curl "http://localhost:3000/api/search?q=blog"
   ```

---

## Performance Issues

### Baseline Performance

**Expected metrics (100 articles, single instance):**
- Cold start: < 1 second
- Memory: ~30MB RSS
- Response time (cached): 2-5ms
- Response time (first): 10-50ms
- Concurrent requests: 1000+ req/s

### Performance Degradation Indicators

Monitor these metrics:

1. **Response time > 100ms** consistently
   - Action: Check system resources, enable caching

2. **Memory > 100MB** for < 500 articles
   - Action: Check for memory leak, restart service

3. **CPU > 50%** at idle
   - Action: Check for runaway goroutines, review logs

4. **Error rate > 1%**
   - Action: Check logs, review recent changes

### Performance Optimization

```bash
# 1. Enable caching (if not already)
CACHE_ENABLED=true
CACHE_TTL=3600

# 2. Adjust rate limiting (if too restrictive)
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60

# 3. Use production mode
ENVIRONMENT=production
GIN_MODE=release

# 4. Consider reverse proxy caching
# Add nginx/caddy with caching for static assets
```

---

## Incident Response

### P1: Complete Service Outage

**Definition:** Service unreachable, 100% error rate

**Immediate Actions (within 5 minutes):**

1. **Check service status**
   ```bash
   sudo systemctl status markgo
   docker-compose ps
   ```

2. **Attempt restart**
   ```bash
   sudo systemctl restart markgo
   # or
   docker-compose restart markgo
   ```

3. **Check health**
   ```bash
   curl http://localhost:3000/health
   ```

4. **If restart fails, rollback** (see [Rollback Procedures](#rollback-procedures))

5. **Notify stakeholders**
   - Update status page
   - Post incident channel message

**Investigation (within 15 minutes):**

```bash
# Check logs for root cause
journalctl -u markgo --since "10 minutes ago" | grep -i error

# Check system resources
free -h
df -h
top
```

### P2: Partial Degradation

**Definition:** Service slow, elevated error rate (5-50%)

**Actions:**

1. **Identify affected functionality**
   ```bash
   # Test endpoints
   curl http://localhost:3000/health
   curl http://localhost:3000/api/articles
   curl http://localhost:3000/
   ```

2. **Check resources**
   ```bash
   # CPU, memory, disk
   top
   free -h
   df -h
   ```

3. **Review recent changes**
   ```bash
   git log --oneline -10
   ```

4. **Consider rollback if recent deployment**

### P3: Minor Issues

**Definition:** Localized issues, < 5% impact

**Actions:**

1. **Document issue**
2. **Create tracking ticket**
3. **Schedule fix during maintenance window**
4. **Implement workaround if possible**

---

## Rollback Procedures

### Quick Rollback (Systemd)

```bash
# 1. Stop service
sudo systemctl stop markgo

# 2. Restore previous binary
cp /usr/local/bin/markgo.backup /usr/local/bin/markgo

# 3. Start service
sudo systemctl start markgo

# 4. Verify
curl http://localhost:3000/health

# 5. Check version
/usr/local/bin/markgo --version
```

**Time to rollback:** ~2 minutes

### Docker Rollback

```bash
# 1. Check previous image tag
docker images markgo

# 2. Update docker-compose.yml to previous tag
vim docker-compose.yml
# Change: image: markgo:v2.1.0 -> image: markgo:v2.0.0

# 3. Restart with old image
docker-compose down
docker-compose up -d

# 4. Verify
curl http://localhost:8080/health
```

**Time to rollback:** ~3 minutes

### Configuration Rollback

```bash
# 1. Restore .env from backup
cp .env.backup .env

# 2. Restart service
sudo systemctl restart markgo

# 3. Verify
curl http://localhost:3000/health
```

### Content Rollback

```bash
# If using Git for articles/
cd articles/
git log --oneline -5
git revert <commit-hash>

# Reload articles
curl -X POST http://localhost:3000/admin/reload
```

---

## Monitoring & Alerts

### Prometheus Metrics

MarkGo exposes Prometheus metrics (if enabled):

```bash
# Scrape metrics
curl http://localhost:3000/metrics
```

**Key Metrics:**

- `markgo_requests_total` - Total HTTP requests
- `markgo_request_duration_seconds` - Request latency
- `markgo_articles_total` - Number of articles loaded
- `markgo_cache_hits_total` - Cache hit rate
- `go_memstats_alloc_bytes` - Memory usage
- `go_goroutines` - Number of goroutines

### Recommended Alerts

**Critical:**
```yaml
# Service Down
- alert: MarkGoDown
  expr: up{job="markgo"} == 0
  for: 1m

# High Error Rate
- alert: MarkGoHighErrorRate
  expr: rate(markgo_requests_total{status=~"5.."}[5m]) > 0.05
  for: 5m
```

**Warning:**
```yaml
# High Response Time
- alert: MarkGoSlowRequests
  expr: histogram_quantile(0.95, markgo_request_duration_seconds) > 0.1
  for: 10m

# High Memory
- alert: MarkGoHighMemory
  expr: go_memstats_alloc_bytes > 100000000
  for: 15m
```

### Log Monitoring

```bash
# Monitor for errors in real-time
tail -f logs/markgo.log | grep -i error

# Count errors in last hour
grep -i error logs/markgo.log | grep "$(date -u +'%Y-%m-%dT%H')" | wc -l

# Find most common errors
grep -i error logs/markgo.log | awk '{print $NF}' | sort | uniq -c | sort -rn | head
```

---

## Emergency Contacts

### Escalation Path

1. **L1: On-call Engineer**
   - Check runbook
   - Attempt restart/rollback
   - Escalate if not resolved in 15 minutes

2. **L2: Senior Engineer / Project Maintainer**
   - Deep troubleshooting
   - Code-level investigation
   - Escalate if architectural issue

3. **L3: Project Lead**
   - Strategic decisions
   - Major architectural changes

### Contact Information

```
Primary Maintainer:
  GitHub: @vnykmshr
  Issues: https://github.com/vnykmshr/markgo/issues

Community:
  Discussions: https://github.com/vnykmshr/markgo/discussions
```

---

## Appendix

### Useful Commands

```bash
# Get process info
ps -p $(pgrep markgo) -o pid,ppid,cmd,%mem,%cpu,etime

# Monitor in real-time
watch -n 2 'curl -s http://localhost:3000/health | jq'

# Test concurrent load
ab -n 1000 -c 10 http://localhost:3000/

# Check open files
lsof -p $(pgrep markgo)

# Network connections
netstat -tnp | grep markgo
```

### Directory Structure Reference

```
Production Deployment:
/usr/local/bin/markgo           # Binary
/etc/markgo/.env                # Configuration
/var/lib/markgo/articles/       # Content
/var/lib/markgo/web/            # Templates/static
/var/log/markgo/                # Logs
/etc/systemd/system/markgo.service  # Systemd unit
```

### Environment Variables Quick Reference

| Variable | Default | Purpose |
|----------|---------|---------|
| `ENVIRONMENT` | development | Runtime environment |
| `PORT` | 3000 | HTTP port |
| `BASE_URL` | - | Public URL (required) |
| `CACHE_ENABLED` | false | Enable in-memory cache |
| `CACHE_TTL` | 3600 | Cache TTL in seconds |
| `SEARCH_ENABLED` | false | Enable full-text search |
| `RSS_ENABLED` | true | Enable RSS feed |

See [Configuration Guide](configuration.md) for complete list.

---

**Document Version:** 1.0
**Next Review:** 2026-01-23
**Feedback:** https://github.com/vnykmshr/markgo/issues
