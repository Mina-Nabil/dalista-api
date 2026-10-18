// Package activityModels holds the GORM models for comments and the activity
// timeline: Comment (per-item threads) and ActivityEvent (the chronological
// action log that powers the Timeline view).
package activityModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comment is a message on an item. ListID is denormalized from the item for
// access scoping and timeline queries. IsSystem marks auto-generated comments
// (e.g. the "need to buy" state transitions added in Phase 6); AuthorID is nil
// for system comments.
type Comment struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ItemID    uuid.UUID  `gorm:"column:item_id;type:uuid;not null"`
	ListID    uuid.UUID  `gorm:"column:list_id;type:uuid;not null"`
	AuthorID  *uuid.UUID `gorm:"column:author_id;type:uuid"`
	Body      string     `gorm:"column:body;not null"`
	IsSystem  bool       `gorm:"column:is_system;not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Comment) TableName() string { return "comments" }

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == (uuid.UUID{}) {
		c.ID = uuid.New()
	}
	return nil
}
