package models

import (
	"github.com/google/uuid"
)

// EmployeeEmergencyContact is one emergency-contact row for an employee (1-N).
// Replaces the former single emergency_contact_{name,relation,phone} columns
// (migration 000017) to match the Python list shape (employees parity #4).
type EmployeeEmergencyContact struct {
	BaseModel
	EmployeeID   uuid.UUID `gorm:"type:uuid;not null;index" json:"employee_id"`
	FullName     string    `gorm:"type:text;not null" json:"full_name"`
	Relationship string    `gorm:"type:text;not null;default:''" json:"relationship"`
	PhoneNumber  string    `gorm:"type:text;not null;default:''" json:"phone_number"`
}

func (EmployeeEmergencyContact) TableName() string { return "employee_emergency_contacts" }
