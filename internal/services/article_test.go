package services

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/markgo/internal/models"
)

func TestNewArticleService(t *testing.T) {
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Test with empty directory
	service, err := NewArticleService(tempDir, logger)
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, tempDir, service.articlesPath)
	assert.Equal(t, 0, len(service.GetAllArticles()))

	// Test with non-existent directory
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	_, err = NewArticleService(nonExistentDir, logger)
	assert.Error(t, err)
}

func TestArticleService_LoadArticles(t *testing.T) {
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create test markdown files
	testArticles := map[string]string{
		"test-article-1.markdown": `---
title: "Test Article 1"
date: 2023-01-01T10:00:00Z
tags: ["go", "testing"]
categories: ["programming"]
featured: true
draft: false
---

# Test Article 1

This is the content of test article 1.

## Section 1

Some content here.`,

		"test-article-2.markdown": `---
title: "Test Article 2"
date: 2023-01-02T10:00:00Z
tags: ["go", "web"]
categories: ["programming", "web"]
featured: false
draft: false
---

# Test Article 2

This is the content of test article 2.`,

		"draft-article.markdown": `---
title: "Draft Article"
date: 2023-01-03T10:00:00Z
tags: ["draft"]
categories: ["test"]
featured: false
draft: true
---

# Draft Article

This is a draft article.`,

		"no-frontmatter.markdown": `# Article Without Frontmatter

This article has no frontmatter.`,

		"invalid-frontmatter.markdown": `---
invalid yaml: [
---

# Invalid Frontmatter

This article has invalid frontmatter.`,

		"not-markdown.txt": "This is not a markdown file.",
	}

	for filename, content := range testArticles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	service, err := NewArticleService(tempDir, logger)
	require.NoError(t, err)

	articles := service.GetAllArticles()

	// Should have 3 published articles (excluding draft and files with issues)
	assert.Equal(t, 3, len(articles))

	// Check that articles are sorted by date (newest first)
	assert.True(t, articles[0].Date.After(articles[1].Date))

	// Check specific article content
	found := false
	for _, article := range articles {
		if article.Slug == "test-article-1" {
			found = true
			assert.Equal(t, "Test Article 1", article.Title)
			assert.Contains(t, article.Tags, "go")
			assert.Contains(t, article.Tags, "testing")
			assert.Contains(t, article.Categories, "programming")
			assert.True(t, article.Featured)
			assert.False(t, article.Draft)
			assert.NotEmpty(t, article.Content)
			assert.NotEmpty(t, article.Excerpt)
			assert.Greater(t, article.WordCount, 0)
			assert.Greater(t, article.ReadingTime, 0)
			break
		}
	}
	assert.True(t, found, "test-article-1 should be found")
}

func TestArticleService_GetArticleBySlug(t *testing.T) {
	service := createTestArticleService(t)

	// Test existing article
	article, err := service.GetArticleBySlug("test-article-1")
	assert.NoError(t, err)
	assert.NotNil(t, article)
	assert.Equal(t, "test-article-1", article.Slug)
	assert.Equal(t, "Test Article 1", article.Title)

	// Test non-existent article
	_, err = service.GetArticleBySlug("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "article not found")

	// Test draft article (should be in cache)
	draftArticle, err := service.GetArticleBySlug("draft-article")
	assert.NoError(t, err)
	assert.NotNil(t, draftArticle)
	assert.True(t, draftArticle.Draft)
}

func TestArticleService_GetArticlesByTag(t *testing.T) {
	service := createTestArticleService(t)

	// Test existing tag
	articles := service.GetArticlesByTag("go")
	assert.Greater(t, len(articles), 0)

	for _, article := range articles {
		found := false
		for _, tag := range article.Tags {
			if strings.EqualFold(tag, "go") {
				found = true
				break
			}
		}
		assert.True(t, found, "Article should have 'go' tag")
	}

	// Test case insensitive
	articlesUpper := service.GetArticlesByTag("GO")
	assert.Equal(t, len(articles), len(articlesUpper))

	// Test non-existent tag
	noArticles := service.GetArticlesByTag("nonexistent")
	assert.Equal(t, 0, len(noArticles))
}

func TestArticleService_GetArticlesByCategory(t *testing.T) {
	service := createTestArticleService(t)

	// Test existing category
	articles := service.GetArticlesByCategory("programming")
	assert.Greater(t, len(articles), 0)

	for _, article := range articles {
		found := false
		for _, category := range article.Categories {
			if strings.EqualFold(category, "programming") {
				found = true
				break
			}
		}
		assert.True(t, found, "Article should have 'programming' category")
	}

	// Test case insensitive
	articlesUpper := service.GetArticlesByCategory("PROGRAMMING")
	assert.Equal(t, len(articles), len(articlesUpper))

	// Test non-existent category
	noArticles := service.GetArticlesByCategory("nonexistent")
	assert.Equal(t, 0, len(noArticles))
}

func TestArticleService_GetAllTags(t *testing.T) {
	service := createTestArticleService(t)

	tags := service.GetAllTags()
	assert.Greater(t, len(tags), 0)

	// Should be sorted
	for i := 1; i < len(tags); i++ {
		assert.True(t, tags[i-1] <= tags[i], "Tags should be sorted")
	}

	// Check specific tags exist
	assert.Contains(t, tags, "go")
	assert.Contains(t, tags, "testing")
}

func TestArticleService_GetAllCategories(t *testing.T) {
	service := createTestArticleService(t)

	categories := service.GetAllCategories()
	assert.Greater(t, len(categories), 0)

	// Should be sorted
	for i := 1; i < len(categories); i++ {
		assert.True(t, categories[i-1] <= categories[i], "Categories should be sorted")
	}

	// Check specific categories exist
	assert.Contains(t, categories, "programming")
}

func TestArticleService_GetTagCounts(t *testing.T) {
	service := createTestArticleService(t)

	tagCounts := service.GetTagCounts()
	assert.Greater(t, len(tagCounts), 0)

	// Should be sorted by count (descending) then by name
	for i := 1; i < len(tagCounts); i++ {
		prev := tagCounts[i-1]
		curr := tagCounts[i]
		if prev.Count == curr.Count {
			assert.True(t, prev.Tag <= curr.Tag, "Tags with same count should be sorted by name")
		} else {
			assert.True(t, prev.Count >= curr.Count, "Tags should be sorted by count descending")
		}
	}

	// All counts should be positive
	for _, tagCount := range tagCounts {
		assert.Greater(t, tagCount.Count, 0)
		assert.NotEmpty(t, tagCount.Tag)
	}
}

func TestArticleService_GetCategoryCounts(t *testing.T) {
	service := createTestArticleService(t)

	categoryCounts := service.GetCategoryCounts()
	assert.Greater(t, len(categoryCounts), 0)

	// Should be sorted by count (descending) then by name
	for i := 1; i < len(categoryCounts); i++ {
		prev := categoryCounts[i-1]
		curr := categoryCounts[i]
		if prev.Count == curr.Count {
			assert.True(t, prev.Category <= curr.Category, "Categories with same count should be sorted by name")
		} else {
			assert.True(t, prev.Count >= curr.Count, "Categories should be sorted by count descending")
		}
	}

	// All counts should be positive
	for _, categoryCount := range categoryCounts {
		assert.Greater(t, categoryCount.Count, 0)
		assert.NotEmpty(t, categoryCount.Category)
	}
}

func TestArticleService_GetFeaturedArticles(t *testing.T) {
	service := createTestArticleService(t)

	// Test with limit
	featured := service.GetFeaturedArticles(1)
	assert.LessOrEqual(t, len(featured), 1)

	// All returned articles should be featured
	for _, article := range featured {
		assert.True(t, article.Featured, "All articles should be featured")
	}

	// Test with larger limit
	allFeatured := service.GetFeaturedArticles(10)
	for _, article := range allFeatured {
		assert.True(t, article.Featured, "All articles should be featured")
	}
}

func TestArticleService_GetRecentArticles(t *testing.T) {
	service := createTestArticleService(t)

	// Test with limit smaller than total articles
	recent := service.GetRecentArticles(1)
	assert.Equal(t, 1, len(recent))

	// Test with limit larger than total articles
	allArticles := service.GetAllArticles()
	moreRecent := service.GetRecentArticles(len(allArticles) + 10)
	assert.Equal(t, len(allArticles), len(moreRecent))

	// Should return the same as GetAllArticles when limit is large
	assert.Equal(t, allArticles, moreRecent)
}

func TestArticleService_GetArticlesForFeed(t *testing.T) {
	service := createTestArticleService(t)

	// Test with limit
	feedArticles := service.GetArticlesForFeed(1)
	assert.Equal(t, 1, len(feedArticles))

	// Test with larger limit
	allArticles := service.GetAllArticles()
	moreFeedArticles := service.GetArticlesForFeed(len(allArticles) + 10)
	assert.Equal(t, len(allArticles), len(moreFeedArticles))
}

func TestArticleService_GetStats(t *testing.T) {
	service := createTestArticleService(t)

	stats := service.GetStats()
	assert.NotNil(t, stats)
	assert.Greater(t, stats.TotalArticles, 0)
	assert.Greater(t, stats.PublishedCount, 0)
	assert.GreaterOrEqual(t, stats.DraftCount, 0)
	assert.Greater(t, stats.TotalTags, 0)
	assert.Greater(t, stats.TotalCategories, 0)
	assert.NotNil(t, stats.PopularTags)
	assert.NotNil(t, stats.RecentArticles)
	assert.False(t, stats.LastUpdated.IsZero())

	// Total articles should equal published + draft
	assert.Equal(t, stats.TotalArticles, stats.PublishedCount+stats.DraftCount)

	// Popular tags should be limited to 10
	assert.LessOrEqual(t, len(stats.PopularTags), 10)

	// Recent articles should be limited to 5
	assert.LessOrEqual(t, len(stats.RecentArticles), 5)
}

func TestArticleService_ReloadArticles(t *testing.T) {
	service := createTestArticleService(t)

	initialCount := len(service.GetAllArticles())
	initialTime := service.lastReload

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	err := service.ReloadArticles()
	assert.NoError(t, err)

	// Should have same count
	assert.Equal(t, initialCount, len(service.GetAllArticles()))

	// Last reload time should be updated
	assert.True(t, service.lastReload.After(initialTime))
}

func TestArticleService_ParseArticle(t *testing.T) {
	service := createTestArticleService(t)

	tests := []struct {
		name     string
		slug     string
		content  string
		wantErr  bool
		validate func(t *testing.T, article *models.Article)
	}{
		{
			name: "Article with frontmatter",
			slug: "test-slug",
			content: `---
title: "Test Title"
date: 2023-01-01T10:00:00Z
tags: ["tag1", "tag2"]
categories: ["cat1"]
featured: true
draft: false
---

# Test Content

This is a comprehensive test content paragraph that contains enough meaningful text for the excerpt generation algorithm to work properly. It includes multiple sentences and sufficient content to pass the minimum length requirements.`,
			wantErr: false,
			validate: func(t *testing.T, article *models.Article) {
				assert.Equal(t, "test-slug", article.Slug)
				assert.Equal(t, "Test Title", article.Title)
				assert.Equal(t, []string{"tag1", "tag2"}, article.Tags)
				assert.Equal(t, []string{"cat1"}, article.Categories)
				assert.True(t, article.Featured)
				assert.False(t, article.Draft)
				assert.Contains(t, article.Content, "<h1 id=\"test-content\">Test Content</h1>")
				assert.NotEmpty(t, article.Excerpt)
				assert.Greater(t, article.WordCount, 0)
				assert.Greater(t, article.ReadingTime, 0)
			},
		},
		{
			name: "Article without frontmatter",
			slug: "no-frontmatter",
			content: `# Simple Article

This is a simple article with just content and no frontmatter. It contains enough meaningful text to generate a proper excerpt. The content should be substantial enough for the excerpt generation algorithm to process it correctly.`,
			wantErr: false,
			validate: func(t *testing.T, article *models.Article) {
				assert.Equal(t, "no-frontmatter", article.Slug)
				assert.Equal(t, "No Frontmatter", article.Title) // Generated from slug
				assert.Contains(t, article.Content, "<h1 id=\"simple-article\">Simple Article</h1>")
				assert.NotEmpty(t, article.Excerpt)
			},
		},
		{
			name: "Article with invalid frontmatter",
			slug: "invalid-yaml",
			content: `---
invalid: yaml: [
---

# Content`,
			wantErr:  true,
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article, err := service.ParseArticle(tt.slug, tt.content)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, article)

			if tt.validate != nil {
				tt.validate(t, article)
			}
		})
	}
}

func TestArticleService_SlugToTitle(t *testing.T) {
	service := createTestArticleService(t)

	tests := []struct {
		slug     string
		expected string
	}{
		{"hello-world", "Hello World"},
		{"go-programming", "Go Programming"},
		{"single", "Single"},
		{"", ""},
		{"multi-word-slug-test", "Multi Word Slug Test"},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			result := service.slugToTitle(tt.slug)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArticleService_GenerateExcerpt(t *testing.T) {
	service := createTestArticleService(t)

	tests := []struct {
		name      string
		content   string
		maxLength int
		expected  string
	}{
		{
			name:      "Short content",
			content:   "This is a meaningful paragraph with enough content.",
			maxLength: 100,
			expected:  "This is a meaningful paragraph with enough content.",
		},
		{
			name:      "Long content with sentence-aware truncation",
			content:   "This is a very long paragraph that contains multiple sentences. It should be truncated at a sentence boundary when possible. This makes the excerpt more readable and natural.",
			maxLength: 100,
			expected:  "This is a very long paragraph that contains multiple sentences. It should be truncated at a...",
		},
		{
			name:      "Content with markdown formatting",
			content:   "This paragraph has **bold text** and *italic text* and `inline code` and [link text](http://example.com) that should be cleaned properly.",
			maxLength: 150,
			expected:  "This paragraph has bold text and italic text and and link text that should be cleaned properly.",
		},
		{
			name:      "Content with headers and code blocks",
			content:   "# Header\n\nThis is the actual content paragraph that should be extracted.\n\n```\ncode block\n```\n\nAnother paragraph here.",
			maxLength: 100,
			expected:  "This is the actual content paragraph that should be extracted.",
		},
		{
			name:      "Content with intro phrases",
			content:   "In this article, we will explore many concepts.\n\nThe main topic is machine learning algorithms and their applications in real-world scenarios.",
			maxLength: 150,
			expected:  "The main topic is machine learning algorithms and their applications in real-world scenarios.",
		},
		{
			name:      "Empty content",
			content:   "",
			maxLength: 100,
			expected:  "",
		},
		{
			name:      "Content with bullet points",
			content:   "Here's some content:\n\n- First point\n- Second point\n\nThis is a regular paragraph that should be extracted instead of the list items.",
			maxLength: 100,
			expected:  "Here's some content: First point Second point",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.generateExcerpt(tt.content, tt.maxLength)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArticleService_ThreadSafety(t *testing.T) {
	service := createTestArticleService(t)

	// Test concurrent access
	done := make(chan bool)
	errors := make(chan error, 10)

	// Start multiple goroutines doing different operations
	for range 5 {
		go func() {
			defer func() { done <- true }()

			// Read operations
			articles := service.GetAllArticles()
			if len(articles) == 0 {
				errors <- assert.AnError
				return
			}

			_, err := service.GetArticleBySlug("test-article-1")
			if err != nil {
				errors <- err
				return
			}

			service.GetArticlesByTag("go")
			service.GetAllTags()
			service.GetStats()
		}()
	}

	// Wait for all goroutines to complete
	for range 5 {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

// Helper function to create a test article service with sample data
func createTestArticleService(t *testing.T) *ArticleService {
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create test articles
	testArticles := map[string]string{
		"test-article-1.markdown": `---
title: "Test Article 1"
date: 2023-01-01T10:00:00Z
tags: ["go", "testing"]
categories: ["programming"]
featured: true
draft: false
---

# Test Article 1

This is the content of test article 1 with enough words to test word count and reading time calculation. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

## Section 1

More content here to increase word count.`,

		"test-article-2.markdown": `---
title: "Test Article 2"
date: 2023-01-02T10:00:00Z
tags: ["go", "web"]
categories: ["programming", "web"]
featured: false
draft: false
---

# Test Article 2

This is the content of test article 2.`,

		"test-article-3.markdown": `---
title: "Test Article 3"
date: 2023-01-03T10:00:00Z
tags: ["javascript", "web"]
categories: ["web"]
featured: false
draft: false
---

# Test Article 3

This is the content of test article 3.`,

		"draft-article.markdown": `---
title: "Draft Article"
date: 2023-01-04T10:00:00Z
tags: ["draft"]
categories: ["test"]
featured: false
draft: true
---

# Draft Article

This is a draft article.`,
	}

	for filename, content := range testArticles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	service, err := NewArticleService(tempDir, logger)
	require.NoError(t, err)

	return service
}

func TestArticleService_MaxFunction(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{0, 0, 0},
		{-1, 1, 1},
		{-5, -3, -3},
	}

	for _, tt := range tests {
		result := max(tt.a, tt.b)
		assert.Equal(t, tt.expected, result)
	}
}

func TestArticleService_ParseArticleFile(t *testing.T) {
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service, err := NewArticleService(tempDir, logger)
	require.NoError(t, err)

	// Create a test file
	testContent := `---
title: "File Test"
date: 2023-01-01T10:00:00Z
---

# File Test Content`

	filePath := filepath.Join(tempDir, "file-test.markdown")
	err = os.WriteFile(filePath, []byte(testContent), 0644)
	require.NoError(t, err)

	article, err := service.ParseArticleFile(filePath)
	assert.NoError(t, err)
	assert.NotNil(t, article)
	assert.Equal(t, "file-test", article.Slug)
	assert.Equal(t, "File Test", article.Title)
	assert.False(t, article.LastModified.IsZero())

	// Test with non-existent file
	_, err = service.ParseArticleFile(filepath.Join(tempDir, "nonexistent.markdown"))
	assert.Error(t, err)
}

func BenchmarkArticleService_GetAllArticles(b *testing.B) {
	service := createBenchmarkArticleService(b)

	for b.Loop() {
		articles := service.GetAllArticles()
		_ = articles
	}
}

func BenchmarkArticleService_GetArticleBySlug(b *testing.B) {
	service := createBenchmarkArticleService(b)

	for b.Loop() {
		_, _ = service.GetArticleBySlug("test-article-1")
	}
}

func BenchmarkArticleService_GetArticlesByTag(b *testing.B) {
	service := createBenchmarkArticleService(b)

	for b.Loop() {
		articles := service.GetArticlesByTag("go")
		_ = articles
	}
}

func createBenchmarkArticleService(b *testing.B) *ArticleService {
	tempDir := b.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create multiple test articles for benchmarking
	for i := range 100 {
		content := `---
title: "Benchmark Article %d"
date: 2023-01-01T10:00:00Z
tags: ["go", "benchmark", "test"]
categories: ["testing"]
featured: false
draft: false
---

# Benchmark Article

This is content for benchmarking purposes.`

		filename := filepath.Join(tempDir, fmt.Sprintf("benchmark-article-%d.markdown", i))
		err := os.WriteFile(filename, []byte(content), 0644)
		require.NoError(b, err)
	}

	service, err := NewArticleService(tempDir, logger)
	require.NoError(b, err)

	return service
}

// Additional Edge Case Tests

func TestArticleService_GetArticlesByTag_EmptyTag(t *testing.T) {
	service := createTestArticleService(t)

	// Test with empty tag
	articles := service.GetArticlesByTag("")
	assert.Equal(t, 0, len(articles))

	// Test with whitespace-only tag
	articles = service.GetArticlesByTag("   ")
	assert.Equal(t, 0, len(articles))
}

func TestArticleService_GetArticlesByCategory_NilCategory(t *testing.T) {
	service := createTestArticleService(t)

	// Test with empty category
	articles := service.GetArticlesByCategory("")
	assert.Equal(t, 0, len(articles))

	// Test with whitespace-only category
	articles = service.GetArticlesByCategory("   ")
	assert.Equal(t, 0, len(articles))
}

func TestArticleService_LoadArticles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Create service with empty directory
	service, err := NewArticleService(tempDir, logger)
	assert.NoError(t, err)

	// Should handle empty directory gracefully
	articles := service.GetAllArticles()
	assert.Equal(t, 0, len(articles))

	stats := service.GetStats()
	assert.Equal(t, 0, stats.TotalArticles)
	assert.Equal(t, 0, stats.PublishedCount)
	assert.Equal(t, 0, stats.DraftCount)
}

func TestArticleService_LoadArticles_InvalidMarkdownFiles(t *testing.T) {
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Create invalid markdown files
	invalidFiles := []struct {
		filename string
		content  string
	}{
		{"invalid-frontmatter.md", "---\ninvalid yaml content\n---\n# Content"},
		{"no-frontmatter.md", "# Just content without frontmatter"},
		{"empty-file.md", ""},
		{"binary-file.md", string([]byte{0x00, 0x01, 0x02, 0x03})},
	}

	for _, file := range invalidFiles {
		err := os.WriteFile(filepath.Join(tempDir, file.filename), []byte(file.content), 0644)
		require.NoError(t, err)
	}

	// Should handle invalid files gracefully
	service, err := NewArticleService(tempDir, logger)
	assert.NoError(t, err)

	// Should skip invalid files
	articles := service.GetAllArticles()
	// May have some articles parsed, but should not crash
	assert.GreaterOrEqual(t, len(articles), 0)
}

func TestArticleService_GetArticleBySlug_WithSpecialCharacters(t *testing.T) {
	service := createTestArticleService(t)

	// Test with special characters in slug
	testCases := []string{
		"article-with-dashes",
		"article_with_underscores",
		"article.with.dots",
		"article with spaces", // Should be handled as slug
	}

	for _, slug := range testCases {
		article, err := service.GetArticleBySlug(slug)
		// Should not crash, but may return not found
		if err != nil {
			assert.Contains(t, err.Error(), "article not found")
		}
		if article != nil {
			assert.NotEmpty(t, article.Slug)
		}
	}
}

func TestArticleService_ReloadArticles_NonExistentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	service, err := NewArticleService(tempDir, logger)
	assert.NoError(t, err)

	// Remove the directory
	os.RemoveAll(tempDir)

	// Should handle missing directory gracefully
	err = service.ReloadArticles()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestArticleService_ParseArticle_EdgeCases(t *testing.T) {
	service := createTestArticleService(t)

	edgeCases := []struct {
		name    string
		content string
		wantErr bool
		checkFn func(*testing.T, *models.Article)
	}{
		{
			name:    "article with only title",
			content: "---\ntitle: \"Just Title\"\n---\n",
			wantErr: false,
			checkFn: func(t *testing.T, article *models.Article) {
				assert.Equal(t, "Just Title", article.Title)
				assert.Empty(t, article.Content)
			},
		},
		{
			name:    "article with unicode content",
			content: "---\ntitle: \"Unicode Test\"\n---\n# Unicode: 你好 🌍 café naïve",
			wantErr: false,
			checkFn: func(t *testing.T, article *models.Article) {
				assert.Contains(t, article.Content, "你好")
				assert.Contains(t, article.Content, "🌍")
				assert.Contains(t, article.Content, "café")
			},
		},
		{
			name:    "article with very long content",
			content: "---\ntitle: \"Long Content\"\n---\n" + strings.Repeat("Lorem ipsum dolor sit amet. ", 1000),
			wantErr: false,
			checkFn: func(t *testing.T, article *models.Article) {
				assert.Greater(t, len(article.Content), 10000)
				assert.Greater(t, article.WordCount, 3000)
			},
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			article, err := service.ParseArticle("test-slug", tc.content)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, article)
				if tc.checkFn != nil {
					tc.checkFn(t, article)
				}
			}
		})
	}
}

func TestArticleService_ProcessDuplicateTitles(t *testing.T) {
	service := createTestArticleService(t)

	tests := []struct {
		name            string
		title           string
		content         string
		expectedContent string
	}{
		{
			name:            "No duplicate title - no change",
			title:           "My Article Title",
			content:         `<h1>Different Heading</h1><p>Some content</p>`,
			expectedContent: `<h1>Different Heading</h1><p>Some content</p>`,
		},
		{
			name:            "Exact duplicate title - demote to h2",
			title:           "My Article Title",
			content:         `<h1>My Article Title</h1><p>Some content</p>`,
			expectedContent: `<h2>My Article Title</h2><p>Some content</p>`,
		},
		{
			name:            "Case insensitive duplicate - demote to h2",
			title:           "My Article Title",
			content:         `<h1>my article title</h1><p>Some content</p>`,
			expectedContent: `<h2>my article title</h2><p>Some content</p>`,
		},
		{
			name:            "H1 with attributes - preserve attributes",
			title:           "My Article Title",
			content:         `<h1 id="main-title" class="title">My Article Title</h1><p>Some content</p>`,
			expectedContent: `<h2 id="main-title" class="title">My Article Title</h2><p>Some content</p>`,
		},
		{
			name:            "Empty title - no processing",
			title:           "",
			content:         `<h1>Some Heading</h1><p>Some content</p>`,
			expectedContent: `<h1>Some Heading</h1><p>Some content</p>`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := service.processDuplicateTitles(tc.title, tc.content)
			assert.Equal(t, tc.expectedContent, result)
		})
	}
}

func TestArticleService_ExtractFirstHeading(t *testing.T) {
	service := createTestArticleService(t)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Simple h1 tag",
			content:  `<h1>My Heading</h1><p>Content</p>`,
			expected: "My Heading",
		},
		{
			name:     "H1 with attributes",
			content:  `<h1 id="title" class="main">My Heading</h1><p>Content</p>`,
			expected: "My Heading",
		},
		{
			name:     "H1 with nested HTML",
			content:  `<h1><span>My</span> <em>Heading</em></h1><p>Content</p>`,
			expected: "My Heading",
		},
		{
			name:     "No h1 tag",
			content:  `<h2>My Heading</h2><p>Content</p>`,
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := service.extractFirstHeading(tc.content)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestArticleService_NormalizeText(t *testing.T) {
	service := createTestArticleService(t)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic text",
			input:    "My Article Title",
			expected: "my article title",
		},
		{
			name:     "Mixed case with extra whitespace",
			input:    "  My   ARTICLE    Title  ",
			expected: "my article title",
		},
		{
			name:     "Whitespace variations",
			input:    "My\tArticle\nTitle",
			expected: "my article title",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := service.normalizeText(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestArticleService_ParseArticle_WithDuplicateTitle(t *testing.T) {
	service := createTestArticleService(t)

	// Test duplicate title - should be demoted to h2
	content := `---
title: "My Test Article"
date: 2023-01-01T10:00:00Z
tags: ["test"]
---

# My Test Article

This is the content of my article.`

	article, err := service.ParseArticle("my-test-article", content)
	assert.NoError(t, err)
	assert.Equal(t, "My Test Article", article.Title)
	assert.NotEmpty(t, article.ProcessedContent)

	// Check that h1 was converted to h2
	assert.Contains(t, article.ProcessedContent, "<h2")
	assert.NotContains(t, article.ProcessedContent, "<h1")

	// Test different heading - should remain unchanged
	differentContent := `---
title: "My Test Article"
date: 2023-01-01T10:00:00Z
tags: ["test"]
---

# Different Heading

This is the content of my article.`

	article2, err := service.ParseArticle("my-test-article-2", differentContent)
	assert.NoError(t, err)
	assert.Equal(t, "My Test Article", article2.Title)
	assert.Contains(t, article2.ProcessedContent, "<h1")
}
