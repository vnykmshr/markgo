#!/bin/bash

# Article validation script for MarkGo
# Validates frontmatter format and structure of markdown articles

set -e

ARTICLES_DIR="${1:-articles}"
ERROR_COUNT=0
WARN_COUNT=0

echo "🔍 Validating articles in $ARTICLES_DIR/"
echo "=================================="

# Check if articles directory exists
if [ ! -d "$ARTICLES_DIR" ]; then
    echo "❌ Articles directory not found: $ARTICLES_DIR"
    exit 1
fi

# Function to validate individual article
validate_article() {
    local file="$1"
    local filename=$(basename "$file")
    local has_frontmatter=false
    local frontmatter_closed=false
    local line_num=0
    local title_found=false
    local date_found=false
    local author_found=false
    
    echo "📄 Validating: $filename"
    
    # Check if file exists and is readable
    if [ ! -r "$file" ]; then
        echo "  ❌ Cannot read file: $file"
        ((ERROR_COUNT++))
        return
    fi
    
    # Check if file is empty
    if [ ! -s "$file" ]; then
        echo "  ⚠️  Empty file: $file"
        ((WARN_COUNT++))
        return
    fi
    
    # Read file line by line
    while IFS= read -r line || [ -n "$line" ]; do
        ((line_num++))
        
        # Check for frontmatter start
        if [ $line_num -eq 1 ] && [ "$line" = "---" ]; then
            has_frontmatter=true
            continue
        fi
        
        # Check for frontmatter end
        if [ "$has_frontmatter" = true ] && [ "$frontmatter_closed" = false ] && [ "$line" = "---" ]; then
            frontmatter_closed=true
            continue
        fi
        
        # Parse frontmatter fields
        if [ "$has_frontmatter" = true ] && [ "$frontmatter_closed" = false ]; then
            case "$line" in
                title:*)
                    title_found=true
                    # Check if title is quoted and not empty
                    if [[ ! "$line" =~ title:\ *\".*\" ]]; then
                        echo "  ⚠️  Title should be quoted: line $line_num"
                        ((WARN_COUNT++))
                    fi
                    ;;
                date:*)
                    date_found=true
                    # Check basic date format (YYYY-MM-DD or RFC3339)
                    if [[ ! "$line" =~ date:\ *[0-9]{4}-[0-9]{2}-[0-9]{2} ]]; then
                        echo "  ⚠️  Date format may be invalid: line $line_num"
                        ((WARN_COUNT++))
                    fi
                    ;;
                author:*)
                    author_found=true
                    ;;
                draft:*)
                    # Check if draft is boolean
                    if [[ ! "$line" =~ draft:\ *(true|false) ]]; then
                        echo "  ⚠️  Draft should be true or false: line $line_num"
                        ((WARN_COUNT++))
                    fi
                    ;;
                featured:*)
                    # Check if featured is boolean
                    if [[ ! "$line" =~ featured:\ *(true|false) ]]; then
                        echo "  ⚠️  Featured should be true or false: line $line_num"
                        ((WARN_COUNT++))
                    fi
                    ;;
            esac
        fi
    done < "$file"
    
    # Validation checks
    if [ "$has_frontmatter" = false ]; then
        echo "  ❌ Missing frontmatter"
        ((ERROR_COUNT++))
    elif [ "$frontmatter_closed" = false ]; then
        echo "  ❌ Frontmatter not properly closed"
        ((ERROR_COUNT++))
    fi
    
    if [ "$title_found" = false ]; then
        echo "  ❌ Missing title field"
        ((ERROR_COUNT++))
    fi
    
    if [ "$date_found" = false ]; then
        echo "  ⚠️  Missing date field"
        ((WARN_COUNT++))
    fi
    
    if [ "$author_found" = false ]; then
        echo "  ⚠️  Missing author field"
        ((WARN_COUNT++))
    fi
    
    # Check file extension
    if [[ ! "$filename" =~ \.(md|markdown)$ ]]; then
        echo "  ⚠️  Unexpected file extension (should be .md or .markdown)"
        ((WARN_COUNT++))
    fi
    
    echo "  ✅ Validation complete"
    echo ""
}

# Find and validate all markdown files
find "$ARTICLES_DIR" -name "*.md" -o -name "*.markdown" | while read -r file; do
    validate_article "$file"
done

# Summary
echo "=================================="
echo "📊 Validation Summary:"
echo "   Errors: $ERROR_COUNT"
echo "   Warnings: $WARN_COUNT"

if [ $ERROR_COUNT -eq 0 ] && [ $WARN_COUNT -eq 0 ]; then
    echo "🎉 All articles validated successfully!"
    exit 0
elif [ $ERROR_COUNT -eq 0 ]; then
    echo "✅ Validation passed with warnings"
    exit 0
else
    echo "❌ Validation failed with errors"
    exit 1
fi