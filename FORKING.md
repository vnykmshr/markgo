# Forking MarkGo

This guide helps you fork MarkGo and customize it for your own use.

## Quick Start

### 1. Fork the Repository

1. Click "Fork" on the GitHub repository page
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/markgo.git
   cd markgo
   ```

### 2. Update Import Paths

MarkGo uses `github.com/vnykmshr/markgo` in all import statements. After forking, you need to update these to point to your repository.

**Easy way (using our script):**
```bash
# Update to your GitHub username
make update-imports USERNAME=yourusername

# Or specify full module path
make update-imports USERNAME=github.com/yourcompany/markgo

# Or run script directly
./scripts/update-imports.sh yourusername
```

**Manual way:**
```bash
# Replace all import paths
find . -name "*.go" -exec sed -i 's|github.com/vnykmshr/markgo|github.com/yourusername/markgo|g' {} \;

# Update go.mod
sed -i 's|github.com/vnykmshr/markgo|github.com/yourusername/markgo|g' go.mod

# Update dependencies
go mod tidy
```

### 3. Verify Everything Works

```bash
# Test compilation
go build ./...

# Run tests
make test

# Start the server
make run
```

## Why Update Import Paths?

Go modules use the import path to identify packages. When you fork a repository, the import paths need to reflect your repository URL so that:

1. **Go modules work correctly** - Dependencies resolve to your fork
2. **Code compiles without errors** - No import resolution issues  
3. **Your changes are isolated** - No conflicts with the original repository
4. **Publishing works** - Others can import your modified version

## What Gets Updated?

The update script changes:

- ✅ **Go import statements** in all `.go` files
- ✅ **go.mod module declaration**
- ✅ **Documentation references** (README, docs)
- ✅ **Script and configuration files**
- ✅ **Docker and deployment configs**

## Import Path Examples

| Input | Resulting Module Path |
|-------|----------------------|
| `johnsmith` | `github.com/johnsmith/markgo` |
| `acme-corp` | `github.com/acme-corp/markgo` |
| `github.com/myorg/blog` | `github.com/myorg/blog` |
| `gitlab.com/team/markgo` | `gitlab.com/team/markgo` |
| `internal.company.com/tools/blog` | `internal.company.com/tools/blog` |

## Customization Ideas

After updating import paths, you can:

### 🎨 **Branding**
- Update `internal/config/config.go` with your blog details
- Modify templates in `web/templates/`
- Replace favicon and logos in `web/static/`

### 🔧 **Features**
- Add custom middleware in `internal/middleware/`
- Extend article metadata in `internal/models/`
- Add new API endpoints in `internal/handlers/`

### 📊 **Analytics**
- Integrate your preferred analytics service
- Add custom metrics collection
- Implement A/B testing

### 🎯 **SEO**
- Customize meta tags and structured data
- Add your Google Search Console integration
- Implement custom sitemap logic

## Best Practices

### 1. Keep Your Fork Updated
```bash
# Add upstream remote
git remote add upstream https://github.com/vnykmshr/markgo.git

# Fetch upstream changes
git fetch upstream

# Merge updates (resolve conflicts as needed)
git merge upstream/main
```

### 2. Preserve Attribution
- Keep the LICENSE file unchanged (MIT License)
- Credit the original project in your README
- Consider contributing improvements back upstream

### 3. Version Your Changes
```bash
# Tag your releases
git tag -a v1.0.0 -m "My custom MarkGo v1.0.0"
git push origin v1.0.0
```

### 4. Document Your Modifications
- Update README.md with your changes
- Document new configuration options
- Add changelog for your modifications

## Troubleshooting

### Import Path Issues
```bash
# If you see import errors:
go mod tidy
go clean -modcache
go build ./...
```

### Build Failures
```bash
# Check for missed import paths
grep -r "github.com/vnykmshr/markgo" . --include="*.go"

# Verify go.mod is correct
head -1 go.mod
```

### Script Issues
```bash
# Run with verbose output
bash -x ./scripts/update-imports.sh yourusername

# Check script permissions
ls -la scripts/update-imports.sh
```

## Advanced Scenarios

### Custom Domain/Private Repo
```bash
# For private repositories or custom domains
./scripts/update-imports.sh company.internal/blog-engine

# Then configure private module access
go env -w GOPRIVATE="company.internal/*"
```

### Multiple Forks
```bash
# Different import path for each use case
./scripts/update-imports.sh github.com/myorg/markgo-blog
./scripts/update-imports.sh github.com/myorg/markgo-docs
```

### Reverting Changes
```bash
# The script creates backups
ls backup-*/

# Restore from backup if needed
cp -r backup-20241204-143022/* ./
```

## Support

### Getting Help
1. **Check the Issues** - Search existing GitHub issues
2. **Documentation** - Review the main README and code comments
3. **Community** - Join discussions in GitHub Discussions
4. **Contribute** - Submit PRs for improvements

### Common Questions

**Q: Do I need to update import paths for every fork?**
A: Yes, this ensures your fork works independently and can be properly versioned.

**Q: Can I use a different repository name?**
A: Yes! Use: `./scripts/update-imports.sh github.com/yourusername/your-blog-name`

**Q: What if I want to contribute back to the original?**
A: Create a separate branch without import path changes for upstream contributions.

**Q: Does this affect performance?**
A: No, import paths are resolved at compile time and don't affect runtime performance.

---

🎉 **Happy forking!** If you build something cool with MarkGo, let us know - we'd love to see what you create!