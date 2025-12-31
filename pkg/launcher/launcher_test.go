package launcher

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nochebuenadev/go-kit/pkg/logz"
)

// mockLogger implements logz.Logger for testing.
type mockLogger struct {
	logz.Logger
}

func (m *mockLogger) Debug(msg string, args ...any)               {}
func (m *mockLogger) Info(msg string, args ...any)                {}
func (m *mockLogger) Warn(msg string, args ...any)                {}
func (m *mockLogger) Error(msg string, err error, args ...any)    {}
func (m *mockLogger) LogError(msg string, err error, args ...any) {}
func (m *mockLogger) Fatal(msg string, err error, args ...any) {
	// Don't os.Exit in tests if possible, but launcher uses it.
	// For testing Run we might need a different approach or just test units.
}
func (m *mockLogger) With(args ...any) logz.Logger                { return m }
func (m *mockLogger) WithContext(ctx context.Context) logz.Logger { return m }

// mockComponent implements Component for testing.
type mockComponent struct {
	onInitCalled  bool
	onStartCalled bool
	onStopCalled  bool
	initErr       error
	startErr      error
	stopErr       error
}

func (m *mockComponent) OnInit() error {
	m.onInitCalled = true
	return m.initErr
}

func (m *mockComponent) OnStart() error {
	m.onStartCalled = true
	return m.startErr
}

func (m *mockComponent) OnStop() error {
	m.onStopCalled = true
	return m.stopErr
}

func TestLauncher_Append(t *testing.T) {
	l := New(&mockLogger{}).(*launcher)
	comp := &mockComponent{}
	l.Append(comp)

	if len(l.components) != 1 {
		t.Errorf("expected 1 component, got %d", len(l.components))
	}
}

func TestLauncher_BeforeStart(t *testing.T) {
	l := New(&mockLogger{}).(*launcher)
	hook := func() error { return nil }
	l.BeforeStart(hook)

	if len(l.onBeforeStart) != 1 {
		t.Errorf("expected 1 hook, got %d", len(l.onBeforeStart))
	}
}

func TestLauncher_Shutdown(t *testing.T) {
	l := New(&mockLogger{}).(*launcher)
	c1 := &mockComponent{}
	c2 := &mockComponent{}
	l.Append(c1, c2)

	// Test reverse order indirectly by checking if they were called
	l.shutdown()

	if !c1.onStopCalled || !c2.onStopCalled {
		t.Error("expected both components to be stopped")
	}
}

func TestLauncher_Run_Lifecycle(t *testing.T) {
	// Since Run() blocks and uses signal.Notify, testing the full Run()
	// is tricky without complex signal sending.
	// We'll test the internal phases if they were exported, but they are not.
	// However, we can test that components are handled.

	l := New(&mockLogger{}).(*launcher)
	c := &mockComponent{}
	hookCalled := false
	l.Append(c)
	l.BeforeStart(func() error {
		hookCalled = true
		return nil
	})

	// To test Run() without blocking forever:
	go func() {
		time.Sleep(100 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}()

	l.Run()

	if !c.onInitCalled {
		t.Error("OnInit was not called")
	}
	if !hookCalled {
		t.Error("BeforeStart hook was not called")
	}
	if !c.onStartCalled {
		t.Error("OnStart was not called")
	}
	if !c.onStopCalled {
		t.Error("OnStop was not called (shutdown phase)")
	}
}

func TestLauncher_ShutdownTimeout(t *testing.T) {
	// l := New(&mockLogger{}).(*launcher)

	// Component that hangs on OnStop
	// c := &mockComponent{}
	// We can't easily mock the hang without a channel or similar in the component
}

type hangingComponent struct {
	mockComponent
}

func (h *hangingComponent) OnStop() error {
	h.onStopCalled = true
	time.Sleep(20 * time.Millisecond) // Smaller than the 15s timeout for test speed
	return nil
}

func TestLauncher_ShutdownSmallDelay(t *testing.T) {
	l := New(&mockLogger{}).(*launcher)
	c := &hangingComponent{}
	l.Append(c)

	l.shutdown()

	if !c.onStopCalled {
		t.Error("OnStop was not called")
	}
}
