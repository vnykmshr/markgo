package services

import (
	"sort"
	"strings"
	"time"

	"github.com/vnykmshr/markgo/internal/models"
)

// Ensure SearchService implements SearchServiceInterface
var _ SearchServiceInterface = (*SearchService)(nil)

type SearchService struct{}

func NewSearchService() *SearchService {
	return &SearchService{}
}

// Search performs a full-text search across articles
func (s *SearchService) Search(articles []*models.Article, query string, limit int) []*models.SearchResult {
	if query == "" || len(articles) == 0 {
		return []*models.SearchResult{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	searchTerms := s.tokenize(query)

	if len(searchTerms) == 0 {
		return []*models.SearchResult{}
	}

	var results []*models.SearchResult

	for _, article := range articles {
		score, matchedFields := s.calculateScore(article, searchTerms)
		if score > 0 {
			results = append(results, &models.SearchResult{
				Article:       article,
				Score:         score,
				MatchedFields: matchedFields,
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
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

func (s *SearchService) calculateScore(article *models.Article, searchTerms []string) (float64, []string) {
	var score float64
	var matchedFields []string
	matchedFieldsMap := make(map[string]bool)

	title := strings.ToLower(article.Title)
	description := strings.ToLower(article.Description)
	excerpt := strings.ToLower(article.Excerpt)
	content := strings.ToLower(s.stripHTML(article.Content))

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
	// Simple tokenization - split by common delimiters
	text = strings.ToLower(text)

	// Replace common punctuation with spaces
	replacer := strings.NewReplacer(
		",", " ",
		".", " ",
		"!", " ",
		"?", " ",
		";", " ",
		":", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"\"", " ",
		"'", " ",
		"-", " ",
		"_", " ",
	)
	text = replacer.Replace(text)

	// Split by whitespace and filter
	words := strings.Fields(text)
	var tokens []string

	for _, word := range words {
		word = strings.TrimSpace(word)
		// Skip very short words and common stop words
		if len(word) > 2 && !s.isStopWord(word) {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

func (s *SearchService) stripHTML(html string) string {
	// Simple HTML tag removal
	result := html
	inTag := false
	var cleaned strings.Builder

	for _, char := range result {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			cleaned.WriteRune(char)
		}
	}

	return cleaned.String()
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
