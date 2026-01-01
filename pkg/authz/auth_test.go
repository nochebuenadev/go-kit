package authz

import (
	"context"
	"testing"
)

func TestIdentity_HasPermission(t *testing.T) {
	id := &Identity{}

	tests := []struct {
		name     string
		mask     int64
		bit      int64
		expected bool
	}{
		{"bit set", 1, 0, true},
		{"bit 4 set", 16, 4, true},
		{"multi bit set", 17, 0, true},
		{"multi bit 4 set", 17, 4, true},
		{"bit not set", 16, 0, false},
		{"out of range low", 1, -1, false},
		{"out of range high", 1, 63, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := id.HasPermission(tt.mask, tt.bit); got != tt.expected {
				t.Errorf("HasPermission(%d, %d) = %v; want %v", tt.mask, tt.bit, got, tt.expected)
			}
		})
	}
}

func TestContextHelpers(t *testing.T) {
	id := &Identity{UID: "user-1"}
	ctx := context.Background()

	ctx = SetInContext(ctx, id)
	retrieved, ok := FromContext(ctx)

	if !ok {
		t.Fatal("expected identity in context")
	}
	if retrieved.UID != id.UID {
		t.Errorf("expected UID %s, got %s", id.UID, retrieved.UID)
	}

	_, ok = FromContext(context.Background())
	if ok {
		t.Error("expected no identity in empty context")
	}
}
