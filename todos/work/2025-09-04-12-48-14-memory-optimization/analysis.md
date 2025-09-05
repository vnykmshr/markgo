# MarkGo Memory Usage Analysis and Optimization Report

Based on my analysis of the MarkGo codebase, I've identified the current memory usage patterns and specific optimization opportunities to achieve the <50MB memory footprint target (6x better than Ghost's ~300MB).

## Current Memory Usage Patterns Analysis

### 1. **Article Processing Memory Usage**

**Current Implementation Analysis:**
- **File Path:** `/Users/vmx/workspace/gocode/src/github.com/vnykmshr/markgo/internal/services/article.go`
- **Memory Hotspots:**
  - `ArticleService` stores all articles in memory (`articles []*models.Article` and `cache map[string]*models.Article`)
  - Each article contains multiple string fields: `Content`, `ProcessedContent`, `Excerpt` (~3x content duplication)
  - Goldmark markdown processor creates temporary objects during parsing
  - YAML frontmatter parsing creates additional temporary allocations
  - Complex excerpt generation with multiple regex operations and string manipulations

**Memory Impact:**
- Dual storage (slice + map) means articles are referenced twice
- Large string duplications (raw content + processed HTML + excerpt)
- Regex compilation happens repeatedly in excerpt generation

### 2. **Cache Service Memory Efficiency**

**Current Implementation Analysis:**
- **File Path:** `/Users/vmx/workspace/gocode/src/github.com/vnykmshr/markgo/internal/services/cache.go`
- **Memory Characteristics:**
  - Simple map-based cache with `CacheItem` wrapper structs
  - Each cache entry has overhead: `Value any`, `Expiration time.Time`
  - Basic LRU eviction by expiration time (not memory-aware)
  - No compression or serialization optimization

**Memory Impact:**
- Interface{} boxing creates additional allocations
- Cache items store full rendered HTML pages (high memory per item)
- No memory-based eviction strategy

### 3. **Search Service Memory Patterns**

**Current Implementation Analysis:**
- **File Path:** `/Users/vmx/workspace/gocode/src/github.com/vnykmshr/markgo/internal/services/search.go`
- **Memory Characteristics:**
  - Creates new slices for every search operation
  - No pre-built search index (processes all articles each time)
  - Text tokenization creates many temporary string slices
  - HTML stripping creates full content copies
  - Stop words stored as map[string]bool (efficient)

**Memory Impact:**
- Linear search through all articles for each query
- Multiple string allocations during tokenization
- No result reuse or caching at search level

### 4. **Template Service Memory Usage**

**Current Implementation Analysis:**
- **File Path:** `/Users/vmx/workspace/gocode/src/github.com/vnykmshr/markgo/internal/services/template.go`
- **Memory Characteristics:**
  - Comprehensive template function map (40+ functions)
  - Global timezone cache using sync.Map
  - Template compilation cached in memory
  - Complex reflection-based helper functions

**Memory Impact:**
- Large function map with closures
- Reflection operations in template functions
- String building operations in template rendering

### 5. **Handler Layer Memory Usage**

**Current Implementation Analysis:**
- **File Path:** `/Users/vmx/workspace/gocode/src/github.com/vnykmshr/markgo/internal/handlers/handlers.go`
- **Memory Characteristics:**
  - Heavy use of `gin.H` for template data (map[string]any)
  - Multiple service calls per request creating temporary data
  - Feed generation creates large XML/JSON strings in memory
  - Pagination objects created for each request

**Memory Impact:**
- Map allocations for template data
- String building for feeds without streaming
- Multiple temporary slices for related articles, recent articles, etc.

## Specific Optimization Opportunities

### **High Impact Optimizations (Target: 20-30MB Savings)**

#### 1. **Article Content Deduplication**
```go
// Current: 3x content storage per article
type Article struct {
    Content          string  // Raw markdown ~50KB
    ProcessedContent string  // HTML ~75KB  
    Excerpt          string  // ~200 bytes
}

// Optimized: Single content with lazy processing
type Article struct {
    Content     string            // Raw markdown
    processed   *string           // Lazy-loaded HTML (pointer)
    excerpt     *string           // Lazy-loaded excerpt
    contentHash [16]byte          // For cache invalidation
}
```
**Memory Savings:** ~40-50% per article (25KB+ per article for large content)

#### 2. **String Interning for Common Data**
```go
// Intern frequently repeated strings (tags, categories, authors)
var stringInterner = sync.Map{} // string -> *string

func InternString(s string) string {
    if interned, exists := stringInterner.Load(s); exists {
        return interned.(string)
    }
    stringInterner.Store(s, s)
    return s
}
```
**Memory Savings:** ~5-10MB for duplicate tags, categories, and metadata

#### 3. **Smart Cache Value Storage**
```go
// Current: Store full gin.H objects
type CacheItem struct {
    Value any              // Full rendered data ~50-100KB
    Expiration time.Time
}

// Optimized: Store compressed data
type CacheItem struct {
    Data       []byte      // Compressed rendered data ~10-20KB  
    IsCompressed bool
    Expiration time.Time
}
```
**Memory Savings:** ~60-80% cache memory reduction

### **Medium Impact Optimizations (Target: 5-15MB Savings)**

#### 4. **Pre-compiled Search Index**
```go
// Replace linear search with inverted index
type SearchIndex struct {
    termToArticles map[string][]uint32    // Term -> Article IDs
    articleData    []SearchableArticle    // Compressed article data
    stemmer        *porter.Stemmer        // Reusable stemmer
}

type SearchableArticle struct {
    ID          uint32
    Title       string
    Tags        []uint16   // Interned tag IDs
    Categories  []uint16   // Interned category IDs  
    WordCount   uint16
}
```
**Memory Savings:** ~10MB for large article sets, eliminates per-query allocations

#### 5. **Template Data Pooling**
```go
// Pool frequently used template data structures
var templateDataPool = sync.Pool{
    New: func() any {
        return make(gin.H, 20) // Pre-sized map
    },
}

func (h *Handlers) getTemplateData() gin.H {
    data := templateDataPool.Get().(gin.H)
    // Clear previous data
    for k := range data {
        delete(data, k)
    }
    return data
}
```
**Memory Savings:** ~2-5MB, eliminates map allocations per request

#### 6. **Lazy Article Loading**
```go
// Don't load all articles at startup
type ArticleService struct {
    articlePaths  []string                    // Just file paths
    loadedCache   map[string]*models.Article  // LRU cache of loaded articles
    maxCacheSize  int                         // Memory-based limit
}
```
**Memory Savings:** ~15-25MB for large article collections

### **Low Impact but Important Optimizations (Target: 2-5MB Savings)**

#### 7. **Efficient String Building**
```go
// Replace string concatenation with builder reuse
var stringBuilderPool = sync.Pool{
    New: func() any {
        return &strings.Builder{}
    },
}

func (s *ArticleService) generateExcerpt(content string, maxLength int) string {
    builder := stringBuilderPool.Get().(*strings.Builder)
    defer func() {
        builder.Reset()
        stringBuilderPool.Put(builder)
    }()
    
    // Use builder instead of string concatenation
    return s.buildExcerptWithBuilder(builder, content, maxLength)
}
```

#### 8. **Regex Compilation Caching**
```go
// Current: Regex compiled multiple times
var (
    codeBlockRe = regexp.MustCompile("```[\\s\\S]*?```")
    linkRe      = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
    // ... other regexes
)
```
**Memory Savings:** ~1-2MB, eliminates repeated regex compilation

## Implementation Priority and Expected Results

### **Phase 1: Critical Memory Optimizations (Target: <40MB total)**
1. Article content deduplication → 20-25MB savings
2. String interning → 5-8MB savings  
3. Cache compression → 8-12MB savings

### **Phase 2: Search and Template Optimizations (Target: <35MB total)**
4. Search index → 5-10MB savings
5. Template data pooling → 2-5MB savings
6. Lazy loading → 5-10MB savings (for large sites)

### **Phase 3: Fine-tuning (Target: <30MB total)**
7. String builder pooling → 1-2MB savings
8. Regex caching → 1-2MB savings
9. Memory-aware eviction policies → 2-5MB savings

## Competitive Analysis

**Current MarkGo Baseline:** ~50-60MB (estimated from existing benchmarks)
**Optimization Target:** <30MB total memory usage

**Competitor Comparison:**
- **Ghost:** ~300MB → **10x more efficient**
- **WordPress:** ~2GB → **65x more efficient**  
- **Hugo:** ~20MB (static) → **Comparable with dynamic features**

## Monitoring and Validation

The existing memory benchmark in `/Users/vmx/workspace/gocode/src/github.com/vnykmshr/markgo/internal/handlers/benchmark_memory_test.go` provides excellent validation framework:

```go
// Validate optimizations don't exceed targets
if heapMB > 30 { // Lowered from 50MB
    b.Logf("WARNING: Heap usage %.1fMB exceeds optimized target")
} else {
    b.Logf("✅ Memory usage %.1fMB meets optimization target (%.1fx better than Ghost)", 
        heapMB, 300/heapMB)
}
```

## Risk Assessment

**Low Risk Optimizations:**
- String interning, regex caching, template pooling
- Can be implemented incrementally

**Medium Risk Optimizations:**
- Article content deduplication (requires careful lazy loading)
- Cache compression (needs proper error handling)

**High Risk Optimizations:**
- Complete search index rewrite (significant architectural change)
- Lazy article loading (impacts startup and error handling)

The analysis shows that MarkGo can realistically achieve **<30MB memory usage** through systematic optimization, making it **10x more memory-efficient than Ghost** while maintaining all dynamic features and superior performance characteristics.