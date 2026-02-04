# MarkGo Deployment Guide

**Version:** 2.3.0

This guide covers production deployment of MarkGo using various methods.

## 🚀 Quick Deploy Options

### Option 1: Docker (Recommended)
```bash
# Clone and configure
git clone https://github.com/vnykmshr/markgo.git
cd markgo
cp .env.example .env  # Edit with your settings

# Deploy with Docker Compose
docker compose up -d

# Access: http://localhost:3000
```

### Option 2: Binary Deployment
```bash
# Download or build binary
make build-release

# Copy to server
scp build/markgo-linux-amd64 server:/usr/local/bin/markgo

# Install systemd service
sudo cp deployments/etc/systemd/system/markgo.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable markgo
sudo systemctl start markgo
```

### Option 3: Nginx + MarkGo
```bash
# Install configurations
sudo cp deployments/etc/nginx/nginx.conf /etc/nginx/
sudo cp deployments/etc/nginx/conf.d/markgo.conf /etc/nginx/conf.d/

# Update domain in markgo.conf
sudo sed -i 's/yourdomain.com/your-actual-domain.com/g' /etc/nginx/conf.d/markgo.conf

# Test and reload
sudo nginx -t && sudo systemctl reload nginx
```

## 📁 Configuration Files

### SystemD Service
Location: `/etc/systemd/system/markgo.service`
- Optimized for security with hardening
- Journal logging integration
- Resource limits configured
- Environment file support

### Nginx Configuration
Location: `/etc/nginx/conf.d/markgo.conf`
- SSL termination ready
- Rate limiting configured
- Static file caching optimized
- Security headers included

## 🛡️ Security Setup

### 1. Create User Account
```bash
sudo useradd --system --no-create-home --shell /bin/false markgo
sudo mkdir -p /opt/markgo/{articles,web,logs}
sudo chown -R markgo:markgo /opt/markgo
```

### 2. SSL Certificate Setup
```bash
# Using Let's Encrypt
sudo certbot certonly --nginx -d yourdomain.com

# Or place your certificates:
# /etc/ssl/certs/yourdomain.com.crt
# /etc/ssl/private/yourdomain.com.key
```

### 3. Environment Configuration
```bash
# Create production environment file
sudo tee /opt/markgo/.env << EOF
ENVIRONMENT=production
PORT=3000
BASE_URL=https://yourdomain.com
BLOG_TITLE="Your Blog Title"
BLOG_AUTHOR="Your Name"
# ... other settings
EOF

sudo chown markgo:markgo /opt/markgo/.env
sudo chmod 600 /opt/markgo/.env
```

## 🔧 Production Checklist

### Before Deployment
- [ ] Set strong admin credentials
- [ ] Configure email settings for contact form
- [ ] Update domain names in configs
- [ ] Set up SSL certificates
- [ ] Configure rate limiting values
- [ ] Set up log rotation
- [ ] Configure firewall rules

### After Deployment
- [ ] Verify service starts: `systemctl status markgo`
- [ ] Check logs: `journalctl -u markgo -f`
- [ ] Test endpoints: `/health`, `/metrics`
- [ ] Verify SSL: `curl -I https://yourdomain.com`
- [ ] Test contact form and admin panel
- [ ] Set up monitoring and backups

## 📊 Monitoring & Maintenance

### Health Checks
```bash
# Service status
systemctl status markgo

# Application health
curl -f http://localhost:3000/health

# Metrics endpoint
curl http://localhost:3000/metrics
```

### Log Management
```bash
# View logs
journalctl -u markgo -f

# Log rotation (systemd handles automatically)
# Custom rotation in /etc/logrotate.d/markgo if needed
```

### Performance Tuning
```bash
# Nginx worker processes (adjust for CPU cores)
worker_processes auto;

# MarkGo concurrency (via GOMAXPROCS)
Environment=GOMAXPROCS=4
```

## 🚨 Troubleshooting

### Common Issues

**Service won't start:**
```bash
# Check service logs
journalctl -u markgo --no-pager
# Check file permissions
ls -la /opt/markgo
```

**High memory usage:**
```bash
# Monitor memory
systemctl status markgo
# Adjust cache settings in .env
CACHE_MAX_SIZE=500
```

**SSL errors:**
```bash
# Test SSL configuration
nginx -t
# Check certificate validity
openssl x509 -in /etc/ssl/certs/yourdomain.com.crt -text -noout
```

## 🔄 Updates & Maintenance

### Updating MarkGo
```bash
# Stop service
sudo systemctl stop markgo

# Backup current binary
sudo cp /usr/local/bin/markgo /usr/local/bin/markgo.bak

# Deploy new binary
sudo cp new-markgo-binary /usr/local/bin/markgo
sudo chown root:root /usr/local/bin/markgo
sudo chmod +x /usr/local/bin/markgo

# Start service
sudo systemctl start markgo
```

### Configuration Updates
```bash
# Edit configuration
sudo nano /opt/markgo/.env

# Reload service
sudo systemctl reload markgo
```

## 🌐 Multi-Instance Setup

For high availability, deploy multiple instances:

```yaml
# docker compose.yml
version: '3.8'
services:
  markgo1:
    image: markgo:latest
    environment:
      PORT: 3001
  markgo2:
    image: markgo:latest  
    environment:
      PORT: 3002
  nginx:
    image: nginx
    # Configure upstream with both instances
```

## 📞 Support

- **Documentation**: [GitHub Repository](https://github.com/vnykmshr/markgo)
- **Issues**: [GitHub Issues](https://github.com/vnykmshr/markgo/issues)
- **Performance**: Sub-second cold start, single-digit ms cached responses
- **Architecture**: Single ~27MB binary, zero external dependencies

---

**MarkGo Engine** - Production-ready, high-performance blog engine built with Go