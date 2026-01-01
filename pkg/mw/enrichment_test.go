package mw

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/authz"
)

func TestEnrichmentMiddleware(t *testing.T) {
	e := echo.New()

	t.Run("missing tenant ID", func(t *testing.T) {
		mw := EnrichmentMiddleware(true, "X-Tenant-ID", nil)
		handler := mw(func(c echo.Context) error { return nil })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err == nil {
			t.Fatal("expected error due to missing tenant ID")
		}
	})

	t.Run("missing identity", func(t *testing.T) {
		mw := EnrichmentMiddleware(true, "X-Tenant-ID", nil)
		handler := mw(func(c echo.Context) error { return nil })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Tenant-ID", "tenant-1")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err != echo.ErrUnauthorized {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("success with optional headers", func(t *testing.T) {
		mw := EnrichmentMiddleware(true, "X-Tenant-ID", map[string]string{"app": "X-App-ID"})
		handler := mw(func(c echo.Context) error {
			id, _ := authz.FromContext(c.Request().Context())
			if id.TenantID != "tenant-1" {
				t.Errorf("expected tenant-1, got %s", id.TenantID)
			}
			enrich, _ := EnrichmentFromContext(c.Request().Context())
			if enrich["app"] != "app-1" {
				t.Errorf("expected app-1, got %s", enrich["app"])
			}
			return nil
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Tenant-ID", "tenant-1")
		req.Header.Set("X-App-ID", "app-1")

		ctx := authz.SetInContext(context.Background(), &authz.Identity{UID: "user-1"})
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := handler(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		mw := EnrichmentMiddleware(false, "X-Tenant-ID", nil)
		handler := mw(func(c echo.Context) error { return nil })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := handler(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
