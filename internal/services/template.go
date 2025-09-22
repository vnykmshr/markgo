package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/microcosm-cc/bluemonday"
	"github.com/vnykmshr/goflow/pkg/scheduling/scheduler"
	"github.com/vnykmshr/goflow/pkg/scheduling/workerpool"
	"github.com/vnykmshr/obcache-go/pkg/obcache"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"

	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
)

var (
	timeZoneCache *obcache.Cache
	titleCaser    = cases.Title(language.English)
)

func init() {
	// Initialize global timezone cache
	config := obcache.NewDefaultConfig()
	config.MaxEntries = 200 // Support many timezones
	config.DefaultTTL = 0   // Timezones don't expire

	cache, err := obcache.New(config)
	if err != nil {
		// Fallback to basic config if creation fails
		basicConfig := obcache.NewDefaultConfig()
		basicConfig.DefaultTTL = 0
		if fallbackCache, err := obcache.New(basicConfig); err == nil {
			timeZoneCache = fallbackCache
		} else {
			// Failed to create cache, timeZoneCache remains nil
			timeZoneCache = nil
		}
		return
	}
	timeZoneCache = cache
}

// CachedTemplateFunctions holds obcache-wrapped template operations
type CachedTemplateFunctions struct {
	RenderToString func(string, any) (string, error)
	ParseTemplate  func(string, string) (*template.Template, error)
}

// TemplateService provides template rendering functionality.
type TemplateService struct {
	templates *template.Template
	config    *config.Config

	// obcache integration
	obcache         *obcache.Cache
	cachedFunctions CachedTemplateFunctions

	// goflow integration
	scheduler scheduler.Scheduler
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewTemplateService creates a new TemplateService instance.
func NewTemplateService(templatesPath string, cfg *config.Config) (*TemplateService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize obcache for template operations
	cacheConfig := obcache.NewDefaultConfig()
	cacheConfig.MaxEntries = 500              // Templates are smaller, fewer entries needed
	cacheConfig.DefaultTTL = 30 * time.Minute // Templates change less frequently

	obcacheInstance, err := obcache.New(cacheConfig)
	if err != nil {
		cancel()
		// Continue without cache if it fails
	}

	// Initialize goflow scheduler for template maintenance
	goflowScheduler := scheduler.New()
	if err := goflowScheduler.Start(); err != nil {
		// Log warning but continue - scheduler is optional
		_ = err // Explicitly ignore error for linting
	}

	service := &TemplateService{
		config:    cfg,
		obcache:   obcacheInstance,
		scheduler: goflowScheduler,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize cached functions
	service.initializeCachedFunctions()

	// Setup background maintenance
	service.setupTemplateMaintenance()

	if err := service.loadTemplates(templatesPath); err != nil {
		return nil, err
	}

	return service, nil
}

func (t *TemplateService) loadTemplates(templatesPath string) error {
	// Use the shared template function map
	funcMap := GetTemplateFuncMap()

	// Load all template files
	pattern := filepath.Join(templatesPath, "*.html")
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return apperrors.NewHTTPError(500, "Failed to parse HTML templates", apperrors.ErrTemplateParseError)
	}

	t.templates = tmpl
	return nil
}

// Render renders a template to the provided writer.
func (t *TemplateService) Render(w io.Writer, templateName string, data any) error {
	if t.templates == nil {
		return apperrors.ErrTemplateNotFound
	}

	tmpl := t.templates.Lookup(templateName)
	if tmpl == nil {
		return apperrors.NewHTTPError(404,
			fmt.Sprintf("Template '%s' not found", templateName),
			apperrors.ErrTemplateNotFound)
	}

	return tmpl.Execute(w, data)
}

// RenderToString renders a template and returns the result as a string.
func (t *TemplateService) RenderToString(templateName string, data any) (string, error) {
	// Use cached function if available
	if t.cachedFunctions.RenderToString != nil {
		return t.cachedFunctions.RenderToString(templateName, data)
	}

	// Fallback to uncached version
	return t.renderToStringUncached(templateName, data)
}

// HasTemplate checks if a template with the given name exists.
func (t *TemplateService) HasTemplate(templateName string) bool {
	if t.templates == nil {
		return false
	}
	return t.templates.Lookup(templateName) != nil
}

// ListTemplates returns a list of all available template names.
func (t *TemplateService) ListTemplates() []string {
	if t.templates == nil {
		return []string{}
	}

	var names []string
	for _, tmpl := range t.templates.Templates() {
		if tmpl.Name() != "" {
			names = append(names, tmpl.Name())
		}
	}
	return names
}

// Reload reloads all templates from the specified path.
func (t *TemplateService) Reload(templatesPath string) error {
	// Clear cache before reloading
	if t.obcache != nil {
		if err := t.obcache.Clear(); err != nil {
			// Log warning but continue - cache clear is optional
			_ = err // Explicitly ignore error for linting
		}
	}
	return t.loadTemplates(templatesPath)
}

// GetTemplate returns the internal template for Gin integration
func (t *TemplateService) GetTemplate() *template.Template {
	return t.templates
}

// initializeCachedFunctions initializes obcache-wrapped template functions
func (t *TemplateService) initializeCachedFunctions() {
	if t.obcache == nil {
		return
	}

	// Wrap template rendering with obcache
	t.cachedFunctions.RenderToString = obcache.Wrap(
		t.obcache,
		t.renderToStringUncached,
		obcache.WithKeyFunc(func(args []any) string {
			if len(args) >= 2 {
				if templateName, ok := args[0].(string); ok {
					// Create hash of data for cache key
					hash := sha256.Sum256([]byte(fmt.Sprintf("%v", args[1])))
					dataHash := fmt.Sprintf("%x", hash[:8])
					return fmt.Sprintf("render:%s:%s", templateName, dataHash)
				}
			}
			return "render:default"
		}),
		obcache.WithTTL(15*time.Minute),
	)
}

// setupTemplateMaintenance sets up background maintenance using goflow
func (t *TemplateService) setupTemplateMaintenance() {
	if t.scheduler == nil {
		return
	}

	// Template cache cleanup every hour
	cleanupTask := workerpool.TaskFunc(func(_ context.Context) error {
		if t.obcache != nil {
			t.obcache.Cleanup()
		}
		return nil
	})

	// Schedule cleanup using proper cron format (6 fields: second, minute, hour, day, month, weekday)
	if err := t.scheduler.ScheduleCron("template-cache-cleanup", "0 0 * * * *", cleanupTask); err != nil {
		// Log warning but continue without scheduled cleanup
		_ = err // Explicitly ignore error for linting
	}
}

// renderToStringUncached renders template without caching
func (t *TemplateService) renderToStringUncached(templateName string, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Render(&buf, templateName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GetCacheStats returns template cache statistics
func (t *TemplateService) GetCacheStats() map[string]int {
	if t.obcache == nil {
		return map[string]int{}
	}

	stats := t.obcache.Stats()
	return map[string]int{
		"templates_cached": int(stats.KeyCount()),
		"cache_hits":       int(stats.Hits()),
		"cache_misses":     int(stats.Misses()),
		"hit_ratio":        int(stats.HitRate() * 100),
	}
}

// GetTimezoneCacheStats returns timezone cache statistics
func GetTimezoneCacheStats() map[string]any {
	if timeZoneCache == nil {
		return map[string]any{
			"error": "timezone cache not initialized",
		}
	}

	stats := timeZoneCache.Stats()
	return map[string]any{
		"timezones_cached": int(stats.KeyCount()),
		"cache_hits":       int(stats.Hits()),
		"cache_misses":     int(stats.Misses()),
		"hit_ratio":        stats.HitRate() * 100,
		"evictions":        int(stats.Evictions()),
		"cache_type":       "obcache-go",
	}
}

// Shutdown gracefully shuts down the template service
func (t *TemplateService) Shutdown() {
	if t.cancel != nil {
		t.cancel()
	}

	if t.scheduler != nil {
		t.scheduler.Stop()
	}

	if t.obcache != nil {
		if err := t.obcache.Close(); err != nil {
			// Log warning but continue - close error is not critical
			_ = err // Explicitly ignore error for linting
		}
	}
}

// GetTemplateFuncMap returns the template function map for reuse in other services
func GetTemplateFuncMap() template.FuncMap {
	return templateFuncs
}

var templateFuncs = template.FuncMap{
	"safeHTML": func(s string) template.HTML {
		s = html.UnescapeString(s)
		// Basic sanitization using bluemonday UGC policy
		policy := bluemonday.UGCPolicy()
		// #nosec G203 - This is intentional HTML output for templates, sanitized by bluemonday
		return template.HTML(policy.Sanitize(s))
	},
	"join": func(sep string, items []string) string {
		return strings.Join(items, sep)
	},
	"slice": func(items any, start, end int) any {
		v := reflect.ValueOf(items)
		if v.Kind() != reflect.Slice {
			return items
		}
		length := v.Len()
		if start < 0 {
			start = 0
		}
		if end > length {
			end = length
		}
		if start >= end {
			return reflect.MakeSlice(v.Type(), 0, 0).Interface()
		}
		return v.Slice(start, end).Interface()
	},
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
	"mul": func(a, b int) int {
		return a * b
	},
	"printf": fmt.Sprintf,
	"le": func(a, b any) bool {
		switch va := a.(type) {
		case int:
			if vb, ok := b.(int); ok {
				return va <= vb
			}
		case float64:
			if vb, ok := b.(float64); ok {
				return va <= vb
			}
		}
		return false
	},
	"div": func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"seq": func(start, end int) []int {
		if start > end {
			return []int{}
		}
		seq := make([]int, end-start+1)
		for i := range seq {
			seq[i] = start + i
		}
		return seq
	},
	"gt": func(a, b int) bool {
		return a > b
	},
	"lt": func(a, b int) bool {
		return a < b
	},
	"eq": reflect.DeepEqual,
	"len": func(items any) int {
		v := reflect.ValueOf(items)
		switch v.Kind() {
		case reflect.Slice, reflect.Array, reflect.String, reflect.Map:
			return v.Len()
		default:
			return 0
		}
	},
	"compare": func(a, b any) int {
		af, aok := toFloat(a)
		bf, bok := toFloat(b)
		if !aok || !bok {
			return 0
		}
		switch {
		case af > bf:
			return 1
		case af < bf:
			return -1
		default:
			return 0
		}
	},
	"formatNumber": func(num any) string {
		switch n := num.(type) {
		case int:
			return humanize.Comma(int64(n))
		case float64:
			return humanize.Ftoa(n)
		default:
			return fmt.Sprint(n)
		}
	},
	"formatDate": func(date time.Time, format string) string {
		return date.Format(format)
	},
	"formatDateInZone": func(date time.Time, zone, format string) string {
		if cachedLoc, found := timeZoneCache.Get(zone); found {
			if loc, ok := cachedLoc.(*time.Location); ok {
				return date.In(loc).Format(format)
			}
			// Fallback if type assertion fails
			if loc, err := time.LoadLocation(zone); err == nil {
				return date.In(loc).Format(format)
			}
			return date.Format(format)
		}

		loc, err := time.LoadLocation(zone)
		if err != nil {
			return date.Format(format) // Fallback to original timezone
		}

		// Store timezone location in cache with no expiration
		if err := timeZoneCache.Set(zone, loc, 0); err != nil {
			// Log warning but continue - cache set is optional
			_ = err // Explicitly ignore error for linting
		}
		return date.In(loc).Format(format)
	},
	"now": time.Now,
	"max": func(a, b int) int {
		if a > b {
			return a
		}
		return b
	},
	"min": func(a, b int) int {
		if a < b {
			return a
		}
		return b
	},
	"mod": func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a % b
	},
	"ne": func(a, b any) bool {
		return !reflect.DeepEqual(a, b)
	},
	"and": func(a, b bool) bool {
		return a && b
	},
	"or": func(a, b bool) bool {
		return a || b
	},
	"not": func(a bool) bool {
		return !a
	},
	"contains":  strings.Contains,
	"hasPrefix": strings.HasPrefix,
	"hasSuffix": strings.HasSuffix,
	"lower":     strings.ToLower,
	"upper":     strings.ToUpper,
	"title":     titleCaser.String,
	"trim":      strings.TrimSpace,
	"truncate": func(s string, length int) string {
		runes := []rune(s)
		if len(runes) <= length {
			return s
		}
		return string(runes[:length]) + "..."
	},
	"truncateHTML": func(s string, length int) template.HTML {
		runes := []rune(s)
		if len(runes) <= length {
			// #nosec G203 - This is intentional HTML output for templates
			return template.HTML(s)
		}
		safe := template.HTMLEscapeString(string(runes[:length])) + "..."
		// #nosec G203 - This is intentional HTML output for templates, content is escaped
		return template.HTML(safe)
	},
	"slugify": func(s string) string {
		s = strings.ToLower(s)
		var b strings.Builder
		for _, r := range norm.NFD.String(s) {
			switch {
			case r >= 'a' && r <= 'z':
				b.WriteRune(r)
			case r >= '0' && r <= '9':
				b.WriteRune(r)
			case r == ' ' || r == '-':
				b.WriteRune('-')
			case unicode.Is(unicode.Mn, r): // Remove diacritics
				continue
			}
		}
		return b.String()
	},
	"timeAgo": func(date time.Time) string {
		duration := time.Since(date)
		switch {
		case duration < time.Minute:
			return "just now"
		case duration < time.Hour:
			return pluralize(int(duration.Minutes()), "minute", "minutes")
		case duration < 24*time.Hour:
			return pluralize(int(duration.Hours()), "hour", "hours")
		case duration < 30*24*time.Hour:
			return pluralize(int(duration.Hours()/24), "day", "days")
		case duration < 365*24*time.Hour:
			return pluralize(int(duration.Hours()/(24*30)), "month", "months")
		default:
			return pluralize(int(duration.Hours()/(24*365)), "year", "years")
		}
	},
	"isNil": func(value any) bool {
		return value == nil
	},
	"isNotNil": func(value any) bool {
		return value != nil
	},
	"default": func(defaultValue, value any) any {
		if value == nil {
			return defaultValue
		}

		val := reflect.ValueOf(value)
		switch val.Kind() {
		case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
			if val.Len() == 0 {
				return defaultValue
			}
		case reflect.Ptr, reflect.Interface:
			if val.IsNil() {
				return defaultValue
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if val.Int() == 0 {
				return defaultValue
			}
		case reflect.Float32, reflect.Float64:
			if val.Float() == 0 {
				return defaultValue
			}
		case reflect.Bool:
			if !val.Bool() {
				return defaultValue
			}
		}

		return value
	},
	"get": func(m map[string]any, key string) any {
		return m[key]
	},
	"ternary": func(cond bool, trueVal, falseVal any) any {
		if cond {
			return trueVal
		}
		return falseVal
	},
	"initials": func(s string) string {
		words := strings.Fields(strings.TrimSpace(s))
		if len(words) == 0 {
			return ""
		}

		initials := ""
		for i, word := range words {
			if i >= 2 { // Only take first 2 initials
				break
			}
			if word != "" {
				if r, _ := utf8.DecodeRuneInString(word); r != utf8.RuneError {
					initials += strings.ToUpper(string(r))
				}
			}
		}
		return initials
	},

	// SEO Template Functions
	"generateJsonLD": func(data map[string]interface{}) template.HTML {
		// Convert map to JSON for structured data
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return template.HTML("")
		}
		// #nosec G203 - This is intentional JSON-LD structured data for SEO
		return template.HTML(fmt.Sprintf(`<script type="application/ld+json">%s</script>`, string(jsonBytes)))
	},
	"renderMetaTags": func(tags map[string]string) template.HTML {
		var buf strings.Builder
		for name, content := range tags {
			if content == "" {
				continue
			}
			// Escape HTML entities in meta content
			escapedContent := html.EscapeString(content)

			// Determine meta tag type
			if strings.HasPrefix(name, "og:") || strings.HasPrefix(name, "twitter:") || strings.HasPrefix(name, "article:") {
				buf.WriteString(fmt.Sprintf(`<meta property="%s" content="%s">`, html.EscapeString(name), escapedContent))
			} else if name == "canonical" {
				buf.WriteString(fmt.Sprintf(`<link rel="canonical" href="%s">`, escapedContent))
			} else {
				buf.WriteString(fmt.Sprintf(`<meta name="%s" content="%s">`, html.EscapeString(name), escapedContent))
			}
			buf.WriteString("\n")
		}
		// #nosec G203 - This is intentional HTML output for meta tags, content is escaped
		return template.HTML(buf.String())
	},
	"seoExcerpt": func(content string, maxLength int) string {
		// Remove markdown formatting for SEO description
		text := content
		text = strings.ReplaceAll(text, "#", "")
		text = strings.ReplaceAll(text, "*", "")
		text = strings.ReplaceAll(text, "_", "")
		text = strings.ReplaceAll(text, "`", "")
		text = strings.ReplaceAll(text, "[", "")
		text = strings.ReplaceAll(text, "]", "")
		text = strings.ReplaceAll(text, "(", "")
		text = strings.ReplaceAll(text, ")", "")

		// Clean up whitespace
		words := strings.Fields(text)
		if len(words) == 0 {
			return ""
		}

		// Build excerpt
		var excerpt strings.Builder
		for _, word := range words {
			if excerpt.Len()+len(word)+1 > maxLength {
				break
			}
			if excerpt.Len() > 0 {
				excerpt.WriteString(" ")
			}
			excerpt.WriteString(word)
		}

		result := excerpt.String()
		if len(result) < len(strings.Join(words, " ")) {
			result += "..."
		}

		return result
	},
	"readingTime": func(content string, wordsPerMinute int) int {
		if wordsPerMinute <= 0 {
			wordsPerMinute = 200 // Default reading speed
		}
		words := strings.Fields(content)
		minutes := len(words) / wordsPerMinute
		if minutes == 0 && len(words) > 0 {
			return 1 // Minimum 1 minute
		}
		return minutes
	},
	"buildURL": func(baseURL, path string) string {
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			return path // Already absolute URL
		}
		baseURL = strings.TrimRight(baseURL, "/")
		path = strings.TrimLeft(path, "/")
		return fmt.Sprintf("%s/%s", baseURL, path)
	},
}

// Helper function
func toFloat(v any) (float64, bool) {
	switch v := v.(type) {
	case int:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s ago", singular)
	}
	return fmt.Sprintf("%d %s ago", n, plural)
}
