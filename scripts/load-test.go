package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// LoadTestConfig holds configuration for load tests
type LoadTestConfig struct {
	BaseURL         string
	ConcurrentUsers int
	Duration        time.Duration
	RampUpTime      time.Duration
	RequestTimeout  time.Duration
	Endpoints       []EndpointConfig
}

type EndpointConfig struct {
	Path   string
	Method string
	Weight int // Relative frequency of this endpoint
}

// LoadTestResult holds results from load testing
type LoadTestResult struct {
	TotalRequests       int
	SuccessfulRequests  int
	FailedRequests      int
	RequestsPerSecond   float64
	AverageResponseTime time.Duration
	MinResponseTime     time.Duration
	MaxResponseTime     time.Duration
	P95ResponseTime     time.Duration
	P99ResponseTime     time.Duration
	ErrorRate           float64
	ResponseTimes       []time.Duration
	StatusCodes         map[int]int
}

// Request represents a single HTTP request result
type Request struct {
	URL          string
	Method       string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
	Timestamp    time.Time
}

func main() {
	// Default load test configuration
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8080",
		ConcurrentUsers: 50,
		Duration:        30 * time.Second,
		RampUpTime:      5 * time.Second,
		RequestTimeout:  10 * time.Second,
		Endpoints: []EndpointConfig{
			{Path: "/", Method: "GET", Weight: 30},                        // Homepage - most frequent
			{Path: "/articles", Method: "GET", Weight: 20},                // Articles listing
			{Path: "/search?q=golang", Method: "GET", Weight: 15},         // Search functionality
			{Path: "/articles/sample-article", Method: "GET", Weight: 15}, // Article details
			{Path: "/tags", Method: "GET", Weight: 10},                    // Tags page
			{Path: "/categories", Method: "GET", Weight: 5},               // Categories page
			{Path: "/health", Method: "GET", Weight: 5},                   // Health checks
		},
	}

	fmt.Println("🚀 Starting MarkGo Load Test")
	fmt.Printf("Target: %s\n", config.BaseURL)
	fmt.Printf("Concurrent Users: %d\n", config.ConcurrentUsers)
	fmt.Printf("Duration: %v\n", config.Duration)
	fmt.Printf("Ramp-up Time: %v\n", config.RampUpTime)
	fmt.Println()

	// Run the load test
	result, err := runLoadTest(config)
	if err != nil {
		log.Fatalf("Load test failed: %v", err)
	}

	// Display results
	displayResults(result, config)

	// Validate performance targets
	validatePerformanceTargets(result)
}

func runLoadTest(config LoadTestConfig) (*LoadTestResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration+config.RampUpTime+10*time.Second)
	defer cancel()

	results := make(chan Request, 10000)
	var wg sync.WaitGroup

	// Start result collector
	var allResults []Request
	var resultsMutex sync.Mutex

	go func() {
		for req := range results {
			resultsMutex.Lock()
			allResults = append(allResults, req)
			resultsMutex.Unlock()
		}
	}()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.RequestTimeout,
	}

	// Start load test workers with ramp-up
	rampUpInterval := config.RampUpTime / time.Duration(config.ConcurrentUsers)

	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)

		// Stagger worker starts for ramp-up
		time.Sleep(rampUpInterval)

		go func(workerID int) {
			defer wg.Done()
			runWorker(ctx, workerID, config, client, results)
		}(i)
	}

	// Start progress reporter
	go reportProgress(ctx, config.Duration)

	// Wait for all workers to complete
	wg.Wait()
	close(results)

	// Allow result collector to finish
	time.Sleep(100 * time.Millisecond)

	return analyzeResults(allResults), nil
}

func runWorker(ctx context.Context, workerID int, config LoadTestConfig, client *http.Client, results chan<- Request) {
	endTime := time.Now().Add(config.Duration)

	for time.Now().Before(endTime) {
		select {
		case <-ctx.Done():
			return
		default:
			// Select endpoint based on weight
			endpoint := selectEndpoint(config.Endpoints)
			url := config.BaseURL + endpoint.Path

			// Make request
			startTime := time.Now()
			resp, err := client.Do(createRequest(endpoint.Method, url))
			responseTime := time.Since(startTime)

			var statusCode int
			if resp != nil {
				statusCode = resp.StatusCode
				_, _ = io.Copy(io.Discard, resp.Body) // Drain and close body
				resp.Body.Close()
			}

			results <- Request{
				URL:          url,
				Method:       endpoint.Method,
				StatusCode:   statusCode,
				ResponseTime: responseTime,
				Error:        err,
				Timestamp:    startTime,
			}

			// Small delay to prevent overwhelming the server
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
		}
	}
}

func selectEndpoint(endpoints []EndpointConfig) EndpointConfig {
	// Calculate total weight
	totalWeight := 0
	for _, ep := range endpoints {
		totalWeight += ep.Weight
	}

	// Select random endpoint based on weight
	r := rand.Intn(totalWeight)
	currentWeight := 0

	for _, ep := range endpoints {
		currentWeight += ep.Weight
		if r < currentWeight {
			return ep
		}
	}

	return endpoints[0] // fallback
}

func createRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil
	}

	// Add realistic headers
	req.Header.Set("User-Agent", "MarkGo-LoadTest/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	return req
}

func reportProgress(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			remaining := duration - elapsed

			if remaining <= 0 {
				fmt.Println("⏰ Load test completed!")
				return
			}

			progress := float64(elapsed) / float64(duration) * 100
			fmt.Printf("📊 Progress: %.1f%% - Remaining: %v\n", progress, remaining.Round(time.Second))
		}
	}
}

func analyzeResults(results []Request) *LoadTestResult {
	if len(results) == 0 {
		return &LoadTestResult{}
	}

	totalRequests := len(results)
	successfulRequests := 0
	failedRequests := 0
	var responseTimes []time.Duration
	statusCodes := make(map[int]int)

	var minTime, maxTime, totalTime time.Duration
	minTime = time.Hour // Initialize to large value

	// Process all results
	firstRequest := results[0].Timestamp
	lastRequest := results[len(results)-1].Timestamp
	testDuration := lastRequest.Sub(firstRequest)

	for _, req := range results {
		if req.Error == nil && req.StatusCode >= 200 && req.StatusCode < 400 {
			successfulRequests++
		} else {
			failedRequests++
		}

		responseTimes = append(responseTimes, req.ResponseTime)
		statusCodes[req.StatusCode]++

		totalTime += req.ResponseTime
		if req.ResponseTime < minTime {
			minTime = req.ResponseTime
		}
		if req.ResponseTime > maxTime {
			maxTime = req.ResponseTime
		}
	}

	avgTime := totalTime / time.Duration(totalRequests)
	rps := float64(totalRequests) / testDuration.Seconds()
	errorRate := float64(failedRequests) / float64(totalRequests) * 100

	// Calculate percentiles
	p95 := calculatePercentile(responseTimes, 0.95)
	p99 := calculatePercentile(responseTimes, 0.99)

	return &LoadTestResult{
		TotalRequests:       totalRequests,
		SuccessfulRequests:  successfulRequests,
		FailedRequests:      failedRequests,
		RequestsPerSecond:   rps,
		AverageResponseTime: avgTime,
		MinResponseTime:     minTime,
		MaxResponseTime:     maxTime,
		P95ResponseTime:     p95,
		P99ResponseTime:     p99,
		ErrorRate:           errorRate,
		ResponseTimes:       responseTimes,
		StatusCodes:         statusCodes,
	}
}

func calculatePercentile(times []time.Duration, percentile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}

	// Simple percentile calculation with sorting
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)

	// Bubble sort (simple but works for our use case)
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

func displayResults(result *LoadTestResult, config LoadTestConfig) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📈 LOAD TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("Total Requests:       %d\n", result.TotalRequests)
	fmt.Printf("Successful Requests:  %d (%.1f%%)\n", result.SuccessfulRequests,
		float64(result.SuccessfulRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("Failed Requests:      %d (%.1f%%)\n", result.FailedRequests, result.ErrorRate)
	fmt.Printf("Requests per Second:  %.2f\n", result.RequestsPerSecond)

	fmt.Println("\n📊 Response Times:")
	fmt.Printf("  Average:  %v\n", result.AverageResponseTime)
	fmt.Printf("  Minimum:  %v\n", result.MinResponseTime)
	fmt.Printf("  Maximum:  %v\n", result.MaxResponseTime)
	fmt.Printf("  95th %%:   %v\n", result.P95ResponseTime)
	fmt.Printf("  99th %%:   %v\n", result.P99ResponseTime)

	fmt.Println("\n🌐 HTTP Status Codes:")
	for code, count := range result.StatusCodes {
		percentage := float64(count) / float64(result.TotalRequests) * 100
		fmt.Printf("  %d: %d requests (%.1f%%)\n", code, count, percentage)
	}

	// Response time distribution
	fmt.Println("\n⏱️  Response Time Distribution:")
	displayResponseTimeDistribution(result.ResponseTimes)

	fmt.Println("\n" + strings.Repeat("=", 60))
}

func displayResponseTimeDistribution(times []time.Duration) {
	if len(times) == 0 {
		return
	}

	ranges := []struct {
		label string
		min   time.Duration
		max   time.Duration
	}{
		{"< 10ms", 0, 10 * time.Millisecond},
		{"10-50ms", 10 * time.Millisecond, 50 * time.Millisecond},
		{"50-100ms", 50 * time.Millisecond, 100 * time.Millisecond},
		{"100-200ms", 100 * time.Millisecond, 200 * time.Millisecond},
		{"200-500ms", 200 * time.Millisecond, 500 * time.Millisecond},
		{"> 500ms", 500 * time.Millisecond, time.Hour},
	}

	total := len(times)

	for _, r := range ranges {
		count := 0
		for _, t := range times {
			if t >= r.min && t < r.max {
				count++
			}
		}

		if count > 0 {
			percentage := float64(count) / float64(total) * 100
			bar := strings.Repeat("█", int(percentage/2)) // Visual bar
			fmt.Printf("  %-10s %6d (%.1f%%) %s\n", r.label, count, percentage, bar)
		}
	}
}

func validatePerformanceTargets(result *LoadTestResult) {
	fmt.Println("\n🎯 PERFORMANCE TARGET VALIDATION")
	fmt.Println(strings.Repeat("=", 40))

	targets := []struct {
		name       string
		target     interface{}
		actual     interface{}
		comparison string
		passed     bool
	}{
		{
			name:       "Requests per Second",
			target:     "≥ 1000 req/s",
			actual:     fmt.Sprintf("%.0f req/s", result.RequestsPerSecond),
			comparison: "Target: >1000 req/s for competitive advantage",
			passed:     result.RequestsPerSecond >= 1000,
		},
		{
			name:       "95th Percentile Response Time",
			target:     "< 50ms",
			actual:     fmt.Sprintf("%.1fms", float64(result.P95ResponseTime.Nanoseconds())/1e6),
			comparison: "Target: <50ms (4x faster than Ghost ~200ms)",
			passed:     result.P95ResponseTime < 50*time.Millisecond,
		},
		{
			name:       "Average Response Time",
			target:     "< 30ms",
			actual:     fmt.Sprintf("%.1fms", float64(result.AverageResponseTime.Nanoseconds())/1e6),
			comparison: "Target: <30ms for excellent user experience",
			passed:     result.AverageResponseTime < 30*time.Millisecond,
		},
		{
			name:       "Error Rate",
			target:     "< 1%",
			actual:     fmt.Sprintf("%.2f%%", result.ErrorRate),
			comparison: "Target: <1% for production reliability",
			passed:     result.ErrorRate < 1.0,
		},
		{
			name:       "99th Percentile Response Time",
			target:     "< 100ms",
			actual:     fmt.Sprintf("%.1fms", float64(result.P99ResponseTime.Nanoseconds())/1e6),
			comparison: "Target: <100ms even for worst-case scenarios",
			passed:     result.P99ResponseTime < 100*time.Millisecond,
		},
	}

	passed := 0
	total := len(targets)

	for _, target := range targets {
		status := "❌ FAIL"
		if target.passed {
			status = "✅ PASS"
			passed++
		}

		fmt.Printf("%s %-30s %s\n", status, target.name+":", target.actual)
		fmt.Printf("    %s\n", target.comparison)
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

	// Competitive analysis summary
	fmt.Println("\n🏆 COMPETITIVE ANALYSIS")
	fmt.Println(strings.Repeat("=", 25))

	if result.P95ResponseTime < 50*time.Millisecond {
		fmt.Println("✅ Faster than Ghost (typical ~200ms)")
		fmt.Println("✅ Faster than WordPress (typical ~300-500ms)")
	}

	if result.RequestsPerSecond > 1000 {
		fmt.Println("✅ Higher throughput than Ghost (400 concurrent user limit)")
		fmt.Println("✅ Excellent scalability for production workloads")
	}

	fmt.Println("✅ Single binary deployment advantage over all competitors")
	fmt.Println("✅ Dynamic features advantage over Hugo")

	// Save results to JSON for further analysis
	saveResultsToJSON(result)
}

func saveResultsToJSON(result *LoadTestResult) {
	// Convert to JSON for external analysis
	data := map[string]interface{}{
		"timestamp":            time.Now().Format(time.RFC3339),
		"total_requests":       result.TotalRequests,
		"successful_requests":  result.SuccessfulRequests,
		"failed_requests":      result.FailedRequests,
		"requests_per_second":  result.RequestsPerSecond,
		"avg_response_time_ms": float64(result.AverageResponseTime.Nanoseconds()) / 1e6,
		"min_response_time_ms": float64(result.MinResponseTime.Nanoseconds()) / 1e6,
		"max_response_time_ms": float64(result.MaxResponseTime.Nanoseconds()) / 1e6,
		"p95_response_time_ms": float64(result.P95ResponseTime.Nanoseconds()) / 1e6,
		"p99_response_time_ms": float64(result.P99ResponseTime.Nanoseconds()) / 1e6,
		"error_rate_percent":   result.ErrorRate,
		"status_codes":         result.StatusCodes,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal results to JSON: %v", err)
		return
	}

	filename := fmt.Sprintf("load-test-results-%d.json", time.Now().Unix())
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		log.Printf("Failed to save results to file: %v", err)
		return
	}

	fmt.Printf("\n💾 Results saved to: %s\n", filename)
}
