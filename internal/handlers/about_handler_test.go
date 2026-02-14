package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/1mb-dev/markgo/internal/config"
)

func createTestAboutHandler(cfg *config.Config) *AboutHandler {
	if cfg == nil {
		cfg = createTestConfig()
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})
	return NewAboutHandler(base, &MockArticleService{}, &MockMarkdownRenderer{})
}

func TestAboutHandler(t *testing.T) {
	t.Run("minimal config shows author name", func(t *testing.T) {
		cfg := &config.Config{
			Environment: "test",
			BaseURL:     "http://localhost:3000",
			Blog: config.BlogConfig{
				Title:  "Test Blog",
				Author: "Test Author",
			},
		}

		handler := createTestAboutHandler(cfg)

		router := gin.New()
		router.GET("/about", handler.ShowAbout)

		req := httptest.NewRequest("GET", "/about", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("full config sets all template data", func(t *testing.T) {
		cfg := &config.Config{
			Environment: "test",
			BaseURL:     "http://localhost:3000",
			Blog: config.BlogConfig{
				Title:       "Test Blog",
				Author:      "Jane Doe",
				AuthorEmail: "jane@example.com",
			},
			About: config.AboutConfig{
				Avatar:   "img/avatar.jpg",
				Tagline:  "Building things",
				Location: "San Francisco, CA",
				GitHub:   "janedoe",
				Twitter:  "@janedoe",
				LinkedIn: "https://linkedin.com/in/janedoe",
				Website:  "janedoe.com",
			},
			Email: config.EmailConfig{
				Host:     "smtp.example.com",
				Username: "user",
			},
		}

		handler := createTestAboutHandler(cfg)

		router := gin.New()
		router.GET("/about", handler.ShowAbout)

		req := httptest.NewRequest("GET", "/about", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("contact section hidden without email", func(t *testing.T) {
		cfg := &config.Config{
			Environment: "test",
			BaseURL:     "http://localhost:3000",
			Blog: config.BlogConfig{
				Title:       "Test Blog",
				Author:      "Test Author",
				AuthorEmail: "", // no email
			},
		}

		handler := createTestAboutHandler(cfg)

		router := gin.New()
		router.GET("/about", handler.ShowAbout)

		req := httptest.NewRequest("GET", "/about", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestBuildSocialLinks(t *testing.T) {
	t.Run("no social links configured", func(t *testing.T) {
		handler := createTestAboutHandler(nil)
		links := handler.buildSocialLinks()
		assert.Empty(t, links)
	})

	t.Run("normalizes github username to full URL", func(t *testing.T) {
		cfg := createTestConfig()
		cfg.About.GitHub = "testuser"
		handler := createTestAboutHandler(cfg)
		links := handler.buildSocialLinks()

		assert.Len(t, links, 1)
		assert.Equal(t, "github", links[0].Platform)
		assert.Equal(t, "https://github.com/testuser", links[0].URL)
	})

	t.Run("preserves full URLs", func(t *testing.T) {
		cfg := createTestConfig()
		cfg.About.GitHub = "https://github.com/testuser"
		handler := createTestAboutHandler(cfg)
		links := handler.buildSocialLinks()

		assert.Len(t, links, 1)
		assert.Equal(t, "https://github.com/testuser", links[0].URL)
	})

	t.Run("normalizes twitter handle", func(t *testing.T) {
		cfg := createTestConfig()
		cfg.About.Twitter = "@janedoe"
		handler := createTestAboutHandler(cfg)
		links := handler.buildSocialLinks()

		assert.Len(t, links, 1)
		assert.Equal(t, "https://x.com/janedoe", links[0].URL)
	})

	t.Run("all platforms configured", func(t *testing.T) {
		cfg := createTestConfig()
		cfg.About.GitHub = "user"
		cfg.About.Twitter = "user"
		cfg.About.LinkedIn = "https://linkedin.com/in/user"
		cfg.About.Mastodon = "https://mastodon.social/@user"
		cfg.About.Website = "example.com"
		handler := createTestAboutHandler(cfg)
		links := handler.buildSocialLinks()

		assert.Len(t, links, 5)
		assert.Equal(t, "github", links[0].Platform)
		assert.Equal(t, "twitter", links[1].Platform)
		assert.Equal(t, "linkedin", links[2].Platform)
		assert.Equal(t, "mastodon", links[3].Platform)
		assert.Equal(t, "website", links[4].Platform)
	})
}

func TestNormalizeURL(t *testing.T) {
	assert.Equal(t, "https://github.com/user", normalizeURL("user", "https://github.com/"))
	assert.Equal(t, "https://github.com/user", normalizeURL("https://github.com/user", "https://github.com/"))
	assert.Equal(t, "https://example.com", normalizeURL("example.com", "https://"))
}
