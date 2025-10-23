# MarkGo Documentation

Complete documentation for MarkGo Engine - a modern, high-performance file-based blog platform built with Go.

## 📖 Getting Started

**New to MarkGo?** Start here:

- [Quick Start](../README.md#quick-start) - Get running in 5 minutes
- [Getting Started Guide](GETTING-STARTED.md) - Detailed setup walkthrough
- [Project Guide](project-guide.md) - Complete feature overview

## ⚙️ Configuration & Deployment

- [Configuration Guide](configuration.md) - All environment variables and options
- [Deployment Guide](deployment.md) - Production deployment strategies
- [Static Export](static-export.md) - GitHub Pages, Netlify, Vercel deployment

## 🏗️ Architecture & Development

- [System Overview](system-overview.md) - Performance characteristics and design
- [Architecture Guide](architecture.md) - Technical architecture details
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute

## 📚 Reference Documentation

- [API Documentation](API.md) - HTTP endpoints and responses
- [Quick Reference](QUICK-REFERENCE.md) - Commands and common tasks

## 🔐 Security

- [Security Guide](../SECURITY.md) - Security best practices and configuration

## 📦 Package Documentation

Code-level documentation:

| Package | Description | Documentation |
|---------|-------------|---------------|
| **Config** | Configuration management and validation | [config-package.md](./config-package.md) |
| **Handlers** | HTTP request handlers and routing | [handlers-package.md](./handlers-package.md) |
| **Middleware** | HTTP middleware pipeline components | [middleware-package.md](./middleware-package.md) |
| **Services** | Business logic and service layer | [services-package.md](./services-package.md) |
| **Models** | Core data structures | [models-package.md](./models-package.md) |
| **Errors** | Domain-specific error handling | [errors-package.md](./errors-package.md) |

## 🚀 CLI Commands

MarkGo uses a unified CLI with subcommands:

| Command | Purpose | Documentation |
|---------|---------|---------------|
| **markgo serve** | Start the web server | Main server command |
| **markgo init** | Initialize a new blog | Quick setup wizard |
| **markgo new** | Create a new article | Interactive article creation |
| **markgo export** | Export to static site | GitHub Pages, Netlify, Vercel |
| **Stress Test** | Performance testing tool | [stress-test-package.md](./stress-test-package.md) |

## 📝 Release Information

- [Changelog](../CHANGELOG.md) - Version history and changes

---

**Looking for something specific?** Use your browser's search (Ctrl+F / Cmd+F) or check the main [README](../README.md).
