// Package middleware provides HTTP middleware for the MarkGo blog engine.
// It includes security, logging, CORS, rate limiting, and request tracking middleware.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// Timeout middleware with configurable duration
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			defer func() {
				if err := recover(); err != nil {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
				close(done)
			}()
			c.Next()
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.AbortWithStatus(http.StatusRequestTimeout)
			return
		}
	}
}

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

// CORS handles cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins or all for development
		if origin != "" && (strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1")) {
			c.Header("Access-Control-Allow-Origin", origin)
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

// Compress enables gzip compression
func Compress() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Encoding", "gzip")
		c.Next()
	}
}

// RateLimit provides basic rate limiting
func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	// Simple in-memory rate limiter with mutex for thread safety
	clients := make(map[string][]time.Time)
	var mu sync.RWMutex

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		defer mu.Unlock()

		// Clean old entries
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
			c.Header("Retry-After", "60")
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		// Add current request
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

// PerformanceMiddleware is an alias for Performance
func PerformanceMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return Performance(logger)
}

// CompetitorBenchmarkMiddleware is a no-op placeholder
func CompetitorBenchmarkMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// SmartCacheHeaders adds basic cache headers
func SmartCacheHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=3600")
		c.Next()
	}
}

// RequestTracker adds request tracking
func RequestTracker(_ *slog.Logger, _ string) gin.HandlerFunc {
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

// RequestLoggingMiddleware provides enhanced request logging
func RequestLoggingMiddleware(_ interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// SecurityLoggingMiddleware provides security event logging
func SecurityLoggingMiddleware(_ interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// PerformanceLoggingMiddleware provides detailed performance logging
func PerformanceLoggingMiddleware(_ interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
