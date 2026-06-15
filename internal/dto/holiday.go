package dto

import (
	"time"

	"github.com/google/uuid"
)

// HolidayCreate is the POST /holidays request body.
type HolidayCreate struct {
	Year     int       `json:"year"      binding:"required,min=2000,max=2100"`
	Name     string    `json:"name"      binding:"required,max=100"`
	FromDate time.Time `json:"from_date" binding:"required"`
	ToDate   time.Time `json:"to_date"   binding:"required"`
}

// HolidayUpdate is the PATCH /holidays/:id request body (all optional).
type HolidayUpdate struct {
	Name     *string    `json:"name"      binding:"omitempty,max=100"`
	FromDate *time.Time `json:"from_date"`
	ToDate   *time.Time `json:"to_date"`
}

// HolidayListQuery is the GET /holidays query string.
type HolidayListQuery struct {
	Year     int    `form:"year"`
	Search   string `form:"search"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// HolidayRead is the canonical response shape for a single holiday.
type HolidayRead struct {
	ID        uuid.UUID `json:"id"`
	Year      int       `json:"year"`
	Name      string    `json:"name"`
	FromDate  time.Time `json:"from_date"`
	ToDate    time.Time `json:"to_date"`
	TotalDays int       `json:"total_days"` // computed: (to_date - from_date).Days + 1
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HolidayTemplateRead is the response shape for a preset template entry.
type HolidayTemplateRead struct {
	ID        uuid.UUID `json:"id"`
	Year      int       `json:"year"`
	Name      string    `json:"name"`
	FromDate  time.Time `json:"from_date"`
	ToDate    time.Time `json:"to_date"`
	TotalDays int       `json:"total_days"`
}

// HolidayImportRequest is the POST /holidays/import request body.
type HolidayImportRequest struct {
	Year        int         `json:"year"         binding:"required,min=2000,max=2100"`
	TemplateIDs []uuid.UUID `json:"template_ids" binding:"required,gt=0,dive,required"`
}

// HolidayImportResult is the POST /holidays/import response data.
type HolidayImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}
