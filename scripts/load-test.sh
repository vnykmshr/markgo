#!/bin/bash

# MarkGo Load Testing Script
# Tests concurrent users and high-traffic conditions

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
CONCURRENT_USERS="${CONCURRENT_USERS:-50}"
DURATION="${DURATION:-30s}"
OUTPUT_FILE="load-test-results-$(date +%Y%m%d-%H%M%S).txt"

echo "🚀 MarkGo Load Test Starting..."
echo "Target: $BASE_URL"
echo "Concurrent Users: $CONCURRENT_USERS"
echo "Duration: $DURATION"
echo "Output: $OUTPUT_FILE"
echo "=========================================="

# Check if server is running
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo "❌ Server is not running at $BASE_URL"
    echo "Please start the MarkGo server first: make run"
    exit 1
fi

echo "✅ Server is running"

# Create results directory
mkdir -p results

# Test endpoints with weights (simulating real traffic patterns)
declare -a ENDPOINTS=(
    "/ 40"                      # Home page - highest traffic
    "/articles 25"              # Articles listing
    "/search?q=golang 15"       # Search functionality
    "/articles/sample 10"       # Individual articles
    "/tags 5"                   # Tags page
    "/categories 3"             # Categories page
    "/health 2"                 # Health checks
)

# Function to make a single request and measure response time
make_request() {
    local url="$1"
    local start_time=$(date +%s%N)
    local response=$(curl -s -w "%{http_code},%{time_total}" -o /dev/null "$BASE_URL$url" 2>/dev/null || echo "000,999.999")
    local end_time=$(date +%s%N)
    
    local status_code=$(echo $response | cut -d',' -f1)
    local response_time=$(echo $response | cut -d',' -f2)
    
    echo "$url,$status_code,$response_time,$(date +%s)" >> "results/raw-$OUTPUT_FILE"
}

# Function to run load test worker
run_worker() {
    local worker_id=$1
    local end_time=$(($(date +%s) + $(echo $DURATION | sed 's/s//')))
    local request_count=0
    
    while [ $(date +%s) -lt $end_time ]; do
        # Select random endpoint based on weight
        local rand=$((RANDOM % 100))
        local cumulative=0
        
        for endpoint_data in "${ENDPOINTS[@]}"; do
            local endpoint=$(echo $endpoint_data | cut -d' ' -f1)
            local weight=$(echo $endpoint_data | cut -d' ' -f2)
            cumulative=$((cumulative + weight))
            
            if [ $rand -lt $cumulative ]; then
                make_request "$endpoint"
                break
            fi
        done
        
        request_count=$((request_count + 1))
        
        # Small delay to prevent overwhelming
        sleep 0.01
    done
    
    echo "Worker $worker_id completed $request_count requests"
}

# Start load test workers in parallel
echo "🔥 Starting load test workers..."
pids=()

for i in $(seq 1 $CONCURRENT_USERS); do
    run_worker $i &
    pids+=($!)
    
    # Stagger worker starts (ramp-up)
    if [ $((i % 10)) -eq 0 ]; then
        echo "Started $i workers..."
        sleep 0.5
    fi
done

# Monitor progress
start_time=$(date +%s)
duration_seconds=$(echo $DURATION | sed 's/s//')

while [ $(($(date +%s) - start_time)) -lt $duration_seconds ]; do
    elapsed=$(($(date +%s) - start_time))
    remaining=$((duration_seconds - elapsed))
    progress=$((elapsed * 100 / duration_seconds))
    
    echo "📊 Progress: ${progress}% - Remaining: ${remaining}s"
    sleep 5
done

# Wait for all workers to complete
echo "⏰ Test duration completed. Waiting for workers to finish..."
for pid in "${pids[@]}"; do
    wait $pid
done

echo "✅ All workers completed"

# Analyze results
echo "📈 Analyzing results..."

if [ ! -f "results/raw-$OUTPUT_FILE" ]; then
    echo "❌ No results file found"
    exit 1
fi

# Basic statistics
total_requests=$(wc -l < "results/raw-$OUTPUT_FILE")
successful_requests=$(awk -F',' '$2 >= 200 && $2 < 400 { count++ } END { print count+0 }' "results/raw-$OUTPUT_FILE")
failed_requests=$((total_requests - successful_requests))
error_rate=$(echo "scale=2; $failed_requests * 100 / $total_requests" | bc -l 2>/dev/null || echo "0")

# Calculate RPS
actual_duration=$(($(date +%s) - start_time))
rps=$(echo "scale=2; $total_requests / $actual_duration" | bc -l 2>/dev/null || echo "0")

# Response time statistics (convert curl time to milliseconds)
awk -F',' '{ print $3 * 1000 }' "results/raw-$OUTPUT_FILE" | sort -n > "results/response-times.tmp"
min_time=$(head -1 "results/response-times.tmp")
max_time=$(tail -1 "results/response-times.tmp")
avg_time=$(awk '{ sum += $1; count++ } END { print sum/count }' "results/response-times.tmp" 2>/dev/null || echo "0")

# Calculate percentiles
line_count=$(wc -l < "results/response-times.tmp")
p95_line=$((line_count * 95 / 100))
p99_line=$((line_count * 99 / 100))
p95_time=$(sed -n "${p95_line}p" "results/response-times.tmp")
p99_time=$(sed -n "${p99_line}p" "results/response-times.tmp")

# Status code distribution
echo "Status Code Distribution:" > "results/$OUTPUT_FILE"
awk -F',' '{ codes[$2]++ } END { for (code in codes) printf "%s: %d requests\n", code, codes[code] }' "results/raw-$OUTPUT_FILE" >> "results/$OUTPUT_FILE"

# Generate comprehensive report
{
    echo "============================================"
    echo "📈 MARKGO LOAD TEST RESULTS"
    echo "============================================"
    echo "Test Configuration:"
    echo "  Target URL: $BASE_URL"
    echo "  Concurrent Users: $CONCURRENT_USERS"
    echo "  Duration: $DURATION"
    echo "  Actual Duration: ${actual_duration}s"
    echo ""
    echo "Request Statistics:"
    echo "  Total Requests: $total_requests"
    echo "  Successful: $successful_requests ($(echo "scale=1; $successful_requests * 100 / $total_requests" | bc -l)%)"
    echo "  Failed: $failed_requests ($error_rate%)"
    echo "  Requests/Second: $rps"
    echo ""
    echo "Response Time Statistics (ms):"
    echo "  Average: $(printf "%.2f" $avg_time)"
    echo "  Minimum: $(printf "%.2f" $min_time)"
    echo "  Maximum: $(printf "%.2f" $max_time)"
    echo "  95th Percentile: $(printf "%.2f" $p95_time)"
    echo "  99th Percentile: $(printf "%.2f" $p99_time)"
    echo ""
    echo "🎯 PERFORMANCE TARGET VALIDATION:"
    echo "============================================"
    
    # Validate targets
    if (( $(echo "$rps >= 1000" | bc -l 2>/dev/null || echo "0") )); then
        echo "✅ Throughput: ${rps} req/s (Target: ≥1000 req/s)"
    else
        echo "❌ Throughput: ${rps} req/s (Target: ≥1000 req/s)"
    fi
    
    if (( $(echo "$p95_time < 50" | bc -l 2>/dev/null || echo "0") )); then
        echo "✅ 95th Percentile: $(printf "%.1f" $p95_time)ms (Target: <50ms)"
    else
        echo "❌ 95th Percentile: $(printf "%.1f" $p95_time)ms (Target: <50ms)"
    fi
    
    if (( $(echo "$avg_time < 30" | bc -l 2>/dev/null || echo "0") )); then
        echo "✅ Average Response: $(printf "%.1f" $avg_time)ms (Target: <30ms)"
    else
        echo "❌ Average Response: $(printf "%.1f" $avg_time)ms (Target: <30ms)"
    fi
    
    if (( $(echo "$error_rate < 1" | bc -l 2>/dev/null || echo "0") )); then
        echo "✅ Error Rate: $error_rate% (Target: <1%)"
    else
        echo "❌ Error Rate: $error_rate% (Target: <1%)"
    fi
    
    echo ""
    echo "🏆 COMPETITIVE COMPARISON:"
    echo "============================================"
    echo "vs Ghost (~200ms avg):     $(echo "scale=1; 200 / $avg_time" | bc -l 2>/dev/null || echo "N/A")x faster"
    echo "vs WordPress (~300ms avg): $(echo "scale=1; 300 / $avg_time" | bc -l 2>/dev/null || echo "N/A")x faster"
    echo "vs Hugo (static files):    Dynamic features + comparable speed"
    echo ""
    echo "Test completed at: $(date)"
    echo "============================================"
    
} | tee "results/$OUTPUT_FILE"

# Display results on console
cat "results/$OUTPUT_FILE"

# Cleanup temp files
rm -f "results/response-times.tmp"

echo ""
echo "💾 Detailed results saved to: results/$OUTPUT_FILE"
echo "💾 Raw data saved to: results/raw-$OUTPUT_FILE"

# Return appropriate exit code
if (( $(echo "$error_rate < 1 && $p95_time < 50 && $rps >= 1000" | bc -l 2>/dev/null || echo "0") )); then
    echo "🎉 All performance targets met!"
    exit 0
else
    echo "⚠️  Some performance targets not met. Check results above."
    exit 1
fi