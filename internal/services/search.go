package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/vnykmshr/goflow/pkg/scheduling/scheduler"
	"github.com/vnykmshr/goflow/pkg/scheduling/workerpool"
	"github.com/vnykmshr/obcache-go/pkg/obcache"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/utils"
)

// Ensure SearchService implements SearchServiceInterface
var _ SearchServiceInterface = (*SearchService)(nil)

// CachedSearchFunctions holds all obcache-wrapped expensive search operations
type CachedSearchFunctions struct {
	SearchArticles      func([]*models.Article, string, int) []*models.SearchResult
	SearchInTitle       func([]*models.Article, string, int) []*models.SearchResult
	ProcessContent      func(string) string
	GenerateSuggestions func(string, []*models.Article) []string
	CalculateScore      func(*models.Article, []string) (float64, []string)
}

// SearchService provides optimized search functionality with obcache and goflow
type SearchService struct {
	// obcache integration
	obcache         *obcache.Cache
	cachedFunctions CachedSearchFunctions

	// goflow integration
	scheduler  scheduler.Scheduler
	workerPool workerpool.Pool
	ctx        context.Context
	cancel     context.CancelFunc

	// utilities
	stringPool  *utils.StringBuilderPool
	slicePool   *utils.SlicePool
	searchLimit int
}

func NewSearchService() *SearchService {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize obcache for search operations
	cacheConfig := obcache.NewDefaultConfig()
	cacheConfig.MaxEntries = 2000
	cacheConfig.DefaultTTL = 10 * time.Minute

	obcacheInstance, err := obcache.New(cacheConfig)
	if err != nil {
		cancel()
		return &SearchService{
			stringPool:  utils.NewStringBuilderPool(),
			slicePool:   utils.NewSlicePool(),
			searchLimit: 1000,
		}
	}

	// Initialize goflow components
	goflowScheduler := scheduler.New()
	_ = goflowScheduler.Start() // Continue even if scheduler fails to start

	workerPoolConfig := workerpool.Config{
		WorkerCount: 4,
		QueueSize:   500,
		TaskTimeout: 30 * time.Second,
	}
	goflowWorkerPool := workerpool.NewWithConfig(workerPoolConfig)

	service := &SearchService{
		obcache:     obcacheInstance,
		scheduler:   goflowScheduler,
		workerPool:  goflowWorkerPool,
		ctx:         ctx,
		cancel:      cancel,
		stringPool:  utils.NewStringBuilderPool(),
		slicePool:   utils.NewSlicePool(),
		searchLimit: 1000,
	}

	// Initialize cached functions
	service.initializeCachedFunctions()

	// Setup background search maintenance
	service.setupSearchMaintenance()

	return service
}

// initializeCachedFunctions initializes all obcache-wrapped functions
func (s *SearchService) initializeCachedFunctions() {
	if s.obcache == nil {
		return
	}

	// Wrap expensive search operations with obcache
	s.cachedFunctions.SearchArticles = obcache.Wrap(
		s.obcache,
		s.searchArticlesUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) >= 2 {
				if query, ok := args[1].(string); ok {
					if limit, ok := args[2].(int); ok {
						return fmt.Sprintf("search:%s:%d", query, limit)
					}
				}
			}
			return "search:default"
		}),
		obcache.WithTTL(5*time.Minute),
	)

	s.cachedFunctions.SearchInTitle = obcache.Wrap(
		s.obcache,
		s.searchInTitleUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) >= 2 {
				if query, ok := args[1].(string); ok {
					if limit, ok := args[2].(int); ok {
						return fmt.Sprintf("title:%s:%d", query, limit)
					}
				}
			}
			return "title:default"
		}),
		obcache.WithTTL(5*time.Minute),
	)

	s.cachedFunctions.ProcessContent = obcache.Wrap(
		s.obcache,
		s.processContentUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) > 0 {
				if content, ok := args[0].(string); ok {
					hash := sha256.Sum256([]byte(content))
					return fmt.Sprintf("content:%x", hash[:8])
				}
			}
			return "content:default"
		}),
		obcache.WithTTL(10*time.Minute),
	)

	s.cachedFunctions.GenerateSuggestions = obcache.Wrap(
		s.obcache,
		s.generateSuggestionsUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) > 0 {
				if query, ok := args[0].(string); ok {
					return fmt.Sprintf("suggestions:%s", query)
				}
			}
			return "suggestions:default"
		}),
		obcache.WithTTL(10*time.Minute),
	)

	s.cachedFunctions.CalculateScore = obcache.Wrap(
		s.obcache,
		s.calculateScoreUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) >= 2 {
				if article, ok := args[0].(*models.Article); ok {
					if terms, ok := args[1].([]string); ok {
						return fmt.Sprintf("score:%s:%s", article.Slug, strings.Join(terms, ","))
					}
				}
			}
			return "score:default"
		}),
		obcache.WithTTL(5*time.Minute),
	)
}

// setupSearchMaintenance sets up background maintenance tasks using goflow
func (s *SearchService) setupSearchMaintenance() {
	if s.scheduler == nil {
		return
	}

	// Cache warming task every 30 minutes
	cacheWarmingTask := workerpool.TaskFunc(func(ctx context.Context) error {
		if s.obcache != nil {
			s.obcache.Cleanup()
		}
		return nil
	})

	// Schedule cache warming
	_ = s.scheduler.ScheduleCron("search-cache-warming", "0 */30 * * * *", cacheWarmingTask) // Continue without cache warming if scheduling fails

	// Cache cleanup task every hour - use basic cleanup instead of Evict
	cleanupTask := workerpool.TaskFunc(func(ctx context.Context) error {
		if s.obcache != nil {
			s.obcache.Cleanup() // Use Cleanup instead of Evict
		}
		return nil
	})

	// Schedule cleanup
	_ = s.scheduler.ScheduleCron("search-cache-cleanup", "0 0 * * * *", cleanupTask) // Continue without scheduled cleanup if scheduling fails
}

// Search performs an optimized full-text search across articles with caching
func (s *SearchService) Search(articles []*models.Article, query string, limit int) []*models.SearchResult {
	if query == "" || len(articles) == 0 {
		return []*models.SearchResult{}
	}

	query = strings.ToLower(strings.TrimSpace(query))

	// Use cached search function if available
	if s.cachedFunctions.SearchArticles != nil {
		return s.cachedFunctions.SearchArticles(articles, query, limit)
	}

	// Fallback to uncached version
	return s.searchArticlesUncached(articles, query, limit)
}

// SearchPaginated performs a paginated search with caching support
func (s *SearchService) SearchPaginated(articles []*models.Article, query string, page, pageSize int) (*models.SearchResultPage, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}

	// Get all search results (potentially from cache)
	allResults := s.Search(articles, query, 0)
	totalResults := len(allResults)

	// Calculate pagination bounds
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	var pageResults []*models.SearchResult
	if startIdx < totalResults {
		if endIdx > totalResults {
			endIdx = totalResults
		}
		pageResults = allResults[startIdx:endIdx]
	}

	pagination := models.NewPagination(page, totalResults, pageSize)

	return &models.SearchResultPage{
		Results:    pageResults,
		Pagination: pagination,
		Query:      query,
		TotalTime:  0,
	}, nil
}

// SearchInTitle searches only in article titles
func (s *SearchService) SearchInTitle(articles []*models.Article, query string, limit int) []*models.SearchResult {
	if query == "" || len(articles) == 0 {
		return []*models.SearchResult{}
	}

	query = strings.ToLower(strings.TrimSpace(query))

	// Use cached function if available
	if s.cachedFunctions.SearchInTitle != nil {
		return s.cachedFunctions.SearchInTitle(articles, query, limit)
	}

	// Fallback to uncached version
	return s.searchInTitleUncached(articles, query, limit)
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

	// Use cached function if available
	if s.cachedFunctions.GenerateSuggestions != nil {
		suggestions := s.cachedFunctions.GenerateSuggestions(query, articles)
		if limit > 0 && len(suggestions) > limit {
			return suggestions[:limit]
		}
		return suggestions
	}

	// Fallback to uncached version
	return s.generateSuggestionsUncached(query, articles)
}

// ClearCache clears all cached search data
func (s *SearchService) ClearCache() {
	if s.obcache != nil {
		_ = s.obcache.Clear()
	}
}

// GetCacheStats returns cache statistics
func (s *SearchService) GetCacheStats() map[string]int {
	if s.obcache == nil {
		return map[string]int{}
	}

	stats := s.obcache.Stats()
	return map[string]int{
		"key_count": int(stats.KeyCount()),
		"hits":      int(stats.Hits()),
		"misses":    int(stats.Misses()),
		"evictions": int(stats.Evictions()),
		"hit_ratio": int(stats.HitRate() * 100),
	}
}

// Shutdown gracefully shuts down the search service
func (s *SearchService) Shutdown() {
	if s.cancel != nil {
		s.cancel()
	}

	if s.scheduler != nil {
		s.scheduler.Stop()
	}

	if s.workerPool != nil {
		<-s.workerPool.Shutdown()
	}

	if s.obcache != nil {
		s.obcache.Close()
	}
}

// =====================
// Uncached Implementations
// =====================

// searchArticlesUncached performs the actual search without caching
func (s *SearchService) searchArticlesUncached(articles []*models.Article, query string, limit int) []*models.SearchResult {
	searchTerms := s.tokenize(query)
	if len(searchTerms) == 0 {
		return []*models.SearchResult{}
	}

	maxResults := limit * 3
	if maxResults <= 0 || maxResults > 100 {
		maxResults = 100
	}

	results := make([]*models.SearchResult, 0, maxResults)
	articlesToProcess := s.prioritizeArticles(articles)

	for _, article := range articlesToProcess {
		if len(results) >= maxResults {
			break
		}

		score, matchedFields := s.calculateScore(article, searchTerms)
		if score > 0 {
			results = append(results, &models.SearchResult{
				Article:       article,
				Score:         score,
				MatchedFields: matchedFields,
			})
		}
	}

	// Sort by score
	if len(results) > 1 {
		if limit > 0 && len(results) > limit {
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

// searchInTitleUncached searches only in titles without caching
func (s *SearchService) searchInTitleUncached(articles []*models.Article, query string, limit int) []*models.SearchResult {
	var results []*models.SearchResult

	for _, article := range articles {
		title := strings.ToLower(article.Title)
		if strings.Contains(title, query) {
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

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// processContentUncached processes content without caching
func (s *SearchService) processContentUncached(content string) string {
	return strings.ToLower(s.stripHTML(content))
}

// generateSuggestionsUncached generates suggestions without caching
func (s *SearchService) generateSuggestionsUncached(query string, articles []*models.Article) []string {
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

	// Search in titles
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

	sort.Strings(suggestions)
	return suggestions
}

// calculateScoreUncached calculates article score without caching
func (s *SearchService) calculateScoreUncached(article *models.Article, searchTerms []string) (float64, []string) {
	return s.calculateScore(article, searchTerms)
}

// calculateScore provides scoring logic
func (s *SearchService) calculateScore(article *models.Article, searchTerms []string) (float64, []string) {
	var score float64
	var matchedFields []string
	matchedFieldsMap := make(map[string]bool)

	title := strings.ToLower(article.Title)
	description := strings.ToLower(article.Description)
	excerpt := strings.ToLower(article.GetExcerpt())

	var content string
	if s.cachedFunctions.ProcessContent != nil {
		content = s.cachedFunctions.ProcessContent(article.GetProcessedContent())
	} else {
		content = s.processContentUncached(article.GetProcessedContent())
	}

	for _, term := range searchTerms {
		termScore := 0.0

		// Title matching (highest weight)
		if strings.Contains(title, term) {
			if title == term {
				termScore += 30.0
			} else if strings.HasPrefix(title, term) {
				termScore += 20.0
			} else {
				termScore += 15.0
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
			occurrences := min(strings.Count(content, term), 10)
			termScore += float64(occurrences) * 0.5
			if !matchedFieldsMap["content"] {
				matchedFields = append(matchedFields, "content")
				matchedFieldsMap["content"] = true
			}
		}

		score += termScore
	}

	// Boost for phrase matches
	if len(searchTerms) > 1 {
		fullQuery := strings.Join(searchTerms, " ")
		if strings.Contains(title, fullQuery) {
			score += 10.0
		} else if strings.Contains(excerpt, fullQuery) {
			score += 5.0
		}
	}

	// Boost for featured articles
	if article.Featured {
		score *= 1.2
	}

	// Boost for recent articles
	if s.isRecent(article) {
		score *= 1.1
	}

	return score, matchedFields
}

// =====================
// Helper Methods
// =====================

func (s *SearchService) tokenize(text string) []string {
	text = strings.ToLower(text)

	builder := s.stringPool.Get()
	defer s.stringPool.Put(builder)

	builder.Reset()
	for _, char := range text {
		switch char {
		case ',', '.', '!', '?', ';', ':', '(', ')', '[', ']', '{', '}', '"', '\'', '-', '_':
			builder.WriteRune(' ')
		default:
			builder.WriteRune(char)
		}
	}

	words := strings.Fields(builder.String())
	tokens := make([]string, 0, len(words))

	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 2 && !s.isStopWord(word) {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

func (s *SearchService) stripHTML(html string) string {
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

func (s *SearchService) partialSort(results []*models.SearchResult, k int) {
	if k >= len(results) {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
		return
	}

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

func (s *SearchService) prioritizeArticles(articles []*models.Article) []*models.Article {
	if len(articles) <= 100 {
		return articles
	}

	prioritized := make([]*models.Article, len(articles))
	copy(prioritized, articles)

	sort.Slice(prioritized, func(i, j int) bool {
		a, b := prioritized[i], prioritized[j]

		if a.Featured != b.Featured {
			return a.Featured
		}

		return a.Date.After(b.Date)
	})

	return prioritized
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
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return article.Date.After(thirtyDaysAgo)
}
