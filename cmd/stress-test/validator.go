package main

import (
	"fmt"
	"strings"
	"time"
)

// PerformanceTarget represents a performance validation target
type PerformanceTarget struct {
	Name        string
	Target      string
	Actual      string
	Description string
	Passed      bool
}

// PerformanceValidator validates test results against performance targets
type PerformanceValidator struct {
	targets []PerformanceTarget
}

// NewPerformanceValidator creates a new performance validator
func NewPerformanceValidator() *PerformanceValidator {
	return &PerformanceValidator{
		targets: make([]PerformanceTarget, 0),
	}
}

// ValidateResults validates the test results against performance targets
func (pv *PerformanceValidator) ValidateResults(results *TestResults) {
	pv.targets = make([]PerformanceTarget, 0)

	// Parse response times for calculations
	responseTimes := make([]time.Duration, len(results.ResponseTimes))
	for i, entry := range results.ResponseTimes {
		responseTimes[i] = entry.ResponseTime
	}

	var avgResponseTime, p95ResponseTime, p99ResponseTime time.Duration
	if len(responseTimes) > 0 {
		// Calculate average
		var total time.Duration
		for _, rt := range responseTimes {
			total += rt
		}
		avgResponseTime = total / time.Duration(len(responseTimes))

		// Calculate percentiles (simple approximation)
		if len(responseTimes) > 0 {
			// Sort response times (simple bubble sort for small datasets)
			for i := 0; i < len(responseTimes)-1; i++ {
				for j := 0; j < len(responseTimes)-i-1; j++ {
					if responseTimes[j] > responseTimes[j+1] {
						responseTimes[j], responseTimes[j+1] = responseTimes[j+1], responseTimes[j]
					}
				}
			}

			p95Index := int(float64(len(responseTimes)) * 0.95)
			p99Index := int(float64(len(responseTimes)) * 0.99)

			if p95Index >= len(responseTimes) {
				p95Index = len(responseTimes) - 1
			}
			if p99Index >= len(responseTimes) {
				p99Index = len(responseTimes) - 1
			}

			p95ResponseTime = responseTimes[p95Index]
			p99ResponseTime = responseTimes[p99Index]
		}
	}

	// Throughput validation
	pv.targets = append(pv.targets, PerformanceTarget{
		Name:        "Requests per Second",
		Target:      "≥ 1000 req/s",
		Actual:      fmt.Sprintf("%.1f req/s", results.RequestsPerSecond),
		Description: "Target: >1000 req/s for competitive advantage vs Ghost (~400 users)",
		Passed:      results.RequestsPerSecond >= 1000,
	})

	// 95th percentile response time
	p95Ms := float64(p95ResponseTime.Nanoseconds()) / 1e6
	pv.targets = append(pv.targets, PerformanceTarget{
		Name:        "95th Percentile Response Time",
		Target:      "< 50ms",
		Actual:      fmt.Sprintf("%.1fms", p95Ms),
		Description: "Target: <50ms (4x faster than Ghost ~200ms)",
		Passed:      p95ResponseTime < 50*time.Millisecond,
	})

	// Average response time
	avgMs := float64(avgResponseTime.Nanoseconds()) / 1e6
	pv.targets = append(pv.targets, PerformanceTarget{
		Name:        "Average Response Time",
		Target:      "< 30ms",
		Actual:      fmt.Sprintf("%.1fms", avgMs),
		Description: "Target: <30ms for excellent user experience",
		Passed:      avgResponseTime < 30*time.Millisecond,
	})

	// Error rate
	errorRate := float64(results.FailedRequests) / float64(results.TotalRequests) * 100
	pv.targets = append(pv.targets, PerformanceTarget{
		Name:        "Error Rate",
		Target:      "< 1%",
		Actual:      fmt.Sprintf("%.2f%%", errorRate),
		Description: "Target: <1% for production reliability",
		Passed:      errorRate < 1.0,
	})

	// 99th percentile response time
	p99Ms := float64(p99ResponseTime.Nanoseconds()) / 1e6
	pv.targets = append(pv.targets, PerformanceTarget{
		Name:        "99th Percentile Response Time",
		Target:      "< 100ms",
		Actual:      fmt.Sprintf("%.1fms", p99Ms),
		Description: "Target: <100ms even for worst-case scenarios",
		Passed:      p99ResponseTime < 100*time.Millisecond,
	})

	// Success rate
	successRate := (float64(results.SuccessfulRequests) / float64(results.TotalRequests)) * 100
	pv.targets = append(pv.targets, PerformanceTarget{
		Name:        "Success Rate",
		Target:      "> 99%",
		Actual:      fmt.Sprintf("%.2f%%", successRate),
		Description: "Target: >99% successful requests for production readiness",
		Passed:      successRate > 99.0,
	})
}

// PrintValidationReport prints a detailed performance validation report
func (pv *PerformanceValidator) PrintValidationReport() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 PERFORMANCE TARGET VALIDATION")
	fmt.Println(strings.Repeat("=", 60))

	passed := 0
	total := len(pv.targets)

	for _, target := range pv.targets {
		status := "❌ FAIL"
		if target.Passed {
			status = "✅ PASS"
			passed++
		}

		fmt.Printf("%s %-30s %s\n", status, target.Name+":", target.Actual)
		fmt.Printf("    %s\n", target.Description)
		fmt.Println()
	}

	fmt.Printf("Overall: %d/%d targets met (%.1f%%)\n", passed, total, float64(passed)/float64(total)*100)

	if passed == total {
		fmt.Println("🎉 ALL PERFORMANCE TARGETS MET! MarkGo is ready for production.")
	} else if passed >= total*3/4 {
		fmt.Println("⚠️  Most targets met, but some optimization may be needed.")
	} else {
		fmt.Println("🔧 Significant performance improvements needed before production.")
	}

	pv.printCompetitiveAnalysis()
}

// printCompetitiveAnalysis prints competitive comparison
func (pv *PerformanceValidator) printCompetitiveAnalysis() {
	fmt.Println("\n🏆 COMPETITIVE ANALYSIS")
	fmt.Println(strings.Repeat("=", 25))

	// Check if we have performance metrics
	hasGoodPerf := false
	hasGoodThroughput := false

	for _, target := range pv.targets {
		if target.Name == "95th Percentile Response Time" && target.Passed {
			fmt.Println("✅ Faster than Ghost (typical ~200ms)")
			fmt.Println("✅ Faster than WordPress (typical ~300-500ms)")
			hasGoodPerf = true
		}
		if target.Name == "Requests per Second" && target.Passed {
			fmt.Println("✅ Higher throughput than Ghost (400 concurrent user limit)")
			fmt.Println("✅ Excellent scalability for production workloads")
			hasGoodThroughput = true
		}
	}

	if !hasGoodPerf {
		fmt.Println("⚠️  Response times need improvement to beat Ghost/WordPress")
	}
	if !hasGoodThroughput {
		fmt.Println("⚠️  Throughput needs improvement for competitive advantage")
	}

	fmt.Println("✅ Single binary deployment advantage over all competitors")
	fmt.Println("✅ Dynamic features advantage over Hugo (static site generator)")
	fmt.Println("✅ Git-friendly content management vs traditional CMSs")
}

// GetValidationSummary returns a summary of the validation results
func (pv *PerformanceValidator) GetValidationSummary() map[string]interface{} {
	passed := 0
	total := len(pv.targets)

	targetDetails := make([]map[string]interface{}, len(pv.targets))
	for i, target := range pv.targets {
		if target.Passed {
			passed++
		}
		targetDetails[i] = map[string]interface{}{
			"name":        target.Name,
			"target":      target.Target,
			"actual":      target.Actual,
			"description": target.Description,
			"passed":      target.Passed,
		}
	}

	return map[string]interface{}{
		"targets_met":    passed,
		"total_targets":  total,
		"success_rate":   float64(passed) / float64(total) * 100,
		"overall_status": pv.getOverallStatus(passed, total),
		"targets":        targetDetails,
	}
}

// getOverallStatus returns the overall validation status
func (pv *PerformanceValidator) getOverallStatus(passed, total int) string {
	if passed == total {
		return "PRODUCTION_READY"
	} else if passed >= total*3/4 {
		return "MOSTLY_READY"
	} else {
		return "NEEDS_IMPROVEMENT"
	}
}
