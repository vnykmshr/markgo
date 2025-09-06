package services

import (
	"bytes"
	"compress/gzip"
	"hash/fnv"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/utils"
)

// ArticleMemoryOptimizer provides memory optimization for article caching
type ArticleMemoryOptimizer struct {
	contentCache   *utils.LRUCache[string, []byte] // Compressed content cache
	excerptCache   *utils.LRUCache[string, string] // Excerpt cache
	processedCache *utils.LRUCache[string, string] // Processed content cache
	stringInterner *utils.StringInterner           // Intern common strings
	bufferPool     *utils.BufferPool               // Buffer pool for processing
	mu             sync.RWMutex
	stats          *MemoryOptimizationStats
}

// MemoryOptimizationStats tracks memory optimization metrics
type MemoryOptimizationStats struct {
	CacheHits          uint64    `json:"cache_hits"`
	CacheMisses        uint64    `json:"cache_misses"`
	CompressionSavings uint64    `json:"compression_savings_bytes"`
	InternerSavings    uint64    `json:"interner_savings_bytes"`
	PoolReuses         uint64    `json:"pool_reuses"`
	LastOptimization   time.Time `json:"last_optimization"`
	MemoryFootprintMB  float64   `json:"memory_footprint_mb"`
}

// NewArticleMemoryOptimizer creates a new memory optimizer for articles
func NewArticleMemoryOptimizer(maxCacheSize int) *ArticleMemoryOptimizer {
	return &ArticleMemoryOptimizer{
		contentCache:   utils.NewLRUCache[string, []byte](maxCacheSize),
		excerptCache:   utils.NewLRUCache[string, string](maxCacheSize * 2), // Excerpts are smaller
		processedCache: utils.NewLRUCache[string, string](maxCacheSize),
		stringInterner: utils.NewStringInterner(),
		bufferPool:     utils.NewBufferPool(),
		stats:          &MemoryOptimizationStats{},
	}
}

// OptimizeArticle applies memory optimizations to an article
func (amo *ArticleMemoryOptimizer) OptimizeArticle(article *models.Article) *models.Article {
	optimized := &models.Article{
		Slug:         amo.stringInterner.Intern(article.Slug),
		Title:        amo.stringInterner.Intern(article.Title),
		Description:  amo.stringInterner.Intern(article.Description),
		Date:         article.Date,
		Tags:         amo.internStringSlice(article.Tags),
		Categories:   amo.internStringSlice(article.Categories),
		Draft:        article.Draft,
		Featured:     article.Featured,
		Author:       amo.stringInterner.Intern(article.Author),
		Content:      article.Content, // Will be compressed separately
		ReadingTime:  article.ReadingTime,
		WordCount:    article.WordCount,
		LastModified: article.LastModified,
	}

	// Compress and cache content if it's large
	if len(article.Content) > 1024 { // Only compress content > 1KB
		compressed := amo.compressContent(article.Content)
		amo.contentCache.Set(article.Slug, compressed)

		// Clear the original content to save memory (lazy loading)
		optimized.Content = ""
	}

	amo.updateStats()
	return optimized
}

// internStringSlice applies string interning to a slice of strings
func (amo *ArticleMemoryOptimizer) internStringSlice(slice []string) []string {
	if len(slice) == 0 {
		return nil // Return nil instead of empty slice to save memory
	}

	interned := make([]string, len(slice))
	for i, s := range slice {
		interned[i] = amo.stringInterner.Intern(s)
	}
	return interned
}

// compressContent compresses article content using gzip
func (amo *ArticleMemoryOptimizer) compressContent(content string) []byte {
	if content == "" {
		return nil
	}

	buf := amo.bufferPool.GetBytesBuffer()
	defer amo.bufferPool.PutBytesBuffer(buf)

	gzipWriter := gzip.NewWriter(buf)
	if _, err := gzipWriter.Write([]byte(content)); err != nil {
		gzipWriter.Close()
		return nil
	}
	gzipWriter.Close()

	compressed := make([]byte, buf.Len())
	copy(compressed, buf.Bytes())

	// Track compression savings
	original := uint64(len(content))
	compressed_size := uint64(len(compressed))
	amo.mu.Lock()
	amo.stats.CompressionSavings += original - compressed_size
	amo.mu.Unlock()

	return compressed
}

// GetContent retrieves content, decompressing if necessary
func (amo *ArticleMemoryOptimizer) GetContent(slug string, originalContent string) string {
	// If we have the content in memory, return it
	if originalContent != "" {
		return originalContent
	}

	// Try to get from compressed cache
	amo.mu.RLock()
	compressed, found := amo.contentCache.Get(slug)
	amo.mu.RUnlock()

	if found {
		amo.mu.Lock()
		amo.stats.CacheHits++
		amo.mu.Unlock()
		return amo.decompressContent(compressed)
	}

	amo.mu.Lock()
	amo.stats.CacheMisses++
	amo.mu.Unlock()

	return "" // Content not available
}

// decompressContent decompresses gzipped content
func (amo *ArticleMemoryOptimizer) decompressContent(compressed []byte) string {
	if len(compressed) == 0 {
		return ""
	}

	buf := amo.bufferPool.GetBytesBuffer()
	defer amo.bufferPool.PutBytesBuffer(buf)

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return ""
	}
	defer reader.Close()

	if _, err := buf.ReadFrom(reader); err != nil {
		return ""
	}
	return buf.String()
}

// CacheProcessedContent caches processed HTML content
func (amo *ArticleMemoryOptimizer) CacheProcessedContent(slug, processed string) {
	if processed == "" {
		return
	}

	// Use hash-based key to save memory on long slugs
	key := amo.hashKey(slug + ":processed")
	amo.processedCache.Set(key, processed)
}

// GetProcessedContent retrieves cached processed content
func (amo *ArticleMemoryOptimizer) GetProcessedContent(slug string) (string, bool) {
	key := amo.hashKey(slug + ":processed")
	content, found := amo.processedCache.Get(key)

	if found {
		amo.mu.Lock()
		amo.stats.CacheHits++
		amo.mu.Unlock()
	} else {
		amo.mu.Lock()
		amo.stats.CacheMisses++
		amo.mu.Unlock()
	}

	return content, found
}

// CacheExcerpt caches article excerpt
func (amo *ArticleMemoryOptimizer) CacheExcerpt(slug, excerpt string) {
	if excerpt == "" {
		return
	}

	key := amo.hashKey(slug + ":excerpt")
	amo.excerptCache.Set(key, excerpt)
}

// GetExcerpt retrieves cached excerpt
func (amo *ArticleMemoryOptimizer) GetExcerpt(slug string) (string, bool) {
	key := amo.hashKey(slug + ":excerpt")
	excerpt, found := amo.excerptCache.Get(key)

	if found {
		amo.mu.Lock()
		amo.stats.CacheHits++
		amo.mu.Unlock()
	} else {
		amo.mu.Lock()
		amo.stats.CacheMisses++
		amo.mu.Unlock()
	}

	return excerpt, found
}

// hashKey creates a hash-based key to reduce memory usage
func (amo *ArticleMemoryOptimizer) hashKey(key string) string {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return amo.stringInterner.Intern(strconv.FormatUint(hasher.Sum64(), 16))
}

// OptimizeArticleSlice optimizes a slice of articles for memory usage
func (amo *ArticleMemoryOptimizer) OptimizeArticleSlice(articles []*models.Article) []*models.Article {
	if len(articles) == 0 {
		return nil
	}

	// Pre-allocate with exact capacity to avoid slice growth
	optimized := make([]*models.Article, 0, len(articles))

	for _, article := range articles {
		optimized = append(optimized, amo.OptimizeArticle(article))
	}

	return optimized
}

// CompactMemory performs memory compaction by clearing unused caches
func (amo *ArticleMemoryOptimizer) CompactMemory() {
	amo.mu.Lock()
	defer amo.mu.Unlock()

	// Clear least recently used items
	amo.contentCache.RemoveOldest(amo.contentCache.Len() / 4) // Remove 25% of oldest
	amo.excerptCache.RemoveOldest(amo.excerptCache.Len() / 4)
	amo.processedCache.RemoveOldest(amo.processedCache.Len() / 4)

	// Compact string interner
	amo.stringInterner.Compact()

	amo.stats.LastOptimization = time.Now()
}

// GetStats returns current memory optimization statistics
func (amo *ArticleMemoryOptimizer) GetStats() *MemoryOptimizationStats {
	amo.mu.RLock()
	defer amo.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := *amo.stats
	stats.MemoryFootprintMB = amo.calculateMemoryFootprint()

	return &stats
}

// calculateMemoryFootprint estimates current memory usage
func (amo *ArticleMemoryOptimizer) calculateMemoryFootprint() float64 {
	var total uintptr

	// Estimate cache sizes
	total += uintptr(amo.contentCache.Len() * 100) // Rough estimate per entry
	total += uintptr(amo.excerptCache.Len() * 50)
	total += uintptr(amo.processedCache.Len() * 200)
	total += uintptr(amo.stringInterner.Size() * int(unsafe.Sizeof("")))

	return float64(total) / (1024 * 1024) // Convert to MB
}

// updateStats updates optimization statistics
func (amo *ArticleMemoryOptimizer) updateStats() {
	amo.mu.Lock()
	defer amo.mu.Unlock()

	amo.stats.InternerSavings += uint64(amo.stringInterner.SavedMemory())
}

// Cleanup performs cleanup of resources
func (amo *ArticleMemoryOptimizer) Cleanup() {
	amo.mu.Lock()
	defer amo.mu.Unlock()

	amo.contentCache.Clear()
	amo.excerptCache.Clear()
	amo.processedCache.Clear()
	amo.stringInterner.Clear()
}

// PrewarmCache pre-loads frequently accessed content into cache
func (amo *ArticleMemoryOptimizer) PrewarmCache(articles []*models.Article) {
	// Pre-warm with most recent articles (typically accessed first)
	for i, article := range articles {
		if i >= 10 { // Limit prewarming to first 10 articles
			break
		}

		if len(article.Content) > 1024 {
			compressed := amo.compressContent(article.Content)
			amo.contentCache.Set(article.Slug, compressed)
		}
	}
}

// GetCacheEfficiency returns cache hit/miss ratio
func (amo *ArticleMemoryOptimizer) GetCacheEfficiency() float64 {
	amo.mu.RLock()
	defer amo.mu.RUnlock()

	total := amo.stats.CacheHits + amo.stats.CacheMisses
	if total == 0 {
		return 0.0
	}

	return float64(amo.stats.CacheHits) / float64(total) * 100.0
}
