package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	defaultTitle       = "Untitled Article"
	defaultDescription = ""
	defaultTags        = "general"
	defaultCategory    = "uncategorized"
	defaultDraft       = true
	defaultFeatured    = false
	articlesDir        = "articles"
)

var (
	title       = flag.String("title", defaultTitle, "Article title")
	description = flag.String("description", defaultDescription, "Article description")
	tags        = flag.String("tags", defaultTags, "Comma-separated tags")
	category    = flag.String("category", defaultCategory, "Article category")
	author      = flag.String("author", "", "Author name (default: current OS username)")
	draft       = flag.Bool("draft", defaultDraft, "Mark article as draft")
	featured    = flag.Bool("featured", defaultFeatured, "Mark article as featured")
	template    = flag.String("template", "default", "Article template to use")
	preview     = flag.Bool("preview", false, "Preview the article without creating file")
	list        = flag.Bool("list", false, "List available templates")
	datePrefix  = flag.Bool("date-prefix", false, "Add date prefix to filename")
	interactive = flag.Bool("interactive", false, "Force interactive mode")
	help        = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *list {
		listTemplates()
		return
	}

	// Check if we should run interactive mode
	if *interactive || shouldRunInteractive() {
		runInteractiveMode()
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
		os.Exit(1)
	}

	// Generate filename from title
	slug := slugify(*title)
	if err := ValidateSlug(slug); err != nil {
		slog.Error("Invalid slug generated from title", "slug", slug, "error", err)
		os.Exit(1)
	}

	// Add date prefix if requested
	filename := slug + ".markdown"
	if *datePrefix {
		dateStr := time.Now().Format("2006-01-02")
		filename = dateStr + "-" + filename
	}

	filepath := filepath.Join(articlesDir, filename)

	// Validate output path
	if err := ValidateOutputPath(filepath); err != nil {
		slog.Error("Cannot create article file", "filepath", filepath, "error", err)
		os.Exit(1)
	}

	// Generate article content using selected template
	templates := GetAvailableTemplates()
	selectedTemplate := templates[*template]
	content := selectedTemplate.Generator(*title, *description, *tags, *category, *author, *draft, *featured)

	// Preview mode - show content without writing file
	if *preview {
		showPreview(content, filepath)
		return
	}

	// Write article content
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		slog.Error("Failed to write article file", "filepath", filepath, "error", err)
		os.Exit(1)
	}

	// Show success message
	showSuccessMessage(filepath, selectedTemplate.Name)
}

func shouldRunInteractive() bool {
	// Run interactive if no flags were provided
	flagsProvided := false
	flag.Visit(func(f *flag.Flag) {
		flagsProvided = true
	})
	return !flagsProvided
}

func runInteractiveMode() {
	fmt.Println("🚀 Interactive Article Creator")
	fmt.Println("Press Enter to use defaults shown in [brackets]")
	fmt.Println()

	defaultAuthor := getDefaultAuthor()

	// Check if input is piped
	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

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
	defaultStr := "false"
	if defaultValue {
		defaultStr = "true"
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
	fmt.Println("new-article - Enhanced markdown blog article generator")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  new-article [OPTIONS]")
	fmt.Println("  new-article                    # Interactive mode")
	fmt.Println("  new-article --interactive      # Force interactive mode")
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
	fmt.Println("  new-article")
	fmt.Println("  new-article --list")
	fmt.Println("  new-article --template tutorial --title \"How to Use Go\"")
	fmt.Println("  new-article --title \"Hello World\" --tags \"golang,tutorial\" --date-prefix")
	fmt.Println("  new-article --title \"My Post\" --template review --preview")
	fmt.Println("  new-article --title \"News Update\" --template news --draft=false --featured=true")
	fmt.Println()
	fmt.Println("AVAILABLE TEMPLATES:")

	templates := GetAvailableTemplates()
	for name, template := range templates {
		fmt.Printf("  %-12s %s\n", name, template.Description)
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
	for name, template := range templates {
		fmt.Printf("  %s\n", name)
		fmt.Printf("    %s: %s\n", template.Name, template.Description)
		fmt.Println()
	}

	fmt.Println("Usage: new-article --template <template-name>")
	fmt.Println("Example: new-article --template tutorial --title \"How to Use Go\"")
}

// showPreview displays the generated article content without creating a file
func showPreview(content, filepath string) {
	fmt.Println("📄 Article Preview")
	fmt.Println("==================")
	fmt.Printf("Would be saved to: %s\n", filepath)
	fmt.Println()
	fmt.Println("Content:")
	fmt.Println("--------")
	fmt.Println(content)
	fmt.Println("--------")
	fmt.Println()
	fmt.Println("💡 Use without --preview flag to create the actual file.")
}

// showSuccessMessage displays a comprehensive success message
func showSuccessMessage(filepath, templateName string) {
	fmt.Println("✅ Article Created Successfully!")
	fmt.Println()
	fmt.Printf("📁 File: %s\n", filepath)
	fmt.Printf("📝 Template: %s\n", templateName)
	fmt.Printf("📄 Title: %s\n", *title)
	fmt.Printf("👤 Author: %s\n", *author)
	fmt.Printf("🏷️  Tags: %s\n", *tags)
	fmt.Printf("📁 Category: %s\n", *category)
	fmt.Printf("📋 Draft: %v\n", *draft)
	fmt.Printf("⭐ Featured: %v\n", *featured)

	if *datePrefix {
		fmt.Println("📅 Filename includes date prefix")
	}

	fmt.Println()
	fmt.Println("🚀 Next steps:")
	fmt.Println("   1. Edit the article content")
	fmt.Println("   2. Set draft: false when ready to publish")
	fmt.Println("   3. Add more tags if needed")
}
