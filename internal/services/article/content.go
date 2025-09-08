package article

import (
	"bytes"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/vnykmshr/markgo/internal/models"
)

// ContentProcessor handles all content-related processing operations
type ContentProcessor interface {
	// Core processing
	ProcessMarkdown(content string) (string, error)
	GenerateExcerpt(content string, maxLength int) string
	ProcessDuplicateTitles(title, htmlContent string) string

	// Content analysis
	CalculateReadingTime(content string) int
	ExtractImageURLs(content string) []string
	ExtractLinks(content string) []string

	// Content validation
	ValidateContent(content string) []string
}

// MarkdownContentProcessor implements ContentProcessor using Goldmark
type MarkdownContentProcessor struct {
	markdown goldmark.Markdown
	logger   *slog.Logger

	// Pre-compiled regexes for performance
	codeBlockRe      *regexp.Regexp
	linkRe           *regexp.Regexp
	imageRe          *regexp.Regexp
	duplicateTitleRe *regexp.Regexp
}

// NewMarkdownContentProcessor creates a new markdown content processor
func NewMarkdownContentProcessor(logger *slog.Logger) *MarkdownContentProcessor {
	// Configure Goldmark with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,            // GitHub Flavored Markdown
			extension.Table,          // Tables
			extension.Strikethrough,  // Strikethrough
			extension.Linkify,        // Auto-link URLs
			extension.TaskList,       // Task lists
			extension.DefinitionList, // Definition lists
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // Hard line breaks
			html.WithXHTML(),     // XHTML output
			html.WithUnsafe(),    // Allow raw HTML
		),
	)

	return &MarkdownContentProcessor{
		markdown: md,
		logger:   logger,

		// Pre-compile regexes for performance
		codeBlockRe:      regexp.MustCompile("```[\\s\\S]*?```"),
		linkRe:           regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`),
		imageRe:          regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`),
		duplicateTitleRe: regexp.MustCompile(`<h1[^>]*>.*?</h1>`),
	}
}

// ProcessMarkdown converts markdown content to HTML
func (p *MarkdownContentProcessor) ProcessMarkdown(content string) (string, error) {
	var buf bytes.Buffer

	if err := p.markdown.Convert([]byte(content), &buf); err != nil {
		p.logger.Error("Failed to process markdown", "error", err)
		return "", fmt.Errorf("markdown processing failed: %w", err)
	}

	return buf.String(), nil
}

// GenerateExcerpt creates a text excerpt from markdown content
func (p *MarkdownContentProcessor) GenerateExcerpt(content string, maxLength int) string {
	if content == "" {
		return ""
	}

	// Remove code blocks first (they can be long and not useful in excerpts)
	cleanContent := p.codeBlockRe.ReplaceAllString(content, " ")

	// Remove links but keep the link text
	cleanContent = p.linkRe.ReplaceAllString(cleanContent, "$1")

	// Remove images but keep alt text
	cleanContent = p.imageRe.ReplaceAllString(cleanContent, "$1")

	// Remove markdown formatting
	cleanContent = strings.ReplaceAll(cleanContent, "**", "")
	cleanContent = strings.ReplaceAll(cleanContent, "*", "")
	cleanContent = strings.ReplaceAll(cleanContent, "_", "")
	cleanContent = strings.ReplaceAll(cleanContent, "`", "")

	// Remove heading markers
	lines := strings.Split(cleanContent, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			cleanLines = append(cleanLines, line)
		}
	}
	cleanContent = strings.Join(cleanLines, " ")

	// Normalize whitespace
	cleanContent = strings.Join(strings.Fields(cleanContent), " ")

	// Truncate to max length
	if len(cleanContent) <= maxLength {
		return cleanContent
	}

	// Find the last complete word within the limit
	truncated := cleanContent[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// ProcessDuplicateTitles removes duplicate H1 tags if the title is already rendered
func (p *MarkdownContentProcessor) ProcessDuplicateTitles(title, htmlContent string) string {
	if title == "" || htmlContent == "" {
		return htmlContent
	}

	// Check if the HTML content starts with an H1 that matches the title
	h1Matches := p.duplicateTitleRe.FindAllString(htmlContent, -1)
	if len(h1Matches) > 0 {
		firstH1 := h1Matches[0]
		// Extract text from the H1 tag
		h1Text := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`).ReplaceAllString(firstH1, "$1")
		h1Text = strings.TrimSpace(strings.ReplaceAll(h1Text, "\n", " "))

		// If the H1 text closely matches the title, remove it
		if strings.EqualFold(h1Text, title) ||
			strings.EqualFold(strings.ReplaceAll(h1Text, " ", ""), strings.ReplaceAll(title, " ", "")) {
			return p.duplicateTitleRe.ReplaceAllString(htmlContent, "")
		}
	}

	return htmlContent
}

// CalculateReadingTime estimates reading time based on word count
func (p *MarkdownContentProcessor) CalculateReadingTime(content string) int {
	const averageWordsPerMinute = 200

	// Remove markdown formatting for accurate word count
	cleanContent := p.codeBlockRe.ReplaceAllString(content, " ")
	cleanContent = p.linkRe.ReplaceAllString(cleanContent, "$1")
	cleanContent = p.imageRe.ReplaceAllString(cleanContent, "")

	words := strings.Fields(cleanContent)
	wordCount := len(words)

	readingTime := wordCount / averageWordsPerMinute
	if readingTime == 0 && wordCount > 0 {
		readingTime = 1 // Minimum 1 minute
	}

	return readingTime
}

// ExtractImageURLs finds all image URLs in the content
func (p *MarkdownContentProcessor) ExtractImageURLs(content string) []string {
	matches := p.imageRe.FindAllStringSubmatch(content, -1)
	var urls []string

	for _, match := range matches {
		if len(match) > 0 {
			// Extract URL from markdown image syntax: ![alt](url)
			urlMatch := regexp.MustCompile(`\(([^)]+)\)`).FindStringSubmatch(match[0])
			if len(urlMatch) > 1 {
				urls = append(urls, urlMatch[1])
			}
		}
	}

	return urls
}

// ExtractLinks finds all links in the content
func (p *MarkdownContentProcessor) ExtractLinks(content string) []string {
	matches := p.linkRe.FindAllStringSubmatch(content, -1)
	var links []string

	for _, match := range matches {
		if len(match) > 0 {
			// Extract URL from markdown link syntax: [text](url)
			urlMatch := regexp.MustCompile(`\(([^)]+)\)`).FindStringSubmatch(match[0])
			if len(urlMatch) > 1 {
				links = append(links, urlMatch[1])
			}
		}
	}

	return links
}

// ValidateContent performs basic content validation
func (p *MarkdownContentProcessor) ValidateContent(content string) []string {
	var issues []string

	if content == "" {
		issues = append(issues, "Content is empty")
		return issues
	}

	// Check for common markdown issues
	lines := strings.Split(content, "\n")

	// Check for missing alt text in images
	imageMatches := p.imageRe.FindAllString(content, -1)
	for _, img := range imageMatches {
		if strings.Contains(img, "![]") {
			issues = append(issues, "Image found without alt text")
		}
	}

	// Check for very long lines (readability)
	for i, line := range lines {
		if len(line) > 120 && !strings.HasPrefix(line, "```") {
			issues = append(issues, fmt.Sprintf("Line %d is very long (%d characters)", i+1, len(line)))
		}
	}

	// Check for too many consecutive headers
	headerCount := 0
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			headerCount++
			if headerCount > 3 {
				issues = append(issues, "Too many consecutive headers detected")
				break
			}
		} else if strings.TrimSpace(line) != "" {
			headerCount = 0
		}
	}

	return issues
}

// Ensure MarkdownContentProcessor implements the interfaces
var _ ContentProcessor = (*MarkdownContentProcessor)(nil)
var _ models.ArticleProcessor = (*MarkdownContentProcessor)(nil)
