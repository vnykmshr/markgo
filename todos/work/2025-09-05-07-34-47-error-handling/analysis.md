# MarkGo Error Handling Analysis Report

Based on my examination of the MarkGo codebase, here's a comprehensive analysis of the current error handling patterns and recommendations for improvement:

## Current Error Handling Patterns

### 1. **CLI Tools (`cmd/` directory)**

**Current State:**
- `/Users/vmx/workspace/gocode/src/github.com/vnykmshr/markgo/cmd/server/main.go`: Uses `slog.Error()` with `os.Exit(1)` for critical failures
- `/Users/vmx/workspace/gocode/src/github.com/vnykmshr/markgo/cmd/new-article/main.go`: Uses `fmt.Printf()` for error messages with `os.Exit(1)`

**Issues Identified:**
- **Inconsistent error logging**: Server uses structured logging (`slog.Error`), but new-article uses plain `fmt.Printf`
- **Non-informative error messages**: Lines 60, 66, 73 have generic error messages that don't help users understand what went wrong

**Improvements Needed:**

```go
// Current (cmd/new-article/main.go:66-67)
if err := os.MkdirAll(articlesDir, 0755); err != nil {
    fmt.Printf("Error creating articles directory: %v\n", err)
    os.Exit(1)
}

// Improved
if err := os.MkdirAll(articlesDir, 0755); err != nil {
    fmt.Printf("Failed to create articles directory '%s': %v\n", 
        articlesDir, err)
    fmt.Printf("Make sure you have write permissions to the current directory.\n")
    os.Exit(1)
}
```

### 2. **HTTP Handlers (`internal/handlers/handlers.go`)**

**Current State:**
- Generally good error handling with proper HTTP status codes
- Uses structured logging with `slog.Error` and `logger.Error`
- Returns JSON responses for API errors

**Issues Identified:**

**Line 148-151**: Article not found handling
```go
article, err := h.articleService.GetArticleBySlug(slug)
if err != nil {
    h.NotFound(c)
    return
}
```
**Problem**: The actual error is lost - users and admins don't know if it was "article not found" vs "file system error"

**Line 350-357**: Form validation errors are generic
```go
if err := c.ShouldBindJSON(&msg); err != nil {
    data := utils.GetTemplateData()
    data["error"] = "Invalid form data"
    data["message"] = err.Error()
    // ...
}
```
**Problem**: Exposes internal validation errors directly to users

**Improvements Needed:**

```go
// Better article error handling
article, err := h.articleService.GetArticleBySlug(slug)
if err != nil {
    h.logger.Error("Failed to retrieve article", "slug", slug, "error", err)
    if errors.Is(err, services.ErrArticleNotFound) {
        h.NotFound(c)
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Unable to load article",
            "message": "Please try again later",
        })
    }
    return
}

// Better form validation
if err := c.ShouldBindJSON(&msg); err != nil {
    h.logger.Warn("Contact form validation failed", "error", err, "ip", c.ClientIP())
    data := utils.GetTemplateData()
    data["error"] = "Invalid form submission"
    data["message"] = "Please check your form data and try again"
    // Don't expose internal validation details
}
```

### 3. **Services (`internal/services/`)**

#### Article Service (`article.go`)
**Current State:**
- Good use of wrapped errors with `fmt.Errorf("failed to load articles: %w", err)`
- Uses structured logging for warnings but continues processing

**Issues Identified:**

**Lines 98-101**: Silent failure handling
```go
article, err := s.ParseArticleFile(path)
if err != nil {
    s.logger.Warn("Failed to parse article", "file", path, "error", err)
    return nil // Continue processing other files
}
```
**Problem**: Parsing failures are logged but not aggregated or reported to administrators

**Line 419**: Generic error message
```go
return nil, fmt.Errorf("article not found: %s", slug)
```
**Problem**: Should use a custom error type for better error handling

#### Email Service (`email.go`)
**Current State:**
- Good error wrapping and detailed logging
- Includes context in error messages

**Issues Identified:**

**Lines 58-60**: Configuration check should be more informative
```go
if e.config.Username == "" || e.config.Password == "" {
    e.logger.Warn("Email service not configured, skipping email send")
    return fmt.Errorf("email service not configured")
}
```
**Problem**: Error message doesn't specify which configuration is missing

**Improvements Needed:**

```go
// Custom error types
var (
    ErrEmailNotConfigured = errors.New("email service not configured")
    ErrMissingCredentials = errors.New("email credentials not provided") 
    ErrSMTPConnection     = errors.New("SMTP connection failed")
)

// Better configuration validation
func (e *EmailService) validateConfig() error {
    var missing []string
    if e.config.Username == "" {
        missing = append(missing, "username")
    }
    if e.config.Password == "" {
        missing = append(missing, "password") 
    }
    if e.config.Host == "" {
        missing = append(missing, "host")
    }
    
    if len(missing) > 0 {
        return fmt.Errorf("%w: missing %s", ErrMissingCredentials, 
            strings.Join(missing, ", "))
    }
    return nil
}
```

### 4. **Configuration Loading (`internal/config/config.go`)**

**Issues Identified:**
- **Line 96**: `_ = godotenv.Load()` - Silently ignores .env loading errors
- **Lines 185-189**: `getEnvInt` fails silently and returns default values without logging

**Improvements Needed:**

```go
// Better env loading
func Load() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        // Don't fail but log the issue for debugging
        slog.Warn("No .env file found or failed to load", "error", err)
    }
    
    // Configuration validation
    cfg := &Config{...}
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return cfg, nil
}

func (c *Config) Validate() error {
    var issues []string
    
    if c.ArticlesPath == "" {
        issues = append(issues, "articles path is required")
    }
    
    if _, err := os.Stat(c.ArticlesPath); os.IsNotExist(err) {
        issues = append(issues, fmt.Sprintf("articles directory does not exist: %s", c.ArticlesPath))
    }
    
    if len(issues) > 0 {
        return errors.New(strings.Join(issues, "; "))
    }
    
    return nil
}
```

### 5. **Middleware (`internal/middleware/middleware.go`)**

**Current State:**
- Good structured logging with context
- Proper error responses with appropriate status codes

**Issues Identified:**

**Lines 156-159**: Rate limit error messages could be more helpful
```go
data["error"] = "Rate limit exceeded"
data["message"] = fmt.Sprintf("Too many requests. Limit: %d requests per %v", limit, window)
```
**Problem**: Doesn't tell users when they can try again

**Improvements Needed:**

```go
// More helpful rate limiting errors
if !limiter.IsAllowed(key) {
    retryAfter := limiter.GetRetryAfter(key) // Need to implement this
    data := utils.GetTemplateData()
    data["error"] = "Rate limit exceeded"
    data["message"] = fmt.Sprintf("Too many requests. Please wait %v before trying again.", retryAfter)
    data["retry_after"] = retryAfter.Seconds()
    c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
    c.JSON(http.StatusTooManyRequests, data)
    utils.PutTemplateData(data)
    c.Abort()
    return
}
```

## Major Areas for Improvement

### 1. **Custom Error Types**

The codebase would benefit from custom error types for better error handling:

```go
// internal/errors/errors.go
package errors

import "errors"

var (
    // Article errors
    ErrArticleNotFound     = errors.New("article not found")
    ErrArticleParseError   = errors.New("article parse error")
    ErrInvalidFrontMatter = errors.New("invalid front matter")
    
    // Email errors  
    ErrEmailNotConfigured = errors.New("email service not configured")
    ErrSMTPAuthFailed     = errors.New("SMTP authentication failed")
    ErrDuplicateEmail     = errors.New("duplicate email detected")
    
    // Template errors
    ErrTemplateNotFound   = errors.New("template not found")
    ErrTemplateParseError = errors.New("template parse error")
    
    // Configuration errors
    ErrInvalidConfig      = errors.New("invalid configuration")
    ErrMissingConfig      = errors.New("missing required configuration")
)

type ArticleError struct {
    File    string
    Message string
    Err     error
}

func (e *ArticleError) Error() string {
    return fmt.Sprintf("article error in %s: %s: %v", e.File, e.Message, e.Err)
}

func (e *ArticleError) Unwrap() error {
    return e.Err
}
```

### 2. **Centralized Error Handling**

Create middleware for consistent error response formatting:

```go
// internal/middleware/error.go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            var statusCode int
            var userMessage string
            
            switch {
            case errors.Is(err, errors.ErrArticleNotFound):
                statusCode = http.StatusNotFound
                userMessage = "The requested article was not found"
            case errors.Is(err, errors.ErrEmailNotConfigured):
                statusCode = http.StatusServiceUnavailable  
                userMessage = "Email service is currently unavailable"
            default:
                statusCode = http.StatusInternalServerError
                userMessage = "An unexpected error occurred"
            }
            
            c.JSON(statusCode, gin.H{
                "error": userMessage,
                "request_id": c.GetString("RequestID"),
            })
        }
    }
}
```

### 3. **Better Error Messages**

**Current issues:**
- Generic messages that don't help users understand the problem
- Inconsistent error formatting between different parts of the application
- Missing context about what users can do to resolve issues

**Recommended improvements:**
- Include specific context in error messages
- Provide actionable guidance where possible
- Use consistent error response format
- Sanitize errors before showing to users (avoid exposing internal details)

### 4. **Error Logging and Monitoring**

**Missing features:**
- Error aggregation and reporting
- No distinction between expected vs unexpected errors  
- Limited context in error logs

**Recommendations:**
- Add error classification (user error vs system error vs external error)
- Include request context in all error logs
- Implement error metrics collection
- Add structured error reporting for administrators

### 5. **Recovery and Resilience**

**Current gaps:**
- No circuit breaker pattern for external services (email)
- Limited retry logic
- No graceful degradation for non-critical features

**Suggested improvements:**
- Add retry logic with exponential backoff for transient failures
- Implement circuit breaker for email service
- Allow the application to continue functioning when non-critical services fail

This analysis shows that while MarkGo has good basic error handling, there are significant opportunities to improve user experience, debugging capabilities, and system resilience through more structured error handling approaches.