package mw

import (
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/authz"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// Authorizer defines the interface for RBAC protection.
	Authorizer interface {
		// Guard returns a middleware that checks if the authenticated user has the required permission bit.
		Guard(requiredBit int64) echo.MiddlewareFunc
	}

	// rbacComponent is the concrete implementation of Authorizer.
	rbacComponent struct {
		logger   logz.Logger
		provider authz.PermissionProvider
		appID    string
	}
)

var (
	// rbacInstance is the singleton authorizer instance.
	rbacInstance Authorizer
	// rbacOnce ensures that the authorizer is initialized only once.
	rbacOnce sync.Once
)

// GetAuthorizer returns the singleton instance of the Authorizer.
func GetAuthorizer(logger logz.Logger, provider authz.PermissionProvider, appID string) Authorizer {
	rbacOnce.Do(func() {
		rbacInstance = &rbacComponent{
			logger:   logger,
			provider: provider,
			appID:    appID,
		}
	})

	return rbacInstance
}

// Guard implements the Authorizer interface.
// It resolves the user's permission mask and checks the required bit.
func (r *rbacComponent) Guard(requiredBit int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			id, ok := authz.FromContext(ctx)
			if !ok {
				r.logger.Warn("mw: intento de acceso RBAC sin identidad en el contexto")
				return echo.ErrUnauthorized
			}

			mask, err := r.provider.ResolveMask(ctx, id.UID, id.TenantID, r.appID)
			if err != nil {
				r.logger.LogError("mw: el proveedor de permisos falló al resolver acceso", err,
					"uid", id.UID, "tenant_id", id.TenantID, "app_id", r.appID)
				return echo.ErrForbidden
			}

			if !id.HasPermission(mask, requiredBit) {
				return echo.ErrForbidden
			}

			return next(c)
		}
	}
}
