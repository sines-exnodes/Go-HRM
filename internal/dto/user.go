package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserMeRead is the auth-profile + embedded employee summary + roles returned by GET /users/me.
type UserMeRead struct {
	ID                   uuid.UUID        `json:"id"`
	Email                string           `json:"email"`
	IsActive             bool             `json:"is_active"`
	Roles                []RoleRead       `json:"roles"`
	NotificationsEnabled bool             `json:"notifications_enabled"`
	Employee             *EmployeeSummary `json:"employee,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

// ---- Auth-side requests (live under /api/v1/users/me* and /api/v1/users/:id*) ----

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"omitempty,min=1"`
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
}

type ChangeEmailRequest struct {
	NewEmail        string `json:"new_email"        binding:"required,email"`
	CurrentPassword string `json:"current_password" binding:"required,min=1"`
}

type RoleAssignmentRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids" binding:"required"`
}

type FcmTokenRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	Token    string `json:"token"     binding:"required"`
	Platform string `json:"platform"  binding:"omitempty,oneof=android ios web unknown"`
}

type NotificationSettingsRequest struct {
	NotificationsEnabled bool `json:"notifications_enabled"`
}

// Admin toggle on /api/v1/users/:id — only is_active for now.
type AdminUserPatch struct {
	IsActive *bool `json:"is_active,omitempty"`
}

type DeleteUserRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=1"`
}
