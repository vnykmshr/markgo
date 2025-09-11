package services // import "github.com/vnykmshr/markgo/internal/services"


FUNCTIONS

func GetTemplateFuncMap() template.FuncMap
    GetFuncMap returns the template function map for reuse in other services

func GetTimezoneCacheStats() map[string]any
    GetTimezoneCacheStats returns timezone cache statistics


TYPES

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
    ArticleServiceInterface defines the interface for article operations

func NewArticleService(articlesPath string, logger *slog.Logger) (ArticleServiceInterface, error)
    NewArticleService creates a new modular article service Built with
    enterprise-grade performance, caching, and modularity

type CachedTemplateFunctions struct {
	RenderToString func(string, any) (string, error)
	ParseTemplate  func(string, string) (*template.Template, error)
}
    CachedTemplateFunctions holds obcache-wrapped template operations

type EmailService struct {
	// Has unexported fields.
}

func NewEmailService(cfg config.EmailConfig, logger *slog.Logger) *EmailService

func (e *EmailService) GetConfig() map[string]any
    GetConfig returns the current email configuration (without sensitive data)

func (e *EmailService) SendContactMessage(msg *models.ContactMessage) error
    SendContactMessage sends a contact form message via email

func (e *EmailService) SendNotification(to, subject, body string) error
    SendNotification sends a general notification email

func (e *EmailService) SendTestEmail() error
    SendTestEmail sends a test email to verify configuration

func (e *EmailService) Shutdown()
    Shutdown gracefully shuts down the email service

func (e *EmailService) TestConnection() error
    TestConnection tests the email configuration

func (e *EmailService) ValidateConfig() []string
    ValidateConfig validates the email configuration

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
    EmailServiceInterface defines the interface for email operations

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
    LogEntry represents a structured log entry with common fields

type LoggingService struct {
	// Has unexported fields.
}
    LoggingService provides enhanced logging functionality with rotation and
    formatting

func NewLoggingService(cfg config.LoggingConfig) (*LoggingService, error)
    NewLoggingService creates a new logging service with the given configuration

func (ls *LoggingService) Close() error
    Close closes any resources used by the logging service

func (ls *LoggingService) Debug(msg string, keyvals ...interface{})
    Debug logs a debug message

func (ls *LoggingService) Error(msg string, keyvals ...interface{})
    Error logs an error message

func (ls *LoggingService) GetLogger() *slog.Logger
    GetLogger returns the configured slog.Logger instance

func (ls *LoggingService) GetMemoryStats() (int64, int64, uint64)
    GetMemoryStats returns current memory statistics for logging

func (ls *LoggingService) Info(msg string, keyvals ...interface{})
    Info logs an info message

func (ls *LoggingService) LogError(ctx context.Context, err error, msg string, keyvals ...interface{})
    LogError logs errors with enhanced context and stack traces

func (ls *LoggingService) LogHTTPRequest(ctx context.Context, entry LogEntry)
    LogHTTPRequest logs HTTP request details with structured data

func (ls *LoggingService) LogPerformance(perfLog PerformanceLog)
    LogPerformance logs performance metrics with structured data

func (ls *LoggingService) LogSecurity(secLog SecurityLog)
    LogSecurity logs security-related events with structured data

func (ls *LoggingService) LogSlowOperation(ctx context.Context, operation string, duration time.Duration, threshold time.Duration, keyvals ...interface{})
    LogSlowOperation logs operations that exceed expected duration

func (ls *LoggingService) Warn(msg string, keyvals ...interface{})
    Warn logs a warning message

func (ls *LoggingService) WithComponent(component string) *slog.Logger
    WithComponent creates a logger with component-specific context

func (ls *LoggingService) WithContext(keyvals ...interface{}) *slog.Logger
    WithContext creates a logger with additional context fields

func (ls *LoggingService) WithRequestContext(ctx context.Context, entry LogEntry) *slog.Logger
    WithRequestContext creates a logger with request-specific context

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
    LoggingServiceInterface defines the interface for logging operations

type PerformanceLog struct {
	Operation    string        `json:"operation"`
	Duration     time.Duration `json:"duration"`
	MemoryBefore int64         `json:"memory_before"`
	MemoryAfter  int64         `json:"memory_after"`
	Goroutines   int           `json:"goroutines"`
	Allocations  uint64        `json:"allocations"`
}
    PerformanceLog represents performance-specific logging data

type SearchService struct{}

func NewSearchService() *SearchService

func (s *SearchService) GetSuggestions(articles []*models.Article, query string, limit int) []string

func (s *SearchService) Search(articles []*models.Article, query string, limit int) []*models.SearchResult

func (s *SearchService) SearchByCategory(articles []*models.Article, category string) []*models.Article

func (s *SearchService) SearchByTag(articles []*models.Article, tag string) []*models.Article

func (s *SearchService) SearchInTitle(articles []*models.Article, query string, limit int) []*models.SearchResult

type SearchServiceInterface interface {
	// Search methods
	Search(articles []*models.Article, query string, limit int) []*models.SearchResult
	SearchInTitle(articles []*models.Article, query string, limit int) []*models.SearchResult
	SearchByTag(articles []*models.Article, tag string) []*models.Article
	SearchByCategory(articles []*models.Article, category string) []*models.Article

	// Suggestions
	GetSuggestions(articles []*models.Article, query string, limit int) []string
}
    SearchServiceInterface defines the interface for search operations

type SecurityLog struct {
	Event       string `json:"event"`
	Severity    string `json:"severity"`
	IP          string `json:"ip,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
}
    SecurityLog represents security-related logging data

type TemplateService struct {
	// Has unexported fields.
}

func NewTemplateService(templatesPath string, cfg *config.Config) (*TemplateService, error)

func (t *TemplateService) GetCacheStats() map[string]int
    GetCacheStats returns template cache statistics

func (t *TemplateService) GetTemplate() *template.Template
    GetTemplate returns the internal template for Gin integration

func (t *TemplateService) HasTemplate(templateName string) bool

func (t *TemplateService) ListTemplates() []string

func (t *TemplateService) Reload(templatesPath string) error

func (t *TemplateService) Render(w io.Writer, templateName string, data any) error

func (t *TemplateService) RenderToString(templateName string, data any) (string, error)

func (t *TemplateService) Shutdown()
    Shutdown gracefully shuts down the template service

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
    TemplateServiceInterface defines the interface for template operations

