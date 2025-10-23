// Package middleware provides HTTP middleware for the MarkGo blog engine.
// It includes security, logging, CORS, rate limiting, and request tracking middleware.
package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Security adds basic security headers
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// Performance logs request timing and basic metrics
func Performance(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		// Log slow requests (over 1 second)
		if duration > time.Second {
			logger.Warn("Slow request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"duration", duration,
				"status", c.Writer.Status(),
			)
		}

		// Add timing header
		c.Header("X-Response-Time", duration.String())
	}
}

// CORS handles cross-origin requests with secure origin validation
func CORS(allowedOrigins []string, isDevelopment bool) gin.HandlerFunc {
	// Build a map of allowed origins for O(1) lookup
	allowedMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		allowedMap[origin] = true
	}

	// In development, add localhost variants explicitly
	if isDevelopment {
		allowedMap["http://localhost:3000"] = true
		allowedMap["http://127.0.0.1:3000"] = true
		allowedMap["http://localhost:3001"] = true
		allowedMap["http://127.0.0.1:3001"] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Only allow explicitly configured origins (exact match - no substring)
		if origin != "" && allowedMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin") // Important for caching
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimit provides sliding window rate limiting with bounded memory
func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	const maxClients = 10000 // Prevent memory exhaustion attacks

	clients := make(map[string][]time.Time)
	var mu sync.RWMutex

	// Background cleanup goroutine to prevent unbounded growth
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			now := time.Now()

			// Clean up old entries and empty slices
			for ip, times := range clients {
				var validTimes []time.Time
				for _, t := range times {
					if now.Sub(t) <= window {
						validTimes = append(validTimes, t)
					}
				}

				if len(validTimes) == 0 {
					delete(clients, ip)
				} else {
					clients[ip] = validTimes
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		// Use RemoteAddr for security (ClientIP can be spoofed via X-Forwarded-For)
		ip := c.Request.RemoteAddr
		// Strip port from RemoteAddr (format is "IP:port")
		if idx := len(ip) - 1; idx >= 0 {
			for i := idx; i >= 0; i-- {
				if ip[i] == ':' {
					ip = ip[:i]
					break
				}
			}
		}

		now := time.Now()

		mu.Lock()
		defer mu.Unlock()

		// Prevent memory exhaustion: reject if too many unique IPs
		if len(clients) >= maxClients && clients[ip] == nil {
			c.Header("Retry-After", "3600")
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		// Clean old entries for this IP
		if times, exists := clients[ip]; exists {
			var validTimes []time.Time
			for _, t := range times {
				if now.Sub(t) <= window {
					validTimes = append(validTimes, t)
				}
			}
			clients[ip] = validTimes
		}

		// Check rate limit
		if len(clients[ip]) >= requests {
			retryAfter := int(window.Seconds())
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		// Add current request
		if clients[ip] == nil {
			clients[ip] = make([]time.Time, 0, requests)
		}
		clients[ip] = append(clients[ip], now)
		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// Logger provides basic request logging
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"duration", param.Latency,
		)
		return ""
	})
}

// SmartCacheHeaders adds basic cache headers
func SmartCacheHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=3600")
		c.Next()
	}
}

// RequestTracker adds request tracking
func RequestTracker() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// BasicAuth provides basic HTTP authentication
func BasicAuth(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}

// NoCache adds no-cache headers
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// RecoveryWithErrorHandler provides recovery with error handling
func RecoveryWithErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.Error("Panic recovered", "error", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	})
}

// ErrorHandler provides centralized error handling
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			logger.Error("Request error", "errors", c.Errors.String())
		}
	}
}
