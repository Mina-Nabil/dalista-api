package authServices

import (
	"context"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/app/repository"
	userModels "dalista/dalista-be/app/repository/models/userModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoadUserPermissions returns the deduplicated union of role-default and
// per-user permissions. Called once per request by the auth middleware.
func LoadUserPermissions(ctx context.Context, userID uuid.UUID, role userModels.Role) ([]string, error) {
	var perms []string
	err := repository.GetDB().WithContext(ctx).Raw(`
		SELECT permission FROM role_permissions WHERE role = ?
		UNION
		SELECT permission FROM user_permissions WHERE user_id = ?
	`, string(role), userID).Scan(&perms).Error
	return perms, err
}

// SyncPermissions upserts the permissions reference table and role_permissions
// from the Go-defined registry, removing stale rows. Idempotent — safe to run
// repeatedly (invoked by the `syncpermissions` command).
func SyncPermissions(allPerms []string, roleDefaults map[userModels.Role][]string) error {
	db := repository.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Upsert the canonical permission list, pruning removed ones.
		for _, p := range allPerms {
			row := userModels.Permission{Permission: p}
			if err := tx.Where("permission = ?", p).FirstOrCreate(&row).Error; err != nil {
				return err
			}
		}
		if len(allPerms) > 0 {
			if err := tx.Where("permission NOT IN ?", allPerms).Delete(&userModels.Permission{}).Error; err != nil {
				return err
			}
		}

		// 2. Upsert role_permissions, pruning stale grants per role.
		for role, perms := range roleDefaults {
			for _, p := range perms {
				row := userModels.RolePermission{Role: string(role), Permission: p}
				if err := tx.Where("role = ? AND permission = ?", string(role), p).FirstOrCreate(&row).Error; err != nil {
					return err
				}
			}
			if len(perms) > 0 {
				if err := tx.Where("role = ? AND permission NOT IN ?", string(role), perms).
					Delete(&userModels.RolePermission{}).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Where("role = ?", string(role)).Delete(&userModels.RolePermission{}).Error; err != nil {
					return err
				}
			}
		}

		logger.InfoNoCtx("[sync-permissions] synced", logger.F("permissions", len(allPerms)), logger.F("roles", len(roleDefaults)))
		return nil
	})
}
