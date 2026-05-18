package models

import "github.com/google/uuid"

// Position is a job role; belongs to exactly one department.
// Employees reference a position through employees.position_id
// (FK added in migration 000005).
type Position struct {
	BaseModel
	Name         string    `gorm:"type:text;not null"            json:"name"`
	Description  string    `gorm:"type:text;not null;default:''" json:"description"`
	DepartmentID uuid.UUID `gorm:"type:uuid;not null;index"      json:"department_id"`

	Department *Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department,omitempty"`
}

func (Position) TableName() string { return "positions" }
