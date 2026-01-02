package mw

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/authz"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// enrichmentKey is a private type for storing optional metadata in the context.
	enrichmentKey struct{}
)

// EnrichmentMiddleware extracts a tenant ID from a header and enriches the identity and logging context.
// It also extracts optional headers and stores them in the context.
// requires: an Identity to be already present in the context (e.g. from FirebaseAuth).
func EnrichmentMiddleware(isEnabled bool, tenantHeader string, optionalHeaders map[string]string) echo.MiddlewareFunc {
	if !isEnabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			tenantID := c.Request().Header.Get(tenantHeader)
			if tenantID == "" {
				return apperr.InvalidInput("el encabezado %q es requerido para identificar el tenant", tenantHeader).
					WithContext("header", tenantHeader)
			}

			identity, ok := authz.FromContext(ctx)
			if !ok {
				return echo.ErrUnauthorized
			}

			identity.TenantID = tenantID

			ctx = authz.SetInContext(ctx, identity)
			ctx = logz.WithField(ctx, "tenant_id", tenantID)

			extraValues := make(map[string]string)
			for key, headerName := range optionalHeaders {
				val := c.Request().Header.Get(headerName)
				if val != "" {
					extraValues[key] = val
					ctx = logz.WithField(ctx, key, val)
				}
			}

			ctx = setEnrichmentInContext(ctx, extraValues)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// EnrichmentFromContext retrieves the optional headers map from the context.
func EnrichmentFromContext(ctx context.Context) (map[string]string, bool) {
	v, ok := ctx.Value(enrichmentKey{}).(map[string]string)
	return v, ok
}

// setEnrichmentInContext injects the optional headers map into the context.
func setEnrichmentInContext(ctx context.Context, values map[string]string) context.Context {
	return context.WithValue(ctx, enrichmentKey{}, values)
}
