package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// DependentRepository defines data access for the dependents of an employee.
type DependentRepository interface {
	Create(ctx context.Context, d *models.Dependent) error
	Update(ctx context.Context, d *models.Dependent) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Dependent, error)
	ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]models.Dependent, error)
}

type dependentRepository struct{ db *gorm.DB }

// NewDependentRepository constructs a Postgres-backed DependentRepository.
func NewDependentRepository(db *gorm.DB) DependentRepository {
	return &dependentRepository{db: db}
}

func (r *dependentRepository) Create(ctx context.Context, d *models.Dependent) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *dependentRepository) Update(ctx context.Context, d *models.Dependent) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *dependentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&models.Dependent{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]interface{}{"is_deleted": true, "deleted_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("dependent not found or already deleted")
	}
	return nil
}

func (r *dependentRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Dependent, error) {
	var d models.Dependent
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&d, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *dependentRepository) ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]models.Dependent, error) {
	var out []models.Dependent
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Where("employee_id = ?", employeeID).
		Order("created_at ASC").
		Find(&out).Error
	return out, err
}
