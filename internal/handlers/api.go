// Package handlers provides HTTP request handlers for the MarkGo blog engine.
// It includes handlers for admin operations, article management, API endpoints, and more.
package handlers

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// CachedAPIFunctions holds cached functions for API operations
type CachedAPIFunctions struct {
	GetRSSFeed  func() (string, error)
	GetJSONFeed func() (string, error)
	GetSitemap  func() (string, error)
}

// APIHandler handles API HTTP requests (RSS, JSON Feed, Sitemap, Health)
type APIHandler struct {
	*BaseHandler
	articleService  services.ArticleServiceInterface
	emailService    services.EmailServiceInterface
	startTime       time.Time
	cachedFunctions CachedAPIFunctions
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(
	config *config.Config,
	logger *slog.Logger,
	templateService services.TemplateServiceInterface,
	articleService services.ArticleServiceInterface,
	emailService services.EmailServiceInterface,
	startTime time.Time,
	cachedFunctions CachedAPIFunctions,
) *APIHandler {
	return &APIHandler{
		BaseHandler:     NewBaseHandler(config, logger, templateService),
		articleService:  articleService,
		emailService:    emailService,
		startTime:       startTime,
		cachedFunctions: cachedFunctions,
	}
}

// Health handles health check requests
func (h *APIHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)

	health := map[string]any{
		"status":      "healthy",
		"timestamp":   time.Now().Unix(),
		"uptime":      uptime.String(),
		"version":     "v1.4.0", // This could come from build info
		"environment": h.config.Environment,
		"services": map[string]any{
			"articles": "healthy", // Could check article service health
			"cache":    "healthy", // Could check cache connectivity
		},
	}

	c.JSON(http.StatusOK, health)
}

// RSS handles RSS feed generation
func (h *APIHandler) RSS(c *gin.Context) {
	cachedFunc := func() (string, error) {
		if h.cachedFunctions.GetRSSFeed != nil {
			return h.cachedFunctions.GetRSSFeed()
		}
		return "", nil
	}

	h.withCachedStringFallback(c, cachedFunc, h.getRSSFeedUncached,
		"application/rss+xml; charset=utf-8", "Failed to generate RSS feed")
}

// JSONFeed handles JSON feed generation
func (h *APIHandler) JSONFeed(c *gin.Context) {
	cachedFunc := func() (string, error) {
		if h.cachedFunctions.GetJSONFeed != nil {
			return h.cachedFunctions.GetJSONFeed()
		}
		return "", nil
	}

	h.withCachedStringFallback(c, cachedFunc, h.getJSONFeedUncached,
		"application/feed+json; charset=utf-8", "Failed to generate JSON feed")
}

// Sitemap handles sitemap.xml generation
func (h *APIHandler) Sitemap(c *gin.Context) {
	cachedFunc := func() (string, error) {
		if h.cachedFunctions.GetSitemap != nil {
			return h.cachedFunctions.GetSitemap()
		}
		return "", nil
	}

	h.withCachedStringFallback(c, cachedFunc, h.getSitemapUncached,
		"application/xml; charset=utf-8", "Failed to generate sitemap")
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

// Uncached data generation methods

func (h *APIHandler) getRSSFeedUncached() (string, error) {
	// Get published articles
	allArticles := h.articleService.GetAllArticles()
	var published []*models.Article
	for _, article := range allArticles {
		if !article.Draft && len(published) < 20 { // Limit to 20 most recent
			published = append(published, article)
		}
	}

	// Build RSS feed using existing Feed model
	feed := models.Feed{
		Title:       h.config.Blog.Title,
		Link:        h.config.BaseURL,
		Description: h.config.Blog.Description,
		FeedURL:     h.config.BaseURL + "/rss.xml",
		Author: models.Author{
			Name:  h.config.Blog.Author,
			Email: h.config.Blog.AuthorEmail,
		},
		Items: make([]models.FeedItem, 0, len(published)),
	}

	for _, article := range published {
		item := models.FeedItem{
			ID:          h.config.BaseURL + "/articles/" + article.Slug,
			Title:       article.Title,
			ContentHTML: article.Content,
			URL:         h.config.BaseURL + "/articles/" + article.Slug,
			Summary:     article.Description,
			Published:   article.Date,
			Tags:        article.Tags,
		}
		feed.Items = append(feed.Items, item)
	}

	// Convert to RSS XML format (simplified)
	rssXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>` + feed.Title + `</title>
<link>` + feed.Link + `</link>
<description>` + feed.Description + `</description>
<language>` + h.config.Blog.Language + `</language>
`
	for _, item := range feed.Items {
		rssXML += `<item>
<title>` + item.Title + `</title>
<link>` + item.URL + `</link>
<description>` + item.Summary + `</description>
<pubDate>` + item.Published.Format(time.RFC1123Z) + `</pubDate>
<guid>` + item.ID + `</guid>
</item>
`
	}
	rssXML += `</channel>
</rss>`

	return rssXML, nil
}

func (h *APIHandler) getJSONFeedUncached() (string, error) {
	// Get published articles
	allArticles := h.articleService.GetAllArticles()
	var published []*models.Article
	for _, article := range allArticles {
		if !article.Draft && len(published) < 20 { // Limit to 20 most recent
			published = append(published, article)
		}
	}

	// Build JSON feed using existing Feed model
	feed := models.Feed{
		Title:       h.config.Blog.Title,
		Link:        h.config.BaseURL,
		FeedURL:     h.config.BaseURL + "/feed.json",
		Description: h.config.Blog.Description,
		Author: models.Author{
			Name:  h.config.Blog.Author,
			Email: h.config.Blog.AuthorEmail,
		},
		Items: make([]models.FeedItem, 0, len(published)),
	}

	for _, article := range published {
		item := models.FeedItem{
			ID:        h.config.BaseURL + "/articles/" + article.Slug,
			URL:       h.config.BaseURL + "/articles/" + article.Slug,
			Title:     article.Title,
			Summary:   article.Description,
			Published: article.Date,
			Tags:      article.Tags,
		}
		feed.Items = append(feed.Items, item)
	}

	// Convert to JSON format
	jsonBytes, err := json.Marshal(map[string]any{
		"version":       "https://jsonfeed.org/version/1.1",
		"title":         feed.Title,
		"home_page_url": feed.Link,
		"feed_url":      feed.FeedURL,
		"description":   feed.Description,
		"author": map[string]string{
			"name":  feed.Author.Name,
			"email": feed.Author.Email,
		},
		"items": feed.Items,
	})
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (h *APIHandler) getSitemapUncached() (string, error) {
	// Get all published articles
	allArticles := h.articleService.GetAllArticles()

	urls := []models.SitemapURL{
		{
			Loc:        h.config.BaseURL,
			LastMod:    time.Now(),
			ChangeFreq: "weekly",
			Priority:   1.0,
		},
		{
			Loc:        h.config.BaseURL + "/articles",
			LastMod:    time.Now(),
			ChangeFreq: "daily",
			Priority:   0.8,
		},
		{
			Loc:        h.config.BaseURL + "/tags",
			LastMod:    time.Now(),
			ChangeFreq: "weekly",
			Priority:   0.6,
		},
		{
			Loc:        h.config.BaseURL + "/categories",
			LastMod:    time.Now(),
			ChangeFreq: "weekly",
			Priority:   0.6,
		},
	}

	// Add published articles
	for _, article := range allArticles {
		if !article.Draft {
			urls = append(urls, models.SitemapURL{
				Loc:        h.config.BaseURL + "/articles/" + article.Slug,
				LastMod:    article.Date,
				ChangeFreq: "monthly",
				Priority:   0.7,
			})
		}
	}

	sitemap := models.Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	xmlBytes, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		return "", err
	}
	return `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(xmlBytes), nil
}
