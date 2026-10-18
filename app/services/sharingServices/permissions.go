// Package sharingServices implements list sharing: sharing a list with a phone
// number (editor/viewer), pending invites for unregistered numbers, auto-attach
// on phone verification, per-list tier caps, and collaborator management. The
// per-record editor/viewer access checks used by the list/item endpoints live
// in listServices (access.go); this package owns the share lifecycle.
package sharingServices

// Permission strings for sharing management. The editor/viewer access checks
// themselves are per-record (resolved from list_shares), not permission strings
// — a shared editor is an ordinary user whose list/item permissions already
// allow mutation; the per-record check gates which lists they apply to.
const (
	// PermShareList lets a user manage shares on lists they own.
	PermShareList = "lists:share"
)

// AllPermissions is every permission string this feature defines.
var AllPermissions = []string{PermShareList}

// UserPermissions are the sharing permissions a standard user gets by default.
var UserPermissions = []string{PermShareList}
