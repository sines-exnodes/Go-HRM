package models

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken stores a single-use token for the self-service
// forgot-password flow. A user may have at most one active (unused +
// unexpired) token at a time — RequestReset soft-deletes any previous
// pending tokens before inserting a new one.
type PasswordResetToken struct {
	BaseModel
	UserID         uuid.UUID `gorm:"type:uuid;not null;index"`
	Token          string    `gorm:"not null;uniqueIndex;size:256"`
	ExpiresAt      time.Time `gorm:"not null"`
	UsedAt         *time.Time
	LastEmailError string `gorm:"size:512"`

	User *User
}
