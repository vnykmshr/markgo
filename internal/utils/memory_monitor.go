package utils

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

// MemoryThresholds defines memory usage thresholds for leak detection
type MemoryThresholds struct {
	HeapAllocMB     float64 // Heap allocation threshold in MB
	SysMemoryMB     float64 // System memory threshold in MB
	GrowthRatioMax  float64 // Maximum acceptable growth ratio
	SampleIntervalS int     // Sample interval in seconds
	AlertIntervalS  int     // Alert cooldown in seconds
}

// DefaultMemoryThresholds returns sensible default thresholds
func DefaultMemoryThresholds() MemoryThresholds {
	return MemoryThresholds{
		HeapAllocMB:     50.0,  // 50MB heap allocation limit
		SysMemoryMB:     100.0, // 100MB system memory limit
		GrowthRatioMax:  2.0,   // 2x growth max between samples
		SampleIntervalS: 60,    // Sample every minute
		AlertIntervalS:  300,   // Alert at most every 5 minutes
	}
}

// MemoryStats holds memory statistics for monitoring
type MemoryStats struct {
	Timestamp    time.Time
	HeapAlloc    uint64
	HeapSys      uint64
	NumGC        uint32
	NumGoroutine int
}

// MemoryMonitor monitors memory usage and detects potential leaks
type MemoryMonitor struct {
	logger      *slog.Logger
	thresholds  MemoryThresholds
	samples     []MemoryStats
	maxSamples  int
	lastAlert   time.Time
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	cleanupFunc func() // Optional cleanup function to call on detection
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(logger *slog.Logger, thresholds MemoryThresholds) *MemoryMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &MemoryMonitor{
		logger:     logger,
		thresholds: thresholds,
		samples:    make([]MemoryStats, 0, 60), // Keep 1 hour of samples
		maxSamples: 60,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// SetCleanupFunc sets a cleanup function to call when memory leak is detected
func (m *MemoryMonitor) SetCleanupFunc(cleanup func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupFunc = cleanup
}

// Start begins memory monitoring
func (m *MemoryMonitor) Start() {
	go m.monitorLoop()
}

// Stop stops memory monitoring
func (m *MemoryMonitor) Stop() {
	m.cancel()
}

// GetCurrentStats returns current memory statistics
func (m *MemoryMonitor) GetCurrentStats() MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return MemoryStats{
		Timestamp:    time.Now(),
		HeapAlloc:    memStats.Alloc,
		HeapSys:      memStats.HeapSys,
		NumGC:        memStats.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
	}
}

// GetRecentSamples returns recent memory samples (thread-safe)
func (m *MemoryMonitor) GetRecentSamples(count int) []MemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if count <= 0 || count > len(m.samples) {
		count = len(m.samples)
	}

	result := make([]MemoryStats, count)
	copy(result, m.samples[len(m.samples)-count:])
	return result
}

// monitorLoop runs the main monitoring loop
func (m *MemoryMonitor) monitorLoop() {
	ticker := time.NewTicker(time.Duration(m.thresholds.SampleIntervalS) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectSample()
		}
	}
}

// collectSample collects a memory sample and checks for leaks
func (m *MemoryMonitor) collectSample() {
	stats := m.GetCurrentStats()

	m.mu.Lock()
	// Add sample to ring buffer
	if len(m.samples) >= m.maxSamples {
		// Shift samples left, drop oldest
		copy(m.samples, m.samples[1:])
		m.samples[len(m.samples)-1] = stats
	} else {
		m.samples = append(m.samples, stats)
	}

	samples := make([]MemoryStats, len(m.samples))
	copy(samples, m.samples)
	cleanupFunc := m.cleanupFunc
	m.mu.Unlock()

	// Check for memory leaks (outside of lock to avoid blocking)
	if len(samples) >= 2 {
		m.checkMemoryLeaks(samples, cleanupFunc)
	}
}

// checkMemoryLeaks analyzes samples for potential memory leaks
func (m *MemoryMonitor) checkMemoryLeaks(samples []MemoryStats, cleanupFunc func()) {
	current := samples[len(samples)-1]
	previous := samples[len(samples)-2]

	// Convert to MB for easier analysis
	currentHeapMB := float64(current.HeapAlloc) / 1024 / 1024
	currentSysMB := float64(current.HeapSys) / 1024 / 1024

	// Check absolute thresholds
	heapExceeded := currentHeapMB > m.thresholds.HeapAllocMB
	sysExceeded := currentSysMB > m.thresholds.SysMemoryMB

	// Check growth ratio
	growthRatio := float64(current.HeapAlloc) / float64(previous.HeapAlloc)
	growthExceeded := growthRatio > m.thresholds.GrowthRatioMax

	// Check if we should alert (cooldown period)
	shouldAlert := time.Since(m.lastAlert) > time.Duration(m.thresholds.AlertIntervalS)*time.Second

	if (heapExceeded || sysExceeded || growthExceeded) && shouldAlert {
		m.lastAlert = time.Now()

		m.logger.Warn("Memory leak detected",
			"heap_alloc_mb", currentHeapMB,
			"heap_sys_mb", currentSysMB,
			"growth_ratio", growthRatio,
			"num_goroutines", current.NumGoroutine,
			"heap_exceeded", heapExceeded,
			"sys_exceeded", sysExceeded,
			"growth_exceeded", growthExceeded,
		)

		// Trigger automatic cleanup if available
		if cleanupFunc != nil {
			m.logger.Info("Triggering automatic cleanup")
			cleanupFunc()
		}

		// Force garbage collection
		runtime.GC()

		// Check if cleanup was effective
		go m.verifyCleanupEffectiveness(currentHeapMB)
	}
}

// verifyCleanupEffectiveness checks if cleanup reduced memory usage
func (m *MemoryMonitor) verifyCleanupEffectiveness(beforeCleanupMB float64) {
	// Wait a bit for cleanup to take effect
	time.Sleep(5 * time.Second)

	afterStats := m.GetCurrentStats()
	afterCleanupMB := float64(afterStats.HeapAlloc) / 1024 / 1024

	reduction := beforeCleanupMB - afterCleanupMB
	reductionPercent := (reduction / beforeCleanupMB) * 100

	if reductionPercent > 10 {
		m.logger.Info("Cleanup effective",
			"before_mb", beforeCleanupMB,
			"after_mb", afterCleanupMB,
			"reduction_mb", reduction,
			"reduction_percent", reductionPercent,
		)
	} else {
		m.logger.Warn("Cleanup ineffective - possible persistent leak",
			"before_mb", beforeCleanupMB,
			"after_mb", afterCleanupMB,
			"reduction_mb", reduction,
			"reduction_percent", reductionPercent,
		)
	}
}

// GetMemoryReport generates a comprehensive memory report
func (m *MemoryMonitor) GetMemoryReport() map[string]interface{} {
	m.mu.RLock()
	samples := make([]MemoryStats, len(m.samples))
	copy(samples, m.samples)
	m.mu.RUnlock()

	if len(samples) == 0 {
		return map[string]interface{}{"error": "no samples available"}
	}

	current := samples[len(samples)-1]

	report := map[string]interface{}{
		"current": map[string]interface{}{
			"heap_alloc_mb":  float64(current.HeapAlloc) / 1024 / 1024,
			"heap_sys_mb":    float64(current.HeapSys) / 1024 / 1024,
			"num_gc":         current.NumGC,
			"num_goroutines": current.NumGoroutine,
			"timestamp":      current.Timestamp.Format(time.RFC3339),
		},
		"thresholds": map[string]interface{}{
			"heap_alloc_mb":     m.thresholds.HeapAllocMB,
			"sys_memory_mb":     m.thresholds.SysMemoryMB,
			"growth_ratio_max":  m.thresholds.GrowthRatioMax,
			"sample_interval_s": m.thresholds.SampleIntervalS,
		},
		"samples_count": len(samples),
	}

	// Add trend analysis if we have enough samples
	if len(samples) >= 3 {
		oldest := samples[0]
		trend := map[string]interface{}{
			"duration_minutes": current.Timestamp.Sub(oldest.Timestamp).Minutes(),
			"heap_growth_mb":   (float64(current.HeapAlloc) - float64(oldest.HeapAlloc)) / 1024 / 1024,
			"sys_growth_mb":    (float64(current.HeapSys) - float64(oldest.HeapSys)) / 1024 / 1024,
			"gc_count_delta":   current.NumGC - oldest.NumGC,
		}
		report["trend"] = trend
	}

	return report
}

// GlobalMemoryMonitor singleton
var globalMemoryMonitor *MemoryMonitor
var globalMemoryMonitorOnce sync.Once

// GetGlobalMemoryMonitor returns the global memory monitor instance
func GetGlobalMemoryMonitor(logger *slog.Logger) *MemoryMonitor {
	globalMemoryMonitorOnce.Do(func() {
		globalMemoryMonitor = NewMemoryMonitor(logger, DefaultMemoryThresholds())
	})
	return globalMemoryMonitor
}

// AutoCleanupManager manages automatic cleanup of pooled resources
type AutoCleanupManager struct {
	pools []PoolCleaner
	mu    sync.RWMutex
}

// PoolCleaner interface for pool cleanup
type PoolCleaner interface {
	Cleanup()
	GetStats() map[string]interface{}
}

// NewAutoCleanupManager creates a new cleanup manager
func NewAutoCleanupManager() *AutoCleanupManager {
	return &AutoCleanupManager{
		pools: make([]PoolCleaner, 0),
	}
}

// RegisterPool registers a pool for automatic cleanup
func (m *AutoCleanupManager) RegisterPool(pool PoolCleaner) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pools = append(m.pools, pool)
}

// CleanupAll triggers cleanup for all registered pools
func (m *AutoCleanupManager) CleanupAll() {
	m.mu.RLock()
	pools := make([]PoolCleaner, len(m.pools))
	copy(pools, m.pools)
	m.mu.RUnlock()

	for _, pool := range pools {
		pool.Cleanup()
	}
}

// GetAllStats returns stats for all registered pools
func (m *AutoCleanupManager) GetAllStats() map[string]interface{} {
	m.mu.RLock()
	pools := make([]PoolCleaner, len(m.pools))
	copy(pools, m.pools)
	m.mu.RUnlock()

	stats := make(map[string]interface{})
	for i, pool := range pools {
		stats[fmt.Sprintf("pool_%d", i)] = pool.GetStats()
	}

	return stats
}

// Global cleanup manager
var globalCleanupManager = NewAutoCleanupManager()

// GetGlobalCleanupManager returns the global cleanup manager
func GetGlobalCleanupManager() *AutoCleanupManager {
	return globalCleanupManager
}
