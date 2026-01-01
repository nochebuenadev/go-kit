package mw

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/authz"
)

type mockPermProvider struct {
	mask int64
	err  error
}

func (m *mockPermProvider) ResolveMask(ctx context.Context, uid, tenantID, appID string) (int64, error) {
	return m.mask, m.err
}

func TestAuthorizer_Guard(t *testing.T) {
	e := echo.New()
	logger := &mockLogger{}

	t.Run("no identity", func(t *testing.T) {
		// Reset singleton
		rbacInstance = nil
		rbacOnce = sync.Once{}

		auth := GetAuthorizer(logger, &mockPermProvider{}, "app-1")
		handler := auth.Guard(1)(func(c echo.Context) error { return nil })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err != echo.ErrUnauthorized {
			t.Errorf("expected unauthorized, got %v", err)
		}
	})

	t.Run("provider failure", func(t *testing.T) {
		rbacInstance = nil
		rbacOnce = sync.Once{}

		auth := GetAuthorizer(logger, &mockPermProvider{err: errors.New("fail")}, "app-1")
		handler := auth.Guard(1)(func(c echo.Context) error { return nil })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := authz.SetInContext(context.Background(), &authz.Identity{UID: "u1"})
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err != echo.ErrForbidden {
			t.Errorf("expected forbidden, got %v", err)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		rbacInstance = nil
		rbacOnce = sync.Once{}

		auth := GetAuthorizer(logger, &mockPermProvider{mask: 0}, "app-1")
		handler := auth.Guard(1)(func(c echo.Context) error { return nil })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := authz.SetInContext(context.Background(), &authz.Identity{UID: "u1"})
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err != echo.ErrForbidden {
			t.Errorf("expected forbidden, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rbacInstance = nil
		rbacOnce = sync.Once{}

		auth := GetAuthorizer(logger, &mockPermProvider{mask: 2}, "app-1") // Bit 1 set (2^1 = 2)
		handler := auth.Guard(1)(func(c echo.Context) error { return c.NoContent(http.StatusOK) })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := authz.SetInContext(context.Background(), &authz.Identity{UID: "u1"})
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := handler(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}
