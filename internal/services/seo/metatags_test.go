package seo

import (
	"strings"
	"testing"
	"time"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

func TestOpenGraphGeneration(t *testing.T) {
	service, _ := createTestService()

	article := &models.Article{
		Slug:        "test-article",
		Title:       "Test Article",
		Description: "Test description",
		Date:        time.Now(),
		Author:      "Test Author",
		Content:     "![Test](image.jpg)\n\nContent here",
		Tags:        []string{"test", "article"},
		Categories:  []string{"tech"},
	}

	tags, err := service.GenerateOpenGraphTags(article, "https://example.com")
	if err != nil {
		t.Fatalf("Failed to generate Open Graph tags: %v", err)
	}

	// Check required tags
	if tags["og:type"] != "article" {
		t.Error("og:type should be article")
	}

	if tags["og:title"] != article.Title {
		t.Error("og:title should match article title")
	}

	if tags["og:url"] != "https://example.com/article/test-article" {
		t.Error("og:url should be correct")
	}

	if tags["og:site_name"] != "Test Blog" {
		t.Error("og:site_name should match site config")
	}

	if tags["og:description"] != article.Description {
		t.Error("og:description should match article description")
	}

	// Check article-specific tags
	if tags["article:author"] != article.Author {
		t.Error("article:author should match")
	}

	if tags["article:section"] != "tech" {
		t.Error("article:section should be first category")
	}

	// Check image extraction
	if !strings.Contains(tags["og:image"], "image.jpg") {
		t.Error("og:image should contain extracted image")
	}
}

func TestTwitterCardGeneration(t *testing.T) {
	service, _ := createTestService()

	article := &models.Article{
		Slug:        "test-article",
		Title:       "Test Article",
		Description: "Test description",
		Content:     "![Test](image.jpg)\n\nContent here",
	}

	tags, err := service.GenerateTwitterCardTags(article, "https://example.com")
	if err != nil {
		t.Fatalf("Failed to generate Twitter Card tags: %v", err)
	}

	// Should use large image card when image present
	if tags["twitter:card"] != "summary_large_image" {
		t.Error("twitter:card should be summary_large_image when image present")
	}

	if tags["twitter:title"] != article.Title {
		t.Error("twitter:title should match article title")
	}

	if tags["twitter:description"] != article.Description {
		t.Error("twitter:description should match article description")
	}

	// Test without image
	article.Content = "No images here"
	tags, err = service.GenerateTwitterCardTags(article, "https://example.com")
	if err != nil {
		t.Fatalf("Failed to generate Twitter Card tags without image: %v", err)
	}

	if tags["twitter:card"] != "summary" {
		t.Error("twitter:card should be summary when no image")
	}
}

func TestMetaTagGeneration(t *testing.T) {
	service, _ := createTestService()

	article := &models.Article{
		Slug:        "test-article",
		Title:       "Test Article",
		Description: "Test description",
		Tags:        []string{"test", "article"},
		Author:      "Test Author",
		Date:        time.Now(),
		Draft:       false,
		WordCount:   200,
		ReadingTime: 1,
	}

	siteConfig := services.SiteConfig{
		Name:     "Test Blog",
		Author:   "Site Author",
		Language: "en",
	}

	tags, err := service.GenerateMetaTags(article, siteConfig)
	if err != nil {
		t.Fatalf("Failed to generate meta tags: %v", err)
	}

	if tags["title"] != article.Title {
		t.Error("title should match article title")
	}

	if tags["description"] != article.Description {
		t.Error("description should match article description")
	}

	if tags["keywords"] != "test, article" {
		t.Error("keywords should be joined tags")
	}

	if tags["author"] != article.Author {
		t.Error("author should match article author")
	}

	if tags["language"] != "en" {
		t.Error("language should match site config")
	}

	if tags["robots"] != "index, follow" {
		t.Error("robots should be index, follow for published articles")
	}

	// Test draft article
	article.Draft = true
	tags, err = service.GenerateMetaTags(article, siteConfig)
	if err != nil {
		t.Fatalf("Failed to generate meta tags for draft: %v", err)
	}

	if tags["robots"] != "noindex, nofollow" {
		t.Error("robots should be noindex, nofollow for drafts")
	}
}

func TestPageMetaTagGeneration(t *testing.T) {
	service, _ := createTestService()

	siteConfig := services.SiteConfig{
		Name:        "Test Blog",
		Description: "Test blog description",
		BaseURL:     "https://example.com",
		Language:    "en",
	}

	tags, err := service.GeneratePageMetaTags("About", "About page description", "/about", siteConfig)
	if err != nil {
		t.Fatalf("Failed to generate page meta tags: %v", err)
	}

	if tags["title"] != "About - Test Blog" {
		t.Error("title should include site name")
	}

	if tags["description"] != "About page description" {
		t.Error("description should match provided description")
	}

	if tags["canonical"] != "https://example.com/about" {
		t.Error("canonical should be built correctly")
	}

	if tags["og:type"] != "website" {
		t.Error("og:type should be website for pages")
	}

	if tags["twitter:card"] != "summary" {
		t.Error("twitter:card should be summary for pages")
	}
}

func TestTwitterHandleExtraction(t *testing.T) {
	service, _ := createTestService()

	// Test valid handle
	handle := service.extractTwitterHandle("John Doe @johndoe")
	if handle != "@johndoe" {
		t.Errorf("Expected @johndoe, got %s", handle)
	}

	// Test no handle
	handle = service.extractTwitterHandle("John Doe")
	if handle != "" {
		t.Errorf("Expected empty string, got %s", handle)
	}

	// Test invalid handle (too long)
	handle = service.extractTwitterHandle("@verylonghandlethatistolong")
	if handle != "" {
		t.Errorf("Expected empty string for invalid handle, got %s", handle)
	}
}
