package seo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

func TestArticleSchemaGeneration(t *testing.T) {
	service, _ := createTestService()

	article := &models.Article{
		Slug:        "test-article",
		Title:       "Test Article Title",
		Description: "Test article description",
		Date:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Tags:        []string{"go", "testing"},
		Categories:  []string{"programming"},
		Author:      "John Doe",
		Content:     "![Test Image](test.jpg)\n\nThis is test content.",
		WordCount:   100,
	}

	schema, err := service.GenerateArticleSchema(article, "https://example.com")
	if err != nil {
		t.Fatalf("Failed to generate article schema: %v", err)
	}

	// Validate required fields
	if schema["@context"] != "https://schema.org" {
		t.Error("Missing or incorrect @context")
	}

	if schema["@type"] != "Article" {
		t.Error("Missing or incorrect @type")
	}

	if schema["headline"] != article.Title {
		t.Error("Incorrect headline")
	}

	if schema["url"] != "https://example.com/article/test-article" {
		t.Error("Incorrect URL")
	}

	// Check author structure
	author, ok := schema["author"].(map[string]interface{})
	if !ok {
		t.Fatal("Author should be a map")
	}

	if author["@type"] != "Person" {
		t.Error("Author type should be Person")
	}

	if author["name"] != "John Doe" {
		t.Error("Incorrect author name")
	}

	// Validate JSON structure
	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Schema should be valid JSON: %v", err)
	}

	// Ensure valid JSON-LD
	var jsonLD map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonLD); err != nil {
		t.Fatalf("Generated schema should be valid JSON-LD: %v", err)
	}
}

func TestWebsiteSchemaGeneration(t *testing.T) {
	service, _ := createTestService()

	siteConfig := services.SiteConfig{
		Name:        "Test Blog",
		Description: "A test blog",
		BaseURL:     "https://example.com",
		Author:      "Site Author",
	}

	schema, err := service.GenerateWebsiteSchema(siteConfig)
	if err != nil {
		t.Fatalf("Failed to generate website schema: %v", err)
	}

	if schema["@type"] != "WebSite" {
		t.Error("Schema type should be WebSite")
	}

	if schema["name"] != siteConfig.Name {
		t.Error("Incorrect site name")
	}

	if schema["url"] != siteConfig.BaseURL {
		t.Error("Incorrect site URL")
	}

	// Check for search action
	if action, ok := schema["potentialAction"]; ok {
		actionMap := action.(map[string]interface{})
		if actionMap["@type"] != "SearchAction" {
			t.Error("Search action type should be SearchAction")
		}
	}
}

func TestBreadcrumbSchemaGeneration(t *testing.T) {
	service, _ := createTestService()

	breadcrumbs := []services.Breadcrumb{
		{Name: "Home", URL: "https://example.com"},
		{Name: "Category", URL: "https://example.com/category"},
		{Name: "Article", URL: "https://example.com/article/test"},
	}

	schema, err := service.GenerateBreadcrumbSchema(breadcrumbs)
	if err != nil {
		t.Fatalf("Failed to generate breadcrumb schema: %v", err)
	}

	if schema["@type"] != "BreadcrumbList" {
		t.Error("Schema type should be BreadcrumbList")
	}

	itemList, ok := schema["itemListElement"].([]map[string]interface{})
	if !ok {
		t.Fatal("itemListElement should be a slice")
	}

	if len(itemList) != 3 {
		t.Errorf("Expected 3 breadcrumb items, got %d", len(itemList))
	}

	// Check first item
	if itemList[0]["position"] != 1 {
		t.Error("First item position should be 1")
	}

	if itemList[0]["name"] != "Home" {
		t.Error("First item name should be Home")
	}
}

func TestImageExtraction(t *testing.T) {
	service, _ := createTestService()

	content := "Some text\n\n![Alt text](https://example.com/image.jpg)\n\nMore text"
	imageURL := service.extractFirstImage(content, "https://example.com")

	if imageURL != "https://example.com/image.jpg" {
		t.Errorf("Expected image URL, got %s", imageURL)
	}

	// Test relative image
	content = "![Alt text](/images/test.jpg)"
	imageURL = service.extractFirstImage(content, "https://example.com")

	if imageURL != "https://example.com/images/test.jpg" {
		t.Errorf("Expected absolute image URL, got %s", imageURL)
	}

	// Test no image
	content = "No images here"
	imageURL = service.extractFirstImage(content, "https://example.com")

	if imageURL != "" {
		t.Errorf("Expected empty string for no images, got %s", imageURL)
	}
}
