package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// RoleRepository defines data access for roles.
type RoleRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Role, error)
	Create(ctx context.Context, role *models.Role) error
	Update(ctx context.Context, role *models.Role) error
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
