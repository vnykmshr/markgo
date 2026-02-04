package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	router := gin.New()
	return router
}

// TestCORS tests the CORS middleware security
func TestCORS(t *testing.T) {
	t.Run("exact origin match - allowed", func(t *testing.T) {
		router := setupTestRouter()
		allowedOrigins := []string{"https://example.com", "https://api.example.com"}
		router.Use(CORS(allowedOrigins, false))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})

	t.Run("exact origin match - not allowed", func(t *testing.T) {
		router := setupTestRouter()
		allowedOrigins := []string{"https://example.com"}
		router.Use(CORS(allowedOrigins, false))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// Should NOT set Access-Control-Allow-Origin for disallowed origin
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("localhost bypass prevented - evil domain", func(t *testing.T) {
		router := setupTestRouter()
		allowedOrigins := []string{"https://example.com"}
		router.Use(CORS(allowedOrigins, false))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		// Try to bypass with localhost.evil.com (should be rejected)
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Origin", "http://localhost.evil.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// Should NOT allow localhost.evil.com even though it contains "localhost"
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("development mode - localhost allowed", func(t *testing.T) {
		router := setupTestRouter()
		allowedOrigins := []string{"https://example.com"}
		router.Use(CORS(allowedOrigins, true)) // Development mode
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		router := setupTestRouter()
		allowedOrigins := []string{"https://example.com"}
		router.Use(CORS(allowedOrigins, false))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("no origin header - no CORS headers", func(t *testing.T) {
		router := setupTestRouter()
		allowedOrigins := []string{"https://example.com"}
		router.Use(CORS(allowedOrigins, false))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		// No Origin header
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// Should not set CORS headers when no Origin header present
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

// TestRateLimit tests the rate limiting middleware
func TestRateLimit(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(RateLimit(5, 1*time.Minute))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		// Make 5 requests (within limit)
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code, "Request %d should succeed", i+1)
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(RateLimit(3, 1*time.Minute))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		// Make 3 requests (at limit)
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			req.RemoteAddr = "192.168.1.2:12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code, "Request %d should succeed", i+1)
		}

		// 4th request should be blocked
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.NotEmpty(t, w.Header().Get("Retry-After"))
	})

	t.Run("different IPs tracked separately", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(RateLimit(2, 1*time.Minute))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		// IP1: 2 requests (at limit)
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			req.RemoteAddr = "192.168.1.3:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}

		// IP2: 2 requests (should also succeed - different IP)
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			req.RemoteAddr = "192.168.1.4:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}
	})

	t.Run("strips port from RemoteAddr", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(RateLimit(2, 1*time.Minute))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "ok")
		})

		// Same IP with different ports should be treated as same client
		req1 := httptest.NewRequest("GET", "/test", http.NoBody)
		req1.RemoteAddr = "192.168.1.5:11111"
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 200, w1.Code)

		req2 := httptest.NewRequest("GET", "/test", http.NoBody)
		req2.RemoteAddr = "192.168.1.5:22222" // Different port, same IP
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, 200, w2.Code)

		// 3rd request should be blocked (same IP, at limit)
		req3 := httptest.NewRequest("GET", "/test", http.NoBody)
		req3.RemoteAddr = "192.168.1.5:33333"
		w3 := httptest.NewRecorder()
		router.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	})
}

// TestSecurity tests the security headers middleware
func TestSecurity(t *testing.T) {
	router := setupTestRouter()
	router.Use(Security())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

// TestLogger tests the logger middleware
func TestLogger(t *testing.T) {
	router := setupTestRouter()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router.Use(Logger(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

// TestNoCache tests the NoCache middleware
func TestNoCache(t *testing.T) {
	router := setupTestRouter()
	router.Use(NoCache())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))
}

// TestSmartCacheHeaders tests the smart cache headers middleware
func TestSmartCacheHeaders(t *testing.T) {
	router := setupTestRouter()
	router.Use(SmartCacheHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "public, max-age=3600", w.Header().Get("Cache-Control"))
}
