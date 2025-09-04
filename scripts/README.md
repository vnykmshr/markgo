# Blog Optimization Scripts

This directory contains automation scripts for optimizing tags and categories across the markgo platform. These tools help consolidate redundant tags, standardize naming conventions, and create a more maintainable content taxonomy.

## Overview

The blog optimization project addresses several key issues:

- **Tag Redundancy**: Multiple tags for same concepts (golang/go/go-programming)
- **Category Inconsistency**: Case sensitivity and naming issues (Technology vs technology)
- **Over-granularity**: Single-use tags that don't provide meaningful classification
- **Semantic Overlap**: Different tags representing similar concepts

## Target Improvements

### Categories: From 30+ to 8 Core Categories

| New Category | Consolidates | Description |
|--------------|-------------|-------------|
| `technology` | technology, Technology, General | Technical articles, tutorials, architecture |
| `personal` | personal, Personal, personal-reflection, thoughts, musings | Personal reflections, life experiences |
| `career` | career, Career, leadership | Career advice, professional growth, leadership |
| `music` | music | Music reviews, artist spotlights, musical reflections |
| `cricket` | cricket | Cricket analysis, player tributes, sports commentary |
| `philosophy` | philosophy, reflection | Philosophical thoughts, existential reflections |
| `tools` | tools, development-tools, devops, infrastructure | Tool reviews, setup guides, infrastructure |
| `culture` | culture, travel, entrepreneurship, community | Cultural observations, travel, community building |

### Tags: From 200+ to ~80 Meaningful Tags

#### Core Technology Tags
- **Languages**: `golang`, `nodejs`, `javascript`
- **Architecture**: `distributed-systems`, `system-architecture`, `backend`, `scalability`
- **Data**: `database`, `redis`, `data-analysis`
- **Infrastructure**: `infrastructure`, `nginx`, `linux`, `automation`

#### Development Process Tags
- **Practices**: `best-practices`, `version-control`, `tools-review`
- **Content Types**: `tutorial`, `case-study`
- **Specialized**: `search`, `networking`, `embedded-systems`, `algorithms`

#### Personal & Career Tags
- **Career**: `career-journey`, `personal-growth`, `leadership`, `entrepreneurship`
- **Content**: `reflection`, `philosophy`, `life-lessons`, `interviews`

#### Cultural & Interest Tags
- **Music**: `music-analysis`, `indian-rock`, `classic-rock`, `music-events`
- **Sports**: `cricket-analysis`, `sachin-tendulkar`
- **Geography**: `nepal`, `india`, `travel`
- **Tech Trends**: `future-tech`, `open-source`

## Scripts

### 1. analyze_tags_categories.py

**Purpose**: Analyze current state of tags and categories across all articles.

**Usage**:
```bash
# Basic analysis
python scripts/analyze_tags_categories.py

# Verbose analysis with minimum usage threshold
python scripts/analyze_tags_categories.py --verbose --min-count 2

# Export to JSON for further processing
python scripts/analyze_tags_categories.py --output json --output-file analysis.json

# Generate CSV report
python scripts/analyze_tags_categories.py --output csv --output-file report.csv

# Sort alphabetically instead of by frequency
python scripts/analyze_tags_categories.py --sort alphabetical
```

**Options**:
- `--articles`: Path to articles directory (default: `../articles`)
- `--output`: Format - `text`, `json`, `csv` (default: `text`)
- `--sort`: Sort order - `frequency`, `alphabetical` (default: `frequency`)
- `--min-count`: Minimum usage count to include (default: `1`)
- `--output-file`: Save to file instead of stdout

**What it analyzes**:
- Total tag and category counts
- Usage frequency distributions
- Singleton tags/categories (used only once)
- Similar tags that might be consolidated
- Tag co-occurrence patterns
- Articles with most/least tags
- Potential redundancies

### 2. migrate_tags_categories.py

**Purpose**: Automatically migrate and consolidate tags and categories according to the optimization plan.

**Usage**:
```bash
# Preview changes without modifying files
python scripts/migrate_tags_categories.py --dry-run --verbose

# Apply changes with backup
python scripts/migrate_tags_categories.py --backup --verbose

# Apply changes to specific directory
python scripts/migrate_tags_categories.py --articles /path/to/articles --backup
```

**Options**:
- `--articles`: Path to articles directory (default: `../articles`)
- `--dry-run`: Preview changes without modifying files
- `--backup`: Create backup files before modification
- `--verbose`: Enable detailed output

**What it does**:
- Parses YAML frontmatter from all markdown files
- Applies category consolidation mappings
- Applies tag consolidation and removal rules
- Maintains hyphenated tag format (e.g., `distributed-systems`, `system-architecture`)
- Outputs tags and categories as CSV within square brackets: `[tag1, tag2, tag3]`
- Preserves date fields exactly as-is to prevent Go parsing errors
- Updates frontmatter with new tags/categories
- Provides detailed migration statistics

**Migration Rules**:

*Categories*:
```
technology, Technology, General → technology
personal, Personal, thoughts, musings → personal
career, Career, leadership → career
tools, devops, infrastructure → tools
culture, travel, entrepreneurship → culture
```

*High-Impact Tag Consolidations*:
```
golang, go, go-programming → golang
nodejs, node-js → nodejs
reflection, reflections, personal-reflection → reflection
career, career-journey, career-transition → career-journey
distributed-systems, microservices, circuit-breaker → distributed-systems
```

**Output Format**: Tags and categories are output as CSV within square brackets:
```yaml
# Before (multiple tags with redundancy)
tags: [golang, distributed-systems, circuit-breaker, resilience-patterns, microservices]
categories: [technology, golang]

# After (consolidated with CSV in brackets)
tags: [golang, distributed-systems, system-architecture]
categories: [technology]
```

## Prerequisites

**Python Dependencies**:
```bash
pip install pyyaml
```

**Required Python Version**: 3.6+

## Workflow

### Step 1: Analyze Current State
```bash
cd scripts
python analyze_tags_categories.py --verbose --output-file current-state.txt
```

Review the analysis to understand:
- Current tag/category distribution
- Singleton and redundant tags
- Most frequently used tags
- Potential consolidation opportunities

### Step 2: Preview Migration
```bash
python migrate_tags_categories.py --dry-run --verbose
```

This shows exactly what changes would be made without modifying any files.

### Step 3: Apply Migration with Backup
```bash
python migrate_tags_categories.py --backup --verbose
```

This applies all changes while creating `.backup` files for safety.

### Step 4: Validate Results
```bash
python analyze_tags_categories.py --verbose --output-file post-migration.txt
```

Compare before/after statistics to validate the optimization.

### Step 5: Clean Up (Optional)
```bash
# Remove backup files if satisfied with results
find ../articles -name "*.backup" -delete
```

## Expected Results

### Quantitative Improvements
- **60% reduction** in total tags (200+ → ~80)
- **75% reduction** in categories (30+ → 8)
- **100% standardized** naming conventions (consistent hyphenated format)
- **Square bracket CSV format** for clean YAML structure
- **Eliminated redundancy** across tag taxonomy

### Qualitative Benefits
- **Better Discoverability**: Clearer content organization
- **Improved Maintainability**: Easier to assign tags to new content
- **Enhanced User Experience**: More intuitive navigation and filtering
- **Better SEO**: Improved content clustering and topical authority
- **Consistent Taxonomy**: Standardized naming across all content

## Troubleshooting

### Common Issues

**1. YAML Parsing Errors**
```
Error parsing YAML frontmatter: ...
```
- Check for malformed YAML in article frontmatter
- Ensure proper indentation and syntax
- Look for special characters that need escaping

**2. Go Date Parsing Errors**
```
failed to parse front matter: parsing time "..." as "2006-01-02T15:04:05Z07:00": cannot parse...
```
- The migration script preserves date fields exactly as-is
- Ensure dates use RFC3339 format with 'T' separator: `2024-12-15T10:00:00Z`
- Avoid space separators: `2024-12-15 10:00:00Z` (will cause Go parsing errors)

**3. Permission Errors**
```
Permission denied: ...
```
- Ensure write permissions on articles directory
- Run with appropriate user privileges
- Check file locks from editors

**4. Encoding Issues**
```
UnicodeDecodeError: ...
```
- Ensure all markdown files are UTF-8 encoded
- Check for binary files in articles directory

### Validation Checks

**Before Migration**:
- Backup important files manually
- Run analysis to understand current state
- Test with `--dry-run` first
- Verify all dates use RFC3339 format with 'T' separator

**After Migration**:
- Compare before/after analysis reports
- Verify blog functionality with new tags/categories
- Test search and filtering features
- Check that all articles load without date parsing errors
- Update any hardcoded references in templates

## Maintenance Guidelines

### New Content Tagging
1. **Maximum 5 tags per article** to maintain focus
2. **Use 1-2 categories maximum** from the 8 core categories
3. **Prefer existing tags** over creating new ones
4. **Use broader tags** unless specific detail adds value
5. **Use hyphens** for multi-word tags (e.g., `distributed-systems`)
6. **Format as CSV within square brackets** in YAML frontmatter

### Periodic Review
- **Quarterly**: Review tag usage analytics
- **Annually**: Consolidate underused tags
- **As needed**: Content audit for proper classification

## File Structure

```
scripts/
├── README.md                    # This file
├── analyze_tags_categories.py   # Analysis tool
└── migrate_tags_categories.py   # Migration tool

../
├── tag-category-optimization-plan.md  # Detailed optimization strategy
└── articles/                    # Target directory for optimization
    ├── *.markdown              # Article files to optimize
    └── *.md                     # Alternative markdown extension
```

## Support

For issues or questions about these scripts:

1. **Check the logs**: Both scripts provide verbose output for debugging
2. **Validate input**: Ensure articles directory contains valid markdown files
3. **Test incrementally**: Use dry-run mode to preview changes
4. **Backup first**: Always create backups before bulk modifications

The optimization process is designed to be safe and reversible, with special care taken to preserve date fields exactly as they appear in the original files. Proper testing and backups are recommended for production use.

## Date Field Preservation

The migration script includes special handling for date fields to prevent Go parsing errors:

- **Preserves exact format**: Date fields are kept as-is without reformatting
- **No datetime conversion**: Avoids converting dates to datetime objects that might change format
- **RFC3339 compatible**: Works with Go's expected `2006-01-02T15:04:05Z07:00` format
- **Error prevention**: Prevents "cannot parse" errors in Go applications

**Supported date formats**:
- ✅ `2024-12-15T10:00:00Z` (RFC3339 UTC)
- ✅ `2010-12-12T03:24:03+00:00` (RFC3339 with timezone)
- ❌ `2010-12-12 03:24:03+00:00` (space separator - causes Go errors)

Run `python test_date_preservation.py` to validate date handling.
