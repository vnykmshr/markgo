// Package seo provides SEO functionality including meta tags, schema markup, and content analysis
package seo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

const (
	// Twitter card types
	twitterCardSummary = "summary"

	// Robots meta tag values
	robotsIndexFollow = "index, follow"
)

// GenerateOpenGraphTags creates Open Graph meta tags for social media sharing
func (s *Service) GenerateOpenGraphTags(article *models.Article, baseURL string) (map[string]string, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	if article == nil {
		return nil, fmt.Errorf("article cannot be nil")
	}

	tags := make(map[string]string)

	// Basic Open Graph tags
	tags["og:type"] = "article"
	tags["og:title"] = article.Title
	tags["og:site_name"] = s.siteConfig.Name

	// URL
	articleURL, err := s.buildArticleURL(article.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to build article URL: %w", err)
	}
	tags["og:url"] = articleURL

	// Description
	description := article.Description
	if description == "" {
		description = s.generateExcerpt(article.Content, 160)
	}
	if description != "" {
		tags["og:description"] = description
	}

	// Image
	imageURL := s.extractFirstImage(article.Content, baseURL)
	if imageURL != "" {
		tags["og:image"] = imageURL
		tags["og:image:alt"] = fmt.Sprintf("Featured image for %s", article.Title)
	}

	// Article-specific tags
	tags["article:published_time"] = article.Date.Format("2006-01-02T15:04:05Z07:00")
	if !article.LastModified.IsZero() {
		tags["article:modified_time"] = article.LastModified.Format("2006-01-02T15:04:05Z07:00")
	}

	// Author
	authorName := article.Author
	if authorName == "" {
		authorName = s.siteConfig.Author
	}
	if authorName != "" {
		tags["article:author"] = authorName
	}

	// Tags and categories
	if len(article.Tags) > 0 {
		for _, tag := range article.Tags {
			// Note: Multiple values for same key should be handled by the template
			tags["article:tag"] = tag
		}
	}

	if len(article.Categories) > 0 {
		tags["article:section"] = article.Categories[0]
	}

	// Locale
	if s.siteConfig.Language != "" {
		tags["og:locale"] = s.siteConfig.Language
	}

	s.logger.Debug("Generated Open Graph tags",
		"slug", article.Slug,
		"tags_count", len(tags))

	return tags, nil
}

// GenerateTwitterCardTags creates Twitter Card meta tags
func (s *Service) GenerateTwitterCardTags(article *models.Article, baseURL string) (map[string]string, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	if article == nil {
		return nil, fmt.Errorf("article cannot be nil")
	}

	tags := make(map[string]string)

	// Determine card type based on content
	imageURL := s.extractFirstImage(article.Content, baseURL)
	if imageURL != "" {
		tags["twitter:card"] = "summary_large_image"
		tags["twitter:image"] = imageURL
		tags["twitter:image:alt"] = fmt.Sprintf("Featured image for %s", article.Title)
	} else {
		tags["twitter:card"] = twitterCardSummary
	}

	// Basic Twitter tags
	tags["twitter:title"] = article.Title

	// Description
	description := article.Description
	if description == "" {
		description = s.generateExcerpt(article.Content, 160)
	}
	if description != "" {
		tags["twitter:description"] = description
	}

	// Site information
	tags["twitter:site"] = s.siteConfig.Name

	// Author (if available as Twitter handle)
	if article.Author != "" {
		authorHandle := s.extractTwitterHandle(article.Author)
		if authorHandle != "" {
			tags["twitter:creator"] = authorHandle
		}
	}

	s.logger.Debug("Generated Twitter Card tags",
		"slug", article.Slug,
		"card_type", tags["twitter:card"],
		"tags_count", len(tags))

	return tags, nil
}

// GenerateMetaTags creates general SEO meta tags
func (s *Service) GenerateMetaTags(article *models.Article, siteConfig services.SiteConfig) (map[string]string, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	if article == nil {
		return nil, fmt.Errorf("article cannot be nil")
	}

	tags := make(map[string]string)

	// Title
	tags["title"] = article.Title

	// Description
	description := article.Description
	if description == "" {
		description = s.generateExcerpt(article.Content, 160)
	}
	if description != "" {
		tags["description"] = description
	}

	// Keywords from tags
	if len(article.Tags) > 0 {
		tags["keywords"] = strings.Join(article.Tags, ", ")
	}

	// Author
	authorName := article.Author
	if authorName == "" {
		authorName = siteConfig.Author
	}
	if authorName != "" {
		tags["author"] = authorName
	}

	// Canonical URL
	articleURL, err := s.buildArticleURL(article.Slug)
	if err == nil {
		tags["canonical"] = articleURL
	}

	// Language
	if siteConfig.Language != "" {
		tags["language"] = siteConfig.Language
	}

	// Publication date
	tags["article:published_time"] = article.Date.Format("2006-01-02")

	// Robots meta tag
	robotsValue := robotsIndexFollow
	if article.Draft {
		robotsValue = "noindex, nofollow"
	}
	tags["robots"] = robotsValue

	// Reading time
	if article.ReadingTime > 0 {
		tags["reading_time"] = strconv.Itoa(article.ReadingTime)
	}

	// Word count
	if article.WordCount > 0 {
		tags["word_count"] = strconv.Itoa(article.WordCount)
	}

	s.logger.Debug("Generated meta tags",
		"slug", article.Slug,
		"tags_count", len(tags))

	return tags, nil
}

// extractTwitterHandle extracts Twitter handle from author name if present
func (s *Service) extractTwitterHandle(author string) string {
	// Look for @username pattern
	if strings.Contains(author, "@") {
		parts := strings.Fields(author)
		for _, part := range parts {
			if strings.HasPrefix(part, "@") {
				handle := strings.TrimPrefix(part, "@")
				// Validate Twitter handle format
				if len(handle) > 0 && len(handle) <= 15 {
					return "@" + handle
				}
			}
		}
	}
	return ""
}

// GeneratePageMetaTags creates meta tags for non-article pages
func (s *Service) GeneratePageMetaTags(title, description, path string, siteConfig services.SiteConfig) (map[string]string, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	tags := make(map[string]string)

	// Basic meta tags
	if title != "" {
		if title != siteConfig.Name {
			tags["title"] = fmt.Sprintf("%s - %s", title, siteConfig.Name)
		} else {
			tags["title"] = title
		}
	}

	if description != "" {
		tags["description"] = description
	} else if siteConfig.Description != "" {
		tags["description"] = siteConfig.Description
	}

	// Canonical URL
	if path != "" && siteConfig.BaseURL != "" {
		canonicalURL := strings.TrimRight(siteConfig.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
		tags["canonical"] = canonicalURL
	}

	// Language
	if siteConfig.Language != "" {
		tags["language"] = siteConfig.Language
	}

	// Robots
	tags["robots"] = robotsIndexFollow

	// Open Graph for pages
	tags["og:type"] = "website"
	if title != "" {
		tags["og:title"] = title
	}
	if description != "" {
		tags["og:description"] = description
	}
	tags["og:site_name"] = siteConfig.Name

	// Twitter Card for pages
	tags["twitter:card"] = twitterCardSummary
	if title != "" {
		tags["twitter:title"] = title
	}
	if description != "" {
		tags["twitter:description"] = description
	}

	s.logger.Debug("Generated page meta tags",
		"title", title,
		"path", path,
		"tags_count", len(tags))

	return tags, nil
}
