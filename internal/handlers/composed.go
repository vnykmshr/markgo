// Package handlers provides HTTP request handlers for the MarkGo blog engine.
// It includes handlers for admin operations, article management, API endpoints, and more.
package handlers

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/obcache-go/pkg/obcache"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/services"
)

// CacheAdapter defines the interface for cache operations
type CacheAdapter interface {
	Clear()
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Stats() map[string]interface{}
}

// ObcacheAdapter provides cache interface using obcache
type ObcacheAdapter struct {
	cache *obcache.Cache
}

func (a *ObcacheAdapter) Clear() {
	if a.cache != nil {
		_ = a.cache.Clear() //nolint:errcheck // Ignore error: cache clear is non-critical
	}
}

func (a *ObcacheAdapter) Get(key string) (interface{}, bool) {
	if a.cache == nil {
		return nil, false
	}
	return a.cache.Get(key)
}

func (a *ObcacheAdapter) Set(key string, value interface{}, ttl time.Duration) {
	if a.cache != nil {
		_ = a.cache.Set(key, value, ttl) //nolint:errcheck // Ignore error: cache set is non-critical
	}
}

func (a *ObcacheAdapter) Delete(key string) {
	if a.cache != nil {
		_ = a.cache.Delete(key) //nolint:errcheck // Ignore error: cache delete is non-critical
	}
}

func (a *ObcacheAdapter) Stats() map[string]interface{} {
	if a.cache == nil {
		return map[string]interface{}{}
	}
	stats := a.cache.Stats()
	return map[string]interface{}{
		"hits":        stats.Hits(),
		"misses":      stats.Misses(),
		"evictions":   stats.Evictions(),
		"hit_rate":    stats.HitRate(),
		"key_count":   stats.KeyCount(),
		"memory_used": "N/A", // obcache doesn't expose memory usage
	}
}

// Handlers composes all handler types for route registration
type Handlers struct {
	ArticleHandler *ArticleHandler
	AdminHandler   *AdminHandler
	APIHandler     *APIHandler
	PreviewHandler *PreviewHandler
	cacheService   CacheAdapter
}

// Config for handler initialization
type Config struct {
	ArticleService  services.ArticleServiceInterface
	EmailService    services.EmailServiceInterface
	SearchService   services.SearchServiceInterface
	TemplateService services.TemplateServiceInterface
	PreviewService  services.PreviewServiceInterface
	Config          *config.Config
	Logger          *slog.Logger
	Cache           *obcache.Cache
}

// New creates a new composed handlers instance
func New(cfg *Config) *Handlers {
	cacheService := &ObcacheAdapter{cache: cfg.Cache}

	// Initialize cached functions for each handler
	cachedFunctions := CachedArticleFunctions{}

	// Create specialized handlers
	articleHandler := NewArticleHandler(
		cfg.Config,
		cfg.Logger,
		cfg.TemplateService,
		cfg.ArticleService,
		cfg.SearchService,
		cachedFunctions,
	)

	adminHandler := NewAdminHandler(
		cfg.Config,
		cfg.Logger,
		cfg.TemplateService,
		cfg.ArticleService,
		time.Now(),
		CachedAdminFunctions{},
	)

	apiHandler := NewAPIHandler(
		cfg.Config,
		cfg.Logger,
		cfg.TemplateService,
		cfg.ArticleService,
		cfg.EmailService,
		time.Now(),
		CachedAPIFunctions{},
	)

	// Create preview handler (if service is available)
	var previewHandler *PreviewHandler
	if cfg.PreviewService != nil {
		previewHandler = NewPreviewHandler(
			cfg.PreviewService,
			cfg.ArticleService,
			cfg.TemplateService,
			BaseHandler{logger: cfg.Logger},
		)
	}

	return &Handlers{
		ArticleHandler: articleHandler,
		AdminHandler:   adminHandler,
		APIHandler:     apiHandler,
		PreviewHandler: previewHandler,
		cacheService:   cacheService,
	}
}

// Home handles the home page route
func (h *Handlers) Home(c *gin.Context) {
	h.ArticleHandler.Home(c)
}

func (h *Handlers) Articles(c *gin.Context) {
	h.ArticleHandler.Articles(c)
}

func (h *Handlers) Article(c *gin.Context) {
	h.ArticleHandler.Article(c)
}

func (h *Handlers) ArticlesByTag(c *gin.Context) {
	h.ArticleHandler.ArticlesByTag(c)
}

func (h *Handlers) ArticlesByCategory(c *gin.Context) {
	h.ArticleHandler.ArticlesByCategory(c)
}

func (h *Handlers) Tags(c *gin.Context) {
	h.ArticleHandler.Tags(c)
}

func (h *Handlers) Categories(c *gin.Context) {
	h.ArticleHandler.Categories(c)
}

func (h *Handlers) Search(c *gin.Context) {
	h.ArticleHandler.Search(c)
}

func (h *Handlers) AboutArticle(c *gin.Context) {
	// Set the slug parameter to "about" for the Article handler
	c.Params = append(c.Params, gin.Param{Key: "slug", Value: "about"})
	h.ArticleHandler.Article(c)
}

func (h *Handlers) ContactForm(c *gin.Context) {
	// Render contact form template
	data := h.ArticleHandler.buildBaseTemplateData("Contact")
	data["recent"] = h.ArticleHandler.getRecentArticles()
	data["template"] = "contact"
	h.ArticleHandler.renderHTML(c, 200, "base.html", data)
}

func (h *Handlers) ContactSubmit(c *gin.Context) {
	h.APIHandler.Contact(c)
}

// RSSFeed handles RSS feed generation
func (h *Handlers) RSSFeed(c *gin.Context) {
	h.APIHandler.RSS(c)
}

func (h *Handlers) JSONFeed(c *gin.Context) {
	h.APIHandler.JSONFeed(c)
}

func (h *Handlers) Sitemap(c *gin.Context) {
	h.APIHandler.Sitemap(c)
}

func (h *Handlers) Health(c *gin.Context) {
	h.APIHandler.Health(c)
}

func (h *Handlers) Metrics(c *gin.Context) {
	h.AdminHandler.Metrics(c)
}

// AdminHome handles the admin dashboard route
func (h *Handlers) AdminHome(c *gin.Context) {
	h.AdminHandler.AdminHome(c)
}

func (h *Handlers) AdminStats(c *gin.Context) {
	h.AdminHandler.Stats(c)
}

func (h *Handlers) ReloadArticles(c *gin.Context) {
	h.AdminHandler.ReloadArticles(c)
}

// GetDrafts handles retrieving draft articles
func (h *Handlers) GetDrafts(c *gin.Context) {
	drafts := h.AdminHandler.articleService.GetDraftArticles()
	c.JSON(200, gin.H{
		"drafts": drafts,
		"count":  len(drafts),
	})
}

func (h *Handlers) GetDraftBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(400, gin.H{"error": "Slug parameter is required"})
		return
	}

	draft, err := h.AdminHandler.articleService.GetDraftBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Draft not found", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"draft": draft})
}

func (h *Handlers) PreviewDraft(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(400, gin.H{"error": "Slug parameter is required"})
		return
	}

	draft, err := h.AdminHandler.articleService.PreviewDraft(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Draft not found", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"preview": draft})
}

func (h *Handlers) PublishDraft(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(400, gin.H{
			"error":   "Slug parameter is required",
			"example": "POST /admin/drafts/my-article/publish"})
		return
	}

	// Additional validation
	if len(slug) > 100 {
		c.JSON(400, gin.H{
			"error":           "Slug too long",
			"max_length":      100,
			"provided_length": len(slug)})
		return
	}

	err := h.AdminHandler.articleService.PublishDraft(slug)
	if err != nil {
		h.AdminHandler.logger.Error("Failed to publish draft", "slug", slug, "error", err)

		// Provide different error codes based on error type
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "not found"):
			c.JSON(404, gin.H{
				"error":      "Draft not found",
				"slug":       slug,
				"suggestion": "Check that the draft exists and slug is correct"})
		case strings.Contains(errMsg, "not a draft"):
			c.JSON(400, gin.H{
				"error": "Article is already published",
				"slug":  slug})
		default:
			c.JSON(500, gin.H{
				"error":   "Failed to publish draft",
				"slug":    slug,
				"details": errMsg})
		}
		return
	}

	c.JSON(200, gin.H{
		"message": "Draft published successfully",
		"slug":    slug,
		"status":  "published"})
}

func (h *Handlers) UnpublishArticle(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(400, gin.H{
			"error":   "Slug parameter is required",
			"example": "POST /admin/articles/my-article/unpublish"})
		return
	}

	// Additional validation
	if len(slug) > 100 {
		c.JSON(400, gin.H{
			"error":           "Slug too long",
			"max_length":      100,
			"provided_length": len(slug)})
		return
	}

	err := h.AdminHandler.articleService.UnpublishArticle(slug)
	if err != nil {
		h.AdminHandler.logger.Error("Failed to unpublish article", "slug", slug, "error", err)

		// Provide different error codes based on error type
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "not found"):
			c.JSON(404, gin.H{
				"error":      "Article not found",
				"slug":       slug,
				"suggestion": "Check that the article exists and slug is correct"})
		case strings.Contains(errMsg, "already"):
			c.JSON(400, gin.H{
				"error": "Article is already a draft",
				"slug":  slug})
		default:
			c.JSON(500, gin.H{
				"error":   "Failed to unpublish article",
				"slug":    slug,
				"details": errMsg})
		}
		return
	}

	c.JSON(200, gin.H{
		"message": "Article unpublished successfully",
		"slug":    slug,
		"status":  "draft"})
}

func (h *Handlers) ClearCache(c *gin.Context) {
	if h.cacheService != nil {
		h.cacheService.Clear()
	}
	c.JSON(200, gin.H{"message": "Cache cleared"})
}

// DebugMemory handles memory debug requests
func (h *Handlers) DebugMemory(c *gin.Context) {
	h.AdminHandler.Debug(c)
}

func (h *Handlers) DebugRuntime(c *gin.Context) {
	h.AdminHandler.Debug(c)
}

func (h *Handlers) DebugConfig(c *gin.Context) {
	h.AdminHandler.Debug(c)
}

func (h *Handlers) DebugRequests(c *gin.Context) {
	h.AdminHandler.Debug(c)
}

func (h *Handlers) SetLogLevel(c *gin.Context) {
	h.AdminHandler.SetLogLevel(c)
}

func (h *Handlers) PprofIndex(c *gin.Context) {
	h.AdminHandler.ProfileIndex(c)
}

func (h *Handlers) PprofProfile(c *gin.Context) {
	h.AdminHandler.ProfileProfile(c)
}

func (h *Handlers) PprofTrace(c *gin.Context) {
	h.AdminHandler.ProfileTrace(c)
}

func (h *Handlers) PprofHeap(c *gin.Context) {
	h.AdminHandler.ProfileHeap(c)
}

func (h *Handlers) PprofGoroutine(c *gin.Context) {
	h.AdminHandler.ProfileGoroutine(c)
}

func (h *Handlers) PprofAllocs(c *gin.Context) {
	h.AdminHandler.ProfileAllocs(c)
}

func (h *Handlers) PprofBlock(c *gin.Context) {
	h.AdminHandler.ProfileBlock(c)
}

func (h *Handlers) PprofMutex(c *gin.Context) {
	h.AdminHandler.ProfileMutex(c)
}

// Logger returns the logger instance (used by middleware)
func (h *Handlers) Logger() *slog.Logger {
	return h.ArticleHandler.logger
}

// NotFound handles 404 errors
func (h *Handlers) NotFound(c *gin.Context) {
	data := h.ArticleHandler.buildBaseTemplateData("Page Not Found")
	data["template"] = "404"
	data["description"] = "The page you're looking for was not found"
	h.ArticleHandler.renderHTML(c, 404, "base.html", data)
}
