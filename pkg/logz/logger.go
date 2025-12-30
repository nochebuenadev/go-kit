package logz

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type (
	// Logger defines the interface for our application logging.
	Logger interface {
		// Debug logs a message at DEBUG level.
		Debug(msg string, args ...any)
		// Info logs a message at INFO level.
		Info(msg string, args ...any)
		// Warn logs a message at WARN level.
		Warn(msg string, args ...any)
		// Error logs a message at ERROR level with an attached error.
		Error(msg string, err error, args ...any)
		// LogError logs an error, automatically extracting code and context if it's an apperr.AppErr.
		LogError(msg string, err error, args ...any)
		// Fatal logs an error and exits the application with code 1.
		Fatal(msg string, err error, args ...any)
		// With returns a new Logger with the given attributes.
		With(args ...any) Logger
		// WithContext returns a new Logger that includes values from the context.
		WithContext(ctx context.Context) Logger
	}

	// appErrorData is a private interface to avoid direct dependency on apperr package internals
	// while still being able to extract its rich information.
	appErrorData interface {
		Error() string
		GetCode() string
		GetContext() map[string]any
	}

	// slogLogger is the concrete implementation of Logger using slog.
	slogLogger struct {
		logger *slog.Logger
	}

	// ctxCorrelationKey is used for request/correlation IDs in context.
	ctxCorrelationKey struct{}
	// ctxExtraFieldsKey is used for storing arbitrary logging fields in context.
	ctxExtraFieldsKey struct{}
)

const (
	// EnvLogLevel is the environment variable to set the minimum log level (e.g., DEBUG, INFO, WARN, ERROR).
	EnvLogLevel = "LOG_LEVEL"
	// EnvLogJSON is the environment variable to enable JSON output (set to "true" or "1").
	EnvLogJSON = "LOG_JSON_OUTPUT"
)

var (
	globalLogger Logger
	once         sync.Once
)

// MustInit initializes the global logger once. It reads configuration from environment variables.
// staticArgs can be used to add global fields to all logs (e.g., service name, environment).
func MustInit(staticArgs ...any) {
	once.Do(func() {
		level := getLogLevelFromEnv()
		isJSON := getLogFormatFromEnv()

		var handler slog.Handler
		opts := &slog.HandlerOptions{
			Level:     level,
			AddSource: false,
		}

		if isJSON {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, opts)
		}

		baseLogger := slog.New(handler)

		if len(staticArgs) > 0 {
			baseLogger = baseLogger.With(staticArgs...)
		}

		globalLogger = &slogLogger{logger: baseLogger}
	})
}

// Global returns the singleton logger instance. It panics if MustInit hasn't been called.
func Global() Logger {
	if globalLogger == nil {
		panic("logz: the logger has not been initialized. Call MustInit() first")
	}
	return globalLogger
}

// WithCorrelationID adds a correlation ID to the context.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxCorrelationKey{}, id)
}

// WithField adds a single key-value pair to the context for logging.
func WithField(ctx context.Context, key string, value any) context.Context {
	return WithFields(ctx, map[string]any{key: value})
}

// WithFields adds multiple key-value pairs to the context for logging.
func WithFields(ctx context.Context, fields map[string]any) context.Context {
	existing, _ := ctx.Value(ctxExtraFieldsKey{}).(map[string]any)
	newMap := make(map[string]any, len(existing)+len(fields))
	for k, v := range existing {
		newMap[k] = v
	}
	for k, v := range fields {
		newMap[k] = v
	}
	return context.WithValue(ctx, ctxExtraFieldsKey{}, newMap)
}

func (l *slogLogger) Debug(msg string, args ...any) { l.logger.Debug(msg, args...) }
func (l *slogLogger) Info(msg string, args ...any)  { l.logger.Info(msg, args...) }
func (l *slogLogger) Warn(msg string, args ...any)  { l.logger.Warn(msg, args...) }

func (l *slogLogger) Error(msg string, err error, args ...any) {
	if err != nil {
		args = append(args, slog.Any("error", err))
	}
	l.logger.Error(msg, args...)
}

func (l *slogLogger) LogError(msg string, err error, args ...any) {
	if err == nil {
		return
	}

	var ae appErrorData
	if errors.As(err, &ae) {
		args = append(args, slog.String("error_code", ae.GetCode()))

		for k, v := range ae.GetContext() {
			args = append(args, slog.Any(k, v))
		}

		l.Error(msg, err, args...)
		return
	}

	l.Error(msg, err, args...)
}

func (l *slogLogger) Fatal(msg string, err error, args ...any) {
	l.Error(msg, err, args...)
	os.Exit(1)
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{logger: l.logger.With(args...)}
}

func (l *slogLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	newLogger := l.logger
	modified := false

	if id, ok := ctx.Value(ctxCorrelationKey{}).(string); ok && id != "" {
		newLogger = newLogger.With(slog.String("request_id", id))
		modified = true
	}

	if fields, ok := ctx.Value(ctxExtraFieldsKey{}).(map[string]any); ok {
		for k, v := range fields {
			newLogger = newLogger.With(k, v)
		}
		modified = true
	}

	if !modified {
		return l
	}

	return &slogLogger{logger: newLogger}
}

func getLogLevelFromEnv() slog.Level {
	switch strings.ToUpper(os.Getenv(EnvLogLevel)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getLogFormatFromEnv() bool {
	val := strings.ToLower(os.Getenv(EnvLogJSON))
	return val == "true" || val == "1"
}
