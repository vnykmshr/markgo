# MarkGo Git Hooks

This directory contains git hooks to ensure code quality before commits reach CI.

## Setup

Run the setup script after cloning:

```bash
./.githooks/setup.sh
```

Or configure Git to use this directory automatically:

```bash
git config core.hooksPath .githooks
```

## Available Hooks

### pre-commit

Runs before every commit to catch issues early:

1. **🔐 Secret Detection**
   - Scans for potential passwords, tokens, secrets
   - Prevents accidental credential commits

2. **✨ Code Formatting (gofmt)**
   - Ensures all Go code is properly formatted
   - **Prevents CI formatting failures**
   - Auto-fix suggestion provided

3. **🔬 Go Vet**
   - Static analysis to catch common mistakes
   - Type errors, suspicious constructs

4. **🧪 Tests**
   - Runs `go test ./... -short`
   - Only runs if Go files changed
   - Catches test failures before CI

## Bypassing Hooks

**Not recommended**, but if needed:

```bash
git commit --no-verify
```

## Why This Matters

**Before hooks**:
- ❌ Formatting issues caught in CI after push
- ❌ Failed CI builds on release commits
- ❌ Extra commits to fix simple issues
- ❌ Annoying red builds on main branch

**With hooks**:
- ✅ Issues caught locally before commit
- ✅ Clean CI builds every time
- ✅ No formatting fix commits
- ✅ Professional commit history

## Example Output

**Successful commit**:
```
🔍 Running pre-commit checks...
🔐 Checking for secrets...
✨ Checking code formatting...
🔬 Running go vet...
🧪 Running tests...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ All pre-commit checks passed!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Failed commit** (formatting):
```
🔍 Running pre-commit checks...
🔐 Checking for secrets...
✨ Checking code formatting...
✗ Code formatting issues detected:
internal/handlers/article_test.go

Run: gofmt -w .

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✗ Pre-commit checks FAILED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Troubleshooting

### Hook not running

```bash
# Check if hook is executable
ls -l .git/hooks/pre-commit

# Make it executable
chmod +x .git/hooks/pre-commit

# Or re-run setup
./.githooks/setup.sh
```

### Hook too slow

The hook runs tests which can be slow. Options:

1. **Use `-short` flag** (already configured)
2. **Disable test step** - Edit `.githooks/pre-commit`, comment out test section
3. **Use `--no-verify`** - Only for urgent fixes (not recommended)

### False positives

If the hook blocks a legitimate commit:

1. Check the error message for details
2. Fix the actual issue (preferred)
3. Use `--no-verify` if truly necessary (rare)

## Maintenance

### Updating hooks

1. Edit files in `.githooks/`
2. Run `./.githooks/setup.sh` to reinstall
3. Or if using `core.hooksPath`, changes apply immediately

### Adding new hooks

1. Create new hook file in `.githooks/`
2. Make it executable: `chmod +x .githooks/hook-name`
3. Update `setup.sh` to install it
4. Document it in this README

## CI Integration

These hooks mirror the CI checks, ensuring:

- Local environment = CI environment
- No surprises in CI
- Fast feedback loop

If the hooks pass locally, CI will pass too.

## Resources

- [Git Hooks Documentation](https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
