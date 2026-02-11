package errors

import (
	"errors"
	"fmt"
	"testing"
)

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
			if unwrapped := err.Unwrap(); unwrapped != tt.err {
				t.Errorf("ValidationError.Unwrap() = %v, want %v", unwrapped, tt.err)
			}
		})
	}
}

func TestHTTPError(t *testing.T) {
	err := NewHTTPError(404, "not found", fmt.Errorf("resource missing"))
	if err.Error() != "HTTP 404: not found" {
		t.Errorf("HTTPError.Error() = %q, want %q", err.Error(), "HTTP 404: not found")
	}
	if err.StatusCode != 404 {
		t.Errorf("HTTPError.StatusCode = %d, want 404", err.StatusCode)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"article not found", ErrArticleNotFound, true},
		{"template not found", ErrTemplateNotFound, true},
		{"wrapped article not found", fmt.Errorf("wrapper: %w", ErrArticleNotFound), true},
		{"other error", ErrInvalidFrontMatter, false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.expected {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.expected)
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
		{"invalid front matter", ErrInvalidFrontMatter, true},
		{"wrapped", fmt.Errorf("wrapper: %w", ErrInvalidFrontMatter), true},
		{"other error", ErrArticleNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidationError(tt.err); got != tt.expected {
				t.Errorf("IsValidationError(%v) = %v, want %v", tt.err, got, tt.expected)
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
		{"missing config", ErrMissingConfig, true},
		{"config validation", ErrConfigValidation, true},
		{"wrapped", fmt.Errorf("wrapper: %w", ErrConfigValidation), true},
		{"other error", ErrArticleNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConfigurationError(tt.err); got != tt.expected {
				t.Errorf("IsConfigurationError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	originalErr := fmt.Errorf("original error")

	configErr := NewConfigError("port", 8080, "test message", originalErr)
	validationErr := NewValidationError("email", "test@example.com", "test message", originalErr)
	httpErr := NewHTTPError(404, "not found", originalErr)

	if !errors.Is(configErr, originalErr) {
		t.Error("errors.Is should work with ConfigError")
	}
	if !errors.Is(validationErr, originalErr) {
		t.Error("errors.Is should work with ValidationError")
	}
	if !errors.Is(httpErr, originalErr) {
		t.Error("errors.Is should work with HTTPError")
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
