/*
Package apperr provides a standardized way to handle application-specific errors in Go.

It defines a robust error structure (AppErr) that includes error codes, human-readable messages,
underlying causes, and contextual information. This helps in maintaining consistency across
the application and simplifies error handling and logging.

The package utilizes ErrorCode to categorize errors, allowing for easier mapping to HTTP status
codes or other external representations.

Example usage:

	err := apperr.New(apperr.ErrInvalidInput, "username is required")
	fmt.Println(err.Detailed())

	// Wrapping an existing error
	wrappedErr := apperr.Wrap(apperr.ErrInternal, "failed to save user", originalErr)
*/
package apperr
