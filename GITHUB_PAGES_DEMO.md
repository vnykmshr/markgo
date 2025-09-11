# MarkGo GitHub Pages Demo

This repository demonstrates MarkGo's static site export capability, automatically deploying to GitHub Pages.

## 🚀 Quick Demo Setup

### Option 1: Fork This Repository
1. **Fork** this repository
2. **Enable GitHub Pages** in Settings > Pages > Source: GitHub Actions  
3. **Push a change** to trigger deployment
4. **Visit** your site at `https://yourusername.github.io/markgo`

### Option 2: Use Template
1. **Use this template** to create a new repository
2. **Clone** your new repository
3. **Enable GitHub Pages** in Settings > Pages > Source: GitHub Actions
4. **Push** to deploy

### Option 3: Add to Existing Project
1. **Copy** `.github/workflows/deploy.yml` to your project
2. **Add** static export functionality
3. **Configure** GitHub Pages
4. **Push** to deploy

## 📁 What's Included

### Static Site Generation
- ✅ Full HTML export of all pages
- ✅ CSS, JavaScript, and image assets
- ✅ RSS and JSON feeds
- ✅ XML sitemap for SEO
- ✅ Optimized for static hosting

### Demo Content
The workflow automatically creates demo content if none exists:
- Welcome post explaining MarkGo
- About page with technical details  
- Sample blog post demonstrating features
- Properly structured frontmatter examples

### Automated Deployment
- ✅ Builds on every push to `main`
- ✅ Generates static site with proper URLs
- ✅ Deploys to GitHub Pages automatically
- ✅ Takes 2-3 minutes from push to live

## ⚡ Performance

### Build Speed
- **Cold build**: ~2-3 seconds
- **Deploy time**: ~2-3 minutes total
- **File generation**: Milliseconds per page

### Runtime Performance  
- **Load time**: <100ms (static HTML)
- **First paint**: <200ms
- **SEO score**: 100/100
- **Lighthouse**: 90+ across all metrics

## 🛠️ Customization

### Adding Your Content

1. **Replace demo articles** in `articles/` directory:
   ```bash
   rm articles/*.md
   # Add your own .md files
   ```

2. **Update site config** in `.env`:
   ```bash
   BLOG_TITLE=Your Blog Name
   BLOG_DESCRIPTION=Your description
   BLOG_AUTHOR=Your Name
   BASE_URL=https://yourusername.github.io/repository-name
   ```

3. **Customize design** in `web/` directory:
   ```
   web/
   ├── static/
   │   ├── css/       # Stylesheets
   │   ├── js/        # JavaScript
   │   └── img/       # Images
   └── templates/     # HTML templates
   ```

4. **Commit and push**:
   ```bash
   git add .
   git commit -m "Customize blog content"
   git push
   ```

### Article Format

Create articles with YAML frontmatter:

```markdown
---
title: "Your Article Title"
description: "SEO description"
date: 2024-01-15T10:00:00Z
published: true
tags: ["tutorial", "example"]
categories: ["Technology"]
author: "Your Name"
---

# Your Article Content

Write your content here using Markdown...
```

## 🌐 Hosting Alternatives

While this demo uses GitHub Pages, MarkGo static export works with any hosting:

### Netlify
```bash
make export-static
# Drag dist/ folder to Netlify dashboard
```

### Vercel
```bash
# Connect GitHub repository to Vercel
# Vercel automatically detects build settings
```

### CloudFlare Pages
```bash
# Connect repository to CloudFlare Pages
# Build command: make export-static
# Output directory: dist
```

### Custom Server
```bash
make export-static
rsync -av dist/ user@server:/var/www/html/
```

## 📊 Features Demonstrated

### Content Management
- ✅ Markdown articles with frontmatter
- ✅ Tag and category organization  
- ✅ Draft and published states
- ✅ Automatic URL generation

### SEO & Discovery
- ✅ XML sitemaps for search engines
- ✅ RSS and JSON feeds
- ✅ Meta tags and structured data
- ✅ Proper heading hierarchy

### User Experience  
- ✅ Responsive design
- ✅ Fast page loads
- ✅ Client-side search
- ✅ Tag/category filtering

### Developer Experience
- ✅ Git-based workflow
- ✅ Automatic deployments
- ✅ Local development server
- ✅ Template hot-reload

## 🔧 GitHub Actions Workflow

The deployment workflow (`.github/workflows/deploy.yml`):

1. **Checks out code** on push to main
2. **Sets up Go** environment 
3. **Builds export tool** (`markgo-export`)
4. **Creates demo content** if articles directory is empty
5. **Exports static site** with proper base URL
6. **Deploys to GitHub Pages** using official action

### Customizing the Workflow

```yaml
# Edit .github/workflows/deploy.yml
- name: Export static site
  run: |
    ./markgo-export \
      --output ./dist \
      --base-url "https://${{ github.repository_owner }}.github.io/${{ github.event.repository.name }}" \
      --include-drafts \  # Add this to include drafts
      --verbose
```

## 📈 Monitoring & Analytics

### Performance Monitoring
- Use GitHub Pages analytics
- Add Google Analytics to templates
- Monitor Core Web Vitals

### SEO Tracking
- Submit sitemap to Google Search Console
- Monitor search rankings
- Track RSS feed subscribers

## 🚨 Troubleshooting

### Common Issues

**Site not loading after deployment:**
- Check GitHub Pages settings are enabled
- Verify base URL in workflow matches repository name
- Wait 5-10 minutes for DNS propagation

**Images not displaying:**
- Ensure images are in `web/static/img/`
- Check base URL configuration
- Verify file extensions and paths

**Search not working:**
- Search is client-side JavaScript
- Ensure `SEARCH_ENABLED=true`
- Check browser console for errors

### Debug Steps

1. **Check GitHub Actions logs** for build errors
2. **Verify exported files** in repository's Pages deployment
3. **Test locally** before pushing:
   ```bash
   make export-static
   cd dist
   python -m http.server 8000
   # Visit http://localhost:8000
   ```

## 🎯 Next Steps

### For Bloggers
1. Replace demo content with your articles
2. Customize the design and branding
3. Add your own images and assets
4. Configure custom domain if desired

### For Developers  
1. Explore the MarkGo codebase
2. Customize templates and functionality
3. Add new features or integrations
4. Contribute back to the project

### For Organizations
1. Use as company blog or documentation site
2. Integrate with existing workflows
3. Add team member management
4. Scale with multiple repositories

## 🔗 Resources

- **[MarkGo Repository](https://github.com/vnykmshr/markgo)** - Main project
- **[Static Export Guide](docs/static-export.md)** - Detailed documentation  
- **[GitHub Pages Docs](https://docs.github.com/en/pages)** - Official GitHub Pages guide
- **[Markdown Guide](https://www.markdownguide.org/)** - Markdown syntax reference

## ⭐ Support

If you find this demo helpful:
- ⭐ **Star the repository**
- 🐛 **Report issues** on GitHub
- 💬 **Join discussions** in the repository
- 🚀 **Share** with others who might benefit

---

**Live Demo**: This repository automatically deploys to GitHub Pages  
**Build Status**: Check the Actions tab for deployment status  
**Performance**: Lighthouse score 90+ across all metrics

*Powered by MarkGo Engine - Fast, Simple, Powerful*