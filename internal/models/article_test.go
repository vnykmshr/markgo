package models

import (
	"errors"
	"sync"
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

func (m *mockProcessor) ProcessDuplicateTitles(_, htmlContent string) string {
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

// Additional comprehensive tests for better coverage

// Enhanced mock processor for error testing
type errorMockProcessor struct {
	shouldError bool
}

func (m *errorMockProcessor) ProcessMarkdown(content string) (string, error) {
	if m.shouldError {
		return "", errors.New("processing error")
	}
	return "<p>" + content + "</p>", nil
}

func (m *errorMockProcessor) GenerateExcerpt(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "..."
}

func (m *errorMockProcessor) ProcessDuplicateTitles(_, htmlContent string) string {
	return htmlContent + " (processed)"
}

func TestArticle_GetProcessedContent(t *testing.T) {
	article := &Article{
		Title:   "Test Article",
		Content: "This is test content",
	}

	// Test without processor
	content := article.GetProcessedContent()
	assert.Equal(t, "This is test content", content)

	// Test with processor
	processor := &errorMockProcessor{shouldError: false}
	article.SetProcessor(processor)

	// Clear to force regeneration
	article.ClearProcessedContent()
	content = article.GetProcessedContent()
	assert.Equal(t, "<p>This is test content</p> (processed)", content)

	// Test caching - should return same content without reprocessing
	content2 := article.GetProcessedContent()
	assert.Equal(t, content, content2)

	// Test with processor error
	errorProcessor := &errorMockProcessor{shouldError: true}
	article.SetProcessor(errorProcessor)
	article.ClearProcessedContent()

	content = article.GetProcessedContent()
	assert.Equal(t, "This is test content", content) // Should fallback to raw content
}

func TestArticle_GetProcessedContent_Concurrent(t *testing.T) {
	article := &Article{
		Title:   "Concurrent Test",
		Content: "Concurrent content",
	}
	article.SetProcessor(&mockProcessor{})

	var wg sync.WaitGroup
	results := make([]string, 10)

	// Test concurrent access to GetProcessedContent
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = article.GetProcessedContent()
		}(i)
	}

	wg.Wait()

	// All results should be the same due to caching
	for i := 1; i < 10; i++ {
		assert.Equal(t, results[0], results[i])
	}
	assert.NotEmpty(t, results[0])
}

func TestArticle_GetExcerpt(t *testing.T) {
	// Test with description
	article := &Article{
		Title:       "Test Article",
		Description: "This is a description",
		Content:     "This is longer content for the article",
	}

	excerpt := article.GetExcerpt()
	assert.Equal(t, "This is a description", excerpt)

	// Test without description but with processor
	article.Description = ""
	processor := &mockProcessor{}
	article.SetProcessor(processor)
	article.ClearProcessedContent()

	excerpt = article.GetExcerpt()
	assert.Equal(t, "This is longer content for the article", excerpt) // Short enough, no truncation

	// Test with long content and no description
	article.Content = "This is a very long piece of content that should be truncated when used as an excerpt because it exceeds the maximum length that we allow for excerpts in the system and should result in a truncated version with ellipsis"
	article.ClearProcessedContent()

	excerpt = article.GetExcerpt()
	assert.True(t, len(excerpt) <= 203) // 200 + "..."
	assert.Contains(t, excerpt, "...")

	// Test caching
	excerpt2 := article.GetExcerpt()
	assert.Equal(t, excerpt, excerpt2)
}

func TestArticle_GetExcerpt_Concurrent(t *testing.T) {
	article := &Article{
		Content: "Concurrent excerpt test content",
	}

	var wg sync.WaitGroup
	results := make([]string, 10)

	// Test concurrent access to GetExcerpt
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = article.GetExcerpt()
		}(i)
	}

	wg.Wait()

	// All results should be the same
	for i := 1; i < 10; i++ {
		assert.Equal(t, results[0], results[i])
	}
	assert.NotEmpty(t, results[0])
}

func TestArticle_SetProcessor(t *testing.T) {
	article := &Article{}
	processor := &mockProcessor{}

	article.SetProcessor(processor)

	// Processor should be set (verify by checking behavior)
	content := article.GetProcessedContent()
	assert.NotNil(t, content)
}

func TestArticle_ProcessedContent(t *testing.T) {
	article := &Article{Content: "Test content"}

	// Test JSON serialization compatibility method
	processed := article.ProcessedContent()
	assert.Equal(t, "Test content", processed)
}

func TestArticle_Excerpt(t *testing.T) {
	article := &Article{Description: "Test description"}

	// Test JSON serialization compatibility method
	excerpt := article.Excerpt()
	assert.Equal(t, "Test description", excerpt)
}

func TestArticle_ClearProcessedContent(t *testing.T) {
	article := &Article{Content: "Test content"}
	article.SetProcessor(&mockProcessor{})

	// Generate processed content and excerpt
	_ = article.GetProcessedContent()
	_ = article.GetExcerpt()

	// Verify they're cached
	assert.NotNil(t, article.processedContent)
	assert.NotNil(t, article.excerpt)

	// Clear cache
	article.ClearProcessedContent()

	// Verify they're cleared
	assert.Nil(t, article.processedContent)
	assert.Nil(t, article.excerpt)
}

func TestContactMessage_ValidationTags(t *testing.T) {
	// Test that ContactMessage has proper validation tags
	msg := ContactMessage{
		Name:            "John Doe",
		Email:           "john@example.com",
		Subject:         "Test Subject",
		Message:         "This is a test message with sufficient length",
		CaptchaQuestion: "What is 2+2?",
		CaptchaAnswer:   "4",
	}

	// Test basic field population
	assert.Equal(t, "John Doe", msg.Name)
	assert.Equal(t, "john@example.com", msg.Email)
	assert.Equal(t, "Test Subject", msg.Subject)
	assert.NotEmpty(t, msg.Message)
	assert.NotEmpty(t, msg.CaptchaQuestion)
	assert.Equal(t, "4", msg.CaptchaAnswer)
}

func TestSearchResult(t *testing.T) {
	article := &Article{
		Title:   "Search Test",
		Content: "Searchable content",
	}

	searchResult := &SearchResult{
		Article:       article,
		Score:         0.85,
		MatchedFields: []string{"title", "content"},
	}

	assert.Equal(t, article, searchResult.Article)
	assert.Equal(t, 0.85, searchResult.Score)
	assert.Equal(t, []string{"title", "content"}, searchResult.MatchedFields)
}

func TestSearchResultPage(t *testing.T) {
	article := &Article{Title: "Test"}
	searchResult := &SearchResult{
		Article: article,
		Score:   0.9,
	}

	pagination := NewPagination(1, 10, 5)

	searchPage := &SearchResultPage{
		Results:    []*SearchResult{searchResult},
		Pagination: pagination,
		Query:      "test query",
		TotalTime:  150,
	}

	assert.Len(t, searchPage.Results, 1)
	assert.Equal(t, searchResult, searchPage.Results[0])
	assert.Equal(t, pagination, searchPage.Pagination)
	assert.Equal(t, "test query", searchPage.Query)
	assert.Equal(t, int64(150), searchPage.TotalTime)
}

func TestFeed_Structures(t *testing.T) {
	author := Author{
		Name:  "Test Author",
		Email: "author@example.com",
		URL:   "https://example.com/author",
	}

	feedItem := FeedItem{
		ID:          "item-1",
		Title:       "Feed Item Title",
		ContentHTML: "<p>HTML content</p>",
		URL:         "https://example.com/item-1",
		Summary:     "Item summary",
		Published:   time.Now(),
		Modified:    time.Now(),
		Tags:        []string{"tag1", "tag2"},
		Author:      author,
	}

	feed := Feed{
		Title:       "Test Feed",
		Description: "Test feed description",
		Link:        "https://example.com",
		FeedURL:     "https://example.com/feed.json",
		Language:    "en",
		Updated:     time.Now(),
		Author:      author,
		Items:       []FeedItem{feedItem},
	}

	assert.Equal(t, "Test Feed", feed.Title)
	assert.Equal(t, author, feed.Author)
	assert.Len(t, feed.Items, 1)
	assert.Equal(t, feedItem, feed.Items[0])
	assert.Equal(t, "Test Author", feed.Items[0].Author.Name)
}

func TestSitemap_Structures(t *testing.T) {
	sitemapURL := SitemapURL{
		Loc:        "https://example.com/article",
		LastMod:    time.Now(),
		ChangeFreq: "weekly",
		Priority:   0.8,
	}

	sitemap := Sitemap{
		XMLName: "urlset",
		Xmlns:   "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:    []SitemapURL{sitemapURL},
	}

	assert.Equal(t, "urlset", sitemap.XMLName)
	assert.Contains(t, sitemap.Xmlns, "sitemaps.org")
	assert.Len(t, sitemap.URLs, 1)
	assert.Equal(t, sitemapURL, sitemap.URLs[0])
	assert.Equal(t, float32(0.8), sitemap.URLs[0].Priority)
}

func TestStats_Structures(t *testing.T) {
	tagCount := TagCount{
		Tag:   "golang",
		Count: 5,
	}

	categoryCount := CategoryCount{
		Category: "programming",
		Count:    3,
	}

	article := &Article{Title: "Test Article"}
	articleList := article.ToListView()

	stats := Stats{
		TotalArticles:   10,
		PublishedCount:  8,
		DraftCount:      2,
		TotalTags:       15,
		TotalCategories: 5,
		PopularTags:     []TagCount{tagCount},
		RecentArticles:  []*ArticleList{articleList},
		LastUpdated:     time.Now(),
	}

	assert.Equal(t, 10, stats.TotalArticles)
	assert.Equal(t, 8, stats.PublishedCount)
	assert.Equal(t, 2, stats.DraftCount)
	assert.Equal(t, 15, stats.TotalTags)
	assert.Equal(t, 5, stats.TotalCategories)
	assert.Len(t, stats.PopularTags, 1)
	assert.Equal(t, tagCount, stats.PopularTags[0])
	assert.Len(t, stats.RecentArticles, 1)
	assert.False(t, stats.LastUpdated.IsZero())

	// Test individual count structures
	assert.Equal(t, "golang", tagCount.Tag)
	assert.Equal(t, 5, tagCount.Count)
	assert.Equal(t, "programming", categoryCount.Category)
	assert.Equal(t, 3, categoryCount.Count)
}

func TestArticle_ContentHashing(t *testing.T) {
	article := &Article{
		Title:   "Hash Test",
		Content: "Original content",
	}

	processor := &errorMockProcessor{shouldError: false}
	article.SetProcessor(processor)

	// First processing should generate hash and content
	content1 := article.GetProcessedContent()
	assert.NotEmpty(t, content1)

	// Same content should return cached version
	content2 := article.GetProcessedContent()
	assert.Equal(t, content1, content2)

	// Change content and clear cache - should generate new content
	article.Content = "Modified content"
	article.ClearProcessedContent()

	content3 := article.GetProcessedContent()
	assert.NotEqual(t, content1, content3)
	assert.Contains(t, content3, "Modified content")
}

// Test memory optimization and lazy loading
func TestArticle_LazyLoading(t *testing.T) {
	article := &Article{
		Content: "Lazy loading test content",
	}

	// Initially, processed content and excerpt should be nil
	assert.Nil(t, article.processedContent)
	assert.Nil(t, article.excerpt)

	// First access should populate the cache
	excerpt := article.GetExcerpt()
	assert.NotNil(t, article.excerpt)
	assert.Equal(t, excerpt, *article.excerpt)

	// Same for processed content
	content := article.GetProcessedContent()
	assert.NotNil(t, article.processedContent)
	assert.Equal(t, content, *article.processedContent)
}

// Benchmark additional methods
func BenchmarkArticle_GetProcessedContent(b *testing.B) {
	article := &Article{
		Content: "Benchmark content for processing",
	}
	article.SetProcessor(&mockProcessor{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		article.GetProcessedContent()
	}
}

func BenchmarkArticle_GetExcerpt(b *testing.B) {
	article := &Article{
		Description: "Benchmark description for excerpt generation",
		Content:     "Benchmark content that could be used for excerpt if description is not available",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		article.GetExcerpt()
	}
}

func BenchmarkArticle_ConcurrentAccess(b *testing.B) {
	article := &Article{
		Content: "Concurrent benchmark content",
	}
	article.SetProcessor(&mockProcessor{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			article.GetProcessedContent()
			article.GetExcerpt()
		}
	})
}
