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
//
//nolint:revive // Interface naming follows established pattern in codebase
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

	// Create service adapter that implements the ArticleServiceInterface
	adapter := &ServiceAdapter{
		service:   container.compositeService,
		container: container,
		logger:    f.logger,
	}

	return adapter, nil
}

// ServiceAdapter adapts the service container to implement the ArticleServiceInterface
type ServiceAdapter struct {
	service   Service
	container *ServiceContainer
	logger    *slog.Logger
}

// Ensure ServiceAdapter implements ArticleServiceInterface
var _ ArticleServiceInterface = (*ServiceAdapter)(nil)

// GetAllArticles returns all articles from the underlying service
func (a *ServiceAdapter) GetAllArticles() []*models.Article {
	return a.service.GetAllArticles()
}

func (a *ServiceAdapter) GetArticleBySlug(slug string) (*models.Article, error) {
	return a.service.GetArticleBySlug(slug)
}

func (a *ServiceAdapter) GetArticlesByTag(tag string) []*models.Article {
	return a.service.GetArticlesByTag(tag)
}

func (a *ServiceAdapter) GetArticlesByCategory(category string) []*models.Article {
	return a.service.GetArticlesByCategory(category)
}

func (a *ServiceAdapter) GetArticlesForFeed(limit int) []*models.Article {
	return a.service.GetArticlesForFeed(limit)
}

func (a *ServiceAdapter) GetFeaturedArticles(limit int) []*models.Article {
	return a.service.GetFeaturedArticles(limit)
}

func (a *ServiceAdapter) GetRecentArticles(limit int) []*models.Article {
	return a.service.GetRecentArticles(limit)
}

func (a *ServiceAdapter) GetAllTags() []string {
	return a.service.GetAllTags()
}

func (a *ServiceAdapter) GetAllCategories() []string {
	return a.service.GetAllCategories()
}

func (a *ServiceAdapter) GetTagCounts() []models.TagCount {
	return a.service.GetTagCounts()
}

func (a *ServiceAdapter) GetCategoryCounts() []models.CategoryCount {
	return a.service.GetCategoryCounts()
}

func (a *ServiceAdapter) GetStats() *models.Stats {
	return a.service.GetStats()
}

func (a *ServiceAdapter) ReloadArticles() error {
	return a.service.ReloadArticles()
}

func (a *ServiceAdapter) GetDraftArticles() []*models.Article {
	return a.service.GetDraftArticles()
}

func (a *ServiceAdapter) GetDraftBySlug(slug string) (*models.Article, error) {
	return a.service.GetDraftBySlug(slug)
}

func (a *ServiceAdapter) PreviewDraft(slug string) (*models.Article, error) {
	return a.service.PreviewDraft(slug)
}

func (a *ServiceAdapter) PublishDraft(slug string) error {
	return a.service.PublishDraft(slug)
}

func (a *ServiceAdapter) UnpublishArticle(slug string) error {
	return a.service.UnpublishArticle(slug)
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

// IsHealthy returns the health status of the service adapter
func (a *ServiceAdapter) IsHealthy() bool {
	return a.container.IsHealthy()
}

func (a *ServiceAdapter) GetHealthStatus() map[string]interface{} {
	return a.container.GetHealthStatus()
}

// SearchArticles performs article search using the underlying service
func (a *ServiceAdapter) SearchArticles(query string, limit int) []*models.SearchResult {
	return a.service.SearchArticles(query, limit)
}

func (a *ServiceAdapter) SearchInTitle(query string, limit int) []*models.SearchResult {
	return a.service.SearchInTitle(query, limit)
}

func (a *ServiceAdapter) GetSearchSuggestions(query string, limit int) []string {
	return a.service.GetSearchSuggestions(query, limit)
}
