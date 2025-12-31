package mw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

func TestWithRequestID(t *testing.T) {
	e := echo.New()
	mw := WithRequestID()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	const testID = "test-request-id"

	// Simulate RequestID middleware setting the header
	c.Response().Header().Set(echo.HeaderXRequestID, testID)

	handler := mw(func(c echo.Context) error {
		// Extract request ID from context using logz helper
		id := logz.GetRequestID(c.Request().Context())

		if id != testID {
			t.Errorf("expected request ID %s in context, got %s", testID, id)
		}

		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestWithRequestID_Empty(t *testing.T) {
	e := echo.New()
	mw := WithRequestID()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// No header set
	handler := mw(func(c echo.Context) error {
		id := logz.GetRequestID(c.Request().Context())

		if id != "" {
			t.Errorf("expected empty request ID in context, got %s", id)
		}

		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
