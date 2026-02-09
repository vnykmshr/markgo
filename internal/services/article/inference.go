package article

import "github.com/vnykmshr/markgo/internal/models"

// inferPostType determines the post type based on article content.
// Explicit type in frontmatter always wins. Otherwise:
//   - Has link_url → "link"
//   - No title and < 100 words → "thought"
//   - Everything else → "article"
func inferPostType(article *models.Article) string {
	if article.Type != "" {
		return article.Type
	}
	if article.LinkURL != "" {
		return "link"
	}
	if article.Title == "" && article.WordCount < 100 {
		return "thought"
	}
	return "article"
}
