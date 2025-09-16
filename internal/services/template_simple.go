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
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/constants"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func (s *SimpleTemplateService) Render(w io.Writer, templateName string, data any) error {
	if s.templates == nil {
		return apperrors.ErrTemplateNotFound
	}

	tmpl := s.templates.Lookup(templateName)
	if tmpl == nil {
		return apperrors.NewHTTPError(404, fmt.Sprintf("Template '%s' not found", templateName), apperrors.ErrTemplateNotFound)
	}

	return tmpl.Execute(w, data)
}

func (s *SimpleTemplateService) RenderToString(templateName string, data any) (string, error) {
	var buf bytes.Buffer
	if err := s.Render(&buf, templateName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *SimpleTemplateService) HasTemplate(templateName string) bool {
	if s.templates == nil {
		return false
	}
	return s.templates.Lookup(templateName) != nil
}

func (s *SimpleTemplateService) GetTemplate() *template.Template {
	return s.templates
}

func (s *SimpleTemplateService) Reload(templatesPath string) error {
	return s.loadTemplates(templatesPath)
}

// getSimpleTemplateFuncMap returns essential template functions without over-engineering
func getSimpleTemplateFuncMap() template.FuncMap {
	titleCaser := cases.Title(language.English)

	return template.FuncMap{
		// Date and time functions
		"formatDate": func(t time.Time) string {
			return t.Format(constants.DefaultDateFormat)
		},
		"formatTime": func(t time.Time) string {
			return t.Format(constants.DefaultTimeFormat)
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format(constants.DefaultDateTimeFormat)
		},
		"humanizeTime": func(t time.Time) string {
			return humanize.Time(t)
		},

		// String manipulation
		"title": titleCaser.String,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"trim":  strings.TrimSpace,
		"join": func(sep string, items []string) string {
			return strings.Join(items, sep)
		},
		"split": func(sep, str string) []string {
			return strings.Split(str, sep)
		},
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,

		// Numeric functions
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},

		// Array/slice functions
		"len": func(v any) int {
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
		},
		"slice": func(start, end int, items []any) []any {
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
		},

		// Utility functions
		"default": func(defaultVal, val any) any {
			if val == nil || val == "" || val == 0 {
				return defaultVal
			}
			return val
		},
		"pluralize": func(count int, singular, plural string) string {
			if count == 1 {
				return singular
			}
			return plural
		},
		"truncate": func(length int, text string) string {
			if len(text) <= length {
				return text
			}
			if length <= 3 {
				return text[:length]
			}
			return text[:length-3] + "..."
		},

		// Reading time estimation
		"readingTime": func(content string) int {
			wordCount := len(strings.Fields(content))
			minutes := wordCount / constants.DefaultReadingSpeed
			if minutes < 1 {
				return 1
			}
			return minutes
		},

		// SEO and meta functions
		"excerpt": func(content string, length int) string {
			if length <= 0 {
				length = constants.DefaultExcerptLength
			}
			if len(content) <= length {
				return content
			}
			return content[:length] + "..."
		},
	}
}
