package services

import (
	"time"

	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// Ensure CacheService implements CacheServiceInterface
var _ CacheServiceInterface = (*CacheService)(nil)

// CacheService is now a wrapper around obcache-go for backward compatibility
type CacheService struct {
	cache      *obcache.Cache
	defaultTTL time.Duration
}

func NewCacheService(defaultTTL time.Duration, maxSize int) *CacheService {
	// Create obcache configuration
	config := obcache.NewDefaultConfig()
	config.DefaultTTL = defaultTTL
	config.MaxEntries = maxSize

	// Create obcache instance
	obcacheInstance, err := obcache.New(config)
	if err != nil {
		// Fallback to a basic configuration if creation fails
		basicConfig := obcache.NewDefaultConfig()
		obcacheInstance, _ = obcache.New(basicConfig)
	}

	return &CacheService{
		cache:      obcacheInstance,
		defaultTTL: defaultTTL,
	}
}

// Set adds or updates an item in the cache
func (c *CacheService) Set(key string, value any, ttl time.Duration) {
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	_ = c.cache.Set(key, value, ttl)
}

// Get retrieves an item from the cache
func (c *CacheService) Get(key string) (any, bool) {
	value, found := c.cache.Get(key)
	return value, found
}

// Delete removes an item from the cache
func (c *CacheService) Delete(key string) {
	_ = c.cache.Delete(key)
}

// Clear removes all items from the cache
func (c *CacheService) Clear() {
	_ = c.cache.Clear()
}

// Size returns the number of items in the cache
func (c *CacheService) Size() int {
	stats := c.cache.Stats()
	return int(stats.KeyCount())
}

// Keys returns all keys in the cache
func (c *CacheService) Keys() []string {
	// Note: obcache doesn't provide direct key enumeration
	// This is a limitation we'll have to accept
	return []string{}
}

// Exists checks if a key exists in the cache
func (c *CacheService) Exists(key string) bool {
	_, found := c.cache.Get(key)
	return found
}

// GetTTL returns the time-to-live for a key (obcache limitation)
func (c *CacheService) GetTTL(key string) time.Duration {
	// obcache doesn't provide TTL introspection
	// Return default TTL as approximation
	if c.Exists(key) {
		return c.defaultTTL
	}
	return 0
}

// Stats returns cache statistics
func (c *CacheService) Stats() map[string]any {
	stats := c.cache.Stats()
	return map[string]any{
		"hit_count":  int(stats.Hits()),
		"miss_count": int(stats.Misses()),
		"hit_ratio":  stats.HitRate() * 100,
		"total_keys": int(stats.KeyCount()),
		"evictions":  int(stats.Evictions()),
		"size":       int(stats.KeyCount()),
		"cache_type": "obcache-go",
	}
}

// GetOrSet gets a value or sets it using the generator function
func (c *CacheService) GetOrSet(key string, generator func() any, ttl time.Duration) any {
	// Check if key exists first
	if value, found := c.cache.Get(key); found {
		return value
	}

	// Generate new value
	value := generator()

	// Set with TTL
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	_ = c.cache.Set(key, value, ttl)

	return value
}

// Stop shuts down the cache service
func (c *CacheService) Stop() {
	if c.cache != nil {
		c.cache.Close()
	}
}
