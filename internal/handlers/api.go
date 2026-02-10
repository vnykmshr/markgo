package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/constants"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// APIHandler handles API HTTP requests (RSS, JSON Feed, Sitemap, Health)
type APIHandler struct {
	*BaseHandler
	articleService services.ArticleServiceInterface
	emailService   services.EmailServiceInterface
	feedService    services.FeedServiceInterface
	startTime      time.Time
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(
	base *BaseHandler,
	articleService services.ArticleServiceInterface,
	emailService services.EmailServiceInterface,
	feedService services.FeedServiceInterface,
	startTime time.Time,
) *APIHandler {
	return &APIHandler{
		BaseHandler:    base,
		articleService: articleService,
		emailService:   emailService,
		feedService:    feedService,
		startTime:      startTime,
	}
}

// Health handles health check requests
func (h *APIHandler) Health(c *gin.Context) {
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

// RSS handles RSS feed generation
func (h *APIHandler) RSS(c *gin.Context) {
	data, err := h.feedService.GenerateRSS()
	if err != nil {
		h.handleError(c, err, "Failed to generate RSS feed")
		return
	}

	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.String(http.StatusOK, data)
}

// JSONFeed handles JSON feed generation
func (h *APIHandler) JSONFeed(c *gin.Context) {
	data, err := h.feedService.GenerateJSONFeed()
	if err != nil {
		h.handleError(c, err, "Failed to generate JSON feed")
		return
	}

	c.Header("Content-Type", "application/feed+json; charset=utf-8")
	c.String(http.StatusOK, data)
}

// Sitemap handles sitemap.xml generation
func (h *APIHandler) Sitemap(c *gin.Context) {
	data, err := h.feedService.GenerateSitemap()
	if err != nil {
		h.handleError(c, err, "Failed to generate sitemap")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, data)
}

// Contact handles contact form submissions
func (h *APIHandler) Contact(c *gin.Context) {
	var form struct {
		Name    string `json:"name" binding:"required"`
		Email   string `json:"email" binding:"required,email"`
		Subject string `json:"subject" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Create contact message model
	contactMsg := &models.ContactMessage{
		Name:    form.Name,
		Email:   form.Email,
		Subject: form.Subject,
		Message: form.Message,
	}

	// Send email through email service
	if err := h.emailService.SendContactMessage(contactMsg); err != nil {
		// Handle missing email configuration gracefully
		if errors.Is(err, apperrors.ErrEmailNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Contact form temporarily unavailable",
				"message": "Email service is not configured. Please try again later or contact us through alternative means.",
				"status":  "unavailable",
			})
			return
		}

		h.handleError(c, err, "Failed to send contact message")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Contact message sent successfully",
		"status":  "success",
	})
}
