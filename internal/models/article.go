package models

import (
	"crypto/md5"
	"sync"
	"time"
)

// Article represents a blog article with memory optimization
type Article struct {
	Slug             string    `yaml:"slug" json:"slug"`
	Title            string    `yaml:"title" json:"title"`
	Description      string    `yaml:"description" json:"description"`
	Date             time.Time `yaml:"date" json:"date"`
	Tags             []string  `yaml:"tags" json:"tags"`
	Categories       []string  `yaml:"categories" json:"categories"`
	Draft            bool      `yaml:"draft" json:"draft"`
	Featured         bool      `yaml:"featured" json:"featured"`
	Author           string    `yaml:"author" json:"author"`
	Content          string    `yaml:"-" json:"content"`
	ReadingTime      int       `yaml:"-" json:"reading_time"`
	WordCount        int       `yaml:"-" json:"word_count"`
	LastModified     time.Time `yaml:"-" json:"last_modified"`
	
	// Memory optimized fields - lazy loaded
	processedContent *string       `yaml:"-" json:"-"`
	excerpt          *string       `yaml:"-" json:"-"`
	contentHash      [16]byte      `yaml:"-" json:"-"`
	mu               sync.RWMutex  `yaml:"-" json:"-"`
	processor        ArticleProcessor `yaml:"-" json:"-"`
}

// ArticleProcessor interface for lazy content processing
type ArticleProcessor interface {
	ProcessMarkdown(content string) (string, error)
	GenerateExcerpt(content string, maxLength int) string
	ProcessDuplicateTitles(title, htmlContent string) string
}

// GetProcessedContent returns the processed HTML content, generating it if needed
func (a *Article) GetProcessedContent() string {
	a.mu.RLock()
	if a.processedContent != nil {
		content := *a.processedContent
		a.mu.RUnlock()
		return content
	}
	a.mu.RUnlock()

	// Generate processed content
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Double-check pattern - another goroutine might have processed it
	if a.processedContent != nil {
		return *a.processedContent
	}
	
	// Generate content hash for cache invalidation
	hash := md5.Sum([]byte(a.Content))
	if a.contentHash != hash && a.processor != nil {
		processed, err := a.processor.ProcessMarkdown(a.Content)
		if err == nil {
			// Process duplicate titles after markdown conversion
			processed = a.processor.ProcessDuplicateTitles(a.Title, processed)
			a.processedContent = &processed
			a.contentHash = hash
			return processed
		}
	}
	
	// Fallback to raw content if processing fails
	a.processedContent = &a.Content
	return a.Content
}

// GetExcerpt returns the article excerpt, generating it if needed
func (a *Article) GetExcerpt() string {
	a.mu.RLock()
	if a.excerpt != nil {
		excerpt := *a.excerpt
		a.mu.RUnlock()
		return excerpt
	}
	a.mu.RUnlock()

	// Generate excerpt
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Double-check pattern
	if a.excerpt != nil {
		return *a.excerpt
	}
	
	if a.processor != nil {
		excerpt := a.processor.GenerateExcerpt(a.Content, 200)
		a.excerpt = &excerpt
		return excerpt
	}
	
	// Fallback to description or truncated content
	if a.Description != "" {
		a.excerpt = &a.Description
		return a.Description
	}
	
	const maxLength = 200
	content := a.Content
	if len(content) > maxLength {
		content = content[:maxLength] + "..."
	}
	a.excerpt = &content
	return content
}

// SetProcessor sets the processor for lazy content processing
func (a *Article) SetProcessor(processor ArticleProcessor) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.processor = processor
}

// ProcessedContent returns the processed content (for JSON serialization compatibility)
func (a *Article) ProcessedContent() string {
	return a.GetProcessedContent()
}

// Excerpt returns the excerpt (for JSON serialization compatibility)  
func (a *Article) Excerpt() string {
	return a.GetExcerpt()
}

// ClearProcessedContent clears cached processed content (for memory optimization)
func (a *Article) ClearProcessedContent() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.processedContent = nil
	a.excerpt = nil
}

// ContactMessage represents a contact form submission
type ContactMessage struct {
	Name            string `json:"name" binding:"required,min=2,max=50"`
	Email           string `json:"email" binding:"required,email"`
	Subject         string `json:"subject" binding:"required,min=5,max=100"`
	Message         string `json:"message" binding:"required,min=10,max=2000"`
	CaptchaQuestion string `json:"captcha_question" binding:"required"`
	CaptchaAnswer   string `json:"captcha_answer" binding:"required"`
}

// SearchResult represents a search result with scoring
type SearchResult struct {
	*Article
	Score         float64  `json:"score"`
	MatchedFields []string `json:"matched_fields"`
}

// ArticleList represents a simplified article for listing pages
type ArticleList struct {
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Tags        []string  `json:"tags"`
	Categories  []string  `json:"categories"`
	Excerpt     string    `json:"excerpt"`
	ReadingTime int       `json:"reading_time"`
	Featured    bool      `json:"featured"`
}

// ToListView converts an Article to ArticleList
func (a *Article) ToListView() *ArticleList {
	return &ArticleList{
		Slug:        a.Slug,
		Title:       a.Title,
		Description: a.Description,
		Date:        a.Date,
		Tags:        a.Tags,
		Categories:  a.Categories,
		Excerpt:     a.GetExcerpt(),
		ReadingTime: a.ReadingTime,
		Featured:    a.Featured,
	}
}

// Feed represents RSS/JSON feed structure
type Feed struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Link        string     `json:"home_page_url"`
	FeedURL     string     `json:"feed_url"`
	Language    string     `json:"language"`
	Updated     time.Time  `json:"date_modified"`
	Author      Author     `json:"author"`
	Items       []FeedItem `json:"items"`
}

// Author represents the blog author
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	URL   string `json:"url"`
}

// FeedItem represents an item in RSS/JSON feed
type FeedItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	ContentHTML string    `json:"content_html"`
	URL         string    `json:"url"`
	Summary     string    `json:"summary"`
	Published   time.Time `json:"date_published"`
	Modified    time.Time `json:"date_modified"`
	Tags        []string  `json:"tags"`
	Author      Author    `json:"author"`
}

// SitemapURL represents a URL in the sitemap
type SitemapURL struct {
	Loc        string    `xml:"loc"`
	LastMod    time.Time `xml:"lastmod"`
	ChangeFreq string    `xml:"changefreq"`
	Priority   float32   `xml:"priority"`
}

// Sitemap represents the XML sitemap structure
type Sitemap struct {
	XMLName string       `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// Stats represents blog statistics
type Stats struct {
	TotalArticles   int            `json:"total_articles"`
	PublishedCount  int            `json:"published_count"`
	DraftCount      int            `json:"draft_count"`
	TotalTags       int            `json:"total_tags"`
	TotalCategories int            `json:"total_categories"`
	PopularTags     []TagCount     `json:"popular_tags"`
	RecentArticles  []*ArticleList `json:"recent_articles"`
	LastUpdated     time.Time      `json:"last_updated"`
}

// TagCount represents tag usage statistics
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// CategoryCount represents category usage statistics
type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// Pagination represents pagination information
type Pagination struct {
	CurrentPage  int  `json:"current_page"`
	TotalPages   int  `json:"total_pages"`
	TotalItems   int  `json:"total_items"`
	ItemsPerPage int  `json:"items_per_page"`
	HasPrevious  bool `json:"has_previous"`
	HasNext      bool `json:"has_next"`
	PreviousPage int  `json:"previous_page"`
	NextPage     int  `json:"next_page"`
}

// NewPagination creates a new pagination struct
func NewPagination(currentPage, totalItems, itemsPerPage int) *Pagination {
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	hasPrevious := currentPage > 1
	hasNext := currentPage < totalPages

	previousPage := max(currentPage-1, 1)

	nextPage := min(currentPage+1, totalPages)

	return &Pagination{
		CurrentPage:  currentPage,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		ItemsPerPage: itemsPerPage,
		HasPrevious:  hasPrevious,
		HasNext:      hasNext,
		PreviousPage: previousPage,
		NextPage:     nextPage,
	}
}
