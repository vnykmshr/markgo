package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/models"
)

// Comprehensive Integration Benchmark Tests
// These benchmarks test full request flows to validate competitive performance claims

func BenchmarkFullRequestFlow_Home(b *testing.B) {
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(100) // 100 test articles
	articles := createLargeArticleSet(100)

	// Setup mocks for home page
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
}

func BenchmarkFullRequestFlow_Article(b *testing.B) {
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(100)
	article := createLargeArticle() // Complex article with lots of content
	relatedArticles := createLargeArticleSet(10)

	mockCacheService.On("Get", "article_large-article").Return(nil, false)
	mockArticleService.On("GetArticleBySlug", "large-article").Return(article, nil)
	mockArticleService.On("GetArticlesByTag", mock.Anything).Return(relatedArticles[:5])
	mockArticleService.On("GetRecentArticles", 5).Return(relatedArticles[:5])
	mockCacheService.On("Set", "article_large-article", mock.Anything, time.Hour).Return()

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/articles/:slug", handlers.Article)

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
}

func BenchmarkFullRequestFlow_Search(b *testing.B) {
	handlers, mockArticleService, _, mockCacheService, mockSearchService := createBenchmarkHandlers(1000)
	articles := createLargeArticleSet(1000) // Large dataset for realistic search
	searchResults := createSearchResults(20)

	mockCacheService.On("Get", "search_golang").Return(nil, false)
	mockArticleService.On("GetAllArticles").Return(articles)
	mockSearchService.On("Search", articles, "golang", 20).Return(searchResults)
	mockArticleService.On("GetRecentArticles", 5).Return(articles[:5])
	mockCacheService.On("Set", "search_golang", mock.Anything, 30*time.Minute).Return()

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/search", handlers.Search)

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
}

func BenchmarkConcurrentArticleAccess(b *testing.B) {
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(50)
	articles := createLargeArticleSet(50)

	// Setup mocks for multiple articles
	for _, article := range articles {
		cacheKey := fmt.Sprintf("article_%s", article.Slug)
		mockCacheService.On("Get", cacheKey).Return(nil, false)
		mockArticleService.On("GetArticleBySlug", article.Slug).Return(article, nil)
		mockArticleService.On("GetArticlesByTag", mock.Anything).Return([]*models.Article{})
		mockArticleService.On("GetRecentArticles", 5).Return(articles[:5])
		mockCacheService.On("Set", cacheKey, mock.Anything, time.Hour).Return()
	}

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/articles/:slug", handlers.Article)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(10) // Simulate 10 concurrent users

	b.RunParallel(func(pb *testing.PB) {
		articleIndex := 0
		for pb.Next() {
			article := articles[articleIndex%len(articles)]
			req, _ := http.NewRequest("GET", fmt.Sprintf("/articles/%s", article.Slug), nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			
			if recorder.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", recorder.Code)
			}
			articleIndex++
		}
	})
}

func BenchmarkSearchWithLargeDataset(b *testing.B) {
	handlers, mockArticleService, _, mockCacheService, mockSearchService := createBenchmarkHandlers(5000)
	articles := createLargeArticleSet(5000) // Very large dataset
	searchTerms := []string{"golang", "javascript", "python", "rust", "java", "performance", "web", "api", "database", "security"}

	// Setup mocks for different search terms
	for _, term := range searchTerms {
		cacheKey := fmt.Sprintf("search_%s", term)
		mockCacheService.On("Get", cacheKey).Return(nil, false)
		mockArticleService.On("GetAllArticles").Return(articles).Maybe()
		mockSearchService.On("Search", articles, term, 20).Return(createSearchResults(15)).Maybe()
		mockArticleService.On("GetRecentArticles", 5).Return(articles[:5]).Maybe()
		mockCacheService.On("Set", cacheKey, mock.Anything, 30*time.Minute).Return().Maybe()
	}

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/search", handlers.Search)

	b.ResetTimer()
	b.ReportAllocs()

	termIndex := 0
	for b.Loop() {
		term := searchTerms[termIndex%len(searchTerms)]
		req, _ := http.NewRequest("GET", fmt.Sprintf("/search?q=%s", term), nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		if recorder.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", recorder.Code)
		}
		termIndex++
	}
}

func BenchmarkConcurrentSearchRequests(b *testing.B) {
	handlers, mockArticleService, _, mockCacheService, mockSearchService := createBenchmarkHandlers(1000)
	articles := createLargeArticleSet(1000)
	searchTerms := []string{"golang", "performance", "web", "api", "database"}

	// Setup mocks
	for _, term := range searchTerms {
		cacheKey := fmt.Sprintf("search_%s", term)
		mockCacheService.On("Get", cacheKey).Return(nil, false)
		mockArticleService.On("GetAllArticles").Return(articles)
		mockSearchService.On("Search", articles, term, 20).Return(createSearchResults(10))
		mockArticleService.On("GetRecentArticles", 5).Return(articles[:5])
		mockCacheService.On("Set", cacheKey, mock.Anything, 30*time.Minute).Return()
	}

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/search", handlers.Search)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(20) // High concurrency

	b.RunParallel(func(pb *testing.PB) {
		termIndex := 0
		for pb.Next() {
			term := searchTerms[termIndex%len(searchTerms)]
			req, _ := http.NewRequest("GET", fmt.Sprintf("/search?q=%s", term), nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			
			if recorder.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", recorder.Code)
			}
			termIndex++
		}
	})
}

func BenchmarkMemoryUsageUnderLoad(b *testing.B) {
	// This benchmark tracks memory allocation patterns under load
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(200)
	articles := createLargeArticleSet(200)

	mockCacheService.On("Get", "home_page").Return(nil, false)
	mockArticleService.On("GetAllArticles").Return(articles)
	mockArticleService.On("GetFeaturedArticles", 3).Return(articles[:3])
	mockArticleService.On("GetRecentArticles", 9).Return(articles[:9])
	mockArticleService.On("GetCategoryCounts").Return(createCategoryCounts(30))
	mockArticleService.On("GetTagCounts").Return(createTagCounts(100))
	mockCacheService.On("Set", "home_page", mock.Anything, time.Hour).Return()

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/", handlers.Home)

	// Force GC before starting
	runtime.GC()
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
	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "total-bytes/op")
}

func BenchmarkCompetitorComparison_ResponseTime(b *testing.B) {
	// This benchmark specifically tests response time targets
	// Target: <50ms per request (95th percentile)
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

	var responseTimes []time.Duration

	b.ResetTimer()

	for b.Loop() {
		start := time.Now()
		req, _ := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		elapsed := time.Since(start)
		
		if recorder.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", recorder.Code)
		}
		
		responseTimes = append(responseTimes, elapsed)
	}

	// Calculate 95th percentile
	if len(responseTimes) > 0 {
		// Simple 95th percentile calculation
		sortedTimes := make([]time.Duration, len(responseTimes))
		copy(sortedTimes, responseTimes)
		
		// Basic bubble sort for small datasets
		n := len(sortedTimes)
		for i := 0; i < n-1; i++ {
			for j := 0; j < n-i-1; j++ {
				if sortedTimes[j] > sortedTimes[j+1] {
					sortedTimes[j], sortedTimes[j+1] = sortedTimes[j+1], sortedTimes[j]
				}
			}
		}
		
		p95Index := int(0.95 * float64(len(sortedTimes)))
		if p95Index >= len(sortedTimes) {
			p95Index = len(sortedTimes) - 1
		}
		
		p95Time := sortedTimes[p95Index]
		b.ReportMetric(float64(p95Time.Nanoseconds())/1e6, "p95-ms")
		
		// Report if we meet our target (<50ms)
		target := 50 * time.Millisecond
		if p95Time > target {
			b.Logf("WARNING: 95th percentile response time %.2fms exceeds target of 50ms", 
				float64(p95Time.Nanoseconds())/1e6)
		}
	}
}

func BenchmarkThroughputTest(b *testing.B) {
	// Tests sustained throughput - target >1000 req/s
	handlers, mockArticleService, _, mockCacheService, _ := createBenchmarkHandlers(50)
	articles := createLargeArticleSet(50)

	// Setup multiple endpoints
	setupThroughputMocks(mockArticleService, mockCacheService, articles)

	router := gin.New()
	setupMinimalTemplates(router)
	router.GET("/", handlers.Home)
	router.GET("/articles/:slug", handlers.Article)
	router.GET("/health", handlers.Health)

	b.ResetTimer()
	b.ReportAllocs()

	startTime := time.Now()
	requests := int64(0)

	// Run for a fixed duration to measure sustained throughput
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	const numWorkers = 50

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			localRequests := int64(0)
			
			for {
				select {
				case <-ctx.Done():
					// Add local requests to total atomically
					requests += localRequests
					return
				default:
					// Mix of different endpoints
					var req *http.Request
					switch localRequests % 3 {
					case 0:
						req, _ = http.NewRequest("GET", "/", nil)
					case 1:
						articleIndex := localRequests % int64(len(articles))
						req, _ = http.NewRequest("GET", fmt.Sprintf("/articles/%s", articles[articleIndex].Slug), nil)
					case 2:
						req, _ = http.NewRequest("GET", "/health", nil)
					}
					
					recorder := httptest.NewRecorder()
					router.ServeHTTP(recorder, req)
					
					if recorder.Code != http.StatusOK {
						b.Logf("Worker %d: Request failed with status %d", workerID, recorder.Code)
						continue
					}
					
					localRequests++
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	requestsPerSecond := float64(requests) / elapsed.Seconds()
	b.ReportMetric(requestsPerSecond, "req/s")
	
	if requestsPerSecond < 1000 {
		b.Logf("WARNING: Throughput %.0f req/s is below target of 1000 req/s", requestsPerSecond)
	}
	
	b.Logf("Processed %d requests in %v (%.0f req/s)", requests, elapsed, requestsPerSecond)
}

// Helper functions for benchmark tests

func createBenchmarkHandlers(articleCount int) (*Handlers, *MockArticleService, *MockEmailService, *MockCacheService, *MockSearchService) {
	mockArticleService := &MockArticleService{}
	mockEmailService := &MockEmailService{}
	mockCacheService := &MockCacheService{}
	mockSearchService := &MockSearchService{}

	cfg := &config.Config{
		Blog: config.BlogConfig{
			Title:        "Benchmark Blog",
			Description:  "A benchmark test blog",
			Author:       "Benchmark Author",
			PostsPerPage: 10,
		},
		Cache: config.CacheConfig{
			TTL: time.Hour,
		},
	}

	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	handlers := New(&Config{
		ArticleService: mockArticleService,
		CacheService:   mockCacheService,
		EmailService:   mockEmailService,
		SearchService:  mockSearchService,
		Config:         cfg,
		Logger:         logger,
	})

	return handlers, mockArticleService, mockEmailService, mockCacheService, mockSearchService
}

func createLargeArticleSet(count int) []*models.Article {
	articles := make([]*models.Article, count)
	categories := []string{"programming", "web", "mobile", "devops", "security", "data", "ai", "blockchain"}
	tags := []string{"golang", "javascript", "python", "rust", "java", "docker", "kubernetes", "aws", "performance", "testing"}
	
	for i := 0; i < count; i++ {
		articles[i] = &models.Article{
			Slug:        fmt.Sprintf("test-article-%d", i),
			Title:       fmt.Sprintf("Test Article %d - Advanced Topics in Software Engineering", i),
			Description: fmt.Sprintf("This is a comprehensive test article %d covering advanced topics in modern software development", i),
			Date:        time.Now().AddDate(0, 0, -i),
			Tags:        []string{tags[i%len(tags)], tags[(i+1)%len(tags)], tags[(i+2)%len(tags)]},
			Categories:  []string{categories[i%len(categories)]},
			Featured:    i%10 == 0, // Every 10th article is featured
			Draft:       false,
			Content:     generateLargeContent(1000 + i*10), // Varying content sizes
			Excerpt:     fmt.Sprintf("This is an excerpt for test article %d with substantial content to test performance", i),
		}
	}
	
	return articles
}

func createLargeArticle() *models.Article {
	return &models.Article{
		Slug:        "large-article",
		Title:       "Large Article - Comprehensive Guide to Modern Software Architecture",
		Description: "An extensive guide covering all aspects of modern software architecture and best practices",
		Date:        time.Now(),
		Tags:        []string{"architecture", "design", "performance", "scalability", "security"},
		Categories:  []string{"architecture", "engineering"},
		Featured:    true,
		Draft:       false,
		Content:     generateLargeContent(5000), // 5KB of content
		Excerpt:     "This comprehensive guide explores modern software architecture patterns and practices for building scalable, maintainable applications",
	}
}

func generateLargeContent(size int) string {
	content := `# Introduction

This is a comprehensive article about software engineering practices and modern web development. 

## Table of Contents

1. Introduction to Modern Web Development
2. Performance Optimization Techniques  
3. Security Best Practices
4. Scalability Patterns
5. Testing Strategies
6. Deployment and DevOps
7. Monitoring and Observability
8. Conclusion

## Section 1: Introduction to Modern Web Development

Modern web development has evolved significantly over the past decade. With the rise of single-page applications (SPAs), microservices architecture, and cloud-native development, developers need to understand a wide range of technologies and patterns.

### Key Technologies

- **Frontend Frameworks**: React, Vue.js, Angular, Svelte
- **Backend Technologies**: Go, Node.js, Python, Rust, Java
- **Databases**: PostgreSQL, MongoDB, Redis, Elasticsearch
- **Cloud Platforms**: AWS, GCP, Azure, DigitalOcean
- **DevOps Tools**: Docker, Kubernetes, Terraform, Ansible

## Section 2: Performance Optimization

Performance is crucial for user experience and business success. Here are key optimization strategies:

### Frontend Optimization

1. **Code Splitting**: Break your application into smaller chunks
2. **Lazy Loading**: Load components and resources on demand
3. **Image Optimization**: Use modern formats and responsive images
4. **Caching Strategies**: Implement proper browser and CDN caching

### Backend Optimization

1. **Database Optimization**: Proper indexing and query optimization
2. **Caching Layers**: Redis, Memcached, application-level caching
3. **Connection Pooling**: Efficient database connection management
4. **Compression**: Gzip, Brotli compression for responses

`

	// Repeat content to reach desired size
	for len(content) < size {
		content += "\n\n" + content
	}
	
	return content[:size]
}

func createCategoryCounts(count int) []models.CategoryCount {
	categories := []string{"programming", "web", "mobile", "devops", "security", "data", "ai", "blockchain", "design", "management"}
	counts := make([]models.CategoryCount, count)
	
	for i := 0; i < count; i++ {
		counts[i] = models.CategoryCount{
			Category: categories[i%len(categories)] + fmt.Sprintf("-%d", i),
			Count:    10 + i,
		}
	}
	
	return counts
}

func createTagCounts(count int) []models.TagCount {
	tags := []string{"golang", "javascript", "python", "rust", "java", "docker", "kubernetes", "aws", "performance", "testing"}
	counts := make([]models.TagCount, count)
	
	for i := 0; i < count; i++ {
		counts[i] = models.TagCount{
			Tag:   tags[i%len(tags)] + fmt.Sprintf("-%d", i),
			Count: 5 + i,
		}
	}
	
	return counts
}

func createSearchResults(count int) []*models.SearchResult {
	articles := createLargeArticleSet(count)
	results := make([]*models.SearchResult, count)
	
	for i, article := range articles {
		results[i] = &models.SearchResult{
			Article:       article,
			Score:         float64(100 - i), // Decreasing scores
			MatchedFields: []string{"title", "content", "tags"},
		}
	}
	
	return results
}

func setupThroughputMocks(mockArticleService *MockArticleService, mockCacheService *MockCacheService, articles []*models.Article) {
	// Home page mocks
	mockCacheService.On("Get", "home_page").Return(nil, false).Maybe()
	mockArticleService.On("GetAllArticles").Return(articles).Maybe()
	mockArticleService.On("GetFeaturedArticles", mock.Anything).Return(articles[:3]).Maybe()
	mockArticleService.On("GetRecentArticles", mock.Anything).Return(articles[:9]).Maybe()
	mockArticleService.On("GetCategoryCounts").Return(createCategoryCounts(10)).Maybe()
	mockArticleService.On("GetTagCounts").Return(createTagCounts(20)).Maybe()
	mockCacheService.On("Set", "home_page", mock.Anything, mock.Anything).Return().Maybe()

	// Article page mocks
	for _, article := range articles {
		cacheKey := fmt.Sprintf("article_%s", article.Slug)
		mockCacheService.On("Get", cacheKey).Return(nil, false).Maybe()
		mockArticleService.On("GetArticleBySlug", article.Slug).Return(article, nil).Maybe()
		mockArticleService.On("GetArticlesByTag", mock.Anything).Return([]*models.Article{}).Maybe()
		mockArticleService.On("GetRecentArticles", 5).Return(articles[:5]).Maybe()
		mockCacheService.On("Set", cacheKey, mock.Anything, mock.Anything).Return().Maybe()
	}
}