// Package comparisonModels holds the GORM model for item comparison:
// ComparisonSession, a persisted, shareable set of manually selected item IDs.
// It is persisted (not stateless) so it can be shared; the items it resolves to
// are access-checked per viewer on every read (see comparisonServices).
package comparisonModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ComparisonSession is a saved comparison: an ordered set of item IDs selected
// by the owner. ItemIDs is a JSONB array of UUID strings (preserving the user's
// selection order); ShareToken is nil until the session is shared.
type ComparisonSession struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID    uuid.UUID `gorm:"column:owner_id;type:uuid;not null"`
	ItemIDs    string    `gorm:"column:item_ids;type:jsonb;not null;default:'[]'"`
	ShareToken *string   `gorm:"column:share_token"`
	CreatedAt  time.Time
}

func (ComparisonSession) TableName() string { return "comparison_sessions" }

func (s *ComparisonSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == (uuid.UUID{}) {
		s.ID = uuid.New()
	}
	return nil
}
