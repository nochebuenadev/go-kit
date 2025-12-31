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
			expectedStatus: http.StatusInternalServerError, // Standardized to internal in current implementation
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
}
