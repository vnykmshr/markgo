package services

import (
	"log/slog"

	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services/article"
)

// NewArticleServiceV2 creates a new modular article service using the new architecture
// This is the transitional service that can be used alongside or instead of the existing ArticleService
func NewArticleServiceV2(articlesPath string, logger *slog.Logger) (ArticleServiceInterface, error) {
	factory := article.NewServiceFactory(logger)
	return factory.CreateLegacyCompatibleService(articlesPath)
}

// ArticleServiceV2Adapter provides a bridge between the old and new service architectures
// This allows for gradual migration while maintaining compatibility
type ArticleServiceV2Adapter struct {
	legacyService ArticleServiceInterface   // Current service
	modernService *article.ServiceContainer // New modular service
	useLegacy     bool                      // Flag to control which service to use
	logger        *slog.Logger
}

// NewArticleServiceV2Adapter creates an adapter that can switch between implementations
func NewArticleServiceV2Adapter(articlesPath string, logger *slog.Logger, useLegacy bool) (*ArticleServiceV2Adapter, error) {
	adapter := &ArticleServiceV2Adapter{
		useLegacy: useLegacy,
		logger:    logger,
	}

	// Initialize legacy service
	legacyService, err := NewArticleService(articlesPath, logger)
	if err != nil {
		return nil, err
	}
	adapter.legacyService = legacyService

	// Initialize modern service
	factory := article.NewServiceFactory(logger)
	modernService, err := factory.CreateModernService(articlesPath)
	if err != nil {
		logger.Warn("Failed to initialize modern service, using legacy only", "error", err)
	} else {
		adapter.modernService = modernService
	}

	return adapter, nil
}

// SwitchToModern switches to using the modern service implementation
func (a *ArticleServiceV2Adapter) SwitchToModern() {
	if a.modernService != nil {
		a.useLegacy = false
		a.logger.Info("Switched to modern article service implementation")
	}
}

// SwitchToLegacy switches to using the legacy service implementation
func (a *ArticleServiceV2Adapter) SwitchToLegacy() {
	a.useLegacy = true
	a.logger.Info("Switched to legacy article service implementation")
}

// GetImplementationStatus returns which implementation is currently active
func (a *ArticleServiceV2Adapter) GetImplementationStatus() map[string]interface{} {
	status := map[string]interface{}{
		"using_legacy":     a.useLegacy,
		"legacy_available": a.legacyService != nil,
		"modern_available": a.modernService != nil,
		"implementation":   "legacy",
	}

	if !a.useLegacy && a.modernService != nil {
		status["implementation"] = "modern"
		status["modern_health"] = a.modernService.GetHealthStatus()
	}

	return status
}

// Implement all ArticleServiceInterface methods by forwarding to the appropriate service

func (a *ArticleServiceV2Adapter) GetAllArticles() []*models.Article {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetAllArticles()
	}
	return a.modernService.GetService().GetAllArticles()
}

func (a *ArticleServiceV2Adapter) GetArticleBySlug(slug string) (*models.Article, error) {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetArticleBySlug(slug)
	}
	return a.modernService.GetService().GetArticleBySlug(slug)
}

func (a *ArticleServiceV2Adapter) GetArticlesByTag(tag string) []*models.Article {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetArticlesByTag(tag)
	}
	return a.modernService.GetService().GetArticlesByTag(tag)
}

func (a *ArticleServiceV2Adapter) GetArticlesByCategory(category string) []*models.Article {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetArticlesByCategory(category)
	}
	return a.modernService.GetService().GetArticlesByCategory(category)
}

func (a *ArticleServiceV2Adapter) GetArticlesForFeed(limit int) []*models.Article {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetArticlesForFeed(limit)
	}
	return a.modernService.GetService().GetArticlesForFeed(limit)
}

func (a *ArticleServiceV2Adapter) GetFeaturedArticles(limit int) []*models.Article {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetFeaturedArticles(limit)
	}
	return a.modernService.GetService().GetFeaturedArticles(limit)
}

func (a *ArticleServiceV2Adapter) GetRecentArticles(limit int) []*models.Article {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetRecentArticles(limit)
	}
	return a.modernService.GetService().GetRecentArticles(limit)
}

func (a *ArticleServiceV2Adapter) GetAllTags() []string {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetAllTags()
	}
	return a.modernService.GetService().GetAllTags()
}

func (a *ArticleServiceV2Adapter) GetAllCategories() []string {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetAllCategories()
	}
	return a.modernService.GetService().GetAllCategories()
}

func (a *ArticleServiceV2Adapter) GetTagCounts() []models.TagCount {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetTagCounts()
	}
	return a.modernService.GetService().GetTagCounts()
}

func (a *ArticleServiceV2Adapter) GetCategoryCounts() []models.CategoryCount {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetCategoryCounts()
	}
	return a.modernService.GetService().GetCategoryCounts()
}

func (a *ArticleServiceV2Adapter) GetStats() *models.Stats {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetStats()
	}
	return a.modernService.GetService().GetStats()
}

func (a *ArticleServiceV2Adapter) ReloadArticles() error {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.ReloadArticles()
	}
	return a.modernService.GetService().ReloadArticles()
}

func (a *ArticleServiceV2Adapter) GetDraftArticles() []*models.Article {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetDraftArticles()
	}
	return a.modernService.GetService().GetDraftArticles()
}

func (a *ArticleServiceV2Adapter) GetDraftBySlug(slug string) (*models.Article, error) {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.GetDraftBySlug(slug)
	}
	return a.modernService.GetService().GetDraftBySlug(slug)
}

func (a *ArticleServiceV2Adapter) PreviewDraft(slug string) (*models.Article, error) {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.PreviewDraft(slug)
	}
	return a.modernService.GetService().PreviewDraft(slug)
}

func (a *ArticleServiceV2Adapter) PublishDraft(slug string) error {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.PublishDraft(slug)
	}
	return a.modernService.GetService().PublishDraft(slug)
}

func (a *ArticleServiceV2Adapter) UnpublishArticle(slug string) error {
	if a.useLegacy || a.modernService == nil {
		return a.legacyService.UnpublishArticle(slug)
	}
	return a.modernService.GetService().UnpublishArticle(slug)
}

// Shutdown gracefully shuts down both services
func (a *ArticleServiceV2Adapter) Shutdown() error {
	var err error

	if a.modernService != nil {
		if modernErr := a.modernService.Stop(); modernErr != nil {
			a.logger.Error("Error shutting down modern service", "error", modernErr)
			err = modernErr
		}
	}

	// If legacy service has a shutdown method, call it
	// Note: The current ArticleService doesn't have a Shutdown method in the interface
	// but individual implementations might have cleanup methods

	return err
}

// GetModernService provides access to the modern service container for advanced usage
func (a *ArticleServiceV2Adapter) GetModernService() *article.ServiceContainer {
	return a.modernService
}

// GetLegacyService provides access to the legacy service for comparison
func (a *ArticleServiceV2Adapter) GetLegacyService() ArticleServiceInterface {
	return a.legacyService
}
