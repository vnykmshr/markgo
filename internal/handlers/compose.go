package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/markgo/internal/services/compose"
)

// validSlug matches URL-safe slugs: lowercase alphanumeric with hyphens, no leading/trailing hyphens.
var validSlug = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

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
	data["csrf_token"] = csrfToken(c)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ShowEdit renders the compose form pre-filled with an existing article.
func (h *ComposeHandler) ShowEdit(c *gin.Context) {
	slug := c.Param("slug")
	if !validSlug.MatchString(slug) {
		c.Status(http.StatusNotFound)
		return
	}

	input, err := h.composeService.LoadArticle(slug)
	if err != nil {
		h.handleError(c, err, "Article not found")
		return
	}

	data := h.buildBaseTemplateData("Edit - " + h.config.Blog.Title)
	data["template"] = templateCompose
	data["input"] = input
	data["editing"] = true
	data["slug"] = slug
	data["csrf_token"] = csrfToken(c)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// csrfToken pulls the CSRF token from the gin context (set by CSRF middleware).
func csrfToken(c *gin.Context) string {
	if token, exists := c.Get("csrf_token"); exists {
		if s, ok := token.(string); ok {
			return s
		}
	}
	return ""
}

// refreshCSRFToken generates a new CSRF token and sets the cookie.
// Used when re-rendering a form after a POST validation error.
func refreshCSRFToken(c *gin.Context) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	token := hex.EncodeToString(b)
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("_csrf", token, 3600, "", "", true, true)
	return token
}

// HandleEdit processes the edit form submission.
func (h *ComposeHandler) HandleEdit(c *gin.Context) {
	slug := c.Param("slug")
	if !validSlug.MatchString(slug) {
		c.Status(http.StatusNotFound)
		return
	}

	input := compose.Input{
		Content: c.PostForm("content"),
		Title:   c.PostForm("title"),
		LinkURL: c.PostForm("link_url"),
		Tags:    c.PostForm("tags"),
		Draft:   c.PostForm("draft") == "on",
	}

	if input.Content == "" {
		data := h.buildBaseTemplateData("Edit - " + h.config.Blog.Title)
		data["template"] = templateCompose
		data["error"] = "Content is required"
		data["input"] = input
		data["editing"] = true
		data["slug"] = slug
		data["csrf_token"] = refreshCSRFToken(c)
		h.renderHTML(c, http.StatusBadRequest, "base.html", data)
		return
	}

	if err := h.composeService.UpdateArticle(slug, input); err != nil {
		h.logger.Error("Failed to update post", "error", err, "slug", slug)
		data := h.buildBaseTemplateData("Edit - " + h.config.Blog.Title)
		data["template"] = templateCompose
		data["error"] = "Failed to update post. Please try again."
		data["input"] = input
		data["editing"] = true
		data["slug"] = slug
		data["csrf_token"] = refreshCSRFToken(c)
		h.renderHTML(c, http.StatusInternalServerError, "base.html", data)
		return
	}

	if err := h.articleService.ReloadArticles(); err != nil {
		h.logger.Error("Failed to reload articles after edit", "error", err)
	}

	c.Redirect(http.StatusSeeOther, "/articles/"+slug)
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
		data["csrf_token"] = refreshCSRFToken(c)
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
		data["csrf_token"] = refreshCSRFToken(c)
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
