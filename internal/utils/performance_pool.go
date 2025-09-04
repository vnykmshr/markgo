package utils

import (
	"sync"
	"time"
)

// PerformanceMetricsPool provides pooled allocations for performance tracking
type PerformanceMetricsPool struct {
	responseTimesPool sync.Pool
	stringMapPool     sync.Pool
	int64MapPool      sync.Pool
	durationSlicePool sync.Pool
}

// NewPerformanceMetricsPool creates a new performance metrics pool
func NewPerformanceMetricsPool() *PerformanceMetricsPool {
	return &PerformanceMetricsPool{
		responseTimesPool: sync.Pool{
			New: func() any {
				// Pre-allocate with capacity for 1000 response times
				slice := make([]time.Duration, 0, 1000)
				return &slice
			},
		},
		stringMapPool: sync.Pool{
			New: func() any {
				return make(map[string]int64, 50) // Pre-allocate for common endpoints
			},
		},
		int64MapPool: sync.Pool{
			New: func() any {
				return make(map[string]time.Duration, 50) // Pre-allocate for common endpoints
			},
		},
		durationSlicePool: sync.Pool{
			New: func() any {
				// For percentile calculations
				slice := make([]time.Duration, 0, 1000)
				return &slice
			},
		},
	}
}

// GetResponseTimesSlice gets a pooled response times slice
func (p *PerformanceMetricsPool) GetResponseTimesSlice() *[]time.Duration {
	slice := p.responseTimesPool.Get().(*[]time.Duration)
	*slice = (*slice)[:0] // Reset length but keep capacity
	return slice
}

// PutResponseTimesSlice returns a response times slice to the pool
func (p *PerformanceMetricsPool) PutResponseTimesSlice(slice *[]time.Duration) {
	if slice != nil && cap(*slice) <= 2000 { // Prevent memory leaks from huge slices
		p.responseTimesPool.Put(slice)
	}
}

// GetStringInt64Map gets a pooled string->int64 map
func (p *PerformanceMetricsPool) GetStringInt64Map() map[string]int64 {
	m := p.stringMapPool.Get().(map[string]int64)
	// Clear the map but keep allocated memory
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutStringInt64Map returns a string->int64 map to the pool
func (p *PerformanceMetricsPool) PutStringInt64Map(m map[string]int64) {
	if m != nil && len(m) <= 100 { // Prevent memory leaks from huge maps
		p.stringMapPool.Put(m)
	}
}

// GetStringDurationMap gets a pooled string->duration map
func (p *PerformanceMetricsPool) GetStringDurationMap() map[string]time.Duration {
	m := p.int64MapPool.Get().(map[string]time.Duration)
	// Clear the map but keep allocated memory
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutStringDurationMap returns a string->duration map to the pool
func (p *PerformanceMetricsPool) PutStringDurationMap(m map[string]time.Duration) {
	if m != nil && len(m) <= 100 { // Prevent memory leaks from huge maps
		p.int64MapPool.Put(m)
	}
}

// GetSortingSlice gets a pooled slice for sorting (percentile calculations)
func (p *PerformanceMetricsPool) GetSortingSlice() *[]time.Duration {
	slice := p.durationSlicePool.Get().(*[]time.Duration)
	*slice = (*slice)[:0] // Reset length but keep capacity
	return slice
}

// PutSortingSlice returns a sorting slice to the pool
func (p *PerformanceMetricsPool) PutSortingSlice(slice *[]time.Duration) {
	if slice != nil && cap(*slice) <= 2000 { // Prevent memory leaks from huge slices
		p.durationSlicePool.Put(slice)
	}
}

// WithResponseTimesSlice provides a convenient way to use pooled response times slice
func (p *PerformanceMetricsPool) WithResponseTimesSlice(fn func(*[]time.Duration)) {
	slice := p.GetResponseTimesSlice()
	defer p.PutResponseTimesSlice(slice)
	fn(slice)
}

// WithStringInt64Map provides a convenient way to use pooled string->int64 map
func (p *PerformanceMetricsPool) WithStringInt64Map(fn func(map[string]int64) map[string]int64) map[string]int64 {
	m := p.GetStringInt64Map()
	defer p.PutStringInt64Map(m)
	return fn(m)
}

// WithStringDurationMap provides a convenient way to use pooled string->duration map
func (p *PerformanceMetricsPool) WithStringDurationMap(fn func(map[string]time.Duration) map[string]time.Duration) map[string]time.Duration {
	m := p.GetStringDurationMap()
	defer p.PutStringDurationMap(m)
	return fn(m)
}

// Global performance pool for application-wide use
var globalPerformancePool = NewPerformanceMetricsPool()

// GetGlobalPerformancePool returns the global performance metrics pool
func GetGlobalPerformancePool() *PerformanceMetricsPool {
	return globalPerformancePool
}

// CircularResponseTimeBuffer provides a memory-efficient circular buffer for response times
type CircularResponseTimeBuffer struct {
	buffer []time.Duration
	size   int
	head   int
	count  int
	mu     sync.RWMutex
}

// NewCircularResponseTimeBuffer creates a circular buffer for response times
func NewCircularResponseTimeBuffer(size int) *CircularResponseTimeBuffer {
	return &CircularResponseTimeBuffer{
		buffer: make([]time.Duration, size),
		size:   size,
	}
}

// Add adds a response time to the buffer
func (c *CircularResponseTimeBuffer) Add(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buffer[c.head] = duration
	c.head = (c.head + 1) % c.size
	if c.count < c.size {
		c.count++
	}
}

// GetAll returns a copy of all current response times
func (c *CircularResponseTimeBuffer) GetAll() []time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.count == 0 {
		return nil
	}

	result := make([]time.Duration, c.count)
	if c.count < c.size {
		// Buffer not full yet, copy from beginning
		copy(result, c.buffer[:c.count])
	} else {
		// Buffer is full, copy in correct order
		copy(result, c.buffer[c.head:])
		copy(result[c.size-c.head:], c.buffer[:c.head])
	}

	return result
}

// GetSorted returns a sorted copy of all current response times using pooled slice
func (c *CircularResponseTimeBuffer) GetSorted() []time.Duration {
	pool := GetGlobalPerformancePool()
	sortingSlice := pool.GetSortingSlice()
	defer pool.PutSortingSlice(sortingSlice)

	c.mu.RLock()
	times := c.GetAll()
	c.mu.RUnlock()

	if len(times) == 0 {
		return nil
	}

	// Use pooled slice for sorting
	*sortingSlice = append(*sortingSlice, times...)

	// Optimized insertion sort for small datasets, quicksort for large
	if len(*sortingSlice) <= 50 {
		insertionSort(*sortingSlice)
	} else {
		quickSort(*sortingSlice, 0, len(*sortingSlice)-1)
	}

	// Return a copy since we'll put the slice back in the pool
	result := make([]time.Duration, len(*sortingSlice))
	copy(result, *sortingSlice)

	return result
}

// Count returns the number of response times currently in the buffer
func (c *CircularResponseTimeBuffer) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}

// insertionSort optimized insertion sort for small slices
func insertionSort(arr []time.Duration) {
	for i := 1; i < len(arr); i++ {
		key := arr[i]
		j := i - 1
		for j >= 0 && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
}

// quickSort optimized quicksort implementation
func quickSort(arr []time.Duration, low, high int) {
	if low < high {
		pivot := partition(arr, low, high)
		quickSort(arr, low, pivot-1)
		quickSort(arr, pivot+1, high)
	}
}

func partition(arr []time.Duration, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] <= pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}
