package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UUIDArray is a typed wrapper around a UUID[] Postgres column. lib/pq's
// pq.StringArray would force callers to cast every element; UUIDArray
// preserves the uuid.UUID type all the way to GORM.
type UUIDArray []uuid.UUID

// Scan implements sql.Scanner. Reads the underlying TEXT[] / UUID[]
// representation produced by lib/pq.
func (a *UUIDArray) Scan(src any) error {
	if src == nil {
		*a = nil
		return nil
	}
	// pq's UUID[] arrives as a string like "{uuid1,uuid2}".
	var raw pq.StringArray
	if err := raw.Scan(src); err != nil {
		return err
	}
	out := make(UUIDArray, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		u, err := uuid.Parse(s)
		if err != nil {
			return fmt.Errorf("UUIDArray.Scan: invalid uuid %q: %w", s, err)
		}
		out = append(out, u)
	}
	*a = out
	return nil
}

// Value implements driver.Valuer. Emits a TEXT[] literal that Postgres
// casts to UUID[] on insert/update.
func (a UUIDArray) Value() (driver.Value, error) {
	strs := make(pq.StringArray, len(a))
	for i, u := range a {
		strs[i] = u.String()
	}
	return strs.Value()
}

// Invite is an admin-generated email invitation. The user row is NOT
// created until the invitee posts the token to /api/v1/invites/accept;
// at that point the InviteService composes the dto.EmployeeCreate from
// these fields and stamps accepted_at + accepted_user_id.
type Invite struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email          string     `gorm:"type:citext;not null;index"                     json:"email"`
	FullName       *string    `gorm:"type:text"                                       json:"full_name,omitempty"`
	Token          string     `gorm:"type:text;not null"                              json:"token"`
	RoleIDs        UUIDArray  `gorm:"type:uuid[];not null;default:'{}'"               json:"role_ids"`
	DepartmentID   *uuid.UUID `gorm:"type:uuid"                                       json:"department_id,omitempty"`
	PositionID     *uuid.UUID `gorm:"type:uuid"                                       json:"position_id,omitempty"`
	ExpiresAt      time.Time  `gorm:"not null"                                        json:"expires_at"`
	AcceptedAt     *time.Time `                                                       json:"accepted_at,omitempty"`
	AcceptedUserID *uuid.UUID `gorm:"type:uuid"                                       json:"accepted_user_id,omitempty"`
	InvitedBy      uuid.UUID  `gorm:"type:uuid;not null"                              json:"invited_by"`
	LastEmailError *string    `gorm:"type:text"                                       json:"last_email_error,omitempty"`

	// Audit columns (declared inline; the row does support soft delete
	// via Revoke — is_deleted=true is the revoke marker).
	CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	IsDeleted bool       `gorm:"not null;default:false" json:"-"`
	DeletedAt *time.Time `                                json:"-"`

	// Optional preload: who invited.
	Inviter *Employee `gorm:"foreignKey:InvitedBy;references:ID" json:"inviter,omitempty"`
}

func (Invite) TableName() string { return "invites" }
