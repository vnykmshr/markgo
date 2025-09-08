package handlers

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// CachedHandlerFunctions holds obcache-wrapped handler operations
type CachedHandlerFunctions struct {
	GetHomeData         func() (map[string]any, error)
	GetArticleData      func(string) (map[string]any, error)
	GetArticlesPage     func(int) (map[string]any, error)
	GetTagArticles      func(string) (map[string]any, error)
	GetCategoryArticles func(string) (map[string]any, error)
	GetSearchResults    func(string) (map[string]any, error)
	GetArchiveData      func(string, string) (map[string]any, error)
	GetTagsPage         func() (map[string]any, error)
	GetCategoriesPage   func() (map[string]any, error)
	GetStatsData        func() (map[string]any, error)
	GetRSSFeed          func() (string, error)
	GetJSONFeed         func() (string, error)
	GetSitemap          func() (string, error)
}

// CacheAdapter provides a cache interface for handlers
type CacheAdapter interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttl time.Duration)
	Delete(key string)
	Clear()
	Size() int
	Stats() map[string]any
}

// ObcacheAdapter provides CacheService-compatible interface using obcache
type ObcacheAdapter struct {
	cache *obcache.Cache
}

func (a *ObcacheAdapter) Get(key string) (any, bool) {
	if a.cache == nil {
		return nil, false
	}
	return a.cache.Get(key)
}

func (a *ObcacheAdapter) Set(key string, value any, ttl time.Duration) {
	if a.cache == nil {
		return
	}
	_ = a.cache.Set(key, value, ttl)
}

func (a *ObcacheAdapter) Delete(key string) {
	if a.cache == nil {
		return
	}
	_ = a.cache.Delete(key)
}

func (a *ObcacheAdapter) Clear() {
	if a.cache == nil {
		return
	}
	_ = a.cache.Clear()
}

func (a *ObcacheAdapter) Size() int {
	if a.cache == nil {
		return 0
	}
	stats := a.cache.Stats()
	return int(stats.KeyCount())
}

func (a *ObcacheAdapter) Stats() map[string]any {
	if a.cache == nil {
		return make(map[string]any)
	}
	stats := a.cache.Stats()
	return map[string]any{
		"key_count":  stats.KeyCount(),
		"hit_count":  stats.Hits(),
		"miss_count": stats.Misses(),
		"hit_ratio":  stats.HitRate(),
	}
}

// Handlers provides the refactored, composed handler structure
type Handlers struct {
	ArticleHandler *ArticleHandler
	AdminHandler   *AdminHandler
	APIHandler     *APIHandler

	// Keep original handlers for backward compatibility during migration
	*LegacyHandlers
}

// LegacyHandlers embeds the original handlers struct for gradual migration
type LegacyHandlers struct {
	// Original handlers struct fields - to be removed after migration
	articleService  services.ArticleServiceInterface
	emailService    services.EmailServiceInterface
	searchService   services.SearchServiceInterface
	templateService services.TemplateServiceInterface
	config          *config.Config
	logger          *slog.Logger
	startTime       time.Time
	cache           *obcache.Cache
	cachedFunctions CachedHandlerFunctions
	cacheService    CacheAdapter
}

// Config for handler initialization - unchanged for compatibility
type Config struct {
	ArticleService  services.ArticleServiceInterface
	EmailService    services.EmailServiceInterface
	SearchService   services.SearchServiceInterface
	TemplateService services.TemplateServiceInterface
	Config          *config.Config
	Logger          *slog.Logger
	Cache           *obcache.Cache
}

// New creates a new composed handlers instance
func New(cfg *Config) *Handlers {
	// Create legacy handlers for backward compatibility
	legacyHandlers := &LegacyHandlers{
		articleService:  cfg.ArticleService,
		emailService:    cfg.EmailService,
		searchService:   cfg.SearchService,
		templateService: cfg.TemplateService,
		config:          cfg.Config,
		logger:          cfg.Logger,
		startTime:       time.Now(),
		cache:           cfg.Cache,
		cacheService:    &ObcacheAdapter{cache: cfg.Cache},
	}

	// Initialize cached functions (reuse existing logic)
	cachedFunctions := initializeCachedFunctions(cfg, legacyHandlers)
	legacyHandlers.cachedFunctions = cachedFunctions

	// Create specialized handlers
	articleHandler := NewArticleHandler(
		cfg.Config,
		cfg.Logger,
		cfg.TemplateService,
		cfg.ArticleService,
		cfg.SearchService,
		CachedArticleFunctions{
			GetHomeData:         cachedFunctions.GetHomeData,
			GetArticleData:      cachedFunctions.GetArticleData,
			GetArticlesPage:     cachedFunctions.GetArticlesPage,
			GetTagArticles:      cachedFunctions.GetTagArticles,
			GetCategoryArticles: cachedFunctions.GetCategoryArticles,
			GetSearchResults:    cachedFunctions.GetSearchResults,
			GetTagsPage:         cachedFunctions.GetTagsPage,
			GetCategoriesPage:   cachedFunctions.GetCategoriesPage,
		},
	)

	adminHandler := NewAdminHandler(
		cfg.Config,
		cfg.Logger,
		cfg.TemplateService,
		cfg.ArticleService,
		legacyHandlers.startTime,
		CachedAdminFunctions{
			GetStatsData: cachedFunctions.GetStatsData,
		},
	)

	apiHandler := NewAPIHandler(
		cfg.Config,
		cfg.Logger,
		cfg.TemplateService,
		cfg.ArticleService,
		cfg.EmailService,
		legacyHandlers.startTime,
		CachedAPIFunctions{
			GetRSSFeed:  cachedFunctions.GetRSSFeed,
			GetJSONFeed: cachedFunctions.GetJSONFeed,
			GetSitemap:  cachedFunctions.GetSitemap,
		},
	)

	return &Handlers{
		ArticleHandler: articleHandler,
		AdminHandler:   adminHandler,
		APIHandler:     apiHandler,
		LegacyHandlers: legacyHandlers,
	}
}

// initializeCachedFunctions creates the cached functions (extracted from original logic)
func initializeCachedFunctions(cfg *Config, legacy *LegacyHandlers) CachedHandlerFunctions {
	// This is a simplified version - the original logic from handlers.go would be extracted here
	// For now, return empty functions to maintain compatibility
	return CachedHandlerFunctions{
		// These would be populated with actual obcache.Wrap calls
		// GetHomeData: obcache.Wrap(cfg.Cache, legacy.getHomeDataUncached, ...),
		// etc.
	}
}

// Legacy method delegation for backward compatibility
// These methods delegate to the appropriate specialized handlers

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

func (h *Handlers) Metrics(c *gin.Context) {
	h.AdminHandler.Metrics(c)
}

func (h *Handlers) Stats(c *gin.Context) {
	h.AdminHandler.Stats(c)
}

func (h *Handlers) Debug(c *gin.Context) {
	h.AdminHandler.Debug(c)
}

func (h *Handlers) ReloadArticles(c *gin.Context) {
	h.AdminHandler.ReloadArticles(c)
}

func (h *Handlers) CompactMemory(c *gin.Context) {
	h.AdminHandler.CompactMemory(c)
}

func (h *Handlers) Health(c *gin.Context) {
	h.APIHandler.Health(c)
}

func (h *Handlers) RSS(c *gin.Context) {
	h.APIHandler.RSS(c)
}

func (h *Handlers) JSONFeed(c *gin.Context) {
	h.APIHandler.JSONFeed(c)
}

func (h *Handlers) Sitemap(c *gin.Context) {
	h.APIHandler.Sitemap(c)
}

func (h *Handlers) Contact(c *gin.Context) {
	h.APIHandler.Contact(c)
}

// Profile endpoints delegation
func (h *Handlers) ProfileIndex(c *gin.Context) {
	h.AdminHandler.ProfileIndex(c)
}

func (h *Handlers) ProfileCmdline(c *gin.Context) {
	h.AdminHandler.ProfileCmdline(c)
}

func (h *Handlers) ProfileProfile(c *gin.Context) {
	h.AdminHandler.ProfileProfile(c)
}

func (h *Handlers) ProfileSymbol(c *gin.Context) {
	h.AdminHandler.ProfileSymbol(c)
}

func (h *Handlers) ProfileTrace(c *gin.Context) {
	h.AdminHandler.ProfileTrace(c)
}

func (h *Handlers) ProfileHeap(c *gin.Context) {
	h.AdminHandler.ProfileHeap(c)
}

func (h *Handlers) ProfileGoroutine(c *gin.Context) {
	h.AdminHandler.ProfileGoroutine(c)
}

func (h *Handlers) ProfileBlock(c *gin.Context) {
	h.AdminHandler.ProfileBlock(c)
}

func (h *Handlers) ProfileMutex(c *gin.Context) {
	h.AdminHandler.ProfileMutex(c)
}

// Utility methods that need to be accessible
func (h *Handlers) GetConfig() *config.Config {
	return h.config
}

func (h *Handlers) GetLogger() *slog.Logger {
	return h.logger
}

// Additional delegated methods
func (h *Handlers) AboutArticle(c *gin.Context) {
	h.ArticleHandler.Article(c)
}

func (h *Handlers) ContactForm(c *gin.Context) {
	// Render contact form template
	data := h.ArticleHandler.buildBaseTemplateData("Contact").
		Set("template", "contact").
		Build()
	h.ArticleHandler.renderHTML(c, 200, "base.html", data)
}

func (h *Handlers) ContactSubmit(c *gin.Context) {
	h.APIHandler.Contact(c)
}

func (h *Handlers) RSSFeed(c *gin.Context) {
	h.APIHandler.RSS(c)
}

func (h *Handlers) ClearCache(c *gin.Context) {
	if h.cacheService != nil {
		h.cacheService.Clear()
	}
	c.JSON(200, gin.H{"message": "Cache cleared"})
}

func (h *Handlers) AdminStats(c *gin.Context) {
	h.AdminHandler.Stats(c)
}

func (h *Handlers) GetDrafts(c *gin.Context) {
	drafts := h.articleService.GetDraftArticles()
	c.JSON(200, drafts)
}

func (h *Handlers) GetDraftBySlug(c *gin.Context) {
	slug := c.Param("slug")
	draft, err := h.articleService.GetDraftBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Draft not found"})
		return
	}
	c.JSON(200, draft)
}

func (h *Handlers) PreviewDraft(c *gin.Context) {
	slug := c.Param("slug")
	draft, err := h.articleService.PreviewDraft(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Draft not found"})
		return
	}
	c.JSON(200, draft)
}

func (h *Handlers) PublishDraft(c *gin.Context) {
	slug := c.Param("slug")
	err := h.articleService.PublishDraft(slug)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to publish draft"})
		return
	}
	c.JSON(200, gin.H{"message": "Draft published"})
}

func (h *Handlers) UnpublishArticle(c *gin.Context) {
	slug := c.Param("slug")
	err := h.articleService.UnpublishArticle(slug)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to unpublish article"})
		return
	}
	c.JSON(200, gin.H{"message": "Article unpublished"})
}

// Debug endpoints
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
	c.JSON(200, gin.H{"message": "Log level not implemented"})
}

// Pprof endpoints
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

func (h *Handlers) PprofCmdline(c *gin.Context) {
	h.AdminHandler.ProfileCmdline(c)
}

func (h *Handlers) PprofSymbol(c *gin.Context) {
	h.AdminHandler.ProfileSymbol(c)
}

func (h *Handlers) PprofBlock(c *gin.Context) {
	h.AdminHandler.ProfileBlock(c)
}

func (h *Handlers) PprofMutex(c *gin.Context) {
	h.AdminHandler.ProfileMutex(c)
}

func (h *Handlers) PprofAllocs(c *gin.Context) {
	// Allocs profile - similar to heap
	h.AdminHandler.ProfileHeap(c)
}

func (h *Handlers) NotFound(c *gin.Context) {
	c.JSON(404, gin.H{"error": "Not found"})
}
