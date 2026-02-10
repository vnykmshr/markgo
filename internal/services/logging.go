package services

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/constants"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
)

// Log format constants
const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

// Log level constants
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// LoggingService provides enhanced logging functionality with rotation and formatting
type LoggingService struct {
	logger *slog.Logger
}

// NewLoggingService creates a new logging service with the given configuration
func NewLoggingService(cfg *config.LoggingConfig) (*LoggingService, error) {
	logger, err := createLogger(cfg)
	if err != nil {
		return nil, apperrors.NewConfigError("logging", *cfg, "Failed to create logger", err)
	}

	// Create base attributes that will be included in all log entries
	baseAttrs := []slog.Attr{
		slog.String("service", "markgo"),
		slog.String("version", constants.AppVersion),
		slog.String("hostname", getHostname()),
		slog.Int("pid", os.Getpid()),
	}

	// Add source information in development
	if cfg.AddSource {
		baseAttrs = append(baseAttrs, slog.String("go_version", runtime.Version()))
	}

	// Convert attributes to interface{} slice for logger.With
	baseArgs := make([]interface{}, 0, len(baseAttrs)*2)
	for _, attr := range baseAttrs {
		baseArgs = append(baseArgs, attr.Key, attr.Value)
	}

	return &LoggingService{
		logger: logger.With(baseArgs...),
	}, nil
}

// GetLogger returns the configured slog.Logger instance
func (ls *LoggingService) GetLogger() *slog.Logger {
	return ls.logger
}

// createLogger creates a configured slog.Logger based on the logging configuration
func createLogger(cfg *config.LoggingConfig) (*slog.Logger, error) {
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
	case LogFormatJSON:
		handler = slog.NewJSONHandler(writer, handlerOpts)
	case LogFormatText:
		handler = slog.NewTextHandler(writer, handlerOpts)
	default:
		return nil, apperrors.NewConfigError("format", cfg.Format, "Unsupported log format", apperrors.ErrConfigValidation)
	}

	return slog.New(handler), nil
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(levelStr string) (slog.Level, error) {
	switch strings.ToLower(levelStr) {
	case LogLevelDebug:
		return slog.LevelDebug, nil
	case LogLevelInfo:
		return slog.LevelInfo, nil
	case LogLevelWarn, "warning":
		return slog.LevelWarn, nil
	case LogLevelError:
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, apperrors.NewConfigError("level", levelStr, "Invalid log level", apperrors.ErrConfigValidation)
	}
}

// getLogWriter returns the appropriate io.Writer for the log output configuration
func getLogWriter(cfg *config.LoggingConfig) (io.Writer, error) {
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
		if err := os.MkdirAll(dir, 0o750); err != nil {
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

// getHostname returns the system hostname, fallback to 'unknown'
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
