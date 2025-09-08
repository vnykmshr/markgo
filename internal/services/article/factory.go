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

// ArticleServiceInterface defines the interface that external packages expect
// This avoids the import cycle by defining the interface locally
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

// CreateLegacyCompatibleService creates a service that's compatible with the existing
// ArticleServiceInterface while using the new modular architecture internally
func (f *ServiceFactory) CreateLegacyCompatibleService(articlesPath string) (ArticleServiceInterface, error) {
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

	// Create legacy wrapper
	wrapper := container.CreateLegacyWrapper()

	// Create an adapter that implements the full legacy interface
	adapter := &LegacyServiceAdapter{
		wrapper:   wrapper,
		container: container,
		logger:    f.logger,
	}

	return adapter, nil
}

// CreateModernService creates a service using the new modular architecture
func (f *ServiceFactory) CreateModernService(articlesPath string) (*ServiceContainer, error) {
	config := &Config{
		ArticlesPath: articlesPath,
		CacheEnabled: true,
		CacheConfig:  DefaultCacheConfig(),
		SearchIndex:  true,
	}

	container, err := NewServiceContainer(config, f.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create service container: %w", err)
	}

	ctx := context.Background()
	if err := container.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start service container: %w", err)
	}

	return container, nil
}

// LegacyServiceAdapter adapts the new service container to implement the legacy interface
type LegacyServiceAdapter struct {
	wrapper   *LegacyServiceWrapper
	container *ServiceContainer
	logger    *slog.Logger
}

// Ensure LegacyServiceAdapter implements ArticleServiceInterface
var _ ArticleServiceInterface = (*LegacyServiceAdapter)(nil)

// Implement all methods from services.ArticleServiceInterface
func (a *LegacyServiceAdapter) GetAllArticles() []*models.Article {
	return a.wrapper.GetAllArticles()
}

func (a *LegacyServiceAdapter) GetArticleBySlug(slug string) (*models.Article, error) {
	return a.wrapper.GetArticleBySlug(slug)
}

func (a *LegacyServiceAdapter) GetArticlesByTag(tag string) []*models.Article {
	return a.wrapper.GetArticlesByTag(tag)
}

func (a *LegacyServiceAdapter) GetArticlesByCategory(category string) []*models.Article {
	return a.wrapper.GetArticlesByCategory(category)
}

func (a *LegacyServiceAdapter) GetArticlesForFeed(limit int) []*models.Article {
	return a.wrapper.GetArticlesForFeed(limit)
}

func (a *LegacyServiceAdapter) GetFeaturedArticles(limit int) []*models.Article {
	return a.wrapper.GetFeaturedArticles(limit)
}

func (a *LegacyServiceAdapter) GetRecentArticles(limit int) []*models.Article {
	return a.wrapper.GetRecentArticles(limit)
}

func (a *LegacyServiceAdapter) GetAllTags() []string {
	return a.wrapper.GetAllTags()
}

func (a *LegacyServiceAdapter) GetAllCategories() []string {
	return a.wrapper.GetAllCategories()
}

func (a *LegacyServiceAdapter) GetTagCounts() []models.TagCount {
	return a.wrapper.GetTagCounts()
}

func (a *LegacyServiceAdapter) GetCategoryCounts() []models.CategoryCount {
	return a.wrapper.GetCategoryCounts()
}

func (a *LegacyServiceAdapter) GetStats() *models.Stats {
	return a.wrapper.GetStats()
}

func (a *LegacyServiceAdapter) ReloadArticles() error {
	return a.wrapper.ReloadArticles()
}

func (a *LegacyServiceAdapter) GetDraftArticles() []*models.Article {
	return a.wrapper.GetDraftArticles()
}

func (a *LegacyServiceAdapter) GetDraftBySlug(slug string) (*models.Article, error) {
	return a.wrapper.GetDraftBySlug(slug)
}

func (a *LegacyServiceAdapter) PreviewDraft(slug string) (*models.Article, error) {
	return a.wrapper.PreviewDraft(slug)
}

func (a *LegacyServiceAdapter) PublishDraft(slug string) error {
	return a.wrapper.PublishDraft(slug)
}

func (a *LegacyServiceAdapter) UnpublishArticle(slug string) error {
	return a.wrapper.UnpublishArticle(slug)
}

// Shutdown gracefully shuts down the service
func (a *LegacyServiceAdapter) Shutdown() error {
	if a.container != nil {
		return a.container.Stop()
	}
	return nil
}

// GetContainer provides access to the underlying container for advanced use cases
func (a *LegacyServiceAdapter) GetContainer() *ServiceContainer {
	return a.container
}

// Health check methods
func (a *LegacyServiceAdapter) IsHealthy() bool {
	return a.container.IsHealthy()
}

func (a *LegacyServiceAdapter) GetHealthStatus() map[string]interface{} {
	return a.container.GetHealthStatus()
}

// Search functionality (these might be accessed through other interfaces)
func (a *LegacyServiceAdapter) SearchArticles(query string, limit int) []*models.SearchResult {
	return a.wrapper.SearchArticles(query, limit)
}

func (a *LegacyServiceAdapter) SearchInTitle(query string, limit int) []*models.SearchResult {
	return a.wrapper.SearchInTitle(query, limit)
}

func (a *LegacyServiceAdapter) GetSearchSuggestions(query string, limit int) []string {
	return a.wrapper.GetSearchSuggestions(query, limit)
}
