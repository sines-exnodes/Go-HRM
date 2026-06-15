package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// HolidayTemplateRepository provides read-only access to the preset templates.
type HolidayTemplateRepository interface {
	ListByYear(ctx context.Context, year int) ([]models.HolidayTemplate, error)
}

type holidayTemplateRepo struct{ db *gorm.DB }

// NewHolidayTemplateRepository constructs a Postgres-backed HolidayTemplateRepository.
func NewHolidayTemplateRepository(db *gorm.DB) HolidayTemplateRepository {
	return &holidayTemplateRepo{db: db}
}

func (r *holidayTemplateRepo) ListByYear(ctx context.Context, year int) ([]models.HolidayTemplate, error) {
	var rows []models.HolidayTemplate
	err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		Where("year = ?", year).
		Order("from_date ASC").
		Find(&rows).Error
	return rows, err
}
