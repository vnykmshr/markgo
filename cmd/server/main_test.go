package main

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/handlers"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// Helper function to create minimal templates for testing
func setupMinimalTemplates(router *gin.Engine) {
	// Create a simple template that can handle basic rendering
	tmpl := `<!DOCTYPE html><html><head><title>{{.title}}</title></head><body>{{.message}}</body></html>`
	router.SetHTMLTemplate(template.Must(template.New("test").Parse(tmpl)))
}

func TestMain(m *testing.M) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Run tests
	code := m.Run()

	// Exit with test result code
	os.Exit(code)
}

// Mock implementations using the new interfaces

type MockArticleService struct {
	mock.Mock
}

func (m *MockArticleService) GetAllArticles() []*models.Article {
	args := m.Called()
	return args.Get(0).([]*models.Article)
}

func (m *MockArticleService) GetArticleBySlug(slug string) (*models.Article, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Article), args.Error(1)
}

func (m *MockArticleService) GetArticlesByTag(tag string) []*models.Article {
	args := m.Called(tag)
	return args.Get(0).([]*models.Article)
}

func (m *MockArticleService) GetArticlesByCategory(category string) []*models.Article {
	args := m.Called(category)
	return args.Get(0).([]*models.Article)
}

func (m *MockArticleService) GetArticlesForFeed(limit int) []*models.Article {
	args := m.Called(limit)
	return args.Get(0).([]*models.Article)
}

func (m *MockArticleService) GetFeaturedArticles(limit int) []*models.Article {
	args := m.Called(limit)
	return args.Get(0).([]*models.Article)
}

func (m *MockArticleService) GetRecentArticles(limit int) []*models.Article {
	args := m.Called(limit)
	return args.Get(0).([]*models.Article)
}

func (m *MockArticleService) GetAllTags() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockArticleService) GetAllCategories() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockArticleService) GetTagCounts() []models.TagCount {
	args := m.Called()
	return args.Get(0).([]models.TagCount)
}

func (m *MockArticleService) GetCategoryCounts() []models.CategoryCount {
	args := m.Called()
	return args.Get(0).([]models.CategoryCount)
}

func (m *MockArticleService) GetStats() *models.Stats {
	args := m.Called()
	return args.Get(0).(*models.Stats)
}

func (m *MockArticleService) ReloadArticles() error {
	args := m.Called()
	return args.Error(0)
}

// Draft operations mock methods
func (m *MockArticleService) GetDraftArticles() []*models.Article {
	args := m.Called()
	return args.Get(0).([]*models.Article)
}

func (m *MockArticleService) GetDraftBySlug(slug string) (*models.Article, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Article), args.Error(1)
}

func (m *MockArticleService) PreviewDraft(slug string) (*models.Article, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Article), args.Error(1)
}

func (m *MockArticleService) PublishDraft(slug string) error {
	args := m.Called(slug)
	return args.Error(0)
}

func (m *MockArticleService) UnpublishArticle(slug string) error {
	args := m.Called(slug)
	return args.Error(0)
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendContactMessage(msg *models.ContactMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockEmailService) SendNotification(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func (m *MockEmailService) SendTestEmail() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEmailService) TestConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEmailService) ValidateConfig() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockEmailService) GetConfig() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func (m *MockEmailService) Shutdown() {
	m.Called()
}

func TestSetupRoutes(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		StaticPath:  "./test/static",
		Environment: "test",
		Cache:       config.CacheConfig{TTL: time.Hour, MaxSize: 100},
		Email:       config.EmailConfig{Host: "test", Port: 587},
		RateLimit: config.RateLimitConfig{
			General: config.RateLimit{Requests: 100, Window: time.Minute},
			Contact: config.RateLimit{Requests: 5, Window: time.Hour},
		},
		Admin: config.AdminConfig{Username: "admin", Password: "test"},
	}

	// Create mock services
	articleService := &MockArticleService{}
	emailService := &MockEmailService{}
	searchService := services.NewSearchService()

	// Create cache
	cacheConfig := obcache.NewDefaultConfig()
	cacheConfig.MaxEntries = 100
	cacheConfig.DefaultTTL = time.Hour
	testCache, _ := obcache.New(cacheConfig)

	// Create handlers
	h := handlers.New(&handlers.Config{
		ArticleService: articleService,
		EmailService:   emailService,
		SearchService:  searchService,
		Config:         cfg,
		Logger:         slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		Cache:          testCache,
	})

	// Create router
	router := gin.New()
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	setupRoutes(router, h, cfg, logger)

	// Test main routes
	testCases := []struct {
		method      string
		path        string
		expected    int
		shouldExist bool
	}{
		{"GET", "/health", http.StatusOK, true},
		{"GET", "/metrics", http.StatusOK, true},
		{"GET", "/admin/stats", http.StatusUnauthorized, true}, // No auth provided
	}

	for _, tc := range testCases {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			// Set up minimal mock expectations for endpoints that need them
			if tc.path == "/metrics" {
				articleService.On("GetStats").Return(&models.Stats{}).Maybe()
				_ = testCache.Set("test", "value", time.Hour) // Ensure cache has some stats
			}

			req, err := http.NewRequest(tc.method, tc.path, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if tc.shouldExist {
				assert.Equal(t, tc.expected, recorder.Code)
			}
		})
	}
}

func TestSetupRoutesWithoutAdmin(t *testing.T) {
	// Create test configuration without admin credentials
	cfg := &config.Config{
		StaticPath:  "./test/static",
		Environment: "test",
		Cache:       config.CacheConfig{TTL: time.Hour, MaxSize: 100},
		Email:       config.EmailConfig{Host: "test", Port: 587},
		RateLimit: config.RateLimitConfig{
			General: config.RateLimit{Requests: 100, Window: time.Minute},
			Contact: config.RateLimit{Requests: 5, Window: time.Hour},
		},
		Admin: config.AdminConfig{Username: "", Password: ""}, // No admin credentials
	}

	// Create mock services
	articleService := &MockArticleService{}
	emailService := &MockEmailService{}
	searchService := services.NewSearchService()

	// Create cache
	cacheConfig := obcache.NewDefaultConfig()
	cacheConfig.MaxEntries = 100
	cacheConfig.DefaultTTL = time.Hour
	testCache, _ := obcache.New(cacheConfig)

	// Create handlers
	h := handlers.New(&handlers.Config{
		ArticleService: articleService,
		EmailService:   emailService,
		SearchService:  searchService,
		Config:         cfg,
		Logger:         slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		Cache:          testCache,
	})

	// Create router
	router := gin.New()
	setupMinimalTemplates(router) // Add minimal templates for testing
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	setupRoutes(router, h, cfg, logger)

	// Test that admin routes don't exist by checking routes directly
	// Since the NotFound handler requires templates, we'll test by examining the router's routes
	routes := router.Routes()

	// Check that no admin routes are registered
	adminRoutesFound := false
	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/admin") {
			adminRoutesFound = true
			break
		}
	}

	// Should not find any admin routes since credentials are empty
	assert.False(t, adminRoutesFound, "Admin routes should not be registered when credentials are empty")
}

func TestSetupTemplates(t *testing.T) {
	// Create temporary template directory
	tmpDir, err := os.MkdirTemp("", "test_templates")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create all required template files
	requiredTemplates := map[string]string{
		"base.html":       `<!DOCTYPE html><html><head><title>{{.Title}}</title></head><body>{{.Content}}</body></html>`,
		"index.html":      `{{template "base.html" .}}`,
		"article.html":    `{{template "base.html" .}}`,
		"articles.html":   `{{template "base.html" .}}`,
		"404.html":        `{{template "base.html" .}}`,
		"contact.html":    `{{template "base.html" .}}`,
		"search.html":     `{{template "base.html" .}}`,
		"tags.html":       `{{template "base.html" .}}`,
		"categories.html": `{{template "base.html" .}}`,
	}

	for filename, content := range requiredTemplates {
		templatePath := filepath.Join(tmpDir, filename)
		err = os.WriteFile(templatePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test setupTemplates - create TemplateService first
	cfg := &config.Config{}
	templateService, err := services.NewTemplateService(tmpDir, cfg)
	require.NoError(t, err)

	router := gin.New()
	err = setupTemplates(router, templateService)
	assert.NoError(t, err)
}

func TestSetupTemplatesWithMissingDirectory(t *testing.T) {
	cfg := &config.Config{}

	// This should return an error when template service creation fails
	_, err := services.NewTemplateService("/nonexistent/path", cfg)
	assert.Error(t, err)
}

func TestSetupTemplatesWithInvalidTemplate(t *testing.T) {
	// Create temporary template directory
	tmpDir, err := os.MkdirTemp("", "test_templates")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create invalid template file
	invalidTemplateContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title</title></head>
<body><h1>{{.Title}}</h1></body>
</html>` // Missing closing brace

	templatePath := filepath.Join(tmpDir, "invalid.html")
	err = os.WriteFile(templatePath, []byte(invalidTemplateContent), 0644)
	require.NoError(t, err)

	// This should return an error when creating template service with invalid template
	cfg := &config.Config{}
	_, err = services.NewTemplateService(tmpDir, cfg)
	assert.Error(t, err)
}

func TestRouteStaticFiles(t *testing.T) {
	// Create temporary static directory
	tmpDir, err := os.MkdirTemp("", "test_static")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test static file
	staticContent := "/* CSS content */"
	staticPath := filepath.Join(tmpDir, "style.css")
	err = os.WriteFile(staticPath, []byte(staticContent), 0644)
	require.NoError(t, err)

	// Create favicon file
	faviconDir := filepath.Join(tmpDir, "img")
	err = os.MkdirAll(faviconDir, 0755)
	require.NoError(t, err)

	faviconPath := filepath.Join(faviconDir, "favicon.ico")
	err = os.WriteFile(faviconPath, []byte("fake favicon"), 0644)
	require.NoError(t, err)

	// Create robots.txt
	robotsPath := filepath.Join(tmpDir, "robots.txt")
	err = os.WriteFile(robotsPath, []byte("User-agent: *\nDisallow:"), 0644)
	require.NoError(t, err)

	// Create configuration
	cfg := &config.Config{
		StaticPath:  tmpDir,
		Environment: "test",
		Cache:       config.CacheConfig{TTL: time.Hour, MaxSize: 100},
		Email:       config.EmailConfig{Host: "test", Port: 587},
		RateLimit: config.RateLimitConfig{
			General: config.RateLimit{Requests: 100, Window: time.Minute},
			Contact: config.RateLimit{Requests: 5, Window: time.Hour},
		},
	}

	// Create mock services
	articleService := &MockArticleService{}
	emailService := &MockEmailService{}
	searchService := services.NewSearchService()

	// Create cache
	cacheConfig := obcache.NewDefaultConfig()
	cacheConfig.MaxEntries = 100
	cacheConfig.DefaultTTL = time.Hour
	testCache, _ := obcache.New(cacheConfig)

	// Create handlers
	h := handlers.New(&handlers.Config{
		ArticleService: articleService,
		EmailService:   emailService,
		SearchService:  searchService,
		Config:         cfg,
		Logger:         slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		Cache:          testCache,
	})

	// Create router
	router := gin.New()
	setupMinimalTemplates(router) // Add minimal templates for testing
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	setupRoutes(router, h, cfg, logger)

	// Test static file serving
	testCases := []struct {
		path     string
		expected int
	}{
		{"/static/style.css", http.StatusOK},
		{"/favicon.ico", http.StatusOK},
		{"/robots.txt", http.StatusOK},
		{"/static/nonexistent.css", http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected, recorder.Code)
		})
	}
}

// Benchmark tests
func BenchmarkSetupRoutes(b *testing.B) {
	cfg := &config.Config{
		StaticPath:  "./test/static",
		Environment: "test",
		Cache:       config.CacheConfig{TTL: time.Hour, MaxSize: 100},
		Email:       config.EmailConfig{Host: "test", Port: 587},
		RateLimit: config.RateLimitConfig{
			General: config.RateLimit{Requests: 100, Window: time.Minute},
			Contact: config.RateLimit{Requests: 5, Window: time.Hour},
		},
	}

	articleService := &MockArticleService{}
	emailService := &MockEmailService{}
	searchService := services.NewSearchService()

	// Create cache
	cacheConfig2 := obcache.NewDefaultConfig()
	cacheConfig2.MaxEntries = 100
	cacheConfig2.DefaultTTL = time.Hour
	testCache2, _ := obcache.New(cacheConfig2)

	h := handlers.New(&handlers.Config{
		ArticleService: articleService,
		EmailService:   emailService,
		SearchService:  searchService,
		Config:         cfg,
		Logger:         slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		Cache:          testCache2,
	})

	for b.Loop() {
		router := gin.New()
		logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
		setupRoutes(router, h, cfg, logger)
	}
}

func BenchmarkSetupTemplates(b *testing.B) {
	// Create temporary template directory
	tmpDir, err := os.MkdirTemp("", "bench_templates")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	// Create test template file
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body><h1>{{.Title}}</h1></body>
</html>`

	templatePath := filepath.Join(tmpDir, "test.html")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(b, err)

	cfg := &config.Config{}
	templateService, err := services.NewTemplateService(tmpDir, cfg)
	require.NoError(b, err)

	for b.Loop() {
		router := gin.New()
		_ = setupTemplates(router, templateService)
	}
}
