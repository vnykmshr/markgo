package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/models"
	"github.com/vnykmshr/markgo/internal/services/compose"
)

// ---------------------------------------------------------------------------
// Test helpers & mocks
// ---------------------------------------------------------------------------

type MockMarkdownRenderer struct {
	Result string
	Err    error
}

func (m *MockMarkdownRenderer) ProcessMarkdown(content string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if m.Result != "" {
		return m.Result, nil
	}
	return "<p>" + content + "</p>", nil
}

func createTestComposeHandler(t *testing.T, renderer MarkdownRenderer) (*ComposeHandler, string) {
	t.Helper()
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Environment: "test",
		BaseURL:     "http://localhost:3000",
		StaticPath:  tmpDir,
		Blog: config.BlogConfig{
			Title: "Test Blog",
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

	if renderer == nil {
		renderer = &MockMarkdownRenderer{}
	}

	handler := NewComposeHandler(base, nil, &MockArticleService{}, renderer)
	return handler, tmpDir
}

// createMultipartRequest builds a multipart/form-data request with a file field.
func createMultipartRequest(t *testing.T, fieldName, fileName string, fileContent []byte, extraFields map[string]string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, fileName)
	require.NoError(t, err)
	_, err = part.Write(fileContent)
	require.NoError(t, err)

	for k, v := range extraFields {
		require.NoError(t, writer.WriteField(k, v))
	}
	require.NoError(t, writer.Close())

	req := httptest.NewRequest("POST", "/compose/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// Minimal 1x1 JPEG (valid JFIF header).
var minimalJPEG = []byte{
	0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
	0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
	0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
	0x09, 0x08, 0x0A, 0x0C, 0x14, 0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12,
	0x13, 0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A, 0x1C, 0x1C, 0x20,
	0x24, 0x2E, 0x27, 0x20, 0x22, 0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29,
	0x2C, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39, 0x3D, 0x38, 0x32,
	0x3C, 0x2E, 0x33, 0x34, 0x32, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x01,
	0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xC4, 0x00, 0x1F, 0x00, 0x00,
	0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	0x09, 0x0A, 0x0B, 0xFF, 0xC4, 0x00, 0xB5, 0x10, 0x00, 0x02, 0x01, 0x03,
	0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7D,
	0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06,
	0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xA1, 0x08,
	0x23, 0x42, 0xB1, 0xC1, 0x15, 0x52, 0xD1, 0xF0, 0x24, 0x33, 0x62, 0x72,
	0x82, 0x09, 0x0A, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x25, 0x26, 0x27, 0x28,
	0x29, 0x2A, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x43, 0x44, 0x45,
	0x46, 0x47, 0x48, 0x49, 0x4A, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
	0x5A, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6A, 0x73, 0x74, 0x75,
	0x76, 0x77, 0x78, 0x79, 0x7A, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
	0x8A, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0xA2, 0xA3,
	0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6,
	0xB7, 0xB8, 0xB9, 0xBA, 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9,
	0xCA, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xE1, 0xE2,
	0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xF1, 0xF2, 0xF3, 0xF4,
	0xF5, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01,
	0x00, 0x00, 0x3F, 0x00, 0x7B, 0x94, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xD9,
}

// ---------------------------------------------------------------------------
// sanitizeFilename tests (pure function)
// ---------------------------------------------------------------------------

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple name", "photo.jpg", "photo"},
		{"uppercase", "MyPhoto.PNG", "myphoto"},
		{"spaces and special chars", "my photo (1).jpg", "myphoto1"},
		{"unicode only", "фото.jpg", "image"},
		{"empty string", "", "image"},
		{"extension only", ".jpg", "image"},
		{"path traversal", "../../etc/passwd", "passwd"},
		{"deep path traversal", "../../../secret.jpg", "secret"},
		{"hyphens and underscores", "my-photo_2024.webp", "my-photo_2024"},
		{"long name", strings.Repeat("a", 100) + ".jpg", strings.Repeat("a", 50)},
		{"dots in name", "file.name.with.dots.jpg", "filenamewithdots"}, // filepath.Ext gets ".jpg", rest keeps dots→stripped
		{"mixed safe and unsafe", "Hello World! @#$.png", "helloworld"}, // special chars stripped
		{"numbers only", "12345.gif", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			assert.Equal(t, tt.expect, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Preview handler tests
// ---------------------------------------------------------------------------

func TestPreview(t *testing.T) {
	t.Run("valid markdown returns HTML", func(t *testing.T) {
		handler, _ := createTestComposeHandler(t, &MockMarkdownRenderer{
			Result: "<h1>Hello</h1>",
		})

		router := gin.New()
		router.POST("/compose/preview", handler.Preview)

		form := url.Values{"content": {"# Hello"}}
		req := httptest.NewRequest("POST", "/compose/preview", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
		assert.Equal(t, "<h1>Hello</h1>", w.Body.String())
	})

	t.Run("empty content returns 200 with empty body", func(t *testing.T) {
		handler, _ := createTestComposeHandler(t, nil)

		router := gin.New()
		router.POST("/compose/preview", handler.Preview)

		form := url.Values{"content": {""}}
		req := httptest.NewRequest("POST", "/compose/preview", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("oversized body rejected by MaxBytesReader", func(t *testing.T) {
		handler, _ := createTestComposeHandler(t, nil)

		router := gin.New()
		router.POST("/compose/preview", handler.Preview)

		largeContent := strings.Repeat("x", 2<<20) // 2MB, exceeds 1MB limit
		form := url.Values{"content": {largeContent}}
		req := httptest.NewRequest("POST", "/compose/preview", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// MaxBytesReader truncates the body; Gin's PostForm returns empty string
		// because form parsing fails on the truncated body → handler sees empty content → 200
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("renderer error returns 500", func(t *testing.T) {
		handler, _ := createTestComposeHandler(t, &MockMarkdownRenderer{
			Err: fmt.Errorf("render failed"),
		})

		router := gin.New()
		router.POST("/compose/preview", handler.Preview)

		form := url.Values{"content": {"# broken"}}
		req := httptest.NewRequest("POST", "/compose/preview", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Preview unavailable")
	})
}

// ---------------------------------------------------------------------------
// Upload handler tests
// ---------------------------------------------------------------------------

func TestUpload(t *testing.T) {
	t.Run("valid JPEG upload", func(t *testing.T) {
		handler, tmpDir := createTestComposeHandler(t, nil)

		req := createMultipartRequest(t, "file", "test-photo.jpg", minimalJPEG, nil)
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/upload", handler.Upload)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["url"], "/static/images/uploads/")
		assert.Contains(t, resp["url"], ".jpg")
		assert.Contains(t, resp["markdown"], "![test-photo]")
		assert.NotEmpty(t, resp["filename"])

		// Verify file exists on disk with correct permissions
		uploadDir := filepath.Join(tmpDir, "images", "uploads")
		files, err := os.ReadDir(uploadDir)
		require.NoError(t, err)
		assert.Len(t, files, 1)

		info, err := os.Stat(filepath.Join(uploadDir, files[0].Name()))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
	})

	t.Run("no file provided", func(t *testing.T) {
		handler, _ := createTestComposeHandler(t, nil)

		form := url.Values{"other": {"data"}}
		req := httptest.NewRequest("POST", "/compose/upload", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/upload", handler.Upload)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["error"], "No file")
	})

	t.Run("disallowed content type", func(t *testing.T) {
		handler, _ := createTestComposeHandler(t, nil)

		// HTML file content — http.DetectContentType will identify as text/html
		htmlContent := []byte("<html><body>hello</body></html>")
		req := createMultipartRequest(t, "file", "sneaky.jpg", htmlContent, nil)
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/upload", handler.Upload)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["error"], "File type not allowed")
	})

	t.Run("path traversal filename is sanitized", func(t *testing.T) {
		handler, _ := createTestComposeHandler(t, nil)

		req := createMultipartRequest(t, "file", "../../etc/passwd.jpg", minimalJPEG, nil)
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/upload", handler.Upload)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		// Filename should be sanitized — no path components
		assert.NotContains(t, resp["filename"], "..")
		assert.NotContains(t, resp["filename"], "/")
		assert.Contains(t, resp["markdown"], "![passwd]")
	})

	t.Run("empty file detected as unsupported type", func(t *testing.T) {
		handler, _ := createTestComposeHandler(t, nil)

		req := createMultipartRequest(t, "file", "empty.jpg", []byte{}, nil)
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/upload", handler.Upload)
		router.ServeHTTP(w, req)

		// Empty file: Read returns 0 bytes + io.EOF. DetectContentType on empty buf
		// returns "text/plain" which is not in allowedImageTypes → 400.
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unicode filename falls back to image", func(t *testing.T) {
		handler, _ := createTestComposeHandler(t, nil)

		req := createMultipartRequest(t, "file", "фото.jpg", minimalJPEG, nil)
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/upload", handler.Upload)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["markdown"], "![image]")
	})
}

// ---------------------------------------------------------------------------
// Quick publish handler tests
// ---------------------------------------------------------------------------

func createQuickPublishHandler(t *testing.T) *ComposeHandler {
	t.Helper()
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Environment:  "test",
		BaseURL:      "http://localhost:3000",
		ArticlesPath: tmpDir,
		Blog:         config.BlogConfig{Title: "Test Blog", Author: "Test Author"},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

	composeSvc := compose.NewService(tmpDir, cfg.Blog.Author)
	return NewComposeHandler(base, composeSvc, &MockArticleService{}, &MockMarkdownRenderer{})
}

func TestQuickPublish(t *testing.T) {
	t.Run("thought — content only, no title", func(t *testing.T) {
		handler := createQuickPublishHandler(t)

		body := `{"content":"Just a quick thought about Go templates."}`
		req := httptest.NewRequest("POST", "/compose/quick", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/quick", handler.HandleQuickPublish)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp["slug"])
		assert.Contains(t, resp["url"], "/writing/")
		assert.Equal(t, "thought", resp["type"])
		assert.Equal(t, "Published", resp["message"])
	})

	t.Run("article — has title", func(t *testing.T) {
		handler := createQuickPublishHandler(t)

		body := `{"content":"Full article content here.","title":"My Article"}`
		req := httptest.NewRequest("POST", "/compose/quick", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/quick", handler.HandleQuickPublish)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "my-article", resp["slug"])
		assert.Equal(t, "/writing/my-article", resp["url"])
		assert.Equal(t, "article", resp["type"])
	})

	t.Run("link — has link_url", func(t *testing.T) {
		handler := createQuickPublishHandler(t)

		body := `{"content":"Check this out","link_url":"https://example.com"}`
		req := httptest.NewRequest("POST", "/compose/quick", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/quick", handler.HandleQuickPublish)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "link", resp["type"])
	})

	t.Run("long content without title is article not thought", func(t *testing.T) {
		handler := createQuickPublishHandler(t)

		// 150 words without a title — should be "article", not "thought"
		words := strings.Repeat("word ", 150)
		body := fmt.Sprintf(`{"content":%q}`, words)
		req := httptest.NewRequest("POST", "/compose/quick", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/quick", handler.HandleQuickPublish)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "article", resp["type"])
	})

	t.Run("empty content returns 400", func(t *testing.T) {
		handler := createQuickPublishHandler(t)

		body := `{"content":""}`
		req := httptest.NewRequest("POST", "/compose/quick", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/quick", handler.HandleQuickPublish)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["error"], "Content is required")
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		handler := createQuickPublishHandler(t)

		req := httptest.NewRequest("POST", "/compose/quick", strings.NewReader("not json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/compose/quick", handler.HandleQuickPublish)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["error"], "Invalid request body")
	})
}

// ---------------------------------------------------------------------------
// Drafts handler tests
// ---------------------------------------------------------------------------

// DraftsArticleService returns canned draft data.
type DraftsArticleService struct {
	MockArticleService
	Drafts []*models.Article
}

func (m *DraftsArticleService) GetDraftArticles() []*models.Article { return m.Drafts }

func TestDrafts(t *testing.T) {
	t.Run("JSON response with drafts", func(t *testing.T) {
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

		svc := &DraftsArticleService{
			Drafts: []*models.Article{
				{Slug: "draft-1", Title: "Draft One", Draft: true, Date: time.Now()},
				{Slug: "draft-2", Title: "Draft Two", Draft: true, Date: time.Now()},
			},
		}
		handler := NewAdminHandler(base, svc, time.Now())

		router := gin.New()
		router.GET("/admin/drafts", handler.Drafts)

		req := httptest.NewRequest("GET", "/admin/drafts", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, float64(2), resp["draft_count"])

		drafts, ok := resp["drafts"].([]any)
		require.True(t, ok)
		assert.Len(t, drafts, 2)
	})

	t.Run("JSON response with no drafts", func(t *testing.T) {
		cfg := createTestConfig()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		base := NewBaseHandler(cfg, logger, &MockTemplateService{}, &BuildInfo{Version: "test"}, &MockSEOService{})

		svc := &DraftsArticleService{Drafts: []*models.Article{}}
		handler := NewAdminHandler(base, svc, time.Now())

		router := gin.New()
		router.GET("/admin/drafts", handler.Drafts)

		req := httptest.NewRequest("GET", "/admin/drafts", http.NoBody)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, float64(0), resp["draft_count"])
	})
}
