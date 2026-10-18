// Package sharingModels holds the GORM models for list sharing: ListShare
// (a list shared with a registered user) and PendingInvite (a share addressed
// to a phone number with no account yet, auto-attached on verification).
package sharingModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role is a collaborator's access level on a shared list.
type Role string

const (
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// ListShare grants a registered user a role on a list. Unique per (list, user).
type ListShare struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ListID    uuid.UUID `gorm:"column:list_id;type:uuid;not null"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	Role      Role      `gorm:"column:role;type:share_role;not null"`
	InvitedBy uuid.UUID `gorm:"column:invited_by;type:uuid;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ListShare) TableName() string { return "list_shares" }

func (s *ListShare) BeforeCreate(tx *gorm.DB) error {
	if s.ID == (uuid.UUID{}) {
		s.ID = uuid.New()
	}
	return nil
}

// PendingInvite is a share addressed to a phone number that has no account yet.
// On phone verification it is resolved into a ListShare (see auto-attach).
// Unique per (list, phone) while unconsumed.
type PendingInvite struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ListID     uuid.UUID  `gorm:"column:list_id;type:uuid;not null"`
	Phone      string     `gorm:"column:phone;not null"`
	Role       Role       `gorm:"column:role;type:share_role;not null"`
	InvitedBy  uuid.UUID  `gorm:"column:invited_by;type:uuid;not null"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time
}

func (PendingInvite) TableName() string { return "pending_invites" }

func (p *PendingInvite) BeforeCreate(tx *gorm.DB) error {
	if p.ID == (uuid.UUID{}) {
		p.ID = uuid.New()
	}
	return nil
}
