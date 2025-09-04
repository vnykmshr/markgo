# Deployment Guide

This directory contains deployment configurations for markgo using Docker Compose.

## Files Overview

- `docker-compose.yml` - Production deployment configuration
- `docker-compose.override.yml.example` - Development override template
- `Dockerfile` - Multi-stage Docker build
- `nginx/` - Nginx reverse proxy configuration
- `monitoring/` - Prometheus and Grafana configuration

## Quick Start

### Production Deployment

1. **Configure environment variables:**
   ```bash
   cp ../.env.example ../.env
   # Edit .env with your production settings
   ```

2. **Start all services:**
   ```bash
   docker compose up -d
   ```

3. **Access the application:**
   - Blog: http://localhost (via Nginx)
   - Direct app: http://localhost:3000
   - Grafana: http://localhost:3001 (admin/admin123)
   - Prometheus: http://localhost:9090

### Development Setup

1. **Create development override:**
   ```bash
   cp docker-compose.override.yml.example docker-compose.override.yml
   ```

2. **Start development environment:**
   ```bash
   docker compose up -d
   ```

   The development override provides:
   - Hot reloading with Air
   - Debug port (2345) for Delve
   - Volume mounting for live code changes
   - Excludes production services (nginx, monitoring)

## Services

### Core Services

**markgo** - Main application
- Port: 3000
- Health check: `/health`
- Volumes: articles, web assets, logs

**redis** - Caching layer
- Port: 6379
- Persistent data volume

### Production Services

**nginx** - Reverse proxy
- Ports: 80 (HTTP), 443 (HTTPS)
- Static file serving
- Rate limiting
- SSL termination

### Monitoring Stack

**prometheus** - Metrics collection
- Port: 9090
- Scrapes application metrics
- 200h data retention

**grafana** - Metrics visualization
- Port: 3001
- Default credentials: admin/admin123
- Pre-configured dashboards

## Configuration

### Environment Variables

Key environment variables (see `.env.example`):

```bash
# Application
ENVIRONMENT=production
PORT=3000
BASE_URL=https://yourdomain.com

# Paths
ARTICLES_PATH=./articles
STATIC_PATH=./web/static
TEMPLATES_PATH=./web/templates

# Email (for contact form)
EMAIL_HOST=smtp.gmail.com
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password

# Cache
CACHE_TTL=1h
CACHE_MAX_SIZE=1000

# Rate Limiting
RATE_LIMIT_GENERAL_REQUESTS=100
RATE_LIMIT_CONTACT_REQUESTS=5

# Blog Settings
BLOG_TITLE="Your Blog Title"
BLOG_AUTHOR="Your Name"
BLOG_AUTHOR_EMAIL="your@email.com"
```

### SSL Configuration

For production SSL:

1. **Place certificates in `nginx/ssl/`:**
   ```
   nginx/ssl/
   ├── yourdomain.com.crt
   ├── yourdomain.com.key
   └── dhparam.pem
   ```

2. **Update `nginx/conf.d/markgo.conf`:**
   ```nginx
   ssl_certificate /etc/nginx/ssl/yourdomain.com.crt;
   ssl_certificate_key /etc/nginx/ssl/yourdomain.com.key;
   ```

## Management Commands

### Application Management

```bash
# View logs
docker compose logs -f markgo

# Restart application
docker compose restart markgo

# Clear cache
curl -X POST -u admin:password http://localhost:3000/admin/cache/clear

# Reload articles
curl -X POST -u admin:password http://localhost:3000/admin/articles/reload
```

### Data Management

```bash
# Backup articles
docker compose exec markgo tar -czf /tmp/articles-backup.tar.gz articles/
docker compose cp markgo:/tmp/articles-backup.tar.gz ./

# Update articles
docker compose cp ./new-articles/ markgo:/app/articles/
docker compose exec markgo curl -X POST -u admin:password http://localhost:3000/admin/articles/reload
```

### Monitoring

```bash
# Check service health
docker compose ps

# View application metrics
curl http://localhost:3000/metrics

# Check nginx status
curl http://localhost/nginx_status
```

## Scaling and Performance

### Horizontal Scaling

To run multiple application instances:

```yaml
# In docker-compose.override.yml
services:
  markgo:
    deploy:
      replicas: 3
    ports:
      - "3000-3002:3000"
```

Update nginx upstream configuration:

```nginx
upstream markgo_backend {
    server markgo_1:3000;
    server markgo_2:3000;
    server markgo_3:3000;
}
```

### Performance Tuning

**Application:**
- Increase cache size: `CACHE_MAX_SIZE=5000`
- Tune cache TTL: `CACHE_TTL=2h`
- Adjust rate limits based on traffic

**Nginx:**
- Enable gzip compression (already configured)
- Set appropriate cache headers
- Tune worker processes

**Redis:**
- Configure persistence based on needs
- Set memory limits
- Enable compression

## Security

### Production Security Checklist

- [ ] Change default admin credentials
- [ ] Configure proper SSL certificates
- [ ] Set secure environment variables
- [ ] Enable firewall rules
- [ ] Configure fail2ban for rate limiting
- [ ] Regular security updates
- [ ] Monitor logs for suspicious activity

### Network Security

Services communicate via Docker network:
- Internal communication on `markgo_network`
- Only necessary ports exposed externally
- Nginx acts as security gateway

## Troubleshooting

### Common Issues

**Application won't start:**
```bash
# Check logs
docker compose logs markgo

# Verify configuration
docker compose config

# Test environment variables
docker compose exec markgo env | grep BLOG
```

**Template errors:**
```bash
# Check template syntax
docker compose exec markgo ls -la web/templates/

# Reload articles
curl -X POST -u admin:password http://localhost:3000/admin/articles/reload
```

**Performance issues:**
```bash
# Check resource usage
docker stats

# Monitor cache hit rate
curl http://localhost:3000/admin/stats
```

### Health Checks

Built-in health checks monitor:
- Application responsiveness
- Redis connectivity
- File system access

Access health status:
```bash
curl http://localhost:3000/health
```

## Backup and Recovery

### Automated Backups

Create backup script:

```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
docker compose exec markgo tar -czf /tmp/backup_$DATE.tar.gz articles/
docker compose cp markgo:/tmp/backup_$DATE.tar.gz ./backups/
```

### Recovery Process

1. Stop services: `docker compose down`
2. Restore articles: `cp -r backup/articles/* articles/`
3. Start services: `docker compose up -d`
4. Reload articles via admin API

## Development Workflow

### Local Development

1. **Setup:**
   ```bash
   cp docker-compose.override.yml.example docker-compose.override.yml
   docker compose up -d
   ```

2. **Development:**
   - Edit files locally
   - Changes auto-reload via Air
   - Debug via port 2345

3. **Testing:**
   ```bash
   # Run tests
   docker compose exec markgo go test ./...

   # Build for production
   docker compose exec markgo make build
   ```

### CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Deploy
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to production
        run: |
          docker compose -f deployments/docker-compose.yml up -d --build
```

## Support

For issues and questions:
- Check application logs: `docker compose logs markgo`
- Review this documentation
- Check project README.md
- Open GitHub issue

## Version Information

- Docker Compose: 3.8+
- Docker: 20.10+
- Application: See go.mod for dependencies
