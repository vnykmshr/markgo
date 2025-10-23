# MarkGo Quick Reference

## 🚀 Getting Started (2 Minutes)

```bash
# 1. Initialize blog
markgo init --quick

# 2. Start server  
markgo

# 3. Visit http://localhost:3000
# 4. Create first article
markgo new
```

## 📋 Commands

| Command | Description |
|---------|-------------|
| `markgo` | Start blog server |
| `markgo init` | Initialize new blog |
| `markgo new` | Create new article |
| `markgo stress-test` | Performance testing |

## 📁 File Structure

```
blog/
├── .env                 # Configuration
├── articles/            # Blog posts (.md files)
├── web/static/css/      # Styles
├── web/templates/       # HTML templates
└── README.md
```

## ✍️ Article Format

```markdown
---
title: "Article Title"
description: "SEO description"
author: "Your Name"
date: 2024-01-01T00:00:00Z
tags: ["tag1", "tag2"]
category: "category"
draft: false
featured: false
---

# Your Content

Write in **Markdown**!
```

## ⚙️ Key Configuration (.env)

```bash
# Blog Info
BLOG_TITLE=My Blog
BLOG_DESCRIPTION=Blog description
BLOG_AUTHOR=Your Name

# Server
PORT=3000
ENVIRONMENT=development

# Email (for contact forms)
EMAIL_HOST=smtp.gmail.com
EMAIL_USERNAME=your-email
EMAIL_PASSWORD=app-password

# Analytics
ANALYTICS_ENABLED=true
ANALYTICS_PROVIDER=google
ANALYTICS_TRACKING_ID=GA_ID
```

## 🎨 Customization

**CSS**: Edit `web/static/css/style.css`
**Templates**: Edit `web/templates/base.html`
**Content**: Add `.md` files to `articles/`

## 📊 Built-in URLs

| URL | Purpose |
|-----|---------|
| `/` | Homepage |
| `/articles` | All articles |
| `/tags` | All tags |
| `/search?q=term` | Search |
| `/rss` | RSS feed |
| `/sitemap.xml` | Sitemap |
| `/health` | Health check |

## 🚀 Performance Targets

- **Throughput**: ≥1000 req/s
- **95th Percentile**: <50ms  
- **Average Response**: <30ms
- **Success Rate**: >99%

## 🛠️ Development

```bash
# Build all tools
make build

# Run tests
make test

# Generate docs
make docs

# Performance test
markgo stress-test --url http://localhost:3000
```

## 🌐 Production Deployment

```bash
# 1. Update config
ENVIRONMENT=production
BASE_URL=https://yourdomain.com

# 2. Build
make build

# 3. Deploy binary + .env + articles + web folders
```

## 💡 Pro Tips

- Use `--date-prefix` for chronological file naming
- Set `TEMPLATE_HOT_RELOAD=true` for development
- Test performance with `markgo stress-test`
- Check `/health` endpoint for monitoring
- Use `draft: true` for work-in-progress articles

---

📖 **Full Guide**: [Getting Started](./GETTING-STARTED.md)  
🐛 **Issues**: [GitHub Issues](https://github.com/vnykmshr/markgo/issues)