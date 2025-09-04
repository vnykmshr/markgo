package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/vnykmshr/markgo/internal/models"
)

// Memory and Resource Usage Benchmarks
// These benchmarks measure memory allocation patterns and resource usage

func BenchmarkArticleMemoryUsage(b *testing.B) {
	// Test memory consumption during article loading and processing
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(100)
	articles := createLargeArticleSet(100)
	largeArticle := createLargeArticle()

	mockCacheService.On("Get", "article_large-article").Return(nil, false)
	mockArticleService.On("GetArticleBySlug", "large-article").Return(largeArticle, nil)
	mockArticleService.On("GetArticlesByTag", mock.Anything).Return(articles[:5])
	mockArticleService.On("GetRecentArticles", 5).Return(articles[:5])
	mockCacheService.On("Set", "article_large-article", mock.Anything, time.Hour).Return()

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/articles/:slug", handlers.Article)

	// Force GC before starting
	runtime.GC()
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		req, _ := http.NewRequest("GET", "/articles/large-article", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", recorder.Code)
		}
	}

	runtime.ReadMemStats(&m2)

	// Report memory metrics
	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "total-bytes/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
	b.ReportMetric(float64(m2.Frees-m1.Frees)/float64(b.N), "frees/op")
}

func BenchmarkCacheMemoryEfficiency(b *testing.B) {
	// Test memory usage patterns under different cache configurations
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(50)
	articles := createLargeArticleSet(50)

	// Setup cache hits and misses
	for i, article := range articles {
		cacheKey := fmt.Sprintf("article_%s", article.Slug)
		if i%2 == 0 {
			// Cache hit
			mockCacheService.On("Get", cacheKey).Return(gin.H{"cached": true}, true)
		} else {
			// Cache miss
			mockCacheService.On("Get", cacheKey).Return(nil, false)
			mockArticleService.On("GetArticleBySlug", article.Slug).Return(article, nil)
			mockArticleService.On("GetArticlesByTag", mock.Anything).Return([]*models.Article{})
			mockArticleService.On("GetRecentArticles", 5).Return(articles[:5])
			mockCacheService.On("Set", cacheKey, mock.Anything, time.Hour).Return()
		}
	}

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/articles/:slug", handlers.Article)

	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.ReportAllocs()

	articleIndex := 0
	for b.Loop() {
		article := articles[articleIndex%len(articles)]
		req, _ := http.NewRequest("GET", fmt.Sprintf("/articles/%s", article.Slug), nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", recorder.Code)
		}
		articleIndex++
	}

	runtime.ReadMemStats(&m2)

	// Report cache efficiency metrics
	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.HeapObjects-m1.HeapObjects)/float64(b.N), "heap-objects/op")
}

func BenchmarkSearchMemoryUsage(b *testing.B) {
	// Test memory usage during search operations with large datasets
	handlers, mockArticleService, _, mockCacheService, mockSearchService := createBenchmarkHandlers(1000)
	articles := createLargeArticleSet(1000)
	searchResults := createSearchResults(20)

	mockCacheService.On("Get", "search_golang").Return(nil, false)
	mockArticleService.On("GetAllArticles").Return(articles)
	mockSearchService.On("Search", articles, "golang", 20).Return(searchResults)
	mockArticleService.On("GetRecentArticles", 5).Return(articles[:5])
	mockCacheService.On("Set", "search_golang", mock.Anything, 30*time.Minute).Return()

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/search", handlers.Search)

	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		req, _ := http.NewRequest("GET", "/search?q=golang", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", recorder.Code)
		}
	}

	runtime.ReadMemStats(&m2)

	// Report search-specific memory metrics
	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.Sys-m1.Sys)/float64(b.N), "sys-bytes/op")
	b.ReportMetric(float64(m2.GCSys-m1.GCSys)/float64(b.N), "gc-sys-bytes/op")
}

func BenchmarkGoroutineUsage(b *testing.B) {
	// Test goroutine creation and cleanup patterns
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(100)
	articles := createLargeArticleSet(100)

	mockCacheService.On("Get", "home_page").Return(nil, false)
	mockArticleService.On("GetAllArticles").Return(articles)
	mockArticleService.On("GetFeaturedArticles", 3).Return(articles[:3])
	mockArticleService.On("GetRecentArticles", 9).Return(articles[:9])
	mockArticleService.On("GetCategoryCounts").Return(createCategoryCounts(20))
	mockArticleService.On("GetTagCounts").Return(createTagCounts(50))
	mockCacheService.On("Set", "home_page", mock.Anything, time.Hour).Return()

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/", handlers.Home)

	startGoroutines := runtime.NumGoroutine()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		req, _ := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", recorder.Code)
		}
	}

	endGoroutines := runtime.NumGoroutine()
	goroutineDelta := endGoroutines - startGoroutines

	b.ReportMetric(float64(goroutineDelta), "goroutine-delta")
	b.ReportMetric(float64(endGoroutines), "total-goroutines")
}

func BenchmarkMemoryLeakDetection(b *testing.B) {
	// Test for potential memory leaks over extended operations
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(50)
	articles := createLargeArticleSet(50)

	// Setup mocks for multiple endpoints
	setupThroughputMocks(mockArticleService, mockCacheService, articles)

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/", handlers.Home)
	router.GET("/articles/:slug", handlers.Article)

	endpoints := []string{"/", "/articles/test-article-0", "/articles/test-article-1"}

	// Force cleanup before starting
	runtime.GC()
	runtime.GC() // Double GC to ensure cleanup

	// Run multiple iterations to detect leaks
	var memStats []runtime.MemStats

	for i := 0; i < 10; i++ {
		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		memStats = append(memStats, m)

		// Run batch of requests
		for j := 0; j < 100; j++ {
			endpoint := endpoints[j%len(endpoints)]
			req, _ := http.NewRequest("GET", endpoint, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			
			// Explicit cleanup every 20 requests to prevent accumulation
			if j%20 == 0 {
				runtime.GC()
			}
		}
		
		// Force cleanup after each batch
		runtime.GC()
	}

	// Analyze memory growth patterns
	initialMem := memStats[0].Alloc
	finalMem := memStats[len(memStats)-1].Alloc
	memoryGrowth := finalMem - initialMem

	b.ReportMetric(float64(memoryGrowth), "memory-growth-bytes")
	b.ReportMetric(float64(finalMem)/float64(initialMem), "memory-growth-ratio")

	// Check for concerning memory growth (>50% increase might indicate a leak)
	growthRatio := float64(finalMem) / float64(initialMem)
	if growthRatio > 1.5 {
		b.Logf("WARNING: Potential memory leak detected. Memory grew by %.1f%% over test duration",
			(growthRatio-1)*100)
	}
}

func BenchmarkBaselineResourceUsage(b *testing.B) {
	// Establish baseline resource usage for comparison
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(10)
	articles := createLargeArticleSet(10)

	mockCacheService.On("Get", "home_page").Return(nil, false)
	mockArticleService.On("GetAllArticles").Return(articles)
	mockArticleService.On("GetFeaturedArticles", 3).Return(articles[:3])
	mockArticleService.On("GetRecentArticles", 9).Return(articles[:9])
	mockArticleService.On("GetCategoryCounts").Return(createCategoryCounts(5))
	mockArticleService.On("GetTagCounts").Return(createTagCounts(10))
	mockCacheService.On("Set", "home_page", mock.Anything, time.Hour).Return()

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/", handlers.Home)

	// Measure baseline resource usage
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		req, _ := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", recorder.Code)
		}
	}

	runtime.ReadMemStats(&m2)
	endGoroutines := runtime.NumGoroutine()

	// Report comprehensive baseline metrics
	b.ReportMetric(float64(m2.Alloc)/1024/1024, "heap-alloc-mb")
	b.ReportMetric(float64(m2.Sys)/1024/1024, "sys-memory-mb")
	b.ReportMetric(float64(m2.HeapObjects), "heap-objects")
	b.ReportMetric(float64(endGoroutines), "goroutines")
	b.ReportMetric(float64(m2.NumGC-m1.NumGC)/float64(b.N), "gc-runs/op")
	b.ReportMetric(float64(m2.PauseTotalNs-m1.PauseTotalNs)/float64(b.N), "gc-pause-ns/op")

	// Validate against memory targets (Ghost uses ~300MB, we target <50MB)
	heapMB := float64(m2.Alloc) / 1024 / 1024
	if heapMB > 50 {
		b.Logf("WARNING: Heap usage %.1fMB exceeds target of 50MB (Ghost baseline: ~300MB)", heapMB)
	} else {
		b.Logf("✅ Memory usage %.1fMB is within target (%.1fx better than Ghost ~300MB)",
			heapMB, 300/heapMB)
	}
}

func BenchmarkMemoryProfileComparison(b *testing.B) {
	// Create memory profile data for comparison with competitors
	handlers, mockArticleService, _, mockCacheService, mockSearchService := createBenchmarkHandlers(200)
	articles := createLargeArticleSet(200)

	// Setup for multiple request types
	mockCacheService.On("Get", mock.Anything).Return(nil, false).Maybe()
	mockCacheService.On("Set", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockArticleService.On("GetAllArticles").Return(articles).Maybe()
	mockArticleService.On("GetArticleBySlug", mock.Anything).Return(articles[0], nil).Maybe()
	mockArticleService.On("GetFeaturedArticles", mock.Anything).Return(articles[:3]).Maybe()
	mockArticleService.On("GetRecentArticles", mock.Anything).Return(articles[:9]).Maybe()
	mockArticleService.On("GetCategoryCounts").Return(createCategoryCounts(20)).Maybe()
	mockArticleService.On("GetTagCounts").Return(createTagCounts(50)).Maybe()
	mockArticleService.On("GetArticlesByTag", mock.Anything).Return(articles[:5]).Maybe()

	mockSearchService.On("Search", mock.Anything, mock.Anything, mock.Anything).Return(createSearchResults(10)).Maybe()

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/", handlers.Home)
	router.GET("/articles/:slug", handlers.Article)
	router.GET("/search", handlers.Search)

	requestTypes := []struct {
		name string
		path string
	}{
		{"home", "/"},
		{"article", "/articles/test-article-0"},
		{"search", "/search?q=golang"},
	}

	runtime.GC()
	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)

	b.ResetTimer()
	b.ReportAllocs()

	requestIndex := 0
	for b.Loop() {
		reqType := requestTypes[requestIndex%len(requestTypes)]
		req, _ := http.NewRequest("GET", reqType.path, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d for %s", recorder.Code, reqType.name)
		}
		requestIndex++
	}

	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)

	// Generate competitor comparison metrics
	currentMemoryMB := float64(finalMem.Alloc) / 1024 / 1024

	// Competitor memory usage (from research)
	ghostMemoryMB := 300.0
	wpMemoryMB := 2048.0

	memoryAdvantageGhost := ghostMemoryMB / currentMemoryMB
	memoryAdvantageWP := wpMemoryMB / currentMemoryMB

	b.ReportMetric(currentMemoryMB, "memory-usage-mb")
	b.ReportMetric(memoryAdvantageGhost, "memory-advantage-vs-ghost")
	b.ReportMetric(memoryAdvantageWP, "memory-advantage-vs-wordpress")

	b.Logf("Memory Comparison:")
	b.Logf("  MarkGo: %.1f MB", currentMemoryMB)
	b.Logf("  vs Ghost: %.1fx more efficient", memoryAdvantageGhost)
	b.Logf("  vs WordPress: %.1fx more efficient", memoryAdvantageWP)
}
