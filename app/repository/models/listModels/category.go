// Package listModels holds the GORM models for the core list domain:
// Category, GroupedList, List, and Item.
package listModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category is a taxonomy node for organizing items and lists. OwnerID is nil
// for seeded system categories (IsSystem = true), which are visible to everyone.
type Category struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OwnerID   *uuid.UUID `gorm:"column:owner_id;type:uuid"`
	Name      string     `gorm:"column:name;not null"`
	IsSystem  bool       `gorm:"column:is_system;not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Category) TableName() string { return "categories" }

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == (uuid.UUID{}) {
		c.ID = uuid.New()
	}
	return nil
}
