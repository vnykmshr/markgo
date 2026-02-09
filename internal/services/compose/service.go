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

	// Generate slug
	var slug string
	if input.Title != "" {
		slug = generateSlug(input.Title)
	} else {
		slug = fmt.Sprintf("thought-%d", now.Unix())
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

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil { //nolint:gosec // article files should be world-readable
		return "", fmt.Errorf("failed to write article file: %w", err)
	}

	return slug, nil
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
