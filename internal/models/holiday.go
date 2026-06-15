package models

import "time"

// Holiday is a company-specific public holiday record.
type Holiday struct {
	BaseModel
	Year     int       `gorm:"not null;index"      json:"year"`
	Name     string    `gorm:"not null"            json:"name"`
	FromDate time.Time `gorm:"type:date;not null"  json:"from_date"`
	ToDate   time.Time `gorm:"type:date;not null"  json:"to_date"`
}

func (Holiday) TableName() string { return "holidays" }

// HolidayTemplate is a read-only preset used for importing holidays.
// Seeded in migration 000023; never mutated at runtime.
type HolidayTemplate struct {
	BaseModel
	Year     int       `gorm:"not null;index"      json:"year"`
	Name     string    `gorm:"not null"            json:"name"`
	FromDate time.Time `gorm:"type:date;not null"  json:"from_date"`
	ToDate   time.Time `gorm:"type:date;not null"  json:"to_date"`
}

func (HolidayTemplate) TableName() string { return "holiday_templates" }
