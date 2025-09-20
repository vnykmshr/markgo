// Package handlers provides HTTP request handlers for the MarkGo blog engine.
// It includes handlers for admin operations, article management, API endpoints, and more.
package handlers

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/http/pprof"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/services"
)

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// CachedAdminFunctions holds cached functions for admin operations
type CachedAdminFunctions struct {
	GetStatsData func() (map[string]any, error)
}

// AdminHandler handles administrative HTTP requests
type AdminHandler struct {
	*BaseHandler
	articleService  services.ArticleServiceInterface
	startTime       time.Time
	cachedFunctions CachedAdminFunctions
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	config *config.Config,
	logger *slog.Logger,
	templateService services.TemplateServiceInterface,
	articleService services.ArticleServiceInterface,
	startTime time.Time,
	cachedFunctions CachedAdminFunctions,
) *AdminHandler {
	return &AdminHandler{
		BaseHandler:     NewBaseHandler(config, logger, templateService),
		articleService:  articleService,
		startTime:       startTime,
		cachedFunctions: cachedFunctions,
	}
}

// formatDuration formats a time duration into human-readable string
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := math.Mod(d.Seconds(), 60)

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %.0fs", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %.0fs", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %.2fs", minutes, seconds)
	}
	return fmt.Sprintf("%.2fs", seconds)
}

// AdminHome handles the admin dashboard/home page
func (h *AdminHandler) AdminHome(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startTime)

	// Get both published and draft articles for accurate counting
	publishedArticles := h.articleService.GetAllArticles()
	draftArticles := h.articleService.GetDraftArticles()

	publishedCount := len(publishedArticles)
	draftCount := len(draftArticles)
	totalArticles := publishedCount + draftCount

	h.logger.Debug("Article counts",
		"published", publishedCount,
		"drafts", draftCount,
		"total", totalArticles)

	adminRoutes := []map[string]any{
		{
			"name":        "Statistics",
			"url":         "/admin/stats",
			"method":      "GET",
			"description": "View detailed blog statistics and analytics",
			"icon":        "📊",
		},
		{
			"name":        "Clear Cache",
			"url":         "/admin/cache/clear",
			"method":      "POST",
			"description": "Clear all cached data to force refresh",
			"icon":        "🗑️",
		},
		{
			"name":        "Reload Articles",
			"url":         "/admin/articles/reload",
			"method":      "POST",
			"description": "Reload all articles from disk",
			"icon":        "🔄",
		},
		{
			"name":        "Draft Articles",
			"url":         "/admin/drafts",
			"method":      "GET",
			"description": "View and manage draft articles",
			"icon":        "📝",
		},
		{
			"name":        "Preview Sessions",
			"url":         "/api/preview/sessions",
			"method":      "GET",
			"description": "View active preview sessions and statistics",
			"icon":        "👁️",
		},
		{
			"name":        "System Metrics",
			"url":         "/metrics",
			"method":      "GET",
			"description": "View system performance metrics",
			"icon":        "⚡",
		},
		{
			"name":        "Health Check",
			"url":         "/health",
			"method":      "GET",
			"description": "Check application health status",
			"icon":        "❤️",
		},
	}

	// Get tag and category counts for additional metrics
	tagCounts := h.articleService.GetTagCounts()
	categoryCounts := h.articleService.GetCategoryCounts()

	systemInfo := map[string]any{
		"uptime":             formatDuration(uptime),
		"uptime_seconds":     int64(uptime.Seconds()),
		"go_version":         runtime.Version(),
		"environment":        h.config.Environment,
		"memory_usage":       formatBytes(m.Alloc),
		"memory_sys":         formatBytes(m.Sys),
		"goroutines":         runtime.NumGoroutine(),
		"articles_total":     totalArticles,
		"articles_published": publishedCount,
		"articles_drafts":    draftCount,
		"tags_total":         len(tagCounts),
		"categories_total":   len(categoryCounts),
		"gc_runs":            m.NumGC,
		"cpu_count":          runtime.NumCPU(),
	}

	if h.shouldReturnJSON(c) {
		c.JSON(http.StatusOK, gin.H{
			"title":       "MarkGo Admin",
			"system_info": systemInfo,
			"routes":      adminRoutes,
			"timestamp":   time.Now().Unix(),
		})
		return
	}

	data := h.buildBaseTemplateData("Admin Dashboard - " + h.config.Blog.Title)
	data["description"] = "Admin dashboard for " + h.config.Blog.Title
	data["system_info"] = systemInfo
	data["admin_routes"] = adminRoutes
	data["template"] = "admin_home"

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Metrics handles the metrics endpoint for performance monitoring
func (h *AdminHandler) Metrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startTime)

	h.respondWithJSON(c, http.StatusOK, func() map[string]any {
		return map[string]any{
			"uptime":         uptime.String(),
			"uptime_seconds": int64(uptime.Seconds()),
			"memory": map[string]any{
				"alloc":       m.Alloc,
				"total_alloc": m.TotalAlloc,
				"sys":         m.Sys,
				"heap_alloc":  m.HeapAlloc,
				"heap_sys":    m.HeapSys,
				"num_gc":      m.NumGC,
			},
			"goroutines":  runtime.NumGoroutine(),
			"go_version":  runtime.Version(),
			"environment": h.config.Environment,
			"timestamp":   time.Now().Unix(),
		}
	})
}

// Stats handles the stats endpoint with cached/fallback pattern
func (h *AdminHandler) Stats(c *gin.Context) {
	h.withCachedFallback(c,
		h.cachedFunctions.GetStatsData,
		h.getStatsDataUncached,
		"admin_stats.html", // This would be JSON for API endpoint
		"Failed to get stats")
}

// Debug handles debug information endpoint
func (h *AdminHandler) Debug(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	debugInfo := map[string]any{
		"runtime": map[string]any{
			"go_version":    runtime.Version(),
			"num_cpu":       runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
			"compiler":      runtime.Compiler,
			"arch":          runtime.GOARCH,
			"os":            runtime.GOOS,
		},
		"memory": map[string]any{
			"alloc":           formatBytes(m.Alloc),
			"total_alloc":     formatBytes(m.TotalAlloc),
			"sys":             formatBytes(m.Sys),
			"heap_alloc":      formatBytes(m.HeapAlloc),
			"heap_sys":        formatBytes(m.HeapSys),
			"heap_idle":       formatBytes(m.HeapIdle),
			"heap_inuse":      formatBytes(m.HeapInuse),
			"heap_released":   formatBytes(m.HeapReleased),
			"heap_objects":    m.HeapObjects,
			"stack_inuse":     formatBytes(m.StackInuse),
			"stack_sys":       formatBytes(m.StackSys),
			"num_gc":          m.NumGC,
			"gc_cpu_fraction": m.GCCPUFraction,
		},
		"config": map[string]any{
			"environment":    h.config.Environment,
			"port":           h.config.Port,
			"base_url":       h.config.BaseURL,
			"cache_ttl":      h.config.Cache.TTL.String(),
			"cache_max_size": h.config.Cache.MaxSize,
		},
		"uptime": time.Since(h.startTime).String(),
	}

	if h.shouldReturnJSON(c) {
		c.JSON(http.StatusOK, debugInfo)
		return
	}

	data := h.buildBaseTemplateData("Debug Information")
	data["debug"] = debugInfo
	data["template"] = "debug"

	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ReloadArticles handles reloading articles (development only)
func (h *AdminHandler) ReloadArticles(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}

	h.logger.Info("Reloading articles requested")

	if err := h.articleService.ReloadArticles(); err != nil {
		h.handleError(c, err, "Failed to reload articles")
		return
	}

	h.logger.Info("Articles reloaded successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":   "Articles reloaded successfully",
		"timestamp": time.Now().Unix(),
	})
}

// CompactMemory handles manual memory compaction (development only)
func (h *AdminHandler) CompactMemory(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	// Force garbage collection
	runtime.GC()
	runtime.GC() // Call twice to ensure cleanup

	// Force garbage collection (article service may not have CompactMemory method)
	runtime.GC()

	runtime.ReadMemStats(&after)

	h.logger.Info("Memory compaction completed",
		"before_alloc", before.Alloc,
		"after_alloc", after.Alloc,
		"freed", before.Alloc-after.Alloc)

	c.JSON(http.StatusOK, gin.H{
		"message": "Memory compaction completed",
		"memory": map[string]any{
			"before_alloc": formatBytes(before.Alloc),
			"after_alloc":  formatBytes(after.Alloc),
			"freed":        formatBytes(before.Alloc - after.Alloc),
		},
		"timestamp": time.Now().Unix(),
	})
}

// ProfileIndex handles pprof profile index (development only)
func (h *AdminHandler) ProfileIndex(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Index(c.Writer, c.Request)
}

func (h *AdminHandler) ProfileCmdline(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Cmdline(c.Writer, c.Request)
}

func (h *AdminHandler) ProfileProfile(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Profile(c.Writer, c.Request)
}

func (h *AdminHandler) ProfileSymbol(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Symbol(c.Writer, c.Request)
}

func (h *AdminHandler) ProfileTrace(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Trace(c.Writer, c.Request)
}

func (h *AdminHandler) ProfileHeap(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
}

func (h *AdminHandler) ProfileGoroutine(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
}

func (h *AdminHandler) ProfileBlock(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
}

func (h *AdminHandler) ProfileMutex(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
}

func (h *AdminHandler) ProfileAllocs(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}
	pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
}

// SetLogLevel handles dynamic log level changes
func (h *AdminHandler) SetLogLevel(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}

	level := c.PostForm("level")
	if level == "" {
		level = c.Query("level")
	}

	if level == "" {
		c.JSON(400, gin.H{
			"error":   "Missing log level parameter",
			"message": "Please provide 'level' parameter with value: debug, info, warn, error",
			"example": "POST /admin/log-level with level=debug or GET /admin/log-level?level=info",
		})
		return
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	level = strings.ToLower(strings.TrimSpace(level))
	if !validLevels[level] {
		c.JSON(400, gin.H{
			"error":    "Invalid log level",
			"message":  "Log level must be one of: debug, info, warn, error",
			"provided": level,
		})
		return
	}

	// Note: slog doesn't support runtime level changes natively
	// This would require a more complex implementation with a custom handler
	h.logger.Info("Log level change requested",
		"requested_level", level,
		"note", "Runtime log level changes require service restart")

	c.JSON(200, gin.H{
		"message":         "Log level change requested",
		"requested_level": level,
		"note":            "Runtime log level changes are not currently supported. Please update config and restart service.",
		"current_status":  "Request logged for future implementation",
	})
}

// Uncached data generation methods

func (h *AdminHandler) getStatsDataUncached() (map[string]any, error) {
	stats := h.articleService.GetStats()
	if stats == nil {
		return nil, fmt.Errorf("failed to get article stats")
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"articles": map[string]any{
			"total":      stats.TotalArticles,
			"published":  stats.PublishedCount,
			"drafts":     stats.DraftCount,
			"tags":       stats.TotalTags,
			"categories": stats.TotalCategories,
		},
		"system": map[string]any{
			"uptime":       time.Since(h.startTime).String(),
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": formatBytes(m.Alloc),
			"memory_sys":   formatBytes(m.Sys),
		},
		"config": map[string]any{
			"environment":   h.config.Environment,
			"cache_enabled": h.config.Cache.TTL > 0,
		},
		"timestamp": time.Now().Unix(),
	}, nil
}
