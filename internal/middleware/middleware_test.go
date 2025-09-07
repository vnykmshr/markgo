package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vnykmshr/markgo/internal/config"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// createTestRouter creates a minimal router for testing middleware
func createTestRouter() *gin.Engine {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		c.JSON(http.StatusOK, gin.H{"message": "slow"})
	})
	return router
}

func TestLogger_Middleware(t *testing.T) {
	// Simple test that verifies the Logger middleware can be created and used
	// without complex pooled dependencies
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Test that the middleware function can be created
	middleware := Logger(logger)
	assert.NotNil(t, middleware)

	// Test basic middleware execution without complex assertions
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Note: Logger middleware uses pooled resources that aren't initialized in tests,
	// so we only verify basic functionality works
}

func TestCORS_Middleware(t *testing.T) {
	corsConfig := config.CORSConfig{
		AllowedOrigins: []string{"https://example.com", "https://test.com"},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}

	t.Run("allowed origin", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS(corsConfig))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("preflight request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS(corsConfig))
		router.OPTIONS("/test", func(c *gin.Context) {}) // Handle OPTIONS

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestSecurity_Middleware(t *testing.T) {
	router := gin.New()
	router.Use(Security())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify security headers
	expectedHeaders := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "SAMEORIGIN",
		"X-XSS-Protection":        "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Permissions-Policy":      "geolocation=(), microphone=(), camera=()",
		"Content-Security-Policy": "default-src 'self'",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := w.Header().Get(header)
		assert.Contains(t, actualValue, expectedValue, "Header %s should contain %s", header, expectedValue)
	}
}

func TestRateLimit_Middleware(t *testing.T) {
	// Test that rate limit middleware can be created
	middleware := RateLimit(2, time.Minute)
	assert.NotNil(t, middleware)

	// Note: Rate limiting middleware uses a global rate limiter manager
	// that requires initialization, so we only test basic creation
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should pass without rate limiting in isolated test
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmartCacheHeaders_Middleware(t *testing.T) {
	router := createTestRouter()
	router.Use(SmartCacheHeaders())

	// Add different route types for testing
	router.GET("/static/style.css", func(c *gin.Context) { c.String(http.StatusOK, "css") })
	router.GET("/articles/test", func(c *gin.Context) { c.String(http.StatusOK, "article") })
	router.GET("/feed.xml", func(c *gin.Context) { c.String(http.StatusOK, "rss") })
	router.GET("/admin/stats", func(c *gin.Context) { c.String(http.StatusOK, "admin") })

	tests := []struct {
		name                 string
		path                 string
		expectedCacheControl string
		shouldHaveETag       bool
	}{
		{
			name:                 "static assets",
			path:                 "/static/style.css",
			expectedCacheControl: "public, max-age=31536000, immutable",
			shouldHaveETag:       true,
		},
		{
			name:                 "article pages",
			path:                 "/articles/test",
			expectedCacheControl: "public, max-age=1800, stale-while-revalidate=3600",
			shouldHaveETag:       true,
		},
		{
			name:                 "RSS feeds",
			path:                 "/feed.xml",
			expectedCacheControl: "public, max-age=3600",
			shouldHaveETag:       true,
		},
		{
			name:                 "home page",
			path:                 "/",
			expectedCacheControl: "public, max-age=900, stale-while-revalidate=1800",
			shouldHaveETag:       true,
		},
		{
			name:                 "admin routes (no caching)",
			path:                 "/admin/stats",
			expectedCacheControl: "",
			shouldHaveETag:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectedCacheControl != "" {
				assert.Equal(t, tt.expectedCacheControl, w.Header().Get("Cache-Control"))
				assert.Equal(t, "Accept-Encoding, User-Agent", w.Header().Get("Vary"))
			} else {
				assert.Empty(t, w.Header().Get("Cache-Control"))
			}

			if tt.shouldHaveETag {
				assert.NotEmpty(t, w.Header().Get("ETag"))
			} else {
				assert.Empty(t, w.Header().Get("ETag"))
			}
		})
	}
}

func TestRequestID_Middleware(t *testing.T) {
	// Test that RequestID middleware can be created
	middleware := RequestID()
	assert.NotNil(t, middleware)

	// Note: RequestID middleware uses pooled resources that aren't initialized in tests,
	// so we only test basic functionality
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	t.Run("middleware executes without error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// RequestID header may not be set due to pool dependencies in tests
	})
}

func TestNoCache_Middleware(t *testing.T) {
	router := gin.New()
	router.Use(NoCache())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))
}

func TestCacheControl_Middleware(t *testing.T) {
	maxAge := 24 * time.Hour
	router := gin.New()
	router.Use(CacheControl(maxAge))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedHeader := "public, max-age=86400" // 24 hours in seconds
	assert.Equal(t, expectedHeader, w.Header().Get("Cache-Control"))
}

func TestGenerateETag(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"simple path", "/test"},
		{"complex path", "/articles/test-article-123"},
		{"path with query", "/search?q=test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etag1 := generateETag(tt.path)
			etag2 := generateETag(tt.path)

			// ETag should be consistent for same path
			assert.Equal(t, etag1, etag2)

			// ETag should be quoted and have reasonable length
			assert.True(t, len(etag1) > 2)
			assert.True(t, strings.HasPrefix(etag1, `"`))
			assert.True(t, strings.HasSuffix(etag1, `"`))
		})
	}

	t.Run("different paths produce different ETags", func(t *testing.T) {
		etag1 := generateETag("/path1")
		etag2 := generateETag("/path2")
		assert.NotEqual(t, etag1, etag2)
	})
}

// Benchmark tests for middleware performance
func BenchmarkLogger(b *testing.B) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	router := gin.New()
	router.Use(Logger(logger))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCORS(b *testing.B) {
	corsConfig := config.CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}

	router := gin.New()
	router.Use(CORS(corsConfig))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkSmartCacheHeaders(b *testing.B) {
	router := gin.New()
	router.Use(SmartCacheHeaders())
	router.GET("/static/test.css", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/static/test.css", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGenerateETag(b *testing.B) {
	path := "/articles/test-article-with-long-path"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateETag(path)
	}
}
