package authz

import "context"

type (
	// Identity represents the authenticated user's information.
	Identity struct {
		UID         string `json:"uid"`
		TenantID    string `json:"tenant_id"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
	}

	// PermissionProvider defines the interface for resolving user permissions.
	PermissionProvider interface {
		// ResolveMask returns the permission mask for a user in a specific tenant and app.
		ResolveMask(ctx context.Context, uid, tenantID, appID string) (int64, error)
	}

	// authContextKey is a private type for storing the identity in the context.
	authContextKey struct{}
)

var (
	// authKey is the unique key used to store/retrieve the Identity from the context.
	authKey = authContextKey{}
)

// HasPermission checks if a specific bit is set in the permission mask.
// The bit must be between 0 and 62.
func (i *Identity) HasPermission(mask, bit int64) bool {
	if bit < 0 || bit >= 63 {
		return false
	}
	return (mask & (1 << uint(bit))) != 0
}

// FromContext retrieves the Identity from the context.
func FromContext(ctx context.Context) (*Identity, bool) {
	id, ok := ctx.Value(authKey).(*Identity)
	return id, ok
}

// SetInContext injects the Identity into the context.
func SetInContext(ctx context.Context, id *Identity) context.Context {
	return context.WithValue(ctx, authKey, id)
}
