// Package services hosts the cross-feature permission registry. Each feature
// package (authServices, listServices, …) owns its permission-string constants
// and an AllPermissions slice; this file aggregates them and defines the
// per-role defaults the `syncpermissions` command writes to the database.
package services

import (
	userModels "dalista/dalista-be/app/repository/models/userModels"
	"dalista/dalista-be/app/services/activityServices"
	"dalista/dalista-be/app/services/authServices"
	"dalista/dalista-be/app/services/billingServices"
	"dalista/dalista-be/app/services/listServices"
	"dalista/dalista-be/app/services/periodicServices"
	"dalista/dalista-be/app/services/sharingServices"
)

// AllPermissions is the union of every feature's permission strings. Append a
// feature's slice here as each module lands.
func AllPermissions() []string {
	var all []string
	all = append(all, authServices.AllPermissions...)
	all = append(all, listServices.AllPermissions...)
	all = append(all, billingServices.AllPermissions...)
	all = append(all, sharingServices.AllPermissions...)
	all = append(all, activityServices.AllPermissions...)
	all = append(all, periodicServices.AllPermissions...)
	// Phase 7+: publicListServices, … register here.
	return all
}

// concat flattens several permission slices into one.
func concat(slices ...[]string) []string {
	var out []string
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}

// RoleDefaults maps each role to the permissions it receives by default.
func RoleDefaults() map[userModels.Role][]string {
	return map[userModels.Role][]string{
		// Admins get everything.
		userModels.RoleAdmin: AllPermissions(),

		// Standard users: own profile + full CRUD over their own list domain.
		// The admin-only category-interest export is excluded. Sharing-based
		// access to others' lists arrives in Phase 4.
		userModels.RoleUser: concat(
			[]string{authServices.PermReadProfile, authServices.PermUpdateProfile},
			listServices.UserPermissions,
			billingServices.UserPermissions,
			sharingServices.UserPermissions,
			activityServices.UserPermissions,
			periodicServices.UserPermissions,
		),
	}
}
