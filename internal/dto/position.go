package dto

import (
	"time"

	"github.com/google/uuid"
)

type PositionCreate struct {
	Name         string    `json:"name"                  binding:"required,min=1,max=100"`
	Description  string    `json:"description,omitempty" binding:"max=1000"`
	DepartmentID uuid.UUID `json:"department_id"         binding:"required"`
}

type PositionUpdate struct {
	Name         *string    `json:"name,omitempty"          binding:"omitempty,min=1,max=100"`
	Description  *string    `json:"description,omitempty"   binding:"omitempty,max=1000"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
}

type PositionRead struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	DepartmentID uuid.UUID       `json:"department_id"`
	Department   *DepartmentRead `json:"department,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type PositionListQuery struct {
	Page     int    `form:"page,default=1"       binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Search   string `form:"search"`
	// DepartmentID is bound as a raw string (gin's query binding cannot
	// populate *uuid.UUID) and parsed in PositionService.List, mirroring
	// DepartmentListQuery.ParentID.
	DepartmentID string `form:"department_id"`
}
