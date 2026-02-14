#!/bin/bash

# Test runner script for markgo
# This script runs all tests with coverage and generates reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_DIR="coverage"
COVERAGE_FILE="$COVERAGE_DIR/coverage.out"
COVERAGE_HTML="$COVERAGE_DIR/coverage.html"
COVERAGE_XML="$COVERAGE_DIR/coverage.xml"
MIN_COVERAGE=80

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}         markgo Test Suite${NC}"
echo -e "${BLUE}===========================================${NC}"

# Create coverage directory
mkdir -p $COVERAGE_DIR

# Clean previous coverage data
rm -f $COVERAGE_FILE $COVERAGE_HTML $COVERAGE_XML

echo -e "${YELLOW}🧹 Cleaning up previous test artifacts...${NC}"

# Run go mod tidy to ensure dependencies are up to date
echo -e "${YELLOW}📦 Ensuring dependencies are up to date...${NC}"
go mod tidy

# Run go mod download to ensure all dependencies are available
go mod download

# Verify go mod
echo -e "${YELLOW}🔍 Verifying go modules...${NC}"
go mod verify

echo -e "${YELLOW}🔧 Running go vet...${NC}"
go vet ./...

echo -e "${YELLOW}🎯 Running tests with coverage...${NC}"

# Run tests with coverage
go test -v -race -coverprofile=$COVERAGE_FILE -covermode=atomic ./...

# Check if tests passed
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
else
    echo -e "${RED}❌ Some tests failed!${NC}"
    exit 1
fi

# Generate coverage report
echo -e "${YELLOW}📊 Generating coverage reports...${NC}"

# Generate HTML coverage report
go tool cover -html=$COVERAGE_FILE -o $COVERAGE_HTML

# Generate text coverage report
COVERAGE_PERCENT=$(go tool cover -func=$COVERAGE_FILE | grep total | awk '{print $3}' | sed 's/%//')

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}         Coverage Report${NC}"
echo -e "${BLUE}===========================================${NC}"

# Display coverage by package
go tool cover -func=$COVERAGE_FILE

echo -e "${BLUE}===========================================${NC}"

# Check coverage threshold
if (( $(echo "$COVERAGE_PERCENT >= $MIN_COVERAGE" | bc -l) )); then
    echo -e "${GREEN}✅ Coverage: ${COVERAGE_PERCENT}% (meets minimum threshold of ${MIN_COVERAGE}%)${NC}"
else
    echo -e "${RED}❌ Coverage: ${COVERAGE_PERCENT}% (below minimum threshold of ${MIN_COVERAGE}%)${NC}"
    echo -e "${RED}Please add more tests to improve coverage.${NC}"
    exit 1
fi

# Generate XML coverage report for CI/CD
if command -v gocov &> /dev/null && command -v gocov-xml &> /dev/null; then
    echo -e "${YELLOW}📋 Generating XML coverage report...${NC}"
    gocov convert $COVERAGE_FILE | gocov-xml > $COVERAGE_XML
fi

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}         Test Summary${NC}"
echo -e "${BLUE}===========================================${NC}"

# Count test files and test functions
TEST_FILES=$(find . -name "*_test.go" | wc -l)
TEST_FUNCTIONS=$(grep -r "^func Test" . --include="*_test.go" | wc -l)
BENCHMARK_FUNCTIONS=$(grep -r "^func Benchmark" . --include="*_test.go" | wc -l)

echo -e "${GREEN}📁 Test files: $TEST_FILES${NC}"
echo -e "${GREEN}🧪 Test functions: $TEST_FUNCTIONS${NC}"
echo -e "${GREEN}⚡ Benchmark functions: $BENCHMARK_FUNCTIONS${NC}"
echo -e "${GREEN}📊 Coverage: ${COVERAGE_PERCENT}%${NC}"

# Show coverage files generated
echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}         Generated Reports${NC}"
echo -e "${BLUE}===========================================${NC}"

echo -e "${GREEN}📄 Coverage profile: $COVERAGE_FILE${NC}"
echo -e "${GREEN}🌐 HTML report: $COVERAGE_HTML${NC}"
if [ -f $COVERAGE_XML ]; then
    echo -e "${GREEN}📋 XML report: $COVERAGE_XML${NC}"
fi

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}         Additional Commands${NC}"
echo -e "${BLUE}===========================================${NC}"

echo -e "${YELLOW}To run specific tests:${NC}"
echo -e "  go test -v ./internal/config"
echo -e "  go test -v ./internal/models"
echo -e "  go test -v ./internal/services"
echo -e "  go test -v ./internal/handlers"
echo -e "  go test -v ./internal/middleware"
echo -e "  go test -v ./cmd/markgo"

echo -e "${YELLOW}To run benchmarks:${NC}"
echo -e "  go test -bench=. ./..."

echo -e "${YELLOW}To run tests with verbose output:${NC}"
echo -e "  go test -v -race ./..."

echo -e "${YELLOW}To open HTML coverage report:${NC}"
echo -e "  open $COVERAGE_HTML"

echo -e "${YELLOW}To run tests continuously:${NC}"
echo -e "  find . -name '*.go' | entr -c go test ./..."

echo -e "${BLUE}===========================================${NC}"
echo -e "${GREEN}✅ Test suite completed successfully!${NC}"
echo -e "${BLUE}===========================================${NC}"

# Optional: Open coverage report in browser (uncomment if desired)
# if command -v open &> /dev/null; then
#     echo -e "${YELLOW}Opening coverage report in browser...${NC}"
#     open $COVERAGE_HTML
# elif command -v xdg-open &> /dev/null; then
#     echo -e "${YELLOW}Opening coverage report in browser...${NC}"
#     xdg-open $COVERAGE_HTML
# fi

exit 0
