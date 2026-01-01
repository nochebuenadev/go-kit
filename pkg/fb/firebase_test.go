package fb

import (
	"testing"

	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type mockLogger struct {
	logz.Logger
}

func (m *mockLogger) Info(msg string, args ...any) {}

func TestGetFirebase(t *testing.T) {
	cfg := &Config{ProjectID: "test-project"}
	f := GetFirebase(&mockLogger{}, cfg)
	if f == nil {
		t.Fatal("expected firebase component, got nil")
	}

	f2 := GetFirebase(&mockLogger{}, cfg)
	if f != f2 {
		t.Error("expected singleton instance")
	}
}

func TestFirebaseComponent_App(t *testing.T) {
	f := &firebaseComponent{app: nil}
	if f.App() != nil {
		t.Error("expected nil app")
	}
}
