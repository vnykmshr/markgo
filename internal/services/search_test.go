package services

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vnykmshr/markgo/internal/models"
)

// mockProcessor implements ArticleProcessor for testing
type mockProcessor struct{}

func (m *mockProcessor) ProcessMarkdown(content string) (string, error) {
	// Simple mock - just return the content as-is (assume it's already HTML)
	return content, nil
}

func (m *mockProcessor) GenerateExcerpt(content string, maxLength int) string {
	// Simple excerpt generation for tests
	if len(content) <= maxLength {
		return content
	}
	// Strip HTML tags for excerpt
	cleaned := strings.ReplaceAll(content, "<p>", "")
	cleaned = strings.ReplaceAll(cleaned, "</p>", "")
	if len(cleaned) <= maxLength {
		return cleaned
	}
	return cleaned[:maxLength] + "..."
}

func (m *mockProcessor) ProcessDuplicateTitles(title, htmlContent string) string {
	// No processing for tests
	return htmlContent
}

func TestNewSearchService(t *testing.T) {
	service := NewSearchService()
	assert.NotNil(t, service)
}

func TestSearchService_Search(t *testing.T) {
	service := NewSearchService()
	articles := createTestArticles()

	// Test basic search for "golang"
	results := service.Search(articles, "golang", 10)
	assert.Len(t, results, 2)
	assert.Equal(t, "golang-tutorial", results[0].Article.Slug)
	// Check for various possible matched fields (title, description, content, tags, etc.)
	assert.NotEmpty(t, results[0].MatchedFields)

	// Test empty query
	results = service.Search(articles, "", 10)
	assert.Empty(t, results)

	// Test empty articles
	results = service.Search([]*models.Article{}, "golang", 10)
	assert.Empty(t, results)

	// Test whitespace query
	results = service.Search(articles, "   ", 10)
	assert.Empty(t, results)

	// Test multiple terms
	results = service.Search(articles, "web development", 10)
	assert.Len(t, results, 1)
	assert.Equal(t, "web-development", results[0].Article.Slug)

	// Test limit
	results = service.Search(articles, "tutorial", 1)
	assert.Len(t, results, 1)

	// Test case insensitive
	results = service.Search(articles, "GOLANG", 10)
	assert.Len(t, results, 2)

	// Test partial match
	results = service.Search(articles, "tutorial", 10)
	assert.Len(t, results, 1)
	assert.Equal(t, "golang-tutorial", results[0].Article.Slug)
}

func TestSearchService_SearchInTitle(t *testing.T) {
	service := NewSearchService()
	articles := createTestArticles()

	// Test title search for "golang"
	results := service.SearchInTitle(articles, "golang", 10)
	assert.Len(t, results, 2) // Two articles have "golang" in title
	assert.Contains(t, results[0].MatchedFields, "title")

	// Test exact title match
	results = service.SearchInTitle(articles, "golang tutorial complete", 10)
	assert.Len(t, results, 1)
	assert.Equal(t, "golang-tutorial", results[0].Article.Slug)
	assert.True(t, results[0].Score > 10.0)

	// Test prefix match
	results = service.SearchInTitle(articles, "golang", 10)
	if len(results) > 0 {
		assert.True(t, results[0].Score > 10.0) // Should have scoring bonus
	}

	// Test empty query
	results = service.SearchInTitle(articles, "", 10)
	assert.Empty(t, results)

	// Test no matches
	results = service.SearchInTitle(articles, "nonexistent", 10)
	assert.Empty(t, results)

	// Test limit
	results = service.SearchInTitle(articles, "golang", 1)
	assert.Len(t, results, 1)

	// Test case insensitive
	results = service.SearchInTitle(articles, "TUTORIAL", 10)
	assert.Len(t, results, 1)
}

func TestSearchService_SearchByTag(t *testing.T) {
	service := NewSearchService()
	articles := createTestArticles()

	// Test tag search
	results := service.SearchByTag(articles, "golang")
	assert.Len(t, results, 2)

	// Test case insensitive
	results = service.SearchByTag(articles, "GOLANG")
	assert.Len(t, results, 2)

	// Test non-existent tag
	results = service.SearchByTag(articles, "nonexistent")
	assert.Empty(t, results)

	// Test empty tag
	results = service.SearchByTag(articles, "")
	assert.Empty(t, results)

	// Test whitespace tag
	results = service.SearchByTag(articles, "   ")
	assert.Empty(t, results)

	// Test specific tag
	results = service.SearchByTag(articles, "tutorial")
	assert.Len(t, results, 1)
	assert.Equal(t, "golang-tutorial", results[0].Slug)
}

func TestSearchService_SearchByCategory(t *testing.T) {
	service := NewSearchService()
	articles := createTestArticles()

	// Test category search
	results := service.SearchByCategory(articles, "programming")
	assert.Len(t, results, 2)

	// Test case insensitive
	results = service.SearchByCategory(articles, "PROGRAMMING")
	assert.Len(t, results, 2)

	// Test non-existent category
	results = service.SearchByCategory(articles, "nonexistent")
	assert.Empty(t, results)

	// Test empty category
	results = service.SearchByCategory(articles, "")
	assert.Empty(t, results)

	// Test specific category
	results = service.SearchByCategory(articles, "web")
	assert.Len(t, results, 1)
	assert.Equal(t, "web-development", results[0].Slug)
}

func TestSearchService_GetSuggestions(t *testing.T) {
	service := NewSearchService()
	articles := createTestArticles()

	// Test suggestions from tags
	suggestions := service.GetSuggestions(articles, "go", 10)
	assert.Contains(t, suggestions, "golang")

	// Test suggestions from titles
	suggestions = service.GetSuggestions(articles, "tutorial", 10)
	assert.Contains(t, suggestions, "tutorial")

	// Test empty query
	suggestions = service.GetSuggestions(articles, "", 10)
	assert.Empty(t, suggestions)

	// Test no matches
	suggestions = service.GetSuggestions(articles, "xyz", 10)
	assert.Empty(t, suggestions)

	// Test limit
	suggestions = service.GetSuggestions(articles, "t", 1)
	assert.True(t, len(suggestions) <= 1)

	// Test case insensitive
	suggestions = service.GetSuggestions(articles, "GO", 10)
	assert.Contains(t, suggestions, "golang")

	// Test partial match
	suggestions = service.GetSuggestions(articles, "web", 10)
	assert.NotEmpty(t, suggestions)
}

func TestSearchService_CalculateScore(t *testing.T) {
	service := NewSearchService()
	article := &models.Article{
		Slug:        "test-article",
		Title:       "golang Programming Tutorial",
		Description: "Learn golang programming language",
		Tags:        []string{"golang", "tutorial"},
		Categories:  []string{"programming"},
		Content:     "<p>golang is a programming language developed by Google</p>",
		Featured:    true,
		Date:        time.Now(),
	}
	article.SetProcessor(&mockProcessor{})

	// Test title match with "golang" (3+ chars, not filtered)
	score, fields := service.calculateScore(article, []string{"golang"})
	assert.True(t, score > 0)
	assert.Contains(t, fields, "title")

	// Test multiple field matches
	score, fields = service.calculateScore(article, []string{"programming"})
	assert.True(t, score > 0)
	assert.Contains(t, fields, "title")
	assert.Contains(t, fields, "description")
	assert.Contains(t, fields, "categories")

	// Test multiple terms match
	score, _ = service.calculateScore(article, []string{"golang", "programming", "tutorial"})
	assert.True(t, score > 20) // Should have high score for multiple matches

	// Test no match
	score, fields = service.calculateScore(article, []string{"python"})
	assert.Equal(t, 0.0, score)
	assert.Empty(t, fields)

	// Test featured article bonus
	article.Featured = true
	score1, _ := service.calculateScore(article, []string{"golang"})
	article.Featured = false
	score2, _ := service.calculateScore(article, []string{"golang"})
	assert.True(t, score1 > score2)
}

func TestSearchService_Tokenize(t *testing.T) {
	service := NewSearchService()

	// Test basic tokenization
	tokens := service.tokenize("Hello world")
	assert.Equal(t, []string{"hello", "world"}, tokens)

	// Test punctuation removal
	tokens = service.tokenize("Hello, world!")
	assert.Equal(t, []string{"hello", "world"}, tokens)

	// Test stop word removal
	tokens = service.tokenize("the quick brown fox")
	assert.Equal(t, []string{"quick", "brown", "fox"}, tokens)

	// Test short word removal - words <= 2 chars and stop words removed
	tokens = service.tokenize("a go to app")
	assert.Equal(t, []string{"app"}, tokens) // Only "app" survives (3+ chars, not stop word)

	// Test mixed case
	tokens = service.tokenize("GoLang Programming")
	assert.Equal(t, []string{"golang", "programming"}, tokens)

	// Test special characters
	tokens = service.tokenize("web-development_tutorial")
	assert.Equal(t, []string{"web", "development", "tutorial"}, tokens)

	// Test empty string
	tokens = service.tokenize("")
	assert.Empty(t, tokens)

	// Test whitespace only
	tokens = service.tokenize("   ")
	assert.Empty(t, tokens)

	// Test numbers and words - numbers are filtered if short
	tokens = service.tokenize("golang 1.19 tutorial")
	assert.Equal(t, []string{"golang", "tutorial"}, tokens)
}

func TestSearchService_StripHTML(t *testing.T) {
	service := NewSearchService()

	// Test basic HTML stripping
	result := service.stripHTML("<p>Hello world</p>")
	assert.Equal(t, "Hello world", result)

	// Test nested tags
	result = service.stripHTML("<div><p>Hello <strong>world</strong></p></div>")
	assert.Equal(t, "Hello world", result)

	// Test self-closing tags
	result = service.stripHTML("Hello <br/> world")
	assert.Equal(t, "Hello  world", result)

	// Test no HTML
	result = service.stripHTML("Hello world")
	assert.Equal(t, "Hello world", result)

	// Test empty string
	result = service.stripHTML("")
	assert.Equal(t, "", result)

	// Test malformed HTML
	result = service.stripHTML("Hello <p world")
	assert.Equal(t, "Hello ", result)

	// Test attributes
	result = service.stripHTML("<p class='test'>Hello</p>")
	assert.Equal(t, "Hello", result)
}

func TestSearchService_IsStopWord(t *testing.T) {
	service := NewSearchService()

	// Test common stop words
	assert.True(t, service.isStopWord("the"))
	assert.True(t, service.isStopWord("and"))
	assert.True(t, service.isStopWord("or"))
	assert.True(t, service.isStopWord("is"))
	assert.True(t, service.isStopWord("are"))

	// Test non-stop words
	assert.False(t, service.isStopWord("golang"))
	assert.False(t, service.isStopWord("programming"))
	assert.False(t, service.isStopWord("tutorial"))

	// Test empty string
	assert.False(t, service.isStopWord(""))

	// Test case sensitivity
	assert.True(t, service.isStopWord("the"))
	assert.False(t, service.isStopWord("THE")) // Stop words are lowercase only
}

func TestSearchService_IsRecent(t *testing.T) {
	service := NewSearchService()

	// Test recent article (10 days ago) - should be true
	recentArticle := &models.Article{
		Date: time.Now().AddDate(0, 0, -10), // 10 days ago
	}
	assert.True(t, service.isRecent(recentArticle))

	// Test old article (40 days ago) - should be false
	oldArticle := &models.Article{
		Date: time.Now().AddDate(0, 0, -40), // 40 days ago
	}
	assert.False(t, service.isRecent(oldArticle))

	// Test today's article - should be true
	todayArticle := &models.Article{
		Date: time.Now(),
	}
	assert.True(t, service.isRecent(todayArticle))

	// Test boundary case (exactly 30 days ago) - should be false
	boundaryArticle := &models.Article{
		Date: time.Now().AddDate(0, 0, -30),
	}
	assert.False(t, service.isRecent(boundaryArticle))

	// Test edge case: 29 days ago (just within the threshold) - should be true
	almostThirtyDaysArticle := &models.Article{
		Date: time.Now().AddDate(0, 0, -29),
	}
	assert.True(t, service.isRecent(almostThirtyDaysArticle))

	// Test edge case: 31 days ago (just outside the threshold) - should be false
	justOverThirtyDaysArticle := &models.Article{
		Date: time.Now().AddDate(0, 0, -31),
	}
	assert.False(t, service.isRecent(justOverThirtyDaysArticle))

	// Test future article - should be true
	futureArticle := &models.Article{
		Date: time.Now().AddDate(0, 0, 1), // 1 day in the future
	}
	assert.True(t, service.isRecent(futureArticle))

	// Test very old article - should be false
	veryOldArticle := &models.Article{
		Date: time.Now().AddDate(-1, 0, 0), // 1 year ago
	}
	assert.False(t, service.isRecent(veryOldArticle))
}

func TestSearchService_IsRecentAffectsScoring(t *testing.T) {
	service := NewSearchService()

	// Create two identical articles with different dates
	recentArticle := &models.Article{
		Slug:        "recent-golang-tutorial",
		Title:       "golang programming tutorial",
		Description: "Learn golang programming",
		Tags:        []string{"golang", "tutorial"},
		Categories:  []string{"programming"},
		Content:     "<p>golang programming tutorial content</p>",
		Featured:    false,                         // Not featured to isolate isRecent effect
		Date:        time.Now().AddDate(0, 0, -10), // 10 days ago (recent)
	}

	oldArticle := &models.Article{
		Slug:        "old-golang-tutorial",
		Title:       "golang programming tutorial",
		Description: "Learn golang programming",
		Tags:        []string{"golang", "tutorial"},
		Categories:  []string{"programming"},
		Content:     "<p>golang programming tutorial content</p>",
		Featured:    false,                         // Not featured to isolate isRecent effect
		Date:        time.Now().AddDate(0, 0, -40), // 40 days ago (old)
	}

	// Calculate scores for both articles with the same search terms
	searchTerms := []string{"golang", "tutorial"}
	recentScore, _ := service.calculateScore(recentArticle, searchTerms)
	oldScore, _ := service.calculateScore(oldArticle, searchTerms)

	// Recent article should have a higher score due to the 1.1x multiplier
	assert.True(t, recentScore > oldScore, "Recent article score (%f) should be higher than old article score (%f)", recentScore, oldScore)

	// Verify the scoring boost is approximately 10% (1.1x multiplier)
	expectedRecentScore := oldScore * 1.1
	tolerance := 0.1 // Allow small floating point differences
	assert.True(t, recentScore >= expectedRecentScore-tolerance && recentScore <= expectedRecentScore+tolerance,
		"Recent score (%f) should be approximately 1.1x old score (%f = %f)", recentScore, oldScore, expectedRecentScore)
}

func TestSearchService_ScoreCalculation(t *testing.T) {
	service := NewSearchService()

	// Test title vs content scoring
	titleArticle := &models.Article{
		Slug:    "title-match",
		Title:   "golang tutorial",
		Content: "This is about python",
	}

	contentArticle := &models.Article{
		Slug:    "content-match",
		Title:   "python tutorial",
		Content: "This is about golang programming",
	}

	titleScore, _ := service.calculateScore(titleArticle, []string{"golang"})
	contentScore, _ := service.calculateScore(contentArticle, []string{"golang"})

	// Title matches should score higher than content matches
	assert.True(t, titleScore > contentScore)

	// Test exact phrase matching
	phraseArticle := &models.Article{
		Title: "golang programming tutorial",
	}

	phraseScore, _ := service.calculateScore(phraseArticle, []string{"golang", "programming"})
	assert.True(t, phraseScore > 20) // Should get bonus for phrase match
}

func TestSearchService_SortingByScore(t *testing.T) {
	service := NewSearchService()

	articles := []*models.Article{
		{
			Slug:    "low-score",
			Title:   "Python tutorial",
			Content: "This mentions golang once",
		},
		{
			Slug:  "high-score",
			Title: "Complete golang Tutorial",
			Tags:  []string{"golang", "tutorial"},
		},
		{
			Slug:  "medium-score",
			Title: "Programming tutorial",
			Tags:  []string{"golang"},
		},
	}

	results := service.Search(articles, "golang", 10)
	assert.Len(t, results, 3)

	// Results should be sorted by score (highest first)
	assert.Equal(t, "high-score", results[0].Article.Slug)
	assert.True(t, results[0].Score > results[1].Score)
	assert.True(t, results[1].Score > results[2].Score)
}

func TestSearchService_EmptyAndNilInputs(t *testing.T) {
	service := NewSearchService()

	// Test with nil articles
	results := service.Search(nil, "golang", 10)
	assert.Empty(t, results)

	// Test with empty articles slice
	results = service.Search([]*models.Article{}, "golang", 10)
	assert.Empty(t, results)

	// Test SearchInTitle with nil
	titleResults := service.SearchInTitle(nil, "golang", 10)
	assert.Empty(t, titleResults)

	// Test SearchByTag with nil
	tagResults := service.SearchByTag(nil, "golang")
	assert.Empty(t, tagResults)

	// Test SearchByCategory with nil
	categoryResults := service.SearchByCategory(nil, "programming")
	assert.Empty(t, categoryResults)

	// Test GetSuggestions with nil
	suggestions := service.GetSuggestions(nil, "golang", 10)
	assert.Empty(t, suggestions)
}

// Helper function to create test articles
func createTestArticles() []*models.Article {
	articles := []*models.Article{
		{
			Slug:        "golang-tutorial",
			Title:       "golang tutorial complete",
			Description: "Learn golang programming basics",
			Tags:        []string{"golang", "tutorial"},
			Categories:  []string{"programming"},
			Content:     "<p>golang is a programming language</p>",
			Featured:    true,
			Date:        time.Now(),
		},
		{
			Slug:        "web-development",
			Title:       "Web Development Guide",
			Description: "Modern web development",
			Tags:        []string{"web", "development"},
			Categories:  []string{"web"},
			Content:     "<p>Web development with modern tools</p>",
			Featured:    false,
			Date:        time.Now().AddDate(0, 0, -10),
		},
		{
			Slug:        "advanced-golang",
			Title:       "Advanced golang Patterns",
			Description: "Advanced golang programming patterns",
			Tags:        []string{"golang", "advanced"},
			Categories:  []string{"programming"},
			Content:     "<p>Advanced patterns in golang programming</p>",
			Featured:    false,
			Date:        time.Now().AddDate(0, 0, -40),
		},
		{
			Slug:        "test-article",
			Title:       "Test Article",
			Description: "A test article",
			Tags:        []string{"test"},
			Categories:  []string{"misc"},
			Content:     "<p>Test content</p>",
			Featured:    false,
			Date:        time.Now().AddDate(0, 0, -5),
		},
	}
	
	// Set processor for all articles
	processor := &mockProcessor{}
	for _, article := range articles {
		article.SetProcessor(processor)
	}
	
	return articles
}

// Benchmark tests
func BenchmarkSearchService_Search(b *testing.B) {
	service := NewSearchService()
	articles := createTestArticles()

	for b.Loop() {
		service.Search(articles, "golang programming", 10)
	}
}

func BenchmarkSearchService_SearchInTitle(b *testing.B) {
	service := NewSearchService()
	articles := createTestArticles()

	for b.Loop() {
		service.SearchInTitle(articles, "golang", 10)
	}
}

func BenchmarkSearchService_SearchByTag(b *testing.B) {
	service := NewSearchService()
	articles := createTestArticles()

	for b.Loop() {
		service.SearchByTag(articles, "golang")
	}
}

func BenchmarkSearchService_Tokenize(b *testing.B) {
	service := NewSearchService()
	text := "This is a sample text with multiple words and punctuation!"

	for b.Loop() {
		service.tokenize(text)
	}
}

func BenchmarkSearchService_CalculateScore(b *testing.B) {
	service := NewSearchService()
	article := &models.Article{
		Slug:        "test-article",
		Title:       "Go Programming Tutorial",
		Description: "Learn Go programming language",
		Tags:        []string{"golang", "tutorial"},
		Categories:  []string{"programming"},
		Content:     "<p>Go is a programming language developed by Google</p>",
	}
	article.SetProcessor(&mockProcessor{})
	terms := []string{"go", "programming", "tutorial"}

	for b.Loop() {
		service.calculateScore(article, terms)
	}
}
