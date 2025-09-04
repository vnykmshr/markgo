package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockProcessor implements ArticleProcessor for testing
type mockProcessor struct{}

func (m *mockProcessor) ProcessMarkdown(content string) (string, error) {
	return content, nil
}

func (m *mockProcessor) GenerateExcerpt(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "..."
}

func (m *mockProcessor) ProcessDuplicateTitles(title, htmlContent string) string {
	return htmlContent
}

func TestArticleToListView(t *testing.T) {
	now := time.Now()
	article := &Article{
		Slug:        "test-article",
		Title:       "Test Article",
		Description: "A test article description",
		Date:        now,
		Tags:        []string{"golang", "testing"},
		Categories:  []string{"programming", "tutorial"},
		Draft:       false,
		Featured:    true,
		Author:      "Test Author",
		Content:     "This is the article content",
		ReadingTime: 5,
		WordCount:   100,
	}
	article.SetProcessor(&mockProcessor{})

	listView := article.ToListView()

	assert.Equal(t, article.Slug, listView.Slug)
	assert.Equal(t, article.Title, listView.Title)
	assert.Equal(t, article.Description, listView.Description)
	assert.Equal(t, article.Date, listView.Date)
	assert.Equal(t, article.Tags, listView.Tags)
	assert.Equal(t, article.Categories, listView.Categories)
	assert.Equal(t, article.GetExcerpt(), listView.Excerpt)
	assert.Equal(t, article.ReadingTime, listView.ReadingTime)
	assert.Equal(t, article.Featured, listView.Featured)

	// Ensure the conversion worked correctly
	assert.NotEmpty(t, listView.Slug) // Should have the slug from the original article
}

func TestNewPagination(t *testing.T) {
	testCases := []struct {
		name                                     string
		currentPage, totalItems, itemsPerPage    int
		expectedCurrentPage, expectedTotalPages  int
		expectedTotalItems, expectedItemsPerPage int
		expectedHasPrevious, expectedHasNext     bool
		expectedPreviousPage, expectedNextPage   int
	}{
		{
			name:                 "First page with multiple pages",
			currentPage:          1,
			totalItems:           25,
			itemsPerPage:         10,
			expectedCurrentPage:  1,
			expectedTotalPages:   3,
			expectedTotalItems:   25,
			expectedItemsPerPage: 10,
			expectedHasPrevious:  false,
			expectedHasNext:      true,
			expectedPreviousPage: 1,
			expectedNextPage:     2,
		},
		{
			name:                 "Middle page",
			currentPage:          2,
			totalItems:           25,
			itemsPerPage:         10,
			expectedCurrentPage:  2,
			expectedTotalPages:   3,
			expectedTotalItems:   25,
			expectedItemsPerPage: 10,
			expectedHasPrevious:  true,
			expectedHasNext:      true,
			expectedPreviousPage: 1,
			expectedNextPage:     3,
		},
		{
			name:                 "Last page",
			currentPage:          3,
			totalItems:           25,
			itemsPerPage:         10,
			expectedCurrentPage:  3,
			expectedTotalPages:   3,
			expectedTotalItems:   25,
			expectedItemsPerPage: 10,
			expectedHasPrevious:  true,
			expectedHasNext:      false,
			expectedPreviousPage: 2,
			expectedNextPage:     3,
		},
		{
			name:                 "Single page",
			currentPage:          1,
			totalItems:           5,
			itemsPerPage:         10,
			expectedCurrentPage:  1,
			expectedTotalPages:   1,
			expectedTotalItems:   5,
			expectedItemsPerPage: 10,
			expectedHasPrevious:  false,
			expectedHasNext:      false,
			expectedPreviousPage: 1,
			expectedNextPage:     1,
		},
		{
			name:                 "Empty items",
			currentPage:          1,
			totalItems:           0,
			itemsPerPage:         10,
			expectedCurrentPage:  1,
			expectedTotalPages:   1,
			expectedTotalItems:   0,
			expectedItemsPerPage: 10,
			expectedHasPrevious:  false,
			expectedHasNext:      false,
			expectedPreviousPage: 1,
			expectedNextPage:     1,
		},
		{
			name:                 "Exact division",
			currentPage:          2,
			totalItems:           20,
			itemsPerPage:         10,
			expectedCurrentPage:  2,
			expectedTotalPages:   2,
			expectedTotalItems:   20,
			expectedItemsPerPage: 10,
			expectedHasPrevious:  true,
			expectedHasNext:      false,
			expectedPreviousPage: 1,
			expectedNextPage:     2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pagination := NewPagination(tc.currentPage, tc.totalItems, tc.itemsPerPage)

			assert.Equal(t, tc.expectedCurrentPage, pagination.CurrentPage)
			assert.Equal(t, tc.expectedTotalPages, pagination.TotalPages)
			assert.Equal(t, tc.expectedTotalItems, pagination.TotalItems)
			assert.Equal(t, tc.expectedItemsPerPage, pagination.ItemsPerPage)
			assert.Equal(t, tc.expectedHasPrevious, pagination.HasPrevious)
			assert.Equal(t, tc.expectedHasNext, pagination.HasNext)
			assert.Equal(t, tc.expectedPreviousPage, pagination.PreviousPage)
			assert.Equal(t, tc.expectedNextPage, pagination.NextPage)
		})
	}
}

func TestPaginationEdgeCases(t *testing.T) {
	// Test with current page beyond total pages
	pagination := NewPagination(5, 10, 5)
	assert.Equal(t, 5, pagination.CurrentPage)
	assert.Equal(t, 2, pagination.TotalPages)
	assert.True(t, pagination.HasPrevious)
	assert.False(t, pagination.HasNext)
	assert.Equal(t, 2, pagination.NextPage) // Should be clamped to total pages
}

// Benchmark tests
func BenchmarkArticleToListView(b *testing.B) {
	article := &Article{
		Slug:        "benchmark-article",
		Title:       "Benchmark Article",
		Description: "A benchmark article description",
		Date:        time.Now(),
		Tags:        []string{"golang", "benchmark"},
		Categories:  []string{"performance"},
		ReadingTime: 5,
		Featured:    true,
	}

	for b.Loop() {
		article.ToListView()
	}
}

func BenchmarkNewPagination(b *testing.B) {

	for b.Loop() {
		NewPagination(2, 100, 10)
	}
}
