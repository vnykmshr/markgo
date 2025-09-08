package article

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vnykmshr/markgo/internal/models"
)

// ServiceFactory creates and configures article services for different use cases
type ServiceFactory struct {
	logger *slog.Logger
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(logger *slog.Logger) *ServiceFactory {
	return &ServiceFactory{
		logger: logger,
	}
}

// ArticleServiceInterface defines the article service interface
// Defined locally to avoid import cycles
type ArticleServiceInterface interface {
	// Article retrieval methods
	GetAllArticles() []*models.Article
	GetArticleBySlug(slug string) (*models.Article, error)
	GetArticlesByTag(tag string) []*models.Article
	GetArticlesByCategory(category string) []*models.Article
	GetArticlesForFeed(limit int) []*models.Article

	// Featured and recent articles
	GetFeaturedArticles(limit int) []*models.Article
	GetRecentArticles(limit int) []*models.Article

	// Metadata methods
	GetAllTags() []string
	GetAllCategories() []string
	GetTagCounts() []models.TagCount
	GetCategoryCounts() []models.CategoryCount

	// Statistics and management
	GetStats() *models.Stats
	ReloadArticles() error

	// Draft operations
	GetDraftArticles() []*models.Article
	GetDraftBySlug(slug string) (*models.Article, error)
	PreviewDraft(slug string) (*models.Article, error)
	PublishDraft(slug string) error
	UnpublishArticle(slug string) error
}

// CreateService creates a new article service using the modular architecture
func (f *ServiceFactory) CreateService(articlesPath string) (ArticleServiceInterface, error) {
	// Create configuration
	config := &Config{
		ArticlesPath: articlesPath,
		CacheEnabled: true,
		CacheConfig:  DefaultCacheConfig(),
		SearchIndex:  true,
	}

	// Create service container
	container, err := NewServiceContainer(config, f.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create service container: %w", err)
	}

	// Initialize the container
	ctx := context.Background()
	if err := container.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start service container: %w", err)
	}

	// Create service wrapper
	wrapper := container.CreateServiceWrapper()

	// Create service adapter that implements the ArticleServiceInterface
	adapter := &ServiceAdapter{
		wrapper:   wrapper,
		container: container,
		logger:    f.logger,
	}

	return adapter, nil
}


// ServiceAdapter adapts the service container to implement the ArticleServiceInterface
type ServiceAdapter struct {
	wrapper   *ServiceWrapper
	container *ServiceContainer
	logger    *slog.Logger
}

// Ensure ServiceAdapter implements ArticleServiceInterface
var _ ArticleServiceInterface = (*ServiceAdapter)(nil)

// Implement all methods from services.ArticleServiceInterface
func (a *ServiceAdapter) GetAllArticles() []*models.Article {
	return a.wrapper.GetAllArticles()
}

func (a *ServiceAdapter) GetArticleBySlug(slug string) (*models.Article, error) {
	return a.wrapper.GetArticleBySlug(slug)
}

func (a *ServiceAdapter) GetArticlesByTag(tag string) []*models.Article {
	return a.wrapper.GetArticlesByTag(tag)
}

func (a *ServiceAdapter) GetArticlesByCategory(category string) []*models.Article {
	return a.wrapper.GetArticlesByCategory(category)
}

func (a *ServiceAdapter) GetArticlesForFeed(limit int) []*models.Article {
	return a.wrapper.GetArticlesForFeed(limit)
}

func (a *ServiceAdapter) GetFeaturedArticles(limit int) []*models.Article {
	return a.wrapper.GetFeaturedArticles(limit)
}

func (a *ServiceAdapter) GetRecentArticles(limit int) []*models.Article {
	return a.wrapper.GetRecentArticles(limit)
}

func (a *ServiceAdapter) GetAllTags() []string {
	return a.wrapper.GetAllTags()
}

func (a *ServiceAdapter) GetAllCategories() []string {
	return a.wrapper.GetAllCategories()
}

func (a *ServiceAdapter) GetTagCounts() []models.TagCount {
	return a.wrapper.GetTagCounts()
}

func (a *ServiceAdapter) GetCategoryCounts() []models.CategoryCount {
	return a.wrapper.GetCategoryCounts()
}

func (a *ServiceAdapter) GetStats() *models.Stats {
	return a.wrapper.GetStats()
}

func (a *ServiceAdapter) ReloadArticles() error {
	return a.wrapper.ReloadArticles()
}

func (a *ServiceAdapter) GetDraftArticles() []*models.Article {
	return a.wrapper.GetDraftArticles()
}

func (a *ServiceAdapter) GetDraftBySlug(slug string) (*models.Article, error) {
	return a.wrapper.GetDraftBySlug(slug)
}

func (a *ServiceAdapter) PreviewDraft(slug string) (*models.Article, error) {
	return a.wrapper.PreviewDraft(slug)
}

func (a *ServiceAdapter) PublishDraft(slug string) error {
	return a.wrapper.PublishDraft(slug)
}

func (a *ServiceAdapter) UnpublishArticle(slug string) error {
	return a.wrapper.UnpublishArticle(slug)
}

// Shutdown gracefully shuts down the service
func (a *ServiceAdapter) Shutdown() error {
	if a.container != nil {
		return a.container.Stop()
	}
	return nil
}

// GetContainer provides access to the underlying container for advanced use cases
func (a *ServiceAdapter) GetContainer() *ServiceContainer {
	return a.container
}

// Health check methods
func (a *ServiceAdapter) IsHealthy() bool {
	return a.container.IsHealthy()
}

func (a *ServiceAdapter) GetHealthStatus() map[string]interface{} {
	return a.container.GetHealthStatus()
}

// Search functionality (these might be accessed through other interfaces)
func (a *ServiceAdapter) SearchArticles(query string, limit int) []*models.SearchResult {
	return a.wrapper.SearchArticles(query, limit)
}

func (a *ServiceAdapter) SearchInTitle(query string, limit int) []*models.SearchResult {
	return a.wrapper.SearchInTitle(query, limit)
}

func (a *ServiceAdapter) GetSearchSuggestions(query string, limit int) []string {
	return a.wrapper.GetSearchSuggestions(query, limit)
}
