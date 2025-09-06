package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	validator := New()
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.htmlPolicy)
}

func TestValidator_Sanitize(t *testing.T) {
	validator := New()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "trim whitespace",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "multiple spaces",
			input:    "hello    world   test",
			expected: "hello world test",
		},
		{
			name:     "tabs and newlines",
			input:    "hello\t\tworld\n\ntest",
			expected: "hello world test",
		},
		{
			name:     "control characters",
			input:    "hello\x00\x08world\x1Ftest",
			expected: "helloworldtest", // All control characters are removed
		},
		{
			name:     "null bytes",
			input:    "hello\x00world",
			expected: "helloworld",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.Sanitize(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidator_SanitizeHTML(t *testing.T) {
	validator := New()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "safe HTML",
			input:    "<p>Hello <strong>world</strong></p>",
			expected: "<p>Hello <strong>world</strong></p>",
		},
		{
			name:     "dangerous script tag",
			input:    "<p>Hello</p><script>alert('xss')</script>",
			expected: "<p>Hello</p>",
		},
		{
			name:     "dangerous onclick",
			input:    `<p onclick="alert('xss')">Hello</p>`,
			expected: "<p>Hello</p>",
		},
		{
			name:     "safe link",
			input:    `<a href="https://example.com">Link</a>`,
			expected: `<a href="https://example.com" rel="noreferrer">Link</a>`,
		},
		{
			name:     "javascript link",
			input:    `<a href="javascript:alert('xss')">Link</a>`,
			expected: "Link",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.SanitizeHTML(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidator_EscapeHTML(t *testing.T) {
	validator := New()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "basic text",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "HTML entities",
			input:    "<script>alert('test')</script>",
			expected: "&lt;script&gt;alert(&#39;test&#39;)&lt;/script&gt;",
		},
		{
			name:     "quotes and ampersand",
			input:    `"test" & 'value'`,
			expected: "&#34;test&#34; &amp; &#39;value&#39;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.EscapeHTML(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidator_ValidateSlug(t *testing.T) {
	validator := New()

	testCases := []struct {
		name      string
		slug      string
		expectErr bool
		errCode   string
	}{
		{
			name:      "valid slug",
			slug:      "hello-world",
			expectErr: false,
		},
		{
			name:      "valid slug with numbers",
			slug:      "hello-world-123",
			expectErr: false,
		},
		{
			name:      "single word",
			slug:      "hello",
			expectErr: false,
		},
		{
			name:      "empty slug",
			slug:      "",
			expectErr: true,
			errCode:   "required",
		},
		{
			name:      "too long",
			slug:      strings.Repeat("a", 201),
			expectErr: true,
			errCode:   "max_length",
		},
		{
			name:      "uppercase letters",
			slug:      "Hello-World",
			expectErr: true,
			errCode:   "invalid_format",
		},
		{
			name:      "spaces",
			slug:      "hello world",
			expectErr: true,
			errCode:   "invalid_format",
		},
		{
			name:      "special characters",
			slug:      "hello_world!",
			expectErr: true,
			errCode:   "invalid_format",
		},
		{
			name:      "consecutive hyphens",
			slug:      "hello--world",
			expectErr: true,
			errCode:   "invalid_format",
		},
		{
			name:      "leading hyphen",
			slug:      "-hello-world",
			expectErr: true,
			errCode:   "invalid_format",
		},
		{
			name:      "trailing hyphen",
			slug:      "hello-world-",
			expectErr: true,
			errCode:   "invalid_format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateSlug(tc.slug)
			if tc.expectErr {
				assert.NotEmpty(t, err.Message)
				assert.Equal(t, tc.errCode, err.Code)
				assert.Equal(t, "slug", err.Field)
			} else {
				assert.Empty(t, err.Message)
			}
		})
	}
}

func TestValidator_ValidateSearchQuery(t *testing.T) {
	validator := New()

	testCases := []struct {
		name      string
		query     string
		expectErr bool
		errCode   string
	}{
		{
			name:      "valid query",
			query:     "hello world",
			expectErr: false,
		},
		{
			name:      "empty query",
			query:     "",
			expectErr: false, // Empty queries are allowed
		},
		{
			name:      "valid long query",
			query:     "this is a very long search query that should still be valid",
			expectErr: false,
		},
		{
			name:      "too long",
			query:     strings.Repeat("a", 501),
			expectErr: true,
			errCode:   "max_length",
		},
		{
			name:      "too short",
			query:     "a",
			expectErr: true,
			errCode:   "min_length",
		},
		{
			name:      "SQL injection attempt",
			query:     "'; DROP TABLE articles; --",
			expectErr: true,
			errCode:   "invalid_characters",
		},
		{
			name:      "XSS attempt",
			query:     "<script>alert('xss')</script>",
			expectErr: true,
			errCode:   "invalid_characters",
		},
		{
			name:      "JavaScript injection",
			query:     "javascript:alert('test')",
			expectErr: true,
			errCode:   "invalid_characters",
		},
		{
			name:      "SQL keywords",
			query:     "SELECT * FROM users",
			expectErr: true,
			errCode:   "invalid_characters",
		},
		{
			name:      "stored procedure attempt",
			query:     "xp_cmdshell",
			expectErr: true,
			errCode:   "invalid_characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateSearchQuery(tc.query)
			if tc.expectErr {
				assert.NotEmpty(t, err.Message)
				assert.Equal(t, tc.errCode, err.Code)
				assert.Equal(t, "query", err.Field)
			} else {
				assert.Empty(t, err.Message)
			}
		})
	}
}

func TestValidator_ValidateEmail(t *testing.T) {
	validator := New()

	testCases := []struct {
		name      string
		email     string
		expectErr bool
		errCode   string
	}{
		{
			name:      "valid email",
			email:     "user@example.com",
			expectErr: false,
		},
		{
			name:      "valid email with subdomain",
			email:     "user@mail.example.com",
			expectErr: false,
		},
		{
			name:      "valid email with plus",
			email:     "user+tag@example.com",
			expectErr: false,
		},
		{
			name:      "empty email",
			email:     "",
			expectErr: true,
			errCode:   "required",
		},
		{
			name:      "too long",
			email:     strings.Repeat("a", 250) + "@example.com",
			expectErr: true,
			errCode:   "max_length",
		},
		{
			name:      "invalid format - no @",
			email:     "userexample.com",
			expectErr: true,
			errCode:   "invalid_format",
		},
		{
			name:      "invalid format - no domain",
			email:     "user@",
			expectErr: true,
			errCode:   "invalid_format",
		},
		{
			name:      "invalid format - no user",
			email:     "@example.com",
			expectErr: true,
			errCode:   "invalid_format",
		},
		{
			name:      "invalid format - spaces",
			email:     "user name@example.com",
			expectErr: true,
			errCode:   "invalid_format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateEmail(tc.email)
			if tc.expectErr {
				assert.NotEmpty(t, err.Message)
				assert.Equal(t, tc.errCode, err.Code)
				assert.Equal(t, "email", err.Field)
			} else {
				assert.Empty(t, err.Message)
			}
		})
	}
}

func TestValidator_ValidateText(t *testing.T) {
	validator := New()

	testCases := []struct {
		name      string
		text      string
		rules     TextRules
		expectErr bool
		errCode   string
	}{
		{
			name:      "valid required text",
			text:      "Hello world",
			rules:     TextRules{Required: true, MinLength: 5, MaxLength: 50, FieldName: "title"},
			expectErr: false,
		},
		{
			name:      "empty required text",
			text:      "",
			rules:     TextRules{Required: true, FieldName: "title"},
			expectErr: true,
			errCode:   "required",
		},
		{
			name:      "empty optional text",
			text:      "",
			rules:     TextRules{Required: false, FieldName: "description"},
			expectErr: false,
		},
		{
			name:      "text too short",
			text:      "Hi",
			rules:     TextRules{MinLength: 5, FieldName: "title"},
			expectErr: true,
			errCode:   "min_length",
		},
		{
			name:      "text too long",
			text:      "This is a very long text that exceeds the limit",
			rules:     TextRules{MaxLength: 20, FieldName: "title"},
			expectErr: true,
			errCode:   "max_length",
		},
		{
			name:      "HTML not allowed",
			text:      "Hello <script>alert('xss')</script>",
			rules:     TextRules{AllowHTML: false, FieldName: "title"},
			expectErr: true,
			errCode:   "invalid_characters",
		},
		{
			name:      "HTML allowed",
			text:      "Hello <strong>world</strong>",
			rules:     TextRules{AllowHTML: true, FieldName: "content"},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateText(tc.text, tc.rules)
			if tc.expectErr {
				assert.NotEmpty(t, err.Message)
				assert.Equal(t, tc.errCode, err.Code)
				assert.Equal(t, tc.rules.FieldName, err.Field)
			} else {
				assert.Empty(t, err.Message)
			}
		})
	}
}

func TestValidator_ValidateURL(t *testing.T) {
	validator := New()

	testCases := []struct {
		name      string
		url       string
		required  bool
		expectErr bool
		errCode   string
	}{
		{
			name:      "valid HTTP URL",
			url:       "http://example.com",
			required:  false,
			expectErr: false,
		},
		{
			name:      "valid HTTPS URL",
			url:       "https://example.com/path",
			required:  false,
			expectErr: false,
		},
		{
			name:      "empty URL not required",
			url:       "",
			required:  false,
			expectErr: false,
		},
		{
			name:      "empty URL required",
			url:       "",
			required:  true,
			expectErr: true,
			errCode:   "required",
		},
		{
			name:      "invalid format",
			url:       "not-a-url",
			required:  false,
			expectErr: true,
			errCode:   "invalid_protocol",
		},
		{
			name:      "invalid protocol",
			url:       "ftp://example.com",
			required:  false,
			expectErr: true,
			errCode:   "invalid_protocol",
		},
		{
			name:      "no host",
			url:       "https://",
			required:  false,
			expectErr: true,
			errCode:   "invalid_host",
		},
		{
			name:      "javascript protocol",
			url:       "javascript:alert('xss')",
			required:  false,
			expectErr: true,
			errCode:   "invalid_protocol",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateURL(tc.url, tc.required)
			if tc.expectErr {
				assert.NotEmpty(t, err.Message)
				assert.Equal(t, tc.errCode, err.Code)
				assert.Equal(t, "url", err.Field)
			} else {
				assert.Empty(t, err.Message)
			}
		})
	}
}

func TestValidator_ValidateBatch(t *testing.T) {
	validator := New()

	// Test successful validation
	validations := map[string]func() ValidationError{
		"email": func() ValidationError {
			return validator.ValidateEmail("user@example.com")
		},
		"slug": func() ValidationError {
			return validator.ValidateSlug("valid-slug")
		},
	}

	result := validator.ValidateBatch(validations)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)

	// Test validation with errors
	validationsWithErrors := map[string]func() ValidationError{
		"email": func() ValidationError {
			return validator.ValidateEmail("invalid-email")
		},
		"slug": func() ValidationError {
			return validator.ValidateSlug("Invalid Slug")
		},
	}

	result = validator.ValidateBatch(validationsWithErrors)
	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 2)

	// Check that field names are set correctly
	emailError := findErrorByField(result.Errors, "email")
	assert.NotNil(t, emailError)
	assert.Equal(t, "invalid_format", emailError.Code)

	slugError := findErrorByField(result.Errors, "slug")
	assert.NotNil(t, slugError)
	assert.Equal(t, "invalid_format", slugError.Code)
}

func TestValidator_SanitizeInput(t *testing.T) {
	validator := New()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "basic text",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "text with HTML",
			input:    "hello <script>alert('xss')</script> world",
			expected: "hello &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; world",
		},
		{
			name:     "text with multiple spaces",
			input:    "hello    world",
			expected: "hello world",
		},
		{
			name:     "text with quotes",
			input:    `hello "quoted" text`,
			expected: "hello &#34;quoted&#34; text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.SanitizeInput(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	// Test with field
	err := ValidationError{
		Field:   "email",
		Message: "invalid format",
	}
	expected := "validation error in email: invalid format"
	assert.Equal(t, expected, err.Error())

	// Test without field
	err = ValidationError{
		Message: "general validation error",
	}
	expected = "validation error: general validation error"
	assert.Equal(t, expected, err.Error())
}

// Helper function to find error by field name
func findErrorByField(errors []ValidationError, field string) *ValidationError {
	for _, err := range errors {
		if err.Field == field {
			return &err
		}
	}
	return nil
}

// Benchmark tests
func BenchmarkValidator_Sanitize(b *testing.B) {
	validator := New()
	input := "  hello    world   with   multiple    spaces  "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Sanitize(input)
	}
}

func BenchmarkValidator_SanitizeHTML(b *testing.B) {
	validator := New()
	input := "<p>Hello <strong>world</strong> with <script>alert('test')</script> content</p>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.SanitizeHTML(input)
	}
}

func BenchmarkValidator_ValidateSlug(b *testing.B) {
	validator := New()
	slug := "valid-slug-example-123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateSlug(slug)
	}
}

func BenchmarkValidator_ValidateSearchQuery(b *testing.B) {
	validator := New()
	query := "this is a valid search query example"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateSearchQuery(query)
	}
}

func BenchmarkValidator_ValidateEmail(b *testing.B) {
	validator := New()
	email := "user@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateEmail(email)
	}
}
