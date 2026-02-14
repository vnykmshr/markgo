package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/1mb-dev/markgo/internal/models"
)

// WritingArticleService returns canned published articles for Writing handler tests.
type WritingArticleService struct {
	MockArticleService
	Articles []*models.Article
}

func (m *WritingArticleService) GetAllArticles() []*models.Article { return m.Articles }

// AdminHomeArticleService returns canned data for AdminHome handler tests.
type AdminHomeArticleService struct {
	MockArticleService
	Published  []*models.Article
	Drafts     []*models.Article
	Tags       []models.TagCount
	Categories []models.CategoryCount
}

func (m *AdminHomeArticleService) GetAllArticles() []*models.Article         { return m.Published }
func (m *AdminHomeArticleService) GetDraftArticles() []*models.Article       { return m.Drafts }
func (m *AdminHomeArticleService) GetTagCounts() []models.TagCount           { return m.Tags }
func (m *AdminHomeArticleService) GetCategoryCounts() []models.CategoryCount { return m.Categories }

// ---------------------------------------------------------------------------
// T3: AdminHandler.Writing()
// ---------------------------------------------------------------------------

func TestWriting(t *testing.T) {
	t.Run("JSON response with articles", func(t *testing.T) {
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

		svc := &WritingArticleService{
			Articles: []*models.Article{
				{Slug: "first-post", Title: "First Post", Draft: false, Date: time.Now()},
				{Slug: "second-post", Title: "Second Post", Draft: false, Date: time.Now()},
				{Slug: "third-post", Title: "Third Post", Draft: false, Date: time.Now()},
			},
		}
		handler := NewAdminHandler(base, svc, time.Now())

		router := gin.New()
		router.GET("/admin/writing", handler.Writing)

		req := httptest.NewRequest("GET", "/admin/writing", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, float64(3), resp["article_count"])

		articles, ok := resp["articles"].([]any)
		require.True(t, ok)
		assert.Len(t, articles, 3)

		// Verify list view structure (ToListView fields)
		first, ok := articles[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "first-post", first["slug"])
		assert.Equal(t, "First Post", first["title"])
	})

	t.Run("JSON response with no articles", func(t *testing.T) {
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

		svc := &WritingArticleService{Articles: []*models.Article{}}
		handler := NewAdminHandler(base, svc, time.Now())

		router := gin.New()
		router.GET("/admin/writing", handler.Writing)

		req := httptest.NewRequest("GET", "/admin/writing", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, float64(0), resp["article_count"])

		articles, ok := resp["articles"].([]any)
		require.True(t, ok)
		assert.Empty(t, articles)
	})
}

// ---------------------------------------------------------------------------
// T6: AdminHandler.AdminHome()
// ---------------------------------------------------------------------------

func TestAdminHome(t *testing.T) {
	t.Run("JSON response with stats", func(t *testing.T) {
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

		svc := &AdminHomeArticleService{
			Published: []*models.Article{
				{Slug: "post-1", Title: "Post One", Draft: false},
				{Slug: "post-2", Title: "Post Two", Draft: false},
			},
			Drafts: []*models.Article{
				{Slug: "draft-1", Title: "Draft One", Draft: true},
			},
			Tags: []models.TagCount{
				{Tag: "golang", Count: 2},
				{Tag: "web", Count: 1},
			},
			Categories: []models.CategoryCount{
				{Category: "Programming", Count: 2},
			},
		}
		handler := NewAdminHandler(base, svc, time.Now())

		router := gin.New()
		router.GET("/admin/home", handler.AdminHome)

		req := httptest.NewRequest("GET", "/admin/home", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		// Verify stats map
		stats, ok := resp["stats"].(map[string]any)
		require.True(t, ok, "response should contain stats map")
		assert.Equal(t, float64(2), stats["published"])
		assert.Equal(t, float64(1), stats["drafts"])
		assert.Equal(t, float64(2), stats["tags"])
		assert.Equal(t, float64(1), stats["categories"])

		// Verify system map exists
		system, ok := resp["system"].(map[string]any)
		require.True(t, ok, "response should contain system map")
		assert.NotEmpty(t, system["uptime"])
		assert.NotEmpty(t, system["memory"])
		assert.NotEmpty(t, system["go_version"])
		assert.Equal(t, "test", system["environment"])

		// Verify timestamp exists
		assert.NotNil(t, resp["timestamp"])
	})

	t.Run("JSON response with empty data", func(t *testing.T) {
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

		svc := &AdminHomeArticleService{
			Published:  []*models.Article{},
			Drafts:     []*models.Article{},
			Tags:       []models.TagCount{},
			Categories: []models.CategoryCount{},
		}
		handler := NewAdminHandler(base, svc, time.Now())

		router := gin.New()
		router.GET("/admin/home", handler.AdminHome)

		req := httptest.NewRequest("GET", "/admin/home", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		stats, ok := resp["stats"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, float64(0), stats["published"])
		assert.Equal(t, float64(0), stats["drafts"])
		assert.Equal(t, float64(0), stats["tags"])
		assert.Equal(t, float64(0), stats["categories"])
	})
}

// ---------------------------------------------------------------------------
// T7: formatDuration (pure function)
// ---------------------------------------------------------------------------

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "0.00s"},
		{"sub-second", 500 * time.Millisecond, "0.50s"},
		{"30 seconds", 30 * time.Second, "30.00s"},
		{"59 seconds", 59 * time.Second, "59.00s"},
		{"90 seconds", 90 * time.Second, "1m 30.00s"},
		{"5 minutes", 5 * time.Minute, "5m 0.00s"},
		{"1 hour", time.Hour, "1h 0m 0s"},
		{"1 hour 30 minutes", time.Hour + 30*time.Minute, "1h 30m 0s"},
		{"2 hours 15 minutes 45 seconds", 2*time.Hour + 15*time.Minute + 45*time.Second, "2h 15m 45s"},
		{"1 day", 24 * time.Hour, "1d 0h 0m 0s"},
		{"1 day 2 hours", 26 * time.Hour, "1d 2h 0m 0s"},
		{"3 days 5 hours 30 minutes", 3*24*time.Hour + 5*time.Hour + 30*time.Minute, "3d 5h 30m 0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}
