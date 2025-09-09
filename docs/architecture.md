# MarkGo Engine v1.0.0 - Technical Architecture Guide

## 🏗️ Architectural Overview

MarkGo Engine follows a **layered architecture** with **clean separation of concerns**, implementing **Domain-Driven Design (DDD)** principles and **dependency injection** patterns for maximum maintainability and testability.

## 📐 Architecture Layers

### **Layer 1: Presentation Layer (HTTP)**
```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Presentation Layer                 │
├─────────────────────────────────────────────────────────────────┤
│  🌐 Gin Router Engine                                          │
│  ├── Static File Serving (/static/*, /favicon.ico)            │
│  ├── API Endpoints (/health, /metrics, /admin/*)              │
│  └── Dynamic Routes (/, /articles/*, /search, /contact)       │
├─────────────────────────────────────────────────────────────────┤
│  🛡️ Middleware Chain (13 Layers)                              │
│  ├── 1. RequestLoggingMiddleware (Structured logging)          │
│  ├── 2. RecoveryWithErrorHandler (Panic recovery)             │
│  ├── 3. Logger (Basic request logging)                        │
│  ├── 4. PerformanceMiddleware (Response time tracking)        │
│  ├── 5. CompetitorBenchmarkMiddleware (Performance comparison)│
│  ├── 6. SmartCacheHeaders (HTTP caching optimization)        │
│  ├── 7. CORS (Cross-origin resource sharing)                 │
│  ├── 8. Security (Security headers and protection)           │
│  ├── 9. SecurityLoggingMiddleware (Security event logging)   │
│  ├── 10. RateLimit (Request rate limiting)                   │
│  ├── 11. ValidationMiddleware (Input validation)             │
│  ├── 12. ErrorHandler (Centralized error handling)           │
│  └── 13. RequestTracker (Development request tracking)       │
└─────────────────────────────────────────────────────────────────┘
```

### **Layer 2: Application Layer (Handlers)**
```
┌─────────────────────────────────────────────────────────────────┐
│                        Application Layer                       │
├─────────────────────────────────────────────────────────────────┤
│  🎯 Handler Groups                                             │
│  ├── Public Handlers                                          │
│  │   ├── Home() → ArticleService.GetRecent()                 │
│  │   ├── Article() → ArticleService.GetBySlug()              │
│  │   ├── Articles() → ArticleService.GetPaginated()          │
│  │   ├── Search() → SearchService.Search()                   │
│  │   └── Contact() → EmailService.SendContactEmail()         │
│  ├── Admin Handlers                                           │
│  │   ├── AdminStats() → Multiple services for metrics       │
│  │   ├── ClearCache() → Cache.Clear()                        │
│  │   └── ReloadArticles() → ArticleService.Reload()          │
│  ├── Debug Handlers (Development only)                       │
│  │   ├── DebugMemory() → runtime.ReadMemStats()             │
│  │   ├── DebugRuntime() → runtime.NumGoroutine()            │
│  │   └── Pprof Integration → net/http/pprof                  │
│  └── Health & Metrics                                         │
│      ├── Health() → System health check                      │
│      └── Metrics() → Prometheus metrics                      │
└─────────────────────────────────────────────────────────────────┘
```

### **Layer 3: Business Logic Layer (Services)**
```
┌─────────────────────────────────────────────────────────────────┐
│                        Business Logic Layer                    │
├─────────────────────────────────────────────────────────────────┤
│  📝 ArticleService                                            │
│  ├── Core Operations                                          │
│  │   ├── LoadArticles() → File system scanning               │
│  │   ├── ParseArticle() → YAML + Markdown processing         │
│  │   ├── GetBySlug() → Cache-first lookup                   │
│  │   └── GetPaginated() → Sorted, filtered results          │
│  ├── Caching Strategy                                         │
│  │   ├── L1 Cache: Parsed articles (obcache-go)             │
│  │   ├── L2 Cache: Rendered HTML (template cache)            │
│  │   └── Invalidation: File modification time based         │
│  └── Performance Optimizations                               │
│      ├── String Interning: Tags, categories, authors        │
│      ├── Memory Pooling: Buffer reuse for parsing           │
│      └── Concurrent Loading: Goroutine-based file processing│
├─────────────────────────────────────────────────────────────────┤
│  🔍 SearchService                                             │
│  ├── Indexing Engine                                          │
│  │   ├── TF-IDF Scoring: Term frequency analysis             │
│  │   ├── Stop Word Filtering: Performance optimization       │
│  │   └── Content Tokenization: Unicode-aware processing     │
│  ├── Query Processing                                         │
│  │   ├── Query Parsing: Operator and phrase detection       │
│  │   ├── Result Ranking: Relevance scoring algorithm        │
│  │   └── Result Caching: Query + content hash keys          │
│  └── Background Tasks                                         │
│      ├── Index Updates: Scheduled via goflow                │
│      └── Cache Warming: Popular query pre-computation       │
├─────────────────────────────────────────────────────────────────┤
│  🎨 TemplateService                                           │
│  ├── Template Management                                      │
│  │   ├── Template Loading: Glob pattern parsing             │
│  │   ├── Function Registration: 30+ custom functions        │
│  │   └── Hot Reload: File system watching (development)     │
│  ├── Rendering Pipeline                                       │
│  │   ├── Context Preparation: Data structure marshaling     │
│  │   ├── Function Execution: Custom template functions      │
│  │   └── Output Generation: HTML generation with caching    │
│  └── Performance Features                                     │
│      ├── Render Caching: Content-based cache keys          │
│      ├── Memory Pooling: Buffer reuse for rendering         │
│      └── Timezone Caching: Location object reuse           │
├─────────────────────────────────────────────────────────────────┤
│  📧 EmailService                                              │
│  ├── Contact Form Processing                                  │
│  │   ├── Input Validation: Sanitization and format checking │
│  │   ├── Spam Protection: Rate limiting and content analysis│
│  │   └── Template Rendering: HTML and text email generation │
│  ├── SMTP Integration                                         │
│  │   ├── Connection Pooling: Reusable SMTP connections      │
│  │   ├── Error Handling: Retry logic and failure logging    │
│  │   └── Configuration: Environment-driven SMTP settings    │
│  └── Background Tasks                                         │
│      └── Cleanup Tasks: Log rotation and temporary file cleanup│
└─────────────────────────────────────────────────────────────────┘
```

### **Layer 4: Infrastructure Layer**
```
┌─────────────────────────────────────────────────────────────────┐
│                       Infrastructure Layer                     │
├─────────────────────────────────────────────────────────────────┤
│  💾 Caching Subsystem (obcache-go)                           │
│  ├── Cache Architecture                                       │
│  │   ├── L1: In-memory LRU cache (hot data, <1ms access)   │
│  │   ├── L2: Compressed memory (warm data, <5ms access)     │
│  │   └── L3: Disk persistence (cold data, <50ms access)    │
│  ├── Cache Strategies                                         │
│  │   ├── Articles: Content hash + TTL invalidation          │
│  │   ├── Templates: File modification time based            │
│  │   ├── Search Results: Query + index version hash         │
│  │   └── Static Assets: Long-term caching (1 year)         │
│  └── Performance Metrics                                      │
│      ├── Hit Ratio: >95% for typical workloads              │
│      ├── Eviction Policy: LRU with size and TTL limits      │
│      └── Memory Management: Zero-allocation design          │
├─────────────────────────────────────────────────────────────────┤
│  🔄 Task Scheduling (goflow)                                  │
│  ├── Scheduled Tasks                                          │
│  │   ├── Cache Warming: "0 */30 * * * *" (every 30 min)    │
│  │   ├── Cache Cleanup: "0 0 * * * *" (hourly)             │
│  │   ├── Template Reload: File system events               │
│  │   └── Email Cleanup: "0 */10 * * * *" (every 10 min)   │
│  ├── Worker Pool Management                                   │
│  │   ├── Concurrent Execution: Goroutine-based workers      │
│  │   ├── Task Queue: Priority-based task scheduling        │
│  │   └── Error Handling: Retry logic and failure logging   │
│  └── Background Operations                                    │
│      ├── Index Updates: Non-blocking search index rebuild   │
│      ├── Cache Preloading: Popular content prefetching      │
│      └── System Maintenance: Log rotation and cleanup       │
├─────────────────────────────────────────────────────────────────┤
│  🗄️ Memory Management                                         │
│  ├── Object Pooling                                           │
│  │   ├── StringBuilderPool: String concatenation operations │
│  │   ├── SlicePool: Dynamic array operations               │
│  │   ├── RuneSlicePool: Unicode text processing             │
│  │   └── BufferPool: I/O operation buffers                 │
│  ├── String Interning                                         │
│  │   ├── Global Interner: Shared string deduplication      │
│  │   ├── Tag Interning: Article tag deduplication          │
│  │   └── Statistics: Hit/miss ratio and memory savings     │
│  └── Garbage Collection Optimization                         │
│      ├── Allocation Reduction: Pool reuse minimizes GC      │
│      ├── Memory Reuse: Buffer recycling for common ops     │
│      └── String Deduplication: Reduced memory footprint    │
└─────────────────────────────────────────────────────────────────┘
```

## 🔄 Data Flow Patterns

### **Request Processing Flow**
```
┌─────────────────────────────────────────────────────────────────┐
│                       Request Flow Diagram                     │
└─────────────────────────────────────────────────────────────────┘

[HTTP Request] → [Nginx Proxy (Optional)]
        ↓
[Gin Router] → [Route Matching]
        ↓
[Middleware Chain Execution]
├── Security Validation → [Continue/Reject]
├── Rate Limiting → [Allow/Block]  
├── Input Validation → [Parse/Error]
└── Performance Tracking → [Metrics Collection]
        ↓
[Handler Execution]
├── [Public Handler] → [Service Layer Call]
├── [Admin Handler] → [Auth Check] → [Service Layer Call]
└── [Debug Handler] → [Dev Environment Check] → [System Info]
        ↓
[Service Layer Processing]
├── [Cache Check] → [Hit: Return Cached] or [Miss: Continue]
├── [Business Logic] → [Data Processing/Transformation]
├── [External Calls] → [File System/Template Engine/Search]
└── [Cache Update] → [Store Result for Future Use]
        ↓
[Response Generation]
├── [Template Rendering] → [HTML Generation]
├── [JSON Serialization] → [API Response]  
├── [Static File Serving] → [Direct File Response]
└── [Error Response] → [Structured Error JSON/HTML]
        ↓
[Response Middleware]
├── [Cache Headers] → [HTTP Caching Instructions]
├── [Security Headers] → [XSS, CSRF, HSTS Protection]
├── [Compression] → [Gzip/Deflate Encoding]
└── [Logging] → [Request Metrics and Access Logs]
        ↓
[HTTP Response] → [Client]
```

### **Article Processing Pipeline**
```
┌─────────────────────────────────────────────────────────────────┐
│                   Article Processing Pipeline                  │
└─────────────────────────────────────────────────────────────────┘

[Markdown File] (.md)
        ↓
[File System Scan] → [File Modification Time Check]
        ↓
[Content Reading] → [UTF-8 Validation] → [Size Limits Check]  
        ↓
[YAML Frontmatter Parsing]
├── [Title] → [String Interning] → [Validation]
├── [Date] → [Time Parsing] → [Timezone Handling] 
├── [Tags] → [String Interning] → [Deduplication]
├── [Categories] → [String Interning] → [Hierarchical Structure]
├── [Author] → [String Interning] → [Default Assignment]
└── [Metadata] → [Custom Field Processing]
        ↓
[Markdown Content Processing]
├── [Goldmark Parser] → [CommonMark + Extensions]
├── [Code Highlighting] → [Syntax Detection]
├── [Link Processing] → [Internal Link Resolution]  
├── [Image Processing] → [Path Resolution + Alt Text]
└── [HTML Generation] → [Security Sanitization]
        ↓
[Article Object Creation]
├── [Slug Generation] → [URL-safe identifier]
├── [Excerpt Generation] → [First N characters/words]
├── [Reading Time Calculation] → [Word count/200 WPM]
├── [Search Index Entry] → [Full-text indexing data]
└── [Cache Key Generation] → [Content hash + timestamp]
        ↓
[Storage & Caching]
├── [Memory Cache] → [obcache-go storage]
├── [Search Index] → [TF-IDF scoring update]
├── [Template Cache] → [Rendered HTML caching]
└── [File System Cache] → [Processed content backup]
```

### **Search Query Processing**
```
┌─────────────────────────────────────────────────────────────────┐
│                  Search Query Processing Flow                  │
└─────────────────────────────────────────────────────────────────┘

[Search Query] (User Input)
        ↓
[Input Validation]
├── [Length Limits] → [Min: 2 chars, Max: 100 chars]
├── [Character Filtering] → [Alphanumeric + spaces only]
├── [XSS Protection] → [HTML entity encoding]
└── [Rate Limiting] → [Query frequency limits]
        ↓
[Query Processing]
├── [Normalization] → [Lowercase, Unicode normalization]
├── [Tokenization] → [Word boundary detection]
├── [Stop Word Removal] → ["the", "and", "or", etc.]
├── [Stemming] → [Root word extraction]
└── [Operator Detection] → [AND, OR, NOT, phrase queries]
        ↓
[Cache Lookup]
├── [Cache Key Generation] → [Query hash + index version]
├── [Cache Hit] → [Return Cached Results] → [Response]
└── [Cache Miss] → [Continue to Index Search]
        ↓
[Index Search Execution] 
├── [Term Lookup] → [Inverted index consultation]
├── [Document Retrieval] → [Matching article identification]
├── [TF-IDF Scoring] → [Relevance score calculation]
├── [Ranking Algorithm] → [Result sorting by relevance]
└── [Result Filtering] → [Published status, permissions]
        ↓
[Result Processing]
├── [Excerpt Generation] → [Highlighted snippets]
├── [URL Generation] → [Article slug to URL mapping]
├── [Metadata Addition] → [Date, author, categories]
├── [Pagination] → [Result set chunking]
└── [JSON Serialization] → [API response format]
        ↓
[Cache Storage] → [Store Results for Future Queries]
        ↓
[Response Delivery] → [JSON API Response to Client]
```

## 🏛️ Design Patterns Implementation

### **1. Dependency Injection Pattern**
```go
// Service dependencies injected via constructor
type Handlers struct {
    articleService  ArticleServiceInterface
    searchService   SearchServiceInterface  
    emailService    EmailServiceInterface
    templateService TemplateServiceInterface
    cache          *obcache.Cache
    config         *config.Config
    logger         *slog.Logger
}

// Interface-based design for testability
type ArticleServiceInterface interface {
    GetBySlug(slug string) (*models.Article, error)
    GetRecent(limit int) ([]*models.Article, error)
    GetPaginated(page, perPage int) (*models.PaginatedArticles, error)
}
```

### **2. Repository Pattern**
```go
// Article repository abstraction
type ArticleRepository interface {
    LoadAll() ([]*models.Article, error)
    GetBySlug(slug string) (*models.Article, error)
    Watch(callback func()) error
}

// File system implementation
type FileSystemRepository struct {
    basePath string
    watcher  *fsnotify.Watcher
}
```

### **3. Strategy Pattern for Caching**
```go
// Different caching strategies based on content type
type CacheStrategy interface {
    GenerateKey(data interface{}) string
    GetTTL() time.Duration
    ShouldCache(data interface{}) bool
}

type ArticleCacheStrategy struct{}
type SearchCacheStrategy struct{}
type TemplateCacheStrategy struct{}
```

### **4. Observer Pattern for File Changes**
```go
// File system watching for hot reload
type FileWatcher struct {
    observers []FileChangeObserver
}

type FileChangeObserver interface {
    OnFileChanged(path string, event fsnotify.Event)
}

// Template service implements observer
func (t *TemplateService) OnFileChanged(path string, event fsnotify.Event) {
    if filepath.Ext(path) == ".html" {
        t.Reload(t.templatesPath)
    }
}
```

### **5. Factory Pattern for Services**
```go
// Service factory for dependency management
type ServiceFactory struct {
    config *config.Config
    logger *slog.Logger
    cache  *obcache.Cache
}

func (f *ServiceFactory) CreateArticleService() (ArticleServiceInterface, error) {
    return services.NewArticleService(f.config.ArticlesPath, f.logger)
}
```

## 🔧 Configuration Management

### **Configuration Hierarchy**
```
Configuration Sources (Priority Order):
1. Command Line Arguments (highest priority)
2. Environment Variables  
3. Configuration Files (.env, config.yaml)
4. Default Values (lowest priority)

Configuration Categories:
├── Server Configuration (ports, timeouts, TLS)
├── Application Configuration (blog settings, features)
├── Performance Configuration (cache, pools, limits)
├── Security Configuration (rate limits, CORS, auth)  
├── Logging Configuration (levels, formats, outputs)
└── Integration Configuration (email, external services)
```

### **Environment Variable Mapping**
```go
// Automatic environment variable binding
type Config struct {
    Environment    string `env:"ENVIRONMENT" default:"development"`
    Port          int    `env:"PORT" default:"3000"`
    BaseURL       string `env:"BASE_URL" default:"http://localhost:3000"`
    
    Blog struct {
        Title       string `env:"BLOG_TITLE" default:"MarkGo Blog"`
        Description string `env:"BLOG_DESCRIPTION" default:""`
        Author      string `env:"BLOG_AUTHOR" default:""`
    }
    
    Cache struct {
        Enabled     bool          `env:"CACHE_ENABLED" default:"true"`
        TTL         time.Duration `env:"CACHE_TTL" default:"1h"`
        MaxSize     int          `env:"CACHE_MAX_SIZE" default:"1000"`
    }
}
```

## 🧪 Testing Architecture

### **Testing Pyramid Structure**
```
                    ┌─────────────┐
                   │   E2E Tests  │  ← 5% (Integration tests)
                  └─────────────┘
                 ┌─────────────────┐
                │  Service Tests   │  ← 25% (Component tests)
               └─────────────────┘  
              ┌─────────────────────┐
             │    Unit Tests        │  ← 70% (Function/method tests)
            └─────────────────────┘

Testing Coverage by Layer:
├── Unit Tests (70%): Individual functions, methods, utilities
├── Service Tests (25%): Service layer integration, mocking
├── Handler Tests (15%): HTTP endpoint testing with test server
├── Integration Tests (5%): Full application testing
└── Performance Tests: Benchmarking and load testing
```

### **Test Organization**
```go
// Test file naming convention
article_service_test.go     // Unit tests for ArticleService
handlers_test.go           // Integration tests for handlers
benchmark_test.go          // Performance benchmarks
integration_test.go        // End-to-end integration tests

// Test helper utilities
test_helpers.go            // Common test setup and utilities
mock_services.go           // Mock implementations for testing
fixtures/                  // Test data and fixtures
```

## 📊 Performance Monitoring

### **Metrics Collection Points**
```go
// Application-level metrics
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds", 
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    
    cacheOperations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_operations_total",
            Help: "Total cache operations",
        },
        []string{"operation", "result"},
    )
)
```

### **Performance Benchmarks**
```go
// Benchmark suite for critical paths
func BenchmarkArticleService_GetBySlug(b *testing.B) {
    service := setupArticleService()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := service.GetBySlug("test-article")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkSearchService_Search(b *testing.B) {
    service := setupSearchService()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := service.Search(articles, "golang performance", 10)
        if err != nil {
            b.Fatal(err) 
        }
    }
}
```

---

**MarkGo Engine v1.0.0** implements a **clean, layered architecture** with **enterprise-grade patterns** and **performance optimizations**, delivering **exceptional performance** in a **maintainable, testable codebase**. 🏗️⚡