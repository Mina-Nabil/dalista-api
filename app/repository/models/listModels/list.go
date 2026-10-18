package listModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// List contains Items. It may be standalone or nested under a GroupedList, and
// carries public/private and periodic/standard flags plus an optional Category.
type List struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OwnerID       uuid.UUID  `gorm:"column:owner_id;type:uuid;not null"`
	GroupedListID *uuid.UUID `gorm:"column:grouped_list_id;type:uuid"`
	CategoryID    *uuid.UUID `gorm:"column:category_id;type:uuid"`
	Name          string     `gorm:"column:name;not null"`
	IsPublic      bool       `gorm:"column:is_public;not null;default:false"`
	IsPeriodic    bool       `gorm:"column:is_periodic;not null;default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (List) TableName() string { return "lists" }

func (l *List) BeforeCreate(tx *gorm.DB) error {
	if l.ID == (uuid.UUID{}) {
		l.ID = uuid.New()
	}
	return nil
}
