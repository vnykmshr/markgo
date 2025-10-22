// Package main provides the main HTTP server for the MarkGo blog platform.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vnykmshr/obcache-go/pkg/obcache"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/constants"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/handlers"
	"github.com/vnykmshr/markgo/internal/middleware"
	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/markgo/internal/services/seo"
)

const (
	envDevelopment = "development"
)

// Build-time variables injected via ldflags
var (
	version   = constants.AppVersion // fallback to constant
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	var logger *slog.Logger
	var server *http.Server

	// Cleanup function for graceful shutdown
	cleanup := func() {
		if server != nil && logger != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			logger.Info("Performing graceful shutdown...")

			if err := server.Shutdown(ctx); err != nil {
				logger.Error("Error during shutdown", "error", err)
			}
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("configuration loading", "Failed to load configuration", err, 1),
			cleanup,
		)
	}

	// Setup enhanced logging with configuration
	loggingService, err := services.NewLoggingService(&cfg.Logging)
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("logging initialization", "Failed to initialize logging service", err, 1),
			cleanup,
		)
	}

	logger = loggingService.GetLogger()
	slog.SetDefault(logger)

	// Initialize services - using modular architecture
	articleService, err := services.NewArticleService(cfg.ArticlesPath, logger)
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("article service initialization", "Failed to initialize article service", err, 1),
			cleanup,
		)
	}

	// Initialize obcache for handlers with performance optimizations
	cacheConfig := obcache.NewDefaultConfig()
	cacheConfig.MaxEntries = cfg.Cache.MaxSize
	cacheConfig.DefaultTTL = cfg.Cache.TTL

	// Set performance-optimized values for available configuration
	// Note: obcache-go uses internal optimizations, so we focus on the key settings
	logger.Info("Initializing cache with performance optimizations",
		"max_entries", cacheConfig.MaxEntries,
		"default_ttl", cacheConfig.DefaultTTL,
		"cache_type", "obcache-go")

	cache, err := obcache.New(cacheConfig)
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("cache initialization", "Failed to initialize cache", err, 1),
			cleanup,
		)
	}
	emailService := services.NewEmailService(&cfg.Email, logger)
	searchService := services.NewSearchService()

	// Initialize SEO service
	siteConfig := services.SiteConfig{
		Name:        cfg.Blog.Title,
		Description: cfg.Blog.Description,
		BaseURL:     cfg.BaseURL,
		Language:    cfg.Blog.Language,
		Author:      cfg.Blog.Author,
	}
	robotsConfig := services.RobotsConfig{
		UserAgent:  "*",
		Allow:      cfg.SEO.RobotsAllowed,
		Disallow:   cfg.SEO.RobotsDisallowed,
		CrawlDelay: cfg.SEO.RobotsCrawlDelay,
		SitemapURL: cfg.BaseURL + "/sitemap.xml",
	}
	seoService := seo.NewService(articleService, siteConfig, robotsConfig, logger, cfg.SEO.Enabled)
	if cfg.SEO.Enabled {
		if err := seoService.Start(); err != nil {
			logger.Error("Failed to start SEO service", "error", err)
		}
	}

	// Setup Gin router - ensure Gin mode matches application environment
	switch cfg.Environment {
	case "production":
		gin.SetMode(gin.ReleaseMode)
		_ = os.Setenv("GIN_MODE", "release")
		logger.Info("Gin router configured for production", "gin_mode", "release")
	case "test":
		gin.SetMode(gin.TestMode)
		_ = os.Setenv("GIN_MODE", "test")
		logger.Info("Gin router configured for testing", "gin_mode", "test")
	default: // development
		gin.SetMode(gin.DebugMode)
		_ = os.Setenv("GIN_MODE", "debug")
		logger.Info("Gin router configured for development", "gin_mode", "debug")
	}

	router := gin.New()

	// Initialize template service
	templateService, err := services.NewTemplateService(cfg.TemplatesPath, cfg)
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("template service initialization", "Failed to initialize template service", err, 1),
			cleanup,
		)
	}

	// Setup HTML templates
	if err := setupTemplates(router, templateService); err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("template setup", "Failed to setup templates", err, 1),
			cleanup,
		)
	}

	// Log rate limiting configuration
	logger.Info("Rate limiting configuration",
		"environment", cfg.Environment,
		"general_requests", cfg.RateLimit.General.Requests,
		"general_window", cfg.RateLimit.General.Window,
		"general_rate_per_sec", float64(cfg.RateLimit.General.Requests)/(cfg.RateLimit.General.Window.Minutes()*60),
		"contact_requests", cfg.RateLimit.Contact.Requests,
		"contact_window", cfg.RateLimit.Contact.Window)

	// Global middleware
	router.Use(
		middleware.RecoveryWithErrorHandler(logger),                                 // Custom recovery with error handling
		middleware.Logger(logger),                                                   // Basic request logging
		middleware.Performance(logger),                                              // Performance monitoring
		middleware.SmartCacheHeaders(),                                              // Intelligent HTTP cache headers
		middleware.CORS(cfg.CORS.AllowedOrigins, cfg.Environment == envDevelopment), // Secure CORS with exact origin matching
		middleware.Security(),
		middleware.RateLimit(cfg.RateLimit.General.Requests, cfg.RateLimit.General.Window),
		middleware.ErrorHandler(logger), // Centralized error handling (must be last)
	)

	// Development-specific enhanced logging
	if cfg.Environment == envDevelopment {
		router.Use(middleware.RequestTracker(logger, cfg.Environment))
		logger.Info("Development logging enhancements enabled")
	}

	// Initialize handlers
	h := handlers.New(&handlers.Config{
		ArticleService:  articleService,
		EmailService:    emailService,
		SearchService:   searchService,
		TemplateService: templateService,
		SEOService:      seoService,
		Config:          cfg,
		Logger:          logger,
		Cache:           cache,
		BuildInfo: &handlers.BuildInfo{
			Version:   version,
			GitCommit: gitCommit,
			BuildTime: buildTime,
		},
	})

	// Setup routes
	setupRoutes(router, h, cfg, logger)

	// Create HTTP server
	server = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting MarkGo server",
			"port", cfg.Port,
			"environment", cfg.Environment,
			"version", version)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			apperrors.HandleCLIError(
				apperrors.NewCLIError("server startup", "Server failed to start", err, 1),
				cleanup,
			)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		apperrors.HandleCLIError(
			apperrors.NewCLIError("server shutdown", "Server forced to shutdown", err, 1),
			nil, // No additional cleanup needed here
		)
	}

	logger.Info("Server exited gracefully")
}

func setupRoutes(router *gin.Engine, h *handlers.Handlers, cfg *config.Config, logger *slog.Logger) {
	// Validation middleware removed - keeping it simple

	// Static files
	router.Static("/static", cfg.StaticPath)
	router.StaticFile("/favicon.ico", cfg.StaticPath+"/img/favicon.ico")

	router.StaticFile("/robots.txt", cfg.StaticPath+"/robots.txt") // Keep static for now, can be made dynamic later

	// Health check and metrics
	router.GET("/health", h.Health)
	router.GET("/metrics", h.Metrics)

	// Main routes
	router.GET("/", h.Home)

	// Articles with pagination validation
	router.GET("/articles", h.Articles)

	// Article by slug with slug validation
	router.GET("/articles/:slug", h.Article)

	router.GET("/tags", h.Tags)

	// Tag filtering with tag validation and pagination
	router.GET("/tags/:tag", h.ArticlesByTag)

	router.GET("/categories", h.Categories)

	// Category filtering with category validation and pagination
	router.GET("/categories/:category", h.ArticlesByCategory)

	// Search with query validation and pagination
	router.GET("/search", h.Search)
	router.GET("/about", h.AboutArticle)

	// Contact form with rate limiting and input validation
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

	// Admin routes (optional) - with minimal middleware chain to avoid header conflicts
	if cfg.Admin.Username != "" && cfg.Admin.Password != "" {
		adminGroup := router.Group("/admin")
		// Use only essential middleware for admin routes to avoid header conflicts
		adminGroup.Use(
			middleware.RecoveryWithErrorHandler(logger),                  // Recovery first
			middleware.BasicAuth(cfg.Admin.Username, cfg.Admin.Password), // Auth second, before any header-writing middleware
			middleware.NoCache(), // No caching for admin
		)
		{
			adminGroup.GET("", h.AdminHome) // Admin dashboard/home page
			adminGroup.POST("/cache/clear", h.ClearCache)
			adminGroup.GET("/stats", h.AdminStats)
			adminGroup.POST("/articles/reload", h.ReloadArticles)

			// SEO management endpoints (placeholder for now)
			// TODO: Implement proper SEO admin interface
			adminGroup.GET("/seo", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "SEO admin interface - under development",
					"enabled": cfg.SEO.Enabled,
				})
			})
		}
	}

	// Debug endpoints (development only, with auth if admin configured)
	if cfg.Environment == envDevelopment {
		debugGroup := router.Group("/debug")

		// Add Basic Auth if admin credentials are configured
		if cfg.Admin.Username != "" && cfg.Admin.Password != "" {
			debugGroup.Use(middleware.BasicAuth(cfg.Admin.Username, cfg.Admin.Password))
			logger.Info("Debug endpoints enabled with authentication", "environment", cfg.Environment)
		} else {
			logger.Warn("Debug endpoints enabled WITHOUT authentication - configure ADMIN_USERNAME/PASSWORD for security")
		}

		{
			// Memory and runtime debugging
			debugGroup.GET("/memory", h.DebugMemory)
			debugGroup.GET("/runtime", h.DebugRuntime)
			debugGroup.GET("/config", h.DebugConfig)
			debugGroup.GET("/requests", h.DebugRequests)
			debugGroup.POST("/log-level", h.SetLogLevel)

			// Go pprof profiling endpoints at /debug/pprof
			pprofGroup := debugGroup.Group("/pprof")
			{
				pprofGroup.GET("/", h.PprofIndex)
				pprofGroup.GET("/cmdline", gin.WrapH(http.HandlerFunc(pprof.Cmdline)))
				pprofGroup.GET("/profile", h.PprofProfile)
				pprofGroup.GET("/symbol", gin.WrapH(http.HandlerFunc(pprof.Symbol)))
				pprofGroup.GET("/trace", h.PprofTrace)
				pprofGroup.GET("/heap", h.PprofHeap)
				pprofGroup.GET("/goroutine", h.PprofGoroutine)
				pprofGroup.GET("/allocs", h.PprofAllocs)
				pprofGroup.GET("/block", h.PprofBlock)
				pprofGroup.GET("/mutex", h.PprofMutex)
			}
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
