// Package publicListServices implements public lists & discovery: making a list
// public (a single, flippable permission so it can later be gated to a paid
// tier), starring/unstarring with a maintained star-count aggregate, the
// star-sorted discovery feed, and admin engagement analytics.
package publicListServices

// Permission strings for public lists & discovery.
const (
	// PermCreatePublic gates making a list public. This is the single,
	// easily-flippable rule (DAL-114): granting it to RoleUser means "any user
	// can create a public list"; to later restrict to a paid tier, drop it from
	// the user default and grant it per-subscription — no code change.
	PermCreatePublic = "public_lists:create"
	// PermStar lets a user star/unstar public lists.
	PermStar = "public_lists:star"
	// PermDiscover lets a user read the public-list discovery feed.
	PermDiscover = "public_lists:discover"
	// PermReadEngagement is admin-only: the dashboard's engagement analytics.
	PermReadEngagement = "analytics:public_list_engagement"
)

// AllPermissions is every permission string this feature defines.
var AllPermissions = []string{
	PermCreatePublic, PermStar, PermDiscover, PermReadEngagement,
}

// UserPermissions are the public-list permissions a standard user gets by
// default (everything except the admin-only engagement analytics).
var UserPermissions = []string{
	PermCreatePublic, PermStar, PermDiscover,
}
