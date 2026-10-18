package listServices

import (
	"context"
	"errors"
	"strings"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	activityServices "dalista/dalista-be/app/services/activityServices"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ListInput carries the mutable fields for creating/updating a List.
type ListInput struct {
	Name          string
	GroupedListID *uuid.UUID
	CategoryID    *uuid.UUID
	IsPublic      *bool
	IsPeriodic    *bool
}

// ListLists returns the caller's lists, optionally filtered to a grouped list.
// When groupedListID is nil, all of the caller's lists are returned.
func ListLists(ctx context.Context, userID uuid.UUID, groupedListID *uuid.UUID) ([]listModels.List, error) {
	q := repository.GetDB().WithContext(ctx).Where("owner_id = ?", userID)
	if groupedListID != nil {
		q = q.Where("grouped_list_id = ?", *groupedListID)
	}
	var lists []listModels.List
	err := q.Order("created_at DESC").Find(&lists).Error
	return lists, err
}

// GetList returns one list the caller can read (owner/editor/viewer).
func GetList(ctx context.Context, userID, id uuid.UUID) (*listModels.List, error) {
	return getListForRead(ctx, userID, id)
}

// CreateList creates a list owned by the caller, validating any referenced
// grouped list (must be owned) and category (own or system).
func CreateList(ctx context.Context, userID uuid.UUID, in ListInput) (*listModels.List, error) {
	if err := validateListRefs(ctx, userID, in.GroupedListID, in.CategoryID); err != nil {
		return nil, err
	}
	list := &listModels.List{
		OwnerID:       userID,
		GroupedListID: in.GroupedListID,
		CategoryID:    in.CategoryID,
		Name:          strings.TrimSpace(in.Name),
	}
	if in.IsPublic != nil {
		list.IsPublic = *in.IsPublic
	}
	if in.IsPeriodic != nil {
		list.IsPeriodic = *in.IsPeriodic
	}
	if err := repository.GetDB().WithContext(ctx).Create(list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpdateList applies a partial update to a list the caller can mutate
// (owner/editor). clearGroupedList detaches it to standalone; clearCategory
// removes its category (both take precedence over the corresponding *ID field).
// Grouped-list and category references are validated against the list's owner,
// so an editor can't reparent the owner's list into their own grouped list.
func UpdateList(ctx context.Context, userID, id uuid.UUID, in ListInput, clearGroupedList, clearCategory bool) (*listModels.List, error) {
	list, err := getListForWrite(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := validateListRefs(ctx, list.OwnerID, in.GroupedListID, in.CategoryID); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if strings.TrimSpace(in.Name) != "" {
		updates["name"] = strings.TrimSpace(in.Name)
	}
	if clearGroupedList {
		updates["grouped_list_id"] = nil
	} else if in.GroupedListID != nil {
		updates["grouped_list_id"] = *in.GroupedListID
	}
	if clearCategory {
		updates["category_id"] = nil
	} else if in.CategoryID != nil {
		updates["category_id"] = *in.CategoryID
	}
	if in.IsPublic != nil {
		updates["is_public"] = *in.IsPublic
	}
	if in.IsPeriodic != nil {
		updates["is_periodic"] = *in.IsPeriodic
	}
	if len(updates) > 0 {
		if err := repository.GetDB().WithContext(ctx).Model(list).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	activityServices.RecordSafe(ctx, id, userID, activityServices.ActionListUpdated, "list", &id, nil)
	return getListForRead(ctx, userID, id)
}

// DeleteList deletes a list. Owner-only — editors can mutate contents but not
// delete the list itself.
func DeleteList(ctx context.Context, userID, id uuid.UUID) error {
	list, err := getOwnedList(ctx, userID, id)
	if err != nil {
		return err
	}
	return repository.GetDB().WithContext(ctx).Delete(list).Error
}

func getOwnedList(ctx context.Context, userID, id uuid.UUID) (*listModels.List, error) {
	var list listModels.List
	err := repository.GetDB().WithContext(ctx).
		Where("id = ? AND owner_id = ?", id, userID).First(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// validateListRefs ensures a referenced grouped list is owned by the caller and
// a referenced category is usable by them.
func validateListRefs(ctx context.Context, userID uuid.UUID, groupedListID, categoryID *uuid.UUID) error {
	if groupedListID != nil {
		if _, err := getOwnedGroupedList(ctx, userID, *groupedListID); err != nil {
			if errors.Is(err, ErrNotFound) {
				return ErrValidation
			}
			return err
		}
	}
	return categoryUsableBy(ctx, userID, categoryID)
}
