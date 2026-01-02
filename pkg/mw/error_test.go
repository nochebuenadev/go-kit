package mw

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...any)               {}
func (m *mockLogger) Info(msg string, args ...any)                {}
func (m *mockLogger) Warn(msg string, args ...any)                {}
func (m *mockLogger) Error(msg string, err error, args ...any)    {}
func (m *mockLogger) LogError(msg string, err error, args ...any) {}
func (m *mockLogger) Fatal(msg string, err error, args ...any)    {}
func (m *mockLogger) With(args ...any) logz.Logger                { return m }
func (m *mockLogger) WithContext(ctx context.Context) logz.Logger { return m }

func TestAppErrorHandler(t *testing.T) {
	e := echo.New()
	handler := AppErrorHandler(&mockLogger{})

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "apperr InvalidInput",
			err:            apperr.New(apperr.ErrInvalidInput, "bad request"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_ARGUMENT",
		},
		{
			name:           "apperr NotFound",
			err:            apperr.New(apperr.ErrResourceNotFound, "not found"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "standard error",
			err:            errors.New("generic error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
		{
			name:           "echo HTTPError",
			err:            echo.NewHTTPError(http.StatusTeapot, "i'm a teapot"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_ARGUMENT",
		},
		{
			name:           "echo HTTPError Not Found",
			err:            echo.NewHTTPError(http.StatusNotFound, "resource not found"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "echo HTTPError Internal",
			err:            echo.NewHTTPError(http.StatusInternalServerError, "internal server error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler(tt.err, c)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var respBody map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &respBody); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// AppErr.GetCode returns the code as string
			if respBody["code"] != tt.expectedCode {
				t.Errorf("expected code %s, got %v", tt.expectedCode, respBody["code"])
			}
		})
	}

	t.Run("committed response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Response().Committed = true

		handler(errors.New("should not be handled"), c)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d (no change), got %d", http.StatusOK, rec.Code)
		}
	})
}

func TestMapHTTPStatusToAppErr(t *testing.T) {
	tests := []struct {
		status   int
		expected apperr.ErrorCode
	}{
		{http.StatusNotFound, apperr.ErrResourceNotFound},
		{http.StatusUnauthorized, apperr.ErrUnauthorized},
		{http.StatusForbidden, apperr.ErrPermissionDenied},
		{http.StatusBadRequest, apperr.ErrInvalidInput},
		{http.StatusConflict, apperr.ErrConflict},
		{http.StatusInternalServerError, apperr.ErrInternal},
		{http.StatusServiceUnavailable, apperr.ErrUnavailable},
		{http.StatusGatewayTimeout, apperr.ErrDeadlineExceeded},
		{http.StatusMethodNotAllowed, apperr.ErrNotImplemented},
		{http.StatusRequestTimeout, apperr.ErrDeadlineExceeded},
		{http.StatusNotImplemented, apperr.ErrNotImplemented},
		{http.StatusTeapot, apperr.ErrInvalidInput},
		{http.StatusBadGateway, apperr.ErrInternal}, // default for 5xx
		{503, apperr.ErrUnavailable},
	}

	for _, tt := range tests {
		he := echo.NewHTTPError(tt.status, "message")
		ae := mapHTTPStatusToAppErr(he)
		if ae.GetCode() != string(tt.expected) {
			t.Errorf("status %d: expected %s, got %s", tt.status, tt.expected, ae.GetCode())
		}
	}
}

func TestMapAppErrToHTTPStatus(t *testing.T) {
	tests := []struct {
		code     apperr.ErrorCode
		expected int
	}{
		{apperr.ErrInvalidInput, http.StatusBadRequest},
		{apperr.ErrResourceNotFound, http.StatusNotFound},
		{apperr.ErrConflict, http.StatusConflict},
		{apperr.ErrUnauthorized, http.StatusUnauthorized},
		{apperr.ErrPermissionDenied, http.StatusForbidden},
		{apperr.ErrInternal, http.StatusInternalServerError},
		{apperr.ErrNotImplemented, http.StatusNotImplemented},
		{apperr.ErrUnavailable, http.StatusServiceUnavailable},
		{apperr.ErrDeadlineExceeded, http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		ae := apperr.New(tt.code, "message")
		status := mapAppErrToHTTPStatus(ae)
		if status != tt.expected {
			t.Errorf("code %s: expected %d, got %d", tt.code, tt.expected, status)
		}
	}
}

func TestMGetRequestHelpers(t *testing.T) {
	e := echo.New()

	t.Run("with request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		c := e.NewContext(req, nil)

		if method := mGetRequestMethod(c); method != http.MethodPost {
			t.Errorf("expected POST, got %s", method)
		}
		// echo.Context.RequestURI might be empty in test if not set manually
		req.RequestURI = "/test"
		if uri := mGetRequestURI(c); uri != "/test" {
			t.Errorf("expected /test, got %s", uri)
		}
	})

	t.Run("without request", func(t *testing.T) {
		c := e.NewContext(nil, nil)
		if method := mGetRequestMethod(c); method != "UNKNOWN" {
			t.Errorf("expected UNKNOWN, got %s", method)
		}
		if uri := mGetRequestURI(c); uri != "UNKNOWN" {
			t.Errorf("expected UNKNOWN, got %s", uri)
		}
	})
}
