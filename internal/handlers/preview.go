package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/services"
)

type PreviewHandler struct {
	previewService  services.PreviewServiceInterface
	articleService  services.ArticleServiceInterface
	templateService services.TemplateServiceInterface
	BaseHandler
}

func NewPreviewHandler(
	previewService services.PreviewServiceInterface,
	articleService services.ArticleServiceInterface,
	templateService services.TemplateServiceInterface,
	baseHandler BaseHandler,
) *PreviewHandler {
	return &PreviewHandler{
		previewService:  previewService,
		articleService:  articleService,
		templateService: templateService,
		BaseHandler:     baseHandler,
	}
}

// CreatePreviewSession creates a new preview session for an article
func (h *PreviewHandler) CreatePreviewSession(c *gin.Context) {
	var req struct {
		ArticleSlug string `json:"article_slug" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Create preview session
	session, err := h.previewService.CreateSession(req.ArticleSlug)
	if err != nil {
		h.logger.Error("Failed to create preview session",
			"article_slug", req.ArticleSlug,
			"error", err)

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "not found"):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Article not found",
			})
		case strings.Contains(errMsg, "maximum"):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Maximum preview sessions exceeded",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create preview session",
			})
		}
		return
	}

	h.logger.Info("Preview session created",
		"session_id", session.ID,
		"article_slug", req.ArticleSlug,
		"url", session.URL)

	c.JSON(http.StatusCreated, gin.H{
		"session": session,
	})
}

// ServePreview serves the preview page for a session
func (h *PreviewHandler) ServePreview(c *gin.Context) {
	sessionID := c.Param("sessionId")

	// Get session
	session, err := h.previewService.GetSession(sessionID)
	if err != nil {
		h.logger.Warn("Preview session not found",
			"session_id", sessionID,
			"error", err)
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Preview Not Found",
			"message": "The preview session you're looking for doesn't exist or has expired.",
		})
		return
	}

	// Get the draft article
	article, err := h.articleService.GetDraftBySlug(session.ArticleSlug)
	if err != nil {
		h.logger.Error("Failed to load draft article",
			"session_id", sessionID,
			"article_slug", session.ArticleSlug,
			"error", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Preview Error",
			"message": "Failed to load the article for preview.",
		})
		return
	}

	// Prepare data for template
	data := gin.H{
		"title":     fmt.Sprintf("Preview: %s", article.Title),
		"article":   article,
		"session":   session,
		"wsURL":     fmt.Sprintf("/api/preview/ws/%s", sessionID),
		"isPreview": true,
	}

	c.HTML(http.StatusOK, "preview.html", data)
}

// WebSocketEndpoint handles WebSocket connections for live preview
func (h *PreviewHandler) WebSocketEndpoint(c *gin.Context) {
	sessionID := c.Param("sessionId")

	// Register WebSocket client
	if err := h.previewService.RegisterWebSocketClient(sessionID, c.Writer, c.Request); err != nil {
		h.logger.Error("Failed to register WebSocket client",
			"session_id", sessionID,
			"error", err)

		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Preview session not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to establish WebSocket connection",
			})
		}
		return
	}

	// Connection handled in RegisterWebSocketClient
	h.logger.Debug("WebSocket connection established",
		"session_id", sessionID,
		"remote_addr", c.Request.RemoteAddr)
}

// ListSessions lists all active preview sessions
func (h *PreviewHandler) ListSessions(c *gin.Context) {
	stats := h.previewService.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"stats":           stats,
		"service_running": h.previewService.IsRunning(),
	})
}

// DeleteSession deletes a preview session
func (h *PreviewHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("sessionId")

	if err := h.previewService.DeleteSession(sessionID); err != nil {
		h.logger.Warn("Failed to delete preview session",
			"session_id", sessionID,
			"error", err)

		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Preview session not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete preview session",
			})
		}
		return
	}

	h.logger.Info("Preview session deleted", "session_id", sessionID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Preview session deleted successfully",
	})
}
