package periodicServices

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	periodicModels "dalista/dalista-be/app/repository/models/periodicModels"
	activityServices "dalista/dalista-be/app/services/activityServices"
	listServices "dalista/dalista-be/app/services/listServices"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ToggleNeedToBuy sets an item's need-to-buy state on a periodic list (editor+).
// On a real change it records an ItemStateEvent and an auto system comment
// ("Marked as need to buy" / "We have enough").
func ToggleNeedToBuy(ctx context.Context, userID, itemID uuid.UUID, need bool) (*listModels.Item, error) {
	item, err := listServices.GetItem(ctx, userID, itemID) // read access + resolves the item
	if err != nil {
		return nil, err
	}
	list, err := listServices.GetListForWrite(ctx, userID, item.ListID) // editor+
	if err != nil {
		return nil, err
	}
	if !list.IsPeriodic {
		return nil, ErrNotPeriodic
	}
	if item.NeedToBuy == need {
		return item, nil // no-op; no event/comment
	}

	db := repository.GetDB().WithContext(ctx)
	if err := db.Model(&listModels.Item{}).Where("id = ?", itemID).Update("need_to_buy", need).Error; err != nil {
		return nil, err
	}

	state := periodicModels.StateEnough
	body := "We have enough"
	if need {
		state = periodicModels.StateNeedToBuy
		body = "Marked as need to buy"
	}
	_ = db.Create(&periodicModels.ItemStateEvent{ItemID: itemID, ActorID: &userID, State: state}).Error
	if _, cerr := activityServices.AddSystemComment(ctx, itemID, list.ID, body); cerr != nil {
		// Non-fatal: the state change already persisted.
		_ = cerr
	}
	activityServices.RecordSafe(ctx, list.ID, userID, activityServices.ActionItemUpdated, "item", &itemID,
		map[string]interface{}{"name": item.Name, "need_to_buy": need})

	item.NeedToBuy = need
	return item, nil
}

// TodayReview returns the periodic list's items currently marked need-to-buy —
// the candidate set for a "today's list" (read access).
func TodayReview(ctx context.Context, userID, listID uuid.UUID) (*listModels.List, []listModels.Item, error) {
	list, err := listServices.GetList(ctx, userID, listID)
	if err != nil {
		return nil, nil, err
	}
	if !list.IsPeriodic {
		return nil, nil, ErrNotPeriodic
	}
	var items []listModels.Item
	err = repository.GetDB().WithContext(ctx).
		Where("list_id = ? AND need_to_buy = TRUE", listID).Order("created_at ASC").Find(&items).Error
	return list, items, err
}

// GenerateToday creates a dated ListVariant from the periodic list's current
// need-to-buy items, excluding any in excludeIDs (which are marked "enough" as
// a side effect, with the usual state event + system comment). Editor+.
func GenerateToday(ctx context.Context, userID, listID uuid.UUID, excludeIDs []uuid.UUID) (*periodicModels.ListVariant, error) {
	list, err := listServices.GetListForWrite(ctx, userID, listID)
	if err != nil {
		return nil, err
	}
	if !list.IsPeriodic {
		return nil, ErrNotPeriodic
	}

	var candidates []listModels.Item
	if err := repository.GetDB().WithContext(ctx).
		Where("list_id = ? AND need_to_buy = TRUE", listID).Order("created_at ASC").Find(&candidates).Error; err != nil {
		return nil, err
	}

	excluded := make(map[uuid.UUID]bool, len(excludeIDs))
	for _, id := range excludeIDs {
		excluded[id] = true
	}

	// Exclude = mark "enough" (with the normal transition side effects).
	for _, it := range candidates {
		if excluded[it.ID] {
			if _, err := ToggleNeedToBuy(ctx, userID, it.ID, false); err != nil {
				return nil, err
			}
		}
	}

	variant := &periodicModels.ListVariant{
		PeriodicListID: listID,
		Title:          fmt.Sprintf("%s — %s", list.Name, time.Now().Format("2006-01-02")),
		GeneratedBy:    &userID,
		GeneratedAt:    time.Now(),
	}

	err = repository.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(variant).Error; err != nil {
			return err
		}
		for _, it := range candidates {
			if excluded[it.ID] {
				continue
			}
			src := it.ID
			snap := &periodicModels.ListVariantItem{
				VariantID:    variant.ID,
				SourceItemID: &src,
				Name:         it.Name,
				Brand:        it.Brand,
				Price:        it.Price,
				CategoryID:   it.CategoryID,
				Notes:        it.Notes,
				Attributes:   it.Attributes,
			}
			if err := tx.Create(snap).Error; err != nil {
				return err
			}
			variant.Items = append(variant.Items, *snap)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	activityServices.RecordSafe(ctx, listID, userID, "variant_created", "list_variant", &variant.ID,
		map[string]interface{}{"title": variant.Title, "item_count": len(variant.Items)})
	return variant, nil
}

// ListVariants returns a periodic list's variants, newest first (read access).
func ListVariants(ctx context.Context, userID, listID uuid.UUID) ([]periodicModels.ListVariant, error) {
	if _, err := listServices.GetList(ctx, userID, listID); err != nil {
		return nil, err
	}
	var variants []periodicModels.ListVariant
	err := repository.GetDB().WithContext(ctx).Preload("Items").
		Where("periodic_list_id = ?", listID).Order("generated_at DESC").Find(&variants).Error
	return variants, err
}

// GetVariant returns one variant with its snapshot items (read access to the
// parent periodic list).
func GetVariant(ctx context.Context, userID, variantID uuid.UUID) (*periodicModels.ListVariant, error) {
	var variant periodicModels.ListVariant
	err := repository.GetDB().WithContext(ctx).Preload("Items").Where("id = ?", variantID).First(&variant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if _, err := listServices.GetList(ctx, userID, variant.PeriodicListID); err != nil {
		return nil, err // no access to the parent list → ErrNotFound from listServices
	}
	return &variant, nil
}
