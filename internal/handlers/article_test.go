package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
)

// Enhanced MockArticleService with more realistic data for article tests
type EnhancedMockArticleService struct {
	articles []*models.Article
}

func (m *EnhancedMockArticleService) GetAllArticles() []*models.Article {
	return m.articles
}

func (m *EnhancedMockArticleService) GetArticleBySlug(slug string) (*models.Article, error) {
	for _, article := range m.articles {
		if article.Slug == slug {
			return article, nil
		}
	}
	return nil, apperrors.ErrArticleNotFound
}

func (m *EnhancedMockArticleService) GetArticlesByTag(tag string) []*models.Article {
	var filtered []*models.Article
	for _, article := range m.articles {
		for _, t := range article.Tags {
			if t == tag {
				filtered = append(filtered, article)
				break
			}
		}
	}
	return filtered
}

func (m *EnhancedMockArticleService) GetArticlesByCategory(category string) []*models.Article {
	var filtered []*models.Article
	for _, article := range m.articles {
		for _, c := range article.Categories {
			if c == category {
				filtered = append(filtered, article)
				break
			}
		}
	}
	return filtered
}

func (m *EnhancedMockArticleService) GetArticlesForFeed(limit int) []*models.Article {
	if limit > len(m.articles) {
		limit = len(m.articles)
	}
	return m.articles[:limit]
}

func (m *EnhancedMockArticleService) GetFeaturedArticles(limit int) []*models.Article {
	return m.articles
}

func (m *EnhancedMockArticleService) GetRecentArticles(limit int) []*models.Article {
	if limit > len(m.articles) {
		limit = len(m.articles)
	}
	return m.articles[:limit]
}

func (m *EnhancedMockArticleService) GetAllTags() []string {
	tagSet := make(map[string]bool)
	for _, article := range m.articles {
		for _, tag := range article.Tags {
			tagSet[tag] = true
		}
	}
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

func (m *EnhancedMockArticleService) GetAllCategories() []string {
	catSet := make(map[string]bool)
	for _, article := range m.articles {
		for _, cat := range article.Categories {
			catSet[cat] = true
		}
	}
	cats := make([]string, 0, len(catSet))
	for cat := range catSet {
		cats = append(cats, cat)
	}
	return cats
}

func (m *EnhancedMockArticleService) GetTagCounts() []models.TagCount {
	counts := make(map[string]int)
	for _, article := range m.articles {
		for _, tag := range article.Tags {
			counts[tag]++
		}
	}
	tagCounts := make([]models.TagCount, 0, len(counts))
	for tag, count := range counts {
		tagCounts = append(tagCounts, models.TagCount{Tag: tag, Count: count})
	}
	return tagCounts
}

func (m *EnhancedMockArticleService) GetCategoryCounts() []models.CategoryCount {
	counts := make(map[string]int)
	for _, article := range m.articles {
		for _, cat := range article.Categories {
			counts[cat]++
		}
	}
	catCounts := make([]models.CategoryCount, 0, len(counts))
	for cat, count := range counts {
		catCounts = append(catCounts, models.CategoryCount{Category: cat, Count: count})
	}
	return catCounts
}

func (m *EnhancedMockArticleService) GetStats() *models.Stats {
	return &models.Stats{
		TotalArticles:   len(m.articles),
		TotalTags:       len(m.GetAllTags()),
		TotalCategories: len(m.GetAllCategories()),
	}
}

func (m *EnhancedMockArticleService) SearchArticles(query string, limit int) []*models.SearchResult {
	if query == "" {
		return []*models.SearchResult{}
	}
	var results []*models.SearchResult
	for _, article := range m.articles {
		if !article.Draft && len(results) < limit {
			results = append(results, &models.SearchResult{Article: article, Score: 1.0})
		}
	}
	return results
}

func (m *EnhancedMockArticleService) ReloadArticles() error {
	return nil
}

func (m *EnhancedMockArticleService) GetDraftArticles() []*models.Article {
	return nil
}

func (m *EnhancedMockArticleService) GetDraftBySlug(_ string) (*models.Article, error) {
	return nil, apperrors.ErrArticleNotFound
}

func createTestArticles() []*models.Article {
	now := time.Now()
	return []*models.Article{
		{
			Slug:        "golang-tutorial",
			Title:       "Getting Started with Go",
			Description: "A beginner's guide to Go programming",
			Content:     "Content about Go programming...",
			Date:        now,
			Draft:       false,
			Tags:        []string{"golang", "tutorial", "programming"},
			Categories:  []string{"Programming", "Tutorials"},
		},
		{
			Slug:        "web-development",
			Title:       "Modern Web Development",
			Description: "Building web applications in 2025",
			Content:     "Content about web development...",
			Date:        now.Add(-24 * time.Hour),
			Draft:       false,
			Tags:        []string{"web", "javascript", "tutorial"},
			Categories:  []string{"Web Development"},
		},
		{
			Slug:        "draft-article",
			Title:       "Draft Article",
			Description: "This is a draft",
			Content:     "Draft content...",
			Date:        now,
			Draft:       true,
			Tags:        []string{"draft"},
			Categories:  []string{"Uncategorized"},
		},
	}
}

func createTestBase() (*BaseHandler, *EnhancedMockArticleService) {
	cfg := &config.Config{
		Environment: "test",
		BaseURL:     "http://localhost:3000",
		Blog: config.BlogConfig{
			Title:       "Test Blog",
			Description: "Test Description",
			Author:      "Test Author",
		},
	}

	articles := createTestArticles()
	mockArticleService := &EnhancedMockArticleService{articles: articles}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

	return base, mockArticleService
}

// TestArticleBySlug tests individual article viewing
func TestArticleBySlug(t *testing.T) {
	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "valid article slug",
			slug:           "golang-tutorial",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "article not found",
			slug:           "nonexistent-article",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "empty slug",
			slug:           "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, svc := createTestBase()
			handler := NewPostHandler(base, svc)
			router := gin.New()
			router.GET("/articles/:slug", handler.Article)

			req := httptest.NewRequest("GET", "/articles/"+tt.slug, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			tt.checkResponse(t, w)
		})
	}
}

// TestArticlesListing tests the articles listing page
func TestArticlesListing(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{
			name:           "default page",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "page 1",
			query:          "?page=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "page 2",
			query:          "?page=2",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid page number",
			query:          "?page=invalid",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "negative page number",
			query:          "?page=-1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, svc := createTestBase()
			handler := NewPostHandler(base, svc)
			router := gin.New()
			router.GET("/articles", handler.Articles)

			req := httptest.NewRequest("GET", "/articles"+tt.query, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestArticlesByTag tests tag filtering
func TestArticlesByTag(t *testing.T) {
	tests := []struct {
		name           string
		tag            string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "valid tag",
			tag:            "golang",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "tag with multiple articles",
			tag:            "tutorial",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "nonexistent tag",
			tag:            "nonexistent",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "URL encoded tag",
			tag:            "web%20development",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, svc := createTestBase()
			handler := NewTaxonomyHandler(base, svc)
			router := gin.New()
			router.GET("/tags/:tag", handler.ArticlesByTag)

			req := httptest.NewRequest("GET", "/tags/"+tt.tag, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			tt.checkResponse(t, w)
		})
	}
}

// TestArticlesByCategory tests category filtering
func TestArticlesByCategory(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "valid category",
			category:       "Programming",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "valid category - web development",
			category:       url.PathEscape("Web Development"),
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "nonexistent category",
			category:       "Nonexistent",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "URL encoded category",
			category:       url.PathEscape("Tutorials"),
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, svc := createTestBase()
			handler := NewTaxonomyHandler(base, svc)
			router := gin.New()
			router.GET("/categories/:category", handler.ArticlesByCategory)

			req := httptest.NewRequest("GET", "/categories/"+tt.category, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			tt.checkResponse(t, w)
		})
	}
}

// TestSearch tests the search functionality
func TestSearch(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "search with query",
			query:          "?q=golang",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "search with empty query",
			query:          "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "search with special characters",
			query:          "?q=go+programming",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name:           "search with long query",
			query:          "?q=thisisaverylongquerystringthatmightcauseissues",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, svc := createTestBase()
			handler := NewSearchHandler(base, svc)
			router := gin.New()
			router.GET("/search", handler.Search)

			req := httptest.NewRequest("GET", "/search"+tt.query, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			tt.checkResponse(t, w)
		})
	}
}

// TestHomePage tests the home page handler
func TestHomePage(t *testing.T) {
	base, svc := createTestBase()
	handler := NewFeedHandler(base, svc)
	router := gin.New()
	router.GET("/", handler.Home)

	req := httptest.NewRequest("GET", "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTagsPage tests the tags listing page
func TestTagsPage(t *testing.T) {
	base, svc := createTestBase()
	handler := NewTaxonomyHandler(base, svc)
	router := gin.New()
	router.GET("/tags", handler.Tags)

	req := httptest.NewRequest("GET", "/tags", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestCategoriesPage tests the categories listing page
func TestCategoriesPage(t *testing.T) {
	base, svc := createTestBase()
	handler := NewTaxonomyHandler(base, svc)
	router := gin.New()
	router.GET("/categories", handler.Categories)

	req := httptest.NewRequest("GET", "/categories", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
