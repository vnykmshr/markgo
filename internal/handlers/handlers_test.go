package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/models"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// TestCacheAdapter provides a simple cache adapter for testing
type TestCacheAdapter struct {
	data map[string]any
}

func (t *TestCacheAdapter) Get(key string) (any, bool) {
	if t.data == nil {
		return nil, false
	}
	value, found := t.data[key]
	return value, found
}

func (t *TestCacheAdapter) Set(key string, value any, ttl time.Duration) {
	if t.data == nil {
		t.data = make(map[string]any)
	}
	t.data[key] = value
}

func (t *TestCacheAdapter) Delete(key string) {
	if t.data != nil {
		delete(t.data, key)
	}
}

func (t *TestCacheAdapter) Clear() {
	t.data = make(map[string]any)
}

func (t *TestCacheAdapter) Size() int {
	if t.data == nil {
		return 0
	}
	return len(t.data)
}

func (t *TestCacheAdapter) Stats() map[string]any {
	return map[string]any{
		"hit_count":  0,
		"miss_count": 0,
		"hit_ratio":  0.0,
		"total_keys": t.Size(),
		"cache_type": "test",
	}
}

// Test helper functions

func createTestHandlers() (*Handlers, *MockArticleService, *MockEmailService, *MockSearchService) {
	mockArticleService := &MockArticleService{}
	mockEmailService := &MockEmailService{}
	mockSearchService := &MockSearchService{}

	cfg := &config.Config{
		Blog: config.BlogConfig{
			Title:        "Test Blog",
			Description:  "A test blog",
			Author:       "Test Author",
			PostsPerPage: 10,
		},
		Cache: config.CacheConfig{
			TTL: time.Hour,
		},
	}

	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	handlers := New(&Config{
		ArticleService: mockArticleService,
		EmailService:   mockEmailService,
		SearchService:  mockSearchService,
		Config:         cfg,
		Logger:         logger,
		Cache:          nil, // Use nil cache for tests
	})

	// Use a simple cache adapter for testing admin functions
	handlers.cacheService = &TestCacheAdapter{}

	return handlers, mockArticleService, mockEmailService, mockSearchService
}

func createTestArticles() []*models.Article {
	return []*models.Article{
		{
			Slug:        "test-article-1",
			Title:       "Test Article 1",
			Description: "First test article",
			Date:        time.Now(),
			Tags:        []string{"test", "golang"},
			Categories:  []string{"programming"},
			Featured:    true,
			Draft:       false,
			Content:     "Test content 1",
		},
		{
			Slug:        "test-article-2",
			Title:       "Test Article 2",
			Description: "Second test article",
			Date:        time.Now().AddDate(0, 0, -1),
			Tags:        []string{"test", "web"},
			Categories:  []string{"development"},
			Featured:    false,
			Draft:       false,
			Content:     "Test content 2",
		},
	}
}

// Handler tests

func TestNew(t *testing.T) {
	handlers, _, _, _ := createTestHandlers()

	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.articleService)
	assert.NotNil(t, handlers.cacheService)
	assert.NotNil(t, handlers.emailService)
	assert.NotNil(t, handlers.searchService)
	assert.NotNil(t, handlers.config)
	assert.NotNil(t, handlers.logger)
}

func TestHome_CacheBehavior(t *testing.T) {
	// NOTE: With obcache integration, the cache behavior is now handled internally by obcache.
	// This test verifies that the home handler works correctly with the integrated cache system.
	testConfig, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	articles := CreateTestArticlesWithVariations()
	SetupArticleServiceMocks(testConfig.Mocks.ArticleService, articles)

	// Setup route
	testConfig.Router.GET("/", testConfig.Handlers.Home)

	// First request - will likely be a cache miss
	req1, _ := http.NewRequest("GET", "/", nil)
	recorder1 := ExecuteRequest(testConfig.Router, req1)
	assert.Equal(t, http.StatusOK, recorder1.Code)
	assert.Contains(t, recorder1.Body.String(), "Test Blog")

	// Second request - may be served from cache
	req2, _ := http.NewRequest("GET", "/", nil)
	recorder2 := ExecuteRequest(testConfig.Router, req2)
	assert.Equal(t, http.StatusOK, recorder2.Code)
	assert.Contains(t, recorder2.Body.String(), "Test Blog")

	// Both requests should return the same content
	assert.Equal(t, recorder1.Body.String(), recorder2.Body.String())
}

func TestArticle(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	article := createTestArticles()[0]

	mockArticleService.On("GetArticleBySlug", "test-article-1").Return(article, nil)
	mockArticleService.On("GetArticlesByTag", "test").Return([]*models.Article{})
	mockArticleService.On("GetArticlesByTag", "golang").Return([]*models.Article{})
	mockArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/articles/:slug", handlers.Article)

	req, _ := http.NewRequest("GET", "/articles/test-article-1", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestArticleNotFound(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	mockArticleService.On("GetArticleBySlug", "non-existent").Return(nil, assert.AnError)

	router := gin.New()
	setupMinimalTemplates(router)
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})
	router.GET("/articles/:slug", handlers.Article)

	req, _ := http.NewRequest("GET", "/articles/non-existent", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestArticlesByTag(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	articles := createTestArticles()

	mockArticleService.On("GetArticlesByTag", "golang").Return([]*models.Article{articles[0]})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/tags/:tag", handlers.ArticlesByTag)

	req, _ := http.NewRequest("GET", "/tags/golang", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestArticlesByCategory(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	articles := createTestArticles()

	mockArticleService.On("GetArticlesByCategory", "programming").Return([]*models.Article{articles[0]})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/categories/:category", handlers.ArticlesByCategory)

	req, _ := http.NewRequest("GET", "/categories/programming", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestTags(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	mockArticleService.On("GetTagCounts").Return([]models.TagCount{
		{Tag: "golang", Count: 5},
		{Tag: "web", Count: 3},
		{Tag: "test", Count: 2},
	})
	mockArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/tags", handlers.Tags)

	req, _ := http.NewRequest("GET", "/tags", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestCategories(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	mockArticleService.On("GetCategoryCounts").Return([]models.CategoryCount{
		{Category: "programming", Count: 5},
		{Category: "development", Count: 3},
	})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/categories", handlers.Categories)

	req, _ := http.NewRequest("GET", "/categories", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestSearch(t *testing.T) {
	handlers, mockArticleService, _, mockSearchService := createTestHandlers()
	articles := createTestArticles()
	searchResults := []*models.SearchResult{
		{
			Article:       articles[0],
			Score:         10.5,
			MatchedFields: []string{"title", "content"},
		},
	}

	mockArticleService.On("GetAllArticles").Return(articles)
	mockSearchService.On("Search", articles, "golang", 20).Return(searchResults)
	mockArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/search", handlers.Search)

	req, _ := http.NewRequest("GET", "/search?q=golang", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
	mockSearchService.AssertExpectations(t)
}

func TestSearchEmptyQuery(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	mockArticleService.On("GetAllArticles").Return([]*models.Article{})
	mockArticleService.On("GetTagCounts").Return([]models.TagCount{})
	mockArticleService.On("GetCategoryCounts").Return([]models.CategoryCount{})
	mockArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/search", handlers.Search)

	req, _ := http.NewRequest("GET", "/search?q=", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestContactForm(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	mockArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/contact", handlers.ContactForm)

	req, _ := http.NewRequest("GET", "/contact", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestContactSubmit(t *testing.T) {
	tests := []struct {
		name           string
		contactData    map[string]any
		expectedStatus int
		setupMocks     func(*MockEmailService)
	}{
		{
			name: "success",
			contactData: map[string]any{
				"name":             "John Doe",
				"email":            "john@example.com",
				"subject":          "Test Message",
				"message":          "This is a test message with sufficient length",
				"captcha_question": "3 + 5",
				"captcha_answer":   "8",
			},
			expectedStatus: http.StatusOK,
			setupMocks: func(mockEmailService *MockEmailService) {
				mockEmailService.On("SendContactMessage", mock.AnythingOfType("*models.ContactMessage")).Return(nil)
			},
		},
		{
			name: "invalid captcha",
			contactData: map[string]any{
				"name":             "John Doe",
				"email":            "john@example.com",
				"subject":          "Test Message",
				"message":          "This is a test message with sufficient length",
				"captcha_question": "3 + 5",
				"captcha_answer":   "wrong",
			},
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(mockEmailService *MockEmailService) {},
		},
		{
			name: "invalid email",
			contactData: map[string]any{
				"name":             "John Doe",
				"email":            "invalid-email",
				"subject":          "Test Message",
				"message":          "This is a test message with sufficient length",
				"captcha_question": "3 + 5",
				"captcha_answer":   "8",
			},
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(mockEmailService *MockEmailService) {},
		},
		{
			name: "missing name",
			contactData: map[string]any{
				"email":            "john@example.com",
				"subject":          "Test Message",
				"message":          "This is a test message with sufficient length",
				"captcha_question": "3 + 5",
				"captcha_answer":   "8",
			},
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(mockEmailService *MockEmailService) {},
		},
		{
			name: "missing email",
			contactData: map[string]any{
				"name":             "John Doe",
				"subject":          "Test Message",
				"message":          "This is a test message with sufficient length",
				"captcha_question": "3 + 5",
				"captcha_answer":   "8",
			},
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(mockEmailService *MockEmailService) {},
		},
		{
			name: "missing message",
			contactData: map[string]any{
				"name":             "John Doe",
				"email":            "john@example.com",
				"subject":          "Test Message",
				"captcha_question": "3 + 5",
				"captcha_answer":   "8",
			},
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(mockEmailService *MockEmailService) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, _, mockEmailService, _ := createTestHandlers()
			tt.setupMocks(mockEmailService)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Next()
				if len(c.Errors) > 0 {
					err := c.Errors.Last()
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				}
			})
			router.POST("/contact", handlers.ContactSubmit)

			jsonData, _ := json.Marshal(tt.contactData)
			req, _ := http.NewRequest("POST", "/contact", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			mockEmailService.AssertExpectations(t)
		})
	}
}

func TestHealth(t *testing.T) {
	handlers, _, _, _ := createTestHandlers()

	router := gin.New()
	router.GET("/health", handlers.Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestMetrics(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	mockArticleService.On("GetStats").Return(&models.Stats{
		TotalArticles:  10,
		PublishedCount: 8,
		DraftCount:     2,
	})

	router := gin.New()
	router.GET("/metrics", handlers.Metrics)

	req, _ := http.NewRequest("GET", "/metrics", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "blog")
	assert.Contains(t, response, "cache")
	mockArticleService.AssertExpectations(t)
}

func TestClearCache(t *testing.T) {
	handlers, _, _, _ := createTestHandlers()

	router := gin.New()
	router.POST("/admin/cache/clear", handlers.ClearCache)

	req, _ := http.NewRequest("POST", "/admin/cache/clear", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Cache cleared successfully", response["message"])
}

func TestAdminStats(t *testing.T) {
	handlers, mockArticleService, mockEmailService, _ := createTestHandlers()

	mockArticleService.On("GetStats").Return(&models.Stats{
		TotalArticles:  10,
		PublishedCount: 8,
		DraftCount:     2,
	})
	mockEmailService.On("GetConfig").Return(map[string]any{
		"host": "smtp.example.com",
	})

	router := gin.New()
	router.GET("/admin/stats", handlers.AdminStats)

	req, _ := http.NewRequest("GET", "/admin/stats", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "articles")
	assert.Contains(t, response, "cache")
	mockArticleService.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestVerifyCaptcha(t *testing.T) {
	handlers, _, _, _ := createTestHandlers()

	testCases := []struct {
		name     string
		question string
		answer   string
		expected bool
	}{
		{
			name:     "Valid addition",
			question: "3 + 5",
			answer:   "8",
			expected: true,
		},
		{
			name:     "Wrong answer",
			question: "3 + 5",
			answer:   "7",
			expected: false,
		},
		{
			name:     "Invalid question format - no spaces",
			question: "3+5",
			answer:   "8",
			expected: false,
		},
		{
			name:     "Invalid question format - wrong operator",
			question: "3 - 5",
			answer:   "8",
			expected: false,
		},
		{
			name:     "Invalid answer format",
			question: "3 + 5",
			answer:   "abc",
			expected: false,
		},
		{
			name:     "Empty answer",
			question: "3 + 5",
			answer:   "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := handlers.verifyCaptcha(tc.question, tc.answer)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestReloadArticles(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	mockArticleService.On("ReloadArticles").Return(nil)

	router := gin.New()
	router.POST("/admin/articles/reload", handlers.ReloadArticles)

	req, _ := http.NewRequest("POST", "/admin/articles/reload", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Articles reloaded successfully", response["message"])
	mockArticleService.AssertExpectations(t)
}

func TestReloadArticlesError(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	mockArticleService.On("ReloadArticles").Return(assert.AnError)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
	})
	router.POST("/admin/articles/reload", handlers.ReloadArticles)

	req, _ := http.NewRequest("POST", "/admin/articles/reload", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestNotFound(t *testing.T) {
	handlers, _, _, _ := createTestHandlers()

	router := gin.New()
	setupMinimalTemplates(router)
	router.NoRoute(handlers.NotFound)

	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestRSSFeed(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	articles := createTestArticles()

	mockArticleService.On("GetArticlesForFeed", 20).Return(articles)

	router := gin.New()
	router.GET("/feed.xml", handlers.RSSFeed)

	req, _ := http.NewRequest("GET", "/feed.xml", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/rss+xml; charset=utf-8", recorder.Header().Get("Content-Type"))
	mockArticleService.AssertExpectations(t)
}

func TestJSONFeed(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	articles := createTestArticles()

	mockArticleService.On("GetArticlesForFeed", 20).Return(articles)

	router := gin.New()
	router.GET("/feed.json", handlers.JSONFeed)

	req, _ := http.NewRequest("GET", "/feed.json", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/feed+json; charset=utf-8", recorder.Header().Get("Content-Type"))
	mockArticleService.AssertExpectations(t)
}

func TestSitemap(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	articles := createTestArticles()

	mockArticleService.On("GetAllArticles").Return(articles, nil)

	router := gin.New()
	router.GET("/sitemap.xml", handlers.Sitemap)

	req, _ := http.NewRequest("GET", "/sitemap.xml", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/xml; charset=utf-8", recorder.Header().Get("Content-Type"))
	mockArticleService.AssertExpectations(t)
}

func TestAboutArticle(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	aboutArticle := &models.Article{
		Slug:    "about",
		Title:   "About",
		Content: "About page content",
	}

	articles := createTestArticles()
	mockArticleService.On("GetArticleBySlug", "about").Return(aboutArticle, nil)
	mockArticleService.On("GetRecentArticles", 5).Return(articles)

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/about", handlers.AboutArticle)

	req, _ := http.NewRequest("GET", "/about", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestAboutArticleNotFound(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	mockArticleService.On("GetArticleBySlug", "about").Return(nil, assert.AnError)

	router := gin.New()
	setupMinimalTemplates(router)
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})
	router.GET("/about", handlers.AboutArticle)

	req, _ := http.NewRequest("GET", "/about", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestArticles(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	articles := createTestArticles()

	mockArticleService.On("GetAllArticles").Return(articles, nil)
	mockArticleService.On("GetRecentArticles", 9).Return(articles)

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/articles", handlers.Articles)

	req, _ := http.NewRequest("GET", "/articles", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestArticlesWithPagination(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	articles := createTestArticles()

	mockArticleService.On("GetAllArticles").Return(articles, nil)
	mockArticleService.On("GetRecentArticles", 9).Return(articles)

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/articles", handlers.Articles)

	req, _ := http.NewRequest("GET", "/articles?page=2", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkHome(b *testing.B) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	articles := createTestArticles()

	mockArticleService.On("GetAllArticles").Return(articles)
	mockArticleService.On("GetFeaturedArticles", 3).Return([]*models.Article{articles[0]})
	mockArticleService.On("GetRecentArticles", 5).Return(articles)
	mockArticleService.On("GetCategoryCounts").Return([]models.CategoryCount{})
	mockArticleService.On("GetTagCounts").Return([]models.TagCount{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/", handlers.Home)

	req, _ := http.NewRequest("GET", "/", nil)

	for b.Loop() {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

func BenchmarkArticle(b *testing.B) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	article := createTestArticles()[0]

	mockArticleService.On("GetArticleBySlug", "test-article-1").Return(article, nil)
	mockArticleService.On("GetArticlesByTag", mock.Anything).Return([]*models.Article{})
	mockArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/articles/:slug", handlers.Article)

	req, _ := http.NewRequest("GET", "/articles/test-article-1", nil)

	for b.Loop() {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

func BenchmarkSearch(b *testing.B) {
	handlers, mockArticleService, _, mockSearchService := createTestHandlers()
	articles := createTestArticles()

	mockArticleService.On("GetAllArticles").Return(articles)
	mockSearchService.On("Search", articles, "golang", 20).Return([]*models.SearchResult{})
	mockArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/search", handlers.Search)

	req, _ := http.NewRequest("GET", "/search?q=golang", nil)

	for b.Loop() {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

// Integration tests
func TestHandlerIntegration(t *testing.T) {
	// Test that handlers work together in a realistic scenario
	handlers, mockArticleService, _, mockSearchService := createTestHandlers()
	articles := createTestArticles()

	// Set up expectations for a typical user journey
	mockArticleService.On("GetAllArticles").Return(articles).Maybe()
	mockArticleService.On("GetArticleBySlug", mock.Anything).Return(articles[0], nil).Maybe()
	mockArticleService.On("GetFeaturedArticles", mock.Anything).Return([]*models.Article{articles[0]}).Maybe()
	mockArticleService.On("GetRecentArticles", mock.Anything).Return(articles).Maybe()
	mockArticleService.On("GetTagCounts").Return([]models.TagCount{}).Maybe()
	mockArticleService.On("GetCategoryCounts").Return([]models.CategoryCount{}).Maybe()
	mockSearchService.On("Search", mock.Anything, mock.Anything, mock.Anything).Return([]*models.SearchResult{}).Maybe()

	router := gin.New()
	setupMinimalTemplates(router)

	// Set up routes
	router.GET("/", handlers.Home)
	router.GET("/articles/:slug", handlers.Article)
	router.GET("/search", handlers.Search)
	router.GET("/health", handlers.Health)

	// Test home page
	req, _ := http.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Test health check
	req, _ = http.NewRequest("GET", "/health", nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Test search
	req, _ = http.NewRequest("GET", "/search?q=test", nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

// Error handling tests
func TestHandlerErrorHandling(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()

	// Test when article service fails
	mockArticleService.On("GetAllArticles").Panic("database error")

	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware for panic handling
	setupMinimalTemplates(router)
	router.GET("/", handlers.Home)

	req, _ := http.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should recover from panic and return 500
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// Template rendering tests
func TestHandlerTemplateRendering(t *testing.T) {
	handlers, mockArticleService, _, _ := createTestHandlers()
	articles := createTestArticles()

	// Only set up article service expectations since this test is for template rendering
	mockArticleService.On("GetAllArticles").Return(articles)
	mockArticleService.On("GetFeaturedArticles", 3).Return([]*models.Article{articles[0]})
	mockArticleService.On("GetRecentArticles", 9).Return(articles)
	mockArticleService.On("GetCategoryCounts").Return([]models.CategoryCount{})
	mockArticleService.On("GetTagCounts").Return([]models.TagCount{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/", handlers.Home)

	req, _ := http.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Content-Type"), "text/html")
	mockArticleService.AssertExpectations(t)
}

// Additional Error Handling and Edge Case Tests

func TestArticle_ShouldReturnError_WhenSlugEmpty(t *testing.T) {
	handlers, _, _, _ := createTestHandlers()

	// No cache expectation needed since route won't match

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/articles/:slug", handlers.Article)

	req, _ := http.NewRequest("GET", "/articles/", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	// No expectations to assert since route doesn't match
}

func TestSearch_ShouldReturnError_WhenQueryTooLong(t *testing.T) {
	handlers, mockArticleService, _, mockSearchService := createTestHandlers()

	// Create a very long query string (over 1000 characters)
	longQuery := strings.Repeat("a", 1001)

	mockArticleService.On("GetAllArticles").Return([]*models.Article{})
	mockSearchService.On("Search", mock.Anything, longQuery, 20).Return([]*models.SearchResult{})
	mockArticleService.On("GetTagCounts").Return([]models.TagCount{}).Maybe()
	mockArticleService.On("GetCategoryCounts").Return([]models.CategoryCount{}).Maybe()
	mockArticleService.On("GetRecentArticles", 5).Return([]*models.Article{})

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/search", handlers.Search)

	// Use the same long query string
	req, _ := http.NewRequest("GET", "/search?q="+longQuery, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should handle gracefully without crashing
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockArticleService.AssertExpectations(t)
}

func TestHandlers_ShouldHandleNilInputs(t *testing.T) {
	handlers, _, _, _ := createTestHandlers()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		}
	})
	router.POST("/contact", handlers.ContactSubmit)

	// Test with nil/empty body
	req, _ := http.NewRequest("POST", "/contact", nil)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
