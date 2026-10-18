package activityServices

import (
	"context"
	"encoding/json"
	"time"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/app/repository"
	activityModels "dalista/dalista-be/app/repository/models/activityModels"

	"github.com/google/uuid"
)

// Activity action constants (ActivityEvent.Action).
const (
	ActionItemCreated  = "item_created"
	ActionItemUpdated  = "item_updated"
	ActionItemDeleted  = "item_deleted"
	ActionListUpdated  = "list_updated"
	ActionCommentAdded = "comment_added"
)

// Record writes an activity event. Best-effort: callers (mutation hooks) should
// log and ignore a non-nil error so activity logging never fails the mutation.
func Record(ctx context.Context, listID uuid.UUID, actorID *uuid.UUID, action, targetType string, targetID *uuid.UUID, payload map[string]interface{}) error {
	body := "{}"
	if payload != nil {
		if raw, err := json.Marshal(payload); err == nil {
			body = string(raw)
		}
	}
	ev := &activityModels.ActivityEvent{
		ListID:     listID,
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Payload:    body,
	}
	return repository.GetDB().WithContext(ctx).Create(ev).Error
}

// RecordSafe is Record with error handling folded in — logs on failure. Use
// from mutation hooks where the activity write must not break the operation.
func RecordSafe(ctx context.Context, listID uuid.UUID, actorID uuid.UUID, action, targetType string, targetID *uuid.UUID, payload map[string]interface{}) {
	if err := Record(ctx, listID, &actorID, action, targetType, targetID, payload); err != nil {
		logger.Error(ctx, err, "failed to record activity event",
			logger.F("list_id", listID.String()), logger.F("action", action))
	}
}

// Timeline returns the access-scoped activity feed for a user: events on lists
// they own or are shared on, newest first. `before` paginates (events strictly
// older than it); limit is clamped to [1, 100].
func Timeline(ctx context.Context, userID uuid.UUID, limit int, before *time.Time) ([]activityModels.ActivityEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q := repository.GetDB().WithContext(ctx).
		Where(`list_id IN (
			SELECT id FROM lists WHERE owner_id = ?
			UNION
			SELECT list_id FROM list_shares WHERE user_id = ?
		)`, userID, userID)
	if before != nil {
		q = q.Where("created_at < ?", *before)
	}
	var events []activityModels.ActivityEvent
	err := q.Order("created_at DESC").Limit(limit).Find(&events).Error
	return events, err
}
