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

// ListCategories returns the caller's own categories plus the shared system
// categories, ordered by name.
func ListCategories(ctx context.Context, userID uuid.UUID) ([]listModels.Category, error) {
	var cats []listModels.Category
	err := repository.GetDB().WithContext(ctx).
		Where("owner_id = ? OR owner_id IS NULL", userID).
		Order("name ASC").Find(&cats).Error
	return cats, err
}

// CreateCategory creates a user-owned category.
func CreateCategory(ctx context.Context, userID uuid.UUID, name string) (*listModels.Category, error) {
	cat := &listModels.Category{OwnerID: &userID, Name: strings.TrimSpace(name)}
	if err := repository.GetDB().WithContext(ctx).Create(cat).Error; err != nil {
		return nil, err
	}
	return cat, nil
}

// UpdateCategory renames a category the caller owns. System categories are not
// visible as owned, so they report ErrNotFound.
func UpdateCategory(ctx context.Context, userID, id uuid.UUID, name string) (*listModels.Category, error) {
	cat, err := getOwnedCategory(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	cat.Name = strings.TrimSpace(name)
	if err := repository.GetDB().WithContext(ctx).Model(cat).Update("name", cat.Name).Error; err != nil {
		return nil, err
	}
	return cat, nil
}

// DeleteCategory deletes a category the caller owns. Items/lists referencing it
// have their category_id set to NULL (ON DELETE SET NULL) → they fall into MISC.
func DeleteCategory(ctx context.Context, userID, id uuid.UUID) error {
	cat, err := getOwnedCategory(ctx, userID, id)
	if err != nil {
		return err
	}
	return repository.GetDB().WithContext(ctx).Delete(cat).Error
}

func getOwnedCategory(ctx context.Context, userID, id uuid.UUID) (*listModels.Category, error) {
	var cat listModels.Category
	err := repository.GetDB().WithContext(ctx).
		Where("id = ? AND owner_id = ?", id, userID).First(&cat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

// categoryUsableBy reports whether the category id is one the user may attach to
// their items/lists: either their own or a system category. Returns ErrValidation
// if not. A nil id is always valid (means MISC / no category).
func categoryUsableBy(ctx context.Context, userID uuid.UUID, id *uuid.UUID) error {
	if id == nil {
		return nil
	}
	var count int64
	err := repository.GetDB().WithContext(ctx).Model(&listModels.Category{}).
		Where("id = ? AND (owner_id = ? OR owner_id IS NULL)", *id, userID).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrValidation
	}
	return nil
}
