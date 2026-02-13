// Package serve provides the HTTP server command for the MarkGo blog platform.
package serve

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/constants"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/handlers"
	"github.com/vnykmshr/markgo/internal/middleware"
	"github.com/vnykmshr/markgo/internal/services"
	"github.com/vnykmshr/markgo/internal/services/article"
	"github.com/vnykmshr/markgo/internal/services/compose"
	"github.com/vnykmshr/markgo/internal/services/feed"
	"github.com/vnykmshr/markgo/internal/services/seo"
	"github.com/vnykmshr/markgo/web"
)

const (
	envDevelopment = "development"
)

// Version is injected via ldflags at build time
var Version = constants.AppVersion

// Run starts the MarkGo HTTP server.
func Run(args []string) {
	// Parse command-line flags
	flagSet := flag.NewFlagSet("serve", flag.ContinueOnError)
	flagSet.SetOutput(os.Stdout)
	port := flagSet.Int("port", 0, "Override server port (default: from .env or 3000)")
	flagSet.Usage = printUsage

	if err := flagSet.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	var logger *slog.Logger
	var server *http.Server
	var templateService *services.TemplateService

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
	var router *gin.Engine
	router, templateService, err = setupServer(cfg, logger)
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

	if templateService != nil {
		templateService.Shutdown()
	}
	logger.Info("Server exited gracefully")
}

func setupServer(cfg *config.Config, logger *slog.Logger) (*gin.Engine, *services.TemplateService, error) {
	// Initialize services
	articleService, err := services.NewArticleService(cfg.ArticlesPath, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("article service: %w", err)
	}

	emailService := services.NewEmailService(&cfg.Email, logger)
	composeService := compose.NewService(cfg.ArticlesPath, cfg.Blog.Author)

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
		return nil, nil, fmt.Errorf("template service: %w", err)
	}

	if err := setupTemplates(router, templateService); err != nil {
		return nil, nil, fmt.Errorf("template setup: %w", err)
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

	// Initialize feed service
	feedService := feed.NewService(articleService, cfg)

	// Initialize markdown renderer for compose preview
	markdownRenderer := article.NewMarkdownContentProcessor(logger)

	// Initialize session store for admin authentication
	sessionStore := middleware.NewSessionStore()
	secureCookie := cfg.Environment != envDevelopment

	// Initialize handlers
	h := handlers.New(&handlers.Config{
		ArticleService:   articleService,
		EmailService:     emailService,
		FeedService:      feedService,
		TemplateService:  templateService,
		SEOService:       seoService,
		ComposeService:   composeService,
		MarkdownRenderer: markdownRenderer,
		SessionStore:     sessionStore,
		SecureCookie:     secureCookie,
		Config:           cfg,
		Logger:           logger,
		BuildInfo: &handlers.BuildInfo{
			Version:   Version,
			GitCommit: constants.GitCommit,
			BuildTime: constants.BuildTime,
		},
	})

	// Session awareness on all routes (sets authenticated=true when valid session exists)
	// Must come after session store init, before route setup
	router.Use(middleware.SessionAware(sessionStore, secureCookie))

	setupRoutes(router, h, sessionStore, secureCookie, cfg, logger)
	return router, templateService, nil
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

func setupRoutes(router *gin.Engine, h *handlers.Router, sessionStore *middleware.SessionStore, secureCookie bool, cfg *config.Config, logger *slog.Logger) { //nolint:funlen // route wiring is inherently long
	// Static files — filesystem first, embedded fallback
	if dirExists(cfg.StaticPath) {
		router.Static("/static", cfg.StaticPath)
		router.StaticFile("/favicon.ico", cfg.StaticPath+"/img/favicon.ico")
		router.StaticFile("/sw.js", cfg.StaticPath+"/sw.js")
	} else {
		staticFS, subErr := fs.Sub(web.Assets, "static")
		if subErr != nil {
			logger.Error("Failed to load embedded static assets — cannot start server", "error", subErr)
			os.Exit(1)
		}
		httpFS := http.FS(staticFS)
		router.StaticFS("/static", httpFS)
		// Serve favicon and sw.js directly from embedded FS (not redirect — SW scope requires root path)
		router.GET("/favicon.ico", func(c *gin.Context) {
			c.FileFromFS("img/favicon.ico", httpFS)
		})
		router.GET("/sw.js", func(c *gin.Context) {
			c.FileFromFS("sw.js", httpFS)
		})
		slog.Info("Using embedded static assets", "checked_path", cfg.StaticPath)
	}
	router.GET("/robots.txt", h.Syndication.RobotsTxt)

	// Health check, metrics, manifest, offline
	router.GET("/health", h.Health.Health)
	router.GET("/manifest.json", h.Health.Manifest)
	router.GET("/offline", h.Health.Offline)
	router.GET("/metrics", h.Admin.Metrics)

	// Public routes
	router.GET("/", h.Feed.Home)
	router.GET("/writing", h.Post.Articles)
	router.GET("/writing/:slug", h.Post.Article)
	router.GET("/tags", h.Taxonomy.Tags)
	router.GET("/tags/:tag", h.Taxonomy.ArticlesByTag)
	router.GET("/categories", h.Taxonomy.Categories)
	router.GET("/categories/:category", h.Taxonomy.ArticlesByCategory)
	router.GET("/search", h.Search.Search)
	router.GET("/about", h.About.ShowAbout)

	// /contact redirects to about page; POST stays for form submission
	router.GET("/contact", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/about#contact")
	})
	contactGroup := router.Group("/contact")
	contactGroup.Use(middleware.RateLimit(cfg.RateLimit.Contact.Requests, cfg.RateLimit.Contact.Window))
	contactGroup.POST("", h.Contact.Submit)

	// Feeds and SEO
	router.GET("/feed.xml", h.Syndication.RSS)
	router.GET("/feed.json", h.Syndication.JSONFeed)
	router.GET("/sitemap.xml", h.Syndication.Sitemap)

	// Login/logout routes (public, CSRF on login POST)
	if h.Auth != nil {
		loginGroup := router.Group("/login")
		loginGroup.Use(middleware.CSRF(secureCookie))
		loginGroup.POST("", h.Auth.HandleLogin)
		router.GET("/logout", h.Auth.HandleLogout)
	}

	// Compose routes — public GET (anyone can draft), protected POST (auth required to publish)
	if cfg.Admin.Username != "" && cfg.Admin.Password != "" && h.Compose != nil {
		// Public compose page (CSRF for form token, no auth required)
		router.GET("/compose", middleware.NoCache(), middleware.CSRF(secureCookie), h.Compose.ShowCompose)

		// Protected compose actions (auth required)
		composeGroup := router.Group("/compose")
		composeGroup.Use(
			middleware.RecoveryWithErrorHandler(logger),
			middleware.SoftSessionAuth(sessionStore, secureCookie),
			middleware.NoCache(),
			middleware.CSRF(secureCookie),
		)
		composeGroup.POST("", h.Compose.HandleSubmit)
		composeGroup.GET("/edit/:slug", h.Compose.ShowEdit)
		composeGroup.POST("/edit/:slug", h.Compose.HandleEdit)
		composeGroup.POST("/preview", h.Compose.Preview)
		composeGroup.POST("/upload", h.Compose.Upload)
		composeGroup.POST("/quick", h.Compose.HandleQuickPublish)
		composeGroup.POST("/publish/:slug", h.Compose.PublishDraft)
	}

	// Admin routes (soft auth — renders login popover when unauthenticated)
	if cfg.Admin.Username != "" && cfg.Admin.Password != "" {
		adminGroup := router.Group("/admin")
		adminGroup.Use(
			middleware.RecoveryWithErrorHandler(logger),
			middleware.SoftSessionAuth(sessionStore, secureCookie),
			middleware.NoCache(),
		)
		adminGroup.GET("", h.Admin.AdminHome)
		adminGroup.GET("/writing", h.Admin.Writing)
		adminGroup.GET("/drafts", middleware.CSRF(secureCookie), h.Admin.Drafts)
		adminGroup.POST("/cache/clear", h.ClearCache)
		adminGroup.GET("/stats", h.Admin.Stats)
		adminGroup.POST("/articles/reload", h.Admin.ReloadArticles)
	}

	// Debug endpoints (development only)
	if cfg.Environment == envDevelopment {
		debugGroup := router.Group("/debug")

		if cfg.Admin.Username != "" && cfg.Admin.Password != "" {
			debugGroup.Use(middleware.SessionAuth(sessionStore))
			logger.Info("Debug endpoints enabled with authentication", "environment", cfg.Environment)
		} else {
			logger.Warn("Debug endpoints enabled WITHOUT authentication - configure ADMIN_USERNAME/PASSWORD for security")
		}

		debugGroup.GET("/memory", h.Admin.Debug)
		debugGroup.GET("/runtime", h.Admin.Debug)
		debugGroup.GET("/config", h.Admin.Debug)
		debugGroup.GET("/requests", h.Admin.Debug)

		// Go pprof profiling endpoints — registered directly via stdlib
		pprofGroup := debugGroup.Group("/pprof")
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
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
		"base.html", "feed.html", "compose.html", "article.html", "articles.html",
		"404.html", "500.html", "offline.html", "about.html", "search.html", "tags.html", "categories.html",
		"drafts.html",
		"admin_writing.html",
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

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
