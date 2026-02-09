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
	slug := fmt.Sprintf("link-%d", now.Unix())
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

	if err := os.WriteFile(filePath, []byte(fileContent), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created link post: %s\n", filePath)
	fmt.Printf("URL: %s\n", linkURL)
}
