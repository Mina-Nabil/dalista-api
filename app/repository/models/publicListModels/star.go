// Package publicListModels holds the GORM models for public lists &
// discovery: PublicListStar (a user starring a public list). The public/private
// flag and the maintained star_count live on the List model itself
// (listModels.List); this package owns the star join table.
package publicListModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PublicListStar is a user's star on a public list. Unique per (user, list);
// the list's star_count aggregate is maintained alongside inserts/deletes.
type PublicListStar struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	ListID    uuid.UUID `gorm:"column:list_id;type:uuid;not null"`
	CreatedAt time.Time
}

func (PublicListStar) TableName() string { return "public_list_stars" }

func (s *PublicListStar) BeforeCreate(tx *gorm.DB) error {
	if s.ID == (uuid.UUID{}) {
		s.ID = uuid.New()
	}
	return nil
}
