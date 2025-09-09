package middleware

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/markgo/internal/services"
)

const (
	RequestIDKey     = "request_id"
	RequestStartTime = "request_start_time"
)

// Define custom types for context keys to avoid collisions
type contextKey string

const (
	loggerContextKey = contextKey("logger")
)

// RequestLoggingMiddleware provides enhanced request logging with structured data
func RequestLoggingMiddleware(loggingService services.LoggingServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate or extract request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Store request ID in context
		c.Set(RequestIDKey, requestID)
		c.Set(RequestStartTime, start)

		// Create request context logger
		entry := services.LogEntry{
			RequestID: requestID,
			IP:        getClientIP(c),
			UserAgent: c.GetHeader("User-Agent"),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
		}

		// Create context with logger
		ctx := context.WithValue(c.Request.Context(), loggerContextKey, loggingService.WithRequestContext(c.Request.Context(), entry))
		c.Request = c.Request.WithContext(ctx)

		// Log request start (debug level)
		if loggingService.GetLogger().Enabled(c.Request.Context(), slog.LevelDebug) {
			loggingService.WithRequestContext(c.Request.Context(), entry).Debug("Request started",
				slog.String("query", c.Request.URL.RawQuery),
				slog.String("referer", c.GetHeader("Referer")),
				slog.String("content_type", c.ContentType()),
			)
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		entry.Duration = duration
		entry.StatusCode = c.Writer.Status()

		// Set request ID header only for successful responses to avoid conflicts with auth
		if c.Writer.Status() < 400 {
			c.Header("X-Request-ID", requestID)
		}

		// Log request completion
		loggingService.LogHTTPRequest(c.Request.Context(), entry)

		// Log slow requests (configurable threshold)
		slowThreshold := 1 * time.Second
		loggingService.LogSlowOperation(c.Request.Context(), "http_request", duration, slowThreshold,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"status_code", c.Writer.Status(),
		)

		// Log security events for suspicious activities
		logSecurityEvents(loggingService, c, entry)
	}
}

// SecurityLoggingMiddleware logs security-related events
func SecurityLoggingMiddleware(loggingService services.LoggingServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		entry := services.LogEntry{
			IP:        getClientIP(c),
			UserAgent: c.GetHeader("User-Agent"),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
		}

		// Check for common attack patterns
		checkForSecurityThreats(loggingService, c, entry)

		c.Next()
	}
}

// PerformanceLoggingMiddleware logs detailed performance metrics
func PerformanceLoggingMiddleware(loggingService services.LoggingServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		memBefore, _, allocsBefore := loggingService.GetMemoryStats()

		c.Next()

		duration := time.Since(start)
		memAfter, _, allocsAfter := loggingService.GetMemoryStats()

		// Log performance metrics for slow or memory-intensive requests
		if duration > 500*time.Millisecond || (memAfter-memBefore) > 10*1024*1024 { // 500ms or 10MB
			perfLog := services.PerformanceLog{
				Operation:    c.Request.Method + " " + c.Request.URL.Path,
				Duration:     duration,
				MemoryBefore: memBefore,
				MemoryAfter:  memAfter,
				Allocations:  allocsAfter - allocsBefore,
			}

			loggingService.LogPerformance(perfLog)
		}
	}
}

// getClientIP extracts the real client IP from headers
func getClientIP(c *gin.Context) string {
	// Check X-Real-IP header
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// Check X-Forwarded-For header
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		if idx := strings.Index(ip, ","); idx > 0 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}

// logSecurityEvents logs potential security threats
func logSecurityEvents(loggingService services.LoggingServiceInterface, c *gin.Context, entry services.LogEntry) {
	// Log failed authentication attempts
	if entry.StatusCode == 401 {
		loggingService.LogSecurity(services.SecurityLog{
			Event:       "authentication_failed",
			Severity:    "medium",
			IP:          entry.IP,
			UserAgent:   entry.UserAgent,
			Path:        entry.Path,
			Description: "Authentication failed for request",
		})
	}

	// Log potential brute force attempts (multiple 401s from same IP would be detected by rate limiting)
	if entry.StatusCode == 403 {
		loggingService.LogSecurity(services.SecurityLog{
			Event:       "access_denied",
			Severity:    "medium",
			IP:          entry.IP,
			UserAgent:   entry.UserAgent,
			Path:        entry.Path,
			Description: "Access denied to protected resource",
		})
	}

	// Log server errors that might indicate attacks
	if entry.StatusCode >= 500 {
		loggingService.LogSecurity(services.SecurityLog{
			Event:       "server_error",
			Severity:    "high",
			IP:          entry.IP,
			UserAgent:   entry.UserAgent,
			Path:        entry.Path,
			Description: "Server error occurred - potential attack or system issue",
		})
	}
}

// checkForSecurityThreats analyzes requests for common attack patterns
func checkForSecurityThreats(loggingService services.LoggingServiceInterface, c *gin.Context, entry services.LogEntry) {
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery

	// Check for SQL injection patterns
	sqlPatterns := []string{
		"union+select", "union%20select", "union select",
		"'; drop table", "%27; drop table",
		"' or '1'='1", "%27 or %271%27=%271",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(strings.ToLower(query), pattern) || strings.Contains(strings.ToLower(path), pattern) {
			loggingService.LogSecurity(services.SecurityLog{
				Event:       "sql_injection_attempt",
				Severity:    "high",
				IP:          entry.IP,
				UserAgent:   entry.UserAgent,
				Path:        entry.Path,
				Description: "Potential SQL injection attempt detected in request",
			})
			break
		}
	}

	// Check for XSS patterns
	xssPatterns := []string{
		"<script", "%3cscript",
		"javascript:", "javascript%3a",
		"onload=", "onload%3d",
		"onerror=", "onerror%3d",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(strings.ToLower(query), pattern) || strings.Contains(strings.ToLower(path), pattern) {
			loggingService.LogSecurity(services.SecurityLog{
				Event:       "xss_attempt",
				Severity:    "high",
				IP:          entry.IP,
				UserAgent:   entry.UserAgent,
				Path:        entry.Path,
				Description: "Potential XSS attempt detected in request",
			})
			break
		}
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") || strings.Contains(path, "%2e%2e") {
		loggingService.LogSecurity(services.SecurityLog{
			Event:       "path_traversal_attempt",
			Severity:    "high",
			IP:          entry.IP,
			UserAgent:   entry.UserAgent,
			Path:        entry.Path,
			Description: "Potential path traversal attempt detected",
		})
	}

	// Check for suspicious User-Agent strings
	suspiciousUAs := []string{
		"sqlmap", "nikto", "nmap", "masscan", "zap", "burp",
		"havij", "acunetix", "netsparker", "appscan",
	}

	userAgent := strings.ToLower(entry.UserAgent)
	for _, ua := range suspiciousUAs {
		if strings.Contains(userAgent, ua) {
			loggingService.LogSecurity(services.SecurityLog{
				Event:       "suspicious_user_agent",
				Severity:    "medium",
				IP:          entry.IP,
				UserAgent:   entry.UserAgent,
				Path:        entry.Path,
				Description: "Suspicious user agent detected - potential scanning tool",
			})
			break
		}
	}

	// Check for excessive request size (potential DoS)
	if c.Request.ContentLength > 10*1024*1024 { // 10MB
		loggingService.LogSecurity(services.SecurityLog{
			Event:       "large_request",
			Severity:    "medium",
			IP:          entry.IP,
			UserAgent:   entry.UserAgent,
			Path:        entry.Path,
			Description: "Large request body detected - potential DoS attempt",
		})
	}
}

// GetRequestLogger extracts the request-scoped logger from context
func GetRequestLogger(c *gin.Context) *slog.Logger {
	if logger, exists := c.Request.Context().Value(loggerContextKey).(*slog.Logger); exists {
		return logger
	}
	return slog.Default() // Fallback to default logger
}

// GetRequestID extracts the request ID from the Gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetRequestDuration calculates the duration since request start
func GetRequestDuration(c *gin.Context) time.Duration {
	if startTime, exists := c.Get(RequestStartTime); exists {
		if start, ok := startTime.(time.Time); ok {
			return time.Since(start)
		}
	}
	return 0
}
