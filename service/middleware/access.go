package middleware

// Access represents a single access level, that is used when permissioning API
// endpoints.
type Access int

const (
	// Admin represents admin level access
	Admin Access = iota + 1

	// Authenticated represents user or service account level access.
	Authenticated

	// Anonymous represents unauthenticated access.
	Anonymous
)
