package activityServices

import (
	"context"
	"errors"
	"strings"

	"dalista/dalista-be/app/repository"
	activityModels "dalista/dalista-be/app/repository/models/activityModels"
	listModels "dalista/dalista-be/app/repository/models/listModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateComment adds a comment to an item and records a comment_added activity
// event. Access (that the author is a member of the item's list) is enforced by
// the caller before this is invoked. itemID/listID come from the resolved item.
func CreateComment(ctx context.Context, authorID, itemID, listID uuid.UUID, body string) (*activityModels.Comment, error) {
	comment := &activityModels.Comment{
		ItemID:   itemID,
		ListID:   listID,
		AuthorID: &authorID,
		Body:     strings.TrimSpace(body),
	}
	if err := repository.GetDB().WithContext(ctx).Create(comment).Error; err != nil {
		return nil, err
	}
	RecordSafe(ctx, listID, authorID, ActionCommentAdded, "item", &itemID, map[string]interface{}{
		"comment_id": comment.ID.String(),
	})
	return comment, nil
}

// AddSystemComment records an auto-generated system comment on an item (used by
// Phase 6 "need to buy" state transitions). No activity event; author is nil.
func AddSystemComment(ctx context.Context, itemID, listID uuid.UUID, body string) (*activityModels.Comment, error) {
	comment := &activityModels.Comment{ItemID: itemID, ListID: listID, Body: body, IsSystem: true}
	if err := repository.GetDB().WithContext(ctx).Create(comment).Error; err != nil {
		return nil, err
	}
	return comment, nil
}

// ListComments returns an item's comments, oldest first.
func ListComments(ctx context.Context, itemID uuid.UUID) ([]activityModels.Comment, error) {
	var comments []activityModels.Comment
	err := repository.GetDB().WithContext(ctx).
		Where("item_id = ?", itemID).Order("created_at ASC").Find(&comments).Error
	return comments, err
}

// UpdateComment edits a comment's body. Author-only; system comments can't be edited.
func UpdateComment(ctx context.Context, userID, commentID uuid.UUID, body string) (*activityModels.Comment, error) {
	comment, err := getComment(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if comment.IsSystem || comment.AuthorID == nil || *comment.AuthorID != userID {
		return nil, ErrForbidden
	}
	comment.Body = strings.TrimSpace(body)
	if err := repository.GetDB().WithContext(ctx).Model(comment).Update("body", comment.Body).Error; err != nil {
		return nil, err
	}
	return comment, nil
}

// DeleteComment removes a comment. Allowed for the comment's author or the
// list's owner.
func DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	comment, err := getComment(ctx, commentID)
	if err != nil {
		return err
	}
	allowed := comment.AuthorID != nil && *comment.AuthorID == userID
	if !allowed {
		var list listModels.List
		if lerr := repository.GetDB().WithContext(ctx).Where("id = ?", comment.ListID).First(&list).Error; lerr == nil {
			allowed = list.OwnerID == userID
		}
	}
	if !allowed {
		return ErrForbidden
	}
	return repository.GetDB().WithContext(ctx).Delete(comment).Error
}

func getComment(ctx context.Context, commentID uuid.UUID) (*activityModels.Comment, error) {
	var comment activityModels.Comment
	err := repository.GetDB().WithContext(ctx).Where("id = ?", commentID).First(&comment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
