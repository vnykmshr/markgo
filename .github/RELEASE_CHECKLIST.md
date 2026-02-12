# Release Checklist for MarkGo

This checklist ensures all critical validations are performed before creating a release. Use this for all production releases to prevent regressions.

**Target:** All maintainers creating releases
**When to use:** Before tagging any version (vX.Y.Z)

---

## Pre-Release Checklist

### 1. Code Quality ✅

- [ ] All CI checks passing on main branch
- [ ] No failing tests: `make test`
- [ ] No lint errors: `make lint`
- [ ] Code formatted: `make fmt` (should show no changes)
- [ ] No security vulnerabilities: `go list -json -m all | nancy sleuth` or wait for CI govulncheck
- [ ] Go module dependencies tidy: `go mod tidy` (should show no changes)

### 2. Build Validation ✅

- [ ] **Local build succeeds:**
  ```bash
  make clean
  make build
  ./build/markgo --version
  ```

- [ ] **Docker build succeeds:**
  ```bash
  docker build -t markgo:test .
  docker run --rm markgo:test --version
  ```

- [ ] **Cross-platform builds succeed:**
  ```bash
  make build-release
  # Verify binaries created for Linux, macOS, Windows
  ls -lh build/
  ```

### 3. Docker Verification 🐳

- [ ] **Dockerfile uses correct build path:**
  ```bash
  grep "cmd/markgo" Dockerfile
  # Should show: ./cmd/markgo (NOT ./cmd/server)
  ```

- [ ] **Docker Compose starts successfully:**
  ```bash
  docker compose up -d
  docker compose ps  # Should show "Up"
  docker compose logs markgo | tail -20
  docker compose down
  ```

- [ ] **Docker healthcheck passes:**
  ```bash
  docker compose up -d
  sleep 30
  docker inspect --format='{{.State.Health.Status}}' markgo-app
  # Should show: healthy
  docker compose down
  ```

### 4. Functional Testing ✅

- [ ] **Server starts and serves requests:**
  ```bash
  ./build/markgo serve &
  SERVER_PID=$!
  sleep 2
  curl http://localhost:3000/health
  # Should return 200 OK
  kill $SERVER_PID
  ```

- [ ] **CLI commands work:**
  ```bash
  ./build/markgo init --quick --output /tmp/markgo-test
  ./build/markgo new --title "Test" --output /tmp/markgo-test/articles
  ```

### 5. GitHub Actions Workflow Validation 🔄

- [ ] **CI workflow references correct paths:**
  ```bash
  grep "cmd/markgo" .github/workflows/ci.yml
  # Should find: go build ... ./cmd/markgo
  ```

- [ ] **Workflow builds use correct ldflags:**
  ```bash
  grep "github.com/vnykmshr/markgo/internal/commands/serve" .github/workflows/ci.yml
  # Should find version injection for serve.Version
  ```

### 6. Documentation Accuracy 📚

- [ ] **README.md metrics are current:**
  ```bash
  # Check binary size
  ls -lh build/markgo-linux-amd64
  # Compare with README.md (~27MB)

  # Check version references
  grep -E "v[0-9]+\.[0-9]+\.[0-9]+" README.md
  ```

- [ ] **CHANGELOG.md updated:**
  - [ ] New version section added
  - [ ] All major changes documented
  - [ ] Breaking changes clearly marked
  - [ ] Contributors credited (if applicable)

- [ ] **Version references updated:**
  ```bash
  # Update these if version changed:
  # - README.md badges
  # - docs/GETTING-STARTED.md examples
  # - docker compose.yml image tags (if publishing)
  ```

### 7. Configuration & Environment 🔧

- [ ] **.env.example up to date:**
  ```bash
  # Verify all config options documented
  diff <(grep "^[A-Z_]*=" .env.example | cut -d= -f1 | sort) \
       <(grep "Get.*Env" internal/config/config.go | grep -oE '[A-Z_]+' | sort)
  ```

- [ ] **Deployment configs validated:**
  - [ ] docker compose.yml environment variables match .env.example
  - [ ] Dockerfile.dev (if exists) uses correct paths
  - [ ] systemd service file (if exists) references correct binary

### 8. Security Check 🔒

- [ ] **No secrets committed:**
  ```bash
  grep -r "password\|secret\|token\|key" .env 2>/dev/null
  # Should fail or return nothing (file shouldn't exist in repo)
  ```

- [ ] **Sensitive files gitignored:**
  ```bash
  grep "\.env$" .gitignore
  grep "\.env\.local" .gitignore
  ```

- [ ] **Dependencies scanned:**
  ```bash
  # Wait for CI govulncheck or run locally:
  govulncheck ./...
  ```

### 9. Version Tagging 🏷️

- [ ] **Choose appropriate version number** (SemVer):
  - MAJOR: Breaking changes
  - MINOR: New features (backward compatible)
  - PATCH: Bug fixes

- [ ] **Verify version constants:**
  ```bash
  grep "AppVersion" internal/constants/constants.go
  # Should match intended release version
  ```

- [ ] **Git working tree is clean:**
  ```bash
  git status
  # Should show: nothing to commit, working tree clean
  ```

### 10. Create Release 🚀

- [ ] **Create annotated tag:**
  ```bash
  git tag -a v2.X.Y -m "Release vX.Y.Z - Brief description

  Major changes:
  - Feature 1
  - Feature 2
  - Bug fix 1
  "
  ```

- [ ] **Push tag:**
  ```bash
  git push origin v2.X.Y
  ```

- [ ] **Monitor CI pipeline:**
  ```bash
  gh run list --limit 3
  gh run watch  # Watch latest run
  ```

- [ ] **Verify release artifacts built:**
  ```bash
  gh run view --log | grep "Build Artifacts"
  ```

### 11. Post-Release Validation ✅

- [ ] **GitHub release created automatically:**
  ```bash
  gh release view v2.X.Y
  ```

- [ ] **All platform binaries attached:**
  - [ ] markgo-linux-amd64
  - [ ] markgo-darwin-amd64
  - [ ] markgo-windows-amd64.exe

- [ ] **Release notes are complete:**
  ```bash
  gh release view v2.X.Y --web
  # Verify notes include:
  # - Summary of changes
  # - Installation instructions
  # - Breaking changes (if any)
  # - Contributors
  ```

- [ ] **Test downloading release binary:**
  ```bash
  wget https://github.com/vnykmshr/markgo/releases/download/v2.X.Y/markgo-linux-amd64
  chmod +x markgo-linux-amd64
  ./markgo-linux-amd64 --version
  ```

- [ ] **Docker image tagged (if applicable):**
  ```bash
  # If publishing to registry
  docker tag markgo:latest markgo:v2.X.Y
  docker push markgo:v2.X.Y
  ```

### 12. Communication 📢

- [ ] **Update project documentation:**
  - [ ] Update installation instructions if needed
  - [ ] Update migration guide for breaking changes

- [ ] **Announce release:**
  - [ ] GitHub Discussions post
  - [ ] Project blog/website (if applicable)
  - [ ] Social media (if applicable)

- [ ] **Monitor for issues:**
  - [ ] Watch GitHub Issues for 24 hours post-release
  - [ ] Check CI status for main branch
  - [ ] Monitor deployment errors in production

---

## Emergency Rollback

If critical issues are discovered post-release:

1. **Delete the faulty release:**
   ```bash
   gh release delete v2.X.Y
   git push origin :refs/tags/v2.X.Y
   git tag -d v2.X.Y
   ```

2. **Fix the issue**

3. **Create a patch release** (vX.Y.Z+1)

4. **Document the issue** in CHANGELOG.md

---

## Common Pitfalls to Avoid

Based on actual incidents:

1. ✅ **Dockerfile build path** - Always verify `./cmd/markgo` not `./cmd/server`
2. ✅ **Formatting before release** - Run `gofmt -s -w .` to avoid CI failures
3. ✅ **Test Docker build locally** - Don't rely solely on CI
4. ✅ **Update CHANGELOG** - Document changes before tagging
5. ✅ **Verify version constants** - Update internal/constants/constants.go
6. ✅ **Clean git state** - Commit all changes before tagging
7. ✅ **Check workflow paths** - Grep for old paths (cmd/server)
8. ✅ **Validate environment variables** - Ensure .env.example is current
9. ✅ **Monitor CI after push** - Don't assume tag push succeeded

---

## Automation Ideas (Future)

Consider automating these checks:

```bash
#!/bin/bash
# scripts/pre-release-check.sh

echo "Running pre-release checks..."

# Run tests
make test || exit 1

# Check formatting
if [ -n "$(gofmt -l .)" ]; then
    echo "Code not formatted"
    exit 1
fi

# Build for all platforms
make build-release || exit 1

# Test Docker build
docker build -t markgo:test . || exit 1

# Verify Dockerfile path
if ! grep -q "cmd/markgo" Dockerfile; then
    echo "Dockerfile uses wrong path"
    exit 1
fi

echo "All checks passed!"
```

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-23 | Initial checklist based on v2.1.0 release experience |

---

**Maintained by:** @vnykmshr
**Feedback:** https://github.com/vnykmshr/markgo/issues
