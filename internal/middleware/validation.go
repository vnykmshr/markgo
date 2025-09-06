package middleware

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/validation"
)

// ValidationMiddleware provides input validation middleware
type ValidationMiddleware struct {
	validator *validation.Validator
	logger    *slog.Logger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(logger *slog.Logger) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validation.New(),
		logger:    logger,
	}
}

// ValidateSlugParam validates slug parameters in routes like /articles/:slug
func (vm *ValidationMiddleware) ValidateSlugParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" {
			// No slug parameter, continue
			c.Next()
			return
		}

		// Sanitize the slug
		slug = vm.validator.Sanitize(slug)

		// Validate the slug
		if err := vm.validator.ValidateSlug(slug); err.Message != "" {
			vm.logger.Warn("Invalid slug parameter",
				"slug", slug,
				"error", err.Message,
				"ip", c.ClientIP(),
				"user_agent", c.GetHeader("User-Agent"))

			_ = c.Error(apperrors.NewValidationError("slug", slug, "Invalid URL format", apperrors.ErrValidationFailed))
			c.Abort()
			return
		}

		// Store sanitized slug back to context
		c.Params = gin.Params{
			{Key: "slug", Value: slug},
		}

		c.Next()
	}
}

// ValidateSearchQuery validates search query parameters
func (vm *ValidationMiddleware) ValidateSearchQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			// No query parameter, continue
			c.Next()
			return
		}

		// Sanitize the query
		query = vm.validator.Sanitize(query)

		// Validate the query
		if err := vm.validator.ValidateSearchQuery(query); err.Message != "" {
			vm.logger.Warn("Invalid search query",
				"query", "[redacted]",
				"error", err.Message,
				"ip", c.ClientIP(),
				"user_agent", c.GetHeader("User-Agent"))

			_ = c.Error(apperrors.NewValidationError("query", nil, "Invalid search query", apperrors.ErrValidationFailed))
			c.Abort()
			return
		}

		// Store sanitized query back to request
		c.Request.URL.RawQuery = strings.Replace(c.Request.URL.RawQuery, c.Query("q"), query, 1)

		c.Next()
	}
}

// ValidateTagCategory validates tag and category parameters
func (vm *ValidationMiddleware) ValidateTagCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for tag parameter
		if tag := c.Param("tag"); tag != "" {
			tag = vm.validator.Sanitize(tag)

			// Validate tag format (similar to slug but more permissive)
			if err := vm.validateTagCategory(tag, "tag"); err.Message != "" {
				vm.logger.Warn("Invalid tag parameter",
					"tag", tag,
					"error", err.Message,
					"ip", c.ClientIP())

				_ = c.Error(apperrors.NewValidationError("tag", tag, "Invalid tag format", apperrors.ErrValidationFailed))
				c.Abort()
				return
			}

			// Update param
			updateParam(c, "tag", tag)
		}

		// Check for category parameter
		if category := c.Param("category"); category != "" {
			category = vm.validator.Sanitize(category)

			// Validate category format
			if err := vm.validateTagCategory(category, "category"); err.Message != "" {
				vm.logger.Warn("Invalid category parameter",
					"category", category,
					"error", err.Message,
					"ip", c.ClientIP())

				_ = c.Error(apperrors.NewValidationError("category", category, "Invalid category format", apperrors.ErrValidationFailed))
				c.Abort()
				return
			}

			// Update param
			updateParam(c, "category", category)
		}

		c.Next()
	}
}

// ValidatePagination validates pagination parameters
func (vm *ValidationMiddleware) ValidatePagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate page parameter
		if pageStr := c.Query("page"); pageStr != "" {
			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 1 || page > 10000 {
				vm.logger.Warn("Invalid page parameter",
					"page", pageStr,
					"ip", c.ClientIP())

				_ = c.Error(apperrors.NewValidationError("page", pageStr, "Page must be a number between 1 and 10000", apperrors.ErrValidationFailed))
				c.Abort()
				return
			}
		}

		// Validate per_page parameter
		if perPageStr := c.Query("per_page"); perPageStr != "" {
			perPage, err := strconv.Atoi(perPageStr)
			if err != nil || perPage < 1 || perPage > 100 {
				vm.logger.Warn("Invalid per_page parameter",
					"per_page", perPageStr,
					"ip", c.ClientIP())

				_ = c.Error(apperrors.NewValidationError("per_page", perPageStr, "Items per page must be between 1 and 100", apperrors.ErrValidationFailed))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ValidateContactMessage validates contact form submissions
func (vm *ValidationMiddleware) ValidateContactMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware runs after JSON binding but before business logic
		// We can access the bound data through c.Get() if needed

		// For now, we rely on the ContactMessage struct validation tags
		// But we could add additional sanitization here if needed

		c.Next()
	}
}

// SanitizeJSONInput sanitizes JSON input fields
func (vm *ValidationMiddleware) SanitizeJSONInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would require more complex JSON parsing and sanitization
		// For now, we rely on struct-level validation and sanitization
		// in individual handlers

		c.Next()
	}
}

// validateTagCategory validates tag and category formats
func (vm *ValidationMiddleware) validateTagCategory(value, fieldType string) validation.ValidationError {
	if value == "" {
		return validation.ValidationError{
			Field:   fieldType,
			Value:   value,
			Message: fieldType + " cannot be empty",
			Code:    "required",
		}
	}

	// Check length
	if len(value) > 50 {
		return validation.ValidationError{
			Field:   fieldType,
			Value:   value,
			Message: fieldType + " cannot exceed 50 characters",
			Code:    "max_length",
		}
	}

	// Allow letters, numbers, hyphens, and underscores
	// More permissive than slugs to support various tag formats
	validFormat := true
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '+') {
			validFormat = false
			break
		}
	}

	if !validFormat {
		return validation.ValidationError{
			Field:   fieldType,
			Value:   value,
			Message: fieldType + " contains invalid characters",
			Code:    "invalid_format",
		}
	}

	return validation.ValidationError{} // Valid
}

// updateParam safely updates a parameter in gin.Context
func updateParam(c *gin.Context, key, value string) {
	for i, param := range c.Params {
		if param.Key == key {
			c.Params[i].Value = value
			return
		}
	}
	// If param doesn't exist, add it
	c.Params = append(c.Params, gin.Param{Key: key, Value: value})
}

// ValidationConfig holds validation middleware configuration
type ValidationConfig struct {
	EnableSlugValidation        bool
	EnableSearchValidation      bool
	EnableTagCategoryValidation bool
	EnablePaginationValidation  bool
	LogFailures                 bool
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		EnableSlugValidation:        true,
		EnableSearchValidation:      true,
		EnableTagCategoryValidation: true,
		EnablePaginationValidation:  true,
		LogFailures:                 true,
	}
}

// ValidationMiddlewareWithConfig creates validation middleware with custom config
func ValidationMiddlewareWithConfig(logger *slog.Logger, config ValidationConfig) gin.HandlerFunc {
	vm := NewValidationMiddleware(logger)

	return func(c *gin.Context) {
		// Apply validations based on configuration and route patterns
		path := c.Request.URL.Path

		if config.EnableSlugValidation && containsSlugParam(path) {
			vm.ValidateSlugParam()(c)
			if c.IsAborted() {
				return
			}
		}

		if config.EnableSearchValidation && strings.Contains(path, "/search") {
			vm.ValidateSearchQuery()(c)
			if c.IsAborted() {
				return
			}
		}

		if config.EnableTagCategoryValidation && (strings.Contains(path, "/tags/") || strings.Contains(path, "/categories/")) {
			vm.ValidateTagCategory()(c)
			if c.IsAborted() {
				return
			}
		}

		if config.EnablePaginationValidation {
			vm.ValidatePagination()(c)
			if c.IsAborted() {
				return
			}
		}

		c.Next()
	}
}

// containsSlugParam checks if path contains slug parameter patterns
func containsSlugParam(path string) bool {
	slugPatterns := []string{
		"/articles/",
		"/drafts/",
	}

	for _, pattern := range slugPatterns {
		if strings.Contains(path, pattern) && !strings.HasSuffix(path, pattern[:len(pattern)-1]) {
			return true
		}
	}

	return false
}
