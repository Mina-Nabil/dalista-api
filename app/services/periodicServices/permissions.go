// Package periodicServices implements periodic-list features: the "need to buy"
// item-state toggle (with auto system comments + a transition log) and the
// "create today's list" flow that snapshots needed items into dated
// ListVariants. Per-record access is gated through listServices (editor+ for
// mutations, read for review/history).
package periodicServices

// Permission strings for periodic-list management. Per-record access (which
// lists a user may toggle/generate on) is enforced via list membership.
const (
	// PermManagePeriodic covers the need-to-buy toggle and today's-list generation.
	PermManagePeriodic = "periodic:manage"
)

// AllPermissions is every permission string this feature defines.
var AllPermissions = []string{PermManagePeriodic}

// UserPermissions are the periodic permissions a standard user gets by default.
var UserPermissions = []string{PermManagePeriodic}
