package apperr

import (
	"errors"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	code := ErrInvalidInput
	msg := "test message"
	err := New(code, msg)

	if err.Code != code {
		t.Errorf("expected code %s, got %s", code, err.Code)
	}
	if err.Message != msg {
		t.Errorf("expected message %s, got %s", msg, err.Message)
	}
}

func TestWrap(t *testing.T) {
	code := ErrInternal
	msg := "wrapped message"
	originalErr := errors.New("original error")
	err := Wrap(code, msg, originalErr)

	if err.Code != code {
		t.Errorf("expected code %s, got %s", code, err.Code)
	}
	if err.Message != msg {
		t.Errorf("expected message %s, got %s", msg, err.Message)
	}
	if err.Err != originalErr {
		t.Errorf("expected error %v, got %v", originalErr, err.Err)
	}
}

func TestEnsureAppError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if err := EnsureAppError(nil); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("already AppErr", func(t *testing.T) {
		original := New(ErrResourceNotFound, "not found")
		ensured := EnsureAppError(original)
		if ensured != original {
			t.Errorf("expected same pointer, got different")
		}
	})

	t.Run("standard error", func(t *testing.T) {
		original := errors.New("something went wrong")
		ensured := EnsureAppError(original)
		if ensured.Code != ErrInternal {
			t.Errorf("expected code %s, got %s", ErrInternal, ensured.Code)
		}
		if ensured.Err != original {
			t.Errorf("expected wrapped error, got %v", ensured.Err)
		}
	})
}

func TestHelpers(t *testing.T) {
	t.Run("InvalidInput", func(t *testing.T) {
		err := InvalidInput("param %s is required", "id")
		if err.Code != ErrInvalidInput {
			t.Errorf("expected code %s, got %s", ErrInvalidInput, err.Code)
		}
		if err.Message != "param id is required" {
			t.Errorf("expected formatted message, got %s", err.Message)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		err := NotFound("user %d not found", 123)
		if err.Code != ErrResourceNotFound {
			t.Errorf("expected code %s, got %s", ErrResourceNotFound, err.Code)
		}
		if err.Message != "user 123 not found" {
			t.Errorf("expected formatted message, got %s", err.Message)
		}
	})

	t.Run("Internal", func(t *testing.T) {
		err := Internal("database error: %v", errors.New("conn lost"))
		if err.Code != ErrInternal {
			t.Errorf("expected code %s, got %s", ErrInternal, err.Code)
		}
		if err.Message != "database error: conn lost" {
			t.Errorf("expected formatted message, got %s", err.Message)
		}
	})
}

func TestAppErr_Error(t *testing.T) {
	err := New(ErrConflict, "conflict occurred").WithError(errors.New("db error"))
	expected := "ALREADY_EXISTS: conflict occurred → db error"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}
}

func TestAppErr_Detailed(t *testing.T) {
	err := New(ErrInvalidInput, "invalid name").
		WithContext("field", "name").
		WithError(errors.New("too short"))

	detailed := err.Detailed()

	if !strings.Contains(detailed, "Code: INVALID_ARGUMENT") {
		t.Error("detailed output missing code")
	}
	if !strings.Contains(detailed, "Message: invalid name") {
		t.Error("detailed output missing message")
	}
	if !strings.Contains(detailed, "Cause: too short") {
		t.Error("detailed output missing cause")
	}
	if !strings.Contains(detailed, "Context: map[field:name]") {
		t.Error("detailed output missing context")
	}
}

func TestErrorCode_Description(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrInvalidInput, "Invalid input provided"},
		{ErrUnauthorized, "Authentication required"},
		{ErrInternal, "Internal server error"},
		{ErrorCode("UNKNOWN"), "UNKNOWN"},
	}

	for _, tt := range tests {
		if tt.code.Description() != tt.expected {
			t.Errorf("for code %s, expected description %s, got %s", tt.code, tt.expected, tt.code.Description())
		}
	}
}
