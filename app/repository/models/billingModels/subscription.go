// Package billingModels holds the GORM models for billing & subscriptions:
// a provider-agnostic Subscription plus per-provider receipt/purchase records
// and per-user StorageUsage.
package billingModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Provider is the billing rail behind a subscription. v1 populates apple_iap
// and google_play; stripe is a reserved value for a possible future web rail.
type Provider string

const (
	ProviderAppleIAP   Provider = "apple_iap"
	ProviderGooglePlay Provider = "google_play"
	ProviderStripe     Provider = "stripe" // reserved, not used in v1
)

// Tier is the subscription tier. Absence of a Subscription row == free.
type Tier string

const (
	TierFree      Tier = "free"
	TierBasic     Tier = "basic"
	TierUnlimited Tier = "unlimited"
)

// Status mirrors the provider's subscription lifecycle.
type Status string

const (
	StatusActive      Status = "active"
	StatusGracePeriod Status = "grace_period"
	StatusExpired     Status = "expired"
	StatusCanceled    Status = "canceled"
	StatusRevoked     Status = "revoked"
)

// Subscription is the single source of truth for a user's entitlement,
// independent of which provider produced it. Tier/quota enforcement reads only
// this model.
type Subscription struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID              uuid.UUID  `gorm:"column:user_id;type:uuid;not null;uniqueIndex"`
	BillingProvider     Provider   `gorm:"column:billing_provider;type:billing_provider;not null"`
	Tier                Tier       `gorm:"column:tier;type:subscription_tier;not null;default:free"`
	Status              Status     `gorm:"column:status;type:subscription_status;not null;default:active"`
	ProductID           *string    `gorm:"column:product_id"`
	ProviderReferenceID *string    `gorm:"column:provider_reference_id"`
	CurrentPeriodEnd    *time.Time `gorm:"column:current_period_end"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (Subscription) TableName() string { return "subscriptions" }

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == (uuid.UUID{}) {
		s.ID = uuid.New()
	}
	return nil
}

// Entitled reports whether the subscription currently grants its tier (active
// or in a grace period).
func (s *Subscription) Entitled() bool {
	return s.Status == StatusActive || s.Status == StatusGracePeriod
}
