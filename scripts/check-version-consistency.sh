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
    "cmd/init/main.go:version.*=.*\""
    "internal/handlers/api.go:\"version\".*\""
    "internal/services/logging.go:\"version\".*\""
)

ERRORS=0

for file_pattern in "${VERSION_FILES[@]}"; do
    IFS=':' read -r file pattern <<< "$file_pattern"

    if [[ ! -f "$file" ]]; then
        echo -e "${YELLOW}⚠️  File not found: $file${NC}"
        continue
    fi

    # Extract version from file
    VERSION_IN_FILE=$(grep -E "$pattern" "$file" | sed -E 's/.*"([^"]+)".*/\1/' | head -1)

    if [[ -z "$VERSION_IN_FILE" ]]; then
        echo -e "${YELLOW}⚠️  No version found in: $file${NC}"
        continue
    fi

    if [[ "$VERSION_IN_FILE" != "$CANONICAL_VERSION" ]]; then
        echo -e "${RED}❌ Version mismatch in $file:${NC}"
        echo -e "   Expected: ${GREEN}$CANONICAL_VERSION${NC}"
        echo -e "   Found:    ${RED}$VERSION_IN_FILE${NC}"
        ((ERRORS++))
    else
        echo -e "${GREEN}✅ $file: $VERSION_IN_FILE${NC}"
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

# Look for version patterns in Go files (excluding test files and generated files)
VERSION_PATTERNS=$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -name "*_test.go" -exec grep -l "\"[0-9]\+\.[0-9]\+\.[0-9]\+\"" {} \; 2>/dev/null || true)

if [[ -n "$VERSION_PATTERNS" ]]; then
    echo -e "${YELLOW}⚠️  Found potential hardcoded versions in:${NC}"
    echo "$VERSION_PATTERNS"
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