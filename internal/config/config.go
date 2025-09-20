// Package config provides configuration management for MarkGo blog engine.
// It handles environment variable parsing, validation, and structured configuration
// for all application components including server, cache, email, and preview services.
package config

import (
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	apperrors "github.com/vnykmshr/markgo/internal/errors"
)

// Environment constants
const (
	DevelopmentEnvironment = "development"
	ProductionEnvironment  = "production"
	TestEnvironment        = "test"
)

type Config struct {
	Environment   string          `json:"environment"`
	Port          int             `json:"port"`
	ArticlesPath  string          `json:"articles_path"`
	StaticPath    string          `json:"static_path"`
	TemplatesPath string          `json:"templates_path"`
	BaseURL       string          `json:"base_url"`
	Server        ServerConfig    `json:"server"`
	Cache         CacheConfig     `json:"cache"`
	Email         EmailConfig     `json:"email"`
	RateLimit     RateLimitConfig `json:"rate_limit"`
	CORS          CORSConfig      `json:"cors"`
	Admin         AdminConfig     `json:"admin"`
	Blog          BlogConfig      `json:"blog"`
	Comments      CommentsConfig  `json:"comments"`
	Logging       LoggingConfig   `json:"logging"`
	Analytics     AnalyticsConfig `json:"analytics"`
	Preview       PreviewConfig   `json:"preview"`
}

type ServerConfig struct {
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

type CacheConfig struct {
	TTL             time.Duration `json:"ttl"`
	MaxSize         int           `json:"max_size"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

type EmailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
	UseSSL   bool   `json:"use_ssl"`
}

type RateLimitConfig struct {
	General RateLimit `json:"general"`
	Contact RateLimit `json:"contact"`
}

type RateLimit struct {
	Requests int           `json:"requests"`
	Window   time.Duration `json:"window"`
}

type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

type AdminConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BlogConfig struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Author       string `json:"author"`
	AuthorEmail  string `json:"author_email"`
	Language     string `json:"language"`
	Theme        string `json:"theme"`
	PostsPerPage int    `json:"posts_per_page"`
}

type CommentsConfig struct {
	Enabled          bool   `json:"enabled"`
	Provider         string `json:"provider"`
	GiscusRepo       string `json:"giscus_repo"`
	GiscusRepoID     string `json:"giscus_repo_id"`
	GiscusCategory   string `json:"giscus_category"`
	GiscusCategoryID string `json:"giscus_category_id"`
	Theme            string `json:"theme"`
	Language         string `json:"language"`
	ReactionsEnabled bool   `json:"reactions_enabled"`
}

type LoggingConfig struct {
	Level      string `json:"level"`          // debug, info, warn, error
	Format     string `json:"format"`         // json, text
	Output     string `json:"output"`         // stdout, stderr, file
	File       string `json:"file,omitempty"` // log file path when output=file
	MaxSize    int    `json:"max_size"`       // max size in MB before rotation
	MaxBackups int    `json:"max_backups"`    // max number of backup files to keep
	MaxAge     int    `json:"max_age"`        // max age in days to keep backups
	Compress   bool   `json:"compress"`       // compress rotated files
	AddSource  bool   `json:"add_source"`     // add source file and line number
	TimeFormat string `json:"time_format"`    // custom time format for text logs
}

type AnalyticsConfig struct {
	Enabled    bool   `json:"enabled"`               // enable/disable analytics
	Provider   string `json:"provider,omitempty"`    // google, plausible, etc.
	TrackingID string `json:"tracking_id,omitempty"` // tracking/site ID
	Domain     string `json:"domain,omitempty"`      // domain for analytics (plausible)
	DataAPI    string `json:"data_api,omitempty"`    // custom API endpoint
	CustomCode string `json:"custom_code,omitempty"` // custom analytics code
}

type PreviewConfig struct {
	Enabled        bool          `json:"enabled"`
	Port           int           `json:"port"`
	BaseURL        string        `json:"base_url"`
	AuthToken      string        `json:"auth_token"`
	SessionTimeout time.Duration `json:"session_timeout"`
	MaxSessions    int           `json:"max_sessions"`
}

// ValidationWarning represents a configuration warning that doesn't prevent startup
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Level   string `json:"level"` // "warning", "recommendation"
}

// ValidationResult contains validation results including errors and warnings
type ValidationResult struct {
	Valid    bool                `json:"valid"`
	Errors   []error             `json:"errors"`
	Warnings []ValidationWarning `json:"warnings"`
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore errors as it's optional)
	if err := godotenv.Load(); err != nil {
		// Silently ignore .env file loading errors as it's optional
		// This is intentional behavior - .env files are not required
		_ = err // Explicitly acknowledge the ignored error
	}

	// Get environment first to determine appropriate defaults
	environment := getEnv("ENVIRONMENT", "development")

	// Set environment-aware rate limit defaults
	var generalRequestsDefault int
	var contactRequestsDefault int

	switch environment {
	case ProductionEnvironment:
		// Conservative limits for production
		generalRequestsDefault = 100 // 100 requests per 15 min = ~0.11 req/sec
		contactRequestsDefault = 5   // 5 contact submissions per hour
	case TestEnvironment:
		// Higher limits for automated testing
		generalRequestsDefault = 5000 // 5000 requests per 15 min = ~5.5 req/sec
		contactRequestsDefault = 50   // 50 contact submissions per hour for test suites
	default: // development
		// Permissive limits for development and manual testing
		generalRequestsDefault = 3000 // 3000 requests per 15 min = ~3.3 req/sec
		contactRequestsDefault = 20   // 20 contact submissions per hour
	}

	cfg := &Config{
		Environment:   environment,
		Port:          getEnvInt("PORT", 3000),
		ArticlesPath:  getEnv("ARTICLES_PATH", "./articles"),
		StaticPath:    getEnv("STATIC_PATH", "./web/static"),
		TemplatesPath: getEnv("TEMPLATES_PATH", "./web/templates"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:3000"),

		Server: ServerConfig{
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},

		Cache: CacheConfig{
			TTL:             getEnvDuration("CACHE_TTL", 1*time.Hour),
			MaxSize:         getEnvInt("CACHE_MAX_SIZE", 1000),
			CleanupInterval: getEnvDuration("CACHE_CLEANUP_INTERVAL", 10*time.Minute),
		},

		Email: EmailConfig{
			Host:     getEnv("EMAIL_HOST", "smtp.gmail.com"),
			Port:     getEnvInt("EMAIL_PORT", 587),
			Username: getEnv("EMAIL_USERNAME", ""),
			Password: getEnv("EMAIL_PASSWORD", ""),
			From:     getEnv("EMAIL_FROM", "noreply@yourdomain.com"),
			To:       getEnv("EMAIL_TO", "your.email@example.com"),
			UseSSL:   getEnvBool("EMAIL_USE_SSL", true),
		},

		RateLimit: RateLimitConfig{
			General: RateLimit{
				Requests: getEnvInt("RATE_LIMIT_GENERAL_REQUESTS", generalRequestsDefault),
				Window:   getEnvDuration("RATE_LIMIT_GENERAL_WINDOW", 15*time.Minute),
			},
			Contact: RateLimit{
				Requests: getEnvInt("RATE_LIMIT_CONTACT_REQUESTS", contactRequestsDefault),
				Window:   getEnvDuration("RATE_LIMIT_CONTACT_WINDOW", 1*time.Hour),
			},
		},

		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "https://yourdomain.com"}),
			AllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"}),
		},

		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", ""),
			Password: getEnv("ADMIN_PASSWORD", ""),
		},

		Blog: BlogConfig{
			Title:        getEnv("BLOG_TITLE", "Your Blog Title"),
			Description:  getEnv("BLOG_DESCRIPTION", "Your blog description goes here"),
			Author:       getEnv("BLOG_AUTHOR", "Your Name"),
			AuthorEmail:  getEnv("BLOG_AUTHOR_EMAIL", "your.email@example.com"),
			Language:     getEnv("BLOG_LANGUAGE", "en"),
			Theme:        getEnv("BLOG_THEME", "default"),
			PostsPerPage: getEnvInt("BLOG_POSTS_PER_PAGE", 10),
		},

		Comments: CommentsConfig{
			Enabled:          getEnvBool("COMMENTS_ENABLED", true),
			Provider:         getEnv("COMMENTS_PROVIDER", "giscus"),
			GiscusRepo:       getEnv("GISCUS_REPO", "yourusername/blog-comments"),
			GiscusRepoID:     getEnv("GISCUS_REPO_ID", ""),
			GiscusCategory:   getEnv("GISCUS_CATEGORY", "General"),
			GiscusCategoryID: getEnv("GISCUS_CATEGORY_ID", ""),
			Theme:            getEnv("GISCUS_THEME", "preferred_color_scheme"),
			Language:         getEnv("GISCUS_LANGUAGE", "en"),
			ReactionsEnabled: getEnvBool("GISCUS_REACTIONS_ENABLED", true),
		},

		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			File:       getEnv("LOG_FILE", ""),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
			Compress:   getEnvBool("LOG_COMPRESS", true),
			AddSource:  getEnvBool("LOG_ADD_SOURCE", false),
			TimeFormat: getEnv("LOG_TIME_FORMAT", "2006-01-02T15:04:05Z07:00"),
		},

		Analytics: AnalyticsConfig{
			Enabled:    getEnvBool("ANALYTICS_ENABLED", false),
			Provider:   getEnv("ANALYTICS_PROVIDER", ""),
			TrackingID: getEnv("ANALYTICS_TRACKING_ID", ""),
			Domain:     getEnv("ANALYTICS_DOMAIN", ""),
			DataAPI:    getEnv("ANALYTICS_DATA_API", ""),
			CustomCode: getEnv("ANALYTICS_CUSTOM_CODE", ""),
		},

		Preview: PreviewConfig{
			Enabled:        getEnvBool("PREVIEW_ENABLED", false),
			Port:           getEnvInt("PREVIEW_PORT", 3001),
			BaseURL:        getEnv("PREVIEW_BASE_URL", ""),
			AuthToken:      getEnv("PREVIEW_AUTH_TOKEN", ""),
			SessionTimeout: getEnvDuration("PREVIEW_SESSION_TIMEOUT", 2*time.Hour),
			MaxSessions:    getEnvInt("PREVIEW_MAX_SESSIONS", 50),
		},
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ValidateWithWarnings performs comprehensive validation and returns both errors and warnings
func (c *Config) ValidateWithWarnings() ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   make([]error, 0),
		Warnings: make([]ValidationWarning, 0),
	}

	// Perform standard validation
	if err := c.Validate(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err)
	}

	// Add warnings and recommendations
	c.addConfigurationWarnings(&result)
	c.addSecurityRecommendations(&result)
	c.addPerformanceRecommendations(&result)
	c.addProductionRecommendations(&result)

	return result
}

// addConfigurationWarnings adds configuration-related warnings
func (c *Config) addConfigurationWarnings(result *ValidationResult) {
	c.addProductionPlaceholderWarnings(result)
	c.addEmailConfigurationWarnings(result)
	c.addCORSConfigurationWarnings(result)
	c.addRateLimitWarnings(result)
}

// addProductionPlaceholderWarnings checks for placeholder values in production
func (c *Config) addProductionPlaceholderWarnings(result *ValidationResult) {
	if c.Environment != ProductionEnvironment {
		return
	}

	if strings.Contains(c.BaseURL, "localhost") {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "base_url",
			Message: "Base URL contains 'localhost' in production environment",
			Level:   "warning",
		})
	}

	if strings.Contains(c.BaseURL, "yourdomain.com") {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "base_url",
			Message: "Base URL contains placeholder domain 'yourdomain.com'",
			Level:   "warning",
		})
	}

	if c.Blog.Title == "Your Blog Title" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "blog.title",
			Message: "Blog title is still the default placeholder value",
			Level:   "warning",
		})
	}

	if c.Blog.Author == "Your Name" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "blog.author",
			Message: "Blog author is still the default placeholder value",
			Level:   "warning",
		})
	}

	if strings.Contains(c.Blog.AuthorEmail, "example.com") {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "blog.author_email",
			Message: "Blog author email contains placeholder domain",
			Level:   "warning",
		})
	}
}

// addEmailConfigurationWarnings checks for email security issues
func (c *Config) addEmailConfigurationWarnings(result *ValidationResult) {
	if c.Email.Username != "" && c.Email.Port == 25 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "email.port",
			Message: "Using port 25 for email may be insecure, consider using 587 (STARTTLS) or 465 (SSL/TLS)",
			Level:   "recommendation",
		})
	}
}

// addCORSConfigurationWarnings checks for CORS security issues
func (c *Config) addCORSConfigurationWarnings(result *ValidationResult) {
	if c.Environment != ProductionEnvironment || len(c.CORS.AllowedOrigins) == 0 {
		return
	}

	for _, origin := range c.CORS.AllowedOrigins {
		if origin == "*" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "cors.allowed_origins",
				Message: "Wildcard (*) in CORS allowed origins is not recommended for production",
				Level:   "warning",
			})
			break
		}
	}
}

// addRateLimitWarnings checks for overly permissive rate limits
func (c *Config) addRateLimitWarnings(result *ValidationResult) {
	if c.Environment != ProductionEnvironment {
		return
	}

	if c.RateLimit.General.Requests > 1000 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "rate_limit.general.requests",
			Message: "General rate limit is very high (>1000), consider lowering for better protection",
			Level:   "recommendation",
		})
	}

	if c.RateLimit.Contact.Requests > 50 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "rate_limit.contact.requests",
			Message: "Contact rate limit is high (>50), consider lowering to prevent spam",
			Level:   "recommendation",
		})
	}
}

// addSecurityRecommendations adds security-related recommendations
func (c *Config) addSecurityRecommendations(result *ValidationResult) {
	// Admin security checks
	if c.Admin.Username != "" && c.Environment == ProductionEnvironment {
		// Check for weak admin usernames
		weakUsernames := []string{"admin", "administrator", "root", "user", "test"}
		for _, weak := range weakUsernames {
			if strings.EqualFold(c.Admin.Username, weak) {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:   "admin.username",
					Message: fmt.Sprintf("Admin username '%s' is commonly used and may be targeted by attackers", weak),
					Level:   "recommendation",
				})
				break
			}
		}

		// Check password complexity (basic check)
		if len(c.Admin.Password) < 12 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "admin.password",
				Message: "Admin password should be at least 12 characters long for better security",
				Level:   "recommendation",
			})
		}

		// Check for simple passwords
		simplePasswords := []string{"password", "123456", "admin", "root"}
		for _, simple := range simplePasswords {
			if strings.EqualFold(c.Admin.Password, simple) {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:   "admin.password",
					Message: "Admin password is too simple and easily guessable",
					Level:   "warning",
				})
				break
			}
		}
	}

	// HTTPS recommendations
	if c.Environment == ProductionEnvironment {
		if strings.HasPrefix(c.BaseURL, "http://") && !strings.Contains(c.BaseURL, "localhost") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "base_url",
				Message: "Using HTTP instead of HTTPS in production is not recommended for security",
				Level:   "warning",
			})
		}
	}

	// Email security
	if c.Email.Username != "" && !c.Email.UseSSL && c.Email.Port != 587 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "email.use_ssl",
			Message: "Email configuration without SSL/TLS may transmit credentials insecurely",
			Level:   "recommendation",
		})
	}
}

// addPerformanceRecommendations adds performance-related recommendations
func (c *Config) addPerformanceRecommendations(result *ValidationResult) {
	// Cache recommendations
	if c.Cache.TTL < 5*time.Minute {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "cache.ttl",
			Message: "Cache TTL is very short (<5min), consider increasing for better performance",
			Level:   "recommendation",
		})
	}

	if c.Cache.TTL > 24*time.Hour {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "cache.ttl",
			Message: "Cache TTL is very long (>24h), users may not see updated content promptly",
			Level:   "recommendation",
		})
	}

	if c.Cache.MaxSize < 100 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "cache.max_size",
			Message: "Cache size is very small (<100), consider increasing for better cache hit ratio",
			Level:   "recommendation",
		})
	}

	// Server timeout recommendations
	if c.Server.ReadTimeout > 30*time.Second {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "server.read_timeout",
			Message: "Read timeout is quite long (>30s), may cause resource exhaustion under load",
			Level:   "recommendation",
		})
	}

	if c.Server.WriteTimeout > 30*time.Second {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "server.write_timeout",
			Message: "Write timeout is quite long (>30s), may cause resource exhaustion under load",
			Level:   "recommendation",
		})
	}

	// Blog performance
	if c.Blog.PostsPerPage > 50 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "blog.posts_per_page",
			Message: "Posts per page is high (>50), may affect page load performance",
			Level:   "recommendation",
		})
	}
}

// addProductionRecommendations adds production-specific recommendations
func (c *Config) addProductionRecommendations(result *ValidationResult) {
	if c.Environment != ProductionEnvironment {
		return
	}

	// Logging recommendations for production
	if c.Logging.Level == "debug" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "logging.level",
			Message: "Debug logging in production may impact performance and expose sensitive information",
			Level:   "warning",
		})
	}

	if c.Logging.Output == "stdout" && c.Logging.Level == "debug" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "logging.output",
			Message: "Consider using file output with log rotation for production debug logs",
			Level:   "recommendation",
		})
	}

	// Admin interface recommendations
	if c.Admin.Username == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "admin.username",
			Message: "Admin interface is disabled in production, consider enabling for management tasks",
			Level:   "recommendation",
		})
	}

	// Email configuration check for production
	if c.Email.Username == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "email.username",
			Message: "Email is not configured, contact form submissions will not be sent",
			Level:   "recommendation",
		})
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple split by comma - in production you might want a more robust parser
		return splitString(value, ",")
	}
	return defaultValue
}

func splitString(s, delimiter string) []string {
	if s == "" {
		return []string{}
	}

	var result []string
	parts := strings.SplitSeq(s, delimiter)
	for part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Validate validates all configuration settings
func (c *Config) Validate() error {
	// Validate basic fields
	if err := c.validateBasic(); err != nil {
		return err
	}

	// Validate server configuration
	if err := c.Server.Validate(); err != nil {
		return apperrors.NewConfigError("server", c.Server, "Invalid server configuration", err)
	}

	// Validate cache configuration
	if err := c.Cache.Validate(); err != nil {
		return apperrors.NewConfigError("cache", c.Cache, "Invalid cache configuration", err)
	}

	// Validate email configuration (only if configured)
	if c.Email.Username != "" || c.Email.Password != "" {
		if err := c.Email.Validate(); err != nil {
			return apperrors.NewConfigError("email", c.Email, "Invalid email configuration", err)
		}
	}

	// Validate rate limit configuration
	if err := c.RateLimit.Validate(); err != nil {
		return apperrors.NewConfigError("rate_limit", c.RateLimit, "Invalid rate limit configuration", err)
	}

	// Validate blog configuration
	if err := c.Blog.Validate(); err != nil {
		return apperrors.NewConfigError("blog", c.Blog, "Invalid blog configuration", err)
	}

	// Validate admin configuration
	if err := c.Admin.Validate(); err != nil {
		return apperrors.NewConfigError("admin", c.Admin, "Invalid admin configuration", err)
	}

	// Validate logging configuration
	if err := c.Logging.Validate(); err != nil {
		return apperrors.NewConfigError("logging", c.Logging, "Invalid logging configuration", err)
	}

	// Validate comments configuration
	if err := c.Comments.Validate(); err != nil {
		return apperrors.NewConfigError("comments", c.Comments, "Invalid comments configuration", err)
	}

	// Validate CORS configuration
	if err := c.CORS.Validate(); err != nil {
		return apperrors.NewConfigError("cors", c.CORS, "Invalid CORS configuration", err)
	}

	// Validate analytics configuration
	if err := c.Analytics.Validate(); err != nil {
		return apperrors.NewConfigError("analytics", c.Analytics, "Invalid analytics configuration", err)
	}

	// Validate preview configuration
	if err := c.Preview.Validate(); err != nil {
		return apperrors.NewConfigError("preview", c.Preview, "Invalid preview configuration", err)
	}

	return nil
}

func (c *Config) validateBasic() error {
	// Validate port
	if c.Port <= 0 || c.Port > 65535 {
		return apperrors.NewConfigError("port", c.Port, "Port must be between 1 and 65535", apperrors.ErrConfigValidation)
	}

	// Validate paths exist and are accessible
	if err := validatePath(c.ArticlesPath, "articles_path"); err != nil {
		return err
	}

	if err := validatePath(c.StaticPath, "static_path"); err != nil {
		return err
	}

	if err := validatePath(c.TemplatesPath, "templates_path"); err != nil {
		return err
	}

	// Validate base URL
	if c.BaseURL == "" {
		return apperrors.NewConfigError("base_url", c.BaseURL, "Base URL cannot be empty", apperrors.ErrMissingConfig)
	}

	if _, err := url.Parse(c.BaseURL); err != nil {
		return apperrors.NewConfigError("base_url", c.BaseURL, "Invalid base URL format", apperrors.ErrConfigValidation)
	}

	// Validate environment
	validEnvs := []string{DevelopmentEnvironment, ProductionEnvironment, TestEnvironment}
	if !contains(validEnvs, c.Environment) {
		return apperrors.NewConfigError("environment", c.Environment,
			"Environment must be one of: development, production, test", apperrors.ErrConfigValidation)
	}

	return nil
}

// Validate server configuration
func (s *ServerConfig) Validate() error {
	if s.ReadTimeout <= 0 {
		return apperrors.NewConfigError("read_timeout", s.ReadTimeout,
			"Read timeout must be positive", apperrors.ErrConfigValidation)
	}
	if s.WriteTimeout <= 0 {
		return apperrors.NewConfigError("write_timeout", s.WriteTimeout,
			"Write timeout must be positive", apperrors.ErrConfigValidation)
	}
	if s.IdleTimeout <= 0 {
		return apperrors.NewConfigError("idle_timeout", s.IdleTimeout,
			"Idle timeout must be positive", apperrors.ErrConfigValidation)
	}
	return nil
}

// Validate cache configuration
func (c *CacheConfig) Validate() error {
	if c.TTL <= 0 {
		return apperrors.NewConfigError("ttl", c.TTL, "Cache TTL must be positive", apperrors.ErrConfigValidation)
	}
	if c.MaxSize <= 0 {
		return apperrors.NewConfigError("max_size", c.MaxSize,
			"Cache max size must be positive", apperrors.ErrConfigValidation)
	}
	if c.CleanupInterval <= 0 {
		return apperrors.NewConfigError("cleanup_interval", c.CleanupInterval,
			"Cleanup interval must be positive", apperrors.ErrConfigValidation)
	}
	return nil
}

// Validate email configuration
func (e *EmailConfig) Validate() error {
	if e.Host == "" {
		return apperrors.NewConfigError("host", e.Host, "Email host cannot be empty", apperrors.ErrMissingConfig)
	}
	if e.Port <= 0 || e.Port > 65535 {
		return apperrors.NewConfigError("port", e.Port,
			"Email port must be between 1 and 65535", apperrors.ErrConfigValidation)
	}
	if e.Username == "" {
		return apperrors.NewConfigError("username", e.Username,
			"Email username cannot be empty when email is configured", apperrors.ErrMissingConfig)
	}
	if e.Password == "" {
		return apperrors.NewConfigError("password", "***",
			"Email password cannot be empty when email is configured", apperrors.ErrMissingConfig)
	}
	if e.From == "" {
		return apperrors.NewConfigError("from", e.From, "Email 'from' address cannot be empty", apperrors.ErrMissingConfig)
	}
	if e.To == "" {
		return apperrors.NewConfigError("to", e.To, "Email 'to' address cannot be empty", apperrors.ErrMissingConfig)
	}

	// Validate email addresses format
	if _, err := mail.ParseAddress(e.From); err != nil {
		return apperrors.NewConfigError("from", e.From, "Invalid 'from' email address format", apperrors.ErrConfigValidation)
	}
	if _, err := mail.ParseAddress(e.To); err != nil {
		return apperrors.NewConfigError("to", e.To, "Invalid 'to' email address format", apperrors.ErrConfigValidation)
	}

	return nil
}

// Validate rate limit configuration
func (r *RateLimitConfig) Validate() error {
	if err := r.General.Validate("general"); err != nil {
		return err
	}
	if err := r.Contact.Validate("contact"); err != nil {
		return err
	}
	return nil
}

// Validate individual rate limit
func (r *RateLimit) Validate(name string) error {
	if r.Requests <= 0 {
		return apperrors.NewConfigError("requests", r.Requests,
			name+" rate limit requests must be positive", apperrors.ErrConfigValidation)
	}
	if r.Window <= 0 {
		return apperrors.NewConfigError("window", r.Window,
			name+" rate limit window must be positive", apperrors.ErrConfigValidation)
	}
	return nil
}

// Validate blog configuration
func (b *BlogConfig) Validate() error {
	if b.Title == "" {
		return apperrors.NewConfigError("title", b.Title, "Blog title cannot be empty", apperrors.ErrMissingConfig)
	}
	if b.Author == "" {
		return apperrors.NewConfigError("author", b.Author, "Blog author cannot be empty", apperrors.ErrMissingConfig)
	}
	if b.PostsPerPage <= 0 {
		return apperrors.NewConfigError("posts_per_page", b.PostsPerPage,
			"Posts per page must be positive", apperrors.ErrConfigValidation)
	}
	if b.PostsPerPage > 100 {
		return apperrors.NewConfigError("posts_per_page", b.PostsPerPage,
			"Posts per page should not exceed 100 for performance reasons",
			apperrors.ErrConfigValidation)
	}
	if b.Language == "" {
		return apperrors.NewConfigError("language", b.Language, "Blog language cannot be empty", apperrors.ErrMissingConfig)
	}

	// Validate language code format (basic check)
	langRegex := regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)
	if !langRegex.MatchString(b.Language) {
		return apperrors.NewConfigError("language", b.Language,
			"Blog language must be a valid language code (e.g., 'en', 'en-US')",
			apperrors.ErrConfigValidation)
	}

	// Validate author email if provided
	if b.AuthorEmail != "" {
		if _, err := mail.ParseAddress(b.AuthorEmail); err != nil {
			return apperrors.NewConfigError("author_email", b.AuthorEmail,
				"Invalid author email address format", apperrors.ErrConfigValidation)
		}
	}

	return nil
}

// Validate admin configuration
func (a *AdminConfig) Validate() error {
	// Admin config is optional, but if username is provided, password must also be provided
	if a.Username != "" && a.Password == "" {
		return apperrors.NewConfigError("password", "***",
			"Admin password is required when username is provided", apperrors.ErrMissingConfig)
	}
	if a.Username != "" && a.Password == "changeme" {
		return apperrors.NewConfigError("password", "***",
			"Admin password must be changed from default value", apperrors.ErrConfigValidation)
	}
	return nil
}

// Validate logging configuration
func (l *LoggingConfig) Validate() error {
	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, l.Level) {
		return apperrors.NewConfigError("level", l.Level,
			"Log level must be one of: debug, info, warn, error", apperrors.ErrConfigValidation)
	}

	// Validate log format
	validFormats := []string{"json", "text"}
	if !contains(validFormats, l.Format) {
		return apperrors.NewConfigError("format", l.Format,
			"Log format must be one of: json, text", apperrors.ErrConfigValidation)
	}

	// Validate log output
	validOutputs := []string{"stdout", "stderr", "file"}
	if !contains(validOutputs, l.Output) {
		return apperrors.NewConfigError("output", l.Output,
			"Log output must be one of: stdout, stderr, file", apperrors.ErrConfigValidation)
	}

	// If output is file, file path is required
	if l.Output == "file" && l.File == "" {
		return apperrors.NewConfigError("file", l.File,
			"Log file path is required when output is 'file'", apperrors.ErrMissingConfig)
	}

	// Validate rotation settings
	if l.MaxSize <= 0 {
		return apperrors.NewConfigError("max_size", l.MaxSize, "Log max size must be positive", apperrors.ErrConfigValidation)
	}
	if l.MaxBackups < 0 {
		return apperrors.NewConfigError("max_backups", l.MaxBackups,
			"Log max backups cannot be negative", apperrors.ErrConfigValidation)
	}
	if l.MaxAge < 0 {
		return apperrors.NewConfigError("max_age", l.MaxAge, "Log max age cannot be negative", apperrors.ErrConfigValidation)
	}

	return nil
}

// Validate comments configuration
func (c *CommentsConfig) Validate() error {
	if !c.Enabled {
		return nil // Comments are disabled, skip validation
	}

	// Validate provider
	validProviders := []string{"giscus", "disqus", "utterances"}
	if !contains(validProviders, c.Provider) {
		return apperrors.NewConfigError("provider", c.Provider,
			"Comments provider must be one of: giscus, disqus, utterances",
			apperrors.ErrConfigValidation)
	}

	// Validate giscus configuration if using giscus
	if c.Provider == "giscus" {
		if c.GiscusRepo == "" {
			return apperrors.NewConfigError("giscus_repo", c.GiscusRepo,
				"Giscus repository is required when using giscus provider",
				apperrors.ErrMissingConfig)
		}

		// Validate repository format (owner/repo)
		repoRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)
		if !repoRegex.MatchString(c.GiscusRepo) {
			return apperrors.NewConfigError("giscus_repo", c.GiscusRepo,
				"Giscus repository must be in format 'owner/repo'",
				apperrors.ErrConfigValidation)
		}

		if c.GiscusCategory == "" {
			return apperrors.NewConfigError("giscus_category", c.GiscusCategory,
				"Giscus category is required when using giscus provider", apperrors.ErrMissingConfig)
		}
	}

	return nil
}

// Validate CORS configuration
func (c *CORSConfig) Validate() error {
	// Validate allowed origins
	for _, origin := range c.AllowedOrigins {
		if origin != "*" {
			if _, err := url.Parse(origin); err != nil {
				return apperrors.NewConfigError("allowed_origins", origin,
					"Invalid CORS allowed origin URL", apperrors.ErrConfigValidation)
			}
		}
	}

	// Validate allowed methods
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for _, method := range c.AllowedMethods {
		if !contains(validMethods, method) {
			return apperrors.NewConfigError("allowed_methods", method,
				"Invalid HTTP method in CORS allowed methods", apperrors.ErrConfigValidation)
		}
	}

	return nil
}

// validatePath validates that a path exists and is accessible
func validatePath(path, fieldName string) error {
	if path == "" {
		return apperrors.NewConfigError(fieldName, path, fieldName+" cannot be empty", apperrors.ErrMissingConfig)
	}

	// Convert to absolute path for validation
	absPath, err := filepath.Abs(path)
	if err != nil {
		return apperrors.NewConfigError(fieldName, path, "Invalid path format", apperrors.ErrConfigValidation)
	}

	// Check if directory exists and is actually a directory
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return apperrors.NewConfigError(fieldName, path, fieldName+" directory does not exist", apperrors.ErrConfigValidation)
	}
	if err != nil {
		return apperrors.NewConfigError(fieldName, path, "Cannot access "+fieldName, apperrors.ErrConfigValidation)
	}
	if !info.IsDir() {
		return apperrors.NewConfigError(fieldName, path, fieldName+" must be a directory", apperrors.ErrConfigValidation)
	}

	return nil
}

// Validate analytics configuration
func (a *AnalyticsConfig) Validate() error {
	// If analytics is disabled, no validation needed
	if !a.Enabled {
		return nil
	}

	// If enabled, provider must be specified
	if a.Provider == "" {
		return apperrors.NewConfigError("provider", a.Provider,
			"Analytics provider must be specified when analytics is enabled", apperrors.ErrConfigValidation)
	}

	// Validate provider and required fields
	switch a.Provider {
	case "google":
		if a.TrackingID == "" {
			return apperrors.NewConfigError("tracking_id", a.TrackingID,
				"Google Analytics requires a tracking ID", apperrors.ErrConfigValidation)
		}
	case "plausible":
		if a.Domain == "" {
			return apperrors.NewConfigError("domain", a.Domain,
				"Plausible Analytics requires a domain", apperrors.ErrConfigValidation)
		}
	case "custom":
		if a.CustomCode == "" {
			return apperrors.NewConfigError("custom_code", a.CustomCode,
				"Custom analytics requires custom code", apperrors.ErrConfigValidation)
		}
	default:
		// Allow other providers without strict validation
		// This allows for future extensibility
	}

	return nil
}

func (p *PreviewConfig) Validate() error {
	// If preview is disabled, no validation needed
	if !p.Enabled {
		return nil
	}

	// Validate port range
	if p.Port < 1 || p.Port > 65535 {
		return apperrors.NewConfigError("port", p.Port,
			"Preview port must be between 1 and 65535", apperrors.ErrConfigValidation)
	}

	// Validate session timeout
	if p.SessionTimeout < time.Minute {
		return apperrors.NewConfigError("session_timeout", p.SessionTimeout,
			"Preview session timeout must be at least 1 minute", apperrors.ErrConfigValidation)
	}

	// Validate max sessions
	if p.MaxSessions < 1 || p.MaxSessions > 1000 {
		return apperrors.NewConfigError("max_sessions", p.MaxSessions,
			"Preview max sessions must be between 1 and 1000", apperrors.ErrConfigValidation)
	}

	return nil
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
