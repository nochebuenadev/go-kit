/*
Package mw provides a collection of middleware for the Echo web framework.

Middlewares:
- AppErrorHandler: Standardized error response mapping.
- WithRequestID: Trace ID propagation.
- FirebaseAuth: Identity validation via Firebase Admin SDK.
- EnrichmentMiddleware: Tenant and metadata extraction.
- Authorizer (RBAC): Permission-based access control.

Each middleware is designed to be easily pluggable and follows the projects structured logging
and error reporting standards.
*/
package mw
