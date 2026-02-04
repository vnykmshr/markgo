# Getting Started with MarkGo

Welcome to MarkGo! This guide will get you from zero to a running blog in under 5 minutes.

## 🚀 Quick Start (5 Minutes)

### Step 1: Install MarkGo

**Option A: Download Release (Recommended)**
1. Visit [MarkGo Releases](https://github.com/vnykmshr/markgo/releases)
2. Download for your platform (Linux, macOS, Windows)
3. Extract and move to your PATH

**Option B: Build from Source**
```bash
git clone https://github.com/vnykmshr/markgo.git
cd markgo
make build
```

### Step 2: Initialize Your Blog

Create a new blog in seconds:

```bash
# Quick setup with defaults
markgo init --quick

# Or interactive setup
markgo init
```

This creates:
- ✅ Complete directory structure
- ✅ Configuration file (`.env`)
- ✅ Sample articles
- ✅ Basic templates and CSS
- ✅ README and .gitignore

### Step 3: Start Your Blog

```bash
markgo serve
```

Visit http://localhost:3000 - Your blog is live! 🎉

### Step 4: Create Your First Article

```bash
markgo new
```

Follow the interactive prompts to create your first blog post.

## 📁 Project Structure

Your blog is organized like this:

```
my-blog/
├── .env                 # Configuration
├── articles/            # Your blog posts (Markdown)
│   ├── welcome.md
│   └── getting-started.md
├── web/
│   ├── static/          # CSS, JS, images
│   │   └── css/style.css
│   └── templates/       # HTML templates
│       └── base.html
├── docs/                # Documentation
├── README.md
└── .gitignore
```

## ✍️ Writing Articles

### Using the CLI (Recommended)

```bash
# Interactive mode - guided setup
markgo new

# Quick creation
markgo new --title "My First Post" --tags "tutorial,markgo"

# With date prefix
markgo new --title "News Update" --date-prefix
```

### Manual Creation

Create a `.md` file in `articles/`:

```markdown
---
title: "Your Awesome Title"
description: "SEO-friendly description"
author: "Your Name"
date: 2024-01-01T00:00:00Z
tags: ["golang", "blog", "tutorial"]
category: "technology"
draft: false
featured: false
---

# Your Content Here

Write your article using **Markdown** syntax!

## Subheadings Work Great

- Bullet points
- Are supported
- Too!

```code blocks
Also work perfectly
```

Images, links, tables - everything Markdown supports!
```

## ⚙️ Configuration

Edit `.env` to customize your blog:

### Essential Settings

```bash
# Blog Information
BLOG_TITLE=My Awesome Blog
BLOG_DESCRIPTION=A blog about amazing things
BLOG_AUTHOR=Your Name
BLOG_AUTHOR_EMAIL=you@example.com

# Server Settings
PORT=3000
BASE_URL=http://localhost:3000

# For production:
# ENVIRONMENT=production
# BASE_URL=https://yourdomain.com
```

### Email Setup (Contact Forms)

```bash
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=noreply@yourdomain.com
EMAIL_TO=contact@yourdomain.com
EMAIL_USE_SSL=true
```

### Performance Tuning

```bash
# Cache Settings
CACHE_TTL=3600              # 1 hour
CACHE_MAX_SIZE=1000         # Max cached items
CACHE_CLEANUP_INTERVAL=600  # 10 minutes

# Rate Limiting
RATE_LIMIT_GENERAL_REQUESTS=100
RATE_LIMIT_GENERAL_WINDOW=900s
```

## 🎨 Customization

### Styling

Edit `web/static/css/style.css`:

```css
/* Add your custom styles */
.my-custom-class {
    color: #007bff;
    font-weight: bold;
}

/* Override defaults */
body {
    font-family: 'Your Preferred Font', sans-serif;
}
```

### Templates

Modify `web/templates/base.html` for layout changes:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}} - My Blog</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <!-- Your custom header -->
    <main>{{template "content" .}}</main>
    <!-- Your custom footer -->
</body>
</html>
```

## 🚀 Development Workflow

### Live Development

```bash
# Start server (auto-reloads templates)
markgo

# In another terminal, create content
markgo new

# Edit files - changes appear instantly!
```

### Testing Performance

```bash
# Run Go benchmarks
go test -bench=. -benchmem ./...

# For load testing, see https://github.com/vnykmshr/webstress
```

## 📊 Monitoring & Analytics

### Built-in Metrics

Visit these endpoints while your blog is running:

- `http://localhost:3000/health` - Health check
- `http://localhost:3000/metrics` - Performance metrics
- `http://localhost:3000/admin/stats` - Detailed statistics

### Google Analytics Setup

```bash
ANALYTICS_ENABLED=true
ANALYTICS_PROVIDER=google
ANALYTICS_TRACKING_ID=GA_MEASUREMENT_ID
```

### Plausible Analytics

```bash
ANALYTICS_ENABLED=true
ANALYTICS_PROVIDER=plausible
ANALYTICS_DOMAIN=yourdomain.com
```

## 🌐 SEO Features

MarkGo includes powerful SEO features out of the box:

- **Automatic Sitemaps**: `yourdomain.com/sitemap.xml`
- **RSS Feeds**: `yourdomain.com/rss` (XML) and `yourdomain.com/feed.json`
- **Meta Tags**: Open Graph and Twitter Cards
- **Structured Data**: JSON-LD for search engines
- **Fast Loading**: Core Web Vitals optimized

## 🚢 Deployment

### Production Setup

1. **Update Configuration**:
   ```bash
   ENVIRONMENT=production
   BASE_URL=https://yourdomain.com
   LOG_LEVEL=warn
   ```

2. **Build for Production**:
   ```bash
   make build
   ```

3. **Deploy Options**:

   **Docker**:
   ```bash
   make docker
   ```

   **Systemd** (Linux):
   ```bash
   sudo cp deployments/etc/systemd/system/markgo.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable markgo
   sudo systemctl start markgo
   ```

   **Manual**:
   ```bash
   # Copy binary and files to server
   scp build/markgo user@server:/opt/markgo/
   scp -r .env articles web user@server:/opt/markgo/
   
   # Start on server
   ./markgo
   ```

### Domain Setup

1. Point your domain to your server
2. Update `BASE_URL` in `.env`
3. Set up HTTPS (recommended: Let's Encrypt)
4. Update CORS settings if needed

## 🔧 Troubleshooting

### Common Issues

**Port Already in Use**:
```bash
# Change port in .env
PORT=3001
```

**Permission Denied**:
```bash
# Make binary executable
chmod +x markgo
```

**Templates Not Loading**:
```bash
# Check templates path in .env
TEMPLATES_PATH=./web/templates
```

**Articles Not Showing**:
```bash
# Check articles path and permissions
ARTICLES_PATH=./articles
ls -la articles/
```

### Getting Help

- 📖 Check `docs/` directory for detailed documentation
- 🐛 Report issues: [GitHub Issues](https://github.com/vnykmshr/markgo/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/vnykmshr/markgo/discussions)
- 📧 Email: maintainer@markgo.dev

## 🎯 Next Steps

Now that you're up and running:

1. **Customize Your Theme**: Modify CSS and templates
2. **Set Up Email**: Enable contact forms
3. **Configure Analytics**: Track your visitors
4. **Write Great Content**: The most important part!
5. **Deploy to Production**: Share with the world

## 📚 Advanced Guides

- [Configuration Reference](./configuration.md)
- [Architecture Guide](./architecture.md)
- [Deployment Guide](./deployment.md)
- [API Documentation](./API.md)
- [Contributing](../CONTRIBUTING.md)

---

**Welcome to the MarkGo community!** 🎉

*Need help? Don't hesitate to ask in our GitHub Discussions.*