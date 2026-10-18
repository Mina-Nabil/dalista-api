package activityModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivityEvent is one entry in the activity log: who (ActorID) did what
// (Action) to which target, on which list, when. Scoped to a list so the
// Timeline feed can be filtered to the lists a user can access. Payload is a
// JSON object of event-specific detail (e.g. the item's name).
type ActivityEvent struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ListID     uuid.UUID  `gorm:"column:list_id;type:uuid;not null"`
	ActorID    *uuid.UUID `gorm:"column:actor_id;type:uuid"`
	Action     string     `gorm:"column:action;not null"`
	TargetType string     `gorm:"column:target_type"`
	TargetID   *uuid.UUID `gorm:"column:target_id;type:uuid"`
	Payload    string     `gorm:"column:payload;type:jsonb;not null;default:'{}'"`
	CreatedAt  time.Time
}

func (ActivityEvent) TableName() string { return "activity_events" }

func (e *ActivityEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == (uuid.UUID{}) {
		e.ID = uuid.New()
	}
	return nil
}
