package listModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ItemImage is a photo attached to an Item. The bytes live in object storage
// (ObjectKey); only metadata is stored in the DB. SizeBytes is counted against
// the owner's storage quota. Position orders the images in the add/edit grid.
type ItemImage struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ItemID      uuid.UUID `gorm:"column:item_id;type:uuid;not null"`
	ObjectKey   string    `gorm:"column:object_key;not null"`
	URL         string    `gorm:"column:url;not null"`
	ContentType string    `gorm:"column:content_type;not null"`
	SizeBytes   int64     `gorm:"column:size_bytes;not null"`
	Position    int       `gorm:"column:position;not null;default:0"`
	CreatedAt   time.Time
}

func (ItemImage) TableName() string { return "item_images" }

func (i *ItemImage) BeforeCreate(tx *gorm.DB) error {
	if i.ID == (uuid.UUID{}) {
		i.ID = uuid.New()
	}
	return nil
}
