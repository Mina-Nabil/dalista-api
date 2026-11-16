package notificationServices

import (
	"context"
	"strings"

	"dalista/dalista-be/app/repository"
	userModels "dalista/dalista-be/app/repository/models/userModels"
	"dalista/dalista-be/integrations/smsMisr"

	"github.com/google/uuid"
)

// AdminSMSResult reports the outcome of an admin SMS broadcast.
type AdminSMSResult struct {
	Requested int `json:"requested"` // user IDs asked for
	Sent      int `json:"sent"`      // messages dispatched (users with a phone)
	Skipped   int `json:"skipped"`   // users not found or without a phone
}

// SendAdminSMS sends a message to each selected user's phone, reusing the SMS
// Misr client (Phase 1). Users without a phone are skipped. Best-effort per
// recipient — a single failure doesn't abort the batch. Admin-gated at the handler.
func SendAdminSMS(ctx context.Context, userIDs []uuid.UUID, message string) (AdminSMSResult, error) {
	message = strings.TrimSpace(message)
	res := AdminSMSResult{Requested: len(userIDs)}
	if message == "" || len(userIDs) == 0 {
		return res, ErrValidation
	}

	var users []userModels.User
	if err := repository.GetDB().WithContext(ctx).
		Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return res, err
	}

	for i := range users {
		if users[i].Phone == nil {
			res.Skipped++
			continue
		}
		phone := strings.TrimSpace(*users[i].Phone)
		if phone == "" {
			res.Skipped++
			continue
		}
		if err := smsMisr.Send(ctx, phone, message); err != nil {
			res.Skipped++
			continue
		}
		res.Sent++
	}
	// Any requested IDs that didn't resolve to a user also count as skipped.
	if missing := res.Requested - (res.Sent + res.Skipped); missing > 0 {
		res.Skipped += missing
	}
	return res, nil
}
