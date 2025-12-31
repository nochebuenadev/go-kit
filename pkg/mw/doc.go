/*
Package mw provides custom middlewares for the Echo web framework.

It includes utilities for standardizing error handling and managing request context
information, such as correlation IDs, to ensure consistent behavior and observability
across the application.

Features:
  - AppErrorHandler: Captures both standard and application errors, returns formatted JSON
    responses, and ensures all failures are logged with rich context.
  - WithRequestID: Middleware that synchronizes the Echo Request ID with the logz context,
    enabling trace-id consistency in all logs.
*/
package mw
