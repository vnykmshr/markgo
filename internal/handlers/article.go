package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// ArticleHandler handles article-related HTTP requests
type ArticleHandler struct {
	*BaseHandler
	articleService services.ArticleServiceInterface
	searchService  services.SearchServiceInterface
}

// NewArticleHandler creates a new article handler
func NewArticleHandler(
	base *BaseHandler,
	articleService services.ArticleServiceInterface,
	searchService services.SearchServiceInterface,
) *ArticleHandler {
	return &ArticleHandler{
		BaseHandler:    base,
		articleService: articleService,
		searchService:  searchService,
	}
}

// Home handles the home page request
func (h *ArticleHandler) Home(c *gin.Context) {
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

	data, err := h.getHomeDataUncached(page, typeFilter)
	if err != nil {
		h.handleError(c, err, "Failed to get home data")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Articles handles the articles listing page
func (h *ArticleHandler) Articles(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	data, err := h.getArticlesPageUncached(page)
	if err != nil {
		h.handleError(c, err, "Failed to get articles page")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Article handles individual article requests
func (h *ArticleHandler) Article(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		h.handleError(c, apperrors.NewValidationError("slug", "", "slug is required", nil), "Invalid article slug")
		return
	}

	data, err := h.getArticleDataUncached(slug)
	if err != nil {
		h.handleError(c, err, "Failed to get article")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ArticlesByTag handles articles filtered by tag
func (h *ArticleHandler) ArticlesByTag(c *gin.Context) {
	tag := c.Param("tag")
	if tag == "" {
		h.handleError(c, apperrors.NewValidationError("tag", "", "tag is required", nil), "Invalid tag")
		return
	}

	// Decode URL-encoded tag name
	decodedTag, err := url.QueryUnescape(tag)
	if err != nil {
		h.handleError(c, fmt.Errorf("invalid tag format: %v", err), "Invalid tag")
		return
	}
	tag = decodedTag

	data, err := h.getTagArticlesUncached(tag)
	if err != nil {
		h.handleError(c, err, "Failed to get articles by tag")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// ArticlesByCategory handles articles filtered by category
func (h *ArticleHandler) ArticlesByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		h.handleError(c, apperrors.NewValidationError("category", "", "category is required", nil), "Invalid category")
		return
	}

	// Decode URL-encoded category name
	decodedCategory, err := url.QueryUnescape(category)
	if err != nil {
		h.handleError(c, fmt.Errorf("invalid category format: %v", err), "Invalid category")
		return
	}
	category = decodedCategory

	data, err := h.getCategoryArticlesUncached(category)
	if err != nil {
		h.handleError(c, err, "Failed to get articles by category")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Tags handles the tags page
func (h *ArticleHandler) Tags(c *gin.Context) {
	data, err := h.getTagsPageUncached()
	if err != nil {
		h.handleError(c, err, "Failed to get tags")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Categories handles the categories page
func (h *ArticleHandler) Categories(c *gin.Context) {
	data, err := h.getCategoriesPageUncached()
	if err != nil {
		h.handleError(c, err, "Failed to get categories")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Search handles search requests
func (h *ArticleHandler) Search(c *gin.Context) {
	query := c.Query("q")

	data, err := h.getSearchResultsUncached(query)
	if err != nil {
		h.handleError(c, err, "Failed to perform search")
		return
	}

	h.enhanceTemplateDataWithSEO(data, c.Request.URL.Path)
	h.renderHTML(c, http.StatusOK, "base.html", data)
}

// Helper methods

const (
	// Template names
	templateArticle = "article"
)

// Uncached data generation methods

func (h *ArticleHandler) getHomeDataUncached(page int, typeFilter string) (map[string]any, error) {
	allArticles := h.articleService.GetAllArticles()

	// Collect published posts, optionally filtered by type
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

	// Slice to current page
	start := (page - 1) * postsPerPage
	if start > len(posts) {
		start = len(posts)
	}
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

	return data, nil
}

func (h *ArticleHandler) getArticleDataUncached(slug string) (map[string]any, error) {
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

	// Use special template for about page
	templateName := templateArticle
	if slug == "about" {
		templateName = "about-article"
	}

	data := h.buildArticlePageData(article.Title+" - "+h.config.Blog.Title, recent)
	data["article"] = article
	data["description"] = article.Description
	data["template"] = templateName

	return data, nil
}

func (h *ArticleHandler) getArticlesPageUncached(page int) (map[string]any, error) {
	allArticles := h.articleService.GetAllArticles()

	// Filter published articles
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

	totalPages := (len(published) + postsPerPage - 1) / postsPerPage
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	start := (page - 1) * postsPerPage
	end := start + postsPerPage
	if end > len(published) {
		end = len(published)
	}

	articles := published[start:end]

	pagination := models.Pagination{
		CurrentPage:  page,
		TotalPages:   totalPages,
		TotalItems:   len(published),
		ItemsPerPage: postsPerPage,
		HasPrevious:  page > 1,
		HasNext:      page < totalPages,
		PreviousPage: page - 1,
		NextPage:     page + 1,
	}

	data := h.buildBaseTemplateData("Articles - " + h.config.Blog.Title)
	data["description"] = "Articles from " + h.config.Blog.Title
	data["articles"] = articles
	data["pagination"] = pagination

	data["template"] = "articles"

	return data, nil
}

func (h *ArticleHandler) getTagArticlesUncached(tag string) (map[string]any, error) {
	articles := h.articleService.GetArticlesByTag(tag)

	// Filter published articles
	var published []*models.Article
	for _, article := range articles {
		if !article.Draft {
			published = append(published, article)
		}
	}

	data := h.buildBaseTemplateData("Articles tagged with " + tag + " - " + h.config.Blog.Title)
	data["description"] = "Articles tagged with " + tag
	data["articles"] = published
	data["tag"] = tag
	data["totalCount"] = len(published)

	data["template"] = "tag"

	return data, nil
}

func (h *ArticleHandler) getCategoryArticlesUncached(category string) (map[string]any, error) {
	articles := h.articleService.GetArticlesByCategory(category)

	// Filter published articles
	var published []*models.Article
	for _, article := range articles {
		if !article.Draft {
			published = append(published, article)
		}
	}

	data := h.buildBaseTemplateData("Articles in " + category + " - " + h.config.Blog.Title)
	data["description"] = "Articles in " + category
	data["articles"] = published
	data["category"] = category
	data["totalCount"] = len(published)

	data["template"] = "category"

	return data, nil
}

func (h *ArticleHandler) getTagsPageUncached() (map[string]any, error) {
	tagCounts := h.articleService.GetTagCounts()

	data := h.buildBaseTemplateData("Tags - " + h.config.Blog.Title)
	data["description"] = "All tags from " + h.config.Blog.Title
	data["tags"] = tagCounts
	data["count"] = len(tagCounts)

	data["template"] = "tags"

	return data, nil
}

func (h *ArticleHandler) getCategoriesPageUncached() (map[string]any, error) {
	categoryCounts := h.articleService.GetCategoryCounts()

	data := h.buildBaseTemplateData("Categories - " + h.config.Blog.Title)
	data["description"] = "All categories from " + h.config.Blog.Title
	data["categories"] = categoryCounts
	data["count"] = len(categoryCounts)

	data["template"] = "categories"

	return data, nil
}

func (h *ArticleHandler) getSearchResultsUncached(query string) (map[string]any, error) {
	allArticles := h.articleService.GetAllArticles()

	// Filter published articles for search
	var published []*models.Article
	for _, article := range allArticles {
		if !article.Draft {
			published = append(published, article)
		}
	}

	var results []*models.SearchResult
	var title, description string

	if query == "" {
		// Empty query - show search page without results
		results = []*models.SearchResult{}
		title = "Search - " + h.config.Blog.Title
		description = "Search through articles on " + h.config.Blog.Title
	} else {
		// Perform search with query
		results = h.searchService.Search(published, query, 50)
		title = "Search results for \"" + query + "\" - " + h.config.Blog.Title
		description = "Search results for \"" + query + "\""
	}

	data := h.buildBaseTemplateData(title)
	data["description"] = description
	data["results"] = results
	data["query"] = query
	data["count"] = len(results)
	data["totalCount"] = len(published)

	data["template"] = "search"

	return data, nil
}
