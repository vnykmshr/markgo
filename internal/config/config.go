package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
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
	GiscusRepoId     string `json:"giscus_repo_id"`
	GiscusCategory   string `json:"giscus_category"`
	GiscusCategoryId string `json:"giscus_category_id"`
	Theme            string `json:"theme"`
	Language         string `json:"language"`
	ReactionsEnabled bool   `json:"reactions_enabled"`
}

type LoggingConfig struct {
	Level          string `json:"level"`           // debug, info, warn, error
	Format         string `json:"format"`          // json, text
	Output         string `json:"output"`          // stdout, stderr, file
	File           string `json:"file,omitempty"`  // log file path when output=file
	MaxSize        int    `json:"max_size"`        // max size in MB before rotation
	MaxBackups     int    `json:"max_backups"`     // max number of backup files to keep
	MaxAge         int    `json:"max_age"`         // max age in days to keep backups
	Compress       bool   `json:"compress"`        // compress rotated files
	AddSource      bool   `json:"add_source"`      // add source file and line number
	TimeFormat     string `json:"time_format"`     // custom time format for text logs
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		Environment:   getEnv("ENVIRONMENT", "development"),
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
				Requests: getEnvInt("RATE_LIMIT_GENERAL_REQUESTS", 100),
				Window:   getEnvDuration("RATE_LIMIT_GENERAL_WINDOW", 15*time.Minute),
			},
			Contact: RateLimit{
				Requests: getEnvInt("RATE_LIMIT_CONTACT_REQUESTS", 5),
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
			GiscusRepoId:     getEnv("GISCUS_REPO_ID", ""),
			GiscusCategory:   getEnv("GISCUS_CATEGORY", "General"),
			GiscusCategoryId: getEnv("GISCUS_CATEGORY_ID", ""),
			Theme:            getEnv("GISCUS_THEME", "preferred_color_scheme"),
			Language:         getEnv("GISCUS_LANGUAGE", "en"),
			ReactionsEnabled: getEnvBool("GISCUS_REACTIONS_ENABLED", true),
		},

		Logging: LoggingConfig{
			Level:          getEnv("LOG_LEVEL", "info"),
			Format:         getEnv("LOG_FORMAT", "json"),
			Output:         getEnv("LOG_OUTPUT", "stdout"),
			File:           getEnv("LOG_FILE", ""),
			MaxSize:        getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups:     getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:         getEnvInt("LOG_MAX_AGE", 28),
			Compress:       getEnvBool("LOG_COMPRESS", true),
			AddSource:      getEnvBool("LOG_ADD_SOURCE", false),
			TimeFormat:     getEnv("LOG_TIME_FORMAT", "2006-01-02T15:04:05Z07:00"),
		},
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
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

	return nil
}

func (c *Config) validateBasic() error {
	// Validate port
	if c.Port <= 0 || c.Port > 65535 {
		return apperrors.NewConfigError("port", c.Port, "Port must be between 1 and 65535", apperrors.ErrConfigValidation)
	}

	// Validate paths exist
	if _, err := os.Stat(c.ArticlesPath); os.IsNotExist(err) {
		return apperrors.NewConfigError("articles_path", c.ArticlesPath, "Articles directory does not exist", apperrors.ErrMissingConfig)
	}

	if _, err := os.Stat(c.StaticPath); os.IsNotExist(err) {
		return apperrors.NewConfigError("static_path", c.StaticPath, "Static files directory does not exist", apperrors.ErrMissingConfig)
	}

	if _, err := os.Stat(c.TemplatesPath); os.IsNotExist(err) {
		return apperrors.NewConfigError("templates_path", c.TemplatesPath, "Templates directory does not exist", apperrors.ErrMissingConfig)
	}

	// Validate base URL
	if c.BaseURL == "" {
		return apperrors.NewConfigError("base_url", c.BaseURL, "Base URL cannot be empty", apperrors.ErrMissingConfig)
	}

	if _, err := url.Parse(c.BaseURL); err != nil {
		return apperrors.NewConfigError("base_url", c.BaseURL, "Invalid base URL format", apperrors.ErrConfigValidation)
	}

	// Validate environment
	validEnvs := []string{"development", "production", "test"}
	if !contains(validEnvs, c.Environment) {
		return apperrors.NewConfigError("environment", c.Environment, "Environment must be one of: development, production, test", apperrors.ErrConfigValidation)
	}

	return nil
}

// Validate server configuration
func (s *ServerConfig) Validate() error {
	if s.ReadTimeout <= 0 {
		return apperrors.NewConfigError("read_timeout", s.ReadTimeout, "Read timeout must be positive", apperrors.ErrConfigValidation)
	}
	if s.WriteTimeout <= 0 {
		return apperrors.NewConfigError("write_timeout", s.WriteTimeout, "Write timeout must be positive", apperrors.ErrConfigValidation)
	}
	if s.IdleTimeout <= 0 {
		return apperrors.NewConfigError("idle_timeout", s.IdleTimeout, "Idle timeout must be positive", apperrors.ErrConfigValidation)
	}
	return nil
}

// Validate cache configuration
func (c *CacheConfig) Validate() error {
	if c.TTL <= 0 {
		return apperrors.NewConfigError("ttl", c.TTL, "Cache TTL must be positive", apperrors.ErrConfigValidation)
	}
	if c.MaxSize <= 0 {
		return apperrors.NewConfigError("max_size", c.MaxSize, "Cache max size must be positive", apperrors.ErrConfigValidation)
	}
	if c.CleanupInterval <= 0 {
		return apperrors.NewConfigError("cleanup_interval", c.CleanupInterval, "Cleanup interval must be positive", apperrors.ErrConfigValidation)
	}
	return nil
}

// Validate email configuration
func (e *EmailConfig) Validate() error {
	if e.Host == "" {
		return apperrors.NewConfigError("host", e.Host, "Email host cannot be empty", apperrors.ErrMissingConfig)
	}
	if e.Port <= 0 || e.Port > 65535 {
		return apperrors.NewConfigError("port", e.Port, "Email port must be between 1 and 65535", apperrors.ErrConfigValidation)
	}
	if e.Username == "" {
		return apperrors.NewConfigError("username", e.Username, "Email username cannot be empty when email is configured", apperrors.ErrMissingConfig)
	}
	if e.Password == "" {
		return apperrors.NewConfigError("password", "***", "Email password cannot be empty when email is configured", apperrors.ErrMissingConfig)
	}
	if e.From == "" {
		return apperrors.NewConfigError("from", e.From, "Email 'from' address cannot be empty", apperrors.ErrMissingConfig)
	}
	if e.To == "" {
		return apperrors.NewConfigError("to", e.To, "Email 'to' address cannot be empty", apperrors.ErrMissingConfig)
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
		return apperrors.NewConfigError("requests", r.Requests, name+" rate limit requests must be positive", apperrors.ErrConfigValidation)
	}
	if r.Window <= 0 {
		return apperrors.NewConfigError("window", r.Window, name+" rate limit window must be positive", apperrors.ErrConfigValidation)
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
		return apperrors.NewConfigError("posts_per_page", b.PostsPerPage, "Posts per page must be positive", apperrors.ErrConfigValidation)
	}
	if b.Language == "" {
		return apperrors.NewConfigError("language", b.Language, "Blog language cannot be empty", apperrors.ErrMissingConfig)
	}
	return nil
}

// Validate admin configuration
func (a *AdminConfig) Validate() error {
	// Admin config is optional, but if username is provided, password must also be provided
	if a.Username != "" && a.Password == "" {
		return apperrors.NewConfigError("password", "***", "Admin password is required when username is provided", apperrors.ErrMissingConfig)
	}
	if a.Username != "" && a.Password == "changeme" {
		return apperrors.NewConfigError("password", "***", "Admin password must be changed from default value", apperrors.ErrConfigValidation)
	}
	return nil
}

// Validate logging configuration
func (l *LoggingConfig) Validate() error {
	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, l.Level) {
		return apperrors.NewConfigError("level", l.Level, "Log level must be one of: debug, info, warn, error", apperrors.ErrConfigValidation)
	}

	// Validate log format
	validFormats := []string{"json", "text"}
	if !contains(validFormats, l.Format) {
		return apperrors.NewConfigError("format", l.Format, "Log format must be one of: json, text", apperrors.ErrConfigValidation)
	}

	// Validate log output
	validOutputs := []string{"stdout", "stderr", "file"}
	if !contains(validOutputs, l.Output) {
		return apperrors.NewConfigError("output", l.Output, "Log output must be one of: stdout, stderr, file", apperrors.ErrConfigValidation)
	}

	// If output is file, file path is required
	if l.Output == "file" && l.File == "" {
		return apperrors.NewConfigError("file", l.File, "Log file path is required when output is 'file'", apperrors.ErrMissingConfig)
	}

	// Validate rotation settings
	if l.MaxSize <= 0 {
		return apperrors.NewConfigError("max_size", l.MaxSize, "Log max size must be positive", apperrors.ErrConfigValidation)
	}
	if l.MaxBackups < 0 {
		return apperrors.NewConfigError("max_backups", l.MaxBackups, "Log max backups cannot be negative", apperrors.ErrConfigValidation)
	}
	if l.MaxAge < 0 {
		return apperrors.NewConfigError("max_age", l.MaxAge, "Log max age cannot be negative", apperrors.ErrConfigValidation)
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
