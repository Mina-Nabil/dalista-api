package publicListServices

import (
	"context"
	"errors"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	publicListModels "dalista/dalista-be/app/repository/models/publicListModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StarState is the result of a star/unstar: whether the caller now has the list
// starred, and the list's current star count.
type StarState struct {
	Starred   bool
	StarCount int
}

// getPublicList loads a list and requires it to be public. Returns ErrNotFound
// (missing) or ErrNotPublic (private).
func getPublicList(ctx context.Context, listID uuid.UUID) (*listModels.List, error) {
	var list listModels.List
	err := repository.GetDB().WithContext(ctx).Where("id = ?", listID).First(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !list.IsPublic {
		return nil, ErrNotPublic
	}
	return &list, nil
}

// Star stars a public list for the caller (idempotent). The list's star_count
// is incremented only when a new star is actually created.
func Star(ctx context.Context, userID, listID uuid.UUID) (StarState, error) {
	if _, err := getPublicList(ctx, listID); err != nil {
		return StarState{}, err
	}

	err := repository.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		star := &publicListModels.PublicListStar{UserID: userID, ListID: listID}
		res := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "list_id"}},
			DoNothing: true,
		}).Create(star)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 1 {
			return tx.Model(&listModels.List{}).Where("id = ?", listID).
				Update("star_count", gorm.Expr("star_count + 1")).Error
		}
		return nil
	})
	if err != nil {
		return StarState{}, err
	}
	return StarState{Starred: true, StarCount: currentStarCount(ctx, listID)}, nil
}

// Unstar removes the caller's star from a public list (idempotent). The list's
// star_count is decremented (floored at zero) only when a star was removed.
func Unstar(ctx context.Context, userID, listID uuid.UUID) (StarState, error) {
	if _, err := getPublicList(ctx, listID); err != nil {
		return StarState{}, err
	}

	err := repository.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("user_id = ? AND list_id = ?", userID, listID).
			Delete(&publicListModels.PublicListStar{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 1 {
			return tx.Model(&listModels.List{}).Where("id = ?", listID).
				Update("star_count", gorm.Expr("GREATEST(star_count - 1, 0)")).Error
		}
		return nil
	})
	if err != nil {
		return StarState{}, err
	}
	return StarState{Starred: false, StarCount: currentStarCount(ctx, listID)}, nil
}

// currentStarCount reads a list's current star_count (0 on any error, best-effort).
func currentStarCount(ctx context.Context, listID uuid.UUID) int {
	var list listModels.List
	if err := repository.GetDB().WithContext(ctx).
		Select("star_count").Where("id = ?", listID).First(&list).Error; err != nil {
		return 0
	}
	return list.StarCount
}
