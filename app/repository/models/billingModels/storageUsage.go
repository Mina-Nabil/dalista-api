package billingModels

import (
	"time"

	"github.com/google/uuid"
)

// StorageUsage tracks a user's consumed storage in bytes, enforced against the
// quota of their current tier.
type StorageUsage struct {
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey"`
	BytesUsed int64     `gorm:"column:bytes_used;not null;default:0"`
	UpdatedAt time.Time
}

func (StorageUsage) TableName() string { return "storage_usage" }
