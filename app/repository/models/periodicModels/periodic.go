// Package periodicModels holds the GORM models for periodic lists: the
// item-state transition log (ItemStateEvent) and dated "today's list"
// snapshots (ListVariant + ListVariantItem).
package periodicModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ItemState is an item's need-to-buy state in a periodic list.
type ItemState string

const (
	StateNeedToBuy ItemState = "need_to_buy"
	StateEnough    ItemState = "enough"
)

// ItemStateEvent logs each need-to-buy/enough transition for an item.
type ItemStateEvent struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ItemID    uuid.UUID  `gorm:"column:item_id;type:uuid;not null"`
	ActorID   *uuid.UUID `gorm:"column:actor_id;type:uuid"`
	State     ItemState  `gorm:"column:state;type:item_state;not null"`
	CreatedAt time.Time
}

func (ItemStateEvent) TableName() string { return "item_state_events" }

func (e *ItemStateEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == (uuid.UUID{}) {
		e.ID = uuid.New()
	}
	return nil
}

// ListVariant is a dated "today's list" snapshot generated from a periodic list.
// One periodic list has many variants (one per generation run).
type ListVariant struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PeriodicListID uuid.UUID  `gorm:"column:periodic_list_id;type:uuid;not null"`
	Title          string     `gorm:"column:title;not null"`
	GeneratedBy    *uuid.UUID `gorm:"column:generated_by;type:uuid"`
	GeneratedAt    time.Time  `gorm:"column:generated_at;not null"`
	CreatedAt      time.Time

	Items []ListVariantItem `gorm:"foreignKey:VariantID"`
}

func (ListVariant) TableName() string { return "list_variants" }

func (v *ListVariant) BeforeCreate(tx *gorm.DB) error {
	if v.ID == (uuid.UUID{}) {
		v.ID = uuid.New()
	}
	return nil
}

// ListVariantItem is a frozen snapshot of an item at generation time, so a
// variant doesn't change when the parent periodic list is later edited.
type ListVariantItem struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	VariantID    uuid.UUID  `gorm:"column:variant_id;type:uuid;not null"`
	SourceItemID *uuid.UUID `gorm:"column:source_item_id;type:uuid"`
	Name         string     `gorm:"column:name;not null"`
	Brand        *string    `gorm:"column:brand"`
	Price        *float64   `gorm:"column:price;type:numeric(12,2)"`
	CategoryID   *uuid.UUID `gorm:"column:category_id;type:uuid"`
	Notes        *string    `gorm:"column:notes"`
	Attributes   string     `gorm:"column:attributes;type:jsonb;not null;default:'{}'"`
	CreatedAt    time.Time
}

func (ListVariantItem) TableName() string { return "list_variant_items" }

func (i *ListVariantItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == (uuid.UUID{}) {
		i.ID = uuid.New()
	}
	return nil
}
