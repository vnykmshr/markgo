package handlers

import (
	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services"
)

// SEODataHelper generates SEO data for templates
type SEODataHelper struct {
	seoService services.SEOServiceInterface
	config     *config.Config
}

// NewSEODataHelper creates a new SEO data helper
func NewSEODataHelper(seoService services.SEOServiceInterface, config *config.Config) *SEODataHelper {
	return &SEODataHelper{
		seoService: seoService,
		config:     config,
	}
}

// GenerateArticleSEOData generates comprehensive SEO data for article pages
func (h *SEODataHelper) GenerateArticleSEOData(article *models.Article) map[string]interface{} {
	if h.seoService == nil || !h.seoService.IsEnabled() {
		return map[string]interface{}{}
	}

	seoData := make(map[string]interface{})
	baseURL := h.config.BaseURL

	// Generate site config for SEO service
	siteConfig := services.SiteConfig{
		Name:        h.config.Blog.Title,
		Description: h.config.Blog.Description,
		BaseURL:     baseURL,
		Language:    h.config.Blog.Language,
		Author:      h.config.Blog.Author,
		Logo:        h.config.SEO.DefaultImage,
		Image:       h.config.SEO.DefaultImage,
	}

	// Generate meta tags
	if metaTags, err := h.seoService.GenerateMetaTags(article, siteConfig); err == nil {
		seoData["seoMetaTags"] = metaTags
	}

	// Generate Open Graph tags
	if ogTags, err := h.seoService.GenerateOpenGraphTags(article, baseURL); err == nil {
		// Merge with existing meta tags
		if seoMetaTags, exists := seoData["seoMetaTags"].(map[string]string); exists {
			for key, value := range ogTags {
				seoMetaTags[key] = value
			}
		} else {
			seoData["seoMetaTags"] = ogTags
		}
	}

	// Generate Twitter Card tags
	if twitterTags, err := h.seoService.GenerateTwitterCardTags(article, baseURL); err == nil {
		// Merge with existing meta tags
		if seoMetaTags, exists := seoData["seoMetaTags"].(map[string]string); exists {
			for key, value := range twitterTags {
				seoMetaTags[key] = value
			}
		}
	}

	// Generate Article Schema
	if articleSchema, err := h.seoService.GenerateArticleSchema(article, baseURL); err == nil {
		seoData["articleSchema"] = articleSchema
	}

	// Generate Website Schema
	if websiteSchema, err := h.seoService.GenerateWebsiteSchema(siteConfig); err == nil {
		seoData["websiteSchema"] = websiteSchema
	}

	// Generate Breadcrumb Schema for article
	breadcrumbs := []services.Breadcrumb{
		{Name: "Home", URL: baseURL},
		{Name: "Articles", URL: baseURL + "/articles"},
		{Name: article.Title, URL: baseURL + "/article/" + article.Slug},
	}
	if breadcrumbSchema, err := h.seoService.GenerateBreadcrumbSchema(breadcrumbs); err == nil {
		seoData["breadcrumbSchema"] = breadcrumbSchema
	}

	return seoData
}

// GeneratePageSEOData generates SEO data for non-article pages
func (h *SEODataHelper) GeneratePageSEOData(title, description, path string) map[string]interface{} {
	if h.seoService == nil || !h.seoService.IsEnabled() {
		return map[string]interface{}{}
	}

	seoData := make(map[string]interface{})

	// Generate site config
	siteConfig := services.SiteConfig{
		Name:        h.config.Blog.Title,
		Description: h.config.Blog.Description,
		BaseURL:     h.config.BaseURL,
		Language:    h.config.Blog.Language,
		Author:      h.config.Blog.Author,
		Logo:        h.config.SEO.DefaultImage,
		Image:       h.config.SEO.DefaultImage,
	}

	// Generate page meta tags (using regular meta tags for pages)
	// Create a temporary article-like structure for page meta generation
	tempPageData := map[string]string{
		"title":               title,
		"description":         description,
		"canonical":           h.config.BaseURL + path,
		"og:type":             "website",
		"og:title":            title,
		"og:description":      description,
		"og:site_name":        siteConfig.Name,
		"twitter:card":        "summary",
		"twitter:title":       title,
		"twitter:description": description,
	}
	seoData["seoMetaTags"] = tempPageData

	// Generate Website Schema
	if websiteSchema, err := h.seoService.GenerateWebsiteSchema(siteConfig); err == nil {
		seoData["websiteSchema"] = websiteSchema
	}

	// Generate Breadcrumb Schema for pages
	breadcrumbs := []services.Breadcrumb{
		{Name: "Home", URL: h.config.BaseURL},
	}
	if path != "/" && path != "" {
		breadcrumbs = append(breadcrumbs, services.Breadcrumb{
			Name: title,
			URL:  h.config.BaseURL + path,
		})
	}
	if breadcrumbSchema, err := h.seoService.GenerateBreadcrumbSchema(breadcrumbs); err == nil {
		seoData["breadcrumbSchema"] = breadcrumbSchema
	}

	return seoData
}

// AddSEODataToTemplateData adds SEO data to existing template data
func (h *SEODataHelper) AddSEODataToTemplateData(templateData map[string]interface{}, seoData map[string]interface{}) {
	for key, value := range seoData {
		templateData[key] = value
	}
}

// GenerateHomeSEOData generates SEO data for the home page
func (h *SEODataHelper) GenerateHomeSEOData() map[string]interface{} {
	return h.GeneratePageSEOData(
		h.config.Blog.Title,
		h.config.Blog.Description,
		"/",
	)
}
