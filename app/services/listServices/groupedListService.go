package listServices

import (
	"context"
	"errors"
	"strings"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ListGroupedLists returns the caller's grouped lists, newest first.
func ListGroupedLists(ctx context.Context, userID uuid.UUID) ([]listModels.GroupedList, error) {
	var gls []listModels.GroupedList
	err := repository.GetDB().WithContext(ctx).
		Where("owner_id = ?", userID).Order("created_at DESC").Find(&gls).Error
	return gls, err
}

// GetGroupedList returns one grouped list the caller owns.
func GetGroupedList(ctx context.Context, userID, id uuid.UUID) (*listModels.GroupedList, error) {
	return getOwnedGroupedList(ctx, userID, id)
}

// CreateGroupedList creates a grouped list owned by the caller.
func CreateGroupedList(ctx context.Context, userID uuid.UUID, name string) (*listModels.GroupedList, error) {
	gl := &listModels.GroupedList{OwnerID: userID, Name: strings.TrimSpace(name)}
	if err := repository.GetDB().WithContext(ctx).Create(gl).Error; err != nil {
		return nil, err
	}
	return gl, nil
}

// UpdateGroupedList renames a grouped list the caller owns.
func UpdateGroupedList(ctx context.Context, userID, id uuid.UUID, name string) (*listModels.GroupedList, error) {
	gl, err := getOwnedGroupedList(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	gl.Name = strings.TrimSpace(name)
	if err := repository.GetDB().WithContext(ctx).Model(gl).Update("name", gl.Name).Error; err != nil {
		return nil, err
	}
	return gl, nil
}

// DeleteGroupedList deletes a grouped list the caller owns. Nested lists are
// detached (grouped_list_id → NULL), becoming standalone, not deleted.
func DeleteGroupedList(ctx context.Context, userID, id uuid.UUID) error {
	gl, err := getOwnedGroupedList(ctx, userID, id)
	if err != nil {
		return err
	}
	return repository.GetDB().WithContext(ctx).Delete(gl).Error
}

func getOwnedGroupedList(ctx context.Context, userID, id uuid.UUID) (*listModels.GroupedList, error) {
	var gl listModels.GroupedList
	err := repository.GetDB().WithContext(ctx).
		Where("id = ? AND owner_id = ?", id, userID).First(&gl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &gl, nil
}
