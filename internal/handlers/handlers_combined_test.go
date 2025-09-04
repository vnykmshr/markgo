package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/markgo/internal/models"
)

// TestHandlerGroup_PageRoutes tests multiple page routes with table-driven approach
func TestHandlerGroup_PageRoutes(t *testing.T) {
	testCases := []TableDrivenTestCase{
		{
			Name: "home page cache miss",
			SetupMocks: func(mocks *TestHandlerMocks) {
				articles := CreateTestArticlesWithVariations()
				SetupCacheMocks(mocks.CacheService, "home_page", false, nil)
				SetupArticleServiceMocks(mocks.ArticleService, articles)
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   []string{"Test Blog"},
		},
		{
			Name: "home page cache hit",
			SetupMocks: func(mocks *TestHandlerMocks) {
				cacheData := gin.H{
					"title":       "Cached Title",
					"description": "Cached Description",
				}
				SetupCacheMocks(mocks.CacheService, "home_page", true, cacheData)
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "article page success",
			SetupMocks: func(mocks *TestHandlerMocks) {
				articles := CreateTestArticlesWithVariations()
				SetupCacheMocks(mocks.CacheService, "article_published-article", false, nil)
				mocks.ArticleService.On("GetArticleBySlug", "published-article").Return(articles[0], nil)
				mocks.ArticleService.On("GetArticlesByTag", "test").Return([]*models.Article{})
				mocks.ArticleService.On("GetArticlesByTag", "golang").Return([]*models.Article{})
				mocks.ArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/articles/published-article", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "article not found",
			SetupMocks: func(mocks *TestHandlerMocks) {
				mocks.CacheService.On("Get", "article_nonexistent").Return(nil, false)
				mocks.ArticleService.On("GetArticleBySlug", "nonexistent").Return(nil, assert.AnError)
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/articles/nonexistent", nil)
				return req
			},
			ExpectedStatus: http.StatusNotFound,
		},
		{
			Name: "tags page",
			SetupMocks: func(mocks *TestHandlerMocks) {
				SetupCacheMocks(mocks.CacheService, "all_tags", false, nil)
				mocks.ArticleService.On("GetTagCounts").Return([]models.TagCount{
					{Tag: "golang", Count: 5},
					{Tag: "web", Count: 3},
				})
				mocks.ArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/tags", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "categories page",
			SetupMocks: func(mocks *TestHandlerMocks) {
				SetupCacheMocks(mocks.CacheService, "all_categories", false, nil)
				mocks.ArticleService.On("GetCategoryCounts").Return([]models.CategoryCount{
					{Category: "programming", Count: 5},
					{Category: "development", Count: 3},
				})
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/categories", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	RunTableDrivenTest(t, testCases, func(router *gin.Engine, handlers *Handlers) {
		router.GET("/", handlers.Home)
		router.GET("/articles/:slug", handlers.Article)
		router.GET("/tags", handlers.Tags)
		router.GET("/categories", handlers.Categories)
	})
}

// TestHandlerGroup_SearchAndFiltering tests search and filtering functionality
func TestHandlerGroup_SearchAndFiltering(t *testing.T) {
	testCases := []TableDrivenTestCase{
		{
			Name: "search with results",
			SetupMocks: func(mocks *TestHandlerMocks) {
				articles := CreateTestArticlesWithVariations()
				searchResults := []*models.SearchResult{
					{
						Article:       articles[0],
						Score:         10.5,
						MatchedFields: []string{"title", "content"},
					},
				}
				SetupCacheMocks(mocks.CacheService, "search_golang", false, nil)
				mocks.ArticleService.On("GetAllArticles").Return(articles)
				mocks.SearchService.On("Search", articles, "golang", 20).Return(searchResults)
				mocks.ArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/search?q=golang", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "search empty query",
			SetupMocks: func(mocks *TestHandlerMocks) {
				SetupDefaultMocks(mocks)
				mocks.ArticleService.On("GetAllArticles").Return([]*models.Article{})
				mocks.ArticleService.On("GetTagCounts").Return([]models.TagCount{})
				mocks.ArticleService.On("GetCategoryCounts").Return([]models.CategoryCount{})
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/search?q=", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "articles by tag",
			SetupMocks: func(mocks *TestHandlerMocks) {
				articles := CreateTestArticlesWithVariations()
				SetupCacheMocks(mocks.CacheService, "articles_tag_golang", false, nil)
				mocks.ArticleService.On("GetArticlesByTag", "golang").Return([]*models.Article{articles[0]})
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/tags/golang", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "articles by category",
			SetupMocks: func(mocks *TestHandlerMocks) {
				articles := CreateTestArticlesWithVariations()
				SetupCacheMocks(mocks.CacheService, "articles_category_programming", false, nil)
				mocks.ArticleService.On("GetArticlesByCategory", "programming").Return([]*models.Article{articles[0]})
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/categories/programming", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	RunTableDrivenTest(t, testCases, func(router *gin.Engine, handlers *Handlers) {
		router.GET("/search", handlers.Search)
		router.GET("/tags/:tag", handlers.ArticlesByTag)
		router.GET("/categories/:category", handlers.ArticlesByCategory)
	})
}

// TestHandlerGroup_ContactForm tests contact form functionality
func TestHandlerGroup_ContactForm(t *testing.T) {
	testCases := []TableDrivenTestCase{
		{
			Name: "contact form display",
			SetupMocks: func(mocks *TestHandlerMocks) {
				mocks.ArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/contact", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "contact form submit success",
			SetupMocks: func(mocks *TestHandlerMocks) {
				mocks.EmailService.On("SendContactMessage", &models.ContactMessage{
					Name:            "John Doe",
					Email:           "john@example.com",
					Subject:         "Test Message",
					Message:         "This is a test message with sufficient length to pass validation",
					CaptchaQuestion: "3 + 5",
					CaptchaAnswer:   "8",
				}).Return(nil)
			},
			RequestFunc: func() *http.Request {
				data := CreateContactFormData()
				req, _ := MakeJSONRequest("POST", "/contact", data)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "contact form submit validation error",
			SetupMocks: func(mocks *TestHandlerMocks) {
				// No email service call expected due to validation failure
			},
			RequestFunc: func() *http.Request {
				data := CreateInvalidContactFormData()
				req, _ := MakeJSONRequest("POST", "/contact", data)
				return req
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "contact form invalid captcha",
			SetupMocks: func(mocks *TestHandlerMocks) {
				// No email service call expected due to captcha failure
			},
			RequestFunc: func() *http.Request {
				data := CreateContactFormData()
				data["captcha_answer"] = "wrong_answer"
				req, _ := MakeJSONRequest("POST", "/contact", data)
				return req
			},
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	RunTableDrivenTest(t, testCases, func(router *gin.Engine, handlers *Handlers) {
		router.GET("/contact", handlers.ContactForm)
		router.POST("/contact", handlers.ContactSubmit)
	})
}

// TestHandlerGroup_AdminOperations tests admin functionality
func TestHandlerGroup_AdminOperations(t *testing.T) {
	testCases := []TableDrivenTestCase{
		{
			Name: "admin stats",
			SetupMocks: func(mocks *TestHandlerMocks) {
				mocks.ArticleService.On("GetStats").Return(&models.Stats{
					TotalArticles:  10,
					PublishedCount: 8,
					DraftCount:     2,
				})
				mocks.CacheService.On("Stats").Return(map[string]interface{}{
					"total_items": 5,
				})
				mocks.EmailService.On("GetConfig").Return(map[string]interface{}{
					"host": "smtp.example.com",
				})
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/admin/stats", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
			ValidationFunc: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				response := AssertJSONResponse(t, recorder, http.StatusOK)
				assert.Contains(t, response, "articles")
				assert.Contains(t, response, "cache")
			},
		},
		{
			Name: "clear cache",
			SetupMocks: func(mocks *TestHandlerMocks) {
				mocks.CacheService.On("Clear").Return()
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("POST", "/admin/cache/clear", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
			ValidationFunc: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				response := AssertJSONResponse(t, recorder, http.StatusOK)
				assert.Equal(t, "Cache cleared successfully", response["message"])
			},
		},
		{
			Name: "reload articles",
			SetupMocks: func(mocks *TestHandlerMocks) {
				mocks.ArticleService.On("ReloadArticles").Return(nil)
				mocks.CacheService.On("Clear").Return()
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("POST", "/admin/reload", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	RunTableDrivenTest(t, testCases, func(router *gin.Engine, handlers *Handlers) {
		router.GET("/admin/stats", handlers.AdminStats)
		router.POST("/admin/cache/clear", handlers.ClearCache)
		router.POST("/admin/reload", handlers.ReloadArticles)
	})
}

// TestHandlerGroup_FeedsAndSitemap tests RSS, JSON feeds and sitemap
func TestHandlerGroup_FeedsAndSitemap(t *testing.T) {
	testCases := []TableDrivenTestCase{
		{
			Name: "RSS feed",
			SetupMocks: func(mocks *TestHandlerMocks) {
				articles := CreateTestArticlesWithVariations()
				mocks.CacheService.On("Get", "rss_feed").Return(nil, false)
				mocks.ArticleService.On("GetArticlesForFeed", 20).Return(articles)
				mocks.CacheService.On("Set", "rss_feed", mock.Anything, mock.Anything).Return()
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/rss", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
			ValidationFunc: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Header().Get("Content-Type"), "application/rss+xml")
			},
		},
		{
			Name: "JSON feed",
			SetupMocks: func(mocks *TestHandlerMocks) {
				articles := CreateTestArticlesWithVariations()
				mocks.CacheService.On("Get", "json_feed").Return(nil, false)
				mocks.ArticleService.On("GetArticlesForFeed", 20).Return(articles)
				mocks.CacheService.On("Set", "json_feed", mock.Anything, mock.Anything).Return()
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/feed.json", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
			ValidationFunc: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Header().Get("Content-Type"), "application/feed+json")
			},
		},
		{
			Name: "sitemap",
			SetupMocks: func(mocks *TestHandlerMocks) {
				articles := CreateTestArticlesWithVariations()
				mocks.CacheService.On("Get", "sitemap").Return(nil, false)
				mocks.ArticleService.On("GetAllArticles").Return(articles)
				mocks.CacheService.On("Set", "sitemap", mock.Anything, mock.Anything).Return()
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/sitemap.xml", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
			ValidationFunc: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Header().Get("Content-Type"), "application/xml")
			},
		},
	}

	RunTableDrivenTest(t, testCases, func(router *gin.Engine, handlers *Handlers) {
		router.GET("/rss", handlers.RSSFeed)
		router.GET("/feed.json", handlers.JSONFeed)
		router.GET("/sitemap.xml", handlers.Sitemap)
	})
}

// TestHandlerGroup_ErrorHandling tests error scenarios
func TestHandlerGroup_ErrorHandling(t *testing.T) {
	testCases := []TableDrivenTestCase{
		{
			Name: "article service error",
			SetupMocks: func(mocks *TestHandlerMocks) {
				mocks.CacheService.On("Get", "home_page").Return(nil, false)
				mocks.ArticleService.On("GetAllArticles").Return([]*models.Article{})
				mocks.ArticleService.On("GetFeaturedArticles", 3).Return([]*models.Article{})
				mocks.ArticleService.On("GetRecentArticles", 9).Return([]*models.Article{})
				mocks.ArticleService.On("GetCategoryCounts").Return([]models.CategoryCount{})
				mocks.ArticleService.On("GetTagCounts").Return([]models.TagCount{})
				mocks.CacheService.On("Set", "home_page", mock.Anything, mock.Anything).Return()
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "health check",
			SetupMocks: func(mocks *TestHandlerMocks) {
				// No mocks needed for health check
			},
			RequestFunc: func() *http.Request {
				req, _ := http.NewRequest("GET", "/health", nil)
				return req
			},
			ExpectedStatus: http.StatusOK,
			ValidationFunc: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				response := AssertJSONResponse(t, recorder, http.StatusOK)
				assert.Equal(t, "healthy", response["status"])
			},
		},
	}

	RunTableDrivenTest(t, testCases, func(router *gin.Engine, handlers *Handlers) {
		router.GET("/", handlers.Home)
		router.GET("/health", handlers.Health)
	})
}
