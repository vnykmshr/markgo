# MarkGo Stress Testing Tool

A comprehensive stress testing and URL validation tool that automatically discovers URLs by crawling your application and validates all responses under concurrent load.

## Features

🚀 **Concurrent Load Testing**: Configurable concurrency levels to simulate real-world traffic
🔍 **Automatic URL Discovery**: Crawls your application to discover all available URLs
📊 **Comprehensive Reporting**: HTML and JSON reports with detailed metrics and visualizations
⚡ **Real-time Monitoring**: Progress updates during test execution
🎯 **Response Validation**: Validates HTTP status codes, response times, and content
🔗 **Link Following**: Automatically follows links to discover deep pages
📈 **Performance Metrics**: Response time distribution, throughput, and success rates
❌ **Error Tracking**: Detailed error reporting and slow request identification

## Quick Start

### Basic Usage

```bash
# Start your MarkGo server
go run cmd/server/main.go

# In another terminal, run a basic stress test
go run cmd/stress-test/main.go -url http://localhost:3000
```

### Advanced Usage

```bash
# High concurrency test with custom duration
go run cmd/stress-test/main.go \
  -url http://localhost:3000 \
  -concurrency 50 \
  -duration 5m \
  -output results.json

# Generate HTML report
go run cmd/stress-test/main.go \
  -url http://localhost:3000 \
  -concurrency 20 \
  -duration 2m \
  -output results.json && \
  go run cmd/stress-test/reporter.go -input results.json -html report.html
```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `-url` | `http://localhost:3000` | Base URL to test |
| `-concurrency` | `10` | Number of concurrent requests |
| `-duration` | `2m` | Test duration (e.g., 30s, 5m, 1h) |
| `-timeout` | `30s` | Request timeout |
| `-user-agent` | `MarkGo-StressTester/1.0` | User agent string |
| `-follow-links` | `true` | Follow links found in pages |
| `-max-depth` | `3` | Maximum crawl depth |
| `-output` | | Output file for results (JSON) |
| `-verbose` | `false` | Verbose logging |
| `-config` | | Path to configuration file |

## Configuration File

Create a JSON configuration file for complex setups:

```json
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
}
```

Use with: `go run cmd/stress-test/main.go -config config.json`

## How It Works

### 1. URL Discovery
The tool starts with your base URL and automatically discovers additional URLs by:
- Parsing HTML responses for `href` attributes
- Following internal links (same domain only)
- Respecting maximum crawl depth
- Avoiding duplicate URLs and infinite loops

### 2. Concurrent Testing
- Maintains a configurable number of concurrent HTTP clients
- Distributes discovered URLs across worker goroutines
- Uses context-based cancellation for clean shutdown
- Implements rate limiting to prevent overwhelming the server

### 3. Response Validation
Each request is validated for:
- HTTP status codes (2xx = success, 3xx = redirect, 4xx/5xx = error)
- Response times and performance metrics
- Content-Type and Content-Length headers
- Link extraction for further crawling
- Error detection and categorization

### 4. Real-time Monitoring
- Progress updates every 10 seconds
- Live metrics: requests sent, success/failure rates, URLs discovered
- Queue size monitoring
- Memory usage tracking

## Report Types

### Console Output
Real-time progress and final summary displayed in the terminal.

### JSON Report
Detailed machine-readable results including:
```json
{
  "duration": "2m0s",
  "urls_discovered": 42,
  "total_requests": 1250,
  "successful_requests": 1200,
  "failed_requests": 50,
  "average_response_time": "150ms",
  "requests_per_second": 10.42,
  "success_rate": 96.0,
  "url_validations": [...],
  "errors": [...],
  "response_times": [...]
}
```

### HTML Report
Beautiful, interactive HTML report with:
- Overview dashboard with key metrics
- Response status distribution charts
- URL validation table with sorting/filtering
- Response time histograms
- Error analysis and slow request identification
- Mobile-responsive design

## Use Cases

### Development Testing
```bash
# Quick validation of all pages
go run cmd/stress-test/main.go -duration 30s -concurrency 5
```

### Pre-deployment Validation
```bash
# Comprehensive test before release
go run cmd/stress-test/main.go \
  -duration 10m \
  -concurrency 25 \
  -max-depth 5 \
  -output pre-deploy-results.json
```

### Performance Regression Testing
```bash
# Compare results across versions
go run cmd/stress-test/main.go \
  -config regression-test.json \
  -output "results-v$(git rev-parse --short HEAD).json"
```

### Load Testing
```bash
# High-load scenario testing
go run cmd/stress-test/main.go \
  -concurrency 100 \
  -duration 15m \
  -follow-links=false \
  -output load-test-results.json
```

## Best Practices

### Server Preparation
1. Ensure your MarkGo server is running with sufficient resources
2. Consider using production-like configuration
3. Monitor server metrics during testing

### Test Configuration
1. Start with low concurrency (5-10) and increase gradually
2. Use shorter durations for initial testing
3. Disable link following for pure load testing
4. Enable verbose logging for troubleshooting

### Result Analysis
1. Focus on success rate (should be >95% for healthy applications)
2. Monitor average response times (<500ms for good UX)
3. Investigate any 5xx errors immediately
4. Check for memory leaks during long-running tests

## Troubleshooting

### Common Issues

**High Error Rate**
- Check server resources and logs
- Reduce concurrency level
- Increase request timeout
- Verify server is properly started

**Slow Performance**
- Monitor server CPU/memory usage
- Check database connection pools
- Review slow request logs
- Consider caching strategies

**Memory Issues**
- Limit crawl depth with `-max-depth`
- Reduce test duration
- Disable link following for simple load tests

### Debug Mode
Enable verbose logging to see detailed request/response information:
```bash
go run cmd/stress-test/main.go -url http://localhost:3000 -verbose
```

## Architecture

The stress tester consists of several components:

- **Main Controller**: Orchestrates the test execution and configuration
- **URL Discovery Engine**: Crawls and discovers URLs using regex parsing
- **Concurrent Workers**: Pool of goroutines handling HTTP requests
- **Results Collector**: Thread-safe aggregation of metrics and validations
- **Report Generator**: Creates HTML and JSON output formats
- **Monitoring System**: Real-time progress tracking and logging

## Integration

### CI/CD Pipeline
```yaml
- name: Stress Test
  run: |
    go run cmd/server/main.go &
    sleep 5
    go run cmd/stress-test/main.go -duration 2m -output ci-results.json
    # Parse results and fail if success rate < 95%
```

### Automated Testing
Integrate into your testing suite to catch performance regressions early.

### Monitoring Integration
Export results to monitoring systems like Prometheus or DataDog for trend analysis.

## Performance Considerations

- **Memory Usage**: ~10MB base + ~1KB per discovered URL
- **CPU Usage**: Scales with concurrency level
- **Network**: Respects server capacity and implements backoff
- **Storage**: JSON results ~1-5MB for typical applications

## Contributing

To extend the stress tester:

1. Add new validation rules in `stress_tester.go`
2. Extend report formats in `reporter.go`
3. Add new configuration options in `main.go`
4. Update this README with new features

## License

This stress testing tool is part of the MarkGo project and follows the same license terms.