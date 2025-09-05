package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/middleware"
	"github.com/vnykmshr/markgo/internal/models"
)

// mockProcessor implements ArticleProcessor for testing
type mockProcessor struct{}

func (m *mockProcessor) ProcessMarkdown(content string) (string, error) {
	return content, nil
}

func (m *mockProcessor) GenerateExcerpt(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "..."
}

func (m *mockProcessor) ProcessDuplicateTitles(title, htmlContent string) string {
	return htmlContent
}

// TestHandlerMocks contains all mocked services for testing
type TestHandlerMocks struct {
	ArticleService *MockArticleService
	EmailService   *MockEmailService
	CacheService   *MockCacheService
	SearchService  *MockSearchService
}

// TestConfig holds test configuration
type TestConfig struct {
	Handlers *Handlers
	Mocks    *TestHandlerMocks
	Router   *gin.Engine
}

// SetupTestEnvironment creates a complete test environment with handlers and mocks
func SetupTestEnvironment(t *testing.T) (*TestConfig, func()) {
	gin.SetMode(gin.TestMode)

	// Create mock services
	mocks := &TestHandlerMocks{
		ArticleService: &MockArticleService{},
		EmailService:   &MockEmailService{},
		CacheService:   &MockCacheService{},
		SearchService:  &MockSearchService{},
	}

	// Create test config
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

	// Create logger that doesn't output during tests
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	// Create handlers
	handlers := New(&Config{
		ArticleService: mocks.ArticleService,
		CacheService:   mocks.CacheService,
		EmailService:   mocks.EmailService,
		SearchService:  mocks.SearchService,
		Config:         cfg,
		Logger:         logger,
	})

	// Create router with minimal templates and error handling
	router := gin.New()
	setupMinimalTemplates(router)
	
	// Add error handling middleware for proper error status codes
	router.Use(middleware.ErrorHandler(logger))

	testConfig := &TestConfig{
		Handlers: handlers,
		Mocks:    mocks,
		Router:   router,
	}

	// Cleanup function
	cleanup := func() {
		// Reset all mocks
		mocks.ArticleService.AssertExpectations(t)
		mocks.EmailService.AssertExpectations(t)
		mocks.CacheService.AssertExpectations(t)
		mocks.SearchService.AssertExpectations(t)
	}

	return testConfig, cleanup
}

// SetupDefaultMocks sets up common mock expectations
func SetupDefaultMocks(mocks *TestHandlerMocks) {
	// Common expectations that many tests use
	mocks.CacheService.On("Get", mock.Anything).Return(nil, false).Maybe()
	mocks.CacheService.On("Set", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mocks.ArticleService.On("GetRecentArticles", mock.Anything).Return([]*models.Article{}).Maybe()
}

// CreateTestArticlesWithVariations creates test articles with different characteristics
func CreateTestArticlesWithVariations() []*models.Article {
	now := time.Now()
	articles := []*models.Article{
		{
			Slug:        "published-article",
			Title:       "Published Article",
			Description: "A published article",
			Date:        now,
			Tags:        []string{"test", "golang"},
			Categories:  []string{"programming"},
			Featured:    true,
			Draft:       false,
			Content:     "Published content",
			WordCount:   100,
			ReadingTime: 1,
		},
		{
			Slug:        "draft-article",
			Title:       "Draft Article",
			Description: "A draft article",
			Date:        now,
			Tags:        []string{"test", "web"},
			Categories:  []string{"development"},
			Featured:    false,
			Draft:       true,
			Content:     "Draft content",
			WordCount:   50,
			ReadingTime: 1,
		},
		{
			Slug:        "featured-article",
			Title:       "Featured Article",
			Description: "A featured article",
			Date:        now.AddDate(0, 0, -1),
			Tags:        []string{"featured", "golang"},
			Categories:  []string{"programming"},
			Featured:    true,
			Draft:       false,
			Content:     "Featured content with more text to make it longer",
			WordCount:   200,
			ReadingTime: 2,
		},
	}

	// Set processor for all articles
	processor := &mockProcessor{}
	for _, article := range articles {
		article.SetProcessor(processor)
	}

	return articles
}

// AssertHTMLResponse asserts common HTML response properties
func AssertHTMLResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, contains ...string) {
	assert.Equal(t, expectedStatus, recorder.Code)

	if expectedStatus == http.StatusOK {
		assert.Contains(t, recorder.Header().Get("Content-Type"), "text/html")

		body := recorder.Body.String()
		for _, containsText := range contains {
			assert.Contains(t, body, containsText)
		}
	}
}

// AssertJSONResponse asserts common JSON response properties
func AssertJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	assert.Equal(t, expectedStatus, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	return response
}

// CreateContactFormData creates valid contact form data for testing
func CreateContactFormData() map[string]interface{} {
	return map[string]interface{}{
		"name":             "John Doe",
		"email":            "john@example.com",
		"subject":          "Test Message",
		"message":          "This is a test message with sufficient length to pass validation",
		"captcha_question": "3 + 5",
		"captcha_answer":   "8",
	}
}

// CreateInvalidContactFormData creates invalid contact form data for testing
func CreateInvalidContactFormData() map[string]interface{} {
	return map[string]interface{}{
		"name":             "",              // Invalid: empty name
		"email":            "invalid-email", // Invalid: bad email format
		"subject":          "Test Message",
		"message":          "Short", // Invalid: too short
		"captcha_question": "3 + 5",
		"captcha_answer":   "7", // Invalid: wrong answer
	}
}

// MakeJSONRequest creates a JSON request for testing
func MakeJSONRequest(method, url string, data interface{}) (*http.Request, error) {
	var body *bytes.Buffer
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// ExecuteRequest executes a request and returns the response recorder
func ExecuteRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

// SetupCacheMocks sets up cache mock expectations for hit/miss scenarios
func SetupCacheMocks(mockCache *MockCacheService, key string, hit bool, data interface{}) {
	if hit {
		mockCache.On("Get", key).Return(data, true)
	} else {
		mockCache.On("Get", key).Return(nil, false)
		mockCache.On("Set", key, mock.Anything, mock.Anything).Return()
	}
}

// SetupArticleServiceMocks sets up common article service mock expectations
func SetupArticleServiceMocks(mockArticle *MockArticleService, articles []*models.Article) {
	mockArticle.On("GetAllArticles").Return(articles).Maybe()
	mockArticle.On("GetFeaturedArticles", mock.Anything).Return(articles[:1]).Maybe()
	mockArticle.On("GetRecentArticles", mock.Anything).Return(articles).Maybe()
	mockArticle.On("GetTagCounts").Return([]models.TagCount{
		{Tag: "golang", Count: 2},
		{Tag: "test", Count: 3},
	}).Maybe()
	mockArticle.On("GetCategoryCounts").Return([]models.CategoryCount{
		{Category: "programming", Count: 2},
		{Category: "development", Count: 1},
	}).Maybe()
}

// TableDrivenTestCase represents a test case for table-driven tests
type TableDrivenTestCase struct {
	Name           string
	SetupMocks     func(*TestHandlerMocks)
	RequestFunc    func() *http.Request
	ExpectedStatus int
	ExpectedBody   []string
	ValidationFunc func(*testing.T, *httptest.ResponseRecorder)
}

// RunTableDrivenTest runs a table-driven test with the given test cases
func RunTableDrivenTest(t *testing.T, testCases []TableDrivenTestCase, routeSetup func(*gin.Engine, *Handlers)) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			testConfig, cleanup := SetupTestEnvironment(t)
			defer cleanup()

			// Setup route
			routeSetup(testConfig.Router, testConfig.Handlers)

			// Setup mocks
			if tc.SetupMocks != nil {
				tc.SetupMocks(testConfig.Mocks)
			}

			// Make request
			req := tc.RequestFunc()
			recorder := ExecuteRequest(testConfig.Router, req)

			// Assert status
			assert.Equal(t, tc.ExpectedStatus, recorder.Code)

			// Assert body contains expected text
			body := recorder.Body.String()
			for _, expected := range tc.ExpectedBody {
				assert.Contains(t, body, expected)
			}

			// Custom validation
			if tc.ValidationFunc != nil {
				tc.ValidationFunc(t, recorder)
			}
		})
	}
}
