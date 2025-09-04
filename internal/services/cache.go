package services

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Ensure CacheService implements CacheServiceInterface
var _ CacheServiceInterface = (*CacheService)(nil)

type CacheItem struct {
	Value          any
	CompressedData []byte // Compressed data when using compression
	Expiration     time.Time
	IsCompressed   bool
	OriginalSize   int64 // Size before compression
	CompressedSize int64 // Size after compression
}

type CacheService struct {
	items       map[string]*CacheItem
	mutex       sync.RWMutex
	defaultTTL  time.Duration
	maxSize     int
	cleanupTick *time.Ticker
	stopCleanup chan bool

	// Memory management
	enableCompression    bool
	compressionThreshold int64 // Minimum size to compress (bytes)
	maxMemoryUsage       int64 // Max memory usage before eviction (bytes)
	currentMemoryUsage   int64 // Current estimated memory usage (atomic)
	hitCount             int64 // Cache hits (atomic)
	missCount            int64 // Cache misses (atomic)
	compressionCount     int64 // Number of compressed items (atomic)
}

func NewCacheService(defaultTTL time.Duration, maxSize int) *CacheService {
	cache := &CacheService{
		items:                make(map[string]*CacheItem),
		defaultTTL:           defaultTTL,
		maxSize:              maxSize,
		cleanupTick:          time.NewTicker(10 * time.Minute),
		stopCleanup:          make(chan bool),
		enableCompression:    true,
		compressionThreshold: 1024,     // Compress items larger than 1KB
		maxMemoryUsage:       50 << 20, // 50MB max cache memory
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
			c.removeItemMemoryUsage(item)
			delete(c.items, key)
		}
	}
}

func (c *CacheService) Set(key string, value any, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check memory usage and evict if necessary
	c.checkMemoryUsageAndEvict()

	// If cache is full, remove oldest item
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	if ttl == 0 {
		ttl = c.defaultTTL
	}

	item := &CacheItem{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}

	// Try to compress the value if compression is enabled
	if c.enableCompression {
		if compressedData, originalSize, compressedSize, err := c.compressValue(value); err == nil {
			if compressedSize < originalSize && originalSize >= c.compressionThreshold {
				item.CompressedData = compressedData
				item.IsCompressed = true
				item.OriginalSize = originalSize
				item.CompressedSize = compressedSize
				item.Value = nil // Clear the original value to save memory
				atomic.AddInt64(&c.compressionCount, 1)
				atomic.AddInt64(&c.currentMemoryUsage, compressedSize)
			} else {
				// Store uncompressed
				item.OriginalSize = originalSize
				atomic.AddInt64(&c.currentMemoryUsage, originalSize)
			}
		} else {
			// Compression failed, store uncompressed
			item.OriginalSize = c.estimateValueSize(value)
			atomic.AddInt64(&c.currentMemoryUsage, item.OriginalSize)
		}
	} else {
		item.OriginalSize = c.estimateValueSize(value)
		atomic.AddInt64(&c.currentMemoryUsage, item.OriginalSize)
	}

	// Remove old item memory usage if it exists
	if oldItem, exists := c.items[key]; exists {
		if oldItem.IsCompressed {
			atomic.AddInt64(&c.currentMemoryUsage, -oldItem.CompressedSize)
		} else {
			atomic.AddInt64(&c.currentMemoryUsage, -oldItem.OriginalSize)
		}
	}

	c.items[key] = item
}

func (c *CacheService) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		atomic.AddInt64(&c.missCount, 1)
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.Expiration) {
		// Remove expired item
		c.mutex.RUnlock()
		c.mutex.Lock()
		c.removeItemMemoryUsage(item)
		delete(c.items, key)
		c.mutex.Unlock()
		c.mutex.RLock()
		atomic.AddInt64(&c.missCount, 1)
		return nil, false
	}

	atomic.AddInt64(&c.hitCount, 1)

	// If item is compressed, decompress it
	if item.IsCompressed {
		if value, err := c.decompressValue(item.CompressedData); err == nil {
			return value, true
		}
		// Decompression failed, return nil
		return nil, false
	}

	return item.Value, true
}

func (c *CacheService) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, exists := c.items[key]; exists {
		c.removeItemMemoryUsage(item)
	}
	delete(c.items, key)
}

func (c *CacheService) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
	atomic.StoreInt64(&c.currentMemoryUsage, 0)
	atomic.StoreInt64(&c.hitCount, 0)
	atomic.StoreInt64(&c.missCount, 0)
	atomic.StoreInt64(&c.compressionCount, 0)
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
		if item, exists := c.items[oldestKey]; exists {
			c.removeItemMemoryUsage(item)
		}
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
	compressed := 0
	totalOriginalSize := int64(0)
	totalCompressedSize := int64(0)

	now := time.Now()
	for _, item := range c.items {
		if now.After(item.Expiration) {
			expired++
		}
		if item.IsCompressed {
			compressed++
			totalCompressedSize += item.CompressedSize
		}
		totalOriginalSize += item.OriginalSize
	}

	hitCount := atomic.LoadInt64(&c.hitCount)
	missCount := atomic.LoadInt64(&c.missCount)
	totalRequests := hitCount + missCount
	hitRatio := 0.0
	if totalRequests > 0 {
		hitRatio = float64(hitCount) / float64(totalRequests)
	}

	compressionRatio := 0.0
	if totalOriginalSize > 0 {
		compressionRatio = float64(totalCompressedSize) / float64(totalOriginalSize)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"total_items":           len(c.items),
		"expired_items":         expired,
		"compressed_items":      compressed,
		"max_size":              c.maxSize,
		"default_ttl":           c.defaultTTL.String(),
		"hit_count":             hitCount,
		"miss_count":            missCount,
		"hit_ratio":             hitRatio,
		"compression_ratio":     compressionRatio,
		"memory_usage":          atomic.LoadInt64(&c.currentMemoryUsage),
		"max_memory":            c.maxMemoryUsage,
		"heap_alloc":            m.Alloc,
		"heap_sys":              m.HeapSys,
		"enable_compression":    c.enableCompression,
		"compression_threshold": c.compressionThreshold,
	}
}

// Compression helper methods

// compressValue compresses a value using gzip and returns the compressed data along with size information
func (c *CacheService) compressValue(value any) ([]byte, int64, int64, error) {
	// For simple types that compress well, use a simpler approach
	switch v := value.(type) {
	case string:
		if len(v) < int(c.compressionThreshold) {
			return nil, int64(len(v)), int64(len(v)), fmt.Errorf("too small to compress")
		}
		return c.compressBytes([]byte(v), int64(len(v)), true)
	case []byte:
		if len(v) < int(c.compressionThreshold) {
			return nil, int64(len(v)), int64(len(v)), fmt.Errorf("too small to compress")
		}
		return c.compressBytes(v, int64(len(v)), false)
	default:
		// Use gob for complex types
		return c.compressWithGob(value)
	}
}

// compressBytes compresses raw bytes and stores type info
func (c *CacheService) compressBytes(data []byte, originalSize int64, isString bool) ([]byte, int64, int64, error) {
	var compressedBuf bytes.Buffer
	writer := gzip.NewWriter(&compressedBuf)
	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, 0, 0, fmt.Errorf("failed to compress data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to close compressor: %w", err)
	}

	var typeMarker byte = 0x01 // []byte
	if isString {
		typeMarker = 0x03 // string
	}
	// Add a type marker at the beginning
	result := append([]byte{typeMarker}, compressedBuf.Bytes()...)
	return result, originalSize, int64(len(result)), nil
}

// compressWithGob compresses a value using gob encoding then gzip
func (c *CacheService) compressWithGob(value any) ([]byte, int64, int64, error) {
	var gobBuf bytes.Buffer
	encoder := gob.NewEncoder(&gobBuf)
	if err := encoder.Encode(value); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to encode value: %w", err)
	}

	originalSize := int64(gobBuf.Len())

	// Compress the serialized data
	var compressedBuf bytes.Buffer
	writer := gzip.NewWriter(&compressedBuf)
	if _, err := writer.Write(gobBuf.Bytes()); err != nil {
		writer.Close()
		return nil, 0, 0, fmt.Errorf("failed to compress data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to close compressor: %w", err)
	}

	// Add a type marker at the beginning to distinguish from raw bytes
	result := append([]byte{0x02}, compressedBuf.Bytes()...)
	return result, originalSize, int64(len(result)), nil
}

// decompressValue decompresses and deserializes a compressed value
func (c *CacheService) decompressValue(compressedData []byte) (any, error) {
	if len(compressedData) == 0 {
		return nil, fmt.Errorf("empty compressed data")
	}

	// Check the type marker
	typeMarker := compressedData[0]
	actualData := compressedData[1:]

	switch typeMarker {
	case 0x01:
		// []byte
		return c.decompressBytes(actualData, false)
	case 0x02:
		// Gob-encoded data
		return c.decompressWithGob(actualData)
	case 0x03:
		// string
		return c.decompressBytes(actualData, true)
	default:
		// For backward compatibility, assume gob if no marker (old format)
		return c.decompressWithGob(compressedData)
	}
}

// decompressBytes decompresses raw bytes and converts to appropriate type
func (c *CacheService) decompressBytes(compressedData []byte, asString bool) (any, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	var decompressedBuf bytes.Buffer
	if _, err := decompressedBuf.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	if asString {
		return decompressedBuf.String(), nil
	}
	return decompressedBuf.Bytes(), nil
}

// decompressWithGob decompresses gob-encoded data
func (c *CacheService) decompressWithGob(compressedData []byte) (any, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	var decompressedBuf bytes.Buffer
	if _, err := decompressedBuf.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Deserialize the value using gob
	decoder := gob.NewDecoder(&decompressedBuf)
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("failed to decode value: %w", err)
	}

	return value, nil
}

// estimateValueSize estimates the memory size of a value
func (c *CacheService) estimateValueSize(value any) int64 {
	// This is a rough estimation - could be improved with reflection for more accuracy
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case map[string]any:
		size := int64(0)
		for k, val := range v {
			size += int64(len(k)) + c.estimateValueSize(val)
		}
		return size
	case []any:
		size := int64(0)
		for _, val := range v {
			size += c.estimateValueSize(val)
		}
		return size
	default:
		// Rough estimate for other types - could use unsafe.Sizeof() for more precision
		return int64(64) // Default estimate
	}
}

// removeItemMemoryUsage removes the memory usage tracking for an item
func (c *CacheService) removeItemMemoryUsage(item *CacheItem) {
	if item.IsCompressed {
		atomic.AddInt64(&c.currentMemoryUsage, -item.CompressedSize)
		atomic.AddInt64(&c.compressionCount, -1)
	} else {
		atomic.AddInt64(&c.currentMemoryUsage, -item.OriginalSize)
	}
}

// checkMemoryUsageAndEvict checks current memory usage and evicts items if necessary
func (c *CacheService) checkMemoryUsageAndEvict() {
	currentUsage := atomic.LoadInt64(&c.currentMemoryUsage)

	// Check if we're over the memory limit
	if currentUsage > c.maxMemoryUsage {
		c.evictByMemoryPressure()
	}

	// Also check system memory pressure
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// If heap usage is over 80% of system memory or growing too fast, be more aggressive
	if m.Alloc > m.Sys*4/5 || currentUsage > c.maxMemoryUsage/2 {
		c.evictByMemoryPressure()
	}
}

// evictByMemoryPressure evicts items based on memory pressure
func (c *CacheService) evictByMemoryPressure() {
	// Collect items sorted by access pattern and size
	type itemInfo struct {
		key      string
		item     *CacheItem
		priority float64 // Higher priority = more likely to be evicted
	}

	var items []itemInfo
	now := time.Now()

	for key, item := range c.items {
		// Calculate priority based on:
		// 1. Time until expiration (items expiring soon have higher priority for eviction)
		// 2. Size (larger items have higher priority for eviction)
		// 3. Compression status (uncompressed items have higher priority)

		timeToExpiration := item.Expiration.Sub(now).Seconds()
		size := item.OriginalSize
		if item.IsCompressed {
			size = item.CompressedSize
		}

		priority := float64(size) / (timeToExpiration + 1) // +1 to avoid division by zero
		if !item.IsCompressed && c.enableCompression {
			priority *= 2 // Prefer evicting uncompressed items
		}

		items = append(items, itemInfo{
			key:      key,
			item:     item,
			priority: priority,
		})
	}

	// Sort by priority (highest first)
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].priority < items[j].priority {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Evict items until we're under the memory limit
	targetUsage := c.maxMemoryUsage * 3 / 4 // Evict until we're at 75% of max
	currentUsage := atomic.LoadInt64(&c.currentMemoryUsage)

	for i := 0; i < len(items) && currentUsage > targetUsage; i++ {
		item := items[i]
		c.removeItemMemoryUsage(item.item)
		delete(c.items, item.key)
		currentUsage = atomic.LoadInt64(&c.currentMemoryUsage)
	}
}

// SetCompressionSettings allows configuration of compression settings
func (c *CacheService) SetCompressionSettings(enabled bool, threshold int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.enableCompression = enabled
	c.compressionThreshold = threshold
}

// SetMaxMemoryUsage sets the maximum memory usage limit
func (c *CacheService) SetMaxMemoryUsage(maxMemory int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.maxMemoryUsage = maxMemory
}
