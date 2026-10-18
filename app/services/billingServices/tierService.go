package billingServices

import (
	"context"
	"errors"

	"dalista/dalista-be/app/repository"
	billingModels "dalista/dalista-be/app/repository/models/billingModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Unlimited is the sentinel for an unbounded limit.
const Unlimited = -1

const gib = int64(1) << 30 // 1 GiB

// TierLimits are the per-tier caps. StorageBytes is the storage quota;
// PerListUserCap is the max collaborators per shared list (consumed by Phase 4
// Sharing). Unlimited (-1) means no bound.
type TierLimits struct {
	StorageBytes   int64 `json:"storage_bytes"`
	PerListUserCap int   `json:"per_list_user_cap"`
}

// tierLimits is the source of truth for tier → limits (DAL-73/74).
//
//	Free: 2 GB / 5 users · Basic: 5 GB / 15 users · Unlimited: 50 GB / ∞ users
var tierLimits = map[billingModels.Tier]TierLimits{
	billingModels.TierFree:      {StorageBytes: 2 * gib, PerListUserCap: 5},
	billingModels.TierBasic:     {StorageBytes: 5 * gib, PerListUserCap: 15},
	billingModels.TierUnlimited: {StorageBytes: 50 * gib, PerListUserCap: Unlimited},
}

// LimitsFor returns the limits for a tier (defaults to free for an unknown tier).
func LimitsFor(tier billingModels.Tier) TierLimits {
	if l, ok := tierLimits[tier]; ok {
		return l
	}
	return tierLimits[billingModels.TierFree]
}

// GetTier returns the user's effective tier, reading through the Subscription
// regardless of billing provider. No row, or a non-entitled subscription
// (expired/canceled/revoked), means free.
func GetTier(ctx context.Context, userID uuid.UUID) (billingModels.Tier, error) {
	sub, err := GetSubscription(ctx, userID)
	if err != nil {
		return "", err
	}
	if sub == nil || !sub.Entitled() {
		return billingModels.TierFree, nil
	}
	return sub.Tier, nil
}

// GetUserLimits returns the user's effective tier and its limits.
func GetUserLimits(ctx context.Context, userID uuid.UUID) (billingModels.Tier, TierLimits, error) {
	tier, err := GetTier(ctx, userID)
	if err != nil {
		return "", TierLimits{}, err
	}
	return tier, LimitsFor(tier), nil
}

// GetSubscription returns the user's subscription row, or (nil, nil) if none.
func GetSubscription(ctx context.Context, userID uuid.UUID) (*billingModels.Subscription, error) {
	var sub billingModels.Subscription
	err := repository.GetDB().WithContext(ctx).Where("user_id = ?", userID).First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}
