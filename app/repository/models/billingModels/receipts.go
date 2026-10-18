package billingModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AppleIAPReceipt stores the verified StoreKit transaction for a user, keyed by
// the stable OriginalTransactionID (Apple's cross-renewal identifier).
type AppleIAPReceipt struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID                uuid.UUID  `gorm:"column:user_id;type:uuid;not null"`
	OriginalTransactionID string     `gorm:"column:original_transaction_id;not null;uniqueIndex"`
	TransactionID         *string    `gorm:"column:transaction_id"`
	ProductID             *string    `gorm:"column:product_id"`
	Environment           *string    `gorm:"column:environment"`
	ExpiresAt             *time.Time `gorm:"column:expires_at"`
	RawPayload            string     `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (AppleIAPReceipt) TableName() string { return "apple_iap_receipts" }

func (r *AppleIAPReceipt) BeforeCreate(tx *gorm.DB) error {
	if r.ID == (uuid.UUID{}) {
		r.ID = uuid.New()
	}
	return nil
}

// GooglePlayPurchase stores the verified Play Billing purchase for a user,
// keyed by the purchase token.
type GooglePlayPurchase struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID               uuid.UUID  `gorm:"column:user_id;type:uuid;not null"`
	PurchaseToken        string     `gorm:"column:purchase_token;not null;uniqueIndex"`
	ProductID            *string    `gorm:"column:product_id"`
	OrderID              *string    `gorm:"column:order_id"`
	ExpiryTime           *time.Time `gorm:"column:expiry_time"`
	AcknowledgementState *int       `gorm:"column:acknowledgement_state"`
	RawPayload           string     `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (GooglePlayPurchase) TableName() string { return "google_play_purchases" }

func (p *GooglePlayPurchase) BeforeCreate(tx *gorm.DB) error {
	if p.ID == (uuid.UUID{}) {
		p.ID = uuid.New()
	}
	return nil
}
