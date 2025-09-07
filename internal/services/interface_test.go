package services

import (
	"testing"

	"github.com/vnykmshr/markgo/internal/config"
	"log/slog"
)

// TestInterfaceCompliance ensures all services properly implement their interfaces
func TestInterfaceCompliance(t *testing.T) {
	// ArticleService interface compliance
	t.Run("ArticleService implements ArticleServiceInterface", func(t *testing.T) {
		tempDir := t.TempDir()
		logger := slog.Default()
		service, err := NewArticleService(tempDir, logger)
		if err != nil {
			t.Fatalf("Failed to create ArticleService: %v", err)
		}

		// This will fail to compile if ArticleService doesn't implement the interface
		var _ ArticleServiceInterface = service
	})

	// EmailService interface compliance
	t.Run("EmailService implements EmailServiceInterface", func(t *testing.T) {
		cfg := config.EmailConfig{
			Host: "test.example.com",
			Port: 587,
		}
		logger := slog.Default()
		service := NewEmailService(cfg, logger)

		// This will fail to compile if EmailService doesn't implement the interface
		var _ EmailServiceInterface = service
	})

	// SearchService interface compliance
	t.Run("SearchService implements SearchServiceInterface", func(t *testing.T) {
		service := NewSearchService()

		// This will fail to compile if SearchService doesn't implement the interface
		var _ SearchServiceInterface = service
	})

	// TemplateService interface compliance
	t.Run("TemplateService implements TemplateServiceInterface", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg := &config.Config{
			Blog: config.BlogConfig{
				Title:       "Test Site",
				Description: "Test Description",
			},
			BaseURL: "https://example.com",
		}
		service, err := NewTemplateService(tempDir, cfg)
		if err != nil {
			// Create a minimal template for testing
			service = &TemplateService{}
		}

		// This will fail to compile if TemplateService doesn't implement the interface
		var _ TemplateServiceInterface = service
	})

	// LoggingService interface compliance
	t.Run("LoggingService implements LoggingServiceInterface", func(t *testing.T) {
		cfg := config.LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout", // Add required output field
		}
		service, err := NewLoggingService(cfg)
		if err != nil {
			t.Fatalf("Failed to create LoggingService: %v", err)
		}

		// This will fail to compile if LoggingService doesn't implement the interface
		var _ LoggingServiceInterface = service
	})
}
