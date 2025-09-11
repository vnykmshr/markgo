package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/markgo/internal/services/export"
)

type ExportConfig struct {
	OutputDir     string
	BaseURL       string
	IncludeDrafts bool
	Verbose       bool
}

func main() {
	// Parse command line arguments
	exportConfig := parseFlags()

	// Load application configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logging
	loggingService, err := services.NewLoggingService(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logging: %v\n", err)
		os.Exit(1)
	}

	logger := loggingService.GetLogger()
	slog.SetDefault(logger)

	// Override base URL if provided
	if exportConfig.BaseURL != "" {
		cfg.BaseURL = exportConfig.BaseURL
	}

	// Set verbose logging if requested
	if exportConfig.Verbose {
		logger.Info("Verbose logging enabled")
	}

	// Initialize services
	articleService, err := services.NewArticleService(cfg.ArticlesPath, logger)
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("article service initialization", "Failed to initialize article service", err, 1),
			nil,
		)
	}

	templateService, err := services.NewTemplateService(cfg.TemplatesPath, cfg)
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("template service initialization", "Failed to initialize template service", err, 1),
			nil,
		)
	}

	searchService := services.NewSearchService()

	// Initialize export service
	exportService, err := export.NewStaticExportService(&export.Config{
		OutputDir:       exportConfig.OutputDir,
		StaticPath:      cfg.StaticPath,
		BaseURL:         cfg.BaseURL,
		ArticleService:  articleService,
		TemplateService: templateService,
		SearchService:   searchService,
		Config:          cfg,
		Logger:          logger,
		IncludeDrafts:   exportConfig.IncludeDrafts,
	})
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("export service initialization", "Failed to initialize export service", err, 1),
			nil,
		)
	}

	// Perform export
	logger.Info("Starting static site export",
		"output_dir", exportConfig.OutputDir,
		"base_url", cfg.BaseURL,
		"include_drafts", exportConfig.IncludeDrafts)

	ctx := context.Background()
	if err := exportService.Export(ctx); err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("export", "Failed to export static site", err, 1),
			nil,
		)
	}

	logger.Info("Static site export completed successfully", "output_dir", exportConfig.OutputDir)
	fmt.Printf("✅ Static site exported successfully to: %s\n", exportConfig.OutputDir)
}

func parseFlags() *ExportConfig {
	config := &ExportConfig{
		OutputDir:     "./dist",
		BaseURL:       "",
		IncludeDrafts: false,
		Verbose:       false,
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				config.OutputDir = args[i+1]
				i++
			}
		case "--base-url", "-u":
			if i+1 < len(args) {
				config.BaseURL = args[i+1]
				i++
			}
		case "--include-drafts", "-d":
			config.IncludeDrafts = true
		case "--verbose", "-v":
			config.Verbose = true
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		}
	}

	// Ensure output directory is absolute
	if !filepath.IsAbs(config.OutputDir) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
			os.Exit(1)
		}
		config.OutputDir = filepath.Join(wd, config.OutputDir)
	}

	return config
}

func printUsage() {
	fmt.Printf(`MarkGo Static Site Export

Usage:
  markgo-export [options]

Options:
  -o, --output DIR        Output directory for static files (default: ./dist)
  -u, --base-url URL      Base URL for the site (overrides config)
  -d, --include-drafts    Include draft articles in export
  -v, --verbose           Enable verbose logging
  -h, --help              Show this help message

Examples:
  markgo-export
  markgo-export --output ./public --base-url https://myblog.github.io
  markgo-export --include-drafts --verbose

`)
}
