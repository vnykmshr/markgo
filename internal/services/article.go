package services

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"

	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/utils"
)

// Ensure ArticleService implements ArticleServiceInterface and ArticleProcessor
var _ ArticleServiceInterface = (*ArticleService)(nil)
var _ models.ArticleProcessor = (*ArticleService)(nil)

// Pre-compiled regexes for excerpt generation (memory optimization)
var (
	codeBlockRe = regexp.MustCompile("```[\\s\\S]*?```")
	linkRe      = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	imageRe     = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)
)

type ArticleService struct {
	articlesPath string
	logger       *slog.Logger
	cache        map[string]*models.Article
	articles     []*models.Article
	mutex        sync.RWMutex
	markdown     goldmark.Markdown
	lastReload   time.Time
	tagInterner  *utils.TagInterner
}

func NewArticleService(articlesPath string, logger *slog.Logger) (*ArticleService, error) {
	service := &ArticleService{
		articlesPath: articlesPath,
		logger:       logger,
		cache:        make(map[string]*models.Article),
		articles:     make([]*models.Article, 0),
		tagInterner:  utils.NewTagInterner(),
		markdown: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.Table,
				extension.Strikethrough,
				extension.TaskList,
				extension.Linkify,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				html.WithXHTML(),
				html.WithUnsafe(),
			),
		),
	}

	if err := service.loadArticles(); err != nil {
		return nil, apperrors.NewArticleError("", "Failed to load articles from directory", err)
	}

	return service, nil
}

func (s *ArticleService) loadArticles() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	articles := make([]*models.Article, 0)
	cache := make(map[string]*models.Article)

	err := filepath.WalkDir(s.articlesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".markdown") {
			return nil
		}

		article, err := s.ParseArticleFile(path)
		if err != nil {
			s.logger.Warn("Failed to parse article", "file", path, "error", err)
			return nil // Continue processing other files
		}

		if !article.Draft {
			articles = append(articles, article)
		}
		cache[article.Slug] = article

		return nil
	})

	if err != nil {
		return err
	}

	// Sort articles by date (newest first)
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date.After(articles[j].Date)
	})

	s.articles = articles
	s.cache = cache
	s.lastReload = time.Now()

	s.logger.Info("Loaded articles", "count", len(articles), "total_with_drafts", len(cache))
	return nil
}

func (s *ArticleService) ParseArticleFile(filePath string) (*models.Article, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Get file info for last modified time
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Extract slug from filename
	filename := filepath.Base(filePath)
	slug := strings.TrimSuffix(filename, ".markdown")

	article, err := s.ParseArticle(slug, string(content))
	if err != nil {
		return nil, err
	}

	article.LastModified = fileInfo.ModTime()
	return article, nil
}

func (s *ArticleService) ParseArticle(slug, content string) (*models.Article, error) {
	article := &models.Article{
		Slug: slug,
		Date: time.Now(),
	}

	// Split front matter and content
	parts := strings.SplitN(content, "---", 3)
	if len(parts) >= 3 && strings.TrimSpace(parts[0]) == "" {
		// Has front matter
		frontMatter := strings.TrimSpace(parts[1])
		content = strings.TrimSpace(parts[2])

		if err := yaml.Unmarshal([]byte(frontMatter), article); err != nil {
			return nil, apperrors.NewArticleError(slug, "Invalid YAML front matter", apperrors.ErrInvalidFrontMatter)
		}
	}

	// Intern strings for memory optimization
	article.Tags = utils.InternStringSlice(article.Tags)
	article.Categories = utils.InternStringSlice(article.Categories)
	article.Author = utils.InternString(article.Author)

	// Set default title if not provided
	if article.Title == "" {
		article.Title = s.slugToTitle(slug)
	}

	// Store markdown content for lazy processing
	article.Content = content

	// Set the processor for lazy content processing
	article.SetProcessor(s)

	// Calculate word count and reading time
	wordCount := len(strings.Fields(content))
	article.WordCount = wordCount
	article.ReadingTime = max(1, wordCount/200) // Average 200 words per minute

	return article, nil
}

func (s *ArticleService) slugToTitle(slug string) string {
	// Convert slug to title case
	words := strings.Split(slug, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// ProcessMarkdown implements ArticleProcessor interface for lazy content processing
func (s *ArticleService) ProcessMarkdown(content string) (string, error) {
	// Parse markdown content to HTML
	buf := utils.GetStringBuilder()
	defer utils.PutStringBuilder(buf)

	if err := s.markdown.Convert([]byte(content), buf); err != nil {
		return "", apperrors.NewArticleError("", "Failed to convert markdown to HTML", err)
	}

	// Note: For duplicate title processing, we need article title context
	// This will be handled in the Article.GetProcessedContent method
	return buf.String(), nil
}

// GenerateExcerpt implements ArticleProcessor interface for lazy excerpt generation
func (s *ArticleService) GenerateExcerpt(content string, maxLength int) string {
	return s.generateExcerpt(content, maxLength)
}

// ProcessDuplicateTitles implements ArticleProcessor interface for duplicate title handling
func (s *ArticleService) ProcessDuplicateTitles(title, htmlContent string) string {
	return s.processDuplicateTitles(title, htmlContent)
}

func (s *ArticleService) generateExcerpt(content string, maxLength int) string {
	// Step 1: Clean markdown syntax properly
	cleaned := s.cleanMarkdown(content)

	// Step 2: Extract meaningful paragraphs
	paragraphs := s.extractParagraphs(cleaned)

	// Step 3: Build excerpt from paragraphs
	excerpt := s.buildExcerpt(paragraphs, maxLength)

	// Step 4: Final cleanup and formatting
	return s.finalizeExcerpt(excerpt, maxLength)
}

func (s *ArticleService) cleanMarkdown(content string) string {
	// Use pre-compiled regex for better memory efficiency
	content = codeBlockRe.ReplaceAllString(content, "")

	// Remove inline code (`code`) but preserve surrounding spaces properly
	inlineCodeRe := regexp.MustCompile("`[^`]*`")
	content = inlineCodeRe.ReplaceAllString(content, " ")

	// Handle links properly [text](url) -> text
	content = linkRe.ReplaceAllString(content, "$1")

	// Remove headers completely - they're usually not good for excerpts
	headerRe := regexp.MustCompile(`(?m)^#{1,6}\s+.*$`)
	content = headerRe.ReplaceAllString(content, "")

	// Remove emphasis markers but keep content (**bold** -> bold, *italic* -> italic)
	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	content = boldRe.ReplaceAllString(content, "$1")
	italicRe := regexp.MustCompile(`\*([^*]+)\*`)
	content = italicRe.ReplaceAllString(content, "$1")
	underlineRe := regexp.MustCompile(`_([^_]+)_`)
	content = underlineRe.ReplaceAllString(content, "$1")

	// Remove image syntax ![alt](url) -> alt
	content = imageRe.ReplaceAllString(content, "$1")

	// Remove horizontal rules and other markdown syntax
	content = regexp.MustCompile(`(?m)^---+\s*$`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`(?m)^\*\*\*+\s*$`).ReplaceAllString(content, "")

	// Remove bullet points and list markers (multiline mode)
	content = regexp.MustCompile(`(?m)^\s*[-*+]\s+`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`(?m)^\s*\d+\.\s+`).ReplaceAllString(content, "")

	return content
}

func (s *ArticleService) extractParagraphs(content string) []string {
	// Split by double newlines to get paragraphs
	paragraphs := regexp.MustCompile(`\n\s*\n`).Split(content, -1)

	var meaningful []string
	for _, para := range paragraphs {
		// Clean up whitespace
		para = strings.TrimSpace(para)
		// Replace multiple spaces/newlines with single space and trim
		para = regexp.MustCompile(`\s+`).ReplaceAllString(para, " ")
		para = strings.TrimSpace(para)

		// Skip empty paragraphs or very short ones
		if len(para) < 30 {
			continue
		}

		// Skip common intro phrases that don't add value to excerpts
		lowerPara := strings.ToLower(para)
		skipPhrases := []string{
			"in this article",
			"this article",
			"this post",
			"today we",
			"in this tutorial",
			"this tutorial",
			"welcome to",
			"let's explore",
			"let's dive",
		}

		shouldSkip := false
		for _, phrase := range skipPhrases {
			if strings.HasPrefix(lowerPara, phrase) {
				shouldSkip = true
				break
			}
		}

		if !shouldSkip && len(para) > 0 {
			meaningful = append(meaningful, para)
		}
	}

	return meaningful
}

func (s *ArticleService) buildExcerpt(paragraphs []string, maxLength int) string {
	if len(paragraphs) == 0 {
		return ""
	}

	// Use pooled string builder for memory efficiency
	excerpt := utils.GetStringBuilder()
	defer utils.PutStringBuilder(excerpt)

	for _, para := range paragraphs {
		// Check if adding this paragraph would exceed limit
		if excerpt.Len() > 0 && excerpt.Len()+len(para)+1 > maxLength {
			break
		}

		if excerpt.Len() > 0 {
			excerpt.WriteString(" ")
		}
		excerpt.WriteString(para)

		// If we have a good amount of content, we can stop
		if excerpt.Len() > maxLength/2 {
			break
		}
	}

	result := excerpt.String()
	return result
}

func (s *ArticleService) finalizeExcerpt(excerpt string, maxLength int) string {
	if len(excerpt) <= maxLength {
		return excerpt
	}

	// Try to truncate at sentence boundary first
	truncated := excerpt[:maxLength]

	// Look for sentence endings (.!?) in the last quarter of the truncated text
	searchStart := len(truncated) * 3 / 4
	if searchStart < 0 {
		searchStart = 0
	}

	sentenceEnd := -1
	for i := len(truncated) - 1; i >= searchStart; i-- {
		if truncated[i] == '.' || truncated[i] == '!' || truncated[i] == '?' {
			// Make sure it's not an abbreviation (simple check)
			if i < len(truncated)-1 && unicode.IsUpper(rune(truncated[i+1])) {
				continue
			}
			sentenceEnd = i
			break
		}
	}

	if sentenceEnd > 0 {
		return truncated[:sentenceEnd+1]
	}

	// Fall back to word boundary
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLength/2 {
		truncated = truncated[:lastSpace]
	}

	// Clean up any trailing punctuation that might look odd
	truncated = strings.TrimRight(truncated, ".,;:-")

	return truncated + "..."
}

// Public methods

func (s *ArticleService) GetAllArticles() []*models.Article {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to prevent external modification
	articles := make([]*models.Article, len(s.articles))
	copy(articles, s.articles)
	return articles
}

func (s *ArticleService) GetArticleBySlug(slug string) (*models.Article, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	article, exists := s.cache[slug]
	if !exists {
		return nil, apperrors.NewArticleError(slug, fmt.Sprintf("Article '%s' does not exist", slug), apperrors.ErrArticleNotFound)
	}

	return article, nil
}

func (s *ArticleService) GetArticlesByTag(tag string) []*models.Article {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var filtered []*models.Article
	for _, article := range s.articles {
		for _, articleTag := range article.Tags {
			if strings.EqualFold(articleTag, tag) {
				filtered = append(filtered, article)
				break
			}
		}
	}

	return filtered
}

func (s *ArticleService) GetArticlesByCategory(category string) []*models.Article {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var filtered []*models.Article
	for _, article := range s.articles {
		for _, articleCategory := range article.Categories {
			if strings.EqualFold(articleCategory, category) {
				filtered = append(filtered, article)
				break
			}
		}
	}

	return filtered
}

func (s *ArticleService) GetAllTags() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tagSet := make(map[string]struct{})
	for _, article := range s.articles {
		for _, tag := range article.Tags {
			tagSet[tag] = struct{}{}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Strings(tags)
	return tags
}

func (s *ArticleService) GetAllCategories() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	categorySet := make(map[string]struct{})
	for _, article := range s.articles {
		for _, category := range article.Categories {
			categorySet[category] = struct{}{}
		}
	}

	categories := make([]string, 0, len(categorySet))
	for category := range categorySet {
		categories = append(categories, category)
	}

	sort.Strings(categories)
	return categories
}

func (s *ArticleService) GetTagCounts() []models.TagCount {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tagCounts := make(map[string]int)
	for _, article := range s.articles {
		for _, tag := range article.Tags {
			tagCounts[tag]++
		}
	}

	var result []models.TagCount
	for tag, count := range tagCounts {
		result = append(result, models.TagCount{
			Tag:   tag,
			Count: count,
		})
	}

	// Sort by count (descending) then by name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Tag < result[j].Tag
		}
		return result[i].Count > result[j].Count
	})

	return result
}

func (s *ArticleService) GetCategoryCounts() []models.CategoryCount {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	categoryCounts := make(map[string]int)
	for _, article := range s.articles {
		for _, category := range article.Categories {
			categoryCounts[category]++
		}
	}

	var result []models.CategoryCount
	for category, count := range categoryCounts {
		result = append(result, models.CategoryCount{
			Category: category,
			Count:    count,
		})
	}

	// Sort by count (descending) then by name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Category < result[j].Category
		}
		return result[i].Count > result[j].Count
	})

	return result
}

func (s *ArticleService) GetFeaturedArticles(limit int) []*models.Article {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var featured []*models.Article
	for _, article := range s.articles {
		if article.Featured {
			featured = append(featured, article)
			if len(featured) >= limit {
				break
			}
		}
	}

	return featured
}

func (s *ArticleService) GetRecentArticles(limit int) []*models.Article {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.articles) <= limit {
		return s.articles
	}

	return s.articles[:limit]
}

func (s *ArticleService) GetStats() *models.Stats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	totalArticles := len(s.cache)
	publishedCount := len(s.articles)
	draftCount := totalArticles - publishedCount

	tagCounts := s.GetTagCounts()
	popularTags := tagCounts
	if len(popularTags) > 10 {
		popularTags = popularTags[:10]
	}

	recentArticles := s.GetRecentArticles(5)
	recentList := make([]*models.ArticleList, len(recentArticles))
	for i, article := range recentArticles {
		recentList[i] = article.ToListView()
	}

	return &models.Stats{
		TotalArticles:   totalArticles,
		PublishedCount:  publishedCount,
		DraftCount:      draftCount,
		TotalTags:       len(s.GetAllTags()),
		TotalCategories: len(s.GetAllCategories()),
		PopularTags:     popularTags,
		RecentArticles:  recentList,
		LastUpdated:     s.lastReload,
	}
}

func (s *ArticleService) ReloadArticles() error {
	s.logger.Info("Reloading articles...")
	return s.loadArticles()
}

func (s *ArticleService) GetArticlesForFeed(limit int) []*models.Article {
	articles := s.GetAllArticles()
	if len(articles) <= limit {
		return articles
	}
	return articles[:limit]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// processDuplicateTitles detects and handles duplicate titles in HTML content
func (s *ArticleService) processDuplicateTitles(title, htmlContent string) string {
	if title == "" {
		return htmlContent
	}

	// Check if the first heading in content matches the title
	firstHeading := s.extractFirstHeading(htmlContent)
	if firstHeading != "" && s.normalizeText(firstHeading) == s.normalizeText(title) {
		// Demote the first heading from h1 to h2
		return s.demoteFirstHeading(htmlContent)
	}

	return htmlContent
}

// extractFirstHeading extracts the text content of the first h1 heading from HTML
func (s *ArticleService) extractFirstHeading(content string) string {
	// Match the first h1 tag and extract its content
	h1Re := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
	matches := h1Re.FindStringSubmatch(content)
	if len(matches) > 1 {
		// Remove any HTML tags from the heading content
		cleanRe := regexp.MustCompile(`<[^>]*>`)
		return strings.TrimSpace(cleanRe.ReplaceAllString(matches[1], ""))
	}
	return ""
}

// normalizeText normalizes text for comparison by removing extra spaces and converting to lowercase
func (s *ArticleService) normalizeText(text string) string {
	// Convert to lowercase and normalize whitespace
	normalized := strings.ToLower(strings.TrimSpace(text))
	spaceRe := regexp.MustCompile(`\s+`)
	return spaceRe.ReplaceAllString(normalized, " ")
}

// demoteFirstHeading converts the first h1 heading to h2
func (s *ArticleService) demoteFirstHeading(content string) string {
	// Replace the first h1 with h2
	h1Re := regexp.MustCompile(`<h1([^>]*)>(.*?)</h1>`)

	// Only replace the first occurrence
	replaced := false
	return h1Re.ReplaceAllStringFunc(content, func(match string) string {
		if replaced {
			return match // Don't replace subsequent h1 tags
		}

		// Extract attributes and content
		matches := h1Re.FindStringSubmatch(match)
		if len(matches) >= 3 {
			attributes := matches[1]
			headingContent := matches[2]
			replaced = true
			return fmt.Sprintf(`<h2%s>%s</h2>`, attributes, headingContent)
		}
		return match
	})
}
