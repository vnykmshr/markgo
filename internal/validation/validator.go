package validation

import (
	"fmt"
	"html"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

// Validator provides centralized input validation and sanitization
type Validator struct {
	htmlPolicy *bluemonday.Policy
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
	Values map[string]string `json:"values,omitempty"`
}

// New creates a new validator instance
func New() *Validator {
	// Create a strict HTML sanitization policy
	policy := bluemonday.NewPolicy()
	// Allow only safe HTML elements for content display
	policy.AllowElements("p", "br", "strong", "em", "u", "h1", "h2", "h3", "h4", "h5", "h6")
	policy.AllowElements("ul", "ol", "li", "blockquote", "code", "pre")
	// Allow links with href attribute
	policy.AllowElements("a")
	policy.AllowAttrs("href").OnElements("a")
	policy.RequireNoReferrerOnLinks(true)
	policy.AllowURLSchemes("http", "https", "mailto")

	return &Validator{
		htmlPolicy: policy,
	}
}

// Sanitize performs basic input sanitization
func (v *Validator) Sanitize(input string) string {
	if input == "" {
		return ""
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Replace multiple whitespace with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	input = spaceRegex.ReplaceAllString(input, " ")

	// Remove null bytes and other control characters
	controlRegex := regexp.MustCompile(`[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F]`)
	input = controlRegex.ReplaceAllString(input, "")

	return input
}

// SanitizeHTML sanitizes HTML content using bluemonday
func (v *Validator) SanitizeHTML(input string) string {
	if input == "" {
		return ""
	}

	// First apply basic sanitization
	input = v.Sanitize(input)

	// Then sanitize HTML with policy
	return v.htmlPolicy.Sanitize(input)
}

// EscapeHTML escapes HTML entities
func (v *Validator) EscapeHTML(input string) string {
	if input == "" {
		return ""
	}
	return html.EscapeString(input)
}

// ValidateSlug validates URL slugs
func (v *Validator) ValidateSlug(slug string) ValidationError {
	if slug == "" {
		return ValidationError{
			Field:   "slug",
			Value:   slug,
			Message: "slug cannot be empty",
			Code:    "required",
		}
	}

	// Check length limits
	if len(slug) > 200 {
		return ValidationError{
			Field:   "slug",
			Value:   slug,
			Message: "slug cannot exceed 200 characters",
			Code:    "max_length",
		}
	}

	// Validate slug format
	slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	if !slugRegex.MatchString(slug) {
		return ValidationError{
			Field:   "slug",
			Value:   slug,
			Message: "slug must contain only lowercase letters, numbers, and hyphens",
			Code:    "invalid_format",
		}
	}

	// Check for consecutive hyphens
	if strings.Contains(slug, "--") {
		return ValidationError{
			Field:   "slug",
			Value:   slug,
			Message: "slug cannot contain consecutive hyphens",
			Code:    "invalid_format",
		}
	}

	// Check for leading/trailing hyphens
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return ValidationError{
			Field:   "slug",
			Value:   slug,
			Message: "slug cannot start or end with hyphens",
			Code:    "invalid_format",
		}
	}

	return ValidationError{} // Valid
}

// ValidateSearchQuery validates search queries with security checks
func (v *Validator) ValidateSearchQuery(query string) ValidationError {
	if query == "" {
		return ValidationError{} // Empty queries are allowed
	}

	// Check length limits
	if len(query) > 500 {
		return ValidationError{
			Field:   "query",
			Value:   query,
			Message: "search query cannot exceed 500 characters",
			Code:    "max_length",
		}
	}

	// Check for minimum length if not empty
	if len(strings.TrimSpace(query)) < 2 {
		return ValidationError{
			Field:   "query",
			Value:   query,
			Message: "search query must be at least 2 characters long",
			Code:    "min_length",
		}
	}

	// Check for dangerous patterns (XSS and injection prevention)
	// Using more precise patterns to avoid blocking legitimate programming terms
	dangerousPatterns := []string{
		// XSS patterns
		"javascript:", "vbscript:", "data:", "blob:",
		"<script", "</script>", "onload=", "onerror=", "onclick=",
		"eval(", "setTimeout(", "setInterval(",
		"document.", "window.", "location.",
		"alert(", "confirm(", "prompt(",
		// SQL injection patterns - more specific to avoid blocking legitimate searches
		"'; ", "';", "'\"", "\";", "/*", "*/", " --", "--",
		" union ", " drop ", " delete ", " truncate ",
	}

	lowerQuery := strings.ToLower(query)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerQuery, pattern) {
			return ValidationError{
				Field:   "query",
				Value:   "[redacted]",
				Message: "search query contains invalid characters",
				Code:    "invalid_characters",
			}
		}
	}

	return ValidationError{} // Valid
}

// ValidateEmail validates email addresses
func (v *Validator) ValidateEmail(email string) ValidationError {
	if email == "" {
		return ValidationError{
			Field:   "email",
			Value:   email,
			Message: "email is required",
			Code:    "required",
		}
	}

	// Basic length check
	if len(email) > 254 {
		return ValidationError{
			Field:   "email",
			Value:   email,
			Message: "email address is too long",
			Code:    "max_length",
		}
	}

	// Validate email format using Go's mail package
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ValidationError{
			Field:   "email",
			Value:   email,
			Message: "invalid email format",
			Code:    "invalid_format",
		}
	}

	return ValidationError{} // Valid
}

// ValidateText validates general text fields with customizable rules
type TextRules struct {
	Required  bool
	MinLength int
	MaxLength int
	AllowHTML bool
	FieldName string
}

func (v *Validator) ValidateText(text string, rules TextRules) ValidationError {
	fieldName := rules.FieldName
	if fieldName == "" {
		fieldName = "text"
	}

	// Check required
	if rules.Required && strings.TrimSpace(text) == "" {
		return ValidationError{
			Field:   fieldName,
			Value:   text,
			Message: fmt.Sprintf("%s is required", fieldName),
			Code:    "required",
		}
	}

	// If empty and not required, it's valid
	if text == "" {
		return ValidationError{}
	}

	// Check length limits
	textLength := utf8.RuneCountInString(text)
	if rules.MinLength > 0 && textLength < rules.MinLength {
		return ValidationError{
			Field:   fieldName,
			Value:   text,
			Message: fmt.Sprintf("%s must be at least %d characters", fieldName, rules.MinLength),
			Code:    "min_length",
		}
	}

	if rules.MaxLength > 0 && textLength > rules.MaxLength {
		return ValidationError{
			Field:   fieldName,
			Value:   text,
			Message: fmt.Sprintf("%s cannot exceed %d characters", fieldName, rules.MaxLength),
			Code:    "max_length",
		}
	}

	// Check for dangerous content if HTML is not allowed
	if !rules.AllowHTML {
		if strings.ContainsAny(text, "<>\"'&") {
			return ValidationError{
				Field:   fieldName,
				Value:   "[redacted]",
				Message: fmt.Sprintf("%s contains invalid characters", fieldName),
				Code:    "invalid_characters",
			}
		}
	}

	return ValidationError{} // Valid
}

// ValidateURL validates URLs
func (v *Validator) ValidateURL(urlStr string, required bool) ValidationError {
	if urlStr == "" {
		if required {
			return ValidationError{
				Field:   "url",
				Value:   urlStr,
				Message: "URL is required",
				Code:    "required",
			}
		}
		return ValidationError{} // Valid if not required
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ValidationError{
			Field:   "url",
			Value:   urlStr,
			Message: "invalid URL format",
			Code:    "invalid_format",
		}
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ValidationError{
			Field:   "url",
			Value:   urlStr,
			Message: "URL must use http or https protocol",
			Code:    "invalid_protocol",
		}
	}

	// Check host
	if parsedURL.Host == "" {
		return ValidationError{
			Field:   "url",
			Value:   urlStr,
			Message: "URL must include a host",
			Code:    "invalid_host",
		}
	}

	return ValidationError{} // Valid
}

// ValidateBatch validates multiple fields at once
func (v *Validator) ValidateBatch(validations map[string]func() ValidationError) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
		Values: make(map[string]string),
	}

	for field, validateFunc := range validations {
		if err := validateFunc(); err.Message != "" {
			result.Valid = false
			err.Field = field
			result.Errors = append(result.Errors, err)
		}
	}

	return result
}

// SanitizeInput applies comprehensive input sanitization
func (v *Validator) SanitizeInput(input string) string {
	if input == "" {
		return ""
	}

	// Apply basic sanitization
	input = v.Sanitize(input)

	// Escape HTML entities for safe display
	input = v.EscapeHTML(input)

	return input
}
