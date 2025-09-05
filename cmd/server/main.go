package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/handlers"
	"github.com/vnykmshr/markgo/internal/middleware"
	"github.com/vnykmshr/markgo/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Setup enhanced logging with configuration
	loggingService, err := services.NewLoggingService(cfg.Logging)
	if err != nil {
		slog.Error("Failed to initialize logging service", "error", err)
		os.Exit(1)
	}

	logger := loggingService.GetLogger()
	slog.SetDefault(logger)

	// Initialize services
	articleService, err := services.NewArticleService(cfg.ArticlesPath, logger)
	if err != nil {
		slog.Error("Failed to initialize article service", "error", err)
		os.Exit(1)
	}

	cacheService := services.NewCacheService(cfg.Cache.TTL, cfg.Cache.MaxSize)
	emailService := services.NewEmailService(cfg.Email, logger)
	searchService := services.NewSearchService()

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Initialize template service
	templateService, err := services.NewTemplateService(cfg.TemplatesPath, cfg)
	if err != nil {
		slog.Error("Failed to initialize template service", "error", err)
		os.Exit(1)
	}

	// Setup HTML templates
	if err := setupTemplates(router, templateService); err != nil {
		slog.Error("Failed to setup templates", "error", err)
		os.Exit(1)
	}

	// Global middleware
	router.Use(
		middleware.RecoveryWithErrorHandler(logger), // Custom recovery with error handling
		middleware.Logger(logger),
		middleware.PerformanceMiddleware(logger),
		middleware.CompetitorBenchmarkMiddleware(),
		middleware.CORS(cfg.CORS),
		middleware.Security(),
		middleware.RateLimit(cfg.RateLimit.General.Requests, cfg.RateLimit.General.Window),
		middleware.ErrorHandler(logger), // Centralized error handling (must be last)
	)

	// Initialize handlers
	h := handlers.New(&handlers.Config{
		ArticleService:  articleService,
		CacheService:    cacheService,
		EmailService:    emailService,
		SearchService:   searchService,
		TemplateService: templateService,
		Config:          cfg,
		Logger:          logger,
	})

	// Setup routes
	setupRoutes(router, h, cfg)

	// Setup template hot-reload for development
	if cfg.Environment == "development" {
		setupTemplateHotReload(templateService, cfg.TemplatesPath, logger)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting MarkGo server",
			"port", cfg.Port,
			"environment", cfg.Environment,
			"version", "2.0.0")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited gracefully")
}

func setupRoutes(router *gin.Engine, h *handlers.Handlers, cfg *config.Config) {
	// Static files
	router.Static("/static", cfg.StaticPath)
	router.StaticFile("/favicon.ico", cfg.StaticPath+"/img/favicon.ico")
	router.StaticFile("/robots.txt", cfg.StaticPath+"/robots.txt")

	// Health check and metrics
	router.GET("/health", h.Health)
	router.GET("/metrics", h.Metrics)

	// Main routes
	router.GET("/", h.Home)
	router.GET("/articles", h.Articles)
	router.GET("/articles/:slug", h.Article)
	router.GET("/tags", h.Tags)
	router.GET("/tags/:tag", h.ArticlesByTag)
	router.GET("/categories", h.Categories)
	router.GET("/categories/:category", h.ArticlesByCategory)
	router.GET("/search", h.Search)
	router.GET("/about", h.AboutArticle)

	// Contact form with rate limiting
	contactGroup := router.Group("/contact")
	contactGroup.Use(middleware.RateLimit(cfg.RateLimit.Contact.Requests, cfg.RateLimit.Contact.Window))
	{
		contactGroup.GET("", h.ContactForm)
		contactGroup.POST("", h.ContactSubmit)
	}

	// Feeds and SEO
	router.GET("/feed.xml", h.RSSFeed)
	router.GET("/feed.json", h.JSONFeed)
	router.GET("/sitemap.xml", h.Sitemap)

	// Admin routes (optional)
	if cfg.Admin.Username != "" && cfg.Admin.Password != "" {
		adminGroup := router.Group("/admin")
		adminGroup.Use(middleware.BasicAuth(cfg.Admin.Username, cfg.Admin.Password))
		{
			adminGroup.POST("/cache/clear", h.ClearCache)
			adminGroup.GET("/stats", h.AdminStats)
			adminGroup.POST("/articles/reload", h.ReloadArticles)
		}
	}

	// 404 handler
	router.NoRoute(h.NotFound)
}

// setupTemplates configures Gin's HTML template renderer using TemplateService
func setupTemplates(router *gin.Engine, templateService *services.TemplateService) error {
	// Validate that required templates exist
	requiredTemplates := []string{
		"base.html", "index.html", "article.html", "articles.html",
		"404.html", "contact.html", "search.html", "tags.html", "categories.html",
	}

	for _, tmplName := range requiredTemplates {
		if !templateService.HasTemplate(tmplName) {
			return fmt.Errorf("required template %s not found", tmplName)
		}
	}

	// Get the internal template from TemplateService
	tmpl := templateService.GetTemplate()
	if tmpl == nil {
		return fmt.Errorf("template service has no loaded templates")
	}

	// Set the HTML template renderer
	router.SetHTMLTemplate(tmpl)

	return nil
}

// setupTemplateHotReload sets up file watching for template hot-reload in development
func setupTemplateHotReload(templateService *services.TemplateService, templatesPath string, logger *slog.Logger) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("Failed to create template watcher", "error", err)
		return
	}

	// Add templates directory to watcher
	err = watcher.Add(templatesPath)
	if err != nil {
		logger.Error("Failed to watch templates directory", "error", err)
		return
	}

	// Start watching in a goroutine
	go func() {
		defer watcher.Close()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Only reload on write/create events for HTML files
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					if filepath.Ext(event.Name) == ".html" {
						logger.Debug("Template file changed, reloading", "file", event.Name)
						if err := templateService.Reload(templatesPath); err != nil {
							logger.Error("Failed to reload templates", "error", err)
						} else {
							logger.Info("Templates reloaded successfully")
						}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("Template watcher error", "error", err)
			}
		}
	}()

	logger.Info("Template hot-reload enabled", "path", templatesPath)
}
