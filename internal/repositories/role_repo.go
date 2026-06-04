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

// RoleFilter mirrors dto.RoleListQuery in a service-agnostic shape.
type RoleFilter struct {
	Page     int
	PageSize int
	Search   string
}

// RoleRepository defines data access for roles.
type RoleRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Role, error)
	Create(ctx context.Context, role *models.Role) error
	Update(ctx context.Context, role *models.Role) error
	List(ctx context.Context, f RoleFilter) ([]models.Role, int64, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// CountUsersWithRole counts non-deleted user_roles rows referencing the role.
	CountUsersWithRole(ctx context.Context, id uuid.UUID) (int64, error)
}

type roleRepository struct{ db *gorm.DB }

// NewRoleRepository constructs a Postgres-backed RoleRepository.
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func notDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

func (r *roleRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&role, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&role, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Role, error) {
	if len(ids) == 0 {
		return []models.Role{}, nil
	}
	var roles []models.Role
	err := r.db.WithContext(ctx).Scopes(notDeleted).Where("id IN ?", ids).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) List(ctx context.Context, f RoleFilter) ([]models.Role, int64, error) {
	q := r.db.WithContext(ctx).Scopes(notDeleted).Model(&models.Role{})
	if s := strings.TrimSpace(f.Search); s != "" {
		q = q.Where("name ILIKE ?", utils.BuildILIKEPattern(s))
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
	var items []models.Role
	err := q.
		Order("level ASC").
		Order("LOWER(name) ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error
	return items, total, err
}

func (r *roleRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Role{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

// CountUsersWithRole counts live user_roles rows referencing the role. The
// join table carries its own is_deleted column (migration 000002).
func (r *roleRepository) CountUsersWithRole(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Where("role_id = ? AND is_deleted = ?", id, false).
		Count(&count).Error
	return count, err
}
