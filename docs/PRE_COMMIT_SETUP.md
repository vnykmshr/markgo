# Pre-commit Hooks Setup Guide

MarkGo Engine uses comprehensive pre-commit hooks to ensure code quality, security, and consistency across all commits. This guide explains how to set up and use the pre-commit framework.

## 🎯 What We Check

Our pre-commit configuration ensures:

- **Code Quality**: Go fmt, imports, vet, linting, and cyclomatic complexity
- **Security**: Static security analysis with gosec
- **Testing**: Unit tests with race detection and coverage requirements (≥80%)
- **Build Verification**: All binaries build successfully
- **Documentation**: Markdown linting and consistency
- **Version Consistency**: All version strings match canonical version
- **Commit Standards**: Conventional commit message format
- **File Standards**: No large files, proper line endings, no merge conflicts

## 📋 Prerequisites

### Required Tools

```bash
# Install pre-commit framework
pip install pre-commit

# Install Go tools (if not already installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Install markdown linter (optional, for documentation checks)
npm install -g markdownlint-cli
```

### System Requirements

- **Go**: 1.21+ (matches project requirements)
- **Python**: 3.8+ (for pre-commit framework)
- **Node.js**: 16+ (for markdown linting, optional)

## 🚀 Installation

### 1. Install Pre-commit Hooks

```bash
# Navigate to project root
cd /path/to/markgo

# Install the pre-commit hooks
pre-commit install

# Install commit-msg hook for conventional commits
pre-commit install --hook-type commit-msg

# Optional: Install for all Git hooks
pre-commit install --install-hooks
```

### 2. Verify Installation

```bash
# Check pre-commit installation
pre-commit --version

# Test configuration
pre-commit run --all-files
```

## 🔧 Configuration Files

### `.pre-commit-config.yaml`
Main configuration defining all hooks and their settings.

### `.golangci.yml`
Comprehensive Go linting configuration with production-grade rules.

### `.markdownlint.yaml`
Markdown documentation linting rules.

### `scripts/check-version-consistency.sh`
Custom script ensuring version consistency across the codebase.

## 🎮 Usage

### Automatic Execution

Pre-commit hooks run automatically on every `git commit`. If any check fails, the commit is blocked.

### Manual Execution

```bash
# Run all hooks on all files
pre-commit run --all-files

# Run specific hook
pre-commit run golangci-lint
pre-commit run go-fmt
pre-commit run gosec

# Run hooks on specific files
pre-commit run --files internal/handlers/api.go

# Skip hooks for emergency commits (not recommended)
git commit --no-verify -m "emergency fix"
```

### Updating Hooks

```bash
# Update hook repositories to latest versions
pre-commit autoupdate

# Clean and reinstall hooks
pre-commit clean
pre-commit install
```

## 🛡️ Hook Details

### Go Code Quality

| Hook | Purpose | Failure Action |
|------|---------|---------------|
| `go-fmt` | Code formatting | Auto-fixes formatting issues |
| `go-imports` | Import organization | Auto-fixes import grouping |
| `go-mod-tidy` | Dependency cleanup | Updates go.mod/go.sum |
| `go-vet-mod` | Static analysis | Reports potential issues |
| `golangci-lint` | Comprehensive linting | Reports code quality issues |
| `go-cyclo` | Complexity check | Fails if functions >15 complexity |

### Security & Testing

| Hook | Purpose | Failure Action |
|------|---------|---------------|
| `gosec` | Security audit | Reports security vulnerabilities |
| `go-unit-tests-mod` | Test execution | Fails if tests don't pass |
| `test-coverage` | Coverage check | Fails if coverage <80% |
| `build-all` | Build verification | Fails if build breaks |

### Code Standards

| Hook | Purpose | Failure Action |
|------|---------|---------------|
| `no-go-debugging` | Debug statement check | Fails if debugging code found |
| `no-todos-in-prod` | TODO/FIXME check | Fails if TODO comments found |
| `version-consistency` | Version string check | Fails if versions inconsistent |
| `conventional-pre-commit` | Commit message format | Fails if non-conventional format |

### File Standards

| Hook | Purpose | Failure Action |
|------|---------|---------------|
| `trailing-whitespace` | Whitespace cleanup | Auto-removes trailing spaces |
| `end-of-file-fixer` | EOF normalization | Auto-adds final newline |
| `check-large-files` | Size limit check | Fails if files >1MB |
| `detect-private-key` | Security scan | Fails if private keys detected |
| `markdownlint` | Documentation lint | Reports markdown issues |

## 🚨 Troubleshooting

### Common Issues

#### Hook Installation Fails
```bash
# Clear cache and reinstall
pre-commit clean
pre-commit install --install-hooks
```

#### Go Tools Not Found
```bash
# Ensure Go tools are in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Reinstall missing tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

#### Test Coverage Fails
```bash
# Check current coverage
make coverage
go tool cover -func=coverage.out | grep total

# Add tests to improve coverage
```

#### Version Consistency Fails
```bash
# Run version check manually
./scripts/check-version-consistency.sh

# Update inconsistent versions to match internal/constants/constants.go
```

### Bypassing Hooks (Emergency Only)

```bash
# Skip all pre-commit hooks (NOT RECOMMENDED)
git commit --no-verify -m "emergency: critical production fix"

# Skip specific hooks
SKIP=gosec,golangci-lint git commit -m "fix: urgent patch"
```

## 🔄 CI/CD Integration

The pre-commit configuration includes CI settings for automated checks:

```yaml
ci:
  autofix_commit_msg: "[pre-commit.ci] auto fixes"
  autofix_prs: true
  autoupdate_schedule: weekly
  skip: [golangci-lint, build-all, test-coverage]  # Resource-intensive checks
```

### GitHub Actions Integration

Add to `.github/workflows/pre-commit.yml`:

```yaml
name: Pre-commit
on: [push, pull_request]
jobs:
  pre-commit:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-python@v4
      with:
        python-version: 3.x
    - uses: actions/setup-go@v4
      with:
        go-version: 1.21
    - uses: pre-commit/action@v3.0.0
```

## 📈 Metrics & Monitoring

### Coverage Reports
```bash
# Generate HTML coverage report
make coverage
go tool cover -html=coverage.out -o coverage.html
```

### Linting Reports
```bash
# Generate detailed linting report
golangci-lint run --out-format=html > lint-report.html
```

### Security Reports
```bash
# Generate security audit report
gosec -fmt=json -out=security-report.json ./...
```

## ⚙️ Customization

### Project-Specific Rules

Edit `.golangci.yml` to adjust linting rules:

```yaml
linters-settings:
  gocyclo:
    min-complexity: 20  # Increase complexity threshold

  lll:
    line-length: 140    # Increase line length limit
```

### Adding Custom Hooks

Add to `.pre-commit-config.yaml`:

```yaml
- repo: local
  hooks:
    - id: custom-check
      name: Custom Project Check
      entry: ./scripts/custom-check.sh
      language: script
      files: \.go$
```

## 🎓 Best Practices

1. **Run hooks locally** before pushing to catch issues early
2. **Keep hooks fast** to maintain developer productivity
3. **Document exceptions** when bypassing hooks is necessary
4. **Regular updates** to keep security and quality checks current
5. **Team training** ensure all developers understand the setup

## 📞 Support

- **Configuration Issues**: Check hook logs with `pre-commit run -v`
- **Performance Issues**: Consider hook parallelization settings
- **Custom Requirements**: Modify configurations in project's `.pre-commit-config.yaml`

---

**Remember**: Pre-commit hooks are your first line of defense for code quality. They catch issues before they reach the repository, saving time and maintaining high standards across the entire development team.