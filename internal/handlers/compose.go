package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/markgo/internal/services/compose"
)

const templateCompose = "compose"

// ComposeHandler handles the compose page for creating new posts.
type ComposeHandler struct {
	*BaseHandler
	composeService *compose.Service
	articleService services.ArticleServiceInterface
}

// NewComposeHandler creates a new compose handler.
func NewComposeHandler(
	base *BaseHandler,
	composeService *compose.Service,
	articleService services.ArticleServiceInterface,
) *ComposeHandler {
	return &ComposeHandler{
		BaseHandler:    base,
		composeService: composeService,
		articleService: articleService,
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
		data["error"] = "Failed to create post. Please try again."
		data["input"] = input
		h.renderHTML(c, http.StatusInternalServerError, "base.html", data)
		return
	}

	// Reload articles so the new post appears in the feed
	reloadOK := true
	if err := h.articleService.ReloadArticles(); err != nil {
		h.logger.Error("Failed to reload articles after compose", "error", err)
		reloadOK = false
	}

	// Redirect to the new post, or feed if reload failed (article won't be in memory)
	if !reloadOK || input.Title == "" {
		c.Redirect(http.StatusSeeOther, "/")
	} else {
		c.Redirect(http.StatusSeeOther, "/articles/"+slug)
	}
}
