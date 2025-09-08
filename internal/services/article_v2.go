package services

import (
	"log/slog"

	"github.com/vnykmshr/markgo/internal/services/article"
)

// NewArticleServiceV2 creates a new modular article service using the new architecture
// This replaces the legacy ArticleService with enterprise-grade performance and modularity
func NewArticleServiceV2(articlesPath string, logger *slog.Logger) (ArticleServiceInterface, error) {
	factory := article.NewServiceFactory(logger)
	return factory.CreateLegacyCompatibleService(articlesPath)
}
