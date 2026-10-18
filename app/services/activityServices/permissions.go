// Package activityServices implements per-item comments and the activity
// timeline. Comments are writable by any list member (including viewers, per
// the Phase 4 role rules); the timeline is an access-scoped feed of
// ActivityEvents. Access scoping is enforced by the handler via
// listServices.CanRead / listServices.GetItem, so this package stays free of a
// listServices import (listServices records events through activityServices.Record).
package activityServices

// Permission strings for comments + activity. The per-record access (which
// lists a user can comment on / see in the timeline) is enforced via list
// membership, not these strings.
const (
	PermCreateComment = "comments:create"
	PermReadComment   = "comments:read"
	PermUpdateComment = "comments:update"
	PermDeleteComment = "comments:delete"
	PermReadActivity  = "activity:read"
)

// AllPermissions is every permission string this feature defines.
var AllPermissions = []string{
	PermCreateComment, PermReadComment, PermUpdateComment, PermDeleteComment, PermReadActivity,
}

// UserPermissions are the activity permissions a standard user gets by default.
var UserPermissions = AllPermissions
