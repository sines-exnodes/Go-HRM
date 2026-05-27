package dto

import (
	"time"

	"github.com/google/uuid"
)

// DepartmentCreate is the request body for POST /api/v1/departments.
type DepartmentCreate struct {
	Name        string     `json:"name"                  binding:"required,min=1,max=100"`
	Description string     `json:"description,omitempty" binding:"max=1000"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

// DepartmentUpdate is the request body for PATCH /api/v1/departments/:id.
// PATCH semantics — only provided fields change. ClearParent makes the
// department a root (distinguishes "no change" from "make root").
type DepartmentUpdate struct {
	Name        *string    `json:"name,omitempty"        binding:"omitempty,min=1,max=100"`
	Description *string    `json:"description,omitempty" binding:"omitempty,max=1000"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	ClearParent bool       `json:"clear_parent,omitempty"`
}

// DepartmentRead is the wire shape returned by every department endpoint.
// employee_count is the number of non-deleted employees whose
// employees.department_id == this department's id (matches Python's
// DepartmentRead).
type DepartmentRead struct {
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	ParentID      *uuid.UUID      `json:"parent_id,omitempty"`
	Parent        *DepartmentRead `json:"parent,omitempty"`
	EmployeeCount int64           `json:"employee_count"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// DepartmentListQuery binds the querystring for GET /api/v1/departments.
// ParentID == "root" (or "null") returns top-level departments only.
type DepartmentListQuery struct {
	Page     int    `form:"page,default=1"       binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Search   string `form:"search"`
	ParentID string `form:"parent_id"`
}
