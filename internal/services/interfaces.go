package services

import (
	"time"

	"github.com/yourusername/markgo/internal/models"
)

// ArticleServiceInterface defines the interface for article operations
type ArticleServiceInterface interface {
	// Article retrieval methods
	GetAllArticles() []*models.Article
	GetArticleBySlug(slug string) (*models.Article, error)
	GetArticlesByTag(tag string) []*models.Article
	GetArticlesByCategory(category string) []*models.Article
	GetArticlesForFeed(limit int) []*models.Article

	// Featured and recent articles
	GetFeaturedArticles(limit int) []*models.Article
	GetRecentArticles(limit int) []*models.Article

	// Metadata methods
	GetAllTags() []string
	GetAllCategories() []string
	GetTagCounts() []models.TagCount
	GetCategoryCounts() []models.CategoryCount

	// Statistics and management
	GetStats() *models.Stats
	ReloadArticles() error
}

// EmailServiceInterface defines the interface for email operations
type EmailServiceInterface interface {
	// Email sending methods
	SendContactMessage(msg *models.ContactMessage) error
	SendNotification(to, subject, body string) error
	SendTestEmail() error

	// Configuration and validation
	TestConnection() error
	ValidateConfig() []string
	GetConfig() map[string]any

	// Lifecycle management
	Shutdown()
}

// CacheServiceInterface defines the interface for cache operations
type CacheServiceInterface interface {
	// Basic cache operations
	Set(key string, value any, ttl time.Duration)
	Get(key string) (any, bool)
	Delete(key string)
	Clear()

	// Cache management
	Size() int
	Keys() []string
	Exists(key string) bool
	GetTTL(key string) time.Duration
	Stats() map[string]any

	// Advanced operations
	GetOrSet(key string, generator func() any, ttl time.Duration) any
	Stop()
}

// SearchServiceInterface defines the interface for search operations
type SearchServiceInterface interface {
	// Search methods
	Search(articles []*models.Article, query string, limit int) []*models.SearchResult
	SearchInTitle(articles []*models.Article, query string, limit int) []*models.SearchResult
	SearchByTag(articles []*models.Article, tag string) []*models.Article
	SearchByCategory(articles []*models.Article, category string) []*models.Article

	// Suggestions
	GetSuggestions(articles []*models.Article, query string, limit int) []string
}
