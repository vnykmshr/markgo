package handlers

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
)

// TestSetup provides minimal test infrastructure
type TestSetup struct {
	handlers *Handlers
	router   *gin.Engine
	articles []*models.Article
}

// NewTestSetup creates a minimal test environment with real services
func NewTestSetup(t *testing.T) *TestSetup {
	gin.SetMode(gin.TestMode)

	// Create minimal config
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
		Blog: config.BlogConfig{
			Title:        "Test Blog",
			Description:  "Test Description",
			Author:       "Test Author",
			PostsPerPage: 5,
			Language:     "en",
		},
		Cache: config.CacheConfig{
			TTL:     time.Hour,
			MaxSize: 1000,
		},
	}

	// Create silent logger for tests
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), &slog.HandlerOptions{
		Level: slog.LevelError, // Only log errors in tests
	}))

	// Create test articles
	articles := []*models.Article{
		{
			Slug:        "test-article-1",
			Title:       "Test Article 1",
			Content:     "Test content 1",
			Tags:        []string{"test", "golang"},
			Categories:  []string{"tech"},
			Date:        time.Now().AddDate(0, 0, -1),
			Draft:       false,
			Featured:    true,
			Description: "Test description 1",
		},
		{
			Slug:        "test-article-2",
			Title:       "Test Article 2",
			Content:     "Test content 2",
			Tags:        []string{"test", "web"},
			Categories:  []string{"web"},
			Date:        time.Now().AddDate(0, 0, -2),
			Draft:       false,
			Featured:    false,
			Description: "Test description 2",
		},
	}

	// Create in-memory services (no mocking needed)
	articleService := NewInMemoryArticleService(articles)
	emailService := NewInMemoryEmailService()
	searchService := NewInMemorySearchService()

	// Create handlers with real services
	handlers := New(&Config{
		ArticleService: articleService,
		EmailService:   emailService,
		SearchService:  searchService,
		Config:         cfg,
		Logger:         logger,
		Cache:          nil, // Tests run without cache for predictability
	})

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Add essential routes (avoid HTML rendering routes for core tests)
	router.GET("/health", handlers.Health)

	// Add data-only routes for testing data generation
	router.GET("/api/home-data", func(c *gin.Context) {
		data, err := handlers.GetHomeDataUncached()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	router.GET("/api/article-data/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		data, err := handlers.GetArticleDataUncached(slug)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	router.GET("/api/articles-page", func(c *gin.Context) {
		page := 1
		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil {
				page = p
			}
		}
		data, err := handlers.GetArticlesPageUncached(page)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	router.GET("/api/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
			return
		}
		data, err := handlers.GetSearchResultsUncached(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	return &TestSetup{
		handlers: handlers,
		router:   router,
		articles: articles,
	}
}

// Helper methods for common test operations
func (ts *TestSetup) GET(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)
	return w
}

func (ts *TestSetup) POST(path string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)
	return w
}

// TestHandlers_CoreEndpoints tests essential functionality with minimal setup
func TestHandlers_CoreEndpoints(t *testing.T) {
	setup := NewTestSetup(t)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "health check",
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"healthy"`,
		},
		{
			name:           "home page data",
			path:           "/api/home-data",
			expectedStatus: http.StatusOK,
			expectedBody:   "Test Blog",
		},
		{
			name:           "articles page data",
			path:           "/api/articles-page",
			expectedStatus: http.StatusOK,
			expectedBody:   "Test Article 1",
		},
		{
			name:           "specific article data",
			path:           "/api/article-data/test-article-1",
			expectedStatus: http.StatusOK,
			expectedBody:   "Test Article 1",
		},
		{
			name:           "search with query",
			path:           "/api/search?q=test",
			expectedStatus: http.StatusOK,
			expectedBody:   "Test Article",
		},
		{
			name:           "article not found",
			path:           "/api/article-data/nonexistent",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "article not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := setup.GET(tt.path)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

// TestHandlers_ErrorHandling tests error scenarios
func TestHandlers_ErrorHandling(t *testing.T) {
	setup := NewTestSetup(t)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{"empty search query", "/api/search", http.StatusBadRequest},
		{"malformed search", "/api/search?q=" + strings.Repeat("x", 1000), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := setup.GET(tt.path)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestHandlers_CachedFunctions tests the cached function behavior
func TestHandlers_CachedFunctions(t *testing.T) {
	setup := NewTestSetup(t)

	t.Run("home data generation", func(t *testing.T) {
		// Test the uncached function directly
		data, err := setup.handlers.GetHomeDataUncached()
		require.NoError(t, err)
		assert.NotNil(t, data)

		// Verify data structure
		assert.Contains(t, data, "totalCount")
		assert.Contains(t, data, "featured")
		assert.Contains(t, data, "recent")
	})

	t.Run("article data generation", func(t *testing.T) {
		data, err := setup.handlers.GetArticleDataUncached("test-article-1")
		require.NoError(t, err)
		assert.NotNil(t, data)

		// Verify article data structure
		assert.Contains(t, data, "article")
		assert.Contains(t, data, "recent")
		assert.Contains(t, data, "relatedArticles")
	})

	t.Run("articles page data generation", func(t *testing.T) {
		data, err := setup.handlers.GetArticlesPageUncached(1)
		require.NoError(t, err)
		assert.NotNil(t, data)

		// Verify pagination data
		assert.Contains(t, data, "articles")
		assert.Contains(t, data, "pagination")
		assert.Contains(t, data, "recent")
	})
}

// In-memory service implementations for testing (no mocking needed)

type InMemoryArticleService struct {
	articles []*models.Article
}

func NewInMemoryArticleService(articles []*models.Article) *InMemoryArticleService {
	return &InMemoryArticleService{articles: articles}
}

func (s *InMemoryArticleService) GetAllArticles() []*models.Article {
	var published []*models.Article
	for _, article := range s.articles {
		if !article.Draft {
			published = append(published, article)
		}
	}
	return published
}

func (s *InMemoryArticleService) GetArticleBySlug(slug string) (*models.Article, error) {
	for _, article := range s.articles {
		if article.Slug == slug && !article.Draft {
			return article, nil
		}
	}
	return nil, errors.ErrArticleNotFound
}

func (s *InMemoryArticleService) GetFeaturedArticles(limit int) []*models.Article {
	var featured []*models.Article
	for _, article := range s.articles {
		if article.Featured && !article.Draft {
			featured = append(featured, article)
			if len(featured) >= limit {
				break
			}
		}
	}
	return featured
}

func (s *InMemoryArticleService) GetRecentArticles(limit int) []*models.Article {
	published := s.GetAllArticles()
	if len(published) <= limit {
		return published
	}
	return published[:limit]
}

func (s *InMemoryArticleService) GetArticlesByTag(tag string) []*models.Article {
	var result []*models.Article
	for _, article := range s.articles {
		if !article.Draft {
			for _, t := range article.Tags {
				if t == tag {
					result = append(result, article)
					break
				}
			}
		}
	}
	return result
}

func (s *InMemoryArticleService) GetArticlesByCategory(category string) []*models.Article {
	var result []*models.Article
	for _, article := range s.articles {
		if !article.Draft {
			for _, cat := range article.Categories {
				if cat == category {
					result = append(result, article)
					break
				}
			}
		}
	}
	return result
}

func (s *InMemoryArticleService) GetAllTags() []string {
	tagSet := make(map[string]bool)
	for _, article := range s.articles {
		if !article.Draft {
			for _, tag := range article.Tags {
				tagSet[tag] = true
			}
		}
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

func (s *InMemoryArticleService) GetAllCategories() []string {
	catSet := make(map[string]bool)
	for _, article := range s.articles {
		if !article.Draft {
			for _, cat := range article.Categories {
				catSet[cat] = true
			}
		}
	}

	var categories []string
	for cat := range catSet {
		categories = append(categories, cat)
	}
	return categories
}

func (s *InMemoryArticleService) GetTagCounts() []models.TagCount {
	counts := make(map[string]int)
	for _, article := range s.articles {
		if !article.Draft {
			for _, tag := range article.Tags {
				counts[tag]++
			}
		}
	}

	var result []models.TagCount
	for tag, count := range counts {
		result = append(result, models.TagCount{Tag: tag, Count: count})
	}
	return result
}

func (s *InMemoryArticleService) GetCategoryCounts() []models.CategoryCount {
	counts := make(map[string]int)
	for _, article := range s.articles {
		if !article.Draft {
			for _, cat := range article.Categories {
				counts[cat]++
			}
		}
	}

	var result []models.CategoryCount
	for cat, count := range counts {
		result = append(result, models.CategoryCount{Category: cat, Count: count})
	}
	return result
}

func (s *InMemoryArticleService) GetArticlesForFeed(limit int) []*models.Article {
	return s.GetRecentArticles(limit)
}

func (s *InMemoryArticleService) ReloadArticles() error {
	return nil
}

func (s *InMemoryArticleService) GetDraftBySlug(slug string) (*models.Article, error) {
	for _, article := range s.articles {
		if article.Slug == slug && article.Draft {
			return article, nil
		}
	}
	return nil, errors.ErrArticleNotFound
}

func (s *InMemoryArticleService) PreviewDraft(slug string) (*models.Article, error) {
	return s.GetDraftBySlug(slug)
}

func (s *InMemoryArticleService) PublishDraft(slug string) error {
	for _, article := range s.articles {
		if article.Slug == slug && article.Draft {
			article.Draft = false
			return nil
		}
	}
	return errors.ErrArticleNotFound
}

func (s *InMemoryArticleService) UnpublishArticle(slug string) error {
	for _, article := range s.articles {
		if article.Slug == slug && !article.Draft {
			article.Draft = true
			return nil
		}
	}
	return errors.ErrArticleNotFound
}

func (s *InMemoryArticleService) GetStats() *models.Stats {
	all := s.GetAllArticles()
	return &models.Stats{
		TotalArticles:   len(all),
		TotalTags:       len(s.GetTagCounts()),
		TotalCategories: len(s.GetCategoryCounts()),
	}
}

func (s *InMemoryArticleService) GetDraftArticles() []*models.Article {
	var drafts []*models.Article
	for _, article := range s.articles {
		if article.Draft {
			drafts = append(drafts, article)
		}
	}
	return drafts
}

func (s *InMemoryArticleService) Reload() error { return nil }
func (s *InMemoryArticleService) ClearCache()   {}

type InMemoryEmailService struct{}

func NewInMemoryEmailService() *InMemoryEmailService {
	return &InMemoryEmailService{}
}

func (s *InMemoryEmailService) SendContactMessage(msg *models.ContactMessage) error {
	return nil
}

func (s *InMemoryEmailService) SendNotification(to, subject, body string) error {
	return nil
}

func (s *InMemoryEmailService) SendTestEmail() error {
	return nil
}

func (s *InMemoryEmailService) TestConnection() error {
	return nil
}

func (s *InMemoryEmailService) ValidateConfig() []string {
	return nil
}

func (s *InMemoryEmailService) GetConfig() map[string]any {
	return map[string]any{"test": true}
}

func (s *InMemoryEmailService) Shutdown() {}

type InMemorySearchService struct{}

func NewInMemorySearchService() *InMemorySearchService {
	return &InMemorySearchService{}
}

func (s *InMemorySearchService) Search(articles []*models.Article, query string, limit int) []*models.SearchResult {
	var results []*models.SearchResult
	query = strings.ToLower(query)

	for _, article := range articles {
		if strings.Contains(strings.ToLower(article.Title), query) ||
			strings.Contains(strings.ToLower(article.Content), query) {
			results = append(results, &models.SearchResult{
				Article:       article,
				Score:         1.0,
				MatchedFields: []string{"title", "content"},
			})
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

func (s *InMemorySearchService) SearchInTitle(articles []*models.Article, query string, limit int) []*models.SearchResult {
	var results []*models.SearchResult
	query = strings.ToLower(query)

	for _, article := range articles {
		if strings.Contains(strings.ToLower(article.Title), query) {
			results = append(results, &models.SearchResult{
				Article:       article,
				Score:         1.0,
				MatchedFields: []string{"title", "content"},
			})
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

func (s *InMemorySearchService) SearchByTag(articles []*models.Article, tag string) []*models.Article {
	var results []*models.Article
	for _, article := range articles {
		for _, t := range article.Tags {
			if t == tag {
				results = append(results, article)
				break
			}
		}
	}
	return results
}

func (s *InMemorySearchService) SearchByCategory(articles []*models.Article, category string) []*models.Article {
	var results []*models.Article
	for _, article := range articles {
		for _, cat := range article.Categories {
			if cat == category {
				results = append(results, article)
				break
			}
		}
	}
	return results
}

func (s *InMemorySearchService) GetSuggestions(articles []*models.Article, query string, limit int) []string {
	var suggestions []string
	query = strings.ToLower(query)

	// Simple suggestion logic for testing
	for _, article := range articles {
		if strings.Contains(strings.ToLower(article.Title), query) {
			suggestions = append(suggestions, article.Title)
			if len(suggestions) >= limit {
				break
			}
		}
	}
	return suggestions
}

// TestHandlers_ShouldHandleNilInputs tests handling of nil inputs
func TestHandlers_ModernErrorHandling(t *testing.T) {
	setup := NewTestSetup(t)

	// Test with non-existent article
	_, err := setup.handlers.GetArticleDataUncached("nonexistent")
	assert.Error(t, err)
}
