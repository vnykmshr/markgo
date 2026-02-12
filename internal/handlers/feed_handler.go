package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// FeedHandler handles the home/feed page.
type FeedHandler struct {
	*BaseHandler
	articleService services.ArticleServiceInterface
}

// NewFeedHandler creates a new feed handler.
func NewFeedHandler(base *BaseHandler, articleService services.ArticleServiceInterface) *FeedHandler {
	return &FeedHandler{
		BaseHandler:    base,
		articleService: articleService,
	}
}

// Home handles the home page request.
func (h *FeedHandler) Home(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	typeFilter := c.Query("type")
	switch typeFilter {
	case "", templateArticle, "thought", "link":
		// valid
	default:
		typeFilter = ""
	}

	data, err := h.getHomeData(page, typeFilter)
	if err != nil {
		h.handleError(c, err, "Failed to get home data")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

func (h *FeedHandler) getHomeData(page int, typeFilter string) (map[string]any, error) {
	allArticles := h.articleService.GetAllArticles()

	postsPerPage := h.config.Blog.PostsPerPage
	if postsPerPage <= 0 {
		postsPerPage = 20
	}
	var posts []*models.Article
	for _, a := range allArticles {
		if !a.Draft && (typeFilter == "" || a.Type == typeFilter) {
			posts = append(posts, a)
		}
	}

	pagination := models.NewPagination(page, len(posts), postsPerPage)

	start := (pagination.CurrentPage - 1) * postsPerPage
	end := start + postsPerPage
	if end > len(posts) {
		end = len(posts)
	}
	pagePosts := posts[start:end]

	data := h.buildBaseTemplateData(h.config.Blog.Title)
	data["description"] = h.config.Blog.Description
	data["posts"] = pagePosts
	data["pagination"] = pagination
	data["activeFilter"] = typeFilter
	data["template"] = "feed"
	data["canonicalPath"] = "/"

	return data, nil
}
