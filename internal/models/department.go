package models

import "github.com/google/uuid"

// Department is an org unit. Self-referential tree via ParentID
// (nullable; FK ON DELETE SET NULL). Employees reference a department
// through employees.department_id (FK added in migration 000005).
type Department struct {
	BaseModel
	Name        string     `gorm:"type:text;not null"            json:"name"`
	Description string     `gorm:"type:text;not null;default:''" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index"               json:"parent_id,omitempty"`

	// Relations — preloaded on demand, omitted from JSON when nil.
	Parent   *Department  `gorm:"foreignKey:ParentID;references:ID" json:"parent,omitempty"`
	Children []Department `gorm:"foreignKey:ParentID;references:ID" json:"children,omitempty"`
}

func (Department) TableName() string { return "departments" }
