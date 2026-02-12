# Contributing to MarkGo Engine

Thank you for your interest in contributing to MarkGo! This document provides guidelines and information for contributors.

## Code of Conduct

Read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## First Time Contributing?

Look for issues labeled [`good first issue`](https://github.com/vnykmshr/markgo/labels/good%20first%20issue).

**Response Times**: Expect PR feedback within 3-5 days. Ping the maintainer if no response after 1 week.

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
- OS: [e.g., macOS 15.0]
- Go Version: [e.g., go1.25.0]
- MarkGo Version: [e.g., v2.3.0]
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
   make tidy
   ```

3. **Set up pre-commit hooks** (recommended):
   ```bash
   ./.githooks/setup.sh
   ```

   This installs hooks that run before each commit:
   - Secret detection (prevents credential leaks)
   - Code formatting check (gofmt)
   - Go vet analysis
   - Tests execution

   **Prevents CI failures** by catching issues locally first.

4. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

#### Development Workflow

1. **Write your code** following our coding standards
2. **Add tests** for your changes
3. **Update documentation** if needed
4. **Run quality checks**:
   ```bash
   make fmt && make lint && make test  # Format, lint, and test
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
- Maintain or improve code coverage (current baseline: ~52%)
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
│   ├── handlers/      # 11 handler types
│   ├── middleware/    # HTTP middleware
│   ├── models/        # Data structures
│   ├── services/      # Business logic
│   ├── constants/     # Build-time metadata
│   └── errors/        # Typed error system
├── web/               # Frontend assets
├── docs/              # Documentation
└── deployments/       # Deployment configs
```

### Local Development

```bash
make dev      # Start development server with hot reload
make test     # Run all tests
make lint     # Run golangci-lint
```

See [README.md](README.md#development) for complete build and test commands.

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