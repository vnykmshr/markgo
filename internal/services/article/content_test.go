package article

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestProcessor() *MarkdownContentProcessor {
	return NewMarkdownContentProcessor(slog.Default())
}

func TestProcessMarkdown_Basic(t *testing.T) {
	p := newTestProcessor()
	html, err := p.ProcessMarkdown("Hello **world**")
	require.NoError(t, err)
	assert.Contains(t, html, "<strong>world</strong>")
}

func TestProcessMarkdown_GFMExtensions(t *testing.T) {
	p := newTestProcessor()

	t.Run("strikethrough", func(t *testing.T) {
		html, err := p.ProcessMarkdown("~~deleted~~")
		require.NoError(t, err)
		assert.Contains(t, html, "<del>deleted</del>")
	})

	t.Run("task list", func(t *testing.T) {
		html, err := p.ProcessMarkdown("- [x] done\n- [ ] todo")
		require.NoError(t, err)
		assert.Contains(t, html, "checked")
	})

	t.Run("table", func(t *testing.T) {
		md := "| A | B |\n|---|---|\n| 1 | 2 |"
		html, err := p.ProcessMarkdown(md)
		require.NoError(t, err)
		assert.Contains(t, html, "<table>")
	})
}

func TestProcessMarkdown_CodeBlocks(t *testing.T) {
	p := newTestProcessor()
	md := "```go\nfunc main() {}\n```"
	html, err := p.ProcessMarkdown(md)
	require.NoError(t, err)
	assert.Contains(t, html, "<code")
	assert.Contains(t, html, "func main()")
}

func TestProcessMarkdown_EmptyInput(t *testing.T) {
	p := newTestProcessor()
	html, err := p.ProcessMarkdown("")
	require.NoError(t, err)
	assert.Empty(t, html)
}

func TestGenerateExcerpt_Empty(t *testing.T) {
	p := newTestProcessor()
	assert.Equal(t, "", p.GenerateExcerpt("", 100))
}

func TestGenerateExcerpt_Short(t *testing.T) {
	p := newTestProcessor()
	// Short content returned as-is (after formatting strip)
	result := p.GenerateExcerpt("Hello world", 100)
	assert.Equal(t, "Hello world", result)
}

func TestGenerateExcerpt_TruncationAtWordBoundary(t *testing.T) {
	p := newTestProcessor()
	content := "The quick brown fox jumps over the lazy dog and runs far away"
	result := p.GenerateExcerpt(content, 30)
	assert.True(t, strings.HasSuffix(result, "..."), "should end with ellipsis")
	assert.LessOrEqual(t, len(result), 34) // 30 + "..."
	// Should not cut mid-word
	assert.False(t, strings.Contains(result[:len(result)-3], " o"), "should not cut 'over' mid-word")
}

func TestGenerateExcerpt_StripsFormatting(t *testing.T) {
	p := newTestProcessor()

	t.Run("bold and italic", func(t *testing.T) {
		result := p.GenerateExcerpt("Hello **bold** and *italic* text", 100)
		assert.NotContains(t, result, "**")
		assert.NotContains(t, result, "*")
	})

	t.Run("code blocks removed", func(t *testing.T) {
		result := p.GenerateExcerpt("Before ```code block``` after", 100)
		assert.NotContains(t, result, "```")
	})

	t.Run("links keep text", func(t *testing.T) {
		result := p.GenerateExcerpt("Check [this link](https://example.com) out", 100)
		assert.Contains(t, result, "this link")
		assert.NotContains(t, result, "https://example.com")
	})

	t.Run("headings removed", func(t *testing.T) {
		result := p.GenerateExcerpt("# Title\n\nBody text here", 100)
		assert.NotContains(t, result, "#")
		assert.Contains(t, result, "Body text here")
	})
}

func TestProcessDuplicateTitles_MatchingH1Removed(t *testing.T) {
	p := newTestProcessor()
	html := `<h1 id="my-title">My Title</h1><p>Content</p>`
	result := p.ProcessDuplicateTitles("My Title", html)
	assert.NotContains(t, result, "<h1")
	assert.Contains(t, result, "<p>Content</p>")
}

func TestProcessDuplicateTitles_NonMatchingPreserved(t *testing.T) {
	p := newTestProcessor()
	html := `<h1>Different Title</h1><p>Content</p>`
	result := p.ProcessDuplicateTitles("My Title", html)
	assert.Contains(t, result, "<h1>Different Title</h1>")
}

func TestProcessDuplicateTitles_EmptyInputs(t *testing.T) {
	p := newTestProcessor()
	assert.Equal(t, "", p.ProcessDuplicateTitles("Title", ""))
	assert.Equal(t, "<p>hi</p>", p.ProcessDuplicateTitles("", "<p>hi</p>"))
}

func TestCalculateReadingTime_Boundary(t *testing.T) {
	p := newTestProcessor()

	t.Run("empty content", func(t *testing.T) {
		assert.Equal(t, 0, p.CalculateReadingTime(""))
	})

	t.Run("one word", func(t *testing.T) {
		assert.Equal(t, 1, p.CalculateReadingTime("hello"))
	})

	t.Run("200 words", func(t *testing.T) {
		words := strings.Repeat("word ", 200)
		assert.Equal(t, 1, p.CalculateReadingTime(words))
	})

	t.Run("400 words", func(t *testing.T) {
		words := strings.Repeat("word ", 400)
		assert.Equal(t, 2, p.CalculateReadingTime(words))
	})

	t.Run("strips markdown formatting from count", func(t *testing.T) {
		// Code blocks should be stripped
		content := "```\nlong code block with many many many words\n```\nShort."
		time := p.CalculateReadingTime(content)
		assert.Equal(t, 1, time)
	})
}

func TestExtractImageURLs(t *testing.T) {
	p := newTestProcessor()

	t.Run("multiple images", func(t *testing.T) {
		content := "![alt1](img1.png) text ![alt2](img2.jpg)"
		urls := p.ExtractImageURLs(content)
		assert.Len(t, urls, 2)
		assert.Contains(t, urls, "img1.png")
		assert.Contains(t, urls, "img2.jpg")
	})

	t.Run("no images", func(t *testing.T) {
		urls := p.ExtractImageURLs("just text, no images")
		assert.Empty(t, urls)
	})
}

func TestExtractLinks(t *testing.T) {
	p := newTestProcessor()

	t.Run("multiple links", func(t *testing.T) {
		content := "[link1](https://a.com) and [link2](https://b.com)"
		links := p.ExtractLinks(content)
		assert.Len(t, links, 2)
		assert.Contains(t, links, "https://a.com")
		assert.Contains(t, links, "https://b.com")
	})

	t.Run("no links", func(t *testing.T) {
		links := p.ExtractLinks("no links here")
		assert.Empty(t, links)
	})
}

func TestValidateContent(t *testing.T) {
	p := newTestProcessor()

	t.Run("empty content", func(t *testing.T) {
		issues := p.ValidateContent("")
		assert.Contains(t, issues, "Content is empty")
	})

	t.Run("missing alt text", func(t *testing.T) {
		issues := p.ValidateContent("![](image.png)")
		found := false
		for _, issue := range issues {
			if strings.Contains(issue, "alt text") {
				found = true
			}
		}
		assert.True(t, found, "should report missing alt text")
	})

	t.Run("long lines", func(t *testing.T) {
		longLine := strings.Repeat("x", 121)
		issues := p.ValidateContent(longLine)
		found := false
		for _, issue := range issues {
			if strings.Contains(issue, "very long") {
				found = true
			}
		}
		assert.True(t, found, "should report long lines")
	})

	t.Run("valid content", func(t *testing.T) {
		issues := p.ValidateContent("# Title\n\nA short paragraph.")
		assert.Empty(t, issues)
	})
}
