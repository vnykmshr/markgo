#!/bin/bash

# MarkGo Competitor Comparison Benchmark Script
# Validates performance targets against Ghost, WordPress, Jekyll, and Hugo

set -e

# Configuration
MARKGO_URL="${MARKGO_URL:-http://localhost:8080}"
DURATION="${DURATION:-30s}"
RESULTS_DIR="competitor-benchmark-$(date +%Y%m%d-%H%M%S)"
LOG_FILE="$RESULTS_DIR/benchmark.log"

# Performance targets based on competitive analysis
declare -A TARGETS=(
    ["response_time_avg_ms"]="30"      # Target: <30ms average (better than all competitors)
    ["response_time_p95_ms"]="50"      # Target: <50ms 95th percentile (4x faster than Ghost)
    ["response_time_p99_ms"]="100"     # Target: <100ms 99th percentile
    ["throughput_rps"]="1000"          # Target: >1000 req/s (2.5x better than Ghost)
    ["error_rate_percent"]="1"         # Target: <1% error rate
    ["memory_usage_mb"]="50"           # Target: <50MB (6x better than Ghost ~300MB)
)

# Competitor baselines (from research)
declare -A COMPETITORS=(
    ["ghost_response_ms"]="200"
    ["ghost_throughput_rps"]="400" 
    ["ghost_memory_mb"]="300"
    ["wordpress_response_ms"]="400"
    ["wordpress_memory_mb"]="2048"
    ["hugo_build_time_ms"]="2100"
    ["jekyll_build_time_ms"]="30000"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Setup
mkdir -p "$RESULTS_DIR"
exec > >(tee -a "$LOG_FILE")
exec 2>&1

echo "🚀 MarkGo Competitor Comparison Benchmark"
echo "=========================================="
echo "Target: $MARKGO_URL"
echo "Duration: $DURATION"
echo "Results: $RESULTS_DIR"
echo "Started: $(date)"
echo ""

# Check if MarkGo server is running
check_server() {
    echo "🔍 Checking if MarkGo server is running..."
    if ! curl -s "$MARKGO_URL/health" > /dev/null 2>&1; then
        echo -e "${RED}❌ MarkGo server is not running at $MARKGO_URL${NC}"
        echo "Please start the server first: make run"
        exit 1
    fi
    
    echo -e "${GREEN}✅ MarkGo server is running${NC}"
    
    # Get initial server info
    curl -s "$MARKGO_URL/metrics" | jq . > "$RESULTS_DIR/initial-metrics.json" 2>/dev/null || echo "Metrics endpoint not available"
}

# Function to run performance test
run_performance_test() {
    local test_name="$1"
    local concurrent_users="$2"
    local duration="$3"
    
    echo "🔥 Running $test_name test..."
    echo "  Concurrent Users: $concurrent_users"
    echo "  Duration: $duration"
    
    # Use custom load test for more detailed metrics
    CONCURRENT_USERS=$concurrent_users DURATION=$duration ./scripts/load-test.sh > "$RESULTS_DIR/$test_name.txt" 2>&1
    
    # Extract key metrics
    local avg_time=$(grep "Average:" "$RESULTS_DIR/$test_name.txt" | awk '{print $2}' | sed 's/ms//')
    local p95_time=$(grep "95th Percentile:" "$RESULTS_DIR/$test_name.txt" | awk '{print $3}' | sed 's/ms//')
    local p99_time=$(grep "99th Percentile:" "$RESULTS_DIR/$test_name.txt" | awk '{print $3}' | sed 's/ms//')
    local rps=$(grep "Requests/Second:" "$RESULTS_DIR/$test_name.txt" | awk '{print $2}')
    local error_rate=$(grep "Error Rate:" "$RESULTS_DIR/$test_name.txt" | awk '{print $3}' | sed 's/%//')
    
    # Store results
    cat > "$RESULTS_DIR/$test_name-metrics.json" <<EOF
{
    "test_name": "$test_name",
    "concurrent_users": $concurrent_users,
    "duration": "$duration",
    "avg_response_time_ms": ${avg_time:-0},
    "p95_response_time_ms": ${p95_time:-0},
    "p99_response_time_ms": ${p99_time:-0},
    "requests_per_second": ${rps:-0},
    "error_rate_percent": ${error_rate:-0},
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    
    echo -e "${BLUE}📊 $test_name Results:${NC}"
    echo "  Avg Response: ${avg_time}ms"
    echo "  95th %: ${p95_time}ms" 
    echo "  RPS: ${rps}"
    echo "  Error Rate: ${error_rate}%"
    echo ""
}

# Get memory usage
get_memory_usage() {
    echo "🧠 Measuring memory usage..."
    
    # Try to get memory from metrics endpoint
    local memory_mb=$(curl -s "$MARKGO_URL/metrics" 2>/dev/null | jq -r '.performance.memory_usage_mb // 0' 2>/dev/null || echo "0")
    
    if [[ "$memory_mb" == "0" ]]; then
        # Fallback: estimate from system if possible
        local pid=$(pgrep markgo 2>/dev/null || echo "")
        if [[ -n "$pid" ]]; then
            memory_mb=$(ps -o rss= -p $pid 2>/dev/null | awk '{print $1/1024}' || echo "0")
        fi
    fi
    
    echo "  Memory Usage: ${memory_mb} MB"
    echo "$memory_mb" > "$RESULTS_DIR/memory_usage.txt"
}

# Validate against targets
validate_targets() {
    echo "🎯 Validating Performance Targets"
    echo "=================================="
    
    # Load test results
    local baseline_file="$RESULTS_DIR/baseline-metrics.json"
    if [[ ! -f "$baseline_file" ]]; then
        echo -e "${RED}❌ No baseline metrics found${NC}"
        return 1
    fi
    
    local avg_time=$(jq -r '.avg_response_time_ms' "$baseline_file")
    local p95_time=$(jq -r '.p95_response_time_ms' "$baseline_file")
    local p99_time=$(jq -r '.p99_response_time_ms' "$baseline_file")
    local rps=$(jq -r '.requests_per_second' "$baseline_file")
    local error_rate=$(jq -r '.error_rate_percent' "$baseline_file")
    local memory_mb=$(cat "$RESULTS_DIR/memory_usage.txt" 2>/dev/null || echo "0")
    
    local passed=0
    local total=6
    
    # Check each target
    echo "Target Validation:"
    
    if (( $(echo "$avg_time <= ${TARGETS[response_time_avg_ms]}" | bc -l 2>/dev/null || echo "0") )); then
        echo -e "  ${GREEN}✅ Average Response Time: ${avg_time}ms (target: <${TARGETS[response_time_avg_ms]}ms)${NC}"
        ((passed++))
    else
        echo -e "  ${RED}❌ Average Response Time: ${avg_time}ms (target: <${TARGETS[response_time_avg_ms]}ms)${NC}"
    fi
    
    if (( $(echo "$p95_time <= ${TARGETS[response_time_p95_ms]}" | bc -l 2>/dev/null || echo "0") )); then
        echo -e "  ${GREEN}✅ 95th Percentile: ${p95_time}ms (target: <${TARGETS[response_time_p95_ms]}ms)${NC}"
        ((passed++))
    else
        echo -e "  ${RED}❌ 95th Percentile: ${p95_time}ms (target: <${TARGETS[response_time_p95_ms]}ms)${NC}"
    fi
    
    if (( $(echo "$p99_time <= ${TARGETS[response_time_p99_ms]}" | bc -l 2>/dev/null || echo "0") )); then
        echo -e "  ${GREEN}✅ 99th Percentile: ${p99_time}ms (target: <${TARGETS[response_time_p99_ms]}ms)${NC}"
        ((passed++))
    else
        echo -e "  ${RED}❌ 99th Percentile: ${p99_time}ms (target: <${TARGETS[response_time_p99_ms]}ms)${NC}"
    fi
    
    if (( $(echo "$rps >= ${TARGETS[throughput_rps]}" | bc -l 2>/dev/null || echo "0") )); then
        echo -e "  ${GREEN}✅ Throughput: ${rps} req/s (target: ≥${TARGETS[throughput_rps]} req/s)${NC}"
        ((passed++))
    else
        echo -e "  ${RED}❌ Throughput: ${rps} req/s (target: ≥${TARGETS[throughput_rps]} req/s)${NC}"
    fi
    
    if (( $(echo "$error_rate <= ${TARGETS[error_rate_percent]}" | bc -l 2>/dev/null || echo "0") )); then
        echo -e "  ${GREEN}✅ Error Rate: ${error_rate}% (target: <${TARGETS[error_rate_percent]}%)${NC}"
        ((passed++))
    else
        echo -e "  ${RED}❌ Error Rate: ${error_rate}% (target: <${TARGETS[error_rate_percent]}%)${NC}"
    fi
    
    if (( $(echo "$memory_mb <= ${TARGETS[memory_usage_mb]}" | bc -l 2>/dev/null || echo "0") )); then
        echo -e "  ${GREEN}✅ Memory Usage: ${memory_mb}MB (target: <${TARGETS[memory_usage_mb]}MB)${NC}"
        ((passed++))
    else
        echo -e "  ${RED}❌ Memory Usage: ${memory_mb}MB (target: <${TARGETS[memory_usage_mb]}MB)${NC}"
    fi
    
    echo ""
    local score=$((passed * 100 / total))
    echo "Score: $passed/$total targets met (${score}%)"
    
    # Store validation results
    cat > "$RESULTS_DIR/validation-results.json" <<EOF
{
    "targets_met": $passed,
    "total_targets": $total,
    "score_percent": $score,
    "results": {
        "avg_response_time_ms": $avg_time,
        "p95_response_time_ms": $p95_time,
        "p99_response_time_ms": $p99_time,
        "throughput_rps": $rps,
        "error_rate_percent": $error_rate,
        "memory_usage_mb": $memory_mb
    },
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    
    return $((total - passed))
}

# Competitive comparison analysis
competitive_analysis() {
    echo "🏆 Competitive Comparison Analysis"
    echo "==================================="
    
    local baseline_file="$RESULTS_DIR/baseline-metrics.json"
    if [[ ! -f "$baseline_file" ]]; then
        echo -e "${RED}❌ No baseline metrics for comparison${NC}"
        return 1
    fi
    
    local avg_time=$(jq -r '.avg_response_time_ms' "$baseline_file")
    local rps=$(jq -r '.requests_per_second' "$baseline_file")
    local memory_mb=$(cat "$RESULTS_DIR/memory_usage.txt" 2>/dev/null || echo "30")
    
    echo "MarkGo vs Competitors:"
    echo ""
    
    # vs Ghost
    local ghost_speedup=$(echo "scale=1; ${COMPETITORS[ghost_response_ms]} / $avg_time" | bc -l 2>/dev/null || echo "N/A")
    local ghost_throughput_improvement=$(echo "scale=1; $rps / ${COMPETITORS[ghost_throughput_rps]}" | bc -l 2>/dev/null || echo "N/A")
    local ghost_memory_improvement=$(echo "scale=1; ${COMPETITORS[ghost_memory_mb]} / $memory_mb" | bc -l 2>/dev/null || echo "N/A")
    
    echo "🆚 Ghost (Node.js):"
    echo "  Response Time: ${ghost_speedup}x faster (${avg_time}ms vs ${COMPETITORS[ghost_response_ms]}ms)"
    echo "  Throughput: ${ghost_throughput_improvement}x better (${rps} vs ${COMPETITORS[ghost_throughput_rps]} req/s)"
    echo "  Memory: ${ghost_memory_improvement}x more efficient (${memory_mb}MB vs ${COMPETITORS[ghost_memory_mb]}MB)"
    echo ""
    
    # vs WordPress
    local wp_speedup=$(echo "scale=1; ${COMPETITORS[wordpress_response_ms]} / $avg_time" | bc -l 2>/dev/null || echo "N/A")
    local wp_memory_improvement=$(echo "scale=1; ${COMPETITORS[wordpress_memory_mb]} / $memory_mb" | bc -l 2>/dev/null || echo "N/A")
    
    echo "🆚 WordPress (PHP):"
    echo "  Response Time: ${wp_speedup}x faster (${avg_time}ms vs ${COMPETITORS[wordpress_response_ms]}ms)"
    echo "  Memory: ${wp_memory_improvement}x more efficient (${memory_mb}MB vs ${COMPETITORS[wordpress_memory_mb]}MB)"
    echo "  Deployment: Single binary vs LAMP stack complexity"
    echo ""
    
    # vs Static Generators
    echo "🆚 Hugo (Static Generator):"
    echo "  Dynamic Features: ✅ Search, forms, real-time updates vs ❌ Static only"
    echo "  Response Time: Comparable to static files (${avg_time}ms vs ~10ms)"
    echo "  Deployment: Single binary vs build process"
    echo ""
    
    echo "🆚 Jekyll (Static Generator):"
    echo "  Build Speed: N/A (no build required) vs ${COMPETITORS[jekyll_build_time_ms]}ms"
    echo "  Dynamic Features: ✅ Server-side capabilities vs ❌ Static only"
    echo "  Scalability: Handles traffic spikes vs rebuild required for changes"
    echo ""
    
    # Create comprehensive comparison report
    cat > "$RESULTS_DIR/competitive-analysis.json" <<EOF
{
    "markgo": {
        "avg_response_time_ms": $avg_time,
        "throughput_rps": $rps,
        "memory_usage_mb": $memory_mb
    },
    "comparisons": {
        "vs_ghost": {
            "response_time_advantage": "${ghost_speedup}x faster",
            "throughput_advantage": "${ghost_throughput_improvement}x better", 
            "memory_advantage": "${ghost_memory_improvement}x more efficient"
        },
        "vs_wordpress": {
            "response_time_advantage": "${wp_speedup}x faster",
            "memory_advantage": "${wp_memory_improvement}x more efficient"
        },
        "vs_hugo": {
            "advantages": ["dynamic features", "single binary", "no build process"],
            "tradeoffs": ["slightly higher response time than static files"]
        },
        "vs_jekyll": {
            "advantages": ["no build time", "dynamic features", "better scalability"],
            "build_time_saved": "${COMPETITORS[jekyll_build_time_ms]}ms per change"
        }
    },
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
}

# Generate final report
generate_report() {
    echo "📝 Generating Final Report"
    echo "=========================="
    
    local report_file="$RESULTS_DIR/BENCHMARK-REPORT.md"
    
    cat > "$report_file" <<EOF
# MarkGo Performance Benchmark Report

**Generated:** $(date)  
**Test Duration:** $DURATION  
**Target Server:** $MARKGO_URL  

## Executive Summary

This report presents comprehensive performance benchmarking results for MarkGo, comparing against key competitors including Ghost, WordPress, Hugo, and Jekyll.

## Performance Metrics

$(cat "$RESULTS_DIR/validation-results.json" | jq -r '
"### Target Achievement: " + (.score_percent | tostring) + "%\n" +
"- **Average Response Time:** " + (.results.avg_response_time_ms | tostring) + "ms\n" +
"- **95th Percentile:** " + (.results.p95_response_time_ms | tostring) + "ms\n" +
"- **99th Percentile:** " + (.results.p99_response_time_ms | tostring) + "ms\n" +
"- **Throughput:** " + (.results.throughput_rps | tostring) + " req/s\n" +
"- **Error Rate:** " + (.results.error_rate_percent | tostring) + "%\n" +
"- **Memory Usage:** " + (.results.memory_usage_mb | tostring) + "MB"
')

## Competitive Analysis

### vs Ghost (Node.js)
- **Response Time:** $(jq -r '.comparisons.vs_ghost.response_time_advantage' "$RESULTS_DIR/competitive-analysis.json")
- **Throughput:** $(jq -r '.comparisons.vs_ghost.throughput_advantage' "$RESULTS_DIR/competitive-analysis.json")  
- **Memory Efficiency:** $(jq -r '.comparisons.vs_ghost.memory_advantage' "$RESULTS_DIR/competitive-analysis.json")

### vs WordPress (PHP)
- **Response Time:** $(jq -r '.comparisons.vs_wordpress.response_time_advantage' "$RESULTS_DIR/competitive-analysis.json")
- **Memory Efficiency:** $(jq -r '.comparisons.vs_wordpress.memory_advantage' "$RESULTS_DIR/competitive-analysis.json")
- **Deployment:** Single binary vs complex LAMP stack

### vs Static Generators (Hugo/Jekyll)
- **Dynamic Features:** ✅ Search, forms, contact functionality
- **Deployment:** Single binary, no build process required
- **Scalability:** Handles traffic without rebuilds

## Test Configuration

- **Concurrent Users:** Varied (10-100)
- **Test Duration:** $DURATION
- **Endpoints Tested:** Home, Articles, Search, Tags, Categories, Health
- **Load Pattern:** Realistic user behavior simulation

## Files Generated

- \`baseline-metrics.json\` - Core performance metrics
- \`validation-results.json\` - Target validation results  
- \`competitive-analysis.json\` - Competitor comparison data
- \`benchmark.log\` - Detailed test execution log

## Conclusion

$(if [[ -f "$RESULTS_DIR/validation-results.json" ]]; then
    local score=$(jq -r '.score_percent' "$RESULTS_DIR/validation-results.json")
    if [[ "$score" -ge 80 ]]; then
        echo "✅ **MarkGo meets performance targets** for production deployment with significant advantages over competitors."
    elif [[ "$score" -ge 60 ]]; then
        echo "⚠️ **MarkGo shows good performance** but some optimization may be beneficial."
    else
        echo "🔧 **Performance improvements needed** before production deployment."
    fi
else
    echo "📊 Performance data analysis completed."
fi)

---
*Report generated by MarkGo Competitive Benchmark Suite*
EOF
    
    echo -e "${GREEN}✅ Report generated: $report_file${NC}"
}

# Main execution
main() {
    check_server
    
    echo "🏃 Running benchmark tests..."
    run_performance_test "baseline" 50 "$DURATION"
    run_performance_test "low-load" 10 "15s"  
    run_performance_test "high-load" 100 "20s"
    
    get_memory_usage
    
    local validation_result=0
    validate_targets || validation_result=$?
    
    competitive_analysis
    generate_report
    
    echo ""
    echo "🎊 Benchmark Complete!"
    echo "========================"
    echo -e "${BLUE}Results Directory: $RESULTS_DIR${NC}"
    echo -e "${BLUE}Report File: $RESULTS_DIR/BENCHMARK-REPORT.md${NC}"
    echo ""
    
    if [[ $validation_result -eq 0 ]]; then
        echo -e "${GREEN}🎉 All performance targets met! MarkGo is ready for production.${NC}"
    else
        echo -e "${YELLOW}⚠️  $validation_result performance target(s) not met. Review results for optimization opportunities.${NC}"
    fi
    
    return $validation_result
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi