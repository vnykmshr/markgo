#!/bin/bash
# MarkGo Engine - Version Consistency Checker
# Ensures all version strings match the canonical version in constants.go

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get canonical version from constants.go
CANONICAL_VERSION=$(grep -E 'AppVersion.*=.*"' internal/constants/constants.go | sed -E 's/.*"(.*)".*/\1/')

if [[ -z "$CANONICAL_VERSION" ]]; then
    echo -e "${RED}❌ Could not find AppVersion in internal/constants/constants.go${NC}"
    exit 1
fi

echo -e "${GREEN}📋 Canonical version: ${CANONICAL_VERSION}${NC}"

# Files to check for version consistency
declare -a VERSION_FILES=(
    "cmd/init/main.go:constants\.AppVersion"
    "internal/handlers/api.go:constants\.AppVersion"
    "internal/services/logging.go:constants\.AppVersion"
)

ERRORS=0

for file_pattern in "${VERSION_FILES[@]}"; do
    IFS=':' read -r file pattern <<< "$file_pattern"

    if [[ ! -f "$file" ]]; then
        echo -e "${YELLOW}⚠️  File not found: $file${NC}"
        continue
    fi

    # Check if file uses constants.AppVersion
    if grep -q "$pattern" "$file"; then
        echo -e "${GREEN}✅ $file: uses constants.AppVersion${NC}"
    else
        echo -e "${RED}❌ $file: does not use constants.AppVersion${NC}"
        ((ERRORS++))
    fi
done

# Check release notes files
RELEASE_NOTES_PATTERN="RELEASE-NOTES-${CANONICAL_VERSION}.md"
if [[ ! -f "$RELEASE_NOTES_PATTERN" ]]; then
    echo -e "${YELLOW}⚠️  Release notes file not found: $RELEASE_NOTES_PATTERN${NC}"
fi

# Check if there are any hardcoded versions that don't match
echo -e "\n${GREEN}🔍 Checking for hardcoded version patterns...${NC}"

# Remove the 'v' prefix for numeric version checks
NUMERIC_VERSION="${CANONICAL_VERSION#v}"

# Look for hardcoded version patterns in Go files (excluding test files, generated files, and expected locations)
echo -e "${GREEN}🔍 Checking for hardcoded version strings...${NC}"
HARDCODED_VERSIONS=$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -name "*_test.go" -not -path "./internal/constants/constants.go" -not -path "./cmd/server/main.go" -exec grep -l "\"v1\.[0-9]\+\.[0-9]\+\"" {} \; 2>/dev/null || true)

if [[ -n "$HARDCODED_VERSIONS" ]]; then
    echo -e "${RED}❌ Found hardcoded version strings in:${NC}"
    echo "$HARDCODED_VERSIONS"
    echo -e "${RED}   These should use constants.AppVersion instead${NC}"
    ((ERRORS++))
else
    echo -e "${GREEN}✅ No hardcoded version strings found (excluding expected locations)${NC}"
fi

# Look for other numeric version patterns that might be unintentional
OTHER_VERSIONS=$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -name "*_test.go" -exec grep -l "\"[0-9]\+\.[0-9]\+\.[0-9]\+\"" {} \; 2>/dev/null || true)

if [[ -n "$OTHER_VERSIONS" ]]; then
    echo -e "${YELLOW}⚠️  Found other numeric version patterns in:${NC}"
    echo "$OTHER_VERSIONS"
    echo -e "${YELLOW}   Please verify these are intentional and not version references${NC}"
fi

if [[ $ERRORS -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 All version strings are consistent!${NC}"
    exit 0
else
    echo -e "\n${RED}💥 Found $ERRORS version inconsistency(ies)${NC}"
    echo -e "${YELLOW}💡 To fix: Update the inconsistent versions to match $CANONICAL_VERSION${NC}"
    exit 1
fi