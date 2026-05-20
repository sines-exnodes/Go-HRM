package models

import (
	"time"

	"github.com/google/uuid"
)

// LeaveType is the string enum for leave_requests.leave_type.
// Mirrors Python's LeaveType StrEnum.
type LeaveType string

// LeavePeriod is the string enum for leave_requests.leave_period.
// Mirrors Python's LeavePeriod StrEnum.
type LeavePeriod string

// LeaveStatus is the string enum for leave_requests.status.
// Mirrors Python's LeaveStatus StrEnum.
type LeaveStatus string

const (
	LeaveTypeAnnual    LeaveType = "annual"
	LeaveTypeSick      LeaveType = "sick"
	LeaveTypePersonal  LeaveType = "personal"
	LeaveTypeMaternity LeaveType = "maternity"
	LeaveTypeUnpaid    LeaveType = "unpaid"
)

const (
	LeavePeriodFullDay       LeavePeriod = "full_day"
	LeavePeriodMorningHalf   LeavePeriod = "morning_half"
	LeavePeriodAfternoonHalf LeavePeriod = "afternoon_half"
)

const (
	LeaveStatusPending   LeaveStatus = "pending"
	LeaveStatusApproved  LeaveStatus = "approved"
	LeaveStatusRejected  LeaveStatus = "rejected"
	LeaveStatusCancelled LeaveStatus = "cancelled"
)

// IsQuotaLeaveType reports whether the type deducts from a stored quota.
// Mirrors Python's QUOTA_TYPES = {annual, sick}. Personal/maternity/unpaid
// have no quota and never trigger the insufficient-quota warning.
func IsQuotaLeaveType(t LeaveType) bool {
	return t == LeaveTypeAnnual || t == LeaveTypeSick
}

// IsHalfDayPeriod reports whether the period is a half-day variant.
func IsHalfDayPeriod(p LeavePeriod) bool {
	return p == LeavePeriodMorningHalf || p == LeavePeriodAfternoonHalf
}

// LeaveRequest maps to the leave_requests table. EmployeeID and CreatedBy
// both reference employees(id) — the HR profile — per the Go schema split
// established in Phase 1. CreatedBy defaults to EmployeeID when the
// creator is also the subject; an admin acting on behalf of another
// employee sets EmployeeID = subject and CreatedBy = admin.
type LeaveRequest struct {
	BaseModel
	EmployeeID    uuid.UUID   `gorm:"type:uuid;not null;index"                       json:"employee_id"`
	FromDate      time.Time   `gorm:"type:date;not null"                              json:"from_date"`
	ToDate        time.Time   `gorm:"type:date;not null"                              json:"to_date"`
	LeavePeriod   LeavePeriod `gorm:"type:text;not null;default:'full_day'"           json:"leave_period"`
	LeaveType     LeaveType   `gorm:"type:text;not null"                              json:"leave_type"`
	TotalDays     float64     `gorm:"type:numeric(5,1);not null"                      json:"total_days"`
	Reason        string      `gorm:"type:text;not null"                              json:"reason"`
	AttachmentURL *string     `gorm:"type:text"                                       json:"attachment_url,omitempty"`
	Status        LeaveStatus `gorm:"type:text;not null;default:'pending';index"      json:"status"`
	CreatedBy     uuid.UUID   `gorm:"type:uuid;not null"                              json:"created_by"`
}

// TableName pins the entity to the leave_requests table so GORM doesn't
// pluralize the type name into the wrong identifier.
func (LeaveRequest) TableName() string { return "leave_requests" }
