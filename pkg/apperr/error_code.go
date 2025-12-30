package apperr

// ErrorCode represents a unique identifier for a specific type of error.
// These codes are intended to be stable and can be used for programmatic error handling
// or mapping to external error representations (e.g., HTTP status codes).
type ErrorCode string

const (
	// ErrInvalidInput indicates that the provided input is malformed or invalid.
	// Typically maps to HTTP 400 Bad Request.
	ErrInvalidInput ErrorCode = "INVALID_ARGUMENT"

	// ErrUnauthorized indicates that the request requires authentication.
	// Typically maps to HTTP 401 Unauthorized.
	ErrUnauthorized ErrorCode = "UNAUTHENTICATED"

	// ErrPermissionDenied indicates that the authenticated user lacks necessary permissions.
	// Typically maps to HTTP 403 Forbidden.
	ErrPermissionDenied ErrorCode = "PERMISSION_DENIED"

	// ErrResourceNotFound indicates that the requested resource could not be found.
	// Typically maps to HTTP 404 Not Found.
	ErrResourceNotFound ErrorCode = "NOT_FOUND"

	// ErrConflict indicates that the request conflicts with the current state of the resource.
	// Typically maps to HTTP 409 Conflict.
	ErrConflict ErrorCode = "ALREADY_EXISTS"

	// ErrInternal indicates an unexpected server-side error.
	// Typically maps to HTTP 500 Internal Server Error.
	ErrInternal ErrorCode = "INTERNAL_ERROR"

	// ErrNotImplemented indicates that the requested functionality is not implemented.
	// Typically maps to HTTP 501 Not Implemented.
	ErrNotImplemented ErrorCode = "NOT_IMPLEMENTED"

	// ErrUnavailable indicates that the service is temporarily unavailable.
	// Typically maps to HTTP 503 Service Unavailable.
	ErrUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// ErrDeadlineExceeded indicates that the request timed out.
	// Typically maps to HTTP 504 Gateway Timeout.
	ErrDeadlineExceeded ErrorCode = "TIMEOUT"
)

// Description returns a human-readable description for the error code.
// If the code is unknown, it returns the code itself as a string.
func (c ErrorCode) Description() string {
	switch c {
	case ErrInvalidInput:
		return "Invalid input provided"
	case ErrUnauthorized:
		return "Authentication required"
	case ErrPermissionDenied:
		return "Insufficient permissions"
	case ErrResourceNotFound:
		return "Resource not found"
	case ErrConflict:
		return "Resource already exists"
	case ErrInternal:
		return "Internal server error"
	case ErrNotImplemented:
		return "Feature not implemented"
	case ErrUnavailable:
		return "Service temporarily unavailable"
	case ErrDeadlineExceeded:
		return "Request timeout"
	default:
		return string(c)
	}
}
