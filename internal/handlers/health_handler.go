package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/constants"
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
		"display":          "standalone",
		"background_color": "#ffffff",
		"theme_color":      themeColor,
		"orientation":      "portrait",
		"icons": []gin.H{
			{"src": "/static/img/favicon-196x196.png", "sizes": "196x196", "type": "image/png"},
			{"src": "/static/img/mstile-310x310.png", "sizes": "310x310", "type": "image/png"},
		},
		"lang": blog.Language,
	}

	c.JSON(http.StatusOK, manifest)
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
