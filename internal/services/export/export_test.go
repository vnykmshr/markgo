package export

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T, outputDir string) *StaticExportService {
	t.Helper()
	return &StaticExportService{
		logger:    slog.Default(),
		outputDir: outputDir,
		baseURL:   "https://example.com",
	}
}

func TestCreateOutputDir_FreshDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "output")
	svc := newTestService(t, dir)

	err := svc.createOutputDir()
	require.NoError(t, err)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestCreateOutputDir_ReplacesExisting(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "output")

	// Create existing directory with a file
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "old.html"), []byte("old"), 0o600))

	svc := newTestService(t, dir)
	err := svc.createOutputDir()
	require.NoError(t, err)

	// The output directory should exist but the old file should be gone
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	_, err = os.Stat(filepath.Join(dir, "old.html"))
	assert.True(t, os.IsNotExist(err), "old file should not exist in new directory")

	// Verify .old and .tmp are cleaned up
	_, err = os.Stat(dir + ".old")
	assert.True(t, os.IsNotExist(err), ".old directory should be cleaned up")
	_, err = os.Stat(dir + ".tmp")
	assert.True(t, os.IsNotExist(err), ".tmp directory should be cleaned up")
}

func TestCreateOutputDir_CleansUpLeftoverTmpDirs(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "output")

	// Simulate leftover dirs from a previous failed run
	require.NoError(t, os.MkdirAll(dir+".tmp", 0o750))
	require.NoError(t, os.MkdirAll(dir+".old", 0o750))

	svc := newTestService(t, dir)
	err := svc.createOutputDir()
	require.NoError(t, err)

	_, err = os.Stat(dir)
	require.NoError(t, err)
}

func TestProcessHTML(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		input   string
		want    string
	}{
		{
			name:    "rewrites href",
			baseURL: "https://example.com",
			input:   `<a href="/articles/hello">link</a>`,
			want:    `<a href="https://example.com/articles/hello">link</a>`,
		},
		{
			name:    "rewrites src",
			baseURL: "https://example.com",
			input:   `<img src="/static/logo.png"/>`,
			want:    `<img src="https://example.com/static/logo.png"/>`,
		},
		{
			name:    "rewrites action",
			baseURL: "https://example.com",
			input:   `<form action="/search">`,
			want:    `<form action="https://example.com/search">`,
		},
		{
			name:    "trailing slash stripped from base",
			baseURL: "https://example.com/",
			input:   `<a href="/about">`,
			want:    `<a href="https://example.com/about">`,
		},
		{
			name:    "no base URL leaves content unchanged",
			baseURL: "",
			input:   `<a href="/about">`,
			want:    `<a href="/about">`,
		},
		{
			name:    "does not rewrite absolute URLs",
			baseURL: "https://example.com",
			input:   `<a href="https://other.com/page">`,
			want:    `<a href="https://other.com/page">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &StaticExportService{baseURL: tt.baseURL}
			got := svc.processHTML(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeForURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase", "Technology", "technology"},
		{"spaces to hyphens", "My Category", "my-category"},
		{"special chars removed", "C++ & Rust!", "c-rust"},
		{"slashes to hyphens", "Front/Back", "front-back"},
		{"underscores to hyphens", "my_tag", "my-tag"},
		{"consecutive hyphens collapsed", "a---b", "a-b"},
		{"leading trailing hyphens trimmed", "-hello-", "hello"},
		{"unicode removed", "日本語", ""},
		{"mixed", "Go (1.21)", "go-121"},
	}

	svc := &StaticExportService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.sanitizeForURL(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetHostFromBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{"valid URL", "https://example.com", "example.com"},
		{"with port", "http://localhost:3000", "localhost:3000"},
		{"empty", "", "localhost"},
		{"invalid", "://bad", "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &StaticExportService{baseURL: tt.baseURL}
			got := svc.getHostFromBaseURL()
			assert.Equal(t, tt.want, got)
		})
	}
}
