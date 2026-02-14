package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/1mb-dev/markgo/internal/constants"
)

// HealthHandler handles health check and metrics requests.
type HealthHandler struct {
	*BaseHandler
	startTime time.Time
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(base *BaseHandler, startTime time.Time) *HealthHandler {
	return &HealthHandler{
		BaseHandler: base,
		startTime:   startTime,
	}
}

// themeColors maps Blog.Theme presets to their primary hex color.
var themeColors = map[string]string{
	"default": "#2563eb",
	"ocean":   "#0891b2",
	"forest":  "#059669",
	"sunset":  "#ea580c",
	"berry":   "#9333ea",
}

// Manifest serves a dynamic web app manifest generated from config.
func (h *HealthHandler) Manifest(c *gin.Context) {
	blog := h.config.Blog

	themeColor := themeColors[blog.Theme]
	if themeColor == "" {
		themeColor = themeColors["default"]
	}

	manifest := gin.H{
		"name":             blog.Title,
		"short_name":       blog.Title,
		"description":      blog.Description,
		"start_url":        "/",
		"scope":            "/",
		"display":          "standalone",
		"background_color": "#ffffff",
		"theme_color":      themeColor,
		"orientation":      "portrait",
		"icons": []gin.H{
			{"src": "/static/img/icon-192x192.png", "sizes": "192x192", "type": "image/png", "purpose": "any"},
			{"src": "/static/img/icon-512x512.png", "sizes": "512x512", "type": "image/png", "purpose": "any maskable"},
		},
		"categories": []string{"blog", "news", "writing"},
		"shortcuts": []gin.H{
			{"name": "Search", "url": "/search", "description": "Search posts"},
			{"name": "Feed", "url": "/", "description": "Latest posts"},
		},
		"share_target": gin.H{
			"action":  "/compose",
			"method":  "GET",
			"enctype": "application/x-www-form-urlencoded",
			"params": gin.H{
				"title": "title",
				"text":  "text",
				"url":   "url",
			},
		},
		"lang": blog.Language,
	}

	c.JSON(http.StatusOK, manifest)
}

// Offline renders the offline fallback page (precached by the Service Worker).
func (h *HealthHandler) Offline(c *gin.Context) {
	data := h.buildBaseTemplateData("Offline - " + h.config.Blog.Title)
	data["template"] = "offline"
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Health handles health check requests.
func (h *HealthHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)

	health := map[string]any{
		"status":      "healthy",
		"timestamp":   time.Now().Unix(),
		"uptime":      uptime.String(),
		"version":     constants.AppVersion,
		"environment": h.config.Environment,
		"services": map[string]any{
			"articles": "healthy",
			"cache":    "healthy",
		},
	}

	c.JSON(http.StatusOK, health)
}
