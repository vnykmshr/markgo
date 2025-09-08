package services

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/vnykmshr/markgo/internal/config"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LoggingService provides enhanced logging functionality with rotation and formatting
type LoggingService struct {
	logger     *slog.Logger
	config     config.LoggingConfig
	baseAttrs  []slog.Attr
	serviceCtx context.Context
}

// NewLoggingService creates a new logging service with the given configuration
func NewLoggingService(cfg config.LoggingConfig) (*LoggingService, error) {
	logger, err := createLogger(cfg)
	if err != nil {
		return nil, apperrors.NewConfigError("logging", cfg, "Failed to create logger", err)
	}

	// Create base attributes that will be included in all log entries
	baseAttrs := []slog.Attr{
		slog.String("service", "markgo"),
		slog.String("version", "1.0.0"),
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
		logger:     logger.With(baseArgs...),
		config:     cfg,
		baseAttrs:  baseAttrs,
		serviceCtx: context.Background(),
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

// WithRequestContext creates a logger with request-specific context
func (ls *LoggingService) WithRequestContext(ctx context.Context, entry LogEntry) *slog.Logger {
	args := make([]interface{}, 0, 16)

	if entry.RequestID != "" {
		args = append(args, "request_id", entry.RequestID)
	}
	if entry.UserID != "" {
		args = append(args, "user_id", entry.UserID)
	}
	if entry.IP != "" {
		args = append(args, "ip", entry.IP)
	}
	if entry.Path != "" {
		args = append(args, "path", entry.Path)
	}
	if entry.Method != "" {
		args = append(args, "method", entry.Method)
	}
	if entry.Component != "" {
		args = append(args, "component", entry.Component)
	}
	if entry.Action != "" {
		args = append(args, "action", entry.Action)
	}

	return ls.logger.With(args...)
}

// WithComponent creates a logger with component-specific context
func (ls *LoggingService) WithComponent(component string) *slog.Logger {
	return ls.logger.With("component", component)
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

// LogPerformance logs performance metrics with structured data
func (ls *LoggingService) LogPerformance(perfLog PerformanceLog) {
	ls.logger.Info("Performance metrics",
		slog.String("operation", perfLog.Operation),
		slog.Duration("duration", perfLog.Duration),
		slog.Int64("memory_before_mb", perfLog.MemoryBefore/(1024*1024)),
		slog.Int64("memory_after_mb", perfLog.MemoryAfter/(1024*1024)),
		slog.Int64("memory_delta_mb", (perfLog.MemoryAfter-perfLog.MemoryBefore)/(1024*1024)),
		slog.Int("goroutines", perfLog.Goroutines),
		slog.Uint64("allocations", perfLog.Allocations),
		slog.String("category", "performance"),
	)
}

// LogSecurity logs security-related events with structured data
func (ls *LoggingService) LogSecurity(secLog SecurityLog) {
	level := slog.LevelWarn
	switch secLog.Severity {
	case "critical":
		level = slog.LevelError
	case "high":
		level = slog.LevelError
	case "medium":
		level = slog.LevelWarn
	case "low":
		level = slog.LevelInfo
	}

	ls.logger.Log(context.Background(), level, "Security event",
		slog.String("event", secLog.Event),
		slog.String("severity", secLog.Severity),
		slog.String("ip", secLog.IP),
		slog.String("user_agent", secLog.UserAgent),
		slog.String("path", secLog.Path),
		slog.String("description", secLog.Description),
		slog.String("category", "security"),
	)
}

// LogHTTPRequest logs HTTP request details with structured data
func (ls *LoggingService) LogHTTPRequest(ctx context.Context, entry LogEntry) {
	attrs := []slog.Attr{
		slog.String("method", entry.Method),
		slog.String("path", entry.Path),
		slog.Int("status_code", entry.StatusCode),
		slog.Duration("duration", entry.Duration),
		slog.String("ip", entry.IP),
		slog.String("user_agent", entry.UserAgent),
		slog.String("category", "http"),
	}

	if entry.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", entry.RequestID))
	}

	level := slog.LevelInfo
	if entry.StatusCode >= 400 {
		level = slog.LevelWarn
	}
	if entry.StatusCode >= 500 {
		level = slog.LevelError
	}

	ls.logger.LogAttrs(ctx, level, "HTTP request", attrs...)
}

// LogError logs errors with enhanced context and stack traces
func (ls *LoggingService) LogError(ctx context.Context, err error, msg string, keyvals ...interface{}) {
	attrs := []slog.Attr{
		slog.String("error", err.Error()),
		slog.String("category", "error"),
	}

	// Add stack trace in debug mode
	if ls.config.Level == "debug" {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			attrs = append(attrs,
				slog.String("caller_file", file),
				slog.Int("caller_line", line),
			)
		}
	}

	// Convert keyvals to attributes
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			if key, ok := keyvals[i].(string); ok {
				attrs = append(attrs, slog.Any(key, keyvals[i+1]))
			}
		}
	}

	ls.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// LogSlowOperation logs operations that exceed expected duration
func (ls *LoggingService) LogSlowOperation(ctx context.Context, operation string, duration time.Duration, threshold time.Duration, keyvals ...interface{}) {
	if duration > threshold {
		attrs := []slog.Attr{
			slog.String("operation", operation),
			slog.Duration("duration", duration),
			slog.Duration("threshold", threshold),
			slog.Duration("exceeded_by", duration-threshold),
			slog.String("category", "performance"),
		}

		// Add additional context
		for i := 0; i < len(keyvals); i += 2 {
			if i+1 < len(keyvals) {
				if key, ok := keyvals[i].(string); ok {
					attrs = append(attrs, slog.Any(key, keyvals[i+1]))
				}
			}
		}

		ls.logger.LogAttrs(ctx, slog.LevelWarn, "Slow operation detected", attrs...)
	}
}

// GetMemoryStats returns current memory statistics for logging
func (ls *LoggingService) GetMemoryStats() (int64, int64, uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc), int64(m.Sys), m.Mallocs
}

// Close closes any resources used by the logging service
func (ls *LoggingService) Close() error {
	// If using file output with lumberjack, we might want to close it
	// For now, we don't need explicit cleanup
	return nil
}

// getHostname returns the system hostname, fallback to 'unknown'
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
