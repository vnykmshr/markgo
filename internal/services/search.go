package services

import (
	"sort"
	"strings"

	"github.com/vnykmshr/markgo/internal/models"
)

type SearchService struct{}

func NewSearchService() *SearchService {
	return &SearchService{}
}

func (s *SearchService) Search(articles []*models.Article, query string, limit int) []*models.SearchResult {
	if query == "" || len(articles) == 0 {
		return []*models.SearchResult{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return []*models.SearchResult{}
	}

	var results []*models.SearchResult
	for _, article := range articles {
		score, fields := s.calculateScore(article, terms)
		if score > 0 {
			results = append(results, &models.SearchResult{
				Article:       article,
				Score:         score,
				MatchedFields: fields,
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

func (s *SearchService) SearchInTitle(articles []*models.Article, query string, limit int) []*models.SearchResult {
	if query == "" || len(articles) == 0 {
		return []*models.SearchResult{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
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

func (s *SearchService) SearchByTag(articles []*models.Article, tag string) []*models.Article {
	tag = strings.ToLower(strings.TrimSpace(tag))
	var results []*models.Article

	for _, article := range articles {
		for _, articleTag := range article.Tags {
			if strings.EqualFold(articleTag, tag) {
				results = append(results, article)
				break
			}
		}
	}

	return results
}

func (s *SearchService) SearchByCategory(articles []*models.Article, category string) []*models.Article {
	category = strings.ToLower(strings.TrimSpace(category))
	var results []*models.Article

	for _, article := range articles {
		for _, articleCategory := range article.Categories {
			if strings.EqualFold(articleCategory, category) {
				results = append(results, article)
				break
			}
		}
	}

	return results
}

func (s *SearchService) GetSuggestions(articles []*models.Article, query string, limit int) []string {
	if query == "" {
		return []string{}
	}

	query = strings.ToLower(query)
	suggestions := make(map[string]bool)

	for _, article := range articles {
		// Add matching tags
		for _, tag := range article.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				suggestions[tag] = true
			}
		}

		// Add matching categories
		for _, category := range article.Categories {
			if strings.Contains(strings.ToLower(category), query) {
				suggestions[category] = true
			}
		}

		// Add matching title words
		titleWords := strings.Fields(strings.ToLower(article.Title))
		for _, word := range titleWords {
			if len(word) > 3 && strings.Contains(word, query) {
				suggestions[word] = true
			}
		}
	}

	result := make([]string, 0, len(suggestions))
	for suggestion := range suggestions {
		result = append(result, suggestion)
	}

	sort.Strings(result)

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

func (s *SearchService) calculateScore(article *models.Article, terms []string) (score float64, fields []string) {

	title := strings.ToLower(article.Title)
	content := strings.ToLower(article.Content)

	for _, term := range terms {
		// Title matches get high score
		if strings.Contains(title, term) {
			score += 10.0
			if !contains(fields, "title") {
				fields = append(fields, "title")
			}
		}

		// Content matches get lower score
		if strings.Contains(content, term) {
			score += 1.0
			if !contains(fields, "content") {
				fields = append(fields, "content")
			}
		}

		// Tag exact matches get high score
		for _, tag := range article.Tags {
			if strings.EqualFold(tag, term) {
				score += 15.0
				if !contains(fields, "tags") {
					fields = append(fields, "tags")
				}
			}
		}

		// Category exact matches get high score
		for _, category := range article.Categories {
			if strings.EqualFold(category, term) {
				score += 12.0
				if !contains(fields, "categories") {
					fields = append(fields, "categories")
				}
			}
		}
	}

	return score, fields
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
