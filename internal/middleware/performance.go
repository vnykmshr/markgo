package middleware

import (
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMetrics holds performance monitoring data
type PerformanceMetrics struct {
	mu                    sync.RWMutex
	RequestCount          int64
	TotalResponseTime     time.Duration
	AverageResponseTime   time.Duration
	MaxResponseTime       time.Duration
	MinResponseTime       time.Duration
	ResponseTimes         []time.Duration
	RequestsPerSecond     float64
	MemoryUsage          uint64
	GoroutineCount       int
	LastUpdateTime       time.Time
	RequestsByEndpoint   map[string]int64
	ResponseTimesByEndpoint map[string]time.Duration
}

var (
	globalMetrics = &PerformanceMetrics{
		MinResponseTime:        time.Hour, // Initialize to high value
		RequestsByEndpoint:     make(map[string]int64),
		ResponseTimesByEndpoint: make(map[string]time.Duration),
		LastUpdateTime:         time.Now(),
	}
)

// PerformanceMiddleware creates a middleware that tracks performance metrics
func PerformanceMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Record memory before request
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		// Process request
		c.Next()

		// Calculate response time
		responseTime := time.Since(startTime)
		endpoint := c.Request.Method + " " + c.FullPath()

		// Update metrics
		updateMetrics(responseTime, endpoint)

		// Record memory after request
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		// Add performance headers
		c.Header("X-Response-Time", responseTime.String())
		c.Header("X-Memory-Delta", strconv.FormatUint(memAfter.Alloc-memBefore.Alloc, 10))
		c.Header("X-Goroutines", strconv.Itoa(runtime.NumGoroutine()))

		// Log performance metrics for slow requests (>100ms)
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

		// Log detailed performance metrics every 100 requests
		globalMetrics.mu.RLock()
		count := globalMetrics.RequestCount
		globalMetrics.mu.RUnlock()

		if count%100 == 0 {
			logPerformanceSummary(logger)
		}
	}
}

func updateMetrics(responseTime time.Duration, endpoint string) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.RequestCount++
	globalMetrics.TotalResponseTime += responseTime

	// Update average
	globalMetrics.AverageResponseTime = globalMetrics.TotalResponseTime / time.Duration(globalMetrics.RequestCount)

	// Update min/max
	if responseTime > globalMetrics.MaxResponseTime {
		globalMetrics.MaxResponseTime = responseTime
	}
	if responseTime < globalMetrics.MinResponseTime {
		globalMetrics.MinResponseTime = responseTime
	}

	// Store recent response times for percentile calculations (keep last 1000)
	globalMetrics.ResponseTimes = append(globalMetrics.ResponseTimes, responseTime)
	if len(globalMetrics.ResponseTimes) > 1000 {
		globalMetrics.ResponseTimes = globalMetrics.ResponseTimes[1:]
	}

	// Update requests per second
	now := time.Now()
	if elapsed := now.Sub(globalMetrics.LastUpdateTime); elapsed > 0 {
		globalMetrics.RequestsPerSecond = float64(globalMetrics.RequestCount) / elapsed.Seconds()
	}

	// Update per-endpoint metrics
	globalMetrics.RequestsByEndpoint[endpoint]++
	if existingTime, exists := globalMetrics.ResponseTimesByEndpoint[endpoint]; exists {
		globalMetrics.ResponseTimesByEndpoint[endpoint] = (existingTime + responseTime) / 2
	} else {
		globalMetrics.ResponseTimesByEndpoint[endpoint] = responseTime
	}

	// Update memory usage
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	globalMetrics.MemoryUsage = mem.Alloc
	globalMetrics.GoroutineCount = runtime.NumGoroutine()
}

func logPerformanceSummary(logger *slog.Logger) {
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()

	// Calculate 95th percentile
	p95 := calculatePercentile(globalMetrics.ResponseTimes, 0.95)
	p99 := calculatePercentile(globalMetrics.ResponseTimes, 0.99)

	logger.Info("Performance Summary",
		"total_requests", globalMetrics.RequestCount,
		"avg_response_time", globalMetrics.AverageResponseTime.String(),
		"min_response_time", globalMetrics.MinResponseTime.String(),
		"max_response_time", globalMetrics.MaxResponseTime.String(),
		"p95_response_time", p95.String(),
		"p99_response_time", p99.String(),
		"requests_per_second", globalMetrics.RequestsPerSecond,
		"memory_usage_mb", globalMetrics.MemoryUsage/1024/1024,
		"goroutine_count", globalMetrics.GoroutineCount,
	)
}

func calculatePercentile(times []time.Duration, percentile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}

	// Simple percentile calculation (not optimized, but good enough for monitoring)
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)

	// Basic bubble sort for small datasets
	n := len(sorted)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	index := int(percentile * float64(len(sorted)-1))
	return sorted[index]
}

// GetPerformanceMetrics returns a copy of current performance metrics
func GetPerformanceMetrics() PerformanceMetrics {
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()

	// Create a deep copy
	metrics := PerformanceMetrics{
		RequestCount:        globalMetrics.RequestCount,
		TotalResponseTime:   globalMetrics.TotalResponseTime,
		AverageResponseTime: globalMetrics.AverageResponseTime,
		MaxResponseTime:     globalMetrics.MaxResponseTime,
		MinResponseTime:     globalMetrics.MinResponseTime,
		RequestsPerSecond:   globalMetrics.RequestsPerSecond,
		MemoryUsage:         globalMetrics.MemoryUsage,
		GoroutineCount:      globalMetrics.GoroutineCount,
		LastUpdateTime:      globalMetrics.LastUpdateTime,
		RequestsByEndpoint:  make(map[string]int64),
		ResponseTimesByEndpoint: make(map[string]time.Duration),
	}

	// Copy maps
	for k, v := range globalMetrics.RequestsByEndpoint {
		metrics.RequestsByEndpoint[k] = v
	}
	for k, v := range globalMetrics.ResponseTimesByEndpoint {
		metrics.ResponseTimesByEndpoint[k] = v
	}

	// Copy response times slice
	metrics.ResponseTimes = make([]time.Duration, len(globalMetrics.ResponseTimes))
	copy(metrics.ResponseTimes, globalMetrics.ResponseTimes)

	return metrics
}

// ResetPerformanceMetrics resets all performance metrics (useful for testing)
func ResetPerformanceMetrics() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.RequestCount = 0
	globalMetrics.TotalResponseTime = 0
	globalMetrics.AverageResponseTime = 0
	globalMetrics.MaxResponseTime = 0
	globalMetrics.MinResponseTime = time.Hour
	globalMetrics.ResponseTimes = nil
	globalMetrics.RequestsPerSecond = 0
	globalMetrics.MemoryUsage = 0
	globalMetrics.GoroutineCount = 0
	globalMetrics.LastUpdateTime = time.Now()
	globalMetrics.RequestsByEndpoint = make(map[string]int64)
	globalMetrics.ResponseTimesByEndpoint = make(map[string]time.Duration)
}

// CompetitorBenchmarkMiddleware adds headers comparing to competitor performance
func CompetitorBenchmarkMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		c.Next()
		
		responseTime := time.Since(startTime)
		
		// Add competitive comparison headers
		c.Header("X-MarkGo-Response-Time", responseTime.String())
		
		// Compare to competitor targets
		if responseTime.Milliseconds() < 50 {
			c.Header("X-Performance-vs-Ghost", "4x faster") // Ghost ~200ms
			c.Header("X-Performance-vs-WordPress", "10x faster") // WordPress ~500ms
		}
		
		if responseTime.Milliseconds() < 30 {
			c.Header("X-Performance-vs-Hugo", "comparable") // Hugo ~10ms (static)
		}
		
		// Performance classification
		if responseTime.Milliseconds() < 10 {
			c.Header("X-Performance-Class", "exceptional")
		} else if responseTime.Milliseconds() < 50 {
			c.Header("X-Performance-Class", "excellent")
		} else if responseTime.Milliseconds() < 100 {
			c.Header("X-Performance-Class", "good")
		} else {
			c.Header("X-Performance-Class", "needs-optimization")
		}
	}
}