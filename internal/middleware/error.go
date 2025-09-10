package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
)

// ErrorHandler provides centralized error handling for HTTP requests
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only handle if there are errors
		if len(c.Errors) == 0 {
			return
		}

		// Get the last error (most recent)
		err := c.Errors.Last().Err

		// Extract request context for logging
		requestID := c.GetString("RequestID")
		if requestID == "" {
			requestID = "unknown"
		}

		method := c.Request.Method
		path := c.Request.URL.Path
		userAgent := c.Request.UserAgent()
		ip := c.ClientIP()

		// Log the error with context
		logger.Error("Request error",
			"error", err,
			"request_id", requestID,
			"method", method,
			"path", path,
			"user_agent", userAgent,
			"client_ip", ip,
		)

		// Determine status code and user message based on error type
		statusCode, userMessage := classifyError(err)

		// Check if response was already written
		if c.Writer.Written() {
			return
		}

		// Set appropriate headers
		if statusCode == http.StatusTooManyRequests {
			// Only set Retry-After if it's not already set by the middleware
			if c.Writer.Header().Get("Retry-After") == "" {
				c.Header("Retry-After", "60") // Default retry after 1 minute
			}
		}

		// Handle different response formats based on Accept header
		accept := c.GetHeader("Accept")
		if shouldReturnJSON(accept, c.Request.URL.Path) {
			// JSON response for API endpoints
			c.JSON(statusCode, gin.H{
				"error":      userMessage,
				"request_id": requestID,
				"timestamp":  time.Now().Unix(),
			})
		} else {
			// HTML response for web pages
			data := map[string]any{
				"error":       getErrorTitle(statusCode),
				"message":     userMessage,
				"request_id":  requestID,
				"status_code": statusCode,
			}

			// Use appropriate template based on status code
			template := getErrorTemplate(statusCode)
			c.HTML(statusCode, template, data)
		}
	}
}

// classifyError determines the appropriate HTTP status code and user message for an error
func classifyError(err error) (statusCode int, userMessage string) {
	switch {
	case apperrors.IsNotFound(err):
		return http.StatusNotFound, apperrors.GetUserFriendlyMessage(err)

	case apperrors.IsValidationError(err):
		return http.StatusBadRequest, apperrors.GetUserFriendlyMessage(err)

	case apperrors.IsConfigurationError(err):
		return http.StatusServiceUnavailable, apperrors.GetUserFriendlyMessage(err)

	// Check for specific error types
	case errors.Is(err, apperrors.ErrEmailNotConfigured):
		return http.StatusServiceUnavailable, "Contact form is temporarily unavailable"

	case errors.Is(err, apperrors.ErrSMTPAuthFailed):
		return http.StatusServiceUnavailable, "Email service is temporarily unavailable"

	case errors.Is(err, apperrors.ErrSearchTimeout):
		return http.StatusRequestTimeout, apperrors.GetUserFriendlyMessage(err)

	// Check for HTTP errors with specific status codes
	case isHTTPError(err):
		if httpErr := getHTTPError(err); httpErr != nil {
			return httpErr.StatusCode, httpErr.Message
		}
		fallthrough

	// Check for permission/authorization errors
	case isPermissionError(err):
		return http.StatusForbidden, "Access denied"

	// Check for rate limiting (would be set by rate limiting middleware)
	case isRateLimitError(err):
		return http.StatusTooManyRequests, "Too many requests. Please wait before trying again"

	// Default to internal server error
	default:
		return http.StatusInternalServerError, "An unexpected error occurred. Please try again later"
	}
}

// shouldReturnJSON determines if the response should be JSON based on request characteristics
func shouldReturnJSON(acceptHeader, path string) bool {
	// API endpoints should return JSON
	if isAPIPath(path) {
		return true
	}

	// Check Accept header for JSON preference
	return acceptHeader == "application/json" ||
		acceptHeader == "application/json, */*" ||
		acceptHeader == "*/*"
}

// isAPIPath checks if the path is an API endpoint
func isAPIPath(path string) bool {
	// Common API path patterns
	return path == "/health" ||
		path == "/metrics" ||
		path == "/admin/stats" ||
		path == "/admin/clear-cache" ||
		path == "/admin/reload" ||
		path == "/contact" // Contact form is typically JSON
}

// getErrorTemplate returns the appropriate template for the status code
func getErrorTemplate(statusCode int) string {
	switch statusCode {
	case http.StatusNotFound:
		return "404.html"
	case http.StatusInternalServerError:
		return "500.html"
	case http.StatusServiceUnavailable:
		return "503.html"
	default:
		return "error.html"
	}
}

// getErrorTitle returns a user-friendly title for the error status
func getErrorTitle(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Invalid Request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Access Denied"
	case http.StatusNotFound:
		return "Page Not Found"
	case http.StatusMethodNotAllowed:
		return "Method Not Allowed"
	case http.StatusRequestTimeout:
		return "Request Timeout"
	case http.StatusTooManyRequests:
		return "Too Many Requests"
	case http.StatusInternalServerError:
		return "Internal Server Error"
	case http.StatusServiceUnavailable:
		return "Service Unavailable"
	default:
		return "Error"
	}
}

// Helper functions for error type checking

func isHTTPError(err error) bool {
	var httpErr *apperrors.HTTPError
	return err != nil && errors.As(err, &httpErr)
}

func getHTTPError(err error) *apperrors.HTTPError {
	var httpErr *apperrors.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}
	return nil
}

func isPermissionError(err error) bool {
	return errors.Is(err, apperrors.ErrFilePermission)
}

func isRateLimitError(err error) bool {
	// This would be set by rate limiting middleware
	// For now, we check if it's a custom rate limit error
	return false // Placeholder - would be implemented with actual rate limiting
}

// RecoveryWithErrorHandler provides panic recovery with integrated error handling
func RecoveryWithErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.Error("Panic recovered",
					"error", err,
					"request_id", c.GetString("RequestID"),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
				)

				// Add error to context so ErrorHandler can process it
				_ = c.Error(apperrors.NewHTTPError(
					http.StatusInternalServerError,
					"Internal server error",
					nil,
				))

				// Don't call c.Next() after panic
				c.Abort()
			}
		}()

		c.Next()
	}
}
