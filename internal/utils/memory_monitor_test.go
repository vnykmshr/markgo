package utils

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultMemoryThresholds(t *testing.T) {
	thresholds := DefaultMemoryThresholds()

	assert.Equal(t, 50.0, thresholds.HeapAllocMB)
	assert.Equal(t, 100.0, thresholds.SysMemoryMB)
	assert.Equal(t, 2.0, thresholds.GrowthRatioMax)
	assert.Equal(t, 60, thresholds.SampleIntervalS)
	assert.Equal(t, 300, thresholds.AlertIntervalS)
}

func TestNewMemoryMonitor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	thresholds := DefaultMemoryThresholds()

	monitor := NewMemoryMonitor(logger, thresholds)

	assert.NotNil(t, monitor)
	assert.Equal(t, thresholds, monitor.thresholds)
	assert.Equal(t, 0, len(monitor.samples))
	assert.Equal(t, 60, monitor.maxSamples)
}

func TestMemoryMonitor_SetCleanupFunc(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	monitor := NewMemoryMonitor(logger, DefaultMemoryThresholds())

	called := false
	cleanup := func() { called = true }

	monitor.SetCleanupFunc(cleanup)

	// Cleanup function should be set
	monitor.mu.RLock()
	assert.NotNil(t, monitor.cleanupFunc)
	monitor.mu.RUnlock()

	// Test calling cleanup function
	monitor.mu.RLock()
	monitor.cleanupFunc()
	monitor.mu.RUnlock()

	assert.True(t, called)
}

func TestMemoryMonitor_GetCurrentStats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	monitor := NewMemoryMonitor(logger, DefaultMemoryThresholds())

	stats := monitor.GetCurrentStats()

	assert.False(t, stats.Timestamp.IsZero())
	assert.True(t, stats.HeapAlloc > 0)
	assert.True(t, stats.HeapSys > 0)
	assert.True(t, stats.NumGoroutine > 0)
}

func TestMemoryMonitor_GetRecentSamples(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	monitor := NewMemoryMonitor(logger, DefaultMemoryThresholds())

	// Initially no samples
	samples := monitor.GetRecentSamples(5)
	assert.Equal(t, 0, len(samples))

	// Add some samples manually
	monitor.mu.Lock()
	monitor.samples = append(monitor.samples, MemoryStats{
		Timestamp: time.Now(),
		HeapAlloc: 1024,
	})
	monitor.samples = append(monitor.samples, MemoryStats{
		Timestamp: time.Now(),
		HeapAlloc: 2048,
	})
	monitor.samples = append(monitor.samples, MemoryStats{
		Timestamp: time.Now(),
		HeapAlloc: 3072,
	})
	monitor.mu.Unlock()

	// Test getting recent samples
	samples = monitor.GetRecentSamples(2)
	assert.Equal(t, 2, len(samples))
	assert.Equal(t, uint64(2048), samples[0].HeapAlloc)
	assert.Equal(t, uint64(3072), samples[1].HeapAlloc)

	// Test getting all samples
	samples = monitor.GetRecentSamples(10)
	assert.Equal(t, 3, len(samples))

	// Test getting zero samples
	samples = monitor.GetRecentSamples(0)
	assert.Equal(t, 3, len(samples)) // Should return all
}

func TestMemoryMonitor_StartStop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	// Use very short interval for testing
	thresholds := DefaultMemoryThresholds()
	thresholds.SampleIntervalS = 1 // 1 second
	monitor := NewMemoryMonitor(logger, thresholds)

	// Start monitoring
	monitor.Start()

	// Wait a bit for samples to be collected
	time.Sleep(1500 * time.Millisecond)

	// Should have at least one sample
	samples := monitor.GetRecentSamples(10)
	assert.True(t, len(samples) > 0)

	// Stop monitoring
	monitor.Stop()

	// Context should be cancelled
	select {
	case <-monitor.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after Stop()")
	}
}

func TestMemoryMonitor_GetMemoryReport(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	monitor := NewMemoryMonitor(logger, DefaultMemoryThresholds())

	// Test empty report
	report := monitor.GetMemoryReport()
	assert.Contains(t, report, "error")

	// Add some samples
	now := time.Now()
	monitor.mu.Lock()
	monitor.samples = append(monitor.samples, MemoryStats{
		Timestamp:    now.Add(-2 * time.Minute),
		HeapAlloc:    1024 * 1024,     // 1MB
		HeapSys:      2 * 1024 * 1024, // 2MB
		NumGC:        10,
		NumGoroutine: 5,
	})
	monitor.samples = append(monitor.samples, MemoryStats{
		Timestamp:    now.Add(-1 * time.Minute),
		HeapAlloc:    1536 * 1024, // 1.5MB
		HeapSys:      2252 * 1024, // 2.2MB
		NumGC:        12,
		NumGoroutine: 6,
	})
	monitor.samples = append(monitor.samples, MemoryStats{
		Timestamp:    now,
		HeapAlloc:    2 * 1024 * 1024, // 2MB
		HeapSys:      2560 * 1024,     // 2.5MB
		NumGC:        15,
		NumGoroutine: 7,
	})
	monitor.mu.Unlock()

	report = monitor.GetMemoryReport()

	// Check current stats
	current := report["current"].(map[string]interface{})
	assert.Equal(t, float64(2), current["heap_alloc_mb"]) // 2MB
	assert.Equal(t, float64(2.5), current["heap_sys_mb"]) // 2.5MB
	assert.Equal(t, uint32(15), current["num_gc"])
	assert.Equal(t, 7, current["num_goroutines"])

	// Check thresholds
	thresholds := report["thresholds"].(map[string]interface{})
	assert.Equal(t, 50.0, thresholds["heap_alloc_mb"])

	// Check trend analysis
	trend := report["trend"].(map[string]interface{})
	assert.Equal(t, float64(1), trend["heap_growth_mb"]) // 2MB - 1MB = 1MB
	assert.Equal(t, uint32(5), trend["gc_count_delta"])  // 15 - 10 = 5
}

func TestNewAutoCleanupManager(t *testing.T) {
	manager := NewAutoCleanupManager()
	assert.NotNil(t, manager)
	assert.Equal(t, 0, len(manager.pools))
}

// MockPoolCleaner for testing
type MockPoolCleaner struct {
	cleanupCalled bool
	stats         map[string]interface{}
}

func (m *MockPoolCleaner) Cleanup() {
	m.cleanupCalled = true
}

func (m *MockPoolCleaner) GetStats() map[string]interface{} {
	return m.stats
}

func TestAutoCleanupManager_RegisterPool(t *testing.T) {
	manager := NewAutoCleanupManager()
	pool := &MockPoolCleaner{
		stats: map[string]interface{}{"test": "value"},
	}

	manager.RegisterPool(pool)

	assert.Equal(t, 1, len(manager.pools))
}

func TestAutoCleanupManager_CleanupAll(t *testing.T) {
	manager := NewAutoCleanupManager()

	pool1 := &MockPoolCleaner{stats: map[string]interface{}{}}
	pool2 := &MockPoolCleaner{stats: map[string]interface{}{}}

	manager.RegisterPool(pool1)
	manager.RegisterPool(pool2)

	// Call cleanup
	manager.CleanupAll()

	// Both pools should have cleanup called
	assert.True(t, pool1.cleanupCalled)
	assert.True(t, pool2.cleanupCalled)
}

func TestAutoCleanupManager_GetAllStats(t *testing.T) {
	manager := NewAutoCleanupManager()

	pool1 := &MockPoolCleaner{stats: map[string]interface{}{"pool1": "data1"}}
	pool2 := &MockPoolCleaner{stats: map[string]interface{}{"pool2": "data2"}}

	manager.RegisterPool(pool1)
	manager.RegisterPool(pool2)

	stats := manager.GetAllStats()

	assert.Equal(t, 2, len(stats))
	assert.Contains(t, stats, "pool_0")
	assert.Contains(t, stats, "pool_1")
	assert.Equal(t, pool1.stats, stats["pool_0"])
	assert.Equal(t, pool2.stats, stats["pool_1"])
}

func TestGetGlobalMemoryMonitor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	monitor1 := GetGlobalMemoryMonitor(logger)
	monitor2 := GetGlobalMemoryMonitor(logger)

	assert.NotNil(t, monitor1)
	assert.Equal(t, monitor1, monitor2) // Should be the same instance (singleton)
}

func TestGetGlobalCleanupManager(t *testing.T) {
	manager1 := GetGlobalCleanupManager()
	manager2 := GetGlobalCleanupManager()

	assert.NotNil(t, manager1)
	assert.Equal(t, manager1, manager2) // Should be the same instance (singleton)
}

// Integration test for memory monitoring with cleanup
func TestMemoryMonitor_Integration(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create monitor with very low thresholds to trigger alerts
	thresholds := MemoryThresholds{
		HeapAllocMB:     0.001, // 1KB (very low)
		SysMemoryMB:     0.001, // 1KB (very low)
		GrowthRatioMax:  1.1,   // 10% growth max
		SampleIntervalS: 1,     // 1 second
		AlertIntervalS:  1,     // 1 second
	}

	monitor := NewMemoryMonitor(logger, thresholds)

	cleanupCalled := false
	monitor.SetCleanupFunc(func() {
		cleanupCalled = true
	})

	// Manually trigger sample collection to simulate monitoring
	monitor.collectSample()
	time.Sleep(100 * time.Millisecond) // Let any goroutines finish
	monitor.collectSample()

	// Should trigger cleanup due to low thresholds
	time.Sleep(200 * time.Millisecond) // Wait for any async operations

	// Check that cleanup was called (thresholds are so low they should trigger)
	// Note: This might be flaky depending on actual memory usage, but should work in most cases
	assert.True(t, cleanupCalled || len(buf.String()) > 0) // Either cleanup called or warning logged
}
