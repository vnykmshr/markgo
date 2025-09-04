package middleware

import (
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/markgo/internal/utils"
)

// PerformanceMetrics holds performance monitoring data with memory optimization
type PerformanceMetrics struct {
	// Atomic fields for high-frequency updates (must be first for alignment)
	requestCount      int64
	totalResponseTime int64 // nanoseconds
	maxResponseTime   int64 // nanoseconds
	minResponseTime   int64 // nanoseconds, use atomic.LoadInt64 with special handling
	memoryUsage       uint64
	
	// Less frequently updated fields protected by mutex
	mu                      sync.RWMutex
	responseTimeBuffer      *utils.CircularResponseTimeBuffer
	requestsPerSecond       float64
	goroutineCount          int
	lastUpdateTime          time.Time
	requestsByEndpoint      map[string]int64
	responseTimesByEndpoint map[string]time.Duration
	pool                    *utils.PerformanceMetricsPool
}

var (
	globalMetrics = &PerformanceMetrics{
		minResponseTime:         int64(time.Hour), // Initialize to high value
		requestsByEndpoint:      make(map[string]int64),
		responseTimesByEndpoint: make(map[string]time.Duration),
		lastUpdateTime:          time.Now(),
		responseTimeBuffer:      utils.NewCircularResponseTimeBuffer(1000),
		pool:                    utils.GetGlobalPerformancePool(),
	}
)

// PerformanceMiddleware creates a middleware that tracks performance metrics
func PerformanceMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Only record detailed memory stats for slow requests or monitoring
		var memBefore runtime.MemStats
		shouldTrackMemory := atomic.LoadInt64(&globalMetrics.requestCount)%10 == 0 // Every 10th request
		if shouldTrackMemory {
			runtime.ReadMemStats(&memBefore)
		}

		// Process request
		c.Next()

		// Calculate response time
		responseTime := time.Since(startTime)
		endpoint := c.Request.Method + " " + c.FullPath()

		// Update metrics
		updateMetrics(responseTime, endpoint)

		// Add basic performance headers (always include response time)
		c.Header("X-Response-Time", responseTime.String())
		c.Header("X-Goroutines", utils.GetGlobalHeaderValuePool().GetGoroutineCountString(runtime.NumGoroutine()))

		// Add memory delta header only when we're tracking memory
		if shouldTrackMemory {
			var memAfter runtime.MemStats
			runtime.ReadMemStats(&memAfter)
			c.Header("X-Memory-Delta", utils.GetGlobalHeaderValuePool().GetMemoryDeltaString(memAfter.Alloc-memBefore.Alloc))

			// Log slow requests with memory details
			if responseTime > 100*time.Millisecond {
				logger.Warn("Slow request detected",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"response_time", responseTime.String(),
					"status", c.Writer.Status(),
					"memory_delta", memAfter.Alloc-memBefore.Alloc,
					"goroutines", runtime.NumGoroutine(),
				)
			}
		} else {
			// Log slow requests without memory details (to avoid ReadMemStats overhead)
			if responseTime > 100*time.Millisecond {
				logger.Warn("Slow request detected",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"response_time", responseTime.String(),
					"status", c.Writer.Status(),
					"goroutines", runtime.NumGoroutine(),
				)
			}
		}

		// Log detailed performance metrics every 100 requests
		if count := atomic.LoadInt64(&globalMetrics.requestCount); count%100 == 0 {
			logPerformanceSummary(logger)
		}
	}
}

func updateMetrics(responseTime time.Duration, endpoint string) {
	responseTimeNs := responseTime.Nanoseconds()
	
	// Update atomic counters (lock-free)
	atomic.AddInt64(&globalMetrics.requestCount, 1)
	atomic.AddInt64(&globalMetrics.totalResponseTime, responseTimeNs)
	
	// Update max response time (lock-free)
	for {
		oldMax := atomic.LoadInt64(&globalMetrics.maxResponseTime)
		if responseTimeNs <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt64(&globalMetrics.maxResponseTime, oldMax, responseTimeNs) {
			break
		}
	}
	
	// Update min response time (lock-free)
	for {
		oldMin := atomic.LoadInt64(&globalMetrics.minResponseTime)
		if responseTimeNs >= oldMin {
			break
		}
		if atomic.CompareAndSwapInt64(&globalMetrics.minResponseTime, oldMin, responseTimeNs) {
			break
		}
	}
	
	// Store in circular buffer (lock-based but efficient)
	globalMetrics.responseTimeBuffer.Add(responseTime)
	
	// Update less frequent metrics with mutex
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	// Update requests per second
	now := time.Now()
	if elapsed := now.Sub(globalMetrics.lastUpdateTime); elapsed > 0 {
		globalMetrics.requestsPerSecond = float64(atomic.LoadInt64(&globalMetrics.requestCount)) / elapsed.Seconds()
	}
	
	// Update per-endpoint metrics
	globalMetrics.requestsByEndpoint[endpoint]++
	if existingTime, exists := globalMetrics.responseTimesByEndpoint[endpoint]; exists {
		globalMetrics.responseTimesByEndpoint[endpoint] = (existingTime + responseTime) / 2
	} else {
		globalMetrics.responseTimesByEndpoint[endpoint] = responseTime
	}
	
	// Update memory usage less frequently (every 10th request to reduce overhead)
	if atomic.LoadInt64(&globalMetrics.requestCount)%10 == 0 {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		atomic.StoreUint64(&globalMetrics.memoryUsage, mem.Alloc)
		globalMetrics.goroutineCount = runtime.NumGoroutine()
	}
}

func logPerformanceSummary(logger *slog.Logger) {
	// Get atomic values
	requestCount := atomic.LoadInt64(&globalMetrics.requestCount)
	totalResponseTime := atomic.LoadInt64(&globalMetrics.totalResponseTime)
	maxResponseTime := atomic.LoadInt64(&globalMetrics.maxResponseTime)
	minResponseTime := atomic.LoadInt64(&globalMetrics.minResponseTime)
	memoryUsage := atomic.LoadUint64(&globalMetrics.memoryUsage)
	
	// Calculate average
	var avgResponseTime time.Duration
	if requestCount > 0 {
		avgResponseTime = time.Duration(totalResponseTime / requestCount)
	}
	
	// Get percentiles from circular buffer
	responseTimes := globalMetrics.responseTimeBuffer.GetSorted()
	p95 := calculatePercentile(responseTimes, 0.95)
	p99 := calculatePercentile(responseTimes, 0.99)
	
	globalMetrics.mu.RLock()
	rps := globalMetrics.requestsPerSecond
	goroutines := globalMetrics.goroutineCount
	globalMetrics.mu.RUnlock()

	logger.Info("Performance Summary",
		"total_requests", requestCount,
		"avg_response_time", avgResponseTime.String(),
		"min_response_time", time.Duration(minResponseTime).String(),
		"max_response_time", time.Duration(maxResponseTime).String(),
		"p95_response_time", p95.String(),
		"p99_response_time", p99.String(),
		"requests_per_second", rps,
		"memory_usage_mb", memoryUsage/1024/1024,
		"goroutine_count", goroutines,
	)
}

func calculatePercentile(sortedTimes []time.Duration, percentile float64) time.Duration {
	if len(sortedTimes) == 0 {
		return 0
	}

	// Times are already sorted from GetSorted() method
	index := int(percentile * float64(len(sortedTimes)-1))
	if index >= len(sortedTimes) {
		index = len(sortedTimes) - 1
	}
	return sortedTimes[index]
}

// PublicPerformanceMetrics is the public interface for performance metrics
type PublicPerformanceMetrics struct {
	RequestCount            int64
	TotalResponseTime       time.Duration
	AverageResponseTime     time.Duration
	MaxResponseTime         time.Duration
	MinResponseTime         time.Duration
	ResponseTimes           []time.Duration
	RequestsPerSecond       float64
	MemoryUsage             uint64
	GoroutineCount          int
	LastUpdateTime          time.Time
	RequestsByEndpoint      map[string]int64
	ResponseTimesByEndpoint map[string]time.Duration
}

// GetPerformanceMetrics returns current performance metrics with pooled maps
func GetPerformanceMetrics() PublicPerformanceMetrics {
	// Get atomic values (lock-free)
	requestCount := atomic.LoadInt64(&globalMetrics.requestCount)
	totalResponseTime := atomic.LoadInt64(&globalMetrics.totalResponseTime)
	maxResponseTime := atomic.LoadInt64(&globalMetrics.maxResponseTime)
	minResponseTime := atomic.LoadInt64(&globalMetrics.minResponseTime)
	memoryUsage := atomic.LoadUint64(&globalMetrics.memoryUsage)
	
	// Calculate average
	var averageResponseTime time.Duration
	if requestCount > 0 {
		averageResponseTime = time.Duration(totalResponseTime / requestCount)
	}
	
	// Get mutex-protected values
	globalMetrics.mu.RLock()
	requestsPerSecond := globalMetrics.requestsPerSecond
	goroutineCount := globalMetrics.goroutineCount
	lastUpdateTime := globalMetrics.lastUpdateTime
	
	// Use pooled maps for efficient copying
	requestsByEndpoint := globalMetrics.pool.WithStringInt64Map(func(pooledMap map[string]int64) map[string]int64 {
		for k, v := range globalMetrics.requestsByEndpoint {
			pooledMap[k] = v
		}
		// Return a copy since the pooled map will be returned to pool
		result := make(map[string]int64, len(pooledMap))
		for k, v := range pooledMap {
			result[k] = v
		}
		return result
	})
	
	responseTimesByEndpoint := globalMetrics.pool.WithStringDurationMap(func(pooledMap map[string]time.Duration) map[string]time.Duration {
		for k, v := range globalMetrics.responseTimesByEndpoint {
			pooledMap[k] = v
		}
		// Return a copy since the pooled map will be returned to pool
		result := make(map[string]time.Duration, len(pooledMap))
		for k, v := range pooledMap {
			result[k] = v
		}
		return result
	})
	globalMetrics.mu.RUnlock()
	
	// Get response times from circular buffer
	responseTimes := globalMetrics.responseTimeBuffer.GetAll()

	return PublicPerformanceMetrics{
		RequestCount:            requestCount,
		TotalResponseTime:       time.Duration(totalResponseTime),
		AverageResponseTime:     averageResponseTime,
		MaxResponseTime:         time.Duration(maxResponseTime),
		MinResponseTime:         time.Duration(minResponseTime),
		ResponseTimes:           responseTimes,
		RequestsPerSecond:       requestsPerSecond,
		MemoryUsage:            memoryUsage,
		GoroutineCount:         goroutineCount,
		LastUpdateTime:         lastUpdateTime,
		RequestsByEndpoint:     requestsByEndpoint,
		ResponseTimesByEndpoint: responseTimesByEndpoint,
	}
}

// ResetPerformanceMetrics resets all performance metrics (useful for testing)
func ResetPerformanceMetrics() {
	// Reset atomic values
	atomic.StoreInt64(&globalMetrics.requestCount, 0)
	atomic.StoreInt64(&globalMetrics.totalResponseTime, 0)
	atomic.StoreInt64(&globalMetrics.maxResponseTime, 0)
	atomic.StoreInt64(&globalMetrics.minResponseTime, int64(time.Hour))
	atomic.StoreUint64(&globalMetrics.memoryUsage, 0)

	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.requestsPerSecond = 0
	globalMetrics.goroutineCount = 0
	globalMetrics.lastUpdateTime = time.Now()
	
	// Clear maps
	for k := range globalMetrics.requestsByEndpoint {
		delete(globalMetrics.requestsByEndpoint, k)
	}
	for k := range globalMetrics.responseTimesByEndpoint {
		delete(globalMetrics.responseTimesByEndpoint, k)
	}
	
	// Reset circular buffer
	globalMetrics.responseTimeBuffer = utils.NewCircularResponseTimeBuffer(1000)
}

// CompetitorBenchmarkMiddleware adds headers comparing to competitor performance
func CompetitorBenchmarkMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		c.Next()
		
		responseTime := time.Since(startTime)
		
		// Add competitive comparison headers
		c.Header("X-MarkGo-Response-Time", responseTime.String())
		
		// Compare to competitor targets (use cached strings)
		classCache := utils.GetGlobalPerformanceClassCache()
		if responseTime.Milliseconds() < 50 {
			c.Header("X-Performance-vs-Ghost", classCache.GetClass("4x faster")) // Ghost ~200ms
			c.Header("X-Performance-vs-WordPress", classCache.GetClass("10x faster")) // WordPress ~500ms
		}
		
		if responseTime.Milliseconds() < 30 {
			c.Header("X-Performance-vs-Hugo", classCache.GetClass("comparable")) // Hugo ~10ms (static)
		}
		
		// Performance classification (use cached strings)
		if responseTime.Milliseconds() < 10 {
			c.Header("X-Performance-Class", classCache.GetClass("exceptional"))
		} else if responseTime.Milliseconds() < 50 {
			c.Header("X-Performance-Class", classCache.GetClass("excellent"))
		} else if responseTime.Milliseconds() < 100 {
			c.Header("X-Performance-Class", classCache.GetClass("good"))
		} else {
			c.Header("X-Performance-Class", classCache.GetClass("needs-optimization"))
		}
	}
}