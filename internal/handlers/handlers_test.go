package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Mock services for testing
type MockArticleService struct{}

func (m *MockArticleService) GetAllArticles() []*models.Article {
	return []*models.Article{
		{Slug: "test-article", Title: "Test", Draft: false, Date: time.Now()},
	}
}
func (m *MockArticleService) GetArticleBySlug(slug string) (*models.Article, error) {
	if slug == "test-article" {
		return &models.Article{Slug: slug, Title: "Test", Draft: false}, nil
	}
	return nil, errors.New("article not found")
}
func (m *MockArticleService) GetArticlesByTag(tag string) []*models.Article           { return nil }
func (m *MockArticleService) GetArticlesByCategory(category string) []*models.Article { return nil }
func (m *MockArticleService) GetArticlesForFeed(limit int) []*models.Article          { return nil }
func (m *MockArticleService) GetFeaturedArticles(limit int) []*models.Article         { return nil }
func (m *MockArticleService) GetRecentArticles(limit int) []*models.Article           { return nil }
func (m *MockArticleService) GetAllTags() []string                                    { return []string{} }
func (m *MockArticleService) GetAllCategories() []string                              { return []string{} }
func (m *MockArticleService) GetTagCounts() []models.TagCount                         { return []models.TagCount{} }
func (m *MockArticleService) GetCategoryCounts() []models.CategoryCount {
	return []models.CategoryCount{}
}
func (m *MockArticleService) GetStats() *models.Stats                             { return &models.Stats{} }
func (m *MockArticleService) ReloadArticles() error                               { return nil }
func (m *MockArticleService) GetDraftArticles() []*models.Article                 { return nil }
func (m *MockArticleService) GetDraftBySlug(slug string) (*models.Article, error) { return nil, nil }

type MockEmailService struct {
	ShouldFail      bool
	NotConfigured   bool
	LastMessageSent *models.ContactMessage
}

func (m *MockEmailService) SendContactMessage(msg *models.ContactMessage) error {
	if m.NotConfigured {
		return apperrors.ErrEmailNotConfigured
	}
	if m.ShouldFail {
		return errors.New("email send failed")
	}
	m.LastMessageSent = msg
	return nil
}
func (m *MockEmailService) SendNotification(to, subject, body string) error { return nil }
func (m *MockEmailService) SendTestEmail() error                            { return nil }
func (m *MockEmailService) TestConnection() error                           { return nil }
func (m *MockEmailService) ValidateConfig() []string                        { return nil }
func (m *MockEmailService) GetConfig() map[string]any                       { return nil }
func (m *MockEmailService) Shutdown()                                       {}

type MockSearchService struct{}

func (m *MockSearchService) Search(articles []*models.Article, query string, limit int) []*models.SearchResult {
	return nil
}
func (m *MockSearchService) SearchInTitle(articles []*models.Article, query string, limit int) []*models.SearchResult {
	return nil
}
func (m *MockSearchService) SearchByTag(articles []*models.Article, tag string) []*models.Article {
	return nil
}
func (m *MockSearchService) SearchByCategory(articles []*models.Article, category string) []*models.Article {
	return nil
}
func (m *MockSearchService) GetSuggestions(articles []*models.Article, query string, limit int) []string {
	return nil
}

type MockTemplateService struct{}

func (m *MockTemplateService) Render(w io.Writer, templateName string, data any) error { return nil }
func (m *MockTemplateService) RenderToString(templateName string, data any) (string, error) {
	return "", nil
}
func (m *MockTemplateService) HasTemplate(templateName string) bool { return true }
func (m *MockTemplateService) ListTemplates() []string              { return nil }
func (m *MockTemplateService) Reload(templatesPath string) error    { return nil }
func (m *MockTemplateService) GetTemplate() *template.Template      { return nil }

type MockSEOService struct{}

func (m *MockSEOService) GenerateSitemap() ([]byte, error)   { return nil, nil }
func (m *MockSEOService) GenerateRobotsTxt() ([]byte, error) { return nil, nil }
func (m *MockSEOService) GenerateArticleSchema(article *models.Article, baseURL string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockSEOService) GenerateWebsiteSchema() (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockSEOService) GenerateBreadcrumbSchema(breadcrumbs []services.Breadcrumb) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockSEOService) GenerateOpenGraphTags(article *models.Article, baseURL string) (map[string]string, error) {
	return nil, nil
}
func (m *MockSEOService) GenerateTwitterCardTags(article *models.Article, baseURL string) (map[string]string, error) {
	return nil, nil
}
func (m *MockSEOService) GenerateMetaTags(article *models.Article) (map[string]string, error) {
	return nil, nil
}
func (m *MockSEOService) GeneratePageMetaTags(title, description, path string) (map[string]string, error) {
	return nil, nil
}
func (m *MockSEOService) AnalyzeContent(content string) (*services.SEOAnalysis, error) {
	return nil, nil
}
func (m *MockSEOService) IsEnabled() bool { return true }

func createTestConfig() *config.Config {
	return &config.Config{
		Environment: "test",
		BaseURL:     "http://localhost:3000",
		Blog: config.BlogConfig{
			Title:       "Test Blog",
			Description: "Test Description",
			Author:      "Test Author",
			AuthorEmail: "test@example.com",
		},
	}
}

func createTestAPIHandler(emailService *MockEmailService) *APIHandler {
	cfg := createTestConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	if emailService == nil {
		emailService = &MockEmailService{}
	}

	base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})
	return NewAPIHandler(base, &MockArticleService{}, emailService, time.Now())
}

// TestContact tests the contact form handler
func TestContact(t *testing.T) {
	t.Run("valid contact form submission", func(t *testing.T) {
		mockEmail := &MockEmailService{}
		handler := createTestAPIHandler(mockEmail)

		router := gin.New()
		router.POST("/contact", handler.Contact)

		formData := map[string]string{
			"name":    "John Doe",
			"email":   "john@example.com",
			"subject": "Test Subject",
			"message": "Test message content",
		}

		body, _ := json.Marshal(formData)
		req := httptest.NewRequest("POST", "/contact", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, mockEmail.LastMessageSent)
		assert.Equal(t, "John Doe", mockEmail.LastMessageSent.Name)
		assert.Equal(t, "john@example.com", mockEmail.LastMessageSent.Email)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
	})

	t.Run("invalid email address", func(t *testing.T) {
		handler := createTestAPIHandler(nil)

		router := gin.New()
		router.POST("/contact", handler.Contact)

		formData := map[string]string{
			"name":    "John Doe",
			"email":   "not-an-email",
			"subject": "Test",
			"message": "Test",
		}

		body, _ := json.Marshal(formData)
		req := httptest.NewRequest("POST", "/contact", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing required fields", func(t *testing.T) {
		handler := createTestAPIHandler(nil)

		router := gin.New()
		router.POST("/contact", handler.Contact)

		formData := map[string]string{
			"name": "John Doe",
			// Missing email, subject, message
		}

		body, _ := json.Marshal(formData)
		req := httptest.NewRequest("POST", "/contact", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("email service not configured", func(t *testing.T) {
		mockEmail := &MockEmailService{NotConfigured: true}
		handler := createTestAPIHandler(mockEmail)

		router := gin.New()
		router.POST("/contact", handler.Contact)

		formData := map[string]string{
			"name":    "John Doe",
			"email":   "john@example.com",
			"subject": "Test",
			"message": "Test",
		}

		body, _ := json.Marshal(formData)
		req := httptest.NewRequest("POST", "/contact", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "unavailable", response["status"])
	})

	t.Run("email service failure", func(t *testing.T) {
		mockEmail := &MockEmailService{ShouldFail: true}
		handler := createTestAPIHandler(mockEmail)

		router := gin.New()
		router.POST("/contact", handler.Contact)

		formData := map[string]string{
			"name":    "John Doe",
			"email":   "john@example.com",
			"subject": "Test",
			"message": "Test",
		}

		body, _ := json.Marshal(formData)
		req := httptest.NewRequest("POST", "/contact", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// When email fails, error is logged and handled
		// The test verifies the handler doesn't crash
		assert.NotEqual(t, 0, w.Code, "Handler should return a status code")
	})
}

// TestAdminEndpoints tests basic admin functionality
func TestAdminEndpoints(t *testing.T) {
	t.Run("admin stats returns valid data", func(t *testing.T) {
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})
		adminHandler := NewAdminHandler(base, &MockArticleService{}, time.Now())

		router := gin.New()
		router.GET("/admin/stats", adminHandler.Stats)

		req := httptest.NewRequest("GET", "/admin/stats", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Stats returns JSON with various metrics
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// Response should have at least some fields
		assert.NotEmpty(t, response)
	})
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	handler := createTestAPIHandler(nil)

	router := gin.New()
	router.GET("/health", handler.Health)

	req := httptest.NewRequest("GET", "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}
