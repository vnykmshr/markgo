package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
)

func TestConfig_Validate(t *testing.T) {
	// Create temporary directories for testing
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(tmpDir+"/articles", 0755))
	require.NoError(t, os.MkdirAll(tmpDir+"/static", 0755))
	require.NoError(t, os.MkdirAll(tmpDir+"/templates", 0755))

	t.Run("valid configuration", func(t *testing.T) {
		cfg := &Config{
			Environment:   "development",
			Port:          3000,
			ArticlesPath:  tmpDir + "/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "http://localhost:3000",
			Server: ServerConfig{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             1 * time.Hour,
				MaxSize:         1000,
				CleanupInterval: 10 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 100, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Test Blog",
				Author:       "Test Author",
				Language:     "en",
				PostsPerPage: 10,
			},
			Admin: AdminConfig{
				Username: "",
				Password: "",
			},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
			},
			Comments: CommentsConfig{
				Enabled: false,
			},
			CORS: CORSConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
				AllowedMethods: []string{"GET", "POST"},
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid port", func(t *testing.T) {
		cfg := &Config{
			Port:          0,
			ArticlesPath:  tmpDir + "/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "http://localhost:3000",
			Environment:   "development",
			Server: ServerConfig{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             1 * time.Hour,
				MaxSize:         1000,
				CleanupInterval: 10 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 100, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Test Blog",
				Author:       "Test Author",
				Language:     "en",
				PostsPerPage: 10,
			},
			Admin: AdminConfig{},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.True(t, apperrors.IsConfigurationError(err))
		assert.Contains(t, err.Error(), "Port must be between 1 and 65535")
	})

	t.Run("missing articles directory", func(t *testing.T) {
		cfg := &Config{
			Port:          3000,
			ArticlesPath:  "/nonexistent/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "http://localhost:3000",
			Environment:   "development",
			Server: ServerConfig{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             1 * time.Hour,
				MaxSize:         1000,
				CleanupInterval: 10 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 100, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Test Blog",
				Author:       "Test Author",
				Language:     "en",
				PostsPerPage: 10,
			},
			Admin: AdminConfig{},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.True(t, apperrors.IsConfigurationError(err))
		assert.Contains(t, err.Error(), "articles_path directory does not exist")
	})

	t.Run("invalid base URL", func(t *testing.T) {
		cfg := &Config{
			Port:          3000,
			ArticlesPath:  tmpDir + "/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "ht!tp://invalid url with spaces",
			Environment:   "development",
			Server: ServerConfig{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             1 * time.Hour,
				MaxSize:         1000,
				CleanupInterval: 10 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 100, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Test Blog",
				Author:       "Test Author",
				Language:     "en",
				PostsPerPage: 10,
			},
			Admin: AdminConfig{},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
			},
			Comments: CommentsConfig{
				Enabled: false,
			},
			CORS: CORSConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
				AllowedMethods: []string{"GET", "POST"},
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.True(t, apperrors.IsConfigurationError(err))
		assert.Contains(t, err.Error(), "Invalid base URL format")
	})

	t.Run("invalid environment", func(t *testing.T) {
		cfg := &Config{
			Environment:   "invalid-env",
			Port:          3000,
			ArticlesPath:  tmpDir + "/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "http://localhost:3000",
			Server: ServerConfig{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             1 * time.Hour,
				MaxSize:         1000,
				CleanupInterval: 10 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 100, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Test Blog",
				Author:       "Test Author",
				Language:     "en",
				PostsPerPage: 10,
			},
			Admin: AdminConfig{},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.True(t, apperrors.IsConfigurationError(err))
		assert.Contains(t, err.Error(), "Environment must be one of: development, production, test")
	})
}

func TestServerConfig_Validate(t *testing.T) {
	t.Run("valid server config", func(t *testing.T) {
		cfg := ServerConfig{
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid read timeout", func(t *testing.T) {
		cfg := ServerConfig{
			ReadTimeout:  0,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Read timeout must be positive")
	})

	t.Run("invalid write timeout", func(t *testing.T) {
		cfg := ServerConfig{
			ReadTimeout:  15 * time.Second,
			WriteTimeout: -5 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Write timeout must be positive")
	})

	t.Run("invalid idle timeout", func(t *testing.T) {
		cfg := ServerConfig{
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  0,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Idle timeout must be positive")
	})
}

func TestCacheConfig_Validate(t *testing.T) {
	t.Run("valid cache config", func(t *testing.T) {
		cfg := CacheConfig{
			TTL:             1 * time.Hour,
			MaxSize:         1000,
			CleanupInterval: 10 * time.Minute,
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid TTL", func(t *testing.T) {
		cfg := CacheConfig{
			TTL:             0,
			MaxSize:         1000,
			CleanupInterval: 10 * time.Minute,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cache TTL must be positive")
	})

	t.Run("invalid max size", func(t *testing.T) {
		cfg := CacheConfig{
			TTL:             1 * time.Hour,
			MaxSize:         0,
			CleanupInterval: 10 * time.Minute,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cache max size must be positive")
	})

	t.Run("invalid cleanup interval", func(t *testing.T) {
		cfg := CacheConfig{
			TTL:             1 * time.Hour,
			MaxSize:         1000,
			CleanupInterval: 0,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cleanup interval must be positive")
	})
}

func TestEmailConfig_Validate(t *testing.T) {
	t.Run("valid email config", func(t *testing.T) {
		cfg := EmailConfig{
			Host:     "smtp.gmail.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			From:     "noreply@example.com",
			To:       "admin@example.com",
			UseSSL:   true,
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty host", func(t *testing.T) {
		cfg := EmailConfig{
			Host:     "",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			From:     "noreply@example.com",
			To:       "admin@example.com",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Email host cannot be empty")
	})

	t.Run("invalid port", func(t *testing.T) {
		cfg := EmailConfig{
			Host:     "smtp.gmail.com",
			Port:     70000,
			Username: "user@example.com",
			Password: "password",
			From:     "noreply@example.com",
			To:       "admin@example.com",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Email port must be between 1 and 65535")
	})

	t.Run("empty username", func(t *testing.T) {
		cfg := EmailConfig{
			Host:     "smtp.gmail.com",
			Port:     587,
			Username: "",
			Password: "password",
			From:     "noreply@example.com",
			To:       "admin@example.com",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Email username cannot be empty")
	})

	t.Run("empty password", func(t *testing.T) {
		cfg := EmailConfig{
			Host:     "smtp.gmail.com",
			Port:     587,
			Username: "user@example.com",
			Password: "",
			From:     "noreply@example.com",
			To:       "admin@example.com",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Email password cannot be empty")
	})
}

func TestRateLimitConfig_Validate(t *testing.T) {
	t.Run("valid rate limit config", func(t *testing.T) {
		cfg := RateLimitConfig{
			General: RateLimit{Requests: 100, Window: 15 * time.Minute},
			Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid general requests", func(t *testing.T) {
		cfg := RateLimitConfig{
			General: RateLimit{Requests: 0, Window: 15 * time.Minute},
			Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "general rate limit requests must be positive")
	})

	t.Run("invalid contact window", func(t *testing.T) {
		cfg := RateLimitConfig{
			General: RateLimit{Requests: 100, Window: 15 * time.Minute},
			Contact: RateLimit{Requests: 5, Window: 0},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "contact rate limit window must be positive")
	})
}

func TestBlogConfig_Validate(t *testing.T) {
	t.Run("valid blog config", func(t *testing.T) {
		cfg := BlogConfig{
			Title:        "My Blog",
			Author:       "John Doe",
			Language:     "en",
			PostsPerPage: 10,
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty title", func(t *testing.T) {
		cfg := BlogConfig{
			Title:        "",
			Author:       "John Doe",
			Language:     "en",
			PostsPerPage: 10,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Blog title cannot be empty")
	})

	t.Run("empty author", func(t *testing.T) {
		cfg := BlogConfig{
			Title:        "My Blog",
			Author:       "",
			Language:     "en",
			PostsPerPage: 10,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Blog author cannot be empty")
	})

	t.Run("invalid posts per page", func(t *testing.T) {
		cfg := BlogConfig{
			Title:        "My Blog",
			Author:       "John Doe",
			Language:     "en",
			PostsPerPage: 0,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Posts per page must be positive")
	})

	t.Run("empty language", func(t *testing.T) {
		cfg := BlogConfig{
			Title:        "My Blog",
			Author:       "John Doe",
			Language:     "",
			PostsPerPage: 10,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Blog language cannot be empty")
	})
}

func TestAdminConfig_Validate(t *testing.T) {
	t.Run("valid admin config - both empty", func(t *testing.T) {
		cfg := AdminConfig{
			Username: "",
			Password: "",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid admin config - both provided", func(t *testing.T) {
		cfg := AdminConfig{
			Username: "admin",
			Password: "securepassword",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("username provided but no password", func(t *testing.T) {
		cfg := AdminConfig{
			Username: "admin",
			Password: "",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Admin password is required when username is provided")
	})

	t.Run("default password", func(t *testing.T) {
		cfg := AdminConfig{
			Username: "admin",
			Password: "changeme",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Admin password must be changed from default value")
	})
}

func TestLoggingConfig_Validate(t *testing.T) {
	t.Run("valid logging config - stdout", func(t *testing.T) {
		cfg := LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid logging config - file", func(t *testing.T) {
		cfg := LoggingConfig{
			Level:      "debug",
			Format:     "text",
			Output:     "file",
			File:       "/var/log/markgo.log",
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     30,
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid log level", func(t *testing.T) {
		cfg := LoggingConfig{
			Level:   "invalid",
			Format:  "json",
			Output:  "stdout",
			MaxSize: 100,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Log level must be one of: debug, info, warn, error")
	})

	t.Run("invalid log format", func(t *testing.T) {
		cfg := LoggingConfig{
			Level:   "info",
			Format:  "xml",
			Output:  "stdout",
			MaxSize: 100,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Log format must be one of: json, text")
	})

	t.Run("invalid log output", func(t *testing.T) {
		cfg := LoggingConfig{
			Level:   "info",
			Format:  "json",
			Output:  "network",
			MaxSize: 100,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Log output must be one of: stdout, stderr, file")
	})

	t.Run("file output without file path", func(t *testing.T) {
		cfg := LoggingConfig{
			Level:   "info",
			Format:  "json",
			Output:  "file",
			File:    "",
			MaxSize: 100,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Log file path is required when output is 'file'")
	})

	t.Run("invalid max size", func(t *testing.T) {
		cfg := LoggingConfig{
			Level:   "info",
			Format:  "json",
			Output:  "stdout",
			MaxSize: 0,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Log max size must be positive")
	})

	t.Run("negative max backups", func(t *testing.T) {
		cfg := LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: -1,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Log max backups cannot be negative")
	})

	t.Run("negative max age", func(t *testing.T) {
		cfg := LoggingConfig{
			Level:   "info",
			Format:  "json",
			Output:  "stdout",
			MaxSize: 100,
			MaxAge:  -5,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Log max age cannot be negative")
	})
}

func TestLoadWithValidationFailures(t *testing.T) {
	clearEnvVars()

	// Test with missing directories
	t.Run("load with missing directories", func(t *testing.T) {
		os.Setenv("ARTICLES_PATH", "/nonexistent/articles")
		defer os.Unsetenv("ARTICLES_PATH")

		_, err := Load()
		assert.Error(t, err)
		assert.True(t, apperrors.IsConfigurationError(err))
		assert.Contains(t, err.Error(), "articles_path directory does not exist")
	})

	// Test with invalid port
	t.Run("load with invalid port", func(t *testing.T) {
		// Create temporary directories
		tmpDir := t.TempDir()
		require.NoError(t, os.MkdirAll(tmpDir+"/articles", 0755))
		require.NoError(t, os.MkdirAll(tmpDir+"/static", 0755))
		require.NoError(t, os.MkdirAll(tmpDir+"/templates", 0755))

		os.Setenv("ARTICLES_PATH", tmpDir+"/articles")
		os.Setenv("STATIC_PATH", tmpDir+"/static")
		os.Setenv("TEMPLATES_PATH", tmpDir+"/templates")
		os.Setenv("PORT", "70000")
		defer func() {
			os.Unsetenv("ARTICLES_PATH")
			os.Unsetenv("STATIC_PATH")
			os.Unsetenv("TEMPLATES_PATH")
			os.Unsetenv("PORT")
		}()

		_, err := Load()
		assert.Error(t, err)
		assert.True(t, apperrors.IsConfigurationError(err))
		assert.Contains(t, err.Error(), "Port must be between 1 and 65535")
	})

	// Test with invalid environment
	t.Run("load with invalid environment", func(t *testing.T) {
		// Create temporary directories
		tmpDir := t.TempDir()
		require.NoError(t, os.MkdirAll(tmpDir+"/articles", 0755))
		require.NoError(t, os.MkdirAll(tmpDir+"/static", 0755))
		require.NoError(t, os.MkdirAll(tmpDir+"/templates", 0755))

		os.Setenv("ARTICLES_PATH", tmpDir+"/articles")
		os.Setenv("STATIC_PATH", tmpDir+"/static")
		os.Setenv("TEMPLATES_PATH", tmpDir+"/templates")
		os.Setenv("ENVIRONMENT", "invalid")
		defer func() {
			os.Unsetenv("ARTICLES_PATH")
			os.Unsetenv("STATIC_PATH")
			os.Unsetenv("TEMPLATES_PATH")
			os.Unsetenv("ENVIRONMENT")
		}()

		_, err := Load()
		assert.Error(t, err)
		assert.True(t, apperrors.IsConfigurationError(err))
		assert.Contains(t, err.Error(), "Environment must be one of: development, production, test")
	})
}

func TestConfig_ValidateWithWarnings(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(tmpDir+"/articles", 0755))
	require.NoError(t, os.MkdirAll(tmpDir+"/static", 0755))
	require.NoError(t, os.MkdirAll(tmpDir+"/templates", 0755))

	t.Run("debug warnings output", func(t *testing.T) {
		cfg := &Config{
			Environment:   "production",
			Port:          3000,
			ArticlesPath:  tmpDir + "/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "http://localhost:3000",
			Server: ServerConfig{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             1 * time.Hour,
				MaxSize:         1000,
				CleanupInterval: 10 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 100, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Test Blog",
				Author:       "Test Author",
				Language:     "en",
				PostsPerPage: 10,
			},
			Admin: AdminConfig{
				Username: "",
				Password: "",
			},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
			},
			Comments: CommentsConfig{
				Enabled: false,
			},
			CORS: CORSConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
				AllowedMethods: []string{"GET", "POST"},
			},
		}

		result := cfg.ValidateWithWarnings()
		t.Logf("Validation result: Valid=%v, Errors=%d, Warnings=%d", result.Valid, len(result.Errors), len(result.Warnings))
		for i, warning := range result.Warnings {
			t.Logf("Warning %d: Field=%s, Message=%s, Level=%s", i, warning.Field, warning.Message, warning.Level)
		}
		for i, err := range result.Errors {
			t.Logf("Error %d: %s", i, err.Error())
		}
		// Just check it doesn't panic for now
		assert.NotNil(t, result)
	})

	t.Run("development environment debug", func(t *testing.T) {
		cfg := &Config{
			Environment:   "development",
			Port:          3000,
			ArticlesPath:  tmpDir + "/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "http://localhost:3000",
			Server: ServerConfig{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             1 * time.Hour,
				MaxSize:         1000,
				CleanupInterval: 10 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 100, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Test Blog",
				Author:       "Test Author",
				Language:     "en",
				PostsPerPage: 10,
			},
			Admin: AdminConfig{
				Username: "",
				Password: "",
			},
			Logging: LoggingConfig{
				Level:      "debug",
				Format:     "text",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
			},
			Comments: CommentsConfig{
				Enabled: false,
			},
			CORS: CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST"},
			},
		}

		result := cfg.ValidateWithWarnings()
		t.Logf("Development validation result: Valid=%v, Errors=%d, Warnings=%d", result.Valid, len(result.Errors), len(result.Warnings))
		for i, warning := range result.Warnings {
			t.Logf("Warning %d: Field=%s, Message=%s, Level=%s", i, warning.Field, warning.Message, warning.Level)
		}
		for i, err := range result.Errors {
			t.Logf("Error %d: %s", i, err.Error())
		}
		// Just check it doesn't panic for now
		assert.NotNil(t, result)
	})

	t.Run("production environment with admin disabled", func(t *testing.T) {
		cfg := &Config{
			Environment:   "production",
			Port:          3000,
			ArticlesPath:  tmpDir + "/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "http://localhost:3000",
			Server: ServerConfig{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             1 * time.Hour,
				MaxSize:         1000,
				CleanupInterval: 10 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 100, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Test Blog",
				Author:       "Test Author",
				Language:     "en",
				PostsPerPage: 10,
			},
			Admin: AdminConfig{
				Username: "",
				Password: "",
			},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
			},
			Comments: CommentsConfig{
				Enabled: false,
			},
			CORS: CORSConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
				AllowedMethods: []string{"GET", "POST"},
			},
		}

		result := cfg.ValidateWithWarnings()
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.NotEmpty(t, result.Warnings)

		// Check for localhost warning in production
		foundLocalhostWarning := false
		foundAdminWarning := false
		for _, warning := range result.Warnings {
			if warning.Field == "base_url" && strings.Contains(warning.Message, "localhost") {
				foundLocalhostWarning = true
			}
			if warning.Field == "admin.username" && warning.Level == "recommendation" {
				foundAdminWarning = true
			}
		}
		assert.True(t, foundLocalhostWarning, "Expected localhost warning in production")
		assert.True(t, foundAdminWarning, "Expected admin disabled warning in production")
	})

	t.Run("production environment with security recommendations", func(t *testing.T) {
		cfg := &Config{
			Environment:   "production",
			Port:          8080,
			ArticlesPath:  tmpDir + "/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "https://example.com",
			Server: ServerConfig{
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             2 * time.Hour,
				MaxSize:         5000,
				CleanupInterval: 30 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 1000, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 10, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Production Blog",
				Author:       "Production Author",
				Language:     "en",
				PostsPerPage: 20,
			},
			Admin: AdminConfig{
				Username: "",
				Password: "",
			},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "file",
				File:       "/var/log/markgo.log",
				MaxSize:    200,
				MaxBackups: 10,
				MaxAge:     30,
			},
			Comments: CommentsConfig{
				Enabled:        true,
				Provider:       "giscus",
				GiscusRepo:     "user/repo",
				GiscusCategory: "General",
			},
			CORS: CORSConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET", "POST"},
			},
		}

		result := cfg.ValidateWithWarnings()
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.NotEmpty(t, result.Warnings)

		// Check for production-specific recommendations
		foundAdminWarning := false
		for _, warning := range result.Warnings {
			if warning.Field == "admin.username" && warning.Level == "recommendation" {
				foundAdminWarning = true
			}
		}
		assert.True(t, foundAdminWarning, "Expected admin disabled warning in production")
	})

	t.Run("basic validation success", func(t *testing.T) {
		cfg := &Config{
			Environment:   "development",
			Port:          3000,
			ArticlesPath:  tmpDir + "/articles",
			StaticPath:    tmpDir + "/static",
			TemplatesPath: tmpDir + "/templates",
			BaseURL:       "http://localhost:3000",
			Server: ServerConfig{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Cache: CacheConfig{
				TTL:             1 * time.Hour,
				MaxSize:         1000,
				CleanupInterval: 10 * time.Minute,
			},
			RateLimit: RateLimitConfig{
				General: RateLimit{Requests: 100, Window: 15 * time.Minute},
				Contact: RateLimit{Requests: 5, Window: 1 * time.Hour},
			},
			Blog: BlogConfig{
				Title:        "Test Blog",
				Author:       "Test Author",
				Language:     "en",
				PostsPerPage: 10,
			},
			Admin: AdminConfig{},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
			},
			Comments: CommentsConfig{Enabled: false},
			CORS: CORSConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
				AllowedMethods: []string{"GET", "POST"},
			},
		}

		result := cfg.ValidateWithWarnings()
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		// Development environment might have few or no warnings
	})
}

func TestCommentsConfig_Validate(t *testing.T) {
	t.Run("valid disabled comments", func(t *testing.T) {
		cfg := CommentsConfig{
			Enabled: false,
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid giscus config", func(t *testing.T) {
		cfg := CommentsConfig{
			Enabled:        true,
			Provider:       "giscus",
			GiscusRepo:     "user/repo",
			GiscusCategory: "General",
			Theme:          "light",
			Language:       "en",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("enabled but no provider", func(t *testing.T) {
		cfg := CommentsConfig{
			Enabled:  true,
			Provider: "",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Comments provider must be one of: giscus, disqus")
	})

	t.Run("invalid provider", func(t *testing.T) {
		cfg := CommentsConfig{
			Enabled:  true,
			Provider: "invalid",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Comments provider must be one of: giscus, disqus")
	})

	t.Run("giscus without repo", func(t *testing.T) {
		cfg := CommentsConfig{
			Enabled:    true,
			Provider:   "giscus",
			GiscusRepo: "",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Giscus repository is required")
	})

	t.Run("giscus without category", func(t *testing.T) {
		cfg := CommentsConfig{
			Enabled:        true,
			Provider:       "giscus",
			GiscusRepo:     "user/repo",
			GiscusCategory: "",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Giscus category is required")
	})
}

func TestCORSConfig_Validate(t *testing.T) {
	t.Run("valid CORS config", func(t *testing.T) {
		cfg := CORSConfig{
			AllowedOrigins: []string{"https://example.com", "https://api.example.com"},
			AllowedMethods: []string{"GET", "POST", "PUT"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty allowed origins", func(t *testing.T) {
		cfg := CORSConfig{
			AllowedOrigins: []string{},
			AllowedMethods: []string{"GET"},
		}
		err := cfg.Validate()
		assert.NoError(t, err) // No error expected since we just loop through origins
	})

	t.Run("empty allowed methods", func(t *testing.T) {
		cfg := CORSConfig{
			AllowedOrigins: []string{"https://example.com"},
			AllowedMethods: []string{},
		}
		err := cfg.Validate()
		assert.NoError(t, err) // No error expected since we just loop through methods
	})

	t.Run("invalid HTTP method", func(t *testing.T) {
		cfg := CORSConfig{
			AllowedOrigins: []string{"https://example.com"},
			AllowedMethods: []string{"GET", "INVALID"},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid HTTP method in CORS allowed methods")
	})

	t.Run("invalid origin format", func(t *testing.T) {
		cfg := CORSConfig{
			AllowedOrigins: []string{"ht!tp://invalid url with spaces"},
			AllowedMethods: []string{"GET"},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid CORS allowed origin URL")
	})
}

func TestValidatePath(t *testing.T) {
	// Create temporary directories for testing
	tmpDir := t.TempDir()
	validPath := tmpDir + "/valid"
	require.NoError(t, os.MkdirAll(validPath, 0755))

	t.Run("valid existing path", func(t *testing.T) {
		err := validatePath(validPath, "test_field", true, false)
		assert.NoError(t, err)
	})

	t.Run("non-existent path", func(t *testing.T) {
		err := validatePath("/nonexistent/path", "test_field", true, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test_field directory does not exist")
	})

	t.Run("empty path", func(t *testing.T) {
		err := validatePath("", "test_field", true, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test_field cannot be empty")
	})

	t.Run("path is file not directory", func(t *testing.T) {
		// Create a file instead of directory
		filePath := tmpDir + "/file.txt"
		require.NoError(t, os.WriteFile(filePath, []byte("test"), 0644))

		err := validatePath(filePath, "test_field", true, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test_field must be a directory")
	})
}
