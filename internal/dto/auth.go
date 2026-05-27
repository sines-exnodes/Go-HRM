package dto

import "github.com/google/uuid"

// LoginRequest is the body for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@exnodes.vn"`
	Password string `json:"password" binding:"required,min=1" example:"ChangeMe!2026"`
	// RememberMe, when true, requests a long-lived refresh token (see
	// REMEMBER_ME_REFRESH_TOKEN_EXPIRE_DAYS, default 30 days) instead of
	// the default refresh-token TTL.
	RememberMe bool `json:"remember_me,omitempty" example:"false"`
}

// RefreshRequest is the body for POST /api/v1/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOi..."`
}

// EmployeeSummary is the slice of Employee fields embedded in auth responses.
// The full Employee resource lives under /api/v1/employees.
type EmployeeSummary struct {
	ID           uuid.UUID  `json:"id"`
	FullName     string     `json:"full_name"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	PositionID   *uuid.UUID `json:"position_id,omitempty"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
}

// RoleSummary is the slim role projection used in auth responses.
type RoleSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	IsSystem    bool      `json:"is_system"`
	Permissions []string  `json:"permissions"`
}

// UserSummary is the user shape returned by the login/refresh endpoints. It
// embeds the HR profile so the frontend never has to fetch full_name from
// /users separately.
type UserSummary struct {
	ID       uuid.UUID        `json:"id"`
	Email    string           `json:"email"`
	IsActive bool             `json:"is_active"`
	Employee *EmployeeSummary `json:"employee,omitempty"`
	Roles    []RoleSummary    `json:"roles"`
}

// LoginResponse is the body of a successful login or refresh.
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type" example:"Bearer"`
	User         UserSummary `json:"user"`
}

// LogoutResponse is the body of a logout call (currently empty acknowledgement).
type LogoutResponse struct {
	Message string `json:"message" example:"Logged out"`
}
