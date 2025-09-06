package errors

import (
	"errors"
	"fmt"
)

// Domain-specific errors for MarkGo application
var (
	// Article errors
	ErrArticleNotFound    = errors.New("article not found")
	ErrArticleParseError  = errors.New("article parse error")
	ErrInvalidFrontMatter = errors.New("invalid front matter")
	ErrArticleExists      = errors.New("article already exists")

	// Email errors
	ErrEmailNotConfigured = errors.New("email credentials not configured")
	ErrSMTPAuthFailed     = errors.New("SMTP authentication failed")
	ErrEmailSendFailed    = errors.New("email send failed")
	ErrInvalidEmail       = errors.New("invalid email address")

	// Template errors
	ErrTemplateNotFound   = errors.New("template not found")
	ErrTemplateParseError = errors.New("template parse error")
	ErrTemplateRender     = errors.New("template render error")

	// Configuration errors
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrMissingConfig    = errors.New("missing required configuration")
	ErrConfigValidation = errors.New("configuration validation failed")

	// Cache errors
	ErrCacheNotFound = errors.New("cache entry not found")
	ErrCacheInvalid  = errors.New("invalid cache data")

	// Search errors
	ErrSearchFailed  = errors.New("search operation failed")
	ErrInvalidQuery  = errors.New("invalid search query")
	ErrSearchTimeout = errors.New("search operation timed out")

	// File system errors
	ErrFileNotFound    = errors.New("file not found")
	ErrFilePermission  = errors.New("file permission denied")
	ErrDirectoryCreate = errors.New("failed to create directory")

	// Validation errors
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidInput     = errors.New("invalid input")
	ErrMissingField     = errors.New("missing required field")
)

// ArticleError represents an error related to article processing
type ArticleError struct {
	File    string
	Message string
	Err     error
}

func (e *ArticleError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("article error in %s: %s", e.File, e.Message)
	}
	return fmt.Sprintf("article error: %s", e.Message)
}

func (e *ArticleError) Unwrap() error {
	return e.Err
}

// NewArticleError creates a new ArticleError
func NewArticleError(file, message string, err error) *ArticleError {
	return &ArticleError{
		File:    file,
		Message: message,
		Err:     err,
	}
}

// ConfigError represents a configuration-related error
type ConfigError struct {
	Field   string
	Value   interface{}
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("config error for field %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError
func NewConfigError(field string, value interface{}, message string, err error) *ConfigError {
	return &ConfigError{
		Field:   field,
		Value:   value,
		Message: message,
		Err:     err,
	}
}

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new ValidationError
func NewValidationError(field string, value interface{}, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Err:     err,
	}
}

// HTTPError represents an HTTP-specific error with status code
type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTPError
func NewHTTPError(statusCode int, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// IsNotFound checks if an error indicates a "not found" condition
func IsNotFound(err error) bool {
	return errors.Is(err, ErrArticleNotFound) ||
		errors.Is(err, ErrTemplateNotFound) ||
		errors.Is(err, ErrFileNotFound) ||
		errors.Is(err, ErrCacheNotFound)
}

// IsValidationError checks if an error is a validation-related error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrMissingField) ||
		errors.Is(err, ErrInvalidFrontMatter)
}

// IsConfigurationError checks if an error is configuration-related
func IsConfigurationError(err error) bool {
	return errors.Is(err, ErrInvalidConfig) ||
		errors.Is(err, ErrMissingConfig) ||
		errors.Is(err, ErrConfigValidation)
}

// IsArticleNotFound checks if an error indicates an article was not found
func IsArticleNotFound(err error) bool {
	return errors.Is(err, ErrArticleNotFound)
}

// GetUserFriendlyMessage returns a user-friendly error message
func GetUserFriendlyMessage(err error) string {
	switch {
	case errors.Is(err, ErrArticleNotFound):
		return "The requested article was not found"
	case errors.Is(err, ErrTemplateNotFound):
		return "Page template not found"
	case errors.Is(err, ErrEmailNotConfigured):
		return "Email service is currently unavailable"
	case errors.Is(err, ErrSMTPAuthFailed):
		return "Email service authentication failed"
	case errors.Is(err, ErrValidationFailed):
		return "Please check your input and try again"
	case errors.Is(err, ErrInvalidQuery):
		return "Invalid search query. Please try different search terms"
	case errors.Is(err, ErrSearchTimeout):
		return "Search took too long. Please try again with more specific terms"
	case IsConfigurationError(err):
		return "Service configuration error. Please contact administrator"
	case IsNotFound(err):
		return "The requested resource was not found"
	case IsValidationError(err):
		return "Invalid input provided. Please check your data and try again"
	default:
		return "An unexpected error occurred. Please try again later"
	}
}
