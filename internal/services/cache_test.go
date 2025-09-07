package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCacheService(t *testing.T) {
	ttl := time.Hour
	maxSize := 100

	cache := NewCacheService(ttl, maxSize)
	defer cache.Stop()

	assert.NotNil(t, cache)
	assert.Equal(t, ttl, cache.defaultTTL)
	assert.NotNil(t, cache.cache)
	assert.Equal(t, 0, cache.Size())
}

func TestCacheService_SetAndGet(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// Test set and get
	cache.Set("key1", "value1", 0) // Use default TTL
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Test non-existent key
	_, found = cache.Get("non-existent")
	assert.False(t, found)
}

func TestCacheService_SetWithCustomTTL(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// Set with custom TTL
	cache.Set("key1", "value1", 100*time.Millisecond)

	// Should exist initially
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	_, found = cache.Get("key1")
	assert.False(t, found)
}

func TestCacheService_Delete(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// Set and verify
	cache.Set("key1", "value1", 0)
	_, found := cache.Get("key1")
	assert.True(t, found)

	// Delete and verify
	cache.Delete("key1")
	_, found = cache.Get("key1")
	assert.False(t, found)
}

func TestCacheService_Clear(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// Add multiple items
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0)

	// Verify items exist
	assert.True(t, cache.Exists("key1"))
	assert.True(t, cache.Exists("key2"))
	assert.True(t, cache.Exists("key3"))

	// Clear cache
	cache.Clear()

	// Verify all items are gone
	assert.False(t, cache.Exists("key1"))
	assert.False(t, cache.Exists("key2"))
	assert.False(t, cache.Exists("key3"))
	assert.Equal(t, 0, cache.Size())
}

func TestCacheService_Size(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// Initial size should be 0
	assert.Equal(t, 0, cache.Size())

	// Add items and check size
	cache.Set("key1", "value1", 0)
	assert.Equal(t, 1, cache.Size())

	cache.Set("key2", "value2", 0)
	assert.Equal(t, 2, cache.Size())

	// Delete and check size
	cache.Delete("key1")
	assert.Equal(t, 1, cache.Size())
}

func TestCacheService_Exists(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// Should not exist initially
	assert.False(t, cache.Exists("key1"))

	// Set and check existence
	cache.Set("key1", "value1", 0)
	assert.True(t, cache.Exists("key1"))

	// Delete and check existence
	cache.Delete("key1")
	assert.False(t, cache.Exists("key1"))
}

func TestCacheService_GetTTL(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// TTL for non-existent key should be 0
	assert.Equal(t, time.Duration(0), cache.GetTTL("non-existent"))

	// Set a key and check TTL (should return default TTL approximation)
	cache.Set("key1", "value1", 0)
	ttl := cache.GetTTL("key1")
	assert.Equal(t, time.Hour, ttl)
}

func TestCacheService_Stats(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// Get initial stats
	stats := cache.Stats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "hit_count")
	assert.Contains(t, stats, "miss_count")
	assert.Contains(t, stats, "hit_ratio")
	assert.Contains(t, stats, "total_keys")
	assert.Contains(t, stats, "cache_type")
	assert.Equal(t, "obcache-go", stats["cache_type"])

	// Add some data and check stats again
	cache.Set("key1", "value1", 0)
	cache.Get("key1")    // Hit
	cache.Get("missing") // Miss

	stats = cache.Stats()
	assert.Equal(t, 1, stats["total_keys"])
}

func TestCacheService_GetOrSet(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	generatorCalled := false
	generator := func() any {
		generatorCalled = true
		return "generated_value"
	}

	// First call should generate value
	value := cache.GetOrSet("key1", generator, 0)
	assert.Equal(t, "generated_value", value)
	assert.True(t, generatorCalled)

	// Reset flag
	generatorCalled = false

	// Second call should get cached value
	value = cache.GetOrSet("key1", generator, 0)
	assert.Equal(t, "generated_value", value)
	assert.False(t, generatorCalled, "Generator should not be called for cached value")
}

func TestCacheService_DifferentTypes(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// Test different value types
	testCases := []struct {
		key   string
		value any
	}{
		{"string", "test string"},
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"slice", []string{"a", "b", "c"}},
		{"map", map[string]int{"a": 1, "b": 2}},
	}

	// Set all values
	for _, tc := range testCases {
		cache.Set(tc.key, tc.value, 0)
	}

	// Verify all values
	for _, tc := range testCases {
		value, found := cache.Get(tc.key)
		assert.True(t, found, "Key %s should exist", tc.key)
		assert.Equal(t, tc.value, value, "Value for key %s should match", tc.key)
	}
}

// Interface compliance test
func TestCacheService_ImplementsCacheServiceInterface(t *testing.T) {
	cache := NewCacheService(time.Hour, 10)
	defer cache.Stop()

	// This will fail to compile if CacheService doesn't implement the interface
	var _ CacheServiceInterface = cache
}
