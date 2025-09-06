package utils

import (
	"sync"
	"time"
)

// LRUCache implements a generic Least Recently Used cache with expiration
type LRUCache[K comparable, V any] struct {
	capacity int
	items    map[K]*lruItem[K, V]
	head     *lruItem[K, V]
	tail     *lruItem[K, V]
	mu       sync.RWMutex
	stats    LRUStats
}

// lruItem represents an item in the LRU cache
type lruItem[K comparable, V any] struct {
	key        K
	value      V
	prev       *lruItem[K, V]
	next       *lruItem[K, V]
	accessTime time.Time
	expiry     time.Time
}

// LRUStats tracks cache performance metrics
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

	cache := &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*lruItem[K, V]),
	}

	// Initialize sentinel nodes
	cache.head = &lruItem[K, V]{}
	cache.tail = &lruItem[K, V]{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// Set adds or updates an item in the cache
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, 0) // 0 means no expiration
}

// SetWithTTL adds or updates an item with a time-to-live
func (c *LRUCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiry := time.Time{} // Zero time means no expiration
	if ttl > 0 {
		expiry = now.Add(ttl)
	}

	if item, exists := c.items[key]; exists {
		// Update existing item
		item.value = value
		item.accessTime = now
		item.expiry = expiry
		c.moveToFront(item)
		return
	}

	// Create new item
	item := &lruItem[K, V]{
		key:        key,
		value:      value,
		accessTime: now,
		expiry:     expiry,
	}

	c.items[key] = item
	c.addToFront(item)
	c.stats.Size++

	// Check capacity and evict if necessary
	if len(c.items) > c.capacity {
		c.evictLRU()
	}
}

// Get retrieves an item from the cache
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		var zero V
		return zero, false
	}

	// Check expiration
	if !item.expiry.IsZero() && time.Now().After(item.expiry) {
		c.removeItem(item)
		c.stats.Expirations++
		c.stats.Misses++
		var zero V
		return zero, false
	}

	// Update access time and move to front
	item.accessTime = time.Now()
	c.moveToFront(item)
	c.stats.Hits++

	return item.value, true
}

// Peek retrieves an item without updating its position
func (c *LRUCache[K, V]) Peek(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		var zero V
		return zero, false
	}

	// Check expiration
	if !item.expiry.IsZero() && time.Now().After(item.expiry) {
		var zero V
		return zero, false
	}

	return item.value, true
}

// Delete removes an item from the cache
func (c *LRUCache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	c.removeItem(item)
	return true
}

// Clear removes all items from the cache
func (c *LRUCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*lruItem[K, V])
	c.head.next = c.tail
	c.tail.prev = c.head
	c.stats.Size = 0
}

// Len returns the number of items in the cache
func (c *LRUCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Keys returns all keys in the cache (from most to least recently used)
func (c *LRUCache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.items))
	for current := c.head.next; current != c.tail; current = current.next {
		keys = append(keys, current.key)
	}
	return keys
}

// RemoveOldest removes the specified number of oldest items
func (c *LRUCache[K, V]) RemoveOldest(count int) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	for i := 0; i < count && len(c.items) > 0; i++ {
		if c.tail.prev != c.head {
			c.removeItem(c.tail.prev)
			c.stats.Evictions++
			removed++
		}
	}
	return removed
}

// CleanupExpired removes all expired items
func (c *LRUCache[K, V]) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expired := make([]*lruItem[K, V], 0)

	// Find expired items
	for current := c.head.next; current != c.tail; current = current.next {
		if !current.expiry.IsZero() && now.After(current.expiry) {
			expired = append(expired, current)
		}
	}

	// Remove expired items
	for _, item := range expired {
		c.removeItem(item)
		c.stats.Expirations++
	}

	return len(expired)
}

// GetStats returns cache statistics
func (c *LRUCache[K, V]) GetStats() LRUStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.items)
	return stats
}

// GetHitRatio returns the cache hit ratio as a percentage
func (c *LRUCache[K, V]) GetHitRatio() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.stats.Hits + c.stats.Misses
	if total == 0 {
		return 0.0
	}
	return float64(c.stats.Hits) / float64(total) * 100.0
}

// addToFront adds an item to the front of the list (most recently used)
func (c *LRUCache[K, V]) addToFront(item *lruItem[K, V]) {
	item.prev = c.head
	item.next = c.head.next
	c.head.next.prev = item
	c.head.next = item
}

// moveToFront moves an item to the front of the list
func (c *LRUCache[K, V]) moveToFront(item *lruItem[K, V]) {
	// Remove from current position
	item.prev.next = item.next
	item.next.prev = item.prev

	// Add to front
	c.addToFront(item)
}

// removeItem removes an item from the cache and linked list
func (c *LRUCache[K, V]) removeItem(item *lruItem[K, V]) {
	// Remove from linked list
	item.prev.next = item.next
	item.next.prev = item.prev

	// Remove from map
	delete(c.items, item.key)
	c.stats.Size--
}

// evictLRU removes the least recently used item
func (c *LRUCache[K, V]) evictLRU() {
	if c.tail.prev != c.head {
		c.removeItem(c.tail.prev)
		c.stats.Evictions++
	}
}

// GetOldestKey returns the key of the oldest (least recently used) item
func (c *LRUCache[K, V]) GetOldestKey() (K, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.tail.prev == c.head {
		var zero K
		return zero, false
	}
	return c.tail.prev.key, true
}

// GetNewestKey returns the key of the newest (most recently used) item
func (c *LRUCache[K, V]) GetNewestKey() (K, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.head.next == c.tail {
		var zero K
		return zero, false
	}
	return c.head.next.key, true
}

// Contains checks if a key exists in the cache without affecting its position
func (c *LRUCache[K, V]) Contains(key K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	// Check expiration
	if !item.expiry.IsZero() && time.Now().After(item.expiry) {
		return false
	}

	return true
}

// ForEach iterates over all items in the cache (from most to least recently used)
func (c *LRUCache[K, V]) ForEach(fn func(key K, value V) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for current := c.head.next; current != c.tail; current = current.next {
		// Check expiration
		if !current.expiry.IsZero() && time.Now().After(current.expiry) {
			continue
		}

		if !fn(current.key, current.value) {
			break
		}
	}
}

// ResizeCapacity changes the cache capacity
func (c *LRUCache[K, V]) ResizeCapacity(newCapacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if newCapacity <= 0 {
		newCapacity = 1
	}

	c.capacity = newCapacity

	// Evict items if necessary
	for len(c.items) > c.capacity {
		c.evictLRU()
	}
}

// GetCapacity returns the current cache capacity
func (c *LRUCache[K, V]) GetCapacity() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.capacity
}
