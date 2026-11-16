package listServices

import (
	"context"
	"errors"
	"fmt"
	"io"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	billingServices "dalista/dalista-be/app/services/billingServices"
	"dalista/dalista-be/integrations/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Item image limits.
const (
	MaxImagesPerItem = 6
	MaxImageBytes    = 10 << 20 // 10 MiB
)

// allowedImageTypes maps an accepted content type to its file extension.
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/heic": ".heic",
	"image/heif": ".heif",
	"image/gif":  ".gif",
}

// ListItemImages returns an item's images (ordered) for a caller who can read
// the item's list.
func ListItemImages(ctx context.Context, userID, itemID uuid.UUID) ([]listModels.ItemImage, error) {
	item, err := getItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if _, err := getListForRead(ctx, userID, item.ListID); err != nil {
		return nil, err
	}
	return imagesForItem(ctx, itemID)
}

// AddItemImage stores an uploaded image for an item (editor/owner), counting its
// size against the item owner's storage quota. Returns ErrUnsupportedMediaType,
// ErrFileTooLarge, ErrTooManyImages, or billingServices.ErrQuotaExceeded on the
// respective rejections.
func AddItemImage(ctx context.Context, userID, itemID uuid.UUID, contentType string, size int64, r io.Reader) (*listModels.ItemImage, error) {
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return nil, ErrUnsupportedMediaType
	}
	if size <= 0 || size > MaxImageBytes {
		return nil, ErrFileTooLarge
	}

	item, err := getItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	// Editor/owner on the item's list.
	if _, err := getListForWrite(ctx, userID, item.ListID); err != nil {
		return nil, err
	}

	// Enforce the per-item image cap.
	var count int64
	if err := repository.GetDB().WithContext(ctx).Model(&listModels.ItemImage{}).
		Where("item_id = ?", itemID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count >= MaxImagesPerItem {
		return nil, ErrTooManyImages
	}

	// Reserve quota first (fails fast when over the tier limit). Rolled back if a
	// later step fails.
	if err := billingServices.AddUsage(ctx, item.OwnerID, size); err != nil {
		return nil, err
	}

	imageID := uuid.New()
	objectKey := fmt.Sprintf("items/%s/%s%s", itemID, imageID, ext)
	store := storage.Get()
	if err := store.Put(ctx, objectKey, r, size, contentType); err != nil {
		_ = billingServices.AddUsage(ctx, item.OwnerID, -size) // roll back quota
		return nil, err
	}

	img := &listModels.ItemImage{
		ID:          imageID,
		ItemID:      itemID,
		ObjectKey:   objectKey,
		URL:         store.URL(objectKey),
		ContentType: contentType,
		SizeBytes:   size,
		Position:    int(count),
	}
	if err := repository.GetDB().WithContext(ctx).Create(img).Error; err != nil {
		_ = store.Delete(ctx, objectKey)
		_ = billingServices.AddUsage(ctx, item.OwnerID, -size)
		return nil, err
	}
	return img, nil
}

// DeleteItemImage removes an item image (editor/owner): deletes the stored
// object, the row, and releases the quota.
func DeleteItemImage(ctx context.Context, userID, itemID, imageID uuid.UUID) error {
	item, err := getItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	if _, err := getListForWrite(ctx, userID, item.ListID); err != nil {
		return err
	}

	var img listModels.ItemImage
	err = repository.GetDB().WithContext(ctx).
		Where("id = ? AND item_id = ?", imageID, itemID).First(&img).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if err := repository.GetDB().WithContext(ctx).Delete(&img).Error; err != nil {
		return err
	}
	// Best-effort object cleanup + quota release; the row (source of truth) is gone.
	_ = storage.Get().Delete(ctx, img.ObjectKey)
	_ = billingServices.AddUsage(ctx, item.OwnerID, -img.SizeBytes)
	return nil
}

func imagesForItem(ctx context.Context, itemID uuid.UUID) ([]listModels.ItemImage, error) {
	var images []listModels.ItemImage
	err := repository.GetDB().WithContext(ctx).
		Where("item_id = ?", itemID).
		Order("position ASC, created_at ASC").
		Find(&images).Error
	return images, err
}
