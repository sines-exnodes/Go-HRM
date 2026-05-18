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

// DepartmentFilter mirrors dto.DepartmentListQuery in a service-agnostic shape.
// ParentID semantics:
//
//	nil           → no filter
//	&uuid.Nil     → top-level only (parent_id IS NULL)
//	&realUUID     → children of that parent
type DepartmentFilter struct {
	Page     int
	PageSize int
	Search   string
	ParentID *uuid.UUID
}

type DepartmentRepository interface {
	Create(ctx context.Context, d *models.Department) error
	Update(ctx context.Context, d *models.Department) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID, preloadParent bool) (*models.Department, error)
	FindByName(ctx context.Context, name string) (*models.Department, error)
	List(ctx context.Context, f DepartmentFilter) ([]models.Department, int64, error)
	HasChildren(ctx context.Context, id uuid.UUID) (bool, error)
	// CountEmployees counts non-deleted employees whose department_id == id.
	CountEmployees(ctx context.Context, id uuid.UUID) (int64, error)
}

type departmentRepository struct{ db *gorm.DB }

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *departmentRepository) Create(ctx context.Context, d *models.Department) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *departmentRepository) Update(ctx context.Context, d *models.Department) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *departmentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Department{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *departmentRepository) FindByID(ctx context.Context, id uuid.UUID, preloadParent bool) (*models.Department, error) {
	q := r.base(ctx)
	if preloadParent {
		q = q.Preload("Parent", models.NotDeleted)
	}
	var d models.Department
	if err := q.Where("id = ?", id).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

// FindByName returns (nil, nil) when no active row matches — callers treat
// that as "available".
func (r *departmentRepository) FindByName(ctx context.Context, name string) (*models.Department, error) {
	var d models.Department
	err := r.base(ctx).
		Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).
		First(&d).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *departmentRepository) List(ctx context.Context, f DepartmentFilter) ([]models.Department, int64, error) {
	q := r.base(ctx).Model(&models.Department{})
	if s := strings.TrimSpace(f.Search); s != "" {
		q = q.Where("name ILIKE ?", utils.BuildILIKEPattern(s))
	}
	if f.ParentID != nil {
		if *f.ParentID == uuid.Nil {
			q = q.Where("parent_id IS NULL")
		} else {
			q = q.Where("parent_id = ?", *f.ParentID)
		}
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
	var items []models.Department
	err := q.
		Preload("Parent", models.NotDeleted).
		Order("LOWER(name) ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error
	return items, total, err
}

func (r *departmentRepository) HasChildren(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.base(ctx).
		Model(&models.Department{}).
		Where("parent_id = ?", id).
		Count(&count).Error
	return count > 0, err
}

func (r *departmentRepository) CountEmployees(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Employee{}).
		Where("department_id = ? AND is_deleted = ?", id, false).
		Count(&count).Error
	return count, err
}
