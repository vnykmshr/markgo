package article

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/vnykmshr/obcache-go/pkg/obcache"

	"github.com/vnykmshr/markgo/internal/models"
)

// CacheCoordinator manages caching across all article services
type CacheCoordinator struct {
	obcache *obcache.Cache
	logger  *slog.Logger
	mu      sync.RWMutex

	// Cache settings
	articleTTL time.Duration
	searchTTL  time.Duration
	contentTTL time.Duration

	// Cache keys
	articlePrefix string
	searchPrefix  string
	contentPrefix string
	statsKey      string

	// Cache statistics
	hits   int64
	misses int64
}

// CacheConfig holds configuration for the cache coordinator
type CacheConfig struct {
	MaxEntries    int           `json:"max_entries"`
	ArticleTTL    time.Duration `json:"article_ttl"`
	SearchTTL     time.Duration `json:"search_ttl"`
	ContentTTL    time.Duration `json:"content_ttl"`
	CleanupPeriod time.Duration `json:"cleanup_period"`
}

// DefaultCacheConfig returns a default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxEntries:    5000,
		ArticleTTL:    30 * time.Minute,
		SearchTTL:     10 * time.Minute,
		ContentTTL:    60 * time.Minute,
		CleanupPeriod: 15 * time.Minute,
	}
}

// NewCacheCoordinator creates a new cache coordinator
func NewCacheCoordinator(config *CacheConfig, logger *slog.Logger) (*CacheCoordinator, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	// Initialize obcache with configuration
	cacheConfig := obcache.NewDefaultConfig().
		WithMaxEntries(config.MaxEntries).
		WithDefaultTTL(config.ArticleTTL)

	cache, err := obcache.New(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create obcache: %w", err)
	}

	coordinator := &CacheCoordinator{
		obcache:       cache,
		logger:        logger,
		articleTTL:    config.ArticleTTL,
		searchTTL:     config.SearchTTL,
		contentTTL:    config.ContentTTL,
		articlePrefix: "article:",
		searchPrefix:  "search:",
		contentPrefix: "content:",
		statsKey:      "stats",
	}

	// Start background cleanup if configured
	if config.CleanupPeriod > 0 {
		go coordinator.startCleanup(config.CleanupPeriod)
	}

	return coordinator, nil
}

// Article caching methods

// GetArticle retrieves an article from cache
func (c *CacheCoordinator) GetArticle(slug string) (*models.Article, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.articlePrefix + slug
	value, found := c.obcache.Get(key)
	if !found {
		c.misses++
		return nil, false
	}

	c.hits++
	if article, ok := value.(*models.Article); ok {
		return article, true
	}

	// Invalid type in cache, remove it
	if err := c.obcache.Delete(key); err != nil {
		c.logger.Warn("Failed to delete invalid article from cache", "key", key, "error", err)
	}
	return nil, false
}

// SetArticle stores an article in cache
func (c *CacheCoordinator) SetArticle(slug string, article *models.Article) {
	if article == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.articlePrefix + slug
	if err := c.obcache.Set(key, article, c.articleTTL); err != nil {
		c.logger.Warn("Failed to cache article", "key", key, "error", err)
	}
}

// InvalidateArticle removes an article from cache
func (c *CacheCoordinator) InvalidateArticle(slug string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.articlePrefix + slug
	if err := c.obcache.Delete(key); err != nil {
		c.logger.Warn("Failed to delete article from cache", "key", key, "error", err)
	}
}

// Search result caching methods

// GetSearchResults retrieves search results from cache
func (c *CacheCoordinator) GetSearchResults(query string, limit int) ([]*models.SearchResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := fmt.Sprintf("%s%s:%d", c.searchPrefix, query, limit)
	value, found := c.obcache.Get(key)
	if !found {
		c.misses++
		return nil, false
	}

	c.hits++
	if results, ok := value.([]*models.SearchResult); ok {
		return results, true
	}

	// Invalid type in cache, remove it
	if err := c.obcache.Delete(key); err != nil {
		c.logger.Warn("Failed to delete invalid search results from cache", "key", key, "error", err)
	}
	return nil, false
}

// SetSearchResults stores search results in cache
func (c *CacheCoordinator) SetSearchResults(query string, limit int, results []*models.SearchResult) {
	if results == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s%s:%d", c.searchPrefix, query, limit)
	if err := c.obcache.Set(key, results, c.searchTTL); err != nil {
		c.logger.Warn("Failed to cache search results", "key", key, "error", err)
	}
}

// InvalidateSearchCache clears all search-related cache entries
func (c *CacheCoordinator) InvalidateSearchCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Since obcache doesn't have prefix-based deletion, we need to clear the entire cache
	// In a production environment, you might want to track search keys separately
	c.logger.Debug("Invalidating search cache (clearing all)")
	if err := c.obcache.Clear(); err != nil {
		c.logger.Warn("Failed to clear search cache", "error", err)
	}
}

// Content caching methods

// GetProcessedContent retrieves processed content from cache
func (c *CacheCoordinator) GetProcessedContent(contentHash string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.contentPrefix + contentHash
	value, found := c.obcache.Get(key)
	if !found {
		c.misses++
		return "", false
	}

	c.hits++
	if content, ok := value.(string); ok {
		return content, true
	}

	// Invalid type in cache, remove it
	if err := c.obcache.Delete(key); err != nil {
		c.logger.Warn("Failed to delete invalid content from cache", "key", key, "error", err)
	}
	return "", false
}

// SetProcessedContent stores processed content in cache
func (c *CacheCoordinator) SetProcessedContent(contentHash, processedContent string) {
	if processedContent == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.contentPrefix + contentHash
	if err := c.obcache.Set(key, processedContent, c.contentTTL); err != nil {
		c.logger.Warn("Failed to cache processed content", "key", key, "error", err)
	}
}

// Stats caching methods

// GetStats retrieves stats from cache
func (c *CacheCoordinator) GetStats() (*models.Stats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, found := c.obcache.Get(c.statsKey)
	if !found {
		c.misses++
		return nil, false
	}

	c.hits++
	if stats, ok := value.(*models.Stats); ok {
		return stats, true
	}

	// Invalid type in cache, remove it
	if err := c.obcache.Delete(c.statsKey); err != nil {
		c.logger.Warn("Failed to delete invalid stats from cache", "error", err)
	}
	return nil, false
}

// SetStats stores stats in cache
func (c *CacheCoordinator) SetStats(stats *models.Stats) {
	if stats == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.obcache.Set(c.statsKey, stats, c.articleTTL); err != nil {
		c.logger.Warn("Failed to cache stats", "error", err)
	}
}

// Cache management methods

// InvalidateAll clears all cache entries
func (c *CacheCoordinator) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info("Invalidating all cached data")
	if err := c.obcache.Clear(); err != nil {
		c.logger.Warn("Failed to clear all cache", "error", err)
	}
	c.hits = 0
	c.misses = 0
}

// InvalidateByTag invalidates cache entries related to a specific tag
func (c *CacheCoordinator) InvalidateByTag(tag string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// For now, clear all search cache since we can't efficiently find tag-related entries
	// In a production system, you might want to maintain tag-to-key mappings
	c.logger.Debug("Invalidating cache for tag", "tag", tag)
	if err := c.obcache.Clear(); err != nil { // This is aggressive but safe
		c.logger.Warn("Failed to clear cache for tag", "tag", tag, "error", err)
	}
}

// GetCacheStats returns cache performance statistics
func (c *CacheCoordinator) GetCacheStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	obcacheStats := c.obcache.Stats()

	hitRate := float64(0)
	total := c.hits + c.misses
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return map[string]interface{}{
		"hits":     c.hits,
		"misses":   c.misses,
		"hit_rate": hitRate,
		"obcache_stats": map[string]interface{}{
			"key_count":        obcacheStats.KeyCount(),
			"obcache_hits":     obcacheStats.Hits(),
			"obcache_misses":   obcacheStats.Misses(),
			"evictions":        obcacheStats.Evictions(),
			"obcache_hit_rate": obcacheStats.HitRate(),
		},
	}
}

// Health check method
func (c *CacheCoordinator) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.obcache != nil
}

// Cleanup and lifecycle management

// startCleanup starts a background goroutine that periodically cleans up expired entries
func (c *CacheCoordinator) startCleanup(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup performs cache maintenance
func (c *CacheCoordinator) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Debug("Performing cache cleanup")
	c.obcache.Cleanup()
}

// Shutdown gracefully shuts down the cache coordinator
func (c *CacheCoordinator) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info("Shutting down cache coordinator")

	if c.obcache != nil {
		c.obcache.Close()
	}

	return nil
}

// CachedRepository wraps a Repository with caching capabilities
type CachedRepository struct {
	repository Repository
	cache      *CacheCoordinator
	logger     *slog.Logger
}

// NewCachedRepository creates a new cached repository
func NewCachedRepository(repository Repository, cache *CacheCoordinator, logger *slog.Logger) *CachedRepository {
	return &CachedRepository{
		repository: repository,
		cache:      cache,
		logger:     logger,
	}
}

// Implement Repository interface with caching

func (cr *CachedRepository) LoadAll(ctx context.Context) ([]*models.Article, error) {
	// Don't cache LoadAll as it's typically called once at startup
	articles, err := cr.repository.LoadAll(ctx)
	if err == nil {
		// Cache individual articles
		for _, article := range articles {
			cr.cache.SetArticle(article.Slug, article)
		}
	}
	return articles, err
}

func (cr *CachedRepository) GetBySlug(slug string) (*models.Article, error) {
	// Check cache first
	if article, found := cr.cache.GetArticle(slug); found {
		return article, nil
	}

	// Not in cache, get from repository
	article, err := cr.repository.GetBySlug(slug)
	if err == nil && article != nil {
		cr.cache.SetArticle(slug, article)
	}

	return article, err
}

func (cr *CachedRepository) GetByTag(tag string) []*models.Article {
	return cr.repository.GetByTag(tag)
}

func (cr *CachedRepository) GetByCategory(category string) []*models.Article {
	return cr.repository.GetByCategory(category)
}

func (cr *CachedRepository) GetPublished() []*models.Article {
	return cr.repository.GetPublished()
}

func (cr *CachedRepository) GetDrafts() []*models.Article {
	return cr.repository.GetDrafts()
}

func (cr *CachedRepository) GetFeatured(limit int) []*models.Article {
	return cr.repository.GetFeatured(limit)
}

func (cr *CachedRepository) GetRecent(limit int) []*models.Article {
	return cr.repository.GetRecent(limit)
}

func (cr *CachedRepository) Reload(ctx context.Context) error {
	err := cr.repository.Reload(ctx)
	if err == nil {
		// Invalidate all cached articles since data has been reloaded
		cr.cache.InvalidateAll()
	}
	return err
}

func (cr *CachedRepository) GetLastModified() time.Time {
	return cr.repository.GetLastModified()
}

func (cr *CachedRepository) GetStats() *models.Stats {
	// Check cache first
	if stats, found := cr.cache.GetStats(); found {
		return stats
	}

	// Not in cache, get from repository
	stats := cr.repository.GetStats()
	if stats != nil {
		cr.cache.SetStats(stats)
	}

	return stats
}

// Ensure CachedRepository implements Repository interface
var _ Repository = (*CachedRepository)(nil)
