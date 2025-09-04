package services

import (
	"sync"
	"time"
)

// Ensure CacheService implements CacheServiceInterface
var _ CacheServiceInterface = (*CacheService)(nil)

type CacheItem struct {
	Value      any
	Expiration time.Time
}

type CacheService struct {
	items       map[string]*CacheItem
	mutex       sync.RWMutex
	defaultTTL  time.Duration
	maxSize     int
	cleanupTick *time.Ticker
	stopCleanup chan bool
}

func NewCacheService(defaultTTL time.Duration, maxSize int) *CacheService {
	cache := &CacheService{
		items:       make(map[string]*CacheItem),
		defaultTTL:  defaultTTL,
		maxSize:     maxSize,
		cleanupTick: time.NewTicker(10 * time.Minute),
		stopCleanup: make(chan bool),
	}

	// Start cleanup goroutine
	go cache.startCleanup()

	return cache
}

func (c *CacheService) startCleanup() {
	for {
		select {
		case <-c.cleanupTick.C:
			c.cleanup()
		case <-c.stopCleanup:
			c.cleanupTick.Stop()
			return
		}
	}
}

func (c *CacheService) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.Expiration) {
			delete(c.items, key)
		}
	}
}

func (c *CacheService) Set(key string, value any, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If cache is full, remove oldest item
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	if ttl == 0 {
		ttl = c.defaultTTL
	}

	c.items[key] = &CacheItem{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

func (c *CacheService) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.Expiration) {
		// Remove expired item
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.items, key)
		c.mutex.Unlock()
		c.mutex.RLock()
		return nil, false
	}

	return item.Value, true
}

func (c *CacheService) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

func (c *CacheService) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
}

func (c *CacheService) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

func (c *CacheService) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}

	return keys
}

func (c *CacheService) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.Expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.Expiration
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

func (c *CacheService) Stop() {
	c.stopCleanup <- true
}

// GetOrSet gets a value from cache, or sets it if it doesn't exist
func (c *CacheService) GetOrSet(key string, generator func() any, ttl time.Duration) any {
	// Try to get from cache first
	if value, found := c.Get(key); found {
		return value
	}

	// Generate value and set in cache
	value := generator()
	c.Set(key, value, ttl)
	return value
}

// Exists checks if a key exists in cache without retrieving the value
func (c *CacheService) Exists(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	// Check if item has expired
	return !time.Now().After(item.Expiration)
}

// GetTTL returns the remaining TTL for a key
func (c *CacheService) GetTTL(key string) time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return 0
	}

	remaining := time.Until(item.Expiration)
	if remaining < 0 {
		return 0
	}

	return remaining
}

// Stats returns cache statistics
func (c *CacheService) Stats() map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	expired := 0
	now := time.Now()
	for _, item := range c.items {
		if now.After(item.Expiration) {
			expired++
		}
	}

	return map[string]any{
		"total_items":   len(c.items),
		"expired_items": expired,
		"max_size":      c.maxSize,
		"default_ttl":   c.defaultTTL.String(),
	}
}
