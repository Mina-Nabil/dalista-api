package listServices

// Permission strings for the core list domain. Ownership is enforced per-record
// in the service layer (owner must match the caller); these strings gate the
// action class. Sharing-based access (editor/viewer on others' lists) arrives
// in Phase 4.
const (
	PermCreateCategory = "categories:create"
	PermReadCategory   = "categories:read"
	PermUpdateCategory = "categories:update"
	PermDeleteCategory = "categories:delete"

	PermCreateGroupedList = "grouped_lists:create"
	PermReadGroupedList   = "grouped_lists:read"
	PermUpdateGroupedList = "grouped_lists:update"
	PermDeleteGroupedList = "grouped_lists:delete"

	PermCreateList = "lists:create"
	PermReadList   = "lists:read"
	PermUpdateList = "lists:update"
	PermDeleteList = "lists:delete"

	PermCreateItem = "items:create"
	PermReadItem   = "items:read"
	PermUpdateItem = "items:update"
	PermDeleteItem = "items:delete"

	// Admin-only: category-interest export for the dashboard.
	PermExportCategoryInterest = "analytics:category_interest_export"
)

// AllPermissions is every permission string this feature defines.
var AllPermissions = []string{
	PermCreateCategory, PermReadCategory, PermUpdateCategory, PermDeleteCategory,
	PermCreateGroupedList, PermReadGroupedList, PermUpdateGroupedList, PermDeleteGroupedList,
	PermCreateList, PermReadList, PermUpdateList, PermDeleteList,
	PermCreateItem, PermReadItem, PermUpdateItem, PermDeleteItem,
	PermExportCategoryInterest,
}

// UserPermissions are the list-domain permissions a standard user receives by
// default (everything except the admin-only export).
var UserPermissions = []string{
	PermCreateCategory, PermReadCategory, PermUpdateCategory, PermDeleteCategory,
	PermCreateGroupedList, PermReadGroupedList, PermUpdateGroupedList, PermDeleteGroupedList,
	PermCreateList, PermReadList, PermUpdateList, PermDeleteList,
	PermCreateItem, PermReadItem, PermUpdateItem, PermDeleteItem,
}
