---
title: "Welcome to MarkGo Engine"
description: "Getting started with MarkGo, a high-performance file-based blog engine written in Go"
date: 2024-01-15T10:00:00Z
tags: ["getting-started", "markgo", "golang", "blogging"]
categories: ["Documentation"]
featured: true
draft: false
author: "MarkGo Team"
---

# Welcome to MarkGo Engine

MarkGo is a modern, high-performance blog engine built with Go, designed for developers who want the simplicity of file-based content management with the power of a dynamic web server.

## Why MarkGo?

### 🚀 Blazing Fast Performance
- **Sub-100ms response times** thanks to Go's native performance
- **Intelligent caching** reduces server load and improves user experience
- **Single binary deployment** with no external dependencies

### 📝 Developer-Friendly Workflow
- **Markdown-first** content creation with YAML frontmatter
- **Git-friendly** workflow - version control your content
- **CLI tools** for creating and managing articles
- **Hot-reload** in development for instant feedback

### 🔧 Production Ready
- **Docker deployment** with comprehensive configuration
- **Built-in security** with rate limiting, CORS, and input validation
- **SEO optimized** with sitemaps, structured data, and meta tags
- **Search functionality** with full-text search across all content

## Getting Started

### 1. Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/markgo
cd markgo

# Install dependencies
make deps

# Start development server with hot reload
make dev
```

### 2. Create Your First Article

Use the built-in CLI tool to create new articles:

```bash
./build/new-article --interactive
```

Or create manually:

```bash
# Build the CLI tool first
make new-article

# Create a new article
./build/new-article --title "My First Post" --tags "tutorial,example"
```

### 3. Customize Your Blog

Edit the configuration in `.env`:

```bash
cp .env.example .env
# Edit .env with your blog details
```

## Key Features

### File-Based Content Management
All your articles are stored as Markdown files in the `/articles` directory. Each article includes YAML frontmatter for metadata:

```yaml
---
title: "Your Article Title"
description: "Article description for SEO"
date: 2024-01-15T10:00:00Z
tags: ["tag1", "tag2"]
categories: ["Category"]
featured: false
draft: false
author: "Your Name"
---

Your article content goes here...
```

### Full-Text Search
MarkGo includes built-in search functionality that indexes all your articles and provides relevant results with scoring.

### Contact Forms
Built-in contact form with email integration for visitor inquiries.

### RSS & JSON Feeds
Automatic generation of RSS and JSON feeds for content syndication.

### Docker Deployment
Production-ready Docker configuration with Nginx reverse proxy.

## Next Steps

- [Configuration Guide](https://github.com/yourusername/markgo/blob/main/docs/configuration.md)
- [Deployment Guide](https://github.com/yourusername/markgo/blob/main/docs/deployment.md)
- [Theme Customization](https://github.com/yourusername/markgo/blob/main/docs/themes.md)
- [Contributing](https://github.com/yourusername/markgo/blob/main/CONTRIBUTING.md)

---

Happy blogging with MarkGo! 🎉
