package pgutil

import (
	"context"
	"testing"

	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

// mockLogger is a simple mock for logz.Logger.
type mockLogger struct {
	logz.Logger
}

func (m *mockLogger) Debug(msg string, args ...any)               {}
func (m *mockLogger) Info(msg string, args ...any)                {}
func (m *mockLogger) Warn(msg string, args ...any)                {}
func (m *mockLogger) Error(msg string, err error, args ...any)    {}
func (m *mockLogger) LogError(msg string, err error, args ...any) {}
func (m *mockLogger) Fatal(msg string, err error, args ...any)    {}
func (m *mockLogger) With(args ...any) logz.Logger                { return m }
func (m *mockLogger) WithContext(ctx context.Context) logz.Logger { return m }

func TestGetPostgresClient(t *testing.T) {
	cfg := &Config{Host: "localhost", Port: 5432}
	logger := &mockLogger{}

	c1 := GetPostgresClient(logger, cfg)
	if c1 == nil {
		t.Fatal("expected client, got nil")
	}

	c2 := GetPostgresClient(logger, cfg)
	if c1 != c2 {
		t.Error("expected same singleton instance")
	}
}

func TestDatabaseConfig_GetConnectionString(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "pass",
		Name:     "dbname",
		SSLMode:  "disable",
		Timezone: "UTC",
	}

	expected := "postgres://user:pass@localhost:5432/dbname?sslmode=disable&timezone=UTC"
	if cfg.GetConnectionString() != expected {
		t.Errorf("expected %s, got %s", expected, cfg.GetConnectionString())
	}
}

func TestPostgresClient_HandleError(t *testing.T) {

	t.Run("nil error", func(t *testing.T) {
		if HandleError(nil) != nil {
			t.Error("expected nil, got error")
		}
	})

	t.Run("apperr mapping", func(t *testing.T) {
		// Testing simple internal error mapping
		err := HandleError(context.DeadlineExceeded)
		ae, ok := err.(*apperr.AppErr)
		if !ok {
			t.Fatalf("expected *apperr.AppErr, got %T", err)
		}
		if ae.GetCode() != "INTERNAL_ERROR" {
			t.Errorf("expected INTERNAL_ERROR, got %s", ae.GetCode())
		}
	})
}

func TestErrRow(t *testing.T) {
	expectedErr := apperr.Internal("test error")
	row := &errRow{err: expectedErr}
	var dest string
	err := row.Scan(&dest)
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}
