package dto

import (
	"time"

	"github.com/google/uuid"
)

// ---- Attendance settings ----

// AttendanceSettingsRead is the response payload for GET
// /organization-settings/attendance.
type AttendanceSettingsRead struct {
	LateThresholdHour       int `json:"late_threshold_hour"`
	LateThresholdMinute     int `json:"late_threshold_minute"`
	CheckoutThresholdHour   int `json:"checkout_threshold_hour"`
	CheckoutThresholdMinute int `json:"checkout_threshold_minute"`
}

// AttendanceSettingsUpdate is the body for PATCH
// /organization-settings/attendance. Pointer types preserve
// "not provided" semantics; nil fields are left unchanged. Numeric
// bounds match the DB CHECK constraints.
type AttendanceSettingsUpdate struct {
	LateThresholdHour       *int `json:"late_threshold_hour,omitempty"       binding:"omitempty,min=0,max=23"`
	LateThresholdMinute     *int `json:"late_threshold_minute,omitempty"     binding:"omitempty,min=0,max=59"`
	CheckoutThresholdHour   *int `json:"checkout_threshold_hour,omitempty"   binding:"omitempty,min=0,max=23"`
	CheckoutThresholdMinute *int `json:"checkout_threshold_minute,omitempty" binding:"omitempty,min=0,max=59"`
}

// ---- Company profile ----

// CompanyProfileRead is the response payload for GET
// /organization-settings/company-profile. UpdatedBy resolves to the
// employee's full name when available (best-effort projection — null
// when the FK is nil or the row is gone).
type CompanyProfileRead struct {
	CompanyAddress          *string    `json:"company_address,omitempty"`
	CompanyLatitude         *float64   `json:"company_latitude,omitempty"`
	CompanyLongitude        *float64   `json:"company_longitude,omitempty"`
	CompanyAddressUpdatedAt *time.Time `json:"company_address_updated_at,omitempty"`
	CompanyAddressUpdatedBy *uuid.UUID `json:"company_address_updated_by,omitempty"`
	UpdatedByName           *string    `json:"updated_by_name,omitempty"`
}

// CompanyProfileUpdate is the body for PATCH
// /organization-settings/company-profile. Pointer types preserve
// "not provided" semantics. Lat/lng have geographic bounds enforced
// by binding tags.
type CompanyProfileUpdate struct {
	CompanyAddress   *string  `json:"company_address,omitempty"`
	CompanyLatitude  *float64 `json:"company_latitude,omitempty"  binding:"omitempty,min=-90,max=90"`
	CompanyLongitude *float64 `json:"company_longitude,omitempty" binding:"omitempty,min=-180,max=180"`
}
