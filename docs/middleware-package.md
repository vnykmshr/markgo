package middleware // import "github.com/vnykmshr/markgo/internal/middleware"


FUNCTIONS

func BasicAuth(username, password string) gin.HandlerFunc
    BasicAuth provides basic HTTP authentication

func CORS() gin.HandlerFunc
    CORS handles cross-origin requests

func CompetitorBenchmarkMiddleware() gin.HandlerFunc
    CompetitorBenchmarkMiddleware is a no-op placeholder

func Compress() gin.HandlerFunc
    Compress enables gzip compression

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc
    ErrorHandler provides centralized error handling

func Logger(logger *slog.Logger) gin.HandlerFunc
    Logger provides basic request logging

func NoCache() gin.HandlerFunc
    NoCache adds no-cache headers

func Performance(logger *slog.Logger) gin.HandlerFunc
    Performance logs request timing and basic metrics

func PerformanceLoggingMiddleware(loggingService interface{}) gin.HandlerFunc
    PerformanceLoggingMiddleware provides detailed performance logging

func PerformanceMiddleware(logger *slog.Logger) gin.HandlerFunc
    PerformanceMiddleware is an alias for Performance

func RateLimit(requests int, window time.Duration) gin.HandlerFunc
    RateLimit provides basic rate limiting

func RecoveryWithErrorHandler(logger *slog.Logger) gin.HandlerFunc
    RecoveryWithErrorHandler provides recovery with error handling

func RequestID() gin.HandlerFunc
    RequestID adds a unique request ID to each request

func RequestLoggingMiddleware(loggingService interface{}) gin.HandlerFunc
    RequestLoggingMiddleware provides enhanced request logging

func RequestTracker(logger *slog.Logger, environment string) gin.HandlerFunc
    RequestTracker adds request tracking

func Security() gin.HandlerFunc
    Security adds basic security headers

func SecurityLoggingMiddleware(loggingService interface{}) gin.HandlerFunc
    SecurityLoggingMiddleware provides security event logging

func SmartCacheHeaders() gin.HandlerFunc
    SmartCacheHeaders adds basic cache headers

func Timeout(timeout time.Duration) gin.HandlerFunc
    Timeout middleware with configurable duration

