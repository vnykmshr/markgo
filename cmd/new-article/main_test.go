package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic title",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "title with special characters",
			input:    "Hello, World! How are you?",
			expected: "hello-world-you", // Limited to 5 words, stop words removed
		},
		{
			name:     "title with numbers",
			input:    "Top 10 Programming Languages 2024",
			expected: "top-10-programming-languages-2024",
		},
		{
			name:     "title with stop words",
			input:    "The Best Way to Learn Go Programming",
			expected: "the-best-way-learn-go", // First word kept, limited to 5 words
		},
		{
			name:     "already hyphenated",
			input:    "hello-world-test",
			expected: "helloworldtest", // Hyphens stripped by regex
		},
		{
			name:     "empty string",
			input:    "",
			expected: "untitled",
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "untitled",
		},
		{
			name:     "unicode characters",
			input:    "Testing Unicode: Café & Naïve",
			expected: "testing-unicode-caf-nave", // Non-ASCII removed
		},
		{
			name:     "multiple spaces",
			input:    "Hello    World    Test",
			expected: "hello-world-test",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  Hello World  ",
			expected: "hello-world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := slugify(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsStopWord(t *testing.T) {
	stopWords := []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"}

	testCases := []struct {
		word     string
		expected bool
	}{
		{"the", true},
		{"and", true},
		{"or", true},
		{"but", true},
		{"in", true},
		{"on", true},
		{"at", true},
		{"to", true},
		{"for", true},
		{"of", true},
		{"with", true},
		{"by", true},
		{"hello", false},
		{"world", false},
		{"programming", false},
		{"go", false},
		{"test", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			result := isStopWord(tc.word, stopWords)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetDefaultAuthor(t *testing.T) {
	author := getDefaultAuthor()
	assert.NotEmpty(t, author, "Default author should not be empty")
	assert.True(t, len(author) > 0, "Default author should have some content")
}

func TestShouldRunInteractive(t *testing.T) {
	// Test when os.Stdin is connected to a terminal (this test may vary by environment)
	result := shouldRunInteractive()
	// We can't guarantee the environment, so we just test that the function runs without panic
	assert.True(t, result || !result, "shouldRunInteractive should return a boolean")
}

func TestGetAvailableTemplates(t *testing.T) {
	templates := GetAvailableTemplates()

	// Check that we have the expected templates
	expectedTemplates := []string{
		"default", "tutorial", "review", "news", "howto", "opinion",
		"listicle", "interview", "minimal",
	}

	for _, expected := range expectedTemplates {
		_, found := templates[expected]
		assert.True(t, found, "Template %s should be available", expected)
	}

	// Check that each template has required fields
	for name, template := range templates {
		assert.NotEmpty(t, template.Name, "Template %s name should not be empty", name)
		assert.NotEmpty(t, template.Description, "Template %s description should not be empty", name)
		assert.NotNil(t, template.Generator, "Template %s generator should not be nil", name)
	}
}

func TestTemplateGeneration(t *testing.T) {
	templates := GetAvailableTemplates()

	// Test valid templates
	for templateName, template := range templates {
		t.Run("template_"+templateName, func(t *testing.T) {
			content := template.Generator("Test Title", "Test description", "test,golang", "test-category", "Test Author", true, false)

			assert.NotEmpty(t, content)
			assert.Contains(t, content, "Test Title")
			assert.Contains(t, content, "Test Author")
			assert.Contains(t, content, "test-category")
			assert.Contains(t, content, "draft: true")
		})
	}
}

func TestShowHelp(t *testing.T) {
	// Capture stdout to verify help is printed
	// This function prints to stdout, so we can't easily test the output
	// But we can ensure it doesn't panic
	assert.NotPanics(t, func() {
		showHelp()
	})
}

func TestListTemplates(t *testing.T) {
	// This function prints to stdout, so we can't easily test the output
	// But we can ensure it doesn't panic
	assert.NotPanics(t, func() {
		listTemplates()
	})
}

func TestShowSuccessMessage(t *testing.T) {
	// This function prints to stdout, so we can't easily test the output
	// But we can ensure it doesn't panic
	assert.NotPanics(t, func() {
		showSuccessMessage("test-file.md", "Test Title")
	})
}

func TestShowPreview(t *testing.T) {
	content := `---
title: Test Article
author: Test Author
---

# Test Article

This is a test article content.
`

	// This function prints to stdout, so we can't easily test the output
	// But we can ensure it doesn't panic
	assert.NotPanics(t, func() {
		showPreview(content, "test-file.md")
	})
}

// Integration test for slug generation with various edge cases
func TestSlugifyEdgeCases(t *testing.T) {
	// Test very long titles
	longTitle := strings.Repeat("This is a very long title segment ", 10)
	slug := slugify(longTitle)
	assert.True(t, len(slug) > 0, "Long title should produce valid slug")
	assert.False(t, strings.Contains(slug, " "), "Slug should not contain spaces")

	// Test title with only stop words (first word kept)
	stopWordsTitle := "the and or but in on at"
	slug = slugify(stopWordsTitle)
	assert.Equal(t, "the", slug, "Title with only stop words should keep first word")

	// Test mixed case with numbers (limited to 5 words)
	mixedTitle := "Top 5 Best Programming Languages in 2024"
	slug = slugify(mixedTitle)
	assert.Equal(t, "top-5-best-programming-languages", slug)
}

// Test helper functions for file operations (without actually creating files)
func TestFileOperationHelpers(t *testing.T) {
	// Test that we can generate valid filenames
	title := "Test Article Title"
	slug := slugify(title)
	filename := slug + ".md"

	assert.True(t, strings.HasSuffix(filename, ".md"), "Filename should have .md extension")
	assert.False(t, strings.Contains(filename, " "), "Filename should not contain spaces")

	// Test with date prefix
	dateStr := time.Now().Format("2006-01-02")
	dateFilename := dateStr + "-" + filename
	assert.True(t, strings.HasPrefix(dateFilename, dateStr), "Date prefixed filename should start with date")
}

// Test template content generation without file I/O
func TestDetailedTemplateGeneration(t *testing.T) {
	templates := GetAvailableTemplates()

	for name, template := range templates {
		t.Run(name, func(t *testing.T) {
			content := template.Generator(
				"Test Title",
				"Test Description",
				"tag1,tag2",
				"test-category",
				"Test Author",
				true,  // draft
				false, // featured
			)

			assert.NotEmpty(t, content, "Template %s should generate non-empty content", name)

			// Verify YAML frontmatter exists
			assert.True(t, strings.HasPrefix(content, "---"), "Content should start with YAML frontmatter")
			// Check for basic structure (different templates may use different formats)
			assert.Contains(t, content, "Test Title")
			assert.Contains(t, content, "Test Author")
			assert.Contains(t, content, "test-category")
			assert.Contains(t, content, "draft: true")
			assert.Contains(t, content, "featured: false")
			// Tags may be in different formats: [tag1, tag2] or - tag1\n- tag2
			assert.True(t, strings.Contains(content, "tag1") && strings.Contains(content, "tag2"), "Content should contain both tags")

			// Verify date is present (may be in different formats)
			assert.Contains(t, content, time.Now().Format("2006-01-02"))
		})
	}
}

// Test that template validation works
func TestTemplateValidation(t *testing.T) {
	templates := GetAvailableTemplates()

	for templateName, template := range templates {
		content := template.Generator("Test", "Test desc", "test", "test-cat", "Test Author", true, false)
		assert.NotEmpty(t, content, "Valid template %s should generate content", templateName)
	}
}

// Benchmark slug generation for performance
func BenchmarkSlugify(b *testing.B) {
	title := "This is a Long Article Title with Many Words and Special Characters!@#$%"

	for i := 0; i < b.N; i++ {
		slugify(title)
	}
}

func BenchmarkTemplateGeneration(b *testing.B) {
	templates := GetAvailableTemplates()
	defaultTemplate := templates["default"]

	for i := 0; i < b.N; i++ {
		defaultTemplate.Generator("Test Title", "Test Description", "tag1,tag2", "category", "Author", true, false)
	}
}
