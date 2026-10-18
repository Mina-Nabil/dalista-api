package userModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Provider identifies the auth method behind an AuthIdentity.
type Provider string

const (
	ProviderEmail    Provider = "email"
	ProviderPhone    Provider = "phone"
	ProviderGoogle   Provider = "google"
	ProviderFacebook Provider = "facebook"
)

// AuthIdentity links a User to one authentication provider. A single user may
// have several (e.g. email + google + phone). ProviderSubject is the external
// identifier for that provider: the email address, phone number, Google `sub`,
// or Facebook user id. PasswordHash is set only for the email provider.
type AuthIdentity struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	Provider        Provider  `gorm:"column:provider;type:auth_provider;not null"`
	ProviderSubject string    `gorm:"column:provider_subject;not null"`
	PasswordHash    *string   `gorm:"column:password_hash"`
	CreatedAt       time.Time
}

func (AuthIdentity) TableName() string { return "auth_identities" }

func (a *AuthIdentity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == (uuid.UUID{}) {
		a.ID = uuid.New()
	}
	return nil
}
