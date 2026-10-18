package authServices

import (
	"context"

	"dalista/dalista-be/api/server/logger"
	sharingServices "dalista/dalista-be/app/services/sharingServices"

	"github.com/google/uuid"
)

// CheckAndAttachPendingInvites resolves any pending list-share invites addressed
// to a user's freshly-verified phone number into real ListShare records.
//
// Invoked from every phone-verification path (standalone sign-in verify +
// post-signup verify). Best-effort and non-fatal: a failure to attach invites
// must never block sign-in / verification (DAL-40 wired the call site in Phase
// 1; DAL-85 fills in the body here now that the sharing tables exist).
func CheckAndAttachPendingInvites(ctx context.Context, userID uuid.UUID, phone string) {
	attached, err := sharingServices.AttachPendingInvites(ctx, userID, phone)
	if err != nil {
		logger.Error(ctx, err, "pending-invite auto-attach failed",
			logger.F("user_id", userID.String()), logger.F("phone", phone))
		return
	}
	if attached > 0 {
		logger.Info(ctx, "auto-attached pending invites on phone verification",
			logger.F("user_id", userID.String()), logger.F("count", attached))
	}
}
