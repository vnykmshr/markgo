package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 5 << 20 // 5MB

var (
	allowedImageTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	safeFilenameRe = regexp.MustCompile(`[^a-z0-9_-]`)
)

// sanitizeFilename strips the extension, lowercases, removes unsafe chars,
// and caps length. Returns "image" if the result is empty.
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	ext := filepath.Ext(name)
	name = strings.TrimSuffix(name, ext)
	name = strings.ToLower(name)
	name = safeFilenameRe.ReplaceAllString(name, "")
	if name == "" {
		name = "image"
	}
	if len(name) > 50 {
		name = name[:50]
	}
	return name
}

// cleanupTempFile removes a temporary file, logging on failure.
func (h *ComposeHandler) cleanupTempFile(path string) {
	if removeErr := os.Remove(path); removeErr != nil {
		h.logger.Error("Failed to remove temp file", "path", path, "error", removeErr)
	}
}

// Upload handles image upload for the compose page.
// Security: content type validated via http.DetectContentType on actual bytes.
// Extension derived from detected MIME type, not from filename.
func (h *ComposeHandler) Upload(c *gin.Context) {
	// Limit request body before Gin parses the multipart form (defense in depth)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error("Upload FormFile error", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided or file too large"})
		return
	}
	defer file.Close()

	// Read first 512 bytes for content type detection
	buf := make([]byte, 512)
	n, readErr := file.Read(buf)
	if readErr != nil && readErr != io.EOF {
		h.logger.Error("Upload file read error", "error", readErr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file"})
		return
	}

	detectedType := http.DetectContentType(buf[:n])
	ext, ok := allowedImageTypes[detectedType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed. Supported: JPEG, PNG, GIF, WebP"})
		return
	}

	// Seek back to start after content detection
	if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
		h.logger.Error("Upload file seek error", "error", seekErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}

	// Build safe filename
	safeName := sanitizeFilename(header.Filename)
	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixMilli(), safeName, ext)

	// Ensure upload directory exists
	uploadDir := filepath.Join(h.config.StaticPath, "images", "uploads")
	if mkdirErr := os.MkdirAll(uploadDir, 0o755); mkdirErr != nil { //nolint:gosec // upload dir needs to be accessible
		h.logger.Error("Failed to create upload directory", "error", mkdirErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	destPath := filepath.Join(uploadDir, filename)

	// Atomic write: temp file + rename
	tmpFile, err := os.CreateTemp(uploadDir, "upload-*")
	if err != nil {
		h.logger.Error("Failed to create temp file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}
	tmpPath := tmpFile.Name()

	if _, copyErr := io.Copy(tmpFile, file); copyErr != nil {
		_ = tmpFile.Close() //nolint:gosec // best-effort close before cleanup
		h.cleanupTempFile(tmpPath)
		h.logger.Error("Failed to write file", "error", copyErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	if closeErr := tmpFile.Close(); closeErr != nil {
		h.cleanupTempFile(tmpPath)
		h.logger.Error("Failed to close temp file", "error", closeErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	if renameErr := os.Rename(tmpPath, destPath); renameErr != nil {
		h.cleanupTempFile(tmpPath)
		h.logger.Error("Failed to rename uploaded file", "error", renameErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	// Make uploaded file world-readable (os.CreateTemp defaults to 0600)
	if chmodErr := os.Chmod(destPath, 0o644); chmodErr != nil { //nolint:gosec // uploaded images must be readable by web server
		h.logger.Error("Failed to chmod uploaded file", "error", chmodErr)
	}

	url := "/static/images/uploads/" + filename
	markdown := fmt.Sprintf("![%s](%s)", safeName, url)

	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"markdown": markdown,
		"filename": filename,
	})
}
