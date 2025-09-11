

FUNCTIONS

func GetAvailableTemplates() map[string]ArticleTemplate
    GetAvailableTemplates returns all available article templates

func SanitizeForYAML(input string) string
    SanitizeForYAML sanitizes strings for safe YAML output

func SanitizeInput(input string) string
    SanitizeInput sanitizes user input to prevent issues

func ShowValidationErrors(errors []ValidationError)
    ShowValidationErrors displays validation errors in a user-friendly format

func ValidateOutputPath(filePath string) error
    ValidateOutputPath validates the output file path

func ValidateSlug(slug string) error
    ValidateSlug validates a generated slug


TYPES

type ArticleTemplate struct {
	Name        string
	Description string
	Generator   func(title, description, tagsStr, category, author string, isDraft, isFeatured bool) string
}
    ArticleTemplate defines a template for creating articles

type ValidationError struct {
	Field   string
	Value   string
	Message string
}
    ValidationError represents a validation error

func (e ValidationError) Error() string

type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}
    ValidationResult contains the results of input validation

func ValidateArticleInput(title, description, tags, category, author string, template string) ValidationResult
    ValidateArticleInput validates all article input parameters

