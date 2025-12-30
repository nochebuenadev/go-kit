package apperr

import (
	"errors"
	"fmt"
)

// AppErr represents a standard application error with rich context.
// It implements the full error interface and provides additional methods
// for structured error reporting and logging.
type AppErr struct {
	Code    ErrorCode      // A high-level error code for programmatic handling.
	Message string         // A human-readable message describing the error.
	Err     error          // The underlying cause of the error (optional).
	Context map[string]any // Key-value pairs providing additional context (optional).
}

// New creates a new AppErr with the given code and message.
func New(code ErrorCode, message string) *AppErr {
	return &AppErr{
		Code:    code,
		Message: message,
	}
}

// Wrap creates a new AppErr that wraps an existing error with a code and message.
func Wrap(code ErrorCode, message string, err error) *AppErr {
	return &AppErr{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// EnsureAppError converts a generic error into an AppErr.
// If the error is already an AppErr, it is returned as is.
// Otherwise, it's wrapped as an ErrInternal.
func EnsureAppError(err error) *AppErr {
	if err == nil {
		return nil
	}

	var appErr *AppErr
	if errors.As(err, &appErr) {
		return appErr
	}

	return Wrap(ErrInternal, "An unexpected error occurred", err)
}

// InvalidInput creates a new AppErr with ErrInvalidInput code.
// It accepts a format string and arguments for the message.
func InvalidInput(msg string, args ...any) *AppErr {
	return New(ErrInvalidInput, fmt.Sprintf(msg, args...))
}

// NotFound creates a new AppErr with ErrResourceNotFound code.
// It accepts a format string and arguments for the message.
func NotFound(msg string, args ...any) *AppErr {
	return New(ErrResourceNotFound, fmt.Sprintf(msg, args...))
}

// Internal creates a new AppErr with ErrInternal code.
// It accepts a format string and arguments for the message.
func Internal(msg string, args ...any) *AppErr {
	return New(ErrInternal, fmt.Sprintf(msg, args...))
}

// Error returns a formatted string representation of the error.
// It includes the code, the message, and the underlying error if present.
func (e *AppErr) Error() string {
	base := fmt.Sprintf("%s: %s", e.Code, e.Message)

	if e.Err != nil {
		base = fmt.Sprintf("%s → %v", base, e.Err)
	}

	return base
}

// Detailed returns a more verbose string representation of the error,
// including its code, message, cause, and full context.
func (e *AppErr) Detailed() string {
	details := fmt.Sprintf("Code: %s | Message: %s", e.Code, e.Message)

	if e.Err != nil {
		details = fmt.Sprintf("%s | Cause: %v", details, e.Err)
	}

	if len(e.Context) > 0 {
		details = fmt.Sprintf("%s | Context: %v", details, e.Context)
	}

	return details
}

// WithContext adds a key-value pair to the error's context.
// It returns the modified AppErr for chaining.
func (e *AppErr) WithContext(key string, value any) *AppErr {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithError sets the underlying cause of the error.
// It returns the modified AppErr for chaining.
func (e *AppErr) WithError(err error) *AppErr {
	e.Err = err
	return e
}
