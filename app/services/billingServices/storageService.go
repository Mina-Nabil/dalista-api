package billingServices

import (
	"context"
	"errors"

	"dalista/dalista-be/app/repository"
	billingModels "dalista/dalista-be/app/repository/models/billingModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrQuotaExceeded is returned when an operation would exceed the tier's storage
// quota. Handlers map it to 413 Payload Too Large.
var ErrQuotaExceeded = errors.New("storage quota exceeded")

// StorageStatus is the user-facing view of storage consumption vs. quota.
type StorageStatus struct {
	Tier       billingModels.Tier `json:"tier"`
	BytesUsed  int64              `json:"bytes_used"`
	QuotaBytes int64              `json:"quota_bytes"` // -1 when unlimited (not used today; max tier is 50 GB)
	Available  int64              `json:"available_bytes"`
}

// GetUsageBytes returns the user's consumed bytes (0 if no row yet).
func GetUsageBytes(ctx context.Context, userID uuid.UUID) (int64, error) {
	var su billingModels.StorageUsage
	err := repository.GetDB().WithContext(ctx).Where("user_id = ?", userID).First(&su).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return su.BytesUsed, nil
}

// GetStatus returns consumption vs. the tier quota.
func GetStatus(ctx context.Context, userID uuid.UUID) (StorageStatus, error) {
	tier, limits, err := GetUserLimits(ctx, userID)
	if err != nil {
		return StorageStatus{}, err
	}
	used, err := GetUsageBytes(ctx, userID)
	if err != nil {
		return StorageStatus{}, err
	}
	available := limits.StorageBytes - used
	if limits.StorageBytes == Unlimited {
		available = Unlimited
	} else if available < 0 {
		available = 0
	}
	return StorageStatus{Tier: tier, BytesUsed: used, QuotaBytes: limits.StorageBytes, Available: available}, nil
}

// AddUsage records a positive or negative delta against the user's usage,
// enforcing the tier quota on increases. Returns ErrQuotaExceeded (without
// writing) when an increase would exceed the quota. Floors usage at zero.
func AddUsage(ctx context.Context, userID uuid.UUID, deltaBytes int64) error {
	_, limits, err := GetUserLimits(ctx, userID)
	if err != nil {
		return err
	}
	used, err := GetUsageBytes(ctx, userID)
	if err != nil {
		return err
	}
	newTotal := used + deltaBytes
	if newTotal < 0 {
		newTotal = 0
	}
	if deltaBytes > 0 && limits.StorageBytes != Unlimited && newTotal > limits.StorageBytes {
		return ErrQuotaExceeded
	}
	return setUsage(ctx, userID, newTotal)
}

// CanStore reports whether additionalBytes would fit within the quota.
func CanStore(ctx context.Context, userID uuid.UUID, additionalBytes int64) (bool, error) {
	_, limits, err := GetUserLimits(ctx, userID)
	if err != nil {
		return false, err
	}
	if limits.StorageBytes == Unlimited {
		return true, nil
	}
	used, err := GetUsageBytes(ctx, userID)
	if err != nil {
		return false, err
	}
	return used+additionalBytes <= limits.StorageBytes, nil
}

func setUsage(ctx context.Context, userID uuid.UUID, bytes int64) error {
	su := billingModels.StorageUsage{UserID: userID, BytesUsed: bytes}
	return repository.GetDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"bytes_used", "updated_at"}),
	}).Create(&su).Error
}
