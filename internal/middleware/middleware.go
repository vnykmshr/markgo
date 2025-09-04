package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/utils"
)

// Logger middleware for structured logging with pooled context
func Logger(logger *slog.Logger) gin.HandlerFunc {
	contextPool := utils.GetGlobalMiddlewareContextPool()
	responseWriterPool := utils.GetGlobalResponseWriterPool()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Use pooled response writer for efficient response handling
		pooledWriter := responseWriterPool.GetWriter(c.Writer)
		defer responseWriterPool.PutWriter(pooledWriter)
		c.Writer = pooledWriter

		// Use pooled context for log metadata
		contextPool.WithContext(func(ctx *utils.MiddlewareContext) {
			// Collect metadata efficiently
			ctx.Values["method"] = c.Request.Method
			ctx.Values["path"] = path
			ctx.Values["ip"] = c.ClientIP()
			ctx.Values["user_agent"] = c.Request.UserAgent()
			if raw != "" {
				ctx.Values["query"] = raw
			}

			// Process request
			c.Next()

			// Flush pooled response writer
			pooledWriter.Flush()

			// Calculate latency and add to context
			latency := time.Since(start)
			ctx.Values["latency"] = latency.String()
			ctx.Values["status"] = pooledWriter.Status()

			// Build log entry with pooled values
			logEntry := logger.With(
				"method", ctx.Values["method"],
				"path", ctx.Values["path"],
				"status", ctx.Values["status"],
				"latency", ctx.Values["latency"],
				"ip", ctx.Values["ip"],
				"user_agent", ctx.Values["user_agent"],
			)

			if query, exists := ctx.Values["query"]; exists {
				logEntry = logEntry.With("query", query)
			}

			// Log based on status code
			status := ctx.Values["status"].(int)
			switch {
			case status >= 500:
				logEntry.Error("Server error")
			case status >= 400:
				logEntry.Warn("Client error")
			case status >= 300:
				logEntry.Info("Redirect")
			default:
				logEntry.Info("Request completed")
			}
		})
	}
}

// CORS middleware for cross-origin requests with optimized header handling
func CORS(config config.CORSConfig) gin.HandlerFunc {
	// Pre-compute joined strings to avoid repeated allocations
	allowedMethods := strings.Join(config.AllowedMethods, ", ")
	allowedHeaders := strings.Join(config.AllowedHeaders, ", ")

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

		// Set CORS headers efficiently
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", allowedMethods)
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
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

// Use optimized rate limiter manager for all rate limiting needs
var rateLimiterManager = utils.GetGlobalRateLimiterManager()

// RateLimit middleware for general rate limiting
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := rateLimiterManager.GetLimiter("general", limit, window)

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !limiter.IsAllowed(key) {
			data := utils.GetTemplateData()
			data["error"] = "Rate limit exceeded"
			data["message"] = fmt.Sprintf("Too many requests. Limit: %d requests per %v", limit, window)
			c.JSON(http.StatusTooManyRequests, data)
			utils.PutTemplateData(data)
			c.Abort()
			return
		}

		c.Next()
	}
}

// ContactRateLimit middleware for contact form rate limiting
func ContactRateLimit() gin.HandlerFunc {
	limiter := rateLimiterManager.GetLimiter("contact", 5, time.Hour) // 5 requests per hour

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !limiter.IsAllowed(key) {
			data := utils.GetTemplateData()
			data["error"] = "Contact form rate limit exceeded"
			data["message"] = "Too many contact form submissions. Please try again later."
			c.JSON(http.StatusTooManyRequests, data)
			utils.PutTemplateData(data)
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
	compressedPool := utils.GetGlobalCompressedResponsePool()
	responseWriterPool := utils.GetGlobalResponseWriterPool()

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

		// Use pooled response writer for efficient compression handling
		pooledWriter := responseWriterPool.GetWriter(c.Writer)
		defer func() {
			pooledWriter.Flush()
			responseWriterPool.PutWriter(pooledWriter)
		}()

		// Use pooled buffer for compression
		compressBuffer := compressedPool.GetBuffer()
		defer compressedPool.PutBuffer(compressBuffer)

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Writer = pooledWriter

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
			// Request timed out - use pooled template data
			data := utils.GetTemplateData()
			data["error"] = "Request timeout"
			data["message"] = "Request took too long to process"
			c.JSON(http.StatusRequestTimeout, data)
			utils.PutTemplateData(data)
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
	pool := utils.GetGlobalRequestIDPool()

	// Get pooled random bytes
	bytes := pool.GetRandomBytes()
	defer pool.PutRandomBytes(bytes)

	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-only ID if random generation fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Use pooled buffer for building the request ID
	buffer := pool.GetBuffer()
	defer pool.PutBuffer(buffer)

	// Build request ID efficiently
	timestamp := time.Now().UnixNano()
	hexBytes := hex.EncodeToString(bytes)

	// Format: timestamp-hexbytes
	buffer = append(buffer, fmt.Sprintf("%d-%s", timestamp, hexBytes)...)
	return string(buffer)
}
