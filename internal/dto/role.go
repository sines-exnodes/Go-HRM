package dto

import (
	"time"

	"github.com/google/uuid"
)

// RoleCreate is the request body for POST /api/v1/roles.
type RoleCreate struct {
	Name        string   `json:"name"                  binding:"required,min=1,max=100"`
	Description string   `json:"description,omitempty" binding:"max=1000"`
	Level       int      `json:"level"                 binding:"required,min=1,max=100"`
	Permissions []string `json:"permissions,omitempty"`
}

// RoleUpdate is the PATCH body for /api/v1/roles/:id — pointer fields, only
// provided fields change. Permissions is a pointer so "omitted" (no change)
// is distinguishable from "[]" (revoke all).
type RoleUpdate struct {
	Name        *string   `json:"name,omitempty"        binding:"omitempty,min=1,max=100"`
	Description *string   `json:"description,omitempty" binding:"omitempty,max=1000"`
	Level       *int      `json:"level,omitempty"       binding:"omitempty,min=1,max=100"`
	Permissions *[]string `json:"permissions,omitempty"`
}

// RoleRead is the wire shape returned by every role endpoint. Superset of the
// two Python list shapes: full permissions[] AND permission_count.
type RoleRead struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Level           int       `json:"level"`
	Permissions     []string  `json:"permissions"`
	PermissionCount int       `json:"permission_count"`
	IsSystem        bool      `json:"is_system"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RoleListQuery binds the querystring for GET /api/v1/roles.
type RoleListQuery struct {
	Page     int    `form:"page,default=1"       binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Search   string `form:"search"`
}
