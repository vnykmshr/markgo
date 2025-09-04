package utils

import (
	"strconv"
	"sync"
)

// HeaderValuePool provides optimized header value generation with caching
type HeaderValuePool struct {
	goroutineCache sync.Map // Cache for goroutine count strings
	memoryCache    sync.Map // Cache for common memory delta strings
	intCache       sync.Map // Cache for common integer strings
}

// NewHeaderValuePool creates a new header value pool
func NewHeaderValuePool() *HeaderValuePool {
	pool := &HeaderValuePool{}
	
	// Pre-populate common values to avoid allocations
	for i := 0; i <= 1000; i++ {
		pool.intCache.Store(i, strconv.Itoa(i))
	}
	
	// Pre-populate common memory deltas (in bytes)
	commonMemoryDeltas := []uint64{0, 1024, 2048, 4096, 8192, 16384, 32768, 65536}
	for _, delta := range commonMemoryDeltas {
		pool.memoryCache.Store(delta, strconv.FormatUint(delta, 10))
	}
	
	return pool
}

// GetGoroutineCountString returns a cached string for goroutine count
func (p *HeaderValuePool) GetGoroutineCountString(count int) string {
	if count <= 1000 {
		if cached, found := p.intCache.Load(count); found {
			return cached.(string)
		}
	}
	
	// Not in cache or too large, check goroutine-specific cache
	if cached, found := p.goroutineCache.Load(count); found {
		return cached.(string)
	}
	
	// Generate and cache
	str := strconv.Itoa(count)
	if count <= 10000 { // Don't cache extremely high values
		p.goroutineCache.Store(count, str)
	}
	return str
}

// GetMemoryDeltaString returns a cached string for memory delta
func (p *HeaderValuePool) GetMemoryDeltaString(delta uint64) string {
	if cached, found := p.memoryCache.Load(delta); found {
		return cached.(string)
	}
	
	// Generate and cache if reasonable size
	str := strconv.FormatUint(delta, 10)
	if delta <= 1048576 { // Don't cache deltas > 1MB to prevent memory leaks
		p.memoryCache.Store(delta, str)
	}
	return str
}

// GetIntString returns a cached string for integers
func (p *HeaderValuePool) GetIntString(value int) string {
	if value <= 1000 && value >= 0 {
		if cached, found := p.intCache.Load(value); found {
			return cached.(string)
		}
	}
	return strconv.Itoa(value)
}

// ClearCache clears the caches (useful for testing or memory management)
func (p *HeaderValuePool) ClearCache() {
	p.goroutineCache = sync.Map{}
	p.memoryCache = sync.Map{}
	// Keep intCache as it has pre-populated common values
}

// Global header value pool
var globalHeaderValuePool = NewHeaderValuePool()

// GetGlobalHeaderValuePool returns the global header value pool
func GetGlobalHeaderValuePool() *HeaderValuePool {
	return globalHeaderValuePool
}

// Performance Classification Cache for performance headers
type PerformanceClassCache struct {
	cache sync.Map
}

// NewPerformanceClassCache creates a new performance class cache
func NewPerformanceClassCache() *PerformanceClassCache {
	cache := &PerformanceClassCache{}
	
	// Pre-populate common performance classes
	commonClasses := map[string]string{
		"exceptional":      "exceptional",
		"excellent":        "excellent", 
		"good":             "good",
		"needs-optimization": "needs-optimization",
		"4x faster":        "4x faster",
		"10x faster":       "10x faster",
		"comparable":       "comparable",
	}
	
	for key, value := range commonClasses {
		cache.cache.Store(key, value)
	}
	
	return cache
}

// GetClass returns a cached performance class string
func (p *PerformanceClassCache) GetClass(class string) string {
	if cached, found := p.cache.Load(class); found {
		return cached.(string)
	}
	
	// Store and return if not found
	p.cache.Store(class, class)
	return class
}

// Global performance class cache
var globalPerformanceClassCache = NewPerformanceClassCache()

// GetGlobalPerformanceClassCache returns the global performance class cache  
func GetGlobalPerformanceClassCache() *PerformanceClassCache {
	return globalPerformanceClassCache
}