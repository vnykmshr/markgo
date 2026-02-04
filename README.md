# MarkGo

[![CI](https://github.com/vnykmshr/markgo/actions/workflows/ci.yml/badge.svg)](https://github.com/vnykmshr/markgo/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25.0+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance blog engine built with Go. MarkGo combines file-based content management with a dynamic web server.

**[Live Demo](https://vnykmshr.github.io/markgo/)**

## Quick Start

**Download a release** from [GitHub Releases](https://github.com/vnykmshr/markgo/releases), or **build from source**:

```bash
git clone https://github.com/vnykmshr/markgo && cd markgo
make build

./build/markgo init --quick
./build/markgo serve
# Visit http://localhost:3000
```

See [Getting Started Guide](docs/GETTING-STARTED.md) for detailed setup instructions.

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

## Usage

```bash
markgo init [--quick]                    # Initialize a new blog
markgo serve                             # Start the web server
markgo new [--title "..." --tags "..."]  # Create an article
markgo export --output ./public          # Export to static site
```

## Configuration

Configure via environment variables. Copy `.env.example` to `.env`:

```bash
BLOG_TITLE="Your Blog Title"
BLOG_AUTHOR="Your Name"
BASE_URL="https://yourdomain.com"
CACHE_ENABLED=true
SEARCH_ENABLED=true
```

See [Configuration Guide](docs/configuration.md) for complete options.

## Deployment

**Static export** (GitHub Pages, Netlify, Vercel):
```bash
markgo export --output ./public --base-url https://yourdomain.com
```

**Docker**:
```bash
docker compose up -d
```

**Manual** (systemd):
```bash
make build
scp build/markgo server:/usr/local/bin/
sudo cp deployments/markgo.service /etc/systemd/system/
sudo systemctl enable --now markgo
```

See [Deployment Guide](docs/deployment.md) and [Static Export Guide](docs/static-export.md) for details.

## Development

```bash
make dev             # Start dev server with hot reload
make build           # Build for current platform
make build-release   # Build for all platforms
make fmt             # Format code
make lint            # Run linter (golangci-lint v2)
make test            # Run tests
make test-race       # Run tests with race detection
make coverage        # Generate coverage report
```

## Documentation

- [Getting Started](docs/GETTING-STARTED.md) - Setup walkthrough
- [Configuration](docs/configuration.md) - All configuration options
- [Deployment](docs/deployment.md) - Production deployment strategies
- [Static Export](docs/static-export.md) - GitHub Pages, Netlify, Vercel
- [Architecture](docs/architecture.md) - Technical architecture and design
- [API](docs/API.md) - HTTP endpoints and responses
- [Runbook](docs/RUNBOOK.md) - Operations and troubleshooting
- [Benchmarks](docs/BENCHMARKS.md) - Performance metrics

## Contributing

Contributions welcome. See [Contributing Guide](.github/CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) file.

---

Made with Go
