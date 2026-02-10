package handlers

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/markgo/internal/services/compose"
)

// Handlers composes all handler types for route registration
type Handlers struct {
	ArticleHandler *ArticleHandler
	AdminHandler   *AdminHandler
	APIHandler     *APIHandler
	ComposeHandler *ComposeHandler
	base           *BaseHandler
	articleService services.ArticleServiceInterface
}

// BuildInfo contains build-time information
type BuildInfo struct {
	Version   string
	GitCommit string
	BuildTime string
}

// Config for handler initialization
type Config struct {
	ArticleService  services.ArticleServiceInterface
	EmailService    services.EmailServiceInterface
	FeedService     services.FeedServiceInterface
	TemplateService services.TemplateServiceInterface
	SEOService      services.SEOServiceInterface
	ComposeService  *compose.Service
	Config          *config.Config
	Logger          *slog.Logger
	BuildInfo       *BuildInfo
}

// New creates a new composed handlers instance
func New(cfg *Config) *Handlers {
	base := NewBaseHandler(cfg.Config, cfg.Logger, cfg.TemplateService, cfg.BuildInfo, cfg.SEOService)

	articleHandler := NewArticleHandler(base, cfg.ArticleService)
	adminHandler := NewAdminHandler(base, cfg.ArticleService, time.Now())
	apiHandler := NewAPIHandler(base, cfg.ArticleService, cfg.EmailService, cfg.FeedService, time.Now())

	var composeHandler *ComposeHandler
	if cfg.ComposeService != nil {
		composeHandler = NewComposeHandler(base, cfg.ComposeService, cfg.ArticleService)
	}

	return &Handlers{
		ArticleHandler: articleHandler,
		AdminHandler:   adminHandler,
		APIHandler:     apiHandler,
		ComposeHandler: composeHandler,
		base:           base,
		articleService: cfg.ArticleService,
	}
}

// Home handles the home page route
func (h *Handlers) Home(c *gin.Context) {
	h.ArticleHandler.Home(c)
}

// Articles handles the articles listing route.
func (h *Handlers) Articles(c *gin.Context) {
	h.ArticleHandler.Articles(c)
}

// Article handles individual article route.
func (h *Handlers) Article(c *gin.Context) {
	h.ArticleHandler.Article(c)
}

// ArticlesByTag handles articles filtered by tag route.
func (h *Handlers) ArticlesByTag(c *gin.Context) {
	h.ArticleHandler.ArticlesByTag(c)
}

// ArticlesByCategory handles articles filtered by category route.
func (h *Handlers) ArticlesByCategory(c *gin.Context) {
	h.ArticleHandler.ArticlesByCategory(c)
}

// Tags handles the tags listing route.
func (h *Handlers) Tags(c *gin.Context) {
	h.ArticleHandler.Tags(c)
}

// Categories handles the categories listing route.
func (h *Handlers) Categories(c *gin.Context) {
	h.ArticleHandler.Categories(c)
}

// Search handles the search route.
func (h *Handlers) Search(c *gin.Context) {
	h.ArticleHandler.Search(c)
}

// AboutArticle handles the about page route.
func (h *Handlers) AboutArticle(c *gin.Context) {
	// Set the slug parameter to "about" for the Article handler
	c.Params = append(c.Params, gin.Param{Key: "slug", Value: "about"})
	h.ArticleHandler.Article(c)
}

// ContactForm handles the contact form route.
func (h *Handlers) ContactForm(c *gin.Context) {
	data := h.base.buildBaseTemplateData("Contact - " + h.base.config.Blog.Title)
	data["template"] = "contact"
	h.base.renderHTML(c, 200, "base.html", data)
}

// ContactSubmit handles contact form submission.
func (h *Handlers) ContactSubmit(c *gin.Context) {
	h.APIHandler.Contact(c)
}

// RSSFeed handles RSS feed generation
func (h *Handlers) RSSFeed(c *gin.Context) {
	h.APIHandler.RSS(c)
}

// JSONFeed handles JSON feed generation.
func (h *Handlers) JSONFeed(c *gin.Context) {
	h.APIHandler.JSONFeed(c)
}

// Sitemap handles sitemap generation.
func (h *Handlers) Sitemap(c *gin.Context) {
	h.APIHandler.Sitemap(c)
}

// Health handles health check requests.
func (h *Handlers) Health(c *gin.Context) {
	h.APIHandler.Health(c)
}

// Metrics handles metrics requests.
func (h *Handlers) Metrics(c *gin.Context) {
	h.AdminHandler.Metrics(c)
}

// AdminHome handles the admin dashboard route
func (h *Handlers) AdminHome(c *gin.Context) {
	h.AdminHandler.AdminHome(c)
}

// AdminStats handles admin statistics requests.
func (h *Handlers) AdminStats(c *gin.Context) {
	h.AdminHandler.Stats(c)
}

// ReloadArticles handles article reload requests.
func (h *Handlers) ReloadArticles(c *gin.Context) {
	h.AdminHandler.ReloadArticles(c)
}

// ClearCache triggers a full article reload from disk.
func (h *Handlers) ClearCache(c *gin.Context) {
	if err := h.articleService.ReloadArticles(); err != nil {
		h.base.logger.Error("Failed to reload articles", "error", err)
		c.JSON(500, gin.H{"error": "Failed to reload articles"})
		return
	}
	c.JSON(200, gin.H{"message": "Cache cleared"})
}

// DebugMemory handles memory debug requests
func (h *Handlers) DebugMemory(c *gin.Context) {
	h.AdminHandler.Debug(c)
}

// DebugRuntime handles runtime debug requests.
func (h *Handlers) DebugRuntime(c *gin.Context) {
	h.AdminHandler.Debug(c)
}

// DebugConfig handles config debug requests.
func (h *Handlers) DebugConfig(c *gin.Context) {
	h.AdminHandler.Debug(c)
}

// DebugRequests handles request debug information.
func (h *Handlers) DebugRequests(c *gin.Context) {
	h.AdminHandler.Debug(c)
}

// PprofIndex handles pprof index requests.
func (h *Handlers) PprofIndex(c *gin.Context) {
	h.AdminHandler.ProfileIndex(c)
}

// PprofProfile handles pprof profile requests.
func (h *Handlers) PprofProfile(c *gin.Context) {
	h.AdminHandler.ProfileProfile(c)
}

// PprofTrace handles pprof trace requests.
func (h *Handlers) PprofTrace(c *gin.Context) {
	h.AdminHandler.ProfileTrace(c)
}

// PprofHeap handles pprof heap requests.
func (h *Handlers) PprofHeap(c *gin.Context) {
	h.AdminHandler.ProfileHeap(c)
}

// PprofGoroutine handles pprof goroutine requests.
func (h *Handlers) PprofGoroutine(c *gin.Context) {
	h.AdminHandler.ProfileGoroutine(c)
}

// PprofAllocs handles pprof allocs requests.
func (h *Handlers) PprofAllocs(c *gin.Context) {
	h.AdminHandler.ProfileAllocs(c)
}

// PprofBlock handles pprof block requests.
func (h *Handlers) PprofBlock(c *gin.Context) {
	h.AdminHandler.ProfileBlock(c)
}

// PprofMutex handles pprof mutex requests.
func (h *Handlers) PprofMutex(c *gin.Context) {
	h.AdminHandler.ProfileMutex(c)
}

// ShowCompose handles the GET /compose route.
func (h *Handlers) ShowCompose(c *gin.Context) {
	if h.ComposeHandler == nil {
		h.NotFound(c)
		return
	}
	h.ComposeHandler.ShowCompose(c)
}

// HandleCompose handles the POST /compose route.
func (h *Handlers) HandleCompose(c *gin.Context) {
	if h.ComposeHandler == nil {
		h.NotFound(c)
		return
	}
	h.ComposeHandler.HandleSubmit(c)
}

// Logger returns the logger instance (used by middleware)
func (h *Handlers) Logger() *slog.Logger {
	return h.base.logger
}

// NotFound handles 404 errors
func (h *Handlers) NotFound(c *gin.Context) {
	data := h.base.buildBaseTemplateData("Page Not Found")
	data["template"] = "404"
	data["description"] = "The page you're looking for was not found"
	h.base.renderHTML(c, 404, "base.html", data)
}
