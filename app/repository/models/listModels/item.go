package listModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Item belongs to a List. A nil CategoryID means the item sits in the system
// MISC bucket (unassigned). OwnerID is denormalized from the list for fast
// ownership checks. Brand and price are first-class because the comparison
// feature keys on them; the remaining optional attributes live in a dedicated
// 1:1 item_attributes side table (see ItemAttributes).
type Item struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OwnerID    uuid.UUID  `gorm:"column:owner_id;type:uuid;not null"`
	ListID     uuid.UUID  `gorm:"column:list_id;type:uuid;not null"`
	CategoryID *uuid.UUID `gorm:"column:category_id;type:uuid"`
	Name       string     `gorm:"column:name;not null"`
	Brand      *string    `gorm:"column:brand"`
	Price      *float64   `gorm:"column:price;type:numeric(12,2)"`
	Notes      *string    `gorm:"column:notes"`
	// NeedToBuy is the periodic-list "need to buy" state (Phase 6); false for
	// standard lists and for sufficiently-stocked items.
	NeedToBuy bool `gorm:"column:need_to_buy;not null;default:false"`
	// IsBought is a generic todo-style "bought / included" checkbox, used on
	// today's lists and other non-periodic lists.
	IsBought bool `gorm:"column:is_bought;not null;default:false"`
	// Quantity is the periodic-list "count" — how many of this item to buy when
	// it's marked need-to-buy. Defaults to 1; carried into generated today's
	// lists and variant snapshots.
	Quantity  int `gorm:"column:quantity;not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// Attributes is the optional 1:1 side record (url/location/socials/phone +
	// custom extras). Nil when the item has no attributes row. Loaded via
	// Preload("Attributes").
	Attributes *ItemAttributes `gorm:"foreignKey:ItemID;references:ID"`
	// Images are the item's photos, ordered by Position. Loaded via
	// Preload("Images").
	Images []ItemImage `gorm:"foreignKey:ItemID;references:ID"`
}

func (Item) TableName() string { return "items" }

func (i *Item) BeforeCreate(tx *gorm.DB) error {
	if i.ID == (uuid.UUID{}) {
		i.ID = uuid.New()
	}
	return nil
}

// ItemAttributes is a 1:1 side table holding an item's optional, well-known
// attributes as typed columns, plus an `extra` jsonb for arbitrary custom
// key/values (the XD "Add Other" row). One row per item at most.
type ItemAttributes struct {
	ItemID    uuid.UUID `gorm:"column:item_id;type:uuid;primaryKey"`
	URL       *string   `gorm:"column:url"`
	Location  *string   `gorm:"column:location"`
	Facebook  *string   `gorm:"column:facebook"`
	Instagram *string   `gorm:"column:instagram"`
	Phone     *string   `gorm:"column:phone"`
	// Extra is a JSON object string of custom key/values; defaults to "{}".
	Extra     string    `gorm:"column:extra;type:jsonb;not null;default:'{}'"`
	UpdatedAt time.Time
}

func (ItemAttributes) TableName() string { return "item_attributes" }
