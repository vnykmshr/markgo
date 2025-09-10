package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StressTestConfig struct {
	BaseURL     string `json:"base_url"`
	Concurrency int    `json:"concurrency"`
	Duration    string `json:"duration"`
	Timeout     string `json:"timeout"`
	UserAgent   string `json:"user_agent"`
	FollowLinks bool   `json:"follow_links"`
	MaxDepth    int    `json:"max_depth"`
	OutputFile  string `json:"output_file"`
	Verbose     bool   `json:"verbose"`
}

func main() {
	var (
		configPath  = flag.String("config", "", "Path to stress test configuration file")
		baseURL     = flag.String("url", "http://localhost:3000", "Base URL to test")
		concurrency = flag.Int("concurrency", 10, "Number of concurrent requests")
		duration    = flag.String("duration", "2m", "Duration to run tests")
		timeout     = flag.String("timeout", "30s", "Request timeout")
		userAgent   = flag.String("user-agent", "MarkGo-StressTester/1.0", "User agent string")
		followLinks = flag.Bool("follow-links", true, "Follow links found in pages")
		maxDepth    = flag.Int("max-depth", 3, "Maximum crawl depth")
		outputFile  = flag.String("output", "", "Output file for results (JSON)")
		verbose     = flag.Bool("verbose", false, "Verbose output")
		help        = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load configuration
	var testConfig StressTestConfig
	if *configPath != "" {
		if err := loadConfig(*configPath, &testConfig); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	} else {
		testConfig = StressTestConfig{
			BaseURL:     *baseURL,
			Concurrency: *concurrency,
			Duration:    *duration,
			Timeout:     *timeout,
			UserAgent:   *userAgent,
			FollowLinks: *followLinks,
			MaxDepth:    *maxDepth,
			OutputFile:  *outputFile,
			Verbose:     *verbose,
		}
	}

	// Setup logger
	logLevel := slog.LevelInfo
	if testConfig.Verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Parse duration
	testDuration, err := time.ParseDuration(testConfig.Duration)
	if err != nil {
		log.Fatalf("Invalid duration: %v", err)
	}

	// Parse timeout
	requestTimeout, err := time.ParseDuration(testConfig.Timeout)
	if err != nil {
		log.Fatalf("Invalid timeout: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	// Initialize stress tester
	tester := NewStressTester(StressTesterConfig{
		BaseURL:        testConfig.BaseURL,
		Concurrency:    testConfig.Concurrency,
		RequestTimeout: requestTimeout,
		UserAgent:      testConfig.UserAgent,
		FollowLinks:    testConfig.FollowLinks,
		MaxDepth:       testConfig.MaxDepth,
		Logger:         logger,
	})

	// Run stress test
	logger.Info("Starting stress test",
		"base_url", testConfig.BaseURL,
		"concurrency", testConfig.Concurrency,
		"duration", testConfig.Duration,
		"follow_links", testConfig.FollowLinks,
		"max_depth", testConfig.MaxDepth)

	results, err := tester.Run(ctx)
	if err != nil {
		log.Fatalf("Stress test failed: %v", err)
	}

	// Output results
	if testConfig.OutputFile != "" {
		if err := saveResults(testConfig.OutputFile, results); err != nil {
			log.Fatalf("Failed to save results: %v", err)
		}
		logger.Info("Results saved", "file", testConfig.OutputFile)

		// Generate HTML report
		htmlFile := strings.TrimSuffix(testConfig.OutputFile, filepath.Ext(testConfig.OutputFile)) + ".html"
		reporter := NewReportGenerator(results)
		if err := reporter.GenerateHTMLReport(htmlFile); err != nil {
			logger.Error("Failed to generate HTML report", "error", err)
		} else {
			logger.Info("HTML report generated", "file", htmlFile)
		}
	}

	// Print summary
	printSummary(results)
}

func showHelp() {
	fmt.Println(`MarkGo Stress Tester

A comprehensive stress testing tool that discovers URLs and validates responses.

USAGE:
    go run cmd/stress-test/main.go [OPTIONS]

OPTIONS:
    -config string
        Path to stress test configuration file (JSON)
    -url string
        Base URL to test (default: http://localhost:3000)
    -concurrency int
        Number of concurrent requests (default: 10)
    -duration string
        Duration to run tests (default: 2m)
    -timeout string
        Request timeout (default: 30s)
    -user-agent string
        User agent string (default: MarkGo-StressTester/1.0)
    -follow-links
        Follow links found in pages (default: true)
    -max-depth int
        Maximum crawl depth (default: 3)
    -output string
        Output file for results (JSON format)
    -verbose
        Verbose output
    -help
        Show this help message

EXAMPLES:
    # Basic stress test
    go run cmd/stress-test/main.go -url http://localhost:3000

    # High concurrency test with custom duration
    go run cmd/stress-test/main.go -url http://localhost:3000 -concurrency 50 -duration 5m

    # Test with link following disabled
    go run cmd/stress-test/main.go -url http://localhost:3000 -follow-links=false

    # Save results to file
    go run cmd/stress-test/main.go -url http://localhost:3000 -output results.json

    # Use configuration file
    go run cmd/stress-test/main.go -config stress-test-config.json

CONFIGURATION FILE EXAMPLE:
    {
        "base_url": "http://localhost:3000",
        "concurrency": 20,
        "duration": "3m",
        "timeout": "30s",
        "user_agent": "MarkGo-StressTester/1.0",
        "follow_links": true,
        "max_depth": 3,
        "output_file": "stress_test_results.json",
        "verbose": true
    }`)
}

func loadConfig(path string, config *StressTestConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("parsing config JSON: %w", err)
	}

	return nil
}

func saveResults(filename string, results *TestResults) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling results: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("writing results file: %w", err)
	}

	return nil
}

func printSummary(results *TestResults) {
	fmt.Printf("\n=== STRESS TEST RESULTS ===\n")
	fmt.Printf("Duration: %s\n", results.Duration)
	fmt.Printf("URLs Discovered: %d\n", results.URLsDiscovered)
	fmt.Printf("Total Requests: %d\n", results.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", results.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", results.FailedRequests)
	fmt.Printf("Average Response Time: %s\n", results.AverageResponseTime)
	fmt.Printf("Min Response Time: %s\n", results.MinResponseTime)
	fmt.Printf("Max Response Time: %s\n", results.MaxResponseTime)
	fmt.Printf("Requests/Second: %.2f\n", results.RequestsPerSecond)
	fmt.Printf("Success Rate: %.2f%%\n", results.SuccessRate)

	if len(results.Errors) > 0 {
		fmt.Printf("\nERRORS:\n")
		errorCounts := make(map[string]int)
		for _, err := range results.Errors {
			errorCounts[err.Error]++
		}
		for errorMsg, count := range errorCounts {
			fmt.Printf("  %s: %d occurrences\n", errorMsg, count)
		}
	}

	if len(results.SlowRequests) > 0 {
		fmt.Printf("\nSLOWEST REQUESTS:\n")
		for i, req := range results.SlowRequests {
			if i >= 10 { // Show top 10
				break
			}
			fmt.Printf("  %s: %s\n", req.URL, req.ResponseTime)
		}
	}

	fmt.Printf("\n=== URL VALIDATION SUMMARY ===\n")
	statusCounts := make(map[int]int)
	for _, validation := range results.URLValidations {
		statusCounts[validation.StatusCode]++
	}

	for status, count := range statusCounts {
		fmt.Printf("HTTP %d: %d URLs\n", status, count)
	}
}
