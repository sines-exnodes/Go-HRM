package models

import "github.com/google/uuid"

// EmployeeLeaveQuota holds the annual / sick leave quota for one employee.
// Phase 5 (Leave Requests) will read from this table for balance calculation.
type EmployeeLeaveQuota struct {
	BaseModel
	EmployeeID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"employee_id"`
	AnnualLeaveQuota float64   `gorm:"type:numeric(6,2);not null;default:12" json:"annual_leave_quota"`
	SickLeaveQuota   float64   `gorm:"type:numeric(6,2);not null;default:6"  json:"sick_leave_quota"`
}

func (EmployeeLeaveQuota) TableName() string { return "employee_leave_quotas" }
