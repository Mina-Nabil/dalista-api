package billingServices

// Permission strings for billing & subscriptions.
const (
	// PermReadSelfBilling lets a user read their own subscription + storage usage.
	PermReadSelfBilling = "billing:read_self"
	// PermReadSubscriptions is the admin oversight permission for the dashboard.
	PermReadSubscriptions = "subscriptions:read"
)

// AllPermissions is every permission string this feature defines.
var AllPermissions = []string{
	PermReadSelfBilling,
	PermReadSubscriptions,
}

// UserPermissions are the billing permissions a standard user gets by default
// (reading their own state; the admin oversight permission is excluded).
var UserPermissions = []string{
	PermReadSelfBilling,
}
