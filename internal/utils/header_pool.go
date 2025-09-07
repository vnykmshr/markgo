package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// HeaderValuePool provides optimized header value generation with caching using obcache-go
type HeaderValuePool struct {
	cache *obcache.Cache
}

// NewHeaderValuePool creates a new header value pool using obcache
func NewHeaderValuePool() *HeaderValuePool {
	// Create obcache configuration optimized for header values
	config := obcache.NewDefaultConfig()
	config.MaxEntries = 2000      // Room for many header values
	config.DefaultTTL = time.Hour // Header values can be cached for an hour

	// Create obcache instance
	obcacheInstance, err := obcache.New(config)
	if err != nil {
		// Fallback to basic config if creation fails
		basicConfig := obcache.NewDefaultConfig()
		obcacheInstance, _ = obcache.New(basicConfig)
	}

	pool := &HeaderValuePool{
		cache: obcacheInstance,
	}

	// Pre-populate common values to avoid allocations
	pool.prePopulateCache()

	return pool
}

// prePopulateCache fills the cache with commonly used values
func (p *HeaderValuePool) prePopulateCache() {
	// Pre-populate common integer values
	for i := 0; i <= 1000; i++ {
		key := fmt.Sprintf("int:%d", i)
		_ = p.cache.Set(key, strconv.Itoa(i), 24*time.Hour) // Long TTL for common values
	}

	// Pre-populate common memory deltas (in bytes)
	commonMemoryDeltas := []uint64{0, 1024, 2048, 4096, 8192, 16384, 32768, 65536}
	for _, delta := range commonMemoryDeltas {
		key := fmt.Sprintf("memory:%d", delta)
		_ = p.cache.Set(key, strconv.FormatUint(delta, 10), 24*time.Hour) // Long TTL for common values
	}
}

// GetGoroutineCountString returns a cached string for goroutine count
func (p *HeaderValuePool) GetGoroutineCountString(count int) string {
	// First check if it's a common integer value
	if count <= 1000 && count >= 0 {
		key := fmt.Sprintf("int:%d", count)
		if cached, found := p.cache.Get(key); found {
			return cached.(string)
		}
	}

	// Check goroutine-specific cache
	goroutineKey := fmt.Sprintf("goroutine:%d", count)
	if cached, found := p.cache.Get(goroutineKey); found {
		return cached.(string)
	}

	// Generate and cache
	str := strconv.Itoa(count)
	if count <= 10000 { // Don't cache extremely high values
		_ = p.cache.Set(goroutineKey, str, time.Hour)
	}
	return str
}

// GetMemoryDeltaString returns a cached string for memory delta
func (p *HeaderValuePool) GetMemoryDeltaString(delta uint64) string {
	key := fmt.Sprintf("memory:%d", delta)
	if cached, found := p.cache.Get(key); found {
		return cached.(string)
	}

	// Generate and cache if reasonable size
	str := strconv.FormatUint(delta, 10)
	if delta <= 1048576 { // Don't cache deltas > 1MB to prevent memory leaks
		_ = p.cache.Set(key, str, time.Hour)
	}
	return str
}

// GetIntString returns a cached string for integers
func (p *HeaderValuePool) GetIntString(value int) string {
	if value <= 1000 && value >= 0 {
		key := fmt.Sprintf("int:%d", value)
		if cached, found := p.cache.Get(key); found {
			return cached.(string)
		}
	}
	return strconv.Itoa(value)
}

// ClearCache clears all cached values
func (p *HeaderValuePool) ClearCache() {
	_ = p.cache.Clear()
	// Re-populate common values after clearing
	p.prePopulateCache()
}

// GetCacheStats returns cache statistics
func (p *HeaderValuePool) GetCacheStats() map[string]any {
	stats := p.cache.Stats()
	return map[string]any{
		"hit_count":  int(stats.Hits()),
		"miss_count": int(stats.Misses()),
		"hit_ratio":  stats.HitRate() * 100,
		"total_keys": int(stats.KeyCount()),
		"evictions":  int(stats.Evictions()),
		"cache_type": "obcache-go",
	}
}

// Close shuts down the header value pool
func (p *HeaderValuePool) Close() {
	if p.cache != nil {
		p.cache.Close()
	}
}

// Global header value pool
var globalHeaderValuePool = NewHeaderValuePool()

// GetGlobalHeaderValuePool returns the global header value pool
func GetGlobalHeaderValuePool() *HeaderValuePool {
	return globalHeaderValuePool
}

// Performance Classification Cache for performance headers
type PerformanceClassCache struct {
	cache *obcache.Cache
}

// NewPerformanceClassCache creates a new performance class cache using obcache
func NewPerformanceClassCache() *PerformanceClassCache {
	// Create obcache configuration
	config := obcache.NewDefaultConfig()
	config.MaxEntries = 500       // Performance classes are limited
	config.DefaultTTL = time.Hour // Performance classifications can be cached

	// Create obcache instance
	obcacheInstance, err := obcache.New(config)
	if err != nil {
		// Fallback to basic config if creation fails
		basicConfig := obcache.NewDefaultConfig()
		obcacheInstance, _ = obcache.New(basicConfig)
	}

	cache := &PerformanceClassCache{
		cache: obcacheInstance,
	}

	// Pre-populate common performance classes
	commonClasses := map[string]string{
		"exceptional":        "exceptional",
		"excellent":          "excellent",
		"good":               "good",
		"needs-optimization": "needs-optimization",
		"4x faster":          "4x faster",
		"10x faster":         "10x faster",
		"comparable":         "comparable",
	}

	for key, value := range commonClasses {
		_ = cache.cache.Set(key, value, 24*time.Hour) // Long TTL for common values
	}

	return cache
}

// GetClass returns a cached performance class string
func (p *PerformanceClassCache) GetClass(class string) string {
	if cached, found := p.cache.Get(class); found {
		return cached.(string)
	}

	// Store and return if not found
	_ = p.cache.Set(class, class, time.Hour)
	return class
}

// Close shuts down the performance class cache
func (p *PerformanceClassCache) Close() {
	if p.cache != nil {
		p.cache.Close()
	}
}

// Global performance class cache
var globalPerformanceClassCache = NewPerformanceClassCache()

// GetGlobalPerformanceClassCache returns the global performance class cache
func GetGlobalPerformanceClassCache() *PerformanceClassCache {
	return globalPerformanceClassCache
}
