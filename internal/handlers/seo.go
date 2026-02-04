package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/services"
)

// SEOHandler handles SEO-related HTTP endpoints
type SEOHandler struct {
	seoService services.SEOServiceInterface
	logger     *slog.Logger
}

// NewSEOHandler creates a new SEO handler instance
func NewSEOHandler(seoService services.SEOServiceInterface, logger *slog.Logger) *SEOHandler {
	return &SEOHandler{
		seoService: seoService,
		logger:     logger,
	}
}

// ServeSitemap serves the XML sitemap (generated on-demand)
func (h *SEOHandler) ServeSitemap(c *gin.Context) {
	if !h.seoService.IsEnabled() {
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusNotFound, "Sitemap not available")
		return
	}

	sitemap, err := h.seoService.GenerateSitemap()
	if err != nil {
		h.logger.Error("Failed to generate sitemap", "error", err)
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusInternalServerError, "Failed to generate sitemap")
		return
	}

	// Set appropriate headers for XML sitemap
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	c.Header("Last-Modified", time.Now().Format(http.TimeFormat))

	c.Data(http.StatusOK, "application/xml; charset=utf-8", sitemap)
}

// ServeRobotsTxt serves the robots.txt file
func (h *SEOHandler) ServeRobotsTxt(c *gin.Context) {
	if !h.seoService.IsEnabled() {
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusNotFound, "robots.txt not available")
		return
	}

	robotsTxt, err := h.seoService.GenerateRobotsTxt()
	if err != nil {
		h.logger.Error("Failed to generate robots.txt", "error", err)
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusInternalServerError, "Failed to generate robots.txt")
		return
	}

	// Set appropriate headers
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	c.Data(http.StatusOK, "text/plain; charset=utf-8", robotsTxt)
}

// AnalyzeContent performs SEO analysis on provided content (admin endpoint)
func (h *SEOHandler) AnalyzeContent(c *gin.Context) {
	if !h.seoService.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "SEO service not enabled",
		})
		return
	}

	var request struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	analysis, err := h.seoService.AnalyzeContent(request.Content)
	if err != nil {
		h.logger.Error("Failed to analyze content", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze content",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"analysis": analysis,
	})
}

// RegisterSEORoutes registers all SEO-related routes
func RegisterSEORoutes(router *gin.Engine, handler *SEOHandler) {
	// Public SEO endpoints
	router.GET("/sitemap.xml", handler.ServeSitemap)
	router.GET("/robots.txt", handler.ServeRobotsTxt)

	// Admin SEO endpoints (should be protected by auth middleware)
	admin := router.Group("/admin/seo")
	admin.POST("/analyze", handler.AnalyzeContent)
}
