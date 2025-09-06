package middleware

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
)

func TestErrorHandler(t *testing.T) {
	// Setup logger
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupError     func(c *gin.Context)
		expectedStatus int
		expectedError  string
		acceptHeader   string
		expectJSON     bool
	}{
		{
			name: "article not found error",
			setupError: func(c *gin.Context) {
				_ = c.Error(apperrors.NewArticleError("test.md", "Article not found", apperrors.ErrArticleNotFound))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "The requested article was not found",
			acceptHeader:   "application/json",
			expectJSON:     true,
		},
		{
			name: "validation error",
			setupError: func(c *gin.Context) {
				_ = c.Error(apperrors.NewValidationError("email", "invalid", "Invalid email format", apperrors.ErrValidationFailed))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Please check your input and try again",
			acceptHeader:   "application/json",
			expectJSON:     true,
		},
		{
			name: "configuration error",
			setupError: func(c *gin.Context) {
				_ = c.Error(apperrors.NewConfigError("port", -1, "Invalid port", apperrors.ErrConfigValidation))
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "Service configuration error. Please contact administrator",
			acceptHeader:   "application/json",
			expectJSON:     true,
		},
		{
			name: "email not configured",
			setupError: func(c *gin.Context) {
				_ = c.Error(apperrors.ErrEmailNotConfigured)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "Contact form is temporarily unavailable",
			acceptHeader:   "application/json",
			expectJSON:     true,
		},
		{
			name: "smtp auth failed",
			setupError: func(c *gin.Context) {
				_ = c.Error(apperrors.ErrSMTPAuthFailed)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "Email service is temporarily unavailable",
			acceptHeader:   "application/json",
			expectJSON:     true,
		},
		{
			name: "search timeout",
			setupError: func(c *gin.Context) {
				_ = c.Error(apperrors.ErrSearchTimeout)
			},
			expectedStatus: http.StatusRequestTimeout,
			expectedError:  "Search took too long. Please try again with more specific terms",
			acceptHeader:   "application/json",
			expectJSON:     true,
		},
		{
			name: "HTTP error",
			setupError: func(c *gin.Context) {
				_ = c.Error(apperrors.NewHTTPError(http.StatusBadGateway, "Bad gateway", errors.New("upstream error")))
			},
			expectedStatus: http.StatusBadGateway,
			expectedError:  "Bad gateway",
			acceptHeader:   "application/json",
			expectJSON:     true,
		},
		{
			name: "unknown error defaults to 500",
			setupError: func(c *gin.Context) {
				_ = c.Error(errors.New("unknown error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "An unexpected error occurred. Please try again later",
			acceptHeader:   "application/json",
			expectJSON:     true,
		},
		// Note: HTML response test removed because it requires template setup
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			router := gin.New()
			router.Use(ErrorHandler(logger))

			router.GET("/test", func(c *gin.Context) {
				tt.setupError(c)
			})

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Accept", tt.acceptHeader)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response content
			responseBody := w.Body.String()
			if tt.expectJSON {
				// For JSON responses, check if it contains the expected error message
				if !strings.Contains(responseBody, tt.expectedError) {
					t.Errorf("Expected error message %q not found in response: %s", tt.expectedError, responseBody)
				}
				// Check content type
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got: %s", contentType)
				}
			} else {
				// For HTML responses, check if it contains the expected error message
				if !strings.Contains(responseBody, tt.expectedError) {
					t.Errorf("Expected error message %q not found in HTML response: %s", tt.expectedError, responseBody)
				}
			}

			// Check that error was logged
			logOutput := buf.String()
			if !strings.Contains(logOutput, "Request error") {
				t.Errorf("Expected error to be logged, but log output was: %s", logOutput)
			}
		})
	}
}

func TestErrorHandlerNoErrors(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should pass through normally
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// No error should be logged
	logOutput := buf.String()
	if strings.Contains(logOutput, "Request error") {
		t.Errorf("No error should be logged, but found: %s", logOutput)
	}
}

func TestRecoveryWithErrorHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler(logger)) // ErrorHandler must come first
	router.Use(RecoveryWithErrorHandler(logger))

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 500 status
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 after panic, got %d", w.Code)
	}

	// Check that panic was logged
	logOutput := buf.String()
	if !strings.Contains(logOutput, "Panic recovered") {
		t.Errorf("Expected panic to be logged, but log output was: %s", logOutput)
	}

	// Check response contains error message
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "Internal server error") {
		t.Errorf("Expected error message in response: %s", responseBody)
	}
}

func TestShouldReturnJSON(t *testing.T) {
	tests := []struct {
		name         string
		acceptHeader string
		path         string
		expectedJSON bool
	}{
		{
			name:         "explicit JSON accept header",
			acceptHeader: "application/json",
			path:         "/test",
			expectedJSON: true,
		},
		{
			name:         "JSON with wildcard",
			acceptHeader: "application/json, */*",
			path:         "/test",
			expectedJSON: true,
		},
		{
			name:         "wildcard accept",
			acceptHeader: "*/*",
			path:         "/test",
			expectedJSON: true,
		},
		{
			name:         "API path - health",
			acceptHeader: "text/html",
			path:         "/health",
			expectedJSON: true,
		},
		{
			name:         "API path - metrics",
			acceptHeader: "text/html",
			path:         "/metrics",
			expectedJSON: true,
		},
		{
			name:         "API path - admin stats",
			acceptHeader: "text/html",
			path:         "/admin/stats",
			expectedJSON: true,
		},
		{
			name:         "API path - contact",
			acceptHeader: "text/html",
			path:         "/contact",
			expectedJSON: true,
		},
		{
			name:         "HTML accept header",
			acceptHeader: "text/html",
			path:         "/articles",
			expectedJSON: false,
		},
		{
			name:         "no accept header",
			acceptHeader: "",
			path:         "/articles",
			expectedJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldReturnJSON(tt.acceptHeader, tt.path)
			if result != tt.expectedJSON {
				t.Errorf("shouldReturnJSON(%q, %q) = %v, want %v", tt.acceptHeader, tt.path, result, tt.expectedJSON)
			}
		})
	}
}

func TestGetErrorTemplate(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   string
	}{
		{
			name:       "404 error",
			statusCode: http.StatusNotFound,
			expected:   "404.html",
		},
		{
			name:       "500 error",
			statusCode: http.StatusInternalServerError,
			expected:   "500.html",
		},
		{
			name:       "503 error",
			statusCode: http.StatusServiceUnavailable,
			expected:   "503.html",
		},
		{
			name:       "other error",
			statusCode: http.StatusBadRequest,
			expected:   "error.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorTemplate(tt.statusCode)
			if result != tt.expected {
				t.Errorf("getErrorTemplate(%d) = %q, want %q", tt.statusCode, result, tt.expected)
			}
		})
	}
}

func TestGetErrorTitle(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   string
	}{
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
			expected:   "Invalid Request",
		},
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			expected:   "Unauthorized",
		},
		{
			name:       "403 Forbidden",
			statusCode: http.StatusForbidden,
			expected:   "Access Denied",
		},
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
			expected:   "Page Not Found",
		},
		{
			name:       "405 Method Not Allowed",
			statusCode: http.StatusMethodNotAllowed,
			expected:   "Method Not Allowed",
		},
		{
			name:       "408 Request Timeout",
			statusCode: http.StatusRequestTimeout,
			expected:   "Request Timeout",
		},
		{
			name:       "429 Too Many Requests",
			statusCode: http.StatusTooManyRequests,
			expected:   "Too Many Requests",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			expected:   "Internal Server Error",
		},
		{
			name:       "503 Service Unavailable",
			statusCode: http.StatusServiceUnavailable,
			expected:   "Service Unavailable",
		},
		{
			name:       "unknown status",
			statusCode: 999,
			expected:   "Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorTitle(tt.statusCode)
			if result != tt.expected {
				t.Errorf("getErrorTitle(%d) = %q, want %q", tt.statusCode, result, tt.expected)
			}
		})
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "article not found",
			err:             apperrors.ErrArticleNotFound,
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "The requested article was not found",
		},
		{
			name:            "validation failed",
			err:             apperrors.ErrValidationFailed,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Please check your input and try again",
		},
		{
			name:            "config validation",
			err:             apperrors.ErrConfigValidation,
			expectedStatus:  http.StatusServiceUnavailable,
			expectedMessage: "Service configuration error. Please contact administrator",
		},
		{
			name:            "email not configured",
			err:             apperrors.ErrEmailNotConfigured,
			expectedStatus:  http.StatusServiceUnavailable,
			expectedMessage: "Contact form is temporarily unavailable",
		},
		{
			name:            "SMTP auth failed",
			err:             apperrors.ErrSMTPAuthFailed,
			expectedStatus:  http.StatusServiceUnavailable,
			expectedMessage: "Email service is temporarily unavailable",
		},
		{
			name:            "search timeout",
			err:             apperrors.ErrSearchTimeout,
			expectedStatus:  http.StatusRequestTimeout,
			expectedMessage: "Search took too long. Please try again with more specific terms",
		},
		{
			name:            "HTTP error",
			err:             apperrors.NewHTTPError(http.StatusBadGateway, "Bad gateway", nil),
			expectedStatus:  http.StatusBadGateway,
			expectedMessage: "Bad gateway",
		},
		{
			name:            "unknown error",
			err:             errors.New("unknown error"),
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "An unexpected error occurred. Please try again later",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message := classifyError(tt.err)
			if status != tt.expectedStatus {
				t.Errorf("classifyError(%v) status = %d, want %d", tt.err, status, tt.expectedStatus)
			}
			if message != tt.expectedMessage {
				t.Errorf("classifyError(%v) message = %q, want %q", tt.err, message, tt.expectedMessage)
			}
		})
	}
}
