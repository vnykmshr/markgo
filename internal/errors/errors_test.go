package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestArticleError(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		message  string
		err      error
		expected string
	}{
		{
			name:     "with file",
			file:     "test.md",
			message:  "parse failed",
			err:      fmt.Errorf("invalid yaml"),
			expected: "article error in test.md: parse failed",
		},
		{
			name:     "without file",
			file:     "",
			message:  "general error",
			err:      fmt.Errorf("something went wrong"),
			expected: "article error: general error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewArticleError(tt.file, tt.message, tt.err)
			if err.Error() != tt.expected {
				t.Errorf("ArticleError.Error() = %q, want %q", err.Error(), tt.expected)
			}

			// Test unwrapping
			if unwrapped := err.Unwrap(); unwrapped != tt.err {
				t.Errorf("ArticleError.Unwrap() = %v, want %v", unwrapped, tt.err)
			}
		})
	}
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    interface{}
		message  string
		err      error
		expected string
	}{
		{
			name:     "with field",
			field:    "port",
			value:    -1,
			message:  "invalid port",
			err:      fmt.Errorf("negative value"),
			expected: "config error for field port: invalid port",
		},
		{
			name:     "without field",
			field:    "",
			value:    nil,
			message:  "general config error",
			err:      fmt.Errorf("config issue"),
			expected: "config error: general config error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigError(tt.field, tt.value, tt.message, tt.err)
			if err.Error() != tt.expected {
				t.Errorf("ConfigError.Error() = %q, want %q", err.Error(), tt.expected)
			}

			// Test unwrapping
			if unwrapped := err.Unwrap(); unwrapped != tt.err {
				t.Errorf("ConfigError.Unwrap() = %v, want %v", unwrapped, tt.err)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    interface{}
		message  string
		err      error
		expected string
	}{
		{
			name:     "with field",
			field:    "email",
			value:    "invalid-email",
			message:  "invalid format",
			err:      fmt.Errorf("parsing failed"),
			expected: "validation error for field email: invalid format",
		},
		{
			name:     "without field",
			field:    "",
			value:    nil,
			message:  "general validation error",
			err:      fmt.Errorf("validation issue"),
			expected: "validation error: general validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.value, tt.message, tt.err)
			if err.Error() != tt.expected {
				t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), tt.expected)
			}

			// Test unwrapping
			if unwrapped := err.Unwrap(); unwrapped != tt.err {
				t.Errorf("ValidationError.Unwrap() = %v, want %v", unwrapped, tt.err)
			}
		})
	}
}

func TestHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		err        error
		expected   string
	}{
		{
			name:       "404 error",
			statusCode: 404,
			message:    "not found",
			err:        fmt.Errorf("resource missing"),
			expected:   "HTTP 404: not found",
		},
		{
			name:       "500 error",
			statusCode: 500,
			message:    "internal error",
			err:        fmt.Errorf("database connection failed"),
			expected:   "HTTP 500: internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewHTTPError(tt.statusCode, tt.message, tt.err)
			if err.Error() != tt.expected {
				t.Errorf("HTTPError.Error() = %q, want %q", err.Error(), tt.expected)
			}

			// Test unwrapping
			if unwrapped := err.Unwrap(); unwrapped != tt.err {
				t.Errorf("HTTPError.Unwrap() = %v, want %v", unwrapped, tt.err)
			}

			// Test status code
			if err.StatusCode != tt.statusCode {
				t.Errorf("HTTPError.StatusCode = %d, want %d", err.StatusCode, tt.statusCode)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "article not found",
			err:      ErrArticleNotFound,
			expected: true,
		},
		{
			name:     "template not found",
			err:      ErrTemplateNotFound,
			expected: true,
		},
		{
			name:     "file not found",
			err:      ErrFileNotFound,
			expected: true,
		},
		{
			name:     "cache not found",
			err:      ErrCacheNotFound,
			expected: true,
		},
		{
			name:     "wrapped article not found",
			err:      fmt.Errorf("wrapper: %w", ErrArticleNotFound),
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrValidationFailed,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "validation failed",
			err:      ErrValidationFailed,
			expected: true,
		},
		{
			name:     "invalid input",
			err:      ErrInvalidInput,
			expected: true,
		},
		{
			name:     "missing field",
			err:      ErrMissingField,
			expected: true,
		},
		{
			name:     "invalid front matter",
			err:      ErrInvalidFrontMatter,
			expected: true,
		},
		{
			name:     "wrapped validation error",
			err:      fmt.Errorf("wrapper: %w", ErrValidationFailed),
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrArticleNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidationError(tt.err)
			if result != tt.expected {
				t.Errorf("IsValidationError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsConfigurationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "invalid config",
			err:      ErrInvalidConfig,
			expected: true,
		},
		{
			name:     "missing config",
			err:      ErrMissingConfig,
			expected: true,
		},
		{
			name:     "config validation",
			err:      ErrConfigValidation,
			expected: true,
		},
		{
			name:     "wrapped config error",
			err:      fmt.Errorf("wrapper: %w", ErrInvalidConfig),
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrArticleNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConfigurationError(tt.err)
			if result != tt.expected {
				t.Errorf("IsConfigurationError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestGetUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "article not found",
			err:      ErrArticleNotFound,
			expected: "The requested article was not found",
		},
		{
			name:     "template not found",
			err:      ErrTemplateNotFound,
			expected: "Page template not found",
		},
		{
			name:     "email not configured",
			err:      ErrEmailNotConfigured,
			expected: "Email service is currently unavailable",
		},
		{
			name:     "smtp auth failed",
			err:      ErrSMTPAuthFailed,
			expected: "Email service authentication failed",
		},
		{
			name:     "validation failed",
			err:      ErrValidationFailed,
			expected: "Please check your input and try again",
		},
		{
			name:     "invalid query",
			err:      ErrInvalidQuery,
			expected: "Invalid search query. Please try different search terms",
		},
		{
			name:     "search timeout",
			err:      ErrSearchTimeout,
			expected: "Search took too long. Please try again with more specific terms",
		},
		{
			name:     "configuration error",
			err:      ErrInvalidConfig,
			expected: "Service configuration error. Please contact administrator",
		},
		{
			name:     "file not found",
			err:      ErrFileNotFound,
			expected: "The requested resource was not found",
		},
		{
			name:     "invalid input",
			err:      ErrInvalidInput,
			expected: "Invalid input provided. Please check your data and try again",
		},
		{
			name:     "unknown error",
			err:      fmt.Errorf("unknown error"),
			expected: "An unexpected error occurred. Please try again later",
		},
		{
			name:     "wrapped known error",
			err:      fmt.Errorf("wrapped: %w", ErrArticleNotFound),
			expected: "The requested article was not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserFriendlyMessage(tt.err)
			if result != tt.expected {
				t.Errorf("GetUserFriendlyMessage(%v) = %q, want %q", tt.err, result, tt.expected)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test that our custom errors work correctly with errors.Is and errors.As
	originalErr := fmt.Errorf("original error")

	articleErr := NewArticleError("test.md", "test message", originalErr)
	configErr := NewConfigError("port", 8080, "test message", originalErr)
	validationErr := NewValidationError("email", "test@example.com", "test message", originalErr)
	httpErr := NewHTTPError(404, "not found", originalErr)

	// Test errors.Is
	if !errors.Is(articleErr, originalErr) {
		t.Error("errors.Is should work with ArticleError")
	}
	if !errors.Is(configErr, originalErr) {
		t.Error("errors.Is should work with ConfigError")
	}
	if !errors.Is(validationErr, originalErr) {
		t.Error("errors.Is should work with ValidationError")
	}
	if !errors.Is(httpErr, originalErr) {
		t.Error("errors.Is should work with HTTPError")
	}

	// Test errors.As
	var ae *ArticleError
	if !errors.As(articleErr, &ae) {
		t.Error("errors.As should work with ArticleError")
	}

	var ce *ConfigError
	if !errors.As(configErr, &ce) {
		t.Error("errors.As should work with ConfigError")
	}

	var ve *ValidationError
	if !errors.As(validationErr, &ve) {
		t.Error("errors.As should work with ValidationError")
	}

	var he *HTTPError
	if !errors.As(httpErr, &he) {
		t.Error("errors.As should work with HTTPError")
	}
}
