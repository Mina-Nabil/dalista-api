package billingServices

import (
	"context"
	"time"

	"dalista/dalista-be/app/repository"
	billingModels "dalista/dalista-be/app/repository/models/billingModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// productTier maps store product identifiers to the tier they grant. Extend as
// products are configured in App Store Connect / Play Console. An unknown paid
// product falls back to basic (never silently grants unlimited).
var productTier = map[string]billingModels.Tier{
	"dalista.basic.monthly":     billingModels.TierBasic,
	"dalista.basic.yearly":      billingModels.TierBasic,
	"dalista.unlimited.monthly": billingModels.TierUnlimited,
	"dalista.unlimited.yearly":  billingModels.TierUnlimited,
}

// TierForProduct resolves a product id to its tier (basic for unknown products).
func TierForProduct(productID string) billingModels.Tier {
	if t, ok := productTier[productID]; ok {
		return t
	}
	return billingModels.TierBasic
}

// SubscriptionInput is the verified provider state used to upsert a subscription.
type SubscriptionInput struct {
	UserID              uuid.UUID
	Provider            billingModels.Provider
	Tier                billingModels.Tier
	Status              billingModels.Status
	ProductID           *string
	ProviderReferenceID *string
	CurrentPeriodEnd    *time.Time
}

// UpsertSubscription writes the user's single subscription row from verified
// provider state (one row per user; updated in place on renewal/cancel/refund).
func UpsertSubscription(ctx context.Context, in SubscriptionInput) (*billingModels.Subscription, error) {
	sub := billingModels.Subscription{
		ID:                  uuid.New(),
		UserID:              in.UserID,
		BillingProvider:     in.Provider,
		Tier:                in.Tier,
		Status:              in.Status,
		ProductID:           in.ProductID,
		ProviderReferenceID: in.ProviderReferenceID,
		CurrentPeriodEnd:    in.CurrentPeriodEnd,
	}
	err := repository.GetDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"billing_provider", "tier", "status", "product_id",
			"provider_reference_id", "current_period_end", "updated_at",
		}),
	}).Create(&sub).Error
	if err != nil {
		return nil, err
	}
	return GetSubscription(ctx, in.UserID)
}

// ApplyAppleReceipt upserts the Apple receipt record keyed by original
// transaction id, then syncs the subscription.
func ApplyAppleReceipt(ctx context.Context, userID uuid.UUID, r billingModels.AppleIAPReceipt, status billingModels.Status) (*billingModels.Subscription, error) {
	r.UserID = userID
	if r.ID == (uuid.UUID{}) {
		r.ID = uuid.New()
	}
	db := repository.GetDB().WithContext(ctx)
	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "original_transaction_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"transaction_id", "product_id", "environment", "expires_at", "raw_payload", "updated_at",
		}),
	}).Create(&r).Error; err != nil {
		return nil, err
	}
	tier := billingModels.TierBasic
	if r.ProductID != nil {
		tier = TierForProduct(*r.ProductID)
	}
	return UpsertSubscription(ctx, SubscriptionInput{
		UserID: userID, Provider: billingModels.ProviderAppleIAP,
		Tier: tier, Status: status,
		ProductID: r.ProductID, ProviderReferenceID: &r.OriginalTransactionID,
		CurrentPeriodEnd: r.ExpiresAt,
	})
}

// ApplyGooglePurchase upserts the Play purchase record keyed by purchase token,
// then syncs the subscription.
func ApplyGooglePurchase(ctx context.Context, userID uuid.UUID, p billingModels.GooglePlayPurchase, status billingModels.Status) (*billingModels.Subscription, error) {
	p.UserID = userID
	if p.ID == (uuid.UUID{}) {
		p.ID = uuid.New()
	}
	db := repository.GetDB().WithContext(ctx)
	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "purchase_token"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"product_id", "order_id", "expiry_time", "acknowledgement_state", "raw_payload", "updated_at",
		}),
	}).Create(&p).Error; err != nil {
		return nil, err
	}
	tier := billingModels.TierBasic
	if p.ProductID != nil {
		tier = TierForProduct(*p.ProductID)
	}
	return UpsertSubscription(ctx, SubscriptionInput{
		UserID: userID, Provider: billingModels.ProviderGooglePlay,
		Tier: tier, Status: status,
		ProductID: p.ProductID, ProviderReferenceID: &p.PurchaseToken,
		CurrentPeriodEnd: p.ExpiryTime,
	})
}

// FindUserByAppleOriginalTxn resolves which user a webhook's original
// transaction id belongs to (notifications don't carry our user id).
func FindUserByAppleOriginalTxn(ctx context.Context, originalTxnID string) (*uuid.UUID, error) {
	var r billingModels.AppleIAPReceipt
	err := repository.GetDB().WithContext(ctx).
		Where("original_transaction_id = ?", originalTxnID).First(&r).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &r.UserID, nil
}

// FindUserByGoogleToken resolves which user a purchase token belongs to.
func FindUserByGoogleToken(ctx context.Context, purchaseToken string) (*uuid.UUID, error) {
	var p billingModels.GooglePlayPurchase
	err := repository.GetDB().WithContext(ctx).
		Where("purchase_token = ?", purchaseToken).First(&p).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p.UserID, nil
}
