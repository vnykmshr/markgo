package services

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/constants"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
)

// SimpleTemplateService provides clean, maintainable template functionality
type SimpleTemplateService struct {
	templates *template.Template
	config    *config.Config
}

// NewSimpleTemplateService creates a new simplified template service
func NewSimpleTemplateService(templatesPath string, cfg *config.Config) (*SimpleTemplateService, error) {
	service := &SimpleTemplateService{
		config: cfg,
	}

	if err := service.loadTemplates(templatesPath); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *SimpleTemplateService) loadTemplates(templatesPath string) error {
	funcMap := getSimpleTemplateFuncMap()

	pattern := filepath.Join(templatesPath, "*.html")
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return apperrors.NewHTTPError(500, "Failed to parse HTML templates", apperrors.ErrTemplateParseError)
	}

	s.templates = tmpl
	return nil
}

// Render renders a template to the provided writer.
func (s *SimpleTemplateService) Render(w io.Writer, templateName string, data any) error {
	if s.templates == nil {
		return apperrors.ErrTemplateNotFound
	}

	tmpl := s.templates.Lookup(templateName)
	if tmpl == nil {
		return apperrors.NewHTTPError(404,
			fmt.Sprintf("Template '%s' not found", templateName),
			apperrors.ErrTemplateNotFound)
	}

	return tmpl.Execute(w, data)
}

// RenderToString renders a template and returns the result as a string.
func (s *SimpleTemplateService) RenderToString(templateName string, data any) (string, error) {
	var buf bytes.Buffer
	if err := s.Render(&buf, templateName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// HasTemplate checks if a template with the given name exists.
func (s *SimpleTemplateService) HasTemplate(templateName string) bool {
	if s.templates == nil {
		return false
	}
	return s.templates.Lookup(templateName) != nil
}

// GetTemplate returns the underlying template instance.
func (s *SimpleTemplateService) GetTemplate() *template.Template {
	return s.templates
}

// Reload reloads all templates from the specified path.
func (s *SimpleTemplateService) Reload(templatesPath string) error {
	return s.loadTemplates(templatesPath)
}

// getSimpleTemplateFuncMap returns essential template functions without over-engineering
func getSimpleTemplateFuncMap() template.FuncMap {
	funcMap := make(template.FuncMap)

	// Add function groups
	addDateTimeFunctions(funcMap)
	addStringFunctions(funcMap)
	addNumericFunctions(funcMap)
	addArrayFunctions(funcMap)
	addUtilityFunctions(funcMap)

	return funcMap
}

// addDateTimeFunctions adds date and time related template functions
func addDateTimeFunctions(funcMap template.FuncMap) {
	funcMap["formatDate"] = func(t time.Time) string {
		return t.Format(constants.DefaultDateFormat)
	}
	funcMap["formatTime"] = func(t time.Time) string {
		return t.Format(constants.DefaultTimeFormat)
	}
	funcMap["formatDateTime"] = func(t time.Time) string {
		return t.Format(constants.DefaultDateTimeFormat)
	}
	funcMap["humanizeTime"] = humanize.Time
}

// addStringFunctions adds string manipulation template functions
func addStringFunctions(funcMap template.FuncMap) {
	titleCaser := cases.Title(language.English)

	funcMap["title"] = titleCaser.String
	funcMap["upper"] = strings.ToUpper
	funcMap["lower"] = strings.ToLower
	funcMap["trim"] = strings.TrimSpace
	funcMap["join"] = func(sep string, items []string) string {
		return strings.Join(items, sep)
	}
	funcMap["split"] = func(sep, str string) []string {
		return strings.Split(str, sep)
	}
	funcMap["contains"] = strings.Contains
	funcMap["hasPrefix"] = strings.HasPrefix
	funcMap["hasSuffix"] = strings.HasSuffix
}

// addNumericFunctions adds numeric calculation template functions
func addNumericFunctions(funcMap template.FuncMap) {
	funcMap["add"] = func(a, b int) int { return a + b }
	funcMap["sub"] = func(a, b int) int { return a - b }
	funcMap["mul"] = func(a, b int) int { return a * b }
	funcMap["div"] = func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	}
}

// addArrayFunctions adds array/slice manipulation template functions
func addArrayFunctions(funcMap template.FuncMap) {
	funcMap["len"] = func(v any) int {
		if v == nil {
			return 0
		}
		switch val := v.(type) {
		case []string:
			return len(val)
		case []any:
			return len(val)
		case string:
			return len(val)
		default:
			return 0
		}
	}
	funcMap["slice"] = func(start, end int, items []any) []any {
		if start < 0 || end < 0 || start >= len(items) {
			return []any{}
		}
		if end > len(items) {
			end = len(items)
		}
		if start >= end {
			return []any{}
		}
		return items[start:end]
	}
}

// addUtilityFunctions adds utility and content-related template functions
func addUtilityFunctions(funcMap template.FuncMap) {
	funcMap["default"] = func(defaultVal, val any) any {
		if val == nil || val == "" || val == 0 {
			return defaultVal
		}
		return val
	}
	funcMap["pluralize"] = func(count int, singular, plural string) string {
		if count == 1 {
			return singular
		}
		return plural
	}
	funcMap["truncate"] = func(length int, text string) string {
		if len(text) <= length {
			return text
		}
		if length <= 3 {
			return text[:length]
		}
		return text[:length-3] + "..."
	}
	funcMap["readingTime"] = func(content string) int {
		wordCount := len(strings.Fields(content))
		minutes := wordCount / constants.DefaultReadingSpeed
		if minutes < 1 {
			return 1
		}
		return minutes
	}
	funcMap["excerpt"] = func(content string, length int) string {
		if length <= 0 {
			length = constants.DefaultExcerptLength
		}
		if len(content) <= length {
			return content
		}
		return content[:length] + "..."
	}
}
