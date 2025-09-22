package seo

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// Service implements the SEOServiceInterface for comprehensive SEO automation
type Service struct {
	articleService services.ArticleServiceInterface
	logger         *slog.Logger
	siteConfig     services.SiteConfig
	robotsConfig   services.RobotsConfig

	// Sitemap caching
	sitemapCache   []byte
	sitemapLastMod time.Time
	sitemapMutex   sync.RWMutex

	// Service state
	enabled bool
	running bool
	mutex   sync.RWMutex
}

// URLSet represents the root sitemap XML structure
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// URL represents a single URL entry in sitemap
type URL struct {
	Location     string `xml:"loc"`
	LastModified string `xml:"lastmod,omitempty"`
	ChangeFreq   string `xml:"changefreq,omitempty"`
	Priority     string `xml:"priority,omitempty"`
}

// NewService creates a new SEO service instance
func NewService(
	articleService services.ArticleServiceInterface,
	siteConfig services.SiteConfig,
	robotsConfig services.RobotsConfig,
	logger *slog.Logger,
	enabled bool,
) *Service {
	return &Service{
		articleService: articleService,
		logger:         logger,
		siteConfig:     siteConfig,
		robotsConfig:   robotsConfig,
		enabled:        enabled,
		running:        false,
	}
}

// Start initializes the SEO service
func (s *Service) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.enabled {
		s.logger.Info("SEO service disabled, skipping start")
		return nil
	}

	if s.running {
		return fmt.Errorf("SEO service already running")
	}

	s.logger.Info("Starting SEO service")

	// Generate initial sitemap
	if err := s.RefreshSitemap(); err != nil {
		s.logger.Error("Failed to generate initial sitemap", "error", err)
		return fmt.Errorf("failed to generate initial sitemap: %w", err)
	}

	s.running = true
	s.logger.Info("SEO service started successfully")
	return nil
}

// Stop gracefully shuts down the SEO service
func (s *Service) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	s.logger.Info("Stopping SEO service")
	s.running = false

	// Clear cache
	s.sitemapMutex.Lock()
	s.sitemapCache = nil
	s.sitemapMutex.Unlock()

	s.logger.Info("SEO service stopped")
	return nil
}

// IsEnabled returns whether the SEO service is enabled
func (s *Service) IsEnabled() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.enabled
}

// GenerateSitemap creates an XML sitemap from all published articles
func (s *Service) GenerateSitemap() ([]byte, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	s.logger.Debug("Generating sitemap")

	// Get all published articles
	articles := s.articleService.GetAllArticles()
	if articles == nil {
		return nil, fmt.Errorf("failed to retrieve articles")
	}

	// Filter out drafts and sort by date
	publishedArticles := make([]*models.Article, 0, len(articles))
	for _, article := range articles {
		if !article.Draft {
			publishedArticles = append(publishedArticles, article)
		}
	}

	// Sort by date (newest first)
	sort.Slice(publishedArticles, func(i, j int) bool {
		return publishedArticles[i].Date.After(publishedArticles[j].Date)
	})

	// Create URL set
	urlSet := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]URL, 0, len(publishedArticles)+3), // +3 for homepage, about, etc.
	}

	// Add homepage
	urlSet.URLs = append(urlSet.URLs, URL{
		Location:   s.siteConfig.BaseURL,
		ChangeFreq: "daily",
		Priority:   "1.0",
	})

	// Add articles
	for _, article := range publishedArticles {
		articleURL, err := s.buildArticleURL(article.Slug)
		if err != nil {
			s.logger.Warn("Failed to build URL for article", "slug", article.Slug, "error", err)
			continue
		}

		priority := "0.8"
		if article.Featured {
			priority = "0.9"
		}

		changeFreq := "monthly"
		if article.Date.After(time.Now().AddDate(0, -1, 0)) {
			changeFreq = "weekly"
		}

		urlSet.URLs = append(urlSet.URLs, URL{
			Location:     articleURL,
			LastModified: article.Date.Format("2006-01-02"),
			ChangeFreq:   changeFreq,
			Priority:     priority,
		})
	}

	// Generate XML
	xmlData, err := xml.MarshalIndent(urlSet, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sitemap XML: %w", err)
	}

	// Add XML header
	result := append([]byte(xml.Header), xmlData...)

	s.logger.Info("Sitemap generated successfully",
		"articles", len(publishedArticles),
		"total_urls", len(urlSet.URLs),
		"size_bytes", len(result))

	return result, nil
}

// RefreshSitemap regenerates and caches the sitemap
func (s *Service) RefreshSitemap() error {
	if !s.enabled {
		return fmt.Errorf("SEO service not enabled")
	}

	sitemap, err := s.GenerateSitemap()
	if err != nil {
		return fmt.Errorf("failed to generate sitemap: %w", err)
	}

	s.sitemapMutex.Lock()
	s.sitemapCache = sitemap
	s.sitemapLastMod = time.Now()
	s.sitemapMutex.Unlock()

	s.logger.Debug("Sitemap cache refreshed")
	return nil
}

// GetSitemapLastModified returns when the sitemap was last generated
func (s *Service) GetSitemapLastModified() time.Time {
	s.sitemapMutex.RLock()
	defer s.sitemapMutex.RUnlock()
	return s.sitemapLastMod
}

// buildArticleURL constructs a complete URL for an article
func (s *Service) buildArticleURL(slug string) (string, error) {
	baseURL, err := url.Parse(s.siteConfig.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	articleURL, err := baseURL.Parse("/article/" + slug)
	if err != nil {
		return "", fmt.Errorf("failed to build article URL: %w", err)
	}

	return articleURL.String(), nil
}

// GenerateRobotsTxt creates a robots.txt file based on configuration
func (s *Service) GenerateRobotsTxt(config services.RobotsConfig) ([]byte, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	var builder strings.Builder

	// User-agent directive
	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = "*"
	}
	builder.WriteString(fmt.Sprintf("User-agent: %s\n", userAgent))

	// Allow directives
	for _, allow := range config.Allow {
		builder.WriteString(fmt.Sprintf("Allow: %s\n", allow))
	}

	// Disallow directives
	for _, disallow := range config.Disallow {
		builder.WriteString(fmt.Sprintf("Disallow: %s\n", disallow))
	}

	// Crawl delay
	if config.CrawlDelay > 0 {
		builder.WriteString(fmt.Sprintf("Crawl-delay: %d\n", config.CrawlDelay))
	}

	// Sitemap URL
	if config.SitemapURL != "" {
		builder.WriteString(fmt.Sprintf("\nSitemap: %s\n", config.SitemapURL))
	}

	s.logger.Debug("Generated robots.txt", "size", builder.Len())
	return []byte(builder.String()), nil
}

// AnalyzeContent performs basic SEO analysis on article content
func (s *Service) AnalyzeContent(content string) (*services.SEOAnalysis, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	analysis := &services.SEOAnalysis{
		Keywords:    make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// Word count and reading time
	words := strings.Fields(content)
	analysis.WordCount = len(words)
	analysis.ReadingTime = analysis.WordCount / 200 // 200 WPM average

	// Count headings (# ## ### etc.)
	headingRegex := regexp.MustCompile(`(?m)^#{1,6}\s+.+$`)
	headings := headingRegex.FindAllString(content, -1)
	analysis.HeadingCount = len(headings)

	// Count images
	imageRegex := regexp.MustCompile(`!\[.*?\]\(.*?\)`)
	images := imageRegex.FindAllString(content, -1)
	analysis.ImageCount = len(images)

	// Count links
	linkRegex := regexp.MustCompile(`\[.*?\]\(.*?\)`)
	links := linkRegex.FindAllString(content, -1)
	analysis.LinkCount = len(links) - analysis.ImageCount // Subtract images

	// Basic SEO scoring
	score := 0.0

	// Word count scoring
	if analysis.WordCount >= 300 {
		score += 20.0
	} else if analysis.WordCount >= 150 {
		score += 10.0
	}

	// Heading scoring
	if analysis.HeadingCount >= 2 {
		score += 15.0
	} else if analysis.HeadingCount >= 1 {
		score += 8.0
	}

	// Image scoring
	if analysis.ImageCount >= 1 {
		score += 10.0
	}

	// Reading time scoring
	if analysis.ReadingTime >= 2 && analysis.ReadingTime <= 10 {
		score += 15.0
	}

	analysis.Score = score

	// Generate suggestions
	if analysis.WordCount < 300 {
		analysis.Suggestions = append(analysis.Suggestions, "Consider adding more content (minimum 300 words recommended)")
	}
	if analysis.HeadingCount == 0 {
		analysis.Suggestions = append(analysis.Suggestions, "Add headings to improve content structure")
	}
	if analysis.ImageCount == 0 {
		analysis.Suggestions = append(analysis.Suggestions, "Consider adding images to enhance visual appeal")
	}
	if analysis.ReadingTime > 15 {
		analysis.Suggestions = append(analysis.Suggestions, "Article might be too long; consider breaking into multiple parts")
	}

	s.logger.Debug("Content analysis completed",
		"word_count", analysis.WordCount,
		"score", analysis.Score,
		"suggestions", len(analysis.Suggestions))

	return analysis, nil
}

// GetPerformanceMetrics returns SEO service metrics
func (s *Service) GetPerformanceMetrics() (*services.SEOMetrics, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SEO service not enabled")
	}

	s.sitemapMutex.RLock()
	sitemapSize := len(s.sitemapCache)
	lastGenerated := s.sitemapLastMod
	s.sitemapMutex.RUnlock()

	// Count published articles
	articles := s.articleService.GetAllArticles()
	articlesIndexed := 0
	for _, article := range articles {
		if !article.Draft {
			articlesIndexed++
		}
	}

	metrics := &services.SEOMetrics{
		SitemapSize:        sitemapSize,
		LastGenerated:      lastGenerated,
		ArticlesIndexed:    articlesIndexed,
		AvgContentScore:    0.0,  // TODO: Calculate from analysis cache
		SchemaValidation:   true, // TODO: Implement schema validation
		OpenGraphEnabled:   s.enabled,
		TwitterCardEnabled: s.enabled,
	}

	s.logger.Debug("Generated SEO metrics",
		"sitemap_size", metrics.SitemapSize,
		"articles_indexed", metrics.ArticlesIndexed)

	return metrics, nil
}
