package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// UserRepository defines data access for users.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByIDWithRoles(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByIDWithRolesAndEmployee(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailWithRoles(ctx context.Context, email string) (*models.User, error)
	FindByEmailWithRolesAndEmployee(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
}

type userRepository struct{ db *gorm.DB }

// NewUserRepository constructs a Postgres-backed UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByIDWithRoles(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("Roles", notDeleted).
		First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&u, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmailWithRoles(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("Roles", notDeleted).
		First(&u, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByIDWithRolesAndEmployee(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("Roles", notDeleted).
		Preload("Employee", notDeleted).
		First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmailWithRolesAndEmployee(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("Roles", notDeleted).
		Preload("Employee", notDeleted).
		First(&u, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]interface{}{"is_deleted": true, "deleted_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("user not found or already deleted")
	}
	return nil
}

// ReplaceRoles atomically replaces the user's role set.
func (r *userRepository) ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Hard delete is acceptable on the join table — soft delete on a
		// many2many is rarely useful and complicates GORM associations.
		if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		rows := make([]map[string]interface{}, 0, len(roleIDs))
		for _, rid := range roleIDs {
			rows = append(rows, map[string]interface{}{
				"user_id": userID,
				"role_id": rid,
			})
		}
		return tx.Table("user_roles").Create(&rows).Error
	})
}
