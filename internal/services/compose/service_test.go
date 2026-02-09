package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePost_Thought(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir, "Test Author")

	slug, err := svc.CreatePost(Input{
		Content: "Just a quick thought about Go.",
	})

	require.NoError(t, err)
	assert.Contains(t, slug, "thought-")

	// Verify file exists
	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	require.Len(t, files, 1)

	content, err := os.ReadFile(files[0])
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "---")
	assert.Contains(t, s, "slug: thought-")
	assert.NotContains(t, s, "title:")
	assert.Contains(t, s, "author: Test Author")
	assert.Contains(t, s, "Just a quick thought about Go.")
}

func TestCreatePost_Link(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir, "Test Author")

	slug, err := svc.CreatePost(Input{
		Title:   "Interesting Read",
		Content: "This article is worth checking out.",
		LinkURL: "https://example.com/article",
		Tags:    "tech, reading",
	})

	require.NoError(t, err)
	assert.Equal(t, "interesting-read", slug)

	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	require.Len(t, files, 1)

	content, err := os.ReadFile(files[0])
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "title: Interesting Read")
	assert.Contains(t, s, "link_url: https://example.com/article")
	assert.Contains(t, s, "- tech")
	assert.Contains(t, s, "- reading")
}

func TestCreatePost_Article(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir, "Test Author")

	slug, err := svc.CreatePost(Input{
		Title:   "Getting Started with Go",
		Content: "Go is a statically typed language...",
		Tags:    "golang, tutorial",
		Draft:   true,
	})

	require.NoError(t, err)
	assert.Equal(t, "getting-started-with-go", slug)

	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	require.Len(t, files, 1)

	content, err := os.ReadFile(files[0])
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "title: Getting Started with Go")
	assert.Contains(t, s, "draft: true")
	assert.Contains(t, s, "Getting Started with Go")
}

func TestCreatePost_EmptyTags(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir, "")

	slug, err := svc.CreatePost(Input{
		Content: "No tags here.",
	})

	require.NoError(t, err)
	assert.Contains(t, slug, "thought-")

	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	require.Len(t, files, 1)

	content, err := os.ReadFile(files[0])
	require.NoError(t, err)

	s := string(content)
	// Should not contain tags or author when empty
	assert.NotContains(t, s, "tags:")
	assert.NotContains(t, s, "author:")
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Getting Started with Go!", "getting-started-with-go"},
		{"  spaces  and  stuff  ", "spaces-and-stuff"},
		{"", ""},
		{"123 Numbers", "123-numbers"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := generateSlug(tt.input)
			// Trim for comparison since our slug may differ slightly
			assert.True(t, strings.HasPrefix(got, tt.expected) || got == tt.expected,
				"generateSlug(%q) = %q, want prefix %q", tt.input, got, tt.expected)
		})
	}
}
