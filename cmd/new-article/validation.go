package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}

// ValidationResult contains the results of input validation
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// ValidateArticleInput validates all article input parameters
func ValidateArticleInput(title, description, tags, category, author string, template string) ValidationResult {
	var errors []ValidationError

	// Validate title
	if titleErr := validateTitle(title); titleErr != nil {
		errors = append(errors, *titleErr)
	}

	// Validate description
	if descErr := validateDescription(description); descErr != nil {
		errors = append(errors, *descErr)
	}

	// Validate tags
	if tagsErr := validateTags(tags); tagsErr != nil {
		errors = append(errors, *tagsErr)
	}

	// Validate category
	if catErr := validateCategory(category); catErr != nil {
		errors = append(errors, *catErr)
	}

	// Validate author
	if authorErr := validateAuthor(author); authorErr != nil {
		errors = append(errors, *authorErr)
	}

	// Validate template
	if templateErr := validateTemplate(template); templateErr != nil {
		errors = append(errors, *templateErr)
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// validateTitle validates the article title
func validateTitle(title string) *ValidationError {
	title = strings.TrimSpace(title)

	if title == "" {
		return &ValidationError{
			Field:   "title",
			Value:   title,
			Message: "title cannot be empty",
		}
	}

	if utf8.RuneCountInString(title) > 200 {
		return &ValidationError{
			Field:   "title",
			Value:   title,
			Message: "title cannot exceed 200 characters",
		}
	}

	if utf8.RuneCountInString(title) < 3 {
		return &ValidationError{
			Field:   "title",
			Value:   title,
			Message: "title must be at least 3 characters long",
		}
	}

	// Note: We check for potentially problematic characters that could break YAML
	// but YAML can handle quotes if escaped, which we do in the generator
	_ = strings.ContainsAny(title, `"'`) // Future: could add warning if needed

	return nil
}

// validateDescription validates the article description
func validateDescription(description string) *ValidationError {
	description = strings.TrimSpace(description)

	if utf8.RuneCountInString(description) > 500 {
		return &ValidationError{
			Field:   "description",
			Value:   description,
			Message: "description cannot exceed 500 characters",
		}
	}

	return nil
}

// validateTags validates the tags string
func validateTags(tags string) *ValidationError {
	tags = strings.TrimSpace(tags)

	if tags == "" {
		return &ValidationError{
			Field:   "tags",
			Value:   tags,
			Message: "at least one tag is required",
		}
	}

	tagList := strings.Split(tags, ",")
	if len(tagList) > 10 {
		return &ValidationError{
			Field:   "tags",
			Value:   tags,
			Message: "cannot have more than 10 tags",
		}
	}

	// Validate individual tags
	validTagPattern := regexp.MustCompile(`^[a-zA-Z0-9\-_\s]+$`)
	for i, tag := range tagList {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue // Skip empty tags
		}

		if utf8.RuneCountInString(tag) > 50 {
			return &ValidationError{
				Field:   "tags",
				Value:   tag,
				Message: fmt.Sprintf("tag %d '%s' cannot exceed 50 characters", i+1, tag),
			}
		}

		if !validTagPattern.MatchString(tag) {
			return &ValidationError{
				Field:   "tags",
				Value:   tag,
				Message: fmt.Sprintf("tag %d '%s' contains invalid characters (use only letters, numbers, hyphens, underscores, and spaces)", i+1, tag),
			}
		}
	}

	return nil
}

// validateCategory validates the category
func validateCategory(category string) *ValidationError {
	category = strings.TrimSpace(category)

	if category == "" {
		return &ValidationError{
			Field:   "category",
			Value:   category,
			Message: "category cannot be empty",
		}
	}

	if utf8.RuneCountInString(category) > 100 {
		return &ValidationError{
			Field:   "category",
			Value:   category,
			Message: "category cannot exceed 100 characters",
		}
	}

	// Category should be URL-friendly
	validCategoryPattern := regexp.MustCompile(`^[a-zA-Z0-9\-_\s]+$`)
	if !validCategoryPattern.MatchString(category) {
		return &ValidationError{
			Field:   "category",
			Value:   category,
			Message: "category contains invalid characters (use only letters, numbers, hyphens, underscores, and spaces)",
		}
	}

	return nil
}

// validateAuthor validates the author name
func validateAuthor(author string) *ValidationError {
	author = strings.TrimSpace(author)

	if author == "" {
		return &ValidationError{
			Field:   "author",
			Value:   author,
			Message: "author name cannot be empty",
		}
	}

	if utf8.RuneCountInString(author) > 100 {
		return &ValidationError{
			Field:   "author",
			Value:   author,
			Message: "author name cannot exceed 100 characters",
		}
	}

	if utf8.RuneCountInString(author) < 2 {
		return &ValidationError{
			Field:   "author",
			Value:   author,
			Message: "author name must be at least 2 characters long",
		}
	}

	return nil
}

// validateTemplate validates the template name
func validateTemplate(template string) *ValidationError {
	if template == "" {
		return nil // Empty is OK, will use default
	}

	templates := GetAvailableTemplates()
	if _, exists := templates[template]; !exists {
		available := make([]string, 0, len(templates))
		for name := range templates {
			available = append(available, name)
		}
		return &ValidationError{
			Field:   "template",
			Value:   template,
			Message: fmt.Sprintf("template '%s' not found. Available templates: %s", template, strings.Join(available, ", ")),
		}
	}

	return nil
}

// ValidateOutputPath validates the output file path
func ValidateOutputPath(filePath string) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("article file already exists: %s", filePath)
	}

	// Check if directory is writable
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	// Try to create a temporary file to test write permissions
	tempFile := filePath + ".tmp"
	file, err := os.Create(tempFile) // #nosec G304
	if err != nil {
		return fmt.Errorf("cannot write to output directory: %w", err)
	}
	_ = file.Close()
	_ = os.Remove(tempFile)

	return nil
}

// SanitizeInput sanitizes user input to prevent issues
func SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	input = spaceRegex.ReplaceAllString(input, " ")

	return input
}

// SanitizeForYAML sanitizes strings for safe YAML output
func SanitizeForYAML(input string) string {
	// Escape quotes in YAML strings
	input = strings.ReplaceAll(input, `"`, `\"`)
	return input
}

// ValidateSlug validates a generated slug
func ValidateSlug(slug string) error {
	if slug == "" {
		return errors.New("slug cannot be empty")
	}

	if len(slug) > 100 {
		return errors.New("slug too long (max 100 characters)")
	}

	// Slug should only contain lowercase letters, numbers, and hyphens
	validSlugPattern := regexp.MustCompile(`^[a-z0-9\-]+$`)
	if !validSlugPattern.MatchString(slug) {
		return fmt.Errorf("invalid slug format: %s (use only lowercase letters, numbers, and hyphens)", slug)
	}

	// Should not start or end with hyphen
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return fmt.Errorf("slug cannot start or end with hyphen: %s", slug)
	}

	// Should not have consecutive hyphens
	if strings.Contains(slug, "--") {
		return fmt.Errorf("slug cannot have consecutive hyphens: %s", slug)
	}

	return nil
}

// ShowValidationErrors displays validation errors in a user-friendly format
func ShowValidationErrors(errors []ValidationError) {
	fmt.Println("❌ Validation failed:")
	fmt.Println()
	for i, err := range errors {
		fmt.Printf("  %d. %s: %s\n", i+1, strings.ToUpper(err.Field[:1])+err.Field[1:], err.Message)
		if err.Value != "" && len(err.Value) < 50 {
			fmt.Printf("     Value: '%s'\n", err.Value)
		}
	}
	fmt.Println()
	fmt.Println("Please fix these issues and try again.")
}
