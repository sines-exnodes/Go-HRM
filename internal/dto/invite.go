package dto

import (
	"time"

	"github.com/google/uuid"
)

// ---- Write inputs ----

// InviteCreate is the body for POST /invites. RoleIDs may be empty; the
// service defaults to the "Employee" role on accept when nothing is
// specified.
type InviteCreate struct {
	Email        string      `json:"email"     binding:"required,email"`
	FullName     *string     `json:"full_name,omitempty"`
	RoleIDs      []uuid.UUID `json:"role_ids,omitempty"`
	DepartmentID *uuid.UUID  `json:"department_id,omitempty"`
	PositionID   *uuid.UUID  `json:"position_id,omitempty"`
}

// InviteListQuery binds the querystring for GET /invites.
type InviteListQuery struct {
	Page     int    `form:"page,default=1"        binding:"min=1"`
	PageSize int    `form:"page_size,default=20"  binding:"min=1,max=100"`
	Email    string `form:"email"`
	Status   string `form:"status"` // pending|accepted|expired|revoked
}

// InviteAccept is the public body for POST /invites/accept. The token
// comes from the email link; the invitee supplies a password +
// (optional) full_name override.
type InviteAccept struct {
	Token    string  `json:"token"     binding:"required"`
	Password string  `json:"password"  binding:"required,min=8"`
	FullName *string `json:"full_name,omitempty"`
}

// ---- Read outputs ----

// InviteInviterBrief is the embedded {id, full_name} projection. The
// inviter is an Employee row per the schema split.
type InviteInviterBrief struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
}

// InviteRead is the canonical wire shape. Status is derived (pending /
// accepted / expired / revoked).
type InviteRead struct {
	ID             uuid.UUID            `json:"id"`
	Email          string               `json:"email"`
	FullName       *string              `json:"full_name,omitempty"`
	RoleIDs        []uuid.UUID          `json:"role_ids"`
	DepartmentID   *uuid.UUID           `json:"department_id,omitempty"`
	PositionID     *uuid.UUID           `json:"position_id,omitempty"`
	ExpiresAt      time.Time            `json:"expires_at"`
	AcceptedAt     *time.Time           `json:"accepted_at,omitempty"`
	AcceptedUserID *uuid.UUID           `json:"accepted_user_id,omitempty"`
	Status         string               `json:"status"`
	InvitedBy      uuid.UUID            `json:"invited_by"`
	Inviter        *InviteInviterBrief  `json:"inviter,omitempty"`
	LastEmailError *string              `json:"last_email_error,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

// InviteAcceptResult is the payload returned by POST /invites/accept —
// the public endpoint that creates the user row.
type InviteAcceptResult struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Message   string    `json:"message"`
}

// ---- Notification test ----

// NotificationTestRequest is the body for POST /notifications/test —
// pushes to the caller's own registered device tokens.
type NotificationTestRequest struct {
	Title string         `json:"title" binding:"required"`
	Body  string         `json:"body"  binding:"required"`
	Data  map[string]any `json:"data,omitempty"`
}

// NotificationTestResult tells the caller how many devices received
// the test push (sent) vs. were skipped (no FCM credentials / send
// errors).
type NotificationTestResult struct {
	Sent    int      `json:"sent"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}
