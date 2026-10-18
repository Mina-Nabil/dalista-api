package userModels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role is the coarse account role. Fine-grained access is permission-string
// based (see permissions.go + the permission registry); roles map to default
// permission sets via RoleDefaults().
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User is the core identity record. Email is optional; phone is unique when
// set and becomes required (verified) after signup regardless of sign-in
// method. Credentials and provider links live in AuthIdentity, not here.
type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email         *string   `gorm:"column:email"`
	Phone         *string   `gorm:"column:phone"`
	Role          Role      `gorm:"column:role;type:user_role;not null;default:user"`
	PhoneVerified bool      `gorm:"column:phone_verified;not null;default:false"`
	IsActive      bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Permissions is loaded per-request by the auth middleware; never persisted.
	Permissions []string `gorm:"-"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == (uuid.UUID{}) {
		u.ID = uuid.New()
	}
	return nil
}

// Can reports whether the user holds the given permission string.
func (u *User) Can(permission string) bool {
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}
