package models

import (
	"time"

	"github.com/google/uuid"
)

// SystemConfigSingletonID is the fixed UUID of the single allowed row in
// the system_config table. The DB enforces this via the
// system_config_singleton CHECK constraint; this constant is used by the
// seed service to upsert the sentinel row and by the repo to read/update
// it.
var SystemConfigSingletonID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// SystemConfig is the singleton organization-wide configuration row.
//
// This struct intentionally does NOT embed BaseModel — the singleton has
// a fixed sentinel UUID, no soft-delete semantics, and no list / get-by-id
// flow. The four audit columns are declared inline to keep the row shape
// consistent with every other entity (spec §5.2), but the repo never
// applies the NotDeleted scope (REVISION NOTES #10).
type SystemConfig struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// Attendance — late arrival threshold
	LateThresholdHour   int16 `gorm:"not null;default:9"  json:"late_threshold_hour"`
	LateThresholdMinute int16 `gorm:"not null;default:0"  json:"late_threshold_minute"`

	// Attendance — checkout threshold (early-leave calc, reserved)
	CheckoutThresholdHour   int16 `gorm:"not null;default:18" json:"checkout_threshold_hour"`
	CheckoutThresholdMinute int16 `gorm:"not null;default:0"  json:"checkout_threshold_minute"`

	// Company profile
	CompanyAddress          *string    `gorm:"type:text"                 json:"company_address,omitempty"`
	CompanyLatitude         *float64   `                                 json:"company_latitude,omitempty"`
	CompanyLongitude        *float64   `                                 json:"company_longitude,omitempty"`
	CompanyAddressUpdatedAt *time.Time `                                 json:"company_address_updated_at,omitempty"`
	CompanyAddressUpdatedBy *uuid.UUID `gorm:"type:uuid"                 json:"company_address_updated_by,omitempty"`

	// Audit columns — present for schema parity but not driven by
	// soft-delete logic.
	CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	IsDeleted bool       `gorm:"not null;default:false" json:"-"`
	DeletedAt *time.Time `                                json:"-"`
}

// TableName pins to system_config (GORM would otherwise pluralize the
// type name into the wrong identifier).
func (SystemConfig) TableName() string { return "system_config" }
