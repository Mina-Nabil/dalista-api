package listServices

import (
	"context"
	"errors"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	sharingModels "dalista/dalista-be/app/repository/models/sharingModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Access is a caller's effective access to a list, from none up to owner.
type Access int

const (
	AccessNone Access = iota
	AccessViewer
	AccessEditor
	AccessOwner
)

// accessLevel resolves the caller's access to a list: owner (owns it), editor
// or viewer (shared via list_shares), or none. A missing list returns
// (AccessNone, nil, nil).
func accessLevel(ctx context.Context, userID, listID uuid.UUID) (Access, *listModels.List, error) {
	var list listModels.List
	err := repository.GetDB().WithContext(ctx).Where("id = ?", listID).First(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AccessNone, nil, nil
	}
	if err != nil {
		return AccessNone, nil, err
	}
	if list.OwnerID == userID {
		return AccessOwner, &list, nil
	}

	var share sharingModels.ListShare
	err = repository.GetDB().WithContext(ctx).
		Where("list_id = ? AND user_id = ?", listID, userID).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AccessNone, &list, nil
	}
	if err != nil {
		return AccessNone, nil, err
	}
	if share.Role == sharingModels.RoleEditor {
		return AccessEditor, &list, nil
	}
	return AccessViewer, &list, nil
}

// CanRead reports whether the caller can read the list (owner/editor/viewer).
// Exported for cross-feature access checks (comments, timeline).
func CanRead(ctx context.Context, userID, listID uuid.UUID) (bool, error) {
	lvl, _, err := accessLevel(ctx, userID, listID)
	if err != nil {
		return false, err
	}
	return lvl >= AccessViewer, nil
}

// getListForRead returns the list if the caller can read it (owner/editor/
// viewer), else ErrNotFound.
func getListForRead(ctx context.Context, userID, listID uuid.UUID) (*listModels.List, error) {
	lvl, list, err := accessLevel(ctx, userID, listID)
	if err != nil {
		return nil, err
	}
	if lvl == AccessNone {
		return nil, ErrNotFound
	}
	return list, nil
}

// GetListForWrite returns the list if the caller can mutate it (owner/editor),
// else ErrForbidden (viewer) / ErrNotFound (no access). Exported for
// cross-feature write gating (e.g. periodic-list toggles/generation).
func GetListForWrite(ctx context.Context, userID, listID uuid.UUID) (*listModels.List, error) {
	return getListForWrite(ctx, userID, listID)
}

// getListForWrite returns the list if the caller can mutate it (owner/editor).
// A viewer gets ErrForbidden; no access gets ErrNotFound.
func getListForWrite(ctx context.Context, userID, listID uuid.UUID) (*listModels.List, error) {
	lvl, list, err := accessLevel(ctx, userID, listID)
	if err != nil {
		return nil, err
	}
	switch {
	case lvl == AccessNone:
		return nil, ErrNotFound
	case lvl < AccessEditor:
		return nil, ErrForbidden
	default:
		return list, nil
	}
}
