package userModels

import (
	"time"

	"github.com/google/uuid"
)

// Permission is the canonical list of permission strings known to the app.
type Permission struct {
	Permission string `gorm:"column:permission;primaryKey"`
}

func (Permission) TableName() string { return "permissions" }

// RolePermission maps a role to one of its default permissions.
type RolePermission struct {
	Role       string `gorm:"column:role;primaryKey"`
	Permission string `gorm:"column:permission;primaryKey"`
}

func (RolePermission) TableName() string { return "role_permissions" }

// UserPermission is a per-user grant layered on top of role defaults.
type UserPermission struct {
	UserID     uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey"`
	Permission string    `gorm:"column:permission;primaryKey"`
	GrantedAt  time.Time `gorm:"column:granted_at;not null"`
}

func (UserPermission) TableName() string { return "user_permissions" }
