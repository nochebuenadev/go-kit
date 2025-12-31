/*
Package mw provides custom middlewares for the Echo web framework.

Currently, it focuses on providing a standardized error handler that maps
application-specific errors (apperr.AppErr) to appropriate HTTP responses
and logs them using the project's logging utility (logz).

Features:
  - AppErrorHandler: Captures both standard and application errors, returns formatted JSON
    responses, and ensures all failures are logged with rich context.
*/
package mw
