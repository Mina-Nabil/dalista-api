package userModels

import (
	"time"

	"github.com/google/uuid"
)

// UserProfile holds display/settings data for a user (Module 2). Tier/
// subscription status is read-through from Billing (Phase 3), not stored here.
type UserProfile struct {
	UserID            uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey"`
	DisplayName       *string   `gorm:"column:display_name"`
	PictureURL        *string   `gorm:"column:picture_url"`
	NotificationPrefs string    `gorm:"column:notification_prefs;type:jsonb;not null;default:'{}'"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (UserProfile) TableName() string { return "user_profiles" }
