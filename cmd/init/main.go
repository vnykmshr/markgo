// Package main provides a command-line tool for initializing a new MarkGo blog.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/vnykmshr/markgo/internal/constants"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
)

const (
	envDevelopment = "development"
	boolFalse      = "false"
)

var (
	help        = flag.Bool("help", false, "Show help message")
	interactive = flag.Bool("interactive", true, "Run in interactive mode")
	quick       = flag.Bool("quick", false, "Quick setup with defaults")
	dir         = flag.String("dir", ".", "Directory to initialize blog in")
	force       = flag.Bool("force", false, "Force overwrite existing files")
)

type BlogConfig struct {
	Title       string
	Description string
	Author      string
	Email       string
	Domain      string
	Port        string
	Environment string
}

func main() {
	cleanup := func() {
		// Cleanup function for error handling
	}

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	fmt.Printf("🚀 MarkGo Blog Engine %s - Quick Setup\n", constants.AppVersion)
	fmt.Println("=====================================")
	fmt.Println()

	// Validate target directory
	targetDir, err := filepath.Abs(*dir)
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("directory validation", fmt.Sprintf("Invalid directory path: %s", *dir), err, 1),
			cleanup,
		)
		return
	}

	// Check if directory exists and is writable
	if err := validateDirectory(targetDir); err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("directory access", fmt.Sprintf("Cannot access directory: %s", targetDir), err, 1),
			cleanup,
		)
		return
	}

	// Check if already initialized
	if !*force && isAlreadyInitialized(targetDir) {
		fmt.Printf("❌ MarkGo blog already exists in %s\n", targetDir)
		fmt.Println("   Use --force flag to overwrite existing files")
		return
	}

	var config BlogConfig

	switch {
	case *quick:
		config = getQuickDefaults()
		fmt.Println("⚡ Quick setup mode - using sensible defaults")
	case *interactive:
		config = runInteractiveSetup()
	default:
		config = getQuickDefaults()
	}

	// Create blog structure
	if err := createBlogStructure(targetDir, config); err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("blog creation", "Failed to create blog structure", err, 1),
			cleanup,
		)
		return
	}

	// Show success message with next steps
	showSuccessMessage(targetDir, config)
}

func validateDirectory(dir string) error {
	// Check if directory exists or can be created
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o750)
	}

	// Check if writable
	testFile := filepath.Join(dir, ".markgo-test")
	if err := os.WriteFile(testFile, []byte("test"), 0o600); err != nil {
		return fmt.Errorf("directory not writable: %w", err)
	}
	_ = os.Remove(testFile)

	return nil
}

func isAlreadyInitialized(dir string) bool {
	markers := []string{".env", "config.yaml", "articles", "web"}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

func getQuickDefaults() BlogConfig {
	currentUser, _ := user.Current()
	username := "Blog Author"
	email := "author@example.com"

	if currentUser != nil {
		username = currentUser.Username
		if currentUser.Name != "" {
			username = currentUser.Name
		}
	}

	return BlogConfig{
		Title:       "My MarkGo Blog",
		Description: "A fast, developer-friendly blog powered by MarkGo",
		Author:      username,
		Email:       email,
		Domain:      "localhost:3000",
		Port:        "3000",
		Environment: envDevelopment,
	}
}

func runInteractiveSetup() BlogConfig {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("📝 Let's set up your new blog! (Press Enter for defaults)")
	fmt.Println()

	defaults := getQuickDefaults()

	config := BlogConfig{
		Title:       getInput(reader, "Blog title", defaults.Title),
		Description: getInput(reader, "Blog description", defaults.Description),
		Author:      getInput(reader, "Author name", defaults.Author),
		Email:       getInput(reader, "Author email", defaults.Email),
		Domain:      getInput(reader, "Domain (for production)", "yourdomain.com"),
		Port:        getInput(reader, "Development port", defaults.Port),
		Environment: getEnvironmentInput(reader),
	}

	fmt.Println()
	fmt.Println("📋 Configuration Summary:")
	fmt.Printf("   Title: %s\n", config.Title)
	fmt.Printf("   Author: %s <%s>\n", config.Author, config.Email)
	fmt.Printf("   Domain: %s\n", config.Domain)
	fmt.Printf("   Environment: %s\n", config.Environment)
	fmt.Println()

	if !getConfirmation(reader, "Proceed with this configuration?") {
		fmt.Println("Setup cancelled.")
		os.Exit(0)
	}

	return config
}

func getInput(reader *bufio.Reader, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}

	return input
}

func getEnvironmentInput(reader *bufio.Reader) string {
	fmt.Println("Environment:")
	fmt.Println("  1. development (recommended for local development)")
	fmt.Println("  2. production (for live deployment)")
	fmt.Printf("Choose environment [1]: ")

	input, err := reader.ReadString('\n')
	if err != nil || strings.TrimSpace(input) == "" || strings.TrimSpace(input) == "1" {
		return envDevelopment
	}

	if strings.TrimSpace(input) == "2" {
		return "production"
	}

	return envDevelopment
}

func getConfirmation(reader *bufio.Reader, prompt string) bool {
	fmt.Printf("%s (y/N): ", prompt)

	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

func createBlogStructure(dir string, config BlogConfig) error {
	// Create directories
	dirs := []string{
		"articles",
		"web/static/css",
		"web/static/js",
		"web/static/images",
		"web/templates",
		"docs",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Create .env file
	if err := createEnvFile(dir, config); err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}

	// Create sample articles
	if err := createSampleArticles(dir, config); err != nil {
		return fmt.Errorf("failed to create sample articles: %w", err)
	}

	// Create basic templates
	if err := createBasicTemplates(dir, config); err != nil {
		return fmt.Errorf("failed to create templates: %w", err)
	}

	// Create basic CSS
	if err := createBasicCSS(dir); err != nil {
		return fmt.Errorf("failed to create CSS: %w", err)
	}

	// Create README
	if err := createReadme(dir, config); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	// Create .gitignore
	if err := createGitignore(dir); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	return nil
}

func createEnvFile(dir string, config BlogConfig) error {
	baseURL := "http://localhost:" + config.Port
	if config.Environment == "production" {
		baseURL = "https://" + config.Domain
	}

	content := fmt.Sprintf(`# MarkGo Blog Configuration
# Generated by markgo init on %s

# Environment Configuration
ENVIRONMENT=%s
PORT=%s
BASE_URL=%s

# Blog Information
BLOG_TITLE=%s
BLOG_DESCRIPTION=%s
BLOG_AUTHOR=%s
BLOG_AUTHOR_EMAIL=%s
BLOG_LANGUAGE=en
BLOG_THEME=default
BLOG_POSTS_PER_PAGE=10

# Paths
ARTICLES_PATH=./articles
STATIC_PATH=./web/static
TEMPLATES_PATH=./web/templates

# Cache Configuration
CACHE_TTL=3600
CACHE_MAX_SIZE=1000
CACHE_CLEANUP_INTERVAL=600

# Rate Limiting
RATE_LIMIT_GENERAL_REQUESTS=100
RATE_LIMIT_GENERAL_WINDOW=900s
RATE_LIMIT_CONTACT_REQUESTS=5
RATE_LIMIT_CONTACT_WINDOW=3600s

# CORS Configuration
CORS_ALLOWED_ORIGINS=%s
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Origin,Content-Type,Accept,Authorization

# Server Timeouts
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_IDLE_TIMEOUT=60s

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_ADD_SOURCE=false

# Email Configuration (Contact Form)
# Configure these to enable contact form functionality
EMAIL_HOST=
EMAIL_PORT=587
EMAIL_USERNAME=
EMAIL_PASSWORD=
EMAIL_FROM=noreply@%s
EMAIL_TO=%s
EMAIL_USE_SSL=true

# Admin Interface (Optional)
# Set credentials to enable admin panel
ADMIN_USERNAME=
ADMIN_PASSWORD=

# Comments System (Optional)
COMMENTS_ENABLED=false
COMMENTS_PROVIDER=giscus

# Analytics (Optional)
ANALYTICS_ENABLED=false
ANALYTICS_PROVIDER=

# Development Settings
TEMPLATE_HOT_RELOAD=%s
DEBUG=%s
`,
		time.Now().Format("2006-01-02 15:04:05"),
		config.Environment,
		config.Port,
		baseURL,
		config.Title,
		config.Description,
		config.Author,
		config.Email,
		baseURL,
		config.Domain,
		config.Email,
		func() string {
			if config.Environment == envDevelopment {
				return "true"
			} else {
				return boolFalse
			}
		}(),
		boolFalse,
	)

	return os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o600)
}

func createSampleArticles(dir string, config BlogConfig) error {
	articlesDir := filepath.Join(dir, "articles")

	// Welcome article - use simpler template
	welcomeContent := "---\n" +
		"title: \"Welcome to " + config.Title + "\"\n" +
		"description: \"Getting started with your new MarkGo blog\"\n" +
		"author: \"" + config.Author + "\"\n" +
		"date: " + time.Now().Format("2006-01-02T15:04:05Z07:00") + "\n" +
		"tags: [\"welcome\", \"getting-started\"]\n" +
		"category: \"general\"\n" +
		"draft: false\n" +
		"featured: true\n" +
		"---\n\n" +
		"# Welcome to Your New MarkGo Blog! 🎉\n\n" +
		"Congratulations on setting up your new MarkGo blog! You now have a lightning-fast, developer-friendly blogging platform powered by Go.\n\n" +
		"## What You Get\n\n" +
		"- **⚡ Blazing Fast Performance**: 4x faster than Ghost with sub-50ms response times\n" +
		"- **📝 Markdown-First**: Write content in familiar Markdown format\n" +
		"- **🛠️ Developer-Friendly**: Git-based workflow, hot reload, comprehensive tooling\n" +
		"- **🚀 Production-Ready**: Built-in caching, rate limiting, metrics, and monitoring\n\n" +
		"## Quick Start\n\n" +
		"1. **Edit this article**: Modify `articles/welcome.md` to customize this post\n" +
		"2. **Create new articles**: Use `markgo new-article` to generate new posts\n" +
		"3. **Start the server**: Run `markgo` to start your blog\n" +
		"4. **Visit your blog**: Open http://localhost:" + config.Port + " in your browser\n\n" +
		"## Writing Articles\n\n" +
		"Articles are stored in the `articles/` directory as Markdown files with frontmatter:\n\n" +
		"```markdown\n" +
		"---\n" +
		"title: \"Your Article Title\"\n" +
		"description: \"Article description for SEO\"\n" +
		"author: \"" + config.Author + "\"\n" +
		"date: 2024-01-01T00:00:00Z\n" +
		"tags: [\"tag1\", \"tag2\"]\n" +
		"category: \"technology\"\n" +
		"draft: false\n" +
		"featured: false\n" +
		"---\n\n" +
		"# Your Article Content\n\n" +
		"Write your content here using Markdown syntax...\n" +
		"```\n\n" +
		"## Next Steps\n\n" +
		"- Customize your blog settings in `.env`\n" +
		"- Modify templates in `web/templates/`\n" +
		"- Add your styles in `web/static/css/`\n" +
		"- Set up email configuration for contact forms\n" +
		"- Configure analytics and comments\n\n" +
		"Happy blogging! 🚀\n\n" +
		"---\n\n" +
		"*This article was generated by `markgo init`. Feel free to edit or delete it.*"

	if err := os.WriteFile(filepath.Join(articlesDir, "welcome.md"), []byte(welcomeContent), 0o600); err != nil {
		return err
	}

	// Getting started article - simple version
	gettingStartedContent := "---\n" +
		"title: \"Getting Started with MarkGo\"\n" +
		"description: \"A comprehensive guide to using your new MarkGo blog\"\n" +
		"author: \"" + config.Author + "\"\n" +
		"date: " + time.Now().Add(-24*time.Hour).Format("2006-01-02T15:04:05Z07:00") + "\n" +
		"tags: [\"guide\", \"tutorial\", \"markgo\"]\n" +
		"category: \"documentation\"\n" +
		"draft: false\n" +
		"featured: false\n" +
		"---\n\n" +
		"# Getting Started with MarkGo\n\n" +
		"This guide will help you get the most out of your new MarkGo blog.\n\n" +
		"## Directory Structure\n\n" +
		"- `articles/` - Your blog posts (Markdown files)\n" +
		"- `web/static/` - CSS, JavaScript, and images\n" +
		"- `web/templates/` - HTML templates\n" +
		"- `.env` - Configuration file\n\n" +
		"## Creating Content\n\n" +
		"Use the CLI tool for the fastest article creation:\n\n" +
		"```bash\n" +
		"# Interactive mode\n" +
		"markgo new-article\n\n" +
		"# Quick creation\n" +
		"markgo new-article --title \"My New Post\" --tags \"tech,go\"\n" +
		"```\n\n" +
		"## Configuration\n\n" +
		"Edit `.env` to customize your blog settings:\n" +
		"- Blog information (title, description, author)\n" +
		"- Server settings (port, timeouts)\n" +
		"- Email configuration for contact forms\n" +
		"- Analytics and comments integration\n\n" +
		"## Development Workflow\n\n" +
		"1. Start the server: `markgo`\n" +
		"2. Create content: `markgo new-article`\n" +
		"3. Edit files (templates auto-reload in development)\n" +
		"4. Test performance: `markgo stress-test`\n\n" +
		"## Performance Features\n\n" +
		"MarkGo is built for speed:\n" +
		"- Template caching and compilation\n" +
		"- Static asset optimization\n" +
		"- Response compression\n" +
		"- Built-in rate limiting\n\n" +
		"## SEO Features\n\n" +
		"- Automatic sitemaps at `/sitemap.xml`\n" +
		"- RSS feeds at `/rss` and `/feed.json`\n" +
		"- Open Graph and Twitter Card meta tags\n" +
		"- JSON-LD structured data\n\n" +
		"## Deployment\n\n" +
		"For production:\n" +
		"1. Set `ENVIRONMENT=production` in `.env`\n" +
		"2. Update `BASE_URL` to your domain\n" +
		"3. Build: `make build`\n" +
		"4. Deploy the binary with your configuration\n\n" +
		"Happy blogging! 🎉"

	return os.WriteFile(filepath.Join(articlesDir, "getting-started.md"), []byte(gettingStartedContent), 0o600)
}

func createBasicTemplates(dir string, config BlogConfig) error {
	// This is a basic implementation - in a real scenario, you'd copy from a templates directory
	// For now, create a simple base template
	templatesDir := filepath.Join(dir, "web", "templates")

	baseTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - ` + config.Title + `</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <header>
        <h1><a href="/">` + config.Title + `</a></h1>
        <nav>
            <a href="/">Home</a>
            <a href="/articles">Articles</a>
            <a href="/tags">Tags</a>
            <a href="/contact">Contact</a>
        </nav>
    </header>
    
    <main>
        {{template "content" .}}
    </main>
    
    <footer>
        <p>&copy; ` + time.Now().Format("2006") + ` ` + config.Author + `. Powered by <a href="https://github.com/vnykmshr/markgo">MarkGo</a>.</p>
    </footer>
</body>
</html>`

	return os.WriteFile(filepath.Join(templatesDir, "base.html"), []byte(baseTemplate), 0o600)
}

func createBasicCSS(dir string) error {
	cssDir := filepath.Join(dir, "web", "static", "css")

	css := `/* MarkGo Default Styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    line-height: 1.6;
    color: #333;
    background: #fff;
}

header {
    background: #f8f9fa;
    padding: 1rem 0;
    border-bottom: 1px solid #e9ecef;
}

header h1 {
    display: inline-block;
    margin-right: 2rem;
}

header h1 a {
    text-decoration: none;
    color: #007bff;
}

nav a {
    margin-right: 1rem;
    text-decoration: none;
    color: #6c757d;
}

nav a:hover {
    color: #007bff;
}

main {
    max-width: 800px;
    margin: 2rem auto;
    padding: 0 1rem;
}

footer {
    text-align: center;
    padding: 2rem;
    margin-top: 4rem;
    background: #f8f9fa;
    border-top: 1px solid #e9ecef;
    color: #6c757d;
}

footer a {
    color: #007bff;
    text-decoration: none;
}

/* Article styles */
article {
    margin-bottom: 3rem;
}

article h1 {
    color: #007bff;
    margin-bottom: 0.5rem;
}

article h2, article h3 {
    margin: 1.5rem 0 0.5rem 0;
}

article p {
    margin-bottom: 1rem;
}

article pre {
    background: #f8f9fa;
    padding: 1rem;
    border-radius: 4px;
    overflow-x: auto;
    margin: 1rem 0;
}

article code {
    background: #f8f9fa;
    padding: 0.2rem 0.4rem;
    border-radius: 3px;
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
}

article pre code {
    background: none;
    padding: 0;
}

/* Responsive design */
@media (max-width: 768px) {
    header h1 {
        display: block;
        margin-bottom: 1rem;
    }
    
    nav a {
        display: inline-block;
        margin-bottom: 0.5rem;
    }
}`

	return os.WriteFile(filepath.Join(cssDir, "style.css"), []byte(css), 0o600)
}

func createReadme(dir string, config BlogConfig) error {
	readme := "# " + config.Title + "\n\n" +
		config.Description + "\n\n" +
		"A fast, developer-friendly blog powered by [MarkGo](https://github.com/vnykmshr/markgo).\n\n" +
		"## Quick Start\n\n" +
		"1. **Start the blog server:** `markgo`\n" +
		"2. **Visit your blog:** Open http://localhost:" + config.Port + "\n" +
		"3. **Create new articles:** `markgo new-article`\n\n" +
		"## Project Structure\n\n" +
		"- `articles/` - Your blog posts (Markdown files)\n" +
		"- `web/static/` - CSS, JavaScript, and images\n" +
		"- `web/templates/` - HTML templates\n" +
		"- `.env` - Configuration file\n\n" +
		"## Writing Articles\n\n" +
		"```bash\n" +
		"# Interactive mode\n" +
		"markgo new-article\n\n" +
		"# Quick creation\n" +
		"markgo new-article --title \"My Post\" --tags \"tech,tutorial\"\n" +
		"```\n\n" +
		"## Performance\n\n" +
		"- Sub-50ms response times\n" +
		"- 4x faster than Ghost\n" +
		"- Built-in caching and optimization\n" +
		"- Test with: `markgo stress-test --url http://localhost:" + config.Port + "`\n\n" +
		"## Features\n\n" +
		"- ✅ Markdown-based content\n" +
		"- ✅ Lightning-fast performance\n" +
		"- ✅ SEO optimization\n" +
		"- ✅ RSS feeds (XML and JSON)\n" +
		"- ✅ Contact forms with email\n" +
		"- ✅ Tag and category organization\n" +
		"- ✅ Search functionality\n" +
		"- ✅ Developer-friendly tooling\n\n" +
		"## Author\n\n" +
		"**" + config.Author + "**  \n" +
		"Email: " + config.Email + "\n\n" +
		"---\n\n" +
		"*Generated by MarkGo " + constants.AppVersion + "*"

	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o600)
}

func createGitignore(dir string) error {
	gitignore := `# Binaries
markgo
markgo.exe
build/
dist/

# Environment files
.env.local
.env.production

# Logs
*.log
logs/

# Cache and temporary files
cache/
tmp/
.tmp/

# Editor and IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Coverage reports
coverage.out
coverage.html

# Test files
test-results/
stress-test-results.json

# Backup files
*.backup
backups/

# Development artifacts
profiles/
debug/
*.pprof
trace.out

# Dependencies (if vendoring)
vendor/

# Local configuration overrides
config.local.yaml
.env.development.local
`

	return os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0o600)
}

func showSuccessMessage(dir string, config BlogConfig) {
	fmt.Println("🎉 MarkGo blog created successfully!")
	fmt.Println()
	fmt.Printf("📁 Location: %s\n", dir)
	fmt.Printf("🌟 Blog: %s\n", config.Title)
	fmt.Printf("👤 Author: %s\n", config.Author)
	fmt.Printf("🌐 Environment: %s\n", config.Environment)
	fmt.Println()

	fmt.Println("📋 What was created:")
	fmt.Println("   ✅ Configuration file (.env)")
	fmt.Println("   ✅ Directory structure (articles/, web/, docs/)")
	fmt.Println("   ✅ Sample articles (welcome.md, getting-started.md)")
	fmt.Println("   ✅ Basic templates and CSS")
	fmt.Println("   ✅ README.md and .gitignore")
	fmt.Println()

	fmt.Println("🚀 Next steps:")
	if dir != "." {
		fmt.Printf("   1. cd %s\n", dir)
	}
	fmt.Println("   2. markgo              # Start your blog server")
	fmt.Printf("   3. Open http://localhost:%s\n", config.Port)
	fmt.Println("   4. markgo new-article  # Create your first article")
	fmt.Println()

	fmt.Println("📖 Learn more:")
	fmt.Println("   • Read articles/getting-started.md")
	fmt.Println("   • Edit .env to customize settings")
	fmt.Println("   • Check out the documentation at docs/")
	fmt.Println()

	fmt.Println("💡 Pro tips:")
	fmt.Println("   • Use 'markgo stress-test' to validate performance")
	fmt.Println("   • Run 'make docs' to generate API documentation")
	fmt.Println("   • Set up email in .env to enable contact forms")
	fmt.Println()

	fmt.Println("Happy blogging! 🎊")
}

func showHelp() {
	fmt.Printf("markgo init - Initialize a new MarkGo blog\n")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  markgo init [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  --dir DIR         Directory to initialize blog in (default: current directory)")
	fmt.Println("  --quick          Quick setup with sensible defaults")
	fmt.Println("  --interactive    Run interactive setup (default)")
	fmt.Println("  --force          Force overwrite existing files")
	fmt.Println("  --help           Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  markgo init                    # Interactive setup in current directory")
	fmt.Println("  markgo init --quick            # Quick setup with defaults")
	fmt.Println("  markgo init --dir my-blog      # Setup in new directory")
	fmt.Println("  markgo init --force            # Overwrite existing files")
	fmt.Println()
	fmt.Println("WHAT IT CREATES:")
	fmt.Println("  .env                  Configuration file")
	fmt.Println("  articles/             Directory for blog posts")
	fmt.Println("  web/static/           Static assets (CSS, JS, images)")
	fmt.Println("  web/templates/        HTML templates")
	fmt.Println("  README.md             Project documentation")
	fmt.Println("  .gitignore            Git ignore file")
	fmt.Println("  Sample articles       Welcome post and getting started guide")
	fmt.Println()
	fmt.Printf("MarkGo %s - The Developer's Blog Engine\n", constants.AppVersion)
}
