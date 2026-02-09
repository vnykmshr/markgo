// Package handlers provides HTTP request handlers for the MarkGo blog engine.
// It includes handlers for admin operations, article management, API endpoints, and more.
package handlers

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// CachedArticleFunctions holds cached functions for article operations
type CachedArticleFunctions struct {
	GetHomeData         func() (map[string]any, error)
	GetArticleData      func(string) (map[string]any, error)
	GetArticlesPage     func(int) (map[string]any, error)
	GetTagArticles      func(string) (map[string]any, error)
	GetCategoryArticles func(string) (map[string]any, error)
	GetSearchResults    func(string) (map[string]any, error)
	GetTagsPage         func() (map[string]any, error)
	GetCategoriesPage   func() (map[string]any, error)
}

// ArticleHandler handles article-related HTTP requests
type ArticleHandler struct {
	*BaseHandler
	articleService  services.ArticleServiceInterface
	searchService   services.SearchServiceInterface
	cachedFunctions CachedArticleFunctions
}

// NewArticleHandler creates a new article handler
func NewArticleHandler(
	cfg *config.Config,
	logger *slog.Logger,
	templateService services.TemplateServiceInterface,
	articleService services.ArticleServiceInterface,
	searchService services.SearchServiceInterface,
	cachedFunctions CachedArticleFunctions,
	buildInfo *BuildInfo,
	seoService services.SEOServiceInterface,
) *ArticleHandler {
	return &ArticleHandler{
		BaseHandler:     NewBaseHandler(cfg, logger, templateService, buildInfo, seoService),
		articleService:  articleService,
		searchService:   searchService,
		cachedFunctions: cachedFunctions,
	}
}

// Home handles the home page request
func (h *ArticleHandler) Home(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	h.withCachedFallback(c,
		h.cachedFunctions.GetHomeData,
		func() (map[string]any, error) { return h.getHomeDataUncached(page) },
		"base.html",
		"Failed to get home data")
}

// Articles handles the articles listing page
func (h *ArticleHandler) Articles(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	cachedFunc := func() (map[string]any, error) {
		if h.cachedFunctions.GetArticlesPage != nil {
			return h.cachedFunctions.GetArticlesPage(page)
		}
		return nil, fmt.Errorf("cache not available")
	}

	uncachedFunc := func() (map[string]any, error) {
		return h.getArticlesPageUncached(page)
	}

	h.withCachedFallback(c, cachedFunc, uncachedFunc, "base.html", "Failed to get articles page")
}

// Article handles individual article requests
func (h *ArticleHandler) Article(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		h.handleError(c, fmt.Errorf("slug is required"), "Invalid article slug")
		return
	}

	cachedFunc := func() (map[string]any, error) {
		if h.cachedFunctions.GetArticleData != nil {
			return h.cachedFunctions.GetArticleData(slug)
		}
		return nil, fmt.Errorf("cache not available")
	}

	uncachedFunc := func() (map[string]any, error) {
		return h.getArticleDataUncached(slug)
	}

	h.withCachedFallback(c, cachedFunc, uncachedFunc, "base.html", "Failed to get article")
}

// ArticlesByTag handles articles filtered by tag
func (h *ArticleHandler) ArticlesByTag(c *gin.Context) {
	tag := c.Param("tag")
	if tag == "" {
		h.handleError(c, fmt.Errorf("tag is required"), "Invalid tag")
		return
	}

	// Decode URL-encoded tag name
	decodedTag, err := url.QueryUnescape(tag)
	if err != nil {
		h.handleError(c, fmt.Errorf("invalid tag format: %v", err), "Invalid tag")
		return
	}
	tag = decodedTag

	cachedFunc := func() (map[string]any, error) {
		if h.cachedFunctions.GetTagArticles != nil {
			return h.cachedFunctions.GetTagArticles(tag)
		}
		return nil, fmt.Errorf("cache not available")
	}

	uncachedFunc := func() (map[string]any, error) {
		return h.getTagArticlesUncached(tag)
	}

	h.withCachedFallback(c, cachedFunc, uncachedFunc, "base.html", "Failed to get articles by tag")
}

// ArticlesByCategory handles articles filtered by category
func (h *ArticleHandler) ArticlesByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		h.handleError(c, fmt.Errorf("category is required"), "Invalid category")
		return
	}

	// Decode URL-encoded category name
	decodedCategory, err := url.QueryUnescape(category)
	if err != nil {
		h.handleError(c, fmt.Errorf("invalid category format: %v", err), "Invalid category")
		return
	}
	category = decodedCategory

	cachedFunc := func() (map[string]any, error) {
		if h.cachedFunctions.GetCategoryArticles != nil {
			return h.cachedFunctions.GetCategoryArticles(category)
		}
		return nil, fmt.Errorf("cache not available")
	}

	uncachedFunc := func() (map[string]any, error) {
		return h.getCategoryArticlesUncached(category)
	}

	h.withCachedFallback(c, cachedFunc, uncachedFunc, "base.html", "Failed to get articles by category")
}

// Tags handles the tags page
func (h *ArticleHandler) Tags(c *gin.Context) {
	h.withCachedFallback(c,
		h.cachedFunctions.GetTagsPage,
		h.getTagsPageUncached,
		"base.html",
		"Failed to get tags")
}

// Categories handles the categories page
func (h *ArticleHandler) Categories(c *gin.Context) {
	h.withCachedFallback(c,
		h.cachedFunctions.GetCategoriesPage,
		h.getCategoriesPageUncached,
		"base.html",
		"Failed to get categories")
}

// Search handles search requests
func (h *ArticleHandler) Search(c *gin.Context) {
	query := c.Query("q")

	// Allow empty queries for initial search page visit
	cachedFunc := func() (map[string]any, error) {
		if h.cachedFunctions.GetSearchResults != nil {
			return h.cachedFunctions.GetSearchResults(query)
		}
		return nil, fmt.Errorf("cache not available")
	}

	uncachedFunc := func() (map[string]any, error) {
		return h.getSearchResultsUncached(query)
	}

	h.withCachedFallback(c, cachedFunc, uncachedFunc, "base.html", "Failed to perform search")
}

// Helper methods

const (
	recentArticlesLimit = 5

	// Template names
	templateArticle = "article"
)

// getRecentArticles returns the most recent published articles for footer display
func (h *ArticleHandler) getRecentArticles() []*models.Article {
	allArticles := h.articleService.GetAllArticles()
	var recent []*models.Article
	for _, article := range allArticles {
		if !article.Draft && len(recent) < recentArticlesLimit {
			recent = append(recent, article)
		}
	}
	return recent
}

// Uncached data generation methods (to be extracted from original handlers.go)

func (h *ArticleHandler) getHomeDataUncached(page int) (map[string]any, error) {
	allArticles := h.articleService.GetAllArticles()

	// Collect published posts (all types)
	postsPerPage := h.config.Blog.PostsPerPage
	if postsPerPage <= 0 {
		postsPerPage = 20
	}
	var posts []*models.Article
	for _, a := range allArticles {
		if !a.Draft {
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
	data["template"] = "feed"

	return data, nil
}

func (h *ArticleHandler) getArticleDataUncached(slug string) (map[string]any, error) {
	article, err := h.articleService.GetArticleBySlug(slug)
	if err != nil {
		return nil, err
	}

	if article.Draft {
		return nil, fmt.Errorf("article not found: %s", slug)
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
	data["recent"] = h.getRecentArticles()
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
	data["recent"] = h.getRecentArticles()
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
	data["recent"] = h.getRecentArticles()
	data["template"] = "category"

	return data, nil
}

func (h *ArticleHandler) getTagsPageUncached() (map[string]any, error) {
	tagCounts := h.articleService.GetTagCounts()

	data := h.buildBaseTemplateData("Tags - " + h.config.Blog.Title)
	data["description"] = "All tags from " + h.config.Blog.Title
	data["tags"] = tagCounts
	data["count"] = len(tagCounts)
	data["recent"] = h.getRecentArticles()
	data["template"] = "tags"

	return data, nil
}

func (h *ArticleHandler) getCategoriesPageUncached() (map[string]any, error) {
	categoryCounts := h.articleService.GetCategoryCounts()

	data := h.buildBaseTemplateData("Categories - " + h.config.Blog.Title)
	data["description"] = "All categories from " + h.config.Blog.Title
	data["categories"] = categoryCounts
	data["count"] = len(categoryCounts)
	data["recent"] = h.getRecentArticles()
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
	data["recent"] = h.getRecentArticles()
	data["template"] = "search"

	return data, nil
}
