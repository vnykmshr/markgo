package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/markgo/internal/services/compose"
)

const templateCompose = "compose"

// ComposeHandler handles the compose page for creating new posts.
type ComposeHandler struct {
	*BaseHandler
	composeService *compose.Service
	articleService services.ArticleServiceInterface
	cacheService   CacheAdapter
}

// NewComposeHandler creates a new compose handler.
func NewComposeHandler(
	cfg *config.Config,
	logger *slog.Logger,
	templateService services.TemplateServiceInterface,
	composeService *compose.Service,
	articleService services.ArticleServiceInterface,
	cacheService CacheAdapter,
	buildInfo *BuildInfo,
	seoService services.SEOServiceInterface,
) *ComposeHandler {
	return &ComposeHandler{
		BaseHandler:    NewBaseHandler(cfg, logger, templateService, buildInfo, seoService),
		composeService: composeService,
		articleService: articleService,
		cacheService:   cacheService,
	}
}

// ShowCompose renders the compose form.
func (h *ComposeHandler) ShowCompose(c *gin.Context) {
	data := h.buildBaseTemplateData("Compose - " + h.config.Blog.Title)
	data["template"] = templateCompose
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// HandleSubmit processes the compose form submission.
func (h *ComposeHandler) HandleSubmit(c *gin.Context) {
	input := compose.Input{
		Content: c.PostForm("content"),
		Title:   c.PostForm("title"),
		LinkURL: c.PostForm("link_url"),
		Tags:    c.PostForm("tags"),
		Draft:   c.PostForm("draft") == "on",
	}

	if input.Content == "" {
		data := h.buildBaseTemplateData("Compose - " + h.config.Blog.Title)
		data["template"] = templateCompose
		data["error"] = "Content is required"
		data["input"] = input
		h.renderHTML(c, http.StatusBadRequest, "base.html", data)
		return
	}

	slug, err := h.composeService.CreatePost(input)
	if err != nil {
		h.logger.Error("Failed to create post", "error", err)
		data := h.buildBaseTemplateData("Compose - " + h.config.Blog.Title)
		data["template"] = templateCompose
		data["error"] = "Failed to create post: " + err.Error()
		data["input"] = input
		h.renderHTML(c, http.StatusInternalServerError, "base.html", data)
		return
	}

	// Reload articles so the new post appears in the feed
	if err := h.articleService.ReloadArticles(); err != nil {
		h.logger.Warn("Failed to reload articles after compose", "error", err)
	}

	// Clear cache so stale pages aren't served
	if h.cacheService != nil {
		h.cacheService.Clear()
	}

	// Redirect to the new post (or feed for thoughts)
	if input.Title == "" {
		c.Redirect(http.StatusSeeOther, "/")
	} else {
		c.Redirect(http.StatusSeeOther, "/articles/"+slug)
	}
}
