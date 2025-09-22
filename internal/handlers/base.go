// Package handlers provides HTTP request handlers for the MarkGo blog engine.
// It includes handlers for admin operations, article management, API endpoints, and more.
package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	config          *config.Config
	logger          *slog.Logger
	templateService services.TemplateServiceInterface
	buildInfo       *BuildInfo
	seoService      services.SEOServiceInterface
	seoHelper       *SEODataHelper
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(
	config *config.Config,
	logger *slog.Logger,
	templateService services.TemplateServiceInterface,
	buildInfo *BuildInfo,
	seoService services.SEOServiceInterface,
) *BaseHandler {
	var seoHelper *SEODataHelper
	if seoService != nil {
		seoHelper = NewSEODataHelper(seoService, config)
	}

	return &BaseHandler{
		config:          config,
		logger:          logger,
		templateService: templateService,
		buildInfo:       buildInfo,
		seoService:      seoService,
		seoHelper:       seoHelper,
	}
}

// withCachedFallback provides the cached/fallback pattern for map[string]any data
func (h *BaseHandler) withCachedFallback(
	c *gin.Context,
	cachedFunc func() (map[string]any, error),
	uncachedFunc func() (map[string]any, error),
	template string,
	errorMsg string,
) {
	// Try cached function first
	if cachedFunc != nil {
		if data, err := cachedFunc(); err == nil {
			// Enhance with SEO data before rendering
			h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
			h.renderHTML(c, http.StatusOK, template, data)
			return
		}
		// Log cache miss but don't fail - fallback to uncached
		h.logger.Debug("Cache miss, falling back to uncached", "template", template)
	}

	// Fallback to uncached
	data, err := uncachedFunc()
	if err != nil {
		h.handleError(c, err, errorMsg)
		return
	}

	// Enhance with SEO data before rendering
	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, template, data)
}

// withCachedStringFallback provides the cached/fallback pattern for string data (RSS, JSON, etc.)
func (h *BaseHandler) withCachedStringFallback(
	c *gin.Context,
	cachedFunc func() (string, error),
	uncachedFunc func() (string, error),
	contentType string,
	errorMsg string,
) {
	// Try cached function first
	if cachedFunc != nil {
		if data, err := cachedFunc(); err == nil && data != "" {
			c.Header("Content-Type", contentType)
			c.String(http.StatusOK, data)
			return
		}
		// Log cache miss but don't fail - fallback to uncached
		h.logger.Debug("Cache miss, falling back to uncached", "content_type", contentType)
	}

	// Fallback to uncached
	data, err := uncachedFunc()
	if err != nil {
		h.handleError(c, err, errorMsg)
		return
	}

	c.Header("Content-Type", contentType)
	c.String(http.StatusOK, data)
}

// enhanceWithSEO adds SEO data to template data
func (h *BaseHandler) enhanceWithSEO(data map[string]any, seoType string) {
	if h.seoHelper == nil || h.seoService == nil || !h.seoService.IsEnabled() {
		return
	}

	var seoData map[string]interface{}

	switch seoType {
	case "article":
		if article, ok := data["article"].(*models.Article); ok && article != nil {
			seoData = h.seoHelper.GenerateArticleSEOData(article)
		}
	case "home":
		seoData = h.seoHelper.GenerateHomeSEOData()
	case "page":
		title, _ := data["title"].(string)
		description, _ := data["description"].(string)
		path := ""
		if pageData, ok := data["path"].(string); ok {
			path = pageData
		}
		seoData = h.seoHelper.GeneratePageSEOData(title, description, path)
	}

	// Merge SEO data into template data
	if seoData != nil {
		h.seoHelper.AddSEODataToTemplateData(data, seoData)
	}
}

// enhanceTemplateDataWithSEO intelligently adds SEO data based on page context
func (h *BaseHandler) enhanceTemplateDataWithSEO(data map[string]any, path string) {
	if h.seoHelper == nil || h.seoService == nil || !h.seoService.IsEnabled() {
		return
	}

	// Determine SEO type based on data content and path
	seoType := "page" // default

	// Check if this is an article page
	if article, ok := data["article"].(*models.Article); ok && article != nil {
		seoType = "article"
	} else if path == "/" || path == "" {
		seoType = "home"
	}

	h.enhanceWithSEO(data, seoType)
}

// renderHTML renders HTML template with common error handling
func (h *BaseHandler) renderHTML(c *gin.Context, status int, template string, data any) {
	if h.shouldReturnJSON(c) {
		c.JSON(status, data)
		return
	}

	if err := h.templateService.Render(c.Writer, template, data); err != nil {
		h.logger.Error("Template rendering failed", "template", template, "error", err)
		h.handleError(c, err, "Template rendering failed")
		return
	}
}

// handleError provides standardized error handling across handlers
func (h *BaseHandler) handleError(c *gin.Context, err error, defaultMsg string) {
	var httpStatus int
	var message string

	switch {
	case err != nil && err.Error() == "article not found":
		httpStatus = http.StatusNotFound
		message = "Resource not found"
	case apperrors.IsValidationError(err):
		httpStatus = http.StatusBadRequest
		message = "Invalid request"
	case apperrors.IsConfigurationError(err):
		httpStatus = http.StatusServiceUnavailable
		message = "Service temporarily unavailable"
	default:
		httpStatus = http.StatusInternalServerError
		message = defaultMsg
	}

	h.logger.Error("Handler error", "error", err, "status", httpStatus)

	if h.shouldReturnJSON(c) {
		c.JSON(httpStatus, gin.H{
			"error":   message,
			"status":  httpStatus,
			"details": err.Error(),
		})
		return
	}

	// For HTML responses, use the error handler middleware
	//nolint:errcheck // Ignore error: adding to gin context errors is non-critical
	_ = c.Error(apperrors.NewHTTPError(httpStatus, message, err))
	c.Abort()
}

// shouldReturnJSON determines if response should be JSON based on request
func (h *BaseHandler) shouldReturnJSON(c *gin.Context) bool {
	// Check Accept header
	accept := c.GetHeader("Accept")
	if accept == "application/json" || accept == "application/*" || accept == "*/*" {
		return true
	}

	// Check for API paths
	path := c.Request.URL.Path
	apiPaths := []string{"/health", "/metrics", "/api/", "/admin/stats"}
	for _, apiPath := range apiPaths {
		if path == apiPath || (len(path) > len(apiPath) && path[:len(apiPath)] == apiPath) {
			return true
		}
	}

	return false
}

// requireDevelopmentEnv checks if we're in development environment
func (h *BaseHandler) requireDevelopmentEnv(c *gin.Context) bool {
	if h.config.Environment != "development" {
		h.handleError(c, apperrors.NewHTTPError(
			http.StatusForbidden,
			"Endpoint only available in development",
			nil,
		), "Development endpoint access denied")
		return false
	}
	return true
}

// buildBaseTemplateData creates common template data that most pages need
func (h *BaseHandler) buildBaseTemplateData(title string) map[string]any {
	appVersion := "unknown"
	if h.buildInfo != nil && h.buildInfo.Version != "" {
		appVersion = h.buildInfo.Version
	}
	return map[string]any{
		"title":       title,
		"config":      h.config,
		"app_version": appVersion,
	}
}

// buildArticlePageData creates template data for article pages
func (h *BaseHandler) buildArticlePageData(title string, recentArticles []*models.Article) map[string]any {
	appVersion := "unknown"
	if h.buildInfo != nil && h.buildInfo.Version != "" {
		appVersion = h.buildInfo.Version
	}
	return map[string]any{
		"title":           title,
		"config":          h.config,
		"recent_articles": recentArticles,
		"app_version":     appVersion,
	}
}

// respondWithJSON provides pooled JSON response handling
func (h *BaseHandler) respondWithJSON(c *gin.Context, status int, dataBuilder func() map[string]any) {
	data := dataBuilder()

	c.JSON(status, data)
}
