package errors // import "github.com/vnykmshr/markgo/internal/errors"


VARIABLES

var (
	// Article errors
	ErrArticleNotFound    = errors.New("article not found")
	ErrArticleParseError  = errors.New("article parse error")
	ErrInvalidFrontMatter = errors.New("invalid front matter")
	ErrArticleExists      = errors.New("article already exists")

	// Email errors
	ErrEmailNotConfigured = errors.New("email credentials not configured")
	ErrSMTPAuthFailed     = errors.New("SMTP authentication failed")
	ErrEmailSendFailed    = errors.New("email send failed")
	ErrInvalidEmail       = errors.New("invalid email address")

	// Template errors
	ErrTemplateNotFound   = errors.New("template not found")
	ErrTemplateParseError = errors.New("template parse error")
	ErrTemplateRender     = errors.New("template render error")

	// Configuration errors
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrMissingConfig    = errors.New("missing required configuration")
	ErrConfigValidation = errors.New("configuration validation failed")

	// Cache errors
	ErrCacheNotFound = errors.New("cache entry not found")
	ErrCacheInvalid  = errors.New("invalid cache data")

	// Search errors
	ErrSearchFailed  = errors.New("search operation failed")
	ErrInvalidQuery  = errors.New("invalid search query")
	ErrSearchTimeout = errors.New("search operation timed out")

	// File system errors
	ErrFileNotFound    = errors.New("file not found")
	ErrFilePermission  = errors.New("file permission denied")
	ErrDirectoryCreate = errors.New("failed to create directory")

	// Validation errors
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidInput     = errors.New("invalid input")
	ErrMissingField     = errors.New("missing required field")

	// CLI errors
	ErrCLIExecution   = errors.New("CLI execution failed")
	ErrCLIValidation  = errors.New("CLI validation failed")
	ErrCLIInterrupted = errors.New("CLI operation interrupted")

	// Service errors
	ErrServiceFailure     = errors.New("service operation failed")
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
	ErrServiceTimeout     = errors.New("service operation timed out")
)
    Domain-specific errors for MarkGo application


FUNCTIONS

func GetUserFriendlyMessage(err error) string
    GetUserFriendlyMessage returns a user-friendly error message

func GracefulExit(exitCode int, message string, cleanup func())
    GracefulExit provides a graceful way to exit CLI applications with cleanup

func HandleCLIError(err error, cleanup func())
    HandleCLIError handles CLI errors gracefully with optional cleanup

func IsArticleNotFound(err error) bool
    IsArticleNotFound checks if an error indicates an article was not found

func IsCLIError(err error) bool
    IsCLIError checks if an error is CLI-related

func IsConfigurationError(err error) bool
    IsConfigurationError checks if an error is configuration-related

func IsNotFound(err error) bool
    IsNotFound checks if an error indicates a "not found" condition

func IsServiceError(err error) bool
    IsServiceError checks if an error is service-related

func IsValidationError(err error) bool
    IsValidationError checks if an error is a validation-related error


TYPES

type ArticleError struct {
	File    string
	Message string
	Err     error
}
    ArticleError represents an error related to article processing

func NewArticleError(file, message string, err error) *ArticleError
    NewArticleError creates a new ArticleError

func (e *ArticleError) Error() string

func (e *ArticleError) Unwrap() error

type CLIError struct {
	Operation string
	Message   string
	Err       error
	ExitCode  int
}
    CLIError represents a CLI-specific error with operation context

func NewCLIError(operation, message string, err error, exitCode int) *CLIError
    NewCLIError creates a new CLIError

func (e *CLIError) Error() string

func (e *CLIError) Unwrap() error

type ConfigError struct {
	Field   string
	Value   interface{}
	Message string
	Err     error
}
    ConfigError represents a configuration-related error

func NewConfigError(field string, value interface{}, message string, err error) *ConfigError
    NewConfigError creates a new ConfigError

func (e *ConfigError) Error() string

func (e *ConfigError) Unwrap() error

type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}
    HTTPError represents an HTTP-specific error with status code

func NewHTTPError(statusCode int, message string, err error) *HTTPError
    NewHTTPError creates a new HTTPError

func (e *HTTPError) Error() string

func (e *HTTPError) Unwrap() error

type ServiceError struct {
	Service   string
	Operation string
	Message   string
	Err       error
	Retryable bool
}
    ServiceError represents a service-layer error with service context

func NewServiceError(service, operation, message string, err error, retryable bool) *ServiceError
    NewServiceError creates a new ServiceError

func (e *ServiceError) Error() string

func (e *ServiceError) IsRetryable() bool
    IsRetryable checks if a service error can be retried

func (e *ServiceError) Unwrap() error

type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
	Err     error
}
    ValidationError represents a validation error with field-specific details

func NewValidationError(field string, value interface{}, message string, err error) *ValidationError
    NewValidationError creates a new ValidationError

func (e *ValidationError) Error() string

func (e *ValidationError) Unwrap() error

