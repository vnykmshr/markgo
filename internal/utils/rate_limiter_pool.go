package utils

import (
	"sync"
	"time"
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

// OptimizedRateLimiter provides memory-efficient rate limiting with circular buffers
type OptimizedRateLimiter struct {
	requests sync.Map // map[string]*CircularTimeBuffer
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

// NewOptimizedRateLimiter creates a new memory-optimized rate limiter
func NewOptimizedRateLimiter(limit int, window time.Duration) *OptimizedRateLimiter {
	rl := &OptimizedRateLimiter{
		limit:   limit,
		window:  window,
		bufPool: NewCircularTimeBufferPool(limit * 2), // Buffer size slightly larger than limit
	}

	// Start cleanup goroutine with less frequent cleanup
	go rl.cleanup()

	return rl
}

// IsAllowed checks if a request is allowed for the given key
func (rl *OptimizedRateLimiter) IsAllowed(key string) bool {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get or create buffer for this key
	bufferInterface, _ := rl.requests.LoadOrStore(key, rl.bufPool.GetBuffer())
	buffer := bufferInterface.(*CircularTimeBuffer)

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

// cleanup periodically removes inactive entries to prevent memory leaks
func (rl *OptimizedRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Less frequent cleanup
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		windowStart := now.Add(-rl.window * 2) // Keep entries a bit longer for efficiency

		var keysToDelete []string

		rl.requests.Range(func(key, value any) bool {
			buffer := value.(*CircularTimeBuffer)

			// Check if this key has any recent activity
			if buffer.CountSince(windowStart) == 0 {
				keysToDelete = append(keysToDelete, key.(string))
				// Return buffer to pool
				rl.bufPool.PutBuffer(buffer)
			}
			return true
		})

		// Delete inactive keys
		for _, key := range keysToDelete {
			rl.requests.Delete(key)
		}
	}
}

// GetStats returns statistics about the rate limiter
func (rl *OptimizedRateLimiter) GetStats() map[string]any {
	activeKeys := 0
	rl.requests.Range(func(key, value any) bool {
		activeKeys++
		return true
	})

	return map[string]any{
		"active_keys": activeKeys,
		"limit":       rl.limit,
		"window":      rl.window.String(),
	}
}

// RateLimiterManager manages multiple rate limiters efficiently
type RateLimiterManager struct {
	limiters sync.Map
	bufPool  *CircularTimeBufferPool
}

// NewRateLimiterManager creates a new rate limiter manager
func NewRateLimiterManager() *RateLimiterManager {
	return &RateLimiterManager{
		bufPool: NewCircularTimeBufferPool(100), // Default buffer size
	}
}

// GetLimiter gets or creates a rate limiter for the given configuration
func (m *RateLimiterManager) GetLimiter(name string, limit int, window time.Duration) *OptimizedRateLimiter {
	limiterInterface, _ := m.limiters.LoadOrStore(name, NewOptimizedRateLimiter(limit, window))
	return limiterInterface.(*OptimizedRateLimiter)
}

// GetStats returns statistics for all managed rate limiters
func (m *RateLimiterManager) GetStats() map[string]any {
	stats := make(map[string]any)

	m.limiters.Range(func(key, value any) bool {
		limiter := value.(*OptimizedRateLimiter)
		stats[key.(string)] = limiter.GetStats()
		return true
	})

	return stats
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
	// Clear all rate limiters
	m.limiters.Range(func(key, value interface{}) bool {
		m.limiters.Delete(key)
		return true
	})

	// Reset buffer pool
	m.bufPool = NewCircularTimeBufferPool(100)
}

// Cleanup clears all request buffers (for OptimizedRateLimiter)
func (rl *OptimizedRateLimiter) Cleanup() {
	// Clear all request tracking
	rl.requests.Range(func(key, value interface{}) bool {
		buffer := value.(*CircularTimeBuffer)
		rl.bufPool.PutBuffer(buffer)
		rl.requests.Delete(key)
		return true
	})
}

// GetStats returns more detailed statistics for RateLimiterManager
func (m *RateLimiterManager) GetStatsDetailed() map[string]interface{} {
	stats := m.GetStats()

	totalActiveKeys := 0
	limiterCount := 0

	m.limiters.Range(func(key, value interface{}) bool {
		limiter := value.(*OptimizedRateLimiter)
		limiterStats := limiter.GetStats()
		if activeKeys, ok := limiterStats["active_keys"].(int); ok {
			totalActiveKeys += activeKeys
		}
		limiterCount++
		return true
	})

	stats["total_active_keys"] = totalActiveKeys
	stats["limiter_count"] = limiterCount
	stats["type"] = "RateLimiterManager"

	return stats
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
