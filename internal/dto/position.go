package dto

import (
	"time"

	"github.com/google/uuid"
)

type PositionCreate struct {
	Name        string `json:"name"                  binding:"required,min=1,max=100"`
	Description string `json:"description,omitempty" binding:"max=1000"`
}

type PositionUpdate struct {
	Name        *string `json:"name,omitempty"        binding:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// PositionRead is the wire shape returned by every position endpoint.
// employee_count is the number of non-deleted employees currently
// referencing this position (matches Python's PositionRead).
type PositionRead struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	EmployeeCount int64     `json:"employee_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PositionListQuery struct {
	Page     int    `form:"page,default=1"       binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Search   string `form:"search"`
}
