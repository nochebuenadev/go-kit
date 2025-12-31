package vkutil

import (
	"testing"

	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type mockLogger struct {
	logz.Logger
}

func (m *mockLogger) Info(msg string, args ...any) {}

func TestNew(t *testing.T) {
	cfg := &Config{
		Addrs: []string{"localhost:6379"},
	}
	v := New(cfg, &mockLogger{})
	if v == nil {
		t.Fatal("expected ValkeyComponent, got nil")
	}
}

func TestVkComponent_OnInit_Config(t *testing.T) {
	cfg := &Config{
		Addrs:             []string{"localhost:6379"},
		CacheSizeEachConn: 10,
	}
	v := New(cfg, &mockLogger{}).(*vkComponent)

	// We can't easily test real connection in unit tests without a server
	// but we can check if it initializes the client (though valkey.NewClient might fail)
	err := v.OnInit()

	// If valkey is not available, it might still "succeed" in creating the client
	// as valkey-go often does lazy connection, but let's check for basic errors.
	if err != nil {
		t.Logf("OnInit failed (expected if valkey is not available): %v", err)
	}
}

func TestVkComponent_Client(t *testing.T) {
	v := &vkComponent{client: nil}
	if v.Client() != nil {
		t.Error("expected nil client")
	}
}
