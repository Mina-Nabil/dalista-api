// Package notificationModels holds the GORM models for notifications: per-user
// push device tokens (DeviceToken) and a log/inbox of notifications sent to a
// user (Notification). Push delivery uses FCM; the device registry is what the
// mobile app registers its token into.
package notificationModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DevicePlatform is the OS a device token belongs to.
type DevicePlatform string

const (
	PlatformIOS     DevicePlatform = "ios"
	PlatformAndroid DevicePlatform = "android"
	PlatformWeb     DevicePlatform = "web"
)

// DeviceToken is a user's FCM registration token for one device. A token is
// unique across users — if the same device re-registers under a new account,
// the row is reassigned to the new user.
type DeviceToken struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID      `gorm:"column:user_id;type:uuid;not null"`
	Token     string         `gorm:"column:token;not null;uniqueIndex"`
	Platform  DevicePlatform `gorm:"column:platform;type:device_platform;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (DeviceToken) TableName() string { return "device_tokens" }

func (d *DeviceToken) BeforeCreate(tx *gorm.DB) error {
	if d.ID == (uuid.UUID{}) {
		d.ID = uuid.New()
	}
	return nil
}

// Notification is a record of a notification sent (or attempted) to a user. It
// doubles as the backing store for an in-app inbox. Data is a JSON object of
// context (e.g. list_id, item_id) the client can deep-link on.
type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	Type      string    `gorm:"column:type;not null"`
	Title     string    `gorm:"column:title;not null"`
	Body      string    `gorm:"column:body;not null"`
	Data      string    `gorm:"column:data;type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time
}

func (Notification) TableName() string { return "notifications" }

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == (uuid.UUID{}) {
		n.ID = uuid.New()
	}
	return nil
}
