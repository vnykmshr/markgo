# MarkGo

[![CI](https://github.com/vnykmshr/markgo/actions/workflows/ci.yml/badge.svg)](https://github.com/vnykmshr/markgo/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25.0+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance blog engine built with Go. MarkGo combines file-based content management with a dynamic web server.

**[Live Demo](https://vnykmshr.github.io/markgo/)**

## Features

**Performance**
- Server ready in under 1 second
- 30MB memory footprint, ~27MB binary
- In-memory caching with no external dependencies
- No runtime requirements (Node.js, PHP, databases)

**Content Management**
- Markdown files with YAML frontmatter
- Git-based workflow for version control
- CLI tools for article creation and export

**Production Ready**
- Docker deployment support
- Static site export (GitHub Pages, Netlify, Vercel)
- Built-in security: rate limiting, CORS, input validation
- SEO automation: sitemaps, Schema.org, Open Graph, Twitter Cards
- Full-text search and RSS/JSON feeds

**Developer Experience**
- Clean architecture with separated concerns
- Unified CLI with subcommands
- Environment variable configuration
- Extensive documentation

## Quick Start

### Option 1: Download Release

```bash
# Download and extract
wget https://github.com/vnykmshr/markgo/releases/latest
tar -xzf markgo-*.tar.gz && cd markgo

# Initialize and start
./markgo init --quick
./markgo serve

# Visit http://localhost:3000
```

### Option 2: Build from Source

```bash
git clone https://github.com/vnykmshr/markgo
cd markgo
make build

./build/markgo init --quick
./build/markgo serve
```

### Create Articles

```bash
# Interactive creation
markgo new

# Quick creation
markgo new --title "Hello World" --tags "introduction,getting-started"
```

### Deploy to GitHub Pages

1. Fork this repository
2. Enable GitHub Pages: Settings > Pages > Source: GitHub Actions
3. Push changes - site auto-deploys to `https://yourusername.github.io/markgo`

See [Getting Started Guide](docs/GETTING-STARTED.md) for detailed setup instructions.

## Project Structure

```
markgo/
├── cmd/markgo/       # Unified CLI binary
├── internal/         # Private packages
│   ├── commands/     # CLI commands (serve, init, new, export)
│   ├── config/       # Configuration management
│   ├── handlers/     # HTTP handlers
│   ├── middleware/   # HTTP middleware
│   ├── models/       # Data structures
│   └── services/     # Business logic
├── web/              # Templates and static files
│   ├── static/       # CSS, JS, images
│   └── templates/    # HTML templates
├── articles/         # Blog content (Markdown)
├── deployments/      # Docker and deployment configs
└── docs/             # Documentation
```

## Writing Articles

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

Write your article using Markdown...
```

### Article Features

- Automatic excerpt generation
- Reading time estimation
- Tag and category organization
- SEO automation with Schema.org structured data
- Dynamic sitemap generation
- Social media optimization (Open Graph, Twitter Cards)
- Multiple Markdown file extensions: `.md`, `.markdown`, `.mdown`, `.mkd`

See [Complete Project Guide](docs/project-guide.md) for advanced features.

## Configuration

MarkGo is configured via environment variables. Copy `.env.example` to `.env`:

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

## Deployment

### Static Sites

```bash
# GitHub Pages (auto-detects repo URL)
make export-github-pages

# Any static host
make export-static

# Custom configuration
./build/markgo export --output ./public --base-url https://yourdomain.com
```

Example: [vnykmshr.github.io/markgo](https://vnykmshr.github.io/markgo/)

See [Static Export Guide](docs/static-export.md) for hosting options.

### Docker

```bash
docker-compose up -d
```

### Manual Deployment

```bash
# Build for production
make prod-build

# Deploy to server
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
make build-all      # Build all tools
make export         # Build static export tool
make export-static  # Export site to ./dist/
make docker         # Build and run Docker container
```

## Testing

```bash
make test           # Run tests
make test-race      # Run tests with race detection
make coverage       # Generate coverage report
make benchmark      # Run benchmarks
make check          # Run all quality checks
```

## Documentation

### Guides
- [Complete Project Guide](docs/project-guide.md) - Everything about MarkGo
- [System Overview](docs/system-overview.md) - Architecture and performance
- [Architecture Guide](docs/architecture.md) - Technical design details
- [Deployment Guide](docs/deployment.md) - Production deployment

### Reference
- [Configuration Guide](docs/configuration.md) - Configuration options
- [Theme Customization](docs/themes.md) - Customizing appearance
- [API Documentation](docs/api.md) - HTTP endpoints
- [Contributing Guide](CONTRIBUTING.md) - How to contribute
- [Troubleshooting](docs/troubleshooting.md) - Common issues

## Development

### Prerequisites

```bash
make install-dev-tools    # Install development tools
make lint                 # Run linting
make fmt                  # Format code
make check                # Run all checks
```

### Hot Reload

```bash
make dev    # Start development server with hot reload
```

Changes to templates and configuration are automatically reloaded.

## Contributing

Contributions welcome. See [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/name`
3. Make changes and add tests
4. Run `make check`
5. Commit: `git commit -m 'Add feature'`
6. Push: `git push origin feature/name`
7. Open a Pull Request

## Performance Comparison

| Platform | Type | Memory | Dependencies |
|----------|------|--------|--------------|
| MarkGo   | Dynamic server | ~30MB | Single binary (~27MB) |
| Ghost    | Dynamic server | ~200MB | Node.js + SQLite |
| WordPress| Dynamic server | ~100MB | PHP + MySQL |
| Hugo     | Static generator | Build-time | Go binary |

See [BENCHMARKS.md](BENCHMARKS.md) for detailed metrics.

## Key Advantages

- **Git-based content**: Version control and easy backups
- **Performance**: Native Go speed and efficiency
- **Deployment flexibility**: Dynamic server or static export
- **Simplicity**: Single binary, no external dependencies
- **Security**: Rate limiting and input validation built-in
- **Portability**: File-based content, easy migration

Read the [System Overview](docs/system-overview.md) and [Architecture Guide](docs/architecture.md) for technical details.

## Links

- [Live Demo](https://vnykmshr.github.io/markgo/)
- [Documentation](docs/)
- [Issue Tracker](https://github.com/vnykmshr/markgo/issues)
- [Discussions](https://github.com/vnykmshr/markgo/discussions)

## License

MIT License - see [LICENSE](LICENSE) file.

---

Made with Go
