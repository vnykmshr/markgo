package services

import (
	"log/slog"

	"github.com/vnykmshr/markgo/internal/services/article"
)

// NewArticleService creates a new modular article service
// Built with enterprise-grade performance, caching, and modularity
func NewArticleService(articlesPath string, logger *slog.Logger) (ArticleServiceInterface, error) {
	factory := article.NewServiceFactory(logger)
	return factory.CreateService(articlesPath)
}
