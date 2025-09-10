package services

import (
	"bytes"
	"context"
	"crypto/sha256"
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
	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/utils"
	"github.com/vnykmshr/obcache-go/pkg/obcache"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
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
		cache, _ = obcache.New(basicConfig)
	}

	timeZoneCache = cache
}

// CachedTemplateFunctions holds obcache-wrapped template operations
type CachedTemplateFunctions struct {
	RenderToString func(string, any) (string, error)
	ParseTemplate  func(string, string) (*template.Template, error)
}

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
	_ = goflowScheduler.Start() // Continue even if scheduler fails to start

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

func (t *TemplateService) Render(w io.Writer, templateName string, data any) error {
	if t.templates == nil {
		return apperrors.ErrTemplateNotFound
	}

	tmpl := t.templates.Lookup(templateName)
	if tmpl == nil {
		return apperrors.NewHTTPError(404, fmt.Sprintf("Template '%s' not found", templateName), apperrors.ErrTemplateNotFound)
	}

	return tmpl.Execute(w, data)
}

func (t *TemplateService) RenderToString(templateName string, data any) (string, error) {
	// Use cached function if available
	if t.cachedFunctions.RenderToString != nil {
		return t.cachedFunctions.RenderToString(templateName, data)
	}

	// Fallback to uncached version
	return t.renderToStringUncached(templateName, data)
}

func (t *TemplateService) HasTemplate(templateName string) bool {
	if t.templates == nil {
		return false
	}
	return t.templates.Lookup(templateName) != nil
}

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

func (t *TemplateService) Reload(templatesPath string) error {
	// Clear cache before reloading
	if t.obcache != nil {
		_ = t.obcache.Clear()
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
	cleanupTask := workerpool.TaskFunc(func(ctx context.Context) error {
		if t.obcache != nil {
			t.obcache.Cleanup()
		}
		return nil
	})

	// Schedule cleanup using proper cron format (6 fields: second, minute, hour, day, month, weekday)
	_ = t.scheduler.ScheduleCron("template-cache-cleanup", "0 0 * * * *", cleanupTask) // Continue without scheduled cleanup if scheduling fails
}

// renderToStringUncached renders template without caching
func (t *TemplateService) renderToStringUncached(templateName string, data any) (string, error) {
	return utils.GetGlobalBufferPool().RenderToString(func(buf *bytes.Buffer) error {
		return t.Render(buf, templateName, data)
	})
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
		t.obcache.Close()
	}
}

// GetFuncMap returns the template function map for reuse in other services
func GetTemplateFuncMap() template.FuncMap {
	return templateFuncs
}

var templateFuncs = template.FuncMap{
	"safeHTML": func(s string) template.HTML {
		s = html.UnescapeString(s)
		// Basic sanitization
		policy := bluemonday.UGCPolicy()
		return template.HTML(policy.Sanitize(s))
	},
	"join": func(sep string, items []string) string {
		return strings.Join(items, sep)
	},
	"slice": func(items any, start, end int) any {
		// Use pooled slice copying for better memory efficiency
		return utils.GetGlobalSlicePool().SliceCopyPooled(items, start, end)
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
	"printf": func(format string, args ...any) string {
		return fmt.Sprintf(format, args...)
	},
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
		// Use pooled int sequence generation
		return utils.GetGlobalSlicePool().IntSequencePooled(start, end)
	},
	"gt": func(a, b int) bool {
		return a > b
	},
	"lt": func(a, b int) bool {
		return a < b
	},
	"eq": func(a, b any) bool {
		return reflect.DeepEqual(a, b)
	},
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
			return date.In(cachedLoc.(*time.Location)).Format(format)
		}

		loc, err := time.LoadLocation(zone)
		if err != nil {
			return date.Format(format) // Fallback to original timezone
		}

		// Store timezone location in cache with no expiration
		_ = timeZoneCache.Set(zone, loc, 0)
		return date.In(loc).Format(format)
	},
	"now": func() time.Time {
		return time.Now()
	},
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
	"contains": func(s, substr string) bool {
		return strings.Contains(s, substr)
	},
	"hasPrefix": func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	},
	"hasSuffix": func(s, suffix string) bool {
		return strings.HasSuffix(s, suffix)
	},
	"lower": func(s string) string {
		return strings.ToLower(s)
	},
	"upper": func(s string) string {
		return strings.ToUpper(s)
	},
	"title": func(s string) string {
		return titleCaser.String(s)
	},
	"trim": func(s string) string {
		return strings.TrimSpace(s)
	},
	"truncate": func(s string, length int) string {
		if utf8.RuneCountInString(s) <= length {
			return s
		}
		// Use pooled rune slice for truncation
		truncatedRunes := utils.GetGlobalRuneSlicePool().WithRuneSlice(func(pooledRunes *[]rune) []rune {
			*pooledRunes = append(*pooledRunes, []rune(s)...)
			if len(*pooledRunes) <= length {
				result := make([]rune, len(*pooledRunes))
				copy(result, *pooledRunes)
				return result
			}
			result := make([]rune, length)
			copy(result, (*pooledRunes)[:length])
			return result
		})
		return string(truncatedRunes) + "..."
	},
	"truncateHTML": func(s string, length int) template.HTML {
		if utf8.RuneCountInString(s) <= length {
			return template.HTML(s)
		}
		// Use pooled rune slice for HTML truncation
		truncatedRunes := utils.GetGlobalRuneSlicePool().WithRuneSlice(func(pooledRunes *[]rune) []rune {
			*pooledRunes = append(*pooledRunes, []rune(s)...)
			result := make([]rune, length)
			copy(result, (*pooledRunes)[:length])
			return result
		})
		safe := template.HTMLEscapeString(string(truncatedRunes)) + "..."
		return template.HTML(safe)
	},
	"slugify": func(s string) string {
		s = strings.ToLower(s)
		// Use pooled string builder for slugification
		return utils.GetGlobalBufferPool().BuildString(func(b *strings.Builder) {
			for _, r := range norm.NFD.String(s) {
				switch {
				case r >= 'a' && r <= 'z':
					b.WriteRune(r)
				case r >= '0' && r <= '9':
					b.WriteRune(r)
				case r == ' ' || r == '-':
					b.WriteRune('-')
				// Remove diacritics
				case unicode.Is(unicode.Mn, r): // Mn: Nonspacing marks
					continue
				}
			}
		})
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
			if len(word) > 0 {
				initials += strings.ToUpper(string([]rune(word)[0]))
			}
		}
		return initials
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
