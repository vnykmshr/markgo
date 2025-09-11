package config // import "github.com/vnykmshr/markgo/internal/config"


TYPES

type AdminConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *AdminConfig) Validate() error
    Validate admin configuration

type AnalyticsConfig struct {
	Enabled    bool   `json:"enabled"`               // enable/disable analytics
	Provider   string `json:"provider,omitempty"`    // google, plausible, etc.
	TrackingID string `json:"tracking_id,omitempty"` // tracking/site ID
	Domain     string `json:"domain,omitempty"`      // domain for analytics (plausible)
	DataAPI    string `json:"data_api,omitempty"`    // custom API endpoint
	CustomCode string `json:"custom_code,omitempty"` // custom analytics code
}

func (a *AnalyticsConfig) Validate() error
    Validate analytics configuration

type BlogConfig struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Author       string `json:"author"`
	AuthorEmail  string `json:"author_email"`
	Language     string `json:"language"`
	Theme        string `json:"theme"`
	PostsPerPage int    `json:"posts_per_page"`
}

func (b *BlogConfig) Validate() error
    Validate blog configuration

type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

func (c *CORSConfig) Validate() error
    Validate CORS configuration

type CacheConfig struct {
	TTL             time.Duration `json:"ttl"`
	MaxSize         int           `json:"max_size"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

func (c *CacheConfig) Validate() error
    Validate cache configuration

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

func (c *CommentsConfig) Validate() error
    Validate comments configuration

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
}

func Load() (*Config, error)

func (c *Config) Validate() error
    Validate validates all configuration settings

func (c *Config) ValidateWithWarnings() ValidationResult
    ValidateWithWarnings performs comprehensive validation and returns both
    errors and warnings

type EmailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
	UseSSL   bool   `json:"use_ssl"`
}

func (e *EmailConfig) Validate() error
    Validate email configuration

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

func (l *LoggingConfig) Validate() error
    Validate logging configuration

type RateLimit struct {
	Requests int           `json:"requests"`
	Window   time.Duration `json:"window"`
}

func (r *RateLimit) Validate(name string) error
    Validate individual rate limit

type RateLimitConfig struct {
	General RateLimit `json:"general"`
	Contact RateLimit `json:"contact"`
}

func (r *RateLimitConfig) Validate() error
    Validate rate limit configuration

type ServerConfig struct {
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

func (s *ServerConfig) Validate() error
    Validate server configuration

type ValidationResult struct {
	Valid    bool                `json:"valid"`
	Errors   []error             `json:"errors"`
	Warnings []ValidationWarning `json:"warnings"`
}
    ValidationResult contains validation results including errors and warnings

type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Level   string `json:"level"` // "warning", "recommendation"
}
    ValidationWarning represents a configuration warning that doesn't prevent
    startup

