package dto

import (
	"time"

	"github.com/google/uuid"
)

// NotificationListQuery is the GET /mobile/notifications query string.
//
// No filter or sort params: DR Rule 16 makes the screen read-and-navigate
// only, and DR Rule 10 fixes the sort to newest-first.
type NotificationListQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// NotificationRead is the canonical response shape for one notification.
//
// Type is returned as a bare string; the mobile app owns the icon/label
// registry from DR section 3. Shipping icon names over the wire would couple
// the API to a client icon set.
//
// Timestamps are RFC3339 UTC. AC-16's "20 July, 2026 - 10:00 AM" in employee
// local time is client-side formatting — the server has no reliable notion of
// the employee's timezone.
type NotificationRead struct {
	ID        uuid.UUID  `json:"id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	SourceID  uuid.UUID  `json:"source_id"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// NotificationUnreadCountRead backs the dashboard header bell (DR Rule 14).
type NotificationUnreadCountRead struct {
	UnreadCount int64 `json:"unread_count"`
}
