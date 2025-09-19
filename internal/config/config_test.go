package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Clear environment variables
	clearEnvVars()

	// Create required directories for validation
	require.NoError(t, os.MkdirAll("./articles", 0755))
	require.NoError(t, os.MkdirAll("./web/static", 0755))
	require.NoError(t, os.MkdirAll("./web/templates", 0755))

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test default values
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "./articles", cfg.ArticlesPath)
	assert.Equal(t, "./web/static", cfg.StaticPath)
	assert.Equal(t, "./web/templates", cfg.TemplatesPath)
	assert.Equal(t, "http://localhost:3000", cfg.BaseURL)

	// Test server config defaults
	assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 15*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout)

	// Test cache config defaults
	assert.Equal(t, 1*time.Hour, cfg.Cache.TTL)
	assert.Equal(t, 1000, cfg.Cache.MaxSize)
	assert.Equal(t, 10*time.Minute, cfg.Cache.CleanupInterval)

	// Test email config defaults
	assert.Equal(t, "smtp.gmail.com", cfg.Email.Host)
	assert.Equal(t, 587, cfg.Email.Port)
	assert.Equal(t, "noreply@yourdomain.com", cfg.Email.From)
	assert.Equal(t, "your.email@example.com", cfg.Email.To)
	assert.True(t, cfg.Email.UseSSL)

	// Test rate limit defaults (development environment)
	assert.Equal(t, 3000, cfg.RateLimit.General.Requests) // Development default
	assert.Equal(t, 15*time.Minute, cfg.RateLimit.General.Window)
	assert.Equal(t, 20, cfg.RateLimit.Contact.Requests) // Development default
	assert.Equal(t, 1*time.Hour, cfg.RateLimit.Contact.Window)

	// Test CORS defaults
	assert.Contains(t, cfg.CORS.AllowedOrigins, "http://localhost:3000")
	assert.Contains(t, cfg.CORS.AllowedOrigins, "https://yourdomain.com")
	assert.Contains(t, cfg.CORS.AllowedMethods, "GET")
	assert.Contains(t, cfg.CORS.AllowedMethods, "POST")

	// Test admin defaults (disabled by default)
	assert.Equal(t, "", cfg.Admin.Username)
	assert.Equal(t, "", cfg.Admin.Password)

	// Test blog defaults
	// Test blog config defaults
	assert.Equal(t, "Your Blog Title", cfg.Blog.Title)
	assert.Equal(t, "Your blog description goes here", cfg.Blog.Description)
	assert.Equal(t, "Your Name", cfg.Blog.Author)
	assert.Equal(t, "your.email@example.com", cfg.Blog.AuthorEmail)
	assert.Equal(t, "en", cfg.Blog.Language)
	assert.Equal(t, "default", cfg.Blog.Theme)
	assert.Equal(t, 10, cfg.Blog.PostsPerPage)

	// Test comments config defaults
	assert.True(t, cfg.Comments.Enabled)
	assert.Equal(t, "giscus", cfg.Comments.Provider)
	assert.Equal(t, "yourusername/blog-comments", cfg.Comments.GiscusRepo)
	assert.Equal(t, "", cfg.Comments.GiscusRepoID)
	assert.Equal(t, "General", cfg.Comments.GiscusCategory)
	assert.Equal(t, "", cfg.Comments.GiscusCategoryID)
	assert.Equal(t, "preferred_color_scheme", cfg.Comments.Theme)
	assert.Equal(t, "en", cfg.Comments.Language)
	assert.True(t, cfg.Comments.ReactionsEnabled)
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Clear environment variables first
	clearEnvVars()

	// Create temporary directories for validation
	tmpDir, err := os.MkdirTemp("", "markgo-test-*")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(tmpDir+"/articles", 0755))
	require.NoError(t, os.MkdirAll(tmpDir+"/static", 0755))
	require.NoError(t, os.MkdirAll(tmpDir+"/templates", 0755))

	// Set environment variables
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("PORT", "8080")
	os.Setenv("ARTICLES_PATH", tmpDir+"/articles")
	os.Setenv("STATIC_PATH", tmpDir+"/static")
	os.Setenv("TEMPLATES_PATH", tmpDir+"/templates")
	os.Setenv("BASE_URL", "https://example.com")

	os.Setenv("SERVER_READ_TIMEOUT", "30s")
	os.Setenv("SERVER_WRITE_TIMEOUT", "30s")
	os.Setenv("SERVER_IDLE_TIMEOUT", "120s")

	os.Setenv("CACHE_TTL", "2h")
	os.Setenv("CACHE_MAX_SIZE", "2000")
	os.Setenv("CACHE_CLEANUP_INTERVAL", "20m")

	os.Setenv("EMAIL_HOST", "smtp.example.com")
	os.Setenv("EMAIL_PORT", "465")
	os.Setenv("EMAIL_USERNAME", "user@example.com")
	os.Setenv("EMAIL_PASSWORD", "password")
	os.Setenv("EMAIL_FROM", "noreply@example.com")
	os.Setenv("EMAIL_TO", "admin@example.com")
	os.Setenv("EMAIL_USE_SSL", "false")

	os.Setenv("RATE_LIMIT_GENERAL_REQUESTS", "200")
	os.Setenv("RATE_LIMIT_GENERAL_WINDOW", "30m")
	os.Setenv("RATE_LIMIT_CONTACT_REQUESTS", "10")
	os.Setenv("RATE_LIMIT_CONTACT_WINDOW", "2h")

	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://api.example.com")
	os.Setenv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE")
	os.Setenv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization")

	os.Setenv("ADMIN_USERNAME", "superuser")
	os.Setenv("ADMIN_PASSWORD", "secretpassword")

	os.Setenv("BLOG_TITLE", "My Custom Blog")
	os.Setenv("BLOG_DESCRIPTION", "A custom blog description")
	os.Setenv("BLOG_AUTHOR", "John Doe")
	os.Setenv("BLOG_AUTHOR_EMAIL", "john@example.com")
	os.Setenv("BLOG_LANGUAGE", "es")
	os.Setenv("BLOG_THEME", "dark")
	os.Setenv("BLOG_POSTS_PER_PAGE", "20")

	defer clearEnvVars()
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test environment variable values
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, tmpDir+"/articles", cfg.ArticlesPath)
	assert.Equal(t, tmpDir+"/static", cfg.StaticPath)
	assert.Equal(t, tmpDir+"/templates", cfg.TemplatesPath)
	assert.Equal(t, "https://example.com", cfg.BaseURL)

	// Test server config
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 120*time.Second, cfg.Server.IdleTimeout)

	// Test cache config
	assert.Equal(t, 2*time.Hour, cfg.Cache.TTL)
	assert.Equal(t, 2000, cfg.Cache.MaxSize)
	assert.Equal(t, 20*time.Minute, cfg.Cache.CleanupInterval)

	// Test email config
	assert.Equal(t, "smtp.example.com", cfg.Email.Host)
	assert.Equal(t, 465, cfg.Email.Port)
	assert.Equal(t, "user@example.com", cfg.Email.Username)
	assert.Equal(t, "password", cfg.Email.Password)
	assert.Equal(t, "noreply@example.com", cfg.Email.From)
	assert.Equal(t, "admin@example.com", cfg.Email.To)
	assert.False(t, cfg.Email.UseSSL)

	// Test rate limit config
	assert.Equal(t, 200, cfg.RateLimit.General.Requests)
	assert.Equal(t, 30*time.Minute, cfg.RateLimit.General.Window)
	assert.Equal(t, 10, cfg.RateLimit.Contact.Requests)
	assert.Equal(t, 2*time.Hour, cfg.RateLimit.Contact.Window)

	// Test CORS config
	assert.Contains(t, cfg.CORS.AllowedOrigins, "https://example.com")
	assert.Contains(t, cfg.CORS.AllowedOrigins, "https://api.example.com")
	assert.Contains(t, cfg.CORS.AllowedMethods, "GET")
	assert.Contains(t, cfg.CORS.AllowedMethods, "PUT")
	assert.Contains(t, cfg.CORS.AllowedHeaders, "Content-Type")
	assert.Contains(t, cfg.CORS.AllowedHeaders, "Authorization")

	// Test admin config
	assert.Equal(t, "superuser", cfg.Admin.Username)
	assert.Equal(t, "secretpassword", cfg.Admin.Password)

	// Test blog config
	assert.Equal(t, "My Custom Blog", cfg.Blog.Title)
	assert.Equal(t, "A custom blog description", cfg.Blog.Description)
	assert.Equal(t, "John Doe", cfg.Blog.Author)
	assert.Equal(t, "john@example.com", cfg.Blog.AuthorEmail)
	assert.Equal(t, "es", cfg.Blog.Language)
	assert.Equal(t, "dark", cfg.Blog.Theme)
	assert.Equal(t, 20, cfg.Blog.PostsPerPage)
}

func TestEnvironmentVariableParsing(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		setValue string
		expected interface{}
		testFunc func(string, interface{}) interface{}
	}{
		// String parsing tests
		{"string with value", "TEST_STR", "test_value", "test_value", func(key string, def interface{}) interface{} { return getEnv(key, def.(string)) }},
		{"string with default", "NON_EXISTING_STR", "", "default", func(key string, def interface{}) interface{} { return getEnv(key, def.(string)) }},

		// Integer parsing tests
		{"int with valid value", "TEST_INT", "42", 42, func(key string, def interface{}) interface{} { return getEnvInt(key, def.(int)) }},
		{"int with invalid value", "TEST_INVALID_INT", "not_a_number", 10, func(key string, def interface{}) interface{} { return getEnvInt(key, def.(int)) }},
		{"int with default", "NON_EXISTING_INT", "", 10, func(key string, def interface{}) interface{} { return getEnvInt(key, def.(int)) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.envVar, tt.setValue)
				defer os.Unsetenv(tt.envVar)
			}

			result := tt.testFunc(tt.envVar, tt.expected)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	testCases := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"TRUE", true},
		{"FALSE", false},
		{"t", true},
		{"f", false},
		{"T", true},
		{"F", false},
	}

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			os.Setenv("TEST_BOOL", tc.value)
			defer os.Unsetenv("TEST_BOOL")

			result := getEnvBool("TEST_BOOL", false)
			assert.Equal(t, tc.expected, result)
		})
	}

	// Test with invalid boolean
	os.Setenv("TEST_INVALID_BOOL", "not_a_bool")
	defer os.Unsetenv("TEST_INVALID_BOOL")

	result := getEnvBool("TEST_INVALID_BOOL", true)
	assert.Equal(t, true, result)

	// Test with non-existing environment variable
	result = getEnvBool("NON_EXISTING_BOOL", true)
	assert.Equal(t, true, result)
}

func TestStringSliceParsing(t *testing.T) {
	tests := []struct {
		name      string
		envVar    string
		setValue  string
		expected  []string
		delimiter string
	}{
		{"valid comma-separated values", "TEST_SLICE", "value1,value2,value3", []string{"value1", "value2", "value3"}, ","},
		{"values with spaces", "TEST_SLICE_SPACES", "value1, value2 , value3", []string{"value1", "value2", "value3"}, ","},
		{"empty values filtered", "TEST_SLICE_EMPTY", "value1,,value3", []string{"value1", "value3"}, ","},
		{"non-existing returns default", "NON_EXISTING_SLICE", "", []string{"default"}, ","},
		{"empty string returns default", "TEST_EMPTY_SLICE", "", []string{"default"}, ","},
		{"semicolon delimiter", "TEST_SEMICOLON", "a;b;c", []string{"a", "b", "c"}, ";"},
		{"single value", "TEST_SINGLE", "single", []string{"single"}, ","},
		{"whitespace handling", "TEST_WHITESPACE", "   a   ,   b   ,   c   ", []string{"a", "b", "c"}, ","},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.envVar, tt.setValue)
				defer os.Unsetenv(tt.envVar)
			}

			if tt.name == "non-existing returns default" || tt.name == "empty string returns default" {
				result := getEnvSlice(tt.envVar, tt.expected)
				assert.Equal(t, tt.expected, result)
			} else {
				result := splitString(tt.setValue, tt.delimiter)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Helper function to clear environment variables
func clearEnvVars() {
	vars := []string{
		"ENVIRONMENT", "PORT", "ARTICLES_PATH", "STATIC_PATH", "TEMPLATES_PATH", "BASE_URL",
		"SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT", "SERVER_IDLE_TIMEOUT",
		"CACHE_TTL", "CACHE_MAX_SIZE", "CACHE_CLEANUP_INTERVAL",
		"EMAIL_HOST", "EMAIL_PORT", "EMAIL_USERNAME", "EMAIL_PASSWORD", "EMAIL_FROM", "EMAIL_TO", "EMAIL_USE_SSL",
		"RATE_LIMIT_GENERAL_REQUESTS", "RATE_LIMIT_GENERAL_WINDOW", "RATE_LIMIT_CONTACT_REQUESTS", "RATE_LIMIT_CONTACT_WINDOW",
		"CORS_ALLOWED_ORIGINS", "CORS_ALLOWED_METHODS", "CORS_ALLOWED_HEADERS",
		"ADMIN_USERNAME", "ADMIN_PASSWORD",
		"BLOG_TITLE", "BLOG_DESCRIPTION", "BLOG_AUTHOR", "BLOG_AUTHOR_EMAIL", "BLOG_LANGUAGE", "BLOG_THEME", "BLOG_POSTS_PER_PAGE",
		"COMMENTS_ENABLED", "COMMENTS_PROVIDER", "GISCUS_REPO", "GISCUS_REPO_ID", "GISCUS_CATEGORY", "GISCUS_CATEGORY_ID", "GISCUS_THEME", "GISCUS_LANGUAGE", "GISCUS_REACTIONS_ENABLED",
	}

	for _, env := range vars {
		os.Unsetenv(env)
	}
}

// Benchmark tests
func BenchmarkLoad(b *testing.B) {
	clearEnvVars()

	for b.Loop() {
		_, err := Load()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetEnv(b *testing.B) {
	os.Setenv("BENCH_VAR", "benchmark_value")
	defer os.Unsetenv("BENCH_VAR")

	for b.Loop() {
		getEnv("BENCH_VAR", "default")
	}
}

func BenchmarkGetEnvInt(b *testing.B) {
	os.Setenv("BENCH_INT", "12345")
	defer os.Unsetenv("BENCH_INT")

	for b.Loop() {
		getEnvInt("BENCH_INT", 0)
	}
}

func BenchmarkGetEnvSlice(b *testing.B) {
	os.Setenv("BENCH_SLICE", "value1,value2,value3,value4,value5")
	defer os.Unsetenv("BENCH_SLICE")

	for b.Loop() {
		getEnvSlice("BENCH_SLICE", []string{"default"})
	}
}

func BenchmarkSplitString(b *testing.B) {
	input := "a,b,c,d,e,f,g,h,i,j"

	for b.Loop() {
		splitString(input, ",")
	}
}

// TestEnvironmentAwareRateLimiting tests that different environments get different rate limiting defaults
func TestEnvironmentAwareRateLimiting(t *testing.T) {
	tests := []struct {
		name                    string
		environment             string
		expectedGeneralRequests int
		expectedContactRequests int
	}{
		{
			name:                    "production environment",
			environment:             "production",
			expectedGeneralRequests: 100,
			expectedContactRequests: 5,
		},
		{
			name:                    "development environment",
			environment:             "development",
			expectedGeneralRequests: 3000,
			expectedContactRequests: 20,
		},
		{
			name:                    "test environment",
			environment:             "test",
			expectedGeneralRequests: 5000,
			expectedContactRequests: 50,
		},
		{
			name:                    "development environment (explicit)",
			environment:             "development",
			expectedGeneralRequests: 3000,
			expectedContactRequests: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			os.Setenv("ENVIRONMENT", tt.environment)
			defer os.Unsetenv("ENVIRONMENT")

			// Load config
			cfg, err := Load()
			assert.NoError(t, err)
			assert.NotNil(t, cfg)

			// Check environment-specific rate limits
			assert.Equal(t, tt.expectedGeneralRequests, cfg.RateLimit.General.Requests)
			assert.Equal(t, tt.expectedContactRequests, cfg.RateLimit.Contact.Requests)
			assert.Equal(t, 15*time.Minute, cfg.RateLimit.General.Window)
			assert.Equal(t, 1*time.Hour, cfg.RateLimit.Contact.Window)
		})
	}
}
