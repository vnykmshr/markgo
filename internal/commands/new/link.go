package new

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func runLink(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: URL is required")
		fmt.Fprintln(os.Stderr, "Usage: markgo new link <url> [commentary]")
		os.Exit(1)
	}

	linkURL := args[0]

	// Validate URL
	parsed, err := url.Parse(linkURL)
	if err != nil || parsed.Host == "" {
		fmt.Fprintf(os.Stderr, "Error: invalid URL: %s\n", linkURL)
		os.Exit(1)
	}

	// Optional commentary from remaining args
	var commentary string
	if len(args) > 1 {
		commentary = strings.Join(args[1:], " ")
	}

	now := time.Now()
	slug := fmt.Sprintf("link-%d", now.UnixMilli())
	title := parsed.Host // Use domain as title

	// Build frontmatter
	fm := map[string]any{
		"slug":     slug,
		"title":    title,
		"link_url": linkURL,
		"date":     now.Format(time.RFC3339),
		"draft":    false,
		"author":   getDefaultAuthor(),
	}

	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal frontmatter: %v\n", err)
		os.Exit(1)
	}

	// Content is the commentary, or a default referencing the link
	content := commentary
	if content == "" {
		content = fmt.Sprintf("[%s](%s)", title, linkURL)
	}

	fileContent := fmt.Sprintf("---\n%s---\n\n%s\n", string(yamlBytes), strings.TrimSpace(content))

	dateStr := now.Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", dateStr, slug)
	filePath := filepath.Join(articlesDir, filename)

	// Verify articles directory exists
	if info, statErr := os.Stat(articlesDir); statErr != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: articles directory does not exist: %s\n", articlesDir)
		os.Exit(1)
	}

	// Use O_EXCL to prevent overwriting existing files
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644) //nolint:gosec // article files should be world-readable
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create file (may already exist): %v\n", err)
		os.Exit(1)
	}
	_, writeErr := f.WriteString(fileContent)
	closeErr := f.Close()
	if writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write file: %v\n", writeErr)
		os.Exit(1)
	}
	if closeErr != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to close file: %v\n", closeErr)
		os.Exit(1)
	}

	fmt.Printf("Created link post: %s\n", filePath)
	fmt.Printf("URL: %s\n", linkURL)
}
