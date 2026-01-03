package httputil

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nochebuenadev/go-kit/pkg/logz"
	"github.com/sony/gobreaker"
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

func TestHttpClient_Resilience(t *testing.T) {
	logger := &mockLogger{}
	cfg := &Config{
		MaxRetries:  2,
		RetryDelay:  10 * time.Millisecond,
		CBThreshold: 3,
		CBTimeout:   100 * time.Millisecond,
		Timeout:     1 * time.Second,
		DialTimeout: 100 * time.Millisecond,
	}

	t.Run("successful request with retries", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"ok"}`)
		}))
		defer server.Close()

		client := &httpClient{
			client: server.Client(),
			logger: logger,
			cfg:    cfg,
			cb:     gobreaker.NewCircuitBreaker(gobreaker.Settings{Name: "test-cb-resilience"}),
		}

		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		// Note: attempts will be 2 because we reset the singleton later maybe?
		// Singleton makes it tricky. I'll use a fresh instance if possible.
	})

	t.Run("circuit breaker stays open", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		// Re-init client to reset CB or use a high failure count
		client := GetClient(logger, cfg)
		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

		// Trigger CB
		for i := 0; i < 4; i++ {
			_, _ = client.Do(req)
		}

		_, err := client.Do(req)
		if err == nil {
			t.Fatal("expected error from open circuit breaker, got nil")
		}
		if err.Error() != "INTERNAL_ERROR: servicio externo no disponible (circuito abierto) → circuit breaker is open" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestDoJSON(t *testing.T) {
	logger := &mockLogger{}
	cfg := DefaultConfig()

	type Response struct {
		Name string `json:"name"`
	}

	t.Run("success decoding", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"name":"test"}`)
		}))
		defer server.Close()

		// Use a fresh client for isolation
		client := &httpClient{
			client: server.Client(),
			logger: logger,
			cfg:    cfg,
			cb:     gobreaker.NewCircuitBreaker(gobreaker.Settings{Name: "test-cb"}),
		}

		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		data, err := DoJSON[Response](context.Background(), client, req)

		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if data.Name != "test" {
			t.Errorf("expected test, got %s", data.Name)
		}
	})
}
