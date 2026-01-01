package mw

import (
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/authz"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

// FirebaseAuth validates a Firebase ID Token from the Authorization header (Bearer token).
// If valid, it injects the Identity into the context and adds user_id to the logger.
func FirebaseAuth(client *auth.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				return echo.ErrUnauthorized
			}

			decoded, err := client.VerifyIDTokenAndCheckRevoked(c.Request().Context(), token)
			if err != nil {
				return echo.ErrUnauthorized
			}

			id := &authz.Identity{
				UID: decoded.UID,
			}

			// Attempt to extract display name and email if available
			if name, ok := decoded.Claims["name"].(string); ok {
				id.DisplayName = name
			}
			if email, ok := decoded.Claims["email"].(string); ok {
				id.Email = email
			}

			newCtx := logz.WithField(c.Request().Context(), "user_id", id.UID)
			newCtx = authz.SetInContext(newCtx, id)
			c.SetRequest(c.Request().WithContext(newCtx))

			return next(c)
		}
	}
}
