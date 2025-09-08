package article

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/vnykmshr/markgo/internal/models"
)

// Repository defines the interface for article data access
type Repository interface {
	// Core CRUD operations
	LoadAll(ctx context.Context) ([]*models.Article, error)
	GetBySlug(slug string) (*models.Article, error)
	GetByTag(tag string) []*models.Article
	GetByCategory(category string) []*models.Article
	GetPublished() []*models.Article
	GetDrafts() []*models.Article
	GetFeatured(limit int) []*models.Article
	GetRecent(limit int) []*models.Article
	
	// File system operations
	Reload(ctx context.Context) error
	GetLastModified() time.Time
	
	// Statistics
	GetStats() *models.Stats
}

// FileSystemRepository implements Repository using file system storage
type FileSystemRepository struct {
	articlesPath string
	logger       *slog.Logger
	articles     []*models.Article
	cache        map[string]*models.Article
	mutex        sync.RWMutex
	lastReload   time.Time
}

// NewFileSystemRepository creates a new file system-based repository
func NewFileSystemRepository(articlesPath string, logger *slog.Logger) *FileSystemRepository {
	return &FileSystemRepository{
		articlesPath: articlesPath,
		logger:       logger,
		cache:        make(map[string]*models.Article),
		articles:     make([]*models.Article, 0),
		lastReload:   time.Now(),
	}
}

// LoadAll loads all articles from the file system
func (r *FileSystemRepository) LoadAll(ctx context.Context) ([]*models.Article, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.logger.Info("Loading articles from file system", "path", r.articlesPath)
	
	var articles []*models.Article
	cache := make(map[string]*models.Article)

	err := filepath.WalkDir(r.articlesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || (!strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".markdown")) {
			return nil
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		article, parseErr := r.parseArticleFile(path)
		if parseErr != nil {
			r.logger.Warn("Failed to parse article", "file", path, "error", parseErr)
			return nil // Continue processing other files
		}

		articles = append(articles, article)
		cache[article.Slug] = article

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load articles: %w", err)
	}

	// Sort articles by date (newest first)
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date.After(articles[j].Date)
	})

	r.articles = articles
	r.cache = cache
	r.lastReload = time.Now()

	r.logger.Info("Articles loaded successfully", "count", len(articles))
	return articles, nil
}

// GetBySlug retrieves an article by its slug
func (r *FileSystemRepository) GetBySlug(slug string) (*models.Article, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	article, exists := r.cache[slug]
	if !exists {
		return nil, fmt.Errorf("article not found: %s", slug)
	}

	return article, nil
}

// GetByTag returns articles that have the specified tag
func (r *FileSystemRepository) GetByTag(tag string) []*models.Article {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*models.Article
	for _, article := range r.articles {
		for _, articleTag := range article.Tags {
			if strings.EqualFold(articleTag, tag) {
				result = append(result, article)
				break
			}
		}
	}

	return result
}

// GetByCategory returns articles in the specified category
func (r *FileSystemRepository) GetByCategory(category string) []*models.Article {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*models.Article
	for _, article := range r.articles {
		for _, articleCategory := range article.Categories {
			if strings.EqualFold(articleCategory, category) {
				result = append(result, article)
				break
			}
		}
	}

	return result
}

// GetPublished returns all published (non-draft) articles
func (r *FileSystemRepository) GetPublished() []*models.Article {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*models.Article
	for _, article := range r.articles {
		if !article.Draft {
			result = append(result, article)
		}
	}

	return result
}

// GetDrafts returns all draft articles
func (r *FileSystemRepository) GetDrafts() []*models.Article {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*models.Article
	for _, article := range r.articles {
		if article.Draft {
			result = append(result, article)
		}
	}

	return result
}

// GetFeatured returns featured articles up to the specified limit
func (r *FileSystemRepository) GetFeatured(limit int) []*models.Article {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*models.Article
	for _, article := range r.articles {
		if article.Featured && !article.Draft {
			result = append(result, article)
			if len(result) >= limit {
				break
			}
		}
	}

	return result
}

// GetRecent returns recent articles up to the specified limit
func (r *FileSystemRepository) GetRecent(limit int) []*models.Article {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*models.Article
	for _, article := range r.articles {
		if !article.Draft {
			result = append(result, article)
			if len(result) >= limit {
				break
			}
		}
	}

	return result
}

// Reload reloads all articles from the file system
func (r *FileSystemRepository) Reload(ctx context.Context) error {
	_, err := r.LoadAll(ctx)
	return err
}

// GetLastModified returns the last reload time
func (r *FileSystemRepository) GetLastModified() time.Time {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.lastReload
}

// GetStats calculates and returns article statistics
func (r *FileSystemRepository) GetStats() *models.Stats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := &models.Stats{
		LastUpdated: time.Now(),
	}

	// Count articles and gather tags/categories
	tagCount := make(map[string]int)
	categoryCount := make(map[string]int)

	for _, article := range r.articles {
		stats.TotalArticles++
		
		if article.Draft {
			stats.DraftCount++
		} else {
			stats.PublishedCount++
		}

		// Count tags
		for _, tag := range article.Tags {
			tagCount[tag]++
		}

		// Count categories
		for _, category := range article.Categories {
			categoryCount[category]++
		}
	}

	stats.TotalTags = len(tagCount)
	stats.TotalCategories = len(categoryCount)

	// Popular tags (top 10)
	type tagCountPair struct {
		tag   string
		count int
	}
	var tagPairs []tagCountPair
	for tag, count := range tagCount {
		tagPairs = append(tagPairs, tagCountPair{tag, count})
	}
	sort.Slice(tagPairs, func(i, j int) bool {
		return tagPairs[i].count > tagPairs[j].count
	})

	maxTags := 10
	if len(tagPairs) < maxTags {
		maxTags = len(tagPairs)
	}
	
	for i := 0; i < maxTags; i++ {
		stats.PopularTags = append(stats.PopularTags, models.TagCount{
			Tag:   tagPairs[i].tag,
			Count: tagPairs[i].count,
		})
	}

	// Recent articles (top 5 published)
	recentCount := 0
	for _, article := range r.articles {
		if !article.Draft && recentCount < 5 {
			stats.RecentArticles = append(stats.RecentArticles, article.ToListView())
			recentCount++
		}
	}

	return stats
}

// parseArticleFile parses a markdown file into an Article model
func (r *FileSystemRepository) parseArticleFile(filePath string) (*models.Article, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Split frontmatter and content
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown file format: missing frontmatter in %s", filePath)
	}

	// Parse frontmatter
	var article models.Article
	if err := yaml.Unmarshal([]byte(parts[1]), &article); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in %s: %w", filePath, err)
	}

	// Set content and basic metadata
	article.Content = strings.TrimSpace(parts[2])
	article.WordCount = len(strings.Fields(article.Content))
	article.ReadingTime = calculateReadingTime(article.WordCount)

	// Get file modification time
	if fileInfo, err := os.Stat(filePath); err == nil {
		article.LastModified = fileInfo.ModTime()
	}

	// Generate slug if not provided
	if article.Slug == "" {
		article.Slug = generateSlug(article.Title)
	}

	return &article, nil
}

// Helper functions

func calculateReadingTime(wordCount int) int {
	const wordsPerMinute = 200
	readingTime := wordCount / wordsPerMinute
	if readingTime == 0 {
		readingTime = 1
	}
	return readingTime
}

func generateSlug(title string) string {
	// Simple slug generation - convert to lowercase, replace spaces with hyphens
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	result := strings.Builder{}
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}