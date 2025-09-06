package utils

import (
	"context"
	"log/slog"
	"runtime"
	"runtime/debug"
	"sort"
	"sync"
	"time"
)

// MemoryProfiler provides detailed memory usage analysis and optimization recommendations
type MemoryProfiler struct {
	logger       *slog.Logger
	samples      []ProfileSample
	gcStats      []GCStats
	allocStats   map[string]*AllocationStats
	mu           sync.RWMutex
	maxSamples   int
	startTime    time.Time
	sampleTicker *time.Ticker
	ctx          context.Context
	cancel       context.CancelFunc
}

// ProfileSample represents a memory usage snapshot
type ProfileSample struct {
	Timestamp    time.Time `json:"timestamp"`
	HeapAlloc    uint64    `json:"heap_alloc"`
	HeapSys      uint64    `json:"heap_sys"`
	HeapIdle     uint64    `json:"heap_idle"`
	HeapInuse    uint64    `json:"heap_inuse"`
	HeapReleased uint64    `json:"heap_released"`
	HeapObjects  uint64    `json:"heap_objects"`
	StackInuse   uint64    `json:"stack_inuse"`
	StackSys     uint64    `json:"stack_sys"`
	MSpanInuse   uint64    `json:"mspan_inuse"`
	MSpanSys     uint64    `json:"mspan_sys"`
	MCacheInuse  uint64    `json:"mcache_inuse"`
	MCacheSys    uint64    `json:"mcache_sys"`
	NextGC       uint64    `json:"next_gc"`
	NumGC        uint32    `json:"num_gc"`
	NumGoroutine int       `json:"num_goroutine"`
	CPUCount     int       `json:"cpu_count"`
	GoVersion    string    `json:"go_version"`
}

// GCStats represents garbage collection statistics
type GCStats struct {
	Timestamp     time.Time     `json:"timestamp"`
	NumGC         uint32        `json:"num_gc"`
	PauseNs       []uint64      `json:"pause_ns"`
	PauseEnd      []uint64      `json:"pause_end"`
	PauseTotal    time.Duration `json:"pause_total"`
	GCCPUFraction float64       `json:"gc_cpu_fraction"`
}

// AllocationStats tracks allocation patterns by type/location
type AllocationStats struct {
	Type       string    `json:"type"`
	Count      uint64    `json:"count"`
	TotalBytes uint64    `json:"total_bytes"`
	AvgSize    uint64    `json:"avg_size"`
	PeakBytes  uint64    `json:"peak_bytes"`
	LastSeen   time.Time `json:"last_seen"`
	GrowthRate float64   `json:"growth_rate"`
}

// MemoryReport contains comprehensive memory analysis
type MemoryReport struct {
	GeneratedAt        time.Time                    `json:"generated_at"`
	UptimeDuration     time.Duration                `json:"uptime_duration"`
	CurrentStats       ProfileSample                `json:"current_stats"`
	PeakStats          ProfileSample                `json:"peak_stats"`
	AverageStats       ProfileSample                `json:"average_stats"`
	GCEfficiency       GCEfficiencyMetrics          `json:"gc_efficiency"`
	AllocationHotspots []*AllocationStats           `json:"allocation_hotspots"`
	Recommendations    []OptimizationRecommendation `json:"recommendations"`
	Trends             TrendAnalysis                `json:"trends"`
}

// GCEfficiencyMetrics provides GC performance analysis
type GCEfficiencyMetrics struct {
	AvgPauseTime    time.Duration `json:"avg_pause_time"`
	MaxPauseTime    time.Duration `json:"max_pause_time"`
	GCFrequency     float64       `json:"gc_frequency_per_min"`
	GCCPUFraction   float64       `json:"gc_cpu_fraction"`
	EfficiencyScore float64       `json:"efficiency_score"`
}

// OptimizationRecommendation suggests specific memory optimizations
type OptimizationRecommendation struct {
	Priority   string `json:"priority"` // "critical", "high", "medium", "low"
	Category   string `json:"category"` // "allocation", "gc", "leak", "pooling"
	Issue      string `json:"issue"`
	Suggestion string `json:"suggestion"`
	Impact     string `json:"impact"`
	Effort     string `json:"effort"`
}

// TrendAnalysis analyzes memory usage trends over time
type TrendAnalysis struct {
	HeapTrend      string   `json:"heap_trend"`      // "increasing", "decreasing", "stable"
	GoroutineTrend string   `json:"goroutine_trend"` // "increasing", "decreasing", "stable"
	GCTrend        string   `json:"gc_trend"`        // "more_frequent", "less_frequent", "stable"
	LeakIndicators []string `json:"leak_indicators"`
	EfficiencyTips []string `json:"efficiency_tips"`
}

// NewMemoryProfiler creates a new memory profiler with monitoring
func NewMemoryProfiler(logger *slog.Logger, sampleInterval time.Duration) *MemoryProfiler {
	ctx, cancel := context.WithCancel(context.Background())

	mp := &MemoryProfiler{
		logger:     logger,
		samples:    make([]ProfileSample, 0, 1000),
		gcStats:    make([]GCStats, 0, 1000),
		allocStats: make(map[string]*AllocationStats),
		maxSamples: 1000,
		startTime:  time.Now(),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start sampling if interval is provided
	if sampleInterval > 0 {
		mp.sampleTicker = time.NewTicker(sampleInterval)
		go mp.startSampling()
	}

	return mp
}

// startSampling begins periodic memory sampling
func (mp *MemoryProfiler) startSampling() {
	defer mp.sampleTicker.Stop()

	for {
		select {
		case <-mp.ctx.Done():
			mp.logger.Debug("Memory profiler sampling stopped")
			return
		case <-mp.sampleTicker.C:
			mp.TakeSample()
		}
	}
}

// TakeSample captures a memory usage snapshot
func (mp *MemoryProfiler) TakeSample() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	sample := ProfileSample{
		Timestamp:    time.Now(),
		HeapAlloc:    memStats.Alloc,
		HeapSys:      memStats.HeapSys,
		HeapIdle:     memStats.HeapIdle,
		HeapInuse:    memStats.HeapInuse,
		HeapReleased: memStats.HeapReleased,
		HeapObjects:  memStats.HeapObjects,
		StackInuse:   memStats.StackInuse,
		StackSys:     memStats.StackSys,
		MSpanInuse:   memStats.MSpanInuse,
		MSpanSys:     memStats.MSpanSys,
		MCacheInuse:  memStats.MCacheInuse,
		MCacheSys:    memStats.MCacheSys,
		NextGC:       memStats.NextGC,
		NumGC:        memStats.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
		CPUCount:     runtime.NumCPU(),
		GoVersion:    runtime.Version(),
	}

	// Add sample and maintain sliding window
	mp.samples = append(mp.samples, sample)
	if len(mp.samples) > mp.maxSamples {
		mp.samples = mp.samples[1:]
	}

	// Collect GC stats
	mp.collectGCStats(&memStats)

	// Log significant changes
	if len(mp.samples) > 1 {
		mp.detectSignificantChanges(sample)
	}
}

// collectGCStats collects garbage collection statistics
func (mp *MemoryProfiler) collectGCStats(memStats *runtime.MemStats) {
	gcStats := GCStats{
		Timestamp:     time.Now(),
		NumGC:         memStats.NumGC,
		PauseNs:       make([]uint64, len(memStats.PauseNs)),
		PauseEnd:      make([]uint64, len(memStats.PauseEnd)),
		PauseTotal:    time.Duration(memStats.PauseTotalNs),
		GCCPUFraction: memStats.GCCPUFraction,
	}

	copy(gcStats.PauseNs, memStats.PauseNs[:])
	copy(gcStats.PauseEnd, memStats.PauseEnd[:])

	mp.gcStats = append(mp.gcStats, gcStats)
	if len(mp.gcStats) > mp.maxSamples {
		mp.gcStats = mp.gcStats[1:]
	}
}

// detectSignificantChanges logs notable memory usage changes
func (mp *MemoryProfiler) detectSignificantChanges(current ProfileSample) {
	prev := mp.samples[len(mp.samples)-2]

	heapGrowth := float64(current.HeapAlloc) / float64(prev.HeapAlloc)
	goroutineGrowth := float64(current.NumGoroutine) / float64(prev.NumGoroutine)

	if heapGrowth > 1.5 {
		mp.logger.Warn("Significant heap allocation increase",
			"previous_mb", prev.HeapAlloc/(1024*1024),
			"current_mb", current.HeapAlloc/(1024*1024),
			"growth_ratio", heapGrowth)
	}

	if goroutineGrowth > 1.3 {
		mp.logger.Warn("Significant goroutine count increase",
			"previous", prev.NumGoroutine,
			"current", current.NumGoroutine,
			"growth_ratio", goroutineGrowth)
	}
}

// GenerateReport creates a comprehensive memory analysis report
func (mp *MemoryProfiler) GenerateReport() *MemoryReport {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if len(mp.samples) == 0 {
		mp.TakeSample() // Ensure we have at least one sample
	}

	current := mp.samples[len(mp.samples)-1]
	peak := mp.findPeakUsage()
	average := mp.calculateAverageStats()
	gcEfficiency := mp.analyzeGCEfficiency()
	trends := mp.analyzeTrends()
	recommendations := mp.generateRecommendations(trends)

	return &MemoryReport{
		GeneratedAt:        time.Now(),
		UptimeDuration:     time.Since(mp.startTime),
		CurrentStats:       current,
		PeakStats:          peak,
		AverageStats:       average,
		GCEfficiency:       gcEfficiency,
		AllocationHotspots: mp.getAllocationHotspots(),
		Recommendations:    recommendations,
		Trends:             trends,
	}
}

// findPeakUsage finds the sample with highest memory usage
func (mp *MemoryProfiler) findPeakUsage() ProfileSample {
	var peak ProfileSample
	for _, sample := range mp.samples {
		if sample.HeapAlloc > peak.HeapAlloc {
			peak = sample
		}
	}
	return peak
}

// calculateAverageStats calculates average memory statistics
func (mp *MemoryProfiler) calculateAverageStats() ProfileSample {
	if len(mp.samples) == 0 {
		return ProfileSample{}
	}

	var avg ProfileSample
	for _, sample := range mp.samples {
		avg.HeapAlloc += sample.HeapAlloc
		avg.HeapSys += sample.HeapSys
		avg.HeapInuse += sample.HeapInuse
		avg.HeapObjects += sample.HeapObjects
		avg.NumGoroutine += sample.NumGoroutine
	}

	count := uint64(len(mp.samples))
	avg.HeapAlloc /= count
	avg.HeapSys /= count
	avg.HeapInuse /= count
	avg.HeapObjects /= count
	avg.NumGoroutine = int(uint64(avg.NumGoroutine) / count)
	avg.Timestamp = time.Now()

	return avg
}

// analyzeGCEfficiency analyzes garbage collection performance
func (mp *MemoryProfiler) analyzeGCEfficiency() GCEfficiencyMetrics {
	if len(mp.gcStats) == 0 {
		return GCEfficiencyMetrics{}
	}

	latest := mp.gcStats[len(mp.gcStats)-1]

	var totalPause uint64
	var maxPause uint64
	var pauseCount int

	for _, pause := range latest.PauseNs {
		if pause > 0 {
			totalPause += pause
			pauseCount++
			if pause > maxPause {
				maxPause = pause
			}
		}
	}

	var avgPause time.Duration
	if pauseCount > 0 {
		avgPause = time.Duration(totalPause / uint64(pauseCount))
	}

	gcFrequency := float64(latest.NumGC) / time.Since(mp.startTime).Minutes()

	// Calculate efficiency score (0-100, higher is better)
	efficiencyScore := mp.calculateEfficiencyScore(avgPause, latest.GCCPUFraction, gcFrequency)

	return GCEfficiencyMetrics{
		AvgPauseTime:    avgPause,
		MaxPauseTime:    time.Duration(maxPause),
		GCFrequency:     gcFrequency,
		GCCPUFraction:   latest.GCCPUFraction,
		EfficiencyScore: efficiencyScore,
	}
}

// calculateEfficiencyScore computes GC efficiency score
func (mp *MemoryProfiler) calculateEfficiencyScore(avgPause time.Duration, cpuFraction, frequency float64) float64 {
	// Ideal values: <1ms pause, <2% CPU, 1-10 GCs per minute
	pauseScore := 100.0
	if avgPause > time.Millisecond {
		pauseScore = 100.0 - (float64(avgPause)/float64(time.Millisecond))*10
	}

	cpuScore := 100.0 - (cpuFraction * 100 * 50) // Penalize high CPU usage

	freqScore := 100.0
	if frequency > 10 {
		freqScore = 100.0 - (frequency-10)*5 // Penalize excessive GC frequency
	}

	// Weight the scores
	return (pauseScore*0.4 + cpuScore*0.4 + freqScore*0.2)
}

// analyzeTrends analyzes memory usage trends over time
func (mp *MemoryProfiler) analyzeTrends() TrendAnalysis {
	if len(mp.samples) < 10 {
		return TrendAnalysis{
			HeapTrend:      "insufficient_data",
			GoroutineTrend: "insufficient_data",
			GCTrend:        "insufficient_data",
		}
	}

	// Analyze heap trend
	heapTrend := mp.analyzeTrend("heap", func(s ProfileSample) float64 { return float64(s.HeapAlloc) })
	goroutineTrend := mp.analyzeTrend("goroutine", func(s ProfileSample) float64 { return float64(s.NumGoroutine) })
	gcTrend := mp.analyzeGCTrend()

	leakIndicators := mp.detectLeakIndicators()
	efficiencyTips := mp.generateEfficiencyTips()

	return TrendAnalysis{
		HeapTrend:      heapTrend,
		GoroutineTrend: goroutineTrend,
		GCTrend:        gcTrend,
		LeakIndicators: leakIndicators,
		EfficiencyTips: efficiencyTips,
	}
}

// analyzeTrend analyzes trend for a specific metric
func (mp *MemoryProfiler) analyzeTrend(name string, extractor func(ProfileSample) float64) string {
	if len(mp.samples) < 10 {
		return "insufficient_data"
	}

	// Take recent samples (last 50% of data)
	start := len(mp.samples) / 2
	recent := mp.samples[start:]

	// Calculate linear regression slope
	slope := mp.calculateSlope(recent, extractor)

	threshold := 0.01 // 1% growth per sample
	if slope > threshold {
		return "increasing"
	} else if slope < -threshold {
		return "decreasing"
	}
	return "stable"
}

// calculateSlope calculates the slope of a trend line
func (mp *MemoryProfiler) calculateSlope(samples []ProfileSample, extractor func(ProfileSample) float64) float64 {
	n := float64(len(samples))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64

	for i, sample := range samples {
		x := float64(i)
		y := extractor(sample)

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	return (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
}

// analyzeGCTrend analyzes garbage collection frequency trend
func (mp *MemoryProfiler) analyzeGCTrend() string {
	if len(mp.gcStats) < 5 {
		return "insufficient_data"
	}

	recent := mp.gcStats[len(mp.gcStats)-5:]
	oldest := recent[0]
	newest := recent[len(recent)-1]

	timeDiff := newest.Timestamp.Sub(oldest.Timestamp)
	gcDiff := newest.NumGC - oldest.NumGC

	frequency := float64(gcDiff) / timeDiff.Minutes()

	if frequency > 2.0 {
		return "more_frequent"
	} else if frequency < 0.5 {
		return "less_frequent"
	}
	return "stable"
}

// detectLeakIndicators identifies potential memory leak patterns
func (mp *MemoryProfiler) detectLeakIndicators() []string {
	indicators := make([]string, 0)

	if len(mp.samples) < 20 {
		return indicators
	}

	current := mp.samples[len(mp.samples)-1]
	older := mp.samples[len(mp.samples)-20]

	// Check for consistent growth
	heapGrowth := float64(current.HeapAlloc) / float64(older.HeapAlloc)
	if heapGrowth > 2.0 {
		indicators = append(indicators, "Consistent heap growth over time")
	}

	goroutineGrowth := float64(current.NumGoroutine) / float64(older.NumGoroutine)
	if goroutineGrowth > 1.5 {
		indicators = append(indicators, "Growing goroutine count")
	}

	// Check for high heap utilization
	heapUtilization := float64(current.HeapInuse) / float64(current.HeapSys)
	if heapUtilization > 0.9 {
		indicators = append(indicators, "High heap utilization (>90%)")
	}

	return indicators
}

// generateEfficiencyTips provides memory efficiency recommendations
func (mp *MemoryProfiler) generateEfficiencyTips() []string {
	tips := make([]string, 0)

	if len(mp.samples) == 0 {
		return tips
	}

	current := mp.samples[len(mp.samples)-1]

	// General tips based on current state
	if current.HeapObjects > 1000000 {
		tips = append(tips, "High object count - consider object pooling")
	}

	if current.NumGoroutine > 10000 {
		tips = append(tips, "High goroutine count - check for goroutine leaks")
	}

	gcEfficiency := mp.analyzeGCEfficiency()
	if gcEfficiency.GCCPUFraction > 0.05 {
		tips = append(tips, "High GC CPU usage - consider GOGC tuning")
	}

	if gcEfficiency.AvgPauseTime > 5*time.Millisecond {
		tips = append(tips, "High GC pause times - consider memory optimization")
	}

	return tips
}

// getAllocationHotspots returns top allocation sources
func (mp *MemoryProfiler) getAllocationHotspots() []*AllocationStats {
	hotspots := make([]*AllocationStats, 0, len(mp.allocStats))

	for _, stats := range mp.allocStats {
		hotspots = append(hotspots, stats)
	}

	// Sort by total bytes allocated
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].TotalBytes > hotspots[j].TotalBytes
	})

	// Return top 10
	if len(hotspots) > 10 {
		hotspots = hotspots[:10]
	}

	return hotspots
}

// generateRecommendations creates optimization recommendations
func (mp *MemoryProfiler) generateRecommendations(trends TrendAnalysis) []OptimizationRecommendation {
	recommendations := make([]OptimizationRecommendation, 0)

	// Based on trends
	if trends.HeapTrend == "increasing" {
		recommendations = append(recommendations, OptimizationRecommendation{
			Priority:   "high",
			Category:   "allocation",
			Issue:      "Heap memory consistently increasing",
			Suggestion: "Implement object pooling for frequently allocated objects",
			Impact:     "Reduce memory allocation and GC pressure",
			Effort:     "medium",
		})
	}

	if trends.GoroutineTrend == "increasing" {
		recommendations = append(recommendations, OptimizationRecommendation{
			Priority:   "critical",
			Category:   "leak",
			Issue:      "Goroutine count consistently increasing",
			Suggestion: "Audit goroutine lifecycle and ensure proper cleanup",
			Impact:     "Prevent goroutine leaks and memory exhaustion",
			Effort:     "high",
		})
	}

	// Based on efficiency metrics
	gcEfficiency := mp.analyzeGCEfficiency()
	if gcEfficiency.EfficiencyScore < 70 {
		recommendations = append(recommendations, OptimizationRecommendation{
			Priority:   "high",
			Category:   "gc",
			Issue:      "Poor garbage collection efficiency",
			Suggestion: "Tune GOGC environment variable and reduce allocation rate",
			Impact:     "Improve application responsiveness and CPU usage",
			Effort:     "medium",
		})
	}

	return recommendations
}

// Stop stops the memory profiler sampling
func (mp *MemoryProfiler) Stop() {
	mp.cancel()
}

// ForceGC triggers garbage collection and waits for completion
func (mp *MemoryProfiler) ForceGC() {
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	runtime.GC()
	debug.FreeOSMemory()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	mp.logger.Info("Manual garbage collection completed",
		"heap_before_mb", before.Alloc/(1024*1024),
		"heap_after_mb", after.Alloc/(1024*1024),
		"freed_mb", (before.Alloc-after.Alloc)/(1024*1024))
}

// GetCurrentStats returns current memory statistics
func (mp *MemoryProfiler) GetCurrentStats() map[string]interface{} {
	mp.TakeSample()
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if len(mp.samples) == 0 {
		return nil
	}

	current := mp.samples[len(mp.samples)-1]
	return map[string]interface{}{
		"heap_alloc_mb":   current.HeapAlloc / (1024 * 1024),
		"heap_sys_mb":     current.HeapSys / (1024 * 1024),
		"heap_objects":    current.HeapObjects,
		"num_goroutine":   current.NumGoroutine,
		"num_gc":          current.NumGC,
		"gc_cpu_fraction": mp.gcStats[len(mp.gcStats)-1].GCCPUFraction,
		"uptime_minutes":  time.Since(mp.startTime).Minutes(),
	}
}
