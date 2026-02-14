package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/1mb-dev/markgo/internal/models"
	"github.com/1mb-dev/markgo/internal/services/compose"
)

// AMAArticleService returns canned drafts for AMA handler tests.
type AMAArticleService struct {
	MockArticleService
	Drafts []*models.Article
}

func (m *AMAArticleService) GetDraftArticles() []*models.Article { return m.Drafts }

func createTestAMAHandler(t *testing.T) (*AMAHandler, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := createTestConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

	composeService := compose.NewService(dir, "Test Author")
	articleService := &AMAArticleService{}

	handler := NewAMAHandler(base, composeService, articleService)
	return handler, dir
}

func TestAMASubmit(t *testing.T) {
	t.Run("valid submission creates file", func(t *testing.T) {
		handler, _ := createTestAMAHandler(t)

		router := gin.New()
		router.POST("/ama/submit", handler.Submit)

		body, _ := json.Marshal(map[string]string{
			"name":     "Alice",
			"email":    "alice@example.com",
			"question": "What is your favorite programming language and why do you prefer it?",
		})
		req := httptest.NewRequest("POST", "/ama/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "success", resp["status"])
	})

	t.Run("honeypot triggers silent success", func(t *testing.T) {
		handler, dir := createTestAMAHandler(t)

		router := gin.New()
		router.POST("/ama/submit", handler.Submit)

		body, _ := json.Marshal(map[string]string{
			"name":     "Bot",
			"question": "This is definitely a real question from a human being, trust me.",
			"website":  "http://spam.example.com",
		})
		req := httptest.NewRequest("POST", "/ama/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Returns success (to not alert the bot) but no file created
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "success", resp["status"])

		// Verify no file was created
		entries, _ := readDir(dir)
		assert.Empty(t, entries)
	})

	t.Run("question too short returns 400", func(t *testing.T) {
		handler, _ := createTestAMAHandler(t)

		router := gin.New()
		router.POST("/ama/submit", handler.Submit)

		body, _ := json.Marshal(map[string]string{
			"name":     "Alice",
			"question": "Too short",
		})
		req := httptest.NewRequest("POST", "/ama/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing required fields returns 400", func(t *testing.T) {
		handler, _ := createTestAMAHandler(t)

		router := gin.New()
		router.POST("/ama/submit", handler.Submit)

		body, _ := json.Marshal(map[string]string{
			"email": "alice@example.com",
		})
		req := httptest.NewRequest("POST", "/ama/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAMAListPending(t *testing.T) {
	t.Run("returns pending AMAs as JSON", func(t *testing.T) {
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

		svc := &AMAArticleService{
			Drafts: []*models.Article{
				{Slug: "ama-1", Type: "ama", Content: "Question 1?", Asker: "Alice", Draft: true, Date: time.Now()},
				{Slug: "ama-2", Type: "ama", Content: "Question 2?", Asker: "Bob", Draft: true, Date: time.Now()},
				{Slug: "regular-draft", Type: "article", Title: "WIP Article", Draft: true, Date: time.Now()},
			},
		}

		composeService := compose.NewService(t.TempDir(), "Test Author")
		handler := NewAMAHandler(base, composeService, svc)

		router := gin.New()
		router.GET("/admin/ama", handler.ListPending)

		req := httptest.NewRequest("GET", "/admin/ama", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		// Only AMA drafts, not the regular article draft
		assert.Equal(t, float64(2), resp["pending_count"])
	})
}

func TestAMAAnswer(t *testing.T) {
	t.Run("publishes answer", func(t *testing.T) {
		dir := t.TempDir()
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

		composeService := compose.NewService(dir, "Test Author")
		articleService := &AMAArticleService{}

		// Create an AMA post first
		slug, err := composeService.CreatePost(&compose.Input{
			Content:    "What is your favorite language?",
			Draft:      true,
			Asker:      "Alice",
			AskerEmail: "alice@example.com",
			Type:       "ama",
		})
		require.NoError(t, err)

		handler := NewAMAHandler(base, composeService, articleService)

		router := gin.New()
		router.POST("/admin/ama/:slug/answer", handler.Answer)

		body, _ := json.Marshal(map[string]string{
			"answer": "Go, because it's simple and powerful.",
		})
		req := httptest.NewRequest("POST", "/admin/ama/"+slug+"/answer", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "success", resp["status"])

		// Verify the article was updated (draft=false, answer appended)
		input, err := composeService.LoadArticle(slug)
		require.NoError(t, err)
		assert.False(t, input.Draft)
		assert.Contains(t, input.Content, "Go, because it's simple and powerful.")
	})
}

func TestAMAAnswer_NotFound(t *testing.T) {
	handler, _ := createTestAMAHandler(t)

	router := gin.New()
	router.POST("/admin/ama/:slug/answer", handler.Answer)

	body, _ := json.Marshal(map[string]string{
		"answer": "This question does not exist.",
	})
	req := httptest.NewRequest("POST", "/admin/ama/nonexistent-slug/answer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
}

func TestAMAAnswer_InvalidSlug(t *testing.T) {
	handler, _ := createTestAMAHandler(t)

	router := gin.New()
	router.POST("/admin/ama/:slug/answer", handler.Answer)

	body, _ := json.Marshal(map[string]string{
		"answer": "My answer.",
	})
	req := httptest.NewRequest("POST", "/admin/ama/INVALID-UPPERCASE/answer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAMADelete_NotFound(t *testing.T) {
	handler, _ := createTestAMAHandler(t)

	router := gin.New()
	router.POST("/admin/ama/:slug/delete", handler.Delete)

	req := httptest.NewRequest("POST", "/admin/ama/nonexistent-slug/delete", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
}

func TestAMAListPending_Empty(t *testing.T) {
	cfg := createTestConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

	svc := &AMAArticleService{
		Drafts: []*models.Article{},
	}

	composeService := compose.NewService(t.TempDir(), "Test Author")
	handler := NewAMAHandler(base, composeService, svc)

	router := gin.New()
	router.GET("/admin/ama", handler.ListPending)

	req := httptest.NewRequest("GET", "/admin/ama", http.NoBody)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["pending_count"])
}

func TestAMADelete(t *testing.T) {
	t.Run("removes file", func(t *testing.T) {
		dir := t.TempDir()
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

		composeService := compose.NewService(dir, "Test Author")
		articleService := &AMAArticleService{}

		slug, err := composeService.CreatePost(&compose.Input{
			Content: "What is your favorite color?",
			Draft:   true,
			Asker:   "Bob",
			Type:    "ama",
		})
		require.NoError(t, err)

		handler := NewAMAHandler(base, composeService, articleService)

		router := gin.New()
		router.POST("/admin/ama/:slug/delete", handler.Delete)

		req := httptest.NewRequest("POST", "/admin/ama/"+slug+"/delete", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify file is gone
		entries, _ := readDir(dir)
		assert.Empty(t, entries)
	})
}

// readDir reads directory entries, filtering out non-markdown files.
func readDir(dir string) ([]string, error) {
	entries, err := dirEntries(dir)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, e := range entries {
		if !e.IsDir() {
			result = append(result, e.Name())
		}
	}
	return result, nil
}

var dirEntries = os.ReadDir
