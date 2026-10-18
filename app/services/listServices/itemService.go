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

// ItemInput carries the mutable fields for creating/updating an Item. A nil
// CategoryID on create means the item goes into the MISC bucket.
type ItemInput struct {
	Name       string
	CategoryID *uuid.UUID
	Brand      *string
	Price      *float64
	Notes      *string
	Attributes *string // JSON object string; nil leaves unchanged on update
}

// ListItems returns the items of a list the caller can read (owner/editor/viewer).
func ListItems(ctx context.Context, userID, listID uuid.UUID, miscOnly bool) ([]listModels.Item, error) {
	if _, err := getListForRead(ctx, userID, listID); err != nil {
		return nil, err
	}
	q := repository.GetDB().WithContext(ctx).Where("list_id = ?", listID)
	if miscOnly {
		q = q.Where("category_id IS NULL")
	}
	var items []listModels.Item
	err := q.Order("created_at ASC").Find(&items).Error
	return items, err
}

// CreateItem adds an item to a list the caller can mutate (owner/editor). The
// item is owned by the list owner, and its category is validated against the
// list owner's (or system) categories.
func CreateItem(ctx context.Context, userID, listID uuid.UUID, in ItemInput) (*listModels.Item, error) {
	list, err := getListForWrite(ctx, userID, listID)
	if err != nil {
		return nil, err
	}
	if err := categoryUsableBy(ctx, list.OwnerID, in.CategoryID); err != nil {
		return nil, err
	}
	item := &listModels.Item{
		OwnerID:    list.OwnerID,
		ListID:     listID,
		CategoryID: in.CategoryID,
		Name:       strings.TrimSpace(in.Name),
		Brand:      in.Brand,
		Price:      in.Price,
		Notes:      in.Notes,
		Attributes: "{}",
	}
	if in.Attributes != nil {
		item.Attributes = *in.Attributes
	}
	if err := repository.GetDB().WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}
	activityServices.RecordSafe(ctx, listID, userID, activityServices.ActionItemCreated, "item", &item.ID,
		map[string]interface{}{"name": item.Name})
	return item, nil
}

// GetItem returns one item on a list the caller can read.
func GetItem(ctx context.Context, userID, id uuid.UUID) (*listModels.Item, error) {
	item, err := getItemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := getListForRead(ctx, userID, item.ListID); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateItem applies a partial update to an item on a list the caller can mutate.
func UpdateItem(ctx context.Context, userID, id uuid.UUID, in ItemInput, clearCategory bool) (*listModels.Item, error) {
	item, err := getItemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	list, err := getListForWrite(ctx, userID, item.ListID)
	if err != nil {
		return nil, err
	}
	if in.CategoryID != nil {
		if err := categoryUsableBy(ctx, list.OwnerID, in.CategoryID); err != nil {
			return nil, err
		}
	}

	updates := map[string]interface{}{}
	if strings.TrimSpace(in.Name) != "" {
		updates["name"] = strings.TrimSpace(in.Name)
	}
	if clearCategory {
		updates["category_id"] = nil // move to MISC
	} else if in.CategoryID != nil {
		updates["category_id"] = *in.CategoryID
	}
	if in.Brand != nil {
		updates["brand"] = *in.Brand
	}
	if in.Price != nil {
		updates["price"] = *in.Price
	}
	if in.Notes != nil {
		updates["notes"] = *in.Notes
	}
	if in.Attributes != nil {
		updates["attributes"] = *in.Attributes
	}
	if len(updates) > 0 {
		if err := repository.GetDB().WithContext(ctx).Model(item).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	updated, err := getItemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	activityServices.RecordSafe(ctx, updated.ListID, userID, activityServices.ActionItemUpdated, "item", &updated.ID,
		map[string]interface{}{"name": updated.Name})
	return updated, nil
}

// DeleteItem deletes an item on a list the caller can mutate.
func DeleteItem(ctx context.Context, userID, id uuid.UUID) error {
	item, err := getItemByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := getListForWrite(ctx, userID, item.ListID); err != nil {
		return err
	}
	if err := repository.GetDB().WithContext(ctx).Delete(item).Error; err != nil {
		return err
	}
	activityServices.RecordSafe(ctx, item.ListID, userID, activityServices.ActionItemDeleted, "item", &item.ID,
		map[string]interface{}{"name": item.Name})
	return nil
}

func getItemByID(ctx context.Context, id uuid.UUID) (*listModels.Item, error) {
	var item listModels.Item
	err := repository.GetDB().WithContext(ctx).Where("id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
