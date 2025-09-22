package seo

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// GenerateArticleSchema creates Schema.org Article structured data
func (s *Service) GenerateArticleSchema(article *models.Article, baseURL string) (map[string]interface{}, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	if article == nil {
		return nil, fmt.Errorf("article cannot be nil")
	}

	articleURL, err := s.buildArticleURL(article.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to build article URL: %w", err)
	}

	// Extract first image from content if available
	imageURL := s.extractFirstImage(article.Content, baseURL)

	schema := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "Article",
		"headline": article.Title,
		"url":      articleURL,
	}

	// Description
	if article.Description != "" {
		schema["description"] = article.Description
	} else {
		// Generate excerpt as fallback
		schema["description"] = s.generateExcerpt(article.Content, 160)
	}

	// Author
	authorName := article.Author
	if authorName == "" {
		authorName = s.siteConfig.Author
	}

	schema["author"] = map[string]interface{}{
		"@type": "Person",
		"name":  authorName,
	}

	// Publisher (organization)
	publisher := map[string]interface{}{
		"@type": "Organization",
		"name":  s.siteConfig.Name,
		"url":   s.siteConfig.BaseURL,
	}

	if s.siteConfig.Logo != "" {
		publisher["logo"] = map[string]interface{}{
			"@type": "ImageObject",
			"url":   s.buildImageURL(s.siteConfig.Logo, baseURL),
		}
	}

	schema["publisher"] = publisher

	// Dates
	schema["datePublished"] = article.Date.Format(time.RFC3339)
	if !article.LastModified.IsZero() {
		schema["dateModified"] = article.LastModified.Format(time.RFC3339)
	} else {
		schema["dateModified"] = article.Date.Format(time.RFC3339)
	}

	// Main entity of page
	schema["mainEntityOfPage"] = map[string]interface{}{
		"@type": "WebPage",
		"@id":   articleURL,
	}

	// Image
	if imageURL != "" {
		schema["image"] = map[string]interface{}{
			"@type": "ImageObject",
			"url":   imageURL,
		}
	}

	// Keywords from tags
	if len(article.Tags) > 0 {
		schema["keywords"] = strings.Join(article.Tags, ", ")
	}

	// Article section (categories)
	if len(article.Categories) > 0 {
		schema["articleSection"] = article.Categories[0] // Primary category
	}

	// Word count and reading time
	if article.WordCount > 0 {
		schema["wordCount"] = article.WordCount
	}

	s.logger.Debug("Generated article schema",
		"slug", article.Slug,
		"title", article.Title,
		"url", articleURL)

	return schema, nil
}

// GenerateWebsiteSchema creates Schema.org WebSite structured data
func (s *Service) GenerateWebsiteSchema(siteConfig services.SiteConfig) (map[string]interface{}, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	schema := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "WebSite",
		"name":     siteConfig.Name,
		"url":      siteConfig.BaseURL,
	}

	if siteConfig.Description != "" {
		schema["description"] = siteConfig.Description
	}

	// Publisher information
	if siteConfig.Author != "" {
		schema["publisher"] = map[string]interface{}{
			"@type": "Person",
			"name":  siteConfig.Author,
		}
	}

	// Potential search action
	searchURL, err := url.Parse(siteConfig.BaseURL)
	if err == nil {
		searchURL.Path = "/search"
		schema["potentialAction"] = map[string]interface{}{
			"@type":       "SearchAction",
			"target":      searchURL.String() + "?q={search_term_string}",
			"query-input": "required name=search_term_string",
		}
	}

	s.logger.Debug("Generated website schema", "name", siteConfig.Name)
	return schema, nil
}

// GenerateBreadcrumbSchema creates Schema.org BreadcrumbList structured data
func (s *Service) GenerateBreadcrumbSchema(breadcrumbs []services.Breadcrumb) (map[string]interface{}, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	if len(breadcrumbs) == 0 {
		return nil, fmt.Errorf("breadcrumbs cannot be empty")
	}

	itemList := make([]map[string]interface{}, len(breadcrumbs))

	for i, crumb := range breadcrumbs {
		itemList[i] = map[string]interface{}{
			"@type":    "ListItem",
			"position": i + 1,
			"name":     crumb.Name,
			"item":     crumb.URL,
		}
	}

	schema := map[string]interface{}{
		"@context":        "https://schema.org",
		"@type":           "BreadcrumbList",
		"itemListElement": itemList,
	}

	s.logger.Debug("Generated breadcrumb schema", "items", len(breadcrumbs))
	return schema, nil
}

// extractFirstImage finds the first image in markdown content
func (s *Service) extractFirstImage(content, baseURL string) string {
	// Look for markdown image syntax: ![alt](url)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "![") && strings.Contains(line, "](") {
			start := strings.Index(line, "](") + 2
			end := strings.Index(line[start:], ")")
			if end > 0 {
				imagePath := line[start : start+end]
				return s.buildImageURL(imagePath, baseURL)
			}
		}
	}

	// Fallback to site default image
	if s.siteConfig.Image != "" {
		return s.buildImageURL(s.siteConfig.Image, baseURL)
	}

	return ""
}

// buildImageURL constructs a complete URL for an image
func (s *Service) buildImageURL(imagePath, baseURL string) string {
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		return imagePath // Already absolute URL
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return imagePath // Return as-is if base URL is invalid
	}

	imageURL, err := base.Parse(imagePath)
	if err != nil {
		return imagePath // Return as-is if parsing fails
	}

	return imageURL.String()
}

// generateExcerpt creates a short excerpt from content
func (s *Service) generateExcerpt(content string, maxLength int) string {
	// Remove markdown formatting
	text := content
	text = strings.ReplaceAll(text, "#", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "_", "")
	text = strings.ReplaceAll(text, "`", "")

	// Clean up whitespace
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	// Build excerpt
	var excerpt strings.Builder
	for _, word := range words {
		if excerpt.Len()+len(word)+1 > maxLength {
			break
		}
		if excerpt.Len() > 0 {
			excerpt.WriteString(" ")
		}
		excerpt.WriteString(word)
	}

	result := excerpt.String()
	if len(result) < len(strings.Join(words, " ")) {
		result += "..."
	}

	return result
}
