package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/pkg/utils"
)

type PositionFilter struct {
	Page         int
	PageSize     int
	Search       string
	DepartmentID *uuid.UUID
}

type PositionRepository interface {
	Create(ctx context.Context, p *models.Position) error
	Update(ctx context.Context, p *models.Position) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID, preloadDept bool) (*models.Position, error)
	FindByNameInDept(ctx context.Context, name string, departmentID uuid.UUID) (*models.Position, error)
	List(ctx context.Context, f PositionFilter) ([]models.Position, int64, error)
	CountByDepartment(ctx context.Context, departmentID uuid.UUID) (int64, error)
	// CountEmployees counts non-deleted employees whose position_id == id.
	CountEmployees(ctx context.Context, id uuid.UUID) (int64, error)
}

type positionRepository struct{ db *gorm.DB }

func NewPositionRepository(db *gorm.DB) PositionRepository {
	return &positionRepository{db: db}
}

func (r *positionRepository) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *positionRepository) Create(ctx context.Context, p *models.Position) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *positionRepository) Update(ctx context.Context, p *models.Position) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *positionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Position{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *positionRepository) FindByID(ctx context.Context, id uuid.UUID, preloadDept bool) (*models.Position, error) {
	q := r.base(ctx)
	if preloadDept {
		q = q.Preload("Department", models.NotDeleted)
	}
	var p models.Position
	if err := q.Where("id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *positionRepository) FindByNameInDept(ctx context.Context, name string, departmentID uuid.UUID) (*models.Position, error) {
	var p models.Position
	err := r.base(ctx).
		Where("LOWER(name) = LOWER(?) AND department_id = ?", strings.TrimSpace(name), departmentID).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *positionRepository) List(ctx context.Context, f PositionFilter) ([]models.Position, int64, error) {
	q := r.base(ctx).Model(&models.Position{})
	if s := strings.TrimSpace(f.Search); s != "" {
		q = q.Where("name ILIKE ?", utils.BuildILIKEPattern(s))
	}
	if f.DepartmentID != nil {
		q = q.Where("department_id = ?", *f.DepartmentID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 10
	}
	var items []models.Position
	err := q.
		Preload("Department", models.NotDeleted).
		Order("LOWER(name) ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error
	return items, total, err
}

func (r *positionRepository) CountByDepartment(ctx context.Context, departmentID uuid.UUID) (int64, error) {
	var count int64
	err := r.base(ctx).
		Model(&models.Position{}).
		Where("department_id = ?", departmentID).
		Count(&count).Error
	return count, err
}

func (r *positionRepository) CountEmployees(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Employee{}).
		Where("position_id = ? AND is_deleted = ?", id, false).
		Count(&count).Error
	return count, err
}
