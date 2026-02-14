package article

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/1mb-dev/markgo/internal/models"
)

func newTestCache(t *testing.T) *CacheCoordinator {
	t.Helper()
	cfg := &CacheConfig{
		MaxEntries:    100,
		ArticleTTL:    5 * time.Minute,
		SearchTTL:     5 * time.Minute,
		ContentTTL:    5 * time.Minute,
		CleanupPeriod: 0, // No background cleanup in tests
	}
	cache, err := NewCacheCoordinator(cfg, slog.Default())
	require.NoError(t, err)
	t.Cleanup(func() { _ = cache.Shutdown(context.Background()) })
	return cache
}

func TestArticleCache(t *testing.T) {
	cache := newTestCache(t)

	article := &models.Article{
		Slug:  "test-article",
		Title: "Test",
	}

	// Miss
	_, found := cache.GetArticle("test-article")
	assert.False(t, found)

	// Set
	cache.SetArticle("test-article", article)

	// Hit
	got, found := cache.GetArticle("test-article")
	assert.True(t, found)
	assert.Equal(t, "Test", got.Title)

	// Invalidate
	cache.InvalidateArticle("test-article")
	_, found = cache.GetArticle("test-article")
	assert.False(t, found)

	// SetArticle with nil is a no-op
	cache.SetArticle("nil-test", nil)
	_, found = cache.GetArticle("nil-test")
	assert.False(t, found)
}

func TestSearchCache(t *testing.T) {
	cache := newTestCache(t)

	results := []*models.SearchResult{
		{Article: &models.Article{Slug: "a"}, Score: 10},
	}

	// Miss
	_, found := cache.GetSearchResults("query", 10)
	assert.False(t, found)

	// Set
	cache.SetSearchResults("query", 10, results)

	// Hit
	got, found := cache.GetSearchResults("query", 10)
	assert.True(t, found)
	assert.Len(t, got, 1)

	// Different limit is a different key
	_, found = cache.GetSearchResults("query", 5)
	assert.False(t, found)

	// SetSearchResults with nil is a no-op
	cache.SetSearchResults("nil-test", 10, nil)
	_, found = cache.GetSearchResults("nil-test", 10)
	assert.False(t, found)
}

func TestContentCache(t *testing.T) {
	cache := newTestCache(t)

	// Miss
	_, found := cache.GetProcessedContent("hash1")
	assert.False(t, found)

	// Set
	cache.SetProcessedContent("hash1", "<p>Hello</p>")

	// Hit
	got, found := cache.GetProcessedContent("hash1")
	assert.True(t, found)
	assert.Equal(t, "<p>Hello</p>", got)

	// SetProcessedContent with empty string is a no-op
	cache.SetProcessedContent("empty", "")
	_, found = cache.GetProcessedContent("empty")
	assert.False(t, found)
}

func TestStatsCache(t *testing.T) {
	cache := newTestCache(t)

	// Miss
	_, found := cache.GetStats()
	assert.False(t, found)

	// Set
	stats := &models.Stats{TotalArticles: 42}
	cache.SetStats(stats)

	// Hit
	got, found := cache.GetStats()
	assert.True(t, found)
	assert.Equal(t, 42, got.TotalArticles)

	// SetStats with nil is a no-op
	cache.SetStats(nil)
}

func TestInvalidateSearchCache(t *testing.T) {
	cache := newTestCache(t)

	// InvalidateSearchCache clears ALL cache — this is a known limitation.
	// The comment in the source acknowledges this: "obcache doesn't have
	// prefix-based deletion, we need to clear the entire cache."
	cache.SetArticle("test", &models.Article{Slug: "test"})
	cache.SetSearchResults("query", 10, []*models.SearchResult{})

	cache.InvalidateSearchCache()

	// Both article and search caches are cleared
	_, found := cache.GetArticle("test")
	assert.False(t, found)
	_, found = cache.GetSearchResults("query", 10)
	assert.False(t, found)
}

func TestInvalidateAll(t *testing.T) {
	cache := newTestCache(t)

	cache.SetArticle("a", &models.Article{Slug: "a"})
	cache.SetProcessedContent("h", "content")

	cache.InvalidateAll()

	_, found := cache.GetArticle("a")
	assert.False(t, found)
	_, found = cache.GetProcessedContent("h")
	assert.False(t, found)
}

func TestGetCacheStats(t *testing.T) {
	cache := newTestCache(t)

	cache.SetArticle("a", &models.Article{Slug: "a"})
	cache.GetArticle("a")           // hit
	cache.GetArticle("nonexistent") // miss

	stats := cache.GetCacheStats()
	assert.Equal(t, int64(1), stats["hits"])
	assert.Equal(t, int64(1), stats["misses"])

	hitRate, ok := stats["hit_rate"].(float64)
	require.True(t, ok)
	assert.InDelta(t, 0.5, hitRate, 0.01)
}

func TestGetCacheStats_ConcurrentAccess(t *testing.T) {
	cache := newTestCache(t)

	article := &models.Article{Slug: "test", Title: "Test"}
	cache.SetArticle("test", article)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.GetArticle("test")
			cache.GetArticle("miss")
			cache.GetCacheStats()
		}()
	}

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.SetArticle("test", article)
			cache.SetProcessedContent("h", "c")
		}()
	}

	wg.Wait()

	stats := cache.GetCacheStats()
	assert.Greater(t, stats["hits"].(int64)+stats["misses"].(int64), int64(0))
}

func TestIsHealthy(t *testing.T) {
	cache := newTestCache(t)
	assert.True(t, cache.IsHealthy())
}

func TestShutdown(t *testing.T) {
	// Test with cleanup goroutine running
	cfg := &CacheConfig{
		MaxEntries:    100,
		ArticleTTL:    5 * time.Minute,
		SearchTTL:     5 * time.Minute,
		ContentTTL:    5 * time.Minute,
		CleanupPeriod: 100 * time.Millisecond, // Fast cleanup for test
	}
	cache, err := NewCacheCoordinator(cfg, slog.Default())
	require.NoError(t, err)

	// Let the cleanup goroutine run at least once
	time.Sleep(150 * time.Millisecond)

	err = cache.Shutdown(context.Background())
	assert.NoError(t, err)

	// Double shutdown should not panic
	err = cache.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestDefaultCacheConfig(t *testing.T) {
	cfg := DefaultCacheConfig()
	assert.Equal(t, 5000, cfg.MaxEntries)
	assert.Equal(t, 30*time.Minute, cfg.ArticleTTL)
	assert.Equal(t, 10*time.Minute, cfg.SearchTTL)
	assert.Equal(t, 60*time.Minute, cfg.ContentTTL)
	assert.Equal(t, 15*time.Minute, cfg.CleanupPeriod)
}

func TestNewCacheCoordinator_NilConfig(t *testing.T) {
	cache, err := NewCacheCoordinator(nil, slog.Default())
	require.NoError(t, err)
	assert.NotNil(t, cache)
	_ = cache.Shutdown(context.Background())
}

func TestInvalidateByTag(t *testing.T) {
	cache := newTestCache(t)

	cache.SetArticle("a", &models.Article{Slug: "a"})
	cache.InvalidateByTag("go")

	// InvalidateByTag clears all cache (aggressive but safe — same as InvalidateSearchCache)
	_, found := cache.GetArticle("a")
	assert.False(t, found)
}
