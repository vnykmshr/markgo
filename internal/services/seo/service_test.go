package seo

import (
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// MockArticleService for testing
type MockArticleService struct {
	articles []*models.Article
}

func (m *MockArticleService) GetAllArticles() []*models.Article {
	return m.articles
}

func (m *MockArticleService) GetArticleBySlug(slug string) (*models.Article, error) {
	for _, article := range m.articles {
		if article.Slug == slug {
			return article, nil
		}
	}
	return nil, nil
}

// Implement other required methods as no-ops
func (m *MockArticleService) GetArticlesByTag(_ string) []*models.Article      { return nil }
func (m *MockArticleService) GetArticlesByCategory(_ string) []*models.Article { return nil }
func (m *MockArticleService) GetArticlesForFeed(_ int) []*models.Article       { return nil }
func (m *MockArticleService) GetFeaturedArticles(_ int) []*models.Article      { return nil }
func (m *MockArticleService) GetRecentArticles(_ int) []*models.Article        { return nil }
func (m *MockArticleService) GetAllTags() []string                             { return nil }
func (m *MockArticleService) GetAllCategories() []string                       { return nil }
func (m *MockArticleService) GetTagCounts() []models.TagCount                  { return nil }
func (m *MockArticleService) GetCategoryCounts() []models.CategoryCount        { return nil }
func (m *MockArticleService) GetStats() *models.Stats                          { return nil }
func (m *MockArticleService) ReloadArticles() error                            { return nil }
func (m *MockArticleService) GetDraftArticles() []*models.Article              { return nil }
func (m *MockArticleService) GetDraftBySlug(_ string) (*models.Article, error) { return nil, nil }

func createTestHelper() (*Helper, *MockArticleService) {
	mockArticles := &MockArticleService{
		articles: []*models.Article{
			{
				Slug:        "test-article",
				Title:       "Test Article",
				Description: "Test description",
				Date:        time.Now().Add(-24 * time.Hour),
				Tags:        []string{"test", "article"},
				Categories:  []string{"tech"},
				Draft:       false,
				Featured:    true,
				Author:      "Test Author",
				Content:     "This is test content with more than 150 words...",
				WordCount:   200,
				ReadingTime: 1,
			},
			{
				Slug:  "draft-article",
				Title: "Draft Article",
				Draft: true,
			},
		},
	}

	siteConfig := services.SiteConfig{
		Name:        "Test Blog",
		Description: "Test blog description",
		BaseURL:     "https://example.com",
		Language:    "en",
		Author:      "Test Author",
	}

	robotsConfig := services.RobotsConfig{
		UserAgent:  "*",
		Allow:      []string{"/"},
		Disallow:   []string{"/admin"},
		CrawlDelay: 1,
		SitemapURL: "https://example.com/sitemap.xml",
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	helper := NewHelper(mockArticles, siteConfig, robotsConfig, logger, true)
	return helper, mockArticles
}

func TestSitemapGeneration(t *testing.T) {
	helper, _ := createTestHelper()

	sitemap, err := helper.GenerateSitemap()
	if err != nil {
		t.Fatalf("Failed to generate sitemap: %v", err)
	}

	sitemapStr := string(sitemap)

	// Basic XML validation
	if !strings.Contains(sitemapStr, "<?xml") {
		t.Error("Sitemap missing XML declaration")
	}

	if !strings.Contains(sitemapStr, "<urlset") {
		t.Error("Sitemap missing urlset tag")
	}

	// Check homepage is included
	if !strings.Contains(sitemapStr, "https://example.com</loc>") {
		t.Error("Sitemap missing homepage")
	}

	// Check published article is included
	if !strings.Contains(sitemapStr, "/article/test-article") {
		t.Error("Sitemap missing published article")
	}

	// Check draft article is NOT included
	if strings.Contains(sitemapStr, "draft-article") {
		t.Error("Sitemap should not include draft articles")
	}
}

func TestRobotsGeneration(t *testing.T) {
	helper, _ := createTestHelper()

	robots, err := helper.GenerateRobotsTxt()
	if err != nil {
		t.Fatalf("Failed to generate robots.txt: %v", err)
	}

	robotsStr := string(robots)

	// Check basic structure
	if !strings.Contains(robotsStr, "User-agent: *") {
		t.Error("robots.txt missing user-agent")
	}

	if !strings.Contains(robotsStr, "Allow: /") {
		t.Error("robots.txt missing allow directive")
	}

	if !strings.Contains(robotsStr, "Disallow: /admin") {
		t.Error("robots.txt missing disallow directive")
	}

	if !strings.Contains(robotsStr, "Sitemap: https://example.com/sitemap.xml") {
		t.Error("robots.txt missing sitemap URL")
	}
}

func TestContentAnalysis(t *testing.T) {
	helper, _ := createTestHelper()

	content := "# Test Article\n\nThis is a test article with multiple paragraphs and some **bold** text.\n\n![Image](image.jpg)\n\n[Link](https://example.com)"

	analysis, err := helper.AnalyzeContent(content)
	if err != nil {
		t.Fatalf("Failed to analyze content: %v", err)
	}

	if analysis.WordCount == 0 {
		t.Error("Word count should be greater than 0")
	}

	if analysis.HeadingCount != 1 {
		t.Errorf("Expected 1 heading, got %d", analysis.HeadingCount)
	}

	if analysis.ImageCount != 1 {
		t.Errorf("Expected 1 image, got %d", analysis.ImageCount)
	}

	if analysis.LinkCount != 1 {
		t.Errorf("Expected 1 link, got %d", analysis.LinkCount)
	}

	if analysis.Score <= 0 {
		t.Error("SEO score should be positive")
	}
}

func TestDisabledHelper(t *testing.T) {
	mockArticles := &MockArticleService{}
	siteConfig := services.SiteConfig{BaseURL: "https://example.com"}
	robotsConfig := services.RobotsConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create disabled helper
	helper := NewHelper(mockArticles, siteConfig, robotsConfig, logger, false)

	if helper.IsEnabled() {
		t.Error("Helper should be disabled")
	}

	_, err := helper.GenerateSitemap()
	if err == nil {
		t.Error("Disabled helper should return error for sitemap generation")
	}
}
