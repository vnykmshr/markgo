// Package new provides the article creation command for MarkGo.
package new

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	apperrors "github.com/vnykmshr/markgo/internal/errors"
)

const (
	defaultTitle       = "Untitled Article"
	defaultDescription = ""
	defaultTags        = "general"
	boolTrue           = "true"
	boolFalse          = "false"
	defaultCategory    = "uncategorized"
	defaultDraft       = true
	defaultFeatured    = false
	articlesDir        = "articles"
)

// Run executes the new article command
func Run(args []string) {
	// Create a new flag set for this command
	fs := flag.NewFlagSet("new", flag.ExitOnError)

	title := fs.String("title", defaultTitle, "Article title")
	description := fs.String("description", defaultDescription, "Article description")
	tags := fs.String("tags", defaultTags, "Comma-separated tags")
	category := fs.String("category", defaultCategory, "Article category")
	author := fs.String("author", "", "Author name (default: current OS username)")
	draft := fs.Bool("draft", defaultDraft, "Mark article as draft")
	featured := fs.Bool("featured", defaultFeatured, "Mark article as featured")
	template := fs.String("template", "default", "Article template to use")
	preview := fs.Bool("preview", false, "Preview the article without creating file")
	list := fs.Bool("list", false, "List available templates")
	datePrefix := fs.Bool("date-prefix", false, "Add date prefix to filename")
	interactive := fs.Bool("interactive", false, "Force interactive mode")
	help := fs.Bool("help", false, "Show help message")

	// Cleanup function for any resources
	cleanup := func() {
		// Add any necessary cleanup here (file handles, temp files, etc.)
	}

	if err := fs.Parse(args[1:]); err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("flag parsing", "Failed to parse command flags", err, 1),
			cleanup,
		)
	}

	if *help {
		showHelp()
		return
	}

	if *list {
		listTemplates()
		return
	}

	// Check if we should run interactive mode
	if *interactive || shouldRunInteractive(fs) {
		runInteractiveMode(title, description, tags, category, author, template, draft, featured, datePrefix)
	}

	// Set default author if not provided
	if *author == "" {
		*author = getDefaultAuthor()
	}

	// Sanitize all inputs
	*title = SanitizeInput(*title)
	*description = SanitizeInput(*description)
	*tags = SanitizeInput(*tags)
	*category = SanitizeInput(*category)
	*author = SanitizeInput(*author)

	// Validate all inputs
	validation := ValidateArticleInput(*title, *description, *tags, *category, *author, *template)
	if !validation.Valid {
		ShowValidationErrors(validation.Errors)
		apperrors.HandleCLIError(
			apperrors.NewCLIError("input validation", "Article input validation failed", apperrors.ErrCLIValidation, 1),
			cleanup,
		)
		return
	}

	// Generate filename from title
	slug := slugify(*title)
	if err := ValidateSlug(slug); err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("slug generation", fmt.Sprintf("Invalid slug generated from title '%s'", *title), err, 1),
			cleanup,
		)
		return
	}

	// Add date prefix if requested
	filename := slug + ".md"
	if *datePrefix {
		dateStr := time.Now().Format("2006-01-02")
		filename = dateStr + "-" + filename
	}

	filePath := filepath.Join(articlesDir, filename)

	// Validate output path
	if err := ValidateOutputPath(filePath); err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("file path validation", fmt.Sprintf("Cannot create article file at '%s'", filePath), err, 1),
			cleanup,
		)
		return
	}

	// Generate article content using selected template
	templates := GetAvailableTemplates()
	selectedTemplate := templates[*template]
	content := selectedTemplate.Generator(*title, *description, *tags, *category, *author, *draft, *featured)

	// Preview mode - show content without writing file
	if *preview {
		showPreview(content, filePath)
		return
	}

	// Write article content
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		apperrors.HandleCLIError(
			apperrors.NewCLIError("file creation", fmt.Sprintf("Failed to write article file '%s'", filePath), err, 1),
			cleanup,
		)
		return
	}

	// Show success message
	showSuccessMessage(filePath, selectedTemplate.Name, *title, *author, *tags, *category, *draft, *featured, *datePrefix)
}

func shouldRunInteractive(fs *flag.FlagSet) bool {
	// Run interactive if no flags were provided
	flagsProvided := false
	fs.Visit(func(_ *flag.Flag) {
		flagsProvided = true
	})
	return !flagsProvided
}

func runInteractiveMode(title, description, tags, category, author, template *string, draft, featured, datePrefix *bool) {
	fmt.Println("🚀 Interactive Article Creator")
	fmt.Println("Press Enter to use defaults shown in [brackets]")
	fmt.Println()

	defaultAuthor := getDefaultAuthor()

	// Check if input is piped
	stat, err := os.Stdin.Stat()
	isPiped := err == nil && (stat.Mode()&os.ModeCharDevice) == 0

	var inputs []string
	if isPiped {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputs = append(inputs, scanner.Text())
		}
	}

	// Get all inputs
	*title = getInputWithPipe("Title", defaultTitle, inputs, 0, isPiped)
	*description = getInputWithPipe("Description", defaultDescription, inputs, 1, isPiped)
	*tags = getInputWithPipe("Tags (comma-separated)", defaultTags, inputs, 2, isPiped)
	*category = getInputWithPipe("Category", defaultCategory, inputs, 3, isPiped)
	*author = getInputWithPipe("Author", defaultAuthor, inputs, 4, isPiped)
	*template = getTemplateInputWithPipe("Template", "default", inputs, 5, isPiped)
	*draft = getBoolInputWithPipe("Draft", defaultDraft, inputs, 6, isPiped)
	*featured = getBoolInputWithPipe("Featured", defaultFeatured, inputs, 7, isPiped)
	*datePrefix = getBoolInputWithPipe("Date prefix filename", false, inputs, 8, isPiped)

	fmt.Println()
}

func getInput(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}

	return input
}

func getBoolInput(prompt string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)
	defaultStr := boolFalse
	if defaultValue {
		defaultStr = boolTrue
	}

	for {
		fmt.Printf("%s (true/false) [%s]: ", prompt, defaultStr)

		input, err := reader.ReadString('\n')
		if err != nil {
			return defaultValue
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" {
			return defaultValue
		}

		switch input {
		case "true", "t", "yes", "y", "1":
			return true
		case "false", "f", "no", "n", "0":
			return false
		default:
			fmt.Println("Please enter 'true' or 'false' (or press Enter for default)")
		}
	}
}

func getInputWithPipe(prompt, defaultValue string, inputs []string, index int, isPiped bool) string {
	if isPiped && index < len(inputs) {
		input := strings.TrimSpace(inputs[index])
		if input != "" {
			return input
		}
		return defaultValue
	}
	return getInput(prompt, defaultValue)
}

func getBoolInputWithPipe(prompt string, defaultValue bool, inputs []string, index int, isPiped bool) bool {
	if isPiped && index < len(inputs) {
		input := strings.TrimSpace(strings.ToLower(inputs[index]))
		if input != "" {
			switch input {
			case "true", "t", "yes", "y", "1":
				return true
			case "false", "f", "no", "n", "0":
				return false
			}
		}
		return defaultValue
	}
	return getBoolInput(prompt, defaultValue)
}

// getTemplateInputWithPipe gets template input with validation
func getTemplateInputWithPipe(prompt, defaultValue string, inputs []string, index int, isPiped bool) string {
	if isPiped && index < len(inputs) {
		input := strings.TrimSpace(inputs[index])
		if input != "" {
			// Validate template exists
			templates := GetAvailableTemplates()
			if _, exists := templates[input]; exists {
				return input
			}
		}
		return defaultValue
	}
	return getTemplateInput(prompt, defaultValue)
}

// getTemplateInput gets template input with validation and help
func getTemplateInput(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	templates := GetAvailableTemplates()

	fmt.Printf("\nAvailable templates:\n")
	for name, template := range templates {
		marker := ""
		if name == defaultValue {
			marker = " (default)"
		}
		fmt.Printf("  %s%s - %s\n", name, marker, template.Description)
	}
	fmt.Println()

	for {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)

		input, err := reader.ReadString('\n')
		if err != nil {
			return defaultValue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			return defaultValue
		}

		// Validate template exists
		if _, exists := templates[input]; exists {
			return input
		}

		fmt.Printf("Template '%s' not found. Available templates: ", input)
		for name := range templates {
			fmt.Printf("%s ", name)
		}
		fmt.Println()
	}
}

func getDefaultAuthor() string {
	if currentUser, err := user.Current(); err == nil {
		return currentUser.Username
	}
	return "Unknown Author"
}

func showHelp() {
	fmt.Println("markgo new - Enhanced markdown blog article generator")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  markgo new [OPTIONS]")
	fmt.Println("  markgo new                    # Interactive mode")
	fmt.Println("  markgo new --interactive      # Force interactive mode")
	fmt.Println()
	fmt.Println("CONTENT OPTIONS:")
	fmt.Printf("  --title       Article title (default: %q)\n", defaultTitle)
	fmt.Printf("  --description Article description (default: %q)\n", defaultDescription)
	fmt.Printf("  --tags        Comma-separated tags (default: %q)\n", defaultTags)
	fmt.Printf("  --category    Article category (default: %q)\n", defaultCategory)
	fmt.Println("  --author      Author name (default: current OS username)")
	fmt.Printf("  --draft       Mark article as draft (default: %v)\n", defaultDraft)
	fmt.Printf("  --featured    Mark article as featured (default: %v)\n", defaultFeatured)
	fmt.Println()
	fmt.Println("TEMPLATE OPTIONS:")
	fmt.Println("  --template    Article template (default: \"default\")")
	fmt.Println("  --list        List available templates")
	fmt.Println()
	fmt.Println("FILE OPTIONS:")
	fmt.Println("  --date-prefix Add date prefix to filename (YYYY-MM-DD-)")
	fmt.Println("  --preview     Preview article without creating file")
	fmt.Println()
	fmt.Println("OTHER OPTIONS:")
	fmt.Println("  --interactive Force interactive mode")
	fmt.Println("  --help        Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  markgo new")
	fmt.Println("  markgo new --list")
	fmt.Println("  markgo new --template tutorial --title \"How to Use Go\"")
	fmt.Println("  markgo new --title \"Hello World\" --tags \"golang,tutorial\" --date-prefix")
	fmt.Println("  markgo new --title \"My Post\" --template review --preview")
	fmt.Println("  markgo new --title \"News Update\" --template news --draft=false --featured=true")
	fmt.Println()
	fmt.Println("AVAILABLE TEMPLATES:")

	templates := GetAvailableTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("  %-12s %s\n", name, templates[name].Description)
	}
}

func slugify(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Define common stop words to remove
	stopWords := []string{"the", "a", "an", "in", "on", "at", "with", "for",
		"to", "of", "is", "are", "and", "or", "but", "from",
		"by", "as", "if", "when", "how", "what", "where", "this"}

	// Split into words and remove stop words (but keep first word if meaningful)
	words := strings.Fields(slug)
	var meaningful []string

	for i, word := range words {
		// Always keep first word, or keep if not a stop word
		if i == 0 || !isStopWord(word, stopWords) {
			// Clean the word of non-alphanumeric characters
			cleanWord := regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(word, "")
			if cleanWord != "" {
				meaningful = append(meaningful, cleanWord)
				if len(meaningful) >= 5 { // Limit to 5 words max
					break
				}
			}
		}
	}

	// Join words with hyphens
	if len(meaningful) == 0 {
		return "untitled"
	}

	return strings.Join(meaningful, "-")
}

func isStopWord(word string, stopWords []string) bool {
	for _, stopWord := range stopWords {
		if word == stopWord {
			return true
		}
	}
	return false
}

// listTemplates shows all available templates
func listTemplates() {
	fmt.Println("📋 Available Article Templates:")
	fmt.Println()

	templates := GetAvailableTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		tmpl := templates[name]
		fmt.Printf("  %s\n", name)
		fmt.Printf("    %s: %s\n", tmpl.Name, tmpl.Description)
		fmt.Println()
	}

	fmt.Println("Usage: markgo new --template <template-name>")
	fmt.Println("Example: markgo new --template tutorial --title \"How to Use Go\"")
}

// showPreview displays the generated article content without creating a file
func showPreview(content, filePath string) {
	fmt.Println("📄 Article Preview")
	fmt.Println("==================")
	fmt.Printf("Would be saved to: %s\n", filePath)
	fmt.Println()
	fmt.Println("Content:")
	fmt.Println("--------")
	fmt.Println(content)
	fmt.Println("--------")
	fmt.Println()
	fmt.Println("💡 Use without --preview flag to create the actual file.")
}

// showSuccessMessage displays a comprehensive success message
func showSuccessMessage(filePath, templateName, title, author, tags, category string, draft, featured, datePrefix bool) {
	fmt.Println("✅ Article Created Successfully!")
	fmt.Println()
	fmt.Printf("📁 File: %s\n", filePath)
	fmt.Printf("📝 Template: %s\n", templateName)
	fmt.Printf("📄 Title: %s\n", title)
	fmt.Printf("👤 Author: %s\n", author)
	fmt.Printf("🏷️  Tags: %s\n", tags)
	fmt.Printf("📁 Category: %s\n", category)
	fmt.Printf("📋 Draft: %v\n", draft)
	fmt.Printf("⭐ Featured: %v\n", featured)

	if datePrefix {
		fmt.Println("📅 Filename includes date prefix")
	}

	fmt.Println()
	fmt.Println("🚀 Next steps:")
	fmt.Println("   1. Edit the article content")
	fmt.Println("   2. Set draft: false when ready to publish")
	fmt.Println("   3. Add more tags if needed")
}
