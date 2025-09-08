package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/markgo/internal/utils"
)

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

// Metrics handles the metrics endpoint for performance monitoring
func (h *AdminHandler) Metrics(c *gin.Context) {
	if !h.requireDevelopmentEnv(c) {
		return
	}

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
	if !h.requireDevelopmentEnv(c) {
		return
	}

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
			"alloc":           utils.FormatBytes(m.Alloc),
			"total_alloc":     utils.FormatBytes(m.TotalAlloc),
			"sys":             utils.FormatBytes(m.Sys),
			"heap_alloc":      utils.FormatBytes(m.HeapAlloc),
			"heap_sys":        utils.FormatBytes(m.HeapSys),
			"heap_idle":       utils.FormatBytes(m.HeapIdle),
			"heap_inuse":      utils.FormatBytes(m.HeapInuse),
			"heap_released":   utils.FormatBytes(m.HeapReleased),
			"heap_objects":    m.HeapObjects,
			"stack_inuse":     utils.FormatBytes(m.StackInuse),
			"stack_sys":       utils.FormatBytes(m.StackSys),
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

	data := h.buildBaseTemplateData("Debug Information").
		Set("debug", debugInfo).
		Set("template", "debug").
		Build()

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
			"before_alloc": utils.FormatBytes(before.Alloc),
			"after_alloc":  utils.FormatBytes(after.Alloc),
			"freed":        utils.FormatBytes(before.Alloc - after.Alloc),
		},
		"timestamp": time.Now().Unix(),
	})
}

// Profile handlers for pprof (development only)
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
			"memory_alloc": utils.FormatBytes(m.Alloc),
			"memory_sys":   utils.FormatBytes(m.Sys),
		},
		"config": map[string]any{
			"environment":   h.config.Environment,
			"cache_enabled": h.config.Cache.TTL > 0,
		},
		"timestamp": time.Now().Unix(),
	}, nil
}
