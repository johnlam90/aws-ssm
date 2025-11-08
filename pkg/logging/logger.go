package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents log levels
type Level int

const (
	// LevelDebug represents debug log level
	LevelDebug Level = iota
	// LevelInfo represents info log level
	LevelInfo
	// LevelWarn represents warning log level
	LevelWarn
	// LevelError represents error log level
	LevelError
	// LevelFatal represents fatal log level
	LevelFatal
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a time.Duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Time creates a time field
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Any creates a field for any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Logger is a structured logger interface
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
	SetLevel(level Level)
	GetLevel() Level
}

// slogLogger implements Logger using slog
type slogLogger struct {
	logger *slog.Logger
	level  Level
	mu     sync.Mutex
	attrs  []any
}

var _ Logger = (*slogLogger)(nil)

// NewLogger creates a new structured logger
func NewLogger(opts ...Option) Logger {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	level := parseLevel(config.Level)
	slogLevel := convertToSlogLevel(level)

	handler := slog.NewJSONHandler(config.Output, &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Keep our custom timestamp if specified
			if a.Key == "time" && config.TimeFormat != "" {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.String()),
				}
			}
			return a
		},
	})

	return &slogLogger{
		logger: slog.New(handler),
		level:  level,
	}
}

// defaultConfig returns default configuration
func defaultConfig() *Config {
	return &Config{
		Level:       "info",
		Output:      os.Stdout,
		AddSource:   false,
		TimeFormat:  time.RFC3339,
		Environment: "development",
	}
}

// Config represents logger configuration
type Config struct {
	Level       string
	Output      io.Writer
	AddSource   bool
	TimeFormat  string
	Environment string
}

// Option is a logger option
type Option func(*Config)

// WithLevel sets the log level
func WithLevel(level string) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithOutput sets the output writer
func WithOutput(output io.Writer) Option {
	return func(c *Config) {
		c.Output = output
	}
}

// WithAddSource enables adding source location
func WithAddSource() Option {
	return func(c *Config) {
		c.AddSource = true
	}
}

// WithEnvironment sets the environment
func WithEnvironment(env string) Option {
	return func(c *Config) {
		c.Environment = env
	}
}

func (l *slogLogger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields...)
}

func (l *slogLogger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields...)
}

func (l *slogLogger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields...)
}

func (l *slogLogger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

func (l *slogLogger) Fatal(msg string, fields ...Field) {
	l.log(LevelFatal, msg, fields...)
}

func (l *slogLogger) With(fields ...Field) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newAttrs := make([]any, len(l.attrs), len(l.attrs)+len(fields))
	copy(newAttrs, l.attrs)
	for _, field := range fields {
		newAttrs = append(newAttrs, field.Key, field.Value)
	}

	return &slogLogger{
		logger: l.logger,
		level:  l.level,
		attrs:  newAttrs,
	}
}

func (l *slogLogger) WithContext(ctx context.Context) Logger {
	// Extract trace information from context if available
	fields := extractContextFields(ctx)
	return l.With(fields...)
}

func (l *slogLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *slogLogger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

func (l *slogLogger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	// Handle Fatal case separately to avoid defer issues
	if level == LevelFatal {
		l.logFatal(msg, fields...)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Build the complete list of attributes
	allAttrs := make([]any, 0, len(l.attrs)+len(fields)+2)
	allAttrs = append(allAttrs, l.attrs...)

	// Add level and timestamp
	allAttrs = append(allAttrs, "level", level.String(), "timestamp", time.Now().Format(time.RFC3339))

	// Add caller information if needed
	if shouldAddCaller(level) {
		if pc, file, line, ok := runtime.Caller(2); ok {
			function := runtime.FuncForPC(pc).Name()
			funcName := strings.TrimPrefix(function, "github.com/aws-ssm/")
			allAttrs = append(allAttrs, "caller", fmt.Sprintf("%s:%d", file, line), "function", funcName)
		}
	}

	// Add user fields
	for _, field := range fields {
		allAttrs = append(allAttrs, field.Key, field.Value)
	}

	// Log the message
	switch level {
	case LevelError:
		l.logger.Error(msg, allAttrs...)
	case LevelWarn:
		l.logger.Warn(msg, allAttrs...)
	case LevelInfo:
		l.logger.Info(msg, allAttrs...)
	case LevelDebug:
		l.logger.Debug(msg, allAttrs...)
	}
}

func (l *slogLogger) logFatal(msg string, fields ...Field) {
	// Build the complete list of attributes
	allAttrs := make([]any, 0, len(l.attrs)+len(fields)+2)
	allAttrs = append(allAttrs, l.attrs...)

	// Add level and timestamp
	allAttrs = append(allAttrs, "level", LevelFatal.String(), "timestamp", time.Now().Format(time.RFC3339))

	// Add user fields
	for _, field := range fields {
		allAttrs = append(allAttrs, field.Key, field.Value)
	}

	// Log the fatal message
	l.logger.Error(msg, allAttrs...)
	os.Exit(1)
}

func shouldAddCaller(level Level) bool {
	// Add caller info for error and above, and in debug mode
	return level >= LevelError || level == LevelDebug
}

func parseLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

func convertToSlogLevel(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelFatal:
		return slog.LevelError + 1
	default:
		return slog.LevelInfo
	}
}

func extractContextFields(ctx context.Context) []Field {
	fields := make([]Field, 0)

	// Extract correlation ID if present
	if corrID := ctx.Value("correlation_id"); corrID != nil {
		if corrIDStr, ok := corrID.(string); ok {
			fields = append(fields, String("correlation_id", corrIDStr))
		}
	}

	// Extract user ID if present
	if userID := ctx.Value("user_id"); userID != nil {
		if userIDStr, ok := userID.(string); ok {
			fields = append(fields, String("user_id", userIDStr))
		}
	}

	// Extract request ID if present
	if reqID := ctx.Value("request_id"); reqID != nil {
		if reqIDStr, ok := reqID.(string); ok {
			fields = append(fields, String("request_id", reqIDStr))
		}
	}

	return fields
}

// Global logger instance
var globalLogger Logger
var once sync.Once

// Init initializes the global logger
func Init(opts ...Option) {
	once.Do(func() {
		globalLogger = NewLogger(opts...)
	})
}

// Default returns the global logger, initializing it if needed
func Default() Logger {
	if globalLogger == nil {
		Init()
	}
	return globalLogger
}

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}

// Fatal logs a fatal message using the global logger
func Fatal(msg string, fields ...Field) {
	Default().Fatal(msg, fields...)
}

// With returns a logger with additional fields using the global logger
func With(fields ...Field) Logger {
	return Default().With(fields...)
}

// WithContext returns a logger with context using the global logger
func WithContext(ctx context.Context) Logger {
	return Default().WithContext(ctx)
}

// SetLevel sets the log level using the global logger
func SetLevel(level Level) {
	Default().SetLevel(level)
}

// GetLevel gets the current log level using the global logger
func GetLevel() Level {
	return Default().GetLevel()
}
