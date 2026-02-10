// Package export provides static site generation functionality for the MarkGo blog engine.
// It handles exporting dynamic blog content to static HTML files for deployment.
package export

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/handlers"
	"github.com/vnykmshr/markgo/internal/services"
)

// StaticExportService provides static site export functionality.
type StaticExportService struct {
	config          *Config
	logger          *slog.Logger
	outputDir       string
	baseURL         string
	articleService  services.ArticleServiceInterface
	templateService *services.TemplateService
	searchService   services.SearchServiceInterface
	appConfig       *config.Config
	includeDrafts   bool
}

// Config holds export configuration options.
type Config struct {
	OutputDir       string
	StaticPath      string
	BaseURL         string
	ArticleService  services.ArticleServiceInterface
	TemplateService *services.TemplateService
	SearchService   services.SearchServiceInterface
	Config          *config.Config
	Logger          *slog.Logger
	IncludeDrafts   bool
	BuildInfo       *handlers.BuildInfo
}

// NewStaticExportService creates a new StaticExportService instance.
func NewStaticExportService(cfg *Config) (*StaticExportService, error) {
	if cfg.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	return &StaticExportService{
		config:          cfg,
		logger:          cfg.Logger,
		outputDir:       cfg.OutputDir,
		baseURL:         cfg.BaseURL,
		articleService:  cfg.ArticleService,
		templateService: cfg.TemplateService,
		searchService:   cfg.SearchService,
		appConfig:       cfg.Config,
		includeDrafts:   cfg.IncludeDrafts,
	}, nil
}

// Export performs static site export.
func (s *StaticExportService) Export(ctx context.Context) error {
	s.logger.Info("Starting static site export")

	// Create output directory
	if err := s.createOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Copy static assets
	if err := s.copyStaticAssets(); err != nil {
		return fmt.Errorf("failed to copy static assets: %w", err)
	}

	// Generate HTML pages
	if err := s.generatePages(ctx); err != nil {
		return fmt.Errorf("failed to generate pages: %w", err)
	}

	// Generate feeds and sitemaps
	if err := s.generateFeeds(ctx); err != nil {
		return fmt.Errorf("failed to generate feeds: %w", err)
	}

	s.logger.Info("Static site export completed")
	return nil
}

func (s *StaticExportService) createOutputDir() error {
	tmpDir := s.outputDir + ".tmp"
	oldDir := s.outputDir + ".old"

	// Clean up any leftover temp/old dirs from a previous failed run
	_ = os.RemoveAll(tmpDir)
	_ = os.RemoveAll(oldDir)

	// Create the temp directory for the new export
	if err := os.MkdirAll(tmpDir, 0o750); err != nil {
		return fmt.Errorf("failed to create temp output directory: %w", err)
	}

	// Swap: existing → .old, .tmp → final
	if _, err := os.Stat(s.outputDir); err == nil {
		if err := os.Rename(s.outputDir, oldDir); err != nil {
			_ = os.RemoveAll(tmpDir)
			return fmt.Errorf("failed to move existing output directory: %w", err)
		}
	}

	if err := os.Rename(tmpDir, s.outputDir); err != nil {
		// Attempt to restore the old directory
		if _, statErr := os.Stat(oldDir); statErr == nil {
			if restoreErr := os.Rename(oldDir, s.outputDir); restoreErr != nil {
				return fmt.Errorf("failed to rename temp directory to output: %w; additionally, failed to restore previous output from %s: %v", err, oldDir, restoreErr)
			}
		}
		return fmt.Errorf("failed to rename temp directory to output: %w", err)
	}

	// Remove the old directory now that the new one is in place
	_ = os.RemoveAll(oldDir)

	s.logger.Info("Created output directory", "path", s.outputDir)
	return nil
}

func (s *StaticExportService) copyStaticAssets() error {
	staticPath := s.appConfig.StaticPath
	if _, err := os.Stat(staticPath); os.IsNotExist(err) {
		s.logger.Warn("Static directory does not exist", "path", staticPath)
		return nil
	}

	destPath := filepath.Join(s.outputDir, "static")
	if err := s.copyDir(staticPath, destPath); err != nil {
		return fmt.Errorf("failed to copy static assets: %w", err)
	}

	s.logger.Info("Copied static assets", "from", staticPath, "to", destPath)
	return nil
}

func (s *StaticExportService) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return s.copyFile(path, dstPath)
	})
}

func (s *StaticExportService) copyFile(src, dst string) error {
	srcFile, err := os.Open(src) // #nosec G304 - Safe: controlled file copy operation
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	if mkdirErr := os.MkdirAll(filepath.Dir(dst), 0o750); mkdirErr != nil {
		return mkdirErr
	}

	dstFile, err := os.Create(dst) // #nosec G304 - Safe: controlled file copy operation
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *StaticExportService) generatePages(_ context.Context) error {
	// Setup a test server to capture HTML responses
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup templates
	tmpl := s.templateService.GetTemplate()
	if tmpl == nil {
		return fmt.Errorf("no templates loaded")
	}
	router.SetHTMLTemplate(tmpl)

	// Initialize handlers with no cache for export
	h := handlers.New(&handlers.Config{
		ArticleService:  s.articleService,
		TemplateService: s.templateService,
		SearchService:   s.searchService,
		Config:          s.appConfig,
		Logger:          s.logger,
		BuildInfo:       s.config.BuildInfo,
	})

	// Setup routes
	s.setupRoutes(router, h)

	// Generate pages
	pages := s.getPagesToGenerate()
	for _, page := range pages {
		if err := s.generatePage(router, page); err != nil {
			return fmt.Errorf("failed to generate page %s: %w", page.Path, err)
		}
	}

	return nil
}

// Page represents a page to be exported.
type Page struct {
	Path     string
	FilePath string
}

func (s *StaticExportService) getPagesToGenerate() []Page {
	pages := []Page{
		{Path: "/", FilePath: "index.html"},
		{Path: "/articles", FilePath: "articles/index.html"},
		{Path: "/tags", FilePath: "tags/index.html"},
		{Path: "/categories", FilePath: "categories/index.html"},
		{Path: "/about", FilePath: "about/index.html"},
		{Path: "/contact", FilePath: "contact/index.html"},
		{Path: "/search", FilePath: "search/index.html"},
		{Path: "/404", FilePath: "404.html"},
	}

	// Add article pages
	articles := s.articleService.GetAllArticles()

	for _, article := range articles {
		if !s.includeDrafts && article.Draft {
			continue
		}
		pages = append(pages, Page{
			Path:     fmt.Sprintf("/articles/%s", article.Slug),
			FilePath: fmt.Sprintf("articles/%s/index.html", article.Slug),
		})
	}

	// Add tag pages
	tags := s.articleService.GetAllTags()

	for _, tag := range tags {
		// Sanitize tag name for URL and file path
		safeTag := s.sanitizeForURL(tag)
		pages = append(pages, Page{
			Path:     fmt.Sprintf("/tags/%s", safeTag),
			FilePath: fmt.Sprintf("tags/%s/index.html", safeTag),
		})
	}

	// Add category pages
	categories := s.articleService.GetAllCategories()

	for _, category := range categories {
		// Sanitize category name for URL and file path
		safeCategory := s.sanitizeForURL(category)
		pages = append(pages, Page{
			Path:     fmt.Sprintf("/categories/%s", safeCategory),
			FilePath: fmt.Sprintf("categories/%s/index.html", safeCategory),
		})
	}

	return pages
}

func (s *StaticExportService) generatePage(router *gin.Engine, page Page) error {
	// Create HTTP request
	req := httptest.NewRequest("GET", page.Path, http.NoBody)
	req.Host = s.getHostFromBaseURL()

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Check response — allow 404 status for the explicit /404 error page
	is404Page := page.Path == "/404" && w.Code == http.StatusNotFound
	if w.Code != http.StatusOK && !is404Page {
		s.logger.Warn("Non-200 response for page", "path", page.Path, "code", w.Code)
		if w.Code == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("received status code %d for path %s", w.Code, page.Path)
	}

	// Process HTML content to fix relative URLs
	htmlContent := s.processHTML(w.Body.String())

	// Write to file
	filePath := filepath.Join(s.outputDir, page.FilePath)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o750); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
	}

	if err := os.WriteFile(filePath, []byte(htmlContent), 0o600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	s.logger.Debug("Generated page", "path", page.Path, "file", filePath)
	return nil
}

func (s *StaticExportService) processHTML(html string) string {
	// Replace relative URLs with absolute URLs based on baseURL
	if s.baseURL != "" {
		// Simple string replacements for common patterns
		html = strings.ReplaceAll(html, `href="/`, fmt.Sprintf(`href="%s/`, strings.TrimSuffix(s.baseURL, "/")))
		html = strings.ReplaceAll(html, `src="/`, fmt.Sprintf(`src="%s/`, strings.TrimSuffix(s.baseURL, "/")))
		html = strings.ReplaceAll(html, `action="/`, fmt.Sprintf(`action="%s/`, strings.TrimSuffix(s.baseURL, "/")))
	}
	return html
}

func (s *StaticExportService) getHostFromBaseURL() string {
	if s.baseURL == "" {
		return "localhost"
	}

	parsedURL, err := url.Parse(s.baseURL)
	if err != nil {
		return "localhost"
	}

	return parsedURL.Host
}

func (s *StaticExportService) setupRoutes(router *gin.Engine, h *handlers.Handlers) {
	// Static files - not needed for export as we copy them separately
	// router.Static("/static", s.appConfig.StaticPath)

	// Main routes
	router.GET("/", h.Home)
	router.GET("/articles", h.Articles)
	router.GET("/articles/:slug", h.Article)
	router.GET("/tags", h.Tags)
	router.GET("/tags/:tag", h.ArticlesByTag)
	router.GET("/categories", h.Categories)
	router.GET("/categories/:category", h.ArticlesByCategory)
	router.GET("/search", h.Search)
	router.GET("/about", h.AboutArticle)
	router.GET("/contact", h.ContactForm)

	// 404 page for static export
	router.NoRoute(h.NotFound)

	// Feeds and SEO - handled separately in generateFeeds
	router.GET("/feed.xml", h.RSSFeed)
	router.GET("/feed.json", h.JSONFeed)
	router.GET("/sitemap.xml", h.Sitemap)
}

func (s *StaticExportService) generateFeeds(_ context.Context) error {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	h := handlers.New(&handlers.Config{
		ArticleService:  s.articleService,
		TemplateService: s.templateService,
		SearchService:   s.searchService,
		Config:          s.appConfig,
		Logger:          s.logger,
		BuildInfo:       s.config.BuildInfo,
	})

	router.GET("/feed.xml", h.RSSFeed)
	router.GET("/feed.json", h.JSONFeed)
	router.GET("/sitemap.xml", h.Sitemap)
	router.StaticFile("/robots.txt", filepath.Join(s.appConfig.StaticPath, "robots.txt"))

	feeds := []struct {
		path string
		file string
	}{
		{"/feed.xml", "feed.xml"},
		{"/feed.json", "feed.json"},
		{"/sitemap.xml", "sitemap.xml"},
		{"/robots.txt", "robots.txt"},
	}

	for _, feed := range feeds {
		req := httptest.NewRequest("GET", feed.path, http.NoBody)
		req.Host = s.getHostFromBaseURL()
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			s.logger.Warn("Failed to generate feed", "path", feed.path, "code", w.Code)
			continue
		}

		filePath := filepath.Join(s.outputDir, feed.file)
		content := w.Body.String()

		// Process XML content to fix URLs if needed
		if strings.HasSuffix(feed.file, ".xml") {
			content = s.processXML(content)
		}

		if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
			return fmt.Errorf("failed to write %s: %w", feed.file, err)
		}

		s.logger.Debug("Generated feed", "file", feed.file)
	}

	return nil
}

func (s *StaticExportService) processXML(xml string) string {
	// Replace relative URLs in XML feeds with absolute URLs
	if s.baseURL != "" {
		baseURLTrimmed := strings.TrimSuffix(s.baseURL, "/")
		xml = strings.ReplaceAll(xml, `<link>/`, fmt.Sprintf(`<link>%s/`, baseURLTrimmed))
		xml = strings.ReplaceAll(xml, `<url><loc>/`, fmt.Sprintf(`<url><loc>%s/`, baseURLTrimmed))
	}
	return xml
}

func (s *StaticExportService) sanitizeForURL(input string) string {
	// Convert to lowercase
	result := strings.ToLower(input)

	// Replace problematic characters with safe alternatives
	result = strings.ReplaceAll(result, "/", "-")
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "_", "-")

	// Remove special characters that could cause issues in URLs or file paths
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	result = reg.ReplaceAllString(result, "")

	// Remove multiple consecutive dashes
	reg = regexp.MustCompile(`-+`)
	result = reg.ReplaceAllString(result, "-")

	// Trim dashes from start and end
	result = strings.Trim(result, "-")

	return result
}
