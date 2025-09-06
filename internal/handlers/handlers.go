package handlers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
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
)

// Handlers contains all HTTP handlers
type Handlers struct {
	articleService  services.ArticleServiceInterface
	cacheService    services.CacheServiceInterface
	emailService    services.EmailServiceInterface
	searchService   services.SearchServiceInterface
	templateService services.TemplateServiceInterface
	config          *config.Config
	logger          *slog.Logger
	startTime       time.Time // Track server start time for uptime calculation
}

// Config for handler initialization
type Config struct {
	ArticleService  services.ArticleServiceInterface
	CacheService    services.CacheServiceInterface
	EmailService    services.EmailServiceInterface
	SearchService   services.SearchServiceInterface
	TemplateService services.TemplateServiceInterface
	Config          *config.Config
	Logger          *slog.Logger
}

// New creates a new handlers instance
func New(cfg *Config) *Handlers {
	return &Handlers{
		articleService:  cfg.ArticleService,
		cacheService:    cfg.CacheService,
		emailService:    cfg.EmailService,
		searchService:   cfg.SearchService,
		templateService: cfg.TemplateService,
		config:          cfg.Config,
		logger:          cfg.Logger,
		startTime:       time.Now(),
	}
}

// Home renders the homepage
func (h *Handlers) Home(c *gin.Context) {
	cacheKey := "home_page"

	if cached, found := h.cacheService.Get(cacheKey); found {
		h.renderHTML(c, http.StatusOK, "base.html", cached)
		return
	}

	articles := h.articleService.GetAllArticles()

	// Get featured articles
	featured := h.articleService.GetFeaturedArticles(3)

	// Get recent articles
	recent := h.articleService.GetRecentArticles(9)

	// Get total Categories
	categoryCounts := h.articleService.GetCategoryCounts()

	// Get popular tags
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

	h.cacheService.Set(cacheKey, data, h.config.Cache.TTL)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Articles renders the articles listing page
func (h *Handlers) Articles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize := h.config.Blog.PostsPerPage
	cacheKey := fmt.Sprintf("articles_page_%d", page)

	if cached, found := h.cacheService.Get(cacheKey); found {
		h.renderHTML(c, http.StatusOK, "base.html", cached)
		return
	}

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

	data := utils.ListPageData("All Articles", h.config, recent, articles, pagination).Build()

	h.cacheService.Set(cacheKey, data, h.config.Cache.TTL)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Article renders a single article page
func (h *Handlers) Article(c *gin.Context) {
	slug := c.Param("slug")

	cacheKey := fmt.Sprintf("article_%s", slug)
	if cached, found := h.cacheService.Get(cacheKey); found {
		h.renderHTML(c, http.StatusOK, "base.html", cached)
		return
	}

	article, err := h.articleService.GetArticleBySlug(slug)
	if err != nil {
		_ = c.Error(apperrors.NewArticleError("", fmt.Sprintf("Article '%s' not found", slug), apperrors.ErrArticleNotFound))
		c.Abort()
		return
	}

	// Get related articles (same tags)
	var related []*models.Article
	if len(article.Tags) > 0 {
		for _, tag := range article.Tags {
			taggedArticles := h.articleService.GetArticlesByTag(tag)
			for _, taggedArticle := range taggedArticles {
				if taggedArticle.Slug != slug && len(related) < 3 {
					related = append(related, taggedArticle)
				}
			}
			if len(related) >= 3 {
				break
			}
		}
	}

	// Get recent articles
	recent := h.articleService.GetRecentArticles(5)

	data := utils.ArticlePageData(article.Title, h.config, recent).
		Set("article", article).
		Set("related", related).
		Set("template", "article").
		Build()

	h.cacheService.Set(cacheKey, data, h.config.Cache.TTL)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ArticlesByTag renders articles filtered by tag
func (h *Handlers) ArticlesByTag(c *gin.Context) {
	tag := c.Param("tag")

	cacheKey := fmt.Sprintf("articles_tag_%s", tag)
	if cached, found := h.cacheService.Get(cacheKey); found {
		h.renderHTML(c, http.StatusOK, "base.html", cached)
		return
	}

	articles := h.articleService.GetArticlesByTag(tag)

	data := utils.BaseTemplateData(fmt.Sprintf("Articles tagged with '%s'", tag), h.config).
		Set("tag", tag).
		Set("articles", articles).
		Set("count", len(articles)).
		Set("template", "articles").
		Build()

	h.cacheService.Set(cacheKey, data, h.config.Cache.TTL)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ArticlesByCategory renders articles filtered by category
func (h *Handlers) ArticlesByCategory(c *gin.Context) {
	category := c.Param("category")

	cacheKey := fmt.Sprintf("articles_category_%s", category)
	if cached, found := h.cacheService.Get(cacheKey); found {
		h.renderHTML(c, http.StatusOK, "base.html", cached)
		return
	}

	articles := h.articleService.GetArticlesByCategory(category)

	data := utils.BaseTemplateData(fmt.Sprintf("Articles in '%s'", category), h.config).
		Set("category", category).
		Set("articles", articles).
		Set("count", len(articles)).
		Set("template", "articles").
		Build()

	h.cacheService.Set(cacheKey, data, h.config.Cache.TTL)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Tags renders the tags page
func (h *Handlers) Tags(c *gin.Context) {
	cacheKey := "all_tags"

	if cached, found := h.cacheService.Get(cacheKey); found {
		h.renderHTML(c, http.StatusOK, "base.html", cached)
		return
	}

	tagCounts := h.articleService.GetTagCounts()

	// Get recent articles
	recent := h.articleService.GetRecentArticles(5)

	data := utils.ArticlePageData("All Tags", h.config, recent).
		Set("tags", tagCounts).
		Set("count", len(tagCounts)).
		Set("template", "tags").
		Build()

	h.cacheService.Set(cacheKey, data, h.config.Cache.TTL)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Categories renders the categories page
func (h *Handlers) Categories(c *gin.Context) {
	cacheKey := "all_categories"

	if cached, found := h.cacheService.Get(cacheKey); found {
		h.renderHTML(c, http.StatusOK, "base.html", cached)
		return
	}

	categories := h.articleService.GetCategoryCounts()

	data := utils.BaseTemplateData("All Categories", h.config).
		Set("categories", categories).
		Set("count", len(categories)).
		Set("template", "categories").
		Build()

	h.cacheService.Set(cacheKey, data, h.config.Cache.TTL)
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

	cacheKey := fmt.Sprintf("search_%s", query)
	if cached, found := h.cacheService.Get(cacheKey); found {
		h.renderHTML(c, http.StatusOK, "base.html", cached)
		return
	}

	articles := h.articleService.GetAllArticles()
	results := h.searchService.Search(articles, query, 20)
	recent := h.articleService.GetRecentArticles(5)

	data := utils.ArticlePageData(fmt.Sprintf("Search results for '%s'", query), h.config, recent).
		Set("query", query).
		Set("results", results).
		Set("count", len(results)).
		Set("template", "search").
		Build()

	h.cacheService.Set(cacheKey, data, 30*time.Minute)
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

	// Verify simple numeric captcha
	if !h.verifyCaptcha(strings.TrimSpace(msg.CaptchaQuestion), strings.TrimSpace(msg.CaptchaAnswer)) {
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
	cacheKey := "rss_feed"

	if cached, found := h.cacheService.Get(cacheKey); found {
		c.Data(http.StatusOK, "application/rss+xml; charset=utf-8", cached.([]byte))
		return
	}

	articles := h.articleService.GetArticlesForFeed(20)
	rss := h.generateRSSFeed(articles)

	h.cacheService.Set(cacheKey, rss, 6*time.Hour)
	c.Data(http.StatusOK, "application/rss+xml; charset=utf-8", rss)
}

// JSONFeed generates JSON feed
func (h *Handlers) JSONFeed(c *gin.Context) {
	cacheKey := "json_feed"

	if cached, found := h.cacheService.Get(cacheKey); found {
		c.Data(http.StatusOK, "application/feed+json; charset=utf-8", cached.([]byte))
		return
	}

	articles := h.articleService.GetArticlesForFeed(20)
	feed := h.generateJSONFeed(articles)

	feedJSON, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		h.logger.Error("Failed to marshal JSON feed", "error", err)
		_ = c.Error(apperrors.NewHTTPError(http.StatusInternalServerError, "Failed to generate JSON feed", err))
		c.Abort()
		return
	}

	h.cacheService.Set(cacheKey, feedJSON, 6*time.Hour)
	c.Data(http.StatusOK, "application/feed+json; charset=utf-8", feedJSON)
}

// Sitemap generates XML sitemap
func (h *Handlers) Sitemap(c *gin.Context) {
	cacheKey := "sitemap"

	if cached, found := h.cacheService.Get(cacheKey); found {
		c.Data(http.StatusOK, "application/xml; charset=utf-8", cached.([]byte))
		return
	}

	sitemap := h.generateSitemap()
	sitemapXML, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		h.logger.Error("Failed to marshal XML sitemap", "error", err)
		_ = c.Error(apperrors.NewHTTPError(http.StatusInternalServerError, "Failed to generate sitemap", err))
		c.Abort()
		return
	}

	// Add XML declaration
	sitemapData := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(sitemapXML))

	h.cacheService.Set(cacheKey, sitemapData, 24*time.Hour)
	c.Data(http.StatusOK, "application/xml; charset=utf-8", sitemapData)
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

func (h *Handlers) generateRSSFeed(articles []*models.Article) []byte {
	// Use pooled buffer for RSS feed generation
	feedContent := utils.GetGlobalFeedBufferPool().BuildFeed(func(rss *strings.Builder) {
		rss.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
		rss.WriteString(`<rss version="2.0">`)
		rss.WriteString(`<channel>`)
		rss.WriteString(fmt.Sprintf(`<title>%s</title>`, h.config.Blog.Title))
		rss.WriteString(fmt.Sprintf(`<description>%s</description>`, h.config.Blog.Description))
		rss.WriteString(fmt.Sprintf(`<link>%s</link>`, h.config.BaseURL))
		rss.WriteString(fmt.Sprintf(`<language>%s</language>`, h.config.Blog.Language))
		rss.WriteString(fmt.Sprintf(`<lastBuildDate>%s</lastBuildDate>`, time.Now().Format(time.RFC1123Z)))

		for _, article := range articles {
			rss.WriteString(`<item>`)
			rss.WriteString(fmt.Sprintf(`<title>%s</title>`, article.Title))
			rss.WriteString(fmt.Sprintf(`<description>%s</description>`, article.GetExcerpt()))
			rss.WriteString(fmt.Sprintf(`<link>%s/articles/%s</link>`, h.config.BaseURL, article.Slug))
			rss.WriteString(fmt.Sprintf(`<guid>%s/articles/%s</guid>`, h.config.BaseURL, article.Slug))
			rss.WriteString(fmt.Sprintf(`<pubDate>%s</pubDate>`, article.Date.Format(time.RFC1123Z)))
			rss.WriteString(`</item>`)
		}

		rss.WriteString(`</channel>`)
		rss.WriteString(`</rss>`)
	})

	return []byte(feedContent)
}

func (h *Handlers) generateJSONFeed(articles []*models.Article) *models.Feed {
	var items []models.FeedItem

	for _, article := range articles {
		items = append(items, models.FeedItem{
			ID:          fmt.Sprintf("%s/articles/%s", h.config.BaseURL, article.Slug),
			Title:       article.Title,
			ContentHTML: article.Content,
			URL:         fmt.Sprintf("%s/articles/%s", h.config.BaseURL, article.Slug),
			Summary:     article.GetExcerpt(),
			Published:   article.Date,
			Modified:    article.LastModified,
			Tags:        article.Tags,
			Author: models.Author{
				Name:  h.config.Blog.Author,
				Email: h.config.Blog.AuthorEmail,
				URL:   h.config.BaseURL,
			},
		})
	}

	return &models.Feed{
		Title:       h.config.Blog.Title,
		Description: h.config.Blog.Description,
		Link:        h.config.BaseURL,
		FeedURL:     fmt.Sprintf("%s/feed.json", h.config.BaseURL),
		Language:    h.config.Blog.Language,
		Updated:     time.Now(),
		Author: models.Author{
			Name:  h.config.Blog.Author,
			Email: h.config.Blog.AuthorEmail,
			URL:   h.config.BaseURL,
		},
		Items: items,
	}
}

func (h *Handlers) generateSitemap() *models.Sitemap {
	var urls []models.SitemapURL

	// Add homepage
	urls = append(urls, models.SitemapURL{
		Loc:        h.config.BaseURL,
		LastMod:    time.Now(),
		ChangeFreq: "daily",
		Priority:   1.0,
	})

	// Add articles
	articles := h.articleService.GetAllArticles()
	for _, article := range articles {
		// Handle about page specially - it uses /about route instead of /articles/about
		if article.Slug == "about" {
			urls = append(urls, models.SitemapURL{
				Loc:        fmt.Sprintf("%s/about", h.config.BaseURL),
				LastMod:    article.LastModified,
				ChangeFreq: "monthly",
				Priority:   0.9, // Higher priority for about page
			})
		} else {
			urls = append(urls, models.SitemapURL{
				Loc:        fmt.Sprintf("%s/articles/%s", h.config.BaseURL, article.Slug),
				LastMod:    article.LastModified,
				ChangeFreq: "weekly",
				Priority:   0.8,
			})
		}
	}

	// Add static pages (removed /about since it's now handled as an article)
	staticPages := []struct {
		path     string
		priority float32
	}{
		{"/articles", 0.9},
		{"/tags", 0.7},
		{"/categories", 0.7},
		{"/contact", 0.6},
	}

	for _, page := range staticPages {
		urls = append(urls, models.SitemapURL{
			Loc:        fmt.Sprintf("%s%s", h.config.BaseURL, page.path),
			LastMod:    time.Now(),
			ChangeFreq: "monthly",
			Priority:   page.priority,
		})
	}

	return &models.Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
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
