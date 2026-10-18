package sharingServices

import (
	"context"
	"strings"
	"time"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/app/repository"
	sharingModels "dalista/dalista-be/app/repository/models/sharingModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AttachPendingInvites resolves any open pending invites addressed to the given
// (now verified) phone into real ListShare rows for the user, marking the
// invites consumed. Called from the phone-verification paths (DAL-85). Returns
// the number of invites attached. Best-effort: safe to call repeatedly.
func AttachPendingInvites(ctx context.Context, userID uuid.UUID, phone string) (int, error) {
	phone = strings.TrimSpace(phone)
	db := repository.GetDB().WithContext(ctx)

	var invites []sharingModels.PendingInvite
	if err := db.Where("phone = ? AND consumed_at IS NULL", phone).Find(&invites).Error; err != nil {
		return 0, err
	}
	if len(invites) == 0 {
		return 0, nil
	}

	attached := 0
	now := time.Now()
	for _, inv := range invites {
		err := db.Transaction(func(tx *gorm.DB) error {
			share := &sharingModels.ListShare{
				ListID:    inv.ListID,
				UserID:    userID,
				Role:      inv.Role,
				InvitedBy: inv.InvitedBy,
			}
			// Idempotent: if a share already exists for (list,user), keep it and
			// just update the role to the invited one.
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "list_id"}, {Name: "user_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"role", "updated_at"}),
			}).Create(share).Error; err != nil {
				return err
			}
			return tx.Model(&sharingModels.PendingInvite{}).
				Where("id = ?", inv.ID).Update("consumed_at", &now).Error
		})
		if err != nil {
			logger.Error(ctx, err, "failed to attach pending invite",
				logger.F("invite_id", inv.ID.String()), logger.F("user_id", userID.String()))
			continue // best-effort; don't block sign-in on one bad invite
		}
		attached++
	}
	return attached, nil
}
