package config

import (
	"os"
	"testing"
)

func TestSEOConfigLoading(t *testing.T) {
	// Set environment variables for testing
	originalEnvs := make(map[string]string)
	testEnvs := map[string]string{
		"SEO_ENABLED":              "true",
		"SEO_SITEMAP_ENABLED":      "true",
		"SEO_SCHEMA_ENABLED":       "true",
		"SEO_OPEN_GRAPH_ENABLED":   "true",
		"SEO_TWITTER_CARD_ENABLED": "true",
		"SEO_DEFAULT_IMAGE":        "/images/default.jpg",
		"SEO_TWITTER_SITE":         "@myblog",
		"SEO_GOOGLE_SITE_VERIFY":   "google123",
		"SEO_BING_SITE_VERIFY":     "bing456",
	}

	// Save original env values and set test values
	for key, value := range testEnvs {
		originalEnvs[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// Restore original environment after test
	defer func() {
		for key, originalValue := range originalEnvs {
			if originalValue == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, originalValue)
			}
		}
	}()

	// Load configuration
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test SEO configuration
	if !cfg.SEO.Enabled {
		t.Error("SEO should be enabled")
	}

	if !cfg.SEO.SitemapEnabled {
		t.Error("Sitemap should be enabled")
	}

	if !cfg.SEO.SchemaEnabled {
		t.Error("Schema should be enabled")
	}

	if !cfg.SEO.OpenGraphEnabled {
		t.Error("Open Graph should be enabled")
	}

	if !cfg.SEO.TwitterCardEnabled {
		t.Error("Twitter Card should be enabled")
	}

	if cfg.SEO.DefaultImage != "/images/default.jpg" {
		t.Errorf("Expected default image '/images/default.jpg', got '%s'", cfg.SEO.DefaultImage)
	}

	if cfg.SEO.TwitterSite != "@myblog" {
		t.Errorf("Expected Twitter site '@myblog', got '%s'", cfg.SEO.TwitterSite)
	}

	if cfg.SEO.GoogleSiteVerify != "google123" {
		t.Errorf("Expected Google verification 'google123', got '%s'", cfg.SEO.GoogleSiteVerify)
	}

	if cfg.SEO.BingSiteVerify != "bing456" {
		t.Errorf("Expected Bing verification 'bing456', got '%s'", cfg.SEO.BingSiteVerify)
	}
}

func TestSEOConfigDefaults(t *testing.T) {
	// Clear all SEO environment variables
	seoEnvVars := []string{
		"SEO_ENABLED", "SEO_SITEMAP_ENABLED", "SEO_SCHEMA_ENABLED",
		"SEO_OPEN_GRAPH_ENABLED", "SEO_TWITTER_CARD_ENABLED",
		"SEO_DEFAULT_IMAGE", "SEO_TWITTER_SITE", "SEO_GOOGLE_SITE_VERIFY",
	}

	originalEnvs := make(map[string]string)
	for _, envVar := range seoEnvVars {
		originalEnvs[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}

	// Restore environment after test
	defer func() {
		for key, originalValue := range originalEnvs {
			if originalValue != "" {
				os.Setenv(key, originalValue)
			}
		}
	}()

	// Load configuration with defaults
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test default values
	if !cfg.SEO.Enabled {
		t.Error("SEO should be enabled by default")
	}

	if !cfg.SEO.SitemapEnabled {
		t.Error("Sitemap should be enabled by default")
	}

	if !cfg.SEO.SchemaEnabled {
		t.Error("Schema should be enabled by default")
	}

	if !cfg.SEO.OpenGraphEnabled {
		t.Error("Open Graph should be enabled by default")
	}

	if !cfg.SEO.TwitterCardEnabled {
		t.Error("Twitter Card should be enabled by default")
	}

	// Check default robots configuration
	expectedDisallowed := []string{"/admin", "/api", "/preview"}
	if len(cfg.SEO.RobotsDisallowed) != len(expectedDisallowed) {
		t.Errorf("Expected %d disallowed paths, got %d", len(expectedDisallowed), len(cfg.SEO.RobotsDisallowed))
	}

	for i, expected := range expectedDisallowed {
		if i >= len(cfg.SEO.RobotsDisallowed) || cfg.SEO.RobotsDisallowed[i] != expected {
			t.Errorf("Expected disallowed path '%s', got '%s'", expected, cfg.SEO.RobotsDisallowed[i])
		}
	}
}

func TestSEOConfigDisabled(t *testing.T) {
	// Set SEO to disabled
	originalEnabled := os.Getenv("SEO_ENABLED")
	os.Setenv("SEO_ENABLED", "false")

	defer func() {
		if originalEnabled == "" {
			os.Unsetenv("SEO_ENABLED")
		} else {
			os.Setenv("SEO_ENABLED", originalEnabled)
		}
	}()

	// Load configuration
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that SEO is disabled
	if cfg.SEO.Enabled {
		t.Error("SEO should be disabled when SEO_ENABLED=false")
	}
}
