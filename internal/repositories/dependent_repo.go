package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// DependentRepository is the Postgres-backed data access for the dependents of
// an employee. (Reconciled with the Phase 1 stub: the Phase 1 version exposed
// an interface with a non-paginated ListByEmployee and a struct-based Update;
// this Phase 2 version is a superset — paginated list + field-map update — and
// has no interface consumers, so it replaces the stub directly.)
type DependentRepository struct {
	db *gorm.DB
}

func NewDependentRepository(db *gorm.DB) *DependentRepository {
	return &DependentRepository{db: db}
}

func (r *DependentRepository) Create(ctx context.Context, d *models.Dependent) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *DependentRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Dependent, error) {
	var d models.Dependent
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &d, err
}

func (r *DependentRepository) ListByEmployee(ctx context.Context, employeeID uuid.UUID, page, pageSize int) ([]models.Dependent, int64, error) {
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

func (r *DependentRepository) Update(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&models.Dependent{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(fields).Error
}

func (r *DependentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.Dependent{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}
