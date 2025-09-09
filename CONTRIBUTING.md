# Contributing to MarkGo Engine

Thank you for your interest in contributing to MarkGo! This document provides guidelines and information for contributors.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- **Be respectful** and inclusive in all interactions
- **Be constructive** when giving feedback
- **Focus on the issue**, not the person
- **Help create** a welcoming environment for all contributors

## How to Contribute

### Reporting Bugs

Before creating a bug report, please:

1. **Check existing issues** to avoid duplicates
2. **Use the latest version** to confirm the bug still exists
3. **Provide detailed information** including:
   - Go version (`go version`)
   - Operating system and version
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs or error messages

Use this template for bug reports:

```markdown
**Bug Description**
A clear description of the bug.

**Steps to Reproduce**
1. Go to '...'
2. Click on '...'
3. See error

**Expected Behavior**
What you expected to happen.

**Screenshots/Logs**
If applicable, add screenshots or log output.

**Environment**
- OS: [e.g., macOS 14.0]
- Go Version: [e.g., go1.24.4]
- MarkGo Version: [e.g., v1.0.0]
```

### Suggesting Features

We welcome feature suggestions! Please:

1. **Check existing issues** for similar requests
2. **Describe the use case** - why is this feature needed?
3. **Explain the proposed solution** - how should it work?
4. **Consider alternatives** - are there other ways to achieve this?

### Pull Requests

#### Development Setup

1. **Fork the repository** and clone your fork:
   ```bash
   git clone https://github.com/vnykmshr/markgo
   cd markgo
   ```

2. **Install dependencies and tools**:
   ```bash
   make deps
   make install-dev-tools
   ```

3. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

#### Development Workflow

1. **Write your code** following our coding standards
2. **Add tests** for your changes
3. **Update documentation** if needed
4. **Run quality checks**:
   ```bash
   make check  # Runs fmt, vet, lint, and tests
   ```

#### Coding Standards

**Go Code Style**:
- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions focused and small
- Handle errors properly

**Testing Requirements**:
- Add unit tests for new functionality
- Maintain or improve code coverage (aim for 80%+)
- Use table-driven tests for multiple scenarios
- Mock external dependencies appropriately

**Git Commit Messages**:
```
type(scope): brief description

Longer description if needed, explaining what and why.

Fixes #issue_number
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
- `feat(search): add fuzzy search algorithm`
- `fix(handlers): prevent panic on malformed requests`
- `docs(readme): update installation instructions`

#### Pull Request Process

1. **Ensure CI passes** - all tests and checks must pass
2. **Update documentation** - README, docs, comments
3. **Describe your changes** - use the PR template
4. **Reference related issues** - use "Fixes #123" or "Relates to #456"
5. **Request review** - we'll review and provide feedback

### Documentation

Documentation improvements are always welcome:

- **Fix typos** or unclear explanations
- **Add examples** for complex features
- **Improve getting started** guides
- **Create tutorials** for common use cases

Documentation files:
- `README.md` - Main project documentation
- `docs/*.md` - Detailed guides and references
- Code comments - Explain complex logic

### Code Review Guidelines

**For Contributors**:
- Be open to feedback and suggestions
- Respond to comments promptly
- Make requested changes or explain why you disagree
- Keep discussions focused on the code

**For Reviewers**:
- Be constructive and respectful
- Explain the "why" behind suggestions
- Acknowledge good practices
- Focus on code quality, not personal style

## Development Guide

### Project Structure

```
markgo/
├── cmd/                # Application entry points
├── internal/           # Private packages
│   ├── config/        # Configuration
│   ├── handlers/      # HTTP handlers
│   ├── middleware/    # HTTP middleware
│   ├── models/        # Data structures
│   ├── services/      # Business logic
│   └── utils/         # Utilities
├── web/               # Frontend assets
├── docs/              # Documentation
└── deployments/       # Deployment configs
```

### Running Tests

```bash
# Run all tests
make test

# Run with race detection
make test-race

# Generate coverage report
make coverage

# Run benchmarks
make benchmark
```

### Local Development

```bash
# Start development server with hot reload
make dev

# Build for current platform
make build

# Build for all platforms
make build-all

# Run linting
make lint

# Format code
make fmt
```

### Adding New Features

1. **Design first** - consider the API and user experience
2. **Start with tests** - write failing tests for the feature
3. **Implement incrementally** - small, focused commits
4. **Document as you go** - update docs and examples

### Common Tasks

#### Adding a New HTTP Handler

1. Create handler function in `internal/handlers/`
2. Add route in handler setup
3. Add tests in `*_test.go` files
4. Update API documentation

#### Adding Configuration Options

1. Add field to config struct in `internal/config/`
2. Update `.env.example` with new option
3. Add validation if needed
4. Document in configuration guide

#### Adding a New Service

1. Define interface in `internal/services/interfaces.go`
2. Implement service in `internal/services/`
3. Add comprehensive tests
4. Update dependency injection in main

## Release Process

Releases are handled by maintainers:

1. **Version bump** following semantic versioning
2. **Update CHANGELOG.md** with new features and fixes
3. **Create release tag** with release notes
4. **Build and publish** release artifacts

## Getting Help

- **Discussions** - Use GitHub Discussions for questions
- **Issues** - Report bugs or request features
- **Code Review** - Ask for feedback on PRs

## Recognition

Contributors will be recognized in:
- Release notes for their contributions
- GitHub contributors list
- Special thanks for significant contributions

## License

By contributing to MarkGo, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to MarkGo! Your help makes this project better for everyone. 🎉