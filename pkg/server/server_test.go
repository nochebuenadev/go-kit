package server

import (
	"testing"

	"github.com/labstack/echo/v4"
)

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...any)               {}
func (m *mockLogger) Info(msg string, args ...any)                {}
func (m *mockLogger) Warn(msg string, args ...any)                {}
func (m *mockLogger) Error(msg string, err error, args ...any)    {}
func (m *mockLogger) LogError(msg string, err error, args ...any) {}
func (m *mockLogger) Fatal(msg string, err error, args ...any)    {}
func (m *mockLogger) With(args ...any) interface{}                { return m } // Simplified for mock
func (m *mockLogger) WithContext(ctx interface{}) interface{}     { return m } // Simplified for mock

// Actually need to match logz.Logger interface exactly if I want to use it
type fullMockLogger struct{}

func (m *fullMockLogger) Debug(msg string, args ...any)               {}
func (m *fullMockLogger) Info(msg string, args ...any)                {}
func (m *fullMockLogger) Warn(msg string, args ...any)                {}
func (m *fullMockLogger) Error(msg string, err error, args ...any)    {}
func (m *fullMockLogger) LogError(msg string, err error, args ...any) {}
func (m *fullMockLogger) Fatal(msg string, err error, args ...any)    {}
func (m *fullMockLogger) With(args ...any) any                        { return m }
func (m *fullMockLogger) WithContext(ctx any) any                     { return m }

// The above is still not quite right because logz.Logger is an interface.
// I'll just use the mock I used in pgutil but adapt it to logz.Logger interface.

func TestGetEchoServer(t *testing.T) {
	cfg := &Config{
		Port:           8080,
		AllowedOrigins: []string{"*"},
	}
	// Using a nil logger for simplicity in this specific test if possible,
	// or I should implement a proper mock.

	srv := GetEchoServer(cfg, nil)
	if srv == nil {
		t.Fatal("expected server, got nil")
	}

	srv2 := GetEchoServer(cfg, nil)
	if srv != srv2 {
		t.Error("expected same singleton instance")
	}
}

func TestEchoServer_OnInit(t *testing.T) {
	cfg := &Config{
		Port:           8080,
		AllowedOrigins: []string{"*"},
	}
	srv := GetEchoServer(cfg, nil).(*echoServer)

	err := srv.OnInit()
	if err != nil {
		t.Fatalf("OnInit failed: %v", err)
	}

	if srv.instance.HideBanner != true {
		t.Error("expected HideBanner to be true")
	}
}

func TestEchoServer_Registry(t *testing.T) {
	cfg := &Config{Port: 8080, AllowedOrigins: []string{"*"}}
	srv := GetEchoServer(cfg, nil)

	called := false
	srv.Registry(func(e *echo.Echo) {
		called = true
	})

	if !called {
		t.Error("registry function was not called")
	}
}

func TestEchoServer_Group(t *testing.T) {
	cfg := &Config{Port: 8080, AllowedOrigins: []string{"*"}}
	srv := GetEchoServer(cfg, nil)

	g := srv.Group("/api")
	if g == nil {
		t.Fatal("expected group, got nil")
	}
}
