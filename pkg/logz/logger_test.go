package logz

import (
	"context"
	"errors"
	"testing"
)

// mockAppErr implements the appErrorData interface for testing purposes.
type mockAppErr struct {
	code    string
	message string
	context map[string]any
}

func (e *mockAppErr) Error() string              { return e.message }
func (e *mockAppErr) GetCode() string            { return e.code }
func (e *mockAppErr) GetContext() map[string]any { return e.context }

func TestMustInit(t *testing.T) {
	// Test normal initialization
	MustInit("service", "test-service")
	logger := Global()
	if logger == nil {
		t.Fatal("Global() returned nil after MustInit()")
	}
}

func TestLoggerMethods(t *testing.T) {
	MustInit()
	logger := Global()

	// These methods primarily call slog, so we just ensure they don't panic.
	// In a more advanced test, we'd capture output (e.g., using a custom buffer).
	logger.Debug("debug message", "key", "value")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message", errors.New("some error"))
}

func TestLogError(t *testing.T) {
	MustInit()
	logger := Global()

	t.Run("standard error", func(t *testing.T) {
		logger.LogError("standard error msg", errors.New("std err"))
	})

	t.Run("app error", func(t *testing.T) {
		ae := &mockAppErr{
			code:    "TEST_CODE",
			message: "app error message",
			context: map[string]any{"user_id": 123},
		}
		logger.LogError("app error msg", ae)
	})

	t.Run("nil error", func(t *testing.T) {
		logger.LogError("nil error msg", nil)
	})
}

func TestContextFeatures(t *testing.T) {
	ctx := context.Background()

	t.Run("CorrelationID", func(t *testing.T) {
		ctx = WithRequestID(ctx, "req-123")
		val := ctx.Value(ctxRequestIDKey{})
		if val != "req-123" {
			t.Errorf("expected correlation ID req-123, got %v", val)
		}
	})

	t.Run("Fields", func(t *testing.T) {
		ctx = WithField(ctx, "foo", "bar")
		ctx = WithFields(ctx, map[string]any{"baz": "qux"})

		fields, ok := ctx.Value(ctxExtraFieldsKey{}).(map[string]any)
		if !ok {
			t.Fatal("no fields found in context")
		}
		if fields["foo"] != "bar" || fields["baz"] != "qux" {
			t.Errorf("unexpected fields: %v", fields)
		}
	})
}

func TestLoggerWith(t *testing.T) {
	MustInit()
	logger := Global().With("extra", "value")
	if logger == nil {
		t.Fatal("With() returned nil")
	}
	logger.Info("info with extra")
}

func TestLoggerWithContext(t *testing.T) {
	MustInit()
	logger := Global()

	ctx := context.Background()
	ctx = WithRequestID(ctx, "123")
	ctx = WithField(ctx, "user", "admin")

	ctxLogger := logger.WithContext(ctx)
	if ctxLogger == nil {
		t.Fatal("WithContext() returned nil")
	}
	ctxLogger.Info("log with context")

	// Test nil context
	if logger.WithContext(context.TODO()) != logger {
		t.Error("WithContext(TODO) should return the original logger")
	}
}
