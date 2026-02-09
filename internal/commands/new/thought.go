package new

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func runThought(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: thought text is required")
		fmt.Fprintln(os.Stderr, "Usage: markgo new thought \"Your thought here\"")
		os.Exit(1)
	}

	content := strings.Join(args, " ")
	now := time.Now()
	slug := fmt.Sprintf("thought-%d", now.Unix())

	// Build frontmatter
	fm := map[string]any{
		"slug":   slug,
		"date":   now.Format(time.RFC3339),
		"draft":  false,
		"author": getDefaultAuthor(),
	}

	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal frontmatter: %v\n", err)
		os.Exit(1)
	}

	fileContent := fmt.Sprintf("---\n%s---\n\n%s\n", string(yamlBytes), strings.TrimSpace(content))

	dateStr := now.Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", dateStr, slug)
	filePath := filepath.Join(articlesDir, filename)

	if err := os.WriteFile(filePath, []byte(fileContent), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created thought: %s\n", filePath)
}
