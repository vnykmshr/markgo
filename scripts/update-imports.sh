#!/bin/bash

# Update Import Paths Script
# This script helps developers who fork MarkGo to update import paths to their own repository

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ORIGINAL_MODULE="github.com/1mb-dev/markgo"
NEW_MODULE=""

show_help() {
    cat << EOF
MarkGo Import Path Updater

This script updates all import paths in the MarkGo codebase to point to your forked repository.

USAGE:
    ./scripts/update-imports.sh <your-username>
    ./scripts/update-imports.sh github.com/yourusername/markgo

EXAMPLES:
    ./scripts/update-imports.sh johnsmith
    ./scripts/update-imports.sh github.com/acme-corp/markgo
    ./scripts/update-imports.sh mycompany.com/projects/markgo

The script will:
1. Update all Go import statements
2. Update go.mod module declaration  
3. Update any references in documentation
4. Validate that the code compiles after changes

EOF
}

validate_input() {
    if [[ -z "$1" ]]; then
        echo -e "${RED}Error: No repository path provided${NC}"
        show_help
        exit 1
    fi
    
    # If just username provided, construct full github path
    if [[ "$1" != *"/"* ]]; then
        NEW_MODULE="github.com/$1/markgo"
    else
        NEW_MODULE="$1"
    fi
    
    # Validate module path format
    if [[ ! "$NEW_MODULE" =~ ^[a-zA-Z0-9][a-zA-Z0-9\-\.]*[a-zA-Z0-9]/[a-zA-Z0-9][a-zA-Z0-9\-_]*[a-zA-Z0-9]/[a-zA-Z0-9][a-zA-Z0-9\-_]*[a-zA-Z0-9]$ ]]; then
        echo -e "${YELLOW}Warning: Module path '$NEW_MODULE' may not be valid${NC}"
        echo -e "Expected format: domain.com/username/repository"
    fi
}

backup_files() {
    echo -e "${BLUE}📁 Creating backup...${NC}"
    
    if [[ -d "backup-$(date +%Y%m%d-%H%M%S)" ]]; then
        echo -e "${YELLOW}Backup directory already exists${NC}"
    else
        mkdir -p "backup-$(date +%Y%m%d-%H%M%S)"
        cp -r . "backup-$(date +%Y%m%d-%H%M%S)/" 2>/dev/null || true
        echo -e "${GREEN}✅ Backup created: backup-$(date +%Y%m%d-%H%M%S)/${NC}"
    fi
}

update_go_files() {
    echo -e "${BLUE}🔄 Updating Go import statements...${NC}"
    
    # Find and update all Go files
    local count=0
    while IFS= read -r -d '' file; do
        if grep -q "$ORIGINAL_MODULE" "$file"; then
            echo "  Updating: $file"
            sed -i.bak "s|$ORIGINAL_MODULE|$NEW_MODULE|g" "$file"
            rm "$file.bak"
            ((count++))
        fi
    done < <(find . -name "*.go" -type f -not -path "./backup-*" -print0)
    
    echo -e "${GREEN}✅ Updated $count Go files${NC}"
}

update_go_mod() {
    echo -e "${BLUE}🔄 Updating go.mod module declaration...${NC}"
    
    if [[ -f "go.mod" ]]; then
        if grep -q "$ORIGINAL_MODULE" "go.mod"; then
            echo "  Updating: go.mod"
            sed -i.bak "s|$ORIGINAL_MODULE|$NEW_MODULE|g" "go.mod"
            rm "go.mod.bak"
            echo -e "${GREEN}✅ Updated go.mod${NC}"
        else
            echo -e "${YELLOW}⚠️  go.mod doesn't contain original module path${NC}"
        fi
    else
        echo -e "${RED}❌ go.mod not found${NC}"
        exit 1
    fi
}

update_documentation() {
    echo -e "${BLUE}📚 Updating documentation...${NC}"
    
    local doc_files=("README.md" ".github/CONTRIBUTING.md" "docs/"*.md)
    local count=0
    
    for pattern in "${doc_files[@]}"; do
        for file in $pattern; do
            if [[ -f "$file" && "$file" != "./backup-"* ]]; then
                if grep -q "$ORIGINAL_MODULE" "$file" 2>/dev/null; then
                    echo "  Updating: $file"
                    sed -i.bak "s|$ORIGINAL_MODULE|$NEW_MODULE|g" "$file"
                    rm "$file.bak" 2>/dev/null || true
                    ((count++))
                fi
            fi
        done
    done
    
    echo -e "${GREEN}✅ Updated $count documentation files${NC}"
}

update_scripts() {
    echo -e "${BLUE}🔧 Updating scripts and configuration...${NC}"
    
    local script_files=("Makefile" "docker-compose.yml" "deployments/"*".yml" "scripts/"*".sh")
    local count=0
    
    for pattern in "${script_files[@]}"; do
        for file in $pattern; do
            if [[ -f "$file" && "$file" != *"/backup-"* ]]; then
                if grep -q "$ORIGINAL_MODULE" "$file" 2>/dev/null; then
                    echo "  Updating: $file"
                    sed -i.bak "s|$ORIGINAL_MODULE|$NEW_MODULE|g" "$file"
                    rm "$file.bak" 2>/dev/null || true
                    ((count++))
                fi
            fi
        done
    done
    
    echo -e "${GREEN}✅ Updated $count script files${NC}"
}

validate_compilation() {
    echo -e "${BLUE}🧪 Validating code compilation...${NC}"
    
    echo "  Running: go mod tidy"
    if go mod tidy; then
        echo -e "${GREEN}✅ go mod tidy successful${NC}"
    else
        echo -e "${RED}❌ go mod tidy failed${NC}"
        return 1
    fi
    
    echo "  Running: go build ./..."
    if go build ./...; then
        echo -e "${GREEN}✅ Build successful${NC}"
    else
        echo -e "${RED}❌ Build failed${NC}"
        return 1
    fi
    
    echo "  Running: go test -build -short ./..."
    if go test -run=^$ ./... >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Tests compile successfully${NC}"
    else
        echo -e "${YELLOW}⚠️  Some tests may have compilation issues${NC}"
    fi
}

show_summary() {
    echo ""
    echo -e "${GREEN}🎉 Import path update completed!${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BLUE}Changes made:${NC}"
    echo -e "  Original: ${RED}$ORIGINAL_MODULE${NC}"
    echo -e "  New:      ${GREEN}$NEW_MODULE${NC}"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo -e "  1. Review the changes: ${YELLOW}git diff${NC}"
    echo -e "  2. Test the application: ${YELLOW}make run${NC}"
    echo -e "  3. Run tests: ${YELLOW}make test${NC}"
    echo -e "  4. Commit changes: ${YELLOW}git add -A && git commit -m 'Update import paths to $NEW_MODULE'${NC}"
    echo ""
    echo -e "${YELLOW}💡 Tip: Keep the backup directory until you're confident everything works!${NC}"
}

# Main execution
main() {
    echo -e "${BLUE}🚀 MarkGo Import Path Updater${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Show help if requested
    if [[ "$1" == "-h" || "$1" == "--help" ]]; then
        show_help
        exit 0
    fi
    
    validate_input "$1"
    
    echo -e "${BLUE}Module path update:${NC}"
    echo -e "  From: ${RED}$ORIGINAL_MODULE${NC}"
    echo -e "  To:   ${GREEN}$NEW_MODULE${NC}"
    echo ""
    
    read -p "Continue with the update? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Update cancelled${NC}"
        exit 0
    fi
    
    # Perform updates
    backup_files
    update_go_files
    update_go_mod
    update_documentation
    update_scripts
    
    # Validate results
    if validate_compilation; then
        show_summary
        exit 0
    else
        echo -e "${RED}❌ Compilation failed. Please check the errors above.${NC}"
        echo -e "${YELLOW}💡 You can restore from the backup if needed.${NC}"
        exit 1
    fi
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi