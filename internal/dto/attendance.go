package dto

import (
	"time"

	"github.com/google/uuid"
)

// ---------- Requests ----------

// AttendanceCheckInReq is the body for POST /attendance/check-in.
// CheckIn defaults to "now" in the company TZ when omitted. WorkLocation
// is the only enum field — the binding tag enforces the domain.
// GPS fields are only inspected when OFFICE_GPS_ENABLED=true.
type AttendanceCheckInReq struct {
	CheckIn      *time.Time `json:"check_in,omitempty"`
	WorkLocation *string    `json:"work_location,omitempty" binding:"omitempty,oneof=office remote hybrid field"`
	Notes        *string    `json:"notes,omitempty"`
	Latitude     *float64   `json:"latitude,omitempty"`
	Longitude    *float64   `json:"longitude,omitempty"`
	Accuracy     *float64   `json:"accuracy,omitempty"`
}

// AttendanceCheckOutReq is the body for POST /attendance/check-out.
type AttendanceCheckOutReq struct {
	CheckOut *time.Time `json:"check_out,omitempty"`
	Notes    *string    `json:"notes,omitempty"`
}

// AttendanceAdminCreateReq is the admin-only manual-entry body.
// EmployeeID is the SUBJECT (HR profile), per the Go schema split — NOT
// a user ID. Date is YYYY-MM-DD in the company TZ. CheckIn/CheckOut are
// optional; when CheckIn is provided and IsLate is nil, the service
// auto-derives IsLate from CheckIn vs the late threshold.
type AttendanceAdminCreateReq struct {
	EmployeeID   uuid.UUID  `json:"employee_id" binding:"required"`
	Date         string     `json:"date"        binding:"required" example:"2026-05-15"`
	CheckIn      *time.Time `json:"check_in,omitempty"`
	CheckOut     *time.Time `json:"check_out,omitempty"`
	IsLate       *bool      `json:"is_late,omitempty"`
	IsHalfDay    *bool      `json:"is_half_day,omitempty"`
	WorkLocation *string    `json:"work_location,omitempty" binding:"omitempty,oneof=office remote hybrid field"`
	Notes        *string    `json:"notes,omitempty"`
}

// AttendanceAdminUpdateReq patches an existing attendance row. Pointer
// types distinguish "not provided" from "explicit zero". CheckIn/CheckOut
// adjust the FIRST session's times (admin correction); a brand-new session
// is appended when no session exists yet and CheckIn is provided.
type AttendanceAdminUpdateReq struct {
	IsLate       *bool      `json:"is_late,omitempty"`
	IsHalfDay    *bool      `json:"is_half_day,omitempty"`
	WorkLocation *string    `json:"work_location,omitempty" binding:"omitempty,oneof=office remote hybrid field"`
	Notes        *string    `json:"notes,omitempty"`
	CheckIn      *time.Time `json:"check_in,omitempty"`
	CheckOut     *time.Time `json:"check_out,omitempty"`
}

// AttendanceListQuery binds the querystring for GET /attendance.
// EmployeeID + DepartmentID are admin-side filters. Status enum: on_time | late.
type AttendanceListQuery struct {
	Page         int    `form:"page,default=1"       binding:"min=1"`
	PageSize     int    `form:"page_size,default=20" binding:"min=1,max=100"`
	EmployeeID   string `form:"employee_id"`
	DepartmentID string `form:"department_id"`
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
	Status       string `form:"status"`
}

// AttendanceMatrixQuery binds the querystring for GET /attendance/matrix.
// Month/Year default to "now in company TZ". Status is a CSV string (the
// Python source mirrors this) — e.g. "on_time,late,absent".
type AttendanceMatrixQuery struct {
	Month        int    `form:"month"                binding:"omitempty,min=1,max=12"`
	Year         int    `form:"year"                 binding:"omitempty,min=2000"`
	Page         int    `form:"page,default=1"       binding:"min=1"`
	PageSize     int    `form:"page_size,default=20" binding:"min=1,max=100"`
	Search       string `form:"search"`
	DepartmentID string `form:"department_id"`
	Status       string `form:"status"`
}

// ---------- Responses ----------

// AttendanceSessionRead is the per-session projection. HoursWorked is nil
// while the session is still open (CheckOut == nil).
type AttendanceSessionRead struct {
	ID             uuid.UUID  `json:"id"`
	CheckIn        time.Time  `json:"check_in"`
	CheckOut       *time.Time `json:"check_out,omitempty"`
	IsAutoCheckout bool       `json:"is_auto_checkout"`
	HoursWorked    *float64   `json:"hours_worked,omitempty"`
}

// AttendanceEmployeeBrief is the embedded {id, full_name, avatar_url}
// projection on AttendanceRead. Department/Position are inflated when
// the employee row carries those FKs and the referenced rows are live.
type AttendanceEmployeeBrief struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	AvatarURL  *string   `json:"avatar_url,omitempty"`
	Department *AttendanceRefRead `json:"department,omitempty"`
	Position   *AttendanceRefRead `json:"position,omitempty"`
}

// AttendanceRefRead is the minimal {id, name} projection used for embedded
// department/position references on AttendanceEmployeeBrief.
type AttendanceRefRead struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// AttendanceRead is the canonical wire shape returned by every attendance
// endpoint. Date is YYYY-MM-DD in the company TZ. CheckIn/CheckOut are
// first/last across all sessions; HoursWorked is the sum across sessions.
type AttendanceRead struct {
	ID           uuid.UUID                `json:"id"`
	EmployeeID   uuid.UUID                `json:"employee_id"`
	Employee     *AttendanceEmployeeBrief `json:"employee,omitempty"`
	Date         string                   `json:"date"`
	IsLate       bool                     `json:"is_late"`
	IsHalfDay    bool                     `json:"is_half_day"`
	WorkLocation *string                  `json:"work_location,omitempty"`
	Notes        *string                  `json:"notes,omitempty"`
	Sessions     []AttendanceSessionRead  `json:"sessions"`
	CheckIn      *time.Time               `json:"check_in,omitempty"`
	CheckOut     *time.Time               `json:"check_out,omitempty"`
	HoursWorked  *float64                 `json:"hours_worked,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

// TodayStatusRead is the response payload for GET /attendance/today.
// Status enum: not_checked_in | checked_in | checked_out.
// Streak counts consecutive workdays (Mon-Fri) with at least one check-in,
// walking backward from today.
type TodayStatusRead struct {
	Status         string                  `json:"status"`
	IsLate         bool                    `json:"is_late"`
	Sessions       []AttendanceSessionRead `json:"sessions"`
	CurrentCheckIn *time.Time              `json:"current_check_in,omitempty"`
	MonthlyCount   int                     `json:"monthly_count"`
	Streak         int                     `json:"streak"`
}

// ---------- Matrix ----------

// AttendanceCellRead is one day in the matrix. Status enum:
// on_time | late | absent | weekend | no_data.
type AttendanceCellRead struct {
	Date        string                  `json:"date"`
	Day         int                     `json:"day"`
	Status      string                  `json:"status"`
	CheckIn     *time.Time              `json:"check_in,omitempty"`
	CheckOut    *time.Time              `json:"check_out,omitempty"`
	HoursWorked *float64                `json:"hours_worked,omitempty"`
	IsLate      bool                    `json:"is_late"`
	Sessions    []AttendanceSessionRead `json:"sessions,omitempty"`
}

// AttendanceRowRead is one employee's row in the matrix. Cells is keyed
// by day-of-month (1..daysInMonth).
type AttendanceRowRead struct {
	EmployeeID        uuid.UUID                  `json:"employee_id"`
	EmployeeName      string                     `json:"employee_name"`
	AvatarURL         *string                    `json:"avatar_url,omitempty"`
	DepartmentName    *string                    `json:"department_name,omitempty"`
	Cells             map[int]AttendanceCellRead `json:"cells"`
	TotalLateMinutes  int                        `json:"total_late_minutes"`
	TotalEarlyMinutes int                        `json:"total_early_minutes"`
}

// AttendanceMatrixRead is the response payload for GET /attendance/matrix.
type AttendanceMatrixRead struct {
	Year        int                 `json:"year"`
	Month       int                 `json:"month"`
	DaysInMonth int                 `json:"days_in_month"`
	Items       []AttendanceRowRead `json:"items"`
	Total       int                 `json:"total"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
	TotalPages  int                 `json:"total_pages"`
}
