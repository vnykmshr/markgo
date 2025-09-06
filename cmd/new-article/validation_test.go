package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "title",
		Value:   "invalid title",
		Message: "title is too short",
	}

	expected := "validation error in title: title is too short"
	assert.Equal(t, expected, err.Error())
}

func TestValidateTitle(t *testing.T) {
	testCases := []struct {
		name        string
		title       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid title",
			title:       "This is a Valid Title",
			expectError: false,
		},
		{
			name:        "empty title",
			title:       "",
			expectError: true,
			errorMsg:    "title cannot be empty",
		},
		{
			name:        "whitespace only title",
			title:       "   ",
			expectError: true,
			errorMsg:    "title cannot be empty",
		},
		{
			name:        "title too short",
			title:       "Hi",
			expectError: true,
			errorMsg:    "title must be at least 3 characters long",
		},
		{
			name:        "title too long",
			title:       strings.Repeat("a", 201),
			expectError: true,
			errorMsg:    "title cannot exceed 200 characters",
		},
		{
			name:        "title with special characters (allowed)",
			title:       "Title with <script>alert('xss')</script>",
			expectError: false, // Title validation doesn't check for invalid characters
		},
		{
			name:        "title with valid special characters",
			title:       "Title: How to Code in Go (Part 1)",
			expectError: false,
		},
		{
			name:        "minimum valid length",
			title:       "abc",
			expectError: false,
		},
		{
			name:        "maximum valid length",
			title:       strings.Repeat("a", 200),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTitle(tc.title)

			if tc.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tc.errorMsg)
				assert.Equal(t, "title", err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	testCases := []struct {
		name        string
		description string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid description",
			description: "This is a valid description for the article",
			expectError: false,
		},
		{
			name:        "empty description",
			description: "",
			expectError: false, // Empty description is allowed
		},
		{
			name:        "description too long",
			description: strings.Repeat("a", 501),
			expectError: true,
			errorMsg:    "description cannot exceed 500 characters",
		},
		{
			name:        "description with special characters (allowed)",
			description: "Description with <script>alert('xss')</script>",
			expectError: false, // Description validation doesn't check for invalid characters
		},
		{
			name:        "maximum valid length",
			description: strings.Repeat("a", 500),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDescription(tc.description)

			if tc.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tc.errorMsg)
				assert.Equal(t, "description", err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateTags(t *testing.T) {
	testCases := []struct {
		name        string
		tags        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid single tag",
			tags:        "programming",
			expectError: false,
		},
		{
			name:        "valid multiple tags",
			tags:        "programming,go,tutorial",
			expectError: false,
		},
		{
			name:        "empty tags",
			tags:        "",
			expectError: true,
			errorMsg:    "at least one tag is required",
		},
		{
			name:        "tags with spaces",
			tags:        "web development, golang, tutorial",
			expectError: false,
		},
		{
			name:        "too many tags",
			tags:        "tag1,tag2,tag3,tag4,tag5,tag6,tag7,tag8,tag9,tag10,tag11",
			expectError: true,
			errorMsg:    "cannot have more than 10 tags",
		},
		{
			name:        "tag too long",
			tags:        strings.Repeat("a", 51),
			expectError: true,
			errorMsg:    "cannot exceed 50 characters",
		},
		{
			name:        "tag with invalid characters",
			tags:        "valid-tag,<script>alert('xss')</script>",
			expectError: true,
			errorMsg:    "contains invalid characters",
		},
		{
			name:        "short tags allowed",
			tags:        "a,bb",
			expectError: false, // Short tags are allowed, empty tags are skipped
		},
		{
			name:        "duplicate tags allowed",
			tags:        "programming,go,programming",
			expectError: false, // Duplicate validation not implemented
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTags(tc.tags)

			if tc.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tc.errorMsg)
				assert.Equal(t, "tags", err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateCategory(t *testing.T) {
	testCases := []struct {
		name        string
		category    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid category",
			category:    "programming",
			expectError: false,
		},
		{
			name:        "empty category",
			category:    "",
			expectError: true,
			errorMsg:    "category cannot be empty",
		},
		{
			name:        "short category allowed",
			category:    "a",
			expectError: false, // No minimum length validation for categories
		},
		{
			name:        "category too long",
			category:    strings.Repeat("a", 101),
			expectError: true,
			errorMsg:    "category cannot exceed 100 characters",
		},
		{
			name:        "category with invalid characters",
			category:    "<script>alert('xss')</script>",
			expectError: true,
			errorMsg:    "category contains invalid characters",
		},
		{
			name:        "category with valid hyphen",
			category:    "web-development",
			expectError: false,
		},
		{
			name:        "category with valid underscore",
			category:    "web_development",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCategory(tc.category)

			if tc.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tc.errorMsg)
				assert.Equal(t, "category", err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateAuthor(t *testing.T) {
	testCases := []struct {
		name        string
		author      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid author",
			author:      "John Doe",
			expectError: false,
		},
		{
			name:        "empty author",
			author:      "",
			expectError: true,
			errorMsg:    "author name cannot be empty",
		},
		{
			name:        "author too short",
			author:      "J",
			expectError: true,
			errorMsg:    "author name must be at least 2 characters long",
		},
		{
			name:        "author too long",
			author:      strings.Repeat("a", 101),
			expectError: true,
			errorMsg:    "author name cannot exceed 100 characters",
		},
		{
			name:        "author with special characters (allowed)",
			author:      "John <script>alert('xss')</script> Doe",
			expectError: false, // Author validation doesn't check for invalid characters
		},
		{
			name:        "author with valid special characters",
			author:      "Jean-Claude Van Damme Jr.",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAuthor(tc.author)

			if tc.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tc.errorMsg)
				assert.Equal(t, "author", err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateTemplate(t *testing.T) {
	testCases := []struct {
		name        string
		template    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid template - default",
			template:    "default",
			expectError: false,
		},
		{
			name:        "valid template - tutorial",
			template:    "tutorial",
			expectError: false,
		},
		{
			name:        "invalid template",
			template:    "nonexistent",
			expectError: true,
			errorMsg:    "template 'nonexistent' not found",
		},
		{
			name:        "empty template allowed",
			template:    "",
			expectError: false, // Empty template is allowed, defaults to "default"
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTemplate(tc.template)

			if tc.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tc.errorMsg)
				assert.Equal(t, "template", err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateSlug(t *testing.T) {
	testCases := []struct {
		name        string
		slug        string
		expectError bool
	}{
		{
			name:        "valid slug",
			slug:        "valid-slug-123",
			expectError: false,
		},
		{
			name:        "empty slug",
			slug:        "",
			expectError: true,
		},
		{
			name:        "slug with spaces",
			slug:        "invalid slug",
			expectError: true,
		},
		{
			name:        "slug with special characters",
			slug:        "invalid@slug#$%",
			expectError: true,
		},
		{
			name:        "slug too long",
			slug:        strings.Repeat("a", 201),
			expectError: true,
		},
		{
			name:        "slug with consecutive hyphens",
			slug:        "invalid--slug",
			expectError: true,
		},
		{
			name:        "slug starting with hyphen",
			slug:        "-invalid-slug",
			expectError: true,
		},
		{
			name:        "slug ending with hyphen",
			slug:        "invalid-slug-",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSlug(tc.slug)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOutputPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_articles")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		name        string
		filepath    string
		expectError bool
		setup       func() string
	}{
		{
			name: "valid path in existing directory",
			setup: func() string {
				return filepath.Join(tempDir, "test-article.md")
			},
			expectError: false,
		},
		{
			name: "existing file",
			setup: func() string {
				path := filepath.Join(tempDir, "existing.md")
				file, _ := os.Create(path)
				file.Close()
				return path
			},
			expectError: true,
		},
		{
			name:        "invalid path with invalid characters",
			filepath:    "/invalid\x00path/file.md",
			expectError: true,
		},
		{
			name: "path in non-existent directory (gets created)",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent", "test.md")
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var testPath string
			if tc.setup != nil {
				testPath = tc.setup()
			} else {
				testPath = tc.filepath
			}

			err := ValidateOutputPath(testPath)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean input",
			input:    "Clean Input Text",
			expected: "Clean Input Text",
		},
		{
			name:     "input with leading/trailing whitespace",
			input:    "  Whitespace Test  ",
			expected: "Whitespace Test",
		},
		{
			name:     "input with multiple spaces",
			input:    "Multiple    Spaces    Test",
			expected: "Multiple Spaces Test",
		},
		{
			name:     "input with tabs and newlines",
			input:    "Text\twith\nnewlines\tand\ttabs",
			expected: "Text with newlines and tabs",
		},
		{
			name:     "input with special characters",
			input:    "Text with special chars: @#$%^&*()",
			expected: "Text with special chars: @#$%^&*()",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \t\n   ",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeInput(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSanitizeForYAML(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean input",
			input:    "Clean Input",
			expected: "Clean Input",
		},
		{
			name:     "input with quotes",
			input:    `Text with "double quotes" and 'single quotes'`,
			expected: "Text with \\\"double quotes\\\" and 'single quotes'",
		},
		{
			name:     "input with backslashes (unchanged)",
			input:    `Text with \ backslashes \`,
			expected: `Text with \ backslashes \`, // SanitizeForYAML doesn't escape backslashes
		},
		{
			name:     "input with newlines (unchanged)",
			input:    "Line 1\nLine 2\r\nLine 3",
			expected: "Line 1\nLine 2\r\nLine 3", // SanitizeForYAML doesn't handle newlines
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeForYAML(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateArticleInput(t *testing.T) {
	testCases := []struct {
		name         string
		title        string
		description  string
		tags         string
		category     string
		author       string
		template     string
		expectValid  bool
		expectErrors int
	}{
		{
			name:         "all valid inputs",
			title:        "Valid Article Title",
			description:  "Valid description",
			tags:         "programming,go,tutorial",
			category:     "programming",
			author:       "John Doe",
			template:     "default",
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name:         "multiple validation errors",
			title:        "", // invalid
			description:  "Valid description",
			tags:         "",        // invalid
			category:     "",        // invalid
			author:       "",        // invalid
			template:     "invalid", // invalid
			expectValid:  false,
			expectErrors: 5,
		},
		{
			name:         "empty description allowed",
			title:        "Valid Title",
			description:  "",
			tags:         "programming",
			category:     "programming",
			author:       "John Doe",
			template:     "default",
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name:         "single validation error",
			title:        "ab", // too short
			description:  "Valid description",
			tags:         "programming",
			category:     "programming",
			author:       "John Doe",
			template:     "default",
			expectValid:  false,
			expectErrors: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateArticleInput(tc.title, tc.description, tc.tags, tc.category, tc.author, tc.template)

			assert.Equal(t, tc.expectValid, result.Valid)
			assert.Len(t, result.Errors, tc.expectErrors)

			if !tc.expectValid {
				// Verify all errors have proper structure
				for _, err := range result.Errors {
					assert.NotEmpty(t, err.Field)
					assert.NotEmpty(t, err.Message)
				}
			}
		})
	}
}

func TestShowValidationErrors(t *testing.T) {
	errors := []ValidationError{
		{Field: "title", Value: "", Message: "title cannot be empty"},
		{Field: "tags", Value: "", Message: "at least one tag is required"},
	}

	// This function prints to stdout, so we can't easily test the output
	// But we can ensure it doesn't panic
	assert.NotPanics(t, func() {
		ShowValidationErrors(errors)
	})
}

// Benchmark validation functions for performance
func BenchmarkValidateTitle(b *testing.B) {
	title := "This is a Sample Article Title for Benchmarking"

	for i := 0; i < b.N; i++ {
		_ = validateTitle(title)
	}
}

func BenchmarkValidateTags(b *testing.B) {
	tags := "programming,golang,tutorial,web-development,backend"

	for i := 0; i < b.N; i++ {
		_ = validateTags(tags)
	}
}

func BenchmarkSanitizeInput(b *testing.B) {
	input := "  Sample input   with   multiple    spaces  and  \t tabs \n newlines  "

	for i := 0; i < b.N; i++ {
		SanitizeInput(input)
	}
}
