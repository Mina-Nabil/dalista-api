package listModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Item belongs to a List. A nil CategoryID means the item sits in the system
// MISC bucket (unassigned). OwnerID is denormalized from the list for fast
// ownership checks. Attributes is a JSON object of custom comparison fields;
// brand and price are first-class because the comparison feature keys on them.
type Item struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OwnerID    uuid.UUID  `gorm:"column:owner_id;type:uuid;not null"`
	ListID     uuid.UUID  `gorm:"column:list_id;type:uuid;not null"`
	CategoryID *uuid.UUID `gorm:"column:category_id;type:uuid"`
	Name       string     `gorm:"column:name;not null"`
	Brand      *string    `gorm:"column:brand"`
	Price      *float64   `gorm:"column:price;type:numeric(12,2)"`
	Notes      *string    `gorm:"column:notes"`
	Attributes string     `gorm:"column:attributes;type:jsonb;not null;default:'{}'"`
	// NeedToBuy is the periodic-list "need to buy" state (Phase 6); false for
	// standard lists and for sufficiently-stocked items.
	NeedToBuy bool `gorm:"column:need_to_buy;not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Item) TableName() string { return "items" }

func (i *Item) BeforeCreate(tx *gorm.DB) error {
	if i.ID == (uuid.UUID{}) {
		i.ID = uuid.New()
	}
	return nil
}
