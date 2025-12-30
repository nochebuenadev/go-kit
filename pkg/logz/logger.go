package logutil

import (
	"capital-link-api/internal/domain"
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, err error, args ...any)
	LogError(msg string, err error, args ...any)
	Fatal(msg string, err error, args ...any)
	With(args ...any) Logger
	WithContext(ctx context.Context) Logger
}

type slogLogger struct {
	logger *slog.Logger
}

type (
	ctxCorrelationKey struct{}
	ctxExtraFieldsKey struct{}
)

const (
	EnvLogLevel = "LOG_LEVEL"
	EnvLogJSON  = "LOG_JSON_OUTPUT"
)

var (
	globalLogger Logger
	once         sync.Once
)

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

func Global() Logger {
	if globalLogger == nil {
		panic("logutil: the logger has not been initialized. Call MustInit() first")
	}
	return globalLogger
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

	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		args = append(args, slog.String("error_code", appErr.Code))
		for k, v := range appErr.Context {
			args = append(args, slog.Any(k, v))
		}

		l.Error(msg, appErr.Err, args...)
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

	if id, ok := ctx.Value(ctxCorrelationKey{}).(string); ok && id != "" {
		newLogger = newLogger.With(slog.String("request_id", id))
	}

	if fields, ok := ctx.Value(ctxExtraFieldsKey{}).(map[string]any); ok {
		for k, v := range fields {
			newLogger = newLogger.With(k, v)
		}
	}

	return &slogLogger{logger: newLogger}
}

func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxCorrelationKey{}, id)
}

func WithField(ctx context.Context, key string, value any) context.Context {
	return WithFields(ctx, map[string]any{key: value})
}

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
