// Package constants provides application-wide constants for the MarkGo blog engine.
// It includes version information, default values, and shared constants used across services.
package constants

import "time"

// Application metadata
const (
	AppName    = "MarkGo"
	AppVersion = "v2.0.0"
)

// File paths and directories
const (
	DefaultArticlesPath  = "articles"
	DefaultStaticPath    = "web/static"
	DefaultTemplatesPath = "web/templates"
	DefaultDistPath      = "dist"
)

// File extensions
var (
	SupportedMarkdownExtensions = []string{".md", ".markdown", ".mdown", ".mkd"}
	SupportedImageExtensions    = []string{".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp"}
)

// HTTP and server defaults
const (
	DefaultPort         = 3000
	DefaultReadTimeout  = 15 * time.Second
	DefaultWriteTimeout = 15 * time.Second
	DefaultIdleTimeout  = 60 * time.Second
)

// Cache defaults
const (
	DefaultCacheTTL             = 1 * time.Hour
	DefaultCacheMaxSize         = 1000
	DefaultCacheCleanupInterval = 10 * time.Minute
)

// Rate limiting defaults
const (
	DefaultGeneralRateLimit = 100
	DefaultContactRateLimit = 5
	DefaultRateLimitWindow  = 15 * time.Minute
)

// Blog defaults
const (
	DefaultPostsPerPage    = 10
	DefaultExcerptLength   = 160
	DefaultReadingSpeed    = 200 // words per minute
	DefaultLanguage        = "en"
	DefaultTheme           = "default"
	DefaultAuthor          = "MarkGo User"
	DefaultBlogTitle       = "My MarkGo Blog"
	DefaultBlogDescription = "A fast, modern blog powered by MarkGo"
)

// Content processing
const (
	DefaultMaxContentLength = 1024 * 1024 // 1MB
	DefaultMaxTitleLength   = 200
	DefaultMinTitleLength   = 3
)

// Template and UI constants
const (
	DefaultDateFormat     = "2006-01-02"
	DefaultDateTimeFormat = "2006-01-02T15:04:05Z07:00"
	DefaultTimeFormat     = "15:04"
)

// Environment values
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTest        = "test"
)

// HTTP headers and content types
const (
	ContentTypeHTML  = "text/html; charset=utf-8"
	ContentTypeJSON  = "application/json"
	ContentTypeXML   = "application/xml"
	ContentTypeRSS   = "application/rss+xml"
	ContentTypeFeed  = "application/json"
	ContentTypeCSS   = "text/css"
	ContentTypeJS    = "application/javascript"
	ContentTypeImage = "image/*"
)

// Search and filtering
const (
	DefaultSearchLimit        = 50
	DefaultSuggestionLimit    = 10
	MinSearchQueryLength      = 2
	MaxSearchQueryLength      = 100
	DefaultSearchScoreBoost   = 2.0
	DefaultTitleScoreWeight   = 3.0
	DefaultContentScoreWeight = 1.0
	DefaultTagScoreWeight     = 2.0
)

// Email defaults
const (
	DefaultEmailTimeout   = 30 * time.Second
	DefaultSMTPPort       = 587
	DefaultSMTPTLSPort    = 465
	DefaultMaxEmailLength = 10000
)

// File system constants
const (
	DefaultFilePermissions = 0o644
	DefaultDirPermissions  = 0o755
	DefaultMaxFileSize     = 50 * 1024 * 1024 // 50MB
)

// Validation constants
const (
	MinUsernameLength = 3
	MaxUsernameLength = 50
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MaxTagLength      = 50
	MaxCategoryLength = 100
)

// Error messages
const (
	ErrInvalidInput      = "invalid input provided"
	ErrFileNotFound      = "file not found"
	ErrPermissionDenied  = "permission denied"
	ErrInternalError     = "internal server error"
	ErrNotFound          = "resource not found"
	ErrUnauthorized      = "unauthorized access"
	ErrRateLimitExceeded = "rate limit exceeded"
)
