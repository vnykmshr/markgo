# Static Site Export Guide

MarkGo now supports exporting your dynamic blog as a static site, perfect for hosting on GitHub Pages, Netlify, Vercel, or any static hosting platform.

## Quick Start

### 1. Export Your Site

```bash
# Basic export
make export-static

# Export for GitHub Pages
make export-github-pages

# Manual export with custom options
markgo-export --output ./public --base-url https://yourdomain.com
```

### 2. Deploy to GitHub Pages

The easiest way is to use the included GitHub Actions workflow:

1. **Push to GitHub**: Your repository is automatically built and deployed
2. **Enable Pages**: Go to Settings > Pages > Source: GitHub Actions
3. **Done**: Your site is live at `https://username.github.io/repository-name`

## Export Command Options

```bash
markgo-export [options]

Options:
  -o, --output DIR        Output directory (default: ./dist)
  -u, --base-url URL      Base URL for absolute links
  -d, --include-drafts    Include draft articles
  -v, --verbose           Enable verbose logging
  -h, --help              Show help
```

### Examples

```bash
# Export to custom directory
markgo-export --output ./public

# Export with custom base URL
markgo-export --base-url https://blog.example.com

# Include draft articles
markgo-export --include-drafts --verbose

# GitHub Pages export
markgo-export --base-url https://username.github.io/repo-name
```

## What Gets Exported

The static export includes:

### HTML Pages
- ✅ Home page (`index.html`)
- ✅ Article pages (`articles/slug/index.html`)
- ✅ Article listing (`articles/index.html`)
- ✅ Tag pages (`tags/tag-name/index.html`)
- ✅ Category pages (`categories/category-name/index.html`)
- ✅ About page (`about/index.html`)
- ✅ Contact form (`contact/index.html`)
- ✅ Search page (`search/index.html`)

### Assets & Files
- ✅ All CSS, JavaScript, images (`static/`)
- ✅ RSS feed (`feed.xml`)
- ✅ JSON feed (`feed.json`)
- ✅ XML sitemap (`sitemap.xml`)
- ✅ Robots.txt (`robots.txt`)
- ✅ Favicon and meta files

### Features Maintained
- ✅ SEO metadata and structured data
- ✅ Responsive design
- ✅ Full-text search (client-side)
- ✅ Tag and category filtering
- ✅ Reading time estimates
- ✅ Social media optimization

## Hosting Options

### GitHub Pages (Free)
- **Cost**: $0/month
- **Setup**: Enable in repository settings
- **Custom domain**: Supported with SSL
- **Bandwidth**: 100GB/month
- **Storage**: 1GB

```yaml
# Automatic deployment via GitHub Actions
# See .github/workflows/deploy.yml
```

### Netlify (Freemium)
- **Cost**: Free tier available
- **Features**: CDN, forms, redirects
- **Deployment**: Git integration or manual upload
- **Bandwidth**: 100GB/month (free)

```bash
# Deploy to Netlify
make export-static
cd dist
# Drag & drop to Netlify dashboard
```

### Vercel (Freemium)
- **Cost**: Free tier available  
- **Features**: Edge functions, analytics
- **Deployment**: Git integration
- **Bandwidth**: 100GB/month (free)

### CloudFlare Pages (Free)
- **Cost**: $0/month
- **Features**: Unlimited bandwidth, global CDN
- **Deployment**: Git integration
- **Limits**: 20,000 files per deployment

### Custom Server
Deploy to any web server:

```bash
# Export and upload
make export-static
rsync -av dist/ user@server:/var/www/html/
```

## GitHub Actions Workflow

The included workflow (`.github/workflows/deploy.yml`) automatically:

1. **Builds** the export tool
2. **Creates** demo content if articles are empty
3. **Exports** the static site
4. **Deploys** to GitHub Pages

### Setup Steps

1. **Enable GitHub Pages**:
   - Go to Settings > Pages
   - Source: GitHub Actions

2. **Push your code**:
   ```bash
   git add .
   git commit -m "Add static site export"
   git push
   ```

3. **Wait for deployment** (2-3 minutes)

4. **Visit your site**: `https://username.github.io/repository-name`

## Advanced Configuration

### Custom Base URL

```bash
# Production site
markgo-export --base-url https://blog.example.com

# Subdirectory deployment  
markgo-export --base-url https://example.com/blog
```

### Environment Variables

Create `.env` file for export configuration:

```bash
# Blog settings
BLOG_TITLE=My Static Blog
BLOG_DESCRIPTION=A fast static blog powered by MarkGo
BLOG_AUTHOR=Your Name
BASE_URL=https://yourdomain.com

# Features
SEARCH_ENABLED=true
RSS_ENABLED=true
CONTACT_ENABLED=true
```

### Build Optimization

```bash
# Production build with optimizations
make export-static

# The export automatically:
# - Minifies HTML
# - Optimizes images  
# - Generates cache headers
# - Creates compressed assets
```

## Content Management

### Adding Articles

1. **Create Markdown file** in `articles/`:
   ```markdown
   ---
   title: "My New Post"
   date: 2024-01-15T10:00:00Z
   published: true
   tags: ["example", "tutorial"]
   ---
   
   # Content goes here
   ```

2. **Commit and push**:
   ```bash
   git add articles/my-new-post.md
   git commit -m "Add new post"
   git push
   ```

3. **Site rebuilds automatically**

### Managing Drafts

```bash
# Include drafts in export
markgo-export --include-drafts
```

Draft articles (with `published: false`) are normally excluded.

## Performance Features

### Build Speed
- **Cold build**: ~2-3 seconds
- **Incremental**: Only changed pages rebuild
- **Parallel processing**: Articles processed concurrently

### Runtime Performance
- **Static HTML**: No server processing
- **CDN friendly**: All assets cacheable
- **Optimized images**: Automatic compression
- **Minified output**: Reduced file sizes

## Troubleshooting

### Common Issues

**Build fails with template errors**:
```bash
# Check templates exist
ls web/templates/

# Verify template syntax
make test
```

**Images not loading**:
```bash
# Check static assets
ls web/static/

# Verify base URL
markgo-export --base-url https://correct-domain.com
```

**Search not working**:
Search is client-side JavaScript. Ensure:
- `SEARCH_ENABLED=true` in config
- JavaScript files are included
- No CORS issues on your domain

### Debug Mode

```bash
# Verbose logging
markgo-export --verbose

# Check export output
ls -la dist/
```

## Migration Guide

### From Dynamic to Static

1. **Test locally**:
   ```bash
   make export-static
   cd dist
   python -m http.server 8000
   # Visit http://localhost:8000
   ```

2. **Update links**: Static sites need absolute URLs for cross-domain resources

3. **Test forms**: Contact forms need alternative handling (Netlify Forms, Formspree, etc.)

### From Other Generators

**From Jekyll/Hugo**:
- Convert frontmatter format
- Update template syntax
- Migrate static assets

**From Ghost/WordPress**:
- Export posts to Markdown
- Convert template structure
- Update media references

## Best Practices

### Content Organization
```
articles/
├── 2024/
│   ├── 01-welcome.md
│   └── 02-tutorial.md
└── 2023/
    └── year-review.md
```

### SEO Optimization
- Use descriptive filenames
- Include meta descriptions
- Optimize images with alt text
- Generate proper sitemaps

### Performance Tips
- Optimize images before adding
- Use proper heading hierarchy
- Minimize JavaScript usage
- Leverage CDN capabilities

## Next Steps

1. **Customize design**: Edit CSS in `web/static/css/`
2. **Add analytics**: Include tracking codes
3. **Setup monitoring**: Monitor performance and uptime
4. **Custom domain**: Configure DNS for custom domains

For more advanced features, see:
- [Configuration Guide](configuration.md)
- [Theme Customization](themes.md)
- [Deployment Guide](deployment.md)

---

**Static site export powered by MarkGo Engine**