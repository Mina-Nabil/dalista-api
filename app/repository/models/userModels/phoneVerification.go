package userModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VerificationPurpose distinguishes a standalone phone sign-in challenge from a
// post-signup phone-verification challenge.
type VerificationPurpose string

const (
	PurposeSignin VerificationPurpose = "signin"
	PurposeVerify VerificationPurpose = "verify"
)

// PhoneVerification is a single OTP challenge. CodeHash is a SHA-256 of the
// OTP (never plaintext). UserID is nil for a standalone sign-in challenge (no
// account yet) and set for a post-signup verification.
type PhoneVerification struct {
	ID         uuid.UUID           `gorm:"type:uuid;primaryKey"`
	Phone      string              `gorm:"column:phone;not null"`
	CodeHash   string              `gorm:"column:code_hash;not null"`
	Purpose    VerificationPurpose `gorm:"column:purpose;type:phone_verification_purpose;not null"`
	UserID     *uuid.UUID          `gorm:"column:user_id;type:uuid"`
	Attempts   int                 `gorm:"column:attempts;not null;default:0"`
	ExpiresAt  time.Time           `gorm:"column:expires_at;not null"`
	ConsumedAt *time.Time          `gorm:"column:consumed_at"`
	CreatedAt  time.Time
}

func (PhoneVerification) TableName() string { return "phone_verifications" }

func (p *PhoneVerification) BeforeCreate(tx *gorm.DB) error {
	if p.ID == (uuid.UUID{}) {
		p.ID = uuid.New()
	}
	return nil
}
