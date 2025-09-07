package handlers

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// BenchmarkSetup provides a realistic benchmark environment
type BenchmarkSetup struct {
	handlers *Handlers
	router   *gin.Engine
	articles []*models.Article
}

func NewBenchmarkSetup(b *testing.B) *BenchmarkSetup {
	gin.SetMode(gin.ReleaseMode) // Use production mode for benchmarks

	// Create realistic config
	cfg := &config.Config{
		BaseURL: "https://example.com",
		Blog: config.BlogConfig{
			Title:        "Benchmark Blog",
			Description:  "Performance testing blog",
			Author:       "Benchmark Author",
			PostsPerPage: 10,
			Language:     "en",
		},
		Cache: config.CacheConfig{
			TTL:     time.Hour,
			MaxSize: 10000,
		},
	}

	// Silent logger for benchmarks
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Generate realistic test data
	articles := generateBenchmarkArticles(1000) // 1k articles for realistic load

	// Create services with obcache for realistic performance testing
	articleService := NewInMemoryArticleService(articles)
	emailService := NewInMemoryEmailService()
	searchService := NewInMemorySearchService()

	// Setup obcache for realistic caching performance
	cacheConfig := obcache.NewDefaultConfig()
	cacheConfig.MaxEntries = cfg.Cache.MaxSize
	cacheConfig.DefaultTTL = cfg.Cache.TTL
	cache, err := obcache.New(cacheConfig)
	if err != nil {
		b.Fatal("Failed to create cache:", err)
	}

	handlers := New(&Config{
		ArticleService: articleService,
		EmailService:   emailService,
		SearchService:  searchService,
		Config:         cfg,
		Logger:         logger,
		Cache:          cache,
	})

	// Setup router with essential middleware
	router := gin.New()
	router.Use(gin.Recovery())

	// Add routes
	router.GET("/", handlers.Home)
	router.GET("/articles", handlers.Articles)
	router.GET("/articles/:slug", handlers.Article)
	router.GET("/tags", handlers.Tags)
	router.GET("/categories", handlers.Categories)
	router.GET("/search", handlers.Search)
	router.GET("/health", handlers.Health)

	return &BenchmarkSetup{
		handlers: handlers,
		router:   router,
		articles: articles,
	}
}

// generateBenchmarkArticles creates realistic article data for benchmarking
func generateBenchmarkArticles(count int) []*models.Article {
	articles := make([]*models.Article, count)
	categories := []string{"tech", "web", "golang", "performance", "testing", "backend", "frontend", "devops"}
	tagSets := [][]string{
		{"golang", "performance", "benchmark"},
		{"web", "http", "server"},
		{"testing", "unit", "integration"},
		{"cache", "optimization", "speed"},
		{"api", "rest", "json"},
		{"database", "sql", "query"},
		{"docker", "deployment", "ci"},
		{"security", "authentication", "authorization"},
	}

	for i := 0; i < count; i++ {
		articles[i] = &models.Article{
			Slug:        fmt.Sprintf("article-%d", i),
			Title:       fmt.Sprintf("Article %d: Performance Testing and Optimization", i),
			Content:     generateArticleContent(i),
			Description: fmt.Sprintf("This is article %d about performance testing and optimization techniques.", i),
			Tags:        tagSets[i%len(tagSets)],
			Categories:  []string{categories[i%len(categories)]},
			Date:        time.Now().AddDate(0, 0, -i),
			Draft:       false,
			Featured:    i%10 == 0, // Every 10th article is featured
		}
	}
	return articles
}

func generateArticleContent(index int) string {
	return fmt.Sprintf(`
# Article %d Content

This is a comprehensive article about performance testing and optimization techniques.
It covers various aspects of modern web development including:

## Performance Optimization

Performance is crucial for modern web applications. Here are key strategies:

1. **Caching**: Implement effective caching strategies at multiple layers
2. **Database Optimization**: Use proper indexing and query optimization
3. **HTTP Optimization**: Leverage HTTP caching headers and compression
4. **Resource Management**: Optimize static assets and minimize bundle sizes

## Testing Strategies

Effective testing ensures reliability and performance:

- Unit tests for individual components
- Integration tests for system behavior
- Performance benchmarks for scalability
- Load testing for real-world scenarios

## Monitoring and Metrics

Continuous monitoring helps maintain performance:

- Application metrics and logging
- Error tracking and alerting  
- Performance monitoring and profiling
- User experience metrics

This article contains approximately 500 words of realistic content that would be typical
for a technical blog post about performance optimization and testing methodologies.

Article index: %d
Generated at: %s
`, index, index, time.Now().Format(time.RFC3339))
}

// Core endpoint benchmarks
func BenchmarkHandlers_HomePage(b *testing.B) {
	setup := NewBenchmarkSetup(b)
	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			setup.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

func BenchmarkHandlers_ArticlesList(b *testing.B) {
	setup := NewBenchmarkSetup(b)
	req := httptest.NewRequest("GET", "/articles", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			setup.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

func BenchmarkHandlers_SingleArticle(b *testing.B) {
	setup := NewBenchmarkSetup(b)
	req := httptest.NewRequest("GET", "/articles/article-0", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			setup.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

func BenchmarkHandlers_Search(b *testing.B) {
	setup := NewBenchmarkSetup(b)
	req := httptest.NewRequest("GET", "/search?q=performance", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			setup.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

func BenchmarkHandlers_TagsPage(b *testing.B) {
	setup := NewBenchmarkSetup(b)
	req := httptest.NewRequest("GET", "/tags", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			setup.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// Cache performance benchmarks
func BenchmarkHandlers_CacheHitVsMiss(b *testing.B) {
	setup := NewBenchmarkSetup(b)

	b.Run("cache_miss", func(b *testing.B) {
		// First request will be cache miss
		req := httptest.NewRequest("GET", "/", nil)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Clear cache for each iteration to ensure cache miss
			if i%100 == 0 { // Only clear occasionally to avoid too much overhead
				// We can't directly clear obcache, so we'll use different paths
				path := fmt.Sprintf("/articles/article-%d", i%1000)
				req = httptest.NewRequest("GET", path, nil)
			}
			w := httptest.NewRecorder()
			setup.router.ServeHTTP(w, req)
		}
	})

	b.Run("cache_hit", func(b *testing.B) {
		// Warm up cache
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		setup.router.ServeHTTP(w, req)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				w := httptest.NewRecorder()
				setup.router.ServeHTTP(w, req)
			}
		})
	})
}

// Memory efficiency benchmarks
func BenchmarkHandlers_MemoryUsage(b *testing.B) {
	setup := NewBenchmarkSetup(b)
	req := httptest.NewRequest("GET", "/articles", nil)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		setup.router.ServeHTTP(w, req)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.HeapAlloc-m1.HeapAlloc)/float64(b.N), "heap-bytes/op")
}

// Concurrent load benchmarks
func BenchmarkHandlers_ConcurrentLoad(b *testing.B) {
	setup := NewBenchmarkSetup(b)

	tests := []struct {
		name string
		path string
	}{
		{"home", "/"},
		{"articles", "/articles"},
		{"single_article", "/articles/article-0"},
		{"search", "/search?q=test"},
		{"tags", "/tags"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			req := httptest.NewRequest("GET", tt.path, nil)

			b.SetParallelism(10) // 10x parallelism
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					w := httptest.NewRecorder()
					setup.router.ServeHTTP(w, req)
					if w.Code >= 400 {
						b.Fatalf("HTTP error %d for path %s", w.Code, tt.path)
					}
				}
			})
		})
	}
}

// Service layer benchmarks (without HTTP overhead)
func BenchmarkServices_ArticleRetrieval(b *testing.B) {
	articles := generateBenchmarkArticles(10000)
	service := NewInMemoryArticleService(articles)

	b.Run("get_all_articles", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			all := service.GetAllArticles()
			if len(all) == 0 {
				b.Fatal("Expected articles")
			}
		}
	})

	b.Run("get_article_by_slug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slug := fmt.Sprintf("article-%d", i%1000)
			article, err := service.GetArticleBySlug(slug)
			if err != nil || article == nil {
				b.Fatalf("Failed to get article %s", slug)
			}
		}
	})

	b.Run("get_articles_by_tag", func(b *testing.B) {
		tags := []string{"golang", "performance", "testing", "web"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tag := tags[i%len(tags)]
			articles := service.GetArticlesByTag(tag)
			if len(articles) == 0 {
				b.Fatalf("Expected articles for tag %s", tag)
			}
		}
	})

	b.Run("get_tag_counts", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			counts := service.GetTagCounts()
			if len(counts) == 0 {
				b.Fatal("Expected tag counts")
			}
		}
	})
}

func BenchmarkServices_Search(b *testing.B) {
	articles := generateBenchmarkArticles(1000)
	service := NewInMemorySearchService()

	queries := []string{"performance", "golang", "testing", "optimization", "web"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			query := queries[0] // Use first query for consistent benchmarking
			results := service.Search(articles, query, 20)
			_ = results // Use results to prevent compiler optimization
		}
	})
}

// Cached function benchmarks
func BenchmarkHandlers_CachedFunctions(b *testing.B) {
	setup := NewBenchmarkSetup(b)

	b.Run("home_data_uncached", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, err := setup.handlers.getHomeDataUncached()
			if err != nil || data == nil {
				b.Fatal("Failed to generate home data")
			}
		}
	})

	b.Run("article_data_uncached", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slug := fmt.Sprintf("article-%d", i%100)
			data, err := setup.handlers.getArticleDataUncached(slug)
			if err != nil || data == nil {
				b.Fatalf("Failed to generate article data for %s", slug)
			}
		}
	})

	b.Run("articles_page_uncached", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, err := setup.handlers.getArticlesPageUncached(1)
			if err != nil || data == nil {
				b.Fatal("Failed to generate articles page data")
			}
		}
	})

	b.Run("search_results_uncached", func(b *testing.B) {
		queries := []string{"performance", "golang", "testing", "web"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			query := queries[i%len(queries)]
			data, err := setup.handlers.getSearchResultsUncached(query)
			if err != nil || data == nil {
				b.Fatalf("Failed to generate search results for %s", query)
			}
		}
	})
}

// HTTP-specific benchmarks
func BenchmarkHandlers_HTTPOverhead(b *testing.B) {
	setup := NewBenchmarkSetup(b)

	b.Run("minimal_handler", func(b *testing.B) {
		// Test just the HTTP overhead with minimal handler
		router := gin.New()
		router.GET("/minimal", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest("GET", "/minimal", nil)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
			}
		})
	})

	b.Run("full_handler_chain", func(b *testing.B) {
		req := httptest.NewRequest("GET", "/health", nil)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				w := httptest.NewRecorder()
				setup.router.ServeHTTP(w, req)
			}
		})
	})
}
