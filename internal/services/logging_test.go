package services

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/1mb-dev/markgo/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoggingService(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.LoggingConfig
		expectError bool
	}{
		{
			name: "Valid JSON config",
			cfg: config.LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
				Compress:   true,
				AddSource:  false,
			},
			expectError: false,
		},
		{
			name: "Valid text config",
			cfg: config.LoggingConfig{
				Level:      "debug",
				Format:     "text",
				Output:     "stderr",
				MaxSize:    50,
				MaxBackups: 5,
				MaxAge:     7,
				Compress:   false,
				AddSource:  true,
			},
			expectError: false,
		},
		{
			name: "Invalid level",
			cfg: config.LoggingConfig{
				Level:   "invalid",
				Format:  "json",
				Output:  "stdout",
				MaxSize: 100,
			},
			expectError: true,
		},
		{
			name: "Invalid format",
			cfg: config.LoggingConfig{
				Level:   "info",
				Format:  "xml",
				Output:  "stdout",
				MaxSize: 100,
			},
			expectError: true,
		},
		{
			name: "Invalid output",
			cfg: config.LoggingConfig{
				Level:   "info",
				Format:  "json",
				Output:  "invalid",
				MaxSize: 100,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewLoggingService(&tt.cfg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.NotNil(t, service.GetLogger())
			}
		})
	}
}

func TestLoggingServiceFileOutput(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir, err := os.MkdirTemp("", "logging-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	logFile := filepath.Join(tmpDir, "test.log")

	cfg := config.LoggingConfig{
		Level:      "info",
		Format:     "json",
		Output:     "file",
		File:       logFile,
		MaxSize:    1, // Small size to test rotation
		MaxBackups: 2,
		MaxAge:     1,
		Compress:   false,
		AddSource:  false,
	}

	service, err := NewLoggingService(&cfg)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Test logging to file
	logger := service.GetLogger()
	logger.Info("Test message", "key", "value")

	// Verify file was created
	assert.FileExists(t, logFile)

	// Read and verify log content
	content, err := os.ReadFile(logFile) // #nosec G304 - Safe: test file reading in temp directory
	require.NoError(t, err)
	assert.Contains(t, string(content), "Test message")
	assert.Contains(t, string(content), "key")
	assert.Contains(t, string(content), "value")
}

func TestLoggingServiceLogLevels(t *testing.T) {
	tests := []struct {
		name         string
		configLevel  string
		logLevel     string
		shouldAppear bool
	}{
		{"Debug config, debug log", "debug", "debug", true},
		{"Debug config, info log", "debug", "info", true},
		{"Debug config, error log", "debug", "error", true},
		{"Info config, debug log", "info", "debug", false},
		{"Info config, info log", "info", "info", true},
		{"Info config, error log", "info", "error", true},
		{"Error config, debug log", "error", "debug", false},
		{"Error config, info log", "error", "info", false},
		{"Error config, error log", "error", "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer

			// Create custom logger that writes to our buffer
			level, _ := parseLogLevel(tt.configLevel)
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: level})
			logger := slog.New(handler)

			// Log at different levels
			switch tt.logLevel {
			case "debug":
				logger.Debug("test message")
			case "info":
				logger.Info("test message")
			case "warn":
				logger.Warn("test message")
			case "error":
				logger.Error("test message")
			}

			output := buf.String()
			if tt.shouldAppear {
				assert.Contains(t, output, "test message")
			} else {
				assert.NotContains(t, output, "test message")
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
		hasError bool
	}{
		{"debug", slog.LevelDebug, false},
		{"info", slog.LevelInfo, false},
		{"warn", slog.LevelWarn, false},
		{"warning", slog.LevelWarn, false},
		{"error", slog.LevelError, false},
		{"DEBUG", slog.LevelDebug, false},
		{"INFO", slog.LevelInfo, false},
		{"invalid", slog.LevelInfo, true},
		{"", slog.LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("level_%s", tt.input), func(t *testing.T) {
			level, err := parseLogLevel(tt.input)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, level)
			}
		})
	}
}

func TestLoggingConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.LoggingConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid config",
			cfg: config.LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
			},
			expectError: false,
		},
		{
			name: "Invalid level",
			cfg: config.LoggingConfig{
				Level:   "invalid",
				Format:  "json",
				Output:  "stdout",
				MaxSize: 100,
			},
			expectError: true,
			errorMsg:    "Log level must be one of",
		},
		{
			name: "Invalid format",
			cfg: config.LoggingConfig{
				Level:   "info",
				Format:  "xml",
				Output:  "stdout",
				MaxSize: 100,
			},
			expectError: true,
			errorMsg:    "Log format must be one of",
		},
		{
			name: "File output without file path",
			cfg: config.LoggingConfig{
				Level:   "info",
				Format:  "json",
				Output:  "file",
				File:    "",
				MaxSize: 100,
			},
			expectError: true,
			errorMsg:    "Log file path is required",
		},
		{
			name: "Invalid max size",
			cfg: config.LoggingConfig{
				Level:   "info",
				Format:  "json",
				Output:  "stdout",
				MaxSize: -1,
			},
			expectError: true,
			errorMsg:    "Log max size must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoggingServiceFormats(t *testing.T) {
	formats := []string{"json", "text"}

	for _, format := range formats {
		t.Run(fmt.Sprintf("format_%s", format), func(t *testing.T) {
			var buf bytes.Buffer

			// Create handler directly for testing
			var handler slog.Handler
			handlerOpts := &slog.HandlerOptions{Level: slog.LevelInfo}

			switch format {
			case "json":
				handler = slog.NewJSONHandler(&buf, handlerOpts)
			case "text":
				handler = slog.NewTextHandler(&buf, handlerOpts)
			}

			logger := slog.New(handler)
			logger.Info("test message", "key", "value")

			output := buf.String()
			assert.Contains(t, output, "test message")

			if format == "json" {
				// JSON format should contain structured data
				assert.Contains(t, output, `"msg":"test message"`)
				assert.Contains(t, output, `"key":"value"`)
			} else {
				// Text format should be human readable
				assert.Contains(t, output, "key=value")
			}
		})
	}
}
