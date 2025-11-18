package logging

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("Level.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFieldCreators(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		field := String("key", "value")
		if field.Key != "key" || field.Value != "value" {
			t.Errorf("String() = %+v, want key='key' value='value'", field)
		}
	})

	t.Run("Int", func(t *testing.T) {
		field := Int("key", 42)
		if field.Key != "key" || field.Value != 42 {
			t.Errorf("Int() = %+v, want key='key' value=42", field)
		}
	})

	t.Run("Int64", func(t *testing.T) {
		field := Int64("key", int64(42))
		if field.Key != "key" || field.Value != int64(42) {
			t.Errorf("Int64() = %+v, want key='key' value=42", field)
		}
	})

	t.Run("Float64", func(t *testing.T) {
		field := Float64("key", 3.14)
		if field.Key != "key" || field.Value != 3.14 {
			t.Errorf("Float64() = %+v, want key='key' value=3.14", field)
		}
	})

	t.Run("Duration", func(t *testing.T) {
		duration := time.Second
		field := Duration("key", duration)
		if field.Key != "key" || field.Value != duration.String() {
			t.Errorf("Duration() = %+v, want key='key' value='1s'", field)
		}
	})

	t.Run("Bool", func(t *testing.T) {
		field := Bool("key", true)
		if field.Key != "key" || field.Value != true {
			t.Errorf("Bool() = %+v, want key='key' value=true", field)
		}
	})

	t.Run("Time", func(t *testing.T) {
		now := time.Now()
		field := Time("key", now)
		if field.Key != "key" || field.Value != now {
			t.Errorf("Time() = %+v, want key='key' value=%v", field, now)
		}
	})

	t.Run("Any", func(t *testing.T) {
		value := map[string]string{"test": "value"}
		field := Any("key", value)
		if field.Key != "key" {
			t.Errorf("Any() key = %q, want 'key'", field.Key)
		}
	})
}

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(
		WithLevel("info"),
		WithOutput(&buf),
	)

	if logger == nil {
		t.Fatal("NewLogger() returned nil")
	}

	// Test logging
	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Logger output doesn't contain expected message. Got: %s", output)
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(
		WithLevel("debug"),
		WithOutput(&buf),
	)

	t.Run("Debug", func(t *testing.T) {
		buf.Reset()
		logger.Debug("debug message")
		if !strings.Contains(buf.String(), "debug message") {
			t.Error("Debug message not logged")
		}
	})

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		logger.Info("info message")
		if !strings.Contains(buf.String(), "info message") {
			t.Error("Info message not logged")
		}
	})

	t.Run("Warn", func(t *testing.T) {
		buf.Reset()
		logger.Warn("warn message")
		if !strings.Contains(buf.String(), "warn message") {
			t.Error("Warn message not logged")
		}
	})

	t.Run("Error", func(t *testing.T) {
		buf.Reset()
		logger.Error("error message")
		if !strings.Contains(buf.String(), "error message") {
			t.Error("Error message not logged")
		}
	})
}

func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithOutput(&buf))

	childLogger := logger.With(String("request_id", "123"))
	if childLogger == nil {
		t.Fatal("With() returned nil")
	}

	buf.Reset()
	childLogger.Info("test message")
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Error("Logger output doesn't contain expected message")
	}
	if !strings.Contains(output, "request_id") {
		t.Error("Logger output doesn't contain request_id field")
	}
}

func TestLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithOutput(&buf))

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-id")
	contextLogger := logger.WithContext(ctx)

	if contextLogger == nil {
		t.Fatal("WithContext() returned nil")
	}

	buf.Reset()
	contextLogger.Info("test message")
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Error("Logger output doesn't contain expected message")
	}
	if !strings.Contains(output, "correlation_id") {
		t.Error("Logger output doesn't contain correlation_id from context")
	}
}

func TestLoggerSetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(
		WithLevel("info"),
		WithOutput(&buf),
	)

	// Debug should not be logged at info level
	logger.Debug("debug message")
	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message should not be logged at info level")
	}

	// Set level to debug
	logger.SetLevel(LevelDebug)

	// Verify the level was set
	if logger.GetLevel() != LevelDebug {
		t.Errorf("Level after SetLevel = %v, want %v", logger.GetLevel(), LevelDebug)
	}
}

func TestLoggerGetLevel(t *testing.T) {
	logger := NewLogger(WithLevel("warn"))

	level := logger.GetLevel()
	if level != LevelWarn {
		t.Errorf("GetLevel() = %v, want %v", level, LevelWarn)
	}
}

func TestWithLevel(t *testing.T) {
	tests := []struct {
		levelStr string
		expected Level
	}{
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"error", LevelError},
		{"fatal", LevelFatal},
		{"unknown", LevelInfo}, // Should default to info
	}

	for _, tt := range tests {
		t.Run(tt.levelStr, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(
				WithLevel(tt.levelStr),
				WithOutput(&buf),
			)
			if logger.GetLevel() != tt.expected {
				t.Errorf("Level = %v, want %v", logger.GetLevel(), tt.expected)
			}
		})
	}
}

func TestWithAddSource(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(
		WithOutput(&buf),
		WithAddSource(),
	)

	logger.Info("test message")
	// Just verify the logger was created successfully
	if logger == nil {
		t.Error("Logger with AddSource should not be nil")
	}
}

func TestWithEnvironment(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(
		WithOutput(&buf),
		WithEnvironment("production"),
	)

	logger.Info("test message")
	// Just verify the logger was created successfully
	if logger == nil {
		t.Error("Logger with Environment should not be nil")
	}
}

func TestGlobalLogger(t *testing.T) {
	// Reset global logger for testing
	globalLogger = nil
	once = *new(sync.Once)

	t.Run("Init", func(t *testing.T) {
		var buf bytes.Buffer
		Init(WithOutput(&buf))

		if globalLogger == nil {
			t.Error("Global logger should be initialized")
		}
	})

	t.Run("Default", func(t *testing.T) {
		logger := Default()
		if logger == nil {
			t.Error("Default() should return a logger")
		}
	})

	t.Run("Debug", func(t *testing.T) {
		var buf bytes.Buffer
		globalLogger = NewLogger(WithLevel("debug"), WithOutput(&buf))
		Debug("test debug")
		if !strings.Contains(buf.String(), "test debug") {
			t.Error("Global Debug() should log message")
		}
	})

	t.Run("Info", func(t *testing.T) {
		var buf bytes.Buffer
		globalLogger = NewLogger(WithOutput(&buf))
		Info("test info")
		if !strings.Contains(buf.String(), "test info") {
			t.Error("Global Info() should log message")
		}
	})

	t.Run("Warn", func(t *testing.T) {
		var buf bytes.Buffer
		globalLogger = NewLogger(WithOutput(&buf))
		Warn("test warn")
		if !strings.Contains(buf.String(), "test warn") {
			t.Error("Global Warn() should log message")
		}
	})

	t.Run("Error", func(t *testing.T) {
		var buf bytes.Buffer
		globalLogger = NewLogger(WithOutput(&buf))
		Error("test error")
		if !strings.Contains(buf.String(), "test error") {
			t.Error("Global Error() should log message")
		}
	})

	t.Run("With", func(t *testing.T) {
		var buf bytes.Buffer
		globalLogger = NewLogger(WithOutput(&buf))
		logger := With(String("key", "value"))
		if logger == nil {
			t.Error("Global With() should return a logger")
		}
	})

	t.Run("WithContext", func(t *testing.T) {
		var buf bytes.Buffer
		globalLogger = NewLogger(WithOutput(&buf))
		ctx := context.Background()
		logger := WithContext(ctx)
		if logger == nil {
			t.Error("Global WithContext() should return a logger")
		}
	})

	t.Run("SetLevel", func(t *testing.T) {
		var buf bytes.Buffer
		globalLogger = NewLogger(WithOutput(&buf))
		SetLevel(LevelWarn)
		if GetLevel() != LevelWarn {
			t.Error("Global SetLevel() should set the level")
		}
	})

	t.Run("GetLevel", func(t *testing.T) {
		var buf bytes.Buffer
		globalLogger = NewLogger(WithLevel("info"), WithOutput(&buf))
		level := GetLevel()
		if level != LevelInfo {
			t.Errorf("Global GetLevel() = %v, want %v", level, LevelInfo)
		}
	})
}

func TestExtractContextFields(t *testing.T) {
	t.Run("correlation_id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "correlation_id", "test-corr-id")
		fields := extractContextFields(ctx)

		found := false
		for _, field := range fields {
			if field.Key == "correlation_id" && field.Value == "test-corr-id" {
				found = true
				break
			}
		}
		if !found {
			t.Error("extractContextFields should extract correlation_id")
		}
	})

	t.Run("user_id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", "test-user-id")
		fields := extractContextFields(ctx)

		found := false
		for _, field := range fields {
			if field.Key == "user_id" && field.Value == "test-user-id" {
				found = true
				break
			}
		}
		if !found {
			t.Error("extractContextFields should extract user_id")
		}
	})

	t.Run("request_id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-req-id")
		fields := extractContextFields(ctx)

		found := false
		for _, field := range fields {
			if field.Key == "request_id" && field.Value == "test-req-id" {
				found = true
				break
			}
		}
		if !found {
			t.Error("extractContextFields should extract request_id")
		}
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		fields := extractContextFields(ctx)

		if len(fields) != 0 {
			t.Errorf("extractContextFields with empty context should return empty slice, got %d fields", len(fields))
		}
	})
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(
		WithLevel("warn"),
		WithOutput(&buf),
	)

	// Debug and Info should not be logged at warn level
	logger.Debug("debug message")
	logger.Info("info message")

	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message should not be logged at warn level")
	}
	if strings.Contains(buf.String(), "info message") {
		t.Error("Info message should not be logged at warn level")
	}

	// Warn should be logged
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn message should be logged at warn level")
	}

	buf.Reset()
	// Error should be logged
	logger.Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error message should be logged at warn level")
	}
}
