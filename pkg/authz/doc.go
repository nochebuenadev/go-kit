/*
Package authz provides unified types and helpers for authorization and identity management.

It defines the standard Identity structure and provides context-based helpers to propagate
user information across different layers of the application. It also supports bitmask-based
permission checks.

Features:
- Standard Identity structure for user tracking.
- Context-safe propagation of user identity.
- Bitmask-based permission evaluation.
- Pluggable PermissionProvider interface.

Example usage:

	identity := &authz.Identity{UID: "user-123", TenantID: "tenant-456"}
	ctx = authz.SetInContext(ctx, identity)

	// Later in another layer:
	if id, ok := authz.FromContext(ctx); ok {
		// Use id.UID
	}
*/
package authz
