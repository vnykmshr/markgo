package services

import (
	"testing"

	"github.com/vnykmshr/markgo/internal/models"
)

func TestSearchService_Search(t *testing.T) {
	service := NewSearchService()

	articles := []*models.Article{
		{
			Title:   "Go Programming Tutorial",
			Content: "Learn Go programming language basics",
			Tags:    []string{"go", "programming"},
		},
		{
			Title:   "JavaScript Guide",
			Content: "JavaScript fundamentals and best practices",
			Tags:    []string{"javascript", "web"},
		},
	}

	// Test title match
	results := service.Search(articles, "Go", 10)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Go Programming Tutorial" {
		t.Errorf("Wrong article returned")
	}

	// Test tag match
	results = service.Search(articles, "programming", 10)
	if len(results) != 1 {
		t.Errorf("Expected 1 result for tag search, got %d", len(results))
	}

	// Test empty query
	results = service.Search(articles, "", 10)
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty query, got %d", len(results))
	}

	// Test no match
	results = service.Search(articles, "python", 10)
	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-matching query, got %d", len(results))
	}
}

func TestSearchService_SearchInTitle(t *testing.T) {
	service := NewSearchService()

	articles := []*models.Article{
		{Title: "Go Programming Tutorial"},
		{Title: "JavaScript Guide"},
	}

	results := service.SearchInTitle(articles, "Go", 10)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestSearchService_SearchByTag(t *testing.T) {
	service := NewSearchService()

	articles := []*models.Article{
		{Tags: []string{"go", "programming"}},
		{Tags: []string{"javascript", "web"}},
	}

	results := service.SearchByTag(articles, "go")
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestSearchService_SearchByCategory(t *testing.T) {
	service := NewSearchService()

	articles := []*models.Article{
		{Categories: []string{"programming", "tutorials"}},
		{Categories: []string{"web", "frontend"}},
	}

	results := service.SearchByCategory(articles, "programming")
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}
