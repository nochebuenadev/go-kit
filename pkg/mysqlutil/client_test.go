package mysqlutil

import (
	"context"
	"testing"

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

func TestGetMySQLClient(t *testing.T) {
	cfg := &Config{Host: "localhost", Port: 3306}
	logger := &mockLogger{}

	c1 := GetMySQLClient(logger, cfg)
	if c1 == nil {
		t.Fatal("expected client, got nil")
	}

	c2 := GetMySQLClient(logger, cfg)
	if c1 != c2 {
		t.Error("expected same singleton instance")
	}
}

func TestConfig_GetConnectionString(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     3306,
		User:     "user",
		Password: "pass",
		Name:     "dbname",
	}

	expected := "user:pass@tcp(localhost:3306)/dbname?parseTime=true"
	if cfg.GetConnectionString() != expected {
		t.Errorf("expected %s, got %s", expected, cfg.GetConnectionString())
	}
}
