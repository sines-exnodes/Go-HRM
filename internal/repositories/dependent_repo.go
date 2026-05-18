package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// DependentRepository defines data access for the dependents of an employee.
type DependentRepository interface {
	Create(ctx context.Context, d *models.Dependent) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Dependent, error)
	ListByEmployee(ctx context.Context, employeeID uuid.UUID, page, pageSize int) ([]models.Dependent, int64, error)
	Update(ctx context.Context, id uuid.UUID, fields map[string]any) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type dependentRepository struct {
	db *gorm.DB
}

// NewDependentRepository constructs a Postgres-backed DependentRepository.
func NewDependentRepository(db *gorm.DB) DependentRepository {
	return &dependentRepository{db: db}
}

func (r *dependentRepository) Create(ctx context.Context, d *models.Dependent) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *dependentRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Dependent, error) {
	var d models.Dependent
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &d, err
}

func (r *dependentRepository) ListByEmployee(ctx context.Context, employeeID uuid.UUID, page, pageSize int) ([]models.Dependent, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	tx := r.db.WithContext(ctx).Model(&models.Dependent{}).
		Where("employee_id = ? AND is_deleted = ?", employeeID, false)

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var deps []models.Dependent
	if err := tx.
		Order("created_at ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&deps).Error; err != nil {
		return nil, 0, err
	}
	return deps, total, nil
}

func (r *dependentRepository) Update(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&models.Dependent{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(fields).Error
}

func (r *dependentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.Dependent{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}
