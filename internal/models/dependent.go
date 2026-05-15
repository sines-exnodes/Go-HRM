package models

import (
	"time"

	"github.com/google/uuid"
)

// Dependent maps to the dependents table — people supported by an employee.
type Dependent struct {
	BaseModel
	EmployeeID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"employee_id"`
	FullName     string     `gorm:"type:text;not null" json:"full_name"`
	DOB          *time.Time `gorm:"type:date" json:"dob,omitempty"`
	Gender       *string    `gorm:"type:text" json:"gender,omitempty"`      // male / female / other
	Relationship string     `gorm:"type:text;not null" json:"relationship"` // child / parent / spouse / other
}

func (Dependent) TableName() string { return "dependents" }
