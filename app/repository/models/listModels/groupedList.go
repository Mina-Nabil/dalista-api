package listModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GroupedList is a home-page container of Lists, owned by a user.
type GroupedList struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID   uuid.UUID `gorm:"column:owner_id;type:uuid;not null"`
	Name      string    `gorm:"column:name;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (GroupedList) TableName() string { return "grouped_lists" }

func (g *GroupedList) BeforeCreate(tx *gorm.DB) error {
	if g.ID == (uuid.UUID{}) {
		g.ID = uuid.New()
	}
	return nil
}
