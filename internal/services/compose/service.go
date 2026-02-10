package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Service handles creating new posts from compose form input.
type Service struct {
	articlesPath  string
	defaultAuthor string
}

// NewService creates a new compose service.
func NewService(articlesPath, defaultAuthor string) *Service {
	return &Service{
		articlesPath:  articlesPath,
		defaultAuthor: defaultAuthor,
	}
}

// Input represents the compose form submission.
type Input struct {
	Content string
	Title   string
	LinkURL string
	Tags    string
	Draft   bool
}

// CreatePost creates a new markdown post file from compose input.
// Returns the generated slug.
func (s *Service) CreatePost(input Input) (string, error) {
	now := time.Now()

	// Generate slug (fall back to timestamp if title produces empty slug, e.g. non-ASCII titles)
	var slug string
	if input.Title != "" {
		slug = generateSlug(input.Title)
	}
	if slug == "" {
		slug = fmt.Sprintf("thought-%d", now.UnixMilli())
	}

	// Parse comma-separated tags
	var tags []string
	for _, t := range strings.Split(input.Tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	// Build frontmatter map (only non-empty fields)
	fm := map[string]any{
		"slug": slug,
		"date": now.Format(time.RFC3339),
	}
	if input.Title != "" {
		fm["title"] = input.Title
	}
	if input.LinkURL != "" {
		fm["link_url"] = input.LinkURL
	}
	if len(tags) > 0 {
		fm["tags"] = tags
	}
	if s.defaultAuthor != "" {
		fm["author"] = s.defaultAuthor
	}
	fm["draft"] = input.Draft

	// Marshal frontmatter to YAML
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Build file content
	content := fmt.Sprintf("---\n%s---\n\n%s\n", string(yamlBytes), strings.TrimSpace(input.Content))

	// Build filename with date prefix
	dateStr := now.Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", dateStr, slug)
	filePath := filepath.Join(s.articlesPath, filename)

	// Write file (O_EXCL prevents overwriting an existing post)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644) //nolint:gosec // article files should be world-readable
	if err != nil {
		return "", fmt.Errorf("failed to create article file (may already exist): %w", err)
	}
	_, writeErr := f.WriteString(content)
	closeErr := f.Close()
	if writeErr != nil {
		return "", fmt.Errorf("failed to write article file: %w", writeErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("failed to close article file: %w", closeErr)
	}

	return slug, nil
}

// LoadArticle reads an existing article file by slug and returns its content for editing.
func (s *Service) LoadArticle(slug string) (*Input, error) {
	_, content, err := s.findFileBySlug(slug)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid article format: missing frontmatter")
	}

	var fm map[string]any
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	input := &Input{
		Content: strings.TrimSpace(parts[2]),
	}
	if title, ok := fm["title"].(string); ok {
		input.Title = title
	}
	if linkURL, ok := fm["link_url"].(string); ok {
		input.LinkURL = linkURL
	}
	if tags, ok := fm["tags"].([]any); ok {
		var tagStrs []string
		for _, t := range tags {
			if str, ok := t.(string); ok {
				tagStrs = append(tagStrs, str)
			}
		}
		input.Tags = strings.Join(tagStrs, ", ")
	}
	if draft, ok := fm["draft"].(bool); ok {
		input.Draft = draft
	}

	return input, nil
}

// UpdateArticle overwrites an existing article file with updated content.
// Preserves fields not exposed in the compose form (slug, date, author, categories).
func (s *Service) UpdateArticle(slug string, input Input) error {
	filePath, existingContent, err := s.findFileBySlug(slug)
	if err != nil {
		return err
	}

	parts := strings.SplitN(string(existingContent), "---", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid article format: missing frontmatter")
	}

	var fm map[string]any
	if unmarshalErr := yaml.Unmarshal([]byte(parts[1]), &fm); unmarshalErr != nil {
		return fmt.Errorf("failed to parse frontmatter: %w", unmarshalErr)
	}

	// Update editable fields, preserve everything else (slug, date, author, categories)
	if input.Title != "" {
		fm["title"] = input.Title
	} else {
		delete(fm, "title")
	}
	if input.LinkURL != "" {
		fm["link_url"] = input.LinkURL
	} else {
		delete(fm, "link_url")
	}

	var tags []string
	for _, t := range strings.Split(input.Tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	if len(tags) > 0 {
		fm["tags"] = tags
	} else {
		delete(fm, "tags")
	}

	fm["draft"] = input.Draft

	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	fileContent := fmt.Sprintf("---\n%s---\n\n%s\n", string(yamlBytes), strings.TrimSpace(input.Content))

	// Atomic write: temp file + rename prevents data loss on crash
	tmpFile, err := os.CreateTemp(filepath.Dir(filePath), ".markgo-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.WriteString(fileContent); err != nil {
		tmpFile.Close()    //nolint:gosec // best-effort cleanup
		os.Remove(tmpPath) //nolint:gosec // best-effort cleanup
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath) //nolint:gosec // best-effort cleanup
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath) //nolint:gosec // best-effort cleanup
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// findFileBySlug scans the articles directory for a file with matching slug in frontmatter.
// Returns the file path and raw content to avoid a redundant re-read by callers.
func (s *Service) findFileBySlug(slug string) (string, []byte, error) {
	entries, err := os.ReadDir(s.articlesPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read articles directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(s.articlesPath, entry.Name())
		content, err := os.ReadFile(filePath) // #nosec G304 -- filePath is built from articlesPath + directory entry
		if err != nil {
			continue // Skip unreadable files — effectively non-existent for this scan
		}

		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) < 3 {
			continue
		}

		var fm struct {
			Slug string `yaml:"slug"`
		}
		if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
			continue
		}

		if fm.Slug == slug {
			return filePath, content, nil
		}
	}

	return "", nil, fmt.Errorf("article not found: %s", slug)
}

// generateSlug creates a URL-friendly slug from a title.
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	return slug
}
