package models

import (
	"time"

	"github.com/google/uuid"
)

// Attendance maps to the attendance table. One row per (employee_id, date)
// — the unique constraint is enforced by the migration. EmployeeID
// references employees(id) per the Go schema split (HR profile, NOT auth
// account). IsLate is computed once from the FIRST check-in vs the
// configured threshold and never re-evaluated on subsequent sessions
// (REVISION NOTES item #5); IsHalfDay is flipped at check-out time when
// total hours-worked falls below the half-day threshold.
type Attendance struct {
	BaseModel
	EmployeeID   uuid.UUID `gorm:"type:uuid;not null;index"        json:"employee_id"`
	Date         time.Time `gorm:"type:date;not null;index"        json:"date"`
	IsLate       bool      `gorm:"not null;default:false"          json:"is_late"`
	IsHalfDay    bool      `gorm:"not null;default:false"          json:"is_half_day"`
	WorkLocation *string   `gorm:"type:text"                       json:"work_location,omitempty"`
	Notes        *string   `gorm:"type:text"                       json:"notes,omitempty"`

	// Relations — loaded via Preload in the repository. Sessions are
	// ordered by check_in ASC (see repo) so the first/last semantics in
	// the service can read row.Sessions[0] / row.Sessions[len-1] directly.
	Employee *Employee           `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Sessions []AttendanceSession `gorm:"foreignKey:AttendanceID;constraint:OnDelete:CASCADE" json:"sessions,omitempty"`
}

// TableName pins to the attendance table — GORM would otherwise pluralize
// to "attendances".
func (Attendance) TableName() string { return "attendance" }

// AttendanceSession maps to the attendance_sessions table. One per
// check-in/check-out pair; multiple per attendance row are supported
// (lunch back). CheckOut is nullable — an open session represents the
// "currently checked in" state. The partial unique index in the migration
// (uq_attendance_sessions_one_open) prevents more than one open session
// per attendance row.
type AttendanceSession struct {
	BaseModel
	AttendanceID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"attendance_id"`
	CheckIn        time.Time  `gorm:"not null"                  json:"check_in"`
	CheckOut       *time.Time `                                  json:"check_out,omitempty"`
	IsAutoCheckout bool       `gorm:"not null;default:false"    json:"is_auto_checkout"`
}

// TableName pins to the attendance_sessions table.
func (AttendanceSession) TableName() string { return "attendance_sessions" }
