# MarkGo Engine

[![CI](https://github.com/vnykmshr/markgo/actions/workflows/ci.yml/badge.svg)](https://github.com/vnykmshr/markgo/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25.0+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Audited-green.svg)](#security)
[![Performance](https://img.shields.io/badge/Cold%20Start-Fast-brightgreen.svg)](#performance)
[![Demo](https://img.shields.io/badge/Demo-Live%20Site-blue.svg)](https://vnykmshr.github.io/markgo/)

A modern, high-performance file-based blog engine built with Go. MarkGo combines the simplicity of file-based content management with the power of a dynamic web server, delivering sub-second cold start and enterprise-grade performance.

🌐 **[Live Demo](https://vnykmshr.github.io/markgo/)** - See MarkGo in action with automatic GitHub Pages deployment

## ✨ Features

### 🚀 Performance
- **Fast startup**: Server ready in <1 second
- **Low memory**: ~30MB resident, 38MB single binary
- **In-memory caching**: obcache-go, no external dependencies
- **No runtime required**: No Node.js, PHP, or database needed
- **Concurrent**: Handles multiple requests efficiently

### 📝 File-Based Content Management
- **Markdown articles** with YAML frontmatter
- **Git-friendly workflow** - version control your content
- **CLI tools** for creating and managing articles
- **Hot-reload** in development for instant feedback

### 🔧 Production Ready
- **Docker deployment** with comprehensive configuration
- **Static site export** for GitHub Pages, Netlify, Vercel deployment
- **Built-in security** with rate limiting, CORS, and input validation
- **Advanced SEO automation** with dynamic sitemaps, Schema.org markup, Open Graph tags, and Twitter Cards
- **Full-text search** across all content
- **RSS/JSON feeds** for content syndication

### 🎨 Developer Experience
- **Clean architecture** with separated concerns
- **Comprehensive testing** with 80%+ coverage
- **Configuration-driven** behavior via environment variables
- **Extensive documentation** and examples

## 🚀 Quick Start (5 Minutes)

### Option 1: Download Release (Recommended)

1. **Download**: [Latest release](https://github.com/vnykmshr/markgo/releases/latest) for your platform
2. **Extract**: `tar -xzf markgo-*.tar.gz && cd markgo`
3. **Initialize**: `./markgo init --quick`
4. **Start**: `./markgo`
5. **Visit**: http://localhost:3000

### Option 2: Build from Source

```bash
# Clone and build
git clone https://github.com/vnykmshr/markgo
cd markgo
make build-all

# Initialize your blog
./build/init --quick

# Start your blog
./build/markgo
```

### Create Your First Article

```bash
# Interactive creation
markgo new-article

# Quick creation
markgo new-article --title "Hello World" --tags "introduction,getting-started"
```

🎉 **Your blog is now running at http://localhost:3000!**

### Option 3: Deploy to GitHub Pages (Static)

1. **Fork this repository** or use as template
2. **Enable GitHub Pages** in Settings > Pages > Source: GitHub Actions  
3. **Push changes** - site auto-deploys at `https://yourusername.github.io/markgo`
4. **See it live**: [Demo Site](https://vnykmshr.github.io/markgo/)

> 📖 **New to MarkGo?** Read our [Getting Started Guide](docs/GETTING-STARTED.md) for the complete 5-minute setup tutorial.

## 📁 Project Structure

```
markgo/
├── cmd/
│   ├── server/          # Main blog server
│   └── new-article/     # CLI article creation tool
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data structures
│   ├── services/        # Business logic
│   └── utils/           # Utility functions
├── web/
│   ├── static/          # CSS, JS, images
│   └── templates/       # HTML templates
├── articles/            # Your blog articles (Markdown)
├── deployments/         # Docker and deployment configs
└── docs/                # Documentation
```

## 📝 Writing Articles

Articles are Markdown files with YAML frontmatter:

```markdown
---
title: "Your Article Title"
description: "Article description for SEO"
date: 2024-01-15T10:00:00Z
published: true
tags: ["golang", "blogging", "tutorial"]
categories: ["Technical"]
author: "Your Name"
---

# Your Article Content

Write your article content here using Markdown...
```

### Article Features

- **Automatic excerpts** with configurable length
- **Reading time estimation**
- **Tag and category organization**
- **Advanced SEO automation** with Schema.org JSON-LD structured data
- **Dynamic sitemap generation** with automatic updates
- **Social media optimization** with Open Graph and Twitter Card tags
- **Multiple Markdown extensions** supported (`.md`, `.markdown`, `.mdown`, `.mkd`)

> 📖 **Learn More**: See the [Complete Project Guide](docs/project-guide.md) for detailed information about content creation, customization, and advanced features.

## 🛠️ Configuration

MarkGo is configured via environment variables. Copy `.env.example` to `.env` and customize:

```bash
# Basic configuration
BLOG_TITLE="Your Blog Title"
BLOG_DESCRIPTION="Your blog description"
BLOG_AUTHOR="Your Name"
BASE_URL="https://yourdomain.com"

# Performance
CACHE_ENABLED=true
CACHE_TTL=3600

# Features
SEARCH_ENABLED=true
RSS_ENABLED=true
CONTACT_ENABLED=true
```

See [Configuration Guide](docs/configuration.md) for complete options.

## 🚀 Deployment

### Static Sites (GitHub Pages, Netlify, Vercel)

```bash
# Export for GitHub Pages (auto-detects repo URL)
make export-github-pages

# Export for any static host
make export-static

# Custom export with options
./build/export --output ./public --base-url https://yourdomain.com
```

**Live Example**: [vnykmshr.github.io/markgo](https://vnykmshr.github.io/markgo/) - Auto-deployed via GitHub Actions

See [Static Export Guide](docs/static-export.md) for complete hosting options.

### Docker (Recommended)

```bash
# Build and run with Docker Compose
docker-compose up -d
```

### Manual Deployment

```bash
# Build for production
make prod-build

# Copy binary to server
scp build/markgo-linux server:/usr/local/bin/

# Install systemd service (Linux)
sudo cp deployments/markgo.service /etc/systemd/system/
sudo systemctl enable markgo
sudo systemctl start markgo
```

### Build Commands

```bash
make build          # Build for current platform
make build-linux    # Build for Linux
make build-all      # Build all tools (server, export, new-article, etc.)
make export         # Build static export tool
make export-static  # Export site to ./dist/
make docker         # Build and run Docker container
```

## 🧪 Testing

```bash
make test           # Run tests
make test-race      # Run tests with race detection
make coverage       # Generate coverage report
make benchmark      # Run benchmarks
make check          # Run all quality checks (fmt, vet, lint, test)
```

## 📚 Documentation

### 📖 Comprehensive Guides
- **[Complete Project Guide](docs/project-guide.md)** - Everything you need to know about MarkGo Engine
- **[System Overview](docs/system-overview.md)** - Technical architecture and performance characteristics  
- **[Architecture Guide](docs/architecture.md)** - In-depth technical architecture details
- **[Deployment Guide](docs/deployment.md)** - Production deployment instructions

### 🛠️ Reference Documentation
- [Configuration Guide](docs/configuration.md) - Complete configuration options
- [Theme Customization](docs/themes.md) - Customizing appearance
- [API Documentation](docs/api.md) - HTTP endpoints and responses
- [Contributing Guide](CONTRIBUTING.md) - How to contribute to the project
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions

## 🔧 Development

### Prerequisites for Development

```bash
# Install development tools
make install-dev-tools

# Run linting
make lint

# Format code
make fmt

# Run all checks
make check
```

### Hot Reload Development

```bash
# Start development server with hot reload
make dev

# Or manually
air
```

Changes to templates and configuration are automatically reloaded.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Run `make check` to ensure quality
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Why MarkGo?

### Performance Characteristics

| Platform | Type | Memory Usage | Dependencies |
|----------|------|--------------|--------------|
| MarkGo   | Dynamic server | ~30MB | Single binary (38MB) |
| Ghost    | Dynamic server | ~200MB | Node.js + SQLite |
| WordPress| Dynamic server | ~100MB | PHP + MySQL |
| Hugo     | Static generator | Build-time only | Go binary |

See [BENCHMARKS.md](BENCHMARKS.md) for detailed performance metrics.

### Key Advantages

- **Developer Workflow**: Git-based content management
- **Performance**: Go's native speed and efficiency
- **Deployment Flexibility**: Dynamic server OR static site export
- **Simplicity**: Single binary, no external dependencies
- **Security**: Built-in rate limiting and input validation
- **Portability**: File-based content, easy backups and migration

> 🏗️ **Architecture Deep-Dive**: Learn about the technical design and performance optimizations in our [System Overview](docs/system-overview.md) and [Architecture Guide](docs/architecture.md).

## 🔗 Links

- [🌐 **Live Demo**](https://vnykmshr.github.io/markgo/) - See MarkGo in action
- [📖 Complete Documentation](docs/) - All guides and references
- [📄 Static Export Guide](docs/static-export.md) - GitHub Pages deployment
- [🐛 Issue Tracker](https://github.com/vnykmshr/markgo/issues) - Bug reports and feature requests
- [💬 Discussions](https://github.com/vnykmshr/markgo/discussions) - Community discussions
- [🚀 Quick Deploy](docs/deployment.md) - Production deployment guide

---

**Made with ❤️ and Go**

⭐ If you find MarkGo useful, please star the repository!