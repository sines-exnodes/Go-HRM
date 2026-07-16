package models

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetOTP stores a single-use 6-digit code for the mobile
// forgot-password flow (DR-001-001-02). Only the most recently issued code
// for a user is valid — RequestOTP soft-deletes any earlier pending rows
// before inserting a new one (SR-03).
//
// CodeHash is a bcrypt hash: the plaintext code only ever exists in the
// dispatched email. Rows are found by UserID and the candidate code is
// compared against the hash, so a database leak does not hand out live codes.
type PasswordResetOTP struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	CodeHash  string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	// ConsumedAt is stamped when the code is successfully verified. A
	// consumed code cannot be verified again (SR-05).
	ConsumedAt *time.Time
	// AttemptCount counts wrong-code submissions. Once it reaches the
	// configured maximum the row is soft-deleted so the code is dead.
	AttemptCount   int    `gorm:"not null;default:0"`
	LastEmailError string `gorm:"size:512"`

	User *User
}

func (PasswordResetOTP) TableName() string { return "password_reset_otps" }
