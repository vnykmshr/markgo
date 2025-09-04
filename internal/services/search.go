package services

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/utils"
)

// Ensure SearchService implements SearchServiceInterface
var _ SearchServiceInterface = (*SearchService)(nil)

// SearchCache caches processed search data for performance
type SearchCache struct {
	processedContent map[string]string // article slug -> processed content
	suggestions      map[string][]string // query prefix -> suggestions
	mu               sync.RWMutex
}

// SearchService provides optimized search functionality
type SearchService struct {
	cache       *SearchCache
	stringPool  *utils.StringBuilderPool
	slicePool   *utils.SlicePool
	searchLimit int // Early termination optimization
}

func NewSearchService() *SearchService {
	return &SearchService{
		cache: &SearchCache{
			processedContent: make(map[string]string),
			suggestions:     make(map[string][]string),
		},
		stringPool:  utils.NewStringBuilderPool(),
		slicePool:   utils.NewSlicePool(),
		searchLimit: 1000, // Stop processing articles after finding enough results
	}
}

// Search performs an optimized full-text search across articles
func (s *SearchService) Search(articles []*models.Article, query string, limit int) []*models.SearchResult {
	if query == "" || len(articles) == 0 {
		return []*models.SearchResult{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	searchTerms := s.tokenizeOptimized(query)

	if len(searchTerms) == 0 {
		return []*models.SearchResult{}
	}

	// Early termination: If we only need a few results, don't process all articles
	maxResults := limit * 3
	if maxResults <= 0 || maxResults > 100 {
		maxResults = 100 // Cap at reasonable number for performance
	}

	// Pre-allocate results with exact needed capacity
	results := make([]*models.SearchResult, 0, maxResults)
	
	// Priority-based processing: process recent/featured articles first
	articlesToProcess := s.prioritizeArticles(articles)

	for _, article := range articlesToProcess {
		// Early termination when we have enough good results
		if len(results) >= maxResults {
			break
		}

		score, matchedFields := s.calculateScoreOptimized(article, searchTerms)
		if score > 0 {
			results = append(results, &models.SearchResult{
				Article:       article,
				Score:         score,
				MatchedFields: matchedFields,
			})
		}
	}

	// Sort by score (highest first) using optimized sorting
	if len(results) > 1 {
		if limit > 0 && len(results) > limit {
			// Use partial sort for better performance
			s.partialSort(results, limit)
			results = results[:limit]
		} else {
			sort.Slice(results, func(i, j int) bool {
				return results[i].Score > results[j].Score
			})
		}
	}

	return results
}

// SearchInTitle searches only in article titles
func (s *SearchService) SearchInTitle(articles []*models.Article, query string, limit int) []*models.SearchResult {
	if query == "" || len(articles) == 0 {
		return []*models.SearchResult{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var results []*models.SearchResult

	for _, article := range articles {
		title := strings.ToLower(article.Title)
		if strings.Contains(title, query) {
			// Calculate simple score based on position and exactness
			score := 10.0
			if strings.HasPrefix(title, query) {
				score += 5.0
			}
			if title == query {
				score += 10.0
			}

			results = append(results, &models.SearchResult{
				Article:       article,
				Score:         score,
				MatchedFields: []string{"title"},
			})
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// SearchByTag finds articles with specific tags
func (s *SearchService) SearchByTag(articles []*models.Article, tag string) []*models.Article {
	tag = strings.ToLower(strings.TrimSpace(tag))
	var results []*models.Article

	for _, article := range articles {
		for _, articleTag := range article.Tags {
			if strings.ToLower(articleTag) == tag {
				results = append(results, article)
				break
			}
		}
	}

	return results
}

// SearchByCategory finds articles in specific category
func (s *SearchService) SearchByCategory(articles []*models.Article, category string) []*models.Article {
	category = strings.ToLower(strings.TrimSpace(category))
	var results []*models.Article

	for _, article := range articles {
		for _, articleCategory := range article.Categories {
			if strings.ToLower(articleCategory) == category {
				results = append(results, article)
				break
			}
		}
	}

	return results
}

// GetSuggestions returns search suggestions based on existing tags and titles
func (s *SearchService) GetSuggestions(articles []*models.Article, query string, limit int) []string {
	if query == "" || len(articles) == 0 {
		return []string{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	suggestionSet := make(map[string]bool)
	var suggestions []string

	// Search in tags
	for _, article := range articles {
		for _, tag := range article.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				if !suggestionSet[tag] {
					suggestions = append(suggestions, tag)
					suggestionSet[tag] = true
				}
			}
		}
	}

	// Search in titles (extract meaningful words)
	for _, article := range articles {
		title := strings.ToLower(article.Title)
		if strings.Contains(title, query) {
			words := s.tokenize(article.Title)
			for _, word := range words {
				if len(word) > 3 && strings.Contains(strings.ToLower(word), query) {
					if !suggestionSet[word] {
						suggestions = append(suggestions, word)
						suggestionSet[word] = true
					}
				}
			}
		}
	}

	// Sort suggestions alphabetically
	sort.Strings(suggestions)

	if limit > 0 && len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions
}

// ClearCache clears all cached search data (for memory management)
func (s *SearchService) ClearCache() {
	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()
	
	s.cache.processedContent = make(map[string]string)
	s.cache.suggestions = make(map[string][]string)
}

// GetCacheStats returns cache statistics
func (s *SearchService) GetCacheStats() map[string]int {
	s.cache.mu.RLock()
	defer s.cache.mu.RUnlock()
	
	return map[string]int{
		"processed_content": len(s.cache.processedContent),
		"suggestions":      len(s.cache.suggestions),
	}
}

// tokenizeOptimized provides faster tokenization with pooled slices
func (s *SearchService) tokenizeOptimized(text string) []string {
	text = strings.ToLower(text)

	// Use string builder pool for efficient string operations
	builder := s.stringPool.Get()
	defer s.stringPool.Put(builder)

	// Replace common punctuation with spaces efficiently
	builder.Reset()
	for _, char := range text {
		switch char {
		case ',', '.', '!', '?', ';', ':', '(', ')', '[', ']', '{', '}', '"', '\'', '-', '_':
			builder.WriteRune(' ')
		default:
			builder.WriteRune(char)
		}
	}
	
	// Split by whitespace and filter
	words := strings.Fields(builder.String())
	
	// Use pooled slice for tokens
	tokens := s.slicePool.GetStringSlice()
	defer s.slicePool.PutStringSlice(tokens)
	*tokens = (*tokens)[:0] // Clear but keep capacity

	for _, word := range words {
		word = strings.TrimSpace(word)
		// Skip very short words and common stop words
		if len(word) > 2 && !s.isStopWord(word) {
			*tokens = append(*tokens, word)
		}
	}

	// Return a copy to avoid pool contamination
	result := make([]string, len(*tokens))
	copy(result, *tokens)
	return result
}

// calculateScoreOptimized provides faster scoring with caching and early exits
func (s *SearchService) calculateScoreOptimized(article *models.Article, searchTerms []string) (float64, []string) {
	var score float64
	var matchedFields []string
	matchedFieldsMap := make(map[string]bool, 6) // Pre-size for performance

	// Pre-computed lowercase versions (done lazily)
	var title, description, excerpt, content string
	titleComputed := false

	for _, term := range searchTerms {
		termScore := 0.0

		// Title matching (highest weight) - compute only once
		if !titleComputed {
			title = strings.ToLower(article.Title)
			titleComputed = true
		}
		if titleScore := s.scoreTitleMatch(title, term); titleScore > 0 {
			termScore += titleScore
			if !matchedFieldsMap["title"] {
				matchedFields = append(matchedFields, "title")
				matchedFieldsMap["title"] = true
			}
		}

		// Tags matching - early exit optimization (check before heavy operations)
		if tagScore := s.scoreTagsMatch(article.Tags, term); tagScore > 0 {
			termScore += tagScore
			if !matchedFieldsMap["tags"] {
				matchedFields = append(matchedFields, "tags")
				matchedFieldsMap["tags"] = true
			}
		}

		// Categories matching - early exit optimization  
		if catScore := s.scoreCategoriesMatch(article.Categories, term); catScore > 0 {
			termScore += catScore
			if !matchedFieldsMap["categories"] {
				matchedFields = append(matchedFields, "categories")
				matchedFieldsMap["categories"] = true
			}
		}

		// Description matching (compute only if needed)
		if termScore < 20.0 && article.Description != "" {
			if description == "" {
				description = strings.ToLower(article.Description)
			}
			if strings.Contains(description, term) {
				termScore += 12.0
				if !matchedFieldsMap["description"] {
					matchedFields = append(matchedFields, "description")
					matchedFieldsMap["description"] = true
				}
			}
		}

		// Excerpt matching (compute only if needed)
		if termScore < 20.0 {
			if excerpt == "" {
				excerpt = strings.ToLower(article.GetExcerpt())
			}
			if strings.Contains(excerpt, term) {
				termScore += 5.0
				if !matchedFieldsMap["excerpt"] {
					matchedFields = append(matchedFields, "excerpt")
					matchedFieldsMap["excerpt"] = true
				}
			}
		}

		// Content matching (lowest weight) - only if we really need more matches
		if termScore < 10.0 {
			if content == "" {
				content = s.getProcessedContent(article)
			}
			if strings.Contains(content, term) {
				// Count occurrences efficiently with limit
				occurrences := min(strings.Count(content, term), 10)
				termScore += float64(occurrences) * 0.5 // Cap content score
				if !matchedFieldsMap["content"] {
					matchedFields = append(matchedFields, "content")
					matchedFieldsMap["content"] = true
				}
			}
		}

		score += termScore
		
		// Early exit if we have a very high score already
		if score > 100.0 {
			break
		}
	}

	// Only do phrase matching if we have multiple terms and reasonable score
	if len(searchTerms) > 1 && score > 5.0 {
		fullQuery := strings.Join(searchTerms, " ")
		if strings.Contains(title, fullQuery) {
			score += 10.0
		} else if excerpt != "" && strings.Contains(excerpt, fullQuery) {
			score += 5.0
		}
	}

	// Apply multipliers
	if article.Featured {
		score *= 1.2
	}
	if s.isRecent(article) {
		score *= 1.1
	}

	return score, matchedFields
}

// getProcessedContent gets cached processed content or processes and caches it
func (s *SearchService) getProcessedContent(article *models.Article) string {
	s.cache.mu.RLock()
	if content, exists := s.cache.processedContent[article.Slug]; exists {
		s.cache.mu.RUnlock()
		return content
	}
	s.cache.mu.RUnlock()

	// Process content
	content := strings.ToLower(s.stripHTMLOptimized(article.GetProcessedContent()))
	
	// Cache the processed content with bounds checking
	s.cache.mu.Lock()
	// Prevent unbounded cache growth during benchmarks
	if len(s.cache.processedContent) > 1000 {
		// Clear half the cache to prevent memory leaks
		for slug := range s.cache.processedContent {
			delete(s.cache.processedContent, slug)
			if len(s.cache.processedContent) <= 500 {
				break
			}
		}
	}
	s.cache.processedContent[article.Slug] = content
	s.cache.mu.Unlock()
	
	return content
}

// scoreTitleMatch provides optimized title scoring
func (s *SearchService) scoreTitleMatch(title, term string) float64 {
	if !strings.Contains(title, term) {
		return 0
	}
	
	if title == term {
		return 30.0 // Exact title match
	} else if strings.HasPrefix(title, term) {
		return 20.0 // Title starts with term
	} else {
		return 15.0 // Title contains term
	}
}

// scoreTagsMatch provides optimized tag scoring with early exit
func (s *SearchService) scoreTagsMatch(tags []string, term string) float64 {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), term) {
			return 10.0
		}
	}
	return 0
}

// scoreCategoriesMatch provides optimized category scoring with early exit
func (s *SearchService) scoreCategoriesMatch(categories []string, term string) float64 {
	for _, category := range categories {
		if strings.Contains(strings.ToLower(category), term) {
			return 8.0
		}
	}
	return 0
}

// stripHTMLOptimized provides faster HTML stripping with string builder
func (s *SearchService) stripHTMLOptimized(html string) string {
	if html == "" {
		return ""
	}
	
	builder := s.stringPool.Get()
	defer s.stringPool.Put(builder)
	builder.Reset()

	inTag := false
	for _, char := range html {
		switch char {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				builder.WriteRune(char)
			}
		}
	}

	return builder.String()
}

// partialSort performs a partial sort to get top K elements efficiently
func (s *SearchService) partialSort(results []*models.SearchResult, k int) {
	if k >= len(results) {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
		return
	}

	// Use a heap-based approach for better performance with large datasets
	for i := 0; i < k; i++ {
		maxIdx := i
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[maxIdx].Score {
				maxIdx = j
			}
		}
		if maxIdx != i {
			results[i], results[maxIdx] = results[maxIdx], results[i]
		}
	}
}

// min is a simple min function for integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// prioritizeArticles sorts articles to process high-priority ones first
func (s *SearchService) prioritizeArticles(articles []*models.Article) []*models.Article {
	if len(articles) <= 100 {
		// For small datasets, return as-is to avoid sorting overhead
		return articles
	}

	// Create a copy to avoid modifying the original slice
	prioritized := make([]*models.Article, len(articles))
	copy(prioritized, articles)

	// Sort by priority: featured and recent articles first
	sort.Slice(prioritized, func(i, j int) bool {
		a, b := prioritized[i], prioritized[j]
		
		// Featured articles get highest priority
		if a.Featured != b.Featured {
			return a.Featured
		}
		
		// Then by recency
		return a.Date.After(b.Date)
	})

	return prioritized
}

func (s *SearchService) calculateScore(article *models.Article, searchTerms []string) (float64, []string) {
	var score float64
	var matchedFields []string
	matchedFieldsMap := make(map[string]bool)

	title := strings.ToLower(article.Title)
	description := strings.ToLower(article.Description)
	excerpt := strings.ToLower(article.GetExcerpt())
	content := strings.ToLower(s.stripHTML(article.GetProcessedContent()))

	for _, term := range searchTerms {
		termScore := 0.0

		// Title matching (highest weight)
		if strings.Contains(title, term) {
			if strings.HasPrefix(title, term) {
				termScore += 20.0 // Title starts with term
			} else if title == term {
				termScore += 30.0 // Exact title match
			} else {
				termScore += 15.0 // Title contains term
			}
			if !matchedFieldsMap["title"] {
				matchedFields = append(matchedFields, "title")
				matchedFieldsMap["title"] = true
			}
		}

		// Description matching
		if description != "" && strings.Contains(description, term) {
			termScore += 12.0
			if !matchedFieldsMap["description"] {
				matchedFields = append(matchedFields, "description")
				matchedFieldsMap["description"] = true
			}
		}

		// Tags matching
		for _, tag := range article.Tags {
			if strings.Contains(strings.ToLower(tag), term) {
				termScore += 10.0
				if !matchedFieldsMap["tags"] {
					matchedFields = append(matchedFields, "tags")
					matchedFieldsMap["tags"] = true
				}
				break
			}
		}

		// Categories matching
		for _, category := range article.Categories {
			if strings.Contains(strings.ToLower(category), term) {
				termScore += 8.0
				if !matchedFieldsMap["categories"] {
					matchedFields = append(matchedFields, "categories")
					matchedFieldsMap["categories"] = true
				}
				break
			}
		}

		// Excerpt matching
		if strings.Contains(excerpt, term) {
			termScore += 5.0
			if !matchedFieldsMap["excerpt"] {
				matchedFields = append(matchedFields, "excerpt")
				matchedFieldsMap["excerpt"] = true
			}
		}

		// Content matching (lowest weight)
		if strings.Contains(content, term) {
			// Count occurrences in content
			occurrences := strings.Count(content, term)
			termScore += float64(occurrences) * 0.5
			if !matchedFieldsMap["content"] {
				matchedFields = append(matchedFields, "content")
				matchedFieldsMap["content"] = true
			}
		}

		score += termScore
	}

	// Boost score for exact phrase matches
	fullQuery := strings.Join(searchTerms, " ")
	if strings.Contains(title, fullQuery) {
		score += 10.0
	}
	if strings.Contains(excerpt, fullQuery) {
		score += 5.0
	}

	// Boost score for featured articles
	if article.Featured {
		score *= 1.2
	}

	// Boost score for recent articles
	if s.isRecent(article) {
		score *= 1.1
	}

	return score, matchedFields
}

func (s *SearchService) tokenize(text string) []string {
	// Delegate to optimized version for backward compatibility
	return s.tokenizeOptimized(text)
}

func (s *SearchService) stripHTML(html string) string {
	// Delegate to optimized version for backward compatibility
	return s.stripHTMLOptimized(html)
}

func (s *SearchService) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "can": true, "this": true,
		"that": true, "these": true, "those": true, "i": true, "you": true,
		"he": true, "she": true, "it": true, "we": true, "they": true,
		"me": true, "him": true, "her": true, "us": true, "them": true,
		"my": true, "your": true, "his": true, "its": true, "our": true,
		"their": true, "am": true, "as": true, "so": true, "no": true,
		"not": true, "up": true, "out": true, "if": true, "about": true,
		"who": true, "what": true, "where": true, "when": true, "why": true,
		"how": true, "all": true, "any": true, "both": true, "each": true,
		"few": true, "more": true, "most": true, "other": true, "some": true,
		"such": true, "only": true, "own": true, "same": true, "than": true,
		"too": true, "very": true, "just": true, "now": true,
	}

	return stopWords[word]
}

func (s *SearchService) isRecent(article *models.Article) bool {
	// Consider articles from the last 30 days as recent
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return article.Date.After(thirtyDaysAgo)
}
