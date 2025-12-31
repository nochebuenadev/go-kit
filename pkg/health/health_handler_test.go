package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
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

type mockCheck struct {
	name     string
	priority Level
	err      error
}

func (m *mockCheck) HealthCheck(ctx context.Context) error { return m.err }
func (m *mockCheck) Name() string                          { return m.name }
func (m *mockCheck) Priority() Level                       { return m.priority }

func TestHealthHandler_HealthCheck(t *testing.T) {
	e := echo.New()
	logger := &mockLogger{}

	tests := []struct {
		name            string
		checks          []Checkable
		expectedStatus  int
		expectedOverall string
	}{
		{
			name: "all up",
			checks: []Checkable{
				&mockCheck{name: "db", priority: LevelCritical, err: nil},
				&mockCheck{name: "cache", priority: LevelDegraded, err: nil},
			},
			expectedStatus:  http.StatusOK,
			expectedOverall: "UP",
		},
		{
			name: "degraded failure",
			checks: []Checkable{
				&mockCheck{name: "db", priority: LevelCritical, err: nil},
				&mockCheck{name: "cache", priority: LevelDegraded, err: errors.New("ping failed")},
			},
			expectedStatus:  http.StatusOK,
			expectedOverall: "DEGRADED",
		},
		{
			name: "critical failure",
			checks: []Checkable{
				&mockCheck{name: "db", priority: LevelCritical, err: errors.New("conn failed")},
				&mockCheck{name: "cache", priority: LevelDegraded, err: nil},
			},
			expectedStatus:  http.StatusServiceUnavailable,
			expectedOverall: "DOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset singleton for test
			handlerInstance = nil
			handlerOnce = sync.Once{}

			h := GetHandler(logger, tt.checks...)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := h.HealthCheck(c); err != nil {
				t.Fatalf("HealthCheck failed: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var resp Response
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if resp.Status != tt.expectedOverall {
				t.Errorf("expected overall status %s, got %s", tt.expectedOverall, resp.Status)
			}
		})
	}
}
