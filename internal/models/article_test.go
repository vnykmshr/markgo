package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestArticleToListView(t *testing.T) {
	now := time.Now()
	article := &Article{
		Slug:             "test-article",
		Title:            "Test Article",
		Description:      "A test article description",
		Date:             now,
		Tags:             []string{"golang", "testing"},
		Categories:       []string{"programming", "tutorial"},
		Draft:            false,
		Featured:         true,
		Author:           "Test Author",
		Content:          "This is the article content",
		ProcessedContent: "<p>This is the article content</p>",
		Excerpt:          "This is the excerpt",
		ReadingTime:      5,
		WordCount:        100,
	}

	listView := article.ToListView()

	assert.Equal(t, article.Slug, listView.Slug)
	assert.Equal(t, article.Title, listView.Title)
	assert.Equal(t, article.Description, listView.Description)
	assert.Equal(t, article.Date, listView.Date)
	assert.Equal(t, article.Tags, listView.Tags)
	assert.Equal(t, article.Categories, listView.Categories)
	assert.Equal(t, article.Excerpt, listView.Excerpt)
	assert.Equal(t, article.ReadingTime, listView.ReadingTime)
	assert.Equal(t, article.Featured, listView.Featured)
}

func TestArticleBasicFields(t *testing.T) {
	article := &Article{
		Slug:        "basic-article",
		Title:       "Basic Article",
		Description: "Basic description",
		Content:     "Basic content",
	}

	assert.Equal(t, "basic-article", article.Slug)
	assert.Equal(t, "Basic Article", article.Title)
	assert.Equal(t, "Basic description", article.Description)
	assert.Equal(t, "Basic content", article.Content)
}

func TestPagination(t *testing.T) {
	// Test with some items
	pagination := NewPagination(1, 50, 10)
	assert.Equal(t, 1, pagination.CurrentPage)
	assert.Equal(t, 5, pagination.TotalPages)
	assert.Equal(t, 50, pagination.TotalItems)
	assert.Equal(t, 10, pagination.ItemsPerPage)
	assert.False(t, pagination.HasPrevious)
	assert.True(t, pagination.HasNext)
	assert.Equal(t, 2, pagination.NextPage)

	// Test middle page
	pagination = NewPagination(3, 50, 10)
	assert.True(t, pagination.HasPrevious)
	assert.True(t, pagination.HasNext)
	assert.Equal(t, 2, pagination.PreviousPage)
	assert.Equal(t, 4, pagination.NextPage)

	// Test last page
	pagination = NewPagination(5, 50, 10)
	assert.True(t, pagination.HasPrevious)
	assert.False(t, pagination.HasNext)

	// Test empty result
	pagination = NewPagination(1, 0, 10)
	assert.Equal(t, 1, pagination.TotalPages)
	assert.False(t, pagination.HasNext)
	assert.False(t, pagination.HasPrevious)
}

func TestSearchResult(t *testing.T) {
	article := &Article{
		Slug:  "search-test",
		Title: "Search Test Article",
	}

	result := &SearchResult{
		Article:       article,
		Score:         0.85,
		MatchedFields: []string{"title", "content"},
	}

	assert.Equal(t, "search-test", result.Slug)
	assert.Equal(t, "Search Test Article", result.Title)
	assert.Equal(t, 0.85, result.Score)
	assert.Len(t, result.MatchedFields, 2)
}
