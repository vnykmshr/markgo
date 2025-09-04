package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Password: getEnv("ADMIN_PASSWORD", "changeme"),
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
