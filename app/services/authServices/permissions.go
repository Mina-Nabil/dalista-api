package authServices

// Permission strings owned by the auth/identity + profile feature. Each feature
// package defines its own constants and AllPermissions slice; the top-level
// permission registry aggregates them (see app/services/permissionRegistry.go).
const (
	PermReadProfile   = "profile:read"
	PermUpdateProfile = "profile:update"
)

// AllPermissions is every permission string this feature defines.
var AllPermissions = []string{
	PermReadProfile,
	PermUpdateProfile,
}
