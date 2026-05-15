package models

import (
	"time"
)

// User maps to the users table. Auth-only — HR / profile fields live on
// Employee (one-to-one via user_id).
type User struct {
	BaseModel
	Email           string     `gorm:"type:citext;not null;uniqueIndex" json:"email"`
	PasswordHash    string     `gorm:"type:text;not null" json:"-"`
	IsActive        bool       `gorm:"not null;default:true" json:"is_active"`
	EmailChangedAt  *time.Time `json:"email_changed_at,omitempty"`
	PasswordResetAt *time.Time `json:"password_reset_at,omitempty"`

	// Many-to-many via user_roles join table.
	Roles []Role `gorm:"many2many:user_roles;joinForeignKey:user_id;joinReferences:role_id" json:"roles,omitempty"`

	// One-to-one with Employee — preloaded by handlers that need the HR
	// profile in the response shape (e.g., login). Pointer so an
	// un-preloaded User does not carry a zero-valued empty Employee.
	Employee *Employee `gorm:"foreignKey:UserID" json:"employee,omitempty"`
}

func (User) TableName() string { return "users" }
