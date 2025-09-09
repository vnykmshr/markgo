# MarkGo Engine

[![Go Version](https://img.shields.io/badge/Go-1.25.0+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Audited-green.svg)](#security)
[![Performance](https://img.shields.io/badge/Cold%20Start-17ms-brightgreen.svg)](#performance)

A modern, high-performance file-based blog engine built with Go. MarkGo combines the simplicity of file-based content management with the power of a dynamic web server, delivering **blazing-fast 17ms cold start** and enterprise-grade performance.

## ✨ Features

### 🚀 Exceptional Performance
- **17ms cold start time** - fastest in class
- **Sub-microsecond response times** for core operations
- **Zero-allocation caching** with enterprise obcache-go integration  
- **38MB single binary** deployment with no dependencies
- **Concurrent request handling** with optimized goroutines

### 📝 File-Based Content Management
- **Markdown articles** with YAML frontmatter
- **Git-friendly workflow** - version control your content
- **CLI tools** for creating and managing articles
- **Hot-reload** in development for instant feedback

### 🔧 Production Ready
- **Docker deployment** with comprehensive configuration
- **Built-in security** with rate limiting, CORS, and input validation
- **SEO optimized** with sitemaps, structured data, and meta tags
- **Full-text search** across all content
- **RSS/JSON feeds** for content syndication

### 🎨 Developer Experience
- **Clean architecture** with separated concerns
- **Comprehensive testing** with 80%+ coverage
- **Configuration-driven** behavior via environment variables
- **Extensive documentation** and examples

## 🚀 Quick Start

### Prerequisites
- Go 1.25.0 or later
- Make (for build automation)

### Installation

```bash
# Clone the repository
git clone https://github.com/vnykmshr/markgo
cd markgo

# Copy environment configuration
cp .env.example .env

# Install dependencies
make deps

# Start development server with hot reload
make dev
```

Visit `http://localhost:8080` to see your blog!

### Create Your First Article

```bash
# Build the CLI tool
make new-article

# Create an article interactively
./build/new-article --interactive

# Or create directly
./build/new-article --title "Hello World" --tags "introduction,getting-started"
```

> 🎯 **New to MarkGo?** Check out our [Complete Project Guide](docs/project-guide.md) for step-by-step tutorials, configuration options, and deployment strategies.

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
- **SEO metadata** generation
- **Social media optimization**
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
make build-all      # Build for all platforms
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

### Performance Comparison

| Platform | Cold Start Time | Memory Usage | Binary Size |
|----------|----------------|--------------|-------------|
| MarkGo   | **17ms**       | ~30MB        | ~38MB       |
| Ghost    | 150-300ms      | ~200MB       | Node.js + deps |
| WordPress| 500-1000ms     | ~100MB       | PHP + MySQL |
| Hugo     | Static only    | Build time   | Go binary   |

### Key Advantages

- **Developer Workflow**: Git-based content management
- **Performance**: Go's native speed and efficiency
- **Simplicity**: Single binary, no external dependencies
- **Security**: Built-in rate limiting and input validation
- **Portability**: File-based content, easy backups and migration

> 🏗️ **Architecture Deep-Dive**: Learn about the technical design and performance optimizations in our [System Overview](docs/system-overview.md) and [Architecture Guide](docs/architecture.md).

## 🔗 Links

- [📖 Complete Documentation](docs/) - All guides and references
- [🐛 Issue Tracker](https://github.com/vnykmshr/markgo/issues) - Bug reports and feature requests
- [💬 Discussions](https://github.com/vnykmshr/markgo/discussions) - Community discussions
- [🚀 Quick Deploy](docs/deployment.md) - Production deployment guide

## 💡 Inspiration

MarkGo was inspired by the need for a high-performance blog engine that combines:
- The simplicity of static site generators
- The flexibility of dynamic web applications  
- The developer experience of modern tools
- The reliability of Go's ecosystem

Perfect for developers, technical writers, and anyone who values performance and simplicity.

## 📚 Comprehensive Documentation

This README provides a quick overview. For detailed information, explore our comprehensive documentation:

- **[📘 Complete Project Guide](docs/project-guide.md)** - Everything you need to know about MarkGo Engine
- **[🏛️ System Overview](docs/system-overview.md)** - Technical architecture and performance characteristics
- **[📐 Architecture Guide](docs/architecture.md)** - In-depth technical design and patterns
- **[🚀 Deployment Guide](docs/deployment.md)** - Production deployment strategies

---

**Made with ❤️ and Go**

⭐ If you find MarkGo useful, please star the repository!