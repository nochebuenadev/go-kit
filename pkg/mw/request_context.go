package mw

import (
	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

// WithRequestID is a middleware that extracts the Request ID from the Echo response header
// (usually injected by the echo.middleware.RequestID middleware) and injects it into the
// request context using logz.WithRequestID. This ensures that all downstream logs
// associated with this request will include the "request_id" attribute.
func WithRequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)

			if reqID != "" {
				ctx := logz.WithRequestID(c.Request().Context(), reqID)

				c.SetRequest(c.Request().WithContext(ctx))
			}

			return next(c)
		}
	}
}
