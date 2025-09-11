package handlers // import "github.com/vnykmshr/markgo/internal/handlers"


TYPES

type APIHandler struct {
	*BaseHandler

	// Has unexported fields.
}
    APIHandler handles API HTTP requests (RSS, JSON Feed, Sitemap, Health)

func NewAPIHandler(
	config *config.Config,
	logger *slog.Logger,
	templateService services.TemplateServiceInterface,
	articleService services.ArticleServiceInterface,
	emailService services.EmailServiceInterface,
	startTime time.Time,
	cachedFunctions CachedAPIFunctions,
) *APIHandler
    NewAPIHandler creates a new API handler

func (h *APIHandler) Contact(c *gin.Context)
    Contact handles contact form submissions

func (h *APIHandler) Health(c *gin.Context)
    Health handles health check requests

func (h *APIHandler) JSONFeed(c *gin.Context)
    JSONFeed handles JSON feed generation

func (h *APIHandler) RSS(c *gin.Context)
    RSS handles RSS feed generation

func (h *APIHandler) Sitemap(c *gin.Context)
    Sitemap handles sitemap.xml generation

type AdminHandler struct {
	*BaseHandler

	// Has unexported fields.
}
    AdminHandler handles administrative HTTP requests

func NewAdminHandler(
	config *config.Config,
	logger *slog.Logger,
	templateService services.TemplateServiceInterface,
	articleService services.ArticleServiceInterface,
	startTime time.Time,
	cachedFunctions CachedAdminFunctions,
) *AdminHandler
    NewAdminHandler creates a new admin handler

func (h *AdminHandler) AdminHome(c *gin.Context)
    AdminHome handles the admin dashboard/home page

func (h *AdminHandler) CompactMemory(c *gin.Context)
    CompactMemory handles manual memory compaction (development only)

func (h *AdminHandler) Debug(c *gin.Context)
    Debug handles debug information endpoint

func (h *AdminHandler) Metrics(c *gin.Context)
    Metrics handles the metrics endpoint for performance monitoring

func (h *AdminHandler) ProfileAllocs(c *gin.Context)

func (h *AdminHandler) ProfileBlock(c *gin.Context)

func (h *AdminHandler) ProfileCmdline(c *gin.Context)

func (h *AdminHandler) ProfileGoroutine(c *gin.Context)

func (h *AdminHandler) ProfileHeap(c *gin.Context)

func (h *AdminHandler) ProfileIndex(c *gin.Context)
    Profile handlers for pprof (development only)

func (h *AdminHandler) ProfileMutex(c *gin.Context)

func (h *AdminHandler) ProfileProfile(c *gin.Context)

func (h *AdminHandler) ProfileSymbol(c *gin.Context)

func (h *AdminHandler) ProfileTrace(c *gin.Context)

func (h *AdminHandler) ReloadArticles(c *gin.Context)
    ReloadArticles handles reloading articles (development only)

func (h *AdminHandler) SetLogLevel(c *gin.Context)
    SetLogLevel handles dynamic log level changes

func (h *AdminHandler) Stats(c *gin.Context)
    Stats handles the stats endpoint with cached/fallback pattern

type ArticleHandler struct {
	*BaseHandler

	// Has unexported fields.
}
    ArticleHandler handles article-related HTTP requests

func NewArticleHandler(
	config *config.Config,
	logger *slog.Logger,
	templateService services.TemplateServiceInterface,
	articleService services.ArticleServiceInterface,
	searchService services.SearchServiceInterface,
	cachedFunctions CachedArticleFunctions,
) *ArticleHandler
    NewArticleHandler creates a new article handler

func (h *ArticleHandler) Article(c *gin.Context)
    Article handles individual article requests

func (h *ArticleHandler) Articles(c *gin.Context)
    Articles handles the articles listing page

func (h *ArticleHandler) ArticlesByCategory(c *gin.Context)
    ArticlesByCategory handles articles filtered by category

func (h *ArticleHandler) ArticlesByTag(c *gin.Context)
    ArticlesByTag handles articles filtered by tag

func (h *ArticleHandler) Categories(c *gin.Context)
    Categories handles the categories page

func (h *ArticleHandler) Home(c *gin.Context)
    Home handles the home page request

func (h *ArticleHandler) Search(c *gin.Context)
    Search handles search requests

func (h *ArticleHandler) Tags(c *gin.Context)
    Tags handles the tags page

type BaseHandler struct {
	// Has unexported fields.
}
    BaseHandler provides common functionality for all handlers

func NewBaseHandler(config *config.Config, logger *slog.Logger, templateService services.TemplateServiceInterface) *BaseHandler
    NewBaseHandler creates a new base handler

type CacheAdapter interface {
	Clear()
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Stats() map[string]interface{}
}
    CacheAdapter defines the interface for cache operations

type CachedAPIFunctions struct {
	GetRSSFeed  func() (string, error)
	GetJSONFeed func() (string, error)
	GetSitemap  func() (string, error)
}
    CachedAPIFunctions holds cached functions for API operations

type CachedAdminFunctions struct {
	GetStatsData func() (map[string]any, error)
}
    CachedAdminFunctions holds cached functions for admin operations

type CachedArticleFunctions struct {
	GetHomeData         func() (map[string]any, error)
	GetArticleData      func(string) (map[string]any, error)
	GetArticlesPage     func(int) (map[string]any, error)
	GetTagArticles      func(string) (map[string]any, error)
	GetCategoryArticles func(string) (map[string]any, error)
	GetSearchResults    func(string) (map[string]any, error)
	GetTagsPage         func() (map[string]any, error)
	GetCategoriesPage   func() (map[string]any, error)
}
    CachedArticleFunctions holds cached functions for article operations

type Config struct {
	ArticleService  services.ArticleServiceInterface
	EmailService    services.EmailServiceInterface
	SearchService   services.SearchServiceInterface
	TemplateService services.TemplateServiceInterface
	Config          *config.Config
	Logger          *slog.Logger
	Cache           *obcache.Cache
}
    Config for handler initialization

type Handlers struct {
	ArticleHandler *ArticleHandler
	AdminHandler   *AdminHandler
	APIHandler     *APIHandler
	// Has unexported fields.
}
    Handlers composes all handler types for route registration

func New(cfg *Config) *Handlers
    New creates a new composed handlers instance

func (h *Handlers) AboutArticle(c *gin.Context)

func (h *Handlers) AdminHome(c *gin.Context)
    Admin route methods

func (h *Handlers) AdminStats(c *gin.Context)

func (h *Handlers) Article(c *gin.Context)

func (h *Handlers) Articles(c *gin.Context)

func (h *Handlers) ArticlesByCategory(c *gin.Context)

func (h *Handlers) ArticlesByTag(c *gin.Context)

func (h *Handlers) Categories(c *gin.Context)

func (h *Handlers) ClearCache(c *gin.Context)

func (h *Handlers) ContactForm(c *gin.Context)

func (h *Handlers) ContactSubmit(c *gin.Context)

func (h *Handlers) DebugConfig(c *gin.Context)

func (h *Handlers) DebugMemory(c *gin.Context)
    Debug route methods

func (h *Handlers) DebugRequests(c *gin.Context)

func (h *Handlers) DebugRuntime(c *gin.Context)

func (h *Handlers) GetDraftBySlug(c *gin.Context)

func (h *Handlers) GetDrafts(c *gin.Context)
    Draft management

func (h *Handlers) Health(c *gin.Context)

func (h *Handlers) Home(c *gin.Context)
    Article route methods

func (h *Handlers) JSONFeed(c *gin.Context)

func (h *Handlers) Logger() *slog.Logger
    Logger returns the logger instance (used by middleware)

func (h *Handlers) Metrics(c *gin.Context)

func (h *Handlers) NotFound(c *gin.Context)
    NotFound handles 404 errors

func (h *Handlers) PprofAllocs(c *gin.Context)

func (h *Handlers) PprofBlock(c *gin.Context)

func (h *Handlers) PprofGoroutine(c *gin.Context)

func (h *Handlers) PprofHeap(c *gin.Context)

func (h *Handlers) PprofIndex(c *gin.Context)

func (h *Handlers) PprofMutex(c *gin.Context)

func (h *Handlers) PprofProfile(c *gin.Context)

func (h *Handlers) PprofTrace(c *gin.Context)

func (h *Handlers) PreviewDraft(c *gin.Context)

func (h *Handlers) PublishDraft(c *gin.Context)

func (h *Handlers) RSSFeed(c *gin.Context)
    API route methods

func (h *Handlers) ReloadArticles(c *gin.Context)

func (h *Handlers) Search(c *gin.Context)

func (h *Handlers) SetLogLevel(c *gin.Context)

func (h *Handlers) Sitemap(c *gin.Context)

func (h *Handlers) Tags(c *gin.Context)

func (h *Handlers) UnpublishArticle(c *gin.Context)

type ObcacheAdapter struct {
	// Has unexported fields.
}
    ObcacheAdapter provides cache interface using obcache

func (a *ObcacheAdapter) Clear()

func (a *ObcacheAdapter) Delete(key string)

func (a *ObcacheAdapter) Get(key string) (interface{}, bool)

func (a *ObcacheAdapter) Set(key string, value interface{}, ttl time.Duration)

func (a *ObcacheAdapter) Stats() map[string]interface{}

