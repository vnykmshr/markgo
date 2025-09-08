package article

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vnykmshr/markgo/internal/models"
)

// ServiceContainer provides dependency injection and service orchestration
type ServiceContainer struct {
	// Core services
	repository       Repository
	contentProcessor ContentProcessor
	searchService    SearchService
	compositeService *CompositeService

	// Caching
	cacheCoordinator *CacheCoordinator

	logger *slog.Logger

	// Configuration
	config *Config
}

// Config holds configuration for the article services
type Config struct {
	ArticlesPath string
	CacheEnabled bool
	CacheConfig  *CacheConfig
	SearchIndex  bool
}

// NewServiceContainer creates and initializes all article-related services
func NewServiceContainer(config *Config, logger *slog.Logger) (*ServiceContainer, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	container := &ServiceContainer{
		config: config,
		logger: logger,
	}

	// Initialize services in dependency order
	if err := container.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	return container, nil
}

// initializeServices initializes all services in the correct order
func (c *ServiceContainer) initializeServices() error {
	// 1. Initialize cache coordinator if caching is enabled
	if c.config.CacheEnabled {
		cacheConfig := c.config.CacheConfig
		if cacheConfig == nil {
			cacheConfig = DefaultCacheConfig()
		}

		var err error
		c.cacheCoordinator, err = NewCacheCoordinator(cacheConfig, c.logger)
		if err != nil {
			c.logger.Warn("Failed to initialize cache coordinator, proceeding without caching", "error", err)
		}
	}

	// 2. Initialize base repository (data access layer)
	baseRepository := NewFileSystemRepository(c.config.ArticlesPath, c.logger)

	// 3. Wrap repository with caching if available
	if c.cacheCoordinator != nil {
		c.repository = NewCachedRepository(baseRepository, c.cacheCoordinator, c.logger)
		c.logger.Info("Initialized cached repository")
	} else {
		c.repository = baseRepository
		c.logger.Info("Initialized non-cached repository")
	}

	// 4. Initialize content processor
	c.contentProcessor = NewMarkdownContentProcessor(c.logger)

	// 5. Initialize search service
	c.searchService = NewTextSearchService(c.logger)

	// 6. Initialize composite service that orchestrates all services
	c.compositeService = NewCompositeService(
		c.repository,
		c.contentProcessor,
		c.searchService,
		c.logger,
	)

	return nil
}

// Start initializes and starts all services
func (c *ServiceContainer) Start(ctx context.Context) error {
	c.logger.Info("Starting article service container")

	// Start the composite service which will handle initialization
	if err := c.compositeService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start composite service: %w", err)
	}

	c.logger.Info("Article service container started successfully")
	return nil
}

// Stop gracefully shuts down all services
func (c *ServiceContainer) Stop() error {
	c.logger.Info("Stopping article service container")

	// Stop composite service first
	if c.compositeService != nil {
		if err := c.compositeService.Stop(); err != nil {
			c.logger.Error("Error stopping composite service", "error", err)
		}
	}

	// Stop cache coordinator
	if c.cacheCoordinator != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := c.cacheCoordinator.Shutdown(ctx); err != nil {
			c.logger.Error("Error stopping cache coordinator", "error", err)
		}
	}

	c.logger.Info("Article service container stopped")
	return nil
}

// GetService returns the main article service interface
func (c *ServiceContainer) GetService() Service {
	return c.compositeService
}

// GetRepository provides access to the repository layer
func (c *ServiceContainer) GetRepository() Repository {
	return c.repository
}

// GetContentProcessor provides access to content processing
func (c *ServiceContainer) GetContentProcessor() ContentProcessor {
	return c.contentProcessor
}

// GetSearchService provides access to search functionality
func (c *ServiceContainer) GetSearchService() SearchService {
	return c.searchService
}

// GetCacheCoordinator provides access to caching functionality
func (c *ServiceContainer) GetCacheCoordinator() *CacheCoordinator {
	return c.cacheCoordinator
}

// Health check methods

// IsHealthy checks if all services are functioning properly
func (c *ServiceContainer) IsHealthy() bool {
	// Check if composite service is started
	if c.compositeService == nil {
		return false
	}

	// Basic health check - try to get stats
	stats := c.compositeService.GetStats()
	return stats != nil
}

// GetHealthStatus returns detailed health information
func (c *ServiceContainer) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"healthy": c.IsHealthy(),
		"services": map[string]interface{}{
			"repository":        c.repository != nil,
			"content_processor": c.contentProcessor != nil,
			"search_service":    c.searchService != nil,
			"composite_service": c.compositeService != nil,
		},
	}

	if c.compositeService != nil {
		stats := c.compositeService.GetStats()
		if stats != nil {
			status["stats"] = map[string]interface{}{
				"total_articles":   stats.TotalArticles,
				"published_count":  stats.PublishedCount,
				"draft_count":      stats.DraftCount,
				"last_reload_time": c.compositeService.GetLastReloadTime(),
			}
		}
	}

	// Add cache statistics if available
	if c.cacheCoordinator != nil {
		status["cache"] = c.cacheCoordinator.GetCacheStats()
		status["cache_healthy"] = c.cacheCoordinator.IsHealthy()
	}

	return status
}

// Utility methods for backward compatibility

// CreateLegacyWrapper creates a wrapper that implements the legacy interfaces
func (c *ServiceContainer) CreateLegacyWrapper() *LegacyServiceWrapper {
	return &LegacyServiceWrapper{
		service: c.compositeService,
		logger:  c.logger,
	}
}

// LegacyServiceWrapper provides backward compatibility with existing interfaces
type LegacyServiceWrapper struct {
	service Service
	logger  *slog.Logger
}

// Implement legacy ArticleProcessor interface for models.Article
func (w *LegacyServiceWrapper) ProcessMarkdown(content string) (string, error) {
	if w.service == nil {
		return "", fmt.Errorf("service not initialized")
	}

	// Get content processor from the composite service
	// For now, create a temporary content processor
	processor := NewMarkdownContentProcessor(w.logger)
	return processor.ProcessMarkdown(content)
}

func (w *LegacyServiceWrapper) GenerateExcerpt(content string, maxLength int) string {
	if w.service == nil {
		return ""
	}

	processor := NewMarkdownContentProcessor(w.logger)
	return processor.GenerateExcerpt(content, maxLength)
}

func (w *LegacyServiceWrapper) CalculateReadingTime(content string) int {
	if w.service == nil {
		return 0
	}

	processor := NewMarkdownContentProcessor(w.logger)
	return processor.CalculateReadingTime(content)
}

// Forward all methods to the composite service
func (w *LegacyServiceWrapper) GetAllArticles() []*models.Article {
	if w.service == nil {
		return []*models.Article{}
	}
	return w.service.GetAllArticles()
}

func (w *LegacyServiceWrapper) GetArticleBySlug(slug string) (*models.Article, error) {
	if w.service == nil {
		return nil, fmt.Errorf("service not initialized")
	}
	return w.service.GetArticleBySlug(slug)
}

func (w *LegacyServiceWrapper) GetArticlesByTag(tag string) []*models.Article {
	if w.service == nil {
		return []*models.Article{}
	}
	return w.service.GetArticlesByTag(tag)
}

func (w *LegacyServiceWrapper) GetArticlesByCategory(category string) []*models.Article {
	if w.service == nil {
		return []*models.Article{}
	}
	return w.service.GetArticlesByCategory(category)
}

func (w *LegacyServiceWrapper) GetArticlesForFeed(limit int) []*models.Article {
	if w.service == nil {
		return []*models.Article{}
	}
	return w.service.GetArticlesForFeed(limit)
}

func (w *LegacyServiceWrapper) GetFeaturedArticles(limit int) []*models.Article {
	if w.service == nil {
		return []*models.Article{}
	}
	return w.service.GetFeaturedArticles(limit)
}

func (w *LegacyServiceWrapper) GetRecentArticles(limit int) []*models.Article {
	if w.service == nil {
		return []*models.Article{}
	}
	return w.service.GetRecentArticles(limit)
}

func (w *LegacyServiceWrapper) GetAllTags() []string {
	if w.service == nil {
		return []string{}
	}
	return w.service.GetAllTags()
}

func (w *LegacyServiceWrapper) GetAllCategories() []string {
	if w.service == nil {
		return []string{}
	}
	return w.service.GetAllCategories()
}

func (w *LegacyServiceWrapper) GetTagCounts() []models.TagCount {
	if w.service == nil {
		return []models.TagCount{}
	}
	return w.service.GetTagCounts()
}

func (w *LegacyServiceWrapper) GetCategoryCounts() []models.CategoryCount {
	if w.service == nil {
		return []models.CategoryCount{}
	}
	return w.service.GetCategoryCounts()
}

func (w *LegacyServiceWrapper) GetStats() *models.Stats {
	if w.service == nil {
		return nil
	}
	return w.service.GetStats()
}

func (w *LegacyServiceWrapper) ReloadArticles() error {
	if w.service == nil {
		return fmt.Errorf("service not initialized")
	}
	return w.service.ReloadArticles()
}

func (w *LegacyServiceWrapper) GetDraftArticles() []*models.Article {
	if w.service == nil {
		return []*models.Article{}
	}
	return w.service.GetDraftArticles()
}

func (w *LegacyServiceWrapper) GetDraftBySlug(slug string) (*models.Article, error) {
	if w.service == nil {
		return nil, fmt.Errorf("service not initialized")
	}
	return w.service.GetDraftBySlug(slug)
}

func (w *LegacyServiceWrapper) PreviewDraft(slug string) (*models.Article, error) {
	if w.service == nil {
		return nil, fmt.Errorf("service not initialized")
	}
	return w.service.PreviewDraft(slug)
}

func (w *LegacyServiceWrapper) PublishDraft(slug string) error {
	if w.service == nil {
		return fmt.Errorf("service not initialized")
	}
	return w.service.PublishDraft(slug)
}

func (w *LegacyServiceWrapper) UnpublishArticle(slug string) error {
	if w.service == nil {
		return fmt.Errorf("service not initialized")
	}
	return w.service.UnpublishArticle(slug)
}

// Search methods
func (w *LegacyServiceWrapper) SearchArticles(query string, limit int) []*models.SearchResult {
	if w.service == nil {
		return []*models.SearchResult{}
	}
	return w.service.SearchArticles(query, limit)
}

func (w *LegacyServiceWrapper) SearchInTitle(query string, limit int) []*models.SearchResult {
	if w.service == nil {
		return []*models.SearchResult{}
	}
	return w.service.SearchInTitle(query, limit)
}

func (w *LegacyServiceWrapper) GetSearchSuggestions(query string, limit int) []string {
	if w.service == nil {
		return []string{}
	}
	return w.service.GetSearchSuggestions(query, limit)
}
