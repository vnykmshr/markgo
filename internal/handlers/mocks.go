package handlers

import (
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/vnykmshr/markgo/internal/models"
)

// setupMinimalTemplates creates minimal templates for testing
func setupMinimalTemplates(router *gin.Engine) {
	// Create simple templates that can handle basic rendering
	tmpl := template.New("base.html")
	tmpl, _ = tmpl.Parse(`<!DOCTYPE html><html><head><title>{{.title}}</title></head><body>{{.message}}</body></html>`)
	router.SetHTMLTemplate(tmpl)
}

// MockArticleService is a mock implementation of ArticleServiceInterface
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

// MockEmailService is a mock implementation of EmailServiceInterface
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

// MockCacheService is a mock implementation of CacheServiceInterface
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Set(key string, value any, ttl time.Duration) {
	m.Called(key, value, ttl)
}

func (m *MockCacheService) Get(key string) (any, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

func (m *MockCacheService) Delete(key string) {
	m.Called(key)
}

func (m *MockCacheService) Clear() {
	m.Called()
}

func (m *MockCacheService) Size() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCacheService) Keys() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockCacheService) Exists(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockCacheService) GetTTL(key string) time.Duration {
	args := m.Called(key)
	return args.Get(0).(time.Duration)
}

func (m *MockCacheService) Stats() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func (m *MockCacheService) GetOrSet(key string, generator func() any, ttl time.Duration) any {
	args := m.Called(key, generator, ttl)
	return args.Get(0)
}

func (m *MockCacheService) Stop() {
	m.Called()
}

// MockSearchService is a mock implementation of SearchServiceInterface
type MockSearchService struct {
	mock.Mock
}

func (m *MockSearchService) Search(articles []*models.Article, query string, limit int) []*models.SearchResult {
	args := m.Called(articles, query, limit)
	return args.Get(0).([]*models.SearchResult)
}

func (m *MockSearchService) SearchInTitle(articles []*models.Article, query string, limit int) []*models.SearchResult {
	args := m.Called(articles, query, limit)
	return args.Get(0).([]*models.SearchResult)
}

func (m *MockSearchService) SearchByTag(articles []*models.Article, tag string) []*models.Article {
	args := m.Called(articles, tag)
	return args.Get(0).([]*models.Article)
}

func (m *MockSearchService) SearchByCategory(articles []*models.Article, category string) []*models.Article {
	args := m.Called(articles, category)
	return args.Get(0).([]*models.Article)
}

func (m *MockSearchService) GetSuggestions(articles []*models.Article, query string, limit int) []string {
	args := m.Called(articles, query, limit)
	return args.Get(0).([]string)
}
