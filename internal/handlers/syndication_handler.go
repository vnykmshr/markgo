package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/services"
)

// SyndicationHandler handles RSS, JSON Feed, and Sitemap generation.
type SyndicationHandler struct {
	*BaseHandler
	feedService services.FeedServiceInterface
}

// NewSyndicationHandler creates a new syndication handler.
func NewSyndicationHandler(base *BaseHandler, feedService services.FeedServiceInterface) *SyndicationHandler {
	return &SyndicationHandler{
		BaseHandler: base,
		feedService: feedService,
	}
}

// RSS handles RSS feed generation.
func (h *SyndicationHandler) RSS(c *gin.Context) {
	data, err := h.feedService.GenerateRSS()
	if err != nil {
		h.handleError(c, err, "Failed to generate RSS feed")
		return
	}

	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.String(http.StatusOK, data)
}

// JSONFeed handles JSON feed generation.
func (h *SyndicationHandler) JSONFeed(c *gin.Context) {
	data, err := h.feedService.GenerateJSONFeed()
	if err != nil {
		h.handleError(c, err, "Failed to generate JSON feed")
		return
	}

	c.Header("Content-Type", "application/feed+json; charset=utf-8")
	c.String(http.StatusOK, data)
}

// Sitemap handles sitemap.xml generation.
func (h *SyndicationHandler) Sitemap(c *gin.Context) {
	data, err := h.feedService.GenerateSitemap()
	if err != nil {
		h.handleError(c, err, "Failed to generate sitemap")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, data)
}
