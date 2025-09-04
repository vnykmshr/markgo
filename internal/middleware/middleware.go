package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/markgo/internal/config"
)

// Logger middleware for structured logging
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build log entry
		logEntry := logger.With(
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		if raw != "" {
			logEntry = logEntry.With("query", raw)
		}

		// Log based on status code
		switch {
		case c.Writer.Status() >= 500:
			logEntry.Error("Server error")
		case c.Writer.Status() >= 400:
			logEntry.Warn("Client error")
		case c.Writer.Status() >= 300:
			logEntry.Info("Redirect")
		default:
			logEntry.Info("Request completed")
		}
	}
}

// CORS middleware for cross-origin requests
func CORS(config config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := "*"
		for _, allowed := range config.AllowedOrigins {
			if allowed == origin || allowed == "*" {
				allowedOrigin = origin
				break
			}
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Security middleware for security headers
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy
		csp := "default-src 'self'; " +
			"style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://fonts.googleapis.com https://giscus.app; " +
			"script-src 'self' https://cdnjs.cloudflare.com https://giscus.app; " +
			"img-src 'self' data: https: https://github.com https://avatars.githubusercontent.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"connect-src 'self' https://giscus.app https://api.github.com; " +
			"frame-src 'self' https://giscus.app; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}

// Rate limiter implementation
type rateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

func (rl *rateLimiter) isAllowed(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this key
	requests := rl.requests[key]

	// Filter out old requests
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit exceeded
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for key, requests := range rl.requests {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if reqTime.After(windowStart) {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
		rl.mutex.Unlock()
	}
}

// Global rate limiters
var (
	generalLimiter *rateLimiter
	contactLimiter *rateLimiter
	limiterMutex   sync.Mutex
)

// RateLimit middleware for general rate limiting
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiterMutex.Lock()
	if generalLimiter == nil {
		generalLimiter = newRateLimiter(limit, window)
	}
	limiterMutex.Unlock()

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !generalLimiter.isAllowed(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Limit: %d requests per %v", limit, window),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ContactRateLimit middleware for contact form rate limiting
func ContactRateLimit() gin.HandlerFunc {
	limiterMutex.Lock()
	if contactLimiter == nil {
		contactLimiter = newRateLimiter(5, time.Hour) // 5 requests per hour
	}
	limiterMutex.Unlock()

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !contactLimiter.isAllowed(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Contact form rate limit exceeded",
				"message": "Too many contact form submissions. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// BasicAuth middleware for admin authentication
func BasicAuth(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}

// NoCache middleware to prevent caching
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// CacheControl middleware for static assets
func CacheControl(maxAge time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
		c.Next()
	}
}

// Compress middleware for response compression
func Compress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if client accepts gzip
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Skip compression for certain content types
		contentType := c.Writer.Header().Get("Content-Type")
		if strings.Contains(contentType, "image/") ||
			strings.Contains(contentType, "video/") ||
			strings.Contains(contentType, "application/zip") {
			c.Next()
			return
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Next()
	}
}

// Timeout middleware for request timeout
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := c.Request.Context(), func() {}
		if timeout > 0 {
			ctx, cancel = c.Request.Context(), func() {}
		}
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Channel to signal completion
		finished := make(chan struct{})

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// Request completed normally
		case <-ctx.Done():
			// Request timed out
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request timeout",
				"message": "Request took too long to process",
			})
			c.Abort()
		}
	}
}

// RequestID middleware to add unique request ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("RequestID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	// Generate 8 random bytes and convert to hex for uniqueness
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-only ID if random generation fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(bytes))
}
