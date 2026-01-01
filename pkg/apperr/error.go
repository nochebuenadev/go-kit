package apperr

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AppErr represents a standard application error with rich context.
// It implements the full error interface and provides additional methods
// for structured error reporting and logging.
type AppErr struct {
	// code is the machine-readable error code.
	code ErrorCode
	// message is the human-readable description of the error.
	message string
	// err is the underlying cause of the error (optional).
	err error
	// context contains additional metadata associated with the error (optional).
	context map[string]any
}

// New creates a new AppErr with the given code and message.
func New(code ErrorCode, message string) *AppErr {
	return &AppErr{
		code:    code,
		message: message,
	}
}

// Wrap creates a new AppErr that wraps an existing error with a code and message.
func Wrap(code ErrorCode, message string, err error) *AppErr {
	return &AppErr{
		code:    code,
		message: message,
		err:     err,
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
	base := fmt.Sprintf("%s: %s", e.code, e.message)

	if e.err != nil {
		base = fmt.Sprintf("%s → %v", base, e.err)
	}

	return base
}

// Detailed returns a more verbose string representation of the error,
// including its code, message, cause, and full context.
func (e *AppErr) Detailed() string {
	details := fmt.Sprintf("code: %s | message: %s", e.code, e.message)

	if e.err != nil {
		details = fmt.Sprintf("%s | Cause: %v", details, e.err)
	}

	if len(e.context) > 0 {
		details = fmt.Sprintf("%s | context: %v", details, e.context)
	}

	return details
}

// GetCode returns the string representation of the error code.
func (e *AppErr) GetCode() string {
	return string(e.code)
}

// GetContext returns the contextual data associated with the error.
// If no context exists, it returns an empty map.
func (e *AppErr) GetContext() map[string]any {
	if e.context == nil {
		return make(map[string]any)
	}

	return e.context
}

// WithContext adds a key-value pair to the error's context.
// It returns the modified AppErr for chaining.
func (e *AppErr) WithContext(key string, value any) *AppErr {
	if e.context == nil {
		e.context = make(map[string]any)
	}
	e.context[key] = value
	return e
}

// WithError sets the underlying cause of the error.
// It returns the modified AppErr for chaining.
func (e *AppErr) WithError(err error) *AppErr {
	e.err = err
	return e
}

// MarshalJSON implements the json.Marshaler interface.
func (e *AppErr) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Context map[string]any `json:"context,omitempty"`
	}{
		Code:    string(e.code),
		Message: e.message,
		Context: e.context,
	})
}
