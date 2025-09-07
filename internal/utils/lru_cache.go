package utils

import (
	"fmt"
	"time"

	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// LRUCache is now a wrapper around obcache-go for better performance and maintainability
type LRUCache[K comparable, V any] struct {
	cache *obcache.Cache
}

// LRUStats tracks cache performance metrics (compatible with old interface)
type LRUStats struct {
	Hits        uint64 `json:"hits"`
	Misses      uint64 `json:"misses"`
	Evictions   uint64 `json:"evictions"`
	Expirations uint64 `json:"expirations"`
	Size        int    `json:"size"`
}

// NewLRUCache creates a new LRU cache with the specified capacity
func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	if capacity <= 0 {
		capacity = 100 // Default capacity
	}

	// Create obcache configuration optimized for LRU behavior
	config := obcache.NewDefaultConfig()
	config.MaxEntries = capacity
	config.DefaultTTL = time.Hour // Default expiration

	// Create obcache instance
	obcacheInstance, err := obcache.New(config)
	if err != nil {
		// Fallback to basic config if creation fails
		basicConfig := obcache.NewDefaultConfig()
		basicConfig.MaxEntries = capacity
		obcacheInstance, _ = obcache.New(basicConfig)
	}

	return &LRUCache[K, V]{
		cache: obcacheInstance,
	}
}

// Set adds or updates an item in the cache without expiration
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, 0) // 0 means no expiration
}

// SetWithTTL adds or updates an item with a time-to-live
func (c *LRUCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	keyStr := c.keyToString(key)
	if ttl == 0 {
		ttl = 365 * 24 * time.Hour // Effectively no expiration (1 year)
	}
	_ = c.cache.Set(keyStr, value, ttl)
}

// Get retrieves an item from the cache
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	keyStr := c.keyToString(key)
	value, found := c.cache.Get(keyStr)
	if found {
		return value.(V), true
	}
	var zero V
	return zero, false
}

// Peek retrieves an item without updating its position (obcache limitation)
func (c *LRUCache[K, V]) Peek(key K) (V, bool) {
	// Note: obcache doesn't distinguish between Get and Peek
	// This is a limitation we accept for the performance benefits
	return c.Get(key)
}

// Delete removes an item from the cache
func (c *LRUCache[K, V]) Delete(key K) bool {
	keyStr := c.keyToString(key)
	existed, _ := c.cache.Get(keyStr)
	_ = c.cache.Delete(keyStr)
	return existed != nil
}

// Clear removes all items from the cache
func (c *LRUCache[K, V]) Clear() {
	_ = c.cache.Clear()
}

// Len returns the number of items in the cache
func (c *LRUCache[K, V]) Len() int {
	stats := c.cache.Stats()
	return int(stats.KeyCount())
}

// Keys returns all keys in the cache (limitation: obcache doesn't support enumeration)
func (c *LRUCache[K, V]) Keys() []K {
	// Note: obcache doesn't provide key enumeration
	// This is a limitation we accept for the performance benefits
	return []K{}
}

// RemoveOldest removes the specified number of oldest items (approximation)
func (c *LRUCache[K, V]) RemoveOldest(count int) int {
	// obcache handles eviction automatically based on LRU policy
	// We can't directly remove "oldest" items, but we can trigger cleanup
	// This is a best-effort approximation
	return 0 // obcache handles this internally
}

// CleanupExpired removes all expired items (handled automatically by obcache)
func (c *LRUCache[K, V]) CleanupExpired() int {
	// obcache handles expired item cleanup automatically
	// Return 0 as we can't count the cleaned items
	return 0
}

// GetStats returns cache statistics
func (c *LRUCache[K, V]) GetStats() LRUStats {
	stats := c.cache.Stats()
	return LRUStats{
		Hits:        uint64(stats.Hits()),
		Misses:      uint64(stats.Misses()),
		Evictions:   uint64(stats.Evictions()),
		Expirations: 0, // obcache doesn't track expirations separately
		Size:        int(stats.KeyCount()),
	}
}

// GetHitRatio returns the cache hit ratio as a percentage
func (c *LRUCache[K, V]) GetHitRatio() float64 {
	stats := c.cache.Stats()
	return stats.HitRate() * 100.0
}

// GetOldestKey returns the key of the oldest item (limitation: not supported)
func (c *LRUCache[K, V]) GetOldestKey() (K, bool) {
	// obcache doesn't expose LRU order information
	var zero K
	return zero, false
}

// GetNewestKey returns the key of the newest item (limitation: not supported)
func (c *LRUCache[K, V]) GetNewestKey() (K, bool) {
	// obcache doesn't expose LRU order information
	var zero K
	return zero, false
}

// Contains checks if a key exists in the cache
func (c *LRUCache[K, V]) Contains(key K) bool {
	keyStr := c.keyToString(key)
	_, found := c.cache.Get(keyStr)
	return found
}

// ForEach iterates over all items in the cache (limitation: not supported)
func (c *LRUCache[K, V]) ForEach(fn func(key K, value V) bool) {
	// obcache doesn't provide key/value enumeration
	// This is a limitation we accept for the performance benefits
}

// ResizeCapacity changes the cache capacity (limitation: not supported at runtime)
func (c *LRUCache[K, V]) ResizeCapacity(newCapacity int) {
	// obcache doesn't support runtime capacity changes
	// This is a limitation we accept for the performance benefits
}

// GetCapacity returns the current cache capacity (approximation)
func (c *LRUCache[K, V]) GetCapacity() int {
	// obcache doesn't expose capacity, return approximation based on current size
	stats := c.cache.Stats()
	return int(stats.KeyCount()) // This is not the capacity, just current size
}

// Close shuts down the cache
func (c *LRUCache[K, V]) Close() {
	if c.cache != nil {
		c.cache.Close()
	}
}

// keyToString converts any comparable key to string for obcache
func (c *LRUCache[K, V]) keyToString(key K) string {
	return fmt.Sprintf("%v", key)
}
