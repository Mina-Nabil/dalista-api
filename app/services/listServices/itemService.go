package listServices

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	activityServices "dalista/dalista-be/app/services/activityServices"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// orderImages is a GORM preload scope that returns an item's images in grid order.
func orderImages(db *gorm.DB) *gorm.DB {
	return db.Order("position ASC, created_at ASC")
}

// FlattenAttributes serializes an item's typed attributes plus its custom extras
// into a single flat JSON object string (well-known keys alongside the extras),
// or "{}" when there are none. Used by the legacy variant-snapshot path, whose
// list_variant_items table still stores attributes as a flat jsonb blob.
func FlattenAttributes(a *listModels.ItemAttributes) string {
	m := map[string]interface{}{}
	if a != nil {
		if a.Extra != "" {
			_ = json.Unmarshal([]byte(a.Extra), &m)
		}
		if a.URL != nil {
			m["url"] = *a.URL
		}
		if a.Location != nil {
			m["location"] = *a.Location
		}
		if a.Facebook != nil {
			m["facebook"] = *a.Facebook
		}
		if a.Instagram != nil {
			m["instagram"] = *a.Instagram
		}
		if a.Phone != nil {
			m["phone"] = *a.Phone
		}
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

// CopyItemAttributes writes the source attributes (if any) as a new
// item_attributes row for toItemID. No-op when src is nil/empty.
func CopyItemAttributes(ctx context.Context, toItemID uuid.UUID, src *listModels.ItemAttributes) error {
	if src == nil {
		return nil
	}
	cols := map[string]interface{}{}
	if src.URL != nil {
		cols["url"] = *src.URL
	}
	if src.Location != nil {
		cols["location"] = *src.Location
	}
	if src.Facebook != nil {
		cols["facebook"] = *src.Facebook
	}
	if src.Instagram != nil {
		cols["instagram"] = *src.Instagram
	}
	if src.Phone != nil {
		cols["phone"] = *src.Phone
	}
	if src.Extra != "" && src.Extra != "{}" {
		cols["extra"] = src.Extra
	}
	return upsertItemAttributes(ctx, toItemID, cols)
}

// ItemInput carries the mutable fields for creating/updating an Item. A nil
// CategoryID on create means the item goes into the MISC bucket.
//
// The attribute fields follow per-field update semantics: a nil pointer leaves
// that column unchanged, and a non-nil pointer sets it (an empty/blank string
// clears the column to NULL). Extra is a JSON object string; nil leaves the
// custom extras unchanged, non-nil replaces them.
type ItemInput struct {
	Name       string
	CategoryID *uuid.UUID
	Brand      *string
	Price      *float64
	Notes      *string
	URL        *string
	Location   *string
	Facebook   *string
	Instagram  *string
	Phone      *string
	Extra      *string
}

// attributeColumns returns the item_attributes columns to write for this input,
// applying per-field semantics (nil = skip, blank = NULL). Returns an empty map
// when no attribute field was provided.
func (in ItemInput) attributeColumns() map[string]interface{} {
	cols := map[string]interface{}{}
	put := func(col string, p *string) {
		if p == nil {
			return
		}
		if v := strings.TrimSpace(*p); v != "" {
			cols[col] = v
		} else {
			cols[col] = nil // explicit clear
		}
	}
	put("url", in.URL)
	put("location", in.Location)
	put("facebook", in.Facebook)
	put("instagram", in.Instagram)
	put("phone", in.Phone)
	if in.Extra != nil {
		cols["extra"] = *in.Extra
	}
	return cols
}

// upsertItemAttributes writes the provided attribute columns for an item,
// creating the 1:1 item_attributes row if absent and updating only the provided
// columns otherwise. No-op when nothing was provided.
func upsertItemAttributes(ctx context.Context, itemID uuid.UUID, cols map[string]interface{}) error {
	if len(cols) == 0 {
		return nil
	}
	assign := make([]string, 0, len(cols)+1)
	for k := range cols {
		assign = append(assign, k)
	}
	assign = append(assign, "updated_at")
	row := map[string]interface{}{"item_id": itemID, "updated_at": time.Now()}
	for k, v := range cols {
		row[k] = v
	}
	return repository.GetDB().WithContext(ctx).
		Table("item_attributes").
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "item_id"}},
			DoUpdates: clause.AssignmentColumns(assign),
		}).
		Create(row).Error
}

// ListItems returns the items of a list the caller can read (owner/editor/viewer).
func ListItems(ctx context.Context, userID, listID uuid.UUID, miscOnly bool) ([]listModels.Item, error) {
	if _, err := getListForRead(ctx, userID, listID); err != nil {
		return nil, err
	}
	q := repository.GetDB().WithContext(ctx).Preload("Attributes").Preload("Images", orderImages).Where("list_id = ?", listID)
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
	}
	if err := repository.GetDB().WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}
	if err := upsertItemAttributes(ctx, item.ID, in.attributeColumns()); err != nil {
		return nil, err
	}
	activityServices.RecordSafe(ctx, listID, userID, activityServices.ActionItemCreated, "item", &item.ID,
		map[string]interface{}{"name": item.Name})
	return getItemByID(ctx, item.ID)
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

// SetBought toggles an item's generic "bought / included" checkbox (todo-style,
// used on today's lists and other non-periodic lists). Editor+ on the item's
// list. No-op if already in the requested state.
func SetBought(ctx context.Context, userID, id uuid.UUID, bought bool) (*listModels.Item, error) {
	item, err := getItemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := getListForWrite(ctx, userID, item.ListID); err != nil {
		return nil, err
	}
	if item.IsBought == bought {
		return item, nil
	}
	if err := repository.GetDB().WithContext(ctx).Model(&listModels.Item{}).
		Where("id = ?", id).Update("is_bought", bought).Error; err != nil {
		return nil, err
	}
	activityServices.RecordSafe(ctx, item.ListID, userID, activityServices.ActionItemUpdated, "item", &id,
		map[string]interface{}{"name": item.Name, "is_bought": bought})
	item.IsBought = bought
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
	if len(updates) > 0 {
		if err := repository.GetDB().WithContext(ctx).Model(item).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	if err := upsertItemAttributes(ctx, id, in.attributeColumns()); err != nil {
		return nil, err
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
	err := repository.GetDB().WithContext(ctx).Preload("Attributes").Preload("Images", orderImages).Where("id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
