package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType is the kind of event a notification represents. It is a
// stored, extensible value with no DB CHECK constraint — DR Rule 15 wants a
// new type to be a data change, not a schema migration. Validation lives
// here, mirroring AnnouncementStatus.
type NotificationType string

const (
	NotificationTypeAnnouncement NotificationType = "announcement"
	NotificationTypeLeaveRequest NotificationType = "leave_request"
)

// IsValid reports whether t is a known notification type.
func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeAnnouncement, NotificationTypeLeaveRequest:
		return true
	}
	return false
}

// Notification is one delivered event for one recipient.
//
// Title and Body are a SNAPSHOT taken at creation time (DR Rule 12) — the
// list is a faithful log of what the employee was told, even if the source
// record is later edited.
//
// SourceID has no foreign key on purpose: it points into announcements or
// leave_requests depending on Type, and the row must outlive its source
// (DR Rule 13).
//
// UserID targets users(id), not employees(id) — notifications are an
// auth-level surface, the same exception made for AnnouncementView.
type Notification struct {
	BaseModel
	UserID   uuid.UUID        `gorm:"type:uuid;not null;index"        json:"user_id"`
	Type     NotificationType `gorm:"type:text;not null"              json:"type"`
	Title    string           `gorm:"type:text;not null"              json:"title"`
	Body     string           `gorm:"type:text;not null;default:''"   json:"body"`
	SourceID uuid.UUID        `gorm:"type:uuid;not null"              json:"source_id"`

	// ReadAt is nil while unread. Read is terminal (DR Rule 8) — nothing
	// ever sets this back to nil. Storing the timestamp rather than a bool
	// costs the same and records when.
	ReadAt *time.Time `json:"read_at,omitempty"`
}

func (Notification) TableName() string { return "notifications" }
