package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// HolidayListQuery is the filter/pagination spec for List.
type HolidayListQuery struct {
	Year     int
	Search   string
	Page     int
	PageSize int
}

// HolidayRepository defines data access for the holidays table.
type HolidayRepository interface {
	List(ctx context.Context, q HolidayListQuery) ([]models.Holiday, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*models.Holiday, error)
	Create(ctx context.Context, h *models.Holiday) error
	Update(ctx context.Context, h *models.Holiday) error
	Delete(ctx context.Context, id uuid.UUID) error

	// FindInRange returns non-deleted holidays whose range overlaps [from, to].
	// Used at leave creation time and during recalculation.
	FindInRange(ctx context.Context, from, to time.Time) ([]models.Holiday, error)

	// YearsWithHolidays returns distinct years that have at least one non-deleted holiday.
	YearsWithHolidays(ctx context.Context) ([]int, error)

	// ExistsByNameAndYear checks for a duplicate name in the same year.
	// excludeID, if non-nil, skips that row (used during Update).
	ExistsByNameAndYear(ctx context.Context, name string, year int, excludeID *uuid.UUID) (bool, error)
}

type holidayRepo struct{ db *gorm.DB }

// NewHolidayRepository constructs a Postgres-backed HolidayRepository.
func NewHolidayRepository(db *gorm.DB) HolidayRepository {
	return &holidayRepo{db: db}
}

func (r *holidayRepo) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *holidayRepo) List(ctx context.Context, q HolidayListQuery) ([]models.Holiday, int64, error) {
	qb := r.base(ctx).Model(&models.Holiday{}).Where("year = ?", q.Year)
	if q.Search != "" {
		qb = qb.Where("name ILIKE ?", "%"+q.Search+"%")
	}
	var total int64
	if err := qb.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	var rows []models.Holiday
	err := qb.Order("from_date ASC").Limit(q.PageSize).Offset((q.Page - 1) * q.PageSize).Find(&rows).Error
	return rows, total, err
}

func (r *holidayRepo) Get(ctx context.Context, id uuid.UUID) (*models.Holiday, error) {
	var h models.Holiday
	err := r.base(ctx).First(&h, "id = ?", id).Error
	return &h, err
}

func (r *holidayRepo) Create(ctx context.Context, h *models.Holiday) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *holidayRepo) Update(ctx context.Context, h *models.Holiday) error {
	return r.db.WithContext(ctx).
		Model(h).
		Where("id = ? AND is_deleted = false", h.ID).
		Updates(map[string]any{
			"name":      h.Name,
			"from_date": h.FromDate,
			"to_date":   h.ToDate,
		}).Error
}

func (r *holidayRepo) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Model(&models.Holiday{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]any{"is_deleted": true, "deleted_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *holidayRepo) FindInRange(ctx context.Context, from, to time.Time) ([]models.Holiday, error) {
	var rows []models.Holiday
	err := r.base(ctx).
		Where("from_date <= ? AND to_date >= ?", to, from).
		Find(&rows).Error
	return rows, err
}

func (r *holidayRepo) YearsWithHolidays(ctx context.Context) ([]int, error) {
	var years []int
	err := r.base(ctx).
		Model(&models.Holiday{}).
		Distinct("year").
		Order("year ASC").
		Pluck("year", &years).Error
	return years, err
}

func (r *holidayRepo) ExistsByNameAndYear(ctx context.Context, name string, year int, excludeID *uuid.UUID) (bool, error) {
	qb := r.base(ctx).Model(&models.Holiday{}).
		Where("year = ? AND LOWER(name) = LOWER(?)", year, name)
	if excludeID != nil {
		qb = qb.Where("id != ?", *excludeID)
	}
	var count int64
	err := qb.Count(&count).Error
	return count > 0, err
}
