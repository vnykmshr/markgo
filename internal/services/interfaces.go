package services

import (
	"context"
	"html/template"
	"io"
	"log/slog"
	"time"

	"github.com/vnykmshr/markgo/internal/models"
)

// LogEntry represents a structured log entry with common fields
type LogEntry struct {
	RequestID  string        `json:"request_id,omitempty"`
	UserID     string        `json:"user_id,omitempty"`
	IP         string        `json:"ip,omitempty"`
	UserAgent  string        `json:"user_agent,omitempty"`
	Path       string        `json:"path,omitempty"`
	Method     string        `json:"method,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
	StatusCode int           `json:"status_code,omitempty"`
	Component  string        `json:"component,omitempty"`
	Action     string        `json:"action,omitempty"`
}

// PerformanceLog represents performance-specific logging data
type PerformanceLog struct {
	Operation    string        `json:"operation"`
	Duration     time.Duration `json:"duration"`
	MemoryBefore int64         `json:"memory_before"`
	MemoryAfter  int64         `json:"memory_after"`
	Goroutines   int           `json:"goroutines"`
	Allocations  uint64        `json:"allocations"`
}

// SecurityLog represents security-related logging data
type SecurityLog struct {
	Event       string `json:"event"`
	Severity    string `json:"severity"`
	IP          string `json:"ip,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
}

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

	// Draft operations
	GetDraftArticles() []*models.Article
	GetDraftBySlug(slug string) (*models.Article, error)
	PreviewDraft(slug string) (*models.Article, error)
	PublishDraft(slug string) error
	UnpublishArticle(slug string) error
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

// TemplateServiceInterface defines the interface for template operations
type TemplateServiceInterface interface {
	// Template rendering
	Render(w io.Writer, templateName string, data any) error
	RenderToString(templateName string, data any) (string, error)

	// Template management
	HasTemplate(templateName string) bool
	ListTemplates() []string
	Reload(templatesPath string) error

	// Internal access (for Gin integration)
	GetTemplate() *template.Template
}

// LoggingServiceInterface defines the interface for logging operations
type LoggingServiceInterface interface {
	// Core logger access
	GetLogger() *slog.Logger

	// Contextual logging
	WithContext(keyvals ...interface{}) *slog.Logger
	WithRequestContext(ctx context.Context, entry LogEntry) *slog.Logger
	WithComponent(component string) *slog.Logger

	// Basic logging methods
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	// Enhanced logging methods
	LogPerformance(perfLog PerformanceLog)
	LogSecurity(secLog SecurityLog)
	LogHTTPRequest(ctx context.Context, entry LogEntry)
	LogError(ctx context.Context, err error, msg string, keyvals ...interface{})
	LogSlowOperation(ctx context.Context, operation string, duration time.Duration, threshold time.Duration, keyvals ...interface{})

	// Utility methods
	GetMemoryStats() (int64, int64, uint64)

	// Lifecycle management
	Close() error
}
