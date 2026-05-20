package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/models"
)

// ---- Leave request: write inputs ----

// LeaveRequestCreate is the JSON body inside the multipart `data` field on
// POST /api/v1/leave-requests. EmployeeID is admin-only — when present and
// non-empty, the creator must hold PermLeaveManage; the request is recorded
// against that subject (employee_id) while created_by stays as the creator.
type LeaveRequestCreate struct {
	EmployeeID  *uuid.UUID         `json:"employee_id,omitempty"`
	FromDate    time.Time          `json:"from_date"    binding:"required"`
	ToDate      time.Time          `json:"to_date"      binding:"required"`
	LeavePeriod models.LeavePeriod `json:"leave_period" binding:"required"`
	LeaveType   models.LeaveType   `json:"leave_type"   binding:"required"`
	Reason      string             `json:"reason"       binding:"required,min=1"`
}

// LeaveRequestUpdate is the JSON body inside the multipart `data` field on
// PATCH /api/v1/leave-requests/:id. Pointer types distinguish "not provided"
// from "explicit zero". When the current row is approved and an admin patches
// it, the service reverts status to pending (Python contract).
type LeaveRequestUpdate struct {
	FromDate    *time.Time          `json:"from_date,omitempty"`
	ToDate      *time.Time          `json:"to_date,omitempty"`
	LeavePeriod *models.LeavePeriod `json:"leave_period,omitempty"`
	LeaveType   *models.LeaveType   `json:"leave_type,omitempty"`
	Reason      *string             `json:"reason,omitempty"`
}

// ---- Leave request: read outputs ----

// LeaveRefRead is the minimal {id, name} projection used for embedded
// references (employee, department, position) on a LeaveRequestRead. Null
// when the FK is nil or the referenced row is missing/soft-deleted.
type LeaveRefRead struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LeaveRequestRead is the canonical wire shape returned by every leave
// endpoint. The embedded employee/department/position projections mirror
// the Python source's response, which pre-resolves them so the FE doesn't
// need a follow-up lookup.
type LeaveRequestRead struct {
	ID            string             `json:"id"`
	Employee      *LeaveRefRead      `json:"employee,omitempty"`
	Department    *LeaveRefRead      `json:"department,omitempty"`
	Position      *LeaveRefRead      `json:"position,omitempty"`
	FromDate      time.Time          `json:"from_date"`
	ToDate        time.Time          `json:"to_date"`
	LeavePeriod   models.LeavePeriod `json:"leave_period"`
	LeaveType     models.LeaveType   `json:"leave_type"`
	TotalDays     float64            `json:"total_days"`
	Reason        string             `json:"reason"`
	AttachmentURL *string            `json:"attachment_url,omitempty"`
	Status        models.LeaveStatus `json:"status"`
	CreatedBy     string             `json:"created_by"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// LeaveRequestWriteResult is the response envelope for Create and Update.
// Warnings are non-blocking; the request was still created/updated. Empty
// slice (not null) when there are no warnings — the FE relies on this for
// safe iteration.
type LeaveRequestWriteResult struct {
	Request  LeaveRequestRead `json:"request"`
	Warnings []string         `json:"warnings"`
}

// ---- Balance + dashboard ----

// LeaveBalanceSummary aggregates an employee's quota usage for a given
// calendar year. Quota figures come from employee_leave_quotas; used
// figures come from SUM(total_days) over status='approved' AND
// is_deleted=false leave_requests in the year.
type LeaveBalanceSummary struct {
	Year            int     `json:"year"`
	AnnualQuota     float64 `json:"annual_quota"`
	AnnualUsed      float64 `json:"annual_used"`
	AnnualRemaining float64 `json:"annual_remaining"`
	SickQuota       float64 `json:"sick_quota"`
	SickUsed        float64 `json:"sick_used"`
	SickRemaining   float64 `json:"sick_remaining"`
	LeavesThisYear  int     `json:"leaves_this_year"`
}

// LeaveDashboardRead is the payload for GET /leave-requests/dashboard/me.
// Upcoming = pending/approved with from_date >= today.
// History = to_date < today.
type LeaveDashboardRead struct {
	Balance  LeaveBalanceSummary `json:"balance"`
	Upcoming []LeaveRequestRead  `json:"upcoming"`
	History  []LeaveRequestRead  `json:"history"`
}

// ---- List queries ----

// LeaveListQuery binds the querystring for GET /api/v1/leave-requests.
// Status is repeat-param (?status=pending&status=approved).
type LeaveListQuery struct {
	Page         int      `form:"page,default=1"        binding:"min=1"`
	PageSize     int      `form:"page_size,default=10"  binding:"min=1,max=100"`
	Search       string   `form:"search"`
	Status       []string `form:"status"`
	DepartmentID string   `form:"department_id"`
	PositionID   string   `form:"position_id"`
}

// LeaveHistoryQuery binds the querystring for GET /api/v1/leave-requests/history/me.
// StartDate/EndDate use the standard YYYY-MM-DD form (matches Python).
type LeaveHistoryQuery struct {
	Page      int        `form:"page,default=1"       binding:"min=1"`
	PageSize  int        `form:"page_size,default=10" binding:"min=1,max=100"`
	Status    []string   `form:"status"`
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"end_date"   time_format:"2006-01-02"`
}
