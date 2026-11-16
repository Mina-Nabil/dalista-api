// Package comparisonServices implements item comparison: creating a persisted
// ComparisonSession from a manually selected set of item IDs, viewing it with
// per-item access gating (public catalog items and lists the viewer can read),
// and sharing it. Sharing grants no new item access — a recipient only ever
// sees the items they could already access; the rest are withheld.
package comparisonServices

// Permission strings for item comparison.
const (
	PermCreateComparison = "comparisons:create"
	PermReadComparison   = "comparisons:read"
	PermShareComparison  = "comparisons:share"
)

// AllPermissions is every permission string this feature defines.
var AllPermissions = []string{
	PermCreateComparison, PermReadComparison, PermShareComparison,
}

// UserPermissions are the comparison permissions a standard user gets by default.
var UserPermissions = []string{
	PermCreateComparison, PermReadComparison, PermShareComparison,
}
