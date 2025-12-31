package worker

import (
	"context"
	"errors"
	"sync"
	"testing"

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

func TestGetWorker(t *testing.T) {
	cfg := &Config{PoolSize: 2, BufferSize: 5}
	w := GetWorker(cfg, &mockLogger{})
	if w == nil {
		t.Fatal("expected worker component, got nil")
	}

	w2 := GetWorker(cfg, &mockLogger{})
	if w != w2 {
		t.Error("expected singleton instance")
	}
}

func TestWorkerComponent_Lifecycle(t *testing.T) {
	// Reset singleton for test
	instance = nil
	once = sync.Once{}

	cfg := &Config{PoolSize: 2, BufferSize: 5}
	w := GetWorker(cfg, &mockLogger{})

	if err := w.OnInit(); err != nil {
		t.Fatalf("OnInit failed: %v", err)
	}

	if err := w.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	var wg sync.WaitGroup
	var taskExecuted bool
	var mu sync.Mutex

	wg.Add(1)
	task := func(ctx context.Context) error {
		mu.Lock()
		taskExecuted = true
		mu.Unlock()
		wg.Done()
		return nil
	}

	if !w.Dispatch(task) {
		t.Error("expected task to be dispatched")
	}

	wg.Wait()

	mu.Lock()
	if !taskExecuted {
		t.Error("expected task to be executed")
	}
	mu.Unlock()

	if err := w.OnStop(); err != nil {
		t.Fatalf("OnStop failed: %v", err)
	}
}

func TestWorkerComponent_Backpressure(t *testing.T) {
	// Reset singleton for test
	instance = nil
	once = sync.Once{}

	cfg := &Config{PoolSize: 0, BufferSize: 1} // No workers, buffer size 1
	w := GetWorker(cfg, &mockLogger{})

	task := func(ctx context.Context) error { return nil }

	if !w.Dispatch(task) {
		t.Error("expected first task to be dispatched")
	}

	if w.Dispatch(task) {
		t.Error("expected second task to fail due to buffer full")
	}
}

func TestWorkerComponent_ErrorHandling(t *testing.T) {
	// Reset singleton for test
	instance = nil
	once = sync.Once{}

	cfg := &Config{PoolSize: 1, BufferSize: 1}
	w := GetWorker(cfg, &mockLogger{})
	w.OnInit()
	w.OnStart()

	var wg sync.WaitGroup
	wg.Add(1)
	task := func(ctx context.Context) error {
		defer wg.Done()
		return errors.New("task failed")
	}

	w.Dispatch(task)
	wg.Wait()
	// Error is logged, we just ensure it doesn't crash

	w.OnStop()
}
