package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// PostHandler handles individual article and article listing pages.
type PostHandler struct {
	*BaseHandler
	articleService services.ArticleServiceInterface
}

// NewPostHandler creates a new post handler.
func NewPostHandler(base *BaseHandler, articleService services.ArticleServiceInterface) *PostHandler {
	return &PostHandler{
		BaseHandler:    base,
		articleService: articleService,
	}
}

// Articles handles the articles listing page.
func (h *PostHandler) Articles(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	data, err := h.getArticlesPage(page)
	if err != nil {
		h.handleError(c, err, "Failed to get articles page")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Article handles individual article requests.
func (h *PostHandler) Article(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		h.handleError(c, apperrors.NewValidationError("slug", "", "slug is required", nil), "Invalid article slug")
		return
	}

	data, err := h.getArticleData(slug)
	if err != nil {
		h.handleError(c, err, "Failed to get article")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

func (h *PostHandler) getArticleData(slug string) (map[string]any, error) {
	article, err := h.articleService.GetArticleBySlug(slug)
	if err != nil {
		return nil, err
	}

	if article.Draft {
		return nil, apperrors.ErrArticleNotFound
	}

	// Get recent articles for sidebar
	allArticles := h.articleService.GetAllArticles()
	var recent []*models.Article
	for _, a := range allArticles {
		if !a.Draft && a.Slug != slug && len(recent) < 5 {
			recent = append(recent, a)
		}
	}

	data := h.buildArticlePageData(article.Title+" - "+h.config.Blog.Title, recent)
	data["article"] = article
	data["description"] = article.Description
	data["template"] = templateArticle
	data["canonicalPath"] = "/writing/" + article.Slug
	data["breadcrumbs"] = []services.Breadcrumb{
		{Name: "Home", URL: "/"},
		{Name: "Writing", URL: "/writing"},
		{Name: article.Title},
	}

	return data, nil
}

func (h *PostHandler) getArticlesPage(page int) (map[string]any, error) {
	allArticles := h.articleService.GetAllArticles()

	var published []*models.Article
	for _, article := range allArticles {
		if !article.Draft {
			published = append(published, article)
		}
	}

	postsPerPage := h.config.Blog.PostsPerPage
	if postsPerPage <= 0 {
		postsPerPage = 10
	}

	pagination := models.NewPagination(page, len(published), postsPerPage)

	start := (pagination.CurrentPage - 1) * postsPerPage
	end := start + postsPerPage
	if end > len(published) {
		end = len(published)
	}

	articles := published[start:end]

	data := h.buildBaseTemplateData("Writing - " + h.config.Blog.Title)
	data["description"] = "Writing from " + h.config.Blog.Title
	data["articles"] = articles
	data["pagination"] = pagination
	data["template"] = "articles"
	data["canonicalPath"] = "/writing"

	return data, nil
}
