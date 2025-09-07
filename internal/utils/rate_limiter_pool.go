package utils

import (
	"sync"
	"time"

	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// CircularTimeBuffer provides a memory-efficient circular buffer for timestamps
type CircularTimeBuffer struct {
	buffer []time.Time
	size   int
	head   int
	count  int
	mu     sync.RWMutex
}

// NewCircularTimeBuffer creates a circular buffer for timestamps
func NewCircularTimeBuffer(size int) *CircularTimeBuffer {
	return &CircularTimeBuffer{
		buffer: make([]time.Time, size),
		size:   size,
	}
}

// Add adds a timestamp to the buffer
func (c *CircularTimeBuffer) Add(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buffer[c.head] = t
	c.head = (c.head + 1) % c.size
	if c.count < c.size {
		c.count++
	}
}

// CountSince returns the number of timestamps since the given time
func (c *CircularTimeBuffer) CountSince(since time.Time) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	validCount := 0
	if c.count == 0 {
		return 0
	}

	// Iterate through all valid entries
	for i := 0; i < c.count; i++ {
		idx := (c.head - 1 - i + c.size) % c.size
		if c.buffer[idx].After(since) {
			validCount++
		} else {
			// Since we're going backwards in time, we can break early
			break
		}
	}

	return validCount
}

// Clear clears all timestamps from the buffer
func (c *CircularTimeBuffer) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.head = 0
	c.count = 0
}

// OptimizedRateLimiter provides memory-efficient rate limiting with circular buffers using obcache-go
type OptimizedRateLimiter struct {
	requests *obcache.Cache
	limit    int
	window   time.Duration
	bufPool  *CircularTimeBufferPool
}

// CircularTimeBufferPool pools circular time buffers for reuse
type CircularTimeBufferPool struct {
	pool sync.Pool
	size int
}

// NewCircularTimeBufferPool creates a new buffer pool
func NewCircularTimeBufferPool(bufferSize int) *CircularTimeBufferPool {
	return &CircularTimeBufferPool{
		size: bufferSize,
		pool: sync.Pool{
			New: func() any {
				return NewCircularTimeBuffer(bufferSize)
			},
		},
	}
}

// GetBuffer gets a buffer from the pool
func (p *CircularTimeBufferPool) GetBuffer() *CircularTimeBuffer {
	buf := p.pool.Get().(*CircularTimeBuffer)
	buf.Clear() // Reset the buffer
	return buf
}

// PutBuffer returns a buffer to the pool
func (p *CircularTimeBufferPool) PutBuffer(buf *CircularTimeBuffer) {
	if buf != nil {
		p.pool.Put(buf)
	}
}

// NewOptimizedRateLimiter creates a new memory-optimized rate limiter using obcache-go
func NewOptimizedRateLimiter(limit int, window time.Duration) *OptimizedRateLimiter {
	// Create obcache configuration for rate limiting
	config := obcache.NewDefaultConfig()
	config.MaxEntries = 10000                // Support many concurrent clients
	config.DefaultTTL = window * 2           // Keep entries longer than window for efficiency
	config.CleanupInterval = 2 * time.Minute // Cleanup interval for expired entries

	// Create obcache instance
	requestsCache, err := obcache.New(config)
	if err != nil {
		// Fallback to basic config if creation fails
		basicConfig := obcache.NewDefaultConfig()
		basicConfig.DefaultTTL = window * 2
		requestsCache, _ = obcache.New(basicConfig)
	}

	rl := &OptimizedRateLimiter{
		requests: requestsCache,
		limit:    limit,
		window:   window,
		bufPool:  NewCircularTimeBufferPool(limit * 2), // Buffer size slightly larger than limit
	}

	return rl
}

// IsAllowed checks if a request is allowed for the given key
func (rl *OptimizedRateLimiter) IsAllowed(key string) bool {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get or create buffer for this key
	var buffer *CircularTimeBuffer
	if cachedBuffer, found := rl.requests.Get(key); found {
		buffer = cachedBuffer.(*CircularTimeBuffer)
	} else {
		buffer = rl.bufPool.GetBuffer()
		// Store buffer in cache with TTL equal to window duration * 2
		_ = rl.requests.Set(key, buffer, rl.window*2)
	}

	// Count valid requests in the time window
	validCount := buffer.CountSince(windowStart)

	// Check if limit exceeded
	if validCount >= rl.limit {
		return false
	}

	// Add current request
	buffer.Add(now)
	return true
}

// GetStats returns statistics about the rate limiter
func (rl *OptimizedRateLimiter) GetStats() map[string]any {
	stats := rl.requests.Stats()

	return map[string]any{
		"active_keys": int(stats.KeyCount()),
		"limit":       rl.limit,
		"window":      rl.window.String(),
		"hit_count":   int(stats.Hits()),
		"miss_count":  int(stats.Misses()),
		"hit_ratio":   stats.HitRate() * 100,
		"evictions":   int(stats.Evictions()),
		"cache_type":  "obcache-go",
	}
}

// RateLimiterManager manages multiple rate limiters efficiently using obcache-go
type RateLimiterManager struct {
	limiters *obcache.Cache
	bufPool  *CircularTimeBufferPool
}

// NewRateLimiterManager creates a new rate limiter manager using obcache-go
func NewRateLimiterManager() *RateLimiterManager {
	// Create obcache configuration for rate limiter management
	config := obcache.NewDefaultConfig()
	config.MaxEntries = 1000 // Support many different rate limiter configurations
	config.DefaultTTL = 0    // Rate limiters don't expire by default

	// Create obcache instance
	limitersCache, err := obcache.New(config)
	if err != nil {
		// Fallback to basic config if creation fails
		basicConfig := obcache.NewDefaultConfig()
		basicConfig.DefaultTTL = 0 // No expiration
		limitersCache, _ = obcache.New(basicConfig)
	}

	return &RateLimiterManager{
		limiters: limitersCache,
		bufPool:  NewCircularTimeBufferPool(100), // Default buffer size
	}
}

// GetLimiter gets or creates a rate limiter for the given configuration
func (m *RateLimiterManager) GetLimiter(name string, limit int, window time.Duration) *OptimizedRateLimiter {
	if cachedLimiter, found := m.limiters.Get(name); found {
		return cachedLimiter.(*OptimizedRateLimiter)
	}

	// Create new rate limiter
	newLimiter := NewOptimizedRateLimiter(limit, window)
	_ = m.limiters.Set(name, newLimiter, 0) // No expiration for rate limiters
	return newLimiter
}

// GetStats returns statistics for all managed rate limiters
func (m *RateLimiterManager) GetStats() map[string]any {
	// Note: obcache doesn't support enumeration, so we return cache stats instead
	// This is a limitation we accept for the performance benefits
	cacheStats := m.limiters.Stats()

	return map[string]any{
		"limiter_count": int(cacheStats.KeyCount()),
		"hit_count":     int(cacheStats.Hits()),
		"miss_count":    int(cacheStats.Misses()),
		"hit_ratio":     cacheStats.HitRate() * 100,
		"cache_type":    "obcache-go",
	}
}

// Global rate limiter manager
var globalRateLimiterManager = NewRateLimiterManager()

// GetGlobalRateLimiterManager returns the global rate limiter manager
func GetGlobalRateLimiterManager() *RateLimiterManager {
	return globalRateLimiterManager
}

// RequestIDPool provides pooled request ID generation to avoid allocations
type RequestIDPool struct {
	bufferPool  sync.Pool
	counterPool sync.Pool
}

// NewRequestIDPool creates a new request ID pool
func NewRequestIDPool() *RequestIDPool {
	return &RequestIDPool{
		bufferPool: sync.Pool{
			New: func() any {
				buf := make([]byte, 8) // For random bytes
				return &buf
			},
		},
		counterPool: sync.Pool{
			New: func() any {
				// Pre-allocated byte slice for building request ID string
				buf := make([]byte, 0, 32) // Enough capacity for typical request ID
				return &buf
			},
		},
	}
}

// GetRandomBytes gets pooled random bytes
func (p *RequestIDPool) GetRandomBytes() []byte {
	bufPtr := p.bufferPool.Get().(*[]byte)
	return *bufPtr
}

// PutRandomBytes returns random bytes to pool
func (p *RequestIDPool) PutRandomBytes(buf []byte) {
	if len(buf) == 8 {
		p.bufferPool.Put(&buf)
	}
}

// GetBuffer gets a pooled buffer for building request ID
func (p *RequestIDPool) GetBuffer() []byte {
	bufPtr := p.counterPool.Get().(*[]byte)
	buf := *bufPtr
	return buf[:0] // Reset length but keep capacity
}

// PutBuffer returns buffer to pool
func (p *RequestIDPool) PutBuffer(buf []byte) {
	if cap(buf) == 32 {
		p.counterPool.Put(&buf)
	}
}

// Global request ID pool
var globalRequestIDPool = NewRequestIDPool()

// GetGlobalRequestIDPool returns the global request ID pool
func GetGlobalRequestIDPool() *RequestIDPool {
	return globalRequestIDPool
}

// Cleanup methods for memory leak prevention

// Cleanup clears all active rate limiters and buffers (for memory leak prevention)
func (m *RateLimiterManager) Cleanup() {
	// Clear all rate limiters using obcache
	_ = m.limiters.Clear()

	// Reset buffer pool
	m.bufPool = NewCircularTimeBufferPool(100)
}

// Cleanup clears all request buffers (for OptimizedRateLimiter)
func (rl *OptimizedRateLimiter) Cleanup() {
	// Clear all request tracking using obcache
	// Note: We can't enumerate and return buffers to pool due to obcache limitations
	// This is acceptable as buffers will be garbage collected
	_ = rl.requests.Clear()
}

// GetStats returns more detailed statistics for RateLimiterManager
func (m *RateLimiterManager) GetStatsDetailed() map[string]interface{} {
	cacheStats := m.limiters.Stats()

	// Enhanced statistics with cache metrics
	return map[string]interface{}{
		"limiter_count":     int(cacheStats.KeyCount()),
		"hit_count":         int(cacheStats.Hits()),
		"miss_count":        int(cacheStats.Misses()),
		"hit_ratio":         cacheStats.HitRate() * 100,
		"evictions":         int(cacheStats.Evictions()),
		"cache_type":        "obcache-go",
		"type":              "RateLimiterManager",
		"total_active_keys": "N/A - obcache limitation",
	}
}

// Cleanup clears all pooled request ID buffers (for memory leak prevention)
func (p *RequestIDPool) Cleanup() {
	p.bufferPool = sync.Pool{
		New: func() any {
			buf := make([]byte, 8)
			return &buf
		},
	}
	p.counterPool = sync.Pool{
		New: func() any {
			buf := make([]byte, 0, 32)
			return &buf
		},
	}
}

// GetStats returns statistics for the request ID pool
func (p *RequestIDPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "RequestIDPool",
		"note": "sync.Pool doesn't expose internal metrics",
	}
}
