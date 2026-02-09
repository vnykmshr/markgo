// Package serve provides the HTTP server command for the MarkGo blog platform.
package serve

import (
	"context"
	"flag"
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

// Version is injected via ldflags at build time
var Version = constants.AppVersion

// Run starts the MarkGo HTTP server.
func Run(args []string) {
	// Parse command-line flags
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	port := fs.Int("port", 0, "Override server port (default: from .env or 3000)")
	fs.Usage = printUsage

	if err := fs.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

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

	// Apply CLI overrides
	if *port != 0 {
		if *port < 1 || *port > 65535 {
			fmt.Fprintf(os.Stderr, "Error: port must be between 1 and 65535, got %d\n", *port)
			os.Exit(1)
		}
		cfg.Port = *port
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

	// Initialize services and configure router
	router, err := setupServer(cfg, logger)
	if err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("server setup", "Failed to set up server", err, 1),
			cleanup,
		)
	}

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
			"version", Version)
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

func setupServer(cfg *config.Config, logger *slog.Logger) (*gin.Engine, error) {
	// Initialize services
	articleService, err := services.NewArticleService(cfg.ArticlesPath, logger)
	if err != nil {
		return nil, fmt.Errorf("article service: %w", err)
	}

	// Initialize obcache for handlers with performance optimizations
	cacheConfig := obcache.NewDefaultConfig()
	cacheConfig.MaxEntries = cfg.Cache.MaxSize
	cacheConfig.DefaultTTL = cfg.Cache.TTL

	logger.Info("Initializing cache with performance optimizations",
		"max_entries", cacheConfig.MaxEntries,
		"default_ttl", cacheConfig.DefaultTTL,
		"cache_type", "obcache-go")

	cache, err := obcache.New(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("cache initialization: %w", err)
	}

	emailService := services.NewEmailService(&cfg.Email, logger)
	searchService := services.NewSearchService()

	// Initialize SEO helper (stateless utility)
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
	seoService := seo.NewHelper(articleService, &siteConfig, &robotsConfig, logger, cfg.SEO.Enabled)
	if cfg.SEO.Enabled {
		logger.Info("SEO features enabled")
	}

	// Setup Gin router
	configureGinMode(cfg, logger)
	router := gin.New()

	// Initialize template service
	templateService, err := services.NewTemplateService(cfg.TemplatesPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("template service: %w", err)
	}

	if err := setupTemplates(router, templateService); err != nil {
		return nil, fmt.Errorf("template setup: %w", err)
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
		middleware.RecoveryWithErrorHandler(logger),
		middleware.Logger(logger),
		middleware.Performance(logger),
		middleware.SmartCacheHeaders(),
		middleware.CORS(cfg.CORS.AllowedOrigins, cfg.Environment == envDevelopment),
		middleware.Security(),
		middleware.RateLimit(cfg.RateLimit.General.Requests, cfg.RateLimit.General.Window),
		middleware.ErrorHandler(logger),
	)

	if cfg.Environment == envDevelopment {
		router.Use(middleware.RequestTracker())
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
			Version:   Version,
			GitCommit: constants.GitCommit,
			BuildTime: constants.BuildTime,
		},
	})

	setupRoutes(router, h, cfg, logger)
	return router, nil
}

func configureGinMode(cfg *config.Config, logger *slog.Logger) {
	switch cfg.Environment {
	case "production":
		gin.SetMode(gin.ReleaseMode)
		_ = os.Setenv("GIN_MODE", "release")
		logger.Info("Gin router configured for production", "gin_mode", "release")
	case "test":
		gin.SetMode(gin.TestMode)
		_ = os.Setenv("GIN_MODE", "test")
		logger.Info("Gin router configured for testing", "gin_mode", "test")
	default:
		gin.SetMode(gin.DebugMode)
		_ = os.Setenv("GIN_MODE", "debug")
		logger.Info("Gin router configured for development", "gin_mode", "debug")
	}
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
	contactGroup.GET("", h.ContactForm)
	contactGroup.POST("", h.ContactSubmit)

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
		adminGroup.GET("", h.AdminHome)
		adminGroup.POST("/cache/clear", h.ClearCache)
		adminGroup.GET("/stats", h.AdminStats)
		adminGroup.POST("/articles/reload", h.ReloadArticles)
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

		// Memory and runtime debugging
		debugGroup.GET("/memory", h.DebugMemory)
		debugGroup.GET("/runtime", h.DebugRuntime)
		debugGroup.GET("/config", h.DebugConfig)
		debugGroup.GET("/requests", h.DebugRequests)

		// Go pprof profiling endpoints at /debug/pprof
		pprofGroup := debugGroup.Group("/pprof")
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

	// 404 handler
	router.NoRoute(h.NotFound)
}

func printUsage() {
	fmt.Printf(`markgo serve - Start the blog server

USAGE:
    markgo serve [options]

OPTIONS:
    --port PORT    Override server port (default: from .env or 3000)
    --help         Show this help message

CONFIGURATION:
    Most server settings are configured via .env file.
    Run 'markgo init' to generate a default configuration.
    See docs/configuration.md for all options.

EXAMPLES:
    markgo serve              # Start with .env configuration
    markgo serve --port 8080  # Start on port 8080

`)
}

// setupTemplates configures Gin's HTML template renderer using TemplateService
func setupTemplates(router *gin.Engine, templateService *services.TemplateService) error {
	// Validate that required templates exist
	requiredTemplates := []string{
		"base.html", "index.html", "feed.html", "article.html", "articles.html",
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
