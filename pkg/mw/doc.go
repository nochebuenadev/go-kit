/*
Package mw provides a collection of reusable middleware for the Echo web framework.

These middlewares implement common cross-cutting concerns such as error handling,
correlation tracking, identity validation, and access control, ensuring that all
services within the go-kit ecosystem behave consistently.

Middlewares:
  - AppErrorHandler: Standardizes error responses by mapping apperr.AppErr and echo.HTTPError
    to consistent HTTP formats and status codes.
  - WithRequestID: propagates correlation IDs from headers to the context for tracing.
  - FirebaseAuth: validates identity tokens using the Firebase Admin SDK.
  - EnrichmentMiddleware: extracts tenant IDs and metadata from JWT claims.
  - Authorizer (RBAC): enforces permission-based access control at the route level.

Each middleware is designed to be easily pluggable and adheres to the project's
structured logging (logz) and error reporting (apperr) standards.
*/
package mw
