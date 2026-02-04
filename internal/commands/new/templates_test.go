package new

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// frontmatter mirrors the YAML frontmatter structure for roundtrip testing
type frontmatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Date        string   `yaml:"date"`
	Tags        []string `yaml:"tags"`
	Categories  []string `yaml:"categories"`
	Featured    bool     `yaml:"featured"`
	Draft       bool     `yaml:"draft"`
	Author      string   `yaml:"author"`
}

// extractFrontmatter splits a generated article and unmarshals its YAML frontmatter.
func extractFrontmatter(t *testing.T, article string) frontmatter {
	t.Helper()
	parts := strings.SplitN(article, "---", 3)
	require.Len(t, parts, 3, "expected frontmatter delimiters")

	var fm frontmatter
	require.NoError(t, yaml.Unmarshal([]byte(parts[1]), &fm))
	return fm
}

func TestGenerateArticleWithTemplate_Roundtrip(t *testing.T) {
	templates := GetAvailableTemplates()
	require.NotEmpty(t, templates)

	for name, tmpl := range templates {
		t.Run(name, func(t *testing.T) {
			article := tmpl.Generator("Test Title", "A description", "go, testing", "tech", "Jane Doe", true, false)

			fm := extractFrontmatter(t, article)
			assert.Equal(t, "Test Title", fm.Title)
			assert.Equal(t, "A description", fm.Description)
			assert.Equal(t, "Jane Doe", fm.Author)
			assert.True(t, fm.Draft)
			assert.False(t, fm.Featured)
			assert.Contains(t, fm.Tags, "go")
			assert.Contains(t, fm.Tags, "testing")
		})
	}
}

func TestGenerateArticleWithTemplate_SpecialCharsInTitle(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		author      string
	}{
		{
			name:        "quotes in title",
			title:       `My Post: "A Guide"`,
			description: "plain",
			author:      "Jane",
		},
		{
			name:        "colon in title",
			title:       "Part 1: Getting Started",
			description: "plain",
			author:      "Jane",
		},
		{
			name:        "backslash in title",
			title:       `Path C:\Users\file`,
			description: "plain",
			author:      "Jane",
		},
		{
			name:        "quotes in description",
			title:       "Normal Title",
			description: `She said "hello"`,
			author:      "Jane",
		},
		{
			name:        "quotes in author",
			title:       "Normal Title",
			description: "plain",
			author:      `O"Brien`,
		},
		{
			name:        "hash in title",
			title:       "Issue #42 Resolved",
			description: "plain",
			author:      "Jane",
		},
		{
			name:        "pipe in description",
			title:       "Normal",
			description: "use cmd | grep",
			author:      "Jane",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := generateDefaultArticle(tt.title, tt.description, "go", "tech", tt.author, false, false)

			// The critical assertion: YAML must parse back correctly
			fm := extractFrontmatter(t, article)
			assert.Equal(t, tt.title, fm.Title, "title roundtrip failed")
			assert.Equal(t, tt.description, fm.Description, "description roundtrip failed")
			assert.Equal(t, tt.author, fm.Author, "author roundtrip failed")
		})
	}
}

func TestGenerateArticleWithTemplate_DraftAndFeatured(t *testing.T) {
	article := generateMinimalArticle("Title", "Desc", "go", "tech", "Author", false, true)
	fm := extractFrontmatter(t, article)
	assert.False(t, fm.Draft)
	assert.True(t, fm.Featured)
}

func TestGetAvailableTemplates(t *testing.T) {
	templates := GetAvailableTemplates()

	expectedNames := []string{"default", "tutorial", "review", "news", "howto", "opinion", "listicle", "interview", "minimal"}
	for _, name := range expectedNames {
		t.Run(name, func(t *testing.T) {
			tmpl, exists := templates[name]
			assert.True(t, exists, "template %q should exist", name)
			assert.NotEmpty(t, tmpl.Name)
			assert.NotEmpty(t, tmpl.Description)
			assert.NotNil(t, tmpl.Generator)
		})
	}
}
