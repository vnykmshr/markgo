package services

import (
	"fmt"
	"html"
	"html/template"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yourusername/markgo/internal/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

var (
	timeZoneCache = sync.Map{}
	titleCaser    = cases.Title(language.English)
)

type TemplateService struct {
	templates *template.Template
	config    *config.Config
}

func NewTemplateService(templatesPath string, cfg *config.Config) (*TemplateService, error) {
	service := &TemplateService{
		config: cfg,
	}

	if err := service.loadTemplates(templatesPath); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
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
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	t.templates = tmpl
	return nil
}

func (t *TemplateService) Render(w io.Writer, templateName string, data any) error {
	if t.templates == nil {
		return fmt.Errorf("templates not loaded")
	}

	tmpl := t.templates.Lookup(templateName)
	if tmpl == nil {
		return fmt.Errorf("template %s not found", templateName)
	}

	return tmpl.Execute(w, data)
}

func (t *TemplateService) RenderToString(templateName string, data any) (string, error) {
	var buf strings.Builder
	if err := t.Render(&buf, templateName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
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
	return t.loadTemplates(templatesPath)
}

// GetTemplate returns the internal template for Gin integration
func (t *TemplateService) GetTemplate() *template.Template {
	return t.templates
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
		val := reflect.ValueOf(items)

		// Check if it's a slice or array
		if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
			return items // Return original if not slice/array
		}

		length := val.Len()

		// Boundary checks
		if start < 0 {
			start = 0
		}
		if start > length {
			start = length
		}
		if end < start {
			end = start
		}
		if end > length {
			end = length
		}

		// Create new slice of the same type
		result := reflect.MakeSlice(val.Type(), end-start, end-start)
		reflect.Copy(result, val.Slice(start, end))

		return result.Interface()
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
		if loc, found := timeZoneCache.Load(zone); found {
			return date.In(loc.(*time.Location)).Format(format)
		}

		loc, err := time.LoadLocation(zone)
		if err != nil {
			return date.Format(format) // Fallback to original timezone
		}

		timeZoneCache.Store(zone, loc)
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
		runes := []rune(s)
		if len(runes) <= length {
			return s
		}
		return string(runes[:length]) + "..."
	},
	"truncateHTML": func(s string, length int) template.HTML {
		if utf8.RuneCountInString(s) <= length {
			return template.HTML(s)
		}
		// Properly handle UTF-8 boundaries
		runes := []rune(s)
		safe := template.HTMLEscapeString(string(runes[:length])) + "..."
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
			// Remove diacritics
			case unicode.Is(unicode.Mn, r): // Mn: Nonspacing marks
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
