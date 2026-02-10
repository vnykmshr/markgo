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

	// Search
	SearchArticles(query string, limit int) []*models.SearchResult

	// Statistics and management
	GetStats() *models.Stats
	ReloadArticles() error

	// Draft read-only operations
	// Note: Drafts are articles with `draft: true` in YAML frontmatter.
	// Users manage draft status via Git (edit file, commit, push).
	// No publish/unpublish API needed - use your editor and Git workflow.
	GetDraftArticles() []*models.Article
	GetDraftBySlug(slug string) (*models.Article, error)
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
	WithRequestContext(ctx context.Context, entry *LogEntry) *slog.Logger
	WithComponent(component string) *slog.Logger

	// Basic logging methods
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	// Enhanced logging methods
	LogPerformance(perfLog PerformanceLog)
	LogSecurity(secLog *SecurityLog)
	LogHTTPRequest(ctx context.Context, entry *LogEntry)
	LogError(ctx context.Context, err error, msg string, keyvals ...interface{})
	LogSlowOperation(
		ctx context.Context,
		operation string,
		duration, threshold time.Duration,
		keyvals ...interface{},
	)

	// Utility methods
	GetMemoryStats() (alloc, sys int64, mallocs uint64)

	// Lifecycle management
	Close() error
}

// FeedServiceInterface defines the interface for feed generation
type FeedServiceInterface interface {
	GenerateRSS() (string, error)
	GenerateJSONFeed() (string, error)
	GenerateSitemap() (string, error)
}

// SEOServiceInterface defines the interface for SEO utilities (stateless)
type SEOServiceInterface interface {
	// Sitemap generation (on-demand, no caching)
	GenerateSitemap() ([]byte, error)
	GenerateRobotsTxt() ([]byte, error)

	// Schema.org structured data
	GenerateArticleSchema(article *models.Article, baseURL string) (map[string]interface{}, error)
	GenerateWebsiteSchema() (map[string]interface{}, error)
	GenerateBreadcrumbSchema(breadcrumbs []Breadcrumb) (map[string]interface{}, error)

	// Open Graph and meta tag optimization
	GenerateOpenGraphTags(article *models.Article, baseURL string) (map[string]string, error)
	GenerateTwitterCardTags(article *models.Article, baseURL string) (map[string]string, error)
	GenerateMetaTags(article *models.Article) (map[string]string, error)
	GeneratePageMetaTags(title, description, path string) (map[string]string, error)

	// SEO analysis
	AnalyzeContent(content string) (*SEOAnalysis, error)

	// Status check
	IsEnabled() bool
}

// SiteConfig represents site-wide configuration for SEO
type SiteConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	BaseURL     string `json:"base_url"`
	Language    string `json:"language"`
	Author      string `json:"author"`
	Logo        string `json:"logo,omitempty"`
	Image       string `json:"image,omitempty"`
}

// Breadcrumb represents a breadcrumb item for structured data
type Breadcrumb struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// SEOAnalysis represents content SEO analysis results
type SEOAnalysis struct {
	TitleLength  int      `json:"title_length"`
	DescLength   int      `json:"description_length"`
	WordCount    int      `json:"word_count"`
	ReadingTime  int      `json:"reading_time"`
	HeadingCount int      `json:"heading_count"`
	ImageCount   int      `json:"image_count"`
	LinkCount    int      `json:"link_count"`
	Keywords     []string `json:"keywords"`
	Suggestions  []string `json:"suggestions"`
	Score        float64  `json:"score"`
}

// RobotsConfig represents robots.txt configuration
type RobotsConfig struct {
	UserAgent  string   `json:"user_agent"`
	Allow      []string `json:"allow"`
	Disallow   []string `json:"disallow"`
	CrawlDelay int      `json:"crawl_delay,omitempty"`
	SitemapURL string   `json:"sitemap_url"`
}
