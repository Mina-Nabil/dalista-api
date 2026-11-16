package notificationServices

import (
	"context"
	"encoding/json"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	notificationModels "dalista/dalista-be/app/repository/models/notificationModels"
	sharingModels "dalista/dalista-be/app/repository/models/sharingModels"
	"dalista/dalista-be/integrations/fcm"

	"github.com/google/uuid"
)

// Notification type constants — the wired push events (DAL-135).
const (
	TypeListShared     = "list_shared"     // a list was shared with a registered user
	TypeItemCommented  = "item_commented"  // a comment was added on a list's item
	TypeAddedToList    = "added_to_list"   // a pending phone-invite auto-attached you to a list
	TypeInviteAccepted = "invite_accepted" // someone you invited by phone joined the list
	TypeRoleChanged    = "role_changed"    // your role on a shared list changed
	TypeNeedToBuy      = "need_to_buy"     // an item was marked "need to buy" on a periodic list
)

// Notify records a notification for a user (inbox) and pushes it to all their
// devices (best-effort). The inbox write is authoritative; push failures are
// logged, and tokens FCM rejects are pruned. Returns an error only if the inbox
// write fails.
func Notify(ctx context.Context, userID uuid.UUID, nType, title, body string, data map[string]string) error {
	if data == nil {
		data = map[string]string{}
	}
	raw, err := json.Marshal(data)
	if err != nil {
		raw = []byte("{}")
	}
	n := &notificationModels.Notification{
		UserID: userID, Type: nType, Title: title, Body: body, Data: string(raw),
	}
	if err := repository.GetDB().WithContext(ctx).Create(n).Error; err != nil {
		return err
	}
	push(ctx, []uuid.UUID{userID}, title, body, data)
	return nil
}

// NotifySafe is Notify with the error swallowed (logged) — for use inside other
// features' mutations, where a notification must never fail the operation.
func NotifySafe(ctx context.Context, userID uuid.UUID, nType, title, body string, data map[string]string) {
	if err := Notify(ctx, userID, nType, title, body, data); err != nil {
		logger.Warning(ctx, "notify failed", logger.F("user_id", userID.String()), logger.F("err", err.Error()))
	}
}

// NotifyListMembers notifies every member of a list (owner + collaborators),
// except exceptUserID (typically the actor). Best-effort.
func NotifyListMembers(ctx context.Context, listID, exceptUserID uuid.UUID, nType, title, body string, data map[string]string) {
	members, err := listMemberIDs(ctx, listID)
	if err != nil {
		logger.Warning(ctx, "notify list members: resolve failed", logger.F("list_id", listID.String()), logger.F("err", err.Error()))
		return
	}
	for _, uid := range members {
		if uid == exceptUserID {
			continue
		}
		NotifySafe(ctx, uid, nType, title, body, data)
	}
}

// ListForUser returns a user's recent notifications, newest first.
func ListForUser(ctx context.Context, userID uuid.UUID, limit int) ([]notificationModels.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var out []notificationModels.Notification
	err := repository.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

// push sends a notification to every device of the given users and prunes any
// tokens FCM reports as invalid. Entirely best-effort.
func push(ctx context.Context, userIDs []uuid.UUID, title, body string, data map[string]string) {
	tokens, err := tokensForUsers(ctx, userIDs)
	if err != nil {
		logger.Warning(ctx, "push: load tokens failed", logger.F("err", err.Error()))
		return
	}
	for i := range tokens {
		err := fcm.Send(ctx, fcm.Message{Token: tokens[i].Token, Title: title, Body: body, Data: data})
		if err == fcm.ErrInvalidToken {
			_ = repository.GetDB().WithContext(ctx).
				Where("id = ?", tokens[i].ID).Delete(&notificationModels.DeviceToken{}).Error
			continue
		}
		if err != nil {
			logger.Warning(ctx, "push: send failed", logger.F("err", err.Error()))
		}
	}
}

// listMemberIDs returns the owner plus all collaborators of a list.
func listMemberIDs(ctx context.Context, listID uuid.UUID) ([]uuid.UUID, error) {
	db := repository.GetDB().WithContext(ctx)

	var list listModels.List
	if err := db.Select("owner_id").Where("id = ?", listID).First(&list).Error; err != nil {
		return nil, err
	}
	var shareUserIDs []uuid.UUID
	if err := db.Model(&sharingModels.ListShare{}).
		Where("list_id = ?", listID).Pluck("user_id", &shareUserIDs).Error; err != nil {
		return nil, err
	}
	ids := append([]uuid.UUID{list.OwnerID}, shareUserIDs...)
	return ids, nil
}
