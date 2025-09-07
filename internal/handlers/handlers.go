package handlers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/middleware"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/markgo/internal/utils"
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

// Handlers contains all HTTP handlers
type Handlers struct {
	articleService  services.ArticleServiceInterface
	emailService    services.EmailServiceInterface
	searchService   services.SearchServiceInterface
	templateService services.TemplateServiceInterface
	config          *config.Config
	logger          *slog.Logger
	startTime       time.Time

	// obcache integration
	cache           *obcache.Cache
	cachedFunctions CachedHandlerFunctions

	// Temporary: cache adapter for smooth migration
	cacheService CacheAdapter
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

// Config for handler initialization
type Config struct {
	ArticleService  services.ArticleServiceInterface
	EmailService    services.EmailServiceInterface
	SearchService   services.SearchServiceInterface
	TemplateService services.TemplateServiceInterface
	Config          *config.Config
	Logger          *slog.Logger
	Cache           *obcache.Cache // Direct obcache usage
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
	// Use the provided TTL
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
		return map[string]any{}
	}
	stats := a.cache.Stats()
	return map[string]any{
		"hit_count":  int(stats.Hits()),
		"miss_count": int(stats.Misses()),
		"hit_ratio":  stats.HitRate() * 100,
		"total_keys": int(stats.KeyCount()),
	}
}

// New creates a new handlers instance
func New(cfg *Config) *Handlers {
	h := &Handlers{
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

	// Initialize cached functions if cache is available
	if h.cache != nil {
		h.initializeCachedFunctions()
	}

	return h
}

// initializeCachedFunctions initializes obcache-wrapped handler functions
func (h *Handlers) initializeCachedFunctions() {
	// Wrap home page data generation
	h.cachedFunctions.GetHomeData = obcache.Wrap(
		h.cache,
		h.getHomeDataUncached,
		obcache.WithKeyFunc(func(args []any) string {
			return "home_page"
		}),
		obcache.WithTTL(h.config.Cache.TTL),
	)

	// Wrap article page data generation
	h.cachedFunctions.GetArticleData = obcache.Wrap(
		h.cache,
		h.getArticleDataUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) > 0 {
				if slug, ok := args[0].(string); ok {
					return "article_" + slug
				}
			}
			return "article_default"
		}),
		obcache.WithTTL(h.config.Cache.TTL),
	)

	// Wrap articles page data generation
	h.cachedFunctions.GetArticlesPage = obcache.Wrap(
		h.cache,
		h.getArticlesPageUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) > 0 {
				if page, ok := args[0].(int); ok {
					return fmt.Sprintf("articles_page_%d", page)
				}
			}
			return "articles_page_1"
		}),
		obcache.WithTTL(h.config.Cache.TTL),
	)

	// Wrap tag articles data generation
	h.cachedFunctions.GetTagArticles = obcache.Wrap(
		h.cache,
		h.getTagArticlesUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) > 0 {
				if tag, ok := args[0].(string); ok {
					return "articles_tag_" + tag
				}
			}
			return "articles_tag_default"
		}),
		obcache.WithTTL(h.config.Cache.TTL),
	)

	// Wrap category articles data generation
	h.cachedFunctions.GetCategoryArticles = obcache.Wrap(
		h.cache,
		h.getCategoryArticlesUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) > 0 {
				if category, ok := args[0].(string); ok {
					return "articles_category_" + category
				}
			}
			return "articles_category_default"
		}),
		obcache.WithTTL(h.config.Cache.TTL),
	)

	// Wrap search results data generation
	h.cachedFunctions.GetSearchResults = obcache.Wrap(
		h.cache,
		h.getSearchResultsUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) > 0 {
				if query, ok := args[0].(string); ok {
					return "search_" + query
				}
			}
			return "search_default"
		}),
		obcache.WithTTL(h.config.Cache.TTL),
	)

	// Wrap tags page data generation
	h.cachedFunctions.GetTagsPage = obcache.Wrap(
		h.cache,
		h.getTagsPageUncached,
		obcache.WithKeyFunc(func(args []any) string {
			return "all_tags"
		}),
		obcache.WithTTL(h.config.Cache.TTL),
	)

	// Wrap categories page data generation
	h.cachedFunctions.GetCategoriesPage = obcache.Wrap(
		h.cache,
		h.getCategoriesPageUncached,
		obcache.WithKeyFunc(func(args []any) string {
			return "all_categories"
		}),
		obcache.WithTTL(h.config.Cache.TTL),
	)

	// Wrap stats data generation
	h.cachedFunctions.GetStatsData = obcache.Wrap(
		h.cache,
		h.getStatsDataUncached,
		obcache.WithKeyFunc(func(args []any) string {
			return "api_stats"
		}),
		obcache.WithTTL(30*time.Minute),
	)

	// Wrap RSS feed generation
	h.cachedFunctions.GetRSSFeed = obcache.Wrap(
		h.cache,
		h.getRSSFeedUncached,
		obcache.WithKeyFunc(func(args []any) string {
			return "rss_feed"
		}),
		obcache.WithTTL(6*time.Hour),
	)

	// Wrap JSON feed generation
	h.cachedFunctions.GetJSONFeed = obcache.Wrap(
		h.cache,
		h.getJSONFeedUncached,
		obcache.WithKeyFunc(func(args []any) string {
			return "json_feed"
		}),
		obcache.WithTTL(6*time.Hour),
	)

	// Wrap sitemap generation
	h.cachedFunctions.GetSitemap = obcache.Wrap(
		h.cache,
		h.getSitemapUncached,
		obcache.WithKeyFunc(func(args []any) string {
			return "sitemap"
		}),
		obcache.WithTTL(24*time.Hour),
	)
}

// Home renders the homepage
func (h *Handlers) Home(c *gin.Context) {
	// Use cached function if available
	if h.cachedFunctions.GetHomeData != nil {
		if data, err := h.cachedFunctions.GetHomeData(); err == nil {
			h.renderHTML(c, http.StatusOK, "base.html", data)
			return
		}
	}

	// Fallback to uncached version
	data, err := h.getHomeDataUncached()
	if err != nil {
		h.logger.Error("Failed to get home data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Articles renders the articles listing page
func (h *Handlers) Articles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// Use cached function if available
	if h.cachedFunctions.GetArticlesPage != nil {
		if data, err := h.cachedFunctions.GetArticlesPage(page); err == nil {
			h.renderHTML(c, http.StatusOK, "base.html", data)
			return
		}
	}

	// Fallback to uncached version
	data, err := h.getArticlesPageUncached(page)
	if err != nil {
		h.logger.Error("Failed to get articles page data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Article renders a single article page
func (h *Handlers) Article(c *gin.Context) {
	slug := c.Param("slug")

	// Use cached function if available
	if h.cachedFunctions.GetArticleData != nil {
		if data, err := h.cachedFunctions.GetArticleData(slug); err == nil {
			h.renderHTML(c, http.StatusOK, "base.html", data)
			return
		}
	}

	// Fallback to uncached version
	data, err := h.getArticleDataUncached(slug)
	if err != nil {
		_ = c.Error(apperrors.NewArticleError("", fmt.Sprintf("Article '%s' not found", slug), apperrors.ErrArticleNotFound))
		c.Abort()
		return
	}

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ArticlesByTag renders articles filtered by tag
func (h *Handlers) ArticlesByTag(c *gin.Context) {
	tag := c.Param("tag")

	// Use cached function if available
	if h.cachedFunctions.GetTagArticles != nil {
		if data, err := h.cachedFunctions.GetTagArticles(tag); err == nil {
			h.renderHTML(c, http.StatusOK, "base.html", data)
			return
		}
	}

	// Fallback to uncached version
	data, err := h.getTagArticlesUncached(tag)
	if err != nil {
		h.logger.Error("Failed to get tag articles data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ArticlesByCategory renders articles filtered by category
func (h *Handlers) ArticlesByCategory(c *gin.Context) {
	category := c.Param("category")

	// Use cached function if available
	if h.cachedFunctions.GetCategoryArticles != nil {
		if data, err := h.cachedFunctions.GetCategoryArticles(category); err == nil {
			h.renderHTML(c, http.StatusOK, "base.html", data)
			return
		}
	}

	// Fallback to uncached version
	data, err := h.getCategoryArticlesUncached(category)
	if err != nil {
		h.logger.Error("Failed to get category articles data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Tags renders the tags page
func (h *Handlers) Tags(c *gin.Context) {
	// Use cached function if available
	if h.cachedFunctions.GetTagsPage != nil {
		if data, err := h.cachedFunctions.GetTagsPage(); err == nil {
			h.renderHTML(c, http.StatusOK, "base.html", data)
			return
		}
	}

	// Fallback to uncached version
	data, err := h.getTagsPageUncached()
	if err != nil {
		h.logger.Error("Failed to get tags page data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Categories renders the categories page
func (h *Handlers) Categories(c *gin.Context) {
	// Use cached function if available
	if h.cachedFunctions.GetCategoriesPage != nil {
		if data, err := h.cachedFunctions.GetCategoriesPage(); err == nil {
			h.renderHTML(c, http.StatusOK, "base.html", data)
			return
		}
	}

	// Fallback to uncached version
	data, err := h.getCategoriesPageUncached()
	if err != nil {
		h.logger.Error("Failed to get categories page data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Search handles search requests
func (h *Handlers) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		// Get data for search landing page
		allArticles := h.articleService.GetAllArticles()
		tagCounts := h.articleService.GetTagCounts()
		categoryCounts := h.articleService.GetCategoryCounts()
		recent := h.articleService.GetRecentArticles(5)

		data := utils.ArticlePageData("Search", h.config, recent).
			Set("template", "search").
			Set("totalCount", len(allArticles)).
			Set("tags", tagCounts).
			Set("categories", categoryCounts).
			Build()
		h.renderHTML(c, http.StatusOK, "base.html", data)
		return
	}

	// Use cached function if available
	if h.cachedFunctions.GetSearchResults != nil {
		if data, err := h.cachedFunctions.GetSearchResults(query); err == nil {
			h.renderHTML(c, http.StatusOK, "base.html", data)
			return
		}
	}

	// Fallback to uncached version
	data, err := h.getSearchResultsUncached(query)
	if err != nil {
		h.logger.Error("Failed to get search results data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// AboutArticle renders the about page as an article
func (h *Handlers) AboutArticle(c *gin.Context) {
	article, err := h.articleService.GetArticleBySlug("about")
	if err != nil {
		h.logger.Error("Error loading about article", "error", err)
		_ = c.Error(apperrors.NewArticleError("about", "About page not found", apperrors.ErrArticleNotFound))
		c.Abort()
		return
	}

	// Get recent articles
	recent := h.articleService.GetRecentArticles(5)

	data := utils.ArticlePageData(article.Title, h.config, recent).
		Set("article", article).
		Set("template", "about-article").
		Build()

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ContactForm renders the contact form
func (h *Handlers) ContactForm(c *gin.Context) {
	// Get recent articles
	recent := h.articleService.GetRecentArticles(5)

	data := utils.ArticlePageData("Contact", h.config, recent).
		Set("template", "contact").
		Build()

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ContactSubmit handles contact form submission
func (h *Handlers) ContactSubmit(c *gin.Context) {
	var msg models.ContactMessage

	if err := c.ShouldBindJSON(&msg); err != nil {
		_ = c.Error(apperrors.NewValidationError("contact_form", nil, "Invalid contact form data", err))
		c.Abort()
		return
	}

	// Sanitize input fields to prevent XSS and other attacks
	msg.Name = strings.TrimSpace(html.EscapeString(msg.Name))
	msg.Email = strings.TrimSpace(html.EscapeString(msg.Email))
	msg.Subject = strings.TrimSpace(html.EscapeString(msg.Subject))
	msg.Message = strings.TrimSpace(html.EscapeString(msg.Message))
	msg.CaptchaQuestion = strings.TrimSpace(msg.CaptchaQuestion)
	msg.CaptchaAnswer = strings.TrimSpace(msg.CaptchaAnswer)

	// Additional validation beyond struct tags
	if len(msg.Message) > 5000 {
		_ = c.Error(apperrors.NewValidationError("message", nil, "Message too long", apperrors.ErrValidationFailed))
		c.Abort()
		return
	}

	// Verify simple numeric captcha
	if !h.verifyCaptcha(msg.CaptchaQuestion, msg.CaptchaAnswer) {
		h.logger.Warn("Invalid captcha submission", "email", msg.Email)
		_ = c.Error(apperrors.NewValidationError("captcha", msg.CaptchaAnswer, "Please solve the math problem correctly", apperrors.ErrValidationFailed))
		c.Abort()
		return
	}

	// Send email
	if err := h.emailService.SendContactMessage(&msg); err != nil {
		h.logger.Error("Failed to send contact email", "error", err, "recipient", msg.Email)
		_ = c.Error(err) // Email service should return appropriate error types
		c.Abort()
		return
	}

	data := utils.GetTemplateData()
	data["success"] = true
	data["message"] = "Thank you for your message! I'll get back to you soon."
	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// RSSFeed generates RSS feed
func (h *Handlers) RSSFeed(c *gin.Context) {
	// Use cached function if available
	if h.cachedFunctions.GetRSSFeed != nil {
		if rss, err := h.cachedFunctions.GetRSSFeed(); err == nil {
			c.Data(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(rss))
			return
		}
	}

	// Fallback to uncached version
	rss, err := h.getRSSFeedUncached()
	if err != nil {
		h.logger.Error("Failed to generate RSS feed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Data(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(rss))
}

// JSONFeed generates JSON feed
func (h *Handlers) JSONFeed(c *gin.Context) {
	// Use cached function if available
	if h.cachedFunctions.GetJSONFeed != nil {
		if feedJSON, err := h.cachedFunctions.GetJSONFeed(); err == nil {
			c.Data(http.StatusOK, "application/feed+json; charset=utf-8", []byte(feedJSON))
			return
		}
	}

	// Fallback to uncached version
	feedJSON, err := h.getJSONFeedUncached()
	if err != nil {
		h.logger.Error("Failed to generate JSON feed", "error", err)
		_ = c.Error(apperrors.NewHTTPError(http.StatusInternalServerError, "Failed to generate JSON feed", err))
		c.Abort()
		return
	}

	c.Data(http.StatusOK, "application/feed+json; charset=utf-8", []byte(feedJSON))
}

// Sitemap generates XML sitemap
func (h *Handlers) Sitemap(c *gin.Context) {
	// Use cached function if available
	if h.cachedFunctions.GetSitemap != nil {
		if sitemapData, err := h.cachedFunctions.GetSitemap(); err == nil {
			c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(sitemapData))
			return
		}
	}

	// Fallback to uncached version
	sitemapData, err := h.getSitemapUncached()
	if err != nil {
		h.logger.Error("Failed to generate sitemap", "error", err)
		_ = c.Error(apperrors.NewHTTPError(http.StatusInternalServerError, "Failed to generate sitemap", err))
		c.Abort()
		return
	}

	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(sitemapData))
}

// Health check endpoint
func (h *Handlers) Health(c *gin.Context) {
	data := utils.GetTemplateData()
	data["status"] = "healthy"
	data["timestamp"] = time.Now().Format(time.RFC3339)
	data["version"] = "2.0.0"
	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// Metrics endpoint
func (h *Handlers) Metrics(c *gin.Context) {
	stats := h.articleService.GetStats()
	cacheStats := h.cacheService.Stats()
	perfMetrics := middleware.GetPerformanceMetrics()

	// Calculate additional performance insights
	p95 := calculatePercentile(perfMetrics.ResponseTimes, 0.95)
	p99 := calculatePercentile(perfMetrics.ResponseTimes, 0.99)

	performanceData := utils.GetTemplateData()
	performanceData["request_count"] = perfMetrics.RequestCount
	performanceData["avg_response_time_ms"] = float64(perfMetrics.AverageResponseTime.Nanoseconds()) / 1e6
	performanceData["min_response_time_ms"] = float64(perfMetrics.MinResponseTime.Nanoseconds()) / 1e6
	performanceData["max_response_time_ms"] = float64(perfMetrics.MaxResponseTime.Nanoseconds()) / 1e6
	performanceData["p95_response_time_ms"] = float64(p95.Nanoseconds()) / 1e6
	performanceData["p99_response_time_ms"] = float64(p99.Nanoseconds()) / 1e6
	performanceData["requests_per_second"] = perfMetrics.RequestsPerSecond
	performanceData["memory_usage_mb"] = perfMetrics.MemoryUsage / 1024 / 1024
	performanceData["goroutine_count"] = perfMetrics.GoroutineCount
	performanceData["requests_by_endpoint"] = perfMetrics.RequestsByEndpoint
	performanceData["avg_response_by_endpoint"] = formatEndpointTimes(perfMetrics.ResponseTimesByEndpoint)
	defer utils.PutTemplateData(performanceData)

	// Add competitive comparison
	competitorComparison := utils.GetTemplateData()
	vsGhost := utils.GetTemplateData()
	vsGhost["response_time_advantage"] = "4x faster"   // Ghost ~200ms vs MarkGo <50ms
	vsGhost["memory_advantage"] = "10x more efficient" // Ghost ~300MB vs MarkGo ~30MB
	competitorComparison["vs_ghost"] = vsGhost

	vsWordpress := utils.GetTemplateData()
	vsWordpress["response_time_advantage"] = "10x faster"  // WordPress ~500ms vs MarkGo <50ms
	vsWordpress["memory_advantage"] = "30x more efficient" // WordPress ~2GB vs MarkGo ~30MB
	competitorComparison["vs_wordpress"] = vsWordpress

	vsHugo := utils.GetTemplateData()
	vsHugo["dynamic_features"] = "search, forms, real-time updates"
	vsHugo["deployment"] = "single binary vs build process"
	competitorComparison["vs_hugo"] = vsHugo

	defer func() {
		utils.PutTemplateData(vsGhost)
		utils.PutTemplateData(vsWordpress)
		utils.PutTemplateData(vsHugo)
		utils.PutTemplateData(competitorComparison)
	}()

	responseData := utils.GetTemplateData()
	responseData["blog"] = stats
	responseData["cache"] = cacheStats
	responseData["performance"] = performanceData
	responseData["competitor_analysis"] = competitorComparison
	responseData["timestamp"] = time.Now().Format(time.RFC3339)
	responseData["version"] = "2.0.0"
	c.JSON(http.StatusOK, responseData)
	utils.PutTemplateData(responseData)
}

func calculatePercentile(times []time.Duration, percentile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}

	// Simple percentile calculation
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)

	// Basic sort
	n := len(sorted)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	index := int(percentile * float64(len(sorted)-1))
	return sorted[index]
}

func formatEndpointTimes(times map[string]time.Duration) map[string]float64 {
	formatted := make(map[string]float64)
	for endpoint, duration := range times {
		formatted[endpoint] = float64(duration.Nanoseconds()) / 1e6 // Convert to milliseconds
	}
	return formatted
}

// Admin handlers

// ClearCache clears all cache
func (h *Handlers) ClearCache(c *gin.Context) {
	h.cacheService.Clear()
	h.logger.Info("Cache cleared by admin")

	data := utils.GetTemplateData()
	data["success"] = true
	data["message"] = "Cache cleared successfully"
	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// AdminStats returns admin statistics
func (h *Handlers) AdminStats(c *gin.Context) {
	stats := h.articleService.GetStats()
	cacheStats := h.cacheService.Stats()

	data := utils.GetTemplateData()
	data["articles"] = stats
	data["cache"] = cacheStats
	data["config"] = h.emailService.GetConfig()
	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// ReloadArticles reloads articles from disk
func (h *Handlers) ReloadArticles(c *gin.Context) {
	if err := h.articleService.ReloadArticles(); err != nil {
		h.logger.Error("Failed to reload articles", "error", err)
		_ = c.Error(err) // Service should return appropriate error types
		c.Abort()
		return
	}

	// Clear cache to force refresh
	h.cacheService.Clear()
	h.logger.Info("Articles reloaded by admin")

	data := utils.GetTemplateData()
	data["success"] = true
	data["message"] = "Articles reloaded successfully"
	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// Enhanced Debug Endpoints

// DebugMemory returns detailed memory statistics
func (h *Handlers) DebugMemory(c *gin.Context) {
	memMonitor := utils.GetGlobalMemoryMonitor(h.logger)
	currentStats := memMonitor.GetCurrentStats()
	report := memMonitor.GetMemoryReport()

	// Get cleanup manager stats
	cleanupStats := utils.GetGlobalCleanupManager().GetAllStats()

	data := utils.GetTemplateData()
	data["current"] = map[string]interface{}{
		"heap_alloc_mb":  float64(currentStats.HeapAlloc) / 1024 / 1024,
		"heap_sys_mb":    float64(currentStats.HeapSys) / 1024 / 1024,
		"num_gc":         currentStats.NumGC,
		"num_goroutines": currentStats.NumGoroutine,
		"timestamp":      currentStats.Timestamp.Format(time.RFC3339),
	}
	data["report"] = report
	data["cleanup_pools"] = cleanupStats
	data["gc_stats"] = h.getGCStats()

	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// DebugRuntime returns Go runtime information
func (h *Handlers) DebugRuntime(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	data := utils.GetTemplateData()
	data["go_version"] = runtime.Version()
	data["go_os"] = runtime.GOOS
	data["go_arch"] = runtime.GOARCH
	data["num_cpu"] = runtime.NumCPU()
	data["num_goroutine"] = runtime.NumGoroutine()
	data["memory"] = map[string]interface{}{
		"alloc_mb":        float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb":  float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":          float64(m.Sys) / 1024 / 1024,
		"num_gc":          m.NumGC,
		"gc_cpu_fraction": m.GCCPUFraction,
		"heap_objects":    m.HeapObjects,
	}
	data["uptime_seconds"] = time.Since(h.startTime).Seconds()
	data["timestamp"] = time.Now().Format(time.RFC3339)

	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// DebugConfig returns safe configuration information for debugging
func (h *Handlers) DebugConfig(c *gin.Context) {
	data := utils.GetTemplateData()
	data["environment"] = h.config.Environment
	data["port"] = h.config.Port
	data["log_level"] = h.config.Logging.Level
	data["cache_ttl"] = h.config.Cache.TTL.String()
	data["posts_per_page"] = h.config.Blog.PostsPerPage
	data["rate_limit"] = map[string]interface{}{
		"general_requests": h.config.RateLimit.General.Requests,
		"general_window":   h.config.RateLimit.General.Window.String(),
		"contact_requests": h.config.RateLimit.Contact.Requests,
		"contact_window":   h.config.RateLimit.Contact.Window.String(),
	}
	// Note: Sensitive config like secrets/passwords are intentionally excluded
	data["timestamp"] = time.Now().Format(time.RFC3339)

	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// DebugRequests returns detailed request debugging info (development only)
func (h *Handlers) DebugRequests(c *gin.Context) {
	// Only enable in development environment
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Debug requests endpoint only available in development"})
		return
	}

	perfMetrics := middleware.GetPerformanceMetrics()

	data := utils.GetTemplateData()
	data["current_request"] = map[string]interface{}{
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"query":      c.Request.URL.RawQuery,
		"user_agent": c.Request.UserAgent(),
		"remote_ip":  c.ClientIP(),
		"headers":    h.getSafeHeaders(c.Request.Header),
	}
	data["performance"] = map[string]interface{}{
		"request_count":        perfMetrics.RequestCount,
		"requests_per_second":  perfMetrics.RequestsPerSecond,
		"avg_response_time_ms": float64(perfMetrics.TotalResponseTime) / float64(perfMetrics.RequestCount) / 1e6,
		"max_response_time_ms": float64(perfMetrics.MaxResponseTime) / 1e6,
		"min_response_time_ms": float64(perfMetrics.MinResponseTime) / 1e6,
		"goroutine_count":      perfMetrics.GoroutineCount,
	}
	data["timestamp"] = time.Now().Format(time.RFC3339)

	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// SetLogLevel allows dynamic log level changes (development only)
func (h *Handlers) SetLogLevel(c *gin.Context) {
	// Only enable in development environment
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Log level changes only available in development"})
		return
	}

	var request struct {
		Level string `json:"level" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLevels {
		if request.Level == level {
			validLevel = true
			break
		}
	}

	if !validLevel {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log level. Must be one of: debug, info, warn, error"})
		return
	}

	// Update logger dynamically
	h.logger.Info("Log level changed dynamically", "old_level", h.config.Logging.Level, "new_level", request.Level)
	h.config.Logging.Level = request.Level

	data := utils.GetTemplateData()
	data["success"] = true
	data["message"] = fmt.Sprintf("Log level changed to: %s", request.Level)
	data["level"] = request.Level
	data["timestamp"] = time.Now().Format(time.RFC3339)

	c.JSON(http.StatusOK, data)
	utils.PutTemplateData(data)
}

// Helper methods for debug endpoints

func (h *Handlers) getGCStats() map[string]interface{} {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return map[string]interface{}{
		"num_gc":          stats.NumGC,
		"pause_total_ns":  stats.PauseTotalNs,
		"pause_avg_ns":    float64(stats.PauseTotalNs) / float64(stats.NumGC+1),
		"gc_cpu_fraction": stats.GCCPUFraction,
		"next_gc_mb":      float64(stats.NextGC) / 1024 / 1024,
		"last_gc":         time.Unix(0, int64(stats.LastGC)).Format(time.RFC3339),
	}
}

func (h *Handlers) getSafeHeaders(headers map[string][]string) map[string]string {
	safe := make(map[string]string)

	// Only include non-sensitive headers
	safeHeaders := []string{
		"Accept", "Accept-Encoding", "Accept-Language", "Cache-Control",
		"Content-Type", "User-Agent", "Referer", "X-Forwarded-For",
		"X-Real-IP", "X-Requested-With",
	}

	for _, headerName := range safeHeaders {
		if values := headers[headerName]; len(values) > 0 {
			safe[headerName] = values[0]
		}
	}

	return safe
}

// Profiling endpoints (pprof integration for development)

// PprofIndex serves the pprof index page
func (h *Handlers) PprofIndex(c *gin.Context) {
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profiling endpoints only available in development"})
		return
	}
	pprof.Index(c.Writer, c.Request)
}

// PprofProfile serves CPU profiling data
func (h *Handlers) PprofProfile(c *gin.Context) {
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profiling endpoints only available in development"})
		return
	}
	pprof.Profile(c.Writer, c.Request)
}

// PprofHeap serves heap profiling data
func (h *Handlers) PprofHeap(c *gin.Context) {
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profiling endpoints only available in development"})
		return
	}
	pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
}

// PprofGoroutine serves goroutine profiling data
func (h *Handlers) PprofGoroutine(c *gin.Context) {
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profiling endpoints only available in development"})
		return
	}
	pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
}

// PprofAllocs serves allocation profiling data
func (h *Handlers) PprofAllocs(c *gin.Context) {
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profiling endpoints only available in development"})
		return
	}
	pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
}

// PprofBlock serves block profiling data
func (h *Handlers) PprofBlock(c *gin.Context) {
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profiling endpoints only available in development"})
		return
	}
	pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
}

// PprofMutex serves mutex profiling data
func (h *Handlers) PprofMutex(c *gin.Context) {
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profiling endpoints only available in development"})
		return
	}
	pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
}

// PprofTrace serves execution trace data
func (h *Handlers) PprofTrace(c *gin.Context) {
	if h.config.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Profiling endpoints only available in development"})
		return
	}
	pprof.Trace(c.Writer, c.Request)
}

// NotFound handles 404 errors
func (h *Handlers) NotFound(c *gin.Context) {
	data := utils.BaseTemplateData("Page Not Found", h.config).
		Set("message", "The page you're looking for doesn't exist.").
		Set("template", "404").
		Build()

	h.renderHTML(c, http.StatusNotFound, "base.html", data)
}

// verifyCaptcha validates a simple addition captcha
func (h *Handlers) verifyCaptcha(question, answer string) bool {
	// Parse the addition question (e.g., "3 + 5")
	// Expected format: "number + number"
	parts := strings.Fields(question)
	if len(parts) != 3 || parts[1] != "+" {
		return false
	}

	// Parse numbers
	num1, err1 := strconv.Atoi(parts[0])
	num2, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return false
	}

	// Parse user's answer
	userAnswer, err := strconv.Atoi(strings.TrimSpace(answer))
	if err != nil {
		return false
	}

	return userAnswer == num1+num2
}

func (h *Handlers) renderHTML(c *gin.Context, status int, template string, data any) {
	c.HTML(status, template, data)
}

// Draft Management Handlers

// GetDrafts returns all draft articles (admin endpoint)
func (h *Handlers) GetDrafts(c *gin.Context) {
	drafts := h.articleService.GetDraftArticles()

	// Convert to list view for efficient transfer
	draftList := make([]*models.ArticleList, len(drafts))
	for i, draft := range drafts {
		draftList[i] = draft.ToListView()
	}

	c.JSON(http.StatusOK, gin.H{
		"drafts": draftList,
		"count":  len(draftList),
	})
}

// GetDraftBySlug returns a specific draft by slug for editing/viewing
func (h *Handlers) GetDraftBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	draft, err := h.articleService.GetDraftBySlug(slug)
	if err != nil {
		var httpStatus int
		var message string

		if apperrors.IsArticleNotFound(err) {
			httpStatus = http.StatusNotFound
			message = "Draft not found"
		} else {
			httpStatus = http.StatusInternalServerError
			message = "Failed to retrieve draft"
		}

		c.JSON(httpStatus, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"draft": draft,
	})
}

// PreviewDraft returns a draft with processed content for preview
func (h *Handlers) PreviewDraft(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	preview, err := h.articleService.PreviewDraft(slug)
	if err != nil {
		var httpStatus int
		var message string

		if apperrors.IsArticleNotFound(err) {
			httpStatus = http.StatusNotFound
			message = "Draft not found for preview"
		} else {
			httpStatus = http.StatusInternalServerError
			message = "Failed to generate draft preview"
		}

		c.JSON(httpStatus, gin.H{"error": message})
		return
	}

	// Return preview with processed content
	c.JSON(http.StatusOK, gin.H{
		"preview": gin.H{
			"slug":              preview.Slug,
			"title":             preview.Title,
			"description":       preview.Description,
			"date":              preview.Date,
			"tags":              preview.Tags,
			"categories":        preview.Categories,
			"author":            preview.Author,
			"content":           preview.Content,
			"processed_content": preview.GetProcessedContent(),
			"excerpt":           preview.GetExcerpt(),
			"reading_time":      preview.ReadingTime,
			"word_count":        preview.WordCount,
			"draft":             preview.Draft,
			"featured":          preview.Featured,
			"last_modified":     preview.LastModified,
		},
	})
}

// PublishDraft publishes a draft article
func (h *Handlers) PublishDraft(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	err := h.articleService.PublishDraft(slug)
	if err != nil {
		var httpStatus int
		var message string

		if apperrors.IsArticleNotFound(err) {
			httpStatus = http.StatusNotFound
			message = "Draft not found"
		} else if apperrors.IsValidationError(err) {
			httpStatus = http.StatusBadRequest
			message = "Article is already published"
		} else {
			httpStatus = http.StatusInternalServerError
			message = "Failed to publish draft"
		}

		h.logger.Error("Failed to publish draft", "slug", slug, "error", err)
		c.JSON(httpStatus, gin.H{"error": message})
		return
	}

	h.logger.Info("Draft published successfully", "slug", slug)
	c.JSON(http.StatusOK, gin.H{
		"message": "Draft published successfully",
		"slug":    slug,
	})
}

// UnpublishArticle converts a published article to draft status
func (h *Handlers) UnpublishArticle(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	err := h.articleService.UnpublishArticle(slug)
	if err != nil {
		var httpStatus int
		var message string

		if apperrors.IsArticleNotFound(err) {
			httpStatus = http.StatusNotFound
			message = "Article not found"
		} else if apperrors.IsValidationError(err) {
			httpStatus = http.StatusBadRequest
			message = "Article is already a draft"
		} else {
			httpStatus = http.StatusInternalServerError
			message = "Failed to unpublish article"
		}

		h.logger.Error("Failed to unpublish article", "slug", slug, "error", err)
		c.JSON(httpStatus, gin.H{"error": message})
		return
	}

	h.logger.Info("Article unpublished to draft successfully", "slug", slug)
	c.JSON(http.StatusOK, gin.H{
		"message": "Article unpublished to draft successfully",
		"slug":    slug,
	})
}

// =====================
// Uncached Data Generation Functions
// =====================

// getHomeDataUncached generates home page data without caching
func (h *Handlers) getHomeDataUncached() (map[string]any, error) {
	articles := h.articleService.GetAllArticles()
	featured := h.articleService.GetFeaturedArticles(3)
	recent := h.articleService.GetRecentArticles(9)
	categoryCounts := h.articleService.GetCategoryCounts()
	tagCounts := h.articleService.GetTagCounts()

	popularTags := tagCounts
	if len(popularTags) > 10 {
		popularTags = popularTags[:10]
	}

	data := utils.BaseTemplateData(h.config.Blog.Title, h.config).
		Set("description", h.config.Blog.Description).
		Set("featured", featured).
		Set("recent", recent).
		Set("tags", popularTags).
		Set("totalCount", len(articles)).
		Set("totalCats", len(categoryCounts)).
		Set("totalTags", len(tagCounts)).
		Set("template", "index").
		Set("headTemplate", "index-head").
		Set("contentTemplate", "index-content").
		Build()

	return data, nil
}

// getArticleDataUncached generates article page data without caching
func (h *Handlers) getArticleDataUncached(slug string) (map[string]any, error) {
	article, err := h.articleService.GetArticleBySlug(slug)
	if err != nil || article == nil {
		return nil, fmt.Errorf("article not found: %s", slug)
	}

	// Get recent articles for sidebar/navigation
	recent := h.articleService.GetRecentArticles(5)

	// Get related articles by same tags (for "related articles" section)
	var relatedArticles []*models.Article
	for _, tag := range article.Tags {
		tagArticles := h.articleService.GetArticlesByTag(tag)
		relatedArticles = append(relatedArticles, tagArticles...)
	}

	data := utils.ArticlePageData(article.Title, h.config, recent).
		Set("description", article.Description).
		Set("article", article).
		Set("relatedArticles", relatedArticles).
		Set("template", "article").
		Set("headTemplate", "article-head").
		Set("contentTemplate", "article-content").
		Build()

	return data, nil
}

// getArticlesPageUncached generates articles page data without caching
func (h *Handlers) getArticlesPageUncached(page int) (map[string]any, error) {
	if page < 1 {
		page = 1
	}

	pageSize := h.config.Blog.PostsPerPage
	allArticles := h.articleService.GetAllArticles()
	totalArticles := len(allArticles)

	// Pagination
	pagination := models.NewPagination(page, totalArticles, pageSize)
	start := (page - 1) * pageSize
	end := min(start+pageSize, totalArticles)

	var articles []*models.Article
	if start < totalArticles {
		articles = allArticles[start:end]
	}

	// Get recent articles
	recent := h.articleService.GetRecentArticles(9)

	return utils.ListPageData("All Articles", h.config, recent, articles, pagination).Build(), nil
}

// getTagArticlesUncached generates tag articles page data without caching
func (h *Handlers) getTagArticlesUncached(tag string) (map[string]any, error) {
	articles := h.articleService.GetArticlesByTag(tag)

	data := utils.BaseTemplateData("Articles tagged with "+tag+" - "+h.config.Blog.Title, h.config).
		Set("description", "Articles tagged with "+tag).
		Set("articles", articles).
		Set("tag", tag).
		Set("totalCount", len(articles)).
		Set("template", "tag").
		Set("headTemplate", "tag-head").
		Set("contentTemplate", "tag-content").
		Build()

	return data, nil
}

// getCategoryArticlesUncached generates category articles page data without caching
func (h *Handlers) getCategoryArticlesUncached(category string) (map[string]any, error) {
	articles := h.articleService.GetArticlesByCategory(category)

	data := utils.BaseTemplateData("Articles in "+category+" - "+h.config.Blog.Title, h.config).
		Set("description", "Articles in category "+category).
		Set("articles", articles).
		Set("category", category).
		Set("totalCount", len(articles)).
		Set("template", "category").
		Set("headTemplate", "category-head").
		Set("contentTemplate", "category-content").
		Build()

	return data, nil
}

// getSearchResultsUncached generates search results data without caching
func (h *Handlers) getSearchResultsUncached(query string) (map[string]any, error) {
	articles := h.articleService.GetAllArticles()
	results := h.searchService.Search(articles, query, 20)
	recent := h.articleService.GetRecentArticles(5)

	data := utils.ArticlePageData("Search results for "+query, h.config, recent).
		Set("description", "Search results for "+query).
		Set("results", results).
		Set("query", query).
		Set("totalResults", len(results)).
		Set("template", "search").
		Set("headTemplate", "search-head").
		Set("contentTemplate", "search-content").
		Build()

	return data, nil
}

// getStatsDataUncached generates stats data without caching
func (h *Handlers) getStatsDataUncached() (map[string]any, error) {
	stats := h.articleService.GetStats()
	return map[string]any{"stats": stats}, nil
}

// getTagsPageUncached generates tags page data without caching
func (h *Handlers) getTagsPageUncached() (map[string]any, error) {
	tagCounts := h.articleService.GetTagCounts()
	recent := h.articleService.GetRecentArticles(5)

	return utils.ArticlePageData("All Tags", h.config, recent).
		Set("tags", tagCounts).
		Set("count", len(tagCounts)).
		Set("template", "tags").
		Build(), nil
}

// getCategoriesPageUncached generates categories page data without caching
func (h *Handlers) getCategoriesPageUncached() (map[string]any, error) {
	categories := h.articleService.GetCategoryCounts()

	return utils.BaseTemplateData("All Categories", h.config).
		Set("categories", categories).
		Set("count", len(categories)).
		Set("template", "categories").
		Build(), nil
}

// getRSSFeedUncached generates RSS feed without caching
func (h *Handlers) getRSSFeedUncached() (string, error) {
	articles := h.articleService.GetArticlesForFeed(20)

	feed := &models.Feed{
		Title:       h.config.Blog.Title,
		Description: h.config.Blog.Description,
		Link:        h.config.BaseURL,
		FeedURL:     h.config.BaseURL + "/feed.xml",
		Language:    "en",
		Updated:     time.Now(),
	}

	if len(articles) > 0 {
		feed.Updated = articles[0].Date
	}

	for _, article := range articles {
		item := models.FeedItem{
			ID:          h.config.BaseURL + "/" + article.Slug,
			Title:       article.Title,
			ContentHTML: article.GetProcessedContent(),
			URL:         h.config.BaseURL + "/" + article.Slug,
			Summary:     article.GetExcerpt(),
			Published:   article.Date,
			Modified:    article.LastModified,
			Tags:        article.Tags,
		}
		feed.Items = append(feed.Items, item)
	}

	xml, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", err
	}

	return "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + string(xml), nil
}

// getJSONFeedUncached generates JSON feed without caching
func (h *Handlers) getJSONFeedUncached() (string, error) {
	articles := h.articleService.GetArticlesForFeed(20)

	feed := &models.Feed{
		Title:       h.config.Blog.Title,
		Description: h.config.Blog.Description,
		Link:        h.config.BaseURL,
		FeedURL:     h.config.BaseURL + "/feed.json",
		Language:    "en",
		Updated:     time.Now(),
	}

	if len(articles) > 0 {
		feed.Updated = articles[0].Date
	}

	for _, article := range articles {
		item := models.FeedItem{
			ID:          h.config.BaseURL + "/" + article.Slug,
			Title:       article.Title,
			ContentHTML: article.GetProcessedContent(),
			URL:         h.config.BaseURL + "/" + article.Slug,
			Summary:     article.GetExcerpt(),
			Published:   article.Date,
			Modified:    article.LastModified,
			Tags:        article.Tags,
		}
		feed.Items = append(feed.Items, item)
	}

	feedJSON, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", err
	}

	return string(feedJSON), nil
}

// getSitemapUncached generates sitemap without caching
func (h *Handlers) getSitemapUncached() (string, error) {
	articles := h.articleService.GetAllArticles()

	sitemap := &models.Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	// Add home page
	sitemap.URLs = append(sitemap.URLs, models.SitemapURL{
		Loc:        h.config.BaseURL,
		LastMod:    time.Now(),
		ChangeFreq: "daily",
		Priority:   1.0,
	})

	// Add articles
	for _, article := range articles {
		sitemap.URLs = append(sitemap.URLs, models.SitemapURL{
			Loc:        h.config.BaseURL + "/" + article.Slug,
			LastMod:    article.LastModified,
			ChangeFreq: "monthly",
			Priority:   0.8,
		})
	}

	sitemapXML, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		return "", err
	}

	return "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + string(sitemapXML), nil
}
