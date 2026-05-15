package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// EmployeeRepository defines data access for the HR profile of a user.
type EmployeeRepository interface {
	Create(ctx context.Context, e *models.Employee) error
	Update(ctx context.Context, e *models.Employee) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Employee, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Employee, error)
	FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Employee, error)
}

type employeeRepository struct{ db *gorm.DB }

// NewEmployeeRepository constructs a Postgres-backed EmployeeRepository.
func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, e *models.Employee) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *employeeRepository) Update(ctx context.Context, e *models.Employee) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *employeeRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&models.Employee{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]interface{}{"is_deleted": true, "deleted_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("employee not found or already deleted")
	}
	return nil
}

func (r *employeeRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *employeeRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&e, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *employeeRepository) FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("User", notDeleted).
		First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}
