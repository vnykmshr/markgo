package services

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LoggingService provides enhanced logging functionality with rotation and formatting
type LoggingService struct {
	logger *slog.Logger
	config config.LoggingConfig
}

// NewLoggingService creates a new logging service with the given configuration
func NewLoggingService(cfg config.LoggingConfig) (*LoggingService, error) {
	logger, err := createLogger(cfg)
	if err != nil {
		return nil, apperrors.NewConfigError("logging", cfg, "Failed to create logger", err)
	}

	return &LoggingService{
		logger: logger,
		config: cfg,
	}, nil
}

// GetLogger returns the configured slog.Logger instance
func (ls *LoggingService) GetLogger() *slog.Logger {
	return ls.logger
}

// createLogger creates a configured slog.Logger based on the logging configuration
func createLogger(cfg config.LoggingConfig) (*slog.Logger, error) {
	// Parse log level
	level, err := parseLogLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// Get output writer
	writer, err := getLogWriter(cfg)
	if err != nil {
		return nil, err
	}

	// Create handler options
	handlerOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	// Create appropriate handler based on format
	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(writer, handlerOpts)
	case "text":
		handler = slog.NewTextHandler(writer, handlerOpts)
	default:
		return nil, apperrors.NewConfigError("format", cfg.Format, "Unsupported log format", apperrors.ErrConfigValidation)
	}

	return slog.New(handler), nil
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(levelStr string) (slog.Level, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, apperrors.NewConfigError("level", levelStr, "Invalid log level", apperrors.ErrConfigValidation)
	}
}

// getLogWriter returns the appropriate io.Writer for the log output configuration
func getLogWriter(cfg config.LoggingConfig) (io.Writer, error) {
	switch cfg.Output {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "file":
		if cfg.File == "" {
			return nil, apperrors.NewConfigError("file", cfg.File, "Log file path is required", apperrors.ErrMissingConfig)
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(cfg.File)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, apperrors.NewConfigError("file", cfg.File, "Failed to create log directory", err)
		}

		// Create lumberjack logger for rotation
		return &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize, // megabytes
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge, // days
			Compress:   cfg.Compress,
		}, nil
	default:
		return nil, apperrors.NewConfigError("output", cfg.Output, "Invalid log output", apperrors.ErrConfigValidation)
	}
}

// WithContext creates a logger with additional context fields
func (ls *LoggingService) WithContext(keyvals ...interface{}) *slog.Logger {
	return ls.logger.With(keyvals...)
}

// Debug logs a debug message
func (ls *LoggingService) Debug(msg string, keyvals ...interface{}) {
	ls.logger.Debug(msg, keyvals...)
}

// Info logs an info message
func (ls *LoggingService) Info(msg string, keyvals ...interface{}) {
	ls.logger.Info(msg, keyvals...)
}

// Warn logs a warning message
func (ls *LoggingService) Warn(msg string, keyvals ...interface{}) {
	ls.logger.Warn(msg, keyvals...)
}

// Error logs an error message
func (ls *LoggingService) Error(msg string, keyvals ...interface{}) {
	ls.logger.Error(msg, keyvals...)
}

// Close closes any resources used by the logging service
func (ls *LoggingService) Close() error {
	// If using file output with lumberjack, we might want to close it
	// For now, we don't need explicit cleanup
	return nil
}
